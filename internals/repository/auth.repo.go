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
	return &Authrepo{
		db: db,
	}
}

func (a *Authrepo) Register(ctx context.Context, email, hashpwd string) (model.User, error) {
	sql := "INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id, email, created_at"
	args := []any{email, hashpwd}
	var user model.User

	if err := a.db.QueryRow(ctx, sql, args...).Scan(&user.Id, &user.Email, &user.Password); err != nil {
		return model.User{}, err
	}
	return user, nil

}

func (a *Authrepo) Login(ctx context.Context, email, hashpwd string) (model.User, error) {
	sql := "SELECT id, password FROM users WHERE email = $1"
	args := []any{email}
	var user model.User

	if err := a.db.QueryRow(ctx, sql, args...).Scan(&user.Id, &user.Password); err != nil {
		return model.User{}, err
	}
	return user, nil

}

// func (a *Authrepo) login(ctx Context.Context) {
// 	sql := ""
// }
