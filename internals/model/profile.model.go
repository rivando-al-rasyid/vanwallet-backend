package model

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID             uuid.UUID  `db:"id"`
	UserID         uuid.UUID  `db:"user_id"`
	FullName       *string    `db:"full_name"`
	Phone          *string    `db:"phone"`
	Photo          *string    `db:"photo"`
	CurrentBalance int64      `db:"current_balance"`
	PinHash        *string    `db:"pin_hash"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      *time.Time `db:"updated_at"`
}
