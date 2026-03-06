package dto

import "github.com/go-playground/validator/v10"

// ReviewPostRequest represents the request to approve/reject a post
type ReviewPostRequest struct {
	ReviewNote string `json:"review_note" validate:"omitempty,max=500"`
}

func (r *ReviewPostRequest) GetValue() *ReviewPostRequest { return r }

func (r *ReviewPostRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	messages := []string{}
	for _, err := range errs {
		switch err.Field() {
		case "ReviewNote":
			messages = append(messages, "Catatan review maksimal 500 karakter")
		}
	}
	return messages, nil
}

// RejectPostRequest represents the request to reject a post (note is required)
type RejectPostRequest struct {
	ReviewNote string `json:"review_note" validate:"required,max=500"`
}

func (r *RejectPostRequest) GetValue() *RejectPostRequest { return r }

func (r *RejectPostRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	messages := []string{}
	for _, err := range errs {
		switch err.Field() {
		case "ReviewNote":
			if err.Tag() == "required" {
				messages = append(messages, "Alasan penolakan wajib diisi")
			} else {
				messages = append(messages, "Catatan review maksimal 500 karakter")
			}
		}
	}
	return messages, nil
}
