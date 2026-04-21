package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/hrm/domain"
)

// OrgRepo implements domain.OrgRepository using pgxpool.
type OrgRepo struct {
	pool *pgxpool.Pool
}

// NewOrgRepo creates a new OrgRepo.
func NewOrgRepo(pool *pgxpool.Pool) *OrgRepo {
	return &OrgRepo{pool: pool}
}

// ─── Branch columns ───────────────────────────────────────────────────────────

const hrmBranchCols = `
	id, code, name, address, phone, is_active,
	is_head_office, city, established_date, head_of_branch_user_id,
	tax_code, authorization_doc_number, authorization_date, authorization_file_id,
	created_at, updated_at, created_by, updated_by `

func scanHRMBranch(row scanner) (*domain.HRMBranch, error) {
	var b domain.HRMBranch
	err := row.Scan(
		&b.ID, &b.Code, &b.Name, &b.Address, &b.Phone, &b.IsActive,
		&b.IsHeadOffice, &b.City, &b.EstablishedDate, &b.HeadOfBranchUserID,
		&b.TaxCode, &b.AuthorizationDocNumber, &b.AuthorizationDate, &b.AuthorizationFileID,
		&b.CreatedAt, &b.UpdatedAt, &b.CreatedBy, &b.UpdatedBy,
	)
	return &b, err
}

// FindBranchByID returns a single active branch with HRM fields.
func (r *OrgRepo) FindBranchByID(ctx context.Context, id uuid.UUID) (*domain.HRMBranch, error) {
	q := `SELECT` + hrmBranchCols + `FROM branches WHERE id = $1 AND is_active = true`
	b, err := scanHRMBranch(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrBranchNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("hrm.FindBranchByID: %w", err)
	}
	return b, nil
}

// ListBranches returns paginated HRM-extended branches.
func (r *OrgRepo) ListBranches(ctx context.Context, f domain.ListHRMBranchesFilter) ([]*domain.HRMBranch, int64, error) {
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
		return nil, 0, fmt.Errorf("hrm.ListBranches count: %w", err)
	}

	args = append(args, f.Size, offset)
	dataQ := fmt.Sprintf(`SELECT`+hrmBranchCols+`FROM branches %s ORDER BY is_head_office DESC, name ASC LIMIT $%d OFFSET $%d`,
		where, idx, idx+1)
	rows, err := r.pool.Query(ctx, dataQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("hrm.ListBranches query: %w", err)
	}
	defer rows.Close()

	var branches []*domain.HRMBranch
	for rows.Next() {
		b, err := scanHRMBranch(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("hrm.ListBranches scan: %w", err)
		}
		branches = append(branches, b)
	}
	if branches == nil {
		branches = []*domain.HRMBranch{}
	}
	return branches, total, rows.Err()
}

// UpdateBranch updates mutable HRM branch fields (name, address, phone, city, tax_code, etc.).
// Code and is_head_office are not updatable via this path (immutable or controlled separately).
func (r *OrgRepo) UpdateBranch(ctx context.Context, p domain.UpdateHRMBranchParams) (*domain.HRMBranch, error) {
	q := `UPDATE branches
	      SET name                    = COALESCE($2, name),
	          address                 = COALESCE($3, address),
	          phone                   = COALESCE($4, phone),
	          city                    = COALESCE($5, city),
	          tax_code                = COALESCE($6, tax_code),
	          established_date        = COALESCE($7, established_date),
	          authorization_doc_number = COALESCE($8, authorization_doc_number),
	          authorization_date      = COALESCE($9, authorization_date),
	          updated_by              = $10,
	          updated_at              = NOW()
	      WHERE id = $1 AND is_active = true
	      RETURNING` + hrmBranchCols
	b, err := scanHRMBranch(r.pool.QueryRow(ctx, q,
		p.ID, p.Name, p.Address, p.Phone, p.City, p.TaxCode,
		p.EstablishedDate, p.AuthorizationDocNumber, p.AuthorizationDate,
		p.UpdatedBy,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrBranchNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("hrm.UpdateBranch: %w", err)
	}
	return b, nil
}

// ─── Department columns ───────────────────────────────────────────────────────

const hrmDeptCols = `
	id, code, name, branch_id, is_active, is_deleted,
	description, dept_type, head_employee_id,
	authorization_doc_number, authorization_date, authorization_file_id,
	created_at, updated_at, created_by, updated_by `

func scanHRMDept(row scanner) (*domain.HRMDepartment, error) {
	var d domain.HRMDepartment
	err := row.Scan(
		&d.ID, &d.Code, &d.Name, &d.BranchID, &d.IsActive, &d.IsDeleted,
		&d.Description, &d.DeptType, &d.HeadEmployeeID,
		&d.AuthorizationDocNumber, &d.AuthorizationDate, &d.AuthorizationFileID,
		&d.CreatedAt, &d.UpdatedAt, &d.CreatedBy, &d.UpdatedBy,
	)
	return &d, err
}

// FindDeptByID returns a single non-deleted department with HRM fields.
func (r *OrgRepo) FindDeptByID(ctx context.Context, id uuid.UUID) (*domain.HRMDepartment, error) {
	q := `SELECT` + hrmDeptCols + `FROM departments WHERE id = $1 AND is_deleted = false`
	d, err := scanHRMDept(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrDeptNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("hrm.FindDeptByID: %w", err)
	}
	return d, nil
}

// ListDepts returns paginated HRM-extended departments.
func (r *OrgRepo) ListDepts(ctx context.Context, f domain.ListHRMDeptsFilter) ([]*domain.HRMDepartment, int64, error) {
	offset := (f.Page - 1) * f.Size
	args := []any{}
	where := "WHERE is_deleted = false"
	idx := 1

	if f.BranchID != nil {
		where += fmt.Sprintf(" AND branch_id = $%d", idx)
		args = append(args, *f.BranchID)
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
		return nil, 0, fmt.Errorf("hrm.ListDepts count: %w", err)
	}

	args = append(args, f.Size, offset)
	dataQ := fmt.Sprintf(`SELECT`+hrmDeptCols+`FROM departments %s ORDER BY name ASC LIMIT $%d OFFSET $%d`,
		where, idx, idx+1)
	rows, err := r.pool.Query(ctx, dataQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("hrm.ListDepts query: %w", err)
	}
	defer rows.Close()

	var depts []*domain.HRMDepartment
	for rows.Next() {
		d, err := scanHRMDept(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("hrm.ListDepts scan: %w", err)
		}
		depts = append(depts, d)
	}
	if depts == nil {
		depts = []*domain.HRMDepartment{}
	}
	return depts, total, rows.Err()
}

// UpdateDept updates mutable HRM department fields.
func (r *OrgRepo) UpdateDept(ctx context.Context, p domain.UpdateHRMDeptParams) (*domain.HRMDepartment, error) {
	q := `UPDATE departments
	      SET name                    = COALESCE($2, name),
	          description             = COALESCE($3, description),
	          dept_type               = COALESCE($4, dept_type),
	          authorization_doc_number = COALESCE($5, authorization_doc_number),
	          authorization_date      = COALESCE($6, authorization_date),
	          updated_by              = $7,
	          updated_at              = NOW()
	      WHERE id = $1 AND is_deleted = false
	      RETURNING` + hrmDeptCols
	d, err := scanHRMDept(r.pool.QueryRow(ctx, q,
		p.ID, p.Name, p.Description, p.DeptType, p.AuthorizationDocNumber, p.AuthorizationDate, p.UpdatedBy,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrDeptNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("hrm.UpdateDept: %w", err)
	}
	return d, nil
}

// ─── BranchDepartment ────────────────────────────────────────────────────────

const bdCols = `branch_id, department_id, head_employee_id, is_active, created_at`

func scanBranchDept(row scanner) (*domain.BranchDepartment, error) {
	var bd domain.BranchDepartment
	err := row.Scan(&bd.BranchID, &bd.DepartmentID, &bd.HeadEmployeeID, &bd.IsActive, &bd.CreatedAt)
	return &bd, err
}

// ListBranchDepts returns the branch-department matrix with optional branch filter.
func (r *OrgRepo) ListBranchDepts(ctx context.Context, f domain.ListBranchDeptsFilter) ([]*domain.BranchDepartment, int64, error) {
	offset := (f.Page - 1) * f.Size
	args := []any{}
	where := "WHERE 1=1"
	idx := 1

	if f.BranchID != nil {
		where += fmt.Sprintf(" AND branch_id = $%d", idx)
		args = append(args, *f.BranchID)
		idx++
	}
	if f.IsActive != nil {
		where += fmt.Sprintf(" AND is_active = $%d", idx)
		args = append(args, *f.IsActive)
		idx++
	}

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM branch_departments "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("hrm.ListBranchDepts count: %w", err)
	}

	args = append(args, f.Size, offset)
	dataQ := fmt.Sprintf(`SELECT `+bdCols+` FROM branch_departments %s ORDER BY branch_id, department_id LIMIT $%d OFFSET $%d`,
		where, idx, idx+1)
	rows, err := r.pool.Query(ctx, dataQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("hrm.ListBranchDepts query: %w", err)
	}
	defer rows.Close()

	var bds []*domain.BranchDepartment
	for rows.Next() {
		bd, err := scanBranchDept(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("hrm.ListBranchDepts scan: %w", err)
		}
		bds = append(bds, bd)
	}
	if bds == nil {
		bds = []*domain.BranchDepartment{}
	}
	return bds, total, rows.Err()
}

// CreateBranchDept links a branch to a department.
func (r *OrgRepo) CreateBranchDept(ctx context.Context, p domain.CreateBranchDeptParams) (*domain.BranchDepartment, error) {
	q := `INSERT INTO branch_departments (branch_id, department_id)
	      VALUES ($1, $2)
	      ON CONFLICT DO NOTHING
	      RETURNING ` + bdCols
	bd, err := scanBranchDept(r.pool.QueryRow(ctx, q, p.BranchID, p.DepartmentID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrDuplicateBranchDept
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrDuplicateBranchDept
		}
		return nil, fmt.Errorf("hrm.CreateBranchDept: %w", err)
	}
	return bd, nil
}

// SoftDeleteBranchDept sets is_active=false for the branch-department pair.
func (r *OrgRepo) SoftDeleteBranchDept(ctx context.Context, branchID, deptID uuid.UUID) error {
	q := `UPDATE branch_departments SET is_active = false
	      WHERE branch_id = $1 AND department_id = $2 AND is_active = true`
	tag, err := r.pool.Exec(ctx, q, branchID, deptID)
	if err != nil {
		return fmt.Errorf("hrm.SoftDeleteBranchDept: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrBranchDeptNotFound
	}
	return nil
}

// AssignBranchHead sets head_of_branch_user_id for a branch.
func (r *OrgRepo) AssignBranchHead(ctx context.Context, p domain.AssignBranchHeadParams) (*domain.HRMBranch, error) {
	q := `UPDATE branches
	      SET head_of_branch_user_id = $2, updated_by = $3, updated_at = NOW()
	      WHERE id = $1 AND is_active = true
	      RETURNING` + hrmBranchCols
	b, err := scanHRMBranch(r.pool.QueryRow(ctx, q, p.BranchID, p.UserID, p.UpdatedBy))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrBranchNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("hrm.AssignBranchHead: %w", err)
	}
	return b, nil
}

// DeactivateBranch sets is_active=false; returns ErrBranchNotFound if already inactive.
// Guards: caller must verify no active employees before calling (employee check in usecase).
func (r *OrgRepo) DeactivateBranch(ctx context.Context, id uuid.UUID, updatedBy *uuid.UUID) error {
	q := `UPDATE branches SET is_active = false, updated_by = $2, updated_at = NOW()
	      WHERE id = $1 AND is_active = true`
	tag, err := r.pool.Exec(ctx, q, id, updatedBy)
	if err != nil {
		return fmt.Errorf("hrm.DeactivateBranch: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrBranchNotFound
	}
	return nil
}

// AssignDeptHead sets head_employee_id for a department.
func (r *OrgRepo) AssignDeptHead(ctx context.Context, p domain.AssignDeptHeadParams) (*domain.HRMDepartment, error) {
	q := `UPDATE departments
	      SET head_employee_id = $2, updated_by = $3, updated_at = NOW()
	      WHERE id = $1 AND is_deleted = false
	      RETURNING` + hrmDeptCols
	d, err := scanHRMDept(r.pool.QueryRow(ctx, q, p.DeptID, p.EmployeeID, p.UpdatedBy))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrDeptNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("hrm.AssignDeptHead: %w", err)
	}
	return d, nil
}

// DeactivateDept sets is_active=false; returns ErrDeptNotFound if already inactive/deleted.
func (r *OrgRepo) DeactivateDept(ctx context.Context, id uuid.UUID, updatedBy *uuid.UUID) error {
	q := `UPDATE departments SET is_active = false, updated_by = $2, updated_at = NOW()
	      WHERE id = $1 AND is_active = true AND is_deleted = false`
	tag, err := r.pool.Exec(ctx, q, id, updatedBy)
	if err != nil {
		return fmt.Errorf("hrm.DeactivateDept: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrDeptNotFound
	}
	return nil
}

// ListBranchesWithDepts returns the full org-chart tree (all active branches + their active depts).
func (r *OrgRepo) ListBranchesWithDepts(ctx context.Context) ([]*domain.OrgChartBranch, error) {
	// Load active branches
	bq := `SELECT id, code, name, is_head_office FROM branches WHERE is_active = true ORDER BY is_head_office DESC, name ASC`
	brows, err := r.pool.Query(ctx, bq)
	if err != nil {
		return nil, fmt.Errorf("hrm.ListBranchesWithDepts branches: %w", err)
	}
	defer brows.Close()

	branchMap := map[uuid.UUID]*domain.OrgChartBranch{}
	var ordered []uuid.UUID
	for brows.Next() {
		var b domain.OrgChartBranch
		if err := brows.Scan(&b.ID, &b.Code, &b.Name, &b.IsHeadOffice); err != nil {
			return nil, fmt.Errorf("hrm.ListBranchesWithDepts branch scan: %w", err)
		}
		b.Departments = []domain.OrgChartDept{}
		branchMap[b.ID] = &b
		ordered = append(ordered, b.ID)
	}
	if err := brows.Err(); err != nil {
		return nil, fmt.Errorf("hrm.ListBranchesWithDepts branch rows: %w", err)
	}

	if len(ordered) == 0 {
		return []*domain.OrgChartBranch{}, nil
	}

	// Load active branch-dept links + dept info
	dq := `SELECT bd.branch_id, d.id, d.code, d.name, d.dept_type
	       FROM branch_departments bd
	       JOIN departments d ON d.id = bd.department_id
	       WHERE bd.is_active = true AND d.is_deleted = false AND d.is_active = true
	       ORDER BY bd.branch_id, d.name ASC`
	drows, err := r.pool.Query(ctx, dq)
	if err != nil {
		return nil, fmt.Errorf("hrm.ListBranchesWithDepts depts: %w", err)
	}
	defer drows.Close()

	for drows.Next() {
		var branchID uuid.UUID
		var dept domain.OrgChartDept
		if err := drows.Scan(&branchID, &dept.ID, &dept.Code, &dept.Name, &dept.DeptType); err != nil {
			return nil, fmt.Errorf("hrm.ListBranchesWithDepts dept scan: %w", err)
		}
		if b, ok := branchMap[branchID]; ok {
			b.Departments = append(b.Departments, dept)
		}
	}
	if err := drows.Err(); err != nil {
		return nil, fmt.Errorf("hrm.ListBranchesWithDepts dept rows: %w", err)
	}

	result := make([]*domain.OrgChartBranch, 0, len(ordered))
	for _, id := range ordered {
		result = append(result, branchMap[id])
	}
	return result, nil
}

