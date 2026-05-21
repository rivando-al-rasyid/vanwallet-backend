package dto

import "github.com/google/uuid"

// RegisterRequest is the payload for creating a new user account.
type RegisterRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest is the payload for authenticating a user.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type ChangePasswordRequest struct {
	Password string `json:"password" validate:"required,min=8"`
}

// LoginResponse is returned after a successful login.
type LoginResponse struct {
	Token string `json:"token"`
}

// UserResponse is the public-safe representation of a user.
type UserResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}
