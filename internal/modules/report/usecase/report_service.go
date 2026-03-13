package services

import (
	"context"
	"math"

	models "fixio/internal/modules/report/entity"
	repositories "fixio/internal/modules/report/repository"
	"fixio/pkg/apperrors"
	"fixio/pkg/network"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReportService defines report business operations
type ReportService interface {
	Create(ctx context.Context, reporterID uuid.UUID, targetType, targetID, reason, description string) (*models.Report, error)
	GetPending(ctx context.Context, pagination network.Pagination) (*network.PaginatedResult[models.Report], error)
	GetAll(ctx context.Context, status string, pagination network.Pagination) (*network.PaginatedResult[models.Report], error)
	Review(ctx context.Context, reportID, reviewerID uuid.UUID, status, reviewNote string) (*models.Report, error)
	CountPending(ctx context.Context) (int64, error)
}

type reportService struct {
	reportRepo repositories.ReportRepository
}

// NewReportService creates a new ReportService
func NewReportService(reportRepo repositories.ReportRepository) ReportService {
	return &reportService{reportRepo: reportRepo}
}

func (s *reportService) Create(ctx context.Context, reporterID uuid.UUID, targetType, targetID, reason, description string) (*models.Report, error) {
	tid, err := uuid.Parse(targetID)
	if err != nil {
		return nil, apperrors.NewBadRequest("ID target tidak valid", err)
	}

	// Check duplicate pending report
	existing, err := s.reportRepo.FindByReporterAndTarget(ctx, reporterID, tid, targetType)
	if err == nil && existing != nil {
		return nil, apperrors.NewBadRequest("Anda sudah melaporkan konten ini", nil)
	}

	report := &models.Report{
		ReporterID:  reporterID,
		TargetType:  targetType,
		TargetID:    tid,
		Reason:      reason,
		Description: description,
	}

	created, err := s.reportRepo.Create(ctx, report)
	if err != nil {
		return nil, apperrors.NewInternal("Gagal membuat laporan", err)
	}
	return created, nil
}

func (s *reportService) GetPending(ctx context.Context, pagination network.Pagination) (*network.PaginatedResult[models.Report], error) {
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.Limit < 1 || pagination.Limit > 100 {
		pagination.Limit = 20
	}

	reports, total, err := s.reportRepo.GetPending(ctx, pagination.Page, pagination.Limit)
	if err != nil {
		return nil, apperrors.NewInternal("Gagal mengambil laporan", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pagination.Limit)))
	return &network.PaginatedResult[models.Report]{
		Data: reports,
		Pagination: network.PaginationMeta{
			Page:       pagination.Page,
			Limit:      pagination.Limit,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    pagination.Page < totalPages,
			HasPrev:    pagination.Page > 1,
		},
	}, nil
}

func (s *reportService) GetAll(ctx context.Context, status string, pagination network.Pagination) (*network.PaginatedResult[models.Report], error) {
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.Limit < 1 || pagination.Limit > 100 {
		pagination.Limit = 20
	}

	reports, total, err := s.reportRepo.GetAll(ctx, status, pagination.Page, pagination.Limit)
	if err != nil {
		return nil, apperrors.NewInternal("Gagal mengambil laporan", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pagination.Limit)))
	return &network.PaginatedResult[models.Report]{
		Data: reports,
		Pagination: network.PaginationMeta{
			Page:       pagination.Page,
			Limit:      pagination.Limit,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    pagination.Page < totalPages,
			HasPrev:    pagination.Page > 1,
		},
	}, nil
}

func (s *reportService) Review(ctx context.Context, reportID, reviewerID uuid.UUID, status, reviewNote string) (*models.Report, error) {
	report, err := s.reportRepo.FindBy(ctx, map[string]any{"id": reportID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewNotFound("Laporan tidak ditemukan", err)
		}
		return nil, apperrors.NewInternal("Gagal mencari laporan", err)
	}

	if report.Status != models.ReportStatusPending {
		return nil, apperrors.NewBadRequest("Laporan sudah di-review sebelumnya", nil)
	}

	report.Status = status
	report.ReviewedBy = &reviewerID
	report.ReviewNote = reviewNote

	updated, err := s.reportRepo.Update(ctx, reportID.String(), report)
	if err != nil {
		return nil, apperrors.NewInternal("Gagal mengupdate laporan", err)
	}
	return updated, nil
}

func (s *reportService) CountPending(ctx context.Context) (int64, error) {
	return s.reportRepo.CountPending(ctx)
}
