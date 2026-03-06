package data

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// Transaction wraps a GORM transaction
type Transaction interface {
	GetTx() *gorm.DB
	Commit() error
	Rollback() error
}

// transaction is the concrete implementation
type transaction struct {
	tx *gorm.DB
}

func (t *transaction) GetTx() *gorm.DB {
	return t.tx
}

func (t *transaction) Commit() error {
	return t.tx.Commit().Error
}

func (t *transaction) Rollback() error {
	return t.tx.Rollback().Error
}

// UnitOfWork manages database transactions
type UnitOfWork interface {
	Begin(ctx context.Context) Transaction
	WithTransaction(ctx context.Context, fn func(tx Transaction) error) error
}

// unitOfWork is the concrete implementation
type unitOfWork struct {
	db *gorm.DB
}

// NewUnitOfWork creates a new UnitOfWork
func NewUnitOfWork(db *gorm.DB) UnitOfWork {
	return &unitOfWork{db: db}
}

// Begin starts a new transaction
func (u *unitOfWork) Begin(ctx context.Context) Transaction {
	tx := u.db.WithContext(ctx).Begin()
	return &transaction{tx: tx}
}

// WithTransaction executes a function within a transaction
// Auto-commits on success, auto-rollbacks on error or panic
func (u *unitOfWork) WithTransaction(ctx context.Context, fn func(tx Transaction) error) error {
	tx := u.Begin(ctx)

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // Re-panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback error: %v, original error: %w", rbErr, err)
		}
		return err
	}

	return tx.Commit()
}
