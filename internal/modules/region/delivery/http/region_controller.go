package controllers

import (
	regionDto "fixio/internal/modules/region/dto"
	models "fixio/internal/modules/region/entity"
	services "fixio/internal/modules/region/usecase"
	"fixio/pkg/helpers"
	"fixio/pkg/network"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type regionController struct {
	network.BaseController
	service services.RegionService
}

// NewRegionController creates a new region controller
func NewRegionController(
	authFn network.AuthenticationProvider,
	authzFn network.AuthorizationProvider,
	service services.RegionService,
) network.Controller {
	return &regionController{
		BaseController: network.NewBaseController("/regions", authFn, authzFn),
		service:        service,
	}
}

// MountRoutes registers all region routes
func (c *regionController) MountRoutes(rg *fiber.Group) {
	// Public
	rg.Get("/", c.GetAll)
	rg.Get("/type/:type", c.GetByType)
	rg.Get("/:id", c.GetByID)
	// Admin only
	rg.Post("/", c.Authentication(), c.Grant("admin:manage_regions"), c.Create)
	rg.Put("/:id", c.Authentication(), c.Grant("admin:manage_regions"), c.Update)
	rg.Delete("/:id", c.Authentication(), c.Grant("admin:manage_regions"), c.Delete)
}

// GetAll godoc
// @Summary     Daftar semua wilayah
// @Description Mengambil daftar semua wilayah (nasional, provinsi, kota) dengan paginasi
// @Tags        Regions
// @Accept      json
// @Produce     json
// @Param       page  query int    false "Nomor halaman" default(1)
// @Param       limit query int    false "Jumlah item per halaman" default(20)
// @Param       sort  query string false "Pengurutan" default(created_at desc)
// @Success     200 {object} network.Response "Daftar wilayah berhasil diambil"
// @Router      /api/regions [get]
func (c *regionController) GetAll(ctx *fiber.Ctx) error {
	pagination := network.DefaultPagination()
	if err := ctx.QueryParser(&pagination); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidPagination, err)
	}

	result, err := c.service.GetAll(ctx.Context(), pagination, nil)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}
	return c.Send(ctx).SuccessDataResponse("Daftar wilayah berhasil diambil", result)
}

// GetByType godoc
// @Summary     Daftar wilayah berdasarkan tipe
// @Description Mengambil daftar wilayah berdasarkan tipe (nasional, provinsi, kota)
// @Tags        Regions
// @Accept      json
// @Produce     json
// @Param       type path string true "Tipe wilayah (nasional, provinsi, kota)"
// @Success     200 {object} network.Response "Daftar wilayah berhasil diambil"
// @Failure     400 {object} network.ErrorResponse "Tipe tidak valid"
// @Router      /api/regions/type/{type} [get]
func (c *regionController) GetByType(ctx *fiber.Ctx) error {
	regionType := ctx.Params("type")
	regions, err := c.service.GetByType(ctx.Context(), regionType)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}
	return c.Send(ctx).SuccessDataResponse("Daftar wilayah berhasil diambil", regions)
}

// GetByID godoc
// @Summary     Detail wilayah
// @Description Mengambil detail satu wilayah berdasarkan ID
// @Tags        Regions
// @Accept      json
// @Produce     json
// @Param       id path string true "Region ID (UUID)"
// @Success     200 {object} network.Response "Detail wilayah berhasil diambil"
// @Failure     404 {object} network.ErrorResponse "Wilayah tidak ditemukan"
// @Router      /api/regions/{id} [get]
func (c *regionController) GetByID(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	result, err := c.service.GetOne(ctx.Context(), map[string]any{"id": id})
	if err != nil {
		return c.Send(ctx).NotFoundError("Wilayah tidak ditemukan", err)
	}
	return c.Send(ctx).SuccessDataResponse("Detail wilayah berhasil diambil", result)
}

// Create godoc
// @Summary     Buat wilayah baru
// @Description Membuat wilayah baru (nasional/provinsi/kota) — khusus Administrator
// @Tags        Regions
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body regionDto.CreateRegionRequest true "Data wilayah baru"
// @Success     201 {object} network.Response "Wilayah berhasil dibuat"
// @Failure     400 {object} network.ErrorResponse "Request body tidak valid"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     403 {object} network.ErrorResponse "Forbidden — bukan administrator"
// @Router      /api/regions [post]
func (c *regionController) Create(ctx *fiber.Ctx) error {
	req := new(regionDto.CreateRegionRequest)
	if err := ctx.BodyParser(req); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidBody, err)
	}
	if result := network.Validate(req); !result.Valid {
		return c.Send(ctx).BadRequestError(result.FirstMessage(), nil)
	}

	region := &models.Region{
		ID:   uuid.New(),
		Name: req.Name,
		Slug: helpers.Slugify(req.Name),
		Type: req.Type,
	}
	if req.ParentID != "" {
		pid, err := uuid.Parse(req.ParentID)
		if err != nil {
			return c.Send(ctx).BadRequestError("ID parent wilayah tidak valid", err)
		}
		region.ParentID = &pid
	}

	created, err := c.service.Create(ctx.Context(), region)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}
	return c.Send(ctx).SuccessCreatedResponse("Wilayah berhasil dibuat", created)
}

// Update godoc
// @Summary     Update wilayah
// @Description Mengupdate data wilayah — khusus Administrator
// @Tags        Regions
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string true "Region ID (UUID)"
// @Param       body body regionDto.UpdateRegionRequest true "Data wilayah yang akan diupdate"
// @Success     200 {object} network.Response "Wilayah berhasil diupdate"
// @Failure     400 {object} network.ErrorResponse "Request body tidak valid"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     403 {object} network.ErrorResponse "Forbidden — bukan administrator"
// @Failure     404 {object} network.ErrorResponse "Wilayah tidak ditemukan"
// @Router      /api/regions/{id} [put]
func (c *regionController) Update(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	req := new(regionDto.UpdateRegionRequest)
	if err := ctx.BodyParser(req); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidBody, err)
	}
	if result := network.Validate(req); !result.Valid {
		return c.Send(ctx).BadRequestError(result.FirstMessage(), nil)
	}

	region := &models.Region{
		Name: req.Name,
		Slug: helpers.Slugify(req.Name),
		Type: req.Type,
	}

	updated, err := c.service.Update(ctx.Context(), id, region)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}
	return c.Send(ctx).SuccessDataResponse("Wilayah berhasil diupdate", updated)
}

// Delete godoc
// @Summary     Hapus wilayah
// @Description Menghapus wilayah berdasarkan ID — khusus Administrator
// @Tags        Regions
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id path string true "Region ID (UUID)"
// @Success     200 {object} network.Response "Wilayah berhasil dihapus"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     403 {object} network.ErrorResponse "Forbidden — bukan administrator"
// @Failure     404 {object} network.ErrorResponse "Wilayah tidak ditemukan"
// @Router      /api/regions/{id} [delete]
func (c *regionController) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return c.Send(ctx).HandleError(err)
	}
	return c.Send(ctx).SuccessMsgResponse("Wilayah berhasil dihapus")
}
