package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/timesheet/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

// AttendanceUseCase manages check-in/check-out records.
type AttendanceUseCase struct {
	repo     domain.AttendanceRepository
	auditLog *audit.Logger
}

// NewAttendanceUseCase constructs an AttendanceUseCase.
func NewAttendanceUseCase(repo domain.AttendanceRepository, auditLog *audit.Logger) *AttendanceUseCase {
	return &AttendanceUseCase{repo: repo, auditLog: auditLog}
}

// CheckIn records a check-in, rejecting if staff already has an open record.
func (uc *AttendanceUseCase) CheckIn(ctx context.Context, req CheckInRequest, callerID uuid.UUID, ip string) (*AttendanceResponse, error) {
	_, err := uc.repo.FindOpenByStaff(ctx, callerID)
	if err == nil {
		return nil, domain.ErrAlreadyCheckedIn
	}

	loc := req.Location
	if loc == "" {
		loc = domain.LocationOnSite
	}

	a, err := uc.repo.CheckIn(ctx, domain.CheckInParams{
		StaffID:  callerID,
		Location: loc,
		Notes:    req.Notes,
	})
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "timesheet", Resource: "attendance",
		ResourceID: &a.ID, Action: "CREATE", IPAddress: ip,
	})

	resp := toAttendanceResponse(a)
	return &resp, nil
}

// CheckOut closes the open attendance record for the caller.
func (uc *AttendanceUseCase) CheckOut(ctx context.Context, callerID uuid.UUID, ip string) (*AttendanceResponse, error) {
	a, err := uc.repo.CheckOut(ctx, callerID)
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "timesheet", Resource: "attendance",
		ResourceID: &a.ID, Action: "UPDATE", IPAddress: ip,
	})

	resp := toAttendanceResponse(a)
	return &resp, nil
}

// MyRecords returns recent attendance records for the caller.
func (uc *AttendanceUseCase) MyRecords(ctx context.Context, callerID uuid.UUID) ([]AttendanceResponse, error) {
	records, err := uc.repo.ListByStaff(ctx, callerID, 50)
	if err != nil {
		return nil, err
	}
	out := make([]AttendanceResponse, len(records))
	for i, a := range records {
		out[i] = toAttendanceResponse(a)
	}
	return out, nil
}
