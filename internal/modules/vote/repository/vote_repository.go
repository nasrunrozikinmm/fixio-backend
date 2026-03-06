package repositories

import (
	"context"

	models "fixio/internal/modules/vote/entity"
	"fixio/pkg/data"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VoteRepository defines vote-specific data access operations
type VoteRepository interface {
	data.BaseRepository[models.Vote]
	FindByUserAndPost(ctx context.Context, userID, postID uuid.UUID) (*models.Vote, error)
	DeleteByUserAndPost(ctx context.Context, userID, postID uuid.UUID) error
	CountByPost(ctx context.Context, postID uuid.UUID) (upvotes int64, downvotes int64, err error)
}

type voteRepository struct {
	data.BaseRepository[models.Vote]
	db *gorm.DB
}

// NewVoteRepository creates a new VoteRepository
func NewVoteRepository(db *gorm.DB) VoteRepository {
	return &voteRepository{
		BaseRepository: data.NewBaseRepository[models.Vote](db),
		db:             db,
	}
}

// FindByUserAndPost finds a vote by user and post
func (r *voteRepository) FindByUserAndPost(ctx context.Context, userID, postID uuid.UUID) (*models.Vote, error) {
	var vote models.Vote
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND post_id = ?", userID, postID).
		First(&vote).Error
	if err != nil {
		return nil, err
	}
	return &vote, nil
}

// DeleteByUserAndPost removes a vote by user and post
func (r *voteRepository) DeleteByUserAndPost(ctx context.Context, userID, postID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Where("user_id = ? AND post_id = ?", userID, postID).
		Delete(&models.Vote{}).Error
}

// CountByPost counts upvotes and downvotes for a post
func (r *voteRepository) CountByPost(ctx context.Context, postID uuid.UUID) (int64, int64, error) {
	var upvotes, downvotes int64

	err := r.db.WithContext(ctx).Model(&models.Vote{}).
		Where("post_id = ? AND type = ?", postID, "up").
		Count(&upvotes).Error
	if err != nil {
		return 0, 0, err
	}

	err = r.db.WithContext(ctx).Model(&models.Vote{}).
		Where("post_id = ? AND type = ?", postID, "down").
		Count(&downvotes).Error
	if err != nil {
		return 0, 0, err
	}

	return upvotes, downvotes, nil
}
