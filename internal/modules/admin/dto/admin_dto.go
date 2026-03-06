package dto

import "github.com/go-playground/validator/v10"

// ChangeRoleRequest represents the request to change a user's role
type ChangeRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=creator moderator administrator"`
}

func (r *ChangeRoleRequest) GetValue() *ChangeRoleRequest { return r }

func (r *ChangeRoleRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	messages := []string{}
	for _, err := range errs {
		switch err.Field() {
		case "Role":
			if err.Tag() == "required" {
				messages = append(messages, "Role wajib diisi")
			} else {
				messages = append(messages, "Role harus 'creator', 'moderator', atau 'administrator'")
			}
		}
	}
	return messages, nil
}

// StatsResponse represents the admin stats overview
type StatsResponse struct {
	TotalUsers         int64 `json:"total_users"`
	TotalPosts         int64 `json:"total_posts"`
	TotalApproved      int64 `json:"total_approved"`
	TotalPendingReview int64 `json:"total_pending_review"`
	TotalRejected      int64 `json:"total_rejected"`
	TotalComments      int64 `json:"total_comments"`
	TotalVotes         int64 `json:"total_votes"`
}
