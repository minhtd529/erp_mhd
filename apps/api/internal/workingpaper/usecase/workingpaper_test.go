package usecase_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/workingpaper/domain"
	"github.com/mdh/erp-audit/api/internal/workingpaper/usecase"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Fakes ─────────────────────────────────────────────────────────────────────

type fakeWPRepo struct {
	papers map[uuid.UUID]*domain.WorkingPaper
}

func newFakeWPRepo() *fakeWPRepo { return &fakeWPRepo{papers: map[uuid.UUID]*domain.WorkingPaper{}} }

func (r *fakeWPRepo) Create(_ context.Context, p domain.CreateWPParams) (*domain.WorkingPaper, error) {
	wp := &domain.WorkingPaper{
		ID:           uuid.New(),
		EngagementID: p.EngagementID,
		FolderID:     p.FolderID,
		DocumentType: p.DocumentType,
		Title:        p.Title,
		Status:       domain.WPStatusDraft,
		FileID:       p.FileID,
		CreatedBy:    p.CreatedBy,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	r.papers[wp.ID] = wp
	return wp, nil
}

func (r *fakeWPRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.WorkingPaper, error) {
	wp, ok := r.papers[id]
	if !ok {
		return nil, domain.ErrWorkingPaperNotFound
	}
	return wp, nil
}

func (r *fakeWPRepo) Update(_ context.Context, p domain.UpdateWPParams) (*domain.WorkingPaper, error) {
	wp, ok := r.papers[p.ID]
	if !ok {
		return nil, domain.ErrWorkingPaperNotFound
	}
	if p.Title != "" {
		wp.Title = p.Title
	}
	wp.FolderID = p.FolderID
	wp.FileID = p.FileID
	wp.UpdatedAt = time.Now()
	return wp, nil
}

func (r *fakeWPRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.WPStatus, _ uuid.UUID) (*domain.WorkingPaper, error) {
	wp, ok := r.papers[id]
	if !ok {
		return nil, domain.ErrWorkingPaperNotFound
	}
	wp.Status = status
	return wp, nil
}

func (r *fakeWPRepo) SoftDelete(_ context.Context, id uuid.UUID, _ uuid.UUID) error {
	if _, ok := r.papers[id]; !ok {
		return domain.ErrWorkingPaperNotFound
	}
	delete(r.papers, id)
	return nil
}

func (r *fakeWPRepo) Finalize(_ context.Context, p domain.FinalizeWPParams) (*domain.WorkingPaper, error) {
	wp, ok := r.papers[p.ID]
	if !ok {
		return nil, domain.ErrWorkingPaperNotFound
	}
	wp.Status = domain.WPStatusFinalized
	wp.SnapshotData = p.SnapshotData
	return wp, nil
}

func (r *fakeWPRepo) List(_ context.Context, _ domain.ListWPFilter) ([]*domain.WorkingPaper, int64, error) {
	var result []*domain.WorkingPaper
	for _, wp := range r.papers {
		result = append(result, wp)
	}
	return result, int64(len(result)), nil
}

func (r *fakeWPRepo) ListPendingReview(_ context.Context, _ domain.ReviewerRole, _, _ int) ([]*domain.WorkingPaper, int64, error) {
	return []*domain.WorkingPaper{}, 0, nil
}

type fakeReviewRepo struct {
	reviews map[uuid.UUID]*domain.WorkingPaperReview // keyed by review ID
}

func newFakeReviewRepo() *fakeReviewRepo {
	return &fakeReviewRepo{reviews: map[uuid.UUID]*domain.WorkingPaperReview{}}
}

func (r *fakeReviewRepo) Create(_ context.Context, p domain.CreateReviewParams) (*domain.WorkingPaperReview, error) {
	rev := &domain.WorkingPaperReview{
		ID:             uuid.New(),
		WorkingPaperID: p.WorkingPaperID,
		ReviewerRole:   p.ReviewerRole,
		ReviewStatus:   domain.ReviewPending,
		CreatedAt:      time.Now(),
	}
	r.reviews[rev.ID] = rev
	return rev, nil
}

func (r *fakeReviewRepo) ListByWP(_ context.Context, wpID uuid.UUID) ([]*domain.WorkingPaperReview, error) {
	var result []*domain.WorkingPaperReview
	for _, rev := range r.reviews {
		if rev.WorkingPaperID == wpID {
			result = append(result, rev)
		}
	}
	return result, nil
}

func (r *fakeReviewRepo) CountUnresolved(_ context.Context, _ uuid.UUID) (int, error) { return 0, nil }

func (r *fakeReviewRepo) Approve(_ context.Context, p domain.ApproveReviewParams) (*domain.WorkingPaperReview, error) {
	for _, rev := range r.reviews {
		if rev.WorkingPaperID == p.WorkingPaperID && rev.ReviewerRole == p.ReviewerRole {
			rev.ReviewStatus = domain.ReviewApproved
			now := time.Now()
			rev.ReviewDate = &now
			rev.ReviewedBy = &p.ReviewerID
			return rev, nil
		}
	}
	return nil, domain.ErrReviewNotFound
}

func (r *fakeReviewRepo) RequestChanges(_ context.Context, p domain.ApproveReviewParams) (*domain.WorkingPaperReview, error) {
	for _, rev := range r.reviews {
		if rev.WorkingPaperID == p.WorkingPaperID && rev.ReviewerRole == p.ReviewerRole {
			rev.ReviewStatus = domain.ReviewRejected
			now := time.Now()
			rev.ReviewDate = &now
			rev.ReviewedBy = &p.ReviewerID
			return rev, nil
		}
	}
	return nil, domain.ErrReviewNotFound
}

func (r *fakeReviewRepo) FindByWPAndRole(_ context.Context, wpID uuid.UUID, role domain.ReviewerRole) (*domain.WorkingPaperReview, error) {
	for _, rev := range r.reviews {
		if rev.WorkingPaperID == wpID && rev.ReviewerRole == role {
			return rev, nil
		}
	}
	return nil, domain.ErrReviewNotFound
}

// approveAll sets all reviews for a WP to APPROVED — used to satisfy Finalize preconditions.
func (r *fakeReviewRepo) approveAll(wpID uuid.UUID) {
	now := time.Now()
	caller := uuid.New()
	for _, rev := range r.reviews {
		if rev.WorkingPaperID == wpID {
			rev.ReviewStatus = domain.ReviewApproved
			rev.ReviewDate = &now
			rev.ReviewedBy = &caller
		}
	}
}

type fakeCommentRepo struct {
	comments map[uuid.UUID]*domain.WorkingPaperComment
}

func newFakeCommentRepo() *fakeCommentRepo {
	return &fakeCommentRepo{comments: map[uuid.UUID]*domain.WorkingPaperComment{}}
}

func (r *fakeCommentRepo) Add(_ context.Context, p domain.AddCommentParams) (*domain.WorkingPaperComment, error) {
	c := &domain.WorkingPaperComment{
		ID: uuid.New(), ReviewID: p.ReviewID,
		CommentText: p.CommentText, IssueStatus: domain.IssueOpen,
		CreatedBy: p.CreatedBy,
	}
	r.comments[c.ID] = c
	return c, nil
}

func (r *fakeCommentRepo) Resolve(_ context.Context, id uuid.UUID) (*domain.WorkingPaperComment, error) {
	c, ok := r.comments[id]
	if !ok {
		return nil, domain.ErrCommentNotFound
	}
	c.IssueStatus = domain.IssueResolved
	now := time.Now()
	c.ResolvedAt = &now
	return c, nil
}

func (r *fakeCommentRepo) ListByReview(_ context.Context, reviewID uuid.UUID) ([]*domain.WorkingPaperComment, error) {
	var result []*domain.WorkingPaperComment
	for _, c := range r.comments {
		if c.ReviewID == reviewID {
			result = append(result, c)
		}
	}
	return result, nil
}

func newAuditLogger() *audit.Logger {
	return nil // nil Logger → no-op (Log checks l == nil)
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestWP_Create(t *testing.T) {
	wpRepo := newFakeWPRepo()
	revRepo := newFakeReviewRepo()
	uc := usecase.NewWorkingPaperUseCase(wpRepo, revRepo, newAuditLogger())

	caller := uuid.New()
	engID := uuid.New()
	resp, err := uc.Create(context.Background(), engID, usecase.WPCreateRequest{
		DocumentType: domain.DocProcedures,
		Title:        "Test WP",
	}, caller, "127.0.0.1")

	require.NoError(t, err)
	assert.Equal(t, "Test WP", resp.Title)
	assert.Equal(t, domain.WPStatusDraft, resp.Status)
	assert.Equal(t, engID, resp.EngagementID)
}

func TestWP_SubmitForReview_seeds_review_chain(t *testing.T) {
	wpRepo := newFakeWPRepo()
	revRepo := newFakeReviewRepo()
	uc := usecase.NewWorkingPaperUseCase(wpRepo, revRepo, newAuditLogger())

	caller := uuid.New()
	engID := uuid.New()
	created, _ := uc.Create(context.Background(), engID, usecase.WPCreateRequest{
		DocumentType: domain.DocProcedures,
		Title:        "WP for review",
	}, caller, "127.0.0.1")

	resp, err := uc.SubmitForReview(context.Background(), created.ID, caller, "127.0.0.1")
	require.NoError(t, err)
	assert.Equal(t, domain.WPStatusInReview, resp.Status)

	reviews, _ := revRepo.ListByWP(context.Background(), created.ID)
	assert.Len(t, reviews, 3)
}

func TestWP_SubmitForReview_rejects_non_draft(t *testing.T) {
	wpRepo := newFakeWPRepo()
	revRepo := newFakeReviewRepo()
	uc := usecase.NewWorkingPaperUseCase(wpRepo, revRepo, newAuditLogger())

	caller := uuid.New()
	engID := uuid.New()
	created, _ := uc.Create(context.Background(), engID, usecase.WPCreateRequest{
		DocumentType: domain.DocProcedures,
		Title:        "WP",
	}, caller, "127.0.0.1")
	uc.SubmitForReview(context.Background(), created.ID, caller, "127.0.0.1")

	// second submit should fail
	_, err := uc.SubmitForReview(context.Background(), created.ID, caller, "127.0.0.1")
	assert.ErrorIs(t, err, domain.ErrInvalidStateTransition)
}

func TestWP_Finalize_requires_all_approved(t *testing.T) {
	wpRepo := newFakeWPRepo()
	revRepo := newFakeReviewRepo()
	uc := usecase.NewWorkingPaperUseCase(wpRepo, revRepo, newAuditLogger())

	caller := uuid.New()
	engID := uuid.New()
	created, _ := uc.Create(context.Background(), engID, usecase.WPCreateRequest{
		DocumentType: domain.DocProcedures,
		Title:        "WP",
	}, caller, "127.0.0.1")
	uc.SubmitForReview(context.Background(), created.ID, caller, "127.0.0.1")

	// not all approved → error
	_, err := uc.Finalize(context.Background(), created.ID, caller, "127.0.0.1")
	assert.ErrorIs(t, err, domain.ErrReviewChainIncomplete)
}

func TestWP_Finalize_captures_snapshot(t *testing.T) {
	wpRepo := newFakeWPRepo()
	revRepo := newFakeReviewRepo()
	uc := usecase.NewWorkingPaperUseCase(wpRepo, revRepo, newAuditLogger())

	caller := uuid.New()
	engID := uuid.New()
	created, _ := uc.Create(context.Background(), engID, usecase.WPCreateRequest{
		DocumentType: domain.DocProcedures,
		Title:        "WP Snapshot",
	}, caller, "127.0.0.1")
	uc.SubmitForReview(context.Background(), created.ID, caller, "127.0.0.1")
	revRepo.approveAll(created.ID)

	resp, err := uc.Finalize(context.Background(), created.ID, caller, "127.0.0.1")
	require.NoError(t, err)
	assert.Equal(t, domain.WPStatusFinalized, resp.Status)
	assert.NotEmpty(t, resp.SnapshotData)

	var snap map[string]any
	require.NoError(t, json.Unmarshal(resp.SnapshotData, &snap))
	assert.Equal(t, "WP Snapshot", snap["title"])
}

func TestWP_SignOff_requires_finalized(t *testing.T) {
	wpRepo := newFakeWPRepo()
	revRepo := newFakeReviewRepo()
	uc := usecase.NewWorkingPaperUseCase(wpRepo, revRepo, newAuditLogger())

	caller := uuid.New()
	engID := uuid.New()
	created, _ := uc.Create(context.Background(), engID, usecase.WPCreateRequest{
		DocumentType: domain.DocProcedures,
		Title:        "WP",
	}, caller, "127.0.0.1")

	_, err := uc.SignOff(context.Background(), created.ID, caller, "127.0.0.1")
	assert.ErrorIs(t, err, domain.ErrInvalidStateTransition)
}

func TestWP_Delete_notfound(t *testing.T) {
	wpRepo := newFakeWPRepo()
	revRepo := newFakeReviewRepo()
	uc := usecase.NewWorkingPaperUseCase(wpRepo, revRepo, newAuditLogger())

	err := uc.Delete(context.Background(), uuid.New(), uuid.New(), "127.0.0.1")
	assert.ErrorIs(t, err, domain.ErrWorkingPaperNotFound)
}

func TestWP_PendingReview_returns_paginated(t *testing.T) {
	wpRepo := newFakeWPRepo()
	revRepo := newFakeReviewRepo()
	uc := usecase.NewWorkingPaperUseCase(wpRepo, revRepo, newAuditLogger())

	result, err := uc.PendingReview(context.Background(), usecase.PendingReviewRequest{
		Role: domain.RoleSeniorAuditor,
		Page: 1, Size: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(0), result.Total)
}

func TestReview_ResolveComment(t *testing.T) {
	wpRepo := newFakeWPRepo()
	revRepo := newFakeReviewRepo()
	commentRepo := newFakeCommentRepo()
	uc := usecase.NewReviewUseCase(revRepo, commentRepo, wpRepo, newAuditLogger())

	caller := uuid.New()
	// seed a comment
	commentID := uuid.New()
	commentRepo.comments[commentID] = &domain.WorkingPaperComment{
		ID: commentID, IssueStatus: domain.IssueOpen,
	}

	resp, err := uc.ResolveComment(context.Background(), commentID, caller, "127.0.0.1")
	require.NoError(t, err)
	assert.Equal(t, domain.IssueResolved, resp.IssueStatus)
}

func TestReview_ResolveComment_notfound(t *testing.T) {
	wpRepo := newFakeWPRepo()
	revRepo := newFakeReviewRepo()
	commentRepo := newFakeCommentRepo()
	uc := usecase.NewReviewUseCase(revRepo, commentRepo, wpRepo, newAuditLogger())

	_, err := uc.ResolveComment(context.Background(), uuid.New(), uuid.New(), "127.0.0.1")
	assert.ErrorIs(t, err, domain.ErrCommentNotFound)
}

func TestWP_Update_Immutable_AfterFinalized(t *testing.T) {
	t.Parallel()
	wpRepo := newFakeWPRepo()
	revRepo := newFakeReviewRepo()
	uc := usecase.NewWorkingPaperUseCase(wpRepo, revRepo, newAuditLogger())

	caller := uuid.New()
	engID := uuid.New()
	created, _ := uc.Create(context.Background(), engID, usecase.WPCreateRequest{
		DocumentType: domain.DocProcedures,
		Title:        "Immutable WP",
	}, caller, "127.0.0.1")

	// Transition to FINALIZED
	uc.SubmitForReview(context.Background(), created.ID, caller, "127.0.0.1")
	revRepo.approveAll(created.ID)
	_, err := uc.Finalize(context.Background(), created.ID, caller, "127.0.0.1")
	require.NoError(t, err)

	// Update on FINALIZED WP must be rejected
	_, err = uc.Update(context.Background(), created.ID, usecase.WPUpdateRequest{Title: "Modified"}, caller, "127.0.0.1")
	assert.ErrorIs(t, err, domain.ErrWorkingPaperNotEditable)
}

func TestWP_Update_Immutable_AfterSignedOff(t *testing.T) {
	t.Parallel()
	wpRepo := newFakeWPRepo()
	revRepo := newFakeReviewRepo()
	uc := usecase.NewWorkingPaperUseCase(wpRepo, revRepo, newAuditLogger())

	caller := uuid.New()
	engID := uuid.New()
	created, _ := uc.Create(context.Background(), engID, usecase.WPCreateRequest{
		DocumentType: domain.DocProcedures,
		Title:        "WP to sign off",
	}, caller, "127.0.0.1")

	// Drive through to SIGNED_OFF
	uc.SubmitForReview(context.Background(), created.ID, caller, "127.0.0.1")
	revRepo.approveAll(created.ID)
	uc.Finalize(context.Background(), created.ID, caller, "127.0.0.1")
	_, err := uc.SignOff(context.Background(), created.ID, caller, "127.0.0.1")
	require.NoError(t, err)

	// Update on SIGNED_OFF WP must be rejected
	_, err = uc.Update(context.Background(), created.ID, usecase.WPUpdateRequest{Title: "Tampered"}, caller, "127.0.0.1")
	assert.ErrorIs(t, err, domain.ErrWorkingPaperNotEditable)
}

func TestWP_Update_Allowed_InDraft(t *testing.T) {
	t.Parallel()
	wpRepo := newFakeWPRepo()
	revRepo := newFakeReviewRepo()
	uc := usecase.NewWorkingPaperUseCase(wpRepo, revRepo, newAuditLogger())

	caller := uuid.New()
	engID := uuid.New()
	created, _ := uc.Create(context.Background(), engID, usecase.WPCreateRequest{
		DocumentType: domain.DocProcedures,
		Title:        "Draft WP",
	}, caller, "127.0.0.1")

	resp, err := uc.Update(context.Background(), created.ID, usecase.WPUpdateRequest{Title: "Updated Draft"}, caller, "127.0.0.1")
	require.NoError(t, err)
	assert.Equal(t, "Updated Draft", resp.Title)
}

func TestWP_GetByID_HappyPath(t *testing.T) {
	wpRepo := newFakeWPRepo()
	revRepo := newFakeReviewRepo()
	uc := usecase.NewWorkingPaperUseCase(wpRepo, revRepo, newAuditLogger())

	caller := uuid.New()
	engID := uuid.New()
	created, _ := uc.Create(context.Background(), engID, usecase.WPCreateRequest{
		DocumentType: domain.DocEvidence,
		Title:        "GetByID test",
	}, caller, "127.0.0.1")

	resp, err := uc.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, resp.ID)
	assert.Equal(t, "GetByID test", resp.Title)
}

func TestWP_GetByID_NotFound(t *testing.T) {
	wpRepo := newFakeWPRepo()
	uc := usecase.NewWorkingPaperUseCase(wpRepo, newFakeReviewRepo(), newAuditLogger())

	_, err := uc.GetByID(context.Background(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrWorkingPaperNotFound)
}

func TestWP_List_ReturnsAll(t *testing.T) {
	wpRepo := newFakeWPRepo()
	uc := usecase.NewWorkingPaperUseCase(wpRepo, newFakeReviewRepo(), newAuditLogger())

	caller := uuid.New()
	engID := uuid.New()
	for i := 0; i < 3; i++ {
		_, _ = uc.Create(context.Background(), engID, usecase.WPCreateRequest{
			DocumentType: domain.DocProcedures,
			Title:        "WP",
		}, caller, "127.0.0.1")
	}

	result, err := uc.List(context.Background(), engID, usecase.WPListRequest{Page: 1, Size: 20})
	require.NoError(t, err)
	assert.Equal(t, 3, len(result.Data))
}
