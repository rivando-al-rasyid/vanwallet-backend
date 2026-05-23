package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type Authrepo struct {
	db *pgxpool.Pool
}

func NewAuthRepo(db *pgxpool.Pool) *Authrepo {
	return &Authrepo{db: db}
}

// Register creates a user, profile, user_pin, and wallet atomically in one transaction.
func (a *Authrepo) Register(ctx context.Context, email, username, hashpwd string) (model.User, error) {
	tx, err := a.db.Begin(ctx)
	if err != nil {
		return model.User{}, fmt.Errorf("Register begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var user model.User
	err = tx.QueryRow(ctx,
		`INSERT INTO users (email, username, password) VALUES ($1, $2, $3)
		 RETURNING id, email, username, created_at`,
		email, username, hashpwd,
	).Scan(&user.ID, &user.Email, &user.Username, &user.CreatedAt)
	if err != nil {
		return model.User{}, fmt.Errorf("Register insert user: %w", err)
	}

	if _, err = tx.Exec(ctx,
		`INSERT INTO profiles (user_id) VALUES ($1)`, user.ID,
	); err != nil {
		return model.User{}, fmt.Errorf("Register insert profile: %w", err)
	}

	if _, err = tx.Exec(ctx,
		`INSERT INTO user_pins (user_id, pin_hash) VALUES ($1, '')`, user.ID,
	); err != nil {
		return model.User{}, fmt.Errorf("Register insert user_pin: %w", err)
	}

	if _, err = tx.Exec(ctx,
		`INSERT INTO wallets (user_id) VALUES ($1)`, user.ID,
	); err != nil {
		return model.User{}, fmt.Errorf("Register insert wallet: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return model.User{}, fmt.Errorf("Register commit: %w", err)
	}

	return user, nil
}

func (a *Authrepo) Login(ctx context.Context, email string) (model.User, error) {
	var user model.User
	err := a.db.QueryRow(ctx,
		`SELECT id, password FROM users WHERE email = $1`, email,
	).Scan(&user.ID, &user.Password)
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (a *Authrepo) GetUserPin(ctx context.Context, email string) (model.UserPin, error) {
	var userpin model.UserPin
	err := a.db.QueryRow(ctx, `
		SELECT up.pin_hash
		FROM user_pins up
		JOIN users u ON up.user_id = u.id
		WHERE u.email = $1`, email,
	).Scan(&userpin.PinHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.UserPin{}, errors.New("user pin not found")
		}
		return model.UserPin{}, err
	}
	return userpin, nil
}

func (a *Authrepo) ClearToken(ctx context.Context, userID string) error {
	_, err := a.db.Exec(ctx,
		`UPDATE users SET token = NULL, updated_at = now() WHERE id = $1`, userID,
	)
	return err
}
