package controllers

import (
	services "fixio/internal/modules/follow/usecase"
	"fixio/pkg/network"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type followController struct {
	network.BaseController
	service services.FollowService
}

// NewFollowController creates a new follow controller
func NewFollowController(
	authFn network.AuthenticationProvider,
	authzFn network.AuthorizationProvider,
	service services.FollowService,
) network.Controller {
	return &followController{
		BaseController: network.NewBaseController("/users", authFn, authzFn),
		service:        service,
	}
}

// MountRoutes registers all follow routes
func (c *followController) MountRoutes(rg *fiber.Group) {
	rg.Post("/:id/follow", c.Authentication(), c.FollowUser)
	rg.Delete("/:id/follow", c.Authentication(), c.UnfollowUser)
	rg.Get("/:id/follow/status", c.Authentication(), c.GetFollowStatus)
	rg.Get("/:id/followers/count", c.GetFollowCounts)
}

// FollowUser handles POST /users/:id/follow
func (c *followController) FollowUser(ctx *fiber.Ctx) error {
	targetID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	userIDStr, _ := ctx.Locals("userId").(string)
	userID, _ := uuid.Parse(userIDStr)

	follow, err := c.service.Follow(ctx.Context(), userID, targetID)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessCreatedResponse("Berhasil follow user", follow)
}

// UnfollowUser handles DELETE /users/:id/follow
func (c *followController) UnfollowUser(ctx *fiber.Ctx) error {
	targetID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	userIDStr, _ := ctx.Locals("userId").(string)
	userID, _ := uuid.Parse(userIDStr)

	if err := c.service.Unfollow(ctx.Context(), userID, targetID); err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessMsgResponse("Berhasil unfollow user")
}

// GetFollowStatus handles GET /users/:id/follow/status
func (c *followController) GetFollowStatus(ctx *fiber.Ctx) error {
	targetID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	userIDStr, _ := ctx.Locals("userId").(string)
	userID, _ := uuid.Parse(userIDStr)

	isFollowing, err := c.service.IsFollowing(ctx.Context(), userID, targetID)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Status follow berhasil diambil", map[string]bool{
		"is_following": isFollowing,
	})
}

// GetFollowCounts handles GET /users/:id/followers/count
func (c *followController) GetFollowCounts(ctx *fiber.Ctx) error {
	userID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidID, err)
	}

	followers, err := c.service.GetFollowerCount(ctx.Context(), userID)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	following, err := c.service.GetFollowingCount(ctx.Context(), userID)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Jumlah follow berhasil diambil", map[string]int64{
		"follower_count":  followers,
		"following_count": following,
	})
}
