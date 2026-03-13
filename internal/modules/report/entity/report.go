package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	userModels "fixio/internal/modules/auth/entity"
)

// Report reason constants
const (
	ReasonSpam          = "spam"
	ReasonHarassment    = "harassment"
	ReasonMisinformation = "misinformation"
	ReasonHateSpeech    = "hate_speech"
	ReasonOther         = "other"
)

// Report status constants
const (
	ReportStatusPending  = "pending"
	ReportStatusReviewed = "reviewed"
	ReportStatusDismissed = "dismissed"
)

// Report represents a content report/flag from a user
type Report struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	ReporterID    uuid.UUID      `json:"reporter_id" gorm:"type:uuid;index;not null"`
	TargetType    string         `json:"target_type" gorm:"not null"` // "post" or "comment"
	TargetID      uuid.UUID      `json:"target_id" gorm:"type:uuid;index;not null"`
	Reason        string         `json:"reason" gorm:"not null"`
	Description   string         `json:"description" gorm:"type:text"`
	Status        string         `json:"status" gorm:"default:pending;not null"`
	ReviewedBy    *uuid.UUID     `json:"reviewed_by" gorm:"type:uuid"`
	ReviewNote    string         `json:"review_note"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Reporter *userModels.User `json:"reporter,omitempty" gorm:"foreignKey:ReporterID"`
	Reviewer *userModels.User `json:"reviewer,omitempty" gorm:"foreignKey:ReviewedBy"`
}

func (row *Report) BeforeCreate(tx *gorm.DB) error {
	if row.ID == uuid.Nil {
		row.ID = uuid.New()
	}
	if row.Status == "" {
		row.Status = ReportStatusPending
	}
	return nil
}

func (Report) TableName() string {
	return "reports"
}

// ValidReasons returns all valid report reasons
func ValidReasons() []string {
	return []string{ReasonSpam, ReasonHarassment, ReasonMisinformation, ReasonHateSpeech, ReasonOther}
}

// IsValidReason validates a reason string
func IsValidReason(reason string) bool {
	for _, r := range ValidReasons() {
		if r == reason {
			return true
		}
	}
	return false
}
