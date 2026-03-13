package dto

import (
	"github.com/google/uuid"
)

// FollowResponse represents a follow relationship in API responses
type FollowResponse struct {
	ID          uuid.UUID `json:"id"`
	FollowerID  uuid.UUID `json:"follower_id"`
	FollowingID uuid.UUID `json:"following_id"`
	CreatedAt   string    `json:"created_at"`
}

// FollowStatusResponse indicates whether the current user follows a target user
type FollowStatusResponse struct {
	IsFollowing bool `json:"is_following"`
}

// FollowCountsResponse returns follower and following counts
type FollowCountsResponse struct {
	FollowerCount  int64 `json:"follower_count"`
	FollowingCount int64 `json:"following_count"`
}
