package server

import (
	"log"

	authModels "fixio/internal/modules/auth/entity"
	bookmarkModels "fixio/internal/modules/bookmark/entity"
	commentModels "fixio/internal/modules/comment/entity"
	followModels "fixio/internal/modules/follow/entity"
	notifModels "fixio/internal/modules/notification/entity"
	postModels "fixio/internal/modules/post/entity"
	regionModels "fixio/internal/modules/region/entity"
	sectorModels "fixio/internal/modules/sector/entity"
	voteModels "fixio/internal/modules/vote/entity"
	"fixio/internal/platform/database"
	redisClient "fixio/internal/platform/redis"
	"fixio/internal/platform/seeder"
	"fixio/pkg/config"
	"fixio/pkg/network"
	"fixio/pkg/storage"
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
		&followModels.Follow{},
		&bookmarkModels.Bookmark{},
		&notifModels.Notification{},
	)

	// 4. Connect to Redis
	cacheStore, _ := redisClient.NewRedisClient(env)

	// 5. Run seeders
	seeder.SeedDefaultAdmin(db, env)
	seeder.SeedDefaultSectors(db)
	seeder.SeedDefaultRegions(db)
	seeder.SeedSampleData(db)

	// 6. Connect to MinIO (optional — non-fatal if unavailable)
	var storageClient *storage.Client
	if env.MinioEndpoint != "" && env.MinioAccessKey != "" {
		sc, err := storage.NewMinioClient(env)
		if err != nil {
			log.Printf("⚠️  MinIO not available: %v (image uploads disabled)", err)
		} else {
			storageClient = sc
		}
	}

	// 7. Create DI module
	module := NewModule(env, db, cacheStore, storageClient)

	// 8. Create Fiber app
	app := network.NewApp(env.ServerReadTimeout, env.FrontendURL)

	// 9. Mount routes
	app.LoadControllersAndMountRoutes(module.Controllers())

	// 10. Setup Swagger
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
