package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	userModels "fixio/internal/modules/auth/entity"
	regionModels "fixio/internal/modules/region/entity"
	sectorModels "fixio/internal/modules/sector/entity"
)

// Post represents a policy critique + solution
type Post struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	UserID         uuid.UUID      `json:"user_id" gorm:"type:uuid;index;not null"`
	Title          string         `json:"title" gorm:"not null"`
	SectorID       *uuid.UUID     `json:"sector_id" gorm:"type:uuid;index"`
	RegionID       *uuid.UUID     `json:"region_id" gorm:"type:uuid;index"`
	Criticism      string         `json:"criticism" gorm:"type:text;not null"`
	Solution       string         `json:"solution" gorm:"type:text;not null"`
	ImpactEstimate string         `json:"impact_estimate" gorm:"type:text"`
	References     string         `json:"references" gorm:"type:text"`
	Status         string         `json:"status" gorm:"default:pending_review;not null"` // draft|pending_review|approved|rejected
	ReviewedBy     *uuid.UUID     `json:"reviewed_by" gorm:"type:uuid"`
	ReviewNote     string         `json:"review_note"`
	VoteCount      int            `json:"vote_count" gorm:"default:0"`
	CommentCount   int            `json:"comment_count" gorm:"default:0"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	User     userModels.User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Sector   *sectorModels.Sector `json:"sector,omitempty" gorm:"foreignKey:SectorID"`
	Region   *regionModels.Region `json:"region,omitempty" gorm:"foreignKey:RegionID"`
	Reviewer *userModels.User     `json:"reviewer,omitempty" gorm:"foreignKey:ReviewedBy"`
}

func (row *Post) BeforeCreate(tx *gorm.DB) error {
	if row.ID == uuid.Nil {
		row.ID = uuid.New()
	}
	if row.Status == "" {
		row.Status = "pending_review"
	}
	return nil
}

func (Post) TableName() string {
	return "posts"
}
