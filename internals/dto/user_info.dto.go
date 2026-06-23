package dto

// UserInfoResponse is returned by GET /auth/me.
// Used to populate the app header: avatar, name, email, wallet balance, and PIN status.
type UserInfoResponse struct {
	ID             string  `json:"id"`
	Email          string  `json:"email"`
	FullName       *string `json:"full_name,omitempty"`
	Phone          *string `json:"phone,omitempty"`
	Photo          *string `json:"photo,omitempty"`
	CurrentBalance int64   `json:"current_balance"`
	PinHash        *string `json:"pin_hash,omitempty"`
}
