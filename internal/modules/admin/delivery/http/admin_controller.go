package controllers

import (
	"fixio/internal/modules/admin/dto"
	services "fixio/internal/modules/admin/usecase"
	"fixio/pkg/network"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type adminController struct {
	network.BaseController
	service services.AdminService
}

// NewAdminController creates a new admin controller
func NewAdminController(
	authFn network.AuthenticationProvider,
	authzFn network.AuthorizationProvider,
	service services.AdminService,
) network.Controller {
	return &adminController{
		BaseController: network.NewBaseController("/admin", authFn, authzFn),
		service:        service,
	}
}

// MountRoutes registers all admin routes
func (c *adminController) MountRoutes(rg *fiber.Group) {
	rg.Use(c.Authentication())
	rg.Get("/users", c.Grant("admin:manage_users"), c.GetAllUsers)
	rg.Put("/users/:id/role", c.Grant("admin:manage_roles"), c.ChangeUserRole)
	rg.Get("/stats", c.Grant("admin:view_stats"), c.GetStats)
}

// GetAllUsers godoc
// @Summary     Daftar semua user
// @Description Mengambil daftar semua user platform dengan paginasi (khusus Administrator)
// @Tags        Admin
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       page  query int    false "Nomor halaman" default(1)
// @Param       limit query int    false "Jumlah item per halaman" default(20)
// @Success     200 {object} network.Response "Daftar user berhasil diambil"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     403 {object} network.ErrorResponse "Forbidden"
// @Router      /api/admin/users [get]
func (c *adminController) GetAllUsers(ctx *fiber.Ctx) error {
	pagination := network.DefaultPagination()
	if err := ctx.QueryParser(&pagination); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidPagination, err)
	}

	result, err := c.service.GetAllUsers(ctx.Context(), pagination)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Daftar user berhasil diambil", result)
}

// ChangeUserRole godoc
// @Summary     Ubah role user
// @Description Mengubah role user (creator/moderator/administrator)
// @Tags        Admin
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string true "Target User ID (UUID)"
// @Param       body body dto.ChangeRoleRequest true "Role baru"
// @Success     200 {object} network.Response "Role user berhasil diubah"
// @Failure     400 {object} network.ErrorResponse "Request tidak valid"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     403 {object} network.ErrorResponse "Forbidden"
// @Router      /api/admin/users/{id}/role [put]
func (c *adminController) ChangeUserRole(ctx *fiber.Ctx) error {
	targetUserID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	adminUserIDStr, _ := ctx.Locals("userId").(string)
	adminUserID, _ := uuid.Parse(adminUserIDStr)

	req := new(dto.ChangeRoleRequest)
	if err := ctx.BodyParser(req); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidBody, err)
	}
	if result := network.Validate(req); !result.Valid {
		return c.Send(ctx).BadRequestError(result.FirstMessage(), nil)
	}

	user, err := c.service.ChangeUserRole(ctx.Context(), targetUserID, adminUserID, req.Role)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Role user berhasil diubah", user)
}

// GetStats godoc
// @Summary     Statistik platform
// @Description Mengambil statistik platform (total users, posts, votes, dll)
// @Tags        Admin
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Success     200 {object} network.Response{data=dto.StatsResponse} "Statistik berhasil diambil"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     403 {object} network.ErrorResponse "Forbidden"
// @Router      /api/admin/stats [get]
func (c *adminController) GetStats(ctx *fiber.Ctx) error {
	stats, err := c.service.GetStats(ctx.Context())
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Statistik platform berhasil diambil", stats)
}
