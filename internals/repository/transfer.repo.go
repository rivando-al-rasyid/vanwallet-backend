package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

// CreateTransfer executes a peer-to-peer transfer atomically.
// Returns (transfer, senderTx, recipientTx, error). Controller sends only senderTx to caller.
func (t *TransactionRepo) CreateTransfer(ctx context.Context, senderWalletID, recipientWalletID uuid.UUID, amount, adminFee int64, note *string) (model.Transfer, model.Transaction, model.Transaction, error) {
	if senderWalletID == recipientWalletID {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, errors.New("cannot transfer to the same wallet")
	}

	tx, err := t.db.Begin(ctx)
	if err != nil {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, fmt.Errorf("CreateTransfer begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	lockFirstID, lockSecondID := senderWalletID, recipientWalletID
	if lockSecondID.String() < lockFirstID.String() {
		lockFirstID, lockSecondID = lockSecondID, lockFirstID
	}

	var lockedFirstID uuid.UUID
	if err = tx.QueryRow(ctx, `SELECT id FROM wallets WHERE id = $1 FOR UPDATE`, lockFirstID).Scan(&lockedFirstID); err != nil {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, fmt.Errorf("CreateTransfer lock first wallet: %w", err)
	}
	var lockedSecondID uuid.UUID
	if err = tx.QueryRow(ctx, `SELECT id FROM wallets WHERE id = $1 FOR UPDATE`, lockSecondID).Scan(&lockedSecondID); err != nil {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, fmt.Errorf("CreateTransfer lock second wallet: %w", err)
	}

	var senderBalance int64
	if err = tx.QueryRow(ctx, `SELECT balance FROM wallets WHERE id = $1`, senderWalletID).Scan(&senderBalance); err != nil {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, fmt.Errorf("CreateTransfer read sender balance: %w", err)
	}
	if senderBalance < amount+adminFee {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, errors.New("insufficient balance")
	}

	var senderTx model.Transaction
	err = tx.QueryRow(ctx, `
		INSERT INTO transactions (wallet_id, type, amount, admin_fee, status, note)
		VALUES ($1, 'TRANSFER_OUT', $2, $3, 'SUCCESS', $4)
		RETURNING id, wallet_id, type, amount, admin_fee, status, idempotency_key, note, created_at, updated_at`,
		senderWalletID, amount, adminFee, note,
	).Scan(&senderTx.ID, &senderTx.WalletID, &senderTx.Type, &senderTx.Amount, &senderTx.AdminFee, &senderTx.Status, &senderTx.IdempotencyKey, &senderTx.Note, &senderTx.CreatedAt, &senderTx.UpdatedAt)
	if err != nil {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, fmt.Errorf("CreateTransfer TRANSFER_OUT: %w", err)
	}

	var recipientTx model.Transaction
	err = tx.QueryRow(ctx, `
		INSERT INTO transactions (wallet_id, type, amount, admin_fee, status, note)
		VALUES ($1, 'TRANSFER_IN', $2, 0, 'SUCCESS', $3)
		RETURNING id, wallet_id, type, amount, admin_fee, status, idempotency_key, note, created_at, updated_at`,
		recipientWalletID, amount, note,
	).Scan(&recipientTx.ID, &recipientTx.WalletID, &recipientTx.Type, &recipientTx.Amount, &recipientTx.AdminFee, &recipientTx.Status, &recipientTx.IdempotencyKey, &recipientTx.Note, &recipientTx.CreatedAt, &recipientTx.UpdatedAt)
	if err != nil {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, fmt.Errorf("CreateTransfer TRANSFER_IN: %w", err)
	}

	transferCode := fmt.Sprintf("TRF-%s", senderTx.ID.String()[:8])
	var transfer model.Transfer
	err = tx.QueryRow(ctx, `
		INSERT INTO transfers (transaction_id, recipient_transaction_id, transfer_code)
		VALUES ($1, $2, $3)
		RETURNING transaction_id, recipient_transaction_id, transfer_code, created_at`,
		senderTx.ID, recipientTx.ID, transferCode,
	).Scan(&transfer.TransactionID, &transfer.RecipientTransactionID, &transfer.TransferCode, &transfer.CreatedAt)
	if err != nil {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, fmt.Errorf("CreateTransfer link: %w", err)
	}

	if _, err = tx.Exec(ctx, `UPDATE wallets SET balance = balance - $1, updated_at = now() WHERE id = $2`, amount+adminFee, senderWalletID); err != nil {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, fmt.Errorf("CreateTransfer debit: %w", err)
	}
	if _, err = tx.Exec(ctx, `UPDATE wallets SET balance = balance + $1, updated_at = now() WHERE id = $2`, amount, recipientWalletID); err != nil {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, fmt.Errorf("CreateTransfer credit: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, fmt.Errorf("CreateTransfer commit: %w", err)
	}
	return transfer, senderTx, recipientTx, nil
}
