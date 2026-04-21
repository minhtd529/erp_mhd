package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ─── HRM-extended Branch ──────────────────────────────────────────────────────

// HRMBranch includes the original branches columns plus HRM fields added in migration 000020.
type HRMBranch struct {
	ID                     uuid.UUID  `db:"id"`
	Code                   string     `db:"code"`
	Name                   string     `db:"name"`
	Address                *string    `db:"address"`
	Phone                  *string    `db:"phone"`
	IsActive               bool       `db:"is_active"`
	IsHeadOffice           bool       `db:"is_head_office"`
	City                   *string    `db:"city"`
	EstablishedDate        *time.Time `db:"established_date"`
	HeadOfBranchUserID     *uuid.UUID `db:"head_of_branch_user_id"`
	TaxCode                *string    `db:"tax_code"`
	AuthorizationDocNumber *string    `db:"authorization_doc_number"`
	AuthorizationDate      *time.Time `db:"authorization_date"`
	AuthorizationFileID    *uuid.UUID `db:"authorization_file_id"`
	CreatedAt              time.Time  `db:"created_at"`
	UpdatedAt              time.Time  `db:"updated_at"`
	CreatedBy              *uuid.UUID `db:"created_by"`
	UpdatedBy              *uuid.UUID `db:"updated_by"`
}

type UpdateHRMBranchParams struct {
	ID                     uuid.UUID
	Name                   *string
	Address                *string
	Phone                  *string
	City                   *string
	TaxCode                *string
	EstablishedDate        *time.Time
	AuthorizationDocNumber *string
	AuthorizationDate      *time.Time
	UpdatedBy              *uuid.UUID
}

type ListHRMBranchesFilter struct {
	Page     int
	Size     int
	IsActive *bool
	Q        string
}

// ─── HRM-extended Department ──────────────────────────────────────────────────

// HRMDepartment includes the original departments columns plus HRM fields added in migration 000020.
type HRMDepartment struct {
	ID                     uuid.UUID  `db:"id"`
	Code                   string     `db:"code"`
	Name                   string     `db:"name"`
	BranchID               *uuid.UUID `db:"branch_id"`
	IsActive               bool       `db:"is_active"`
	IsDeleted              bool       `db:"is_deleted"`
	Description            *string    `db:"description"`
	DeptType               string     `db:"dept_type"`
	HeadEmployeeID         *uuid.UUID `db:"head_employee_id"`
	AuthorizationDocNumber *string    `db:"authorization_doc_number"`
	AuthorizationDate      *time.Time `db:"authorization_date"`
	AuthorizationFileID    *uuid.UUID `db:"authorization_file_id"`
	CreatedAt              time.Time  `db:"created_at"`
	UpdatedAt              time.Time  `db:"updated_at"`
	CreatedBy              *uuid.UUID `db:"created_by"`
	UpdatedBy              *uuid.UUID `db:"updated_by"`
}

type UpdateHRMDeptParams struct {
	ID                     uuid.UUID
	Name                   *string
	Description            *string
	DeptType               *string
	AuthorizationDocNumber *string
	AuthorizationDate      *time.Time
	UpdatedBy              *uuid.UUID
}

type ListHRMDeptsFilter struct {
	Page     int
	Size     int
	BranchID *uuid.UUID
	IsActive *bool
	Q        string
}

// ─── BranchDepartment ────────────────────────────────────────────────────────

type BranchDepartment struct {
	BranchID       uuid.UUID  `db:"branch_id"`
	DepartmentID   uuid.UUID  `db:"department_id"`
	HeadEmployeeID *uuid.UUID `db:"head_employee_id"`
	IsActive       bool       `db:"is_active"`
	CreatedAt      time.Time  `db:"created_at"`
}

type CreateBranchDeptParams struct {
	BranchID     uuid.UUID
	DepartmentID uuid.UUID
	CreatedBy    *uuid.UUID
}

type ListBranchDeptsFilter struct {
	Page     int
	Size     int
	BranchID *uuid.UUID
	IsActive *bool
}

// ─── Org Chart ────────────────────────────────────────────────────────────────

// OrgChartBranch is a branch node in the org chart tree.
type OrgChartBranch struct {
	ID          uuid.UUID          `json:"id"`
	Code        string             `json:"code"`
	Name        string             `json:"name"`
	IsHeadOffice bool              `json:"is_head_office"`
	Departments []OrgChartDept     `json:"departments"`
}

// OrgChartDept is a department leaf in the org chart tree.
type OrgChartDept struct {
	ID       uuid.UUID `json:"id"`
	Code     string    `json:"code"`
	Name     string    `json:"name"`
	DeptType string    `json:"dept_type"`
}

// AssignBranchHeadParams carries the user to assign as head-of-branch.
type AssignBranchHeadParams struct {
	BranchID  uuid.UUID
	UserID    uuid.UUID
	UpdatedBy *uuid.UUID
}

// AssignDeptHeadParams carries the employee to assign as department head.
type AssignDeptHeadParams struct {
	DeptID     uuid.UUID
	EmployeeID uuid.UUID
	UpdatedBy  *uuid.UUID
}

// ─── Repository interface ────────────────────────────────────────────────────

// OrgRepository defines the data-access contract for HRM organization operations.
type OrgRepository interface {
	FindBranchByID(ctx context.Context, id uuid.UUID) (*HRMBranch, error)
	ListBranches(ctx context.Context, f ListHRMBranchesFilter) ([]*HRMBranch, int64, error)
	UpdateBranch(ctx context.Context, p UpdateHRMBranchParams) (*HRMBranch, error)
	AssignBranchHead(ctx context.Context, p AssignBranchHeadParams) (*HRMBranch, error)
	DeactivateBranch(ctx context.Context, id uuid.UUID, updatedBy *uuid.UUID) error

	FindDeptByID(ctx context.Context, id uuid.UUID) (*HRMDepartment, error)
	ListDepts(ctx context.Context, f ListHRMDeptsFilter) ([]*HRMDepartment, int64, error)
	UpdateDept(ctx context.Context, p UpdateHRMDeptParams) (*HRMDepartment, error)
	AssignDeptHead(ctx context.Context, p AssignDeptHeadParams) (*HRMDepartment, error)
	DeactivateDept(ctx context.Context, id uuid.UUID, updatedBy *uuid.UUID) error

	ListBranchDepts(ctx context.Context, f ListBranchDeptsFilter) ([]*BranchDepartment, int64, error)
	CreateBranchDept(ctx context.Context, p CreateBranchDeptParams) (*BranchDepartment, error)
	SoftDeleteBranchDept(ctx context.Context, branchID, deptID uuid.UUID) error

	ListBranchesWithDepts(ctx context.Context) ([]*OrgChartBranch, error)
}
