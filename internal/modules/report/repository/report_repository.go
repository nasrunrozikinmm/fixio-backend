package repositories

import (
	"context"

	models "fixio/internal/modules/report/entity"
	"fixio/pkg/data"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReportRepository defines report data access operations
type ReportRepository interface {
	data.BaseRepository[models.Report]
	GetPending(ctx context.Context, page, limit int) ([]models.Report, int64, error)
	GetAll(ctx context.Context, status string, page, limit int) ([]models.Report, int64, error)
	FindByReporterAndTarget(ctx context.Context, reporterID, targetID uuid.UUID, targetType string) (*models.Report, error)
	CountPending(ctx context.Context) (int64, error)
}

type reportRepository struct {
	data.BaseRepository[models.Report]
	db *gorm.DB
}

// NewReportRepository creates a new ReportRepository
func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{
		BaseRepository: data.NewBaseRepository[models.Report](db),
		db:             db,
	}
}

func (r *reportRepository) GetPending(ctx context.Context, page, limit int) ([]models.Report, int64, error) {
	var reports []models.Report
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Report{}).Where("status = ?", models.ReportStatusPending)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := r.db.WithContext(ctx).
		Where("status = ?", models.ReportStatusPending).
		Preload("Reporter").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

func (r *reportRepository) GetAll(ctx context.Context, status string, page, limit int) ([]models.Report, int64, error) {
	var reports []models.Report
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Report{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	fetchQuery := r.db.WithContext(ctx)
	if status != "" {
		fetchQuery = fetchQuery.Where("status = ?", status)
	}

	offset := (page - 1) * limit
	if err := fetchQuery.
		Preload("Reporter").
		Preload("Reviewer").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

func (r *reportRepository) FindByReporterAndTarget(ctx context.Context, reporterID, targetID uuid.UUID, targetType string) (*models.Report, error) {
	var report models.Report
	err := r.db.WithContext(ctx).
		Where("reporter_id = ? AND target_id = ? AND target_type = ? AND status = ?",
			reporterID, targetID, targetType, models.ReportStatusPending).
		First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *reportRepository) CountPending(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Report{}).
		Where("status = ?", models.ReportStatusPending).
		Count(&count).Error
	return count, err
}
