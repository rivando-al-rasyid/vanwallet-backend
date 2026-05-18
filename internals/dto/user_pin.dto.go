package dto

// SetPinRequest is the payload for setting or updating a user's PIN.
type SetPinRequest struct {
	Pin        string `json:"pin"         validate:"required,len=6,numeric"`
	ConfirmPin string `json:"confirm_pin" validate:"required,eqfield=Pin"`
}

// VerifyPinRequest is the payload for verifying a user's PIN.
type VerifyPinRequest struct {
	Pin string `json:"pin" validate:"required,len=6,numeric"`
}
