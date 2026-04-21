package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/hrm/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

// ─── DTOs ────────────────────────────────────────────────────────────────────

type BranchHRMResponse struct {
	ID                     uuid.UUID  `json:"id"`
	Code                   string     `json:"code"`
	Name                   string     `json:"name"`
	Address                *string    `json:"address,omitempty"`
	Phone                  *string    `json:"phone,omitempty"`
	IsActive               bool       `json:"is_active"`
	IsHeadOffice           bool       `json:"is_head_office"`
	City                   *string    `json:"city,omitempty"`
	EstablishedDate        *string    `json:"established_date,omitempty"`
	HeadOfBranchUserID     *uuid.UUID `json:"head_of_branch_user_id,omitempty"`
	TaxCode                *string    `json:"tax_code,omitempty"`
	AuthorizationDocNumber *string    `json:"authorization_doc_number,omitempty"`
	AuthorizationDate      *string    `json:"authorization_date,omitempty"`
	CreatedAt              string     `json:"created_at"`
	UpdatedAt              string     `json:"updated_at"`
}

type UpdateBranchHRMRequest struct {
	Name                   *string `json:"name"`
	Address                *string `json:"address"`
	Phone                  *string `json:"phone"`
	City                   *string `json:"city"`
	TaxCode                *string `json:"tax_code"`
	EstablishedDate        *string `json:"established_date"`
	AuthorizationDocNumber *string `json:"authorization_doc_number"`
	AuthorizationDate      *string `json:"authorization_date"`
}

type ListBranchHRMRequest struct {
	Page     int    `form:"page,default=1"  binding:"min=1"`
	Size     int    `form:"size,default=20" binding:"min=1,max=100"`
	IsActive *bool  `form:"is_active"`
	Q        string `form:"q"`
}

type DeptHRMResponse struct {
	ID                     uuid.UUID  `json:"id"`
	Code                   string     `json:"code"`
	Name                   string     `json:"name"`
	BranchID               *uuid.UUID `json:"branch_id,omitempty"`
	IsActive               bool       `json:"is_active"`
	IsDeleted              bool       `json:"is_deleted"`
	Description            *string    `json:"description,omitempty"`
	DeptType               string     `json:"dept_type"`
	HeadEmployeeID         *uuid.UUID `json:"head_employee_id,omitempty"`
	AuthorizationDocNumber *string    `json:"authorization_doc_number,omitempty"`
	AuthorizationDate      *string    `json:"authorization_date,omitempty"`
	CreatedAt              string     `json:"created_at"`
	UpdatedAt              string     `json:"updated_at"`
}

type UpdateDeptHRMRequest struct {
	Name                   *string `json:"name"`
	Description            *string `json:"description"`
	DeptType               *string `json:"dept_type"`
	AuthorizationDocNumber *string `json:"authorization_doc_number"`
	AuthorizationDate      *string `json:"authorization_date"`
}

type ListDeptHRMRequest struct {
	Page     int        `form:"page,default=1"  binding:"min=1"`
	Size     int        `form:"size,default=20" binding:"min=1,max=100"`
	BranchID *uuid.UUID `form:"branch_id"`
	IsActive *bool      `form:"is_active"`
	Q        string     `form:"q"`
}

type BranchDeptResponse struct {
	BranchID       uuid.UUID  `json:"branch_id"`
	DepartmentID   uuid.UUID  `json:"department_id"`
	HeadEmployeeID *uuid.UUID `json:"head_employee_id,omitempty"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      string     `json:"created_at"`
}

type LinkBranchDeptRequest struct {
	BranchID     uuid.UUID `json:"branch_id"     binding:"required"`
	DepartmentID uuid.UUID `json:"department_id" binding:"required"`
}

type ListBranchDeptRequest struct {
	Page     int        `form:"page,default=1"  binding:"min=1"`
	Size     int        `form:"size,default=20" binding:"min=1,max=100"`
	BranchID *uuid.UUID `form:"branch_id"`
	IsActive *bool      `form:"is_active"`
}

// OrgChartResponse is the full tree returned by GET /hrm/organization/org-chart.
type OrgChartResponse struct {
	Branches []domain.OrgChartBranch `json:"branches"`
}

type AssignBranchHeadRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type AssignDeptHeadRequest struct {
	EmployeeID string `json:"employee_id" binding:"required"`
}

// ─── UseCase ─────────────────────────────────────────────────────────────────

// OrgUseCase bundles all HRM organization operations.
type OrgUseCase struct {
	repo     domain.OrgRepository
	auditLog *audit.Logger
}

// NewOrgUseCase constructs an OrgUseCase.
func NewOrgUseCase(repo domain.OrgRepository, auditLog *audit.Logger) *OrgUseCase {
	return &OrgUseCase{repo: repo, auditLog: auditLog}
}

// ─── Branch methods ───────────────────────────────────────────────────────────

func (uc *OrgUseCase) GetBranch(ctx context.Context, id uuid.UUID) (*BranchHRMResponse, error) {
	b, err := uc.repo.FindBranchByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toBranchHRMResponse(b)
	return &resp, nil
}

func (uc *OrgUseCase) ListBranches(ctx context.Context, req ListBranchHRMRequest) (pagination.OffsetResult[BranchHRMResponse], error) {
	branches, total, err := uc.repo.ListBranches(ctx, domain.ListHRMBranchesFilter{
		Page: req.Page, Size: req.Size, IsActive: req.IsActive, Q: req.Q,
	})
	if err != nil {
		return pagination.OffsetResult[BranchHRMResponse]{}, fmt.Errorf("hrm.ListBranches: %w", err)
	}
	items := make([]BranchHRMResponse, len(branches))
	for i, b := range branches {
		items[i] = toBranchHRMResponse(b)
	}
	return pagination.NewOffsetResult(items, total, req.Page, req.Size), nil
}

func (uc *OrgUseCase) UpdateBranch(ctx context.Context, id uuid.UUID, req UpdateBranchHRMRequest, callerID *uuid.UUID, ip string, callerRoles []string, callerBranchID *uuid.UUID) (*BranchHRMResponse, error) {
	// HEAD_OF_BRANCH may only edit non-critical fields on their own branch.
	if isOnlyHoB(callerRoles) {
		if callerBranchID == nil || *callerBranchID != id {
			return nil, domain.ErrInsufficientPermission
		}
		if req.Name != nil || req.TaxCode != nil || req.EstablishedDate != nil ||
			req.AuthorizationDocNumber != nil || req.AuthorizationDate != nil {
			return nil, domain.ErrInsufficientPermission
		}
	}

	p := domain.UpdateHRMBranchParams{
		ID:                     id,
		Name:                   req.Name,
		Address:                req.Address,
		Phone:                  req.Phone,
		City:                   req.City,
		TaxCode:                req.TaxCode,
		AuthorizationDocNumber: req.AuthorizationDocNumber,
		UpdatedBy:              callerID,
	}
	if req.EstablishedDate != nil {
		if t, err := time.Parse("2006-01-02", *req.EstablishedDate); err == nil {
			p.EstablishedDate = &t
		}
	}
	if req.AuthorizationDate != nil {
		if t, err := time.Parse("2006-01-02", *req.AuthorizationDate); err == nil {
			p.AuthorizationDate = &t
		}
	}

	b, err := uc.repo.UpdateBranch(ctx, p)
	if err != nil {
		return nil, err
	}

	if uc.auditLog != nil {
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID:     callerID,
			Module:     "hrm",
			Resource:   "branches",
			ResourceID: &id,
			Action:     "UPDATE",
			IPAddress:  ip,
		})
	}

	resp := toBranchHRMResponse(b)
	return &resp, nil
}

// ─── Department methods ───────────────────────────────────────────────────────

func (uc *OrgUseCase) GetDept(ctx context.Context, id uuid.UUID) (*DeptHRMResponse, error) {
	d, err := uc.repo.FindDeptByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toDeptHRMResponse(d)
	return &resp, nil
}

func (uc *OrgUseCase) ListDepts(ctx context.Context, req ListDeptHRMRequest) (pagination.OffsetResult[DeptHRMResponse], error) {
	depts, total, err := uc.repo.ListDepts(ctx, domain.ListHRMDeptsFilter{
		Page: req.Page, Size: req.Size, BranchID: req.BranchID, IsActive: req.IsActive, Q: req.Q,
	})
	if err != nil {
		return pagination.OffsetResult[DeptHRMResponse]{}, fmt.Errorf("hrm.ListDepts: %w", err)
	}
	items := make([]DeptHRMResponse, len(depts))
	for i, d := range depts {
		items[i] = toDeptHRMResponse(d)
	}
	return pagination.NewOffsetResult(items, total, req.Page, req.Size), nil
}

func (uc *OrgUseCase) UpdateDept(ctx context.Context, id uuid.UUID, req UpdateDeptHRMRequest, callerID *uuid.UUID, ip string) (*DeptHRMResponse, error) {
	p := domain.UpdateHRMDeptParams{
		ID:                     id,
		Name:                   req.Name,
		Description:            req.Description,
		DeptType:               req.DeptType,
		AuthorizationDocNumber: req.AuthorizationDocNumber,
		UpdatedBy:              callerID,
	}
	if req.AuthorizationDate != nil {
		if t, err := time.Parse("2006-01-02", *req.AuthorizationDate); err == nil {
			p.AuthorizationDate = &t
		}
	}

	d, err := uc.repo.UpdateDept(ctx, p)
	if err != nil {
		return nil, err
	}

	if uc.auditLog != nil {
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID:     callerID,
			Module:     "hrm",
			Resource:   "departments",
			ResourceID: &id,
			Action:     "UPDATE",
			IPAddress:  ip,
		})
	}

	resp := toDeptHRMResponse(d)
	return &resp, nil
}

// ─── BranchDepartment methods ─────────────────────────────────────────────────

func (uc *OrgUseCase) ListBranchDepts(ctx context.Context, req ListBranchDeptRequest) (pagination.OffsetResult[BranchDeptResponse], error) {
	bds, total, err := uc.repo.ListBranchDepts(ctx, domain.ListBranchDeptsFilter{
		Page: req.Page, Size: req.Size, BranchID: req.BranchID, IsActive: req.IsActive,
	})
	if err != nil {
		return pagination.OffsetResult[BranchDeptResponse]{}, fmt.Errorf("hrm.ListBranchDepts: %w", err)
	}
	items := make([]BranchDeptResponse, len(bds))
	for i, bd := range bds {
		items[i] = toBranchDeptResponse(bd)
	}
	return pagination.NewOffsetResult(items, total, req.Page, req.Size), nil
}

func (uc *OrgUseCase) LinkBranchDept(ctx context.Context, req LinkBranchDeptRequest, callerID *uuid.UUID, ip string) (*BranchDeptResponse, error) {
	bd, err := uc.repo.CreateBranchDept(ctx, domain.CreateBranchDeptParams{
		BranchID:     req.BranchID,
		DepartmentID: req.DepartmentID,
		CreatedBy:    callerID,
	})
	if err != nil {
		return nil, err
	}

	if uc.auditLog != nil {
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID:    callerID,
			Module:    "hrm",
			Resource:  "branch_departments",
			Action:    "CREATE",
			IPAddress: ip,
		})
	}

	resp := toBranchDeptResponse(bd)
	return &resp, nil
}

func (uc *OrgUseCase) UnlinkBranchDept(ctx context.Context, branchID, deptID uuid.UUID, callerID *uuid.UUID, ip string) error {
	if err := uc.repo.SoftDeleteBranchDept(ctx, branchID, deptID); err != nil {
		return err
	}

	if uc.auditLog != nil {
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID:    callerID,
			Module:    "hrm",
			Resource:  "branch_departments",
			Action:    "DELETE",
			IPAddress: ip,
		})
	}

	return nil
}

// ─── OrgChart ────────────────────────────────────────────────────────────────

func (uc *OrgUseCase) GetOrgChart(ctx context.Context) (*OrgChartResponse, error) {
	branches, err := uc.repo.ListBranchesWithDepts(ctx)
	if err != nil {
		return nil, fmt.Errorf("hrm.GetOrgChart: %w", err)
	}
	nodes := make([]domain.OrgChartBranch, len(branches))
	for i, b := range branches {
		nodes[i] = *b
	}
	return &OrgChartResponse{Branches: nodes}, nil
}

// ─── Assign-head / Deactivate methods ────────────────────────────────────────

func (uc *OrgUseCase) AssignBranchHead(ctx context.Context, branchID uuid.UUID, req AssignBranchHeadRequest, callerID *uuid.UUID, ip string) (*BranchHRMResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}
	b, err := uc.repo.AssignBranchHead(ctx, domain.AssignBranchHeadParams{
		BranchID:  branchID,
		UserID:    userID,
		UpdatedBy: callerID,
	})
	if err != nil {
		return nil, err
	}
	if uc.auditLog != nil {
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID: callerID, Module: "hrm", Resource: "branches",
			ResourceID: &branchID, Action: "ASSIGN_HEAD", IPAddress: ip,
		})
	}
	resp := toBranchHRMResponse(b)
	return &resp, nil
}

func (uc *OrgUseCase) DeactivateBranch(ctx context.Context, branchID uuid.UUID, callerID *uuid.UUID, ip string) error {
	// Note: employee-count guard deferred until migration 000021 adds dept FK to employees.
	if err := uc.repo.DeactivateBranch(ctx, branchID, callerID); err != nil {
		return err
	}
	if uc.auditLog != nil {
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID: callerID, Module: "hrm", Resource: "branches",
			ResourceID: &branchID, Action: "DEACTIVATE", IPAddress: ip,
		})
	}
	return nil
}

func (uc *OrgUseCase) AssignDeptHead(ctx context.Context, deptID uuid.UUID, req AssignDeptHeadRequest, callerID *uuid.UUID, ip string) (*DeptHRMResponse, error) {
	empID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("invalid employee_id: %w", err)
	}
	d, err := uc.repo.AssignDeptHead(ctx, domain.AssignDeptHeadParams{
		DeptID:     deptID,
		EmployeeID: empID,
		UpdatedBy:  callerID,
	})
	if err != nil {
		return nil, err
	}
	if uc.auditLog != nil {
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID: callerID, Module: "hrm", Resource: "departments",
			ResourceID: &deptID, Action: "ASSIGN_HEAD", IPAddress: ip,
		})
	}
	resp := toDeptHRMResponse(d)
	return &resp, nil
}

func (uc *OrgUseCase) DeactivateDept(ctx context.Context, deptID uuid.UUID, callerID *uuid.UUID, ip string) error {
	if err := uc.repo.DeactivateDept(ctx, deptID, callerID); err != nil {
		return err
	}
	if uc.auditLog != nil {
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID: callerID, Module: "hrm", Resource: "departments",
			ResourceID: &deptID, Action: "DEACTIVATE", IPAddress: ip,
		})
	}
	return nil
}

// isOnlyHoB returns true when the caller's roles contain HEAD_OF_BRANCH but none
// of the higher-privilege roles that would bypass branch scope restrictions.
func isOnlyHoB(roles []string) bool {
	hasHoB := false
	for _, r := range roles {
		switch r {
		case "SUPER_ADMIN", "CHAIRMAN", "CEO":
			return false
		case "HEAD_OF_BRANCH":
			hasHoB = true
		}
	}
	return hasHoB
}

// ─── Converters ───────────────────────────────────────────────────────────────

func toBranchHRMResponse(b *domain.HRMBranch) BranchHRMResponse {
	r := BranchHRMResponse{
		ID:                     b.ID,
		Code:                   b.Code,
		Name:                   b.Name,
		Address:                b.Address,
		Phone:                  b.Phone,
		IsActive:               b.IsActive,
		IsHeadOffice:           b.IsHeadOffice,
		City:                   b.City,
		HeadOfBranchUserID:     b.HeadOfBranchUserID,
		TaxCode:                b.TaxCode,
		AuthorizationDocNumber: b.AuthorizationDocNumber,
		CreatedAt:              b.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              b.UpdatedAt.Format(time.RFC3339),
	}
	if b.EstablishedDate != nil {
		s := b.EstablishedDate.Format("2006-01-02")
		r.EstablishedDate = &s
	}
	if b.AuthorizationDate != nil {
		s := b.AuthorizationDate.Format("2006-01-02")
		r.AuthorizationDate = &s
	}
	return r
}

func toDeptHRMResponse(d *domain.HRMDepartment) DeptHRMResponse {
	r := DeptHRMResponse{
		ID:                     d.ID,
		Code:                   d.Code,
		Name:                   d.Name,
		BranchID:               d.BranchID,
		IsActive:               d.IsActive,
		IsDeleted:              d.IsDeleted,
		Description:            d.Description,
		DeptType:               d.DeptType,
		HeadEmployeeID:         d.HeadEmployeeID,
		AuthorizationDocNumber: d.AuthorizationDocNumber,
		CreatedAt:              d.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              d.UpdatedAt.Format(time.RFC3339),
	}
	if d.AuthorizationDate != nil {
		s := d.AuthorizationDate.Format("2006-01-02")
		r.AuthorizationDate = &s
	}
	return r
}

func toBranchDeptResponse(bd *domain.BranchDepartment) BranchDeptResponse {
	return BranchDeptResponse{
		BranchID:       bd.BranchID,
		DepartmentID:   bd.DepartmentID,
		HeadEmployeeID: bd.HeadEmployeeID,
		IsActive:       bd.IsActive,
		CreatedAt:      bd.CreatedAt.Format(time.RFC3339),
	}
}
