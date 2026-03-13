package controllers

import (
	"fixio/internal/modules/report/dto"
	services "fixio/internal/modules/report/usecase"
	"fixio/pkg/network"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type reportController struct {
	network.BaseController
	service services.ReportService
}

// NewReportController creates a new report controller
func NewReportController(
	authFn network.AuthenticationProvider,
	authzFn network.AuthorizationProvider,
	service services.ReportService,
) network.Controller {
	return &reportController{
		BaseController: network.NewBaseController("/reports", authFn, authzFn),
		service:        service,
	}
}

// MountRoutes registers report routes
func (c *reportController) MountRoutes(rg *fiber.Group) {
	// Authenticated users can report
	rg.Post("/", c.Authentication(), c.Grant("comment:create"), c.Create)
	// Moderator/Admin can review reports
	rg.Get("/", c.Authentication(), c.Grant("moderation:review"), c.GetAll)
	rg.Get("/pending", c.Authentication(), c.Grant("moderation:review"), c.GetPending)
	rg.Get("/count", c.Authentication(), c.Grant("moderation:review"), c.CountPending)
	rg.Put("/:id/review", c.Authentication(), c.Grant("moderation:review"), c.Review)
}

// Create godoc
// @Summary     Laporkan konten
// @Description Melaporkan post atau komentar yang melanggar aturan
// @Tags        Reports
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body dto.CreateReportRequest true "Data laporan"
// @Success     201 {object} network.Response "Laporan berhasil dibuat"
// @Failure     400 {object} network.ErrorResponse "Request tidak valid"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Router      /api/reports [post]
func (c *reportController) Create(ctx *fiber.Ctx) error {
	userIDStr, _ := ctx.Locals("userId").(string)
	userID, _ := uuid.Parse(userIDStr)

	req := new(dto.CreateReportRequest)
	if err := ctx.BodyParser(req); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidBody, err)
	}
	if result := network.Validate(req); !result.Valid {
		return c.Send(ctx).BadRequestError(result.FirstMessage(), nil)
	}

	report, err := c.service.Create(ctx.Context(), userID, req.TargetType, req.TargetID, req.Reason, req.Description)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessCreatedResponse("Laporan berhasil dibuat", report)
}

// GetAll godoc
// @Summary     Daftar semua laporan
// @Description Mengambil semua laporan dengan filter status opsional (moderator/admin)
// @Tags        Reports
// @Security    BearerAuth
// @Produce     json
// @Param       status query string false "Filter status: pending, reviewed, dismissed"
// @Param       page   query int    false "Halaman" default(1)
// @Param       limit  query int    false "Limit" default(20)
// @Success     200 {object} network.Response "Daftar laporan"
// @Router      /api/reports [get]
func (c *reportController) GetAll(ctx *fiber.Ctx) error {
	pagination := network.DefaultPagination()
	if err := ctx.QueryParser(&pagination); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidPagination, err)
	}
	status := ctx.Query("status")

	result, err := c.service.GetAll(ctx.Context(), status, pagination)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Daftar laporan berhasil diambil", result)
}

// GetPending godoc
// @Summary     Daftar laporan pending
// @Description Mengambil laporan yang belum di-review (moderator/admin)
// @Tags        Reports
// @Security    BearerAuth
// @Produce     json
// @Param       page  query int false "Halaman" default(1)
// @Param       limit query int false "Limit" default(20)
// @Success     200 {object} network.Response "Daftar laporan pending"
// @Router      /api/reports/pending [get]
func (c *reportController) GetPending(ctx *fiber.Ctx) error {
	pagination := network.DefaultPagination()
	if err := ctx.QueryParser(&pagination); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidPagination, err)
	}

	result, err := c.service.GetPending(ctx.Context(), pagination)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Laporan pending berhasil diambil", result)
}

// CountPending godoc
// @Summary     Jumlah laporan pending
// @Description Mengambil jumlah laporan yang belum di-review
// @Tags        Reports
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} network.Response "Jumlah laporan pending"
// @Router      /api/reports/count [get]
func (c *reportController) CountPending(ctx *fiber.Ctx) error {
	count, err := c.service.CountPending(ctx.Context())
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Jumlah laporan pending", fiber.Map{
		"pending_count": count,
	})
}

// Review godoc
// @Summary     Review laporan
// @Description Moderator/Admin mereview laporan (reviewed/dismissed)
// @Tags        Reports
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string true "Report ID (UUID)"
// @Param       body body dto.ReviewReportRequest true "Data review"
// @Success     200 {object} network.Response "Laporan berhasil di-review"
// @Failure     400 {object} network.ErrorResponse "Request tidak valid"
// @Failure     404 {object} network.ErrorResponse "Laporan tidak ditemukan"
// @Router      /api/reports/{id}/review [put]
func (c *reportController) Review(ctx *fiber.Ctx) error {
	reportID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	userIDStr, _ := ctx.Locals("userId").(string)
	reviewerID, _ := uuid.Parse(userIDStr)

	req := new(dto.ReviewReportRequest)
	if err := ctx.BodyParser(req); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidBody, err)
	}
	if result := network.Validate(req); !result.Valid {
		return c.Send(ctx).BadRequestError(result.FirstMessage(), nil)
	}

	report, err := c.service.Review(ctx.Context(), reportID, reviewerID, req.Status, req.ReviewNote)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Laporan berhasil di-review", report)
}
