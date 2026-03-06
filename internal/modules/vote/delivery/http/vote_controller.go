package controllers

import (
	"fixio/internal/modules/vote/dto"
	services "fixio/internal/modules/vote/usecase"
	"fixio/pkg/network"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type voteController struct {
	network.BaseController
	service services.VoteService
}

// NewVoteController creates a new vote controller
func NewVoteController(
	authFn network.AuthenticationProvider,
	authzFn network.AuthorizationProvider,
	service services.VoteService,
) network.Controller {
	return &voteController{
		BaseController: network.NewBaseController("/posts", authFn, authzFn),
		service:        service,
	}
}

// MountRoutes registers all vote routes
func (c *voteController) MountRoutes(rg *fiber.Group) {
	rg.Get("/:id/vote", c.Authentication(), c.GetUserVote)
	rg.Post("/:id/vote", c.Authentication(), c.Grant("vote:create"), c.Vote)
	rg.Delete("/:id/vote", c.Authentication(), c.Grant("vote:delete"), c.RemoveVote)
}

// GetUserVote godoc
// @Summary     Ambil vote user pada post
// @Description Mengambil vote user yang sedang login pada sebuah post. Mengembalikan null jika belum vote.
// @Tags        Votes
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id path string true "Post ID (UUID)"
// @Success     200 {object} network.Response "Vote user berhasil diambil"
// @Failure     400 {object} network.ErrorResponse "ID tidak valid"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Router      /api/posts/{id}/vote [get]
func (c *voteController) GetUserVote(ctx *fiber.Ctx) error {
	postID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	userIDStr, _ := ctx.Locals("userId").(string)
	userID, _ := uuid.Parse(userIDStr)

	vote, err := c.service.GetUserVote(ctx.Context(), userID, postID)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Vote user berhasil diambil", vote)
}

// Vote godoc
// @Summary     Vote pada post
// @Description Memberikan vote (up/down) pada sebuah post. Jika sudah ada vote dengan tipe yang sama, vote akan dihapus (toggle).
// @Tags        Votes
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string true "Post ID (UUID)"
// @Param       body body dto.VoteRequest true "Tipe vote (up/down)"
// @Success     200 {object} network.Response "Vote berhasil"
// @Failure     400 {object} network.ErrorResponse "ID atau body tidak valid"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     403 {object} network.ErrorResponse "Forbidden"
// @Router      /api/posts/{id}/vote [post]
func (c *voteController) Vote(ctx *fiber.Ctx) error {
	postID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	userIDStr, _ := ctx.Locals("userId").(string)
	userID, _ := uuid.Parse(userIDStr)

	req := new(dto.VoteRequest)
	if err := ctx.BodyParser(req); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidBody, err)
	}
	if result := network.Validate(req); !result.Valid {
		return c.Send(ctx).BadRequestError(result.FirstMessage(), nil)
	}

	vote, err := c.service.Vote(ctx.Context(), userID, postID, req.Type)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	if vote == nil {
		return c.Send(ctx).SuccessMsgResponse("Vote berhasil dihapus")
	}

	return c.Send(ctx).SuccessDataResponse("Vote berhasil", vote)
}

// RemoveVote godoc
// @Summary     Hapus vote dari post
// @Description Menghapus vote user dari sebuah post
// @Tags        Votes
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id path string true "Post ID (UUID)"
// @Success     200 {object} network.Response "Vote berhasil dihapus"
// @Failure     400 {object} network.ErrorResponse "ID tidak valid"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     403 {object} network.ErrorResponse "Forbidden"
// @Router      /api/posts/{id}/vote [delete]
func (c *voteController) RemoveVote(ctx *fiber.Ctx) error {
	postID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	userIDStr, _ := ctx.Locals("userId").(string)
	userID, _ := uuid.Parse(userIDStr)

	if err := c.service.RemoveVote(ctx.Context(), userID, postID); err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessMsgResponse("Vote berhasil dihapus")
}
