// Package repository provides the PostgreSQL implementation of the org domain
// repository interfaces (BranchRepository and DepartmentRepository).
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/org/domain"
)

// ─── BranchRepo ──────────────────────────────────────────────────────────────

type BranchRepo struct {
	pool *pgxpool.Pool
}

func NewBranchRepo(pool *pgxpool.Pool) *BranchRepo {
	return &BranchRepo{pool: pool}
}

const branchCols = `id, code, name, address, phone, is_active, created_at, updated_at, created_by, updated_by`

func (r *BranchRepo) Create(ctx context.Context, p domain.CreateBranchParams) (*domain.Branch, error) {
	q := `INSERT INTO branches (code, name, address, phone, created_by, updated_by)
	      VALUES ($1, $2, $3, $4, $5, $5)
	      RETURNING ` + branchCols
	b, err := scanBranch(r.pool.QueryRow(ctx, q, p.Code, p.Name, p.Address, p.Phone, p.CreatedBy))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrDuplicateCode
		}
		return nil, fmt.Errorf("org.CreateBranch: %w", err)
	}
	return b, nil
}

func (r *BranchRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Branch, error) {
	q := `SELECT ` + branchCols + ` FROM branches WHERE id = $1`
	b, err := scanBranch(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrBranchNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("org.FindBranchByID: %w", err)
	}
	return b, nil
}

func (r *BranchRepo) Update(ctx context.Context, p domain.UpdateBranchParams) (*domain.Branch, error) {
	q := `UPDATE branches
	      SET code=$2, name=$3, address=$4, phone=$5, is_active=$6, updated_by=$7, updated_at=NOW()
	      WHERE id=$1
	      RETURNING ` + branchCols
	b, err := scanBranch(r.pool.QueryRow(ctx, q,
		p.ID, p.Code, p.Name, p.Address, p.Phone, p.IsActive, p.UpdatedBy,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrBranchNotFound
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrDuplicateCode
		}
		return nil, fmt.Errorf("org.UpdateBranch: %w", err)
	}
	return b, nil
}

func (r *BranchRepo) List(ctx context.Context, f domain.ListBranchesFilter) ([]*domain.Branch, int64, error) {
	offset := (f.Page - 1) * f.Size
	args := []any{}
	where := "WHERE 1=1"
	idx := 1

	if f.IsActive != nil {
		where += fmt.Sprintf(" AND is_active = $%d", idx)
		args = append(args, *f.IsActive)
		idx++
	}
	if f.Q != "" {
		where += fmt.Sprintf(` AND (name ILIKE $%d OR code ILIKE $%d)`, idx, idx+1)
		args = append(args, "%"+f.Q+"%", "%"+f.Q+"%")
		idx += 2
	}

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM branches "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("org.ListBranches count: %w", err)
	}

	args = append(args, f.Size, offset)
	dataQ := fmt.Sprintf(`SELECT `+branchCols+` FROM branches %s ORDER BY name ASC LIMIT $%d OFFSET $%d`,
		where, idx, idx+1)
	rows, err := r.pool.Query(ctx, dataQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("org.ListBranches query: %w", err)
	}
	defer rows.Close()

	var branches []*domain.Branch
	for rows.Next() {
		b, err := scanBranch(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("org.ListBranches scan: %w", err)
		}
		branches = append(branches, b)
	}
	if branches == nil {
		branches = []*domain.Branch{}
	}
	return branches, total, rows.Err()
}

// ─── DeptRepo ────────────────────────────────────────────────────────────────

type DeptRepo struct {
	pool *pgxpool.Pool
}

func NewDeptRepo(pool *pgxpool.Pool) *DeptRepo {
	return &DeptRepo{pool: pool}
}

const deptCols = `id, branch_id, code, name, is_active, created_at, updated_at, created_by, updated_by`

func (r *DeptRepo) Create(ctx context.Context, p domain.CreateDepartmentParams) (*domain.Department, error) {
	q := `INSERT INTO departments (branch_id, code, name, created_by, updated_by)
	      VALUES ($1, $2, $3, $4, $4)
	      RETURNING ` + deptCols
	d, err := scanDept(r.pool.QueryRow(ctx, q, p.BranchID, p.Code, p.Name, p.CreatedBy))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrDuplicateCode
		}
		return nil, fmt.Errorf("org.CreateDepartment: %w", err)
	}
	return d, nil
}

func (r *DeptRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Department, error) {
	q := `SELECT ` + deptCols + ` FROM departments WHERE id = $1`
	d, err := scanDept(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrDepartmentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("org.FindDepartmentByID: %w", err)
	}
	return d, nil
}

func (r *DeptRepo) Update(ctx context.Context, p domain.UpdateDepartmentParams) (*domain.Department, error) {
	q := `UPDATE departments
	      SET code=$2, name=$3, branch_id=$4, is_active=$5, updated_by=$6, updated_at=NOW()
	      WHERE id=$1
	      RETURNING ` + deptCols
	d, err := scanDept(r.pool.QueryRow(ctx, q,
		p.ID, p.Code, p.Name, p.BranchID, p.IsActive, p.UpdatedBy,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrDepartmentNotFound
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrDuplicateCode
		}
		return nil, fmt.Errorf("org.UpdateDepartment: %w", err)
	}
	return d, nil
}

func (r *DeptRepo) List(ctx context.Context, f domain.ListDepartmentsFilter) ([]*domain.Department, int64, error) {
	offset := (f.Page - 1) * f.Size
	args := []any{}
	where := "WHERE 1=1"
	idx := 1

	if f.BranchID != nil {
		where += fmt.Sprintf(" AND branch_id = $%d", idx)
		args = append(args, f.BranchID)
		idx++
	}
	if f.IsActive != nil {
		where += fmt.Sprintf(" AND is_active = $%d", idx)
		args = append(args, *f.IsActive)
		idx++
	}
	if f.Q != "" {
		where += fmt.Sprintf(` AND (name ILIKE $%d OR code ILIKE $%d)`, idx, idx+1)
		args = append(args, "%"+f.Q+"%", "%"+f.Q+"%")
		idx += 2
	}

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM departments "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("org.ListDepartments count: %w", err)
	}

	args = append(args, f.Size, offset)
	dataQ := fmt.Sprintf(`SELECT `+deptCols+` FROM departments %s ORDER BY name ASC LIMIT $%d OFFSET $%d`,
		where, idx, idx+1)
	rows, err := r.pool.Query(ctx, dataQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("org.ListDepartments query: %w", err)
	}
	defer rows.Close()

	var depts []*domain.Department
	for rows.Next() {
		d, err := scanDept(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("org.ListDepartments scan: %w", err)
		}
		depts = append(depts, d)
	}
	if depts == nil {
		depts = []*domain.Department{}
	}
	return depts, total, rows.Err()
}

// ─── Scanners ─────────────────────────────────────────────────────────────────

type scanner interface {
	Scan(dest ...any) error
}

func scanBranch(row scanner) (*domain.Branch, error) {
	var b domain.Branch
	return &b, row.Scan(
		&b.ID, &b.Code, &b.Name, &b.Address, &b.Phone,
		&b.IsActive, &b.CreatedAt, &b.UpdatedAt, &b.CreatedBy, &b.UpdatedBy,
	)
}

func scanDept(row scanner) (*domain.Department, error) {
	var d domain.Department
	return &d, row.Scan(
		&d.ID, &d.BranchID, &d.Code, &d.Name,
		&d.IsActive, &d.CreatedAt, &d.UpdatedAt, &d.CreatedBy, &d.UpdatedBy,
	)
}
