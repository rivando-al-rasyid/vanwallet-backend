package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

// Misalkan Anda mendefinisikan ini di package model
// var ErrUserNotFound = errors.New("user not found")

type ProfileRepo struct {
	db *pgxpool.Pool
}

func NewProfileRepo(db *pgxpool.Pool) *ProfileRepo {
	return &ProfileRepo{db: db}
}

func (p *ProfileRepo) UserProfile(ctx context.Context, email string) (model.Profile, error) {
	sql := `SELECT
    p.full_name,
    p.phone,
    p.photo,
    p.created_at,
    p.updated_at
FROM
    profiles p
JOIN
    users u ON p.user_id = u.id
WHERE
    u.email = $1;`

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

func (p *ProfileRepo) EditProfile(ctx context.Context, email string, updates map[string]any) (model.Profile, error) {
	if len(updates) == 0 {
		return model.Profile{}, fmt.Errorf("EditProfile: no fields to update")
	}

	allowed := map[string]bool{
		"full_name": true,
		"phone":     true,
		"photo":     true,
	}

	var (
		sb      strings.Builder
		args    []any
		counter int
	)

	args = append(args, email)
	counter = 1

	sb.WriteString(`UPDATE profiles SET `)

	first := true

	for col, val := range updates {
		if !allowed[col] {
			return model.Profile{}, fmt.Errorf("EditProfile: column '%s' is not updatable", col)
		}
		if !first {
			sb.WriteString(", ")
		}
		counter++
		fmt.Fprintf(&sb, "%s = $%d", col, counter)
		args = append(args, val)
		first = false
	}

	sb.WriteString(`, updated_at = now()`)
	sb.WriteString(`
        FROM users u
        WHERE profiles.user_id = u.id
          AND u.email = $1
        RETURNING
            profiles.user_id,
            profiles.full_name,
            profiles.phone,
            profiles.photo,
            profiles.updated_at`)

	var profile model.Profile
	err := p.db.QueryRow(ctx, sb.String(), args...).Scan(
		&profile.UserID,
		&profile.FullName,
		&profile.Phone,
		&profile.Photo,
		&profile.UpdatedAt,
	)
	if err != nil {
		return model.Profile{}, fmt.Errorf("EditProfile: %w", err)
	}

	return profile, nil
}
