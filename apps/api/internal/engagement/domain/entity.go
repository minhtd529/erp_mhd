// Package domain defines the Engagement bounded context aggregates.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// EngagementStatus represents the lifecycle stage of an engagement.
type EngagementStatus string

const (
	StatusDraft      EngagementStatus = "DRAFT"
	StatusProposal   EngagementStatus = "PROPOSAL"
	StatusContracted EngagementStatus = "CONTRACTED"
	StatusActive     EngagementStatus = "ACTIVE"
	StatusCompleted  EngagementStatus = "COMPLETED"
	StatusSettled    EngagementStatus = "SETTLED"
)

// ServiceType enumerates the kinds of audit/advisory services.
type ServiceType string

const (
	ServiceAudit            ServiceType = "AUDIT"
	ServiceReview           ServiceType = "REVIEW"
	ServiceCompilation      ServiceType = "COMPILATION"
	ServiceTaxAdvisory      ServiceType = "TAX_ADVISORY"
	ServiceBusinessAdvisory ServiceType = "BUSINESS_ADVISORY"
)

// FeeType enumerates billing models.
type FeeType string

const (
	FeeFixed          FeeType = "FIXED"
	FeeTimeAndMaterial FeeType = "TIME_AND_MATERIAL"
	FeeRetainer       FeeType = "RETAINER"
	FeeSuccess        FeeType = "SUCCESS"
)

// Engagement is the aggregate root for the engagement bounded context.
type Engagement struct {
	ID                   uuid.UUID        `json:"id"                     db:"id"`
	ClientID             uuid.UUID        `json:"client_id"              db:"client_id"`
	ServiceType          ServiceType      `json:"service_type"           db:"service_type"`
	FeeType              FeeType          `json:"fee_type"               db:"fee_type"`
	FeeAmount            float64          `json:"fee_amount"             db:"fee_amount"`
	Status               EngagementStatus `json:"status"                 db:"status"`
	PartnerID            *uuid.UUID       `json:"partner_id"             db:"partner_id"`
	PrimarySalespersonID *uuid.UUID       `json:"primary_salesperson_id" db:"primary_salesperson_id"`
	StartDate            *time.Time       `json:"start_date"             db:"start_date"`
	EndDate              *time.Time       `json:"end_date"               db:"end_date"`
	Description          *string          `json:"description"            db:"description"`
	IsDeleted            bool             `json:"is_deleted"             db:"is_deleted"`
	CreatedAt            time.Time        `json:"created_at"             db:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"             db:"updated_at"`
	CreatedBy            uuid.UUID        `json:"created_by"             db:"created_by"`
	UpdatedBy            uuid.UUID        `json:"updated_by"             db:"updated_by"`
}

// MemberRole enumerates team member roles within an engagement.
type MemberRole string

const (
	RolePartner       MemberRole = "PARTNER"
	RoleManager       MemberRole = "MANAGER"
	RoleSeniorAuditor MemberRole = "SENIOR_AUDITOR"
	RoleAuditor       MemberRole = "AUDITOR"
	RoleIntern        MemberRole = "INTERN"
)

// MemberStatus enumerates the assignment lifecycle.
type MemberStatus string

const (
	MemberAssigned  MemberStatus = "ASSIGNED"
	MemberActive    MemberStatus = "ACTIVE"
	MemberCompleted MemberStatus = "COMPLETED"
)

// EngagementMember represents a team member assigned to an engagement.
type EngagementMember struct {
	ID               uuid.UUID    `json:"id"                db:"id"`
	EngagementID     uuid.UUID    `json:"engagement_id"     db:"engagement_id"`
	StaffID          uuid.UUID    `json:"staff_id"          db:"staff_id"`
	Role             MemberRole   `json:"role"              db:"role"`
	HourlyRate       *float64     `json:"hourly_rate"       db:"hourly_rate"`
	AllocationPercent int         `json:"allocation_percent" db:"allocation_percent"`
	Status           MemberStatus `json:"status"            db:"status"`
	IsDeleted        bool         `json:"is_deleted"        db:"is_deleted"`
	CreatedAt        time.Time    `json:"created_at"        db:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"        db:"updated_at"`
	CreatedBy        uuid.UUID    `json:"created_by"        db:"created_by"`
	UpdatedBy        uuid.UUID    `json:"updated_by"        db:"updated_by"`
}

// TaskPhase enumerates engagement phases.
type TaskPhase string

const (
	PhasePlanning  TaskPhase = "PLANNING"
	PhaseFieldwork TaskPhase = "FIELDWORK"
	PhaseReporting TaskPhase = "REPORTING"
)

// TaskStatus enumerates task progress states.
type TaskStatus string

const (
	TaskNotStarted TaskStatus = "NOT_STARTED"
	TaskInProgress TaskStatus = "IN_PROGRESS"
	TaskCompleted  TaskStatus = "COMPLETED"
)

// EngagementTask represents a task within a phase of an engagement.
type EngagementTask struct {
	ID           uuid.UUID  `json:"id"            db:"id"`
	EngagementID uuid.UUID  `json:"engagement_id" db:"engagement_id"`
	Phase        TaskPhase  `json:"phase"         db:"phase"`
	Title        string     `json:"title"         db:"title"`
	AssignedTo   *uuid.UUID `json:"assigned_to"   db:"assigned_to"`
	Status       TaskStatus `json:"status"        db:"status"`
	DueDate      *time.Time `json:"due_date"      db:"due_date"`
	IsDeleted    bool       `json:"is_deleted"    db:"is_deleted"`
	CreatedAt    time.Time  `json:"created_at"    db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"    db:"updated_at"`
	CreatedBy    uuid.UUID  `json:"created_by"    db:"created_by"`
	UpdatedBy    uuid.UUID  `json:"updated_by"    db:"updated_by"`
}

// CostType enumerates direct cost categories.
type CostType string

const (
	CostTravel        CostType = "TRAVEL"
	CostAccommodation CostType = "ACCOMMODATION"
	CostMeals         CostType = "MEALS"
	CostMaterials     CostType = "MATERIALS"
	CostOther         CostType = "OTHER"
)

// CostStatus enumerates the approval lifecycle of a direct cost.
type CostStatus string

const (
	CostDraft     CostStatus = "DRAFT"
	CostSubmitted CostStatus = "SUBMITTED"
	CostApproved  CostStatus = "APPROVED"
	CostRejected  CostStatus = "REJECTED"
)

// DirectCost represents a reimbursable expense on an engagement.
type DirectCost struct {
	ID           uuid.UUID  `json:"id"            db:"id"`
	EngagementID uuid.UUID  `json:"engagement_id" db:"engagement_id"`
	CostType     CostType   `json:"cost_type"     db:"cost_type"`
	Description  string     `json:"description"   db:"description"`
	Amount       float64    `json:"amount"        db:"amount"`
	Status       CostStatus `json:"status"        db:"status"`
	ReceiptURL   *string    `json:"receipt_url"   db:"receipt_url"`
	SubmittedAt  *time.Time `json:"submitted_at"  db:"submitted_at"`
	SubmittedBy  *uuid.UUID `json:"submitted_by"  db:"submitted_by"`
	ApprovedAt   *time.Time `json:"approved_at"   db:"approved_at"`
	ApprovedBy   *uuid.UUID `json:"approved_by"   db:"approved_by"`
	RejectReason *string    `json:"reject_reason" db:"reject_reason"`
	IsDeleted    bool       `json:"is_deleted"    db:"is_deleted"`
	CreatedAt    time.Time  `json:"created_at"    db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"    db:"updated_at"`
	CreatedBy    uuid.UUID  `json:"created_by"    db:"created_by"`
	UpdatedBy    uuid.UUID  `json:"updated_by"    db:"updated_by"`
}

// ─── Command params ───────────────────────────────────────────────────────────

type CreateEngagementParams struct {
	ClientID             uuid.UUID
	ServiceType          ServiceType
	FeeType              FeeType
	FeeAmount            float64
	PartnerID            *uuid.UUID
	PrimarySalespersonID *uuid.UUID
	StartDate            *time.Time
	EndDate              *time.Time
	Description          *string
	CreatedBy            uuid.UUID
}

type UpdateEngagementParams struct {
	ID                   uuid.UUID
	ServiceType          ServiceType
	FeeType              FeeType
	FeeAmount            float64
	PartnerID            *uuid.UUID
	PrimarySalespersonID *uuid.UUID
	StartDate            *time.Time
	EndDate              *time.Time
	Description          *string
	UpdatedBy            uuid.UUID
}

type ListEngagementsFilter struct {
	Page     int
	Size     int
	ClientID *uuid.UUID
	Status   EngagementStatus
	Q        string
}

// CursorFilter is used by cursor-based list queries.
type CursorFilter struct {
	// AfterID and AfterCreatedAt define the exclusive start position.
	AfterID        *uuid.UUID
	AfterCreatedAt *time.Time
	Size           int
	ClientID       *uuid.UUID
	Status         EngagementStatus
	Q              string
}

type AssignMemberParams struct {
	EngagementID      uuid.UUID
	StaffID           uuid.UUID
	Role              MemberRole
	HourlyRate        *float64
	AllocationPercent int
	CreatedBy         uuid.UUID
}

type UpdateMemberParams struct {
	ID                uuid.UUID
	EngagementID      uuid.UUID
	Role              MemberRole
	HourlyRate        *float64
	AllocationPercent int
	UpdatedBy         uuid.UUID
}

type CreateTaskParams struct {
	EngagementID uuid.UUID
	Phase        TaskPhase
	Title        string
	AssignedTo   *uuid.UUID
	DueDate      *time.Time
	CreatedBy    uuid.UUID
}

type UpdateTaskParams struct {
	ID           uuid.UUID
	EngagementID uuid.UUID
	Title        string
	AssignedTo   *uuid.UUID
	Status       TaskStatus
	DueDate      *time.Time
	UpdatedBy    uuid.UUID
}

type CreateCostParams struct {
	EngagementID uuid.UUID
	CostType     CostType
	Description  string
	Amount       float64
	ReceiptURL   *string
	CreatedBy    uuid.UUID
}
