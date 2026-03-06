package repositories

import (
	models "fixio/internal/modules/sector/entity"
	"fixio/pkg/data"

	"gorm.io/gorm"
)

// SectorRepository defines sector-specific data access operations
type SectorRepository interface {
	data.BaseRepository[models.Sector]
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
