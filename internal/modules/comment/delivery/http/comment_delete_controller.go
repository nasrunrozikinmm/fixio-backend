package controllers

import (
	"fixio/internal/modules/comment/dto"
	services "fixio/internal/modules/comment/usecase"
	"fixio/pkg/network"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type commentDeleteController struct {
	network.BaseController
	service services.CommentService
}

// NewCommentDeleteController creates a controller for comment deletion (different path)
func NewCommentDeleteController(
	authFn network.AuthenticationProvider,
	authzFn network.AuthorizationProvider,
	service services.CommentService,
) network.Controller {
	_ = dto.CreateCommentRequest{} // ensure dto import
	return &commentDeleteController{
		BaseController: network.NewBaseController("/comments", authFn, authzFn),
		service:        service,
	}
}

// MountRoutes registers comment deletion routes
func (c *commentDeleteController) MountRoutes(rg *fiber.Group) {
	rg.Delete("/:id", c.Authentication(), c.DeleteComment)
}

// DeleteComment godoc
// @Summary     Hapus komentar
// @Description Menghapus komentar berdasarkan ID (pemilik, moderator, atau administrator)
// @Tags        Comments
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id path string true "Comment ID (UUID)"
// @Success     200 {object} network.Response "Komentar berhasil dihapus"
// @Failure     400 {object} network.ErrorResponse "ID tidak valid"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     403 {object} network.ErrorResponse "Forbidden"
// @Failure     404 {object} network.ErrorResponse "Komentar tidak ditemukan"
// @Router      /api/comments/{id} [delete]
func (c *commentDeleteController) DeleteComment(ctx *fiber.Ctx) error {
	commentID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	userIDStr, _ := ctx.Locals("userId").(string)
	userID, _ := uuid.Parse(userIDStr)
	userRole, _ := ctx.Locals("userRole").(string)

	if err := c.service.Delete(ctx.Context(), commentID, userID, userRole); err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessMsgResponse("Komentar berhasil dihapus")
}
