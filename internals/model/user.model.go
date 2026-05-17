package model

import (
	"time"
)

type User struct {
	Id        string     `db:"id"`
	Email     string     `db:"email"`
	Password  string     `db:"password"`
	token     string     `db:"token"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}
