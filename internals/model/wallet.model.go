package model

import "time"

type wallet struct {
	Id        string     `db:"id"`
	UserId    string     `db:"user_id"`
	PinHash   string     `db:"pin_hash"`
	FailedAtt int        `db:"failed_attempts"`
	LockedUnt *time.Time `db:"locked_until"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}
