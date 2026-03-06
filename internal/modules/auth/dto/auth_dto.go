package dto

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// OAuthUserInfo represents user info from OAuth provider
type OAuthUserInfo struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Provider  string `json:"provider"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token string      `json:"token"`
	User  UserProfile `json:"user"`
}

// UserProfile represents the user profile response
type UserProfile struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url"`
	Bio       string    `json:"bio"`
	Location  string    `json:"location"`
	Role      string    `json:"role"`
	Provider  string    `json:"provider"`
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	Name     string `json:"name" validate:"omitempty,max=255"`
	Bio      string `json:"bio" validate:"omitempty,max=500"`
	Location string `json:"location" validate:"omitempty,max=255"`
}

func (r *UpdateProfileRequest) GetValue() *UpdateProfileRequest { return r }

func (r *UpdateProfileRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	messages := []string{}
	for _, err := range errs {
		switch err.Field() {
		case "Name":
			messages = append(messages, "Nama maksimal 255 karakter")
		case "Bio":
			messages = append(messages, "Bio maksimal 500 karakter")
		case "Location":
			messages = append(messages, "Lokasi maksimal 255 karakter")
		}
	}
	return messages, nil
}
