package server

import (
	"log"

	authModels "fixio/internal/modules/auth/entity"
	commentModels "fixio/internal/modules/comment/entity"
	postModels "fixio/internal/modules/post/entity"
	regionModels "fixio/internal/modules/region/entity"
	sectorModels "fixio/internal/modules/sector/entity"
	voteModels "fixio/internal/modules/vote/entity"
	"fixio/internal/platform/database"
	redisClient "fixio/internal/platform/redis"
	"fixio/internal/platform/seeder"
	"fixio/pkg/config"
	"fixio/pkg/network"
)

// Server represents the application server
type Server struct {
	App    *network.App
	Module *appCore
	Env    *config.Environment
}

// NewServer creates and bootstraps a new server
func NewServer() *Server {
	log.Println("🚀 Starting Fixio Backend...")

	// 1. Load configuration
	env := config.NewEnvironment()
	log.Printf("📋 Environment: %s", env.Environment)

	// 2. Connect to database
	db := database.NewPostgresDB(env)

	// 3. Run migrations
	database.MigrateDatabase(db,
		&authModels.User{},
		&sectorModels.Sector{},
		&regionModels.Region{},
		&postModels.Post{},
		&voteModels.Vote{},
		&commentModels.Comment{},
	)

	// 4. Connect to Redis
	cacheStore, _ := redisClient.NewRedisClient(env)

	// 5. Run seeders
	seeder.SeedDefaultAdmin(db, env)
	seeder.SeedDefaultSectors(db)
	seeder.SeedDefaultRegions(db)
	seeder.SeedSampleData(db)

	// 6. Create DI module
	module := NewModule(env, db, cacheStore)

	// 7. Create Fiber app
	app := network.NewApp(env.ServerReadTimeout, env.FrontendURL)

	// 8. Mount routes
	app.LoadControllersAndMountRoutes(module.Controllers())

	// 9. Setup Swagger
	app.LoadSwagger()

	log.Println("✅ Server bootstrap completed")

	return &Server{
		App:    app,
		Module: module,
		Env:    env,
	}
}

// Start begins listening for requests
func (s *Server) Start() error {
	log.Printf("🌐 Server starting on %s:%d", s.Env.AppHost, s.Env.AppPort)
	return s.App.Start(s.Env.AppHost, s.Env.AppPort)
}
