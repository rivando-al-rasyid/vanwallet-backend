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

type Transaction struct {
	ID        uuid.UUID         `json:"id"`
	WalletID  uuid.UUID         `json:"wallet_id"`
	Amount    int64             `json:"amount"`
	Direction Direction         `json:"direction"`
	AdminFee  int64             `json:"admin_fee"`
	Status    TransactionStatus `json:"status"`
	Note      *string           `json:"note,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}
