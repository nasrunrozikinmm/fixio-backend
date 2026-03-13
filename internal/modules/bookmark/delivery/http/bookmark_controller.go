package controllers

import (
	services "fixio/internal/modules/bookmark/usecase"
	"fixio/pkg/network"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type bookmarkController struct {
	network.BaseController
	service services.BookmarkService
}

// NewBookmarkController creates a new bookmark controller
func NewBookmarkController(
	authFn network.AuthenticationProvider,
	authzFn network.AuthorizationProvider,
	service services.BookmarkService,
) network.Controller {
	return &bookmarkController{
		BaseController: network.NewBaseController("/bookmarks", authFn, authzFn),
		service:        service,
	}
}

// MountRoutes registers all bookmark routes
func (c *bookmarkController) MountRoutes(rg *fiber.Group) {
	rg.Post("/posts/:id", c.Authentication(), c.AddBookmark)
	rg.Delete("/posts/:id", c.Authentication(), c.RemoveBookmark)
	rg.Get("/posts/:id/status", c.Authentication(), c.GetBookmarkStatus)
	rg.Get("/", c.Authentication(), c.GetMyBookmarks)
}

// AddBookmark godoc
// @Summary     Bookmark post
// @Description Tambahkan post ke daftar bookmark user.
// @Tags        Bookmark
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Post ID (UUID)"
// @Success     201 {object} network.Response "Post berhasil di-bookmark"
// @Failure     400 {object} network.ErrorResponse "ID tidak valid / sudah di-bookmark"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     404 {object} network.ErrorResponse "Post tidak ditemukan"
// @Router      /api/bookmarks/posts/{id} [post]
func (c *bookmarkController) AddBookmark(ctx *fiber.Ctx) error {
	postID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	userIDStr, _ := ctx.Locals("userId").(string)
	userID, _ := uuid.Parse(userIDStr)

	bookmark, err := c.service.AddBookmark(ctx.Context(), userID, postID)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessCreatedResponse("Post berhasil di-bookmark", bookmark)
}

// RemoveBookmark godoc
// @Summary     Hapus bookmark
// @Description Hapus post dari daftar bookmark user.
// @Tags        Bookmark
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Post ID (UUID)"
// @Success     200 {object} network.Response "Bookmark berhasil dihapus"
// @Failure     400 {object} network.ErrorResponse "ID tidak valid"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     404 {object} network.ErrorResponse "Bookmark tidak ditemukan"
// @Router      /api/bookmarks/posts/{id} [delete]
func (c *bookmarkController) RemoveBookmark(ctx *fiber.Ctx) error {
	postID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	userIDStr, _ := ctx.Locals("userId").(string)
	userID, _ := uuid.Parse(userIDStr)

	if err := c.service.RemoveBookmark(ctx.Context(), userID, postID); err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessMsgResponse("Bookmark berhasil dihapus")
}

// GetBookmarkStatus godoc
// @Summary     Cek status bookmark
// @Description Cek apakah user sudah bookmark post tertentu.
// @Tags        Bookmark
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Post ID (UUID)"
// @Success     200 {object} network.Response "Status bookmark berhasil diambil"
// @Failure     400 {object} network.ErrorResponse "ID tidak valid"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Router      /api/bookmarks/posts/{id}/status [get]
func (c *bookmarkController) GetBookmarkStatus(ctx *fiber.Ctx) error {
	postID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	userIDStr, _ := ctx.Locals("userId").(string)
	userID, _ := uuid.Parse(userIDStr)

	isBookmarked, err := c.service.IsBookmarked(ctx.Context(), userID, postID)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Status bookmark berhasil diambil", map[string]bool{
		"is_bookmarked": isBookmarked,
	})
}

// GetMyBookmarks godoc
// @Summary     Ambil daftar bookmark saya
// @Description Mengambil daftar post yang di-bookmark oleh user yang login.
// @Tags        Bookmark
// @Security    BearerAuth
// @Produce     json
// @Param       page  query int false "Nomor halaman" default(1)
// @Param       limit query int false "Jumlah per halaman" default(10)
// @Success     200 {object} network.Response "Daftar bookmark berhasil diambil"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Router      /api/bookmarks [get]
func (c *bookmarkController) GetMyBookmarks(ctx *fiber.Ctx) error {
	userIDStr, _ := ctx.Locals("userId").(string)
	userID, _ := uuid.Parse(userIDStr)

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))

	posts, total, err := c.service.GetUserBookmarkedPosts(ctx.Context(), userID, page, limit)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Daftar bookmark berhasil diambil", map[string]interface{}{
		"posts": posts,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
