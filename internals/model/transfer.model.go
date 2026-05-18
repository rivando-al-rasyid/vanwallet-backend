package model

import "github.com/google/uuid"

// Transfer links both sides of a peer-to-peer transfer.
//
//   - TransactionID        → the OUT transaction row (sender's ledger entry)
//   - SenderTransactionID  → the IN  transaction row (recipient's ledger entry)
//   - RecipientWalletID    → destination wallet
type Transfer struct {
	TransactionID       uuid.UUID `db:"transaction_id"`
	SenderTransactionID uuid.UUID `db:"sender_transaction_id"`
	RecipientWalletID   uuid.UUID `db:"recipient_wallet_id"`
	TransferCode        *string   `db:"transfer_code"`
}
