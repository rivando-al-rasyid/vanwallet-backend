package model

type Dashboard struct {
	CurrentBalance int64 `json:"current_balance"`
	TotalIncome    int64 `json:"total_income"`
	TotalExpense   int64 `json:"total_expense"`
}

type ChartPoint struct {
	Label   string `json:"label"`
	Income  int64  `json:"income"`
	Expense int64  `json:"expense"`
}
