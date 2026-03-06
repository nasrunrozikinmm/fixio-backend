package apperrors

import "errors"

// Sentinel errors — domain-specific errors

// Auth errors
var (
	ErrInvalidToken      = errors.New("invalid or expired token")
	ErrTokenExpired      = errors.New("token has expired")
	ErrUnauthorized      = errors.New("unauthorized access")
	ErrForbidden         = errors.New("forbidden: insufficient permissions")
	ErrInvalidOAuthCode  = errors.New("invalid oauth authorization code")
	ErrOAuthProviderFail = errors.New("failed to get user info from oauth provider")
)

// User errors
var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUserExists       = errors.New("user already exists")
	ErrInvalidRole      = errors.New("invalid role")
	ErrCannotSelfDemote = errors.New("cannot change own role")
)

// Post errors
var (
	ErrPostNotFound        = errors.New("post not found")
	ErrPostNotApproved     = errors.New("post is not approved")
	ErrPostNotOwned        = errors.New("you do not own this post")
	ErrPostAlreadyReviewed = errors.New("post has already been reviewed")
)

// Vote errors
var (
	ErrAlreadyVoted = errors.New("you have already voted on this post")
	ErrVoteNotFound = errors.New("vote not found")
)

// Comment errors
var (
	ErrCommentNotFound = errors.New("comment not found")
	ErrCommentNotOwned = errors.New("you do not own this comment")
	ErrNestedReply     = errors.New("nested replies beyond 1 level are not allowed")
)

// Sector & Region errors
var (
	ErrSectorNotFound = errors.New("sector not found")
	ErrSectorExists   = errors.New("sector already exists")
	ErrRegionNotFound = errors.New("region not found")
	ErrRegionExists   = errors.New("region already exists")
)

// General errors
var (
	ErrRecordNotFound = errors.New("record not found")
	ErrInvalidInput   = errors.New("invalid input")
)
