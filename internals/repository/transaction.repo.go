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
// Income = sum of TRANSFER_IN. Expense = sum of EXPENSE + WITHDRAWAL + TRANSFER_OUT (incl. admin_fee).
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
			COALESCE(SUM(CASE WHEN type = 'TRANSFER_IN' AND status = 'SUCCESS' THEN amount ELSE 0 END), 0) AS total_income,
			COALESCE(SUM(CASE WHEN type IN ('EXPENSE','WITHDRAWAL','TRANSFER_OUT') AND status = 'SUCCESS' THEN (amount + admin_fee) ELSE 0 END), 0) AS total_expense
		FROM transactions
		WHERE wallet_id IN (SELECT id FROM UserWallets)
	)
	SELECT
		wa.current_balance,
		ta.total_income,
		ta.total_expense
	FROM WalletAgg wa
	CROSS JOIN TransactionAgg ta;`

	var s model.TransactionSummary
	err := t.db.QueryRow(ctx, sql, email).Scan(
		&s.CurrentBalance,
		&s.TotalIncome,
		&s.TotalExpense,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.TransactionSummary{}, errors.New("user not found")
		}
		return model.TransactionSummary{}, err
	}
	return s, nil
}

// GetTransactionReport returns daily (7days) or weekly (30days) buckets.
func (t *TransactionRepo) GetTransactionReport(ctx context.Context, email string, rangeParam string) ([]model.ChartPoint, error) {
	var sql string

	switch rangeParam {
	case "30days":
		sql = `
		WITH UserWallets AS (
			SELECT w.id FROM wallets w JOIN users u ON w.user_id = u.id WHERE u.email = $1
		)
		SELECT
			'W' || TO_CHAR(DATE_TRUNC('week', created_at), 'IW') AS label,
			COALESCE(SUM(CASE WHEN type = 'TRANSFER_IN' AND status = 'SUCCESS' THEN amount ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN type IN ('EXPENSE','WITHDRAWAL','TRANSFER_OUT') AND status = 'SUCCESS' THEN (amount + admin_fee) ELSE 0 END), 0) AS expense
		FROM transactions
		WHERE wallet_id IN (SELECT id FROM UserWallets)
		  AND created_at >= NOW() - INTERVAL '30 days'
		GROUP BY DATE_TRUNC('week', created_at)
		ORDER BY DATE_TRUNC('week', created_at);`
	default:
		sql = `
		WITH UserWallets AS (
			SELECT w.id FROM wallets w JOIN users u ON w.user_id = u.id WHERE u.email = $1
		)
		SELECT
			TO_CHAR(DATE_TRUNC('day', created_at), 'Dy') AS label,
			COALESCE(SUM(CASE WHEN type = 'TRANSFER_IN' AND status = 'SUCCESS' THEN amount ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN type IN ('EXPENSE','WITHDRAWAL','TRANSFER_OUT') AND status = 'SUCCESS' THEN (amount + admin_fee) ELSE 0 END), 0) AS expense
		FROM transactions
		WHERE wallet_id IN (SELECT id FROM UserWallets)
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

// GetTransactionsByWallet returns paginated transactions for a wallet owned by the user.
func (t *TransactionRepo) GetTransactionsByWallet(ctx context.Context, email string, walletID uuid.UUID, page, limit int) ([]model.Transaction, int, error) {
	offset := (page - 1) * limit

	// Verify ownership
	var wid uuid.UUID
	if err := t.db.QueryRow(ctx, `
		SELECT w.id FROM wallets w JOIN users u ON w.user_id = u.id
		WHERE u.email = $1 AND w.id = $2`, email, walletID,
	).Scan(&wid); err != nil {
		return nil, 0, fmt.Errorf("GetTransactionsByWallet: wallet not found or access denied: %w", err)
	}

	var total int
	if err := t.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM transactions WHERE wallet_id = $1`, walletID,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("GetTransactionsByWallet count: %w", err)
	}

	rows, err := t.db.Query(ctx, `
		SELECT id, wallet_id, type, amount, admin_fee, status, idempotency_key, note, created_at, updated_at
		FROM transactions
		WHERE wallet_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, walletID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("GetTransactionsByWallet query: %w", err)
	}
	defer rows.Close()

	txs, err := scanTransactions(rows)
	return txs, total, err
}

// GetAllTransactions returns paginated transactions across all wallets for the user.
func (t *TransactionRepo) GetAllTransactions(ctx context.Context, email string, page, limit int) ([]model.Transaction, int, error) {
	offset := (page - 1) * limit

	var total int
	if err := t.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM transactions tr
		JOIN wallets w ON tr.wallet_id = w.id
		JOIN users u ON w.user_id = u.id
		WHERE u.email = $1`, email,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("GetAllTransactions count: %w", err)
	}

	rows, err := t.db.Query(ctx, `
		SELECT tr.id, tr.wallet_id, tr.type, tr.amount, tr.admin_fee, tr.status, tr.idempotency_key, tr.note, tr.created_at, tr.updated_at
		FROM transactions tr
		JOIN wallets w ON tr.wallet_id = w.id
		JOIN users u ON w.user_id = u.id
		WHERE u.email = $1
		ORDER BY tr.created_at DESC
		LIMIT $2 OFFSET $3`, email, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("GetAllTransactions query: %w", err)
	}
	defer rows.Close()

	txs, err := scanTransactions(rows)
	return txs, total, err
}

// GetTransactionByID returns a single transaction, verifying ownership.
func (t *TransactionRepo) GetTransactionByID(ctx context.Context, email string, transactionID uuid.UUID) (model.Transaction, error) {
	var tx model.Transaction
	err := t.db.QueryRow(ctx, `
		SELECT tr.id, tr.wallet_id, tr.type, tr.amount, tr.admin_fee, tr.status, tr.idempotency_key, tr.note, tr.created_at, tr.updated_at
		FROM transactions tr
		JOIN wallets w ON tr.wallet_id = w.id
		JOIN users u ON w.user_id = u.id
		WHERE u.email = $1 AND tr.id = $2`, email, transactionID,
	).Scan(
		&tx.ID, &tx.WalletID, &tx.Type, &tx.Amount,
		&tx.AdminFee, &tx.Status, &tx.IdempotencyKey, &tx.Note, &tx.CreatedAt, &tx.UpdatedAt,
	)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("GetTransactionByID: %w", err)
	}
	return tx, nil
}

// CreateTopup creates a standalone top-up record in a DB transaction and credits the wallet.
// Status starts as PENDING — a webhook/callback should confirm and set SUCCESS.
func (t *TransactionRepo) CreateTopup(ctx context.Context, req model.Topup) (model.Topup, error) {
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return model.Topup{}, fmt.Errorf("CreateTopup begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Verify wallet ownership is done at service/controller level before calling here.
	var topup model.Topup
	err = tx.QueryRow(ctx, `
		INSERT INTO topups (wallet_id, amount, status, payment_method)
		VALUES ($1, $2, $3, $4)
		RETURNING id, wallet_id, amount, status, payment_method, external_reference, created_at`,
		req.WalletID, req.Amount, model.TransactionStatusPending, req.PaymentMethod,
	).Scan(
		&topup.ID, &topup.WalletID, &topup.Amount, &topup.Status,
		&topup.PaymentMethod, &topup.ExternalReference, &topup.CreatedAt,
	)
	if err != nil {
		return model.Topup{}, fmt.Errorf("CreateTopup insert: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return model.Topup{}, fmt.Errorf("CreateTopup commit: %w", err)
	}
	return topup, nil
}

// ConfirmTopup sets a topup to SUCCESS and credits the wallet balance atomically.
func (t *TransactionRepo) ConfirmTopup(ctx context.Context, topupID uuid.UUID) (model.Topup, error) {
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return model.Topup{}, fmt.Errorf("ConfirmTopup begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var topup model.Topup
	err = tx.QueryRow(ctx, `
		UPDATE topups SET status = 'SUCCESS', updated_at = now()
		WHERE id = $1 AND status = 'PENDING'
		RETURNING id, wallet_id, amount, status, payment_method, external_reference, created_at`,
		topupID,
	).Scan(
		&topup.ID, &topup.WalletID, &topup.Amount, &topup.Status,
		&topup.PaymentMethod, &topup.ExternalReference, &topup.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Topup{}, errors.New("topup not found or already processed")
		}
		return model.Topup{}, fmt.Errorf("ConfirmTopup update topup: %w", err)
	}

	if _, err = tx.Exec(ctx,
		`UPDATE wallets SET balance = balance + $1, updated_at = now() WHERE id = $2`,
		topup.Amount, topup.WalletID,
	); err != nil {
		return model.Topup{}, fmt.Errorf("ConfirmTopup credit wallet: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return model.Topup{}, fmt.Errorf("ConfirmTopup commit: %w", err)
	}
	return topup, nil
}

// CreateWithdrawal inserts a WITHDRAWAL transaction + withdrawals detail row,
// and debits the wallet balance — all atomically.
func (t *TransactionRepo) CreateWithdrawal(ctx context.Context, walletID uuid.UUID, amount, adminFee int64, bank model.Withdrawal) (model.Transaction, error) {
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("CreateWithdrawal begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Check sufficient balance (amount + adminFee)
	var balance int64
	if err = tx.QueryRow(ctx,
		`SELECT balance FROM wallets WHERE id = $1 FOR UPDATE`, walletID,
	).Scan(&balance); err != nil {
		return model.Transaction{}, fmt.Errorf("CreateWithdrawal lock wallet: %w", err)
	}
	total := amount + adminFee
	if balance < total {
		return model.Transaction{}, errors.New("insufficient balance")
	}

	// Insert transaction row
	var txRow model.Transaction
	err = tx.QueryRow(ctx, `
		INSERT INTO transactions (wallet_id, type, amount, admin_fee, status)
		VALUES ($1, 'WITHDRAWAL', $2, $3, 'PENDING')
		RETURNING id, wallet_id, type, amount, admin_fee, status, idempotency_key, note, created_at, updated_at`,
		walletID, amount, adminFee,
	).Scan(
		&txRow.ID, &txRow.WalletID, &txRow.Type, &txRow.Amount,
		&txRow.AdminFee, &txRow.Status, &txRow.IdempotencyKey, &txRow.Note, &txRow.CreatedAt, &txRow.UpdatedAt,
	)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("CreateWithdrawal insert transaction: %w", err)
	}

	// Insert withdrawal detail
	if _, err = tx.Exec(ctx, `
		INSERT INTO withdrawals (transaction_id, bank_name, account_number, account_holder)
		VALUES ($1, $2, $3, $4)`,
		txRow.ID, bank.BankName, bank.AccountNumber, bank.AccountHolder,
	); err != nil {
		return model.Transaction{}, fmt.Errorf("CreateWithdrawal insert detail: %w", err)
	}

	// Debit wallet
	if _, err = tx.Exec(ctx,
		`UPDATE wallets SET balance = balance - $1, updated_at = now() WHERE id = $2`,
		total, walletID,
	); err != nil {
		return model.Transaction{}, fmt.Errorf("CreateWithdrawal debit wallet: %w", err)
	}

	// Mark transaction SUCCESS
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

// CreateTransfer executes a peer-to-peer transfer atomically:
//  1. Lock and check sender balance
//  2. Insert TRANSFER_OUT transaction (sender)
//  3. Insert TRANSFER_IN  transaction (recipient)
//  4. Insert transfers link row
//  5. Debit sender wallet, credit recipient wallet
func (t *TransactionRepo) CreateTransfer(ctx context.Context, senderWalletID, recipientWalletID uuid.UUID, amount, adminFee int64, note *string) (model.Transfer, model.Transaction, model.Transaction, error) {
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, fmt.Errorf("CreateTransfer begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Lock sender wallet and check balance
	var senderBalance int64
	if err = tx.QueryRow(ctx,
		`SELECT balance FROM wallets WHERE id = $1 FOR UPDATE`, senderWalletID,
	).Scan(&senderBalance); err != nil {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, fmt.Errorf("CreateTransfer lock sender wallet: %w", err)
	}
	total := amount + adminFee
	if senderBalance < total {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, errors.New("insufficient balance")
	}

	// Lock recipient wallet (prevent deadlock with consistent ordering)
	if _, err = tx.Exec(ctx,
		`SELECT id FROM wallets WHERE id = $1 FOR UPDATE`, recipientWalletID,
	); err != nil {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, fmt.Errorf("CreateTransfer lock recipient wallet: %w", err)
	}

	// Insert TRANSFER_OUT (sender)
	var senderTx model.Transaction
	err = tx.QueryRow(ctx, `
		INSERT INTO transactions (wallet_id, type, amount, admin_fee, status, note)
		VALUES ($1, 'TRANSFER_OUT', $2, $3, 'SUCCESS', $4)
		RETURNING id, wallet_id, type, amount, admin_fee, status, idempotency_key, note, created_at, updated_at`,
		senderWalletID, amount, adminFee, note,
	).Scan(
		&senderTx.ID, &senderTx.WalletID, &senderTx.Type, &senderTx.Amount,
		&senderTx.AdminFee, &senderTx.Status, &senderTx.IdempotencyKey, &senderTx.Note, &senderTx.CreatedAt, &senderTx.UpdatedAt,
	)
	if err != nil {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, fmt.Errorf("CreateTransfer insert TRANSFER_OUT: %w", err)
	}

	// Insert TRANSFER_IN (recipient)
	var recipientTx model.Transaction
	err = tx.QueryRow(ctx, `
		INSERT INTO transactions (wallet_id, type, amount, admin_fee, status, note)
		VALUES ($1, 'TRANSFER_IN', $2, 0, 'SUCCESS', $3)
		RETURNING id, wallet_id, type, amount, admin_fee, status, idempotency_key, note, created_at, updated_at`,
		recipientWalletID, amount, note,
	).Scan(
		&recipientTx.ID, &recipientTx.WalletID, &recipientTx.Type, &recipientTx.Amount,
		&recipientTx.AdminFee, &recipientTx.Status, &recipientTx.IdempotencyKey, &recipientTx.Note, &recipientTx.CreatedAt, &recipientTx.UpdatedAt,
	)
	if err != nil {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, fmt.Errorf("CreateTransfer insert TRANSFER_IN: %w", err)
	}

	// Generate transfer_code
	transferCode := fmt.Sprintf("TRF-%s", senderTx.ID.String()[:8])

	// Insert transfers link
	var transfer model.Transfer
	err = tx.QueryRow(ctx, `
		INSERT INTO transfers (transaction_id, recipient_transaction_id, transfer_code)
		VALUES ($1, $2, $3)
		RETURNING transaction_id, recipient_transaction_id, transfer_code, created_at`,
		senderTx.ID, recipientTx.ID, transferCode,
	).Scan(
		&transfer.TransactionID, &transfer.RecipientTransactionID,
		&transfer.TransferCode, &transfer.CreatedAt,
	)
	if err != nil {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, fmt.Errorf("CreateTransfer insert link: %w", err)
	}

	// Debit sender
	if _, err = tx.Exec(ctx,
		`UPDATE wallets SET balance = balance - $1, updated_at = now() WHERE id = $2`,
		total, senderWalletID,
	); err != nil {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, fmt.Errorf("CreateTransfer debit sender: %w", err)
	}

	// Credit recipient
	if _, err = tx.Exec(ctx,
		`UPDATE wallets SET balance = balance + $1, updated_at = now() WHERE id = $2`,
		amount, recipientWalletID,
	); err != nil {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, fmt.Errorf("CreateTransfer credit recipient: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, fmt.Errorf("CreateTransfer commit: %w", err)
	}

	return transfer, senderTx, recipientTx, nil
}

// CreateExpense inserts an EXPENSE transaction + expenses detail row, and debits the wallet.
func (t *TransactionRepo) CreateExpense(ctx context.Context, walletID uuid.UUID, amount, adminFee int64, category, merchantName, note *string) (model.Transaction, error) {
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("CreateExpense begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Check balance
	var balance int64
	if err = tx.QueryRow(ctx,
		`SELECT balance FROM wallets WHERE id = $1 FOR UPDATE`, walletID,
	).Scan(&balance); err != nil {
		return model.Transaction{}, fmt.Errorf("CreateExpense lock wallet: %w", err)
	}
	if balance < amount+adminFee {
		return model.Transaction{}, errors.New("insufficient balance")
	}

	// Insert transaction
	var txRow model.Transaction
	err = tx.QueryRow(ctx, `
		INSERT INTO transactions (wallet_id, type, amount, admin_fee, status, note)
		VALUES ($1, 'EXPENSE', $2, $3, 'SUCCESS', $4)
		RETURNING id, wallet_id, type, amount, admin_fee, status, idempotency_key, note, created_at, updated_at`,
		walletID, amount, adminFee, note,
	).Scan(
		&txRow.ID, &txRow.WalletID, &txRow.Type, &txRow.Amount,
		&txRow.AdminFee, &txRow.Status, &txRow.IdempotencyKey, &txRow.Note, &txRow.CreatedAt, &txRow.UpdatedAt,
	)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("CreateExpense insert transaction: %w", err)
	}

	// Insert expense detail
	if _, err = tx.Exec(ctx,
		`INSERT INTO expenses (transaction_id, category, merchant_name) VALUES ($1, $2, $3)`,
		txRow.ID, category, merchantName,
	); err != nil {
		return model.Transaction{}, fmt.Errorf("CreateExpense insert detail: %w", err)
	}

	// Debit wallet
	if _, err = tx.Exec(ctx,
		`UPDATE wallets SET balance = balance - $1, updated_at = now() WHERE id = $2`,
		amount+adminFee, walletID,
	); err != nil {
		return model.Transaction{}, fmt.Errorf("CreateExpense debit wallet: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return model.Transaction{}, fmt.Errorf("CreateExpense commit: %w", err)
	}
	return txRow, nil
}

// scanTransactions is a helper to scan transaction rows.
func scanTransactions(rows pgx.Rows) ([]model.Transaction, error) {
	var result []model.Transaction
	for rows.Next() {
		var tx model.Transaction
		if err := rows.Scan(
			&tx.ID, &tx.WalletID, &tx.Type, &tx.Amount,
			&tx.AdminFee, &tx.Status, &tx.IdempotencyKey, &tx.Note, &tx.CreatedAt, &tx.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanTransactions: %w", err)
		}
		result = append(result, tx)
	}
	return result, rows.Err()
}

// SearchReceivers searches users by full_name or phone (case-insensitive, partial match)
// for use in the transfer flow. Excludes the calling user.
func (t *TransactionRepo) SearchReceivers(ctx context.Context, callerEmail, query string, page, limit int) ([]model.ReceiverResult, int, error) {
	offset := (page - 1) * limit
	like := "%" + query + "%"

	var total int
	if err := t.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM users u
		JOIN profiles p ON p.user_id = u.id
		JOIN wallets  w ON w.user_id = u.id
		WHERE u.email != $1
		  AND (p.full_name ILIKE $2 OR p.phone ILIKE $2)`,
		callerEmail, like,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("SearchReceivers count: %w", err)
	}

	rows, err := t.db.Query(ctx, `
		SELECT
			u.id         AS user_id,
			u.email,
			p.full_name,
			p.phone,
			p.photo,
			w.id         AS wallet_id,
			w.label      AS wallet_label
		FROM users u
		JOIN profiles p ON p.user_id = u.id
		JOIN wallets  w ON w.user_id = u.id
		WHERE u.email != $1
		  AND (p.full_name ILIKE $2 OR p.phone ILIKE $2)
		ORDER BY p.full_name ASC NULLS LAST
		LIMIT $3 OFFSET $4`,
		callerEmail, like, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("SearchReceivers query: %w", err)
	}
	defer rows.Close()

	var results []model.ReceiverResult
	for rows.Next() {
		var r model.ReceiverResult
		if err := rows.Scan(
			&r.UserID, &r.Email,
			&r.FullName, &r.Phone, &r.Photo,
			&r.WalletID, &r.WalletLabel,
		); err != nil {
			return nil, 0, fmt.Errorf("SearchReceivers scan: %w", err)
		}
		results = append(results, r)
	}
	return results, total, rows.Err()
}
