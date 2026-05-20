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

func (p *DashboardRepo) UserProfile(ctx context.Context, email string) (model.Profile, error) {
	sql := `WITH TargetUser AS (
	SELECT id FROM users WHERE email = $1
),
UserWallets AS (
    SELECT w.id, w.balance
    FROM wallets w
    INNER JOIN TargetUser tu ON w.user_id = tu.id
)
SELECT 
    (SELECT COALESCE(SUM(balance), 0) FROM UserWallets) AS current_balance,
    
    COALESCE(SUM(CASE 
        WHEN t.direction = 'IN' AND t.status = 'SUCCESS' THEN t.amount 
        ELSE 0 
    END), 0) AS total_income,
    
    COALESCE(SUM(CASE 
        WHEN t.direction = 'OUT' AND t.status = 'SUCCESS' THEN (t.amount + t.admin_fee) 
        ELSE 0 
    END), 0) AS total_expense

FROM transactions t
WHERE t.wallet_id IN (SELECT id FROM UserWallets);`

	var profile model.Profile

	err := p.db.QueryRow(ctx, sql, email).Scan(
		&profile.FullName, &profile.Phone,
		&profile.Photo, &profile.CreatedAt, &profile.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Profile{}, errors.New("user profile not found")
		}
		return model.Profile{}, err
	}

	return profile, nil
}
