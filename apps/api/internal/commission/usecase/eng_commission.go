package usecase

import (
	"context"
	"net"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/commission/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

const maxTotalRate = 1.0

type EngCommissionUseCase struct {
	ecRepo   domain.EngCommissionRepository
	auditLog *audit.Logger
}

func NewEngCommissionUseCase(ecRepo domain.EngCommissionRepository, auditLog *audit.Logger) *EngCommissionUseCase {
	return &EngCommissionUseCase{ecRepo: ecRepo, auditLog: auditLog}
}

func (uc *EngCommissionUseCase) Create(ctx context.Context, req EngCommissionCreateRequest, callerID uuid.UUID, ip net.IP) (*EngCommissionResponse, error) {
	// Validate total rate ≤ 100% per engagement
	if req.RateType == domain.CommissionTypeFlat || req.RateType == domain.CommissionTypeCustom {
		current, err := uc.ecRepo.SumRateByEngagement(ctx, req.EngagementID)
		if err != nil {
			return nil, err
		}
		if current+req.Rate > maxTotalRate {
			return nil, domain.ErrEngCommissionRateExceeds
		}
	}

	tiers := req.Tiers
	if tiers == nil {
		tiers = []domain.CommissionTier{}
	}

	ec, err := uc.ecRepo.Create(ctx, domain.CreateEngCommissionParams{
		EngagementID:  req.EngagementID,
		SalespersonID: req.SalespersonID,
		Role:          req.Role,
		PlanID:        req.PlanID,
		RateType:      req.RateType,
		Rate:          req.Rate,
		FixedAmount:   req.FixedAmount,
		Tiers:         tiers,
		ApplyBase:     req.ApplyBase,
		TriggerOn:     req.TriggerOn,
		MaxAmount:     req.MaxAmount,
		HoldbackPct:   req.HoldbackPct,
		Notes:         req.Notes,
		CreatedBy:     callerID,
	})
	if err != nil {
		return nil, err
	}

	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, IPAddress: ip.String(),
		Module: "commission", Resource: "engagement_commission", ResourceID: &ec.ID,
		Action: "create", NewValue: ec,
	})
	r := toEngCommissionResponse(ec)
	return &r, nil
}

func (uc *EngCommissionUseCase) GetByID(ctx context.Context, id uuid.UUID) (*EngCommissionResponse, error) {
	ec, err := uc.ecRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	r := toEngCommissionResponse(ec)
	return &r, nil
}

func (uc *EngCommissionUseCase) List(ctx context.Context, f domain.ListEngCommissionsFilter, page, size int) (pagination.OffsetResult[EngCommissionResponse], error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	ecs, total, err := uc.ecRepo.List(ctx, f, page, size)
	if err != nil {
		return pagination.OffsetResult[EngCommissionResponse]{}, err
	}
	items := make([]EngCommissionResponse, len(ecs))
	for i, ec := range ecs {
		items[i] = toEngCommissionResponse(ec)
	}
	return pagination.NewOffsetResult(items, total, page, size), nil
}

func (uc *EngCommissionUseCase) Cancel(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip net.IP) (*EngCommissionResponse, error) {
	ec, err := uc.ecRepo.Cancel(ctx, id, callerID)
	if err != nil {
		return nil, err
	}

	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, IPAddress: ip.String(),
		Module: "commission", Resource: "engagement_commission", ResourceID: &ec.ID,
		Action: "cancel", NewValue: ec,
	})
	r := toEngCommissionResponse(ec)
	return &r, nil
}

func (uc *EngCommissionUseCase) Approve(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip net.IP) (*EngCommissionResponse, error) {
	ec, err := uc.ecRepo.Approve(ctx, id, callerID)
	if err != nil {
		return nil, err
	}

	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, IPAddress: ip.String(),
		Module: "commission", Resource: "engagement_commission", ResourceID: &ec.ID,
		Action: "approve", NewValue: ec,
	})
	r := toEngCommissionResponse(ec)
	return &r, nil
}
