// Package domain defines the WorkingPaper bounded context aggregates.
package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// WPStatus represents the lifecycle of a working paper.
type WPStatus string

const (
	WPStatusDraft     WPStatus = "DRAFT"
	WPStatusInReview  WPStatus = "IN_REVIEW"
	WPStatusCommented WPStatus = "COMMENTED"
	WPStatusFinalized WPStatus = "FINALIZED"
	WPStatusSignedOff WPStatus = "SIGNED_OFF"
)

// DocumentType enumerates working paper categories.
type DocumentType string

const (
	DocProcedures DocumentType = "PROCEDURES"
	DocEvidence   DocumentType = "EVIDENCE"
	DocAnalysis   DocumentType = "ANALYSIS"
	DocConclusion DocumentType = "CONCLUSION"
	DocMgmtLetter DocumentType = "MANAGEMENT_LETTER"
)

// ReviewerRole enumerates review chain levels.
type ReviewerRole string

const (
	RoleAuditor       ReviewerRole = "AUDITOR"
	RoleSeniorAuditor ReviewerRole = "SENIOR_AUDITOR"
	RoleManager       ReviewerRole = "MANAGER"
	RolePartner       ReviewerRole = "PARTNER"
)

// ReviewStatus represents the outcome of a review step.
type ReviewStatus string

const (
	ReviewPending  ReviewStatus = "PENDING"
	ReviewReviewed ReviewStatus = "REVIEWED"
	ReviewApproved ReviewStatus = "APPROVED"
	ReviewRejected ReviewStatus = "REJECTED"
)

// IssueStatus represents the state of a review comment.
type IssueStatus string

const (
	IssueOpen     IssueStatus = "OPEN"
	IssueResolved IssueStatus = "RESOLVED"
)

// WorkingPaper is the aggregate root for the WorkingPapers context.
type WorkingPaper struct {
	ID           uuid.UUID       `json:"id"`
	EngagementID uuid.UUID       `json:"engagement_id"`
	FolderID     *uuid.UUID      `json:"folder_id"`
	DocumentType DocumentType    `json:"document_type"`
	Title        string          `json:"title"`
	Status       WPStatus        `json:"status"`
	FileID       *uuid.UUID      `json:"file_id"`
	SnapshotData json.RawMessage `json:"snapshot_data"`
	IsDeleted    bool            `json:"is_deleted"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	CreatedBy    uuid.UUID       `json:"created_by"`
	UpdatedBy    uuid.UUID       `json:"updated_by"`
}

// WorkingPaperReview records a review step in the approval chain.
type WorkingPaperReview struct {
	ID             uuid.UUID    `json:"id"`
	WorkingPaperID uuid.UUID    `json:"working_paper_id"`
	ReviewerRole   ReviewerRole `json:"reviewer_role"`
	ReviewStatus   ReviewStatus `json:"review_status"`
	ReviewDate     *time.Time   `json:"review_date"`
	ReviewedBy     *uuid.UUID   `json:"reviewed_by"`
	CreatedAt      time.Time    `json:"created_at"`
}

// WorkingPaperComment is an issue raised during review.
type WorkingPaperComment struct {
	ID          uuid.UUID   `json:"id"`
	ReviewID    uuid.UUID   `json:"review_id"`
	CommentText string      `json:"comment_text"`
	IssueStatus IssueStatus `json:"issue_status"`
	RaisedAt    time.Time   `json:"raised_at"`
	ResolvedAt  *time.Time  `json:"resolved_at"`
	CreatedBy   uuid.UUID   `json:"created_by"`
}

// WorkingPaperFolder groups working papers within an engagement.
type WorkingPaperFolder struct {
	ID           uuid.UUID `json:"id"`
	EngagementID uuid.UUID `json:"engagement_id"`
	FolderName   string    `json:"folder_name"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    uuid.UUID `json:"created_by"`
}

// AuditTemplate is a reusable working paper template.
type AuditTemplate struct {
	ID           uuid.UUID       `json:"id"`
	TemplateType string          `json:"template_type"`
	Title        string          `json:"title"`
	Version      string          `json:"version"`
	Content      json.RawMessage `json:"content"`
	VSACompliant bool            `json:"vsa_compliant"`
	IsActive     bool            `json:"is_active"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	CreatedBy    uuid.UUID       `json:"created_by"`
	UpdatedBy    uuid.UUID       `json:"updated_by"`
}
