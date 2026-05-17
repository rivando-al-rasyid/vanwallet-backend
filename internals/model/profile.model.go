package model

import "time"

type Profile struct {
	Id        string     `db:"id"`
	UserId    string     `db:"user_id"`
	FullName  string     `db:"full_name"`
	Phone     string     `db:"phone"`
	Photo     string     `db:"photo"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}
