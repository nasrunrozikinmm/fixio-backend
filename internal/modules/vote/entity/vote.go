package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Vote represents a user's vote on a post
type Vote struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;uniqueIndex:idx_user_post;not null"`
	PostID    uuid.UUID      `json:"post_id" gorm:"type:uuid;uniqueIndex:idx_user_post;index;not null"`
	Type      string         `json:"type" gorm:"not null"` // up | down
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (row *Vote) BeforeCreate(tx *gorm.DB) error {
	if row.ID == uuid.Nil {
		row.ID = uuid.New()
	}
	return nil
}

func (Vote) TableName() string {
	return "votes"
}
