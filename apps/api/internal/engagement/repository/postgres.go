// Package repository provides the PostgreSQL implementation of the Engagement
// domain repository interfaces using raw SQL (CQRS: same connection pool for reads and writes).
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/engagement/domain"
)

// ─── EngagementRepo ──────────────────────────────────────────────────────────

// EngagementRepo implements domain.EngagementRepository.
type EngagementRepo struct{ pool *pgxpool.Pool }

// NewEngagementRepo creates an EngagementRepo.
func NewEngagementRepo(pool *pgxpool.Pool) *EngagementRepo { return &EngagementRepo{pool: pool} }

func (r *EngagementRepo) Create(ctx context.Context, p domain.CreateEngagementParams) (*domain.Engagement, error) {
	const q = `
		INSERT INTO engagements
		    (client_id, service_type, fee_type, fee_amount, status,
		     partner_id, primary_salesperson_id, start_date, end_date, description,
		     created_by, updated_by)
		VALUES ($1, $2, $3, $4, 'DRAFT', $5, $6, $7, $8, $9, $10, $10)
		RETURNING ` + engagementCols

	e, err := scanEngagement(r.pool.QueryRow(ctx, q,
		p.ClientID, string(p.ServiceType), string(p.FeeType), p.FeeAmount,
		p.PartnerID, p.PrimarySalespersonID, p.StartDate, p.EndDate, p.Description,
		p.CreatedBy,
	))
	if err != nil {
		return nil, fmt.Errorf("engagement.Create: %w", err)
	}
	return e, nil
}

func (r *EngagementRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Engagement, error) {
	q := "SELECT " + engagementCols + " FROM engagements WHERE id = $1 AND is_deleted = false"
	e, err := scanEngagement(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrEngagementNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("engagement.FindByID: %w", err)
	}
	return e, nil
}

func (r *EngagementRepo) Update(ctx context.Context, p domain.UpdateEngagementParams) (*domain.Engagement, error) {
	const q = `
		UPDATE engagements
		SET service_type            = $2,
		    fee_type                = $3,
		    fee_amount              = $4,
		    partner_id              = $5,
		    primary_salesperson_id  = $6,
		    start_date              = $7,
		    end_date                = $8,
		    description             = $9,
		    updated_by              = $10,
		    updated_at              = NOW()
		WHERE id = $1 AND is_deleted = false
		RETURNING ` + engagementCols

	e, err := scanEngagement(r.pool.QueryRow(ctx, q,
		p.ID, string(p.ServiceType), string(p.FeeType), p.FeeAmount,
		p.PartnerID, p.PrimarySalespersonID, p.StartDate, p.EndDate, p.Description,
		p.UpdatedBy,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrEngagementNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("engagement.Update: %w", err)
	}
	return e, nil
}

func (r *EngagementRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.EngagementStatus, updatedBy uuid.UUID) (*domain.Engagement, error) {
	const q = `
		UPDATE engagements SET status = $2, updated_by = $3, updated_at = NOW()
		WHERE id = $1 AND is_deleted = false
		RETURNING ` + engagementCols

	e, err := scanEngagement(r.pool.QueryRow(ctx, q, id, string(status), updatedBy))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrEngagementNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("engagement.UpdateStatus: %w", err)
	}
	return e, nil
}

func (r *EngagementRepo) SoftDelete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE engagements SET is_deleted = true, updated_by = $2, updated_at = NOW() WHERE id = $1 AND is_deleted = false`,
		id, deletedBy)
	if err != nil {
		return fmt.Errorf("engagement.SoftDelete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrEngagementNotFound
	}
	return nil
}

func (r *EngagementRepo) List(ctx context.Context, f domain.ListEngagementsFilter) ([]*domain.Engagement, int64, error) {
	offset := (f.Page - 1) * f.Size
	args := []any{}
	where := "WHERE is_deleted = false"
	idx := 1

	if f.ClientID != nil {
		where += fmt.Sprintf(" AND client_id = $%d", idx)
		args = append(args, *f.ClientID)
		idx++
	}
	if f.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, string(f.Status))
		idx++
	}
	if f.ServiceType != "" {
		where += fmt.Sprintf(" AND service_type = $%d", idx)
		args = append(args, string(f.ServiceType))
		idx++
	}
	if f.FeeType != "" {
		where += fmt.Sprintf(" AND fee_type = $%d", idx)
		args = append(args, string(f.FeeType))
		idx++
	}
	if f.PartnerID != nil {
		where += fmt.Sprintf(" AND partner_id = $%d", idx)
		args = append(args, *f.PartnerID)
		idx++
	}
	if f.DateFrom != nil {
		where += fmt.Sprintf(" AND start_date >= $%d", idx)
		args = append(args, *f.DateFrom)
		idx++
	}
	if f.DateTo != nil {
		where += fmt.Sprintf(" AND start_date <= $%d", idx)
		args = append(args, *f.DateTo)
		idx++
	}
	if f.Q != "" {
		where += fmt.Sprintf(` AND (search_vector @@ plainto_tsquery('simple', $%d) OR description ILIKE $%d)`, idx, idx+1)
		args = append(args, f.Q, "%"+f.Q+"%")
		idx += 2
	}

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM engagements "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("engagement.List count: %w", err)
	}

	args = append(args, f.Size, offset)
	dataQ := fmt.Sprintf("SELECT %s FROM engagements %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		engagementCols, where, idx, idx+1)

	rows, err := r.pool.Query(ctx, dataQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("engagement.List query: %w", err)
	}
	defer rows.Close()

	var list []*domain.Engagement
	for rows.Next() {
		e, err := scanEngagement(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("engagement.List scan: %w", err)
		}
		list = append(list, e)
	}
	if list == nil {
		list = []*domain.Engagement{}
	}
	return list, total, rows.Err()
}

// ListCursor returns up to f.Size+1 engagements ordered by (created_at DESC, id DESC)
// starting after the cursor position (exclusive). Callers use the extra row to detect
// whether a next page exists.
func (r *EngagementRepo) ListCursor(ctx context.Context, f domain.CursorFilter) ([]*domain.Engagement, error) {
	args := []any{}
	where := "WHERE is_deleted = false"
	idx := 1

	if f.AfterCreatedAt != nil && f.AfterID != nil {
		where += fmt.Sprintf(" AND (created_at, id) < ($%d, $%d)", idx, idx+1)
		args = append(args, *f.AfterCreatedAt, *f.AfterID)
		idx += 2
	}
	if f.ClientID != nil {
		where += fmt.Sprintf(" AND client_id = $%d", idx)
		args = append(args, *f.ClientID)
		idx++
	}
	if f.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, string(f.Status))
		idx++
	}
	if f.ServiceType != "" {
		where += fmt.Sprintf(" AND service_type = $%d", idx)
		args = append(args, string(f.ServiceType))
		idx++
	}
	if f.FeeType != "" {
		where += fmt.Sprintf(" AND fee_type = $%d", idx)
		args = append(args, string(f.FeeType))
		idx++
	}
	if f.PartnerID != nil {
		where += fmt.Sprintf(" AND partner_id = $%d", idx)
		args = append(args, *f.PartnerID)
		idx++
	}
	if f.DateFrom != nil {
		where += fmt.Sprintf(" AND start_date >= $%d", idx)
		args = append(args, *f.DateFrom)
		idx++
	}
	if f.DateTo != nil {
		where += fmt.Sprintf(" AND start_date <= $%d", idx)
		args = append(args, *f.DateTo)
		idx++
	}
	if f.Q != "" {
		where += fmt.Sprintf(` AND (search_vector @@ plainto_tsquery('simple', $%d) OR description ILIKE $%d)`, idx, idx+1)
		args = append(args, f.Q, "%"+f.Q+"%")
		idx += 2
	}

	limit := f.Size + 1 // fetch one extra to detect hasMore
	args = append(args, limit)
	q := fmt.Sprintf("SELECT %s FROM engagements %s ORDER BY created_at DESC, id DESC LIMIT $%d",
		engagementCols, where, idx)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("engagement.ListCursor query: %w", err)
	}
	defer rows.Close()

	var list []*domain.Engagement
	for rows.Next() {
		e, err := scanEngagement(rows)
		if err != nil {
			return nil, fmt.Errorf("engagement.ListCursor scan: %w", err)
		}
		list = append(list, e)
	}
	if list == nil {
		list = []*domain.Engagement{}
	}
	return list, rows.Err()
}

// ─── MemberRepo ──────────────────────────────────────────────────────────────

// MemberRepo implements domain.MemberRepository.
type MemberRepo struct{ pool *pgxpool.Pool }

// NewMemberRepo creates a MemberRepo.
func NewMemberRepo(pool *pgxpool.Pool) *MemberRepo { return &MemberRepo{pool: pool} }

func (r *MemberRepo) Assign(ctx context.Context, p domain.AssignMemberParams) (*domain.EngagementMember, error) {
	const q = `
		INSERT INTO engagement_members
		    (engagement_id, staff_id, role, hourly_rate, allocation_percent, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		RETURNING ` + memberCols

	m, err := scanMember(r.pool.QueryRow(ctx, q,
		p.EngagementID, p.StaffID, string(p.Role), p.HourlyRate, p.AllocationPercent, p.CreatedBy,
	))
	if err != nil {
		return nil, fmt.Errorf("engagement.Assign: %w", err)
	}
	return m, nil
}

func (r *MemberRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.EngagementMember, error) {
	q := "SELECT " + memberCols + " FROM engagement_members WHERE id = $1 AND is_deleted = false"
	m, err := scanMember(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrMemberNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("engagement.FindMemberByID: %w", err)
	}
	return m, nil
}

func (r *MemberRepo) Update(ctx context.Context, p domain.UpdateMemberParams) (*domain.EngagementMember, error) {
	const q = `
		UPDATE engagement_members
		SET role = $3, hourly_rate = $4, allocation_percent = $5, updated_by = $6, updated_at = NOW()
		WHERE id = $1 AND engagement_id = $2 AND is_deleted = false
		RETURNING ` + memberCols

	m, err := scanMember(r.pool.QueryRow(ctx, q,
		p.ID, p.EngagementID, string(p.Role), p.HourlyRate, p.AllocationPercent, p.UpdatedBy,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrMemberNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("engagement.UpdateMember: %w", err)
	}
	return m, nil
}

func (r *MemberRepo) SoftDelete(ctx context.Context, id uuid.UUID, engagementID uuid.UUID, deletedBy uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE engagement_members SET is_deleted = true, updated_by = $3, updated_at = NOW() WHERE id = $1 AND engagement_id = $2 AND is_deleted = false`,
		id, engagementID, deletedBy)
	if err != nil {
		return fmt.Errorf("engagement.SoftDeleteMember: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrMemberNotFound
	}
	return nil
}

func (r *MemberRepo) ListByEngagement(ctx context.Context, engagementID uuid.UUID) ([]*domain.EngagementMember, error) {
	q := "SELECT " + memberCols + " FROM engagement_members WHERE engagement_id = $1 AND is_deleted = false ORDER BY created_at"
	rows, err := r.pool.Query(ctx, q, engagementID)
	if err != nil {
		return nil, fmt.Errorf("engagement.ListMembers: %w", err)
	}
	defer rows.Close()

	var list []*domain.EngagementMember
	for rows.Next() {
		m, err := scanMember(rows)
		if err != nil {
			return nil, fmt.Errorf("engagement.ListMembers scan: %w", err)
		}
		list = append(list, m)
	}
	if list == nil {
		list = []*domain.EngagementMember{}
	}
	return list, rows.Err()
}

func (r *MemberRepo) SumAllocation(ctx context.Context, engagementID uuid.UUID, excludeID *uuid.UUID) (int, error) {
	var sum int
	if excludeID != nil {
		err := r.pool.QueryRow(ctx,
			`SELECT COALESCE(SUM(allocation_percent),0) FROM engagement_members WHERE engagement_id = $1 AND id != $2 AND is_deleted = false`,
			engagementID, *excludeID).Scan(&sum)
		return sum, err
	}
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(allocation_percent),0) FROM engagement_members WHERE engagement_id = $1 AND is_deleted = false`,
		engagementID).Scan(&sum)
	return sum, err
}

// ─── TaskRepo ────────────────────────────────────────────────────────────────

// TaskRepo implements domain.TaskRepository.
type TaskRepo struct{ pool *pgxpool.Pool }

// NewTaskRepo creates a TaskRepo.
func NewTaskRepo(pool *pgxpool.Pool) *TaskRepo { return &TaskRepo{pool: pool} }

func (r *TaskRepo) Create(ctx context.Context, p domain.CreateTaskParams) (*domain.EngagementTask, error) {
	const q = `
		INSERT INTO engagement_tasks (engagement_id, phase, title, assigned_to, due_date, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		RETURNING ` + taskCols

	t, err := scanTask(r.pool.QueryRow(ctx, q,
		p.EngagementID, string(p.Phase), p.Title, p.AssignedTo, p.DueDate, p.CreatedBy,
	))
	if err != nil {
		return nil, fmt.Errorf("engagement.CreateTask: %w", err)
	}
	return t, nil
}

func (r *TaskRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.EngagementTask, error) {
	q := "SELECT " + taskCols + " FROM engagement_tasks WHERE id = $1 AND is_deleted = false"
	t, err := scanTask(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTaskNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("engagement.FindTaskByID: %w", err)
	}
	return t, nil
}

func (r *TaskRepo) Update(ctx context.Context, p domain.UpdateTaskParams) (*domain.EngagementTask, error) {
	const q = `
		UPDATE engagement_tasks
		SET title = $3, assigned_to = $4, status = $5, due_date = $6, updated_by = $7, updated_at = NOW()
		WHERE id = $1 AND engagement_id = $2 AND is_deleted = false
		RETURNING ` + taskCols

	t, err := scanTask(r.pool.QueryRow(ctx, q,
		p.ID, p.EngagementID, p.Title, p.AssignedTo, string(p.Status), p.DueDate, p.UpdatedBy,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTaskNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("engagement.UpdateTask: %w", err)
	}
	return t, nil
}

func (r *TaskRepo) ListByEngagement(ctx context.Context, engagementID uuid.UUID, phase domain.TaskPhase) ([]*domain.EngagementTask, error) {
	args := []any{engagementID}
	q := "SELECT " + taskCols + " FROM engagement_tasks WHERE engagement_id = $1 AND is_deleted = false"
	if phase != "" {
		q += " AND phase = $2"
		args = append(args, string(phase))
	}
	q += " ORDER BY created_at"

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("engagement.ListTasks: %w", err)
	}
	defer rows.Close()

	var list []*domain.EngagementTask
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("engagement.ListTasks scan: %w", err)
		}
		list = append(list, t)
	}
	if list == nil {
		list = []*domain.EngagementTask{}
	}
	return list, rows.Err()
}

// ─── CostRepo ────────────────────────────────────────────────────────────────

// CostRepo implements domain.CostRepository.
type CostRepo struct{ pool *pgxpool.Pool }

// NewCostRepo creates a CostRepo.
func NewCostRepo(pool *pgxpool.Pool) *CostRepo { return &CostRepo{pool: pool} }

func (r *CostRepo) Create(ctx context.Context, p domain.CreateCostParams) (*domain.DirectCost, error) {
	const q = `
		INSERT INTO direct_costs (engagement_id, cost_type, description, amount, receipt_url, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		RETURNING ` + costCols

	c, err := scanCost(r.pool.QueryRow(ctx, q,
		p.EngagementID, string(p.CostType), p.Description, p.Amount, p.ReceiptURL, p.CreatedBy,
	))
	if err != nil {
		return nil, fmt.Errorf("engagement.CreateCost: %w", err)
	}
	return c, nil
}

func (r *CostRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.DirectCost, error) {
	q := "SELECT " + costCols + " FROM direct_costs WHERE id = $1 AND is_deleted = false"
	c, err := scanCost(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrCostNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("engagement.FindCostByID: %w", err)
	}
	return c, nil
}

func (r *CostRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.CostStatus, actorID uuid.UUID, rejectReason *string) (*domain.DirectCost, error) {
	now := time.Now()
	var q string
	var args []any

	switch status {
	case domain.CostSubmitted:
		q = `UPDATE direct_costs SET status = $2, submitted_at = $3, submitted_by = $4, updated_by = $4, updated_at = NOW()
		     WHERE id = $1 AND is_deleted = false RETURNING ` + costCols
		args = []any{id, string(status), now, actorID}
	case domain.CostApproved:
		q = `UPDATE direct_costs SET status = $2, approved_at = $3, approved_by = $4, updated_by = $4, updated_at = NOW()
		     WHERE id = $1 AND is_deleted = false RETURNING ` + costCols
		args = []any{id, string(status), now, actorID}
	case domain.CostRejected:
		q = `UPDATE direct_costs SET status = $2, reject_reason = $3, updated_by = $4, updated_at = NOW()
		     WHERE id = $1 AND is_deleted = false RETURNING ` + costCols
		args = []any{id, string(status), rejectReason, actorID}
	default:
		return nil, fmt.Errorf("engagement.UpdateCostStatus: unsupported status %s", status)
	}

	c, err := scanCost(r.pool.QueryRow(ctx, q, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrCostNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("engagement.UpdateCostStatus: %w", err)
	}
	return c, nil
}

func (r *CostRepo) ListByEngagement(ctx context.Context, engagementID uuid.UUID) ([]*domain.DirectCost, error) {
	q := "SELECT " + costCols + " FROM direct_costs WHERE engagement_id = $1 AND is_deleted = false ORDER BY created_at DESC"
	rows, err := r.pool.Query(ctx, q, engagementID)
	if err != nil {
		return nil, fmt.Errorf("engagement.ListCosts: %w", err)
	}
	defer rows.Close()

	var list []*domain.DirectCost
	for rows.Next() {
		c, err := scanCost(rows)
		if err != nil {
			return nil, fmt.Errorf("engagement.ListCosts scan: %w", err)
		}
		list = append(list, c)
	}
	if list == nil {
		list = []*domain.DirectCost{}
	}
	return list, rows.Err()
}

// ─── Column lists & scanners ─────────────────────────────────────────────────

const engagementCols = `id, client_id, service_type, fee_type, fee_amount, status,
    partner_id, primary_salesperson_id, start_date, end_date, description,
    is_deleted, created_at, updated_at, created_by, updated_by`

const memberCols = `id, engagement_id, staff_id, role, hourly_rate, allocation_percent,
    status, is_deleted, created_at, updated_at, created_by, updated_by`

const taskCols = `id, engagement_id, phase, title, assigned_to, status,
    due_date, is_deleted, created_at, updated_at, created_by, updated_by`

const costCols = `id, engagement_id, cost_type, description, amount, status,
    receipt_url, submitted_at, submitted_by, approved_at, approved_by, reject_reason,
    is_deleted, created_at, updated_at, created_by, updated_by`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanEngagement(row rowScanner) (*domain.Engagement, error) {
	var e domain.Engagement
	var svcType, feeType, status string
	err := row.Scan(
		&e.ID, &e.ClientID, &svcType, &feeType, &e.FeeAmount, &status,
		&e.PartnerID, &e.PrimarySalespersonID, &e.StartDate, &e.EndDate, &e.Description,
		&e.IsDeleted, &e.CreatedAt, &e.UpdatedAt, &e.CreatedBy, &e.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	e.ServiceType = domain.ServiceType(svcType)
	e.FeeType = domain.FeeType(feeType)
	e.Status = domain.EngagementStatus(status)
	return &e, nil
}

func scanMember(row rowScanner) (*domain.EngagementMember, error) {
	var m domain.EngagementMember
	var role, status string
	err := row.Scan(
		&m.ID, &m.EngagementID, &m.StaffID, &role, &m.HourlyRate, &m.AllocationPercent,
		&status, &m.IsDeleted, &m.CreatedAt, &m.UpdatedAt, &m.CreatedBy, &m.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	m.Role = domain.MemberRole(role)
	m.Status = domain.MemberStatus(status)
	return &m, nil
}

func scanTask(row rowScanner) (*domain.EngagementTask, error) {
	var t domain.EngagementTask
	var phase, status string
	err := row.Scan(
		&t.ID, &t.EngagementID, &phase, &t.Title, &t.AssignedTo, &status,
		&t.DueDate, &t.IsDeleted, &t.CreatedAt, &t.UpdatedAt, &t.CreatedBy, &t.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	t.Phase = domain.TaskPhase(phase)
	t.Status = domain.TaskStatus(status)
	return &t, nil
}

func scanCost(row rowScanner) (*domain.DirectCost, error) {
	var c domain.DirectCost
	var costType, status string
	err := row.Scan(
		&c.ID, &c.EngagementID, &costType, &c.Description, &c.Amount, &status,
		&c.ReceiptURL, &c.SubmittedAt, &c.SubmittedBy, &c.ApprovedAt, &c.ApprovedBy, &c.RejectReason,
		&c.IsDeleted, &c.CreatedAt, &c.UpdatedAt, &c.CreatedBy, &c.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	c.CostType = domain.CostType(costType)
	c.Status = domain.CostStatus(status)
	return &c, nil
}
