// Package usecase implements the Timesheet application layer.
package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/timesheet/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/distlock"
	"github.com/mdh/erp-audit/api/pkg/outbox"
	"github.com/mdh/erp-audit/api/pkg/pagination"
	"github.com/mdh/erp-audit/api/pkg/worker"
	"github.com/mdh/erp-audit/api/pkg/ws"
)

// TimesheetUseCase handles weekly timesheet CRUD and approval workflow.
type TimesheetUseCase struct {
	tsRepo    domain.TimesheetRepository
	locker    distlock.Acquirer
	auditLog  *audit.Logger
	publisher *outbox.Publisher
	broadcast ws.Broadcaster
}

// NewTimesheetUseCase constructs a TimesheetUseCase.
func NewTimesheetUseCase(tsRepo domain.TimesheetRepository, locker distlock.Acquirer, auditLog *audit.Logger, publisher *outbox.Publisher, broadcast ws.Broadcaster) *TimesheetUseCase {
	return &TimesheetUseCase{tsRepo: tsRepo, locker: locker, auditLog: auditLog, publisher: publisher, broadcast: broadcast}
}

// GetOrCreate returns (or creates) the timesheet for the given staff+week.
func (uc *TimesheetUseCase) GetOrCreate(ctx context.Context, staffID uuid.UUID, weekStart time.Time, callerID uuid.UUID) (*TimesheetResponse, error) {
	ts, err := uc.tsRepo.GetOrCreate(ctx, domain.GetOrCreateTimesheetParams{
		StaffID:         staffID,
		PeriodStartDate: weekStart,
		CreatedBy:       callerID,
	})
	if err != nil {
		return nil, err
	}
	resp := toTimesheetResponse(ts)
	return &resp, nil
}

// GetByID retrieves a timesheet by primary key.
func (uc *TimesheetUseCase) GetByID(ctx context.Context, id uuid.UUID) (*TimesheetResponse, error) {
	ts, err := uc.tsRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toTimesheetResponse(ts)
	return &resp, nil
}

// List returns a paginated list of timesheets.
func (uc *TimesheetUseCase) List(ctx context.Context, req TimesheetListRequest) (PaginatedResult[TimesheetResponse], error) {
	items, total, err := uc.tsRepo.List(ctx, domain.ListTimesheetsFilter{
		Page:    req.Page,
		Size:    req.Size,
		StaffID: req.StaffID,
		Status:  req.Status,
	})
	if err != nil {
		return PaginatedResult[TimesheetResponse]{}, fmt.Errorf("timesheet.List: %w", err)
	}
	responses := make([]TimesheetResponse, len(items))
	for i, ts := range items {
		responses[i] = toTimesheetResponse(ts)
	}
	return newPaginatedResult(responses, total, req.Page, req.Size), nil
}

// ListCursor returns cursor-paginated timesheets.
func (uc *TimesheetUseCase) ListCursor(ctx context.Context, req TimesheetCursorListRequest) (TimesheetCursorResult, error) {
	f := domain.TimesheetCursorFilter{
		Size:    req.Size,
		StaffID: req.StaffID,
		Status:  req.Status,
	}
	if req.Cursor != "" {
		c, err := pagination.DecodeCursor(req.Cursor)
		if err != nil {
			return TimesheetCursorResult{}, err
		}
		f.AfterID = &c.ID
		f.AfterCreatedAt = &c.CreatedAt
	}

	items, err := uc.tsRepo.ListCursor(ctx, f)
	if err != nil {
		return TimesheetCursorResult{}, fmt.Errorf("timesheet.ListCursor: %w", err)
	}

	responses := make([]TimesheetResponse, len(items))
	for i, ts := range items {
		responses[i] = toTimesheetResponse(ts)
	}
	return pagination.NewCursorResult(responses, req.Size, func(r TimesheetResponse) pagination.Cursor {
		return pagination.Cursor{ID: r.ID, CreatedAt: r.CreatedAt}
	}), nil
}

// Submit transitions a DRAFT or REJECTED timesheet to SUBMITTED.
func (uc *TimesheetUseCase) Submit(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip string) (*TimesheetResponse, error) {
	ts, err := uc.tsRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ts.Status != domain.TSStatusDraft && ts.Status != domain.TSStatusRejected {
		return nil, domain.ErrInvalidStateTransition
	}
	updated, err := uc.tsRepo.UpdateStatus(ctx, id, domain.TSStatusSubmitted, callerID, nil)
	if err != nil {
		return nil, err
	}
	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "timesheet", Resource: "timesheets",
		ResourceID: &id, Action: "STATE_TRANSITION", IPAddress: ip,
	})
	_ = uc.publisher.Publish(ctx, "timesheet", id, outbox.EventTimesheetSubmitted, worker.TimesheetEventPayload{
		TimesheetID: id.String(), StaffID: updated.StaffID.String(),
		Status: string(updated.Status), CallerID: callerID.String(),
	})
	if uc.broadcast != nil {
		_ = uc.broadcast.Broadcast("timesheet", "timesheet.submitted", map[string]string{
			"timesheet_id": id.String(), "staff_id": updated.StaffID.String(), "status": string(updated.Status),
		})
	}
	resp := toTimesheetResponse(updated)
	return &resp, nil
}

// Approve transitions a SUBMITTED timesheet to APPROVED, with distributed lock.
func (uc *TimesheetUseCase) Approve(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip string) (*TimesheetResponse, error) {
	lockKey := fmt.Sprintf("timesheet:%s:approve", id)
	lock, err := uc.locker.Acquire(ctx, lockKey)
	if err != nil {
		return nil, domain.ErrLockNotAcquired
	}
	defer lock.Release(ctx) //nolint:errcheck

	ts, err := uc.tsRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ts.Status != domain.TSStatusSubmitted {
		return nil, domain.ErrInvalidStateTransition
	}
	updated, err := uc.tsRepo.UpdateStatus(ctx, id, domain.TSStatusApproved, callerID, nil)
	if err != nil {
		return nil, err
	}
	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "timesheet", Resource: "timesheets",
		ResourceID: &id, Action: "APPROVE", IPAddress: ip,
	})
	_ = uc.publisher.Publish(ctx, "timesheet", id, outbox.EventTimesheetApproved, worker.TimesheetEventPayload{
		TimesheetID: id.String(), StaffID: updated.StaffID.String(),
		Status: string(updated.Status), CallerID: callerID.String(),
	})
	if uc.broadcast != nil {
		_ = uc.broadcast.Broadcast("timesheet", "timesheet.approved", map[string]string{
			"timesheet_id": id.String(), "staff_id": updated.StaffID.String(), "status": string(updated.Status),
		})
	}
	resp := toTimesheetResponse(updated)
	return &resp, nil
}

// Reject transitions a SUBMITTED timesheet back to REJECTED with a reason.
func (uc *TimesheetUseCase) Reject(ctx context.Context, id uuid.UUID, reason string, callerID uuid.UUID, ip string) (*TimesheetResponse, error) {
	lockKey := fmt.Sprintf("timesheet:%s:approve", id)
	lock, err := uc.locker.Acquire(ctx, lockKey)
	if err != nil {
		return nil, domain.ErrLockNotAcquired
	}
	defer lock.Release(ctx) //nolint:errcheck

	ts, err := uc.tsRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ts.Status != domain.TSStatusSubmitted {
		return nil, domain.ErrInvalidStateTransition
	}
	updated, err := uc.tsRepo.UpdateStatus(ctx, id, domain.TSStatusRejected, callerID, &reason)
	if err != nil {
		return nil, err
	}
	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "timesheet", Resource: "timesheets",
		ResourceID: &id, Action: "REJECT", IPAddress: ip,
	})
	_ = uc.publisher.Publish(ctx, "timesheet", id, outbox.EventTimesheetRejected, worker.TimesheetEventPayload{
		TimesheetID: id.String(), StaffID: updated.StaffID.String(),
		Status: string(updated.Status), CallerID: callerID.String(),
	})
	if uc.broadcast != nil {
		_ = uc.broadcast.Broadcast("timesheet", "timesheet.rejected", map[string]string{
			"timesheet_id": id.String(), "staff_id": updated.StaffID.String(), "status": string(updated.Status),
		})
	}
	resp := toTimesheetResponse(updated)
	return &resp, nil
}

// Lock transitions an APPROVED timesheet to LOCKED (immutable for billing).
func (uc *TimesheetUseCase) Lock(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip string) (*TimesheetResponse, error) {
	lockKey := fmt.Sprintf("timesheet:%s:lock", id)
	lock, err := uc.locker.Acquire(ctx, lockKey)
	if err != nil {
		return nil, domain.ErrLockNotAcquired
	}
	defer lock.Release(ctx) //nolint:errcheck

	ts, err := uc.tsRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ts.Status != domain.TSStatusApproved {
		return nil, domain.ErrInvalidStateTransition
	}
	updated, err := uc.tsRepo.UpdateStatus(ctx, id, domain.TSStatusLocked, callerID, nil)
	if err != nil {
		return nil, err
	}
	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "timesheet", Resource: "timesheets",
		ResourceID: &id, Action: "STATE_TRANSITION", IPAddress: ip,
	})
	_ = uc.publisher.Publish(ctx, "timesheet", id, outbox.EventTimesheetLocked, worker.TimesheetEventPayload{
		TimesheetID: id.String(), StaffID: updated.StaffID.String(),
		Status: string(updated.Status), CallerID: callerID.String(),
	})
	if uc.broadcast != nil {
		_ = uc.broadcast.Broadcast("timesheet", "timesheet.locked", map[string]string{
			"timesheet_id": id.String(), "staff_id": updated.StaffID.String(), "status": string(updated.Status),
		})
	}
	resp := toTimesheetResponse(updated)
	return &resp, nil
}
