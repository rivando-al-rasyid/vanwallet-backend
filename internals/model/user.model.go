package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents the core identity record.
// Credentials only — no PII stored here.
type User struct {
	ID        uuid.UUID  `db:"id"`
	Email     string     `db:"email"`
	Username  string     `db:"username"`
	Password  string     `db:"password"`
	Token     *string    `db:"token"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}
