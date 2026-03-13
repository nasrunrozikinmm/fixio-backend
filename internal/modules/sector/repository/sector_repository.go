package repositories

import (
	"context"
	models "fixio/internal/modules/sector/entity"
	"fixio/pkg/data"
	"time"

	"gorm.io/gorm"
)

// TrendingSector holds a sector with its recent post count
type TrendingSector struct {
	models.Sector
	PostCount int `json:"post_count"`
}

// SectorRepository defines sector-specific data access operations
type SectorRepository interface {
	data.BaseRepository[models.Sector]
	GetTrending(ctx context.Context, days int, limit int) ([]TrendingSector, error)
}

type sectorRepository struct {
	data.BaseRepository[models.Sector]
	db *gorm.DB
}

// NewSectorRepository creates a new SectorRepository
func NewSectorRepository(db *gorm.DB) SectorRepository {
	return &sectorRepository{
		BaseRepository: data.NewBaseRepository[models.Sector](db),
		db:             db,
	}
}

// GetTrending returns top sectors by post count within the given time window
func (r *sectorRepository) GetTrending(ctx context.Context, days int, limit int) ([]TrendingSector, error) {
	since := time.Now().AddDate(0, 0, -days)

	var results []TrendingSector
	err := r.db.WithContext(ctx).
		Table("sectors").
		Select("sectors.*, COUNT(posts.id) AS post_count").
		Joins("LEFT JOIN posts ON posts.sector_id = sectors.id AND posts.deleted_at IS NULL AND posts.created_at >= ?", since).
		Where("sectors.deleted_at IS NULL").
		Group("sectors.id").
		Having("COUNT(posts.id) > 0").
		Order("post_count DESC").
		Limit(limit).
		Find(&results).Error

	return results, err
}
