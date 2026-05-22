package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type TransactionRepo struct {
	db *pgxpool.Pool
}

func NewTransactionRepo(db *pgxpool.Pool) *TransactionRepo {
	return &TransactionRepo{db: db}
}

// CreateTransaction inserts a new transaction row and returns the created record.
func (t *TransactionRepo) CreateTransaction(ctx context.Context, tx model.Transaction) (model.Transaction, error) {
	sql := `
		INSERT INTO transactions (wallet_id, amount, direction, admin_fee, status, note)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, wallet_id, amount, direction, admin_fee, status, note, created_at
	`
	var created model.Transaction
	err := t.db.QueryRow(ctx, sql,
		tx.WalletID,
		tx.Amount,
		tx.Direction,
		tx.AdminFee,
		tx.Status,
		tx.Note,
	).Scan(
		&created.ID,
		&created.WalletID,
		&created.Amount,
		&created.Direction,
		&created.AdminFee,
		&created.Status,
		&created.Note,
		&created.CreatedAt,
	)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("CreateTransaction: %w", err)
	}
	return created, nil
}

// GetTransactionsByWallet returns paginated transactions for a wallet belonging to the given user email.
func (t *TransactionRepo) GetTransactionsByWallet(ctx context.Context, email string, walletID uuid.UUID, page, limit int) ([]model.Transaction, int, error) {
	offset := (page - 1) * limit

	// Verify wallet belongs to user
	var wid uuid.UUID
	checkSQL := `
		SELECT w.id FROM wallets w
		JOIN users u ON w.user_id = u.id
		WHERE u.email = $1 AND w.id = $2
	`
	if err := t.db.QueryRow(ctx, checkSQL, email, walletID).Scan(&wid); err != nil {
		return nil, 0, fmt.Errorf("GetTransactionsByWallet: wallet not found or access denied: %w", err)
	}

	var total int
	if err := t.db.QueryRow(ctx, `SELECT COUNT(*) FROM transactions WHERE wallet_id = $1`, walletID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("GetTransactionsByWallet count: %w", err)
	}

	rows, err := t.db.Query(ctx, `
		SELECT id, wallet_id, amount, direction, admin_fee, status, note, created_at
		FROM transactions
		WHERE wallet_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, walletID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("GetTransactionsByWallet query: %w", err)
	}
	defer rows.Close()

	var result []model.Transaction
	for rows.Next() {
		var tx model.Transaction
		if err := rows.Scan(
			&tx.ID, &tx.WalletID, &tx.Amount, &tx.Direction,
			&tx.AdminFee, &tx.Status, &tx.Note, &tx.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("GetTransactionsByWallet scan: %w", err)
		}
		result = append(result, tx)
	}
	return result, total, nil
}

// GetAllTransactions returns paginated transactions across all wallets for the user.
func (t *TransactionRepo) GetAllTransactions(ctx context.Context, email string, page, limit int) ([]model.Transaction, int, error) {
	offset := (page - 1) * limit

	var total int
	countSQL := `
		SELECT COUNT(*) FROM transactions tr
		JOIN wallets w ON tr.wallet_id = w.id
		JOIN users u ON w.user_id = u.id
		WHERE u.email = $1
	`
	if err := t.db.QueryRow(ctx, countSQL, email).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("GetAllTransactions count: %w", err)
	}

	rows, err := t.db.Query(ctx, `
		SELECT tr.id, tr.wallet_id, tr.amount, tr.direction, tr.admin_fee, tr.status, tr.note, tr.created_at
		FROM transactions tr
		JOIN wallets w ON tr.wallet_id = w.id
		JOIN users u ON w.user_id = u.id
		WHERE u.email = $1
		ORDER BY tr.created_at DESC
		LIMIT $2 OFFSET $3
	`, email, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("GetAllTransactions query: %w", err)
	}
	defer rows.Close()

	var result []model.Transaction
	for rows.Next() {
		var tx model.Transaction
		if err := rows.Scan(
			&tx.ID, &tx.WalletID, &tx.Amount, &tx.Direction,
			&tx.AdminFee, &tx.Status, &tx.Note, &tx.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("GetAllTransactions scan: %w", err)
		}
		result = append(result, tx)
	}
	return result, total, nil
}

// GetTransactionByID returns a single transaction, verifying it belongs to the user.
func (t *TransactionRepo) GetTransactionByID(ctx context.Context, email string, transactionID uuid.UUID) (model.Transaction, error) {
	var tx model.Transaction
	err := t.db.QueryRow(ctx, `
		SELECT tr.id, tr.wallet_id, tr.amount, tr.direction, tr.admin_fee, tr.status, tr.note, tr.created_at
		FROM transactions tr
		JOIN wallets w ON tr.wallet_id = w.id
		JOIN users u ON w.user_id = u.id
		WHERE u.email = $1 AND tr.id = $2
	`, email, transactionID).Scan(
		&tx.ID, &tx.WalletID, &tx.Amount, &tx.Direction,
		&tx.AdminFee, &tx.Status, &tx.Note, &tx.CreatedAt,
	)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("GetTransactionByID: %w", err)
	}
	return tx, nil
}
