package domain

import "errors"

var (
	ErrEngagementNotFound     = errors.New("ENGAGEMENT_NOT_FOUND")
	ErrInvalidStateTransition = errors.New("INVALID_STATE_TRANSITION")
	ErrEngagementLocked       = errors.New("ENGAGEMENT_LOCKED")
	ErrTeamAllocationExceeds  = errors.New("TEAM_ALLOCATION_EXCEEDS_100")
	ErrPartnerNotAssigned     = errors.New("PARTNER_NOT_ASSIGNED")
	ErrMemberNotFound         = errors.New("MEMBER_NOT_FOUND")
	ErrTaskNotFound           = errors.New("TASK_NOT_FOUND")
	ErrCostNotFound           = errors.New("COST_NOT_FOUND")
	ErrCostApprovalRequired   = errors.New("COST_APPROVAL_REQUIRED")
	ErrInvalidCostTransition  = errors.New("INVALID_COST_TRANSITION")
)
