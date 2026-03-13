package dto

import (
	"time"

	"github.com/google/uuid"
)

// NotificationResponse is the response DTO for a single notification
type NotificationResponse struct {
	ID            uuid.UUID `json:"id"`
	Type          string    `json:"type"`
	ActorID       uuid.UUID `json:"actor_id"`
	ActorName     string    `json:"actor_name,omitempty"`
	ReferenceID   uuid.UUID `json:"reference_id"`
	ReferenceType string    `json:"reference_type"`
	Message       string    `json:"message"`
	IsRead        bool      `json:"is_read"`
	CreatedAt     time.Time `json:"created_at"`
}

// UnreadCountResponse is the response DTO for unread count
type UnreadCountResponse struct {
	Count int64 `json:"count"`
}
