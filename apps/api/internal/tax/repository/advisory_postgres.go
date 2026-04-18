package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/tax/domain"
)

type AdvisoryRepo struct{ pool *pgxpool.Pool }

func NewAdvisoryRepo(pool *pgxpool.Pool) *AdvisoryRepo { return &AdvisoryRepo{pool: pool} }

const advisoryCols = `id, client_id, engagement_id, advisory_type, recommendation, findings,
	status, delivered_date, client_feedback, created_by, updated_by, created_at, updated_at`

func scanAdvisory(row pgx.Row) (*domain.AdvisoryRecord, error) {
	var a domain.AdvisoryRecord
	err := row.Scan(
		&a.ID, &a.ClientID, &a.EngagementID, &a.AdvisoryType, &a.Recommendation, &a.Findings,
		&a.Status, &a.DeliveredDate, &a.ClientFeedback, &a.CreatedBy, &a.UpdatedBy, &a.CreatedAt, &a.UpdatedAt,
	)
	return &a, err
}

func (r *AdvisoryRepo) Create(ctx context.Context, p domain.CreateAdvisoryParams) (*domain.AdvisoryRecord, error) {
	const q = `INSERT INTO advisory_records
		(client_id, engagement_id, advisory_type, recommendation, findings, created_by)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING ` + advisoryCols
	row := r.pool.QueryRow(ctx, q,
		p.ClientID, p.EngagementID, string(p.AdvisoryType),
		p.Recommendation, p.Findings, p.CreatedBy,
	)
	a, err := scanAdvisory(row)
	if err != nil {
		return nil, fmt.Errorf("advisory.Create: %w", err)
	}
	return a, nil
}

func (r *AdvisoryRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.AdvisoryRecord, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+advisoryCols+` FROM advisory_records WHERE id=$1`, id)
	a, err := scanAdvisory(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrAdvisoryRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("advisory.FindByID: %w", err)
	}
	return a, nil
}

func (r *AdvisoryRepo) List(ctx context.Context, f domain.ListAdvisoryFilter, page, size int) ([]*domain.AdvisoryRecord, int64, error) {
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
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM advisory_records `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("advisory.List count: %w", err)
	}

	offset := (page - 1) * size
	args = append(args, size, offset)
	q := fmt.Sprintf(`SELECT `+advisoryCols+` FROM advisory_records `+where+` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("advisory.List: %w", err)
	}
	defer rows.Close()

	var list []*domain.AdvisoryRecord
	for rows.Next() {
		a, err := scanAdvisory(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("advisory.List scan: %w", err)
		}
		list = append(list, a)
	}
	if list == nil {
		list = []*domain.AdvisoryRecord{}
	}
	return list, total, rows.Err()
}

func (r *AdvisoryRepo) Update(ctx context.Context, p domain.UpdateAdvisoryParams) (*domain.AdvisoryRecord, error) {
	const q = `UPDATE advisory_records
		SET recommendation=$2, findings=$3, updated_by=$4, updated_at=NOW()
		WHERE id=$1 AND status='DRAFTED'
		RETURNING ` + advisoryCols
	row := r.pool.QueryRow(ctx, q, p.ID, p.Recommendation, p.Findings, p.UpdatedBy)
	a, err := scanAdvisory(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrAdvisoryRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("advisory.Update: %w", err)
	}
	return a, nil
}

func (r *AdvisoryRepo) Deliver(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) (*domain.AdvisoryRecord, error) {
	const q = `UPDATE advisory_records
		SET status='DELIVERED', delivered_date=NOW(), updated_by=$2, updated_at=NOW()
		WHERE id=$1 AND status='DRAFTED'
		RETURNING ` + advisoryCols
	row := r.pool.QueryRow(ctx, q, id, updatedBy)
	a, err := scanAdvisory(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrAdvisoryNotDeliverable
	}
	if err != nil {
		return nil, fmt.Errorf("advisory.Deliver: %w", err)
	}
	return a, nil
}

func (r *AdvisoryRepo) AttachFile(ctx context.Context, p domain.AttachFileParams) (*domain.AdvisoryFile, error) {
	const q = `INSERT INTO advisory_files (advisory_id, file_name, file_path, created_by)
		VALUES ($1,$2,$3,$4)
		RETURNING id, advisory_id, file_name, file_path, created_by, created_at`
	var f domain.AdvisoryFile
	err := r.pool.QueryRow(ctx, q, p.AdvisoryID, p.FileName, p.FilePath, p.CreatedBy).
		Scan(&f.ID, &f.AdvisoryID, &f.FileName, &f.FilePath, &f.CreatedBy, &f.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("advisory.AttachFile: %w", err)
	}
	return &f, nil
}

func (r *AdvisoryRepo) ListFiles(ctx context.Context, advisoryID uuid.UUID) ([]*domain.AdvisoryFile, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, advisory_id, file_name, file_path, created_by, created_at FROM advisory_files WHERE advisory_id=$1 ORDER BY created_at ASC`,
		advisoryID,
	)
	if err != nil {
		return nil, fmt.Errorf("advisory.ListFiles: %w", err)
	}
	defer rows.Close()

	var list []*domain.AdvisoryFile
	for rows.Next() {
		var f domain.AdvisoryFile
		if err := rows.Scan(&f.ID, &f.AdvisoryID, &f.FileName, &f.FilePath, &f.CreatedBy, &f.CreatedAt); err != nil {
			return nil, fmt.Errorf("advisory.ListFiles scan: %w", err)
		}
		list = append(list, &f)
	}
	if list == nil {
		list = []*domain.AdvisoryFile{}
	}
	return list, rows.Err()
}
