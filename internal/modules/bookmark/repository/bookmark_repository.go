package repository

import (
	"context"

	models "fixio/internal/modules/bookmark/entity"
	"fixio/pkg/data"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BookmarkRepository defines the data access interface for bookmarks
type BookmarkRepository interface {
	data.BaseRepository[models.Bookmark]
	FindByUserAndPost(ctx context.Context, userID, postID uuid.UUID) (*models.Bookmark, error)
	DeleteByUserAndPost(ctx context.Context, userID, postID uuid.UUID) error
	GetUserBookmarkPostIDs(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]uuid.UUID, int64, error)
	IsBookmarked(ctx context.Context, userID, postID uuid.UUID) (bool, error)
}

type bookmarkRepository struct {
	data.BaseRepository[models.Bookmark]
	db *gorm.DB
}

// NewBookmarkRepository creates a new BookmarkRepository
func NewBookmarkRepository(db *gorm.DB) BookmarkRepository {
	return &bookmarkRepository{
		BaseRepository: data.NewBaseRepository[models.Bookmark](db),
		db:             db,
	}
}

func (r *bookmarkRepository) FindByUserAndPost(ctx context.Context, userID, postID uuid.UUID) (*models.Bookmark, error) {
	var bookmark models.Bookmark
	err := r.db.WithContext(ctx).Where("user_id = ? AND post_id = ?", userID, postID).First(&bookmark).Error
	if err != nil {
		return nil, err
	}
	return &bookmark, nil
}

func (r *bookmarkRepository) DeleteByUserAndPost(ctx context.Context, userID, postID uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Where("user_id = ? AND post_id = ?", userID, postID).Delete(&models.Bookmark{}).Error
}

func (r *bookmarkRepository) GetUserBookmarkPostIDs(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]uuid.UUID, int64, error) {
	var total int64
	var postIDs []uuid.UUID

	query := r.db.WithContext(ctx).Model(&models.Bookmark{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Pluck("post_id", &postIDs).Error
	if err != nil {
		return nil, 0, err
	}

	return postIDs, total, nil
}

func (r *bookmarkRepository) IsBookmarked(ctx context.Context, userID, postID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Bookmark{}).Where("user_id = ? AND post_id = ?", userID, postID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
