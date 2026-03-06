package controllers

import (
	services "fixio/internal/modules/auth/usecase"
	postServices "fixio/internal/modules/post/usecase"
	"fixio/pkg/network"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type userController struct {
	network.BaseController
	authService services.AuthService
	postService postServices.PostService
}

// NewUserController creates a new user/profile controller
func NewUserController(
	authFn network.AuthenticationProvider,
	authzFn network.AuthorizationProvider,
	authService services.AuthService,
	postService postServices.PostService,
) network.Controller {
	return &userController{
		BaseController: network.NewBaseController("/users", authFn, authzFn),
		authService:    authService,
		postService:    postService,
	}
}

// MountRoutes registers user profile routes
func (c *userController) MountRoutes(rg *fiber.Group) {
	rg.Get("/:id", c.GetUserProfile)
}

// GetUserProfile godoc
// @Summary     Get profil user publik
// @Description Mengambil profil publik user beserta daftar post-nya berdasarkan user ID
// @Tags        Users
// @Accept      json
// @Produce     json
// @Param       id   path string true "User ID (UUID)"
// @Param       page  query int false "Nomor halaman" default(1)
// @Param       limit query int false "Jumlah item per halaman" default(20)
// @Success     200 {object} network.Response "Profil user berhasil diambil"
// @Failure     400 {object} network.ErrorResponse "ID tidak valid"
// @Failure     404 {object} network.ErrorResponse "User tidak ditemukan"
// @Router      /api/users/{id} [get]
func (c *userController) GetUserProfile(ctx *fiber.Ctx) error {
	userID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	profile, err := c.authService.GetCurrentUser(ctx.Context(), userID)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	pagination := network.DefaultPagination()
	if err := ctx.QueryParser(&pagination); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidPagination, err)
	}

	posts, err := c.postService.GetByUserID(ctx.Context(), userID, pagination)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	result := map[string]interface{}{
		"user":  profile,
		"posts": posts,
	}

	return c.Send(ctx).SuccessDataResponse("Profil user berhasil diambil", result)
}
