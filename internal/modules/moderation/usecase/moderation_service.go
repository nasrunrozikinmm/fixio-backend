package services

import (
	"context"
	"math"

	userRepo "fixio/internal/modules/auth/repository"
	postModels "fixio/internal/modules/post/entity"
	postRepo "fixio/internal/modules/post/repository"
	"fixio/pkg/apperrors"
	"fixio/pkg/helpers"
	"fixio/pkg/network"

	"github.com/google/uuid"
)

// ModerationService defines content moderation business operations
type ModerationService interface {
	GetQueue(ctx context.Context, pagination network.Pagination) (*network.PaginatedResult[postModels.Post], error)
	ApprovePost(ctx context.Context, postID, reviewerID uuid.UUID, reviewNote string) (*postModels.Post, error)
	RejectPost(ctx context.Context, postID, reviewerID uuid.UUID, reviewNote string) (*postModels.Post, error)
	GetHistory(ctx context.Context, pagination network.Pagination) (*network.PaginatedResult[postModels.Post], error)
}

type moderationService struct {
	postRepo postRepo.PostRepository
	userRepo userRepo.UserRepository
}

// NewModerationService creates a new ModerationService
func NewModerationService(postRepo postRepo.PostRepository, userRepo userRepo.UserRepository) ModerationService {
	return &moderationService{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}

// GetQueue retrieves posts pending review
func (s *moderationService) GetQueue(ctx context.Context, pagination network.Pagination) (*network.PaginatedResult[postModels.Post], error) {
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.Limit < 1 || pagination.Limit > 100 {
		pagination.Limit = 20
	}

	posts, total, err := s.postRepo.GetByStatus(ctx, helpers.StatusPendingReview, pagination.Page, pagination.Limit)
	if err != nil {
		return nil, apperrors.NewInternal("Gagal mengambil antrian moderasi", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pagination.Limit)))

	return &network.PaginatedResult[postModels.Post]{
		Data: posts,
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

// ApprovePost approves a pending post
func (s *moderationService) ApprovePost(ctx context.Context, postID, reviewerID uuid.UUID, reviewNote string) (*postModels.Post, error) {
	post, err := s.postRepo.FindByIDWithRelations(ctx, postID)
	if err != nil {
		return nil, apperrors.NewNotFound("Post tidak ditemukan", err)
	}

	if post.Status != helpers.StatusPendingReview {
		return nil, apperrors.NewBadRequest("Post ini tidak dalam status pending review", nil)
	}

	// Update post status
	if err := s.postRepo.UpdateFields(ctx, postID.String(), map[string]any{
		"status":      helpers.StatusApproved,
		"reviewed_by": reviewerID,
		"review_note": reviewNote,
	}); err != nil {
		return nil, apperrors.NewInternal("Gagal approve post", err)
	}

	// Fetch updated post
	return s.postRepo.FindByIDWithRelations(ctx, postID)
}

// RejectPost rejects a pending post
func (s *moderationService) RejectPost(ctx context.Context, postID, reviewerID uuid.UUID, reviewNote string) (*postModels.Post, error) {
	post, err := s.postRepo.FindByIDWithRelations(ctx, postID)
	if err != nil {
		return nil, apperrors.NewNotFound("Post tidak ditemukan", err)
	}

	if post.Status != helpers.StatusPendingReview {
		return nil, apperrors.NewBadRequest("Post ini tidak dalam status pending review", nil)
	}

	// Update post status
	if err := s.postRepo.UpdateFields(ctx, postID.String(), map[string]any{
		"status":      helpers.StatusRejected,
		"reviewed_by": reviewerID,
		"review_note": reviewNote,
	}); err != nil {
		return nil, apperrors.NewInternal("Gagal reject post", err)
	}

	// Fetch updated post
	return s.postRepo.FindByIDWithRelations(ctx, postID)
}

// GetHistory retrieves reviewed posts (approved + rejected)
func (s *moderationService) GetHistory(ctx context.Context, pagination network.Pagination) (*network.PaginatedResult[postModels.Post], error) {
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.Limit < 1 || pagination.Limit > 100 {
		pagination.Limit = 20
	}

	// Get posts that have been reviewed (have a reviewed_by)
	filter := map[string]any{}
	posts, total, err := s.postRepo.GetWithRelations(ctx, pagination.Page, pagination.Limit, "updated_at DESC", filter, "")
	if err != nil {
		return nil, apperrors.NewInternal("Gagal mengambil riwayat moderasi", err)
	}

	// Filter to only reviewed posts
	var reviewedPosts []postModels.Post
	for _, p := range posts {
		if p.ReviewedBy != nil {
			reviewedPosts = append(reviewedPosts, p)
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(pagination.Limit)))

	return &network.PaginatedResult[postModels.Post]{
		Data: reviewedPosts,
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
