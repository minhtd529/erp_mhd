package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/timesheet/domain"
	"github.com/mdh/erp-audit/api/internal/timesheet/usecase"
)

// ── mock entry repo ───────────────────────────────────────────────────────────

type mockEntryRepo struct {
	created   *domain.TimesheetEntry
	found     *domain.TimesheetEntry
	updated   *domain.TimesheetEntry
	createErr error
	findErr   error
	updateErr error
	deleteErr error
	listItems []*domain.TimesheetEntry
}

func (m *mockEntryRepo) Create(_ context.Context, _ domain.CreateEntryParams) (*domain.TimesheetEntry, error) {
	return m.created, m.createErr
}
func (m *mockEntryRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.TimesheetEntry, error) {
	return m.found, m.findErr
}
func (m *mockEntryRepo) Update(_ context.Context, _ domain.UpdateEntryParams) (*domain.TimesheetEntry, error) {
	return m.updated, m.updateErr
}
func (m *mockEntryRepo) SoftDelete(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ uuid.UUID) error {
	return m.deleteErr
}
func (m *mockEntryRepo) ListByTimesheet(_ context.Context, _ uuid.UUID) ([]*domain.TimesheetEntry, error) {
	return m.listItems, m.findErr
}
func (m *mockEntryRepo) ListLockedByEngagement(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]*domain.TimesheetEntry, error) {
	return m.listItems, m.findErr
}

// ── helper ────────────────────────────────────────────────────────────────────

func newEntry(tsID uuid.UUID, date time.Time) *domain.TimesheetEntry {
	return &domain.TimesheetEntry{
		ID:           uuid.New(),
		TimesheetID:  tsID,
		EntryDate:    date,
		EngagementID: uuid.New(),
		HoursWorked:  8,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		CreatedBy:    uuid.New(),
		UpdatedBy:    uuid.New(),
	}
}

// monday returns the Monday of the current week.
func monday() time.Time {
	now := time.Now()
	wd := int(now.Weekday())
	if wd == 0 {
		wd = 7
	}
	return now.AddDate(0, 0, -(wd - 1)).Truncate(24 * time.Hour)
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestEntryUseCase_Create(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()
	mon := monday()
	tsID := uuid.New()

	tests := []struct {
		name      string
		tsFound   *domain.Timesheet
		tsFindErr error
		entryDate string
		wantErr   error
	}{
		{
			name:      "success: entry on Monday",
			tsFound:   &domain.Timesheet{ID: tsID, Status: domain.TSStatusDraft, PeriodStartDate: mon},
			entryDate: mon.Format("2006-01-02"),
		},
		{
			name:      "success: entry on Sunday (last day of week)",
			tsFound:   &domain.Timesheet{ID: tsID, Status: domain.TSStatusDraft, PeriodStartDate: mon},
			entryDate: mon.AddDate(0, 0, 6).Format("2006-01-02"),
		},
		{
			name:      "entry date before period → ENTRY_DATE_OUTSIDE_PERIOD",
			tsFound:   &domain.Timesheet{ID: tsID, Status: domain.TSStatusDraft, PeriodStartDate: mon},
			entryDate: mon.AddDate(0, 0, -1).Format("2006-01-02"),
			wantErr:   domain.ErrEntryDateOutsidePeriod,
		},
		{
			name:      "approved timesheet → TIMESHEET_NOT_EDITABLE",
			tsFound:   &domain.Timesheet{ID: tsID, Status: domain.TSStatusApproved, PeriodStartDate: mon},
			entryDate: mon.Format("2006-01-02"),
			wantErr:   domain.ErrTimesheetNotEditable,
		},
		{
			name:      "locked timesheet → TIMESHEET_NOT_EDITABLE",
			tsFound:   &domain.Timesheet{ID: tsID, Status: domain.TSStatusLocked, PeriodStartDate: mon},
			entryDate: mon.Format("2006-01-02"),
			wantErr:   domain.ErrTimesheetNotEditable,
		},
		{
			name:      "timesheet not found",
			tsFindErr: domain.ErrTimesheetNotFound,
			entryDate: mon.Format("2006-01-02"),
			wantErr:   domain.ErrTimesheetNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tsRepo := &mockTSRepo{found: tt.tsFound, findErr: tt.tsFindErr}
			entryRepo := &mockEntryRepo{created: newEntry(tsID, mon)}
			uc := usecase.NewEntryUseCase(entryRepo, tsRepo, nil)
			_, err := uc.Create(context.Background(), tsID, usecase.EntryCreateRequest{
				EntryDate:    tt.entryDate,
				EngagementID: uuid.New(),
				HoursWorked:  8,
			}, callerID, "")
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("want %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestEntryUseCase_Delete(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()
	tsID := uuid.New()
	mon := monday()

	tests := []struct {
		name      string
		tsFound   *domain.Timesheet
		deleteErr error
		wantErr   error
	}{
		{
			name:    "success",
			tsFound: &domain.Timesheet{ID: tsID, Status: domain.TSStatusDraft, PeriodStartDate: mon},
		},
		{
			name:    "locked → TIMESHEET_NOT_EDITABLE",
			tsFound: &domain.Timesheet{ID: tsID, Status: domain.TSStatusLocked, PeriodStartDate: mon},
			wantErr: domain.ErrTimesheetNotEditable,
		},
		{
			name:      "entry not found",
			tsFound:   &domain.Timesheet{ID: tsID, Status: domain.TSStatusDraft, PeriodStartDate: mon},
			deleteErr: domain.ErrEntryNotFound,
			wantErr:   domain.ErrEntryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tsRepo := &mockTSRepo{found: tt.tsFound}
			entryRepo := &mockEntryRepo{deleteErr: tt.deleteErr}
			uc := usecase.NewEntryUseCase(entryRepo, tsRepo, nil)
			err := uc.Delete(context.Background(), tsID, uuid.New(), callerID, "")
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("want %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
