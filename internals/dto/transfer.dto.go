package dto

// TransferRequest is the payload for transferring funds to another wallet.
type TransferRequest struct {
	SenderWalletID    string `json:"sender_wallet_id"    binding:"required,uuid4"`
	RecipientWalletID string `json:"recipient_wallet_id" binding:"required,uuid4"`
	Amount            int64  `json:"amount"              binding:"required,gt=0"`
	Note              string `json:"note"                binding:"omitempty,max=255"`
}
