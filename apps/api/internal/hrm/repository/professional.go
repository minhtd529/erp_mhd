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

// pgCheckViolation maps PostgreSQL check_violation (23514) to domain.ErrValidation.
// Returns nil for any other error so callers can use: if e := pgCheckViolation(err); e != nil { return nil, e }
func pgCheckViolation(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23514" {
		return fmt.Errorf("%w: %s", domain.ErrValidation, pgErr.ConstraintName)
	}
	return nil
}

// ─── CertRepo ─────────────────────────────────────────────────────────────────

type CertRepo struct{ pool *pgxpool.Pool }

func NewCertRepo(pool *pgxpool.Pool) *CertRepo { return &CertRepo{pool: pool} }

const certCols = `
	id, employee_id, cert_type, cert_name, cert_number,
	issued_date, expiry_date, issuing_authority,
	status, document_url, notes, is_deleted, created_by, created_at, updated_at`

func scanCert(row scanner) (*domain.Certification, error) {
	var c domain.Certification
	err := row.Scan(
		&c.ID, &c.EmployeeID, &c.CertType, &c.CertName, &c.CertNumber,
		&c.IssuedDate, &c.ExpiryDate, &c.IssuingAuthority,
		&c.Status, &c.DocumentURL, &c.Notes, &c.IsDeleted, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt,
	)
	return &c, err
}

func (r *CertRepo) Create(ctx context.Context, p domain.CreateCertificationParams) (*domain.Certification, error) {
	const q = `
		INSERT INTO certifications
			(employee_id, cert_type, cert_name, cert_number, issued_date, expiry_date,
			 issuing_authority, status, document_url, notes, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING ` + certCols
	c, err := scanCert(r.pool.QueryRow(ctx, q,
		p.EmployeeID, p.CertType, p.CertName, p.CertNumber, p.IssuedDate, p.ExpiryDate,
		p.IssuingAuthority, p.Status, p.DocumentURL, p.Notes, p.CreatedBy,
	))
	if err != nil {
		if e := pgCheckViolation(err); e != nil {
			return nil, e
		}
		return nil, fmt.Errorf("CertRepo.Create: %w", err)
	}
	return c, nil
}

func (r *CertRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Certification, error) {
	q := `SELECT ` + certCols + ` FROM certifications WHERE id = $1`
	c, err := scanCert(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrCertificationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("CertRepo.FindByID: %w", err)
	}
	return c, nil
}

func (r *CertRepo) ListByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*domain.Certification, error) {
	q := `SELECT ` + certCols + ` FROM certifications
		WHERE employee_id = $1 AND is_deleted = false
		ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, employeeID)
	if err != nil {
		return nil, fmt.Errorf("CertRepo.ListByEmployee: %w", err)
	}
	defer rows.Close()
	var out []*domain.Certification
	for rows.Next() {
		c, err := scanCert(rows)
		if err != nil {
			return nil, fmt.Errorf("CertRepo.ListByEmployee scan: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CertRepo) Update(ctx context.Context, p domain.UpdateCertificationParams) (*domain.Certification, error) {
	const q = `
		UPDATE certifications SET
			cert_type=$2, cert_name=$3, cert_number=$4,
			issued_date=$5, expiry_date=$6, issuing_authority=$7,
			status=$8, document_url=$9, notes=$10, updated_at=now()
		WHERE id=$1 AND is_deleted=false
		RETURNING ` + certCols
	c, err := scanCert(r.pool.QueryRow(ctx, q,
		p.ID, p.CertType, p.CertName, p.CertNumber,
		p.IssuedDate, p.ExpiryDate, p.IssuingAuthority,
		p.Status, p.DocumentURL, p.Notes,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrCertificationNotFound
	}
	if err != nil {
		if e := pgCheckViolation(err); e != nil {
			return nil, e
		}
		return nil, fmt.Errorf("CertRepo.Update: %w", err)
	}
	return c, nil
}

func (r *CertRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE certifications SET is_deleted=true, updated_at=now() WHERE id=$1 AND is_deleted=false`, id)
	if err != nil {
		return fmt.Errorf("CertRepo.SoftDelete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrCertificationNotFound
	}
	return nil
}

func (r *CertRepo) ListExpiring(ctx context.Context, withinDays int) ([]*domain.Certification, error) {
	cutoff := time.Now().AddDate(0, 0, withinDays)
	q := `SELECT ` + certCols + ` FROM certifications
		WHERE is_deleted=false AND status='ACTIVE'
		  AND expiry_date IS NOT NULL AND expiry_date <= $1
		ORDER BY expiry_date ASC`
	rows, err := r.pool.Query(ctx, q, cutoff)
	if err != nil {
		return nil, fmt.Errorf("CertRepo.ListExpiring: %w", err)
	}
	defer rows.Close()
	var out []*domain.Certification
	for rows.Next() {
		c, err := scanCert(rows)
		if err != nil {
			return nil, fmt.Errorf("CertRepo.ListExpiring scan: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ─── TrainingCourseRepo ───────────────────────────────────────────────────────

type TrainingCourseRepo struct{ pool *pgxpool.Pool }

func NewTrainingCourseRepo(pool *pgxpool.Pool) *TrainingCourseRepo {
	return &TrainingCourseRepo{pool: pool}
}

const courseCols = `
	id, code, name, provider, description,
	cpe_hours, course_type, is_active, notes, created_by, created_at, updated_at`

func scanCourse(row scanner) (*domain.TrainingCourse, error) {
	var c domain.TrainingCourse
	err := row.Scan(
		&c.ID, &c.Code, &c.Name, &c.Provider, &c.Description,
		&c.CpeHours, &c.CourseType, &c.IsActive, &c.Notes, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt,
	)
	return &c, err
}

func (r *TrainingCourseRepo) Create(ctx context.Context, p domain.CreateTrainingCourseParams) (*domain.TrainingCourse, error) {
	const q = `
		INSERT INTO training_courses
			(code, name, provider, description, cpe_hours, course_type, is_active, notes, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING ` + courseCols
	c, err := scanCourse(r.pool.QueryRow(ctx, q,
		p.Code, p.Name, p.Provider, p.Description, p.CpeHours,
		p.CourseType, p.IsActive, p.Notes, p.CreatedBy,
	))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrDuplicateCourseCode
		}
		if e := pgCheckViolation(err); e != nil {
			return nil, e
		}
		return nil, fmt.Errorf("TrainingCourseRepo.Create: %w", err)
	}
	return c, nil
}

func (r *TrainingCourseRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.TrainingCourse, error) {
	q := `SELECT ` + courseCols + ` FROM training_courses WHERE id = $1`
	c, err := scanCourse(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTrainingCourseNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("TrainingCourseRepo.FindByID: %w", err)
	}
	return c, nil
}

func (r *TrainingCourseRepo) List(ctx context.Context, f domain.ListTrainingCoursesFilter) ([]*domain.TrainingCourse, int64, error) {
	where := []string{"1=1"}
	args := []any{}
	i := 1

	if f.CourseType != "" {
		where = append(where, fmt.Sprintf("course_type=$%d", i))
		args = append(args, f.CourseType)
		i++
	}
	if f.IsActive != nil {
		where = append(where, fmt.Sprintf("is_active=$%d", i))
		args = append(args, *f.IsActive)
		i++
	}
	if f.Q != "" {
		where = append(where, fmt.Sprintf("(name ILIKE $%d OR code ILIKE $%d OR provider ILIKE $%d)", i, i, i))
		args = append(args, "%"+f.Q+"%")
		i++
	}

	pred := strings.Join(where, " AND ")

	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM training_courses WHERE `+pred, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("TrainingCourseRepo.List count: %w", err)
	}

	if f.Page < 1 {
		f.Page = 1
	}
	if f.Size < 1 {
		f.Size = 20
	}
	offset := (f.Page - 1) * f.Size
	query := fmt.Sprintf(`SELECT `+courseCols+` FROM training_courses WHERE %s ORDER BY name ASC LIMIT $%d OFFSET $%d`, pred, i, i+1)
	args = append(args, f.Size, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("TrainingCourseRepo.List: %w", err)
	}
	defer rows.Close()

	var out []*domain.TrainingCourse
	for rows.Next() {
		c, err := scanCourse(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("TrainingCourseRepo.List scan: %w", err)
		}
		out = append(out, c)
	}
	return out, total, rows.Err()
}

func (r *TrainingCourseRepo) Update(ctx context.Context, p domain.UpdateTrainingCourseParams) (*domain.TrainingCourse, error) {
	const q = `
		UPDATE training_courses SET
			name=$2, provider=$3, description=$4, cpe_hours=$5,
			course_type=$6, is_active=$7, notes=$8, updated_at=now()
		WHERE id=$1
		RETURNING ` + courseCols
	c, err := scanCourse(r.pool.QueryRow(ctx, q,
		p.ID, p.Name, p.Provider, p.Description, p.CpeHours,
		p.CourseType, p.IsActive, p.Notes,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTrainingCourseNotFound
	}
	if err != nil {
		if e := pgCheckViolation(err); e != nil {
			return nil, e
		}
		return nil, fmt.Errorf("TrainingCourseRepo.Update: %w", err)
	}
	return c, nil
}

func (r *TrainingCourseRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM training_courses WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("TrainingCourseRepo.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTrainingCourseNotFound
	}
	return nil
}

// ─── TrainingRecordRepo ───────────────────────────────────────────────────────

type TrainingRecordRepo struct{ pool *pgxpool.Pool }

func NewTrainingRecordRepo(pool *pgxpool.Pool) *TrainingRecordRepo {
	return &TrainingRecordRepo{pool: pool}
}

const recordCols = `
	id, employee_id, course_id, completion_date, cpe_hours_earned,
	certificate_url, status, notes, is_deleted, created_by, created_at, updated_at`

func scanRecord(row scanner) (*domain.TrainingRecord, error) {
	var r domain.TrainingRecord
	err := row.Scan(
		&r.ID, &r.EmployeeID, &r.CourseID, &r.CompletionDate, &r.CpeHoursEarned,
		&r.CertificateURL, &r.Status, &r.Notes, &r.IsDeleted, &r.CreatedBy, &r.CreatedAt, &r.UpdatedAt,
	)
	return &r, err
}

func (r *TrainingRecordRepo) Create(ctx context.Context, p domain.CreateTrainingRecordParams) (*domain.TrainingRecord, error) {
	const q = `
		INSERT INTO training_records
			(employee_id, course_id, completion_date, cpe_hours_earned,
			 certificate_url, status, notes, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING ` + recordCols
	rec, err := scanRecord(r.pool.QueryRow(ctx, q,
		p.EmployeeID, p.CourseID, p.CompletionDate, p.CpeHoursEarned,
		p.CertificateURL, p.Status, p.Notes, p.CreatedBy,
	))
	if err != nil {
		if e := pgCheckViolation(err); e != nil {
			return nil, e
		}
		return nil, fmt.Errorf("TrainingRecordRepo.Create: %w", err)
	}
	return rec, nil
}

func (r *TrainingRecordRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.TrainingRecord, error) {
	q := `SELECT ` + recordCols + ` FROM training_records WHERE id = $1`
	rec, err := scanRecord(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTrainingRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("TrainingRecordRepo.FindByID: %w", err)
	}
	return rec, nil
}

func (r *TrainingRecordRepo) ListByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*domain.TrainingRecord, error) {
	q := `SELECT ` + recordCols + ` FROM training_records
		WHERE employee_id=$1 AND is_deleted=false
		ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, employeeID)
	if err != nil {
		return nil, fmt.Errorf("TrainingRecordRepo.ListByEmployee: %w", err)
	}
	defer rows.Close()
	var out []*domain.TrainingRecord
	for rows.Next() {
		rec, err := scanRecord(rows)
		if err != nil {
			return nil, fmt.Errorf("TrainingRecordRepo.ListByEmployee scan: %w", err)
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (r *TrainingRecordRepo) Update(ctx context.Context, p domain.UpdateTrainingRecordParams) (*domain.TrainingRecord, error) {
	const q = `
		UPDATE training_records SET
			completion_date=$2, cpe_hours_earned=$3, certificate_url=$4,
			status=$5, notes=$6, updated_at=now()
		WHERE id=$1 AND is_deleted=false
		RETURNING ` + recordCols
	rec, err := scanRecord(r.pool.QueryRow(ctx, q,
		p.ID, p.CompletionDate, p.CpeHoursEarned, p.CertificateURL, p.Status, p.Notes,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTrainingRecordNotFound
	}
	if err != nil {
		if e := pgCheckViolation(err); e != nil {
			return nil, e
		}
		return nil, fmt.Errorf("TrainingRecordRepo.Update: %w", err)
	}
	return rec, nil
}

func (r *TrainingRecordRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE training_records SET is_deleted=true, updated_at=now() WHERE id=$1 AND is_deleted=false`, id)
	if err != nil {
		return fmt.Errorf("TrainingRecordRepo.SoftDelete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTrainingRecordNotFound
	}
	return nil
}

// GetCPESummary returns total CPE hours earned by an employee in a given year,
// grouped by course_type. The required_hours come from cpe_requirements_by_role
// joined on the employee's roles via users → user_roles → roles.
func (r *TrainingRecordRepo) GetCPESummary(ctx context.Context, employeeID uuid.UUID, year int) (*domain.CPESummary, error) {
	const q = `
		WITH completed AS (
			SELECT tr.cpe_hours_earned, tc.course_type
			FROM   training_records tr
			JOIN   training_courses tc ON tc.id = tr.course_id
			WHERE  tr.employee_id = $1
			  AND  tr.status      = 'COMPLETED'
			  AND  tr.is_deleted  = false
			  AND  EXTRACT(YEAR FROM tr.completion_date) = $2
		),
		by_category AS (
			SELECT   course_type, SUM(cpe_hours_earned) AS hours
			FROM     completed
			GROUP BY course_type
		),
		total AS (SELECT COALESCE(SUM(cpe_hours_earned),0) AS hours FROM completed),
		req AS (
			SELECT COALESCE(MAX(cpr.required_hours),0) AS hours
			FROM   employees      e
			JOIN   users          u   ON u.id = e.user_id
			JOIN   user_roles     ur  ON ur.user_id = u.id
			JOIN   roles          ro  ON ro.id = ur.role_id
			JOIN   cpe_requirements_by_role cpr ON cpr.role_code = ro.code AND cpr.year = $2
			WHERE  e.id = $1
		)
		SELECT
			total.hours,
			req.hours,
			COALESCE(
				(SELECT jsonb_object_agg(course_type, hours) FROM by_category),
				'{}'::jsonb
			)
		FROM total, req`

	var s domain.CPESummary
	s.EmployeeID = employeeID
	s.Year = year

	err := r.pool.QueryRow(ctx, q, employeeID, year).Scan(
		&s.TotalHours, &s.RequiredHours, &s.ByCategory,
	)
	if err != nil {
		return nil, fmt.Errorf("TrainingRecordRepo.GetCPESummary: %w", err)
	}
	return &s, nil
}

// ─── CPERequirementRepo ───────────────────────────────────────────────────────

type CPERequirementRepo struct{ pool *pgxpool.Pool }

func NewCPERequirementRepo(pool *pgxpool.Pool) *CPERequirementRepo {
	return &CPERequirementRepo{pool: pool}
}

const cpeCols = `
	id, role_code, year, required_hours, category_breakdown, notes, created_by, created_at, updated_at`

func scanCPE(row scanner) (*domain.CPERequirement, error) {
	var c domain.CPERequirement
	err := row.Scan(
		&c.ID, &c.RoleCode, &c.Year, &c.RequiredHours, &c.CategoryBreakdown,
		&c.Notes, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt,
	)
	return &c, err
}

func (r *CPERequirementRepo) Create(ctx context.Context, p domain.CreateCPERequirementParams) (*domain.CPERequirement, error) {
	const q = `
		INSERT INTO cpe_requirements_by_role
			(role_code, year, required_hours, category_breakdown, notes, created_by)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING ` + cpeCols
	c, err := scanCPE(r.pool.QueryRow(ctx, q,
		p.RoleCode, p.Year, p.RequiredHours, p.CategoryBreakdown, p.Notes, p.CreatedBy,
	))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrDuplicateCPERequirement
		}
		if e := pgCheckViolation(err); e != nil {
			return nil, e
		}
		return nil, fmt.Errorf("CPERequirementRepo.Create: %w", err)
	}
	return c, nil
}

func (r *CPERequirementRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.CPERequirement, error) {
	q := `SELECT ` + cpeCols + ` FROM cpe_requirements_by_role WHERE id = $1`
	c, err := scanCPE(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrCPERequirementNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("CPERequirementRepo.FindByID: %w", err)
	}
	return c, nil
}

func (r *CPERequirementRepo) List(ctx context.Context, roleCode string, year int) ([]*domain.CPERequirement, error) {
	where := []string{"1=1"}
	args := []any{}
	i := 1
	if roleCode != "" {
		where = append(where, fmt.Sprintf("role_code=$%d", i))
		args = append(args, roleCode)
		i++
	}
	if year > 0 {
		where = append(where, fmt.Sprintf("year=$%d", i))
		args = append(args, int16(year))
		i++
	}
	_ = i
	q := `SELECT ` + cpeCols + ` FROM cpe_requirements_by_role WHERE ` + strings.Join(where, " AND ") + ` ORDER BY role_code, year`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("CPERequirementRepo.List: %w", err)
	}
	defer rows.Close()
	var out []*domain.CPERequirement
	for rows.Next() {
		c, err := scanCPE(rows)
		if err != nil {
			return nil, fmt.Errorf("CPERequirementRepo.List scan: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CPERequirementRepo) Update(ctx context.Context, p domain.UpdateCPERequirementParams) (*domain.CPERequirement, error) {
	const q = `
		UPDATE cpe_requirements_by_role SET
			required_hours=$2, category_breakdown=$3, notes=$4, updated_at=now()
		WHERE id=$1
		RETURNING ` + cpeCols
	c, err := scanCPE(r.pool.QueryRow(ctx, q,
		p.ID, p.RequiredHours, p.CategoryBreakdown, p.Notes,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrCPERequirementNotFound
	}
	if err != nil {
		if e := pgCheckViolation(err); e != nil {
			return nil, e
		}
		return nil, fmt.Errorf("CPERequirementRepo.Update: %w", err)
	}
	return c, nil
}
