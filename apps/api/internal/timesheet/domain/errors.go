package domain

import "errors"

var (
	ErrTimesheetNotFound      = errors.New("TIMESHEET_NOT_FOUND")
	ErrTimesheetLocked        = errors.New("TIMESHEET_LOCKED")
	ErrInvalidStateTransition = errors.New("INVALID_STATE_TRANSITION")
	ErrEntryNotFound          = errors.New("ENTRY_NOT_FOUND")
	ErrEntryDateOutsidePeriod = errors.New("ENTRY_DATE_OUTSIDE_PERIOD")
	ErrTimesheetNotEditable   = errors.New("TIMESHEET_NOT_EDITABLE")
	ErrAttendanceNotFound     = errors.New("ATTENDANCE_NOT_FOUND")
	ErrAlreadyCheckedIn       = errors.New("ALREADY_CHECKED_IN")
	ErrNotCheckedIn           = errors.New("NOT_CHECKED_IN")
	ErrLockNotAcquired        = errors.New("LOCK_NOT_ACQUIRED")
)
