package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Sector represents a policy sector category
type Sector struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	Name      string         `json:"name" gorm:"uniqueIndex;not null"`
	Slug      string         `json:"slug" gorm:"uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (row *Sector) BeforeCreate(tx *gorm.DB) error {
	if row.ID == uuid.Nil {
		row.ID = uuid.New()
	}
	return nil
}

func (Sector) TableName() string {
	return "sectors"
}
