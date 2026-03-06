package dto

import "github.com/go-playground/validator/v10"

// VoteRequest represents the request to vote on a post
type VoteRequest struct {
	Type string `json:"type" validate:"required,oneof=up down"`
}

func (r *VoteRequest) GetValue() *VoteRequest { return r }

func (r *VoteRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	messages := []string{}
	for _, err := range errs {
		switch err.Field() {
		case "Type":
			if err.Tag() == "required" {
				messages = append(messages, "Tipe vote wajib diisi")
			} else {
				messages = append(messages, "Tipe vote harus 'up' atau 'down'")
			}
		}
	}
	return messages, nil
}
