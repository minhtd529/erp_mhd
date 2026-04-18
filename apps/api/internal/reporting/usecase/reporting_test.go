package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/reporting/domain"
	"github.com/mdh/erp-audit/api/internal/reporting/usecase"
)

// ── Mock ──────────────────────────────────────────────────────────────────────

type mockRepo struct {
	revenueYTD          int64
	revenueMonth        int64
	outstanding         int64
	activeEng           int64
	engByStatus         map[string]int64
	avgUtil             float64
	revByService        []domain.RevenueByService
	commKPIs            *domain.CommissionKPIs
	teamSize            int64
	teamUtil            float64
	teamProgress        []domain.EngagementProgress
	teamOutstanding     int64
	staffActiveEng      int64
	staffHours          float64
	isSalesperson       bool
	commByStaff         *domain.CommissionKPIs
	revByStaff          []domain.RevenueByStaffRow
	utilRates           []domain.UtilizationRate
	arAging             []domain.ARAgingRow
	arAgingByClient     *domain.ARAgingRow
	engProgress         []domain.EngagementProgress
	commSummary         []domain.CommissionMonthlySummary
	refreshLogs         []domain.MVRefreshLog
	commPayout          []domain.CommissionPayoutRow
	commByService       []domain.CommissionByServiceRow
	commPending         *domain.CommissionPendingSummary
	commClawback        *domain.CommissionClawbackSummary
	errRevenueYTD       error
	errRevenueMonth     error
	errOutstanding      error
	errActiveEng        error
	errCommKPIs         error
	errRefreshAllViews  error
}

func (m *mockRepo) GetRevenueYTD(_ context.Context, _ int) (int64, error) {
	return m.revenueYTD, m.errRevenueYTD
}
func (m *mockRepo) GetRevenueMonth(_ context.Context, _, _ int) (int64, error) {
	return m.revenueMonth, m.errRevenueMonth
}
func (m *mockRepo) GetTotalOutstandingReceivables(_ context.Context) (int64, error) {
	return m.outstanding, m.errOutstanding
}
func (m *mockRepo) GetActiveEngagementsCount(_ context.Context) (int64, error) {
	return m.activeEng, m.errActiveEng
}
func (m *mockRepo) GetEngagementsByStatus(_ context.Context) (map[string]int64, error) {
	return m.engByStatus, nil
}
func (m *mockRepo) GetAvgUtilizationMonth(_ context.Context, _, _ int) (float64, error) {
	return m.avgUtil, nil
}
func (m *mockRepo) GetRevenueByService(_ context.Context) ([]domain.RevenueByService, error) {
	return m.revByService, nil
}
func (m *mockRepo) GetCommissionKPIs(_ context.Context, _, _ int) (*domain.CommissionKPIs, error) {
	return m.commKPIs, m.errCommKPIs
}
func (m *mockRepo) GetCommissionByStaff(_ context.Context, _ uuid.UUID, _ int) (*domain.CommissionKPIs, error) {
	return m.commByStaff, nil
}
func (m *mockRepo) GetTeamSize(_ context.Context, _ uuid.UUID) (int64, error) {
	return m.teamSize, nil
}
func (m *mockRepo) GetTeamUtilization(_ context.Context, _ uuid.UUID, _, _ int) (float64, error) {
	return m.teamUtil, nil
}
func (m *mockRepo) GetTeamEngagementProgress(_ context.Context, _ uuid.UUID) ([]domain.EngagementProgress, error) {
	return m.teamProgress, nil
}
func (m *mockRepo) GetTeamOutstandingReceivables(_ context.Context, _ uuid.UUID) (int64, error) {
	return m.teamOutstanding, nil
}
func (m *mockRepo) GetStaffActiveEngagements(_ context.Context, _ uuid.UUID) (int64, error) {
	return m.staffActiveEng, nil
}
func (m *mockRepo) GetStaffHoursMonth(_ context.Context, _ uuid.UUID, _, _ int) (float64, error) {
	return m.staffHours, nil
}
func (m *mockRepo) IsSalesperson(_ context.Context, _ uuid.UUID) (bool, error) {
	return m.isSalesperson, nil
}
func (m *mockRepo) GetRevenueByStaff(_ context.Context, _ domain.ReportFilter) ([]domain.RevenueByStaffRow, error) {
	return m.revByStaff, nil
}
func (m *mockRepo) GetUtilizationRates(_ context.Context, _ time.Time) ([]domain.UtilizationRate, error) {
	return m.utilRates, nil
}
func (m *mockRepo) GetARAgingAll(_ context.Context) ([]domain.ARAgingRow, error) {
	return m.arAging, nil
}
func (m *mockRepo) GetARAgingByClient(_ context.Context, _ uuid.UUID) (*domain.ARAgingRow, error) {
	return m.arAgingByClient, nil
}
func (m *mockRepo) GetEngagementProgress(_ context.Context, _ int) ([]domain.EngagementProgress, error) {
	return m.engProgress, nil
}
func (m *mockRepo) GetCommissionMonthlySummary(_ context.Context, _ int) ([]domain.CommissionMonthlySummary, error) {
	return m.commSummary, nil
}
func (m *mockRepo) RefreshAllViews(_ context.Context) error { return m.errRefreshAllViews }
func (m *mockRepo) GetLastRefreshLog(_ context.Context) ([]domain.MVRefreshLog, error) {
	return m.refreshLogs, nil
}
func (m *mockRepo) GetCommissionPayoutReport(_ context.Context, _ int) ([]domain.CommissionPayoutRow, error) {
	return m.commPayout, nil
}
func (m *mockRepo) GetCommissionByServiceReport(_ context.Context, _ int) ([]domain.CommissionByServiceRow, error) {
	return m.commByService, nil
}
func (m *mockRepo) GetCommissionPendingReport(_ context.Context) (*domain.CommissionPendingSummary, error) {
	return m.commPending, nil
}
func (m *mockRepo) GetCommissionClawbackReport(_ context.Context, _ int) (*domain.CommissionClawbackSummary, error) {
	return m.commClawback, nil
}

// ── Dashboard tests ───────────────────────────────────────────────────────────

func TestExecutiveDashboard_Happy(t *testing.T) {
	repo := &mockRepo{
		revenueYTD:    1_000_000_00,
		revenueMonth:  100_000_00,
		outstanding:   50_000_00,
		activeEng:     12,
		engByStatus:   map[string]int64{"active": 10, "draft": 2},
		avgUtil:       75.5,
		revByService:  []domain.RevenueByService{{ServiceType: "audit", TotalRevenue: 500_000_00}},
		commKPIs:      &domain.CommissionKPIs{TotalAccruedMonth: 5_000_00, TotalPending: 2_000_00},
	}
	uc := usecase.NewDashboardUseCase(repo)
	dash, err := uc.ExecutiveDashboard(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dash.TotalRevenueYTD != 1_000_000_00 {
		t.Errorf("expected 1_000_000_00, got %d", dash.TotalRevenueYTD)
	}
	if dash.ActiveEngagements != 12 {
		t.Errorf("expected 12, got %d", dash.ActiveEngagements)
	}
	if dash.CommissionKPIs.CommissionPctRevenue == 0 {
		t.Error("expected commission pct revenue to be computed")
	}
	if len(dash.RevenueByService) != 1 {
		t.Errorf("expected 1 service, got %d", len(dash.RevenueByService))
	}
}

func TestExecutiveDashboard_RevenueYTDError(t *testing.T) {
	repo := &mockRepo{errRevenueYTD: errors.New("db error")}
	uc := usecase.NewDashboardUseCase(repo)
	_, err := uc.ExecutiveDashboard(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExecutiveDashboard_CommissionKPIsError(t *testing.T) {
	repo := &mockRepo{
		revenueYTD:   100,
		revenueMonth: 100,
		engByStatus:  map[string]int64{},
		commKPIs:     nil,
		errCommKPIs:  errors.New("db error"),
	}
	uc := usecase.NewDashboardUseCase(repo)
	_, err := uc.ExecutiveDashboard(context.Background())
	if err == nil {
		t.Fatal("expected error from commKPIs")
	}
}

func TestManagerDashboard_Happy(t *testing.T) {
	managerID := uuid.New()
	repo := &mockRepo{
		teamSize:        5,
		teamUtil:        68.0,
		activeEng:       8,
		teamOutstanding: 30_000_00,
		teamProgress: []domain.EngagementProgress{
			{EngagementID: uuid.New(), CompletionPercent: 50},
		},
	}
	uc := usecase.NewDashboardUseCase(repo)
	dash, err := uc.ManagerDashboard(context.Background(), managerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dash.ManagerID != managerID {
		t.Error("manager ID mismatch")
	}
	if dash.TeamSize != 5 {
		t.Errorf("expected team size 5, got %d", dash.TeamSize)
	}
	if len(dash.EngagementsProgress) != 1 {
		t.Errorf("expected 1 engagement, got %d", len(dash.EngagementsProgress))
	}
}

func TestPersonalDashboard_NonSalesperson(t *testing.T) {
	staffID := uuid.New()
	repo := &mockRepo{
		avgUtil:        60.0,
		staffHours:     120.5,
		staffActiveEng: 3,
		isSalesperson:  false,
	}
	uc := usecase.NewDashboardUseCase(repo)
	dash, err := uc.PersonalDashboard(context.Background(), staffID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dash.IsSalesperson {
		t.Error("expected non-salesperson")
	}
	if dash.CommissionYTD != 0 {
		t.Error("expected no commission data for non-salesperson")
	}
}

func TestPersonalDashboard_Salesperson(t *testing.T) {
	staffID := uuid.New()
	repo := &mockRepo{
		avgUtil:        60.0,
		staffHours:     120.5,
		staffActiveEng: 3,
		isSalesperson:  true,
		commByStaff: &domain.CommissionKPIs{
			TotalAccruedMonth: 10_000_00,
			TotalPaidMonth:    8_000_00,
			TotalPending:      2_000_00,
		},
	}
	uc := usecase.NewDashboardUseCase(repo)
	dash, err := uc.PersonalDashboard(context.Background(), staffID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !dash.IsSalesperson {
		t.Error("expected salesperson flag true")
	}
	if dash.CommissionYTD != 10_000_00 {
		t.Errorf("expected commission YTD 10_000_00, got %d", dash.CommissionYTD)
	}
}

// ── Report tests ──────────────────────────────────────────────────────────────

func TestRevenueReport_Happy(t *testing.T) {
	repo := &mockRepo{
		revByService: []domain.RevenueByService{
			{ServiceType: "audit", TotalRevenue: 500_000_00, InvoiceCount: 10},
		},
	}
	uc := usecase.NewReportUseCase(repo)
	rows, err := uc.RevenueReport(context.Background(), domain.ReportFilter{Year: 2026})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}
}

func TestUtilizationReport_Happy(t *testing.T) {
	repo := &mockRepo{
		utilRates: []domain.UtilizationRate{
			{StaffID: uuid.New(), UtilizationPercent: 80.0},
		},
	}
	uc := usecase.NewReportUseCase(repo)
	rows, err := uc.UtilizationReport(context.Background(), domain.ReportFilter{Year: 2026, Month: 4})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}
}

func TestARAgingReport_Happy(t *testing.T) {
	repo := &mockRepo{
		arAging: []domain.ARAgingRow{
			{ClientID: uuid.New(), TotalOutstanding: 50_000_00},
		},
	}
	uc := usecase.NewReportUseCase(repo)
	rows, err := uc.ARAgingReport(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}
}

func TestCommissionSummaryReport_Happy(t *testing.T) {
	repo := &mockRepo{
		commSummary: []domain.CommissionMonthlySummary{
			{Month: time.Now(), TotalAccrued: 5_000_00},
		},
	}
	uc := usecase.NewReportUseCase(repo)
	rows, err := uc.CommissionSummaryReport(context.Background(), 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}
}

func TestRefreshMaterializedViews_Happy(t *testing.T) {
	repo := &mockRepo{}
	uc := usecase.NewReportUseCase(repo)
	if err := uc.RefreshMaterializedViews(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRefreshMaterializedViews_Error(t *testing.T) {
	repo := &mockRepo{errRefreshAllViews: errors.New("pg error")}
	uc := usecase.NewReportUseCase(repo)
	if err := uc.RefreshMaterializedViews(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetMVRefreshStatus_Happy(t *testing.T) {
	repo := &mockRepo{
		refreshLogs: []domain.MVRefreshLog{
			{ViewName: "mv_revenue_by_service", Success: true},
		},
	}
	uc := usecase.NewReportUseCase(repo)
	logs, err := uc.GetMVRefreshStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(logs))
	}
}

// ── Commission report tests ───────────────────────────────────────────────────

func TestCommissionPayoutReport_Happy(t *testing.T) {
	repo := &mockRepo{
		commPayout: []domain.CommissionPayoutRow{
			{Month: time.Now(), TotalApproved: 10_000_00, TotalPaid: 8_000_00, RecordCount: 5},
		},
	}
	uc := usecase.NewReportUseCase(repo)
	rows, err := uc.CommissionPayoutReport(context.Background(), 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}
	if rows[0].RecordCount != 5 {
		t.Errorf("expected record count 5, got %d", rows[0].RecordCount)
	}
}

func TestCommissionPayoutReport_DefaultMonths(t *testing.T) {
	repo := &mockRepo{commPayout: []domain.CommissionPayoutRow{}}
	uc := usecase.NewReportUseCase(repo)
	// months=0 should default to 12
	rows, err := uc.CommissionPayoutReport(context.Background(), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rows == nil {
		rows = []domain.CommissionPayoutRow{}
	}
	// just ensure no panic
}

func TestCommissionByServiceReport_Happy(t *testing.T) {
	repo := &mockRepo{
		commByService: []domain.CommissionByServiceRow{
			{ServiceType: "AUDIT", TotalAccrued: 50_000_00, RecordCount: 10, AvgRate: 0.05},
			{ServiceType: "TAX_ADVISORY", TotalAccrued: 20_000_00, RecordCount: 4, AvgRate: 0.03},
		},
	}
	uc := usecase.NewReportUseCase(repo)
	rows, err := uc.CommissionByServiceReport(context.Background(), 2026)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(rows))
	}
	if rows[0].ServiceType != "AUDIT" {
		t.Errorf("expected first service AUDIT, got %s", rows[0].ServiceType)
	}
}

func TestCommissionByServiceReport_DefaultYear(t *testing.T) {
	repo := &mockRepo{commByService: nil}
	uc := usecase.NewReportUseCase(repo)
	// year=0 should default to current year
	_, err := uc.CommissionByServiceReport(context.Background(), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCommissionPendingReport_Happy(t *testing.T) {
	repo := &mockRepo{
		commPending: &domain.CommissionPendingSummary{
			TotalPendingApproval: 5_000_00,
			TotalPendingPayout:   3_000_00,
			Records: []domain.CommissionPendingRow{
				{RecordID: uuid.New(), Status: "accrued", PayableAmount: 5_000_00},
				{RecordID: uuid.New(), Status: "approved", PayableAmount: 3_000_00},
			},
		},
	}
	uc := usecase.NewReportUseCase(repo)
	summary, err := uc.CommissionPendingReport(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.TotalPendingApproval != 5_000_00 {
		t.Errorf("expected 5_000_00 pending approval, got %d", summary.TotalPendingApproval)
	}
	if len(summary.Records) != 2 {
		t.Errorf("expected 2 records, got %d", len(summary.Records))
	}
}

func TestCommissionClawbackReport_Happy(t *testing.T) {
	repo := &mockRepo{
		commClawback: &domain.CommissionClawbackSummary{
			TotalClawback: 2_500_00,
			RecordCount:   2,
			Records: []domain.CommissionClawbackRow{
				{RecordID: uuid.New(), ClawbackAmount: 1_500_00, Reason: "invoice cancelled"},
				{RecordID: uuid.New(), ClawbackAmount: 1_000_00, Reason: "manual clawback"},
			},
		},
	}
	uc := usecase.NewReportUseCase(repo)
	summary, err := uc.CommissionClawbackReport(context.Background(), 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.TotalClawback != 2_500_00 {
		t.Errorf("expected 2_500_00 total clawback, got %d", summary.TotalClawback)
	}
	if summary.RecordCount != 2 {
		t.Errorf("expected 2 records, got %d", summary.RecordCount)
	}
}

func TestCommissionClawbackReport_DefaultMonths(t *testing.T) {
	repo := &mockRepo{commClawback: &domain.CommissionClawbackSummary{}}
	uc := usecase.NewReportUseCase(repo)
	// months=0 should default to 3
	summary, err := uc.CommissionClawbackReport(context.Background(), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary == nil {
		t.Error("expected non-nil summary")
	}
}
