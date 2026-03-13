package repositories

import (
	"context"

	models "fixio/internal/modules/follow/entity"
	"fixio/pkg/data"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FollowRepository defines follow-specific data access operations
type FollowRepository interface {
	data.BaseRepository[models.Follow]
	FindByFollowerAndFollowing(ctx context.Context, followerID, followingID uuid.UUID) (*models.Follow, error)
	DeleteByFollowerAndFollowing(ctx context.Context, followerID, followingID uuid.UUID) error
	CountFollowers(ctx context.Context, userID uuid.UUID) (int64, error)
	CountFollowing(ctx context.Context, userID uuid.UUID) (int64, error)
	GetFollowerIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	GetFollowingIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

type followRepository struct {
	data.BaseRepository[models.Follow]
	db *gorm.DB
}

// NewFollowRepository creates a new FollowRepository
func NewFollowRepository(db *gorm.DB) FollowRepository {
	return &followRepository{
		BaseRepository: data.NewBaseRepository[models.Follow](db),
		db:             db,
	}
}

func (r *followRepository) FindByFollowerAndFollowing(ctx context.Context, followerID, followingID uuid.UUID) (*models.Follow, error) {
	var follow models.Follow
	err := r.db.WithContext(ctx).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		First(&follow).Error
	if err != nil {
		return nil, err
	}
	return &follow, nil
}

func (r *followRepository) DeleteByFollowerAndFollowing(ctx context.Context, followerID, followingID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Delete(&models.Follow{}).Error
}

func (r *followRepository) CountFollowers(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Follow{}).
		Where("following_id = ?", userID).
		Count(&count).Error
	return count, err
}

func (r *followRepository) CountFollowing(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Follow{}).
		Where("follower_id = ?", userID).
		Count(&count).Error
	return count, err
}

func (r *followRepository) GetFollowerIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Model(&models.Follow{}).
		Where("following_id = ?", userID).
		Pluck("follower_id", &ids).Error
	return ids, err
}

func (r *followRepository) GetFollowingIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Model(&models.Follow{}).
		Where("follower_id = ?", userID).
		Pluck("following_id", &ids).Error
	return ids, err
}
