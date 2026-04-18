package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TimesheetRepository is the data-access contract for Timesheet.
type TimesheetRepository interface {
	GetOrCreate(ctx context.Context, p GetOrCreateTimesheetParams) (*Timesheet, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Timesheet, error)
	FindByStaffAndWeek(ctx context.Context, staffID uuid.UUID, weekStart time.Time) (*Timesheet, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status TimesheetStatus, actorID uuid.UUID, rejectReason *string) (*Timesheet, error)
	UpdateTotalHours(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, f ListTimesheetsFilter) ([]*Timesheet, int64, error)
	ListCursor(ctx context.Context, f TimesheetCursorFilter) ([]*Timesheet, error)
}

// EntryRepository is the data-access contract for TimesheetEntry.
type EntryRepository interface {
	Create(ctx context.Context, p CreateEntryParams) (*TimesheetEntry, error)
	FindByID(ctx context.Context, id uuid.UUID) (*TimesheetEntry, error)
	Update(ctx context.Context, p UpdateEntryParams) (*TimesheetEntry, error)
	SoftDelete(ctx context.Context, id uuid.UUID, timesheetID uuid.UUID, deletedBy uuid.UUID) error
	ListByTimesheet(ctx context.Context, timesheetID uuid.UUID) ([]*TimesheetEntry, error)
	// ListLockedByEngagement returns entries whose parent timesheet is LOCKED,
	// filtered by engagement and date range. Used for invoice generation.
	ListLockedByEngagement(ctx context.Context, engagementID uuid.UUID, start, end time.Time) ([]*TimesheetEntry, error)
}

// AttendanceRepository is the data-access contract for Attendance.
type AttendanceRepository interface {
	CheckIn(ctx context.Context, p CheckInParams) (*Attendance, error)
	CheckOut(ctx context.Context, staffID uuid.UUID) (*Attendance, error)
	FindOpenByStaff(ctx context.Context, staffID uuid.UUID) (*Attendance, error)
	ListByStaff(ctx context.Context, staffID uuid.UUID, limit int) ([]*Attendance, error)
}
