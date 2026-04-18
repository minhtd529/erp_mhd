package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/tax/domain"
)

type DeadlineRepo struct{ pool *pgxpool.Pool }

func NewDeadlineRepo(pool *pgxpool.Pool) *DeadlineRepo { return &DeadlineRepo{pool: pool} }

const deadlineCols = `id, client_id, deadline_type, deadline_name, due_date, status,
	expected_submission_date, actual_submission_date, submission_status,
	notes, created_by, updated_by, created_at, updated_at`

func scanDeadline(row pgx.Row) (*domain.TaxDeadline, error) {
	var d domain.TaxDeadline
	err := row.Scan(
		&d.ID, &d.ClientID, &d.DeadlineType, &d.DeadlineName, &d.DueDate, &d.Status,
		&d.ExpectedSubmissionDate, &d.ActualSubmissionDate, &d.SubmissionStatus,
		&d.Notes, &d.CreatedBy, &d.UpdatedBy, &d.CreatedAt, &d.UpdatedAt,
	)
	return &d, err
}

func (r *DeadlineRepo) Create(ctx context.Context, p domain.CreateDeadlineParams) (*domain.TaxDeadline, error) {
	const q = `INSERT INTO tax_deadlines
		(client_id, deadline_type, deadline_name, due_date, expected_submission_date, notes, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING ` + deadlineCols
	row := r.pool.QueryRow(ctx, q,
		p.ClientID, string(p.DeadlineType), p.DeadlineName, p.DueDate,
		p.ExpectedSubmissionDate, p.Notes, p.CreatedBy,
	)
	d, err := scanDeadline(row)
	if err != nil {
		return nil, fmt.Errorf("deadline.Create: %w", err)
	}
	return d, nil
}

func (r *DeadlineRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.TaxDeadline, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+deadlineCols+` FROM tax_deadlines WHERE id=$1`, id)
	d, err := scanDeadline(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTaxDeadlineNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("deadline.FindByID: %w", err)
	}
	return d, nil
}

func (r *DeadlineRepo) List(ctx context.Context, f domain.ListDeadlinesFilter, page, size int) ([]*domain.TaxDeadline, int64, error) {
	where := "WHERE 1=1"
	args := []any{}
	idx := 1

	if f.ClientID != nil {
		where += fmt.Sprintf(" AND client_id=$%d", idx)
		args = append(args, f.ClientID)
		idx++
	}
	if f.Status != "" {
		where += fmt.Sprintf(" AND status=$%d", idx)
		args = append(args, string(f.Status))
		idx++
	}

	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM tax_deadlines `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("deadline.List count: %w", err)
	}

	offset := (page - 1) * size
	args = append(args, size, offset)
	q := fmt.Sprintf(`SELECT `+deadlineCols+` FROM tax_deadlines `+where+` ORDER BY due_date ASC LIMIT $%d OFFSET $%d`, idx, idx+1)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("deadline.List: %w", err)
	}
	defer rows.Close()

	var list []*domain.TaxDeadline
	for rows.Next() {
		d, err := scanDeadline(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("deadline.List scan: %w", err)
		}
		list = append(list, d)
	}
	if list == nil {
		list = []*domain.TaxDeadline{}
	}
	return list, total, rows.Err()
}

func (r *DeadlineRepo) Update(ctx context.Context, p domain.UpdateDeadlineParams) (*domain.TaxDeadline, error) {
	const q = `UPDATE tax_deadlines
		SET deadline_name=$2, due_date=$3, expected_submission_date=$4, notes=$5, updated_by=$6, updated_at=NOW()
		WHERE id=$1
		RETURNING ` + deadlineCols
	row := r.pool.QueryRow(ctx, q, p.ID, p.DeadlineName, p.DueDate, p.ExpectedSubmissionDate, p.Notes, p.UpdatedBy)
	d, err := scanDeadline(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTaxDeadlineNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("deadline.Update: %w", err)
	}
	return d, nil
}

func (r *DeadlineRepo) MarkCompleted(ctx context.Context, id uuid.UUID, actualDate time.Time, updatedBy uuid.UUID) (*domain.TaxDeadline, error) {
	const q = `UPDATE tax_deadlines
		SET status='COMPLETED', actual_submission_date=$2, submission_status='SUBMITTED', updated_by=$3, updated_at=NOW()
		WHERE id=$1
		RETURNING ` + deadlineCols
	row := r.pool.QueryRow(ctx, q, id, actualDate, updatedBy)
	d, err := scanDeadline(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTaxDeadlineNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("deadline.MarkCompleted: %w", err)
	}
	return d, nil
}

func (r *DeadlineRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.DeadlineStatus) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE tax_deadlines SET status=$2, updated_at=NOW() WHERE id=$1`,
		id, string(status),
	)
	return err
}

func (r *DeadlineRepo) ListDueSoon(ctx context.Context, beforeDate time.Time) ([]*domain.TaxDeadline, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+deadlineCols+` FROM tax_deadlines WHERE status='NOT_DUE' AND due_date <= $1 ORDER BY due_date ASC`,
		beforeDate,
	)
	if err != nil {
		return nil, fmt.Errorf("deadline.ListDueSoon: %w", err)
	}
	defer rows.Close()
	return collectDeadlines(rows)
}

func (r *DeadlineRepo) ListOverdue(ctx context.Context) ([]*domain.TaxDeadline, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+deadlineCols+` FROM tax_deadlines WHERE status NOT IN ('COMPLETED','OVERDUE') AND due_date < NOW() ORDER BY due_date ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("deadline.ListOverdue: %w", err)
	}
	defer rows.Close()
	return collectDeadlines(rows)
}

func collectDeadlines(rows pgx.Rows) ([]*domain.TaxDeadline, error) {
	var list []*domain.TaxDeadline
	for rows.Next() {
		d, err := scanDeadline(rows)
		if err != nil {
			return nil, fmt.Errorf("deadline scan: %w", err)
		}
		list = append(list, d)
	}
	if list == nil {
		list = []*domain.TaxDeadline{}
	}
	return list, rows.Err()
}
