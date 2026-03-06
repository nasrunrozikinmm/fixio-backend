package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a platform user authenticated via OAuth
type User struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	Name       string         `json:"name" gorm:"not null"`
	Email      string         `json:"email" gorm:"uniqueIndex;not null"`
	AvatarURL  string         `json:"avatar_url"`
	Bio        string         `json:"bio"`
	Location   string         `json:"location"`
	Role       string         `json:"role" gorm:"default:creator;not null"` // creator | moderator | administrator
	Provider   string         `json:"provider" gorm:"not null"`             // google | facebook
	ProviderID string         `json:"-" gorm:"uniqueIndex;not null"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

func (row *User) BeforeCreate(tx *gorm.DB) error {
	if row.ID == uuid.Nil {
		row.ID = uuid.New()
	}
	if row.Role == "" {
		row.Role = "creator"
	}
	return nil
}

func (User) TableName() string {
	return "users"
}
