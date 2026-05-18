package model

// TransactionStatus represents the status of a transaction.
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "PENDING"
	TransactionStatusSuccess   TransactionStatus = "SUCCESS"
	TransactionStatusFailed    TransactionStatus = "FAILED"
	TransactionStatusCancelled TransactionStatus = "CANCELLED"
)

// PaymentMethod represents supported top-up payment methods.
type PaymentMethod string

const (
	PaymentMethodBRI   PaymentMethod = "BRI"
	PaymentMethodBCA   PaymentMethod = "BCA"
	PaymentMethodDANA  PaymentMethod = "DANA"
	PaymentMethodGoPay PaymentMethod = "GOPAY"
	PaymentMethodOVO   PaymentMethod = "OVO"
)

// Direction represents the flow direction of a transaction.
type Direction string

const (
	DirectionIn  Direction = "IN"
	DirectionOut Direction = "OUT"
)
