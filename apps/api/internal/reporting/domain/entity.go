package domain

import (
	"time"

	"github.com/google/uuid"
)

// ── Materialized view results ─────────────────────────────────────────────────

type RevenueByService struct {
	ServiceType   string `json:"service_type"`
	InvoiceCount  int64  `json:"invoice_count"`
	TotalRevenue  int64  `json:"total_revenue"`
	TotalTax      int64  `json:"total_tax"`
}

type UtilizationRate struct {
	StaffID            uuid.UUID `json:"staff_id"`
	Month              time.Time `json:"month"`
	TotalHours         float64   `json:"total_hours"`
	UtilizationPercent float64   `json:"utilization_percent"`
}

type ARAgingRow struct {
	ClientID        uuid.UUID `json:"client_id"`
	CurrentAmount   int64     `json:"current_amount"`
	Days1To30       int64     `json:"days_1_30"`
	Days31To60      int64     `json:"days_31_60"`
	Days61To90      int64     `json:"days_61_90"`
	DaysOver90      int64     `json:"days_over_90"`
	TotalOutstanding int64    `json:"total_outstanding"`
}

type EngagementProgress struct {
	EngagementID      uuid.UUID `json:"engagement_id"`
	ClientID          uuid.UUID `json:"client_id"`
	Status            string    `json:"status"`
	BudgetedHours     float64   `json:"budgeted_hours"`
	HoursLogged       float64   `json:"hours_logged"`
	CompletionPercent float64   `json:"completion_percent"`
}

type CommissionMonthlySummary struct {
	Month         time.Time `json:"month"`
	TotalAccrued  int64     `json:"total_accrued"`
	TotalApproved int64     `json:"total_approved"`
	TotalPaid     int64     `json:"total_paid"`
	TotalPending  int64     `json:"total_pending"`
	TotalOnHold   int64     `json:"total_on_hold"`
	ClawbackCount int64     `json:"clawback_count"`
}

// ── Dashboard aggregates ──────────────────────────────────────────────────────

type ExecutiveDashboard struct {
	TotalRevenueYTD          int64                      `json:"total_revenue_ytd"`
	TotalRevenueMonth        int64                      `json:"total_revenue_month"`
	OutstandingReceivables   int64                      `json:"outstanding_receivables"`
	ActiveEngagements        int64                      `json:"active_engagements"`
	EngagementsByStatus      map[string]int64           `json:"engagements_by_status"`
	AvgUtilizationPercent    float64                    `json:"avg_utilization_percent"`
	RevenueByService         []RevenueByService         `json:"revenue_by_service"`
	CommissionKPIs           CommissionKPIs             `json:"commission_kpis"`
	LastRefreshedAt          *time.Time                 `json:"last_refreshed_at"`
}

type CommissionKPIs struct {
	TotalAccruedMonth     int64   `json:"total_accrued_month"`
	TotalPaidMonth        int64   `json:"total_paid_month"`
	TotalPending          int64   `json:"total_pending"`
	TotalOnHold           int64   `json:"total_on_hold"`
	CommissionPctRevenue  float64 `json:"commission_pct_revenue"`
}

type ManagerDashboard struct {
	ManagerID               uuid.UUID `json:"manager_id"`
	TeamSize                int64     `json:"team_size"`
	TeamUtilizationPercent  float64   `json:"team_utilization_percent"`
	ActiveEngagements       int64     `json:"active_engagements"`
	OutstandingReceivables  int64     `json:"outstanding_receivables"`
	EngagementsProgress     []EngagementProgress `json:"engagements_progress"`
}

type PersonalDashboard struct {
	StaffID              uuid.UUID `json:"staff_id"`
	UtilizationPercent   float64   `json:"utilization_percent"`
	HoursLoggedMonth     float64   `json:"hours_logged_month"`
	ActiveEngagements    int64     `json:"active_engagements"`
	IsSalesperson        bool      `json:"is_salesperson"`
	CommissionYTD        int64     `json:"commission_ytd,omitempty"`
	CommissionMonth      int64     `json:"commission_month,omitempty"`
	CommissionPending    int64     `json:"commission_pending,omitempty"`
	CommissionOnHold     int64     `json:"commission_on_hold,omitempty"`
}

// ── Report filter ─────────────────────────────────────────────────────────────

type ReportFilter struct {
	Year     int
	Month    int        // 0 = all months
	From     *time.Time
	To       *time.Time
	StaffID  *uuid.UUID
	ClientID *uuid.UUID
	Format   string // "json", "csv"
}

// MVRefreshLog records the last refresh of a materialized view.
type MVRefreshLog struct {
	ViewName    string     `json:"view_name"`
	RefreshedAt time.Time  `json:"refreshed_at"`
	DurationMs  int        `json:"duration_ms"`
	Success     bool       `json:"success"`
	ErrorMsg    *string    `json:"error_msg"`
}

// ── Commission reports ────────────────────────────────────────────────────────

// CommissionPayoutRow is one month's payout summary (Chi hoa hồng tổng hợp).
type CommissionPayoutRow struct {
	Month         time.Time `json:"month"`
	TotalApproved int64     `json:"total_approved"`
	TotalPaid     int64     `json:"total_paid"`
	RecordCount   int64     `json:"record_count"`
}

// CommissionByServiceRow groups commission by engagement service type.
type CommissionByServiceRow struct {
	ServiceType   string  `json:"service_type"`
	TotalAccrued  int64   `json:"total_accrued"`
	TotalPayable  int64   `json:"total_payable"`
	TotalPaid     int64   `json:"total_paid"`
	RecordCount   int64   `json:"record_count"`
	AvgRate       float64 `json:"avg_rate"`
}

// CommissionPendingRow represents a single pending commission record.
type CommissionPendingRow struct {
	RecordID      uuid.UUID `json:"record_id"`
	SalespersonID uuid.UUID `json:"salesperson_id"`
	EngagementID  uuid.UUID `json:"engagement_id"`
	Status        string    `json:"status"`
	PayableAmount int64     `json:"payable_amount"`
	AccruedAt     time.Time `json:"accrued_at"`
	ApprovedAt    *time.Time `json:"approved_at,omitempty"`
}

// CommissionClawbackRow is one month's clawback summary.
type CommissionClawbackRow struct {
	RecordID       uuid.UUID  `json:"record_id"`
	SalespersonID  uuid.UUID  `json:"salesperson_id"`
	EngagementID   uuid.UUID  `json:"engagement_id"`
	ClawbackAmount int64      `json:"clawback_amount"`
	Reason         string     `json:"reason"`
	ClawbackAt     time.Time  `json:"clawback_at"`
}

// CommissionPendingSummary wraps pending records with aggregate totals.
type CommissionPendingSummary struct {
	TotalPendingApproval int64                  `json:"total_pending_approval"`
	TotalPendingPayout   int64                  `json:"total_pending_payout"`
	Records              []CommissionPendingRow  `json:"records"`
}

// CommissionClawbackSummary wraps clawback records with total.
type CommissionClawbackSummary struct {
	TotalClawback int64                   `json:"total_clawback"`
	RecordCount   int64                   `json:"record_count"`
	Records       []CommissionClawbackRow `json:"records"`
}
