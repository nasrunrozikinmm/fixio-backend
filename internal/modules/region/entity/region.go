package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Region represents a geographical area (nasional, provinsi, kota)
type Region struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	Name      string         `json:"name" gorm:"not null"`
	Slug      string         `json:"slug" gorm:"uniqueIndex;not null"`
	Type      string         `json:"type" gorm:"not null"` // nasional | provinsi | kota
	ParentID  *uuid.UUID     `json:"parent_id" gorm:"type:uuid;index"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Parent   *Region  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []Region `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

func (row *Region) BeforeCreate(tx *gorm.DB) error {
	if row.ID == uuid.Nil {
		row.ID = uuid.New()
	}
	return nil
}

func (Region) TableName() string {
	return "regions"
}
