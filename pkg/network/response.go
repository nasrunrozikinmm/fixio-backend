package network

import "github.com/gofiber/fiber/v2"

// Response is the standard API response structure
type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse is the standard error response structure
type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// NewSuccessResponse creates a success response
func NewSuccessResponse(status int, message string, data interface{}) *Response {
	return &Response{
		Status:  status,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(status int, message string) *ErrorResponse {
	return &ErrorResponse{
		Status:  status,
		Message: message,
	}
}

// SendSuccess sends a success JSON response
func SendSuccess(ctx *fiber.Ctx, status int, message string, data interface{}) error {
	return ctx.Status(status).JSON(NewSuccessResponse(status, message, data))
}

// SendError sends an error JSON response
func SendError(ctx *fiber.Ctx, status int, message string) error {
	return ctx.Status(status).JSON(NewErrorResponse(status, message))
}
