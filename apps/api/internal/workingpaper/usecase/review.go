package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/workingpaper/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

// ReviewUseCase handles review approval workflow.
type ReviewUseCase struct {
	reviewRepo  domain.ReviewRepository
	commentRepo domain.CommentRepository
	wpRepo      domain.WorkingPaperRepository
	auditLog    *audit.Logger
}

// NewReviewUseCase constructs a ReviewUseCase.
func NewReviewUseCase(
	reviewRepo domain.ReviewRepository,
	commentRepo domain.CommentRepository,
	wpRepo domain.WorkingPaperRepository,
	auditLog *audit.Logger,
) *ReviewUseCase {
	return &ReviewUseCase{
		reviewRepo:  reviewRepo,
		commentRepo: commentRepo,
		wpRepo:      wpRepo,
		auditLog:    auditLog,
	}
}

// ReviewListRequest carries pagination params for review listing.
type ReviewListRequest struct {
	Page int
	Size int
}

func (uc *ReviewUseCase) ListReviews(ctx context.Context, wpID uuid.UUID, req ReviewListRequest) (PaginatedResult[ReviewResponse], error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	reviews, err := uc.reviewRepo.ListByWP(ctx, wpID)
	if err != nil {
		return PaginatedResult[ReviewResponse]{}, err
	}
	all := make([]ReviewResponse, len(reviews))
	for i, r := range reviews {
		all[i] = toReviewResponse(r)
	}
	return paginateSlice(all, req.Page, req.Size), nil
}

func (uc *ReviewUseCase) Approve(ctx context.Context, wpID uuid.UUID, role domain.ReviewerRole, callerID uuid.UUID, ip string) (*ReviewResponse, error) {
	if err := uc.checkWPInReview(ctx, wpID); err != nil {
		return nil, err
	}

	rev, err := uc.reviewRepo.Approve(ctx, domain.ApproveReviewParams{
		WorkingPaperID: wpID,
		ReviewerRole:   role,
		ReviewerID:     callerID,
	})
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "working_papers", Resource: "working_paper_reviews",
		ResourceID: &rev.ID, Action: "APPROVE", IPAddress: ip,
	})

	resp := toReviewResponse(rev)
	return &resp, nil
}

func (uc *ReviewUseCase) RequestChanges(ctx context.Context, wpID uuid.UUID, role domain.ReviewerRole, callerID uuid.UUID, ip string) (*ReviewResponse, error) {
	if err := uc.checkWPInReview(ctx, wpID); err != nil {
		return nil, err
	}

	rev, err := uc.reviewRepo.RequestChanges(ctx, domain.ApproveReviewParams{
		WorkingPaperID: wpID,
		ReviewerRole:   role,
		ReviewerID:     callerID,
	})
	if err != nil {
		return nil, err
	}

	// Transition WP to COMMENTED status
	_, _ = uc.wpRepo.UpdateStatus(ctx, wpID, domain.WPStatusCommented, callerID)

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "working_papers", Resource: "working_paper_reviews",
		ResourceID: &rev.ID, Action: "REJECT", IPAddress: ip,
	})

	resp := toReviewResponse(rev)
	return &resp, nil
}

func (uc *ReviewUseCase) AddComment(ctx context.Context, wpID uuid.UUID, role domain.ReviewerRole, req CommentAddRequest, callerID uuid.UUID, ip string) (*CommentResponse, error) {
	rev, err := uc.reviewRepo.FindByWPAndRole(ctx, wpID, role)
	if err != nil {
		return nil, err
	}

	comment, err := uc.commentRepo.Add(ctx, domain.AddCommentParams{
		ReviewID:    rev.ID,
		CommentText: req.CommentText,
		CreatedBy:   callerID,
	})
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "working_papers", Resource: "working_paper_comments",
		ResourceID: &comment.ID, Action: "CREATE", IPAddress: ip,
	})

	resp := toCommentResponse(comment)
	return &resp, nil
}

// CommentListRequest carries pagination params for comment listing.
type CommentListRequest struct {
	Page int
	Size int
}

func (uc *ReviewUseCase) ListComments(ctx context.Context, wpID uuid.UUID, role domain.ReviewerRole, req CommentListRequest) (PaginatedResult[CommentResponse], error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	rev, err := uc.reviewRepo.FindByWPAndRole(ctx, wpID, role)
	if err != nil {
		return PaginatedResult[CommentResponse]{}, err
	}
	comments, err := uc.commentRepo.ListByReview(ctx, rev.ID)
	if err != nil {
		return PaginatedResult[CommentResponse]{}, err
	}
	all := make([]CommentResponse, len(comments))
	for i, c := range comments {
		all[i] = toCommentResponse(c)
	}
	return paginateSlice(all, req.Page, req.Size), nil
}

func (uc *ReviewUseCase) ResolveComment(ctx context.Context, commentID uuid.UUID, callerID uuid.UUID, ip string) (*CommentResponse, error) {
	comment, err := uc.commentRepo.Resolve(ctx, commentID)
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "working_papers", Resource: "working_paper_comments",
		ResourceID: &commentID, Action: "UPDATE", IPAddress: ip,
		NewValue: map[string]string{"issue_status": "RESOLVED"},
	})

	resp := toCommentResponse(comment)
	return &resp, nil
}

func (uc *ReviewUseCase) checkWPInReview(ctx context.Context, wpID uuid.UUID) error {
	wp, err := uc.wpRepo.FindByID(ctx, wpID)
	if err != nil {
		return err
	}
	if wp.Status != domain.WPStatusInReview && wp.Status != domain.WPStatusCommented {
		return domain.ErrInvalidStateTransition
	}
	return nil
}
