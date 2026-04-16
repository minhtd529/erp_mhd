// Package repository provides the PostgreSQL implementation of the HRM domain
// repository interfaces.
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

// Repo implements domain.EmployeeRepository using pgxpool.
type Repo struct {
	pool *pgxpool.Pool
}

// New creates a new HRM Repo.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

const employeeCols = `
	id, full_name, email, phone, date_of_birth, grade, position,
	office_id, manager_id, hourly_rate, status,
	employment_date, contract_end_date,
	is_salesperson, sales_commission_eligible,
	bank_account_number_enc, bank_account_name,
	is_deleted, created_at, updated_at, created_by, updated_by`

// Create inserts a new employee and returns the full entity.
func (r *Repo) Create(ctx context.Context, p domain.CreateEmployeeParams) (*domain.Employee, error) {
	q := `
		INSERT INTO employees
			(full_name, email, phone, date_of_birth, grade, position,
			 office_id, manager_id, hourly_rate, employment_date, contract_end_date,
			 is_salesperson, sales_commission_eligible,
			 created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$14)
		RETURNING ` + employeeCols

	row := r.pool.QueryRow(ctx, q,
		p.FullName, p.Email, p.Phone, p.DateOfBirth, string(p.Grade), p.Position,
		p.OfficeID, p.ManagerID, p.HourlyRate,
		p.EmploymentDate, p.ContractEndDate,
		p.IsSalesperson, p.SalesCommissionEligible,
		p.CreatedBy,
	)
	e, err := scanEmployee(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrDuplicateEmail
		}
		return nil, fmt.Errorf("hrm.Create: %w", err)
	}
	return e, nil
}

// FindByID returns a single non-deleted employee.
func (r *Repo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error) {
	q := `SELECT ` + employeeCols + ` FROM employees WHERE id = $1 AND is_deleted = false`
	e, err := scanEmployee(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrEmployeeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("hrm.FindByID: %w", err)
	}
	return e, nil
}

// Update patches mutable profile fields and returns the refreshed entity.
// Bank details are updated separately via UpdateBankDetails.
func (r *Repo) Update(ctx context.Context, p domain.UpdateEmployeeParams) (*domain.Employee, error) {
	q := `
		UPDATE employees SET
			full_name                = $2,
			phone                    = $3,
			date_of_birth            = $4,
			grade                    = $5,
			position                 = $6,
			office_id                = $7,
			manager_id               = $8,
			hourly_rate              = $9,
			employment_date          = $10,
			contract_end_date        = $11,
			is_salesperson           = $12,
			sales_commission_eligible = $13,
			updated_by               = $14,
			updated_at               = NOW()
		WHERE id = $1 AND is_deleted = false
		RETURNING ` + employeeCols

	e, err := scanEmployee(r.pool.QueryRow(ctx, q,
		p.ID, p.FullName, p.Phone, p.DateOfBirth, string(p.Grade), p.Position,
		p.OfficeID, p.ManagerID, p.HourlyRate,
		p.EmploymentDate, p.ContractEndDate,
		p.IsSalesperson, p.SalesCommissionEligible,
		p.UpdatedBy,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrEmployeeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("hrm.Update: %w", err)
	}
	return e, nil
}

// UpdateBankDetails sets the encrypted bank account fields for an employee.
func (r *Repo) UpdateBankDetails(ctx context.Context, p domain.UpdateBankDetailsParams) (*domain.Employee, error) {
	q := `
		UPDATE employees SET
			bank_account_number_enc = $2,
			bank_account_name       = $3,
			updated_by              = $4,
			updated_at              = NOW()
		WHERE id = $1 AND is_deleted = false
		RETURNING ` + employeeCols

	e, err := scanEmployee(r.pool.QueryRow(ctx, q,
		p.ID, p.BankAccountNumberEnc, p.BankAccountName, p.UpdatedBy,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrEmployeeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("hrm.UpdateBankDetails: %w", err)
	}
	return e, nil
}

// SoftDelete marks an employee as deleted with status RESIGNED.
func (r *Repo) SoftDelete(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	const q = `
		UPDATE employees
		SET is_deleted = true, status = 'RESIGNED', updated_by = $2, updated_at = NOW()
		WHERE id = $1 AND is_deleted = false`

	tag, err := r.pool.Exec(ctx, q, id, deletedBy)
	if err != nil {
		return fmt.Errorf("hrm.SoftDelete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrEmployeeNotFound
	}
	return nil
}

// List returns a paginated list of non-deleted employees.
func (r *Repo) List(ctx context.Context, f domain.ListEmployeesFilter) ([]*domain.Employee, int64, error) {
	offset := (f.Page - 1) * f.Size
	args := []any{}
	where := "WHERE is_deleted = false"
	idx := 1

	if f.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, string(f.Status))
		idx++
	}
	if f.OfficeID != nil {
		where += fmt.Sprintf(" AND office_id = $%d", idx)
		args = append(args, f.OfficeID)
		idx++
	}
	if f.Q != "" {
		where += fmt.Sprintf(" AND (full_name ILIKE $%d OR email ILIKE $%d)", idx, idx)
		args = append(args, "%"+f.Q+"%")
		idx++
	}

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM employees "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("hrm.List count: %w", err)
	}

	args = append(args, f.Size, offset)
	dataQ := fmt.Sprintf(
		`SELECT `+employeeCols+` FROM employees %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, idx, idx+1,
	)

	rows, err := r.pool.Query(ctx, dataQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("hrm.List query: %w", err)
	}
	defer rows.Close()

	var employees []*domain.Employee
	for rows.Next() {
		e, err := scanEmployee(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("hrm.List scan: %w", err)
		}
		employees = append(employees, e)
	}
	if employees == nil {
		employees = []*domain.Employee{}
	}
	return employees, total, rows.Err()
}

// ─── helpers ─────────────────────────────────────────────────────────────────

type scanner interface {
	Scan(dest ...any) error
}

func scanEmployee(row scanner) (*domain.Employee, error) {
	var e domain.Employee
	var grade string
	var status string
	err := row.Scan(
		&e.ID, &e.FullName, &e.Email, &e.Phone, &e.DateOfBirth,
		&grade, &e.Position,
		&e.OfficeID, &e.ManagerID, &e.HourlyRate, &status,
		&e.EmploymentDate, &e.ContractEndDate,
		&e.IsSalesperson, &e.SalesCommissionEligible,
		&e.BankAccountNumberEnc, &e.BankAccountName,
		&e.IsDeleted, &e.CreatedAt, &e.UpdatedAt, &e.CreatedBy, &e.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	e.Grade = domain.EmployeeGrade(grade)
	e.Status = domain.EmployeeStatus(status)
	return &e, nil
}
