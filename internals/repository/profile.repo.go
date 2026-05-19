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

// Opsi struktur yang lebih baik, definisikan di model:
// type UserProfile struct {
//     User    model.User
//     Profile model.Profile
// }

func (p *ProfileRepo) UserProfile(ctx context.Context, email string) (model.User, model.Profile, error) {
	sql := `SELECT
        u.id, u.email,
        p.id, p.user_id, p.full_name, p.phone, p.photo, p.created_at, p.updated_at
    FROM users u
    JOIN profiles p ON u.id = p.user_id
    WHERE u.email = $1`

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
			return model.User{}, model.Profile{}, errors.New("user profile not found") // Gunakan custom error dari model Anda
		}
		return model.User{}, model.Profile{}, err
	}

	return user, profile, nil
}
