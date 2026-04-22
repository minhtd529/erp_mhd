package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Provisioning errors
var (
	ErrDuplicatePendingRequest    = errors.New("DUPLICATE_PENDING_REQUEST")
	ErrRequestNotFound            = errors.New("REQUEST_NOT_FOUND")
	ErrInvalidRoleForProvisioning = errors.New("INVALID_ROLE_FOR_PROVISIONING")
	ErrRequestExpired             = errors.New("REQUEST_EXPIRED")
	ErrRequestAlreadyExecuted     = errors.New("REQUEST_ALREADY_EXECUTED")
	ErrOffboardingNotFound        = errors.New("OFFBOARDING_NOT_FOUND")
	ErrInvalidRequestStatus       = errors.New("INVALID_REQUEST_STATUS")
)

// ProvisioningRequest maps to user_provisioning_requests.
type ProvisioningRequest struct {
	ID                    uuid.UUID
	EmployeeID            uuid.UUID
	RequestedBy           uuid.UUID
	RequestedRole         string
	RequestedBranchID     *uuid.UUID
	Status                string
	ApprovalLevel         int16
	BranchApproverID      *uuid.UUID
	BranchApprovedAt      *time.Time
	BranchRejectionReason *string
	HRApproverID          *uuid.UUID
	HRApprovedAt          *time.Time
	HRRejectionReason     *string
	ExecutedBy            *uuid.UUID
	ExecutedAt            *time.Time
	IsEmergency           bool
	EmergencyReason       *string
	Notes                 *string
	ExpiresAt             time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// OffboardingChecklist maps to offboarding_checklists.
type OffboardingChecklist struct {
	ID            uuid.UUID
	EmployeeID    uuid.UUID
	ChecklistType string
	InitiatedBy   uuid.UUID
	TargetDate    *time.Time
	Items         []byte
	Status        string
	CompletedAt   *time.Time
	Notes         *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ─── Params ───────────────────────────────────────────────────────────────────

type CreateProvisioningParams struct {
	EmployeeID        uuid.UUID
	RequestedBy       uuid.UUID
	RequestedRole     string
	RequestedBranchID *uuid.UUID
	IsEmergency       bool
	EmergencyReason   *string
	Notes             *string
}

type ListProvisioningFilter struct {
	Status     string
	EmployeeID *uuid.UUID
	Page       int
	Size       int
}

type BranchApproveParams struct {
	RequestID  uuid.UUID
	ApproverID uuid.UUID
}

type HRApproveParams struct {
	RequestID  uuid.UUID
	ApproverID uuid.UUID
}

type RejectParams struct {
	RequestID  uuid.UUID
	RejectedBy uuid.UUID
	Reason     string
}

type ExecuteParams struct {
	RequestID  uuid.UUID
	ExecutorID uuid.UUID
}

// ExecuteResult is returned after atomic user provisioning.
type ExecuteResult struct {
	Request      *ProvisioningRequest
	UserID       uuid.UUID
	TempPassword string // plain text, returned once; never persisted
}

type CreateOffboardingParams struct {
	EmployeeID    uuid.UUID
	ChecklistType string
	InitiatedBy   uuid.UUID
	TargetDate    *time.Time
	Notes         *string
}

type UpdateOffboardingItemParams struct {
	ChecklistID uuid.UUID
	ItemKey     string
	Completed   bool
	CompletedBy *uuid.UUID
	Notes       *string
}

type ListOffboardingFilter struct {
	Status     string
	EmployeeID *uuid.UUID
	Page       int
	Size       int
}

// ─── Repository interfaces ────────────────────────────────────────────────────

// ProvisioningExpiredAlert is returned by ListExpiredPending for the HRM reminder job.
type ProvisioningExpiredAlert struct {
	RequestID   uuid.UUID
	EmployeeID  uuid.UUID
	RequestedBy uuid.UUID // notify the submitter
	ExpiresAt   time.Time
}

// ProvisioningRepository defines data access for user_provisioning_requests.
type ProvisioningRepository interface {
	Create(ctx context.Context, p CreateProvisioningParams) (*ProvisioningRequest, error)
	FindByID(ctx context.Context, id uuid.UUID) (*ProvisioningRequest, error)
	List(ctx context.Context, f ListProvisioningFilter) ([]*ProvisioningRequest, int64, error)
	HasPendingForEmployee(ctx context.Context, employeeID uuid.UUID) (bool, error)
	BranchApprove(ctx context.Context, p BranchApproveParams) (*ProvisioningRequest, error)
	BranchReject(ctx context.Context, p RejectParams) (*ProvisioningRequest, error)
	HRApprove(ctx context.Context, p HRApproveParams) (*ProvisioningRequest, error)
	HRReject(ctx context.Context, p RejectParams) (*ProvisioningRequest, error)
	Cancel(ctx context.Context, requestID, callerID uuid.UUID) (*ProvisioningRequest, error)
	MarkExecuted(ctx context.Context, requestID, executorID uuid.UUID) (*ProvisioningRequest, error)
	// ListExpiredPending returns PENDING requests whose expires_at has passed.
	// Used by the HRM daily reminder job to notify the requester.
	ListExpiredPending(ctx context.Context) ([]ProvisioningExpiredAlert, error)
}

// OffboardingRepository defines data access for offboarding_checklists.
type OffboardingRepository interface {
	Create(ctx context.Context, p CreateOffboardingParams) (*OffboardingChecklist, error)
	FindByID(ctx context.Context, id uuid.UUID) (*OffboardingChecklist, error)
	List(ctx context.Context, f ListOffboardingFilter) ([]*OffboardingChecklist, int64, error)
	UpdateItem(ctx context.Context, p UpdateOffboardingItemParams) (*OffboardingChecklist, error)
	Complete(ctx context.Context, checklistID, callerID uuid.UUID) (*OffboardingChecklist, error)
}
