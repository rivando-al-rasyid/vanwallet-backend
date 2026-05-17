package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type ProfileRepo struct {
	db *pgxpool.Pool
}

func NewProfileRepo(db *pgxpool.Pool) *ProfileRepo {
	return &ProfileRepo{db: db}
}

func (a *ProfileRepo) UserProfile(ctx context.Context, email, hashpwd string) (model.User, error) {
	sql := `	sql := "SELECT id, password FROM users WHERE email = $1"`
	var user model.User

	err := a.db.QueryRow(ctx, sql, email, hashpwd).Scan(
		&user.Id,
		&user.Email,
		&user.CreatedAt,
	)
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}
