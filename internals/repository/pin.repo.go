package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type PinRepo struct {
	db *pgxpool.Pool
}

func NewPinRepo(db *pgxpool.Pool) *PinRepo {
	return &PinRepo{db: db}
}

func (p *PinRepo) GetPinByEmail(ctx context.Context, email string) (model.UserPin, error) {
	var up model.UserPin
	err := p.db.QueryRow(ctx, `
		SELECT up.id, up.user_id, up.pin_hash, up.failed_attempts, up.locked_until,
		       up.created_at, up.updated_at
		FROM user_pins up
		JOIN users u ON up.user_id = u.id
		WHERE u.email = $1`, email,
	).Scan(
		&up.ID, &up.UserID, &up.PinHash,
		&up.FailedAttempts, &up.LockedUntil,
		&up.CreatedAt, &up.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.UserPin{}, fmt.Errorf("GetPinByEmail: no pin record for %s", email)
		}
		return model.UserPin{}, fmt.Errorf("GetPinByEmail: %w", err)
	}
	return up, nil
}

func (p *PinRepo) GetPinByUserID(ctx context.Context, userID string) (model.UserPin, error) {
	var up model.UserPin
	err := p.db.QueryRow(ctx, `
		SELECT id, user_id, pin_hash, failed_attempts, locked_until, created_at, updated_at
		FROM user_pins
		WHERE user_id = $1`, userID,
	).Scan(
		&up.ID, &up.UserID, &up.PinHash,
		&up.FailedAttempts, &up.LockedUntil,
		&up.CreatedAt, &up.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.UserPin{}, fmt.Errorf("GetPinByUserID: no pin record for user %s", userID)
		}
		return model.UserPin{}, fmt.Errorf("GetPinByUserID: %w", err)
	}
	return up, nil
}

func (p *PinRepo) SetPin(ctx context.Context, email, hashedPin string) error {
	tag, err := p.db.Exec(ctx, `
		UPDATE user_pins up
		SET    pin_hash = $1,
		       updated_at = now()
		FROM   users u
		WHERE  up.user_id = u.id
		  AND  u.email    = $2`,
		hashedPin, email,
	)
	if err != nil {
		return fmt.Errorf("SetPin: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("SetPin: no pin record found for email %s", email)
	}
	return nil
}
