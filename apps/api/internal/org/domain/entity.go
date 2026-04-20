package domain

import (
	"time"

	"github.com/google/uuid"
)

// ─── Branch ──────────────────────────────────────────────────────────────────

type Branch struct {
	ID        uuid.UUID  `json:"id"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	Address   *string    `json:"address"`
	Phone     *string    `json:"phone"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy *uuid.UUID `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by"`
}

type CreateBranchParams struct {
	Code      string
	Name      string
	Address   *string
	Phone     *string
	CreatedBy *uuid.UUID
}

type UpdateBranchParams struct {
	ID        uuid.UUID
	Code      string
	Name      string
	Address   *string
	Phone     *string
	IsActive  bool
	UpdatedBy *uuid.UUID
}

type ListBranchesFilter struct {
	Page     int
	Size     int
	IsActive *bool
	Q        string
}

// ─── Department ──────────────────────────────────────────────────────────────

type Department struct {
	ID        uuid.UUID  `json:"id"`
	BranchID  *uuid.UUID `json:"branch_id"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy *uuid.UUID `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by"`
}

type CreateDepartmentParams struct {
	BranchID  *uuid.UUID
	Code      string
	Name      string
	CreatedBy *uuid.UUID
}

type UpdateDepartmentParams struct {
	ID        uuid.UUID
	Code      string
	Name      string
	BranchID  *uuid.UUID
	IsActive  bool
	UpdatedBy *uuid.UUID
}

type ListDepartmentsFilter struct {
	Page     int
	Size     int
	BranchID *uuid.UUID
	IsActive *bool
	Q        string
}
