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

	orghandler "github.com/mdh/erp-audit/api/internal/org/handler"
	orgrepo "github.com/mdh/erp-audit/api/internal/org/repository"
	orgusecase "github.com/mdh/erp-audit/api/internal/org/usecase"

	"github.com/mdh/erp-audit/api/pkg/ws"

	enghandler "github.com/mdh/erp-audit/api/internal/engagement/handler"
	engrepo "github.com/mdh/erp-audit/api/internal/engagement/repository"
	engusecase "github.com/mdh/erp-audit/api/internal/engagement/usecase"

	tshandler "github.com/mdh/erp-audit/api/internal/timesheet/handler"
	tsrepo "github.com/mdh/erp-audit/api/internal/timesheet/repository"
	tsusecase "github.com/mdh/erp-audit/api/internal/timesheet/usecase"

	billinghandler "github.com/mdh/erp-audit/api/internal/billing/handler"
	billingrepo "github.com/mdh/erp-audit/api/internal/billing/repository"
	billingusecase "github.com/mdh/erp-audit/api/internal/billing/usecase"

	wphandler "github.com/mdh/erp-audit/api/internal/workingpaper/handler"
	wprepo "github.com/mdh/erp-audit/api/internal/workingpaper/repository"
	wpusecase "github.com/mdh/erp-audit/api/internal/workingpaper/usecase"

	commhandler "github.com/mdh/erp-audit/api/internal/commission/handler"
	commrepo "github.com/mdh/erp-audit/api/internal/commission/repository"
	commusecase "github.com/mdh/erp-audit/api/internal/commission/usecase"
	commworker "github.com/mdh/erp-audit/api/internal/commission/worker"

	taxhandler "github.com/mdh/erp-audit/api/internal/tax/handler"
	taxrepo "github.com/mdh/erp-audit/api/internal/tax/repository"
	taxusecase "github.com/mdh/erp-audit/api/internal/tax/usecase"
	taxworker "github.com/mdh/erp-audit/api/internal/tax/worker"

	"github.com/mdh/erp-audit/api/pkg/notification"
	"github.com/mdh/erp-audit/api/pkg/push"

	reportinghandler "github.com/mdh/erp-audit/api/internal/reporting/handler"
	reportingrepo "github.com/mdh/erp-audit/api/internal/reporting/repository"
	reportingusecase "github.com/mdh/erp-audit/api/internal/reporting/usecase"
	reportingworker "github.com/mdh/erp-audit/api/internal/reporting/worker"

	"github.com/mdh/erp-audit/api/pkg/crypto"
	"github.com/mdh/erp-audit/api/pkg/distlock"
	"github.com/mdh/erp-audit/api/pkg/metrics"
	"github.com/mdh/erp-audit/api/pkg/outbox"
	"github.com/mdh/erp-audit/api/pkg/ratelimit"
	"github.com/mdh/erp-audit/api/pkg/worker"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

func main() {
	// ── Config ───────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	if err := config.ValidateProductionConfig(cfg); err != nil {
		log.Fatalf("%v", err)
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

	// ── Redis ────────────────────────────────────────────────────────────────
	redisOpts, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		l.Fatal("failed to parse Redis URL", logger.Error(err))
	}
	redisClient := redis.NewClient(redisOpts)
	defer redisClient.Close()

	locker := distlock.New(redisClient, 30*time.Second)
	outboxPublisher := outbox.New(db.Pool)

	// ── Asynq worker + outbox poller ─────────────────────────────────────────
	asynqRedisOpt := asynq.RedisClientOpt{
		Addr:     redisOpts.Addr,
		Password: redisOpts.Password,
		DB:       redisOpts.DB,
	}
	asynqClient := asynq.NewClient(asynqRedisOpt)
	defer asynqClient.Close()
	outboxPoller := outbox.NewPoller(db.Pool, asynqClient, 5*time.Second, 50)
	workerSrv := worker.New(worker.Config{RedisAddr: redisOpts.Addr, Concurrency: 10})

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

	auditLogRepo := authrepo.NewAuditLogRepo(db.Pool)
	listAuditLogsUC := authusecase.NewListAuditLogsUseCase(auditLogRepo)

	authH := authhandler.NewAuthHandler(loginUC, refreshUC, logoutUC)
	userH := authhandler.NewUserHandler(createUserUC, assignRoleUC, listUsersUC, updateUserUC, deleteUserUC)
	twoFAH := authhandler.NewTwoFAHandler(enable2FAUC, verifySetupUC, disable2FAUC, verify2FALoginUC, verifyBackupUC, regenBackupUC)
	auditH := authhandler.NewAuditHandler(listAuditLogsUC)

	// ── CRM module ────────────────────────────────────────────────────────────
	crmRepo := crmrepo.New(db.Pool)
	clientUC := crmusecase.NewClientUseCase(crmRepo, auditLogger)
	clientH := crmhandler.NewClientHandler(clientUC)
	contactRepo := crmrepo.NewContactRepo(db.Pool)
	contactUC := crmusecase.NewContactUseCase(contactRepo, auditLogger)
	contactH := crmhandler.NewContactHandler(contactUC)

	// ── HRM encryption — fail fast if key is missing or invalid ──────────────
	hrmCipher, err := crypto.NewAESGCM(cfg.HRM.EncryptionKey)
	if err != nil {
		log.Fatalf("failed to init HRM encryption: %v", err)
	}
	log.Println("HRM encryption service initialized")

	// ── HRM module ────────────────────────────────────────────────────────────
	hrmOrgRepo := hrmrepo.NewOrgRepo(db.Pool)
	hrmOrgUC := hrmusecase.NewOrgUseCase(hrmOrgRepo, auditLogger)
	hrmOrgH := hrmhandler.NewOrgHandler(hrmOrgUC)

	hrmEmpRepo := hrmrepo.NewEmployeeRepo(db.Pool)
	hrmDepRepo := hrmrepo.NewDependentRepo(db.Pool)
	hrmConRepo := hrmrepo.NewContractRepo(db.Pool)
	hrmSalRepo := hrmrepo.NewSalaryHistoryRepo(db.Pool)
	hrmEmpUC := hrmusecase.NewEmployeeUseCase(hrmEmpRepo, auditLogger)
	hrmDepUC := hrmusecase.NewDependentUseCase(hrmDepRepo, auditLogger)
	hrmConUC := hrmusecase.NewContractUseCase(hrmConRepo, hrmEmpRepo, auditLogger)
	hrmProfUC := hrmusecase.NewProfileUseCase(hrmEmpRepo, auditLogger)
	hrmSensUC := hrmusecase.NewSensitiveUseCase(hrmEmpRepo, hrmCipher, auditLogger)
	hrmSalUC := hrmusecase.NewSalaryHistoryUseCase(hrmSalRepo, hrmEmpRepo, auditLogger)
	hrmEmpH := hrmhandler.NewEmployeeHandler(hrmEmpUC)
	hrmDepH := hrmhandler.NewDependentHandler(hrmDepUC)
	hrmConH := hrmhandler.NewContractHandler(hrmConUC)
	hrmProfH := hrmhandler.NewProfileHandler(hrmProfUC)
	hrmSensH := hrmhandler.NewSensitiveHandler(hrmSensUC)
	hrmSalH := hrmhandler.NewSalaryHistoryHandler(hrmSalUC)

	hrmProvRepo := hrmrepo.NewProvisioningRepo(db.Pool)
	hrmOffRepo := hrmrepo.NewOffboardingRepo(db.Pool)
	hrmProvUC := hrmusecase.NewProvisioningUseCase(hrmProvRepo, hrmOffRepo, db.Pool, auditLogger)
	hrmProvH := hrmhandler.NewProvisioningHandler(hrmProvUC)

	// ── Org module (branches & departments) ───────────────────────────────────
	branchRepo := orgrepo.NewBranchRepo(db.Pool)
	deptRepo := orgrepo.NewDeptRepo(db.Pool)
	branchUC := orgusecase.NewBranchUseCase(branchRepo, auditLogger)
	deptUC := orgusecase.NewDepartmentUseCase(deptRepo, auditLogger)
	branchH := orghandler.NewBranchHandler(branchUC)
	deptH := orghandler.NewDepartmentHandler(deptUC)

	// ── WebSocket hub (created early so use cases can broadcast) ─────────────
	wsHub := ws.NewHub()
	go wsHub.Run()
	wsH := ws.NewHandler(wsHub, jwtSvc)

	// ── Engagement module ─────────────────────────────────────────────────────
	engRepo := engrepo.NewEngagementRepo(db.Pool)
	memberRepo := engrepo.NewMemberRepo(db.Pool)
	taskRepo := engrepo.NewTaskRepo(db.Pool)
	costRepo := engrepo.NewCostRepo(db.Pool)
	engagementUC := engusecase.NewEngagementUseCase(engRepo, auditLogger, wsHub)
	teamUC := engusecase.NewTeamUseCase(memberRepo, engRepo, auditLogger)
	taskUC := engusecase.NewTaskUseCase(taskRepo, engRepo, auditLogger)
	costUC := engusecase.NewCostUseCase(costRepo, engRepo, auditLogger)
	engH := enghandler.NewEngagementHandler(engagementUC)
	teamH := enghandler.NewTeamHandler(teamUC)
	taskH := enghandler.NewTaskHandler(taskUC)
	costH := enghandler.NewCostHandler(costUC)

	// ── Timesheet module ──────────────────────────────────────────────────────
	timesheetRepo := tsrepo.NewTimesheetRepo(db.Pool)
	entryRepo := tsrepo.NewEntryRepo(db.Pool)
	attendanceRepo := tsrepo.NewAttendanceRepo(db.Pool)
	timesheetUC := tsusecase.NewTimesheetUseCase(timesheetRepo, locker, auditLogger, outboxPublisher, wsHub)
	entryUC := tsusecase.NewEntryUseCase(entryRepo, timesheetRepo, auditLogger)
	attendanceUC := tsusecase.NewAttendanceUseCase(attendanceRepo, auditLogger)
	tsH := tshandler.NewTimesheetHandler(timesheetUC)
	entryH := tshandler.NewEntryHandler(entryUC)
	attendanceH := tshandler.NewAttendanceHandler(attendanceUC)

	// ── Billing module ────────────────────────────────────────────────────────
	billingInvRepo := billingrepo.NewInvoiceRepo(db.Pool)
	billingLineRepo := billingrepo.NewLineItemRepo(db.Pool)
	billingPayRepo := billingrepo.NewPaymentRepo(db.Pool)
	billingMemoRepo := billingrepo.NewMemoRepo(db.Pool)
	billingARRepo := billingrepo.NewARRepo(db.Pool)
	invoiceUC := billingusecase.NewInvoiceUseCase(billingInvRepo, billingLineRepo, auditLogger, outboxPublisher)
	paymentUC := billingusecase.NewPaymentUseCase(billingPayRepo, billingInvRepo, auditLogger, outboxPublisher)
	memoUC := billingusecase.NewMemoUseCase(billingMemoRepo, billingInvRepo, auditLogger, outboxPublisher)
	arUC := billingusecase.NewARUseCase(billingARRepo)
	engAdapter := billingusecase.NewEngagementAdapter(engRepo, memberRepo, costRepo)
	tsEntryAdapter := billingusecase.NewTimesheetEntryAdapter(entryRepo)
	generateUC := billingusecase.NewGenerateUseCase(billingInvRepo, billingLineRepo, engAdapter, tsEntryAdapter, auditLogger)
	invoiceH := billinghandler.NewInvoiceHandler(invoiceUC, generateUC)
	paymentH := billinghandler.NewPaymentHandler(paymentUC)
	memoH := billinghandler.NewMemoHandler(memoUC)
	arH := billinghandler.NewARHandler(arUC)
	billingReportRepo := billingrepo.NewReportRepo(db.Pool)
	reportUC := billingusecase.NewReportUseCase(billingReportRepo)
	reportH := billinghandler.NewReportHandler(reportUC)

	// ── Working Paper module ──────────────────────────────────────────────────
	wpWPRepo := wprepo.NewWPRepo(db.Pool)
	wpReviewRepo := wprepo.NewReviewRepo(db.Pool)
	wpCommentRepo := wprepo.NewCommentRepo(db.Pool)
	wpFolderRepo := wprepo.NewFolderRepo(db.Pool)
	wpTemplateRepo := wprepo.NewTemplateRepo(db.Pool)
	wpUC := wpusecase.NewWorkingPaperUseCase(wpWPRepo, wpReviewRepo, auditLogger)
	wpReviewUC := wpusecase.NewReviewUseCase(wpReviewRepo, wpCommentRepo, wpWPRepo, auditLogger)
	wpFolderUC := wpusecase.NewFolderUseCase(wpFolderRepo, auditLogger)
	wpTemplateUC := wpusecase.NewTemplateUseCase(wpTemplateRepo, wpWPRepo, auditLogger)
	wpH := wphandler.NewWPHandler(wpUC)
	wpReviewH := wphandler.NewReviewHandler(wpReviewUC)
	wpFolderH := wphandler.NewFolderHandler(wpFolderUC)
	wpTmplH := wphandler.NewTemplateHandler(wpTemplateUC)

	// ── Commission module ─────────────────────────────────────────────────────
	commPlanRepo := commrepo.NewPlanRepo(db.Pool)
	commECRepo := commrepo.NewEngCommissionRepo(db.Pool)
	commRecordRepo := commrepo.NewRecordRepo(db.Pool)
	commBillingReader := commrepo.NewBillingReader(db.Pool)
	commPlanUC := commusecase.NewPlanUseCase(commPlanRepo, auditLogger)
	commECUC := commusecase.NewEngCommissionUseCase(commECRepo, auditLogger)
	commAccrualUC := commusecase.NewAccrualUseCase(commECRepo, commRecordRepo, commBillingReader, auditLogger)
	commRecordUC := commusecase.NewRecordUseCase(commRecordRepo, auditLogger)
	commPlanH := commhandler.NewPlanHandler(commPlanUC)
	commECH := commhandler.NewEngCommissionHandler(commECUC)
	commRecordH := commhandler.NewRecordHandler(commRecordUC)

	// ── Tax Advisory module ───────────────────────────────────────────────────
	taxDeadlineRepo := taxrepo.NewDeadlineRepo(db.Pool)
	taxAdvisoryRepo := taxrepo.NewAdvisoryRepo(db.Pool)
	taxComplianceRepo := taxrepo.NewComplianceRepo(db.Pool)
	taxDeadlineUC := taxusecase.NewTaxDeadlineUseCase(taxDeadlineRepo, auditLogger)
	taxAdvisoryUC := taxusecase.NewAdvisoryUseCase(taxAdvisoryRepo, auditLogger)
	taxComplianceUC := taxusecase.NewComplianceUseCase(taxComplianceRepo)
	taxDeadlineH := taxhandler.NewDeadlineHandler(taxDeadlineUC)
	taxAdvisoryH := taxhandler.NewAdvisoryHandler(taxAdvisoryUC)
	taxComplianceH := taxhandler.NewComplianceHandler(taxComplianceUC)

	// ── Push notification module ───────────────────────────────────────────────
	pushRelay := push.NewRelay()
	pushDeviceRepo := push.NewDeviceRepo(db.Pool)
	pushDeviceUC := authusecase.NewPushDeviceUseCase(pushDeviceRepo)
	push2FAUC := authusecase.NewPush2FAUseCase(repo, repo, repo, jwtSvc, pushRelay)
	pushH := authhandler.NewPushHandler(pushDeviceUC, push2FAUC, pushRelay, pushDeviceRepo)
	pushNotifier := notification.New(pushDeviceRepo, pushRelay)

	// ── Reporting module ──────────────────────────────────────────────────────
	repRepo := reportingrepo.NewReportingRepo(db.Pool)
	repDashUC := reportingusecase.NewDashboardUseCase(repRepo)
	repReportUC := reportingusecase.NewReportUseCase(repRepo)
	repDashH := reportinghandler.NewDashboardHandler(repDashUC)
	repReportH := reportinghandler.NewReportHandler(repReportUC)

	// Register event handlers — timesheet push notifications
	workerSrv.Register(outbox.EventTimesheetSubmitted, worker.NewTimesheetSubmittedHandler(pushNotifier))
	workerSrv.Register(outbox.EventTimesheetApproved, worker.NewTimesheetApprovedHandler(pushNotifier))
	workerSrv.Register(outbox.EventTimesheetRejected, worker.NewTimesheetRejectedHandler(pushNotifier))
	workerSrv.Register(outbox.EventTimesheetLocked, worker.NewTimesheetLockedHandler(pushNotifier))
	workerSrv.Register(outbox.EventEngagementActivated, worker.NewEngagementActivatedHandler(pushNotifier))
	// Tax deadline reminder job
	workerSrv.Register(taxworker.TaskDeadlineReminder, taxworker.NewDeadlineReminderHandler(taxDeadlineUC))

	// Commission accrual event handlers
	workerSrv.Register(outbox.EventInvoiceIssued, commworker.NewInvoiceIssuedHandler(commAccrualUC))
	workerSrv.Register(outbox.EventInvoiceCancelled, commworker.NewInvoiceCancelledHandler(commAccrualUC))
	workerSrv.Register(outbox.EventPaymentReceived, commworker.NewPaymentReceivedHandler(commAccrualUC))
	workerSrv.Register(outbox.EventEngagementSettled, commworker.NewEngagementSettledHandler(commAccrualUC))
	// Reporting MV nightly refresh job
	workerSrv.Register(reportingworker.TaskRefreshMVs, reportingworker.NewMVRefreshHandler(repReportUC))

	// ── Router ────────────────────────────────────────────────────────────────
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger(l))
	r.Use(middleware.CORS([]string{"*"}))
	r.Use(ratelimit.New(redisClient).Middleware())
	r.Use(metrics.Middleware())
	r.Use(middleware.AuditIDMiddleware())

	// Prometheus metrics (internal — restrict via network policy or ingress in production)
	r.GET("/metrics", metrics.Handler())

	// Health check (public)
	r.GET("/health", func(c *gin.Context) {
		if err := db.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": "database unreachable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "version": "v1", "env": cfg.App.Env})
	})

	authMW := middleware.AuthMiddleware(jwtSvc)
	require2FA := middleware.Require2FA()

	v1 := r.Group("/api/v1")
	// Apply 2FA enforcement globally; it only rejects FIRM_PARTNER/SUPER_ADMIN without 2fa_verified.
	v1.Use(require2FA)
	authhandler.RegisterRoutes(v1, authH, userH, twoFAH, pushH, auditH, authMW)
	crmhandler.RegisterRoutes(v1, clientH, contactH, authMW)
	hrmhandler.RegisterRoutes(v1, hrmOrgH, hrmEmpH, hrmDepH, hrmConH, hrmProfH, hrmSensH, hrmSalH, hrmProvH, authMW)
	orghandler.RegisterRoutes(v1, branchH, deptH, authMW)
	enghandler.RegisterRoutes(v1, engH, teamH, taskH, costH, authMW)
	tshandler.RegisterRoutes(v1, tsH, entryH, attendanceH, authMW)
	billinghandler.RegisterRoutes(v1, invoiceH, paymentH, memoH, arH, reportH, authMW)
	wphandler.RegisterRoutes(v1, wpH, wpReviewH, wpTmplH, wpFolderH, authMW)
	commhandler.RegisterRoutes(v1, commPlanH, commECH, commRecordH, authMW)
	taxhandler.RegisterRoutes(v1, taxDeadlineH, taxAdvisoryH, taxComplianceH, authMW)
	reportinghandler.RegisterRoutes(v1, repDashH, repReportH, authMW)
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

	// ── Start background services ─────────────────────────────────────────────
	bgCtx, bgCancel := context.WithCancel(context.Background())
	defer bgCancel()

	go outboxPoller.Run(bgCtx)
	go func() {
		if err := workerSrv.Run(); err != nil {
			l.Info("worker stopped: " + err.Error())
		}
	}()

	go func() {
		l.Info("server starting on :" + cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Fatal("server error", logger.Error(err))
		}
	}()

	<-quit
	l.Info("shutting down server...")

	bgCancel()
	workerSrv.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	wsHub.Stop()

	if err := srv.Shutdown(ctx); err != nil {
		l.Fatal("server shutdown failed", logger.Error(err))
	}
	l.Info("server exited")
}
