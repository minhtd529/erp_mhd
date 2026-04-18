// Package repository provides the PostgreSQL implementation of the Timesheet
// domain repository interfaces.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/timesheet/domain"
)

// ─── TimesheetRepo ───────────────────────────────────────────────────────────

// TimesheetRepo implements domain.TimesheetRepository.
type TimesheetRepo struct{ pool *pgxpool.Pool }

// NewTimesheetRepo creates a TimesheetRepo.
func NewTimesheetRepo(pool *pgxpool.Pool) *TimesheetRepo { return &TimesheetRepo{pool: pool} }

func (r *TimesheetRepo) GetOrCreate(ctx context.Context, p domain.GetOrCreateTimesheetParams) (*domain.Timesheet, error) {
	const q = `
		INSERT INTO timesheets (staff_id, period_start_date, created_by, updated_by)
		VALUES ($1, $2, $3, $3)
		ON CONFLICT (staff_id, period_start_date) WHERE is_deleted = false
		DO UPDATE SET updated_at = NOW()
		RETURNING ` + tsCols

	ts, err := scanTimesheet(r.pool.QueryRow(ctx, q, p.StaffID, p.PeriodStartDate, p.CreatedBy))
	if err != nil {
		return nil, fmt.Errorf("timesheet.GetOrCreate: %w", err)
	}
	return ts, nil
}

func (r *TimesheetRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Timesheet, error) {
	q := "SELECT " + tsCols + " FROM timesheets WHERE id = $1 AND is_deleted = false"
	ts, err := scanTimesheet(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTimesheetNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("timesheet.FindByID: %w", err)
	}
	return ts, nil
}

func (r *TimesheetRepo) FindByStaffAndWeek(ctx context.Context, staffID uuid.UUID, weekStart time.Time) (*domain.Timesheet, error) {
	q := "SELECT " + tsCols + " FROM timesheets WHERE staff_id = $1 AND period_start_date = $2 AND is_deleted = false"
	ts, err := scanTimesheet(r.pool.QueryRow(ctx, q, staffID, weekStart))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTimesheetNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("timesheet.FindByStaffAndWeek: %w", err)
	}
	return ts, nil
}

func (r *TimesheetRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.TimesheetStatus, actorID uuid.UUID, rejectReason *string) (*domain.Timesheet, error) {
	now := time.Now()
	var q string
	var args []any

	switch status {
	case domain.TSStatusSubmitted:
		q = `UPDATE timesheets SET status=$2, submitted_at=$3, submitted_by=$4, updated_by=$4, updated_at=NOW() WHERE id=$1 AND is_deleted=false RETURNING ` + tsCols
		args = []any{id, string(status), now, actorID}
	case domain.TSStatusApproved:
		q = `UPDATE timesheets SET status=$2, approved_at=$3, approved_by=$4, updated_by=$4, updated_at=NOW() WHERE id=$1 AND is_deleted=false RETURNING ` + tsCols
		args = []any{id, string(status), now, actorID}
	case domain.TSStatusRejected:
		q = `UPDATE timesheets SET status=$2, reject_reason=$3, updated_by=$4, updated_at=NOW() WHERE id=$1 AND is_deleted=false RETURNING ` + tsCols
		args = []any{id, string(status), rejectReason, actorID}
	case domain.TSStatusLocked:
		q = `UPDATE timesheets SET status=$2, locked_at=$3, updated_by=$4, updated_at=NOW() WHERE id=$1 AND is_deleted=false RETURNING ` + tsCols
		args = []any{id, string(status), now, actorID}
	default:
		q = `UPDATE timesheets SET status=$2, updated_by=$3, updated_at=NOW() WHERE id=$1 AND is_deleted=false RETURNING ` + tsCols
		args = []any{id, string(status), actorID}
	}

	ts, err := scanTimesheet(r.pool.QueryRow(ctx, q, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTimesheetNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("timesheet.UpdateStatus: %w", err)
	}
	return ts, nil
}

func (r *TimesheetRepo) UpdateTotalHours(ctx context.Context, id uuid.UUID) error {
	const q = `
		UPDATE timesheets t
		SET total_hours = (
			SELECT COALESCE(SUM(hours_worked),0) FROM timesheet_entries WHERE timesheet_id = t.id AND is_deleted = false
		), updated_at = NOW()
		WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id)
	return err
}

func (r *TimesheetRepo) List(ctx context.Context, f domain.ListTimesheetsFilter) ([]*domain.Timesheet, int64, error) {
	offset := (f.Page - 1) * f.Size
	args := []any{}
	where := "WHERE is_deleted = false"
	idx := 1

	if f.StaffID != nil {
		where += fmt.Sprintf(" AND staff_id = $%d", idx)
		args = append(args, *f.StaffID)
		idx++
	}
	if f.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, string(f.Status))
		idx++
	}

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM timesheets "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("timesheet.List count: %w", err)
	}

	args = append(args, f.Size, offset)
	dataQ := fmt.Sprintf("SELECT %s FROM timesheets %s ORDER BY period_start_date DESC LIMIT $%d OFFSET $%d",
		tsCols, where, idx, idx+1)

	rows, err := r.pool.Query(ctx, dataQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("timesheet.List: %w", err)
	}
	defer rows.Close()

	var list []*domain.Timesheet
	for rows.Next() {
		ts, err := scanTimesheet(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("timesheet.List scan: %w", err)
		}
		list = append(list, ts)
	}
	if list == nil {
		list = []*domain.Timesheet{}
	}
	return list, total, rows.Err()
}

func (r *TimesheetRepo) ListCursor(ctx context.Context, f domain.TimesheetCursorFilter) ([]*domain.Timesheet, error) {
	args := []any{}
	where := "WHERE is_deleted = false"
	idx := 1

	if f.AfterCreatedAt != nil && f.AfterID != nil {
		where += fmt.Sprintf(" AND (created_at, id) < ($%d, $%d)", idx, idx+1)
		args = append(args, *f.AfterCreatedAt, *f.AfterID)
		idx += 2
	}
	if f.StaffID != nil {
		where += fmt.Sprintf(" AND staff_id = $%d", idx)
		args = append(args, *f.StaffID)
		idx++
	}
	if f.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, string(f.Status))
		idx++
	}

	limit := f.Size + 1
	args = append(args, limit)
	q := fmt.Sprintf("SELECT %s FROM timesheets %s ORDER BY created_at DESC, id DESC LIMIT $%d",
		tsCols, where, idx)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("timesheet.ListCursor: %w", err)
	}
	defer rows.Close()

	var list []*domain.Timesheet
	for rows.Next() {
		ts, err := scanTimesheet(rows)
		if err != nil {
			return nil, fmt.Errorf("timesheet.ListCursor scan: %w", err)
		}
		list = append(list, ts)
	}
	if list == nil {
		list = []*domain.Timesheet{}
	}
	return list, rows.Err()
}

// ─── EntryRepo ───────────────────────────────────────────────────────────────

// EntryRepo implements domain.EntryRepository.
type EntryRepo struct{ pool *pgxpool.Pool }

// NewEntryRepo creates an EntryRepo.
func NewEntryRepo(pool *pgxpool.Pool) *EntryRepo { return &EntryRepo{pool: pool} }

func (r *EntryRepo) Create(ctx context.Context, p domain.CreateEntryParams) (*domain.TimesheetEntry, error) {
	const q = `
		INSERT INTO timesheet_entries (timesheet_id, entry_date, engagement_id, task_id, hours_worked, description, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		RETURNING ` + entryCols

	e, err := scanEntry(r.pool.QueryRow(ctx, q,
		p.TimesheetID, p.EntryDate, p.EngagementID, p.TaskID, p.HoursWorked, p.Description, p.CreatedBy,
	))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, fmt.Errorf("timesheet.entry.Create foreign key violation: %w", err)
		}
		return nil, fmt.Errorf("timesheet.entry.Create: %w", err)
	}
	return e, nil
}

func (r *EntryRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.TimesheetEntry, error) {
	q := "SELECT " + entryCols + " FROM timesheet_entries WHERE id = $1 AND is_deleted = false"
	e, err := scanEntry(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrEntryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("timesheet.entry.FindByID: %w", err)
	}
	return e, nil
}

func (r *EntryRepo) Update(ctx context.Context, p domain.UpdateEntryParams) (*domain.TimesheetEntry, error) {
	const q = `
		UPDATE timesheet_entries
		SET engagement_id=$3, task_id=$4, hours_worked=$5, description=$6, updated_by=$7, updated_at=NOW()
		WHERE id=$1 AND timesheet_id=$2 AND is_deleted=false
		RETURNING ` + entryCols

	e, err := scanEntry(r.pool.QueryRow(ctx, q,
		p.ID, p.TimesheetID, p.EngagementID, p.TaskID, p.HoursWorked, p.Description, p.UpdatedBy,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrEntryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("timesheet.entry.Update: %w", err)
	}
	return e, nil
}

func (r *EntryRepo) SoftDelete(ctx context.Context, id uuid.UUID, timesheetID uuid.UUID, deletedBy uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE timesheet_entries SET is_deleted=true, updated_by=$3, updated_at=NOW() WHERE id=$1 AND timesheet_id=$2 AND is_deleted=false`,
		id, timesheetID, deletedBy)
	if err != nil {
		return fmt.Errorf("timesheet.entry.SoftDelete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrEntryNotFound
	}
	return nil
}

func (r *EntryRepo) ListByTimesheet(ctx context.Context, timesheetID uuid.UUID) ([]*domain.TimesheetEntry, error) {
	q := "SELECT " + entryCols + " FROM timesheet_entries WHERE timesheet_id=$1 AND is_deleted=false ORDER BY entry_date, created_at"
	rows, err := r.pool.Query(ctx, q, timesheetID)
	if err != nil {
		return nil, fmt.Errorf("timesheet.entry.List: %w", err)
	}
	defer rows.Close()

	var list []*domain.TimesheetEntry
	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return nil, fmt.Errorf("timesheet.entry.List scan: %w", err)
		}
		list = append(list, e)
	}
	if list == nil {
		list = []*domain.TimesheetEntry{}
	}
	return list, rows.Err()
}

func (r *EntryRepo) ListLockedByEngagement(ctx context.Context, engagementID uuid.UUID, start, end time.Time) ([]*domain.TimesheetEntry, error) {
	q := `
		SELECT ` + entryCols + `
		FROM timesheet_entries
		WHERE engagement_id = $1
		  AND entry_date BETWEEN $2 AND $3
		  AND is_deleted = false
		  AND timesheet_id IN (
		      SELECT id FROM timesheets WHERE status = 'LOCKED' AND is_deleted = false
		  )
		ORDER BY entry_date, created_at`

	rows, err := r.pool.Query(ctx, q, engagementID, start, end)
	if err != nil {
		return nil, fmt.Errorf("timesheet.entry.ListLockedByEngagement: %w", err)
	}
	defer rows.Close()

	var list []*domain.TimesheetEntry
	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return nil, fmt.Errorf("timesheet.entry.ListLockedByEngagement scan: %w", err)
		}
		list = append(list, e)
	}
	if list == nil {
		list = []*domain.TimesheetEntry{}
	}
	return list, rows.Err()
}

// ─── AttendanceRepo ──────────────────────────────────────────────────────────

// AttendanceRepo implements domain.AttendanceRepository.
type AttendanceRepo struct{ pool *pgxpool.Pool }

// NewAttendanceRepo creates an AttendanceRepo.
func NewAttendanceRepo(pool *pgxpool.Pool) *AttendanceRepo { return &AttendanceRepo{pool: pool} }

func (r *AttendanceRepo) CheckIn(ctx context.Context, p domain.CheckInParams) (*domain.Attendance, error) {
	const q = `
		INSERT INTO attendance (staff_id, check_in_time, location, status, notes)
		VALUES ($1, NOW(), $2, 'PRESENT', $3)
		RETURNING ` + attendanceCols

	a, err := scanAttendance(r.pool.QueryRow(ctx, q, p.StaffID, string(p.Location), p.Notes))
	if err != nil {
		return nil, fmt.Errorf("attendance.CheckIn: %w", err)
	}
	return a, nil
}

func (r *AttendanceRepo) CheckOut(ctx context.Context, staffID uuid.UUID) (*domain.Attendance, error) {
	const q = `
		UPDATE attendance SET check_out_time=NOW(), updated_at=NOW()
		WHERE staff_id=$1 AND check_out_time IS NULL
		RETURNING ` + attendanceCols

	a, err := scanAttendance(r.pool.QueryRow(ctx, q, staffID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotCheckedIn
	}
	if err != nil {
		return nil, fmt.Errorf("attendance.CheckOut: %w", err)
	}
	return a, nil
}

func (r *AttendanceRepo) FindOpenByStaff(ctx context.Context, staffID uuid.UUID) (*domain.Attendance, error) {
	q := "SELECT " + attendanceCols + " FROM attendance WHERE staff_id=$1 AND check_out_time IS NULL ORDER BY check_in_time DESC LIMIT 1"
	a, err := scanAttendance(r.pool.QueryRow(ctx, q, staffID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrAttendanceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("attendance.FindOpen: %w", err)
	}
	return a, nil
}

func (r *AttendanceRepo) ListByStaff(ctx context.Context, staffID uuid.UUID, limit int) ([]*domain.Attendance, error) {
	q := "SELECT " + attendanceCols + " FROM attendance WHERE staff_id=$1 ORDER BY check_in_time DESC LIMIT $2"
	rows, err := r.pool.Query(ctx, q, staffID, limit)
	if err != nil {
		return nil, fmt.Errorf("attendance.List: %w", err)
	}
	defer rows.Close()

	var list []*domain.Attendance
	for rows.Next() {
		a, err := scanAttendance(rows)
		if err != nil {
			return nil, fmt.Errorf("attendance.List scan: %w", err)
		}
		list = append(list, a)
	}
	if list == nil {
		list = []*domain.Attendance{}
	}
	return list, rows.Err()
}

// ─── Column lists & scanners ─────────────────────────────────────────────────

const tsCols = `id, staff_id, period_start_date, status, total_hours,
    submitted_at, submitted_by, approved_at, approved_by, reject_reason, locked_at,
    is_deleted, created_at, updated_at, created_by, updated_by`

const entryCols = `id, timesheet_id, entry_date, engagement_id, task_id,
    hours_worked, description, is_deleted, created_at, updated_at, created_by, updated_by`

const attendanceCols = `id, staff_id, check_in_time, check_out_time, location, status, notes, created_at, updated_at`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTimesheet(row rowScanner) (*domain.Timesheet, error) {
	var ts domain.Timesheet
	var status string
	err := row.Scan(
		&ts.ID, &ts.StaffID, &ts.PeriodStartDate, &status, &ts.TotalHours,
		&ts.SubmittedAt, &ts.SubmittedBy, &ts.ApprovedAt, &ts.ApprovedBy, &ts.RejectReason, &ts.LockedAt,
		&ts.IsDeleted, &ts.CreatedAt, &ts.UpdatedAt, &ts.CreatedBy, &ts.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	ts.Status = domain.TimesheetStatus(status)
	return &ts, nil
}

func scanEntry(row rowScanner) (*domain.TimesheetEntry, error) {
	var e domain.TimesheetEntry
	err := row.Scan(
		&e.ID, &e.TimesheetID, &e.EntryDate, &e.EngagementID, &e.TaskID,
		&e.HoursWorked, &e.Description, &e.IsDeleted, &e.CreatedAt, &e.UpdatedAt, &e.CreatedBy, &e.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func scanAttendance(row rowScanner) (*domain.Attendance, error) {
	var a domain.Attendance
	var loc, status string
	err := row.Scan(
		&a.ID, &a.StaffID, &a.CheckInTime, &a.CheckOutTime, &loc, &status, &a.Notes,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	a.Location = domain.AttendanceLocation(loc)
	a.Status = domain.AttendanceStatus(status)
	return &a, nil
}
