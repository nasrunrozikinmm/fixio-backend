package dto

import "github.com/go-playground/validator/v10"

// CreateReportRequest is the DTO for reporting content
type CreateReportRequest struct {
	TargetType  string `json:"target_type" validate:"required,oneof=post comment"`
	TargetID    string `json:"target_id" validate:"required,uuid"`
	Reason      string `json:"reason" validate:"required,oneof=spam harassment misinformation hate_speech other"`
	Description string `json:"description" validate:"max=1000"`
}

func (r *CreateReportRequest) GetValue() *CreateReportRequest { return r }

func (r *CreateReportRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	messages := []string{}
	for _, err := range errs {
		switch err.Field() {
		case "TargetType":
			messages = append(messages, "Tipe target harus 'post' atau 'comment'")
		case "TargetID":
			messages = append(messages, "ID target tidak valid")
		case "Reason":
			messages = append(messages, "Alasan laporan tidak valid")
		case "Description":
			messages = append(messages, "Deskripsi maksimal 1000 karakter")
		}
	}
	return messages, nil
}

// ReviewReportRequest is the DTO for reviewing a report (admin/moderator)
type ReviewReportRequest struct {
	Status     string `json:"status" validate:"required,oneof=reviewed dismissed"`
	ReviewNote string `json:"review_note" validate:"max=500"`
}

func (r *ReviewReportRequest) GetValue() *ReviewReportRequest { return r }

func (r *ReviewReportRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	messages := []string{}
	for _, err := range errs {
		switch err.Field() {
		case "Status":
			messages = append(messages, "Status harus 'reviewed' atau 'dismissed'")
		case "ReviewNote":
			messages = append(messages, "Catatan review maksimal 500 karakter")
		}
	}
	return messages, nil
}
