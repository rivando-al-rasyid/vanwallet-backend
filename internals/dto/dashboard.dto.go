package dto

// DashboardResponse combines balance summary and recent transaction history.
type DashboardResponse struct {
	CurrentBalance int64                `json:"current_balance"`
	TotalIncome    int64                `json:"total_income"`
	TotalExpense   int64                `json:"total_expense"`
	RecentTx       []TransactionResponse `json:"recent_transactions"`
}

// ChartPointResponse is a single bar in the financial chart.
type ChartPointResponse struct {
	Label   string `json:"label"`
	Income  int64  `json:"income"`
	Expense int64  `json:"expense"`
}

// TransactionReportResponse wraps the full chart dataset.
type TransactionReportResponse struct {
	Range  string               `json:"range"`  // "7days" | "30days"
	Points []ChartPointResponse `json:"points"`
}
