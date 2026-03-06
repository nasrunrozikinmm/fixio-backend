package network

import (
	"context"
	"math"

	"fixio/pkg/data"
)

// crudService provides a generic CRUD service implementation
type crudService[T any] struct {
	repo data.BaseRepository[T]
}

// NewCrudServiceWithRepo creates a new CrudService with the provided repository
func NewCrudServiceWithRepo[T any](repo data.BaseRepository[T]) CrudService[T] {
	return &crudService[T]{repo: repo}
}

// GetAll retrieves all records with pagination and optional filters
func (s *crudService[T]) GetAll(ctx context.Context, pagination Pagination, filter map[string]any) (*PaginatedResult[T], error) {
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.Limit < 1 {
		pagination.Limit = 20
	}
	if pagination.Limit > 100 {
		pagination.Limit = 100
	}

	items, total, err := s.repo.Get(ctx, pagination.Page, pagination.Limit, pagination.Sort, filter)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(pagination.Limit)))

	return &PaginatedResult[T]{
		Data: items,
		Pagination: PaginationMeta{
			Page:       pagination.Page,
			Limit:      pagination.Limit,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    pagination.Page < totalPages,
			HasPrev:    pagination.Page > 1,
		},
	}, nil
}

// GetOne retrieves a single record by filter
func (s *crudService[T]) GetOne(ctx context.Context, filter map[string]any) (*T, error) {
	return s.repo.FindBy(ctx, filter)
}

// Create creates a new record
func (s *crudService[T]) Create(ctx context.Context, data *T) (*T, error) {
	return s.repo.Create(ctx, data)
}

// Update updates an existing record by ID
func (s *crudService[T]) Update(ctx context.Context, id string, data *T) (*T, error) {
	return s.repo.Update(ctx, id, data)
}

// Delete soft-deletes a record by ID
func (s *crudService[T]) Delete(ctx context.Context, id string) error {
	return s.repo.SoftDelete(ctx, id)
}
