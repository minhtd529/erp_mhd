// Package usecase implements the WorkingPaper application layer.
package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/workingpaper/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

// WorkingPaperUseCase handles WP CRUD and lifecycle transitions.
type WorkingPaperUseCase struct {
	wpRepo     domain.WorkingPaperRepository
	reviewRepo domain.ReviewRepository
	auditLog   *audit.Logger
}

// NewWorkingPaperUseCase constructs a WorkingPaperUseCase.
func NewWorkingPaperUseCase(
	wpRepo domain.WorkingPaperRepository,
	reviewRepo domain.ReviewRepository,
	auditLog *audit.Logger,
) *WorkingPaperUseCase {
	return &WorkingPaperUseCase{wpRepo: wpRepo, reviewRepo: reviewRepo, auditLog: auditLog}
}

func (uc *WorkingPaperUseCase) Create(ctx context.Context, engagementID uuid.UUID, req WPCreateRequest, callerID uuid.UUID, ip string) (*WPResponse, error) {
	wp, err := uc.wpRepo.Create(ctx, domain.CreateWPParams{
		EngagementID: engagementID,
		FolderID:     req.FolderID,
		DocumentType: req.DocumentType,
		Title:        req.Title,
		FileID:       req.FileID,
		CreatedBy:    callerID,
	})
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "working_papers", Resource: "working_papers",
		ResourceID: &wp.ID, Action: "CREATE", IPAddress: ip,
	})

	resp := toWPResponse(wp)
	return &resp, nil
}

func (uc *WorkingPaperUseCase) GetByID(ctx context.Context, id uuid.UUID) (*WPResponse, error) {
	wp, err := uc.wpRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toWPResponse(wp)
	return &resp, nil
}

func (uc *WorkingPaperUseCase) Update(ctx context.Context, id uuid.UUID, req WPUpdateRequest, callerID uuid.UUID, ip string) (*WPResponse, error) {
	existing, err := uc.wpRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing.Status == domain.WPStatusFinalized || existing.Status == domain.WPStatusSignedOff {
		return nil, domain.ErrWorkingPaperNotEditable
	}

	wp, err := uc.wpRepo.Update(ctx, domain.UpdateWPParams{
		ID:        id,
		Title:     req.Title,
		FolderID:  req.FolderID,
		FileID:    req.FileID,
		UpdatedBy: callerID,
	})
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "working_papers", Resource: "working_papers",
		ResourceID: &id, Action: "UPDATE", IPAddress: ip,
	})

	resp := toWPResponse(wp)
	return &resp, nil
}

func (uc *WorkingPaperUseCase) Delete(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip string) error {
	if err := uc.wpRepo.SoftDelete(ctx, id, callerID); err != nil {
		return err
	}
	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "working_papers", Resource: "working_papers",
		ResourceID: &id, Action: "DELETE", IPAddress: ip,
	})
	return nil
}

// SubmitForReview transitions DRAFT → IN_REVIEW and seeds the review chain.
func (uc *WorkingPaperUseCase) SubmitForReview(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip string) (*WPResponse, error) {
	wp, err := uc.wpRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if wp.Status != domain.WPStatusDraft {
		return nil, domain.ErrInvalidStateTransition
	}

	updated, err := uc.wpRepo.UpdateStatus(ctx, id, domain.WPStatusInReview, callerID)
	if err != nil {
		return nil, err
	}

	// Seed review chain: SENIOR_AUDITOR → MANAGER → PARTNER
	for _, role := range []domain.ReviewerRole{domain.RoleSeniorAuditor, domain.RoleManager, domain.RolePartner} {
		_, _ = uc.reviewRepo.Create(ctx, domain.CreateReviewParams{
			WorkingPaperID: id,
			ReviewerRole:   role,
		})
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "working_papers", Resource: "working_papers",
		ResourceID: &id, Action: "STATE_TRANSITION", IPAddress: ip,
		NewValue: map[string]string{"status": "IN_REVIEW"},
	})

	resp := toWPResponse(updated)
	return &resp, nil
}

// Finalize captures a JSONB snapshot and transitions IN_REVIEW → FINALIZED.
func (uc *WorkingPaperUseCase) Finalize(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip string) (*WPResponse, error) {
	wp, err := uc.wpRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if wp.Status != domain.WPStatusInReview && wp.Status != domain.WPStatusCommented {
		return nil, domain.ErrInvalidStateTransition
	}

	// All reviews must be APPROVED before finalizing
	reviews, err := uc.reviewRepo.ListByWP(ctx, id)
	if err != nil {
		return nil, err
	}
	for _, r := range reviews {
		if r.ReviewStatus != domain.ReviewApproved {
			return nil, domain.ErrReviewChainIncomplete
		}
	}

	// Check no open comments
	unresolved, err := uc.reviewRepo.CountUnresolved(ctx, id)
	if err != nil {
		return nil, err
	}
	if unresolved > 0 {
		return nil, domain.ErrCommentsNotResolved
	}

	snapshot, _ := json.Marshal(map[string]any{
		"title":         wp.Title,
		"document_type": wp.DocumentType,
		"file_id":       wp.FileID,
		"finalized_at":  time.Now().UTC(),
		"finalized_by":  callerID.String(),
	})

	finalized, err := uc.wpRepo.Finalize(ctx, domain.FinalizeWPParams{
		ID:           id,
		SnapshotData: snapshot,
		FinalizedAt:  time.Now(),
		UpdatedBy:    callerID,
	})
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "working_papers", Resource: "working_papers",
		ResourceID: &id, Action: "STATE_TRANSITION", IPAddress: ip,
		NewValue: map[string]string{"status": "FINALIZED"},
	})

	resp := toWPResponse(finalized)
	return &resp, nil
}

// SignOff transitions FINALIZED → SIGNED_OFF (Partner only).
func (uc *WorkingPaperUseCase) SignOff(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip string) (*WPResponse, error) {
	wp, err := uc.wpRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if wp.Status != domain.WPStatusFinalized {
		return nil, domain.ErrInvalidStateTransition
	}

	updated, err := uc.wpRepo.UpdateStatus(ctx, id, domain.WPStatusSignedOff, callerID)
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "working_papers", Resource: "working_papers",
		ResourceID: &id, Action: "APPROVE", IPAddress: ip,
		NewValue: map[string]string{"status": "SIGNED_OFF"},
	})

	resp := toWPResponse(updated)
	return &resp, nil
}

// PendingReview returns working papers that are waiting for the given reviewer role to act.
func (uc *WorkingPaperUseCase) PendingReview(ctx context.Context, req PendingReviewRequest) (*PaginatedResult[WPResponse], error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 20
	}
	wps, total, err := uc.wpRepo.ListPendingReview(ctx, req.Role, req.Page, req.Size)
	if err != nil {
		return nil, err
	}
	data := make([]WPResponse, len(wps))
	for i, wp := range wps {
		data[i] = toWPResponse(wp)
	}
	result := pagination.NewOffsetResult(data, total, req.Page, req.Size)
	return &result, nil
}

func (uc *WorkingPaperUseCase) List(ctx context.Context, engagementID uuid.UUID, req WPListRequest) (*PaginatedResult[WPResponse], error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 20
	}
	wps, total, err := uc.wpRepo.List(ctx, domain.ListWPFilter{
		EngagementID: engagementID,
		Status:       req.Status,
		Page:         req.Page,
		Size:         req.Size,
	})
	if err != nil {
		return nil, err
	}
	data := make([]WPResponse, len(wps))
	for i, wp := range wps {
		data[i] = toWPResponse(wp)
	}
	result := pagination.NewOffsetResult(data, total, req.Page, req.Size)
	return &result, nil
}
