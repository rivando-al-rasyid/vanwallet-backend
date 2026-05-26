package dto

// WithdrawalRequest is the payload for withdrawing funds to a bank account.
type WithdrawalRequest struct {
	WalletID      string `json:"wallet_id"      binding:"required,uuid4"`
	Amount        int64  `json:"amount"         binding:"required,gt=0"`
	BankName      string `json:"bank_name"      binding:"required,min=2,max=50"`
	AccountNumber string `json:"account_number" binding:"required,min=6,max=20"`
	AccountHolder string `json:"account_holder" binding:"required,min=2,max=100"`
}

// WithdrawalResponse is returned after a withdrawal is submitted.
type WithdrawalResponse struct {
	TransactionID string `json:"transaction_id"`
	WalletID      string `json:"wallet_id"`
	Amount        int64  `json:"amount"`
	AdminFee      int64  `json:"admin_fee"`
	BankName      string `json:"bank_name"`
	AccountNumber string `json:"account_number"`
	AccountHolder string `json:"account_holder"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}
