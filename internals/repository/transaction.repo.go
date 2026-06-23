package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/pkg"
)

type TransactionRepo struct {
	db *pgxpool.Pool
}

func NewTransactionRepo(db *pgxpool.Pool) *TransactionRepo {
	return &TransactionRepo{db: db}
}

const (
	maxPINFailedAttempts = 5
	pinLockDuration      = 5 * time.Minute
)

// VerifyPIN fetches the user's stored argon2 PIN hash and compares it with rawPin.
// It also enforces brute-force protection using failed_attempts and locked_until.
func (t *TransactionRepo) VerifyPIN(ctx context.Context, email, rawPin string) error {
	var (
		pinID          uuid.UUID
		pinHash        string
		failedAttempts int
		lockedUntil    *time.Time
	)

	err := t.db.QueryRow(ctx, `
		SELECT up.id, COALESCE(up.pin_hash, ''), up.failed_attempts, up.locked_until
		FROM user_pins up
		JOIN users u ON up.user_id = u.id
		WHERE u.email = $1`, email,
	).Scan(&pinID, &pinHash, &failedAttempts, &lockedUntil)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("pin not set")
		}
		return fmt.Errorf("VerifyPIN query: %w", err)
	}

	if pinHash == "" {
		return errors.New("pin not set")
	}

	if lockedUntil != nil && lockedUntil.After(time.Now()) {
		return errors.New("pin is temporarily locked")
	}

	var hc pkg.HashConfig
	hc.UseRecommended()
	if err := hc.Compare(rawPin, pinHash); err != nil {
		failedAttempts++
		var nextLockedUntil *time.Time
		if failedAttempts >= maxPINFailedAttempts {
			locked := time.Now().Add(pinLockDuration)
			nextLockedUntil = &locked
		}

		if _, updateErr := t.db.Exec(ctx, `
			UPDATE user_pins
			SET failed_attempts = $2, locked_until = $3, updated_at = now()
			WHERE id = $1`,
			pinID, failedAttempts, nextLockedUntil,
		); updateErr != nil {
			return fmt.Errorf("VerifyPIN update failed attempts: %w", updateErr)
		}

		if nextLockedUntil != nil {
			return errors.New("invalid pin; pin is temporarily locked")
		}

		return errors.New("invalid pin")
	}

	if _, err := t.db.Exec(ctx, `
		UPDATE user_pins
		SET failed_attempts = 0, locked_until = NULL, last_used_at = now(), updated_at = now()
		WHERE id = $1`, pinID,
	); err != nil {
		return fmt.Errorf("VerifyPIN reset attempts: %w", err)
	}

	return nil
}

// WalletBelongsToUser returns true when walletID is owned by the user identified by email.
func (t *TransactionRepo) WalletBelongsToUser(ctx context.Context, email string, walletID uuid.UUID) (bool, error) {
	var exists bool
	err := t.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM wallets w
			JOIN users u ON w.user_id = u.id
			WHERE u.email = $1 AND w.id = $2
		)`, email, walletID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("WalletBelongsToUser: %w", err)
	}
	return exists, nil
}

func (t *TransactionRepo) GetWalletOwnerEmail(ctx context.Context, walletID uuid.UUID) (string, error) {
	var email string
	err := t.db.QueryRow(ctx, `
		SELECT u.email
		FROM wallets w
		JOIN users u ON w.user_id = u.id
		WHERE w.id = $1`, walletID,
	).Scan(&email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("wallet not found")
		}
		return "", fmt.Errorf("GetWalletOwnerEmail: %w", err)
	}
	return email, nil
}
