package dto

// UpdateProfileRequest is the payload for updating a user's profile.
type UpdateProfileRequest struct {
	FullName *string `json:"full_name" validate:"omitempty,min=2"`
	Phone    *string `json:"phone"     validate:"omitempty,e164"`
	Photo    *string `json:"photo"     validate:"omitempty,url"`
}

// ProfileResponse is the public representation of a user profile.
type ProfileResponse struct {
	ID       string  `json:"id"`
	UserID   string  `json:"user_id"`
	FullName *string `json:"full_name,omitempty"`
	Phone    *string `json:"phone,omitempty"`
	Photo    *string `json:"photo,omitempty"`
}
