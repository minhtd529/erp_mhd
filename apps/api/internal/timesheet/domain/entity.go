// Package domain defines the Timesheet bounded context aggregates.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// TimesheetStatus represents the approval lifecycle of a weekly timesheet.
type TimesheetStatus string

const (
	TSStatusDraft     TimesheetStatus = "DRAFT"
	TSStatusSubmitted TimesheetStatus = "SUBMITTED"
	TSStatusApproved  TimesheetStatus = "APPROVED"
	TSStatusRejected  TimesheetStatus = "REJECTED"
	TSStatusLocked    TimesheetStatus = "LOCKED"
)

// AttendanceLocation enumerates where work was performed.
type AttendanceLocation string

const (
	LocationOnSite AttendanceLocation = "ON_SITE"
	LocationRemote AttendanceLocation = "REMOTE"
)

// AttendanceStatus enumerates attendance types.
type AttendanceStatus string

const (
	AttendancePresent AttendanceStatus = "PRESENT"
	AttendanceLeave   AttendanceStatus = "LEAVE"
	AttendanceHoliday AttendanceStatus = "HOLIDAY"
	AttendanceAbsent  AttendanceStatus = "ABSENT"
)

// Timesheet is the aggregate root: one record per staff member per week.
type Timesheet struct {
	ID              uuid.UUID       `json:"id"                db:"id"`
	StaffID         uuid.UUID       `json:"staff_id"          db:"staff_id"`
	PeriodStartDate time.Time       `json:"period_start_date" db:"period_start_date"`
	Status          TimesheetStatus `json:"status"            db:"status"`
	TotalHours      float64         `json:"total_hours"       db:"total_hours"`
	SubmittedAt     *time.Time      `json:"submitted_at"      db:"submitted_at"`
	SubmittedBy     *uuid.UUID      `json:"submitted_by"      db:"submitted_by"`
	ApprovedAt      *time.Time      `json:"approved_at"       db:"approved_at"`
	ApprovedBy      *uuid.UUID      `json:"approved_by"       db:"approved_by"`
	RejectReason    *string         `json:"reject_reason"     db:"reject_reason"`
	LockedAt        *time.Time      `json:"locked_at"         db:"locked_at"`
	IsDeleted       bool            `json:"is_deleted"        db:"is_deleted"`
	CreatedAt       time.Time       `json:"created_at"        db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"        db:"updated_at"`
	CreatedBy       uuid.UUID       `json:"created_by"        db:"created_by"`
	UpdatedBy       uuid.UUID       `json:"updated_by"        db:"updated_by"`
}

// TimesheetEntry is a single daily time record within a weekly timesheet.
type TimesheetEntry struct {
	ID           uuid.UUID  `json:"id"            db:"id"`
	TimesheetID  uuid.UUID  `json:"timesheet_id"  db:"timesheet_id"`
	EntryDate    time.Time  `json:"entry_date"    db:"entry_date"`
	EngagementID uuid.UUID  `json:"engagement_id" db:"engagement_id"`
	TaskID       *uuid.UUID `json:"task_id"       db:"task_id"`
	HoursWorked  float64    `json:"hours_worked"  db:"hours_worked"`
	Description  *string    `json:"description"   db:"description"`
	IsDeleted    bool       `json:"is_deleted"    db:"is_deleted"`
	CreatedAt    time.Time  `json:"created_at"    db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"    db:"updated_at"`
	CreatedBy    uuid.UUID  `json:"created_by"    db:"created_by"`
	UpdatedBy    uuid.UUID  `json:"updated_by"    db:"updated_by"`
}

// Attendance records a staff member's check-in/out for a day.
type Attendance struct {
	ID           uuid.UUID          `json:"id"             db:"id"`
	StaffID      uuid.UUID          `json:"staff_id"       db:"staff_id"`
	CheckInTime  time.Time          `json:"check_in_time"  db:"check_in_time"`
	CheckOutTime *time.Time         `json:"check_out_time" db:"check_out_time"`
	Location     AttendanceLocation `json:"location"       db:"location"`
	Status       AttendanceStatus   `json:"status"         db:"status"`
	Notes        *string            `json:"notes"          db:"notes"`
	CreatedAt    time.Time          `json:"created_at"     db:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"     db:"updated_at"`
}

// ─── Command params ───────────────────────────────────────────────────────────

type GetOrCreateTimesheetParams struct {
	StaffID         uuid.UUID
	PeriodStartDate time.Time // must be a Monday
	CreatedBy       uuid.UUID
}

type ListTimesheetsFilter struct {
	Page    int
	Size    int
	StaffID *uuid.UUID
	Status  TimesheetStatus
}

// TimesheetCursorFilter is for cursor-based list queries.
type TimesheetCursorFilter struct {
	AfterID        *uuid.UUID
	AfterCreatedAt *time.Time
	Size           int
	StaffID        *uuid.UUID
	Status         TimesheetStatus
}

type CreateEntryParams struct {
	TimesheetID  uuid.UUID
	EntryDate    time.Time
	EngagementID uuid.UUID
	TaskID       *uuid.UUID
	HoursWorked  float64
	Description  *string
	CreatedBy    uuid.UUID
}

type UpdateEntryParams struct {
	ID           uuid.UUID
	TimesheetID  uuid.UUID
	EngagementID uuid.UUID
	TaskID       *uuid.UUID
	HoursWorked  float64
	Description  *string
	UpdatedBy    uuid.UUID
}

type CheckInParams struct {
	StaffID  uuid.UUID
	Location AttendanceLocation
	Notes    *string
}
