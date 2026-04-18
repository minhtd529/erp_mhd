// Package repository provides the PostgreSQL implementation of the WorkingPaper domain.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/workingpaper/domain"
)

// ─── WorkingPaper ─────────────────────────────────────────────────────────────

type WPRepo struct{ pool *pgxpool.Pool }

func NewWPRepo(pool *pgxpool.Pool) *WPRepo { return &WPRepo{pool: pool} }

const wpCols = `id, engagement_id, folder_id, document_type, title, status,
	file_id, snapshot_data, is_deleted, created_at, updated_at, created_by, updated_by`

func (r *WPRepo) Create(ctx context.Context, p domain.CreateWPParams) (*domain.WorkingPaper, error) {
	const q = `
		INSERT INTO working_papers (engagement_id, folder_id, document_type, title, file_id, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$6)
		RETURNING ` + wpCols

	wp, err := scanWP(r.pool.QueryRow(ctx, q,
		p.EngagementID, p.FolderID, string(p.DocumentType), p.Title, p.FileID, p.CreatedBy,
	))
	if err != nil {
		return nil, fmt.Errorf("wp.Create: %w", err)
	}
	return wp, nil
}

func (r *WPRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.WorkingPaper, error) {
	q := `SELECT ` + wpCols + ` FROM working_papers WHERE id = $1 AND is_deleted = false`
	wp, err := scanWP(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrWorkingPaperNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("wp.FindByID: %w", err)
	}
	return wp, nil
}

func (r *WPRepo) Update(ctx context.Context, p domain.UpdateWPParams) (*domain.WorkingPaper, error) {
	const q = `
		UPDATE working_papers SET title=$2, folder_id=$3, file_id=$4, updated_by=$5, updated_at=NOW()
		WHERE id=$1 AND is_deleted=false AND status='DRAFT'
		RETURNING ` + wpCols

	wp, err := scanWP(r.pool.QueryRow(ctx, q, p.ID, p.Title, p.FolderID, p.FileID, p.UpdatedBy))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrWorkingPaperNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("wp.Update: %w", err)
	}
	return wp, nil
}

func (r *WPRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.WPStatus, updatedBy uuid.UUID) (*domain.WorkingPaper, error) {
	const q = `
		UPDATE working_papers SET status=$2, updated_by=$3, updated_at=NOW()
		WHERE id=$1 AND is_deleted=false
		RETURNING ` + wpCols

	wp, err := scanWP(r.pool.QueryRow(ctx, q, id, string(status), updatedBy))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrWorkingPaperNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("wp.UpdateStatus: %w", err)
	}
	return wp, nil
}

func (r *WPRepo) Finalize(ctx context.Context, p domain.FinalizeWPParams) (*domain.WorkingPaper, error) {
	const q = `
		UPDATE working_papers
		SET status='FINALIZED', snapshot_data=$2, updated_by=$3, updated_at=NOW()
		WHERE id=$1 AND is_deleted=false
		RETURNING ` + wpCols

	wp, err := scanWP(r.pool.QueryRow(ctx, q, p.ID, p.SnapshotData, p.UpdatedBy))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrWorkingPaperNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("wp.Finalize: %w", err)
	}
	return wp, nil
}

func (r *WPRepo) SoftDelete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	const q = `
		UPDATE working_papers SET is_deleted=true, updated_by=$2, updated_at=NOW()
		WHERE id=$1 AND is_deleted=false AND status='DRAFT'`
	tag, err := r.pool.Exec(ctx, q, id, deletedBy)
	if err != nil {
		return fmt.Errorf("wp.SoftDelete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrWorkingPaperNotFound
	}
	return nil
}

func (r *WPRepo) List(ctx context.Context, f domain.ListWPFilter) ([]*domain.WorkingPaper, int64, error) {
	offset := (f.Page - 1) * f.Size
	args := []any{f.EngagementID}
	where := "WHERE is_deleted=false AND engagement_id=$1"
	idx := 2

	if f.Status != "" {
		where += fmt.Sprintf(" AND status=$%d", idx)
		args = append(args, string(f.Status))
		idx++
	}

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM working_papers "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("wp.List count: %w", err)
	}

	args = append(args, f.Size, offset)
	dataQ := fmt.Sprintf("SELECT %s FROM working_papers %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		wpCols, where, idx, idx+1)

	rows, err := r.pool.Query(ctx, dataQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("wp.List query: %w", err)
	}
	defer rows.Close()

	var list []*domain.WorkingPaper
	for rows.Next() {
		wp, err := scanWP(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("wp.List scan: %w", err)
		}
		list = append(list, wp)
	}
	if list == nil {
		list = []*domain.WorkingPaper{}
	}
	return list, total, rows.Err()
}

func (r *WPRepo) ListPendingReview(ctx context.Context, role domain.ReviewerRole, page, size int) ([]*domain.WorkingPaper, int64, error) {
	offset := (page - 1) * size
	const where = `
		WHERE wp.is_deleted=false
		  AND wp.status IN ('IN_REVIEW','COMMENTED')
		  AND EXISTS (
		        SELECT 1 FROM working_paper_reviews r
		        WHERE r.working_paper_id=wp.id
		          AND r.reviewer_role=$1
		          AND r.review_status='PENDING'
		  )`

	var total int64
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM working_papers wp`+where, string(role),
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("wp.ListPendingReview count: %w", err)
	}

	dataQ := fmt.Sprintf(`SELECT wp.%s FROM working_papers wp%s ORDER BY wp.created_at DESC LIMIT $2 OFFSET $3`,
		wpCols, where)
	rows, err := r.pool.Query(ctx, dataQ, string(role), size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("wp.ListPendingReview query: %w", err)
	}
	defer rows.Close()

	var list []*domain.WorkingPaper
	for rows.Next() {
		wp, err := scanWP(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("wp.ListPendingReview scan: %w", err)
		}
		list = append(list, wp)
	}
	if list == nil {
		list = []*domain.WorkingPaper{}
	}
	return list, total, rows.Err()
}

// ─── Review ───────────────────────────────────────────────────────────────────

type ReviewRepo struct{ pool *pgxpool.Pool }

func NewReviewRepo(pool *pgxpool.Pool) *ReviewRepo { return &ReviewRepo{pool: pool} }

const reviewCols = `id, working_paper_id, reviewer_role, review_status, review_date, reviewed_by, created_at`

func (r *ReviewRepo) Create(ctx context.Context, p domain.CreateReviewParams) (*domain.WorkingPaperReview, error) {
	const q = `
		INSERT INTO working_paper_reviews (working_paper_id, reviewer_role)
		VALUES ($1, $2)
		RETURNING ` + reviewCols

	rev, err := scanReview(r.pool.QueryRow(ctx, q, p.WorkingPaperID, string(p.ReviewerRole)))
	if err != nil {
		return nil, fmt.Errorf("review.Create: %w", err)
	}
	return rev, nil
}

func (r *ReviewRepo) FindByWPAndRole(ctx context.Context, wpID uuid.UUID, role domain.ReviewerRole) (*domain.WorkingPaperReview, error) {
	q := `SELECT ` + reviewCols + ` FROM working_paper_reviews WHERE working_paper_id=$1 AND reviewer_role=$2`
	rev, err := scanReview(r.pool.QueryRow(ctx, q, wpID, string(role)))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrReviewNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("review.FindByWPAndRole: %w", err)
	}
	return rev, nil
}

func (r *ReviewRepo) Approve(ctx context.Context, p domain.ApproveReviewParams) (*domain.WorkingPaperReview, error) {
	now := time.Now()
	const q = `
		UPDATE working_paper_reviews
		SET review_status='APPROVED', review_date=$3, reviewed_by=$4
		WHERE working_paper_id=$1 AND reviewer_role=$2
		RETURNING ` + reviewCols

	rev, err := scanReview(r.pool.QueryRow(ctx, q, p.WorkingPaperID, string(p.ReviewerRole), now, p.ReviewerID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrReviewNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("review.Approve: %w", err)
	}
	return rev, nil
}

func (r *ReviewRepo) RequestChanges(ctx context.Context, p domain.ApproveReviewParams) (*domain.WorkingPaperReview, error) {
	now := time.Now()
	const q = `
		UPDATE working_paper_reviews
		SET review_status='REJECTED', review_date=$3, reviewed_by=$4
		WHERE working_paper_id=$1 AND reviewer_role=$2
		RETURNING ` + reviewCols

	rev, err := scanReview(r.pool.QueryRow(ctx, q, p.WorkingPaperID, string(p.ReviewerRole), now, p.ReviewerID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrReviewNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("review.RequestChanges: %w", err)
	}
	return rev, nil
}

func (r *ReviewRepo) ListByWP(ctx context.Context, wpID uuid.UUID) ([]*domain.WorkingPaperReview, error) {
	q := `SELECT ` + reviewCols + ` FROM working_paper_reviews WHERE working_paper_id=$1 ORDER BY created_at`
	rows, err := r.pool.Query(ctx, q, wpID)
	if err != nil {
		return nil, fmt.Errorf("review.ListByWP: %w", err)
	}
	defer rows.Close()
	var list []*domain.WorkingPaperReview
	for rows.Next() {
		rev, err := scanReview(rows)
		if err != nil {
			return nil, fmt.Errorf("review.ListByWP scan: %w", err)
		}
		list = append(list, rev)
	}
	if list == nil {
		list = []*domain.WorkingPaperReview{}
	}
	return list, rows.Err()
}

func (r *ReviewRepo) CountUnresolved(ctx context.Context, wpID uuid.UUID) (int, error) {
	const q = `
		SELECT COUNT(*) FROM working_paper_comments c
		JOIN working_paper_reviews r ON r.id = c.review_id
		WHERE r.working_paper_id = $1 AND c.issue_status = 'OPEN'`
	var count int
	if err := r.pool.QueryRow(ctx, q, wpID).Scan(&count); err != nil {
		return 0, fmt.Errorf("review.CountUnresolved: %w", err)
	}
	return count, nil
}

// ─── Comment ─────────────────────────────────────────────────────────────────

type CommentRepo struct{ pool *pgxpool.Pool }

func NewCommentRepo(pool *pgxpool.Pool) *CommentRepo { return &CommentRepo{pool: pool} }

func (r *CommentRepo) Add(ctx context.Context, p domain.AddCommentParams) (*domain.WorkingPaperComment, error) {
	const q = `
		INSERT INTO working_paper_comments (review_id, comment_text, created_by)
		VALUES ($1,$2,$3)
		RETURNING id, review_id, comment_text, issue_status, raised_at, resolved_at, created_by`

	var c domain.WorkingPaperComment
	var status string
	err := r.pool.QueryRow(ctx, q, p.ReviewID, p.CommentText, p.CreatedBy).Scan(
		&c.ID, &c.ReviewID, &c.CommentText, &status, &c.RaisedAt, &c.ResolvedAt, &c.CreatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("comment.Add: %w", err)
	}
	c.IssueStatus = domain.IssueStatus(status)
	return &c, nil
}

func (r *CommentRepo) Resolve(ctx context.Context, id uuid.UUID) (*domain.WorkingPaperComment, error) {
	now := time.Now()
	const q = `
		UPDATE working_paper_comments SET issue_status='RESOLVED', resolved_at=$2
		WHERE id=$1
		RETURNING id, review_id, comment_text, issue_status, raised_at, resolved_at, created_by`

	var c domain.WorkingPaperComment
	var status string
	err := r.pool.QueryRow(ctx, q, id, now).Scan(
		&c.ID, &c.ReviewID, &c.CommentText, &status, &c.RaisedAt, &c.ResolvedAt, &c.CreatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrCommentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("comment.Resolve: %w", err)
	}
	c.IssueStatus = domain.IssueStatus(status)
	return &c, nil
}

func (r *CommentRepo) ListByReview(ctx context.Context, reviewID uuid.UUID) ([]*domain.WorkingPaperComment, error) {
	const q = `SELECT id, review_id, comment_text, issue_status, raised_at, resolved_at, created_by
		FROM working_paper_comments WHERE review_id=$1 ORDER BY raised_at`
	rows, err := r.pool.Query(ctx, q, reviewID)
	if err != nil {
		return nil, fmt.Errorf("comment.List: %w", err)
	}
	defer rows.Close()
	var list []*domain.WorkingPaperComment
	for rows.Next() {
		var c domain.WorkingPaperComment
		var status string
		if err := rows.Scan(&c.ID, &c.ReviewID, &c.CommentText, &status, &c.RaisedAt, &c.ResolvedAt, &c.CreatedBy); err != nil {
			return nil, fmt.Errorf("comment.List scan: %w", err)
		}
		c.IssueStatus = domain.IssueStatus(status)
		list = append(list, &c)
	}
	if list == nil {
		list = []*domain.WorkingPaperComment{}
	}
	return list, rows.Err()
}

// ─── Folder ───────────────────────────────────────────────────────────────────

type FolderRepo struct{ pool *pgxpool.Pool }

func NewFolderRepo(pool *pgxpool.Pool) *FolderRepo { return &FolderRepo{pool: pool} }

func (r *FolderRepo) Create(ctx context.Context, p domain.CreateFolderParams) (*domain.WorkingPaperFolder, error) {
	const q = `
		INSERT INTO working_paper_folders (engagement_id, folder_name, created_by)
		VALUES ($1,$2,$3)
		RETURNING id, engagement_id, folder_name, created_at, created_by`

	var f domain.WorkingPaperFolder
	err := r.pool.QueryRow(ctx, q, p.EngagementID, p.FolderName, p.CreatedBy).Scan(
		&f.ID, &f.EngagementID, &f.FolderName, &f.CreatedAt, &f.CreatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("folder.Create: %w", err)
	}
	return &f, nil
}

func (r *FolderRepo) ListByEngagement(ctx context.Context, engagementID uuid.UUID) ([]*domain.WorkingPaperFolder, error) {
	const q = `SELECT id, engagement_id, folder_name, created_at, created_by
		FROM working_paper_folders WHERE engagement_id=$1 ORDER BY created_at`
	rows, err := r.pool.Query(ctx, q, engagementID)
	if err != nil {
		return nil, fmt.Errorf("folder.List: %w", err)
	}
	defer rows.Close()
	var list []*domain.WorkingPaperFolder
	for rows.Next() {
		var f domain.WorkingPaperFolder
		if err := rows.Scan(&f.ID, &f.EngagementID, &f.FolderName, &f.CreatedAt, &f.CreatedBy); err != nil {
			return nil, fmt.Errorf("folder.List scan: %w", err)
		}
		list = append(list, &f)
	}
	if list == nil {
		list = []*domain.WorkingPaperFolder{}
	}
	return list, rows.Err()
}

// ─── Template ─────────────────────────────────────────────────────────────────

type TemplateRepo struct{ pool *pgxpool.Pool }

func NewTemplateRepo(pool *pgxpool.Pool) *TemplateRepo { return &TemplateRepo{pool: pool} }

const tmplCols = `id, template_type, title, version, content, vsa_compliant, is_active, created_at, updated_at, created_by, updated_by`

func (r *TemplateRepo) Create(ctx context.Context, p domain.CreateTemplateParams) (*domain.AuditTemplate, error) {
	const q = `
		INSERT INTO audit_templates (template_type, title, version, content, vsa_compliant, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$6)
		RETURNING ` + tmplCols

	content := p.Content
	if content == nil {
		content = []byte("{}")
	}
	tmpl, err := scanTemplate(r.pool.QueryRow(ctx, q,
		p.TemplateType, p.Title, p.Version, content, p.VSACompliant, p.CreatedBy,
	))
	if err != nil {
		return nil, fmt.Errorf("template.Create: %w", err)
	}
	return tmpl, nil
}

func (r *TemplateRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.AuditTemplate, error) {
	q := `SELECT ` + tmplCols + ` FROM audit_templates WHERE id=$1`
	tmpl, err := scanTemplate(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTemplateNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("template.FindByID: %w", err)
	}
	return tmpl, nil
}

func (r *TemplateRepo) Update(ctx context.Context, p domain.UpdateTemplateParams) (*domain.AuditTemplate, error) {
	const q = `
		UPDATE audit_templates SET title=$2, content=$3, vsa_compliant=$4, updated_by=$5, updated_at=NOW()
		WHERE id=$1
		RETURNING ` + tmplCols

	tmpl, err := scanTemplate(r.pool.QueryRow(ctx, q,
		p.ID, p.Title, p.Content, p.VSACompliant, p.UpdatedBy,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTemplateNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("template.Update: %w", err)
	}
	return tmpl, nil
}

func (r *TemplateRepo) Retire(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) error {
	const q = `UPDATE audit_templates SET is_active=false, updated_by=$2, updated_at=NOW() WHERE id=$1`
	tag, err := r.pool.Exec(ctx, q, id, updatedBy)
	if err != nil {
		return fmt.Errorf("template.Retire: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTemplateNotFound
	}
	return nil
}

func (r *TemplateRepo) List(ctx context.Context, activeOnly bool, page, size int) ([]*domain.AuditTemplate, int64, error) {
	offset := (page - 1) * size
	where := ""
	if activeOnly {
		where = "WHERE is_active=true"
	}
	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM audit_templates "+where).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("template.List count: %w", err)
	}
	q := fmt.Sprintf("SELECT %s FROM audit_templates %s ORDER BY created_at DESC LIMIT $1 OFFSET $2", tmplCols, where)
	rows, err := r.pool.Query(ctx, q, size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("template.List: %w", err)
	}
	defer rows.Close()
	var list []*domain.AuditTemplate
	for rows.Next() {
		tmpl, err := scanTemplate(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("template.List scan: %w", err)
		}
		list = append(list, tmpl)
	}
	if list == nil {
		list = []*domain.AuditTemplate{}
	}
	return list, total, rows.Err()
}

// ─── helpers ──────────────────────────────────────────────────────────────────

type scanner interface{ Scan(dest ...any) error }

func scanWP(row scanner) (*domain.WorkingPaper, error) {
	var wp domain.WorkingPaper
	var docType, status string
	err := row.Scan(
		&wp.ID, &wp.EngagementID, &wp.FolderID, &docType, &wp.Title, &status,
		&wp.FileID, &wp.SnapshotData,
		&wp.IsDeleted, &wp.CreatedAt, &wp.UpdatedAt, &wp.CreatedBy, &wp.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	wp.DocumentType = domain.DocumentType(docType)
	wp.Status = domain.WPStatus(status)
	return &wp, nil
}

func scanReview(row scanner) (*domain.WorkingPaperReview, error) {
	var r domain.WorkingPaperReview
	var role, status string
	err := row.Scan(
		&r.ID, &r.WorkingPaperID, &role, &status, &r.ReviewDate, &r.ReviewedBy, &r.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	r.ReviewerRole = domain.ReviewerRole(role)
	r.ReviewStatus = domain.ReviewStatus(status)
	return &r, nil
}

func scanTemplate(row scanner) (*domain.AuditTemplate, error) {
	var t domain.AuditTemplate
	err := row.Scan(
		&t.ID, &t.TemplateType, &t.Title, &t.Version, &t.Content,
		&t.VSACompliant, &t.IsActive, &t.CreatedAt, &t.UpdatedAt, &t.CreatedBy, &t.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
