package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

// GetSummary returns aggregated balance, income, expense, and per-wallet breakdown.
func (t *TransactionRepo) GetSummary(ctx context.Context, email string) (model.TransactionSummary, error) {
	aggSQL := `
	WITH TargetUser AS (
		SELECT id FROM users WHERE email = $1
	),
	UserWallets AS (
		SELECT id, balance FROM wallets WHERE user_id IN (SELECT id FROM TargetUser)
	),
	WalletAgg AS (
		SELECT COALESCE(SUM(balance), 0) AS current_balance FROM UserWallets
	),
	TransactionAgg AS (
		SELECT
			COALESCE(SUM(CASE WHEN type = 'TRANSFER_IN'  AND status = 'SUCCESS' THEN amount ELSE 0 END), 0) AS total_income,
			COALESCE(SUM(CASE WHEN type IN ('EXPENSE','WITHDRAWAL','TRANSFER_OUT') AND status = 'SUCCESS' THEN (amount + admin_fee) ELSE 0 END), 0) AS total_expense
		FROM transactions
		WHERE wallet_id IN (SELECT id FROM UserWallets)
	)
	SELECT wa.current_balance, ta.total_income, ta.total_expense
	FROM WalletAgg wa CROSS JOIN TransactionAgg ta`

	var s model.TransactionSummary
	if err := t.db.QueryRow(ctx, aggSQL, email).Scan(
		&s.CurrentBalance, &s.TotalIncome, &s.TotalExpense,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.TransactionSummary{}, errors.New("user not found")
		}
		return model.TransactionSummary{}, fmt.Errorf("GetSummary agg: %w", err)
	}

	rows, err := t.db.Query(ctx, `
		SELECT w.id, w.label, w.balance
		FROM wallets w
		JOIN users u ON w.user_id = u.id
		WHERE u.email = $1
		ORDER BY w.created_at ASC`, email)
	if err != nil {
		return model.TransactionSummary{}, fmt.Errorf("GetSummary wallets: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var w model.WalletSummary
		if err := rows.Scan(&w.ID, &w.Label, &w.Balance); err != nil {
			return model.TransactionSummary{}, fmt.Errorf("GetSummary wallet scan: %w", err)
		}
		s.Wallets = append(s.Wallets, w)
	}
	return s, rows.Err()
}

// GetTransactionReport returns daily (7days) or weekly (30days) chart buckets.
func (t *TransactionRepo) GetTransactionReport(ctx context.Context, email, rangeParam, typeFilter string) ([]model.ChartPoint, error) {
	var dateTrunc, dateLabel, interval string
	switch rangeParam {
	case "30days":
		dateTrunc = "week"
		dateLabel = "'W' || TO_CHAR(DATE_TRUNC('week', created_at), 'IW')"
		interval = "30 days"
	default:
		dateTrunc = "day"
		dateLabel = "TO_CHAR(DATE_TRUNC('day', created_at), 'Dy')"
		interval = "7 days"
	}
	var incomeExpr, expenseExpr string
	switch typeFilter {
	case "income":
		incomeExpr = "COALESCE(SUM(CASE WHEN type = 'TRANSFER_IN' AND status = 'SUCCESS' THEN amount ELSE 0 END), 0)"
		expenseExpr = "0"
	case "expense":
		incomeExpr = "0"
		expenseExpr = "COALESCE(SUM(CASE WHEN type IN ('EXPENSE','WITHDRAWAL','TRANSFER_OUT') AND status = 'SUCCESS' THEN (amount + admin_fee) ELSE 0 END), 0)"
	default:
		incomeExpr = "COALESCE(SUM(CASE WHEN type = 'TRANSFER_IN' AND status = 'SUCCESS' THEN amount ELSE 0 END), 0)"
		expenseExpr = "COALESCE(SUM(CASE WHEN type IN ('EXPENSE','WITHDRAWAL','TRANSFER_OUT') AND status = 'SUCCESS' THEN (amount + admin_fee) ELSE 0 END), 0)"
	}
	sql := fmt.Sprintf(`
		WITH UserWallets AS (
			SELECT w.id FROM wallets w JOIN users u ON w.user_id = u.id WHERE u.email = $1
		)
		SELECT %s AS label, %s AS income, %s AS expense
		FROM transactions
		WHERE wallet_id IN (SELECT id FROM UserWallets)
		  AND created_at >= NOW() - INTERVAL '%s'
		GROUP BY DATE_TRUNC('%s', created_at)
		ORDER BY DATE_TRUNC('%s', created_at)`,
		dateLabel, incomeExpr, expenseExpr, interval, dateTrunc, dateTrunc,
	)
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
	return result, rows.Err()
}

func (t *TransactionRepo) GetAllHistory(ctx context.Context, email string, filter model.HistoryFilter) ([]model.HistoryItem, int, error) {
	offset := (filter.Page - 1) * filter.Limit

	baseSQL := `
		WITH UserWallets AS (
			SELECT w.id, w.label
			FROM wallets w
			JOIN users u ON w.user_id = u.id
			WHERE u.email = $1
		), UnifiedHistory AS (
			SELECT
				tr.id::text                         AS id,
				'transaction'                       AS source,
				tr.type::text                       AS type,
				CASE
					WHEN tr.type = 'TRANSFER_IN' THEN 'income'
					ELSE 'expense'
				END                                AS direction,
				tr.amount                           AS amount,
				tr.admin_fee                        AS admin_fee,
				tr.status::text                     AS status,
				''                                  AS payment_method,
				COALESCE(tr.note, '')               AS note,
				tr.wallet_id::text                  AS wallet_id,
				uw.label                            AS wallet_label,
				tr.created_at                       AS created_at
			FROM transactions tr
			JOIN UserWallets uw ON tr.wallet_id = uw.id

			UNION ALL

			SELECT
				tp.id::text                         AS id,
				'topup'                             AS source,
				'TOPUP'                             AS type,
				'income'                            AS direction,
				tp.amount                           AS amount,
				0                                   AS admin_fee,
				tp.status::text                     AS status,
				COALESCE(tp.payment_method::text, '') AS payment_method,
				''                                  AS note,
				tp.wallet_id::text                  AS wallet_id,
				uw.label                            AS wallet_label,
				tp.created_at                       AS created_at
			FROM topups tp
			JOIN UserWallets uw ON tp.wallet_id = uw.id
		)
		SELECT * FROM UnifiedHistory`

	args := []any{email}
	conditions := make([]string, 0)
	addArg := func(value any) string {
		args = append(args, value)
		return fmt.Sprintf("$%d", len(args))
	}

	if filter.WalletID != "" {
		conditions = append(conditions, "wallet_id = "+addArg(filter.WalletID))
	}
	if filter.Source != "" {
		conditions = append(conditions, "LOWER(source) = LOWER("+addArg(filter.Source)+")")
	}
	if filter.Type != "" {
		conditions = append(conditions, "UPPER(type) = UPPER("+addArg(filter.Type)+")")
	}
	if filter.Status != "" {
		conditions = append(conditions, "UPPER(status) = UPPER("+addArg(filter.Status)+")")
	}
	if filter.Direction != "" {
		conditions = append(conditions, "LOWER(direction) = LOWER("+addArg(filter.Direction)+")")
	}
	if filter.StartDate != "" {
		conditions = append(conditions, "created_at >= "+addArg(filter.StartDate)+"::date")
	}
	if filter.EndDate != "" {
		conditions = append(conditions, "created_at < ("+addArg(filter.EndDate)+"::date + INTERVAL '1 day')")
	}
	if filter.Query != "" {
		q := "%" + filter.Query + "%"
		placeholder := addArg(q)
		conditions = append(conditions, `(
			id ILIKE `+placeholder+` OR
			source ILIKE `+placeholder+` OR
			type ILIKE `+placeholder+` OR
			direction ILIKE `+placeholder+` OR
			status ILIKE `+placeholder+` OR
			payment_method ILIKE `+placeholder+` OR
			note ILIKE `+placeholder+` OR
			wallet_id ILIKE `+placeholder+` OR
			wallet_label ILIKE `+placeholder+`
		)`)
	}

	whereSQL := ""
	if len(conditions) > 0 {
		whereSQL = " WHERE " + strings.Join(conditions, " AND ")
	}

	countSQL := `SELECT COUNT(*) FROM (` + baseSQL + whereSQL + `) AS filtered_history`
	var total int
	if err := t.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("GetAllHistory count: %w", err)
	}

	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, filter.Limit, offset)
	querySQL := baseSQL + whereSQL + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(queryArgs)-1, len(queryArgs))

	rows, err := t.db.Query(ctx, querySQL, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("GetAllHistory query: %w", err)
	}
	defer rows.Close()

	var items []model.HistoryItem
	for rows.Next() {
		var h model.HistoryItem
		if err := rows.Scan(
			&h.ID, &h.Source, &h.Type, &h.Direction, &h.Amount, &h.AdminFee,
			&h.Status, &h.PaymentMethod, &h.Note, &h.WalletID, &h.WalletLabel, &h.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("GetAllHistory scan: %w", err)
		}
		items = append(items, h)
	}
	return items, total, rows.Err()
}
