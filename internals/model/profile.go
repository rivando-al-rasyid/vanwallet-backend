package model

import "time"

// Profile maps to the "profiles" table.
type Profile struct {
	ID        string     `db:"id"`
	UserID    string     `db:"user_id"`
	FullName  *string    `db:"full_name"`
	Phone     *string    `db:"phone"`
	Photo     *string    `db:"photo"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}
