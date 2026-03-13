package services

import (
	"context"
	"math"

	models "fixio/internal/modules/comment/entity"
	repositories "fixio/internal/modules/comment/repository"
	postRepo "fixio/internal/modules/post/repository"
	"fixio/pkg/apperrors"
	"fixio/pkg/helpers"
	"fixio/pkg/network"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CommentService defines comment business operations
type CommentService interface {
	GetByPostID(ctx context.Context, postID uuid.UUID, pagination network.Pagination) (*network.PaginatedResult[models.Comment], error)
	Create(ctx context.Context, userID, postID uuid.UUID, content string, parentID *uuid.UUID) (*models.Comment, error)
	Delete(ctx context.Context, commentID, userID uuid.UUID, userRole string) error
	ToggleVote(ctx context.Context, userID, commentID uuid.UUID) (voted bool, newCount int, err error)
	GetUserVotes(ctx context.Context, userID uuid.UUID, commentIDs []uuid.UUID) ([]uuid.UUID, error)
}

type commentService struct {
	commentRepo     repositories.CommentRepository
	commentVoteRepo repositories.CommentVoteRepository
	postRepo        postRepo.PostRepository
}

// NewCommentService creates a new CommentService
func NewCommentService(commentRepo repositories.CommentRepository, commentVoteRepo repositories.CommentVoteRepository, postRepo postRepo.PostRepository) CommentService {
	return &commentService{
		commentRepo:     commentRepo,
		commentVoteRepo: commentVoteRepo,
		postRepo:        postRepo,
	}
}

// GetByPostID retrieves comments for a post
func (s *commentService) GetByPostID(ctx context.Context, postID uuid.UUID, pagination network.Pagination) (*network.PaginatedResult[models.Comment], error) {
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.Limit < 1 || pagination.Limit > 100 {
		pagination.Limit = 20
	}

	// Check if post is approved
	post, err := s.postRepo.FindBy(ctx, map[string]any{"id": postID})
	if err != nil {
		return nil, apperrors.NewNotFound("Post tidak ditemukan", err)
	}
	if post.Status != helpers.StatusApproved {
		return nil, apperrors.NewBadRequest("Komentar hanya tersedia untuk post yang sudah approved", nil)
	}

	comments, total, err := s.commentRepo.GetByPostID(ctx, postID, pagination.Page, pagination.Limit)
	if err != nil {
		return nil, apperrors.NewInternal("Gagal mengambil komentar", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pagination.Limit)))

	return &network.PaginatedResult[models.Comment]{
		Data: comments,
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

// Create creates a new comment on a post
func (s *commentService) Create(ctx context.Context, userID, postID uuid.UUID, content string, parentID *uuid.UUID) (*models.Comment, error) {
	// Check if post is approved
	post, err := s.postRepo.FindBy(ctx, map[string]any{"id": postID})
	if err != nil {
		return nil, apperrors.NewNotFound("Post tidak ditemukan", err)
	}
	if post.Status != helpers.StatusApproved {
		return nil, apperrors.NewBadRequest("Komentar hanya bisa ditambahkan pada post yang sudah approved", nil)
	}

	// Check if it's a nested reply (only 1 level allowed)
	if parentID != nil {
		parent, err := s.commentRepo.FindByIDWithUser(ctx, *parentID)
		if err != nil {
			return nil, apperrors.NewNotFound("Komentar parent tidak ditemukan", err)
		}
		// Prevent nested replies beyond 1 level
		if parent.ParentID != nil {
			return nil, apperrors.NewBadRequest("Reply bersarang lebih dari 1 level tidak diperbolehkan", nil)
		}
	}

	comment := &models.Comment{
		UserID:   userID,
		PostID:   postID,
		ParentID: parentID,
		Content:  content,
	}

	created, err := s.commentRepo.Create(ctx, comment)
	if err != nil {
		return nil, apperrors.NewInternal("Gagal membuat komentar", err)
	}

	// Increment post comment count
	s.postRepo.IncrementCommentCount(ctx, postID, 1)

	return created, nil
}

// Delete deletes a comment (owner or admin only)
func (s *commentService) Delete(ctx context.Context, commentID, userID uuid.UUID, userRole string) error {
	comment, err := s.commentRepo.FindByIDWithUser(ctx, commentID)
	if err != nil {
		return apperrors.NewNotFound("Komentar tidak ditemukan", err)
	}

	if comment.UserID != userID && !helpers.IsAdmin(userRole) {
		return apperrors.NewForbidden("Anda tidak memiliki akses untuk menghapus komentar ini", nil)
	}

	if err := s.commentRepo.SoftDelete(ctx, commentID.String()); err != nil {
		return apperrors.NewInternal("Gagal menghapus komentar", err)
	}

	// Decrement post comment count
	s.postRepo.IncrementCommentCount(ctx, comment.PostID, -1)

	return nil
}

// ToggleVote toggles a user's upvote on a comment. Returns new voted state and updated vote count.
func (s *commentService) ToggleVote(ctx context.Context, userID, commentID uuid.UUID) (bool, int, error) {
	// Check comment exists
	comment, err := s.commentRepo.FindByIDWithUser(ctx, commentID)
	if err != nil {
		return false, 0, apperrors.NewNotFound("Komentar tidak ditemukan", err)
	}

	existing, err := s.commentVoteRepo.FindByUserAndComment(ctx, userID, commentID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, 0, apperrors.NewInternal("Gagal memeriksa vote", err)
	}

	voted := false
	if existing != nil {
		// Already voted → remove
		if err := s.commentVoteRepo.DeleteByUserAndComment(ctx, userID, commentID); err != nil {
			return false, 0, apperrors.NewInternal("Gagal menghapus vote", err)
		}
		_ = s.commentVoteRepo.IncrementVoteCount(ctx, commentID, -1)
		voted = false
	} else {
		// Not voted → add
		vote := &models.CommentVote{
			UserID:    userID,
			CommentID: commentID,
		}
		if err := s.commentVoteRepo.Create(ctx, vote); err != nil {
			return false, 0, apperrors.NewInternal("Gagal menyimpan vote", err)
		}
		_ = s.commentVoteRepo.IncrementVoteCount(ctx, commentID, 1)
		voted = true
	}

	// Reload to get fresh vote_count
	updated, err := s.commentRepo.FindByIDWithUser(ctx, commentID)
	if err != nil {
		return voted, comment.VoteCount, nil // fallback to stale count
	}
	return voted, updated.VoteCount, nil
}

// GetUserVotes returns comment IDs from the given set that the user has voted on
func (s *commentService) GetUserVotes(ctx context.Context, userID uuid.UUID, commentIDs []uuid.UUID) ([]uuid.UUID, error) {
	return s.commentVoteRepo.GetUserVotes(ctx, userID, commentIDs)
}
