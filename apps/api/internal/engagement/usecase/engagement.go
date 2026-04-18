// Package usecase implements the Engagement application layer.
package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/engagement/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/pagination"
	"github.com/mdh/erp-audit/api/pkg/ws"
)

// EngagementUseCase bundles Engagement CRUD + state transition operations.
type EngagementUseCase struct {
	repo      domain.EngagementRepository
	auditLog  *audit.Logger
	broadcast ws.Broadcaster
}

// NewEngagementUseCase constructs an EngagementUseCase.
func NewEngagementUseCase(repo domain.EngagementRepository, auditLog *audit.Logger, broadcast ws.Broadcaster) *EngagementUseCase {
	return &EngagementUseCase{repo: repo, auditLog: auditLog, broadcast: broadcast}
}

func (uc *EngagementUseCase) Create(ctx context.Context, req EngagementCreateRequest, callerID uuid.UUID, ip string) (*EngagementResponse, error) {
	e, err := uc.repo.Create(ctx, domain.CreateEngagementParams{
		ClientID:             req.ClientID,
		ServiceType:          req.ServiceType,
		FeeType:              req.FeeType,
		FeeAmount:            req.FeeAmount,
		PartnerID:            req.PartnerID,
		PrimarySalespersonID: req.PrimarySalespersonID,
		StartDate:            req.StartDate,
		EndDate:              req.EndDate,
		Description:          req.Description,
		CreatedBy:            callerID,
	})
	if err != nil {
		return nil, err
	}

	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "engagement", Resource: "engagements",
		ResourceID: &e.ID, Action: "CREATE", IPAddress: ip,
	})
	if uc.broadcast != nil {
		_ = uc.broadcast.Broadcast("engagement", "engagement.created", map[string]string{
			"id": e.ID.String(), "status": string(e.Status),
		})
	}

	resp := toEngagementResponse(e)
	return &resp, nil
}

func (uc *EngagementUseCase) GetByID(ctx context.Context, id uuid.UUID) (*EngagementResponse, error) {
	e, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toEngagementResponse(e)
	return &resp, nil
}

func (uc *EngagementUseCase) Update(ctx context.Context, id uuid.UUID, req EngagementUpdateRequest, callerID uuid.UUID, ip string) (*EngagementResponse, error) {
	e, err := uc.repo.Update(ctx, domain.UpdateEngagementParams{
		ID:                   id,
		ServiceType:          req.ServiceType,
		FeeType:              req.FeeType,
		FeeAmount:            req.FeeAmount,
		PartnerID:            req.PartnerID,
		PrimarySalespersonID: req.PrimarySalespersonID,
		StartDate:            req.StartDate,
		EndDate:              req.EndDate,
		Description:          req.Description,
		UpdatedBy:            callerID,
	})
	if err != nil {
		return nil, err
	}

	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "engagement", Resource: "engagements",
		ResourceID: &id, Action: "UPDATE", IPAddress: ip,
	})

	resp := toEngagementResponse(e)
	return &resp, nil
}

func (uc *EngagementUseCase) Delete(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip string) error {
	if err := uc.repo.SoftDelete(ctx, id, callerID); err != nil {
		return err
	}
	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "engagement", Resource: "engagements",
		ResourceID: &id, Action: "DELETE", IPAddress: ip,
	})
	return nil
}

func (uc *EngagementUseCase) List(ctx context.Context, req EngagementListRequest) (PaginatedResult[EngagementResponse], error) {
	items, total, err := uc.repo.List(ctx, domain.ListEngagementsFilter{
		Page:     req.Page,
		Size:     req.Size,
		ClientID: req.ClientID,
		Status:   req.Status,
		Q:        req.Q,
	})
	if err != nil {
		return PaginatedResult[EngagementResponse]{}, fmt.Errorf("engagement.List: %w", err)
	}
	responses := make([]EngagementResponse, len(items))
	for i, e := range items {
		responses[i] = toEngagementResponse(e)
	}
	return newPaginatedResult(responses, total, req.Page, req.Size), nil
}

// ListCursor returns cursor-paginated engagements. If req.Cursor is non-empty it
// decodes the position; otherwise it starts from the beginning.
func (uc *EngagementUseCase) ListCursor(ctx context.Context, req EngagementCursorListRequest) (EngagementCursorResult, error) {
	f := domain.CursorFilter{
		Size:     req.Size,
		ClientID: req.ClientID,
		Status:   req.Status,
		Q:        req.Q,
	}
	if req.Cursor != "" {
		c, err := pagination.DecodeCursor(req.Cursor)
		if err != nil {
			return EngagementCursorResult{}, err
		}
		f.AfterID = &c.ID
		f.AfterCreatedAt = &c.CreatedAt
	}

	items, err := uc.repo.ListCursor(ctx, f)
	if err != nil {
		return EngagementCursorResult{}, fmt.Errorf("engagement.ListCursor: %w", err)
	}

	responses := make([]EngagementResponse, len(items))
	for i, e := range items {
		responses[i] = toEngagementResponse(e)
	}
	return pagination.NewCursorResult(responses, req.Size, func(r EngagementResponse) pagination.Cursor {
		return pagination.Cursor{ID: r.ID, CreatedAt: r.CreatedAt}
	}), nil
}

// validTransitions defines allowed status state machine moves.
var validTransitions = map[domain.EngagementStatus][]domain.EngagementStatus{
	domain.StatusDraft:      {domain.StatusProposal, domain.StatusActive},
	domain.StatusProposal:   {domain.StatusContracted, domain.StatusDraft},
	domain.StatusContracted: {domain.StatusActive},
	domain.StatusActive:     {domain.StatusCompleted},
	domain.StatusCompleted:  {domain.StatusSettled},
}

func (uc *EngagementUseCase) Activate(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip string) (*EngagementResponse, error) {
	return uc.transition(ctx, id, domain.StatusActive, "STATE_TRANSITION", callerID, ip)
}

func (uc *EngagementUseCase) Complete(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip string) (*EngagementResponse, error) {
	return uc.transition(ctx, id, domain.StatusCompleted, "STATE_TRANSITION", callerID, ip)
}

func (uc *EngagementUseCase) transition(ctx context.Context, id uuid.UUID, next domain.EngagementStatus, action string, callerID uuid.UUID, ip string) (*EngagementResponse, error) {
	current, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	allowed := false
	for _, s := range validTransitions[current.Status] {
		if s == next {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, domain.ErrInvalidStateTransition
	}

	// CONTRACTED/ACTIVE requires a partner to be assigned.
	if next == domain.StatusActive && current.PartnerID == nil {
		return nil, domain.ErrPartnerNotAssigned
	}

	e, err := uc.repo.UpdateStatus(ctx, id, next, callerID)
	if err != nil {
		return nil, err
	}

	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "engagement", Resource: "engagements",
		ResourceID: &id, Action: action, IPAddress: ip,
	})
	if uc.broadcast != nil {
		_ = uc.broadcast.Broadcast("engagement", "engagement.state_changed", map[string]string{
			"id": id.String(), "status": string(next), "previous_status": string(current.Status),
		})
	}

	resp := toEngagementResponse(e)
	return &resp, nil
}
