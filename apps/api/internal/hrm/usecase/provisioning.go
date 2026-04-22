package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/hrm/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

// blockedRoles are role codes that cannot be provisioned through this workflow
// (SPEC §8.4: SUPER_ADMIN and CHAIRMAN can only be assigned by SA directly).
var blockedRoles = map[string]struct{}{
	"SUPER_ADMIN": {},
	"CHAIRMAN":    {},
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type ProvisioningRequestResponse struct {
	ID                    uuid.UUID  `json:"id"`
	EmployeeID            uuid.UUID  `json:"employee_id"`
	RequestedBy           uuid.UUID  `json:"requested_by"`
	RequestedRole         string     `json:"requested_role"`
	RequestedBranchID     *uuid.UUID `json:"requested_branch_id,omitempty"`
	Status                string     `json:"status"`
	ApprovalLevel         int16      `json:"approval_level"`
	BranchApproverID      *uuid.UUID `json:"branch_approver_id,omitempty"`
	BranchApprovedAt      *string    `json:"branch_approved_at,omitempty"`
	BranchRejectionReason *string    `json:"branch_rejection_reason,omitempty"`
	HRApproverID          *uuid.UUID `json:"hr_approver_id,omitempty"`
	HRApprovedAt          *string    `json:"hr_approved_at,omitempty"`
	HRRejectionReason     *string    `json:"hr_rejection_reason,omitempty"`
	ExecutedBy            *uuid.UUID `json:"executed_by,omitempty"`
	ExecutedAt            *string    `json:"executed_at,omitempty"`
	IsEmergency           bool       `json:"is_emergency"`
	EmergencyReason       *string    `json:"emergency_reason,omitempty"`
	Notes                 *string    `json:"notes,omitempty"`
	ExpiresAt             string     `json:"expires_at"`
	CreatedAt             string     `json:"created_at"`
	UpdatedAt             string     `json:"updated_at"`
}

type ExecuteProvisioningResponse struct {
	Request      ProvisioningRequestResponse `json:"request"`
	UserID       uuid.UUID                   `json:"user_id"`
	TempPassword string                      `json:"temp_password"`
}

type OffboardingResponse struct {
	ID            uuid.UUID  `json:"id"`
	EmployeeID    uuid.UUID  `json:"employee_id"`
	ChecklistType string     `json:"checklist_type"`
	InitiatedBy   uuid.UUID  `json:"initiated_by"`
	TargetDate    *string    `json:"target_date,omitempty"`
	Items         any        `json:"items"`
	Status        string     `json:"status"`
	CompletedAt   *string    `json:"completed_at,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
	CreatedAt     string     `json:"created_at"`
	UpdatedAt     string     `json:"updated_at"`
}

// ─── Request DTOs ─────────────────────────────────────────────────────────────

type CreateProvisioningRequest struct {
	EmployeeID        uuid.UUID  `json:"employee_id" binding:"required"`
	RequestedRole     string     `json:"requested_role" binding:"required"`
	RequestedBranchID *uuid.UUID `json:"requested_branch_id"`
	IsEmergency       bool       `json:"is_emergency"`
	EmergencyReason   *string    `json:"emergency_reason"`
	Notes             *string    `json:"notes"`
}

type ListProvisioningRequest struct {
	Status     string `form:"status"`
	EmployeeID string `form:"employee_id"`
	Page       int    `form:"page"`
	Size       int    `form:"size"`
}

type RejectRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type CreateOffboardingRequest struct {
	EmployeeID    uuid.UUID  `json:"employee_id" binding:"required"`
	ChecklistType string     `json:"checklist_type" binding:"required,oneof=ONBOARDING OFFBOARDING"`
	TargetDate    *string    `json:"target_date"`
	Notes         *string    `json:"notes"`
}

type ListOffboardingRequest struct {
	Status     string `form:"status"`
	EmployeeID string `form:"employee_id"`
	Page       int    `form:"page"`
	Size       int    `form:"size"`
}

type UpdateOffboardingItemRequest struct {
	Completed bool    `json:"completed"`
	Notes     *string `json:"notes"`
}

// ─── ProvisioningUseCase ─────────────────────────────────────────────────────

// ProvisioningUseCase holds provisioning + offboarding logic.
// pool is injected for the atomic Execute transaction.
type ProvisioningUseCase struct {
	repo         domain.ProvisioningRepository
	offRepo      domain.OffboardingRepository
	pool         *pgxpool.Pool
	auditLog     *audit.Logger
}

func NewProvisioningUseCase(
	repo domain.ProvisioningRepository,
	offRepo domain.OffboardingRepository,
	pool *pgxpool.Pool,
	auditLog *audit.Logger,
) *ProvisioningUseCase {
	return &ProvisioningUseCase{repo: repo, offRepo: offRepo, pool: pool, auditLog: auditLog}
}

// ─── Provisioning methods ─────────────────────────────────────────────────────

func (uc *ProvisioningUseCase) CreateRequest(
	ctx context.Context,
	req CreateProvisioningRequest,
	callerID uuid.UUID,
	ip string,
) (*ProvisioningRequestResponse, error) {
	// Blocked roles check (SPEC §8.4)
	if _, blocked := blockedRoles[req.RequestedRole]; blocked {
		return nil, domain.ErrInvalidRoleForProvisioning
	}

	// Emergency requires reason
	if req.IsEmergency && (req.EmergencyReason == nil || *req.EmergencyReason == "") {
		return nil, domain.ErrValidation
	}

	// Duplicate PENDING check
	hasPending, err := uc.repo.HasPendingForEmployee(ctx, req.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("CreateRequest: check pending: %w", err)
	}
	if hasPending {
		return nil, domain.ErrDuplicatePendingRequest
	}

	r, err := uc.repo.Create(ctx, domain.CreateProvisioningParams{
		EmployeeID:        req.EmployeeID,
		RequestedBy:       callerID,
		RequestedRole:     req.RequestedRole,
		RequestedBranchID: req.RequestedBranchID,
		IsEmergency:       req.IsEmergency,
		EmergencyReason:   req.EmergencyReason,
		Notes:             req.Notes,
	})
	if err != nil {
		return nil, fmt.Errorf("CreateRequest: %w", err)
	}

	action := "PROVISIONING_REQUESTED"
	if req.IsEmergency {
		action = "PROVISIONING_EMERGENCY"
	}
	uc.auditLog.Log(ctx, audit.Entry{ //nolint:errcheck
		UserID:     &callerID,
		Module:     "hrm",
		Resource:   "provisioning_request",
		ResourceID: &r.ID,
		Action:     action,
		NewValue:   r,
		IPAddress:  ip,
	})

	return toProvisioningResponse(r), nil
}

func (uc *ProvisioningUseCase) ListRequests(
	ctx context.Context,
	req ListProvisioningRequest,
) (pagination.OffsetResult[ProvisioningRequestResponse], error) {
	empty := pagination.OffsetResult[ProvisioningRequestResponse]{}

	f := domain.ListProvisioningFilter{
		Status: req.Status,
		Page:   req.Page,
		Size:   req.Size,
	}
	if req.EmployeeID != "" {
		id, err := uuid.Parse(req.EmployeeID)
		if err != nil {
			return empty, domain.ErrValidation
		}
		f.EmployeeID = &id
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Size < 1 || f.Size > 100 {
		f.Size = 20
	}

	items, total, err := uc.repo.List(ctx, f)
	if err != nil {
		return empty, fmt.Errorf("ListRequests: %w", err)
	}

	resp := make([]ProvisioningRequestResponse, len(items))
	for i, item := range items {
		resp[i] = *toProvisioningResponse(item)
	}
	return pagination.NewOffsetResult(resp, total, f.Page, f.Size), nil
}

func (uc *ProvisioningUseCase) GetRequest(
	ctx context.Context,
	id uuid.UUID,
) (*ProvisioningRequestResponse, error) {
	r, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toProvisioningResponse(r), nil
}

func (uc *ProvisioningUseCase) BranchApprove(
	ctx context.Context,
	requestID, callerID uuid.UUID,
	callerBranchID *uuid.UUID,
	ip string,
) (*ProvisioningRequestResponse, error) {
	req, err := uc.repo.FindByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req.Status != "PENDING" {
		return nil, domain.ErrInvalidRequestStatus
	}
	if time.Now().After(req.ExpiresAt) {
		return nil, domain.ErrRequestExpired
	}
	// Branch scope: HoB can only approve requests for their own branch
	if callerBranchID != nil && req.RequestedBranchID != nil && *callerBranchID != *req.RequestedBranchID {
		return nil, domain.ErrInsufficientPermission
	}

	r, err := uc.repo.BranchApprove(ctx, domain.BranchApproveParams{
		RequestID:  requestID,
		ApproverID: callerID,
	})
	if err != nil {
		return nil, err
	}

	uc.auditLog.Log(ctx, audit.Entry{ //nolint:errcheck
		UserID:     &callerID,
		Module:     "hrm",
		Resource:   "provisioning_request",
		ResourceID: &requestID,
		Action:     "PROVISIONING_BRANCH_APPROVED",
		NewValue:   r,
		IPAddress:  ip,
	})
	return toProvisioningResponse(r), nil
}

func (uc *ProvisioningUseCase) BranchReject(
	ctx context.Context,
	requestID, callerID uuid.UUID,
	callerBranchID *uuid.UUID,
	reason, ip string,
) (*ProvisioningRequestResponse, error) {
	req, err := uc.repo.FindByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req.Status != "PENDING" {
		return nil, domain.ErrInvalidRequestStatus
	}
	// Branch scope enforcement
	if callerBranchID != nil && req.RequestedBranchID != nil && *callerBranchID != *req.RequestedBranchID {
		return nil, domain.ErrInsufficientPermission
	}

	r, err := uc.repo.BranchReject(ctx, domain.RejectParams{
		RequestID:  requestID,
		RejectedBy: callerID,
		Reason:     reason,
	})
	if err != nil {
		return nil, err
	}

	uc.auditLog.Log(ctx, audit.Entry{ //nolint:errcheck
		UserID:     &callerID,
		Module:     "hrm",
		Resource:   "provisioning_request",
		ResourceID: &requestID,
		Action:     "PROVISIONING_BRANCH_REJECTED",
		NewValue:   r,
		IPAddress:  ip,
	})
	return toProvisioningResponse(r), nil
}

func (uc *ProvisioningUseCase) HRApprove(
	ctx context.Context,
	requestID, callerID uuid.UUID,
	ip string,
) (*ProvisioningRequestResponse, error) {
	req, err := uc.repo.FindByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req.Status != "PENDING" {
		return nil, domain.ErrInvalidRequestStatus
	}
	if time.Now().After(req.ExpiresAt) {
		return nil, domain.ErrRequestExpired
	}
	// HCM flow: branch approval must be done first if requested_branch_id is set
	if req.RequestedBranchID != nil && req.BranchApprovedAt == nil {
		return nil, domain.ErrInvalidRequestStatus
	}

	r, err := uc.repo.HRApprove(ctx, domain.HRApproveParams{
		RequestID:  requestID,
		ApproverID: callerID,
	})
	if err != nil {
		return nil, err
	}

	uc.auditLog.Log(ctx, audit.Entry{ //nolint:errcheck
		UserID:     &callerID,
		Module:     "hrm",
		Resource:   "provisioning_request",
		ResourceID: &requestID,
		Action:     "PROVISIONING_HR_APPROVED",
		NewValue:   r,
		IPAddress:  ip,
	})
	return toProvisioningResponse(r), nil
}

func (uc *ProvisioningUseCase) HRReject(
	ctx context.Context,
	requestID, callerID uuid.UUID,
	reason, ip string,
) (*ProvisioningRequestResponse, error) {
	req, err := uc.repo.FindByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req.Status != "PENDING" {
		return nil, domain.ErrInvalidRequestStatus
	}

	r, err := uc.repo.HRReject(ctx, domain.RejectParams{
		RequestID:  requestID,
		RejectedBy: callerID,
		Reason:     reason,
	})
	if err != nil {
		return nil, err
	}

	uc.auditLog.Log(ctx, audit.Entry{ //nolint:errcheck
		UserID:     &callerID,
		Module:     "hrm",
		Resource:   "provisioning_request",
		ResourceID: &requestID,
		Action:     "PROVISIONING_HR_REJECTED",
		NewValue:   r,
		IPAddress:  ip,
	})
	return toProvisioningResponse(r), nil
}

// Execute atomically creates a system user, assigns the requested role,
// links employee.user_id, and marks the request EXECUTED.
// SPEC §8.3: Execute is only allowed when status=APPROVED or is_emergency=true.
func (uc *ProvisioningUseCase) Execute(
	ctx context.Context,
	requestID, callerID uuid.UUID,
	employeeFullName, employeeEmail string,
	ip string,
) (*ExecuteProvisioningResponse, error) {
	req, err := uc.repo.FindByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req.Status == "EXECUTED" {
		return nil, domain.ErrRequestAlreadyExecuted
	}
	if req.Status == "CANCELLED" || req.Status == "REJECTED" {
		return nil, domain.ErrInvalidRequestStatus
	}
	// Non-emergency: must be APPROVED
	if !req.IsEmergency && req.Status != "APPROVED" {
		return nil, domain.ErrInvalidRequestStatus
	}
	if time.Now().After(req.ExpiresAt) {
		return nil, domain.ErrRequestExpired
	}

	// Generate temporary password
	tempPassword, err := generateTempPassword()
	if err != nil {
		return nil, fmt.Errorf("Execute: generate password: %w", err)
	}
	hashedPassword, err := pkgauth.HashPassword(tempPassword)
	if err != nil {
		return nil, fmt.Errorf("Execute: hash password: %w", err)
	}

	// Atomic transaction: create user + assign role + link employee + mark executed
	tx, err := uc.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("Execute: begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// 1. Create user
	var userID uuid.UUID
	if err := tx.QueryRow(ctx,
		`INSERT INTO users (email, hashed_password, full_name, branch_id, created_by)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		employeeEmail, hashedPassword, employeeFullName,
		req.RequestedBranchID, callerID,
	).Scan(&userID); err != nil {
		return nil, fmt.Errorf("Execute: insert user: %w", err)
	}

	// 2. Find role ID and assign
	var roleID uuid.UUID
	if err := tx.QueryRow(ctx,
		`SELECT id FROM roles WHERE code = $1`, req.RequestedRole,
	).Scan(&roleID); err != nil {
		return nil, fmt.Errorf("Execute: find role %q: %w", req.RequestedRole, err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, roleID,
	); err != nil {
		return nil, fmt.Errorf("Execute: assign role: %w", err)
	}

	// 3. Link employee.user_id
	if _, err := tx.Exec(ctx,
		`UPDATE employees SET user_id = $1, updated_at = now() WHERE id = $2`,
		userID, req.EmployeeID,
	); err != nil {
		return nil, fmt.Errorf("Execute: link employee: %w", err)
	}

	// 4. Mark request EXECUTED
	now := time.Now()
	var updated domain.ProvisioningRequest
	if err := tx.QueryRow(ctx,
		`UPDATE user_provisioning_requests
		 SET status = 'EXECUTED', executed_by = $2, executed_at = $3, updated_at = now()
		 WHERE id = $1
		 RETURNING `+provisioningColsInline,
		requestID, callerID, now,
	).Scan(
		&updated.ID, &updated.EmployeeID, &updated.RequestedBy, &updated.RequestedRole, &updated.RequestedBranchID,
		&updated.Status, &updated.ApprovalLevel,
		&updated.BranchApproverID, &updated.BranchApprovedAt, &updated.BranchRejectionReason,
		&updated.HRApproverID, &updated.HRApprovedAt, &updated.HRRejectionReason,
		&updated.ExecutedBy, &updated.ExecutedAt,
		&updated.IsEmergency, &updated.EmergencyReason, &updated.Notes,
		&updated.ExpiresAt, &updated.CreatedAt, &updated.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("Execute: mark executed: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("Execute: commit: %w", err)
	}

	uc.auditLog.Log(ctx, audit.Entry{ //nolint:errcheck
		UserID:     &callerID,
		Module:     "hrm",
		Resource:   "provisioning_request",
		ResourceID: &requestID,
		Action:     "PROVISIONING_EXECUTED",
		NewValue:   map[string]any{"user_id": userID, "role": req.RequestedRole, "employee_id": req.EmployeeID},
		IPAddress:  ip,
	})

	return &ExecuteProvisioningResponse{
		Request:      *toProvisioningResponse(&updated),
		UserID:       userID,
		TempPassword: tempPassword,
	}, nil
}

func (uc *ProvisioningUseCase) CancelRequest(
	ctx context.Context,
	requestID, callerID uuid.UUID,
	callerRoles []string,
	ip string,
) (*ProvisioningRequestResponse, error) {
	req, err := uc.repo.FindByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req.Status == "EXECUTED" {
		return nil, domain.ErrRequestAlreadyExecuted
	}

	// Only requester or SUPER_ADMIN may cancel
	isSA := hasRole(callerRoles, "SUPER_ADMIN")
	if req.RequestedBy != callerID && !isSA {
		return nil, domain.ErrInsufficientPermission
	}

	r, err := uc.repo.Cancel(ctx, requestID, callerID)
	if err != nil {
		return nil, err
	}

	uc.auditLog.Log(ctx, audit.Entry{ //nolint:errcheck
		UserID:     &callerID,
		Module:     "hrm",
		Resource:   "provisioning_request",
		ResourceID: &requestID,
		Action:     "PROVISIONING_CANCELLED",
		NewValue:   r,
		IPAddress:  ip,
	})
	return toProvisioningResponse(r), nil
}

// ─── Offboarding methods ──────────────────────────────────────────────────────

func (uc *ProvisioningUseCase) CreateOffboarding(
	ctx context.Context,
	req CreateOffboardingRequest,
	callerID uuid.UUID,
	ip string,
) (*OffboardingResponse, error) {
	var targetDate *time.Time
	if req.TargetDate != nil {
		t, err := time.Parse("2006-01-02", *req.TargetDate)
		if err != nil {
			return nil, domain.ErrValidation
		}
		targetDate = &t
	}

	c, err := uc.offRepo.Create(ctx, domain.CreateOffboardingParams{
		EmployeeID:    req.EmployeeID,
		ChecklistType: req.ChecklistType,
		InitiatedBy:   callerID,
		TargetDate:    targetDate,
		Notes:         req.Notes,
	})
	if err != nil {
		return nil, fmt.Errorf("CreateOffboarding: %w", err)
	}

	uc.auditLog.Log(ctx, audit.Entry{ //nolint:errcheck
		UserID:     &callerID,
		Module:     "hrm",
		Resource:   "offboarding_checklist",
		ResourceID: &c.ID,
		Action:     "OFFBOARDING_INITIATED",
		NewValue:   c,
		IPAddress:  ip,
	})
	return toOffboardingResponse(c), nil
}

func (uc *ProvisioningUseCase) ListOffboarding(
	ctx context.Context,
	req ListOffboardingRequest,
) (pagination.OffsetResult[OffboardingResponse], error) {
	empty := pagination.OffsetResult[OffboardingResponse]{}

	f := domain.ListOffboardingFilter{
		Status: req.Status,
		Page:   req.Page,
		Size:   req.Size,
	}
	if req.EmployeeID != "" {
		id, err := uuid.Parse(req.EmployeeID)
		if err != nil {
			return empty, domain.ErrValidation
		}
		f.EmployeeID = &id
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Size < 1 || f.Size > 100 {
		f.Size = 20
	}

	items, total, err := uc.offRepo.List(ctx, f)
	if err != nil {
		return empty, fmt.Errorf("ListOffboarding: %w", err)
	}

	resp := make([]OffboardingResponse, len(items))
	for i, item := range items {
		resp[i] = *toOffboardingResponse(item)
	}
	return pagination.NewOffsetResult(resp, total, f.Page, f.Size), nil
}

func (uc *ProvisioningUseCase) GetOffboarding(
	ctx context.Context,
	id uuid.UUID,
) (*OffboardingResponse, error) {
	c, err := uc.offRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toOffboardingResponse(c), nil
}

func (uc *ProvisioningUseCase) UpdateOffboardingItem(
	ctx context.Context,
	checklistID uuid.UUID,
	itemKey string,
	req UpdateOffboardingItemRequest,
	callerID uuid.UUID,
	ip string,
) (*OffboardingResponse, error) {
	c, err := uc.offRepo.UpdateItem(ctx, domain.UpdateOffboardingItemParams{
		ChecklistID: checklistID,
		ItemKey:     itemKey,
		Completed:   req.Completed,
		CompletedBy: &callerID,
		Notes:       req.Notes,
	})
	if err != nil {
		return nil, err
	}

	uc.auditLog.Log(ctx, audit.Entry{ //nolint:errcheck
		UserID:     &callerID,
		Module:     "hrm",
		Resource:   "offboarding_checklist",
		ResourceID: &checklistID,
		Action:     "OFFBOARDING_ITEM_COMPLETED",
		NewValue:   map[string]any{"key": itemKey, "completed": req.Completed},
		IPAddress:  ip,
	})
	return toOffboardingResponse(c), nil
}

func (uc *ProvisioningUseCase) CompleteOffboarding(
	ctx context.Context,
	checklistID, callerID uuid.UUID,
	ip string,
) (*OffboardingResponse, error) {
	c, err := uc.offRepo.Complete(ctx, checklistID, callerID)
	if err != nil {
		return nil, err
	}

	uc.auditLog.Log(ctx, audit.Entry{ //nolint:errcheck
		UserID:     &callerID,
		Module:     "hrm",
		Resource:   "offboarding_checklist",
		ResourceID: &checklistID,
		Action:     "OFFBOARDING_COMPLETED",
		NewValue:   c,
		IPAddress:  ip,
	})
	return toOffboardingResponse(c), nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// provisioningColsInline is used inside Execute's inline tx.QueryRow (must match scanProvisioning order).
const provisioningColsInline = `
	id, employee_id, requested_by, requested_role, requested_branch_id,
	status, approval_level,
	branch_approver_id, branch_approved_at, branch_rejection_reason,
	hr_approver_id, hr_approved_at, hr_rejection_reason,
	executed_by, executed_at,
	is_emergency, emergency_reason, notes,
	expires_at, created_at, updated_at`

func toProvisioningResponse(r *domain.ProvisioningRequest) *ProvisioningRequestResponse {
	resp := &ProvisioningRequestResponse{
		ID:                    r.ID,
		EmployeeID:            r.EmployeeID,
		RequestedBy:           r.RequestedBy,
		RequestedRole:         r.RequestedRole,
		RequestedBranchID:     r.RequestedBranchID,
		Status:                r.Status,
		ApprovalLevel:         r.ApprovalLevel,
		BranchApproverID:      r.BranchApproverID,
		BranchRejectionReason: r.BranchRejectionReason,
		HRApproverID:          r.HRApproverID,
		HRRejectionReason:     r.HRRejectionReason,
		ExecutedBy:            r.ExecutedBy,
		IsEmergency:           r.IsEmergency,
		EmergencyReason:       r.EmergencyReason,
		Notes:                 r.Notes,
		ExpiresAt:             r.ExpiresAt.Format(time.RFC3339),
		CreatedAt:             r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:             r.UpdatedAt.Format(time.RFC3339),
	}
	if r.BranchApprovedAt != nil {
		s := r.BranchApprovedAt.Format(time.RFC3339)
		resp.BranchApprovedAt = &s
	}
	if r.HRApprovedAt != nil {
		s := r.HRApprovedAt.Format(time.RFC3339)
		resp.HRApprovedAt = &s
	}
	if r.ExecutedAt != nil {
		s := r.ExecutedAt.Format(time.RFC3339)
		resp.ExecutedAt = &s
	}
	return resp
}

func toOffboardingResponse(c *domain.OffboardingChecklist) *OffboardingResponse {
	resp := &OffboardingResponse{
		ID:            c.ID,
		EmployeeID:    c.EmployeeID,
		ChecklistType: c.ChecklistType,
		InitiatedBy:   c.InitiatedBy,
		Items:         json.RawMessage(c.Items),
		Status:        c.Status,
		Notes:         c.Notes,
		CreatedAt:     c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     c.UpdatedAt.Format(time.RFC3339),
	}
	if c.TargetDate != nil {
		s := c.TargetDate.Format("2006-01-02")
		resp.TargetDate = &s
	}
	if c.CompletedAt != nil {
		s := c.CompletedAt.Format(time.RFC3339)
		resp.CompletedAt = &s
	}
	return resp
}

func generateTempPassword() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b)[:16], nil
}

func hasRole(roles []string, code string) bool {
	for _, r := range roles {
		if r == code {
			return true
		}
	}
	return false
}
