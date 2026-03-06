package controllers

import (
	"fixio/internal/modules/post/dto"
	services "fixio/internal/modules/post/usecase"
	"fixio/pkg/network"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type postController struct {
	network.BaseController
	service services.PostService
}

// NewPostController creates a new post controller
func NewPostController(
	authFn network.AuthenticationProvider,
	authzFn network.AuthorizationProvider,
	service services.PostService,
) network.Controller {
	return &postController{
		BaseController: network.NewBaseController("/posts", authFn, authzFn),
		service:        service,
	}
}

// MountRoutes registers all post routes
func (c *postController) MountRoutes(rg *fiber.Group) {
	rg.Get("/", c.GetAll)
	rg.Get("/:id", c.GetByID)
	rg.Post("/", c.Authentication(), c.Grant("post:create"), c.Create)
	rg.Put("/:id", c.Authentication(), c.Grant("post:edit_own"), c.Update)
	rg.Delete("/:id", c.Authentication(), c.Delete)
}

// GetAll godoc
// @Summary     Daftar semua post
// @Description Mengambil daftar semua post yang sudah disetujui, dengan filter dan paginasi
// @Tags        Posts
// @Accept      json
// @Produce     json
// @Param       page      query int    false "Nomor halaman" default(1)
// @Param       limit     query int    false "Jumlah item per halaman" default(20)
// @Param       sort      query string false "Pengurutan" default(created_at desc)
// @Param       sector_id query string false "Filter berdasarkan sektor (UUID)"
// @Param       region_id query string false "Filter berdasarkan wilayah (UUID)"
// @Param       status    query string false "Filter berdasarkan status (draft, pending_review, approved, rejected)"
// @Param       user_id   query string false "Filter berdasarkan user (UUID)"
// @Param       search    query string false "Kata kunci pencarian"
// @Success     200 {object} network.Response "Daftar post berhasil diambil"
// @Router      /api/posts [get]
func (c *postController) GetAll(ctx *fiber.Ctx) error {
	pagination := network.DefaultPagination()
	if err := ctx.QueryParser(&pagination); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidPagination, err)
	}

	filter := &dto.PostFilter{}
	if err := ctx.QueryParser(filter); err != nil {
		return c.Send(ctx).BadRequestError("Filter tidak valid", err)
	}

	result, err := c.service.GetAll(ctx.Context(), pagination, filter)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Daftar post berhasil diambil", result)
}

// GetByID godoc
// @Summary     Detail post
// @Description Mengambil detail satu post berdasarkan ID
// @Tags        Posts
// @Accept      json
// @Produce     json
// @Param       id path string true "Post ID (UUID)"
// @Success     200 {object} network.Response "Detail post berhasil diambil"
// @Failure     400 {object} network.ErrorResponse "ID tidak valid"
// @Failure     404 {object} network.ErrorResponse "Post tidak ditemukan"
// @Router      /api/posts/{id} [get]
func (c *postController) GetByID(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	post, err := c.service.GetByID(ctx.Context(), id)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Detail post berhasil diambil", post)
}

// Create godoc
// @Summary     Buat post baru
// @Description Membuat post baru (kritik + solusi kebijakan). Post akan masuk ke antrian moderasi.
// @Tags        Posts
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body dto.CreatePostRequest true "Data post baru"
// @Success     201 {object} network.Response "Post berhasil dibuat"
// @Failure     400 {object} network.ErrorResponse "Request body tidak valid"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     403 {object} network.ErrorResponse "Forbidden"
// @Router      /api/posts [post]
func (c *postController) Create(ctx *fiber.Ctx) error {
	userIDStr, _ := ctx.Locals("userId").(string)
	userID, _ := uuid.Parse(userIDStr)

	req := new(dto.CreatePostRequest)
	if err := ctx.BodyParser(req); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidBody, err)
	}
	if result := network.Validate(req); !result.Valid {
		return c.Send(ctx).BadRequestError(result.FirstMessage(), nil)
	}

	post, err := c.service.Create(ctx.Context(), userID, req)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessCreatedResponse("Post berhasil dibuat", post)
}

// Update godoc
// @Summary     Update post
// @Description Mengupdate post milik sendiri (hanya pemilik yang bisa edit)
// @Tags        Posts
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string true "Post ID (UUID)"
// @Param       body body dto.UpdatePostRequest true "Data post yang akan diupdate"
// @Success     200 {object} network.Response "Post berhasil diupdate"
// @Failure     400 {object} network.ErrorResponse "Request body tidak valid"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     403 {object} network.ErrorResponse "Forbidden — bukan pemilik post"
// @Failure     404 {object} network.ErrorResponse "Post tidak ditemukan"
// @Router      /api/posts/{id} [put]
func (c *postController) Update(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	userIDStr, _ := ctx.Locals("userId").(string)
	userID, _ := uuid.Parse(userIDStr)

	req := new(dto.UpdatePostRequest)
	if err := ctx.BodyParser(req); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidBody, err)
	}
	if result := network.Validate(req); !result.Valid {
		return c.Send(ctx).BadRequestError(result.FirstMessage(), nil)
	}

	post, err := c.service.Update(ctx.Context(), id, userID, req)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Post berhasil diupdate", post)
}

// Delete godoc
// @Summary     Hapus post
// @Description Menghapus post (pemilik, moderator, atau administrator)
// @Tags        Posts
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id path string true "Post ID (UUID)"
// @Success     200 {object} network.Response "Post berhasil dihapus"
// @Failure     400 {object} network.ErrorResponse "ID tidak valid"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     403 {object} network.ErrorResponse "Forbidden"
// @Failure     404 {object} network.ErrorResponse "Post tidak ditemukan"
// @Router      /api/posts/{id} [delete]
func (c *postController) Delete(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	userIDStr, _ := ctx.Locals("userId").(string)
	userID, _ := uuid.Parse(userIDStr)
	userRole, _ := ctx.Locals("userRole").(string)

	if err := c.service.Delete(ctx.Context(), id, userID, userRole); err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessMsgResponse("Post berhasil dihapus")
}
