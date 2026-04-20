package usecase

import (
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

// ─── Branch DTOs ──────────────────────────────────────────────────────────────

type BranchCreateRequest struct {
	Code    string  `json:"code"    binding:"required,max=20"`
	Name    string  `json:"name"    binding:"required,max=200"`
	Address *string `json:"address"`
	Phone   *string `json:"phone"`
}

type BranchUpdateRequest struct {
	Code     string  `json:"code"      binding:"required,max=20"`
	Name     string  `json:"name"      binding:"required,max=200"`
	Address  *string `json:"address"`
	Phone    *string `json:"phone"`
	IsActive bool    `json:"is_active"`
}

type BranchListRequest struct {
	Page     int    `form:"page,default=1"  binding:"min=1"`
	Size     int    `form:"size,default=20" binding:"min=1,max=100"`
	IsActive *bool  `form:"is_active"`
	Q        string `form:"q"`
}

type BranchResponse struct {
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

// ─── Department DTOs ──────────────────────────────────────────────────────────

type DepartmentCreateRequest struct {
	Code     string     `json:"code"      binding:"required,max=20"`
	Name     string     `json:"name"      binding:"required,max=200"`
	BranchID *uuid.UUID `json:"branch_id"`
}

type DepartmentUpdateRequest struct {
	Code     string     `json:"code"      binding:"required,max=20"`
	Name     string     `json:"name"      binding:"required,max=200"`
	BranchID *uuid.UUID `json:"branch_id"`
	IsActive bool       `json:"is_active"`
}

type DepartmentListRequest struct {
	Page     int        `form:"page,default=1"  binding:"min=1"`
	Size     int        `form:"size,default=20" binding:"min=1,max=100"`
	BranchID *uuid.UUID `form:"branch_id"`
	IsActive *bool      `form:"is_active"`
	Q        string     `form:"q"`
}

type DepartmentResponse struct {
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

// ─── Shared ───────────────────────────────────────────────────────────────────

type PaginatedResult[T any] = pagination.OffsetResult[T]
