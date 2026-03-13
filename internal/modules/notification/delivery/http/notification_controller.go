package controllers

import (
	services "fixio/internal/modules/notification/usecase"
	"fixio/pkg/network"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type notificationController struct {
	network.BaseController
	service services.NotificationService
}

// NewNotificationController creates a new notification controller
func NewNotificationController(
	authFn network.AuthenticationProvider,
	authzFn network.AuthorizationProvider,
	service services.NotificationService,
) network.Controller {
	return &notificationController{
		BaseController: network.NewBaseController("/notifications", authFn, authzFn),
		service:        service,
	}
}

// MountRoutes registers all notification routes
func (c *notificationController) MountRoutes(rg *fiber.Group) {
	rg.Get("/", c.Authentication(), c.GetNotifications)
	rg.Get("/count", c.Authentication(), c.GetUnreadCount)
	rg.Put("/:id/read", c.Authentication(), c.MarkAsRead)
	rg.Put("/read-all", c.Authentication(), c.MarkAllAsRead)
}

// GetNotifications godoc
// @Summary     Ambil daftar notifikasi
// @Description Mengambil daftar notifikasi milik user yang login dengan pagination.
// @Tags        Notification
// @Security    BearerAuth
// @Produce     json
// @Param       page  query int false "Nomor halaman" default(1)
// @Param       limit query int false "Jumlah per halaman" default(10)
// @Success     200 {object} network.Response "Daftar notifikasi berhasil diambil"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Router      /api/notifications [get]
func (c *notificationController) GetNotifications(ctx *fiber.Ctx) error {
	userIDStr, _ := ctx.Locals("userId").(string)
	userID, _ := uuid.Parse(userIDStr)

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))

	notifications, total, err := c.service.GetNotifications(ctx.Context(), userID, page, limit)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Daftar notifikasi berhasil diambil", map[string]interface{}{
		"notifications": notifications,
		"total":         total,
		"page":          page,
		"limit":         limit,
	})
}

// GetUnreadCount godoc
// @Summary     Ambil jumlah notifikasi belum dibaca
// @Description Mengambil jumlah notifikasi yang belum dibaca oleh user yang login.
// @Tags        Notification
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} network.Response "Jumlah notifikasi belum dibaca berhasil diambil"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Router      /api/notifications/count [get]
func (c *notificationController) GetUnreadCount(ctx *fiber.Ctx) error {
	userIDStr, _ := ctx.Locals("userId").(string)
	userID, _ := uuid.Parse(userIDStr)

	count, err := c.service.GetUnreadCount(ctx.Context(), userID)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Jumlah notifikasi belum dibaca berhasil diambil", map[string]int64{
		"count": count,
	})
}

// MarkAsRead godoc
// @Summary     Tandai notifikasi sudah dibaca
// @Description Menandai satu notifikasi sudah dibaca berdasarkan ID.
// @Tags        Notification
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Notification ID (UUID)"
// @Success     200 {object} network.Response "Notifikasi ditandai sudah dibaca"
// @Failure     400 {object} network.ErrorResponse "ID tidak valid"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Failure     403 {object} network.ErrorResponse "Forbidden"
// @Failure     404 {object} network.ErrorResponse "Notifikasi tidak ditemukan"
// @Router      /api/notifications/{id}/read [put]
func (c *notificationController) MarkAsRead(ctx *fiber.Ctx) error {
	notifID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	userIDStr, _ := ctx.Locals("userId").(string)
	userID, _ := uuid.Parse(userIDStr)

	if err := c.service.MarkAsRead(ctx.Context(), notifID, userID); err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessMsgResponse("Notifikasi ditandai sudah dibaca")
}

// MarkAllAsRead godoc
// @Summary     Tandai semua notifikasi sudah dibaca
// @Description Menandai semua notifikasi milik user yang login sebagai sudah dibaca.
// @Tags        Notification
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} network.Response "Semua notifikasi ditandai sudah dibaca"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Router      /api/notifications/read-all [put]
func (c *notificationController) MarkAllAsRead(ctx *fiber.Ctx) error {
	userIDStr, _ := ctx.Locals("userId").(string)
	userID, _ := uuid.Parse(userIDStr)

	if err := c.service.MarkAllAsRead(ctx.Context(), userID); err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessMsgResponse("Semua notifikasi ditandai sudah dibaca")
}
