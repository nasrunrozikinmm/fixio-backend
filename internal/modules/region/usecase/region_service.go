package services

import (
	"context"

	models "fixio/internal/modules/region/entity"
	repositories "fixio/internal/modules/region/repository"
	"fixio/pkg/network"
)

// RegionService defines region business operations
type RegionService interface {
	network.CrudService[models.Region]
	GetByType(ctx context.Context, regionType string) ([]models.Region, error)
}

type regionService struct {
	network.CrudService[models.Region]
	repo repositories.RegionRepository
}

// NewRegionService creates a new RegionService
func NewRegionService(repo repositories.RegionRepository) RegionService {
	return &regionService{
		CrudService: network.NewCrudServiceWithRepo[models.Region](repo),
		repo:        repo,
	}
}

// GetByType retrieves all regions of a specific type
func (s *regionService) GetByType(ctx context.Context, regionType string) ([]models.Region, error) {
	return s.repo.GetByType(ctx, regionType)
}
