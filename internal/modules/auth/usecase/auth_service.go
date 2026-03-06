package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"fixio/internal/modules/auth/dto"
	models "fixio/internal/modules/auth/entity"
	repositories "fixio/internal/modules/auth/repository"
	"fixio/pkg/apperrors"
	"fixio/pkg/cache"
	"fixio/pkg/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AuthService defines authentication business operations
type AuthService interface {
	HandleOAuthCallback(ctx context.Context, userInfo *dto.OAuthUserInfo) (*dto.AuthResponse, error)
	GetCurrentUser(ctx context.Context, userID uuid.UUID) (*dto.UserProfile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req *dto.UpdateProfileRequest) (*dto.UserProfile, error)
	GenerateToken(user *models.User) (string, error)
	ValidateToken(tokenString string) (*JWTClaims, error)
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	Email  string    `json:"email"`
	Name   string    `json:"name"`
	jwt.RegisteredClaims
}

type authService struct {
	userRepo repositories.UserRepository
	env      *config.Environment
	cache    cache.Cache
}

// NewAuthService creates a new AuthService
func NewAuthService(userRepo repositories.UserRepository, env *config.Environment, cache cache.Cache) AuthService {
	return &authService{
		userRepo: userRepo,
		env:      env,
		cache:    cache,
	}
}

// HandleOAuthCallback processes the OAuth callback and returns auth response
func (s *authService) HandleOAuthCallback(ctx context.Context, userInfo *dto.OAuthUserInfo) (*dto.AuthResponse, error) {
	if userInfo == nil {
		return nil, apperrors.NewBadRequest("User info is required", nil)
	}

	// Upsert user
	user := &models.User{
		Name:       userInfo.Name,
		Email:      userInfo.Email,
		AvatarURL:  userInfo.AvatarURL,
		Provider:   userInfo.Provider,
		ProviderID: userInfo.ID,
		Role:       "creator", // Default role for new users
	}

	savedUser, err := s.userRepo.UpsertOAuthUser(ctx, user)
	if err != nil {
		return nil, apperrors.NewInternal("Failed to save user", err)
	}

	// Generate JWT
	token, err := s.GenerateToken(savedUser)
	if err != nil {
		return nil, apperrors.NewInternal("Failed to generate token", err)
	}

	return &dto.AuthResponse{
		Token: token,
		User:  userToProfile(savedUser),
	}, nil
}

// GetCurrentUser retrieves the current user's profile
func (s *authService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*dto.UserProfile, error) {
	user, err := s.userRepo.FindBy(ctx, map[string]any{"id": userID})
	if err != nil {
		return nil, apperrors.NewNotFound("User tidak ditemukan", err)
	}
	profile := userToProfile(user)
	return &profile, nil
}

// UpdateProfile updates a user's profile
func (s *authService) UpdateProfile(ctx context.Context, userID uuid.UUID, req *dto.UpdateProfileRequest) (*dto.UserProfile, error) {
	user, err := s.userRepo.FindBy(ctx, map[string]any{"id": userID})
	if err != nil {
		return nil, apperrors.NewNotFound("User tidak ditemukan", err)
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}
	if req.Location != "" {
		user.Location = req.Location
	}

	updated, err := s.userRepo.Update(ctx, userID.String(), user)
	if err != nil {
		return nil, apperrors.NewInternal("Gagal update profil", err)
	}

	profile := userToProfile(updated)
	return &profile, nil
}

// GenerateToken creates a JWT token for a user
func (s *authService) GenerateToken(user *models.User) (string, error) {
	expiryDays := s.env.JwtTokenExpiryDays
	if expiryDays == 0 {
		expiryDays = 7
	}

	claims := &JWTClaims{
		UserID: user.ID,
		Role:   user.Role,
		Email:  user.Email,
		Name:   user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiryDays) * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "fixio",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.env.JwtSecret))
}

// ValidateToken validates a JWT token and returns the claims
func (s *authService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.env.JwtSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperrors.NewUnauthorized("Token telah kadaluarsa", err)
		}
		return nil, apperrors.NewUnauthorized("Token tidak valid", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, apperrors.NewUnauthorized("Token tidak valid", nil)
	}

	return claims, nil
}

// userToProfile converts a User entity to UserProfile DTO
func userToProfile(user *models.User) dto.UserProfile {
	return dto.UserProfile{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		Bio:       user.Bio,
		Location:  user.Location,
		Role:      user.Role,
		Provider:  user.Provider,
	}
}
