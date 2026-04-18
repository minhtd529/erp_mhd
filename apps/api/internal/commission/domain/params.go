package domain

import "github.com/google/uuid"

type CreatePlanParams struct {
	Code         string
	Name         string
	Description  string
	Type         CommissionType
	DefaultRate  float64
	Tiers        []CommissionTier
	ApplyBase    CommissionBase
	TriggerOn    CommissionTrigger
	ServiceTypes []string
	CreatedBy    uuid.UUID
}

type UpdatePlanParams struct {
	ID          uuid.UUID
	Name        string
	Description string
	DefaultRate float64
	Tiers       []CommissionTier
	ApplyBase   CommissionBase
	TriggerOn   CommissionTrigger
	ServiceTypes []string
	UpdatedBy   uuid.UUID
}

type ListPlansFilter struct {
	IsActive *bool
	Type     CommissionType
}

type CreateEngCommissionParams struct {
	EngagementID  uuid.UUID
	SalespersonID uuid.UUID
	Role          SalesRole
	PlanID        *uuid.UUID
	RateType      CommissionType
	Rate          float64
	FixedAmount   *int64
	Tiers         []CommissionTier
	ApplyBase     CommissionBase
	TriggerOn     CommissionTrigger
	MaxAmount     *int64
	HoldbackPct   float64
	Notes         string
	CreatedBy     uuid.UUID
}

type ListEngCommissionsFilter struct {
	EngagementID  *uuid.UUID
	SalespersonID *uuid.UUID
	Status        string
}
