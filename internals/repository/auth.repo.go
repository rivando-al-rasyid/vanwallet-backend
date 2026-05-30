package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

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
func (a *Authrepo) Register(ctx context.Context, email, hashpwd string) (model.User, error) {
	tx, err := a.db.Begin(ctx)
	if err != nil {
		return model.User{}, fmt.Errorf("Register begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var user model.User
	err = tx.QueryRow(ctx,
		`INSERT INTO users (email, password) VALUES ($1, $2)
		 RETURNING id, email, created_at`,
		email, hashpwd,
	).Scan(&user.ID, &user.Email, &user.CreatedAt)
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

// SaveToken inserts a new token row into the tokens table.
// tokenType should be model.TokenTypeRefresh for login-issued access tokens.
func (a *Authrepo) SaveToken(ctx context.Context, userID, rawToken string, tokenType model.TokenType, expiresAt time.Time) error {
	_, err := a.db.Exec(ctx,
		`INSERT INTO tokens (user_id, token, type, expires_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (token) DO NOTHING`,
		userID, rawToken, tokenType, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("SaveToken: %w", err)
	}
	return nil
}

// RevokeToken marks a specific token as revoked (logout current device).
func (a *Authrepo) RevokeToken(ctx context.Context, rawToken string) error {
	_, err := a.db.Exec(ctx,
		`UPDATE tokens SET is_revoked = true WHERE token = $1`, rawToken,
	)
	if err != nil {
		return fmt.Errorf("RevokeToken: %w", err)
	}
	return nil
}

// IsTokenValid returns true if the token exists, is not revoked, and has not expired.
func (a *Authrepo) IsTokenValid(ctx context.Context, rawToken string) (bool, error) {
	var valid bool
	err := a.db.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM tokens
			WHERE token = $1
			  AND is_revoked = false
			  AND expires_at > now()
		)`, rawToken,
	).Scan(&valid)
	if err != nil {
		return false, fmt.Errorf("IsTokenValid: %w", err)
	}
	return valid, nil
}

func (a *Authrepo) GetTokenByEmail(ctx context.Context, email string) (string, error) {
	var token string

	query := `
		SELECT t.token 
		FROM tokens t
		JOIN users u ON t.user_id = u.id
		WHERE u.email = $1
		  AND type = EMAIL_VERIFICATION
		  AND t.is_revoked = false
		  AND t.expires_at > now()
		ORDER BY created_at DESC 
		LIMIT 1;
	`

	err := a.db.QueryRow(ctx, query, email).Scan(&token)
	if err != nil {
		// It's good practice to handle the "no rows found" case specifically
		if errors.Is(err, pgx.ErrNoRows) { // or sql.ErrNoRows depending on your driver
			return "", nil // Or return a custom error like ErrTokenNotFound
		}
		return "", fmt.Errorf("GetTokenByEmail: %w", err)
	}

	return token, nil
}
