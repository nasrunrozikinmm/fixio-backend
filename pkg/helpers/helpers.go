package helpers

import "strings"

// Roles constants
const (
	RoleCreator       = "creator"
	RoleModerator     = "moderator"
	RoleAdministrator = "administrator"
)

// Post status constants
const (
	StatusDraft         = "draft"
	StatusPendingReview = "pending_review"
	StatusApproved      = "approved"
	StatusRejected      = "rejected"
)

// Vote type constants
const (
	VoteUp   = "up"
	VoteDown = "down"
)

// OAuth provider constants
const (
	ProviderGoogle = "google"
)

// Region type constants
const (
	RegionNasional = "nasional"
	RegionProvinsi = "provinsi"
	RegionKota     = "kota"
)

// ValidRoles returns all valid role strings
func ValidRoles() []string {
	return []string{RoleCreator, RoleModerator, RoleAdministrator}
}

// IsValidRole checks if a role string is valid
func IsValidRole(role string) bool {
	for _, r := range ValidRoles() {
		if r == strings.ToLower(role) {
			return true
		}
	}
	return false
}

// IsModeratorOrAbove checks if role has moderator or admin privileges
func IsModeratorOrAbove(role string) bool {
	r := strings.ToLower(role)
	return r == RoleModerator || r == RoleAdministrator
}

// IsAdmin checks if role is administrator
func IsAdmin(role string) bool {
	return strings.ToLower(role) == RoleAdministrator
}

// ValidPostStatuses returns all valid post statuses
func ValidPostStatuses() []string {
	return []string{StatusDraft, StatusPendingReview, StatusApproved, StatusRejected}
}

// IsValidPostStatus checks if a post status string is valid
func IsValidPostStatus(status string) bool {
	for _, s := range ValidPostStatuses() {
		if s == strings.ToLower(status) {
			return true
		}
	}
	return false
}

// Slugify creates a URL-safe slug from a string
func Slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "-")
	// Remove non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			result.WriteRune(c)
		}
	}
	return result.String()
}
