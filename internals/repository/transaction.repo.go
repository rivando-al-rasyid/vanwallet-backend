package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type TransactionRepo struct {
	db *pgxpool.Pool
}

func NewTransactionRepo(db *pgxpool.Pool) *TransactionRepo {
	return &TransactionRepo{db: db}
}

// GetSummary returns aggregated balance, income, and expense for the user.
func (t *TransactionRepo) GetSummary(ctx context.Context, email string) (model.TransactionSummary, error) {
	sql := `
	WITH TargetUser AS (
		SELECT id FROM users WHERE email = $1
	),
	UserWallets AS (
		SELECT id, balance FROM wallets WHERE user_id IN (SELECT id FROM TargetUser)
	),
	WalletAgg AS (
		SELECT COALESCE(SUM(balance), 0) AS current_balance
		FROM UserWallets
	),
	TransactionAgg AS (
		SELECT
			COALESCE(SUM(CASE WHEN direction = 'IN'  AND status = 'SUCCESS' THEN amount            ELSE 0 END), 0) AS total_income,
			COALESCE(SUM(CASE WHEN direction = 'OUT' AND status = 'SUCCESS' THEN (amount+admin_fee) ELSE 0 END), 0) AS total_expense
		FROM transactions
		WHERE wallet_id IN (SELECT id FROM UserWallets)
	)
	SELECT
		wa.current_balance,
		ta.total_income,
		ta.total_expense
	FROM TargetUser tu
	CROSS JOIN WalletAgg wa
	CROSS JOIN TransactionAgg ta;`

	var s model.TransactionSummary
	err := t.db.QueryRow(ctx, sql, email).Scan(
		&s.CurrentBalance,
		&s.TotalIncome,
		&s.TotalExpense,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.TransactionSummary{}, errors.New("user profile not found")
		}
		return model.TransactionSummary{}, err
	}
	return s, nil
}

// GetTransactionReport returns daily buckets for the last 7 days,
// or weekly buckets for the last 30 days, for all wallets owned by the user.
// rangeParam must be "7days" or "30days".
func (t *TransactionRepo) GetTransactionReport(ctx context.Context, email string, rangeParam string) ([]model.ChartPoint, error) {
	var sql string

	switch rangeParam {
	case "30days":
		sql = `
		WITH UserWallets AS (
			SELECT w.id
			FROM wallets w
			JOIN users u ON w.user_id = u.id
			WHERE u.email = $1
		)
		SELECT
			'W' || TO_CHAR(DATE_TRUNC('week', created_at), 'IW') AS label,
			COALESCE(SUM(CASE WHEN direction = 'IN'  AND status = 'SUCCESS' THEN amount             ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN direction = 'OUT' AND status = 'SUCCESS' THEN (amount+admin_fee) ELSE 0 END), 0) AS expense
		FROM transactions
		WHERE
			wallet_id IN (SELECT id FROM UserWallets)
			AND created_at >= NOW() - INTERVAL '30 days'
		GROUP BY DATE_TRUNC('week', created_at)
		ORDER BY DATE_TRUNC('week', created_at);`

	default: // "7days"
		sql = `
		WITH UserWallets AS (
			SELECT w.id
			FROM wallets w
			JOIN users u ON w.user_id = u.id
			WHERE u.email = $1
		)
		SELECT
			TO_CHAR(DATE_TRUNC('day', created_at), 'Dy') AS label,
			COALESCE(SUM(CASE WHEN direction = 'IN'  AND status = 'SUCCESS' THEN amount             ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN direction = 'OUT' AND status = 'SUCCESS' THEN (amount+admin_fee) ELSE 0 END), 0) AS expense
		FROM transactions
		WHERE
			wallet_id IN (SELECT id FROM UserWallets)
			AND created_at >= NOW() - INTERVAL '7 days'
		GROUP BY DATE_TRUNC('day', created_at)
		ORDER BY DATE_TRUNC('day', created_at);`
	}

	rows, err := t.db.Query(ctx, sql, email)
	if err != nil {
		return nil, fmt.Errorf("GetTransactionReport: %w", err)
	}
	defer rows.Close()

	var result []model.ChartPoint
	for rows.Next() {
		var cp model.ChartPoint
		if err := rows.Scan(&cp.Label, &cp.Income, &cp.Expense); err != nil {
			return nil, fmt.Errorf("GetTransactionReport scan: %w", err)
		}
		result = append(result, cp)
	}
	return result, nil
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
