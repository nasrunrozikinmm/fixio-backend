package services

import (
	"context"
	"fmt"

	models "fixio/internal/modules/notification/entity"
	"fixio/internal/modules/notification/repository"
	"fixio/pkg/apperrors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationService defines the business logic interface for notifications
type NotificationService interface {
	GetNotifications(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]models.Notification, int64, error)
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
	MarkAsRead(ctx context.Context, id, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	CreateNotification(ctx context.Context, userID uuid.UUID, notifType string, actorID uuid.UUID, refID uuid.UUID, refType, message string) (*models.Notification, error)
}

type notificationService struct {
	notifRepo repository.NotificationRepository
}

// NewNotificationService creates a new NotificationService
func NewNotificationService(
	notifRepo repository.NotificationRepository,
) NotificationService {
	return &notificationService{
		notifRepo: notifRepo,
	}
}

func (s *notificationService) GetNotifications(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]models.Notification, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}
	return s.notifRepo.GetByUserID(ctx, userID, page, pageSize)
}

func (s *notificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.notifRepo.GetUnreadCount(ctx, userID)
}

func (s *notificationService) MarkAsRead(ctx context.Context, id, userID uuid.UUID) error {
	// Verify notification exists and belongs to user
	notif, err := s.notifRepo.FindBy(ctx, map[string]any{"id": id})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperrors.NewNotFound("Notifikasi tidak ditemukan", err)
		}
		return apperrors.NewInternal("Gagal memeriksa notifikasi", err)
	}

	if notif.UserID != userID {
		return apperrors.NewForbidden("Tidak berhak mengakses notifikasi ini", nil)
	}

	if err := s.notifRepo.MarkAsRead(ctx, id, userID); err != nil {
		return apperrors.NewInternal("Gagal menandai notifikasi sudah dibaca", err)
	}

	return nil
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	if err := s.notifRepo.MarkAllAsRead(ctx, userID); err != nil {
		return apperrors.NewInternal("Gagal menandai semua notifikasi sudah dibaca", err)
	}
	return nil
}

func (s *notificationService) CreateNotification(ctx context.Context, userID uuid.UUID, notifType string, actorID uuid.UUID, refID uuid.UUID, refType, message string) (*models.Notification, error) {
	// Don't notify self
	if userID == actorID {
		return nil, nil
	}

	notif := &models.Notification{
		UserID:        userID,
		Type:          notifType,
		ActorID:       actorID,
		ReferenceID:   refID,
		ReferenceType: refType,
		Message:       message,
		IsRead:        false,
	}

	created, err := s.notifRepo.Create(ctx, notif)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat notifikasi: %w", err)
	}

	return created, nil
}
