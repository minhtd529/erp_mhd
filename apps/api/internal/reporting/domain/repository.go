package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ReportingRepository interface {
	// Materialized view reads
	GetRevenueByService(ctx context.Context) ([]RevenueByService, error)
	GetUtilizationRates(ctx context.Context, month time.Time) ([]UtilizationRate, error)
	GetARAgingAll(ctx context.Context) ([]ARAgingRow, error)
	GetARAgingByClient(ctx context.Context, clientID uuid.UUID) (*ARAgingRow, error)
	GetEngagementProgress(ctx context.Context, limit int) ([]EngagementProgress, error)
	GetCommissionMonthlySummary(ctx context.Context, months int) ([]CommissionMonthlySummary, error)

	// Aggregated metrics for dashboards
	GetRevenueYTD(ctx context.Context, year int) (int64, error)
	GetRevenueMonth(ctx context.Context, year, month int) (int64, error)
	GetTotalOutstandingReceivables(ctx context.Context) (int64, error)
	GetActiveEngagementsCount(ctx context.Context) (int64, error)
	GetEngagementsByStatus(ctx context.Context) (map[string]int64, error)
	GetAvgUtilizationMonth(ctx context.Context, year, month int) (float64, error)

	// Commission KPIs (from commission_records directly for freshness)
	GetCommissionKPIs(ctx context.Context, year, month int) (*CommissionKPIs, error)
	GetCommissionByStaff(ctx context.Context, staffID uuid.UUID, year int) (*CommissionKPIs, error)

	// Manager-scoped
	GetTeamSize(ctx context.Context, managerID uuid.UUID) (int64, error)
	GetTeamUtilization(ctx context.Context, managerID uuid.UUID, year, month int) (float64, error)
	GetTeamEngagementProgress(ctx context.Context, managerID uuid.UUID) ([]EngagementProgress, error)
	GetTeamOutstandingReceivables(ctx context.Context, managerID uuid.UUID) (int64, error)
	GetStaffActiveEngagements(ctx context.Context, staffID uuid.UUID) (int64, error)
	GetStaffHoursMonth(ctx context.Context, staffID uuid.UUID, year, month int) (float64, error)
	IsSalesperson(ctx context.Context, staffID uuid.UUID) (bool, error)

	// Revenue by salesperson report
	GetRevenueByStaff(ctx context.Context, f ReportFilter) ([]RevenueByStaffRow, error)

	// MV refresh management
	RefreshAllViews(ctx context.Context) error
	GetLastRefreshLog(ctx context.Context) ([]MVRefreshLog, error)

	// Commission reports
	GetCommissionPayoutReport(ctx context.Context, months int) ([]CommissionPayoutRow, error)
	GetCommissionByServiceReport(ctx context.Context, year int) ([]CommissionByServiceRow, error)
	GetCommissionPendingReport(ctx context.Context) (*CommissionPendingSummary, error)
	GetCommissionClawbackReport(ctx context.Context, months int) (*CommissionClawbackSummary, error)
}

// RevenueByStaffRow is used in the revenue-by-salesperson report.
type RevenueByStaffRow struct {
	StaffID       uuid.UUID `json:"staff_id"`
	TotalRevenue  int64     `json:"total_revenue"`
	InvoiceCount  int64     `json:"invoice_count"`
	EngagementCount int64   `json:"engagement_count"`
}
