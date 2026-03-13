package server

import (
	"fixio/internal/app/middleware"
	"fixio/pkg/cache"
	"fixio/pkg/config"
	"fixio/pkg/network"
	"fixio/pkg/storage"

	// Repositories
	authRepo "fixio/internal/modules/auth/repository"
	bookmarkRepo "fixio/internal/modules/bookmark/repository"
	commentRepo "fixio/internal/modules/comment/repository"
	followRepo "fixio/internal/modules/follow/repository"
	notifRepo "fixio/internal/modules/notification/repository"
	postRepo "fixio/internal/modules/post/repository"
	regionRepo "fixio/internal/modules/region/repository"
	sectorRepo "fixio/internal/modules/sector/repository"
	voteRepo "fixio/internal/modules/vote/repository"

	// Services
	adminService "fixio/internal/modules/admin/usecase"
	authService "fixio/internal/modules/auth/usecase"
	bookmarkService "fixio/internal/modules/bookmark/usecase"
	commentService "fixio/internal/modules/comment/usecase"
	followService "fixio/internal/modules/follow/usecase"
	modService "fixio/internal/modules/moderation/usecase"
	notifService "fixio/internal/modules/notification/usecase"
	postService "fixio/internal/modules/post/usecase"
	regionService "fixio/internal/modules/region/usecase"
	sectorService "fixio/internal/modules/sector/usecase"
	voteService "fixio/internal/modules/vote/usecase"

	// Controllers
	adminCtrl "fixio/internal/modules/admin/delivery/http"
	authCtrl "fixio/internal/modules/auth/delivery/http"
	bookmarkCtrl "fixio/internal/modules/bookmark/delivery/http"
	commentCtrl "fixio/internal/modules/comment/delivery/http"
	followCtrl "fixio/internal/modules/follow/delivery/http"
	modCtrl "fixio/internal/modules/moderation/delivery/http"
	notifCtrl "fixio/internal/modules/notification/delivery/http"
	postCtrl "fixio/internal/modules/post/delivery/http"
	regionCtrl "fixio/internal/modules/region/delivery/http"
	sectorCtrl "fixio/internal/modules/sector/delivery/http"
	uploadCtrl "fixio/internal/modules/upload/delivery/http"
	voteCtrl "fixio/internal/modules/vote/delivery/http"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// appCore is the DI container that holds all wired dependencies
type appCore struct {
	env            *config.Environment
	authMiddleware *middleware.AuthMiddleware
	storageClient  *storage.Client

	// Services
	AuthService         authService.AuthService
	PostService         postService.PostService
	VoteService         voteService.VoteService
	CommentService      commentService.CommentService
	SectorService       sectorService.SectorService
	RegionService       regionService.RegionService
	ModerationService   modService.ModerationService
	AdminService        adminService.AdminService
	FollowService       followService.FollowService
	BookmarkService     bookmarkService.BookmarkService
	NotificationService notifService.NotificationService
}

// NewModule creates the DI container and wires all dependencies
func NewModule(env *config.Environment, db *gorm.DB, cacheStore cache.Cache, storageClient *storage.Client) *appCore {
	// 1. Initialize Repositories
	userRepository := authRepo.NewUserRepository(db)
	postRepository := postRepo.NewPostRepository(db)
	voteRepository := voteRepo.NewVoteRepository(db)
	commentRepository := commentRepo.NewCommentRepository(db)
	sectorRepository := sectorRepo.NewSectorRepository(db)
	regionRepository := regionRepo.NewRegionRepository(db)
	followRepository := followRepo.NewFollowRepository(db)
	bookmarkRepository := bookmarkRepo.NewBookmarkRepository(db)
	notifRepository := notifRepo.NewNotificationRepository(db)

	// 2. Initialize Services
	authSvc := authService.NewAuthService(userRepository, env, cacheStore)
	postSvc := postService.NewPostService(postRepository)
	voteSvc := voteService.NewVoteService(voteRepository, postRepository)
	commentSvc := commentService.NewCommentService(commentRepository, postRepository)
	sectorSvc := sectorService.NewSectorService(sectorRepository)
	regionSvc := regionService.NewRegionService(regionRepository)
	moderationSvc := modService.NewModerationService(postRepository, userRepository)
	adminSvc := adminService.NewAdminService(userRepository, postRepository, db)
	followSvc := followService.NewFollowService(followRepository, userRepository)
	bookmarkSvc := bookmarkService.NewBookmarkService(bookmarkRepository, postRepository)
	notifSvc := notifService.NewNotificationService(notifRepository)

	// 3. Initialize Auth Middleware
	authMw := middleware.NewAuthMiddleware(authSvc)

	return &appCore{
		env:                 env,
		authMiddleware:      authMw,
		storageClient:       storageClient,
		AuthService:         authSvc,
		PostService:         postSvc,
		VoteService:         voteSvc,
		CommentService:      commentSvc,
		SectorService:       sectorSvc,
		RegionService:       regionSvc,
		ModerationService:   moderationSvc,
		AdminService:        adminSvc,
		FollowService:       followSvc,
		BookmarkService:     bookmarkSvc,
		NotificationService: notifSvc,
	}
}

// AuthenticationProvider returns a function that creates authentication middleware
func (m *appCore) AuthenticationProvider() network.AuthenticationProvider {
	return func() fiber.Handler {
		return m.authMiddleware.Authentication()
	}
}

// AuthorizationProvider returns a function that creates authorization middleware
func (m *appCore) AuthorizationProvider() network.AuthorizationProvider {
	return func(roleOrPermission string) fiber.Handler {
		return m.authMiddleware.Authorization(roleOrPermission)
	}
}

// Controllers returns all wired controllers
func (m *appCore) Controllers() []network.Controller {
	auth := m.AuthenticationProvider()
	authz := m.AuthorizationProvider()

	controllers := []network.Controller{
		authCtrl.NewAuthController(auth, authz, m.AuthService, m.env),
		authCtrl.NewUserController(auth, authz, m.AuthService, m.PostService),
		postCtrl.NewPostController(auth, authz, m.PostService, m.FollowService),
		voteCtrl.NewVoteController(auth, authz, m.VoteService),
		commentCtrl.NewCommentController(auth, authz, m.CommentService),
		commentCtrl.NewCommentDeleteController(auth, authz, m.CommentService),
		sectorCtrl.NewSectorController(auth, authz, m.SectorService),
		regionCtrl.NewRegionController(auth, authz, m.RegionService),
		modCtrl.NewModerationController(auth, authz, m.ModerationService),
		adminCtrl.NewAdminController(auth, authz, m.AdminService),
		followCtrl.NewFollowController(auth, authz, m.FollowService),
		bookmarkCtrl.NewBookmarkController(auth, authz, m.BookmarkService),
		notifCtrl.NewNotificationController(auth, authz, m.NotificationService),
	}

	// Register upload controller only when MinIO storage is available
	if m.storageClient != nil {
		controllers = append(controllers, uploadCtrl.NewUploadController(auth, authz, m.storageClient))
	}

	return controllers
}
