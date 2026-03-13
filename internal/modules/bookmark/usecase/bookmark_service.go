package services

import (
	"context"
	models "fixio/internal/modules/bookmark/entity"
	"fixio/internal/modules/bookmark/repository"
	postModels "fixio/internal/modules/post/entity"
	postRepo "fixio/internal/modules/post/repository"
	"fixio/pkg/apperrors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BookmarkService defines the business logic interface for bookmarks
type BookmarkService interface {
	AddBookmark(ctx context.Context, userID, postID uuid.UUID) (*models.Bookmark, error)
	RemoveBookmark(ctx context.Context, userID, postID uuid.UUID) error
	IsBookmarked(ctx context.Context, userID, postID uuid.UUID) (bool, error)
	GetUserBookmarkedPosts(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]postModels.Post, int64, error)
}

type bookmarkService struct {
	bookmarkRepo repository.BookmarkRepository
	postRepo     postRepo.PostRepository
}

// NewBookmarkService creates a new BookmarkService
func NewBookmarkService(
	bookmarkRepo repository.BookmarkRepository,
	postRepo postRepo.PostRepository,
) BookmarkService {
	return &bookmarkService{
		bookmarkRepo: bookmarkRepo,
		postRepo:     postRepo,
	}
}

func (s *bookmarkService) AddBookmark(ctx context.Context, userID, postID uuid.UUID) (*models.Bookmark, error) {
	// Validate post exists
	_, err := s.postRepo.FindBy(ctx, map[string]any{"id": postID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewNotFound("Post tidak ditemukan", err)
		}
		return nil, apperrors.NewInternal("Gagal memeriksa post", err)
	}

	// Check if already bookmarked
	existing, err := s.bookmarkRepo.FindByUserAndPost(ctx, userID, postID)
	if err == nil && existing != nil {
		return nil, apperrors.NewBadRequest("Post sudah di-bookmark", nil)
	}

	bookmark := &models.Bookmark{
		UserID: userID,
		PostID: postID,
	}

	created, err := s.bookmarkRepo.Create(ctx, bookmark)
	if err != nil {
		return nil, apperrors.NewInternal("Gagal menambahkan bookmark", err)
	}

	return created, nil
}

func (s *bookmarkService) RemoveBookmark(ctx context.Context, userID, postID uuid.UUID) error {
	// Verify the bookmark exists
	_, err := s.bookmarkRepo.FindByUserAndPost(ctx, userID, postID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperrors.NewNotFound("Bookmark tidak ditemukan", err)
		}
		return apperrors.NewInternal("Gagal memeriksa bookmark", err)
	}

	if err := s.bookmarkRepo.DeleteByUserAndPost(ctx, userID, postID); err != nil {
		return apperrors.NewInternal("Gagal menghapus bookmark", err)
	}

	return nil
}

func (s *bookmarkService) IsBookmarked(ctx context.Context, userID, postID uuid.UUID) (bool, error) {
	return s.bookmarkRepo.IsBookmarked(ctx, userID, postID)
}

func (s *bookmarkService) GetUserBookmarkedPosts(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]postModels.Post, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	postIDs, total, err := s.bookmarkRepo.GetUserBookmarkPostIDs(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, apperrors.NewInternal("Gagal mengambil bookmark", err)
	}

	if len(postIDs) == 0 {
		return []postModels.Post{}, total, nil
	}

	posts, err := s.postRepo.GetByIDs(ctx, postIDs)
	if err != nil {
		return nil, 0, apperrors.NewInternal("Gagal mengambil post bookmark", err)
	}

	return posts, total, nil
}
