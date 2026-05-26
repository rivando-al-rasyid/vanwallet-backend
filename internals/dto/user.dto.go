package dto

import "github.com/google/uuid"

// RegisterRequest — only email and password required.
// No username field; username concept has been removed entirely.
type RegisterRequest struct {
	Email    string `json:"email"    binding:"required,email"  example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8"  example:"P@ssw0rd123" minLength:"8"`
}

// LoginRequest is the payload for authenticating.
type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required"       example:"P@ssw0rd123"`
}

// UserResponse is the public representation after register/login.
type UserResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}
