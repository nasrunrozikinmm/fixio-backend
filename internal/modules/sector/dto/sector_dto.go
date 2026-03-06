package dto

import "github.com/go-playground/validator/v10"

// CreateSectorRequest represents the request to create a sector
type CreateSectorRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}

func (r *CreateSectorRequest) GetValue() *CreateSectorRequest { return r }

func (r *CreateSectorRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	messages := []string{}
	for _, err := range errs {
		switch err.Field() {
		case "Name":
			if err.Tag() == "required" {
				messages = append(messages, "Nama sektor wajib diisi")
			} else {
				messages = append(messages, "Nama sektor maksimal 100 karakter")
			}
		}
	}
	return messages, nil
}

// UpdateSectorRequest represents the request to update a sector
type UpdateSectorRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}

func (r *UpdateSectorRequest) GetValue() *UpdateSectorRequest { return r }

func (r *UpdateSectorRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	messages := []string{}
	for _, err := range errs {
		switch err.Field() {
		case "Name":
			if err.Tag() == "required" {
				messages = append(messages, "Nama sektor wajib diisi")
			} else {
				messages = append(messages, "Nama sektor maksimal 100 karakter")
			}
		}
	}
	return messages, nil
}
