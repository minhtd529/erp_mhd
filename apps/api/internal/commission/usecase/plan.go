package usecase

import (
	"context"
	"net"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/commission/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

type PlanUseCase struct {
	planRepo domain.PlanRepository
	auditLog *audit.Logger
}

func NewPlanUseCase(planRepo domain.PlanRepository, auditLog *audit.Logger) *PlanUseCase {
	return &PlanUseCase{planRepo: planRepo, auditLog: auditLog}
}

func (uc *PlanUseCase) Create(ctx context.Context, req PlanCreateRequest, callerID uuid.UUID, ip net.IP) (*PlanResponse, error) {
	tiers := req.Tiers
	if tiers == nil {
		tiers = []domain.CommissionTier{}
	}
	st := req.ServiceTypes
	if st == nil {
		st = []string{}
	}

	plan, err := uc.planRepo.Create(ctx, domain.CreatePlanParams{
		Code:         req.Code,
		Name:         req.Name,
		Description:  req.Description,
		Type:         req.Type,
		DefaultRate:  req.DefaultRate,
		Tiers:        tiers,
		ApplyBase:    req.ApplyBase,
		TriggerOn:    req.TriggerOn,
		ServiceTypes: st,
		CreatedBy:    callerID,
	})
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, IPAddress: ip.String(),
		Module: "commission", Resource: "commission_plan", ResourceID: &plan.ID,
		Action: "CREATE", NewValue: plan,
	})
	r := toPlanResponse(plan)
	return &r, nil
}

func (uc *PlanUseCase) GetByID(ctx context.Context, id uuid.UUID) (*PlanResponse, error) {
	plan, err := uc.planRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	r := toPlanResponse(plan)
	return &r, nil
}

func (uc *PlanUseCase) Update(ctx context.Context, id uuid.UUID, req PlanUpdateRequest, callerID uuid.UUID, ip net.IP) (*PlanResponse, error) {
	tiers := req.Tiers
	if tiers == nil {
		tiers = []domain.CommissionTier{}
	}
	st := req.ServiceTypes
	if st == nil {
		st = []string{}
	}

	plan, err := uc.planRepo.Update(ctx, domain.UpdatePlanParams{
		ID:           id,
		Name:         req.Name,
		Description:  req.Description,
		DefaultRate:  req.DefaultRate,
		Tiers:        tiers,
		ApplyBase:    req.ApplyBase,
		TriggerOn:    req.TriggerOn,
		ServiceTypes: st,
		UpdatedBy:    callerID,
	})
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, IPAddress: ip.String(),
		Module: "commission", Resource: "commission_plan", ResourceID: &plan.ID,
		Action: "UPDATE", NewValue: plan,
	})
	r := toPlanResponse(plan)
	return &r, nil
}

func (uc *PlanUseCase) Deactivate(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip net.IP) (*PlanResponse, error) {
	plan, err := uc.planRepo.Deactivate(ctx, id, callerID)
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, IPAddress: ip.String(),
		Module: "commission", Resource: "commission_plan", ResourceID: &plan.ID,
		Action: "DEACTIVATE", NewValue: plan,
	})
	r := toPlanResponse(plan)
	return &r, nil
}

func (uc *PlanUseCase) List(ctx context.Context, f domain.ListPlansFilter, page, size int) (pagination.OffsetResult[PlanResponse], error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	plans, total, err := uc.planRepo.List(ctx, f, page, size)
	if err != nil {
		return pagination.OffsetResult[PlanResponse]{}, err
	}
	items := make([]PlanResponse, len(plans))
	for i, p := range plans {
		items[i] = toPlanResponse(p)
	}
	return pagination.NewOffsetResult(items, total, page, size), nil
}
