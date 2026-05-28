package model

import "time"

// HistoryItem is a flat unified record used by GetAllHistory.
// Source is "topup" or "transaction".
type HistoryItem struct {
	ID            string
	Source        string
	Type          string
	Amount        int64
	AdminFee      int64
	Status        string
	PaymentMethod string
	Note          string
	WalletID      string
	CreatedAt     time.Time
}
