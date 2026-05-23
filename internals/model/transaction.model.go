package model

import (
	"time"

	"github.com/google/uuid"
)

// TransactionSummary holds aggregated financial data for a user.
type TransactionSummary struct {
	CurrentBalance int64 `json:"current_balance"`
	TotalIncome    int64 `json:"total_income"`
	TotalExpense   int64 `json:"total_expense"`
}

// ChartPoint is a single data point in the income/expense report chart.
type ChartPoint struct {
	Label   string `json:"label"`
	Income  int64  `json:"income"`
	Expense int64  `json:"expense"`
}

// Transaction maps to the central ledger table.
// Type matches the DB enum: EXPENSE, WITHDRAWAL, TRANSFER_IN, TRANSFER_OUT.
// Amount is always positive; direction is inferred from Type.
type Transaction struct {
	ID             uuid.UUID         `db:"id"              json:"id"`
	WalletID       uuid.UUID         `db:"wallet_id"       json:"wallet_id"`
	Type           TransactionType   `db:"type"            json:"type"`
	Amount         int64             `db:"amount"          json:"amount"`
	AdminFee       int64             `db:"admin_fee"       json:"admin_fee"`
	Status         TransactionStatus `db:"status"          json:"status"`
	IdempotencyKey *string           `db:"idempotency_key" json:"idempotency_key,omitempty"`
	Note           *string           `db:"note"            json:"note,omitempty"`
	CreatedAt      time.Time         `db:"created_at"      json:"created_at"`
	UpdatedAt      *time.Time        `db:"updated_at"      json:"updated_at,omitempty"`
}
