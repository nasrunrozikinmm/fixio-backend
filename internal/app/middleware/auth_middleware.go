package middleware

import (
	"strings"

	authService "fixio/internal/modules/auth/usecase"
	"fixio/pkg/helpers"
	"fixio/pkg/network"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware provides JWT authentication and role-based authorization
type AuthMiddleware struct {
	authService authService.AuthService
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(authService authService.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

// Authentication returns a middleware handler that validates JWT tokens
func (m *AuthMiddleware) Authentication() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := extractToken(c)
		if tokenString == "" {
			return network.SendError(c, fiber.StatusUnauthorized, "Token tidak ditemukan. Silakan login terlebih dahulu")
		}

		claims, err := m.authService.ValidateToken(tokenString)
		if err != nil {
			return network.SendError(c, fiber.StatusUnauthorized, "Token tidak valid atau telah kadaluarsa")
		}

		// Set user info in context locals
		c.Locals("userId", claims.UserID.String())
		c.Locals("userRole", claims.Role)
		c.Locals("userEmail", claims.Email)
		c.Locals("userName", claims.Name)

		return c.Next()
	}
}

// Authorization returns a middleware handler that checks user role or permission
func (m *AuthMiddleware) Authorization(roleOrPermission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals("userRole").(string)
		if !ok || userRole == "" {
			return network.SendError(c, fiber.StatusUnauthorized, "Anda belum login")
		}

		// Check if it's a grant permission check
		if strings.HasPrefix(roleOrPermission, "grant:") {
			permission := strings.TrimPrefix(roleOrPermission, "grant:")
			if !hasPermission(userRole, permission) {
				return network.SendError(c, fiber.StatusForbidden, "Anda tidak memiliki akses untuk melakukan aksi ini")
			}
			return c.Next()
		}

		// Role-based check
		if !hasRole(userRole, roleOrPermission) {
			return network.SendError(c, fiber.StatusForbidden, "Anda tidak memiliki akses")
		}

		return c.Next()
	}
}

// extractToken extracts the JWT token from Authorization header or cookie
func extractToken(c *fiber.Ctx) string {
	// Check Authorization header
	auth := c.Get("Authorization")
	if auth != "" {
		parts := strings.Split(auth, " ")
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// Check cookie
	return c.Cookies("access_token")
}

// hasRole checks if user has the required role or higher
func hasRole(userRole, requiredRole string) bool {
	roleHierarchy := map[string]int{
		helpers.RoleCreator:       1,
		helpers.RoleModerator:     2,
		helpers.RoleAdministrator: 3,
	}

	userLevel := roleHierarchy[strings.ToLower(userRole)]
	requiredLevel := roleHierarchy[strings.ToLower(requiredRole)]

	return userLevel >= requiredLevel
}

// hasPermission checks if a role has a specific permission
func hasPermission(role, permission string) bool {
	permissions := getPermissionsForRole(role)
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// getPermissionsForRole returns the permissions for a given role
func getPermissionsForRole(role string) []string {
	creatorPerms := []string{
		"post:create", "post:edit_own", "post:delete_own",
		"comment:create", "comment:delete_own",
		"vote:create", "vote:delete",
	}

	modPerms := append(creatorPerms,
		"moderation:view_queue", "moderation:approve", "moderation:reject",
		"moderation:view_history", "post:hide",
	)

	adminPerms := append(modPerms,
		"admin:manage_users", "admin:manage_roles",
		"admin:manage_sectors", "admin:manage_regions",
		"admin:view_stats", "post:delete_any", "comment:delete_any",
	)

	switch strings.ToLower(role) {
	case helpers.RoleAdministrator:
		return adminPerms
	case helpers.RoleModerator:
		return modPerms
	default:
		return creatorPerms
	}
}
