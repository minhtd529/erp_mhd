package usecase

import (
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/timesheet/domain"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

// ─── Timesheet DTOs ──────────────────────────────────────────────────────────

type TimesheetGetRequest struct {
	WeekDate string `uri:"week_id" binding:"required"` // YYYY-MM-DD (Monday)
}

type TimesheetListRequest struct {
	Page    int                    `form:"page,default=1"  binding:"min=1"`
	Size    int                    `form:"size,default=20" binding:"min=1,max=100"`
	StaffID *uuid.UUID             `form:"staff_id"`
	Status  domain.TimesheetStatus `form:"status"`
}

type RejectRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type TimesheetResponse struct {
	ID              uuid.UUID              `json:"id"`
	StaffID         uuid.UUID              `json:"staff_id"`
	PeriodStartDate time.Time              `json:"period_start_date"`
	Status          domain.TimesheetStatus `json:"status"`
	TotalHours      float64                `json:"total_hours"`
	SubmittedAt     *time.Time             `json:"submitted_at"`
	SubmittedBy     *uuid.UUID             `json:"submitted_by"`
	ApprovedAt      *time.Time             `json:"approved_at"`
	ApprovedBy      *uuid.UUID             `json:"approved_by"`
	RejectReason    *string                `json:"reject_reason"`
	LockedAt        *time.Time             `json:"locked_at"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// ─── Entry DTOs ──────────────────────────────────────────────────────────────

type EntryCreateRequest struct {
	EntryDate    string     `json:"entry_date"    binding:"required"` // YYYY-MM-DD
	EngagementID uuid.UUID  `json:"engagement_id" binding:"required"`
	TaskID       *uuid.UUID `json:"task_id"`
	HoursWorked  float64    `json:"hours_worked"  binding:"required,gt=0,max=24"`
	Description  *string    `json:"description"`
}

type EntryUpdateRequest struct {
	EngagementID uuid.UUID  `json:"engagement_id" binding:"required"`
	TaskID       *uuid.UUID `json:"task_id"`
	HoursWorked  float64    `json:"hours_worked"  binding:"required,gt=0,max=24"`
	Description  *string    `json:"description"`
}

type EntryResponse struct {
	ID           uuid.UUID  `json:"id"`
	TimesheetID  uuid.UUID  `json:"timesheet_id"`
	EntryDate    time.Time  `json:"entry_date"`
	EngagementID uuid.UUID  `json:"engagement_id"`
	TaskID       *uuid.UUID `json:"task_id"`
	HoursWorked  float64    `json:"hours_worked"`
	Description  *string    `json:"description"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	CreatedBy    uuid.UUID  `json:"created_by"`
	UpdatedBy    uuid.UUID  `json:"updated_by"`
}

// ─── Attendance DTOs ─────────────────────────────────────────────────────────

type CheckInRequest struct {
	Location domain.AttendanceLocation `json:"location"`
	Notes    *string                   `json:"notes"`
}

type AttendanceResponse struct {
	ID           uuid.UUID                  `json:"id"`
	StaffID      uuid.UUID                  `json:"staff_id"`
	CheckInTime  time.Time                  `json:"check_in_time"`
	CheckOutTime *time.Time                 `json:"check_out_time"`
	Location     domain.AttendanceLocation  `json:"location"`
	Status       domain.AttendanceStatus    `json:"status"`
	Notes        *string                    `json:"notes"`
	CreatedAt    time.Time                  `json:"created_at"`
}

// ─── Pagination ──────────────────────────────────────────────────────────────

// PaginatedResult is the shared offset pagination type.
type PaginatedResult[T any] = pagination.OffsetResult[T]

func newPaginatedResult[T any](data []T, total int64, page, size int) PaginatedResult[T] {
	return pagination.NewOffsetResult(data, total, page, size)
}

// TimesheetCursorListRequest is for cursor-paginated GET /timesheets.
type TimesheetCursorListRequest struct {
	Cursor  string                 `form:"cursor"`
	Size    int                    `form:"size,default=20" binding:"min=1,max=100"`
	StaffID *uuid.UUID             `form:"staff_id"`
	Status  domain.TimesheetStatus `form:"status"`
}

// TimesheetCursorResult wraps cursor-paginated timesheet results.
type TimesheetCursorResult = pagination.CursorResult[TimesheetResponse]

// ─── Converters ──────────────────────────────────────────────────────────────

func toTimesheetResponse(ts *domain.Timesheet) TimesheetResponse {
	return TimesheetResponse{
		ID: ts.ID, StaffID: ts.StaffID, PeriodStartDate: ts.PeriodStartDate,
		Status: ts.Status, TotalHours: ts.TotalHours,
		SubmittedAt: ts.SubmittedAt, SubmittedBy: ts.SubmittedBy,
		ApprovedAt: ts.ApprovedAt, ApprovedBy: ts.ApprovedBy,
		RejectReason: ts.RejectReason, LockedAt: ts.LockedAt,
		CreatedAt: ts.CreatedAt, UpdatedAt: ts.UpdatedAt,
	}
}

func toEntryResponse(e *domain.TimesheetEntry) EntryResponse {
	return EntryResponse{
		ID: e.ID, TimesheetID: e.TimesheetID, EntryDate: e.EntryDate,
		EngagementID: e.EngagementID, TaskID: e.TaskID,
		HoursWorked: e.HoursWorked, Description: e.Description,
		CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
		CreatedBy: e.CreatedBy, UpdatedBy: e.UpdatedBy,
	}
}

func toAttendanceResponse(a *domain.Attendance) AttendanceResponse {
	return AttendanceResponse{
		ID: a.ID, StaffID: a.StaffID, CheckInTime: a.CheckInTime,
		CheckOutTime: a.CheckOutTime, Location: a.Location,
		Status: a.Status, Notes: a.Notes, CreatedAt: a.CreatedAt,
	}
}
