package model

import (
	"time"

	"github.com/google/uuid"
)

// Profile stores PII separately from users for easier data management.
// One-to-one relationship with User.
type Profile struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	FullName  *string    `db:"full_name"`
	Phone     *string    `db:"phone"`
	Photo     *string    `db:"photo"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}
