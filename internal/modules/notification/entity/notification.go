package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Notification types
const (
	NotifTypeNewFollower  = "new_follower"
	NotifTypePostVote     = "post_vote"
	NotifTypePostComment  = "post_comment"
	NotifTypeCommentVote  = "comment_vote"
)

// Notification represents an in-app notification for a user
type Notification struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	UserID        uuid.UUID      `json:"user_id" gorm:"type:uuid;index;not null"`
	Type          string         `json:"type" gorm:"not null"`                   // new_follower | post_vote | post_comment
	ActorID       uuid.UUID      `json:"actor_id" gorm:"type:uuid;not null"`     // who triggered this notification
	ReferenceID   uuid.UUID      `json:"reference_id" gorm:"type:uuid"`          // the entity being referenced (post, follow, etc.)
	ReferenceType string         `json:"reference_type" gorm:"type:varchar(50)"` // post | user
	Message       string         `json:"message" gorm:"type:text;not null"`      // human-readable message
	IsRead        bool           `json:"is_read" gorm:"default:false;index"`
	CreatedAt     time.Time      `json:"created_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

func (row *Notification) BeforeCreate(tx *gorm.DB) error {
	if row.ID == uuid.Nil {
		row.ID = uuid.New()
	}
	return nil
}

func (Notification) TableName() string {
	return "notifications"
}
