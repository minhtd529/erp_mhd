package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/timesheet/domain"
	"github.com/mdh/erp-audit/api/internal/timesheet/usecase"
	"github.com/mdh/erp-audit/api/pkg/distlock"
)

// ── mock repos ────────────────────────────────────────────────────────────────

type mockTSRepo struct {
	created   *domain.Timesheet
	found     *domain.Timesheet
	statusUpd *domain.Timesheet
	createErr error
	findErr   error
	statusErr error
	listItems []*domain.Timesheet
	listTotal int64
}

func (m *mockTSRepo) GetOrCreate(_ context.Context, _ domain.GetOrCreateTimesheetParams) (*domain.Timesheet, error) {
	return m.created, m.createErr
}
func (m *mockTSRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.Timesheet, error) {
	return m.found, m.findErr
}
func (m *mockTSRepo) FindByStaffAndWeek(_ context.Context, _ uuid.UUID, _ time.Time) (*domain.Timesheet, error) {
	return m.found, m.findErr
}
func (m *mockTSRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ domain.TimesheetStatus, _ uuid.UUID, _ *string) (*domain.Timesheet, error) {
	return m.statusUpd, m.statusErr
}
func (m *mockTSRepo) UpdateTotalHours(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockTSRepo) List(_ context.Context, _ domain.ListTimesheetsFilter) ([]*domain.Timesheet, int64, error) {
	return m.listItems, m.listTotal, m.findErr
}
func (m *mockTSRepo) ListCursor(_ context.Context, _ domain.TimesheetCursorFilter) ([]*domain.Timesheet, error) {
	return m.listItems, m.findErr
}

// ── mock broadcaster ──────────────────────────────────────────────────────────

type captureBroadcaster struct {
	channel   string
	eventType string
	called    bool
}

func (b *captureBroadcaster) Broadcast(channel string, eventType string, _ any) error {
	b.channel = channel
	b.eventType = eventType
	b.called = true
	return nil
}

// ── mock locker ───────────────────────────────────────────────────────────────

// noopLocker always succeeds without hitting Redis — safe for unit tests.
type noopLocker struct{ acquireErr error }

func (l *noopLocker) Acquire(_ context.Context, _ string) (*distlock.Lock, error) {
	if l.acquireErr != nil {
		return nil, l.acquireErr
	}
	return distlock.NewNoopLock(), nil
}

// ── helper ────────────────────────────────────────────────────────────────────

func newTS(status domain.TimesheetStatus) *domain.Timesheet {
	now := time.Now()
	return &domain.Timesheet{
		ID:              uuid.New(),
		StaffID:         uuid.New(),
		PeriodStartDate: now.Truncate(24 * time.Hour),
		Status:          status,
		TotalHours:      8,
		CreatedAt:       now,
		UpdatedAt:       now,
		CreatedBy:       uuid.New(),
		UpdatedBy:       uuid.New(),
	}
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestTimesheetUseCase_Submit(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()

	tests := []struct {
		name    string
		found   *domain.Timesheet
		findErr error
		wantErr error
	}{
		{
			name:  "DRAFT → SUBMITTED",
			found: newTS(domain.TSStatusDraft),
		},
		{
			name:  "REJECTED → SUBMITTED (re-submission)",
			found: newTS(domain.TSStatusRejected),
		},
		{
			name:    "APPROVED → SUBMITTED → INVALID_STATE_TRANSITION",
			found:   newTS(domain.TSStatusApproved),
			wantErr: domain.ErrInvalidStateTransition,
		},
		{
			name:    "not found",
			findErr: domain.ErrTimesheetNotFound,
			wantErr: domain.ErrTimesheetNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &mockTSRepo{
				found:     tt.found,
				findErr:   tt.findErr,
				statusUpd: newTS(domain.TSStatusSubmitted),
			}
			uc := usecase.NewTimesheetUseCase(repo, &noopLocker{}, nil, nil, nil)
			_, err := uc.Submit(context.Background(), uuid.New(), callerID, "")
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

func TestTimesheetUseCase_Approve(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()

	tests := []struct {
		name       string
		found      *domain.Timesheet
		findErr    error
		acquireErr error
		wantErr    error
	}{
		{
			name:  "success: SUBMITTED → APPROVED",
			found: newTS(domain.TSStatusSubmitted),
		},
		{
			name:    "DRAFT → APPROVED → INVALID_STATE_TRANSITION",
			found:   newTS(domain.TSStatusDraft),
			wantErr: domain.ErrInvalidStateTransition,
		},
		{
			name:       "lock not acquired → LOCK_NOT_ACQUIRED",
			found:      newTS(domain.TSStatusSubmitted),
			acquireErr: distlock.ErrLockNotAcquired,
			wantErr:    domain.ErrLockNotAcquired,
		},
		{
			name:    "not found",
			findErr: domain.ErrTimesheetNotFound,
			wantErr: domain.ErrTimesheetNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &mockTSRepo{
				found:     tt.found,
				findErr:   tt.findErr,
				statusUpd: newTS(domain.TSStatusApproved),
			}
			uc := usecase.NewTimesheetUseCase(repo, &noopLocker{acquireErr: tt.acquireErr}, nil, nil, nil)
			_, err := uc.Approve(context.Background(), uuid.New(), callerID, "")
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

func TestTimesheetUseCase_Lock(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()

	tests := []struct {
		name    string
		found   *domain.Timesheet
		wantErr error
	}{
		{
			name:  "success: APPROVED → LOCKED",
			found: newTS(domain.TSStatusApproved),
		},
		{
			name:    "SUBMITTED → LOCKED → INVALID_STATE_TRANSITION",
			found:   newTS(domain.TSStatusSubmitted),
			wantErr: domain.ErrInvalidStateTransition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &mockTSRepo{
				found:     tt.found,
				statusUpd: newTS(domain.TSStatusLocked),
			}
			uc := usecase.NewTimesheetUseCase(repo, &noopLocker{}, nil, nil, nil)
			_, err := uc.Lock(context.Background(), uuid.New(), callerID, "")
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

func TestTimesheetUseCase_List(t *testing.T) {
	t.Parallel()
	items := []*domain.Timesheet{
		newTS(domain.TSStatusDraft),
		newTS(domain.TSStatusSubmitted),
	}
	uc := usecase.NewTimesheetUseCase(&mockTSRepo{listItems: items, listTotal: 2}, &noopLocker{}, nil, nil, nil)
	result, err := uc.List(context.Background(), usecase.TimesheetListRequest{Page: 1, Size: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("want total 2, got %d", result.Total)
	}
}

func TestTimesheetUseCase_ListCursor(t *testing.T) {
	t.Parallel()
	items := []*domain.Timesheet{
		newTS(domain.TSStatusDraft),
		newTS(domain.TSStatusSubmitted),
	}
	uc := usecase.NewTimesheetUseCase(&mockTSRepo{listItems: items}, &noopLocker{}, nil, nil, nil)

	t.Run("no cursor returns first page", func(t *testing.T) {
		t.Parallel()
		result, err := uc.ListCursor(context.Background(), usecase.TimesheetCursorListRequest{Size: 20})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Data) != 2 {
			t.Errorf("want 2 items, got %d", len(result.Data))
		}
	})

	t.Run("invalid cursor returns error", func(t *testing.T) {
		t.Parallel()
		_, err := uc.ListCursor(context.Background(), usecase.TimesheetCursorListRequest{
			Size:   20,
			Cursor: "!!!invalid!!!",
		})
		if err == nil {
			t.Fatal("expected error for invalid cursor")
		}
	})
}

func TestTimesheetUseCase_BroadcastOnApprove(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()
	broadcaster := &captureBroadcaster{}
	repo := &mockTSRepo{
		found:     newTS(domain.TSStatusSubmitted),
		statusUpd: newTS(domain.TSStatusApproved),
	}
	uc := usecase.NewTimesheetUseCase(repo, &noopLocker{}, nil, nil, broadcaster)
	_, err := uc.Approve(context.Background(), uuid.New(), callerID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !broadcaster.called {
		t.Fatal("expected Broadcast to be called after Approve")
	}
	if broadcaster.channel != "timesheet" {
		t.Errorf("want channel=timesheet, got %q", broadcaster.channel)
	}
	if broadcaster.eventType != "timesheet.approved" {
		t.Errorf("want eventType=timesheet.approved, got %q", broadcaster.eventType)
	}
}
