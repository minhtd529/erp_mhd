package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/config"
	"github.com/mdh/erp-audit/api/pkg/database"
	"github.com/mdh/erp-audit/api/pkg/logger"
	"github.com/mdh/erp-audit/api/pkg/middleware"

	authhandler "github.com/mdh/erp-audit/api/internal/global/auth/handler"
	authrepo "github.com/mdh/erp-audit/api/internal/global/auth/repository"
	authusecase "github.com/mdh/erp-audit/api/internal/global/auth/usecase"

	crmhandler "github.com/mdh/erp-audit/api/internal/crm/handler"
	crmrepo "github.com/mdh/erp-audit/api/internal/crm/repository"
	crmusecase "github.com/mdh/erp-audit/api/internal/crm/usecase"

	hrmhandler "github.com/mdh/erp-audit/api/internal/hrm/handler"
	hrmrepo "github.com/mdh/erp-audit/api/internal/hrm/repository"
	hrmusecase "github.com/mdh/erp-audit/api/internal/hrm/usecase"

	"github.com/mdh/erp-audit/api/pkg/ws"
)

func main() {
	// ── Config ───────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// ── Logger ───────────────────────────────────────────────────────────────
	l, err := logger.New(cfg.App.Env)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer l.Sync() //nolint:errcheck

	// ── Database ─────────────────────────────────────────────────────────────
	db, err := database.New(cfg.Database)
	if err != nil {
		l.Fatal("failed to connect to database", logger.Error(err))
	}
	defer db.Close()

	if err := database.Migrate(cfg.Database.URL); err != nil {
		l.Fatal("failed to run migrations", logger.Error(err))
	}

	// ── Shared services ───────────────────────────────────────────────────────
	jwtSvc := pkgauth.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)
	auditLogger := audit.New(db.Pool)

	// ── Auth / Global user module ─────────────────────────────────────────────
	repo := authrepo.New(db.Pool)
	twoFACfg := cfg.TwoFA

	loginUC := authusecase.NewLoginUseCase(repo, repo, repo, repo, jwtSvc, auditLogger)
	refreshUC := authusecase.NewRefreshTokenUseCase(repo, repo, jwtSvc)
	logoutUC := authusecase.NewLogoutUseCase(repo)
	createUserUC := authusecase.NewCreateUserUseCase(repo, repo, auditLogger)
	assignRoleUC := authusecase.NewAssignRoleUseCase(repo, repo, auditLogger)
	listUsersUC := authusecase.NewListUsersUseCase(repo, repo)
	updateUserUC := authusecase.NewUpdateUserUseCase(repo, auditLogger)
	deleteUserUC := authusecase.NewDeleteUserUseCase(repo, auditLogger)

	enable2FAUC := authusecase.NewEnable2FAUseCase(repo, repo, twoFACfg.EncryptionKey, twoFACfg.Issuer)
	verifySetupUC := authusecase.NewVerifySetupUseCase(repo, twoFACfg.EncryptionKey, auditLogger)
	disable2FAUC := authusecase.NewDisable2FAUseCase(repo, repo, repo, auditLogger)
	verify2FALoginUC := authusecase.NewVerify2FALoginUseCase(
		repo, repo, repo, jwtSvc,
		twoFACfg.EncryptionKey,
		twoFACfg.MaxAttempts,
		twoFACfg.TrustDeviceDays,
		twoFACfg.MaxTrustedDevices,
		auditLogger,
	)
	verifyBackupUC := authusecase.NewVerifyBackupCodeUseCase(repo, repo, repo, jwtSvc, auditLogger)
	regenBackupUC := authusecase.NewRegenBackupCodesUseCase(repo, repo, auditLogger)

	authH := authhandler.NewAuthHandler(loginUC, refreshUC, logoutUC)
	userH := authhandler.NewUserHandler(createUserUC, assignRoleUC, listUsersUC, updateUserUC, deleteUserUC)
	twoFAH := authhandler.NewTwoFAHandler(enable2FAUC, verifySetupUC, disable2FAUC, verify2FALoginUC, verifyBackupUC, regenBackupUC)

	// ── CRM module ────────────────────────────────────────────────────────────
	crmRepo := crmrepo.New(db.Pool)
	clientUC := crmusecase.NewClientUseCase(crmRepo, auditLogger)
	clientH := crmhandler.NewClientHandler(clientUC)

	// ── HRM module ────────────────────────────────────────────────────────────
	hrmRepo := hrmrepo.New(db.Pool)
	employeeUC := hrmusecase.NewEmployeeUseCase(hrmRepo, auditLogger, cfg.HRM.BankEncryptionKey)
	employeeH := hrmhandler.NewEmployeeHandler(employeeUC)

	// ── WebSocket hub ─────────────────────────────────────────────────────────
	wsHub := ws.NewHub()
	go wsHub.Run()
	wsH := ws.NewHandler(wsHub, jwtSvc)

	// ── Router ────────────────────────────────────────────────────────────────
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger(l))
	r.Use(middleware.CORS([]string{"*"}))

	// Health check (public)
	r.GET("/health", func(c *gin.Context) {
		if err := db.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": "database unreachable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "version": "v1", "env": cfg.App.Env})
	})

	authMW := middleware.AuthMiddleware(jwtSvc)

	v1 := r.Group("/api/v1")
	authhandler.RegisterRoutes(v1, authH, userH, twoFAH, authMW)
	crmhandler.RegisterRoutes(v1, clientH, authMW)
	hrmhandler.RegisterRoutes(v1, employeeH, authMW)
	ws.RegisterRoutes(v1, wsH) // GET /api/v1/events/stream — public upgrade, auth via ?token=

	// ── HTTP server ───────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		l.Info("server starting on :" + cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Fatal("server error", logger.Error(err))
		}
	}()

	<-quit
	l.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	wsHub.Stop()

	if err := srv.Shutdown(ctx); err != nil {
		l.Fatal("server shutdown failed", logger.Error(err))
	}
	l.Info("server exited")
}
