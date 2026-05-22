package dto

// TransactionResponse is the public representation of a single transaction ledger entry.
type TransactionResponse struct {
	ID        string `json:"id"`
	WalletID  string `json:"wallet_id"`
	Amount    int64  `json:"amount"`
	Direction string `json:"direction"`
	AdminFee  int64  `json:"admin_fee"`
	Status    string `json:"status"`
	Note      string `json:"note,omitempty"`
	CreatedAt string `json:"created_at"`
}

// TransactionListResponse wraps a paginated list of transactions.
type TransactionListResponse struct {
	Data  []TransactionResponse `json:"data"`
	Total int                   `json:"total"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
}

// CreateTransactionRequest is the payload for creating a new transaction.
type CreateTransactionRequest struct {
	WalletID  string  `json:"wallet_id"  validate:"required,uuid4"`
	Amount    int64   `json:"amount"     validate:"required,gt=0"`
	Direction string  `json:"direction"  validate:"required,oneof=TOPUP IN OUT"`
	AdminFee  int64   `json:"admin_fee"  validate:"omitempty,gte=0"`
	Note      *string `json:"note"       validate:"omitempty,max=255"`
}
