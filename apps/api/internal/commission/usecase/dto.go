package usecase

import (
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/commission/domain"
)

// ── CommissionPlan DTOs ──────────────────────────────────────────────────────

type PlanCreateRequest struct {
	Code         string                 `json:"code" binding:"required"`
	Name         string                 `json:"name" binding:"required"`
	Description  string                 `json:"description"`
	Type         domain.CommissionType  `json:"type" binding:"required"`
	DefaultRate  float64                `json:"default_rate"`
	Tiers        []domain.CommissionTier `json:"tiers"`
	ApplyBase    domain.CommissionBase  `json:"apply_base" binding:"required"`
	TriggerOn    domain.CommissionTrigger `json:"trigger_on" binding:"required"`
	ServiceTypes []string               `json:"service_types"`
}

type PlanUpdateRequest struct {
	Name         string                 `json:"name" binding:"required"`
	Description  string                 `json:"description"`
	DefaultRate  float64                `json:"default_rate"`
	Tiers        []domain.CommissionTier `json:"tiers"`
	ApplyBase    domain.CommissionBase  `json:"apply_base" binding:"required"`
	TriggerOn    domain.CommissionTrigger `json:"trigger_on" binding:"required"`
	ServiceTypes []string               `json:"service_types"`
}

type PlanResponse struct {
	ID           uuid.UUID               `json:"id"`
	Code         string                  `json:"code"`
	Name         string                  `json:"name"`
	Description  string                  `json:"description"`
	Type         domain.CommissionType   `json:"type"`
	DefaultRate  float64                 `json:"default_rate"`
	Tiers        []domain.CommissionTier `json:"tiers"`
	ApplyBase    domain.CommissionBase   `json:"apply_base"`
	TriggerOn    domain.CommissionTrigger `json:"trigger_on"`
	ServiceTypes []string                `json:"service_types"`
	IsActive     bool                    `json:"is_active"`
	CreatedBy    uuid.UUID               `json:"created_by"`
	CreatedAt    time.Time               `json:"created_at"`
	UpdatedAt    time.Time               `json:"updated_at"`
}

func toPlanResponse(p *domain.CommissionPlan) PlanResponse {
	return PlanResponse{
		ID:           p.ID,
		Code:         p.Code,
		Name:         p.Name,
		Description:  p.Description,
		Type:         p.Type,
		DefaultRate:  p.DefaultRate,
		Tiers:        p.Tiers,
		ApplyBase:    p.ApplyBase,
		TriggerOn:    p.TriggerOn,
		ServiceTypes: p.ServiceTypes,
		IsActive:     p.IsActive,
		CreatedBy:    p.CreatedBy,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

// ── EngagementCommission DTOs ────────────────────────────────────────────────

type EngCommissionCreateRequest struct {
	EngagementID  uuid.UUID               `json:"engagement_id" binding:"required"`
	SalespersonID uuid.UUID               `json:"salesperson_id" binding:"required"`
	Role          domain.SalesRole        `json:"role" binding:"required"`
	PlanID        *uuid.UUID              `json:"plan_id"`
	RateType      domain.CommissionType   `json:"rate_type" binding:"required"`
	Rate          float64                 `json:"rate"`
	FixedAmount   *int64                  `json:"fixed_amount"`
	Tiers         []domain.CommissionTier `json:"tiers"`
	ApplyBase     domain.CommissionBase   `json:"apply_base" binding:"required"`
	TriggerOn     domain.CommissionTrigger `json:"trigger_on" binding:"required"`
	MaxAmount     *int64                  `json:"max_amount"`
	HoldbackPct   float64                 `json:"holdback_pct"`
	Notes         string                  `json:"notes"`
}

type EngCommissionResponse struct {
	ID            uuid.UUID               `json:"id"`
	EngagementID  uuid.UUID               `json:"engagement_id"`
	SalespersonID uuid.UUID               `json:"salesperson_id"`
	Role          domain.SalesRole        `json:"role"`
	PlanID        *uuid.UUID              `json:"plan_id"`
	RateType      domain.CommissionType   `json:"rate_type"`
	Rate          float64                 `json:"rate"`
	FixedAmount   *int64                  `json:"fixed_amount"`
	Tiers         []domain.CommissionTier `json:"tiers"`
	ApplyBase     domain.CommissionBase   `json:"apply_base"`
	TriggerOn     domain.CommissionTrigger `json:"trigger_on"`
	MaxAmount     *int64                  `json:"max_amount"`
	HoldbackPct   float64                 `json:"holdback_pct"`
	Status        string                  `json:"status"`
	Notes         string                  `json:"notes"`
	ApprovedBy    *uuid.UUID              `json:"approved_by"`
	ApprovedAt    *time.Time              `json:"approved_at"`
	CreatedBy     uuid.UUID               `json:"created_by"`
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at"`
}

func toEngCommissionResponse(ec *domain.EngagementCommission) EngCommissionResponse {
	return EngCommissionResponse{
		ID:            ec.ID,
		EngagementID:  ec.EngagementID,
		SalespersonID: ec.SalespersonID,
		Role:          ec.Role,
		PlanID:        ec.PlanID,
		RateType:      ec.RateType,
		Rate:          ec.Rate,
		FixedAmount:   ec.FixedAmount,
		Tiers:         ec.Tiers,
		ApplyBase:     ec.ApplyBase,
		TriggerOn:     ec.TriggerOn,
		MaxAmount:     ec.MaxAmount,
		HoldbackPct:   ec.HoldbackPct,
		Status:        ec.Status,
		Notes:         ec.Notes,
		ApprovedBy:    ec.ApprovedBy,
		ApprovedAt:    ec.ApprovedAt,
		CreatedBy:     ec.CreatedBy,
		CreatedAt:     ec.CreatedAt,
		UpdatedAt:     ec.UpdatedAt,
	}
}

// ── CommissionRecord DTOs ────────────────────────────────────────────────────

type RecordResponse struct {
	ID                     uuid.UUID              `json:"id"`
	EngagementCommissionID uuid.UUID              `json:"engagement_commission_id"`
	EngagementID           uuid.UUID              `json:"engagement_id"`
	SalespersonID          uuid.UUID              `json:"salesperson_id"`
	InvoiceID              *uuid.UUID             `json:"invoice_id"`
	PaymentID              *uuid.UUID             `json:"payment_id"`
	BaseAmount             int64                  `json:"base_amount"`
	Rate                   float64                `json:"rate"`
	CalculatedAmount       int64                  `json:"calculated_amount"`
	HoldbackAmount         int64                  `json:"holdback_amount"`
	PayableAmount          int64                  `json:"payable_amount"`
	Status                 domain.CommissionStatus `json:"status"`
	AccruedAt              time.Time              `json:"accrued_at"`
	ApprovedBy             *uuid.UUID             `json:"approved_by"`
	ApprovedAt             *time.Time             `json:"approved_at"`
	PaidAt                 *time.Time             `json:"paid_at"`
	PayoutReference        string                 `json:"payout_reference"`
	IsClawback             bool                   `json:"is_clawback"`
	ClawbackReason         string                 `json:"clawback_reason"`
	Notes                  string                 `json:"notes"`
	CreatedAt              time.Time              `json:"created_at"`
}

func toRecordResponse(r *domain.CommissionRecord) RecordResponse {
	return RecordResponse{
		ID:                     r.ID,
		EngagementCommissionID: r.EngagementCommissionID,
		EngagementID:           r.EngagementID,
		SalespersonID:          r.SalespersonID,
		InvoiceID:              r.InvoiceID,
		PaymentID:              r.PaymentID,
		BaseAmount:             r.BaseAmount,
		Rate:                   r.Rate,
		CalculatedAmount:       r.CalculatedAmount,
		HoldbackAmount:         r.HoldbackAmount,
		PayableAmount:          r.PayableAmount,
		Status:                 r.Status,
		AccruedAt:              r.AccruedAt,
		ApprovedBy:             r.ApprovedBy,
		ApprovedAt:             r.ApprovedAt,
		PaidAt:                 r.PaidAt,
		PayoutReference:        r.PayoutReference,
		IsClawback:             r.IsClawback,
		ClawbackReason:         r.ClawbackReason,
		Notes:                  r.Notes,
		CreatedAt:              r.CreatedAt,
	}
}
