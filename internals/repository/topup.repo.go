package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

// CreateTopup creates a PENDING topup record.
func (t *TransactionRepo) CreateTopup(ctx context.Context, req model.Topup) (model.Topup, error) {
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return model.Topup{}, fmt.Errorf("CreateTopup begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	var topup model.Topup
	err = tx.QueryRow(ctx, `
		INSERT INTO topups (wallet_id, amount, status, payment_method)
		VALUES ($1, $2, $3, $4)
		RETURNING id, wallet_id, amount, status, payment_method, external_reference, created_at`,
		req.WalletID, req.Amount, model.TransactionStatusPending, req.PaymentMethod,
	).Scan(&topup.ID, &topup.WalletID, &topup.Amount, &topup.Status, &topup.PaymentMethod, &topup.ExternalReference, &topup.CreatedAt)
	if err != nil {
		return model.Topup{}, fmt.Errorf("CreateTopup insert: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return model.Topup{}, fmt.Errorf("CreateTopup commit: %w", err)
	}
	return topup, nil
}

func (t *TransactionRepo) UpdateTopupPayment(ctx context.Context, topupID uuid.UUID, externalRef string, metadata []byte) error {
	_, err := t.db.Exec(ctx, `
		UPDATE topups
		SET external_reference = $2, payment_metadata = $3, updated_at = now()
		WHERE id = $1`,
		topupID, externalRef, metadata,
	)
	if err != nil {
		return fmt.Errorf("UpdateTopupPayment: %w", err)
	}
	return nil
}

func (t *TransactionRepo) SettleTopup(ctx context.Context, topupID uuid.UUID, metadata []byte) (bool, uuid.UUID, error) {
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return false, uuid.Nil, fmt.Errorf("SettleTopup begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var (
		status   model.TransactionStatus
		walletID uuid.UUID
		amount   int64
	)
	err = tx.QueryRow(ctx, `
		SELECT status, wallet_id, amount
		FROM topups
		WHERE id = $1
		FOR UPDATE`, topupID,
	).Scan(&status, &walletID, &amount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, uuid.Nil, errors.New("topup not found")
		}
		return false, uuid.Nil, fmt.Errorf("SettleTopup select: %w", err)
	}

	if status == model.TransactionStatusSuccess {
		if err = tx.Commit(ctx); err != nil {
			return false, uuid.Nil, fmt.Errorf("SettleTopup commit: %w", err)
		}
		return false, walletID, nil
	}
	if status != model.TransactionStatusPending {
		return false, uuid.Nil, fmt.Errorf("topup cannot be settled from status %s", status)
	}

	if _, err = tx.Exec(ctx, `
		UPDATE topups
		SET status = $2, payment_metadata = $3, updated_at = now()
		WHERE id = $1`,
		topupID, model.TransactionStatusSuccess, metadata,
	); err != nil {
		return false, uuid.Nil, fmt.Errorf("SettleTopup update topup: %w", err)
	}

	if _, err = tx.Exec(ctx, `
		UPDATE wallets
		SET balance = balance + $1, updated_at = now()
		WHERE id = $2`,
		amount, walletID,
	); err != nil {
		return false, uuid.Nil, fmt.Errorf("SettleTopup credit wallet: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return false, uuid.Nil, fmt.Errorf("SettleTopup commit: %w", err)
	}
	return true, walletID, nil
}

func (t *TransactionRepo) UpdateTopupStatus(ctx context.Context, topupID uuid.UUID, status model.TransactionStatus, metadata []byte) error {
	var currentStatus model.TransactionStatus
	err := t.db.QueryRow(ctx, `SELECT status FROM topups WHERE id = $1`, topupID).Scan(&currentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("topup not found")
		}
		return fmt.Errorf("UpdateTopupStatus select: %w", err)
	}
	if currentStatus != model.TransactionStatusPending {
		return nil
	}

	_, err = t.db.Exec(ctx, `
		UPDATE topups
		SET status = $2, payment_metadata = $3, updated_at = now()
		WHERE id = $1`,
		topupID, status, metadata,
	)
	if err != nil {
		return fmt.Errorf("UpdateTopupStatus: %w", err)
	}
	return nil
}

func (t *TransactionRepo) GetTopupWalletID(ctx context.Context, topupID uuid.UUID) (uuid.UUID, error) {
	var walletID uuid.UUID
	err := t.db.QueryRow(ctx, `SELECT wallet_id FROM topups WHERE id = $1`, topupID).Scan(&walletID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, errors.New("topup not found")
		}
		return uuid.Nil, fmt.Errorf("GetTopupWalletID: %w", err)
	}
	return walletID, nil
}
