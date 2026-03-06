package repositories

import (
	"context"

	models "fixio/internal/modules/region/entity"
	"fixio/pkg/data"

	"gorm.io/gorm"
)

// RegionRepository defines region-specific data access operations
type RegionRepository interface {
	data.BaseRepository[models.Region]
	GetByType(ctx context.Context, regionType string) ([]models.Region, error)
	GetWithChildren(ctx context.Context, page, limit int) ([]models.Region, int64, error)
}

type regionRepository struct {
	data.BaseRepository[models.Region]
	db *gorm.DB
}

// NewRegionRepository creates a new RegionRepository
func NewRegionRepository(db *gorm.DB) RegionRepository {
	return &regionRepository{
		BaseRepository: data.NewBaseRepository[models.Region](db),
		db:             db,
	}
}

// GetByType retrieves all regions of a specific type
func (r *regionRepository) GetByType(ctx context.Context, regionType string) ([]models.Region, error) {
	var regions []models.Region
	err := r.db.WithContext(ctx).
		Where("type = ?", regionType).
		Order("name ASC").
		Find(&regions).Error
	return regions, err
}

// GetWithChildren retrieves regions with their children
func (r *regionRepository) GetWithChildren(ctx context.Context, page, limit int) ([]models.Region, int64, error) {
	var regions []models.Region
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Region{}).Where("parent_id IS NULL")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := r.db.WithContext(ctx).
		Where("parent_id IS NULL").
		Preload("Children").
		Preload("Children.Children").
		Order("name ASC").
		Offset(offset).
		Limit(limit).
		Find(&regions).Error; err != nil {
		return nil, 0, err
	}

	return regions, total, nil
}
