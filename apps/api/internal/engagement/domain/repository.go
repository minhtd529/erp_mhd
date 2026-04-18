package domain

import (
	"context"

	"github.com/google/uuid"
)

// EngagementRepository defines the data-access contract for Engagement.
type EngagementRepository interface {
	Create(ctx context.Context, p CreateEngagementParams) (*Engagement, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Engagement, error)
	Update(ctx context.Context, p UpdateEngagementParams) (*Engagement, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status EngagementStatus, updatedBy uuid.UUID) (*Engagement, error)
	SoftDelete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error
	List(ctx context.Context, f ListEngagementsFilter) ([]*Engagement, int64, error)
	// ListCursor returns size+1 rows starting after the cursor position (for cursor pagination).
	ListCursor(ctx context.Context, f CursorFilter) ([]*Engagement, error)
}

// MemberRepository defines the data-access contract for EngagementMember.
type MemberRepository interface {
	Assign(ctx context.Context, p AssignMemberParams) (*EngagementMember, error)
	FindByID(ctx context.Context, id uuid.UUID) (*EngagementMember, error)
	Update(ctx context.Context, p UpdateMemberParams) (*EngagementMember, error)
	SoftDelete(ctx context.Context, id uuid.UUID, engagementID uuid.UUID, deletedBy uuid.UUID) error
	ListByEngagement(ctx context.Context, engagementID uuid.UUID) ([]*EngagementMember, error)
	SumAllocation(ctx context.Context, engagementID uuid.UUID, excludeID *uuid.UUID) (int, error)
}

// TaskRepository defines the data-access contract for EngagementTask.
type TaskRepository interface {
	Create(ctx context.Context, p CreateTaskParams) (*EngagementTask, error)
	FindByID(ctx context.Context, id uuid.UUID) (*EngagementTask, error)
	Update(ctx context.Context, p UpdateTaskParams) (*EngagementTask, error)
	ListByEngagement(ctx context.Context, engagementID uuid.UUID, phase TaskPhase) ([]*EngagementTask, error)
}

// CostRepository defines the data-access contract for DirectCost.
type CostRepository interface {
	Create(ctx context.Context, p CreateCostParams) (*DirectCost, error)
	FindByID(ctx context.Context, id uuid.UUID) (*DirectCost, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status CostStatus, actorID uuid.UUID, rejectReason *string) (*DirectCost, error)
	ListByEngagement(ctx context.Context, engagementID uuid.UUID) ([]*DirectCost, error)
}
