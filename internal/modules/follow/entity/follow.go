package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Follow represents a user following another user
type Follow struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	FollowerID  uuid.UUID      `json:"follower_id" gorm:"type:uuid;uniqueIndex:idx_follower_following;not null"`
	FollowingID uuid.UUID      `json:"following_id" gorm:"type:uuid;uniqueIndex:idx_follower_following;index;not null"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (row *Follow) BeforeCreate(tx *gorm.DB) error {
	if row.ID == uuid.Nil {
		row.ID = uuid.New()
	}
	return nil
}

func (Follow) TableName() string {
	return "follows"
}
