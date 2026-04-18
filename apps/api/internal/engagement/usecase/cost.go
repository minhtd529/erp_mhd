package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/engagement/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

// CostUseCase manages direct cost recording and approval.
type CostUseCase struct {
	costRepo       domain.CostRepository
	engagementRepo domain.EngagementRepository
	auditLog       *audit.Logger
}

// NewCostUseCase constructs a CostUseCase.
func NewCostUseCase(costRepo domain.CostRepository, engagementRepo domain.EngagementRepository, auditLog *audit.Logger) *CostUseCase {
	return &CostUseCase{costRepo: costRepo, engagementRepo: engagementRepo, auditLog: auditLog}
}

// CostListRequest carries pagination params for cost listing.
type CostListRequest struct {
	Page int
	Size int
}

func (uc *CostUseCase) List(ctx context.Context, engagementID uuid.UUID, req CostListRequest) (PaginatedResult[CostResponse], error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	costs, err := uc.costRepo.ListByEngagement(ctx, engagementID)
	if err != nil {
		return PaginatedResult[CostResponse]{}, err
	}
	all := make([]CostResponse, len(costs))
	for i, c := range costs {
		all[i] = toCostResponse(c)
	}
	page, size := req.Page, req.Size
	start := (page - 1) * size
	if start > len(all) {
		start = len(all)
	}
	end := start + size
	if end > len(all) {
		end = len(all)
	}
	return newPaginatedResult(all[start:end], int64(len(all)), page, size), nil
}

func (uc *CostUseCase) Create(ctx context.Context, engagementID uuid.UUID, req CostCreateRequest, callerID uuid.UUID, ip string) (*CostResponse, error) {
	if _, err := uc.engagementRepo.FindByID(ctx, engagementID); err != nil {
		return nil, err
	}

	c, err := uc.costRepo.Create(ctx, domain.CreateCostParams{
		EngagementID: engagementID,
		CostType:     req.CostType,
		Description:  req.Description,
		Amount:       req.Amount,
		ReceiptURL:   req.ReceiptURL,
		CreatedBy:    callerID,
	})
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "engagement", Resource: "direct_costs",
		ResourceID: &c.ID, Action: "CREATE", IPAddress: ip,
	})

	resp := toCostResponse(c)
	return &resp, nil
}

func (uc *CostUseCase) Submit(ctx context.Context, engagementID uuid.UUID, costID uuid.UUID, callerID uuid.UUID, ip string) (*CostResponse, error) {
	existing, err := uc.costRepo.FindByID(ctx, costID)
	if err != nil {
		return nil, err
	}
	if existing.EngagementID != engagementID {
		return nil, domain.ErrCostNotFound
	}
	if existing.Status != domain.CostDraft {
		return nil, domain.ErrInvalidCostTransition
	}

	c, err := uc.costRepo.UpdateStatus(ctx, costID, domain.CostSubmitted, callerID, nil)
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "engagement", Resource: "direct_costs",
		ResourceID: &costID, Action: "STATE_TRANSITION", IPAddress: ip,
	})

	resp := toCostResponse(c)
	return &resp, nil
}

func (uc *CostUseCase) Approve(ctx context.Context, engagementID uuid.UUID, costID uuid.UUID, callerID uuid.UUID, ip string) (*CostResponse, error) {
	existing, err := uc.costRepo.FindByID(ctx, costID)
	if err != nil {
		return nil, err
	}
	if existing.EngagementID != engagementID {
		return nil, domain.ErrCostNotFound
	}
	if existing.Status != domain.CostSubmitted {
		return nil, domain.ErrCostApprovalRequired
	}

	c, err := uc.costRepo.UpdateStatus(ctx, costID, domain.CostApproved, callerID, nil)
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "engagement", Resource: "direct_costs",
		ResourceID: &costID, Action: "APPROVE", IPAddress: ip,
	})

	resp := toCostResponse(c)
	return &resp, nil
}

func (uc *CostUseCase) Reject(ctx context.Context, engagementID uuid.UUID, costID uuid.UUID, reason string, callerID uuid.UUID, ip string) (*CostResponse, error) {
	existing, err := uc.costRepo.FindByID(ctx, costID)
	if err != nil {
		return nil, err
	}
	if existing.EngagementID != engagementID {
		return nil, domain.ErrCostNotFound
	}
	if existing.Status != domain.CostSubmitted {
		return nil, domain.ErrCostApprovalRequired
	}

	c, err := uc.costRepo.UpdateStatus(ctx, costID, domain.CostRejected, callerID, &reason)
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "engagement", Resource: "direct_costs",
		ResourceID: &costID, Action: "REJECT", IPAddress: ip,
	})

	resp := toCostResponse(c)
	return &resp, nil
}
