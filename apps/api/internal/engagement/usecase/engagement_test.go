package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/engagement/domain"
	"github.com/mdh/erp-audit/api/internal/engagement/usecase"
)

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

// ── mock repos ────────────────────────────────────────────────────────────────

type mockEngagementRepo struct {
	created    *domain.Engagement
	found      *domain.Engagement
	updated    *domain.Engagement
	statusUpd  *domain.Engagement
	createErr  error
	findErr    error
	updateErr  error
	statusErr  error
	deleteErr  error
	listItems  []*domain.Engagement
	listTotal  int64
}

func (m *mockEngagementRepo) Create(_ context.Context, _ domain.CreateEngagementParams) (*domain.Engagement, error) {
	return m.created, m.createErr
}
func (m *mockEngagementRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.Engagement, error) {
	return m.found, m.findErr
}
func (m *mockEngagementRepo) Update(_ context.Context, _ domain.UpdateEngagementParams) (*domain.Engagement, error) {
	return m.updated, m.updateErr
}
func (m *mockEngagementRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ domain.EngagementStatus, _ uuid.UUID) (*domain.Engagement, error) {
	return m.statusUpd, m.statusErr
}
func (m *mockEngagementRepo) SoftDelete(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return m.deleteErr
}
func (m *mockEngagementRepo) List(_ context.Context, _ domain.ListEngagementsFilter) ([]*domain.Engagement, int64, error) {
	return m.listItems, m.listTotal, m.findErr
}
func (m *mockEngagementRepo) ListCursor(_ context.Context, _ domain.CursorFilter) ([]*domain.Engagement, error) {
	return m.listItems, m.findErr
}

// ── helper ───────────────────────────────────────────────────────────────────

func newEngagement(status domain.EngagementStatus, partnerID *uuid.UUID) *domain.Engagement {
	return &domain.Engagement{
		ID:          uuid.New(),
		ClientID:    uuid.New(),
		ServiceType: domain.ServiceAudit,
		FeeType:     domain.FeeFixed,
		FeeAmount:   1000,
		Status:      status,
		PartnerID:   partnerID,
		CreatedBy:   uuid.New(),
		UpdatedBy:   uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestEngagementUseCase_Create(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()
	eng := newEngagement(domain.StatusDraft, nil)

	tests := []struct {
		name    string
		repo    *mockEngagementRepo
		wantErr error
	}{
		{
			name: "success",
			repo: &mockEngagementRepo{created: eng},
		},
		{
			name:    "repo error propagated",
			repo:    &mockEngagementRepo{createErr: errors.New("db error")},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := usecase.NewEngagementUseCase(tt.repo, nil, nil)
			resp, err := uc.Create(context.Background(), usecase.EngagementCreateRequest{
				ClientID:    uuid.New(),
				ServiceType: domain.ServiceAudit,
				FeeType:     domain.FeeFixed,
				FeeAmount:   1000,
			}, callerID, "")
			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Status != domain.StatusDraft {
				t.Errorf("want status DRAFT, got %s", resp.Status)
			}
		})
	}
}

func TestEngagementUseCase_GetByID(t *testing.T) {
	t.Parallel()
	id := uuid.New()

	tests := []struct {
		name    string
		repo    *mockEngagementRepo
		wantErr error
	}{
		{
			name: "success",
			repo: &mockEngagementRepo{found: newEngagement(domain.StatusDraft, nil)},
		},
		{
			name:    "not found",
			repo:    &mockEngagementRepo{findErr: domain.ErrEngagementNotFound},
			wantErr: domain.ErrEngagementNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := usecase.NewEngagementUseCase(tt.repo, nil, nil)
			_, err := uc.GetByID(context.Background(), id)
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

func TestEngagementUseCase_Activate(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()
	partnerID := uuid.New()

	tests := []struct {
		name    string
		found   *domain.Engagement
		findErr error
		updated *domain.Engagement
		wantErr error
	}{
		{
			name:    "success: DRAFT → ACTIVE with partner",
			found:   newEngagement(domain.StatusDraft, &partnerID),
			updated: newEngagement(domain.StatusActive, &partnerID),
		},
		{
			name:    "no partner → PARTNER_NOT_ASSIGNED",
			found:   newEngagement(domain.StatusDraft, nil),
			wantErr: domain.ErrPartnerNotAssigned,
		},
		{
			name:    "COMPLETED → ACTIVE → INVALID_STATE_TRANSITION",
			found:   newEngagement(domain.StatusCompleted, &partnerID),
			wantErr: domain.ErrInvalidStateTransition,
		},
		{
			name:    "not found",
			findErr: domain.ErrEngagementNotFound,
			wantErr: domain.ErrEngagementNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &mockEngagementRepo{
				found:     tt.found,
				findErr:   tt.findErr,
				statusUpd: tt.updated,
			}
			uc := usecase.NewEngagementUseCase(repo, nil, nil)
			_, err := uc.Activate(context.Background(), uuid.New(), callerID, "")
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

func TestEngagementUseCase_Complete(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()
	partnerID := uuid.New()

	tests := []struct {
		name    string
		found   *domain.Engagement
		wantErr error
	}{
		{
			name:  "success: ACTIVE → COMPLETED",
			found: newEngagement(domain.StatusActive, &partnerID),
		},
		{
			name:    "DRAFT → COMPLETED → INVALID_STATE_TRANSITION",
			found:   newEngagement(domain.StatusDraft, &partnerID),
			wantErr: domain.ErrInvalidStateTransition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &mockEngagementRepo{
				found:     tt.found,
				statusUpd: newEngagement(domain.StatusCompleted, &partnerID),
			}
			uc := usecase.NewEngagementUseCase(repo, nil, nil)
			_, err := uc.Complete(context.Background(), uuid.New(), callerID, "")
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

func TestEngagementUseCase_List(t *testing.T) {
	t.Parallel()
	items := []*domain.Engagement{
		newEngagement(domain.StatusDraft, nil),
		newEngagement(domain.StatusActive, nil),
	}
	uc := usecase.NewEngagementUseCase(&mockEngagementRepo{listItems: items, listTotal: 2}, nil, nil)
	result, err := uc.List(context.Background(), usecase.EngagementListRequest{Page: 1, Size: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("want total 2, got %d", result.Total)
	}
	if len(result.Data) != 2 {
		t.Errorf("want 2 items, got %d", len(result.Data))
	}
}

func TestEngagementUseCase_ListCursor(t *testing.T) {
	t.Parallel()
	items := []*domain.Engagement{
		newEngagement(domain.StatusDraft, nil),
		newEngagement(domain.StatusActive, nil),
	}
	uc := usecase.NewEngagementUseCase(&mockEngagementRepo{listItems: items}, nil, nil)

	t.Run("no cursor returns first page", func(t *testing.T) {
		t.Parallel()
		result, err := uc.ListCursor(context.Background(), usecase.EngagementCursorListRequest{Size: 20})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Data) != 2 {
			t.Errorf("want 2 items, got %d", len(result.Data))
		}
		if result.HasMore {
			t.Error("expected HasMore=false")
		}
	})

	t.Run("invalid cursor returns error", func(t *testing.T) {
		t.Parallel()
		_, err := uc.ListCursor(context.Background(), usecase.EngagementCursorListRequest{
			Size:   20,
			Cursor: "!!!invalid!!!",
		})
		if err == nil {
			t.Fatal("expected error for invalid cursor")
		}
	})
}

func TestEngagementUseCase_Search(t *testing.T) {
	t.Parallel()
	matched := []*domain.Engagement{newEngagement(domain.StatusActive, nil)}
	uc := usecase.NewEngagementUseCase(&mockEngagementRepo{listItems: matched, listTotal: 1}, nil, nil)

	result, err := uc.List(context.Background(), usecase.EngagementListRequest{
		Page: 1, Size: 20, Q: "annual audit",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("want 1 result, got %d", result.Total)
	}
}

func TestEngagementUseCase_BroadcastOnActivate(t *testing.T) {
	t.Parallel()
	partnerID := uuid.New()
	broadcaster := &captureBroadcaster{}
	repo := &mockEngagementRepo{
		found:     newEngagement(domain.StatusDraft, &partnerID),
		statusUpd: newEngagement(domain.StatusActive, &partnerID),
	}
	uc := usecase.NewEngagementUseCase(repo, nil, broadcaster)
	_, err := uc.Activate(context.Background(), uuid.New(), uuid.New(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !broadcaster.called {
		t.Fatal("expected Broadcast to be called after Activate")
	}
	if broadcaster.channel != "engagement" {
		t.Errorf("want channel=engagement, got %q", broadcaster.channel)
	}
	if broadcaster.eventType != "engagement.state_changed" {
		t.Errorf("want eventType=engagement.state_changed, got %q", broadcaster.eventType)
	}
}
