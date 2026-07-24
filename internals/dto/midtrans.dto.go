package dto

// MidtransNotification is the payload sent by Midtrans payment notifications.
type MidtransNotification struct {
	OrderID           string `json:"order_id"`
	TransactionStatus string `json:"transaction_status"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
	FraudStatus       string `json:"fraud_status"`
	PaymentType       string `json:"payment_type"`
	TransactionID     string `json:"transaction_id"`
	StatusMessage     string `json:"status_message"`
}
