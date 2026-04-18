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

// ── mock cost repo ────────────────────────────────────────────────────────────

type mockCostRepo struct {
	created    *domain.DirectCost
	found      *domain.DirectCost
	updated    *domain.DirectCost
	createErr  error
	findErr    error
	updateErr  error
	listItems  []*domain.DirectCost
}

func (m *mockCostRepo) Create(_ context.Context, _ domain.CreateCostParams) (*domain.DirectCost, error) {
	return m.created, m.createErr
}
func (m *mockCostRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.DirectCost, error) {
	return m.found, m.findErr
}
func (m *mockCostRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ domain.CostStatus, _ uuid.UUID, _ *string) (*domain.DirectCost, error) {
	return m.updated, m.updateErr
}
func (m *mockCostRepo) ListByEngagement(_ context.Context, _ uuid.UUID) ([]*domain.DirectCost, error) {
	return m.listItems, m.findErr
}

// ── helper ────────────────────────────────────────────────────────────────────

func newCost(engID uuid.UUID, status domain.CostStatus) *domain.DirectCost {
	return &domain.DirectCost{
		ID:           uuid.New(),
		EngagementID: engID,
		CostType:     domain.CostTravel,
		Description:  "taxi",
		Amount:       50,
		Status:       status,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		CreatedBy:    uuid.New(),
		UpdatedBy:    uuid.New(),
	}
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestCostUseCase_Submit(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()
	engID := uuid.New()

	tests := []struct {
		name    string
		found   *domain.DirectCost
		findErr error
		wantErr error
	}{
		{
			name:  "success: DRAFT → SUBMITTED",
			found: newCost(engID, domain.CostDraft),
		},
		{
			name:    "already submitted → INVALID_COST_TRANSITION",
			found:   newCost(engID, domain.CostSubmitted),
			wantErr: domain.ErrInvalidCostTransition,
		},
		{
			name:    "wrong engagement",
			found:   newCost(uuid.New(), domain.CostDraft), // different engID
			wantErr: domain.ErrCostNotFound,
		},
		{
			name:    "not found",
			findErr: domain.ErrCostNotFound,
			wantErr: domain.ErrCostNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			costRepo := &mockCostRepo{
				found:   tt.found,
				findErr: tt.findErr,
				updated: newCost(engID, domain.CostSubmitted),
			}
			uc := usecase.NewCostUseCase(costRepo, &mockEngagementRepo{found: newEngagement(domain.StatusActive, nil)}, nil)
			_, err := uc.Submit(context.Background(), engID, uuid.New(), callerID, "")
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

func TestCostUseCase_Approve(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()
	engID := uuid.New()

	tests := []struct {
		name    string
		found   *domain.DirectCost
		wantErr error
	}{
		{
			name:  "success: SUBMITTED → APPROVED",
			found: newCost(engID, domain.CostSubmitted),
		},
		{
			name:    "draft → COST_APPROVAL_REQUIRED",
			found:   newCost(engID, domain.CostDraft),
			wantErr: domain.ErrCostApprovalRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			costRepo := &mockCostRepo{
				found:   tt.found,
				updated: newCost(engID, domain.CostApproved),
			}
			uc := usecase.NewCostUseCase(costRepo, &mockEngagementRepo{found: newEngagement(domain.StatusActive, nil)}, nil)
			_, err := uc.Approve(context.Background(), engID, uuid.New(), callerID, "")
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
