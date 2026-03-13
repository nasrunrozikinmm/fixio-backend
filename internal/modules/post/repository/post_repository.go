package repositories

import (
	"context"
	"fmt"

	models "fixio/internal/modules/post/entity"
	"fixio/pkg/data"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PostRepository defines post-specific data access operations
type PostRepository interface {
	data.BaseRepository[models.Post]
	GetWithRelations(ctx context.Context, page, limit int, sort string, filter map[string]any, search string) ([]models.Post, int64, error)
	FindByIDWithRelations(ctx context.Context, id uuid.UUID) (*models.Post, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]models.Post, int64, error)
	GetByStatus(ctx context.Context, status string, page, limit int) ([]models.Post, int64, error)
	IncrementVoteCount(ctx context.Context, postID uuid.UUID, delta int) error
	IncrementCommentCount(ctx context.Context, postID uuid.UUID, delta int) error
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Post, error)
	GetRelated(ctx context.Context, postID uuid.UUID, sectorID *uuid.UUID, limit int) ([]models.Post, error)
}

type postRepository struct {
	data.BaseRepository[models.Post]
	db *gorm.DB
}

// NewPostRepository creates a new PostRepository
func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepository{
		BaseRepository: data.NewBaseRepository[models.Post](db),
		db:             db,
	}
}

// GetWithRelations retrieves posts with preloaded relations, filtering, and search
func (r *postRepository) GetWithRelations(ctx context.Context, page, limit int, sort string, filter map[string]any, search string) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Post{})

	// Apply filters
	for key, value := range filter {
		if value != nil && value != "" {
			if key == "user_ids" {
				query = query.Where("user_id IN (?)", value)
			} else {
				query = query.Where(fmt.Sprintf("%s = ?", key), value)
			}
		}
	}

	// Apply search
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("title ILIKE ? OR criticism ILIKE ? OR solution ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	if sort != "" {
		query = query.Order(sort)
	} else {
		query = query.Order("created_at DESC")
	}

	// Apply pagination & preload
	offset := (page - 1) * limit
	if err := query.
		Preload("User").
		Preload("Sector").
		Preload("Region").
		Offset(offset).
		Limit(limit).
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// FindByIDWithRelations finds a post by ID with all relations preloaded
func (r *postRepository) FindByIDWithRelations(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	var post models.Post
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Sector").
		Preload("Region").
		Preload("Reviewer").
		Where("id = ?", id).
		First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// GetByUserID retrieves posts by a specific user
func (r *postRepository) GetByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Post{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.
		Preload("Sector").
		Preload("Region").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// GetByStatus retrieves posts by status (for moderation queue)
func (r *postRepository) GetByStatus(ctx context.Context, status string, page, limit int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Post{}).Where("status = ?", status)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.
		Preload("User").
		Preload("Sector").
		Preload("Region").
		Order("created_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// IncrementVoteCount atomically updates the vote count
func (r *postRepository) IncrementVoteCount(ctx context.Context, postID uuid.UUID, delta int) error {
	return r.db.WithContext(ctx).
		Model(&models.Post{}).
		Where("id = ?", postID).
		UpdateColumn("vote_count", gorm.Expr("vote_count + ?", delta)).
		Error
}

// IncrementCommentCount atomically updates the comment count
func (r *postRepository) IncrementCommentCount(ctx context.Context, postID uuid.UUID, delta int) error {
	return r.db.WithContext(ctx).
		Model(&models.Post{}).
		Where("id = ?", postID).
		UpdateColumn("comment_count", gorm.Expr("comment_count + ?", delta)).
		Error
}

// GetByIDs retrieves posts by a list of IDs with relations preloaded
func (r *postRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Post, error) {
	if len(ids) == 0 {
		return []models.Post{}, nil
	}
	var posts []models.Post
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Sector").
		Preload("Region").
		Where("id IN ?", ids).
		Find(&posts).Error
	if err != nil {
		return nil, err
	}
	return posts, nil
}

// GetRelated retrieves related posts based on the same sector, excluding the current post.
// Falls back to latest approved posts if no sector match is found.
func (r *postRepository) GetRelated(ctx context.Context, postID uuid.UUID, sectorID *uuid.UUID, limit int) ([]models.Post, error) {
	var posts []models.Post

	query := r.db.WithContext(ctx).
		Where("id != ? AND status = ?", postID, "approved").
		Preload("User").
		Preload("Sector").
		Preload("Region")

	if sectorID != nil {
		// Primary: same sector, ordered by vote_count desc
		query = query.Where("sector_id = ?", *sectorID)
	}

	err := query.
		Order("vote_count DESC, created_at DESC").
		Limit(limit).
		Find(&posts).Error
	if err != nil {
		return nil, err
	}

	// Fallback: if sector-based query yields too few results, fill with latest approved posts
	if sectorID != nil && len(posts) < limit {
		existingIDs := make([]uuid.UUID, len(posts))
		for i, p := range posts {
			existingIDs[i] = p.ID
		}
		excludeIDs := append(existingIDs, postID)
		remaining := limit - len(posts)

		var fallbackPosts []models.Post
		err = r.db.WithContext(ctx).
			Where("id NOT IN ? AND status = ?", excludeIDs, "approved").
			Preload("User").
			Preload("Sector").
			Preload("Region").
			Order("vote_count DESC, created_at DESC").
			Limit(remaining).
			Find(&fallbackPosts).Error
		if err != nil {
			return posts, nil // Return what we have; don't fail on fallback
		}
		posts = append(posts, fallbackPosts...)
	}

	return posts, nil
}
