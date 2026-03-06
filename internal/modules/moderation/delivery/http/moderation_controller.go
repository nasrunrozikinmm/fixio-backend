package controllers

import (
	"fixio/internal/modules/moderation/dto"
	services "fixio/internal/modules/moderation/usecase"
	"fixio/pkg/network"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type moderationController struct {
	network.BaseController
	service services.ModerationService
}

// NewModerationController creates a new moderation controller
func NewModerationController(
	authFn network.AuthenticationProvider,
	authzFn network.AuthorizationProvider,
	service services.ModerationService,
) network.Controller {
	return &moderationController{
		BaseController: network.NewBaseController("/moderation", authFn, authzFn),
		service:        service,
	}
}

// MountRoutes registers all moderation routes
func (c *moderationController) MountRoutes(rg *fiber.Group) {
	rg.Use(c.Authentication())
	rg.Get("/queue", c.Grant("moderation:view_queue"), c.GetQueue)
	rg.Put("/posts/:id/approve", c.Grant("moderation:approve"), c.ApprovePost)
	rg.Put("/posts/:id/reject", c.Grant("moderation:reject"), c.RejectPost)
	rg.Get("/history", c.Grant("moderation:view_history"), c.GetHistory)
}

// GetQueue godoc
// @Summary     Antrian moderasi
// @Description Mengambil daftar post yang menunggu review (khusus Moderator/Admin)
// @Tags        Moderation
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       page  query int false "Nomor halaman" default(1)
// @Param       limit query int false "Jumlah item per halaman" default(20)
// @Success     200 {object} network.Response "Antrian moderasi berhasil diambil"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     403 {object} network.ErrorResponse "Forbidden"
// @Router      /api/moderation/queue [get]
func (c *moderationController) GetQueue(ctx *fiber.Ctx) error {
	pagination := network.DefaultPagination()
	if err := ctx.QueryParser(&pagination); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidPagination, err)
	}

	result, err := c.service.GetQueue(ctx.Context(), pagination)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Antrian moderasi berhasil diambil", result)
}

// ApprovePost godoc
// @Summary     Approve post
// @Description Menyetujui post yang sedang menunggu review
// @Tags        Moderation
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string true "Post ID (UUID)"
// @Param       body body dto.ReviewPostRequest false "Catatan review (opsional)"
// @Success     200 {object} network.Response "Post berhasil diapprove"
// @Failure     400 {object} network.ErrorResponse "ID tidak valid"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     403 {object} network.ErrorResponse "Forbidden"
// @Router      /api/moderation/posts/{id}/approve [put]
func (c *moderationController) ApprovePost(ctx *fiber.Ctx) error {
	postID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	reviewerIDStr, _ := ctx.Locals("userId").(string)
	reviewerID, _ := uuid.Parse(reviewerIDStr)

	req := new(dto.ReviewPostRequest)
	ctx.BodyParser(req) // Optional body

	post, err := c.service.ApprovePost(ctx.Context(), postID, reviewerID, req.ReviewNote)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Post berhasil diapprove", post)
}

// RejectPost godoc
// @Summary     Tolak post
// @Description Menolak post yang sedang menunggu review (alasan wajib diisi)
// @Tags        Moderation
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string true "Post ID (UUID)"
// @Param       body body dto.RejectPostRequest true "Alasan penolakan"
// @Success     200 {object} network.Response "Post berhasil ditolak"
// @Failure     400 {object} network.ErrorResponse "Request tidak valid"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     403 {object} network.ErrorResponse "Forbidden"
// @Router      /api/moderation/posts/{id}/reject [put]
func (c *moderationController) RejectPost(ctx *fiber.Ctx) error {
	postID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	reviewerIDStr, _ := ctx.Locals("userId").(string)
	reviewerID, _ := uuid.Parse(reviewerIDStr)

	req := new(dto.RejectPostRequest)
	if err := ctx.BodyParser(req); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidBody, err)
	}
	if result := network.Validate(req); !result.Valid {
		return c.Send(ctx).BadRequestError(result.FirstMessage(), nil)
	}

	post, err := c.service.RejectPost(ctx.Context(), postID, reviewerID, req.ReviewNote)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Post berhasil ditolak", post)
}

// GetHistory godoc
// @Summary     Riwayat moderasi
// @Description Mengambil daftar post yang sudah direview (approved/rejected)
// @Tags        Moderation
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       page  query int false "Nomor halaman" default(1)
// @Param       limit query int false "Jumlah item per halaman" default(20)
// @Success     200 {object} network.Response "Riwayat moderasi berhasil diambil"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     403 {object} network.ErrorResponse "Forbidden"
// @Router      /api/moderation/history [get]
func (c *moderationController) GetHistory(ctx *fiber.Ctx) error {
	pagination := network.DefaultPagination()
	if err := ctx.QueryParser(&pagination); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidPagination, err)
	}

	result, err := c.service.GetHistory(ctx.Context(), pagination)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Riwayat moderasi berhasil diambil", result)
}
