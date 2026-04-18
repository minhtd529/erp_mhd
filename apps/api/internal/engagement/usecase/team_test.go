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

// ── mock member repo ──────────────────────────────────────────────────────────

type mockMemberRepo struct {
	assigned   *domain.EngagementMember
	found      *domain.EngagementMember
	updated    *domain.EngagementMember
	assignErr  error
	findErr    error
	updateErr  error
	deleteErr  error
	listItems  []*domain.EngagementMember
	sumValue   int
	sumErr     error
}

func (m *mockMemberRepo) Assign(_ context.Context, _ domain.AssignMemberParams) (*domain.EngagementMember, error) {
	return m.assigned, m.assignErr
}
func (m *mockMemberRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.EngagementMember, error) {
	return m.found, m.findErr
}
func (m *mockMemberRepo) Update(_ context.Context, _ domain.UpdateMemberParams) (*domain.EngagementMember, error) {
	return m.updated, m.updateErr
}
func (m *mockMemberRepo) SoftDelete(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ uuid.UUID) error {
	return m.deleteErr
}
func (m *mockMemberRepo) ListByEngagement(_ context.Context, _ uuid.UUID) ([]*domain.EngagementMember, error) {
	return m.listItems, m.findErr
}
func (m *mockMemberRepo) SumAllocation(_ context.Context, _ uuid.UUID, _ *uuid.UUID) (int, error) {
	return m.sumValue, m.sumErr
}

// ── helper ────────────────────────────────────────────────────────────────────

func newMember(engID uuid.UUID, alloc int) *domain.EngagementMember {
	return &domain.EngagementMember{
		ID:                uuid.New(),
		EngagementID:      engID,
		StaffID:           uuid.New(),
		Role:              domain.RoleAuditor,
		AllocationPercent: alloc,
		Status:            domain.MemberAssigned,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		CreatedBy:         uuid.New(),
		UpdatedBy:         uuid.New(),
	}
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestTeamUseCase_Assign(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()
	engID := uuid.New()

	tests := []struct {
		name     string
		sum      int
		alloc    int
		engFound *domain.Engagement
		engErr   error
		wantErr  error
	}{
		{
			name:     "success: 50% allocation, existing sum 30%",
			sum:      30,
			alloc:    50,
			engFound: newEngagement(domain.StatusActive, nil),
		},
		{
			name:    "exceeds 100%: existing 60 + new 50",
			sum:     60,
			alloc:   50,
			engFound: newEngagement(domain.StatusActive, nil),
			wantErr: domain.ErrTeamAllocationExceeds,
		},
		{
			name:    "engagement not found",
			engErr:  domain.ErrEngagementNotFound,
			wantErr: domain.ErrEngagementNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			engRepo := &mockEngagementRepo{found: tt.engFound, findErr: tt.engErr}
			memberRepo := &mockMemberRepo{
				sumValue: tt.sum,
				assigned: newMember(engID, tt.alloc),
			}
			uc := usecase.NewTeamUseCase(memberRepo, engRepo, nil)
			_, err := uc.Assign(context.Background(), engID, usecase.MemberAssignRequest{
				StaffID:           uuid.New(),
				Role:              domain.RoleAuditor,
				AllocationPercent: tt.alloc,
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

func TestTeamUseCase_Unassign(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()

	tests := []struct {
		name      string
		deleteErr error
		wantErr   error
	}{
		{name: "success"},
		{
			name:      "not found",
			deleteErr: domain.ErrMemberNotFound,
			wantErr:   domain.ErrMemberNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			memberRepo := &mockMemberRepo{deleteErr: tt.deleteErr}
			uc := usecase.NewTeamUseCase(memberRepo, &mockEngagementRepo{}, nil)
			err := uc.Unassign(context.Background(), uuid.New(), uuid.New(), callerID, "")
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
