package services

import (
	"context"
	models "fixio/internal/modules/sector/entity"
	repositories "fixio/internal/modules/sector/repository"
	"fixio/pkg/network"
)

// SectorService defines sector business operations
type SectorService interface {
	network.CrudService[models.Sector]
	GetTrending(ctx context.Context, days int, limit int) ([]repositories.TrendingSector, error)
}

type sectorService struct {
	network.CrudService[models.Sector]
	repo repositories.SectorRepository
}

// NewSectorService creates a new SectorService
func NewSectorService(repo repositories.SectorRepository) SectorService {
	return &sectorService{
		CrudService: network.NewCrudServiceWithRepo[models.Sector](repo),
		repo:        repo,
	}
}

// GetTrending returns top sectors by post count within the given time window
func (s *sectorService) GetTrending(ctx context.Context, days int, limit int) ([]repositories.TrendingSector, error) {
	return s.repo.GetTrending(ctx, days, limit)
}
