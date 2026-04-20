package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/org/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

type BranchUseCase struct {
	repo     domain.BranchRepository
	auditLog *audit.Logger
}

func NewBranchUseCase(repo domain.BranchRepository, auditLog *audit.Logger) *BranchUseCase {
	return &BranchUseCase{repo: repo, auditLog: auditLog}
}

func (uc *BranchUseCase) Create(ctx context.Context, req BranchCreateRequest, callerID *uuid.UUID, ip string) (*BranchResponse, error) {
	b, err := uc.repo.Create(ctx, domain.CreateBranchParams{
		Code: req.Code, Name: req.Name,
		Address: req.Address, Phone: req.Phone,
		CreatedBy: callerID,
	})
	if err != nil {
		return nil, err
	}
	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: callerID, Module: "org", Resource: "branches",
		ResourceID: &b.ID, Action: "CREATE", IPAddress: ip,
	})
	resp := toBranchResponse(b)
	return &resp, nil
}

func (uc *BranchUseCase) GetByID(ctx context.Context, id uuid.UUID) (*BranchResponse, error) {
	b, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toBranchResponse(b)
	return &resp, nil
}

func (uc *BranchUseCase) Update(ctx context.Context, id uuid.UUID, req BranchUpdateRequest, callerID *uuid.UUID, ip string) (*BranchResponse, error) {
	b, err := uc.repo.Update(ctx, domain.UpdateBranchParams{
		ID: id, Code: req.Code, Name: req.Name,
		Address: req.Address, Phone: req.Phone,
		IsActive: req.IsActive, UpdatedBy: callerID,
	})
	if err != nil {
		return nil, err
	}
	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: callerID, Module: "org", Resource: "branches",
		ResourceID: &id, Action: "UPDATE", IPAddress: ip,
	})
	resp := toBranchResponse(b)
	return &resp, nil
}

func (uc *BranchUseCase) List(ctx context.Context, req BranchListRequest) (PaginatedResult[BranchResponse], error) {
	branches, total, err := uc.repo.List(ctx, domain.ListBranchesFilter{
		Page: req.Page, Size: req.Size, IsActive: req.IsActive, Q: req.Q,
	})
	if err != nil {
		return PaginatedResult[BranchResponse]{}, fmt.Errorf("org.ListBranches: %w", err)
	}
	items := make([]BranchResponse, len(branches))
	for i, b := range branches {
		items[i] = toBranchResponse(b)
	}
	return pagination.NewOffsetResult(items, total, req.Page, req.Size), nil
}

func toBranchResponse(b *domain.Branch) BranchResponse {
	return BranchResponse{
		ID: b.ID, Code: b.Code, Name: b.Name,
		Address: b.Address, Phone: b.Phone, IsActive: b.IsActive,
		CreatedAt: b.CreatedAt, UpdatedAt: b.UpdatedAt,
		CreatedBy: b.CreatedBy, UpdatedBy: b.UpdatedBy,
	}
}
