package dto

// TopupRequest is the payload for initiating a wallet top-up.
type TopupRequest struct {
	Amount        int64  `json:"amount"         validate:"required,gt=0"`
	PaymentMethod string `json:"payment_method" validate:"required,oneof=BRI BCA DANA GOPAY OVO"`
}

// TopupResponse is returned after a top-up is initiated.
type TopupResponse struct {
	TransactionID     string `json:"transaction_id"`
	Amount            int64  `json:"amount"`
	PaymentMethod     string `json:"payment_method"`
	ExternalReference string `json:"external_reference,omitempty"`
	Status            string `json:"status"`
}
