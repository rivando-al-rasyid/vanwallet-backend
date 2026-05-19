package repository

import (
	"context"
	"errors"

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
        p.user_id, p.full_name, p.phone, p.photo, p.created_at, p.updated_at
	FROM users u 
	WHERE profiles.user_id = u.id 
  	AND u.email = $1;`

	var user model.User
	var profile model.Profile

	err := p.db.QueryRow(ctx, sql, email).Scan(
		&user.ID, &user.Email,
		&profile.ID, &profile.UserID, &profile.FullName, &profile.Phone,
		&profile.Photo, &profile.CreatedAt, &profile.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Mapping error database ke error domain aplikasi
			return model.Profile{}, errors.New("user profile not found")
		}
		return model.Profile{}, err
	}

	return profile, nil
}

func (p *ProfileRepo) EditProfile(ctx context.Context, email string) (model.Profile, error) {

	sql := `UPDATE profiles
	SET 
    full_name = $2, 
    phone = $3, 
    photo = $4, 
    updated_at = now() 
	FROM users u 
	WHERE profiles.user_id = u.id 
  	AND u.email = $1;
`
	var user model.User
	var profile model.Profile

	err := p.db.QueryRow(ctx, sql, email).Scan(
		&user.ID, &user.Email, &profile.UserID, &profile.FullName, &profile.Phone,
		&profile.Photo, &profile.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Mapping error database ke error domain aplikasi
			return model.Profile{}, errors.New("user profile not found") // Gunakan custom error dari model Anda
		}
		return model.Profile{}, err
	}

	return profile, nil
}
