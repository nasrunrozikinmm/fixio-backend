package dto

import (
	"time"

	"github.com/google/uuid"
)

// BookmarkResponse is the response DTO for a single bookmark
type BookmarkResponse struct {
	ID        uuid.UUID `json:"id"`
	PostID    uuid.UUID `json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}

// BookmarkStatusResponse is the response DTO for bookmark status check
type BookmarkStatusResponse struct {
	IsBookmarked bool `json:"is_bookmarked"`
}
