package services

import (
	models "fixio/internal/modules/sector/entity"
	repositories "fixio/internal/modules/sector/repository"
	"fixio/pkg/network"
)

// SectorService defines sector business operations
type SectorService interface {
	network.CrudService[models.Sector]
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
