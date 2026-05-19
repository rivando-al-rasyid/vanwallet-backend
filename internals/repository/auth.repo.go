package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type Authrepo struct {
	db *pgxpool.Pool
}

func NewAuthRepo(db *pgxpool.Pool) *Authrepo {
	return &Authrepo{db: db}
}

func (a *Authrepo) Register(ctx context.Context, email, hashpwd string) (model.User, error) {
	sql := `
		WITH new_user AS (
			INSERT INTO users (email, password)
			VALUES ($1, $2)
			RETURNING id, email, created_at
		),
		new_profile AS (
			INSERT INTO profiles (user_id)
			SELECT id FROM new_user
		),
		new_pin AS (
			INSERT INTO user_pins (user_id)
			SELECT id FROM new_user
		),
		new_wallet AS (
			INSERT INTO wallets (user_id)
			SELECT id FROM new_user
		)
		SELECT id, email, created_at FROM new_user
	`

	var user model.User

	err := a.db.QueryRow(ctx, sql, email, hashpwd).Scan(
		&user.ID,
		&user.Email,
		&user.CreatedAt,
	)
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (a *Authrepo) Login(ctx context.Context, email string) (model.User, error) {
	sql := "SELECT id, password FROM users WHERE email = $1"
	args := []any{email}
	var user model.User
	if err := a.db.QueryRow(ctx, sql, args...).Scan(&user.ID, &user.Password); err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (a *Authrepo) ClearToken(ctx context.Context, userID string) error {
	sql := "UPDATE users SET token = NULL, updated_at = now() WHERE id = $1"
	_, err := a.db.Exec(ctx, sql, userID)
	return err
}
