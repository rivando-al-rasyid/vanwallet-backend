package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type Authrepo struct {
	db *pgxpool.Pool
}

func NewAuthRepo(db *pgxpool.Pool) *Authrepo {
	return &Authrepo{db: db}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func tokenDigest(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}

// Register creates a user, profile, and default wallet atomically in one transaction.
func (a *Authrepo) Register(ctx context.Context, email, hashpwd string) (model.User, error) {
	email = normalizeEmail(email)

	tx, err := a.db.Begin(ctx)
	if err != nil {
		return model.User{}, fmt.Errorf("Register begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	var user model.User

	err = tx.QueryRow(ctx,
		`INSERT INTO users (email, password) VALUES ($1, $2)
		 RETURNING id, email, created_at, updated_at`,
		email, hashpwd,
	).Scan(&user.ID, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return model.User{}, fmt.Errorf("Register insert user: %w", err)
	}

	fullName := strings.Split(email, "@")[0]

	if _, err = tx.Exec(ctx,
		`INSERT INTO profiles (user_id, full_name) VALUES ($1, $2)`,
		user.ID,
		fullName,
	); err != nil {
		return model.User{}, fmt.Errorf("Register insert profile: %w", err)
	}

	if _, err = tx.Exec(ctx,
		`INSERT INTO wallets (user_id, label, balance) VALUES ($1, $2, 0)`,
		user.ID,
		"Main Wallet",
	); err != nil {
		return model.User{}, fmt.Errorf("Register insert default wallet: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return model.User{}, fmt.Errorf("Register commit: %w", err)
	}

	return user, nil
}

func (a *Authrepo) Login(ctx context.Context, email string) (model.User, error) {
	email = normalizeEmail(email)

	var user model.User
	err := a.db.QueryRow(ctx,
		`SELECT id, email, password, created_at, updated_at FROM users WHERE email = $1`, email,
	).Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, errors.New("user not found")
		}
		return model.User{}, err
	}
	return user, nil
}

// GetUserInfo returns account identity, profile summary, wallet balance, and PIN status.
// Used by GET /auth/me.
func (a *Authrepo) GetUserInfo(ctx context.Context, email string) (model.Profile, error) {
	email = normalizeEmail(email)

	var profile model.Profile
	err := a.db.QueryRow(ctx, `
		SELECT
			p.full_name,
			p.phone,
			p.photo,
			COALESCE(SUM(w.balance), 0) AS current_balance,
			CASE
				WHEN COALESCE(up.pin_hash, '') = '' THEN NULL
				ELSE 'set'
			END AS pin_hash
		FROM users u
		JOIN profiles p ON p.user_id = u.id
		LEFT JOIN wallets w ON w.user_id = u.id
		LEFT JOIN user_pins up ON up.user_id = u.id
		WHERE u.email = $1
		GROUP BY p.full_name, p.phone, p.photo, up.pin_hash
		LIMIT 1`, email,
	).Scan(
		&profile.FullName,
		&profile.Phone,
		&profile.Photo,
		&profile.CurrentBalance,
		&profile.PinHash,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Profile{}, errors.New("user info not found")
		}
		return model.Profile{}, fmt.Errorf("GetUserInfo: %w", err)
	}

	return profile, nil
}

// SaveToken inserts a hashed token row into the tokens table.
// The raw JWT/opaque token is never persisted directly.
func (a *Authrepo) SaveToken(ctx context.Context, userID uuid.UUID, rawToken string, tokenType model.TokenType, expiresAt time.Time) error {
	_, err := a.db.Exec(ctx,
		`INSERT INTO tokens (user_id, token, type, expires_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (token) DO NOTHING`,
		userID, tokenDigest(rawToken), tokenType, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("SaveToken: %w", err)
	}
	return nil
}

// RevokeToken marks a specific token as revoked.
func (a *Authrepo) RevokeToken(ctx context.Context, rawToken string) error {
	_, err := a.db.Exec(ctx,
		`UPDATE tokens SET is_revoked = true WHERE token = $1`, tokenDigest(rawToken),
	)
	if err != nil {
		return fmt.Errorf("RevokeToken: %w", err)
	}
	return nil
}

// IsTokenValid returns true if the token exists, is not revoked, and has not expired.
func (a *Authrepo) IsTokenValid(ctx context.Context, rawToken string) (bool, error) {
	var valid bool
	err := a.db.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM tokens
			WHERE token = $1
			  AND is_revoked = false
			  AND expires_at > now()
		)`, tokenDigest(rawToken),
	).Scan(&valid)
	if err != nil {
		return false, fmt.Errorf("IsTokenValid: %w", err)
	}
	return valid, nil
}

// GetUserByResetToken validates that rawToken is a live PASSWORD_RESET token and
// returns the associated user. The token is revoked immediately (single-use) so
// it cannot be replayed.
func (a *Authrepo) GetUserByResetToken(ctx context.Context, rawToken string) (model.User, error) {
	var user model.User
	digest := tokenDigest(rawToken)

	err := a.db.QueryRow(ctx, `
		SELECT u.id, u.email
		FROM tokens t
		JOIN users u ON t.user_id = u.id
		WHERE t.token     = $1
		  AND t.type      = $2
		  AND t.is_revoked = false
		  AND t.expires_at > now()
		LIMIT 1
	`, digest, model.TokenTypePasswordReset,
	).Scan(&user.ID, &user.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, errors.New("invalid or expired reset token")
		}
		return model.User{}, fmt.Errorf("GetUserByResetToken: %w", err)
	}

	// Revoke immediately — single-use token
	if _, err = a.db.Exec(ctx,
		`UPDATE tokens SET is_revoked = true WHERE token = $1`, digest,
	); err != nil {
		return model.User{}, fmt.Errorf("GetUserByResetToken revoke: %w", err)
	}

	return user, nil
}

// UpdatePassword sets a new hashed password for the given user ID.
func (a *Authrepo) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	result, err := a.db.Exec(ctx,
		`UPDATE users SET password = $1, updated_at = NOW() WHERE id = $2`,
		hashedPassword, userID,
	)
	if err != nil {
		return fmt.Errorf("UpdatePassword: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	return nil
}
