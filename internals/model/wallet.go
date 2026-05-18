package model

import "time"

// Wallet maps to the "wallets" table.
type Wallet struct {
	ID        string     `db:"id"`
	UserID    string     `db:"user_id"`
	Label     string     `db:"label"`
	Balance   int64      `db:"balance"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}
