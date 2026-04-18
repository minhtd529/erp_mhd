package usecase

import (
	"context"
	"net"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/tax/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

type AdvisoryUseCase struct {
	repo     domain.AdvisoryRepository
	auditLog *audit.Logger
}

func NewAdvisoryUseCase(repo domain.AdvisoryRepository, auditLog *audit.Logger) *AdvisoryUseCase {
	return &AdvisoryUseCase{repo: repo, auditLog: auditLog}
}

type CreateAdvisoryRequest struct {
	EngagementID   *uuid.UUID          `json:"engagement_id"`
	AdvisoryType   domain.AdvisoryType `json:"advisory_type" binding:"required"`
	Recommendation string              `json:"recommendation" binding:"required"`
	Findings       string              `json:"findings"`
}

type UpdateAdvisoryRequest struct {
	Recommendation string `json:"recommendation" binding:"required"`
	Findings       string `json:"findings"`
}

type AttachFileRequest struct {
	FileName string `json:"file_name" binding:"required"`
	FilePath string `json:"file_path" binding:"required"`
}

func (uc *AdvisoryUseCase) Create(ctx context.Context, clientID uuid.UUID, req CreateAdvisoryRequest, callerID uuid.UUID, ip net.IP) (*domain.AdvisoryRecord, error) {
	a, err := uc.repo.Create(ctx, domain.CreateAdvisoryParams{
		ClientID:       clientID,
		EngagementID:   req.EngagementID,
		AdvisoryType:   req.AdvisoryType,
		Recommendation: req.Recommendation,
		Findings:       req.Findings,
		CreatedBy:      callerID,
	})
	if err != nil {
		return nil, err
	}
	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, IPAddress: ip.String(),
		Module: "tax", Resource: "advisory_record", ResourceID: &a.ID,
		Action: "CREATE", NewValue: a,
	})
	return a, nil
}

func (uc *AdvisoryUseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.AdvisoryRecord, error) {
	return uc.repo.FindByID(ctx, id)
}

func (uc *AdvisoryUseCase) List(ctx context.Context, clientID uuid.UUID, status domain.AdvisoryStatus, page, size int) (pagination.OffsetResult[domain.AdvisoryRecord], error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	f := domain.ListAdvisoryFilter{ClientID: &clientID, Status: status}
	list, total, err := uc.repo.List(ctx, f, page, size)
	if err != nil {
		return pagination.OffsetResult[domain.AdvisoryRecord]{}, err
	}
	items := make([]domain.AdvisoryRecord, len(list))
	for i, a := range list {
		items[i] = *a
	}
	return pagination.NewOffsetResult(items, total, page, size), nil
}

func (uc *AdvisoryUseCase) Update(ctx context.Context, id uuid.UUID, req UpdateAdvisoryRequest, callerID uuid.UUID, ip net.IP) (*domain.AdvisoryRecord, error) {
	a, err := uc.repo.Update(ctx, domain.UpdateAdvisoryParams{
		ID:             id,
		Recommendation: req.Recommendation,
		Findings:       req.Findings,
		UpdatedBy:      callerID,
	})
	if err != nil {
		return nil, err
	}
	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, IPAddress: ip.String(),
		Module: "tax", Resource: "advisory_record", ResourceID: &a.ID,
		Action: "UPDATE", NewValue: a,
	})
	return a, nil
}

func (uc *AdvisoryUseCase) Deliver(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip net.IP) (*domain.AdvisoryRecord, error) {
	a, err := uc.repo.Deliver(ctx, id, callerID)
	if err != nil {
		return nil, err
	}
	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, IPAddress: ip.String(),
		Module: "tax", Resource: "advisory_record", ResourceID: &a.ID,
		Action: "DELIVER", NewValue: map[string]string{"status": "DELIVERED"},
	})
	return a, nil
}

func (uc *AdvisoryUseCase) AttachFile(ctx context.Context, id uuid.UUID, req AttachFileRequest, callerID uuid.UUID, ip net.IP) (*domain.AdvisoryFile, error) {
	f, err := uc.repo.AttachFile(ctx, domain.AttachFileParams{
		AdvisoryID: id,
		FileName:   req.FileName,
		FilePath:   req.FilePath,
		CreatedBy:  callerID,
	})
	if err != nil {
		return nil, err
	}
	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, IPAddress: ip.String(),
		Module: "tax", Resource: "advisory_file", ResourceID: &f.ID,
		Action: "ATTACH_FILE", NewValue: f,
	})
	return f, nil
}

func (uc *AdvisoryUseCase) ListFiles(ctx context.Context, id uuid.UUID) ([]*domain.AdvisoryFile, error) {
	return uc.repo.ListFiles(ctx, id)
}
