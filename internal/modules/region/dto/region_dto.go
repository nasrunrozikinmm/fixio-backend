package dto

import "github.com/go-playground/validator/v10"

// CreateRegionRequest represents the request to create a region
type CreateRegionRequest struct {
	Name     string `json:"name" validate:"required,max=100"`
	Type     string `json:"type" validate:"required,oneof=nasional provinsi kota"`
	ParentID string `json:"parent_id" validate:"omitempty,uuid"`
}

func (r *CreateRegionRequest) GetValue() *CreateRegionRequest { return r }

func (r *CreateRegionRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	messages := []string{}
	for _, err := range errs {
		switch err.Field() {
		case "Name":
			if err.Tag() == "required" {
				messages = append(messages, "Nama wilayah wajib diisi")
			} else {
				messages = append(messages, "Nama wilayah maksimal 100 karakter")
			}
		case "Type":
			if err.Tag() == "required" {
				messages = append(messages, "Tipe wilayah wajib diisi")
			} else {
				messages = append(messages, "Tipe wilayah harus 'nasional', 'provinsi', atau 'kota'")
			}
		case "ParentID":
			messages = append(messages, "ID parent wilayah tidak valid")
		}
	}
	return messages, nil
}

// UpdateRegionRequest represents the request to update a region
type UpdateRegionRequest struct {
	Name     string `json:"name" validate:"omitempty,max=100"`
	Type     string `json:"type" validate:"omitempty,oneof=nasional provinsi kota"`
	ParentID string `json:"parent_id" validate:"omitempty,uuid"`
}

func (r *UpdateRegionRequest) GetValue() *UpdateRegionRequest { return r }

func (r *UpdateRegionRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	messages := []string{}
	for _, err := range errs {
		switch err.Field() {
		case "Name":
			messages = append(messages, "Nama wilayah maksimal 100 karakter")
		case "Type":
			messages = append(messages, "Tipe wilayah harus 'nasional', 'provinsi', atau 'kota'")
		}
	}
	return messages, nil
}
