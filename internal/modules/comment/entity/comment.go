package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	userModels "fixio/internal/modules/auth/entity"
)

// Comment represents a comment on a post with 1-level nesting support
type Comment struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	PostID    uuid.UUID      `json:"post_id" gorm:"type:uuid;index;not null"`
	ParentID  *uuid.UUID     `json:"parent_id" gorm:"type:uuid;index"` // nullable — reply
	Content   string         `json:"content" gorm:"type:text;not null"`
	VoteCount int            `json:"vote_count" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	User    userModels.User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Replies []Comment       `json:"replies,omitempty" gorm:"foreignKey:ParentID"`
}

func (row *Comment) BeforeCreate(tx *gorm.DB) error {
	if row.ID == uuid.Nil {
		row.ID = uuid.New()
	}
	return nil
}

func (Comment) TableName() string {
	return "comments"
}
