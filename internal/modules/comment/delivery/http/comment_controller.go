package controllers

import (
	"fixio/internal/modules/comment/dto"
	services "fixio/internal/modules/comment/usecase"
	"fixio/pkg/network"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type commentController struct {
	network.BaseController
	service services.CommentService
}

// NewCommentController creates a new comment controller
func NewCommentController(
	authFn network.AuthenticationProvider,
	authzFn network.AuthorizationProvider,
	service services.CommentService,
) network.Controller {
	return &commentController{
		BaseController: network.NewBaseController("/posts", authFn, authzFn),
		service:        service,
	}
}

// MountRoutes registers all comment routes
func (c *commentController) MountRoutes(rg *fiber.Group) {
	rg.Get("/:id/comments", c.GetByPostID)
	rg.Post("/:id/comments", c.Authentication(), c.Grant("comment:create"), c.Create)
}

// GetByPostID godoc
// @Summary     Daftar komentar post
// @Description Mengambil daftar komentar pada sebuah post dengan paginasi
// @Tags        Comments
// @Accept      json
// @Produce     json
// @Param       id    path  string true  "Post ID (UUID)"
// @Param       page  query int    false "Nomor halaman" default(1)
// @Param       limit query int    false "Jumlah item per halaman" default(20)
// @Success     200 {object} network.Response "Komentar berhasil diambil"
// @Failure     400 {object} network.ErrorResponse "ID tidak valid"
// @Failure     404 {object} network.ErrorResponse "Post tidak ditemukan"
// @Router      /api/posts/{id}/comments [get]
func (c *commentController) GetByPostID(ctx *fiber.Ctx) error {
	postID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	pagination := network.DefaultPagination()
	if err := ctx.QueryParser(&pagination); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidPagination, err)
	}

	result, err := c.service.GetByPostID(ctx.Context(), postID, pagination)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Komentar berhasil diambil", result)
}

// Create godoc
// @Summary     Tambah komentar
// @Description Menambahkan komentar baru pada sebuah post. Mendukung reply (1 level nesting) via parent_id.
// @Tags        Comments
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string true "Post ID (UUID)"
// @Param       body body dto.CreateCommentRequest true "Data komentar baru"
// @Success     201 {object} network.Response "Komentar berhasil ditambahkan"
// @Failure     400 {object} network.ErrorResponse "Request body tidak valid"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     403 {object} network.ErrorResponse "Forbidden"
// @Failure     404 {object} network.ErrorResponse "Post tidak ditemukan"
// @Router      /api/posts/{id}/comments [post]
func (c *commentController) Create(ctx *fiber.Ctx) error {
	postID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	userIDStr, _ := ctx.Locals("userId").(string)
	userID, _ := uuid.Parse(userIDStr)

	req := new(dto.CreateCommentRequest)
	if err := ctx.BodyParser(req); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidBody, err)
	}
	if result := network.Validate(req); !result.Valid {
		return c.Send(ctx).BadRequestError(result.FirstMessage(), nil)
	}

	var parentID *uuid.UUID
	if req.ParentID != "" {
		pid, err := uuid.Parse(req.ParentID)
		if err != nil {
			return c.Send(ctx).BadRequestError("ID parent komentar tidak valid", err)
		}
		parentID = &pid
	}

	comment, err := c.service.Create(ctx.Context(), userID, postID, req.Content, parentID)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessCreatedResponse("Komentar berhasil ditambahkan", comment)
}
