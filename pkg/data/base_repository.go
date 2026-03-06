package data

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseRepository provides generic CRUD operations using Go Generics
type BaseRepository[T any] interface {
	Get(ctx context.Context, page, limit int, sort string, filter map[string]any) ([]T, int64, error)
	Create(ctx context.Context, data *T) (*T, error)
	Update(ctx context.Context, id string, data *T) (*T, error)
	Delete(ctx context.Context, id string) error
	SoftDelete(ctx context.Context, id string) error
	FindBy(ctx context.Context, filter map[string]any) (*T, error)
	FindAllBy(ctx context.Context, filter map[string]any) ([]T, error)
	DeleteBy(ctx context.Context, filter map[string]any) error
	UpdateFields(ctx context.Context, id string, fields map[string]any) error
	CountBy(ctx context.Context, filter map[string]any) (int64, error)
	ExistsBy(ctx context.Context, filter map[string]any) (bool, error)
	WithTx(tx Transaction) BaseRepository[T]
	DB() *gorm.DB
}

// baseRepository is the concrete implementation of BaseRepository
type baseRepository[T any] struct {
	db *gorm.DB
}

// NewBaseRepository creates a new BaseRepository instance
func NewBaseRepository[T any](db *gorm.DB) BaseRepository[T] {
	return &baseRepository[T]{db: db}
}

// DB returns the underlying GORM DB instance
func (r *baseRepository[T]) DB() *gorm.DB {
	return r.db
}

// WithTx returns a new repository that uses the given transaction
func (r *baseRepository[T]) WithTx(tx Transaction) BaseRepository[T] {
	return &baseRepository[T]{db: tx.GetTx()}
}

// Get retrieves records with pagination, sorting, and filtering
func (r *baseRepository[T]) Get(ctx context.Context, page, limit int, sort string, filter map[string]any) ([]T, int64, error) {
	var items []T
	var total int64

	query := r.db.WithContext(ctx).Model(new(T))

	// Apply filters
	for key, value := range filter {
		if value != nil && value != "" {
			query = query.Where(fmt.Sprintf("%s = ?", key), value)
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	if sort != "" {
		query = query.Order(sort)
	}

	// Apply pagination
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// Create inserts a new record
func (r *baseRepository[T]) Create(ctx context.Context, data *T) (*T, error) {
	if err := r.db.WithContext(ctx).Create(data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

// Update updates an existing record by ID
func (r *baseRepository[T]) Update(ctx context.Context, id string, data *T) (*T, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	result := r.db.WithContext(ctx).Where("id = ?", uid).Updates(data)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	// Fetch updated record
	var updated T
	if err := r.db.WithContext(ctx).Where("id = ?", uid).First(&updated).Error; err != nil {
		return nil, err
	}
	return &updated, nil
}

// Delete permanently deletes a record by ID
func (r *baseRepository[T]) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	result := r.db.WithContext(ctx).Unscoped().Where("id = ?", uid).Delete(new(T))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// SoftDelete soft-deletes a record by ID
func (r *baseRepository[T]) SoftDelete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	result := r.db.WithContext(ctx).Where("id = ?", uid).Delete(new(T))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// FindBy finds a single record matching the filter
func (r *baseRepository[T]) FindBy(ctx context.Context, filter map[string]any) (*T, error) {
	var item T
	query := r.db.WithContext(ctx)
	for key, value := range filter {
		query = query.Where(fmt.Sprintf("%s = ?", key), value)
	}
	if err := query.First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// FindAllBy finds all records matching the filter
func (r *baseRepository[T]) FindAllBy(ctx context.Context, filter map[string]any) ([]T, error) {
	var items []T
	query := r.db.WithContext(ctx)
	for key, value := range filter {
		query = query.Where(fmt.Sprintf("%s = ?", key), value)
	}
	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// DeleteBy deletes records matching the filter
func (r *baseRepository[T]) DeleteBy(ctx context.Context, filter map[string]any) error {
	query := r.db.WithContext(ctx)
	for key, value := range filter {
		query = query.Where(fmt.Sprintf("%s = ?", key), value)
	}
	return query.Delete(new(T)).Error
}

// UpdateFields updates specific fields of a record by ID
func (r *baseRepository[T]) UpdateFields(ctx context.Context, id string, fields map[string]any) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	result := r.db.WithContext(ctx).Model(new(T)).Where("id = ?", uid).Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CountBy counts records matching the filter
func (r *baseRepository[T]) CountBy(ctx context.Context, filter map[string]any) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(new(T))
	for key, value := range filter {
		query = query.Where(fmt.Sprintf("%s = ?", key), value)
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ExistsBy checks if any record matches the filter
func (r *baseRepository[T]) ExistsBy(ctx context.Context, filter map[string]any) (bool, error) {
	count, err := r.CountBy(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
