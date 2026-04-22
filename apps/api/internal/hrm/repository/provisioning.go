package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/hrm/domain"
)

// ─── ProvisioningRepo ─────────────────────────────────────────────────────────

type ProvisioningRepo struct{ pool *pgxpool.Pool }

func NewProvisioningRepo(pool *pgxpool.Pool) *ProvisioningRepo {
	return &ProvisioningRepo{pool: pool}
}

const provisioningCols = `
	id, employee_id, requested_by, requested_role, requested_branch_id,
	status, approval_level,
	branch_approver_id, branch_approved_at, branch_rejection_reason,
	hr_approver_id, hr_approved_at, hr_rejection_reason,
	executed_by, executed_at,
	is_emergency, emergency_reason, notes,
	expires_at, created_at, updated_at`

func scanProvisioning(row scanner) (*domain.ProvisioningRequest, error) {
	var r domain.ProvisioningRequest
	err := row.Scan(
		&r.ID, &r.EmployeeID, &r.RequestedBy, &r.RequestedRole, &r.RequestedBranchID,
		&r.Status, &r.ApprovalLevel,
		&r.BranchApproverID, &r.BranchApprovedAt, &r.BranchRejectionReason,
		&r.HRApproverID, &r.HRApprovedAt, &r.HRRejectionReason,
		&r.ExecutedBy, &r.ExecutedAt,
		&r.IsEmergency, &r.EmergencyReason, &r.Notes,
		&r.ExpiresAt, &r.CreatedAt, &r.UpdatedAt,
	)
	return &r, err
}

func (repo *ProvisioningRepo) Create(ctx context.Context, p domain.CreateProvisioningParams) (*domain.ProvisioningRequest, error) {
	const q = `
		INSERT INTO user_provisioning_requests
			(employee_id, requested_by, requested_role, requested_branch_id, is_emergency, emergency_reason, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING ` + provisioningCols

	row := repo.pool.QueryRow(ctx, q,
		p.EmployeeID, p.RequestedBy, p.RequestedRole, p.RequestedBranchID,
		p.IsEmergency, p.EmergencyReason, p.Notes,
	)
	r, err := scanProvisioning(row)
	if err != nil {
		return nil, fmt.Errorf("ProvisioningRepo.Create: %w", err)
	}
	return r, nil
}

func (repo *ProvisioningRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.ProvisioningRequest, error) {
	q := `SELECT ` + provisioningCols + ` FROM user_provisioning_requests WHERE id = $1`
	row := repo.pool.QueryRow(ctx, q, id)
	r, err := scanProvisioning(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrRequestNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("ProvisioningRepo.FindByID: %w", err)
	}
	return r, nil
}

func (repo *ProvisioningRepo) List(ctx context.Context, f domain.ListProvisioningFilter) ([]*domain.ProvisioningRequest, int64, error) {
	where := []string{"1=1"}
	args := []any{}
	idx := 1

	if f.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", idx))
		args = append(args, f.Status)
		idx++
	}
	if f.EmployeeID != nil {
		where = append(where, fmt.Sprintf("employee_id = $%d", idx))
		args = append(args, *f.EmployeeID)
		idx++
	}

	cond := strings.Join(where, " AND ")

	var total int64
	if err := repo.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM user_provisioning_requests WHERE `+cond, args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("ProvisioningRepo.List count: %w", err)
	}

	if f.Page < 1 {
		f.Page = 1
	}
	if f.Size < 1 {
		f.Size = 20
	}
	offset := (f.Page - 1) * f.Size
	listArgs := append(args, f.Size, offset)

	q := `SELECT ` + provisioningCols + ` FROM user_provisioning_requests WHERE ` + cond +
		fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1)

	rows, err := repo.pool.Query(ctx, q, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("ProvisioningRepo.List query: %w", err)
	}
	defer rows.Close()

	var result []*domain.ProvisioningRequest
	for rows.Next() {
		r, err := scanProvisioning(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("ProvisioningRepo.List scan: %w", err)
		}
		result = append(result, r)
	}
	return result, total, rows.Err()
}

func (repo *ProvisioningRepo) HasPendingForEmployee(ctx context.Context, employeeID uuid.UUID) (bool, error) {
	var exists bool
	err := repo.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM user_provisioning_requests WHERE employee_id = $1 AND status = 'PENDING')`,
		employeeID,
	).Scan(&exists)
	return exists, err
}

func (repo *ProvisioningRepo) BranchApprove(ctx context.Context, p domain.BranchApproveParams) (*domain.ProvisioningRequest, error) {
	now := time.Now()
	q := `
		UPDATE user_provisioning_requests
		SET branch_approver_id = $2, branch_approved_at = $3, approval_level = 2, updated_at = now()
		WHERE id = $1 AND status = 'PENDING' AND branch_approved_at IS NULL
		RETURNING ` + provisioningCols

	row := repo.pool.QueryRow(ctx, q, p.RequestID, p.ApproverID, now)
	r, err := scanProvisioning(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrInvalidRequestStatus
	}
	if err != nil {
		return nil, fmt.Errorf("ProvisioningRepo.BranchApprove: %w", err)
	}
	return r, nil
}

func (repo *ProvisioningRepo) BranchReject(ctx context.Context, p domain.RejectParams) (*domain.ProvisioningRequest, error) {
	q := `
		UPDATE user_provisioning_requests
		SET status = 'REJECTED', branch_rejection_reason = $2, updated_at = now()
		WHERE id = $1 AND status = 'PENDING' AND branch_approved_at IS NULL
		RETURNING ` + provisioningCols

	row := repo.pool.QueryRow(ctx, q, p.RequestID, p.Reason)
	r, err := scanProvisioning(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrInvalidRequestStatus
	}
	if err != nil {
		return nil, fmt.Errorf("ProvisioningRepo.BranchReject: %w", err)
	}
	return r, nil
}

func (repo *ProvisioningRepo) HRApprove(ctx context.Context, p domain.HRApproveParams) (*domain.ProvisioningRequest, error) {
	now := time.Now()
	q := `
		UPDATE user_provisioning_requests
		SET hr_approver_id = $2, hr_approved_at = $3, status = 'APPROVED', updated_at = now()
		WHERE id = $1 AND status = 'PENDING'
		RETURNING ` + provisioningCols

	row := repo.pool.QueryRow(ctx, q, p.RequestID, p.ApproverID, now)
	r, err := scanProvisioning(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrInvalidRequestStatus
	}
	if err != nil {
		return nil, fmt.Errorf("ProvisioningRepo.HRApprove: %w", err)
	}
	return r, nil
}

func (repo *ProvisioningRepo) HRReject(ctx context.Context, p domain.RejectParams) (*domain.ProvisioningRequest, error) {
	q := `
		UPDATE user_provisioning_requests
		SET status = 'REJECTED', hr_rejection_reason = $2, updated_at = now()
		WHERE id = $1 AND status = 'PENDING'
		RETURNING ` + provisioningCols

	row := repo.pool.QueryRow(ctx, q, p.RequestID, p.Reason)
	r, err := scanProvisioning(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrInvalidRequestStatus
	}
	if err != nil {
		return nil, fmt.Errorf("ProvisioningRepo.HRReject: %w", err)
	}
	return r, nil
}

func (repo *ProvisioningRepo) Cancel(ctx context.Context, requestID, callerID uuid.UUID) (*domain.ProvisioningRequest, error) {
	q := `
		UPDATE user_provisioning_requests
		SET status = 'CANCELLED', updated_at = now()
		WHERE id = $1 AND status IN ('PENDING','APPROVED')
		RETURNING ` + provisioningCols

	row := repo.pool.QueryRow(ctx, q, requestID)
	r, err := scanProvisioning(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrInvalidRequestStatus
	}
	if err != nil {
		return nil, fmt.Errorf("ProvisioningRepo.Cancel: %w", err)
	}
	return r, nil
}

func (repo *ProvisioningRepo) MarkExecuted(ctx context.Context, requestID, executorID uuid.UUID) (*domain.ProvisioningRequest, error) {
	now := time.Now()
	q := `
		UPDATE user_provisioning_requests
		SET status = 'EXECUTED', executed_by = $2, executed_at = $3, updated_at = now()
		WHERE id = $1
		RETURNING ` + provisioningCols

	row := repo.pool.QueryRow(ctx, q, requestID, executorID, now)
	r, err := scanProvisioning(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrRequestNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("ProvisioningRepo.MarkExecuted: %w", err)
	}
	return r, nil
}

// ─── OffboardingRepo ──────────────────────────────────────────────────────────

type OffboardingRepo struct{ pool *pgxpool.Pool }

func NewOffboardingRepo(pool *pgxpool.Pool) *OffboardingRepo {
	return &OffboardingRepo{pool: pool}
}

const offboardingCols = `
	id, employee_id, checklist_type, initiated_by, target_date,
	items, status, completed_at, notes, created_at, updated_at`

func scanOffboarding(row scanner) (*domain.OffboardingChecklist, error) {
	var c domain.OffboardingChecklist
	err := row.Scan(
		&c.ID, &c.EmployeeID, &c.ChecklistType, &c.InitiatedBy, &c.TargetDate,
		&c.Items, &c.Status, &c.CompletedAt, &c.Notes, &c.CreatedAt, &c.UpdatedAt,
	)
	return &c, err
}

func (repo *OffboardingRepo) Create(ctx context.Context, p domain.CreateOffboardingParams) (*domain.OffboardingChecklist, error) {
	const q = `
		INSERT INTO offboarding_checklists
			(employee_id, checklist_type, initiated_by, target_date, notes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING ` + offboardingCols

	row := repo.pool.QueryRow(ctx, q,
		p.EmployeeID, p.ChecklistType, p.InitiatedBy, p.TargetDate, p.Notes,
	)
	c, err := scanOffboarding(row)
	if err != nil {
		return nil, fmt.Errorf("OffboardingRepo.Create: %w", err)
	}
	return c, nil
}

func (repo *OffboardingRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.OffboardingChecklist, error) {
	q := `SELECT ` + offboardingCols + ` FROM offboarding_checklists WHERE id = $1`
	row := repo.pool.QueryRow(ctx, q, id)
	c, err := scanOffboarding(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrOffboardingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("OffboardingRepo.FindByID: %w", err)
	}
	return c, nil
}

func (repo *OffboardingRepo) List(ctx context.Context, f domain.ListOffboardingFilter) ([]*domain.OffboardingChecklist, int64, error) {
	where := []string{"1=1"}
	args := []any{}
	idx := 1

	if f.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", idx))
		args = append(args, f.Status)
		idx++
	}
	if f.EmployeeID != nil {
		where = append(where, fmt.Sprintf("employee_id = $%d", idx))
		args = append(args, *f.EmployeeID)
		idx++
	}

	cond := strings.Join(where, " AND ")

	var total int64
	if err := repo.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM offboarding_checklists WHERE `+cond, args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("OffboardingRepo.List count: %w", err)
	}

	if f.Page < 1 {
		f.Page = 1
	}
	if f.Size < 1 {
		f.Size = 20
	}
	offset := (f.Page - 1) * f.Size
	listArgs := append(args, f.Size, offset)

	q := `SELECT ` + offboardingCols + ` FROM offboarding_checklists WHERE ` + cond +
		fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1)

	rows, err := repo.pool.Query(ctx, q, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("OffboardingRepo.List query: %w", err)
	}
	defer rows.Close()

	var result []*domain.OffboardingChecklist
	for rows.Next() {
		c, err := scanOffboarding(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("OffboardingRepo.List scan: %w", err)
		}
		result = append(result, c)
	}
	return result, total, rows.Err()
}

func (repo *OffboardingRepo) UpdateItem(ctx context.Context, p domain.UpdateOffboardingItemParams) (*domain.OffboardingChecklist, error) {
	// Update a single item inside the JSONB items array by key.
	// items is {"items": [{"key": "...", "completed": bool, ...}]}
	// We use jsonb_set to toggle the item matching the given key.
	const q = `
		UPDATE offboarding_checklists
		SET items = jsonb_set(
			items,
			ARRAY['items'],
			(
				SELECT jsonb_agg(
					CASE WHEN elem->>'key' = $2
					THEN elem || jsonb_build_object('completed', $3::boolean, 'notes', $4::text)
					ELSE elem END
				)
				FROM jsonb_array_elements(items->'items') AS elem
			),
			true
		),
		updated_at = now()
		WHERE id = $1 AND status = 'IN_PROGRESS'
		RETURNING ` + offboardingCols

	row := repo.pool.QueryRow(ctx, q, p.ChecklistID, p.ItemKey, p.Completed, p.Notes)
	c, err := scanOffboarding(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrOffboardingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("OffboardingRepo.UpdateItem: %w", err)
	}
	return c, nil
}

func (repo *OffboardingRepo) Complete(ctx context.Context, checklistID, callerID uuid.UUID) (*domain.OffboardingChecklist, error) {
	now := time.Now()
	q := `
		UPDATE offboarding_checklists
		SET status = 'COMPLETED', completed_at = $2, updated_at = now()
		WHERE id = $1 AND status = 'IN_PROGRESS'
		RETURNING ` + offboardingCols

	row := repo.pool.QueryRow(ctx, q, checklistID, now)
	c, err := scanOffboarding(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrOffboardingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("OffboardingRepo.Complete: %w", err)
	}
	return c, nil
}
