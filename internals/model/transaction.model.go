package model

import "time"

type Transaction struct {
	Id        string    `db:"id"`
	WalletID  string    `db:"wallet_id"`
	Direction string    `db:"directipn"`
	Fee       string    `db:"admin_fee"`
	Status    string    `db:"status"`
	Note      string    `db:"note"`
	CreatedAt time.Time `db:"created_at"`
}
