package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

// CreateWithdrawal inserts a WITHDRAWAL transaction + detail, debits wallet atomically.
func (t *TransactionRepo) CreateWithdrawal(ctx context.Context, walletID uuid.UUID, amount, adminFee int64, bank model.Withdrawal) (model.Transaction, error) {
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("CreateWithdrawal begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	var balance int64
	if err = tx.QueryRow(ctx, `SELECT balance FROM wallets WHERE id = $1 FOR UPDATE`, walletID).Scan(&balance); err != nil {
		return model.Transaction{}, fmt.Errorf("CreateWithdrawal lock wallet: %w", err)
	}
	if balance < amount+adminFee {
		return model.Transaction{}, errors.New("insufficient balance")
	}
	var txRow model.Transaction
	err = tx.QueryRow(ctx, `
		INSERT INTO transactions (wallet_id, type, amount, admin_fee, status)
		VALUES ($1, 'WITHDRAWAL', $2, $3, 'PENDING')
		RETURNING id, wallet_id, type, amount, admin_fee, status, idempotency_key, note, created_at, updated_at`,
		walletID, amount, adminFee,
	).Scan(&txRow.ID, &txRow.WalletID, &txRow.Type, &txRow.Amount, &txRow.AdminFee, &txRow.Status, &txRow.IdempotencyKey, &txRow.Note, &txRow.CreatedAt, &txRow.UpdatedAt)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("CreateWithdrawal insert transaction: %w", err)
	}
	if _, err = tx.Exec(ctx,
		`INSERT INTO withdrawals (transaction_id, bank_name, account_number, account_holder) VALUES ($1, $2, $3, $4)`,
		txRow.ID, bank.BankName, bank.AccountNumber, bank.AccountHolder,
	); err != nil {
		return model.Transaction{}, fmt.Errorf("CreateWithdrawal insert detail: %w", err)
	}
	if _, err = tx.Exec(ctx,
		`UPDATE wallets SET balance = balance - $1, updated_at = now() WHERE id = $2`,
		amount+adminFee, walletID,
	); err != nil {
		return model.Transaction{}, fmt.Errorf("CreateWithdrawal debit wallet: %w", err)
	}
	if _, err = tx.Exec(ctx,
		`UPDATE transactions SET status = 'SUCCESS', updated_at = now() WHERE id = $1`, txRow.ID,
	); err != nil {
		return model.Transaction{}, fmt.Errorf("CreateWithdrawal update status: %w", err)
	}
	txRow.Status = model.TransactionStatusSuccess
	if err = tx.Commit(ctx); err != nil {
		return model.Transaction{}, fmt.Errorf("CreateWithdrawal commit: %w", err)
	}
	return txRow, nil
}
