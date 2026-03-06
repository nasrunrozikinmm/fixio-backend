package dto

import "github.com/go-playground/validator/v10"

// CreateCommentRequest represents the request to create a comment
type CreateCommentRequest struct {
	Content  string `json:"content" validate:"required,min=1,max=2000"`
	ParentID string `json:"parent_id" validate:"omitempty,uuid"`
}

func (r *CreateCommentRequest) GetValue() *CreateCommentRequest { return r }

func (r *CreateCommentRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	messages := []string{}
	for _, err := range errs {
		switch err.Field() {
		case "Content":
			if err.Tag() == "required" || err.Tag() == "min" {
				messages = append(messages, "Komentar wajib diisi")
			} else {
				messages = append(messages, "Komentar maksimal 2000 karakter")
			}
		case "ParentID":
			messages = append(messages, "ID parent komentar tidak valid")
		}
	}
	return messages, nil
}
