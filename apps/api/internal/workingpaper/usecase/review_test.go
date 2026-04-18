package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/workingpaper/domain"
	"github.com/mdh/erp-audit/api/internal/workingpaper/usecase"
)

// ── ReviewUseCase tests ───────────────────────────────────────────────────────

func buildReviewUC() (*fakeWPRepo, *fakeReviewRepo, *fakeCommentRepo, *usecase.ReviewUseCase) {
	wpRepo := newFakeWPRepo()
	revRepo := newFakeReviewRepo()
	commentRepo := newFakeCommentRepo()
	uc := usecase.NewReviewUseCase(revRepo, commentRepo, wpRepo, nil)
	return wpRepo, revRepo, commentRepo, uc
}

// seedInReview creates a WP in IN_REVIEW status with a seeded review chain entry.
func seedInReview(t *testing.T, wpRepo *fakeWPRepo, revRepo *fakeReviewRepo, role domain.ReviewerRole) (uuid.UUID, uuid.UUID) {
	t.Helper()
	caller := uuid.New()
	engID := uuid.New()
	wpUC := usecase.NewWorkingPaperUseCase(wpRepo, revRepo, nil)
	wp, err := wpUC.Create(context.Background(), engID, usecase.WPCreateRequest{
		DocumentType: domain.DocProcedures, Title: "Review WP",
	}, caller, "127.0.0.1")
	if err != nil {
		t.Fatalf("Create WP: %v", err)
	}
	// Force status to IN_REVIEW by calling SubmitForReview (which seeds review chain)
	_, err = wpUC.SubmitForReview(context.Background(), wp.ID, caller, "127.0.0.1")
	if err != nil {
		t.Fatalf("SubmitForReview: %v", err)
	}
	return wp.ID, caller
}

func TestReview_ListReviews_HappyPath(t *testing.T) {
	wpRepo, revRepo, _, uc := buildReviewUC()

	wpID, _ := seedInReview(t, wpRepo, revRepo, domain.RoleSeniorAuditor)

	result, err := uc.ListReviews(context.Background(), wpID, usecase.ReviewListRequest{Page: 1, Size: 20})
	if err != nil {
		t.Fatalf("ListReviews: %v", err)
	}
	// SubmitForReview seeds 3 review chain entries
	if len(result.Data) != 3 {
		t.Errorf("want 3 reviews, got %d", len(result.Data))
	}
}

func TestReview_ListReviews_Empty(t *testing.T) {
	_, revRepo, _, uc := buildReviewUC()
	_ = revRepo

	result, err := uc.ListReviews(context.Background(), uuid.New(), usecase.ReviewListRequest{Page: 1, Size: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Data) != 0 {
		t.Errorf("want 0 reviews, got %d", len(result.Data))
	}
}

func TestReview_Approve_HappyPath(t *testing.T) {
	wpRepo, revRepo, _, uc := buildReviewUC()

	wpID, caller := seedInReview(t, wpRepo, revRepo, domain.RoleSeniorAuditor)

	resp, err := uc.Approve(context.Background(), wpID, domain.RoleSeniorAuditor, caller, "127.0.0.1")
	if err != nil {
		t.Fatalf("Approve: %v", err)
	}
	if resp.ReviewStatus != domain.ReviewApproved {
		t.Errorf("want APPROVED, got %s", resp.ReviewStatus)
	}
}

func TestReview_Approve_WPNotInReview(t *testing.T) {
	wpRepo, revRepo, _, uc := buildReviewUC()

	// Create WP but do NOT submit for review (stays DRAFT)
	wpUC := usecase.NewWorkingPaperUseCase(wpRepo, revRepo, nil)
	wp, _ := wpUC.Create(context.Background(), uuid.New(), usecase.WPCreateRequest{
		DocumentType: domain.DocProcedures, Title: "Draft WP",
	}, uuid.New(), "127.0.0.1")

	_, err := uc.Approve(context.Background(), wp.ID, domain.RoleSeniorAuditor, uuid.New(), "127.0.0.1")
	if err != domain.ErrInvalidStateTransition {
		t.Errorf("want ErrInvalidStateTransition, got %v", err)
	}
}

func TestReview_RequestChanges_HappyPath(t *testing.T) {
	wpRepo, revRepo, _, uc := buildReviewUC()

	wpID, caller := seedInReview(t, wpRepo, revRepo, domain.RoleManager)

	resp, err := uc.RequestChanges(context.Background(), wpID, domain.RoleManager, caller, "127.0.0.1")
	if err != nil {
		t.Fatalf("RequestChanges: %v", err)
	}
	if resp.ReviewStatus != domain.ReviewRejected {
		t.Errorf("want REJECTED, got %s", resp.ReviewStatus)
	}
}

func TestReview_AddComment_HappyPath(t *testing.T) {
	wpRepo, revRepo, _, uc := buildReviewUC()

	wpID, caller := seedInReview(t, wpRepo, revRepo, domain.RoleSeniorAuditor)

	comment, err := uc.AddComment(context.Background(), wpID, domain.RoleSeniorAuditor, usecase.CommentAddRequest{
		CommentText: "Please fix the variance analysis",
	}, caller, "127.0.0.1")
	if err != nil {
		t.Fatalf("AddComment: %v", err)
	}
	if comment.CommentText != "Please fix the variance analysis" {
		t.Errorf("wrong comment text: %q", comment.CommentText)
	}
}

func TestReview_AddComment_ReviewNotFound(t *testing.T) {
	wpRepo, revRepo, _, uc := buildReviewUC()

	wpID, caller := seedInReview(t, wpRepo, revRepo, domain.RoleSeniorAuditor)

	// RolePartner has no review entry seeded for this WP in fakeReviewRepo
	// Actually SubmitForReview seeds all 3, so try a non-existent WP
	_, err := uc.AddComment(context.Background(), uuid.New(), domain.RoleSeniorAuditor, usecase.CommentAddRequest{
		CommentText: "comment",
	}, caller, "127.0.0.1")
	if err == nil {
		t.Error("expected error for non-existent WP review")
	}
	_ = wpID
}

func TestReview_ListComments_HappyPath(t *testing.T) {
	wpRepo, revRepo, _, uc := buildReviewUC()

	wpID, caller := seedInReview(t, wpRepo, revRepo, domain.RoleSeniorAuditor)

	// Add 2 comments for senior auditor role
	_, _ = uc.AddComment(context.Background(), wpID, domain.RoleSeniorAuditor, usecase.CommentAddRequest{
		CommentText: "First comment",
	}, caller, "127.0.0.1")
	_, _ = uc.AddComment(context.Background(), wpID, domain.RoleSeniorAuditor, usecase.CommentAddRequest{
		CommentText: "Second comment",
	}, caller, "127.0.0.1")

	result, err := uc.ListComments(context.Background(), wpID, domain.RoleSeniorAuditor, usecase.CommentListRequest{Page: 1, Size: 20})
	if err != nil {
		t.Fatalf("ListComments: %v", err)
	}
	if len(result.Data) != 2 {
		t.Errorf("want 2 comments, got %d", len(result.Data))
	}
}

func TestReview_ListComments_NoReview_ReturnsError(t *testing.T) {
	_, _, _, uc := buildReviewUC()

	_, err := uc.ListComments(context.Background(), uuid.New(), domain.RoleSeniorAuditor, usecase.CommentListRequest{Page: 1, Size: 20})
	if err == nil {
		t.Error("expected error for non-existent review")
	}
}
