package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

// CreateExpense inserts an EXPENSE transaction + detail, debits wallet atomically.
func (t *TransactionRepo) CreateExpense(ctx context.Context, walletID uuid.UUID, amount, adminFee int64, category, merchantName, note *string) (model.Transaction, error) {
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("CreateExpense begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	var balance int64
	if err = tx.QueryRow(ctx, `SELECT balance FROM wallets WHERE id = $1 FOR UPDATE`, walletID).Scan(&balance); err != nil {
		return model.Transaction{}, fmt.Errorf("CreateExpense lock wallet: %w", err)
	}
	if balance < amount+adminFee {
		return model.Transaction{}, errors.New("insufficient balance")
	}
	var txRow model.Transaction
	err = tx.QueryRow(ctx, `
		INSERT INTO transactions (wallet_id, type, amount, admin_fee, status, note)
		VALUES ($1, 'EXPENSE', $2, $3, 'SUCCESS', $4)
		RETURNING id, wallet_id, type, amount, admin_fee, status, idempotency_key, note, created_at, updated_at`,
		walletID, amount, adminFee, note,
	).Scan(&txRow.ID, &txRow.WalletID, &txRow.Type, &txRow.Amount, &txRow.AdminFee, &txRow.Status, &txRow.IdempotencyKey, &txRow.Note, &txRow.CreatedAt, &txRow.UpdatedAt)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("CreateExpense insert transaction: %w", err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO expenses (transaction_id, category, merchant_name) VALUES ($1, $2, $3)`, txRow.ID, category, merchantName); err != nil {
		return model.Transaction{}, fmt.Errorf("CreateExpense insert detail: %w", err)
	}
	if _, err = tx.Exec(ctx, `UPDATE wallets SET balance = balance - $1, updated_at = now() WHERE id = $2`, amount+adminFee, walletID); err != nil {
		return model.Transaction{}, fmt.Errorf("CreateExpense debit wallet: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return model.Transaction{}, fmt.Errorf("CreateExpense commit: %w", err)
	}
	return txRow, nil
}
