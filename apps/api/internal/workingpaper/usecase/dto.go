package usecase

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/workingpaper/domain"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

type PaginatedResult[T any] = pagination.OffsetResult[T]

func paginateSlice[T any](all []T, page, size int) PaginatedResult[T] {
	start := (page - 1) * size
	if start > len(all) {
		start = len(all)
	}
	end := start + size
	if end > len(all) {
		end = len(all)
	}
	return pagination.NewOffsetResult(all[start:end], int64(len(all)), page, size)
}

// ── Working Paper DTOs ────────────────────────────────────────────────────────

type WPCreateRequest struct {
	FolderID     *uuid.UUID          `json:"folder_id"`
	DocumentType domain.DocumentType `json:"document_type" binding:"required"`
	Title        string              `json:"title"         binding:"required"`
	FileID       *uuid.UUID          `json:"file_id"`
}

type WPUpdateRequest struct {
	Title    string     `json:"title"`
	FolderID *uuid.UUID `json:"folder_id"`
	FileID   *uuid.UUID `json:"file_id"`
}

type WPListRequest struct {
	Status domain.WPStatus `form:"status"`
	Page   int             `form:"page"`
	Size   int             `form:"size"`
}

type WPResponse struct {
	ID           uuid.UUID           `json:"id"`
	EngagementID uuid.UUID           `json:"engagement_id"`
	FolderID     *uuid.UUID          `json:"folder_id"`
	DocumentType domain.DocumentType `json:"document_type"`
	Title        string              `json:"title"`
	Status       domain.WPStatus     `json:"status"`
	FileID       *uuid.UUID          `json:"file_id"`
	SnapshotData json.RawMessage     `json:"snapshot_data"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
	CreatedBy    uuid.UUID           `json:"created_by"`
}

func toWPResponse(wp *domain.WorkingPaper) WPResponse {
	return WPResponse{
		ID:           wp.ID,
		EngagementID: wp.EngagementID,
		FolderID:     wp.FolderID,
		DocumentType: wp.DocumentType,
		Title:        wp.Title,
		Status:       wp.Status,
		FileID:       wp.FileID,
		SnapshotData: wp.SnapshotData,
		CreatedAt:    wp.CreatedAt,
		UpdatedAt:    wp.UpdatedAt,
		CreatedBy:    wp.CreatedBy,
	}
}

// ── Review DTOs ───────────────────────────────────────────────────────────────

type ReviewResponse struct {
	ID             uuid.UUID           `json:"id"`
	WorkingPaperID uuid.UUID           `json:"working_paper_id"`
	ReviewerRole   domain.ReviewerRole `json:"reviewer_role"`
	ReviewStatus   domain.ReviewStatus `json:"review_status"`
	ReviewDate     *time.Time          `json:"review_date"`
	ReviewedBy     *uuid.UUID          `json:"reviewed_by"`
	CreatedAt      time.Time           `json:"created_at"`
}

func toReviewResponse(r *domain.WorkingPaperReview) ReviewResponse {
	return ReviewResponse{
		ID:             r.ID,
		WorkingPaperID: r.WorkingPaperID,
		ReviewerRole:   r.ReviewerRole,
		ReviewStatus:   r.ReviewStatus,
		ReviewDate:     r.ReviewDate,
		ReviewedBy:     r.ReviewedBy,
		CreatedAt:      r.CreatedAt,
	}
}

// ── Comment DTOs ──────────────────────────────────────────────────────────────

type CommentAddRequest struct {
	CommentText string `json:"comment_text" binding:"required"`
}

type CommentResponse struct {
	ID          uuid.UUID           `json:"id"`
	ReviewID    uuid.UUID           `json:"review_id"`
	CommentText string              `json:"comment_text"`
	IssueStatus domain.IssueStatus  `json:"issue_status"`
	RaisedAt    time.Time           `json:"raised_at"`
	ResolvedAt  *time.Time          `json:"resolved_at"`
	CreatedBy   uuid.UUID           `json:"created_by"`
}

func toCommentResponse(c *domain.WorkingPaperComment) CommentResponse {
	return CommentResponse{
		ID:          c.ID,
		ReviewID:    c.ReviewID,
		CommentText: c.CommentText,
		IssueStatus: c.IssueStatus,
		RaisedAt:    c.RaisedAt,
		ResolvedAt:  c.ResolvedAt,
		CreatedBy:   c.CreatedBy,
	}
}

// ── Folder DTOs ───────────────────────────────────────────────────────────────

type FolderCreateRequest struct {
	FolderName string `json:"folder_name" binding:"required"`
}

type FolderResponse struct {
	ID           uuid.UUID `json:"id"`
	EngagementID uuid.UUID `json:"engagement_id"`
	FolderName   string    `json:"folder_name"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    uuid.UUID `json:"created_by"`
}

func toFolderResponse(f *domain.WorkingPaperFolder) FolderResponse {
	return FolderResponse{
		ID:           f.ID,
		EngagementID: f.EngagementID,
		FolderName:   f.FolderName,
		CreatedAt:    f.CreatedAt,
		CreatedBy:    f.CreatedBy,
	}
}

// ── Pending review DTOs ───────────────────────────────────────────────────────

type PendingReviewRequest struct {
	Role domain.ReviewerRole `form:"role"`
	Page int                 `form:"page"`
	Size int                 `form:"size"`
}

// ── Template DTOs ─────────────────────────────────────────────────────────────

type TemplateCreateRequest struct {
	TemplateType string          `json:"template_type" binding:"required"`
	Title        string          `json:"title"         binding:"required"`
	Version      string          `json:"version"`
	Content      json.RawMessage `json:"content"`
	VSACompliant bool            `json:"vsa_compliant"`
}

type TemplateUpdateRequest struct {
	Title        string          `json:"title"`
	Content      json.RawMessage `json:"content"`
	VSACompliant bool            `json:"vsa_compliant"`
}

type TemplateListRequest struct {
	ActiveOnly bool `form:"active_only"`
	Page       int  `form:"page"`
	Size       int  `form:"size"`
}

type TemplateResponse struct {
	ID           uuid.UUID       `json:"id"`
	TemplateType string          `json:"template_type"`
	Title        string          `json:"title"`
	Version      string          `json:"version"`
	Content      json.RawMessage `json:"content"`
	VSACompliant bool            `json:"vsa_compliant"`
	IsActive     bool            `json:"is_active"`
	CreatedAt    time.Time       `json:"created_at"`
	CreatedBy    uuid.UUID       `json:"created_by"`
}

func toTemplateResponse(t *domain.AuditTemplate) TemplateResponse {
	return TemplateResponse{
		ID:           t.ID,
		TemplateType: t.TemplateType,
		Title:        t.Title,
		Version:      t.Version,
		Content:      t.Content,
		VSACompliant: t.VSACompliant,
		IsActive:     t.IsActive,
		CreatedAt:    t.CreatedAt,
		CreatedBy:    t.CreatedBy,
	}
}

// ── Apply template response ───────────────────────────────────────────────────

type ApplyTemplateResponse struct {
	Created int          `json:"created"`
	Papers  []WPResponse `json:"papers"`
}
