package repository

import (
	"context"
	"fmt"

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
