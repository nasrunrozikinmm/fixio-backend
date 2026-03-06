package repositories

import (
	"context"

	models "fixio/internal/modules/auth/entity"
	"fixio/pkg/data"

	"gorm.io/gorm"
)

// UserRepository defines user-specific data access operations
type UserRepository interface {
	data.BaseRepository[models.User]
	FindByProviderID(ctx context.Context, provider, providerID string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	UpsertOAuthUser(ctx context.Context, user *models.User) (*models.User, error)
}

type userRepository struct {
	data.BaseRepository[models.User]
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		BaseRepository: data.NewBaseRepository[models.User](db),
		db:             db,
	}
}

// FindByProviderID finds a user by OAuth provider and provider ID
func (r *userRepository) FindByProviderID(ctx context.Context, provider, providerID string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Where("provider = ? AND provider_id = ?", provider, providerID).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail finds a user by email
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpsertOAuthUser creates or updates a user from OAuth data
func (r *userRepository) UpsertOAuthUser(ctx context.Context, user *models.User) (*models.User, error) {
	existing, err := r.FindByProviderID(ctx, user.Provider, user.ProviderID)
	if err == nil && existing != nil {
		// Update existing user info
		existing.Name = user.Name
		existing.AvatarURL = user.AvatarURL
		existing.Email = user.Email
		if err := r.db.WithContext(ctx).Save(existing).Error; err != nil {
			return nil, err
		}
		return existing, nil
	}

	// Create new user
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}
