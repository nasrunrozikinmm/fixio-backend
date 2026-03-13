package services

import (
	"context"
	"math"

	"fixio/internal/modules/post/dto"
	models "fixio/internal/modules/post/entity"
	repositories "fixio/internal/modules/post/repository"
	"fixio/pkg/apperrors"
	"fixio/pkg/helpers"
	"fixio/pkg/network"
	"fixio/pkg/sanitizer"

	"github.com/google/uuid"
)

// PostService defines post business operations
type PostService interface {
	GetAll(ctx context.Context, pagination network.Pagination, filter *dto.PostFilter) (*network.PaginatedResult[models.Post], error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Post, error)
	Create(ctx context.Context, userID uuid.UUID, req *dto.CreatePostRequest) (*models.Post, error)
	Update(ctx context.Context, id, userID uuid.UUID, req *dto.UpdatePostRequest) (*models.Post, error)
	Delete(ctx context.Context, id, userID uuid.UUID, userRole string) error
	GetByUserID(ctx context.Context, userID uuid.UUID, pagination network.Pagination) (*network.PaginatedResult[models.Post], error)
	GetRelated(ctx context.Context, id uuid.UUID, limit int) ([]models.Post, error)
}

type postService struct {
	repo repositories.PostRepository
}

// NewPostService creates a new PostService
func NewPostService(repo repositories.PostRepository) PostService {
	return &postService{repo: repo}
}

// GetAll retrieves approved posts with filters
func (s *postService) GetAll(ctx context.Context, pagination network.Pagination, filter *dto.PostFilter) (*network.PaginatedResult[models.Post], error) {
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.Limit < 1 || pagination.Limit > 100 {
		pagination.Limit = 20
	}

	filterMap := map[string]any{
		"status": helpers.StatusApproved, // Only show approved posts publicly
	}

	search := ""
	if filter != nil {
		if filter.SectorID != nil {
			filterMap["sector_id"] = *filter.SectorID
		}
		if filter.RegionID != nil {
			filterMap["region_id"] = *filter.RegionID
		}
		if filter.Status != "" {
			filterMap["status"] = filter.Status
		}
		if filter.UserID != nil {
			filterMap["user_id"] = *filter.UserID
		}
		if len(filter.UserIDs) > 0 {
			filterMap["user_ids"] = filter.UserIDs
		}
		search = filter.Search
	}

	posts, total, err := s.repo.GetWithRelations(ctx, pagination.Page, pagination.Limit, pagination.Sort, filterMap, search)
	if err != nil {
		return nil, apperrors.NewInternal("Gagal mengambil data post", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pagination.Limit)))

	return &network.PaginatedResult[models.Post]{
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

// GetByID retrieves a single post by ID
func (s *postService) GetByID(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	post, err := s.repo.FindByIDWithRelations(ctx, id)
	if err != nil {
		return nil, apperrors.NewNotFound("Post tidak ditemukan", err)
	}
	return post, nil
}

// Create creates a new post
func (s *postService) Create(ctx context.Context, userID uuid.UUID, req *dto.CreatePostRequest) (*models.Post, error) {
	status := helpers.StatusPendingReview
	if req.Status == helpers.StatusDraft {
		status = helpers.StatusDraft
	}

	post := &models.Post{
		UserID:         userID,
		Title:          req.Title,
		SectorID:       req.SectorID,
		RegionID:       req.RegionID,
		Criticism:      sanitizer.SanitizeHTML(req.Criticism),
		Solution:       sanitizer.SanitizeHTML(req.Solution),
		ImpactEstimate: req.ImpactEstimate,
		References:     req.References,
		Images:         models.StringArray(req.Images),
		Status:         status,
	}

	created, err := s.repo.Create(ctx, post)
	if err != nil {
		return nil, apperrors.NewInternal("Gagal membuat post", err)
	}

	return created, nil
}

// Update updates an existing post (owner only)
func (s *postService) Update(ctx context.Context, id, userID uuid.UUID, req *dto.UpdatePostRequest) (*models.Post, error) {
	post, err := s.repo.FindByIDWithRelations(ctx, id)
	if err != nil {
		return nil, apperrors.NewNotFound("Post tidak ditemukan", err)
	}

	if post.UserID != userID {
		return nil, apperrors.NewForbidden("Anda tidak memiliki akses untuk mengedit post ini", nil)
	}

	// Update fields
	if req.Title != "" {
		post.Title = req.Title
	}
	if req.SectorID != nil {
		post.SectorID = req.SectorID
	}
	if req.RegionID != nil {
		post.RegionID = req.RegionID
	}
	if req.Criticism != "" {
		post.Criticism = sanitizer.SanitizeHTML(req.Criticism)
	}
	if req.Solution != "" {
		post.Solution = sanitizer.SanitizeHTML(req.Solution)
	}
	if req.ImpactEstimate != "" {
		post.ImpactEstimate = req.ImpactEstimate
	}
	if req.References != "" {
		post.References = req.References
	}
	if req.Images != nil {
		post.Images = models.StringArray(req.Images)
	}

	// Reset status to pending_review when edited
	post.Status = helpers.StatusPendingReview
	post.ReviewedBy = nil
	post.ReviewNote = ""

	updated, err := s.repo.Update(ctx, id.String(), post)
	if err != nil {
		return nil, apperrors.NewInternal("Gagal mengupdate post", err)
	}

	return updated, nil
}

// Delete deletes a post (owner or admin only)
func (s *postService) Delete(ctx context.Context, id, userID uuid.UUID, userRole string) error {
	post, err := s.repo.FindByIDWithRelations(ctx, id)
	if err != nil {
		return apperrors.NewNotFound("Post tidak ditemukan", err)
	}

	if post.UserID != userID && !helpers.IsAdmin(userRole) {
		return apperrors.NewForbidden("Anda tidak memiliki akses untuk menghapus post ini", nil)
	}

	if err := s.repo.SoftDelete(ctx, id.String()); err != nil {
		return apperrors.NewInternal("Gagal menghapus post", err)
	}

	return nil
}

// GetByUserID retrieves posts by a specific user
func (s *postService) GetByUserID(ctx context.Context, userID uuid.UUID, pagination network.Pagination) (*network.PaginatedResult[models.Post], error) {
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.Limit < 1 || pagination.Limit > 100 {
		pagination.Limit = 20
	}

	posts, total, err := s.repo.GetByUserID(ctx, userID, pagination.Page, pagination.Limit)
	if err != nil {
		return nil, apperrors.NewInternal("Gagal mengambil data post", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pagination.Limit)))

	return &network.PaginatedResult[models.Post]{
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

// GetRelated retrieves related posts for a given post (same sector, fallback to popular)
func (s *postService) GetRelated(ctx context.Context, id uuid.UUID, limit int) ([]models.Post, error) {
	if limit < 1 || limit > 10 {
		limit = 5
	}

	// Get the source post to find its sector
	post, err := s.repo.FindByIDWithRelations(ctx, id)
	if err != nil {
		return nil, apperrors.NewNotFound("Post tidak ditemukan", err)
	}

	related, err := s.repo.GetRelated(ctx, id, post.SectorID, limit)
	if err != nil {
		return nil, apperrors.NewInternal("Gagal mengambil post terkait", err)
	}

	return related, nil
}
