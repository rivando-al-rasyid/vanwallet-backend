package dto

import "time"

// SetPinRequest is the payload for setting or updating a user's PIN.
type SetPinRequest struct {
	PinHash        *string    `json:"pin_hash"         validate:"required,len=6,numeric"`
	FailedAttempts *int       `json:"failed_attempts"`
	LockedUntil    *time.Time `json:"locked_until"`
}

// VerifyPinRequest is the payload for verifying a user's PIN.
type VerifyPinRequest struct {
	Pin string `json:"pin" validate:"required,len=6,numeric"`
}
