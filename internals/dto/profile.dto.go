package dto

import "time"

// Profile defines the exact JSON structure sent back to the client
type Profile struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	FullName  *string    `json:"full_name"`
	Phone     *string    `json:"phone"`
	Photo     *string    `json:"photo"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}
