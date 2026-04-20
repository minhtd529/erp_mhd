package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/org/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

type DepartmentUseCase struct {
	repo     domain.DepartmentRepository
	auditLog *audit.Logger
}

func NewDepartmentUseCase(repo domain.DepartmentRepository, auditLog *audit.Logger) *DepartmentUseCase {
	return &DepartmentUseCase{repo: repo, auditLog: auditLog}
}

func (uc *DepartmentUseCase) Create(ctx context.Context, req DepartmentCreateRequest, callerID *uuid.UUID, ip string) (*DepartmentResponse, error) {
	d, err := uc.repo.Create(ctx, domain.CreateDepartmentParams{
		Code: req.Code, Name: req.Name,
		BranchID: req.BranchID, CreatedBy: callerID,
	})
	if err != nil {
		return nil, err
	}
	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: callerID, Module: "org", Resource: "departments",
		ResourceID: &d.ID, Action: "CREATE", IPAddress: ip,
	})
	resp := toDeptResponse(d)
	return &resp, nil
}

func (uc *DepartmentUseCase) GetByID(ctx context.Context, id uuid.UUID) (*DepartmentResponse, error) {
	d, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toDeptResponse(d)
	return &resp, nil
}

func (uc *DepartmentUseCase) Update(ctx context.Context, id uuid.UUID, req DepartmentUpdateRequest, callerID *uuid.UUID, ip string) (*DepartmentResponse, error) {
	d, err := uc.repo.Update(ctx, domain.UpdateDepartmentParams{
		ID: id, Code: req.Code, Name: req.Name,
		BranchID: req.BranchID, IsActive: req.IsActive,
		UpdatedBy: callerID,
	})
	if err != nil {
		return nil, err
	}
	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: callerID, Module: "org", Resource: "departments",
		ResourceID: &id, Action: "UPDATE", IPAddress: ip,
	})
	resp := toDeptResponse(d)
	return &resp, nil
}

func (uc *DepartmentUseCase) List(ctx context.Context, req DepartmentListRequest) (PaginatedResult[DepartmentResponse], error) {
	depts, total, err := uc.repo.List(ctx, domain.ListDepartmentsFilter{
		Page: req.Page, Size: req.Size,
		BranchID: req.BranchID, IsActive: req.IsActive, Q: req.Q,
	})
	if err != nil {
		return PaginatedResult[DepartmentResponse]{}, fmt.Errorf("org.ListDepartments: %w", err)
	}
	items := make([]DepartmentResponse, len(depts))
	for i, d := range depts {
		items[i] = toDeptResponse(d)
	}
	return pagination.NewOffsetResult(items, total, req.Page, req.Size), nil
}

func toDeptResponse(d *domain.Department) DepartmentResponse {
	return DepartmentResponse{
		ID: d.ID, BranchID: d.BranchID, Code: d.Code, Name: d.Name,
		IsActive: d.IsActive,
		CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
		CreatedBy: d.CreatedBy, UpdatedBy: d.UpdatedBy,
	}
}
