package dto

// UpdateProfileRequest is the payload for updating a user's profile.
import "mime/multipart"

type UpdateProfileRequest struct {
	FullName *string               `form:"full_name" validate:"omitempty,min=2"`
	Phone    *string               `form:"phone"     validate:"omitempty,e164"`
	Photo    *multipart.FileHeader `form:"photo"     validate:"omitempty"` // Swapped to accept a file
}

// ProfileResponse is the public representation of a user profile.
type ProfileResponse struct {
	FullName *string `json:"full_name,omitempty"`
	Phone    *string `json:"phone,omitempty"`
	Photo    *string `json:"photo,omitempty"`
}
