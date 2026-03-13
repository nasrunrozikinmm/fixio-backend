package network

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	swagger "github.com/swaggo/fiber-swagger"
)

// App is the main application wrapper around Fiber
type App struct {
	fiber *fiber.App
}

// NewApp creates a new App with default Fiber configuration
func NewApp(readTimeout int, frontendURL string) *App {
	app := fiber.New(fiber.Config{
		ReadTimeout:  time.Duration(readTimeout) * time.Second,
		WriteTimeout: time.Duration(readTimeout) * time.Second,
		BodyLimit:    25 * 1024 * 1024, // 25 MB for image uploads
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return ctx.Status(code).JSON(ErrorResponse{
				Status:  code,
				Message: err.Error(),
			})
		},
	})

	// Built-in middlewares
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${method} | ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     frontendURL,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	return &App{fiber: app}
}

// LoadRootMiddlewares applies root-level middlewares to the Fiber app
func (a *App) LoadRootMiddlewares(middlewares []RootMiddleware) {
	for _, mw := range middlewares {
		a.fiber.Use(mw.Handle())
	}
}

// LoadControllersAndMountRoutes mounts all controller routes under /api prefix
func (a *App) LoadControllersAndMountRoutes(controllers []Controller) {
	api := a.fiber.Group("/api")
	for _, ctrl := range controllers {
		group := api.Group(ctrl.Path()).(*fiber.Group)
		ctrl.MountRoutes(group)
	}
}

// LoadSwagger sets up swagger documentation endpoint
func (a *App) LoadSwagger() {
	a.fiber.Get("/swagger/*", swagger.WrapHandler)
}

// Start begins listening on the specified address
func (a *App) Start(ip string, port uint16) error {
	addr := fmt.Sprintf("%s:%d", ip, port)
	return a.fiber.Listen(addr)
}

// GetApp returns the underlying Fiber app instance
func (a *App) GetApp() *fiber.App {
	return a.fiber
}

// Shutdown gracefully shuts down the server
func (a *App) Shutdown() error {
	return a.fiber.Shutdown()
}
