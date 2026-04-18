package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CreateWPParams struct {
	EngagementID uuid.UUID
	FolderID     *uuid.UUID
	DocumentType DocumentType
	Title        string
	FileID       *uuid.UUID
	CreatedBy    uuid.UUID
}

type UpdateWPParams struct {
	ID       uuid.UUID
	Title    string
	FolderID *uuid.UUID
	FileID   *uuid.UUID
	UpdatedBy uuid.UUID
}

type ListWPFilter struct {
	EngagementID uuid.UUID
	Status       WPStatus
	Page         int
	Size         int
}

type CreateReviewParams struct {
	WorkingPaperID uuid.UUID
	ReviewerRole   ReviewerRole
}

type ApproveReviewParams struct {
	WorkingPaperID uuid.UUID
	ReviewerRole   ReviewerRole
	ReviewerID     uuid.UUID
}

type AddCommentParams struct {
	ReviewID    uuid.UUID
	CommentText string
	CreatedBy   uuid.UUID
}

type CreateFolderParams struct {
	EngagementID uuid.UUID
	FolderName   string
	CreatedBy    uuid.UUID
}

type CreateTemplateParams struct {
	TemplateType string
	Title        string
	Version      string
	Content      json.RawMessage
	VSACompliant bool
	CreatedBy    uuid.UUID
}

type UpdateTemplateParams struct {
	ID           uuid.UUID
	Title        string
	Content      json.RawMessage
	VSACompliant bool
	UpdatedBy    uuid.UUID
}

// FinalizeWPParams carries snapshot data to freeze.
type FinalizeWPParams struct {
	ID           uuid.UUID
	SnapshotData json.RawMessage
	FinalizedAt  time.Time
	UpdatedBy    uuid.UUID
}
