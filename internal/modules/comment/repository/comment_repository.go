package repositories

import (
	"context"

	models "fixio/internal/modules/comment/entity"
	"fixio/pkg/data"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CommentRepository defines comment-specific data access operations
type CommentRepository interface {
	data.BaseRepository[models.Comment]
	GetByPostID(ctx context.Context, postID uuid.UUID, page, limit int) ([]models.Comment, int64, error)
	FindByIDWithUser(ctx context.Context, id uuid.UUID) (*models.Comment, error)
}

type commentRepository struct {
	data.BaseRepository[models.Comment]
	db *gorm.DB
}

// NewCommentRepository creates a new CommentRepository
func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{
		BaseRepository: data.NewBaseRepository[models.Comment](db),
		db:             db,
	}
}

// GetByPostID retrieves top-level comments for a post with nested replies
func (r *commentRepository) GetByPostID(ctx context.Context, postID uuid.UUID, page, limit int) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var total int64

	// Count only top-level comments
	query := r.db.WithContext(ctx).Model(&models.Comment{}).
		Where("post_id = ? AND parent_id IS NULL", postID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := r.db.WithContext(ctx).
		Where("post_id = ? AND parent_id IS NULL", postID).
		Preload("User").
		Preload("Replies").
		Preload("Replies.User").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// FindByIDWithUser finds a comment by ID with user preloaded
func (r *commentRepository) FindByIDWithUser(ctx context.Context, id uuid.UUID) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("id = ?", id).
		First(&comment).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}
