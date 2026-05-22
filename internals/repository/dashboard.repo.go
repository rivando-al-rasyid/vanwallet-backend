package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type DashboardRepo struct {
	db *pgxpool.Pool
}

func NewDashboardRepo(db *pgxpool.Pool) *DashboardRepo {
	return &DashboardRepo{db: db}
}

// GetData returns aggregated balance, income, and expense for the user.
func (p *DashboardRepo) GetData(ctx context.Context, email string) (model.Dashboard, error) {
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

	var d model.Dashboard
	err := p.db.QueryRow(ctx, sql, email).Scan(
		&d.CurrentBalance,
		&d.TotalIncome,
		&d.TotalExpense,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Dashboard{}, errors.New("user profile not found")
		}
		return model.Dashboard{}, err
	}
	return d, nil
}

// GetTransactionReport returns daily buckets for the last 7 days,
// or weekly buckets for the last 30 days, for all wallets owned by the user.
// range_ must be "7days" or "30days".
func (p *DashboardRepo) GetTransactionReport(ctx context.Context, email string, rangeParam string) ([]model.ChartPoint, error) {
	var sql string

	switch rangeParam {
	case "30days":
		// Group by ISO week number within the last 30 days
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

	rows, err := p.db.Query(ctx, sql, email)
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
