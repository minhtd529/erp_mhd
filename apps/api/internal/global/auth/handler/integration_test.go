package handler_test

// Integration test for the /login endpoint.
// Requires a live PostgreSQL database; skips automatically when DATABASE_URL is not set.
// Run manually: DATABASE_URL=postgres://erp:erp@localhost:5433/erp_audit?sslmode=disable go test ./internal/global/auth/handler/... -v -run Integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/middleware"

	authhandler "github.com/mdh/erp-audit/api/internal/global/auth/handler"
	authrepo "github.com/mdh/erp-audit/api/internal/global/auth/repository"
	authusecase "github.com/mdh/erp-audit/api/internal/global/auth/usecase"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set — skipping integration tests")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("ping test db: %v", err)
	}

	// Migrations are expected to have been applied by `make migrate-up` before running tests.
	// We verify the schema by checking a known table rather than re-running migrations.
	var exists bool
	err = pool.QueryRow(context.Background(),
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'refresh_tokens')`).
		Scan(&exists)
	if err != nil || !exists {
		t.Skip("refresh_tokens table missing — run `make migrate-up` before integration tests")
	}

	t.Cleanup(pool.Close)
	return pool
}

func setupTestRouter(pool *pgxpool.Pool) *gin.Engine {
	gin.SetMode(gin.TestMode)

	jwtSvc := pkgauth.NewJWTService("test-secret-32chars-minimum-ok!", 15, 7)
	auditLog := audit.New(pool)

	repo := authrepo.New(pool)

	loginUC := authusecase.NewLoginUseCase(repo, repo, repo, repo, jwtSvc, auditLog)
	refreshUC := authusecase.NewRefreshTokenUseCase(repo, repo, jwtSvc)
	logoutUC := authusecase.NewLogoutUseCase(repo)
	createUserUC := authusecase.NewCreateUserUseCase(repo, repo, auditLog)
	assignRoleUC := authusecase.NewAssignRoleUseCase(repo, repo, auditLog)

	const testEncKey = "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"
	enable2FAUC := authusecase.NewEnable2FAUseCase(repo, repo, testEncKey, "ERP Audit Test")
	verifySetupUC := authusecase.NewVerifySetupUseCase(repo, testEncKey, auditLog)
	disable2FAUC := authusecase.NewDisable2FAUseCase(repo, repo, repo, auditLog)
	verify2FAUC := authusecase.NewVerify2FALoginUseCase(repo, repo, repo, jwtSvc, testEncKey, 5, 30, 5, auditLog)
	verifyBackupUC := authusecase.NewVerifyBackupCodeUseCase(repo, repo, repo, jwtSvc, auditLog)
	regenBackupUC := authusecase.NewRegenBackupCodesUseCase(repo, repo, auditLog)

	listUsersUC := authusecase.NewListUsersUseCase(repo, repo)
	updateUserUC := authusecase.NewUpdateUserUseCase(repo, auditLog)
	deleteUserUC := authusecase.NewDeleteUserUseCase(repo, auditLog)

	authH := authhandler.NewAuthHandler(loginUC, refreshUC, logoutUC)
	userH := authhandler.NewUserHandler(createUserUC, assignRoleUC, listUsersUC, updateUserUC, deleteUserUC)
	twoFAH := authhandler.NewTwoFAHandler(enable2FAUC, verifySetupUC, disable2FAUC, verify2FAUC, verifyBackupUC, regenBackupUC)
	authMW := middleware.AuthMiddleware(jwtSvc)

	r := gin.New()
	r.Use(gin.Recovery())
	v1 := r.Group("/api/v1")
	authhandler.RegisterRoutes(v1, authH, userH, twoFAH, authMW)

	return r
}

func postJSON(t *testing.T, r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ─── Integration tests ───────────────────────────────────────────────────────

func TestLoginIntegration(t *testing.T) {
	pool := setupTestDB(t)
	r := setupTestRouter(pool)

	// Seed a test user (idempotent — ignore duplicate-email errors)
	const (
		testEmail    = "integ_test_user@example.com"
		testPassword = "TestPassword123!"
	)

	// Use CreateUser endpoint to seed
	seedBody := map[string]any{
		"email":     testEmail,
		"password":  testPassword,
		"full_name": "Integration Test User",
		"role_code": "AUDIT_STAFF",
	}

	// Seed via DB directly to avoid needing an admin token
	hashed, err := pkgauth.HashPassword(testPassword)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	_, err = pool.Exec(context.Background(), `
		INSERT INTO users (email, hashed_password, full_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO NOTHING
	`, testEmail, hashed, "Integration Test User")
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	_ = seedBody // used above for comment context

	t.Run("valid credentials → 200 with tokens", func(t *testing.T) {
		w := postJSON(t, r, "/api/v1/auth/login", map[string]any{
			"email":    testEmail,
			"password": testPassword,
		})

		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("parse response: %v", err)
		}
		if resp["access_token"] == "" || resp["access_token"] == nil {
			t.Error("expected non-empty access_token")
		}
		if resp["refresh_token"] == "" || resp["refresh_token"] == nil {
			t.Error("expected non-empty refresh_token")
		}
	})

	t.Run("wrong password → 401 INVALID_CREDENTIALS", func(t *testing.T) {
		w := postJSON(t, r, "/api/v1/auth/login", map[string]any{
			"email":    testEmail,
			"password": "WrongPassword!",
		})

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d: %s", w.Code, w.Body.String())
		}

		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["error"] != "INVALID_CREDENTIALS" {
			t.Errorf("want INVALID_CREDENTIALS error code, got %v", resp["error"])
		}
	})

	t.Run("non-existent email → 401 INVALID_CREDENTIALS", func(t *testing.T) {
		w := postJSON(t, r, "/api/v1/auth/login", map[string]any{
			"email":    "nobody@nowhere.com",
			"password": "Irrelevant!",
		})

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("malformed body → 400", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login",
			bytes.NewReader([]byte(`{not json`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})

	t.Run("refresh flow → 200 new access token", func(t *testing.T) {
		// Login first
		loginW := postJSON(t, r, "/api/v1/auth/login", map[string]any{
			"email":    testEmail,
			"password": testPassword,
		})
		if loginW.Code != http.StatusOK {
			t.Skipf("login failed (%d), skipping refresh test", loginW.Code)
		}

		var loginResp map[string]any
		_ = json.Unmarshal(loginW.Body.Bytes(), &loginResp)
		rt, _ := loginResp["refresh_token"].(string)

		// Now refresh
		refreshW := postJSON(t, r, "/api/v1/auth/refresh", map[string]any{
			"refresh_token": rt,
		})
		if refreshW.Code != http.StatusOK {
			t.Fatalf("want 200 on refresh, got %d: %s", refreshW.Code, refreshW.Body.String())
		}
		var refreshResp map[string]any
		_ = json.Unmarshal(refreshW.Body.Bytes(), &refreshResp)
		if refreshResp["access_token"] == nil || refreshResp["access_token"] == "" {
			t.Error("expected non-empty access_token in refresh response")
		}
	})
}
