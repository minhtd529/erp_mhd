package domain

import (
	"context"

	"github.com/google/uuid"
)

type PlanRepository interface {
	Create(ctx context.Context, p CreatePlanParams) (*CommissionPlan, error)
	FindByID(ctx context.Context, id uuid.UUID) (*CommissionPlan, error)
	FindByCode(ctx context.Context, code string) (*CommissionPlan, error)
	Update(ctx context.Context, p UpdatePlanParams) (*CommissionPlan, error)
	Deactivate(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) (*CommissionPlan, error)
	List(ctx context.Context, f ListPlansFilter, page, size int) ([]*CommissionPlan, int64, error)
}

type EngCommissionRepository interface {
	Create(ctx context.Context, p CreateEngCommissionParams) (*EngagementCommission, error)
	FindByID(ctx context.Context, id uuid.UUID) (*EngagementCommission, error)
	List(ctx context.Context, f ListEngCommissionsFilter, page, size int) ([]*EngagementCommission, int64, error)
	SumRateByEngagement(ctx context.Context, engagementID uuid.UUID) (float64, error)
	Cancel(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) (*EngagementCommission, error)
	Approve(ctx context.Context, id uuid.UUID, approvedBy uuid.UUID) (*EngagementCommission, error)
	ListActiveByTrigger(ctx context.Context, engagementID uuid.UUID, trigger CommissionTrigger) ([]*EngagementCommission, error)
	SumHoldbackByEngagement(ctx context.Context, engagementID uuid.UUID) (int64, error)
}

// BillingDataReader provides the commission accrual engine with read-only
// access to billing tables without coupling commission to the billing usecase.
type BillingDataReader interface {
	GetInvoiceForAccrual(ctx context.Context, invoiceID uuid.UUID) (*InvoiceAccrualData, error)
	GetPaymentForAccrual(ctx context.Context, paymentID uuid.UUID) (*PaymentAccrualData, error)
}

type ListRecordsFilter struct {
	SalespersonID *uuid.UUID
	EngagementID  *uuid.UUID
	Status        CommissionStatus
}

// StatementFilter carries date-range parameters for commission statement queries.
type StatementFilter struct {
	SalespersonID uuid.UUID
	From          string // RFC3339 date
	To            string // RFC3339 date
}

type RecordRepository interface {
	Create(ctx context.Context, r *CommissionRecord) (*CommissionRecord, error)
	FindByID(ctx context.Context, id uuid.UUID) (*CommissionRecord, error)
	List(ctx context.Context, f ListRecordsFilter, page, size int) ([]*CommissionRecord, int64, error)
	ListByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]*CommissionRecord, error)
	ListForStatement(ctx context.Context, f StatementFilter) ([]*CommissionRecord, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status CommissionStatus, approvedBy *uuid.UUID, payoutRef string) (*CommissionRecord, error)
	BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status CommissionStatus, approvedBy *uuid.UUID, payoutRef string) (int64, error)
	SummarySalesperson(ctx context.Context, salespersonID uuid.UUID) (*SalespersonSummary, error)
	ListByTeam(ctx context.Context, managerID uuid.UUID, page, size int) ([]*CommissionRecord, int64, error)
}

// SalespersonSummary aggregates commission data for a single salesperson.
type SalespersonSummary struct {
	TotalYTD      int64 `json:"total_ytd"`
	TotalMonth    int64 `json:"total_month"`
	TotalPending  int64 `json:"total_pending"`
	TotalOnHold   int64 `json:"total_on_hold"`
	TotalApproved int64 `json:"total_approved"`
	TotalPaid     int64 `json:"total_paid"`
}
