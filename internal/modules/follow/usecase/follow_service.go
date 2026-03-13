package services

import (
	"context"
	"fmt"

	authRepo "fixio/internal/modules/auth/repository"
	models "fixio/internal/modules/follow/entity"
	repositories "fixio/internal/modules/follow/repository"
	notifModels "fixio/internal/modules/notification/entity"
	notifService "fixio/internal/modules/notification/usecase"
	"fixio/pkg/apperrors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FollowService defines follow business operations
type FollowService interface {
	Follow(ctx context.Context, followerID, followingID uuid.UUID) (*models.Follow, error)
	Unfollow(ctx context.Context, followerID, followingID uuid.UUID) error
	IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error)
	GetFollowerCount(ctx context.Context, userID uuid.UUID) (int64, error)
	GetFollowingCount(ctx context.Context, userID uuid.UUID) (int64, error)
	GetFollowingIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

type followService struct {
	followRepo repositories.FollowRepository
	userRepo   authRepo.UserRepository
	notifSvc   notifService.NotificationService
}

// NewFollowService creates a new FollowService
func NewFollowService(followRepo repositories.FollowRepository, userRepo authRepo.UserRepository, notifSvc notifService.NotificationService) FollowService {
	return &followService{
		followRepo: followRepo,
		userRepo:   userRepo,
		notifSvc:   notifSvc,
	}
}

// Follow creates a follow relationship
func (s *followService) Follow(ctx context.Context, followerID, followingID uuid.UUID) (*models.Follow, error) {
	// Cannot follow yourself
	if followerID == followingID {
		return nil, apperrors.NewBadRequest("Tidak bisa follow diri sendiri", nil)
	}

	// Check target user exists
	_, err := s.userRepo.FindBy(ctx, map[string]any{"id": followingID})
	if err != nil {
		return nil, apperrors.NewNotFound("User tidak ditemukan", err)
	}

	// Check if already following
	existing, err := s.followRepo.FindByFollowerAndFollowing(ctx, followerID, followingID)
	if err == nil && existing != nil {
		return nil, apperrors.NewBadRequest("Sudah mengikuti user ini", nil)
	}

	follow := &models.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}

	created, err := s.followRepo.Create(ctx, follow)
	if err != nil {
		return nil, apperrors.NewInternal("Gagal follow user", err)
	}

	// Send notification to the followed user
	follower, _ := s.userRepo.FindBy(ctx, map[string]any{"id": followerID})
	actorName := "Seseorang"
	if follower != nil {
		actorName = follower.Name
	}
	_, _ = s.notifSvc.CreateNotification(
		ctx,
		followingID,
		notifModels.NotifTypeNewFollower,
		followerID,
		followerID,
		"user",
		fmt.Sprintf("%s mulai mengikuti Anda", actorName),
	)

	return created, nil
}

// Unfollow removes a follow relationship
func (s *followService) Unfollow(ctx context.Context, followerID, followingID uuid.UUID) error {
	_, err := s.followRepo.FindByFollowerAndFollowing(ctx, followerID, followingID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperrors.NewNotFound("Follow tidak ditemukan", err)
		}
		return apperrors.NewInternal("Gagal memeriksa follow", err)
	}

	if err := s.followRepo.DeleteByFollowerAndFollowing(ctx, followerID, followingID); err != nil {
		return apperrors.NewInternal("Gagal unfollow user", err)
	}

	return nil
}

// IsFollowing checks if follower follows following
func (s *followService) IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	_, err := s.followRepo.FindByFollowerAndFollowing(ctx, followerID, followingID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, apperrors.NewInternal("Gagal memeriksa follow status", err)
	}
	return true, nil
}

// GetFollowerCount returns the number of followers for a user
func (s *followService) GetFollowerCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := s.followRepo.CountFollowers(ctx, userID)
	if err != nil {
		return 0, apperrors.NewInternal("Gagal menghitung followers", err)
	}
	return count, nil
}

// GetFollowingCount returns the number of users a user is following
func (s *followService) GetFollowingCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := s.followRepo.CountFollowing(ctx, userID)
	if err != nil {
		return 0, apperrors.NewInternal("Gagal menghitung following", err)
	}
	return count, nil
}

// GetFollowingIDs returns the list of user IDs the user is following
func (s *followService) GetFollowingIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	ids, err := s.followRepo.GetFollowingIDs(ctx, userID)
	if err != nil {
		return nil, apperrors.NewInternal("Gagal mengambil following list", err)
	}
	return ids, nil
}
