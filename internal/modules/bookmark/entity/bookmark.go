package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Bookmark represents a user's bookmark on a post
type Bookmark struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;uniqueIndex:idx_bookmark_user_post;not null"`
	PostID    uuid.UUID      `json:"post_id" gorm:"type:uuid;uniqueIndex:idx_bookmark_user_post;index;not null"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (row *Bookmark) BeforeCreate(tx *gorm.DB) error {
	if row.ID == uuid.Nil {
		row.ID = uuid.New()
	}
	return nil
}

func (Bookmark) TableName() string {
	return "bookmarks"
}
