package controllers

import (
	"fmt"
	"path/filepath"
	"strings"

	"fixio/pkg/network"
	"fixio/pkg/storage"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	maxFileSize   = 5 << 20 // 5 MB per file
	maxFiles      = 5
	uploadsPrefix = "posts"
)

var allowedMIMETypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

type uploadController struct {
	network.BaseController
	storage *storage.Client
}

// NewUploadController creates a new upload controller
func NewUploadController(
	authFn network.AuthenticationProvider,
	authzFn network.AuthorizationProvider,
	storageClient *storage.Client,
) network.Controller {
	return &uploadController{
		BaseController: network.NewBaseController("/uploads", authFn, authzFn),
		storage:        storageClient,
	}
}

// MountRoutes registers upload routes
func (c *uploadController) MountRoutes(rg *fiber.Group) {
	rg.Post("/images", c.Authentication(), c.UploadImages)
}

// UploadImages godoc
// @Summary     Upload images
// @Description Upload up to 5 images (max 5MB each). Returns array of public URLs.
// @Tags        Uploads
// @Security    BearerAuth
// @Accept      multipart/form-data
// @Produce     json
// @Param       images formData file true "Image files (max 5)"
// @Success     200 {object} network.Response "Images uploaded successfully"
// @Failure     400 {object} network.ErrorResponse "Invalid file or too many files"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Router      /api/uploads/images [post]
func (c *uploadController) UploadImages(ctx *fiber.Ctx) error {
	form, err := ctx.MultipartForm()
	if err != nil {
		return c.Send(ctx).BadRequestError("Failed to parse multipart form", err)
	}

	files := form.File["images"]
	if len(files) == 0 {
		return c.Send(ctx).BadRequestError("No images provided", nil)
	}
	if len(files) > maxFiles {
		return c.Send(ctx).BadRequestError(
			fmt.Sprintf("Maximum %d images allowed", maxFiles), nil,
		)
	}

	var urls []string

	for _, file := range files {
		// Validate file size
		if file.Size > maxFileSize {
			return c.Send(ctx).BadRequestError(
				fmt.Sprintf("File %s exceeds maximum size of 5MB", file.Filename), nil,
			)
		}

		// Validate MIME type
		contentType := file.Header.Get("Content-Type")
		if !allowedMIMETypes[contentType] {
			return c.Send(ctx).BadRequestError(
				fmt.Sprintf("File %s has unsupported type %s. Allowed: JPEG, PNG, WebP, GIF", file.Filename, contentType), nil,
			)
		}

		// Open file
		src, err := file.Open()
		if err != nil {
			return c.Send(ctx).BadRequestError("Failed to read uploaded file", err)
		}

		// Generate unique object name
		ext := strings.ToLower(filepath.Ext(file.Filename))
		objectName := fmt.Sprintf("%s/%s%s", uploadsPrefix, uuid.New().String(), ext)

		// Upload to MinIO
		url, err := c.storage.Upload(ctx.Context(), objectName, contentType, src, file.Size)
		src.Close()
		if err != nil {
			return c.Send(ctx).InternalServerError("Failed to upload image", err)
		}

		urls = append(urls, url)
	}

	return c.Send(ctx).SuccessDataResponse("Images uploaded successfully", fiber.Map{
		"urls": urls,
	})
}
