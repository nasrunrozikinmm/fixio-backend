package network

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

// Application represents the main HTTP application wrapper
type Application interface {
	LoadSwagger()
	LoadRootMiddlewares(middlewares []RootMiddleware)
	LoadControllersAndMountRoutes(controllers []Controller)
	Start(ip string, port uint16) error
	GetApp() *fiber.App
}

// RootMiddleware is applied to all routes globally
type RootMiddleware interface {
	Handle() fiber.Handler
}

// Controller represents an HTTP controller for a specific module
type Controller interface {
	BaseControllerInterface
	MountRoutes(group *fiber.Group)
}

// BaseControllerInterface provides common controller capabilities
type BaseControllerInterface interface {
	ResponseSender
	Path() string
	Authentication() fiber.Handler
	Authorization(role string) fiber.Handler
	Grant(permission string) fiber.Handler
}

// ResponseSender provides fluent response methods
type ResponseSender interface {
	Send(ctx *fiber.Ctx) *Sender
}

// AuthenticationProvider is a function that returns an authentication middleware handler
type AuthenticationProvider func() fiber.Handler

// AuthorizationProvider is a function that returns an authorization middleware handler by role
type AuthorizationProvider func(role string) fiber.Handler

// GrantProvider is a function that returns a permission grant middleware handler
type GrantProvider func(permission string) fiber.Handler

// ApplicationCore is the DI container that provides Controllers & Middlewares
type ApplicationCore interface {
	Controllers() []Controller
	RootMiddlewares() []RootMiddleware
}

// Pagination holds pagination parameters
type Pagination struct {
	Page  int    `json:"page" query:"page"`
	Limit int    `json:"limit" query:"limit"`
	Sort  string `json:"sort" query:"sort"`
}

// DefaultPagination returns default pagination values
func DefaultPagination() Pagination {
	return Pagination{
		Page:  1,
		Limit: 20,
		Sort:  "created_at desc",
	}
}

// PaginatedResult holds paginated data with metadata
type PaginatedResult[T any] struct {
	Data       []T            `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// PaginationMeta holds pagination metadata
type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// CrudService provides generic CRUD operations
type CrudService[T any] interface {
	GetAll(ctx context.Context, pagination Pagination, filter map[string]any) (*PaginatedResult[T], error)
	GetOne(ctx context.Context, filter map[string]any) (*T, error)
	Create(ctx context.Context, data *T) (*T, error)
	Update(ctx context.Context, id string, data *T) (*T, error)
	Delete(ctx context.Context, id string) error
}
