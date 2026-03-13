package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	userModels "fixio/internal/modules/auth/entity"
	regionModels "fixio/internal/modules/region/entity"
	sectorModels "fixio/internal/modules/sector/entity"
)

// StringArray is a custom type for storing JSON string arrays in PostgreSQL jsonb columns
type StringArray []string

// Scan implements the sql.Scanner interface for reading from the database
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("StringArray.Scan: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, a)
}

// Value implements the driver.Valuer interface for writing to the database
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

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
	Images         StringArray    `json:"images" gorm:"type:jsonb;default:'[]'"`
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
