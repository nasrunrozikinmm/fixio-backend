package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CommentVote represents a user's upvote on a comment (one vote per user per comment)
type CommentVote struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;uniqueIndex:idx_user_comment_vote;not null"`
	CommentID uuid.UUID      `json:"comment_id" gorm:"type:uuid;uniqueIndex:idx_user_comment_vote;index;not null"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (row *CommentVote) BeforeCreate(tx *gorm.DB) error {
	if row.ID == uuid.Nil {
		row.ID = uuid.New()
	}
	return nil
}

func (CommentVote) TableName() string {
	return "comment_votes"
}
