package domain

import (
	"time"

	"github.com/google/uuid"
)

type CommissionType string

const (
	CommissionTypeFlat    CommissionType = "flat"
	CommissionTypeTiered  CommissionType = "tiered"
	CommissionTypeFixed   CommissionType = "fixed"
	CommissionTypeCustom  CommissionType = "custom"
)

type CommissionBase string

const (
	CommBaseFeeContracted CommissionBase = "fee_contracted"
	CommBaseFeeInvoiced   CommissionBase = "fee_invoiced"
	CommBaseFeePaid       CommissionBase = "fee_paid"
	CommBaseGrossMargin   CommissionBase = "gross_margin"
)

type CommissionTrigger string

const (
	CommTriggerContractSigned  CommissionTrigger = "contract_signed"
	CommTriggerInvoiceIssued   CommissionTrigger = "invoice_issued"
	CommTriggerPaymentReceived CommissionTrigger = "payment_received"
	CommTriggerEngCompleted    CommissionTrigger = "eng_completed"
)

type SalesRole string

const (
	SalesRolePrimary        SalesRole = "primary"
	SalesRoleReferrer       SalesRole = "referrer"
	SalesRoleAccountManager SalesRole = "account_manager"
	SalesRoleTechnicalLead  SalesRole = "technical_lead"
)

type CommissionStatus string

const (
	CommStatusAccrued   CommissionStatus = "accrued"
	CommStatusApproved  CommissionStatus = "approved"
	CommStatusOnHold    CommissionStatus = "on_hold"
	CommStatusPaid      CommissionStatus = "paid"
	CommStatusClawback  CommissionStatus = "clawback"
	CommStatusCancelled CommissionStatus = "cancelled"
)

type CommissionTier struct {
	MinAmount int64   `json:"min_amount"`
	MaxAmount *int64  `json:"max_amount"`
	Rate      float64 `json:"rate"`
}

type CommissionPlan struct {
	ID           uuid.UUID        `json:"id"`
	Code         string           `json:"code"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Type         CommissionType   `json:"type"`
	DefaultRate  float64          `json:"default_rate"`
	Tiers        []CommissionTier `json:"tiers"`
	ApplyBase    CommissionBase   `json:"apply_base"`
	TriggerOn    CommissionTrigger `json:"trigger_on"`
	ServiceTypes []string         `json:"service_types"`
	IsActive     bool             `json:"is_active"`
	CreatedBy    uuid.UUID        `json:"created_by"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	UpdatedBy    *uuid.UUID       `json:"updated_by"`
}

type EngagementCommission struct {
	ID            uuid.UUID         `json:"id"`
	EngagementID  uuid.UUID         `json:"engagement_id"`
	SalespersonID uuid.UUID         `json:"salesperson_id"`
	Role          SalesRole         `json:"role"`
	PlanID        *uuid.UUID        `json:"plan_id"`
	RateType      CommissionType    `json:"rate_type"`
	Rate          float64           `json:"rate"`
	FixedAmount   *int64            `json:"fixed_amount"`
	Tiers         []CommissionTier  `json:"tiers"`
	ApplyBase     CommissionBase    `json:"apply_base"`
	TriggerOn     CommissionTrigger `json:"trigger_on"`
	MaxAmount     *int64            `json:"max_amount"`
	HoldbackPct   float64           `json:"holdback_pct"`
	Status        string            `json:"status"`
	Notes         string            `json:"notes"`
	ApprovedBy    *uuid.UUID        `json:"approved_by"`
	ApprovedAt    *time.Time        `json:"approved_at"`
	CreatedBy     uuid.UUID         `json:"created_by"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type CommissionRecord struct {
	ID                     uuid.UUID        `json:"id"`
	EngagementCommissionID uuid.UUID        `json:"engagement_commission_id"`
	EngagementID           uuid.UUID        `json:"engagement_id"`
	SalespersonID          uuid.UUID        `json:"salesperson_id"`
	InvoiceID              *uuid.UUID       `json:"invoice_id"`
	PaymentID              *uuid.UUID       `json:"payment_id"`
	BaseAmount             int64            `json:"base_amount"`
	Rate                   float64          `json:"rate"`
	CalculatedAmount       int64            `json:"calculated_amount"`
	HoldbackAmount         int64            `json:"holdback_amount"`
	PayableAmount          int64            `json:"payable_amount"`
	Status                 CommissionStatus `json:"status"`
	AccruedAt              time.Time        `json:"accrued_at"`
	ApprovedBy             *uuid.UUID       `json:"approved_by"`
	ApprovedAt             *time.Time       `json:"approved_at"`
	PaidAt                 *time.Time       `json:"paid_at"`
	PaidByPayrollID        *uuid.UUID       `json:"paid_by_payroll_id"`
	PayoutReference        string           `json:"payout_reference"`
	ClawbackRecordID       *uuid.UUID       `json:"clawback_record_id"`
	IsClawback             bool             `json:"is_clawback"`
	ClawbackReason         string           `json:"clawback_reason"`
	Notes                  string           `json:"notes"`
	CreatedAt              time.Time        `json:"created_at"`
	UpdatedAt              time.Time        `json:"updated_at"`
	UpdatedBy              *uuid.UUID       `json:"updated_by"`
}
