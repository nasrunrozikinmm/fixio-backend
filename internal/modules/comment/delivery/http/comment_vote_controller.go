package controllers

import (
	services "fixio/internal/modules/comment/usecase"
	"fixio/pkg/network"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type commentVoteController struct {
	network.BaseController
	service services.CommentService
}

// NewCommentVoteController creates a new controller for comment voting
func NewCommentVoteController(
	authFn network.AuthenticationProvider,
	authzFn network.AuthorizationProvider,
	service services.CommentService,
) network.Controller {
	return &commentVoteController{
		BaseController: network.NewBaseController("/comments", authFn, authzFn),
		service:        service,
	}
}

// MountRoutes registers comment vote routes
func (c *commentVoteController) MountRoutes(rg *fiber.Group) {
	rg.Post("/:id/vote", c.Authentication(), c.Grant("comment:create"), c.ToggleVote)
	rg.Get("/:id/votes/me", c.Authentication(), c.GetMyVote)
}

// ToggleVote godoc
// @Summary     Toggle upvote komentar
// @Description Toggle upvote pada komentar. Jika sudah vote → hapus, belum → tambahkan.
// @Tags        Comments
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id path string true "Comment ID (UUID)"
// @Success     200 {object} network.Response "Vote berhasil di-toggle"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     404 {object} network.ErrorResponse "Komentar tidak ditemukan"
// @Router      /api/comments/{id}/vote [post]
func (c *commentVoteController) ToggleVote(ctx *fiber.Ctx) error {
	commentID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	userIDStr, _ := ctx.Locals("userId").(string)
	userID, _ := uuid.Parse(userIDStr)

	voted, newCount, err := c.service.ToggleVote(ctx.Context(), userID, commentID)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	msg := "Vote berhasil ditambahkan"
	if !voted {
		msg = "Vote berhasil dihapus"
	}

	return c.Send(ctx).SuccessDataResponse(msg, fiber.Map{
		"voted":      voted,
		"vote_count": newCount,
	})
}

// GetMyVote godoc
// @Summary     Cek vote user pada komentar
// @Description Mengambil status vote user terhadap suatu komentar
// @Tags        Comments
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Comment ID (UUID)"
// @Success     200 {object} network.Response "Status vote berhasil diambil"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Router      /api/comments/{id}/votes/me [get]
func (c *commentVoteController) GetMyVote(ctx *fiber.Ctx) error {
	commentID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	userIDStr, _ := ctx.Locals("userId").(string)
	userID, _ := uuid.Parse(userIDStr)

	votedIDs, err := c.service.GetUserVotes(ctx.Context(), userID, []uuid.UUID{commentID})
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	voted := false
	for _, id := range votedIDs {
		if id == commentID {
			voted = true
			break
		}
	}

	return c.Send(ctx).SuccessDataResponse("Status vote berhasil diambil", fiber.Map{
		"voted": voted,
	})
}
