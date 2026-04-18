package usecase

import (
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/engagement/domain"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

// ─── Engagement DTOs ─────────────────────────────────────────────────────────

type EngagementCreateRequest struct {
	ClientID             uuid.UUID          `json:"client_id"              binding:"required"`
	ServiceType          domain.ServiceType `json:"service_type"           binding:"required"`
	FeeType              domain.FeeType     `json:"fee_type"               binding:"required"`
	FeeAmount            float64            `json:"fee_amount"             binding:"min=0"`
	PartnerID            *uuid.UUID         `json:"partner_id"`
	PrimarySalespersonID *uuid.UUID         `json:"primary_salesperson_id"`
	StartDate            *time.Time         `json:"start_date"`
	EndDate              *time.Time         `json:"end_date"`
	Description          *string            `json:"description"`
}

type EngagementUpdateRequest struct {
	ServiceType          domain.ServiceType `json:"service_type"           binding:"required"`
	FeeType              domain.FeeType     `json:"fee_type"               binding:"required"`
	FeeAmount            float64            `json:"fee_amount"             binding:"min=0"`
	PartnerID            *uuid.UUID         `json:"partner_id"`
	PrimarySalespersonID *uuid.UUID         `json:"primary_salesperson_id"`
	StartDate            *time.Time         `json:"start_date"`
	EndDate              *time.Time         `json:"end_date"`
	Description          *string            `json:"description"`
}

type EngagementListRequest struct {
	Page     int                      `form:"page,default=1"  binding:"min=1"`
	Size     int                      `form:"size,default=20" binding:"min=1,max=100"`
	ClientID *uuid.UUID               `form:"client_id"`
	Status   domain.EngagementStatus  `form:"status"`
	Q        string                   `form:"q"`
}

type EngagementResponse struct {
	ID                   uuid.UUID              `json:"id"`
	ClientID             uuid.UUID              `json:"client_id"`
	ServiceType          domain.ServiceType     `json:"service_type"`
	FeeType              domain.FeeType         `json:"fee_type"`
	FeeAmount            float64                `json:"fee_amount"`
	Status               domain.EngagementStatus `json:"status"`
	PartnerID            *uuid.UUID             `json:"partner_id"`
	PrimarySalespersonID *uuid.UUID             `json:"primary_salesperson_id"`
	StartDate            *time.Time             `json:"start_date"`
	EndDate              *time.Time             `json:"end_date"`
	Description          *string                `json:"description"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
	CreatedBy            uuid.UUID              `json:"created_by"`
	UpdatedBy            uuid.UUID              `json:"updated_by"`
}

// ─── Member DTOs ─────────────────────────────────────────────────────────────

type MemberAssignRequest struct {
	StaffID           uuid.UUID          `json:"staff_id"           binding:"required"`
	Role              domain.MemberRole  `json:"role"               binding:"required"`
	HourlyRate        *float64           `json:"hourly_rate"`
	AllocationPercent int                `json:"allocation_percent" binding:"min=1,max=100"`
}

type MemberUpdateRequest struct {
	Role              domain.MemberRole `json:"role"               binding:"required"`
	HourlyRate        *float64          `json:"hourly_rate"`
	AllocationPercent int               `json:"allocation_percent" binding:"min=1,max=100"`
}

type MemberResponse struct {
	ID                uuid.UUID           `json:"id"`
	EngagementID      uuid.UUID           `json:"engagement_id"`
	StaffID           uuid.UUID           `json:"staff_id"`
	Role              domain.MemberRole   `json:"role"`
	HourlyRate        *float64            `json:"hourly_rate"`
	AllocationPercent int                 `json:"allocation_percent"`
	Status            domain.MemberStatus `json:"status"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
	CreatedBy         uuid.UUID           `json:"created_by"`
	UpdatedBy         uuid.UUID           `json:"updated_by"`
}

// ─── Task DTOs ───────────────────────────────────────────────────────────────

type TaskCreateRequest struct {
	Phase      domain.TaskPhase `json:"phase"       binding:"required"`
	Title      string           `json:"title"       binding:"required"`
	AssignedTo *uuid.UUID       `json:"assigned_to"`
	DueDate    *time.Time       `json:"due_date"`
}

type TaskUpdateRequest struct {
	Title      string           `json:"title"       binding:"required"`
	AssignedTo *uuid.UUID       `json:"assigned_to"`
	Status     domain.TaskStatus `json:"status"      binding:"required"`
	DueDate    *time.Time       `json:"due_date"`
}

type TaskResponse struct {
	ID           uuid.UUID         `json:"id"`
	EngagementID uuid.UUID         `json:"engagement_id"`
	Phase        domain.TaskPhase  `json:"phase"`
	Title        string            `json:"title"`
	AssignedTo   *uuid.UUID        `json:"assigned_to"`
	Status       domain.TaskStatus `json:"status"`
	DueDate      *time.Time        `json:"due_date"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	CreatedBy    uuid.UUID         `json:"created_by"`
	UpdatedBy    uuid.UUID         `json:"updated_by"`
}

// ─── Cost DTOs ───────────────────────────────────────────────────────────────

type CostCreateRequest struct {
	CostType    domain.CostType `json:"cost_type"   binding:"required"`
	Description string          `json:"description" binding:"required"`
	Amount      float64         `json:"amount"      binding:"required,gt=0"`
	ReceiptURL  *string         `json:"receipt_url"`
}

type CostRejectRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type CostResponse struct {
	ID           uuid.UUID          `json:"id"`
	EngagementID uuid.UUID          `json:"engagement_id"`
	CostType     domain.CostType    `json:"cost_type"`
	Description  string             `json:"description"`
	Amount       float64            `json:"amount"`
	Status       domain.CostStatus  `json:"status"`
	ReceiptURL   *string            `json:"receipt_url"`
	SubmittedAt  *time.Time         `json:"submitted_at"`
	SubmittedBy  *uuid.UUID         `json:"submitted_by"`
	ApprovedAt   *time.Time         `json:"approved_at"`
	ApprovedBy   *uuid.UUID         `json:"approved_by"`
	RejectReason *string            `json:"reject_reason"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	CreatedBy    uuid.UUID          `json:"created_by"`
	UpdatedBy    uuid.UUID          `json:"updated_by"`
}

// ─── Pagination ──────────────────────────────────────────────────────────────

// PaginatedResult is the shared offset pagination type.
type PaginatedResult[T any] = pagination.OffsetResult[T]

func newPaginatedResult[T any](data []T, total int64, page, size int) PaginatedResult[T] {
	return pagination.NewOffsetResult(data, total, page, size)
}

// EngagementCursorListRequest is the body for cursor-paginated GET /engagements.
type EngagementCursorListRequest struct {
	Cursor   string                  `form:"cursor"`
	Size     int                     `form:"size,default=20" binding:"min=1,max=100"`
	ClientID *uuid.UUID              `form:"client_id"`
	Status   domain.EngagementStatus `form:"status"`
	Q        string                  `form:"q"`
}

// CursorResult wraps cursor-paginated engagement results.
type EngagementCursorResult = pagination.CursorResult[EngagementResponse]

// ─── Converters ──────────────────────────────────────────────────────────────

func toEngagementResponse(e *domain.Engagement) EngagementResponse {
	return EngagementResponse{
		ID: e.ID, ClientID: e.ClientID, ServiceType: e.ServiceType,
		FeeType: e.FeeType, FeeAmount: e.FeeAmount, Status: e.Status,
		PartnerID: e.PartnerID, PrimarySalespersonID: e.PrimarySalespersonID,
		StartDate: e.StartDate, EndDate: e.EndDate, Description: e.Description,
		CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
		CreatedBy: e.CreatedBy, UpdatedBy: e.UpdatedBy,
	}
}

func toMemberResponse(m *domain.EngagementMember) MemberResponse {
	return MemberResponse{
		ID: m.ID, EngagementID: m.EngagementID, StaffID: m.StaffID,
		Role: m.Role, HourlyRate: m.HourlyRate, AllocationPercent: m.AllocationPercent,
		Status: m.Status, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
		CreatedBy: m.CreatedBy, UpdatedBy: m.UpdatedBy,
	}
}

func toTaskResponse(t *domain.EngagementTask) TaskResponse {
	return TaskResponse{
		ID: t.ID, EngagementID: t.EngagementID, Phase: t.Phase,
		Title: t.Title, AssignedTo: t.AssignedTo, Status: t.Status,
		DueDate: t.DueDate, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
		CreatedBy: t.CreatedBy, UpdatedBy: t.UpdatedBy,
	}
}

func toCostResponse(c *domain.DirectCost) CostResponse {
	return CostResponse{
		ID: c.ID, EngagementID: c.EngagementID, CostType: c.CostType,
		Description: c.Description, Amount: c.Amount, Status: c.Status,
		ReceiptURL: c.ReceiptURL, SubmittedAt: c.SubmittedAt, SubmittedBy: c.SubmittedBy,
		ApprovedAt: c.ApprovedAt, ApprovedBy: c.ApprovedBy, RejectReason: c.RejectReason,
		CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
		CreatedBy: c.CreatedBy, UpdatedBy: c.UpdatedBy,
	}
}
