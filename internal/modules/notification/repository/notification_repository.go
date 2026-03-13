package repository

import (
	"context"

	models "fixio/internal/modules/notification/entity"
	"fixio/pkg/data"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationRepository defines the data access interface for notifications
type NotificationRepository interface {
	data.BaseRepository[models.Notification]
	GetByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]models.Notification, int64, error)
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
	MarkAsRead(ctx context.Context, id, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
}

type notificationRepository struct {
	data.BaseRepository[models.Notification]
	db *gorm.DB
}

// NewNotificationRepository creates a new NotificationRepository
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{
		BaseRepository: data.NewBaseRepository[models.Notification](db),
		db:             db,
	}
}

func (r *notificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	query.Model(&models.Notification{}).Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&notifications).Error
	if err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

func (r *notificationRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&count).Error
	return count, err
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true).Error
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}
