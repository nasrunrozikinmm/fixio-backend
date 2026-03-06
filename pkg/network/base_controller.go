package network

import (
	"github.com/gofiber/fiber/v2"
)

// BaseController provides common controller functionality
type BaseController struct {
	path    string
	authFn  AuthenticationProvider
	authzFn AuthorizationProvider
}

// NewBaseController creates a new BaseController
func NewBaseController(path string, authFn AuthenticationProvider, authzFn AuthorizationProvider) BaseController {
	return BaseController{
		path:    path,
		authFn:  authFn,
		authzFn: authzFn,
	}
}

// Path returns the base path for this controller's routes
func (b *BaseController) Path() string {
	return b.path
}

// Authentication returns the authentication middleware handler
func (b *BaseController) Authentication() fiber.Handler {
	if b.authFn != nil {
		return b.authFn()
	}
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}

// Authorization returns the role-based authorization middleware handler
func (b *BaseController) Authorization(role string) fiber.Handler {
	if b.authzFn != nil {
		return b.authzFn(role)
	}
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}

// Grant returns the permission-based authorization middleware handler
func (b *BaseController) Grant(permission string) fiber.Handler {
	// Permission check via authorization provider with "grant:" prefix
	if b.authzFn != nil {
		return b.authzFn("grant:" + permission)
	}
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}

// Send creates a new fluent response sender for this request context
func (b *BaseController) Send(ctx *fiber.Ctx) *Sender {
	return NewSender(ctx)
}
