package usecase_test

// Integration tests for the WorkingPaper domain workflow.
// Requires a live PostgreSQL database; skips automatically when DATABASE_URL is not set.
// Run manually: DATABASE_URL=postgres://erp:erp@localhost:5433/erp_audit?sslmode=disable go test ./internal/workingpaper/usecase/... -v -run Integration

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	engagedomain "github.com/mdh/erp-audit/api/internal/engagement/domain"
	engagerepo "github.com/mdh/erp-audit/api/internal/engagement/repository"
	"github.com/mdh/erp-audit/api/internal/workingpaper/domain"
	wprepo "github.com/mdh/erp-audit/api/internal/workingpaper/repository"
	"github.com/mdh/erp-audit/api/internal/workingpaper/usecase"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

func setupWPIntegrationDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set — skipping working paper integration tests")
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("ping db: %v", err)
	}
	var exists bool
	if err := pool.QueryRow(context.Background(),
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'working_papers')`).
		Scan(&exists); err != nil || !exists {
		t.Skip("working_papers table missing — run `make migrate-up` first")
	}
	t.Cleanup(pool.Close)
	return pool
}

func seedWPUser(t *testing.T, pool *pgxpool.Pool, email string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := pool.QueryRow(context.Background(), `
		INSERT INTO users (email, hashed_password, full_name)
		VALUES ($1, '$2a$10$dummyhashforintegrationtest', 'Integration WP User')
		ON CONFLICT (email) DO UPDATE SET email=EXCLUDED.email
		RETURNING id`, email).Scan(&id)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return id
}

func seedWPClient(t *testing.T, pool *pgxpool.Pool, callerID uuid.UUID) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := pool.QueryRow(context.Background(), `
		INSERT INTO clients (client_name, client_type, created_by, updated_by)
		VALUES ('WP Integration Client', 'CORPORATE', $1, $1)
		RETURNING id`, callerID).Scan(&id)
	if err != nil {
		t.Fatalf("seed client: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `UPDATE clients SET is_deleted=true WHERE id=$1`, id)
	})
	return id
}

func seedEngagementForWP(t *testing.T, pool *pgxpool.Pool, clientID, callerID uuid.UUID) uuid.UUID {
	t.Helper()
	engageRepo := engagerepo.NewEngagementRepo(pool)
	auditLog := audit.New(pool)
	_ = auditLog

	e, err := engageRepo.Create(context.Background(), engagedomain.CreateEngagementParams{
		ClientID:    clientID,
		ServiceType: engagedomain.ServiceAudit,
		FeeType:     engagedomain.FeeFixed,
		FeeAmount:   10_000_000,
		CreatedBy:   callerID,
	})
	if err != nil {
		t.Fatalf("seed engagement: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `UPDATE engagements SET is_deleted=true WHERE id=$1`, e.ID)
	})
	return e.ID
}

func TestWPIntegration_Lifecycle(t *testing.T) {
	pool := setupWPIntegrationDB(t)
	ctx := context.Background()

	callerID := seedWPUser(t, pool, "integ_wp_user@example.com")
	clientID := seedWPClient(t, pool, callerID)
	engagementID := seedEngagementForWP(t, pool, clientID, callerID)

	auditLog := audit.New(pool)
	wpRepo := wprepo.NewWPRepo(pool)
	reviewRepo := wprepo.NewReviewRepo(pool)
	uc := usecase.NewWorkingPaperUseCase(wpRepo, reviewRepo, auditLog)

	t.Run("create → status is DRAFT", func(t *testing.T) {
		wp, err := uc.Create(ctx, engagementID, usecase.WPCreateRequest{
			DocumentType: domain.DocProcedures,
			Title:        "Integration Test WP",
		}, callerID, "127.0.0.1")
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if wp.Status != domain.WPStatusDraft {
			t.Errorf("want DRAFT, got %s", wp.Status)
		}
		t.Cleanup(func() {
			pool.Exec(ctx, `UPDATE working_papers SET is_deleted=true WHERE id=$1`, wp.ID)
		})
	})

	t.Run("update allowed in DRAFT", func(t *testing.T) {
		wp, err := uc.Create(ctx, engagementID, usecase.WPCreateRequest{
			DocumentType: domain.DocEvidence,
			Title:        "Draft WP for update",
		}, callerID, "127.0.0.1")
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		t.Cleanup(func() {
			pool.Exec(ctx, `UPDATE working_papers SET is_deleted=true WHERE id=$1`, wp.ID)
		})

		updated, err := uc.Update(ctx, wp.ID, usecase.WPUpdateRequest{
			Title: "Updated Draft Title",
		}, callerID, "127.0.0.1")
		if err != nil {
			t.Fatalf("Update DRAFT: %v", err)
		}
		if updated.Title != "Updated Draft Title" {
			t.Errorf("title not updated, got %q", updated.Title)
		}
	})

	t.Run("DRAFT → IN_REVIEW via SubmitForReview", func(t *testing.T) {
		wp, err := uc.Create(ctx, engagementID, usecase.WPCreateRequest{
			DocumentType: domain.DocAnalysis,
			Title:        "Submit for review WP",
		}, callerID, "127.0.0.1")
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		t.Cleanup(func() {
			pool.Exec(ctx, `UPDATE working_papers SET is_deleted=true WHERE id=$1`, wp.ID)
		})

		reviewed, err := uc.SubmitForReview(ctx, wp.ID, callerID, "127.0.0.1")
		if err != nil {
			t.Fatalf("SubmitForReview: %v", err)
		}
		if reviewed.Status != domain.WPStatusInReview {
			t.Errorf("want IN_REVIEW, got %s", reviewed.Status)
		}
	})

	t.Run("update blocked after SubmitForReview (IN_REVIEW)", func(t *testing.T) {
		wp, err := uc.Create(ctx, engagementID, usecase.WPCreateRequest{
			DocumentType: domain.DocConclusion,
			Title:        "Immutability test WP",
		}, callerID, "127.0.0.1")
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		t.Cleanup(func() {
			pool.Exec(ctx, `UPDATE working_papers SET is_deleted=true WHERE id=$1`, wp.ID)
		})

		// Force status to FINALIZED directly to test immutability guard.
		_, err = pool.Exec(ctx, `UPDATE working_papers SET status='FINALIZED' WHERE id=$1`, wp.ID)
		if err != nil {
			t.Fatalf("force finalized: %v", err)
		}

		_, err = uc.Update(ctx, wp.ID, usecase.WPUpdateRequest{
			Title: "Should be rejected",
		}, callerID, "127.0.0.1")
		if err != domain.ErrWorkingPaperNotEditable {
			t.Errorf("want ErrWorkingPaperNotEditable, got %v", err)
		}
	})

	t.Run("update blocked after SIGNED_OFF", func(t *testing.T) {
		wp, err := uc.Create(ctx, engagementID, usecase.WPCreateRequest{
			DocumentType: domain.DocMgmtLetter,
			Title:        "Signed off WP",
		}, callerID, "127.0.0.1")
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		t.Cleanup(func() {
			pool.Exec(ctx, `UPDATE working_papers SET is_deleted=true WHERE id=$1`, wp.ID)
		})

		_, err = pool.Exec(ctx, `UPDATE working_papers SET status='SIGNED_OFF' WHERE id=$1`, wp.ID)
		if err != nil {
			t.Fatalf("force signed_off: %v", err)
		}

		_, err = uc.Update(ctx, wp.ID, usecase.WPUpdateRequest{
			Title: "Should also be rejected",
		}, callerID, "127.0.0.1")
		if err != domain.ErrWorkingPaperNotEditable {
			t.Errorf("want ErrWorkingPaperNotEditable, got %v", err)
		}
	})
}
