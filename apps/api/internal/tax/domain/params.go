package domain

import (
	"time"

	"github.com/google/uuid"
)

type CreateDeadlineParams struct {
	ClientID               uuid.UUID
	DeadlineType           DeadlineType
	DeadlineName           string
	DueDate                time.Time
	ExpectedSubmissionDate *time.Time
	Notes                  string
	CreatedBy              uuid.UUID
}

type UpdateDeadlineParams struct {
	ID                     uuid.UUID
	DeadlineName           string
	DueDate                time.Time
	ExpectedSubmissionDate *time.Time
	Notes                  string
	UpdatedBy              uuid.UUID
}

type ListDeadlinesFilter struct {
	ClientID *uuid.UUID
	Status   DeadlineStatus
}

type CreateAdvisoryParams struct {
	ClientID       uuid.UUID
	EngagementID   *uuid.UUID
	AdvisoryType   AdvisoryType
	Recommendation string
	Findings       string
	CreatedBy      uuid.UUID
}

type UpdateAdvisoryParams struct {
	ID             uuid.UUID
	Recommendation string
	Findings       string
	UpdatedBy      uuid.UUID
}

type AttachFileParams struct {
	AdvisoryID uuid.UUID
	FileName   string
	FilePath   string
	CreatedBy  uuid.UUID
}

type ListAdvisoryFilter struct {
	ClientID *uuid.UUID
	Status   AdvisoryStatus
}
