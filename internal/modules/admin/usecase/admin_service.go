package services

import (
	"context"
	"math"

	"fixio/internal/modules/admin/dto"
	userModels "fixio/internal/modules/auth/entity"
	userRepo "fixio/internal/modules/auth/repository"
	commentModels "fixio/internal/modules/comment/entity"
	postModels "fixio/internal/modules/post/entity"
	postRepo "fixio/internal/modules/post/repository"
	voteModels "fixio/internal/modules/vote/entity"
	"fixio/pkg/apperrors"
	"fixio/pkg/helpers"
	"fixio/pkg/network"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdminService defines admin business operations
type AdminService interface {
	GetAllUsers(ctx context.Context, pagination network.Pagination) (*network.PaginatedResult[userModels.User], error)
	ChangeUserRole(ctx context.Context, targetUserID, adminUserID uuid.UUID, newRole string) (*userModels.User, error)
	GetStats(ctx context.Context) (*dto.StatsResponse, error)
}

type adminService struct {
	userRepo userRepo.UserRepository
	postRepo postRepo.PostRepository
	db       *gorm.DB
}

// NewAdminService creates a new AdminService
func NewAdminService(userRepo userRepo.UserRepository, postRepo postRepo.PostRepository, db *gorm.DB) AdminService {
	return &adminService{
		userRepo: userRepo,
		postRepo: postRepo,
		db:       db,
	}
}

// GetAllUsers retrieves all users with pagination
func (s *adminService) GetAllUsers(ctx context.Context, pagination network.Pagination) (*network.PaginatedResult[userModels.User], error) {
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.Limit < 1 || pagination.Limit > 100 {
		pagination.Limit = 20
	}

	users, total, err := s.userRepo.Get(ctx, pagination.Page, pagination.Limit, pagination.Sort, nil)
	if err != nil {
		return nil, apperrors.NewInternal("Gagal mengambil data user", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pagination.Limit)))

	return &network.PaginatedResult[userModels.User]{
		Data: users,
		Pagination: network.PaginationMeta{
			Page:       pagination.Page,
			Limit:      pagination.Limit,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    pagination.Page < totalPages,
			HasPrev:    pagination.Page > 1,
		},
	}, nil
}

// ChangeUserRole changes a user's role
func (s *adminService) ChangeUserRole(ctx context.Context, targetUserID, adminUserID uuid.UUID, newRole string) (*userModels.User, error) {
	if targetUserID == adminUserID {
		return nil, apperrors.NewBadRequest("Tidak bisa mengubah role sendiri", nil)
	}

	if !helpers.IsValidRole(newRole) {
		return nil, apperrors.NewBadRequest("Role tidak valid. Pilih: creator, moderator, atau administrator", nil)
	}

	user, err := s.userRepo.FindBy(ctx, map[string]any{"id": targetUserID})
	if err != nil {
		return nil, apperrors.NewNotFound("User tidak ditemukan", err)
	}

	if err := s.userRepo.UpdateFields(ctx, targetUserID.String(), map[string]any{
		"role": newRole,
	}); err != nil {
		return nil, apperrors.NewInternal("Gagal mengubah role user", err)
	}

	user.Role = newRole
	return user, nil
}

// GetStats retrieves platform statistics
func (s *adminService) GetStats(ctx context.Context) (*dto.StatsResponse, error) {
	var stats dto.StatsResponse

	// Count users
	s.db.WithContext(ctx).Model(&userModels.User{}).Count(&stats.TotalUsers)

	// Count posts by status
	s.db.WithContext(ctx).Model(&postModels.Post{}).Count(&stats.TotalPosts)
	s.db.WithContext(ctx).Model(&postModels.Post{}).Where("status = ?", helpers.StatusApproved).Count(&stats.TotalApproved)
	s.db.WithContext(ctx).Model(&postModels.Post{}).Where("status = ?", helpers.StatusPendingReview).Count(&stats.TotalPendingReview)
	s.db.WithContext(ctx).Model(&postModels.Post{}).Where("status = ?", helpers.StatusRejected).Count(&stats.TotalRejected)

	// Count comments & votes
	s.db.WithContext(ctx).Model(&commentModels.Comment{}).Count(&stats.TotalComments)
	s.db.WithContext(ctx).Model(&voteModels.Vote{}).Count(&stats.TotalVotes)

	return &stats, nil
}
