package model

import "time"

type User struct {
	Id        string     `db:"id"`
	Email     string     `db:"email"`
	Password  string     `db:"password"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}

// type Profile struct {
// 	Id        string     `db:"id"`
// 	UserId    string     `db:"user_id"`
// 	FullName  string     `db:"full_name"`
// 	Phone     string     `db:"phone"`
// 	Photo     string     `db:"photo"`
// 	CreatedAt time.Time  `db:"created_at"`
// 	UpdatedAt *time.Time `db:"updated_at"`
// }

// type UserPin struct {
// 	Id        string     `db:"id"`
// 	UserId    string     `db:"user_id"`
// 	PinHash   string     `db:"pin_hash"`
// 	FailedAtt int        `db:"failed_attempts"`
// 	LockedUnt *time.Time `db:"locked_until"`
// 	CreatedAt time.Time  `db:"created_at"`
// 	UpdatedAt *time.Time `db:"updated_at"`
// }

// type wallet struct {
// 	Id        string     `db:"id"`
// 	UserId    string     `db:"user_id"`
// 	label     string     `db:"user_id"`
// 	balance   string     `db:"user_id"`
// 	CreatedAt time.Time  `db:"created_at"`
// 	UpdatedAt *time.Time `db:"updated_at"`
// }

// type transaction struct {
// 	Id        string `db:"id"`
// 	WalletID  string `db:"wallet_id"`
// 	Direction string `db:"directipn"`
// 	balance   string `db:"user_id"`

// 	Fee       string    `db:"admin_fee"`
// 	Status    string    `db:"status"`
// 	Note      string    `db:"note"`
// 	CreatedAt time.Time `db:"created_at"`
// }
