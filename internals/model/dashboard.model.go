package model

type Dashboard struct {
	CurrentBalance int64 `json:"current_balance"`
	TotalIncome    int64 `json:"total_income"`
	TotalExpense   int64 `json:"total_expense"`
}
