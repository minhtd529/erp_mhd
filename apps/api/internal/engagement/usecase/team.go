package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/engagement/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

// TeamUseCase handles team member assignment within an engagement.
type TeamUseCase struct {
	memberRepo     domain.MemberRepository
	engagementRepo domain.EngagementRepository
	auditLog       *audit.Logger
}

// NewTeamUseCase constructs a TeamUseCase.
func NewTeamUseCase(memberRepo domain.MemberRepository, engagementRepo domain.EngagementRepository, auditLog *audit.Logger) *TeamUseCase {
	return &TeamUseCase{memberRepo: memberRepo, engagementRepo: engagementRepo, auditLog: auditLog}
}

func (uc *TeamUseCase) List(ctx context.Context, engagementID uuid.UUID) ([]MemberResponse, error) {
	members, err := uc.memberRepo.ListByEngagement(ctx, engagementID)
	if err != nil {
		return nil, err
	}
	out := make([]MemberResponse, len(members))
	for i, m := range members {
		out[i] = toMemberResponse(m)
	}
	return out, nil
}

func (uc *TeamUseCase) Assign(ctx context.Context, engagementID uuid.UUID, req MemberAssignRequest, callerID uuid.UUID, ip string) (*MemberResponse, error) {
	if _, err := uc.engagementRepo.FindByID(ctx, engagementID); err != nil {
		return nil, err
	}

	sum, err := uc.memberRepo.SumAllocation(ctx, engagementID, nil)
	if err != nil {
		return nil, err
	}
	if sum+req.AllocationPercent > 100 {
		return nil, domain.ErrTeamAllocationExceeds
	}

	m, err := uc.memberRepo.Assign(ctx, domain.AssignMemberParams{
		EngagementID:      engagementID,
		StaffID:           req.StaffID,
		Role:              req.Role,
		HourlyRate:        req.HourlyRate,
		AllocationPercent: req.AllocationPercent,
		CreatedBy:         callerID,
	})
	if err != nil {
		return nil, err
	}

	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "engagement", Resource: "engagement_members",
		ResourceID: &m.ID, Action: "CREATE", IPAddress: ip,
	})

	resp := toMemberResponse(m)
	return &resp, nil
}

func (uc *TeamUseCase) Update(ctx context.Context, engagementID uuid.UUID, memberID uuid.UUID, req MemberUpdateRequest, callerID uuid.UUID, ip string) (*MemberResponse, error) {
	sum, err := uc.memberRepo.SumAllocation(ctx, engagementID, &memberID)
	if err != nil {
		return nil, err
	}
	if sum+req.AllocationPercent > 100 {
		return nil, domain.ErrTeamAllocationExceeds
	}

	m, err := uc.memberRepo.Update(ctx, domain.UpdateMemberParams{
		ID:                memberID,
		EngagementID:      engagementID,
		Role:              req.Role,
		HourlyRate:        req.HourlyRate,
		AllocationPercent: req.AllocationPercent,
		UpdatedBy:         callerID,
	})
	if err != nil {
		return nil, err
	}

	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "engagement", Resource: "engagement_members",
		ResourceID: &memberID, Action: "UPDATE", IPAddress: ip,
	})

	resp := toMemberResponse(m)
	return &resp, nil
}

func (uc *TeamUseCase) Unassign(ctx context.Context, engagementID uuid.UUID, memberID uuid.UUID, callerID uuid.UUID, ip string) error {
	if err := uc.memberRepo.SoftDelete(ctx, memberID, engagementID, callerID); err != nil {
		return err
	}
	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "engagement", Resource: "engagement_members",
		ResourceID: &memberID, Action: "DELETE", IPAddress: ip,
	})
	return nil
}
