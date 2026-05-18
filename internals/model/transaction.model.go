package model

import (
	"time"

	"github.com/google/uuid"
)

// Transaction is the central ledger entry.
// Direction is the single source of truth for IN/OUT.
// Each row represents ONE side of a transaction.
// For transfers, two rows are created — one OUT for the sender and one IN
// for the recipient — linked via the Transfer table.
type Transaction struct {
	ID        uuid.UUID         `db:"id"`
	WalletID  uuid.UUID         `db:"wallet_id"`
	Amount    int64             `db:"amount"`
	Direction Direction         `db:"direction"`
	AdminFee  int64             `db:"admin_fee"`
	Status    TransactionStatus `db:"status"`
	Note      *string           `db:"note"`
	CreatedAt time.Time         `db:"created_at"`
}
