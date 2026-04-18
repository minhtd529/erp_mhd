package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/reporting/domain"
)

type DashboardUseCase struct {
	repo domain.ReportingRepository
}

func NewDashboardUseCase(repo domain.ReportingRepository) *DashboardUseCase {
	return &DashboardUseCase{repo: repo}
}

// ExecutiveDashboard returns company-wide KPIs for FIRM_PARTNER view.
func (uc *DashboardUseCase) ExecutiveDashboard(ctx context.Context) (*domain.ExecutiveDashboard, error) {
	now := time.Now().UTC()
	year, month := now.Year(), int(now.Month())

	revenueYTD, err := uc.repo.GetRevenueYTD(ctx, year)
	if err != nil {
		return nil, err
	}
	revenueMonth, err := uc.repo.GetRevenueMonth(ctx, year, month)
	if err != nil {
		return nil, err
	}
	outstanding, err := uc.repo.GetTotalOutstandingReceivables(ctx)
	if err != nil {
		return nil, err
	}
	activeEng, err := uc.repo.GetActiveEngagementsCount(ctx)
	if err != nil {
		return nil, err
	}
	engByStatus, err := uc.repo.GetEngagementsByStatus(ctx)
	if err != nil {
		return nil, err
	}
	avgUtil, err := uc.repo.GetAvgUtilizationMonth(ctx, year, month)
	if err != nil {
		return nil, err
	}
	revByService, err := uc.repo.GetRevenueByService(ctx)
	if err != nil {
		return nil, err
	}
	commKPIs, err := uc.repo.GetCommissionKPIs(ctx, year, month)
	if err != nil {
		return nil, err
	}

	// Commission % of revenue
	if revenueMonth > 0 {
		commKPIs.CommissionPctRevenue = float64(commKPIs.TotalAccruedMonth) / float64(revenueMonth) * 100
	}

	return &domain.ExecutiveDashboard{
		TotalRevenueYTD:        revenueYTD,
		TotalRevenueMonth:      revenueMonth,
		OutstandingReceivables: outstanding,
		ActiveEngagements:      activeEng,
		EngagementsByStatus:    engByStatus,
		AvgUtilizationPercent:  avgUtil,
		RevenueByService:       revByService,
		CommissionKPIs:         *commKPIs,
	}, nil
}

// ManagerDashboard returns team-scoped KPIs for AUDIT_MANAGER.
func (uc *DashboardUseCase) ManagerDashboard(ctx context.Context, managerID uuid.UUID) (*domain.ManagerDashboard, error) {
	now := time.Now().UTC()
	year, month := now.Year(), int(now.Month())

	teamSize, err := uc.repo.GetTeamSize(ctx, managerID)
	if err != nil {
		return nil, err
	}
	teamUtil, err := uc.repo.GetTeamUtilization(ctx, managerID, year, month)
	if err != nil {
		return nil, err
	}
	activeEng, err := uc.repo.GetActiveEngagementsCount(ctx)
	if err != nil {
		return nil, err
	}
	teamReceivables, err := uc.repo.GetTeamOutstandingReceivables(ctx, managerID)
	if err != nil {
		return nil, err
	}
	progress, err := uc.repo.GetTeamEngagementProgress(ctx, managerID)
	if err != nil {
		return nil, err
	}

	return &domain.ManagerDashboard{
		ManagerID:              managerID,
		TeamSize:               teamSize,
		TeamUtilizationPercent: teamUtil,
		ActiveEngagements:      activeEng,
		OutstandingReceivables: teamReceivables,
		EngagementsProgress:    progress,
	}, nil
}

// PersonalDashboard returns individual metrics for AUDIT_STAFF.
func (uc *DashboardUseCase) PersonalDashboard(ctx context.Context, staffID uuid.UUID) (*domain.PersonalDashboard, error) {
	now := time.Now().UTC()
	year, month := now.Year(), int(now.Month())

	util, err := uc.repo.GetAvgUtilizationMonth(ctx, year, month)
	if err != nil {
		return nil, err
	}
	hours, err := uc.repo.GetStaffHoursMonth(ctx, staffID, year, month)
	if err != nil {
		return nil, err
	}
	activeEng, err := uc.repo.GetStaffActiveEngagements(ctx, staffID)
	if err != nil {
		return nil, err
	}
	isSp, err := uc.repo.IsSalesperson(ctx, staffID)
	if err != nil {
		// non-fatal — might not be an employee record
		isSp = false
	}

	dash := &domain.PersonalDashboard{
		StaffID:            staffID,
		UtilizationPercent: util,
		HoursLoggedMonth:   hours,
		ActiveEngagements:  activeEng,
		IsSalesperson:      isSp,
	}

	if isSp {
		commKPIs, err := uc.repo.GetCommissionByStaff(ctx, staffID, year)
		if err == nil {
			dash.CommissionYTD = commKPIs.TotalAccruedMonth
			dash.CommissionMonth = commKPIs.TotalPaidMonth
			dash.CommissionPending = commKPIs.TotalPending
			dash.CommissionOnHold = commKPIs.TotalOnHold
		}
	}

	return dash, nil
}
