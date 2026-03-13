package repositories

import (
	"context"

	models "fixio/internal/modules/comment/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CommentVoteRepository defines comment vote data access operations
type CommentVoteRepository interface {
	FindByUserAndComment(ctx context.Context, userID, commentID uuid.UUID) (*models.CommentVote, error)
	Create(ctx context.Context, vote *models.CommentVote) error
	DeleteByUserAndComment(ctx context.Context, userID, commentID uuid.UUID) error
	IncrementVoteCount(ctx context.Context, commentID uuid.UUID, delta int) error
	GetUserVotes(ctx context.Context, userID uuid.UUID, commentIDs []uuid.UUID) ([]uuid.UUID, error)
}

type commentVoteRepository struct {
	db *gorm.DB
}

// NewCommentVoteRepository creates a new CommentVoteRepository
func NewCommentVoteRepository(db *gorm.DB) CommentVoteRepository {
	return &commentVoteRepository{db: db}
}

func (r *commentVoteRepository) FindByUserAndComment(ctx context.Context, userID, commentID uuid.UUID) (*models.CommentVote, error) {
	var vote models.CommentVote
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND comment_id = ?", userID, commentID).
		First(&vote).Error
	if err != nil {
		return nil, err
	}
	return &vote, nil
}

func (r *commentVoteRepository) Create(ctx context.Context, vote *models.CommentVote) error {
	return r.db.WithContext(ctx).Create(vote).Error
}

func (r *commentVoteRepository) DeleteByUserAndComment(ctx context.Context, userID, commentID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Where("user_id = ? AND comment_id = ?", userID, commentID).
		Delete(&models.CommentVote{}).Error
}

func (r *commentVoteRepository) IncrementVoteCount(ctx context.Context, commentID uuid.UUID, delta int) error {
	return r.db.WithContext(ctx).
		Model(&models.Comment{}).
		Where("id = ?", commentID).
		UpdateColumn("vote_count", gorm.Expr("vote_count + ?", delta)).
		Error
}

// GetUserVotes returns the comment IDs that a user has voted on (from a given set)
func (r *commentVoteRepository) GetUserVotes(ctx context.Context, userID uuid.UUID, commentIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(commentIDs) == 0 {
		return nil, nil
	}
	var votedIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&models.CommentVote{}).
		Where("user_id = ? AND comment_id IN (?)", userID, commentIDs).
		Pluck("comment_id", &votedIDs).Error
	return votedIDs, err
}
