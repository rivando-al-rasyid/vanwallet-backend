package model

import "github.com/google/uuid"

// Topup holds payment-gateway-specific data for a top-up transaction.
// ExternalReference is the reference ID from the payment gateway (DANA, GoPay, etc.)
// and its uniqueness constraint enables idempotency checks.
type Topup struct {
	TransactionID     uuid.UUID      `db:"transaction_id"`
	PaymentMethod     *PaymentMethod `db:"payment_method"`
	ExternalReference *string        `db:"external_reference"`
}
