package domain

import (
	"context"

	"github.com/google/uuid"
)

// WorkingPaperRepository defines data-access for WorkingPaper aggregates.
type WorkingPaperRepository interface {
	Create(ctx context.Context, p CreateWPParams) (*WorkingPaper, error)
	FindByID(ctx context.Context, id uuid.UUID) (*WorkingPaper, error)
	Update(ctx context.Context, p UpdateWPParams) (*WorkingPaper, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status WPStatus, updatedBy uuid.UUID) (*WorkingPaper, error)
	Finalize(ctx context.Context, p FinalizeWPParams) (*WorkingPaper, error)
	SoftDelete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error
	List(ctx context.Context, f ListWPFilter) ([]*WorkingPaper, int64, error)
	// ListPendingReview returns WPs in IN_REVIEW/COMMENTED status where the given role
	// has not yet approved, so the reviewer can see their queue.
	ListPendingReview(ctx context.Context, role ReviewerRole, page, size int) ([]*WorkingPaper, int64, error)
}

// ReviewRepository defines data-access for WorkingPaperReview aggregates.
type ReviewRepository interface {
	Create(ctx context.Context, p CreateReviewParams) (*WorkingPaperReview, error)
	FindByWPAndRole(ctx context.Context, wpID uuid.UUID, role ReviewerRole) (*WorkingPaperReview, error)
	Approve(ctx context.Context, p ApproveReviewParams) (*WorkingPaperReview, error)
	RequestChanges(ctx context.Context, p ApproveReviewParams) (*WorkingPaperReview, error)
	ListByWP(ctx context.Context, wpID uuid.UUID) ([]*WorkingPaperReview, error)
	CountUnresolved(ctx context.Context, wpID uuid.UUID) (int, error)
}

// CommentRepository defines data-access for WorkingPaperComment aggregates.
type CommentRepository interface {
	Add(ctx context.Context, p AddCommentParams) (*WorkingPaperComment, error)
	Resolve(ctx context.Context, id uuid.UUID) (*WorkingPaperComment, error)
	ListByReview(ctx context.Context, reviewID uuid.UUID) ([]*WorkingPaperComment, error)
}

// FolderRepository defines data-access for WorkingPaperFolder.
type FolderRepository interface {
	Create(ctx context.Context, p CreateFolderParams) (*WorkingPaperFolder, error)
	ListByEngagement(ctx context.Context, engagementID uuid.UUID) ([]*WorkingPaperFolder, error)
}

// TemplateRepository defines data-access for AuditTemplate.
type TemplateRepository interface {
	Create(ctx context.Context, p CreateTemplateParams) (*AuditTemplate, error)
	FindByID(ctx context.Context, id uuid.UUID) (*AuditTemplate, error)
	Update(ctx context.Context, p UpdateTemplateParams) (*AuditTemplate, error)
	Retire(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) error
	List(ctx context.Context, activeOnly bool, page, size int) ([]*AuditTemplate, int64, error)
}
