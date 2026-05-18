package dto

// TransferRequest is the payload for transferring funds to another wallet.
type TransferRequest struct {
	RecipientWalletID string `json:"recipient_wallet_id" validate:"required,uuid4"`
	Amount            int64  `json:"amount"              validate:"required,gt=0"`
	Note              string `json:"note"                validate:"omitempty,max=255"`
	Pin               string `json:"pin"                 validate:"required,len=6,numeric"`
}

// TransferResponse is returned after a transfer is completed.
type TransferResponse struct {
	TransferCode        string `json:"transfer_code"`
	SenderTransaction   TransactionResponse `json:"sender_transaction"`
	RecipientTransaction TransactionResponse `json:"recipient_transaction"`
}
