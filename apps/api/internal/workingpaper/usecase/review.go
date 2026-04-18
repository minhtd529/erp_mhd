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

func (uc *ReviewUseCase) ListReviews(ctx context.Context, wpID uuid.UUID) ([]ReviewResponse, error) {
	reviews, err := uc.reviewRepo.ListByWP(ctx, wpID)
	if err != nil {
		return nil, err
	}
	data := make([]ReviewResponse, len(reviews))
	for i, r := range reviews {
		data[i] = toReviewResponse(r)
	}
	return data, nil
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

	_ = uc.auditLog.Log(ctx, audit.Entry{
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

	_ = uc.auditLog.Log(ctx, audit.Entry{
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

	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "working_papers", Resource: "working_paper_comments",
		ResourceID: &comment.ID, Action: "CREATE", IPAddress: ip,
	})

	resp := toCommentResponse(comment)
	return &resp, nil
}

func (uc *ReviewUseCase) ListComments(ctx context.Context, wpID uuid.UUID, role domain.ReviewerRole) ([]CommentResponse, error) {
	rev, err := uc.reviewRepo.FindByWPAndRole(ctx, wpID, role)
	if err != nil {
		return nil, err
	}
	comments, err := uc.commentRepo.ListByReview(ctx, rev.ID)
	if err != nil {
		return nil, err
	}
	data := make([]CommentResponse, len(comments))
	for i, c := range comments {
		data[i] = toCommentResponse(c)
	}
	return data, nil
}

func (uc *ReviewUseCase) ResolveComment(ctx context.Context, commentID uuid.UUID, callerID uuid.UUID, ip string) (*CommentResponse, error) {
	comment, err := uc.commentRepo.Resolve(ctx, commentID)
	if err != nil {
		return nil, err
	}

	_ = uc.auditLog.Log(ctx, audit.Entry{
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
