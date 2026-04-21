package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/hrm/domain"
)

// ─── EmployeeRepo ─────────────────────────────────────────────────────────────

type EmployeeRepo struct{ pool *pgxpool.Pool }

func NewEmployeeRepo(pool *pgxpool.Pool) *EmployeeRepo { return &EmployeeRepo{pool: pool} }

// employeeCols is the full ordered column list for SELECT on employees.
// Column order MUST match scanEmployee exactly.
const employeeCols = `
	id, full_name, email, phone, date_of_birth, grade, position, office_id, manager_id,
	hourly_rate, status, employment_date, contract_end_date, is_deleted,
	created_at, updated_at, created_by, updated_by,
	employee_code, user_id, branch_id, department_id, position_title,
	employment_type, hired_date, probation_end_date, termination_date,
	termination_reason, current_contract_id,
	display_name, gender, place_of_birth, nationality, ethnicity,
	personal_email, personal_phone, work_phone, current_address, permanent_address,
	cccd_encrypted, cccd_issued_date, cccd_issued_place, passport_number, passport_expiry,
	hired_source, referrer_employee_id, probation_salary_pct, work_location, remote_days_per_week,
	education_level, education_major, education_school, education_graduation_year,
	vn_cpa_number, vn_cpa_issued_date, vn_cpa_expiry_date,
	practicing_certificate_number, practicing_certificate_expiry,
	base_salary, salary_currency, salary_effective_date,
	bank_account_encrypted, bank_name, bank_branch, mst_ca_nhan_encrypted,
	commission_rate, commission_type, sales_target_yearly, biz_dev_region,
	so_bhxh_encrypted, bhxh_registered_date, bhxh_province_code,
	bhyt_card_number, bhyt_expiry_date, bhyt_registered_hospital_code, bhyt_registered_hospital_name,
	tncn_registered `

func scanEmployee(row scanner) (*domain.Employee, error) {
	var e domain.Employee
	err := row.Scan(
		// base (18 cols)
		&e.ID, &e.FullName, &e.Email, &e.Phone, &e.DateOfBirth, &e.Grade, &e.Position,
		&e.OfficeID, &e.ManagerID, &e.HourlyRate, &e.Status, &e.EmploymentDate,
		&e.ContractEndDate, &e.IsDeleted, &e.CreatedAt, &e.UpdatedAt, &e.CreatedBy, &e.UpdatedBy,
		// extended identity (11 cols)
		&e.EmployeeCode, &e.UserID, &e.BranchID, &e.DepartmentID, &e.PositionTitle,
		&e.EmploymentType, &e.HiredDate, &e.ProbationEndDate, &e.TerminationDate,
		&e.TerminationReason, &e.CurrentContractID,
		// personal (15 cols)
		&e.DisplayName, &e.Gender, &e.PlaceOfBirth, &e.Nationality, &e.Ethnicity,
		&e.PersonalEmail, &e.PersonalPhone, &e.WorkPhone, &e.CurrentAddress, &e.PermanentAddress,
		&e.CccdEncrypted, &e.CccdIssuedDate, &e.CccdIssuedPlace, &e.PassportNumber, &e.PassportExpiry,
		// employment (5 cols)
		&e.HiredSource, &e.ReferrerEmployeeID, &e.ProbationSalaryPct, &e.WorkLocation, &e.RemoteDaysPerWeek,
		// qualifications (9 cols)
		&e.EducationLevel, &e.EducationMajor, &e.EducationSchool, &e.EducationGraduationYear,
		&e.VnCpaNumber, &e.VnCpaIssuedDate, &e.VnCpaExpiryDate,
		&e.PracticingCertNumber, &e.PracticingCertExpiry,
		// salary/bank (7 cols)
		&e.BaseSalary, &e.SalaryCurrency, &e.SalaryEffectiveDate,
		&e.BankAccountEncrypted, &e.BankName, &e.BankBranch, &e.MstCaNhanEncrypted,
		// commission (4 cols)
		&e.CommissionRate, &e.CommissionType, &e.SalesTargetYearly, &e.BizDevRegion,
		// insurance (8 cols)
		&e.SoBhxhEncrypted, &e.BhxhRegisteredDate, &e.BhxhProvinceCode,
		&e.BhytCardNumber, &e.BhytExpiryDate,
		&e.BhytRegisteredHospitalCode, &e.BhytRegisteredHospitalName,
		&e.TncnRegistered,
	)
	return &e, err
}

func (r *EmployeeRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error) {
	q := `SELECT` + employeeCols + `FROM employees WHERE id = $1 AND is_deleted = false`
	e, err := scanEmployee(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrEmployeeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("employee.FindByID: %w", err)
	}
	return e, nil
}

func (r *EmployeeRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*domain.Employee, error) {
	q := `SELECT` + employeeCols + `FROM employees WHERE user_id = $1 AND is_deleted = false`
	e, err := scanEmployee(r.pool.QueryRow(ctx, q, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrEmployeeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("employee.FindByUserID: %w", err)
	}
	return e, nil
}

func (r *EmployeeRepo) List(ctx context.Context, f domain.ListEmployeesFilter) ([]*domain.Employee, int64, error) {
	offset := (f.Page - 1) * f.Size
	args := []any{}
	where := "WHERE is_deleted = false"
	idx := 1

	if f.UserID != nil {
		where += fmt.Sprintf(" AND user_id = $%d", idx)
		args = append(args, *f.UserID)
		idx++
	}
	if f.BranchScope != nil {
		where += fmt.Sprintf(" AND branch_id = $%d", idx)
		args = append(args, *f.BranchScope)
		idx++
	}
	if f.BranchID != nil {
		where += fmt.Sprintf(" AND branch_id = $%d", idx)
		args = append(args, *f.BranchID)
		idx++
	}
	if f.DepartmentID != nil {
		where += fmt.Sprintf(" AND department_id = $%d", idx)
		args = append(args, *f.DepartmentID)
		idx++
	}
	if f.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, *f.Status)
		idx++
	}
	if f.Grade != nil {
		where += fmt.Sprintf(" AND grade = $%d", idx)
		args = append(args, *f.Grade)
		idx++
	}
	if f.Q != "" {
		where += fmt.Sprintf(` AND (full_name ILIKE $%d OR email ILIKE $%d OR employee_code ILIKE $%d)`, idx, idx+1, idx+2)
		q := "%" + f.Q + "%"
		args = append(args, q, q, q)
		idx += 3
	}

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM employees "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("employee.List count: %w", err)
	}

	args = append(args, f.Size, offset)
	q := `SELECT` + employeeCols + `FROM employees ` + where +
		fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("employee.List query: %w", err)
	}
	defer rows.Close()

	var employees []*domain.Employee
	for rows.Next() {
		e, err := scanEmployee(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("employee.List scan: %w", err)
		}
		employees = append(employees, e)
	}
	if employees == nil {
		employees = []*domain.Employee{}
	}
	return employees, total, nil
}

func (r *EmployeeRepo) Create(ctx context.Context, p domain.CreateEmployeeParams) (*domain.Employee, error) {
	q := `INSERT INTO employees
		(full_name, email, phone, date_of_birth, grade, manager_id, status,
		 branch_id, department_id, position_title, employment_type, hired_date,
		 display_name, gender, personal_email, personal_phone,
		 work_location, hired_source, education_level, commission_type, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
		RETURNING ` + employeeCols
	e, err := scanEmployee(r.pool.QueryRow(ctx, q,
		p.FullName, p.Email, p.Phone, p.DateOfBirth, p.Grade, p.ManagerID, p.Status,
		p.BranchID, p.DepartmentID, p.PositionTitle, p.EmploymentType, p.HiredDate,
		p.DisplayName, p.Gender, p.PersonalEmail, p.PersonalPhone,
		p.WorkLocation, p.HiredSource, p.EducationLevel, p.CommissionType, p.CreatedBy,
	))
	if err != nil {
		if isEmpUniqueViolation(err) {
			return nil, domain.ErrEmployeeEmailConflict
		}
		return nil, fmt.Errorf("employee.Create: %w", err)
	}
	return e, nil
}

func (r *EmployeeRepo) Update(ctx context.Context, p domain.UpdateEmployeeParams) (*domain.Employee, error) {
	parts := []string{}
	args := []any{}
	idx := 1

	appendStr := func(col string, v *string) {
		if v != nil {
			parts = append(parts, fmt.Sprintf("%s = $%d", col, idx))
			args = append(args, *v)
			idx++
		}
	}
	appendUUID := func(col string, v *uuid.UUID) {
		if v != nil {
			parts = append(parts, fmt.Sprintf("%s = $%d", col, idx))
			args = append(args, *v)
			idx++
		}
	}
	appendTime := func(col string, v *time.Time) {
		if v != nil {
			parts = append(parts, fmt.Sprintf("%s = $%d", col, idx))
			args = append(args, *v)
			idx++
		}
	}
	appendInt16 := func(col string, v *int16) {
		if v != nil {
			parts = append(parts, fmt.Sprintf("%s = $%d", col, idx))
			args = append(args, *v)
			idx++
		}
	}
	appendFloat := func(col string, v *float64) {
		if v != nil {
			parts = append(parts, fmt.Sprintf("%s = $%d", col, idx))
			args = append(args, *v)
			idx++
		}
	}

	appendStr("full_name", p.FullName)
	appendStr("phone", p.Phone)
	appendStr("grade", p.Grade)
	appendUUID("manager_id", p.ManagerID)
	appendStr("status", p.Status)
	appendUUID("branch_id", p.BranchID)
	appendUUID("department_id", p.DepartmentID)
	appendStr("position_title", p.PositionTitle)
	appendStr("employment_type", p.EmploymentType)
	appendTime("hired_date", p.HiredDate)
	appendTime("probation_end_date", p.ProbationEndDate)
	appendTime("termination_date", p.TerminationDate)
	appendStr("termination_reason", p.TerminationReason)
	appendUUID("current_contract_id", p.CurrentContractID)
	appendStr("display_name", p.DisplayName)
	appendStr("gender", p.Gender)
	appendStr("personal_email", p.PersonalEmail)
	appendStr("personal_phone", p.PersonalPhone)
	appendStr("work_phone", p.WorkPhone)
	appendStr("current_address", p.CurrentAddress)
	appendStr("permanent_address", p.PermanentAddress)
	appendStr("work_location", p.WorkLocation)
	appendInt16("remote_days_per_week", p.RemoteDaysPerWeek)
	appendStr("hired_source", p.HiredSource)
	appendStr("education_level", p.EducationLevel)
	appendStr("education_major", p.EducationMajor)
	appendStr("education_school", p.EducationSchool)
	appendInt16("education_graduation_year", p.EducationGraduationYear)
	appendStr("vn_cpa_number", p.VnCpaNumber)
	appendStr("practicing_certificate_number", p.PracticingCertNumber)
	appendStr("commission_type", p.CommissionType)
	appendFloat("commission_rate", p.CommissionRate)
	appendStr("biz_dev_region", p.BizDevRegion)
	appendStr("nationality", p.Nationality)
	appendStr("ethnicity", p.Ethnicity)

	if len(parts) == 0 {
		return r.FindByID(ctx, p.ID)
	}

	parts = append(parts, fmt.Sprintf("updated_at = $%d", idx), fmt.Sprintf("updated_by = $%d", idx+1))
	args = append(args, time.Now(), p.UpdatedBy, p.ID)
	idx += 2

	q := `UPDATE employees SET ` + strings.Join(parts, ", ") +
		fmt.Sprintf(` WHERE id = $%d AND is_deleted = false RETURNING `, idx) + employeeCols

	e, err := scanEmployee(r.pool.QueryRow(ctx, q, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrEmployeeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("employee.Update: %w", err)
	}
	return e, nil
}

func (r *EmployeeRepo) SoftDelete(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE employees SET is_deleted = true, updated_at = NOW(), updated_by = $2
		 WHERE id = $1 AND is_deleted = false`,
		id, deletedBy,
	)
	if err != nil {
		return fmt.Errorf("employee.SoftDelete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrEmployeeNotFound
	}
	return nil
}

func (r *EmployeeRepo) UpdateProfile(ctx context.Context, p domain.UpdateProfileParams) (*domain.Employee, error) {
	q := `UPDATE employees SET
		display_name      = COALESCE($1, display_name),
		personal_phone    = COALESCE($2, personal_phone),
		personal_email    = COALESCE($3, personal_email),
		current_address   = COALESCE($4, current_address),
		permanent_address = COALESCE($5, permanent_address),
		updated_at        = NOW(),
		updated_by        = $6
		WHERE user_id = $7 AND is_deleted = false
		RETURNING ` + employeeCols
	e, err := scanEmployee(r.pool.QueryRow(ctx, q,
		p.DisplayName, p.PersonalPhone, p.PersonalEmail,
		p.CurrentAddress, p.PermanentAddress, p.UpdatedBy, p.UserID,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrEmployeeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("employee.UpdateProfile: %w", err)
	}
	return e, nil
}

func isEmpUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// ─── DependentRepo ────────────────────────────────────────────────────────────

type DependentRepo struct{ pool *pgxpool.Pool }

func NewDependentRepo(pool *pgxpool.Pool) *DependentRepo { return &DependentRepo{pool: pool} }

const dependentCols = `
	id, employee_id, full_name, relationship, date_of_birth, cccd_or_birth_cert,
	tax_deduction_registered, tax_deduction_from, tax_deduction_to, notes,
	created_at, updated_at `

func scanDependent(row scanner) (*domain.EmployeeDependent, error) {
	var d domain.EmployeeDependent
	err := row.Scan(
		&d.ID, &d.EmployeeID, &d.FullName, &d.Relationship, &d.DateOfBirth, &d.CccdOrBirthCert,
		&d.TaxDeductionRegistered, &d.TaxDeductionFrom, &d.TaxDeductionTo, &d.Notes,
		&d.CreatedAt, &d.UpdatedAt,
	)
	return &d, err
}

func (r *DependentRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeDependent, error) {
	q := `SELECT` + dependentCols + `FROM employee_dependents WHERE id = $1`
	d, err := scanDependent(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrDependentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("dependent.FindByID: %w", err)
	}
	return d, nil
}

func (r *DependentRepo) ListByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]*domain.EmployeeDependent, error) {
	q := `SELECT` + dependentCols + `FROM employee_dependents WHERE employee_id = $1 ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, q, employeeID)
	if err != nil {
		return nil, fmt.Errorf("dependent.List: %w", err)
	}
	defer rows.Close()

	var deps []*domain.EmployeeDependent
	for rows.Next() {
		d, err := scanDependent(rows)
		if err != nil {
			return nil, fmt.Errorf("dependent.List scan: %w", err)
		}
		deps = append(deps, d)
	}
	if deps == nil {
		deps = []*domain.EmployeeDependent{}
	}
	return deps, nil
}

func (r *DependentRepo) Create(ctx context.Context, p domain.CreateDependentParams) (*domain.EmployeeDependent, error) {
	q := `INSERT INTO employee_dependents
		(employee_id, full_name, relationship, date_of_birth, cccd_or_birth_cert,
		 tax_deduction_registered, tax_deduction_from, tax_deduction_to, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING ` + dependentCols
	d, err := scanDependent(r.pool.QueryRow(ctx, q,
		p.EmployeeID, p.FullName, p.Relationship, p.DateOfBirth, p.CccdOrBirthCert,
		p.TaxDeductionRegistered, p.TaxDeductionFrom, p.TaxDeductionTo, p.Notes,
	))
	if err != nil {
		return nil, fmt.Errorf("dependent.Create: %w", err)
	}
	return d, nil
}

func (r *DependentRepo) Update(ctx context.Context, p domain.UpdateDependentParams) (*domain.EmployeeDependent, error) {
	parts := []string{}
	args := []any{}
	idx := 1

	if p.FullName != nil {
		parts = append(parts, fmt.Sprintf("full_name = $%d", idx)); args = append(args, *p.FullName); idx++
	}
	if p.Relationship != nil {
		parts = append(parts, fmt.Sprintf("relationship = $%d", idx)); args = append(args, *p.Relationship); idx++
	}
	if p.DateOfBirth != nil {
		parts = append(parts, fmt.Sprintf("date_of_birth = $%d", idx)); args = append(args, *p.DateOfBirth); idx++
	}
	if p.CccdOrBirthCert != nil {
		parts = append(parts, fmt.Sprintf("cccd_or_birth_cert = $%d", idx)); args = append(args, *p.CccdOrBirthCert); idx++
	}
	if p.TaxDeductionRegistered != nil {
		parts = append(parts, fmt.Sprintf("tax_deduction_registered = $%d", idx)); args = append(args, *p.TaxDeductionRegistered); idx++
	}
	if p.TaxDeductionFrom != nil {
		parts = append(parts, fmt.Sprintf("tax_deduction_from = $%d", idx)); args = append(args, *p.TaxDeductionFrom); idx++
	}
	if p.TaxDeductionTo != nil {
		parts = append(parts, fmt.Sprintf("tax_deduction_to = $%d", idx)); args = append(args, *p.TaxDeductionTo); idx++
	}
	if p.Notes != nil {
		parts = append(parts, fmt.Sprintf("notes = $%d", idx)); args = append(args, *p.Notes); idx++
	}

	if len(parts) == 0 {
		return r.FindByID(ctx, p.ID)
	}

	parts = append(parts, fmt.Sprintf("updated_at = $%d", idx))
	args = append(args, time.Now(), p.ID, p.EmployeeID)
	idx++

	q := `UPDATE employee_dependents SET ` + strings.Join(parts, ", ") +
		fmt.Sprintf(` WHERE id = $%d AND employee_id = $%d RETURNING `, idx, idx+1) + dependentCols

	d, err := scanDependent(r.pool.QueryRow(ctx, q, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrDependentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("dependent.Update: %w", err)
	}
	return d, nil
}

func (r *DependentRepo) Delete(ctx context.Context, id, employeeID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM employee_dependents WHERE id = $1 AND employee_id = $2`, id, employeeID)
	if err != nil {
		return fmt.Errorf("dependent.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrDependentNotFound
	}
	return nil
}

// ─── ContractRepo ─────────────────────────────────────────────────────────────

type ContractRepo struct{ pool *pgxpool.Pool }

func NewContractRepo(pool *pgxpool.Pool) *ContractRepo { return &ContractRepo{pool: pool} }

const contractCols = `
	id, employee_id, contract_number, contract_type, start_date, end_date, signed_date,
	salary_at_signing, position_at_signing, notes, document_url, is_current,
	created_by, created_at, updated_at `

func scanContract(row scanner) (*domain.EmploymentContract, error) {
	var c domain.EmploymentContract
	err := row.Scan(
		&c.ID, &c.EmployeeID, &c.ContractNumber, &c.ContractType, &c.StartDate, &c.EndDate, &c.SignedDate,
		&c.SalaryAtSigning, &c.PositionAtSigning, &c.Notes, &c.DocumentURL, &c.IsCurrent,
		&c.CreatedBy, &c.CreatedAt, &c.UpdatedAt,
	)
	return &c, err
}

func (r *ContractRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.EmploymentContract, error) {
	q := `SELECT` + contractCols + `FROM employment_contracts WHERE id = $1`
	c, err := scanContract(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrContractNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("contract.FindByID: %w", err)
	}
	return c, nil
}

func (r *ContractRepo) ListByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]*domain.EmploymentContract, error) {
	q := `SELECT` + contractCols + `FROM employment_contracts WHERE employee_id = $1 ORDER BY start_date DESC`
	rows, err := r.pool.Query(ctx, q, employeeID)
	if err != nil {
		return nil, fmt.Errorf("contract.List: %w", err)
	}
	defer rows.Close()

	var contracts []*domain.EmploymentContract
	for rows.Next() {
		c, err := scanContract(rows)
		if err != nil {
			return nil, fmt.Errorf("contract.List scan: %w", err)
		}
		contracts = append(contracts, c)
	}
	if contracts == nil {
		contracts = []*domain.EmploymentContract{}
	}
	return contracts, nil
}

func (r *ContractRepo) HasActiveContract(ctx context.Context, employeeID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM employment_contracts WHERE employee_id = $1 AND is_current = true)`,
		employeeID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("contract.HasActive: %w", err)
	}
	return exists, nil
}

func (r *ContractRepo) Create(ctx context.Context, p domain.CreateContractParams) (*domain.EmploymentContract, error) {
	q := `INSERT INTO employment_contracts
		(employee_id, contract_number, contract_type, start_date, end_date, signed_date,
		 salary_at_signing, position_at_signing, notes, document_url, is_current, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,true,$11)
		RETURNING ` + contractCols
	c, err := scanContract(r.pool.QueryRow(ctx, q,
		p.EmployeeID, p.ContractNumber, p.ContractType, p.StartDate, p.EndDate, p.SignedDate,
		p.SalaryAtSigning, p.PositionAtSigning, p.Notes, p.DocumentURL, p.CreatedBy,
	))
	if err != nil {
		return nil, fmt.Errorf("contract.Create: %w", err)
	}
	return c, nil
}

func (r *ContractRepo) Update(ctx context.Context, p domain.UpdateContractParams) (*domain.EmploymentContract, error) {
	parts := []string{}
	args := []any{}
	idx := 1

	if p.ContractType != nil {
		parts = append(parts, fmt.Sprintf("contract_type = $%d", idx)); args = append(args, *p.ContractType); idx++
	}
	if p.EndDate != nil {
		parts = append(parts, fmt.Sprintf("end_date = $%d", idx)); args = append(args, *p.EndDate); idx++
	}
	if p.SignedDate != nil {
		parts = append(parts, fmt.Sprintf("signed_date = $%d", idx)); args = append(args, *p.SignedDate); idx++
	}
	if p.SalaryAtSigning != nil {
		parts = append(parts, fmt.Sprintf("salary_at_signing = $%d", idx)); args = append(args, *p.SalaryAtSigning); idx++
	}
	if p.PositionAtSigning != nil {
		parts = append(parts, fmt.Sprintf("position_at_signing = $%d", idx)); args = append(args, *p.PositionAtSigning); idx++
	}
	if p.Notes != nil {
		parts = append(parts, fmt.Sprintf("notes = $%d", idx)); args = append(args, *p.Notes); idx++
	}
	if p.DocumentURL != nil {
		parts = append(parts, fmt.Sprintf("document_url = $%d", idx)); args = append(args, *p.DocumentURL); idx++
	}

	if len(parts) == 0 {
		return r.FindByID(ctx, p.ID)
	}

	parts = append(parts, fmt.Sprintf("updated_at = $%d", idx))
	args = append(args, time.Now(), p.ID, p.EmployeeID)
	idx++

	q := `UPDATE employment_contracts SET ` + strings.Join(parts, ", ") +
		fmt.Sprintf(` WHERE id = $%d AND employee_id = $%d RETURNING `, idx, idx+1) + contractCols

	c, err := scanContract(r.pool.QueryRow(ctx, q, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrContractNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("contract.Update: %w", err)
	}
	return c, nil
}

func (r *ContractRepo) Terminate(ctx context.Context, id, employeeID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE employment_contracts SET is_current = false, updated_at = NOW()
		 WHERE id = $1 AND employee_id = $2 AND is_current = true`,
		id, employeeID,
	)
	if err != nil {
		return fmt.Errorf("contract.Terminate: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrContractNotFound
	}
	return nil
}
