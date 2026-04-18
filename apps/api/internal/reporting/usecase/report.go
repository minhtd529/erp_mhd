package usecase

import (
	"context"
	"time"

	"github.com/mdh/erp-audit/api/internal/reporting/domain"
)

type ReportUseCase struct {
	repo domain.ReportingRepository
}

func NewReportUseCase(repo domain.ReportingRepository) *ReportUseCase {
	return &ReportUseCase{repo: repo}
}

func (uc *ReportUseCase) RevenueReport(ctx context.Context, f domain.ReportFilter) ([]domain.RevenueByService, error) {
	return uc.repo.GetRevenueByService(ctx)
}

func (uc *ReportUseCase) UtilizationReport(ctx context.Context, f domain.ReportFilter) ([]domain.UtilizationRate, error) {
	month := time.Now().UTC()
	if f.Year > 0 && f.Month > 0 {
		month = time.Date(f.Year, time.Month(f.Month), 1, 0, 0, 0, 0, time.UTC)
	}
	return uc.repo.GetUtilizationRates(ctx, month)
}

func (uc *ReportUseCase) ARAgingReport(ctx context.Context) ([]domain.ARAgingRow, error) {
	return uc.repo.GetARAgingAll(ctx)
}

func (uc *ReportUseCase) EngagementStatusReport(ctx context.Context, limit int) ([]domain.EngagementProgress, error) {
	if limit <= 0 {
		limit = 50
	}
	return uc.repo.GetEngagementProgress(ctx, limit)
}

func (uc *ReportUseCase) CommissionSummaryReport(ctx context.Context, months int) ([]domain.CommissionMonthlySummary, error) {
	if months <= 0 {
		months = 12
	}
	return uc.repo.GetCommissionMonthlySummary(ctx, months)
}

func (uc *ReportUseCase) RevenueByStaffReport(ctx context.Context, f domain.ReportFilter) ([]domain.RevenueByStaffRow, error) {
	return uc.repo.GetRevenueByStaff(ctx, f)
}

func (uc *ReportUseCase) GetMVRefreshStatus(ctx context.Context) ([]domain.MVRefreshLog, error) {
	return uc.repo.GetLastRefreshLog(ctx)
}

func (uc *ReportUseCase) RefreshMaterializedViews(ctx context.Context) error {
	return uc.repo.RefreshAllViews(ctx)
}

// ── Commission reports ────────────────────────────────────────────────────────

func (uc *ReportUseCase) CommissionPayoutReport(ctx context.Context, months int) ([]domain.CommissionPayoutRow, error) {
	if months <= 0 {
		months = 12
	}
	return uc.repo.GetCommissionPayoutReport(ctx, months)
}

func (uc *ReportUseCase) CommissionByServiceReport(ctx context.Context, year int) ([]domain.CommissionByServiceRow, error) {
	if year <= 0 {
		year = time.Now().Year()
	}
	return uc.repo.GetCommissionByServiceReport(ctx, year)
}

func (uc *ReportUseCase) CommissionPendingReport(ctx context.Context) (*domain.CommissionPendingSummary, error) {
	return uc.repo.GetCommissionPendingReport(ctx)
}

func (uc *ReportUseCase) CommissionClawbackReport(ctx context.Context, months int) (*domain.CommissionClawbackSummary, error) {
	if months <= 0 {
		months = 3
	}
	return uc.repo.GetCommissionClawbackReport(ctx, months)
}
