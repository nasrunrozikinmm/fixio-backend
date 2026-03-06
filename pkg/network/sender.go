package network

import (
	"fixio/pkg/apperrors"

	"github.com/gofiber/fiber/v2"
)

// Sender provides a fluent API for sending HTTP responses
type Sender struct {
	ctx *fiber.Ctx
}

// NewSender creates a new Sender instance
func NewSender(ctx *fiber.Ctx) *Sender {
	return &Sender{ctx: ctx}
}

// SuccessDataResponse sends a 200 success response with data
func (s *Sender) SuccessDataResponse(message string, data interface{}) error {
	return s.ctx.Status(fiber.StatusOK).JSON(Response{
		Status:  fiber.StatusOK,
		Message: message,
		Data:    data,
	})
}

// SuccessCreatedResponse sends a 201 created response with data
func (s *Sender) SuccessCreatedResponse(message string, data interface{}) error {
	return s.ctx.Status(fiber.StatusCreated).JSON(Response{
		Status:  fiber.StatusCreated,
		Message: message,
		Data:    data,
	})
}

// SuccessMsgResponse sends a 200 success response with message only
func (s *Sender) SuccessMsgResponse(message string) error {
	return s.ctx.Status(fiber.StatusOK).JSON(Response{
		Status:  fiber.StatusOK,
		Message: message,
	})
}

// NoContentResponse sends a 204 no content response
func (s *Sender) NoContentResponse() error {
	return s.ctx.SendStatus(fiber.StatusNoContent)
}

// BadRequestError sends a 400 bad request error response
func (s *Sender) BadRequestError(message string, err error) error {
	return s.ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
		Status:  fiber.StatusBadRequest,
		Message: message,
	})
}

// UnauthorizedError sends a 401 unauthorized error response
func (s *Sender) UnauthorizedError(message string, err error) error {
	return s.ctx.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
		Status:  fiber.StatusUnauthorized,
		Message: message,
	})
}

// ForbiddenError sends a 403 forbidden error response
func (s *Sender) ForbiddenError(message string, err error) error {
	return s.ctx.Status(fiber.StatusForbidden).JSON(ErrorResponse{
		Status:  fiber.StatusForbidden,
		Message: message,
	})
}

// NotFoundError sends a 404 not found error response
func (s *Sender) NotFoundError(message string, err error) error {
	return s.ctx.Status(fiber.StatusNotFound).JSON(ErrorResponse{
		Status:  fiber.StatusNotFound,
		Message: message,
	})
}

// ConflictError sends a 409 conflict error response
func (s *Sender) ConflictError(message string, err error) error {
	return s.ctx.Status(fiber.StatusConflict).JSON(ErrorResponse{
		Status:  fiber.StatusConflict,
		Message: message,
	})
}

// InternalServerError sends a 500 internal server error response
func (s *Sender) InternalServerError(message string, err error) error {
	return s.ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
		Status:  fiber.StatusInternalServerError,
		Message: message,
	})
}

// HandleError automatically maps error types to appropriate HTTP responses
func (s *Sender) HandleError(err error) error {
	if err == nil {
		return s.InternalServerError("An unexpected error occurred", nil)
	}

	// Check for AppError
	if appErr, ok := err.(*apperrors.AppError); ok {
		status := appErr.Code
		if status == 0 {
			status = fiber.StatusInternalServerError
		}
		return s.ctx.Status(status).JSON(ErrorResponse{
			Status:  status,
			Message: appErr.Message,
		})
	}

	// Default to 500
	return s.InternalServerError(err.Error(), err)
}
