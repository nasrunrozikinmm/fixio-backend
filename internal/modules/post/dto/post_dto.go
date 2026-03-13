package dto

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// CreatePostRequest represents the request to create a new post
type CreatePostRequest struct {
	Title          string     `json:"title" validate:"required,max=255"`
	SectorID       *uuid.UUID `json:"sector_id" validate:"omitempty"`
	RegionID       *uuid.UUID `json:"region_id" validate:"omitempty"`
	Criticism      string     `json:"criticism" validate:"required,min=10,max=10000"`
	Solution       string     `json:"solution" validate:"required,min=10,max=10000"`
	ImpactEstimate string     `json:"impact_estimate" validate:"omitempty"`
	References     string     `json:"references" validate:"omitempty"`
	Images         []string   `json:"images" validate:"omitempty,max=5,dive,url"`
	Status         string     `json:"status" validate:"omitempty,oneof=draft pending_review"`
}

func (r *CreatePostRequest) GetValue() *CreatePostRequest { return r }

func (r *CreatePostRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	messages := []string{}
	for _, err := range errs {
		switch err.Field() {
		case "Title":
			if err.Tag() == "required" {
				messages = append(messages, "Judul wajib diisi")
			} else {
				messages = append(messages, "Judul maksimal 255 karakter")
			}
		case "Criticism":
			if err.Tag() == "required" {
				messages = append(messages, "Kritik wajib diisi")
			} else {
				messages = append(messages, "Kritik minimal 10 karakter")
			}
		case "Solution":
			if err.Tag() == "required" {
				messages = append(messages, "Solusi wajib diisi")
			} else {
				messages = append(messages, "Solusi minimal 10 karakter")
			}
		}
	}
	return messages, nil
}

// UpdatePostRequest represents the request to update a post
type UpdatePostRequest struct {
	Title          string     `json:"title" validate:"omitempty,max=255"`
	SectorID       *uuid.UUID `json:"sector_id" validate:"omitempty"`
	RegionID       *uuid.UUID `json:"region_id" validate:"omitempty"`
	Criticism      string     `json:"criticism" validate:"omitempty,min=10,max=10000"`
	Solution       string     `json:"solution" validate:"omitempty,min=10,max=10000"`
	ImpactEstimate string     `json:"impact_estimate" validate:"omitempty"`
	References     string     `json:"references" validate:"omitempty"`
	Images         []string   `json:"images" validate:"omitempty,max=5,dive,url"`
}

func (r *UpdatePostRequest) GetValue() *UpdatePostRequest { return r }

func (r *UpdatePostRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	messages := []string{}
	for _, err := range errs {
		switch err.Field() {
		case "Title":
			messages = append(messages, "Judul maksimal 255 karakter")
		case "Criticism":
			messages = append(messages, "Kritik minimal 10 karakter")
		case "Solution":
			messages = append(messages, "Solusi minimal 10 karakter")
		}
	}
	return messages, nil
}

// PostFilter represents query filters for listing posts
type PostFilter struct {
	SectorID *uuid.UUID  `json:"sector_id" query:"sector_id"`
	RegionID *uuid.UUID  `json:"region_id" query:"region_id"`
	Status   string      `json:"status" query:"status"`
	UserID   *uuid.UUID  `json:"user_id" query:"user_id"`
	Search   string      `json:"search" query:"search"`
	UserIDs  []uuid.UUID `json:"-" query:"-"` // Programmatic only, not from query string
}
