package dto

// ExpenseResponse is the public representation of an expense record.
type ExpenseResponse struct {
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	Category      string `json:"category,omitempty"`
	MerchantName  string `json:"merchant_name,omitempty"`
	Note          string `json:"note,omitempty"`
	CreatedAt     string `json:"created_at"`
}

// UpdateExpenseRequest allows enriching category and merchant name after the fact.
type UpdateExpenseRequest struct {
	Category     *string `json:"category"      validate:"omitempty,max=100"`
	MerchantName *string `json:"merchant_name" validate:"omitempty,max=100"`
}
