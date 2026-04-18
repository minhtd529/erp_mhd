package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/timesheet/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

// EntryUseCase manages time entries within a weekly timesheet.
type EntryUseCase struct {
	entryRepo domain.EntryRepository
	tsRepo    domain.TimesheetRepository
	auditLog  *audit.Logger
}

// NewEntryUseCase constructs an EntryUseCase.
func NewEntryUseCase(entryRepo domain.EntryRepository, tsRepo domain.TimesheetRepository, auditLog *audit.Logger) *EntryUseCase {
	return &EntryUseCase{entryRepo: entryRepo, tsRepo: tsRepo, auditLog: auditLog}
}

// List returns all non-deleted entries for a timesheet.
func (uc *EntryUseCase) List(ctx context.Context, timesheetID uuid.UUID) ([]EntryResponse, error) {
	entries, err := uc.entryRepo.ListByTimesheet(ctx, timesheetID)
	if err != nil {
		return nil, err
	}
	out := make([]EntryResponse, len(entries))
	for i, e := range entries {
		out[i] = toEntryResponse(e)
	}
	return out, nil
}

// Create records a new time entry, validating the timesheet is editable and
// the entry date falls within the week period.
func (uc *EntryUseCase) Create(ctx context.Context, timesheetID uuid.UUID, req EntryCreateRequest, callerID uuid.UUID, ip string) (*EntryResponse, error) {
	ts, err := uc.tsRepo.FindByID(ctx, timesheetID)
	if err != nil {
		return nil, err
	}
	if ts.Status == domain.TSStatusApproved || ts.Status == domain.TSStatusLocked {
		return nil, domain.ErrTimesheetNotEditable
	}

	entryDate, err := time.Parse("2006-01-02", req.EntryDate)
	if err != nil {
		return nil, domain.ErrEntryDateOutsidePeriod
	}
	weekEnd := ts.PeriodStartDate.AddDate(0, 0, 6)
	if entryDate.Before(ts.PeriodStartDate) || entryDate.After(weekEnd) {
		return nil, domain.ErrEntryDateOutsidePeriod
	}

	e, err := uc.entryRepo.Create(ctx, domain.CreateEntryParams{
		TimesheetID:  timesheetID,
		EntryDate:    entryDate,
		EngagementID: req.EngagementID,
		TaskID:       req.TaskID,
		HoursWorked:  req.HoursWorked,
		Description:  req.Description,
		CreatedBy:    callerID,
	})
	if err != nil {
		return nil, err
	}

	_ = uc.tsRepo.UpdateTotalHours(ctx, timesheetID)

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "timesheet", Resource: "timesheet_entries",
		ResourceID: &e.ID, Action: "CREATE", IPAddress: ip,
	})

	resp := toEntryResponse(e)
	return &resp, nil
}

// Update modifies an existing entry if the timesheet is still editable.
func (uc *EntryUseCase) Update(ctx context.Context, timesheetID uuid.UUID, entryID uuid.UUID, req EntryUpdateRequest, callerID uuid.UUID, ip string) (*EntryResponse, error) {
	ts, err := uc.tsRepo.FindByID(ctx, timesheetID)
	if err != nil {
		return nil, err
	}
	if ts.Status == domain.TSStatusApproved || ts.Status == domain.TSStatusLocked {
		return nil, domain.ErrTimesheetNotEditable
	}

	e, err := uc.entryRepo.Update(ctx, domain.UpdateEntryParams{
		ID:           entryID,
		TimesheetID:  timesheetID,
		EngagementID: req.EngagementID,
		TaskID:       req.TaskID,
		HoursWorked:  req.HoursWorked,
		Description:  req.Description,
		UpdatedBy:    callerID,
	})
	if err != nil {
		return nil, err
	}

	_ = uc.tsRepo.UpdateTotalHours(ctx, timesheetID)

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "timesheet", Resource: "timesheet_entries",
		ResourceID: &entryID, Action: "UPDATE", IPAddress: ip,
	})

	resp := toEntryResponse(e)
	return &resp, nil
}

// Delete soft-deletes an entry if the timesheet is still editable.
func (uc *EntryUseCase) Delete(ctx context.Context, timesheetID uuid.UUID, entryID uuid.UUID, callerID uuid.UUID, ip string) error {
	ts, err := uc.tsRepo.FindByID(ctx, timesheetID)
	if err != nil {
		return err
	}
	if ts.Status == domain.TSStatusApproved || ts.Status == domain.TSStatusLocked {
		return domain.ErrTimesheetNotEditable
	}

	if err := uc.entryRepo.SoftDelete(ctx, entryID, timesheetID, callerID); err != nil {
		return err
	}

	_ = uc.tsRepo.UpdateTotalHours(ctx, timesheetID)

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "timesheet", Resource: "timesheet_entries",
		ResourceID: &entryID, Action: "DELETE", IPAddress: ip,
	})
	return nil
}
