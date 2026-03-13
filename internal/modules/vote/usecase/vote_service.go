package services

import (
	"context"
	"fmt"

	authRepo "fixio/internal/modules/auth/repository"
	notifModels "fixio/internal/modules/notification/entity"
	notifService "fixio/internal/modules/notification/usecase"
	postRepo "fixio/internal/modules/post/repository"
	models "fixio/internal/modules/vote/entity"
	repositories "fixio/internal/modules/vote/repository"
	"fixio/pkg/apperrors"
	"fixio/pkg/helpers"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VoteService defines vote business operations
type VoteService interface {
	Vote(ctx context.Context, userID, postID uuid.UUID, voteType string) (*models.Vote, error)
	RemoveVote(ctx context.Context, userID, postID uuid.UUID) error
	GetUserVote(ctx context.Context, userID, postID uuid.UUID) (*models.Vote, error)
}

type voteService struct {
	voteRepo repositories.VoteRepository
	postRepo postRepo.PostRepository
	userRepo authRepo.UserRepository
	notifSvc notifService.NotificationService
}

// NewVoteService creates a new VoteService
func NewVoteService(voteRepo repositories.VoteRepository, postRepo postRepo.PostRepository, userRepo authRepo.UserRepository, notifSvc notifService.NotificationService) VoteService {
	return &voteService{
		voteRepo: voteRepo,
		postRepo: postRepo,
		userRepo: userRepo,
		notifSvc: notifSvc,
	}
}

// Vote creates or toggles a vote on a post
func (s *voteService) Vote(ctx context.Context, userID, postID uuid.UUID, voteType string) (*models.Vote, error) {
	// Check if post is approved
	post, err := s.postRepo.FindBy(ctx, map[string]any{"id": postID})
	if err != nil {
		return nil, apperrors.NewNotFound("Post tidak ditemukan", err)
	}
	if post.Status != helpers.StatusApproved {
		return nil, apperrors.NewBadRequest("Voting hanya bisa dilakukan pada post yang sudah approved", nil)
	}

	// Check existing vote
	existingVote, err := s.voteRepo.FindByUserAndPost(ctx, userID, postID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, apperrors.NewInternal("Gagal memeriksa vote", err)
	}

	if existingVote != nil {
		// Same vote type → remove vote
		if existingVote.Type == voteType {
			if err := s.voteRepo.DeleteByUserAndPost(ctx, userID, postID); err != nil {
				return nil, apperrors.NewInternal("Gagal menghapus vote", err)
			}
			// Update post vote count
			delta := -1
			if voteType == helpers.VoteDown {
				delta = 1
			}
			s.postRepo.IncrementVoteCount(ctx, postID, delta)
			return nil, nil // Vote removed
		}

		// Different vote type → update
		if err := s.voteRepo.DeleteByUserAndPost(ctx, userID, postID); err != nil {
			return nil, apperrors.NewInternal("Gagal mengupdate vote", err)
		}
		// Delta is +2 or -2 because we're switching from up to down or vice versa
		delta := 2
		if voteType == helpers.VoteDown {
			delta = -2
		}
		s.postRepo.IncrementVoteCount(ctx, postID, delta)
	} else {
		// New vote → update count
		delta := 1
		if voteType == helpers.VoteDown {
			delta = -1
		}
		s.postRepo.IncrementVoteCount(ctx, postID, delta)
	}

	// Create new vote
	vote := &models.Vote{
		UserID: userID,
		PostID: postID,
		Type:   voteType,
	}

	created, err := s.voteRepo.Create(ctx, vote)
	if err != nil {
		return nil, apperrors.NewInternal("Gagal membuat vote", err)
	}

	// Send notification to post author on upvote (not on toggle/switch, not on downvote)
	if existingVote == nil && voteType == helpers.VoteUp {
		voter, _ := s.userRepo.FindBy(ctx, map[string]any{"id": userID})
		actorName := "Seseorang"
		if voter != nil {
			actorName = voter.Name
		}
		_, _ = s.notifSvc.CreateNotification(
			ctx,
			post.UserID,
			notifModels.NotifTypePostVote,
			userID,
			postID,
			"post",
			fmt.Sprintf("%s menyukai post Anda", actorName),
		)
	}

	return created, nil
}

// RemoveVote removes a user's vote from a post
func (s *voteService) RemoveVote(ctx context.Context, userID, postID uuid.UUID) error {
	existing, err := s.voteRepo.FindByUserAndPost(ctx, userID, postID)
	if err != nil {
		return apperrors.NewNotFound("Vote tidak ditemukan", err)
	}

	if err := s.voteRepo.DeleteByUserAndPost(ctx, userID, postID); err != nil {
		return apperrors.NewInternal("Gagal menghapus vote", err)
	}

	// Update post vote count
	delta := -1
	if existing.Type == helpers.VoteDown {
		delta = 1
	}
	s.postRepo.IncrementVoteCount(ctx, postID, delta)

	return nil
}

// GetUserVote retrieves the user's vote on a post
func (s *voteService) GetUserVote(ctx context.Context, userID, postID uuid.UUID) (*models.Vote, error) {
	vote, err := s.voteRepo.FindByUserAndPost(ctx, userID, postID)
	if err != nil {
		return nil, nil // No vote found is not an error
	}
	return vote, nil
}
