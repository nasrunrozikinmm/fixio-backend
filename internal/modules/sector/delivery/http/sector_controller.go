package controllers

import (
	sectorDto "fixio/internal/modules/sector/dto"
	models "fixio/internal/modules/sector/entity"
	services "fixio/internal/modules/sector/usecase"
	"fixio/pkg/helpers"
	"fixio/pkg/network"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type sectorController struct {
	network.BaseController
	service services.SectorService
}

// NewSectorController creates a new sector controller
func NewSectorController(
	authFn network.AuthenticationProvider,
	authzFn network.AuthorizationProvider,
	service services.SectorService,
) network.Controller {
	return &sectorController{
		BaseController: network.NewBaseController("/sectors", authFn, authzFn),
		service:        service,
	}
}

// MountRoutes registers all sector routes
func (c *sectorController) MountRoutes(rg *fiber.Group) {
	// Public
	rg.Get("/", c.GetAll)
	rg.Get("/:id", c.GetByID)
	// Admin only
	rg.Post("/", c.Authentication(), c.Grant("admin:manage_sectors"), c.Create)
	rg.Put("/:id", c.Authentication(), c.Grant("admin:manage_sectors"), c.Update)
	rg.Delete("/:id", c.Authentication(), c.Grant("admin:manage_sectors"), c.Delete)
}

// GetAll godoc
// @Summary     Daftar semua sektor
// @Description Mengambil daftar semua sektor kebijakan dengan paginasi
// @Tags        Sectors
// @Accept      json
// @Produce     json
// @Param       page  query int    false "Nomor halaman" default(1)
// @Param       limit query int    false "Jumlah item per halaman" default(20)
// @Param       sort  query string false "Pengurutan" default(created_at desc)
// @Success     200 {object} network.Response "Daftar sektor berhasil diambil"
// @Router      /api/sectors [get]
func (c *sectorController) GetAll(ctx *fiber.Ctx) error {
	pagination := network.DefaultPagination()
	if err := ctx.QueryParser(&pagination); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidPagination, err)
	}

	result, err := c.service.GetAll(ctx.Context(), pagination, nil)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}
	return c.Send(ctx).SuccessDataResponse("Daftar sektor berhasil diambil", result)
}

// GetByID godoc
// @Summary     Detail sektor
// @Description Mengambil detail satu sektor berdasarkan ID
// @Tags        Sectors
// @Accept      json
// @Produce     json
// @Param       id path string true "Sector ID (UUID)"
// @Success     200 {object} network.Response "Detail sektor berhasil diambil"
// @Failure     404 {object} network.ErrorResponse "Sektor tidak ditemukan"
// @Router      /api/sectors/{id} [get]
func (c *sectorController) GetByID(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	result, err := c.service.GetOne(ctx.Context(), map[string]any{"id": id})
	if err != nil {
		return c.Send(ctx).NotFoundError("Sektor tidak ditemukan", err)
	}
	return c.Send(ctx).SuccessDataResponse("Detail sektor berhasil diambil", result)
}

// Create godoc
// @Summary     Buat sektor baru
// @Description Membuat sektor kebijakan baru (khusus Administrator)
// @Tags        Sectors
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body sectorDto.CreateSectorRequest true "Data sektor baru"
// @Success     201 {object} network.Response "Sektor berhasil dibuat"
// @Failure     400 {object} network.ErrorResponse "Request body tidak valid"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     403 {object} network.ErrorResponse "Forbidden — bukan administrator"
// @Router      /api/sectors [post]
func (c *sectorController) Create(ctx *fiber.Ctx) error {
	req := new(sectorDto.CreateSectorRequest)
	if err := ctx.BodyParser(req); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidBody, err)
	}
	if result := network.Validate(req); !result.Valid {
		return c.Send(ctx).BadRequestError(result.FirstMessage(), nil)
	}

	sector := &models.Sector{
		ID:   uuid.New(),
		Name: req.Name,
		Slug: helpers.Slugify(req.Name),
	}

	created, err := c.service.Create(ctx.Context(), sector)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}
	return c.Send(ctx).SuccessCreatedResponse("Sektor berhasil dibuat", created)
}

// Update godoc
// @Summary     Update sektor
// @Description Mengupdate data sektor (khusus Administrator)
// @Tags        Sectors
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string true "Sector ID (UUID)"
// @Param       body body sectorDto.UpdateSectorRequest true "Data sektor yang akan diupdate"
// @Success     200 {object} network.Response "Sektor berhasil diupdate"
// @Failure     400 {object} network.ErrorResponse "Request body tidak valid"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     403 {object} network.ErrorResponse "Forbidden — bukan administrator"
// @Failure     404 {object} network.ErrorResponse "Sektor tidak ditemukan"
// @Router      /api/sectors/{id} [put]
func (c *sectorController) Update(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	req := new(sectorDto.UpdateSectorRequest)
	if err := ctx.BodyParser(req); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidBody, err)
	}
	if result := network.Validate(req); !result.Valid {
		return c.Send(ctx).BadRequestError(result.FirstMessage(), nil)
	}

	sector := &models.Sector{
		Name: req.Name,
		Slug: helpers.Slugify(req.Name),
	}

	updated, err := c.service.Update(ctx.Context(), id, sector)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}
	return c.Send(ctx).SuccessDataResponse("Sektor berhasil diupdate", updated)
}

// Delete godoc
// @Summary     Hapus sektor
// @Description Menghapus sektor berdasarkan ID (khusus Administrator)
// @Tags        Sectors
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id path string true "Sector ID (UUID)"
// @Success     200 {object} network.Response "Sektor berhasil dihapus"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     403 {object} network.ErrorResponse "Forbidden — bukan administrator"
// @Failure     404 {object} network.ErrorResponse "Sektor tidak ditemukan"
// @Router      /api/sectors/{id} [delete]
func (c *sectorController) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return c.Send(ctx).HandleError(err)
	}
	return c.Send(ctx).SuccessMsgResponse("Sektor berhasil dihapus")
}
