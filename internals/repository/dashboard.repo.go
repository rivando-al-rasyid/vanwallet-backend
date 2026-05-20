package repository

import (
	"context"
	"errors"

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

func (p *DashboardRepo) GetData(ctx context.Context, email string) (model.Dashboard, error) {
	// We isolate the aggregations into CTEs to avoid duplicating wallet balances
	// if a user has multiple transactions, then tie it all back to the TargetUser.
	sql := `WITH TargetUser AS (
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
            COALESCE(SUM(CASE WHEN direction = 'IN' AND status = 'SUCCESS' THEN amount ELSE 0 END), 0) AS total_income,
            COALESCE(SUM(CASE WHEN direction = 'OUT' AND status = 'SUCCESS' THEN (amount + admin_fee) ELSE 0 END), 0) AS total_expense
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

	var dashboard model.Dashboard

	err := p.db.QueryRow(ctx, sql, email).Scan(
		&dashboard.CurrentBalance,
		&dashboard.TotalIncome,
		&dashboard.TotalExpense,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Dashboard{}, errors.New("user profile not found")
		}
		return model.Dashboard{}, err
	}

	return dashboard, nil
}
