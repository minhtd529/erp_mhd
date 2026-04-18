package usecase_test

// Integration tests for the Engagement domain workflow.
// Requires a live PostgreSQL database; skips automatically when DATABASE_URL is not set.
// Run manually: DATABASE_URL=postgres://erp:erp@localhost:5433/erp_audit?sslmode=disable go test ./internal/engagement/usecase/... -v -run Integration

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/engagement/domain"
	engagerepo "github.com/mdh/erp-audit/api/internal/engagement/repository"
	"github.com/mdh/erp-audit/api/internal/engagement/usecase"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

func setupIntegrationDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set — skipping engagement integration tests")
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
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'engagements')`).
		Scan(&exists); err != nil || !exists {
		t.Skip("engagements table missing — run `make migrate-up` first")
	}
	t.Cleanup(pool.Close)
	return pool
}

// seedClient inserts a minimal client row and returns its id.
func seedClient(t *testing.T, pool *pgxpool.Pool, callerID uuid.UUID) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := pool.QueryRow(context.Background(), `
		INSERT INTO clients (client_name, client_type, created_by, updated_by)
		VALUES ('Integration Test Client', 'CORPORATE', $1, $1)
		RETURNING id`, callerID).Scan(&id)
	if err != nil {
		t.Fatalf("seed client: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `UPDATE clients SET is_deleted=true WHERE id=$1`, id)
	})
	return id
}

// seedUser inserts a minimal user row (idempotent by email) and returns its id.
func seedUser(t *testing.T, pool *pgxpool.Pool, email string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := pool.QueryRow(context.Background(), `
		INSERT INTO users (email, hashed_password, full_name)
		VALUES ($1, '$2a$10$dummyhashforintegrationtest', 'Integration Partner')
		ON CONFLICT (email) DO UPDATE SET email=EXCLUDED.email
		RETURNING id`, email).Scan(&id)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return id
}

func TestEngagementIntegration_Lifecycle(t *testing.T) {
	pool := setupIntegrationDB(t)
	ctx := context.Background()

	callerID := uuid.New()
	// Seed a system user as the caller (needed for FK on audit_logs).
	callerID = seedUser(t, pool, "integ_engagement_partner@example.com")
	clientID := seedClient(t, pool, callerID)

	auditLog := audit.New(pool)
	repo := engagerepo.NewEngagementRepo(pool)
	uc := usecase.NewEngagementUseCase(repo, auditLog, nil)

	svcType := domain.ServiceAudit
	feeType := domain.FeeFixed

	t.Run("create → status is DRAFT", func(t *testing.T) {
		resp, err := uc.Create(ctx, usecase.EngagementCreateRequest{
			ClientID:    clientID,
			ServiceType: svcType,
			FeeType:     feeType,
			FeeAmount:   50_000_000,
		}, callerID, "127.0.0.1")
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if resp.Status != domain.StatusDraft {
			t.Errorf("want DRAFT, got %s", resp.Status)
		}
		t.Cleanup(func() {
			pool.Exec(ctx, `UPDATE engagements SET is_deleted=true WHERE id=$1`, resp.ID)
		})
	})

	t.Run("activate without partner → PARTNER_NOT_ASSIGNED", func(t *testing.T) {
		resp, err := uc.Create(ctx, usecase.EngagementCreateRequest{
			ClientID:    clientID,
			ServiceType: svcType,
			FeeType:     feeType,
			FeeAmount:   1,
		}, callerID, "127.0.0.1")
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		t.Cleanup(func() {
			pool.Exec(ctx, `UPDATE engagements SET is_deleted=true WHERE id=$1`, resp.ID)
		})

		_, err = uc.Activate(ctx, resp.ID, callerID, "127.0.0.1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != domain.ErrPartnerNotAssigned {
			t.Errorf("want ErrPartnerNotAssigned, got %v", err)
		}
	})

	t.Run("DRAFT → ACTIVE → COMPLETED full lifecycle", func(t *testing.T) {
		resp, err := uc.Create(ctx, usecase.EngagementCreateRequest{
			ClientID:    clientID,
			ServiceType: svcType,
			FeeType:     feeType,
			FeeAmount:   10_000_000,
			PartnerID:   &callerID,
		}, callerID, "127.0.0.1")
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		id := resp.ID
		t.Cleanup(func() {
			pool.Exec(ctx, `UPDATE engagements SET is_deleted=true WHERE id=$1`, id)
		})

		// Activate
		activated, err := uc.Activate(ctx, id, callerID, "127.0.0.1")
		if err != nil {
			t.Fatalf("Activate: %v", err)
		}
		if activated.Status != domain.StatusActive {
			t.Errorf("want ACTIVE, got %s", activated.Status)
		}

		// Complete
		completed, err := uc.Complete(ctx, id, callerID, "127.0.0.1")
		if err != nil {
			t.Fatalf("Complete: %v", err)
		}
		if completed.Status != domain.StatusCompleted {
			t.Errorf("want COMPLETED, got %s", completed.Status)
		}
	})

	t.Run("invalid transition DRAFT → COMPLETED", func(t *testing.T) {
		resp, err := uc.Create(ctx, usecase.EngagementCreateRequest{
			ClientID:    clientID,
			ServiceType: svcType,
			FeeType:     feeType,
			FeeAmount:   1,
		}, callerID, "127.0.0.1")
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		t.Cleanup(func() {
			pool.Exec(ctx, `UPDATE engagements SET is_deleted=true WHERE id=$1`, resp.ID)
		})

		_, err = uc.Complete(ctx, resp.ID, callerID, "127.0.0.1")
		if err != domain.ErrInvalidStateTransition {
			t.Errorf("want ErrInvalidStateTransition, got %v", err)
		}
	})
}
