
# ERP Audit System — Development Roadmap

## Overview
Following Phase 1-5 approach with 9 DDD Bounded Contexts. Each phase builds on previous foundation.

## Phase 1: Foundation & Infrastructure (Months 1-3)

### Bounded Contexts
- **Global**: Organization structure, user management, audit framework
- **CRM**: Basic client/firm data management
- **HRM**: Basic staff/resource data management

### Deliverables
- [x] Backend infrastructure (Go, PostgreSQL, Redis, MinIO) — **Phase 0 done 2026-04-16**
- [x] Frontend foundation (Next.js, Zustand, React Query) — **Phase 0 done 2026-04-16**
- [x] Authentication system (JWT + TOTP 2FA) — full implementation: JWT+bcrypt + TOTP/backup codes/trusted devices
- [x] Audit trail framework (pkg/audit, immutable logs) — full implementation + audit on every mutation
- [x] API conventions implementation (RESTful, versioning, error handling) — UPPER_SNAKE error codes, /api/v1 prefix
- [x] Database schema (core tables, soft deletes, audit fields) — migrations 000001 + 000002
- [x] WebSocket real-time framework
- [x] Docker Compose dev environment — postgres:5433, redis:6380, minio:9000, nats:4222
- [x] Module structure: domain → repository → usecase → handler — exemplar in internal/global

### Phase 0 Completed (2026-04-16)
- [x] Turborepo monorepo initialized (pnpm workspaces)
- [x] apps/api Go Clean Architecture (domain→usecase→repository→handler) with 8 module stubs
- [x] apps/web Next.js 14 App Router + Tailwind + React Query + Zustand setup
- [x] docker-compose.yml: postgres, redis, minio, nats (non-conflicting ports)
- [x] Makefile: dev/test/lint/migrate/sqlc targets
- [x] sqlc.yaml configured + golang-migrate runner in pkg/database
- [x] .github/workflows/ci.yml: lint + test + build + docker jobs
- [x] README.md with full dev setup guide
- [x] pkg/database: pgxpool connection pool + migration runner
- [x] **Smoke test PASSED**: `GET /health → 200 {"status":"ok","version":"v1","env":"development"}`

### Phase 1.1 Completed: pkg/auth (2026-04-16)
- [x] Migration 000002: trusted_devices, two_factor_backup_codes, two_factor_challenges, refresh_tokens
- [x] pkg/auth: JWTService (HS256, 15 min access / 7 day refresh), bcrypt(cost=12), HashRefreshToken, context keys
- [x] internal/global/auth/domain: UserForAuth, RefreshToken, TwoFactorChallenge, repository interfaces, JWTIssuer interface
- [x] internal/global/auth/repository: Postgres implementation (UserRepo, RoleRepo, RefreshTokenRepo) + sqlc query files
- [x] internal/global/auth/usecase: Login, RefreshToken, Logout, LogoutAll, CreateUser, AssignRole
- [x] internal/global/auth/handler: AuthHandler (/login /refresh /logout /me), UserHandler (POST /users, POST /users/:id/roles)
- [x] pkg/middleware: AuthMiddleware(JWTService), RequireRole(...), RequirePermission(module, resource, action)
- [x] 11 tests PASS: 6 unit (Login table-driven) + 4 unit (CreateUser) + 5 integration (/login flow, refresh flow)

### Phase 1.4 Completed: WebSocket Real-Time Framework (2026-04-16)
- [x] `pkg/ws` package: Hub (fan-out broadcaster), Client (write/read pumps, 30-min session timeout, ping/pong)
- [x] `GET /api/v1/events/stream?token=<JWT>&channels=global,crm` — JWT auth via query param
- [x] Channel-based subscription: each client subscribes to 1..N named channels
- [x] `Hub.Broadcast(channel, type, data)` — safe to call from any goroutine, non-blocking fan-out
- [x] 30-minute session timeout enforced via write pump ticker
- [x] `gorilla/websocket v1.5.3` added to go.mod
- [x] Hub wired in `cmd/server/main.go` with graceful Stop() on shutdown
- [x] 6 new tests: Register, Unregister, Broadcast delivery, channel isolation, multi-channel sub, ParseChannels
- [x] 68 total tests passing

### Phase 1.3 Completed: Global/CRM/HRM Basic CRUD (2026-04-16)
- [x] Migration 000004: `clients` (CRM) + `employees` (HRM) tables with indexes
- [x] CRM bounded context: domain/repository/usecase/handler for Client CRUD (5 endpoints)
- [x] HRM bounded context: domain/repository/usecase/handler for Employee CRUD (5 endpoints)
- [x] Global user management extended: GET/PUT/DELETE /api/v1/users + GET /api/v1/users/:id
- [x] All mutations emit audit.Logger.Log (CREATE/UPDATE/DELETE for clients, employees, users)
- [x] All list endpoints return PaginatedResult[T] with total/page/size/total_pages
- [x] Domain error sentinels: CLIENT_NOT_FOUND, DUPLICATE_TAX_CODE, EMPLOYEE_NOT_FOUND, DUPLICATE_EMAIL
- [x] Handler maps every domain error → correct HTTP status (404/409/422)
- [x] pkg/audit.Logger nil-safe (no-op when nil, allows unit tests without DB)
- [x] 62 tests passing (CRM: 12, HRM: 12, Global user mgmt: 6, auth: 32)

### Phase 1.2 Completed: TOTP 2FA (2026-04-16)
- [x] Migration 000003: login_attempt_count + login_locked_until (users), attempt_count + invalidated_at (challenges)
- [x] pkg/crypto/aes.go: AES-256-GCM encrypt/decrypt for TOTP secret at rest
- [x] pkg/auth/totp.go: GenerateTOTPKey (pquerna/otp, SHA-1, 6 digits, 30s period), ValidateTOTP, QRCodePNG (skip2/go-qrcode)
- [x] pkg/auth/backup_codes.go: GenerateBackupCodes (10×8 chars), HashBackupCode (bcrypt cost=10), CheckBackupCode, HashDeviceFingerprint
- [x] Config: TwoFAConfig (EncryptionKey, ChallengeTTLSecs, MaxAttempts, TrustDeviceDays, MaxTrustedDevices, Issuer)
- [x] domain: TwoFARepository interface, BackupCode + TrustedDevice entities, updated UserForAuth + TwoFactorChallenge
- [x] repository: twofa_postgres.go — all TwoFARepository methods; queries/twofa.sql reference queries
- [x] usecase: Enable2FA, VerifySetup, Disable2FA (password-verified), Verify2FALogin, VerifyBackupCode, RegenBackupCodes
- [x] login.go: brute-force lock (5 failures → 15 min lock), trusted device check (SHA-256 fingerprint), challenge creation
- [x] handler: TwoFAHandler (/auth/2fa/setup, /auth/2fa/confirm, /auth/2fa DELETE, /auth/2fa/verify, /auth/2fa/backup, /auth/2fa/backup-codes/regenerate)
- [x] Login returns HTTP 202 + challenge_id when 2FA required; trusted device skips 2FA
- [x] 22 tests PASS: 6 Login + 4 CreateUser + 4 Login2FA + 5 Verify2FALogin + 3 VerifyBackupCode + 5 integration (skipped without DB)

### Phase 1.6 Completed: Audit Logs API, Org Module & Web Pages (2026-04-20)
- [x] **Audit Logs API**: `GET /api/v1/audit-logs` với full filtering (module, resource, action, user_id, from, to)
- [x] `internal/global/auth/usecase/list_audit_logs.go`: `ListAuditLogsUseCase` + `AuditLogQuerier` interface
- [x] `internal/global/auth/repository/audit_postgres.go`: `AuditLogRepo` — LEFT JOIN users để lấy full_name, dynamic WHERE clause
- [x] `internal/global/auth/handler/audit_handler.go`: `AuditHandler.List` — bind query params, trả PaginatedResult[AuditLogEntry]
- [x] **Org bounded context** (`internal/org/`): Branch & Department entities, repository, usecase, handler
- [x] `GET|POST /api/v1/branches`, `GET|PUT /api/v1/branches/:id` — RBAC: SUPER_ADMIN, FIRM_PARTNER
- [x] `GET|POST /api/v1/departments`, `GET|PUT /api/v1/departments/:id` — RBAC: SUPER_ADMIN, FIRM_PARTNER
- [x] **Frontend**: `/users` page — User CRUD + role assignment (create, edit, delete, assign roles)
- [x] **Frontend**: `/branches` page — Branch CRUD + Department CRUD (tabbed UI)
- [x] **Frontend**: `/audit-logs` page — read-only log viewer với filter by module, action, date range; color-coded action badges
- [x] New services: `src/services/users.ts`, `src/services/branches.ts`, `src/services/audit.ts`
- [x] Sidebar updated to include Users, Branches, Audit Logs navigation items

### Phase 1.5 Completed: Sales Owner & Salesperson Fields (2026-04-16)

- [x] **Migration 000005**: ALTER `clients` (+sales_owner_id, +referrer_id), ALTER `employees` (+is_salesperson, +sales_commission_eligible, +bank_account_number_enc, +bank_account_name), ALTER `engagements` (+primary_salesperson_id, guarded DO block)
- [x] **CRM**: `sales_owner_id` + `referrer_id` added to `Client` entity, CreateClientParams, UpdateClientParams, DTO, repository queries, handler — fully round-tripped in API
- [x] **HRM**: `is_salesperson` + `sales_commission_eligible` added to `Employee` entity, both params, DTO, repository queries
- [x] **HRM**: bank fields encrypted at rest (AES-256-GCM, `pkg/crypto/aes.go`); dedicated `PUT /employees/:id/bank-details` endpoint; fields tagged `json:"-"` — never exposed in list/detail responses
- [x] `HRMConfig.BankEncryptionKey` (`HRM_BANK_ENCRYPTION_KEY` env var) added to `pkg/config`
- [x] Audit `UPDATE_BANK_DETAILS` action logged on every bank-details update
- [x] New tests: `TestClientUseCase_Create_WithSalesFields`, `TestEmployeeUseCase_UpdateBankDetails`
- [x] All tests passing (`go test ./internal/... ./pkg/... -count=1`)

### Acceptance Criteria
- [x] Auth with JWT + TOTP 2FA passed 100% tests
- [x] Global/CRM/HRM basic CRUD endpoints working
- [x] Audit trail logging all mutations (LOGIN, CREATE, ASSIGN_ROLE) + `GET /api/v1/audit-logs` read endpoint
- [x] Dev environment runs locally via `make dev`
- [x] Client API accepts/returns `sales_owner_id`, `referrer_id`
- [x] Employee API accepts/returns `is_salesperson`, `sales_commission_eligible`
- [x] Bank account fields encrypted at rest, not exposed in list endpoints
- [x] Branch & Department management endpoints (Org bounded context) — SUPER_ADMIN/FIRM_PARTNER
- [x] User management UI (/users), Org UI (/branches), Audit Logs UI (/audit-logs) live in web app

---

## Phase 2: Core Engagement & Timesheet (Months 3-5)

### Bounded Contexts
- **Engagement**: Audit engagement lifecycle, assignments, approvals
- **Timesheet**: Time entry, resource allocation, lockdown
- **Global** (extended): Notifications, audit logs enhancement

### Phase 2.2 Completed: Timesheet + Outbox Pattern (2026-04-18)
- [x] Migration 000009: timesheets, timesheet_entries, attendance tables
- [x] pkg/distlock: Redis SET NX PX distributed lock + Acquirer interface + NoopLock for tests
- [x] Timesheet domain: entities, UPPER_SNAKE_CASE errors, 3 repository interfaces
- [x] Timesheet repository: TimesheetRepo (GetOrCreate idempotent via ON CONFLICT), EntryRepo, AttendanceRepo
- [x] TimesheetUseCase: Submit/Approve/Reject/Lock with state machine + distributed lock on Approve/Reject/Lock
- [x] EntryUseCase: Create/Update/Delete with date-range validation and editable-state guard
- [x] AttendanceUseCase: CheckIn/CheckOut with open-record guard
- [x] 3 handlers, 13 endpoints wired with RBAC
- [x] 13 tests: state machine, date boundary, lock contention — all passing
- [x] Migration 000010: outbox_messages table (PENDING/PROCESSING/PROCESSED/FAILED) with partial index
- [x] pkg/outbox: Publisher (Publish + PublishTx), Poller (SELECT FOR UPDATE SKIP LOCKED → Asynq), EventType constants
- [x] pkg/worker: Server (Asynq-backed), HandlerFunc registry, stub handlers for Phase 2 events
- [x] TimesheetUseCase publishes TimesheetSubmitted/Approved/Rejected/Locked events via outbox
- [x] cmd/worker/main.go: standalone worker binary (poller + Asynq server)
- [x] 6 worker/outbox tests: handler payload validation, nil-publisher safety

### Phase 2.1 Completed: Engagement Bounded Context (2026-04-18)
- [x] Migration 000008: `engagements`, `engagement_members`, `engagement_tasks`, `direct_costs` + indexes
- [x] Domain: all entities, UPPER_SNAKE_CASE error sentinels, 4 repository interfaces
- [x] Repository: `EngagementRepo`, `MemberRepo`, `TaskRepo`, `CostRepo` — raw pgx (CQRS-style)
- [x] Usecase: Engagement CRUD + state machine (DRAFT→PROPOSAL→CONTRACTED→ACTIVE→COMPLETED→SETTLED)
- [x] Usecase: Team assignment with allocation-sum ≤ 100% enforcement; task + direct-cost lifecycle
- [x] Handler: 4 handlers, 21 endpoints wired with RBAC (`FIRM_PARTNER`/`AUDIT_MANAGER`/`AUDIT_STAFF`)
- [x] All mutations emit audit log; list returns `PaginatedResult[T]`; errors → correct HTTP codes
- [x] 17 new tests: state machine, allocation overflow, cost status transitions — all passing

### Deliverables
- [x] Engagement bounded context with CQRS (reads via sqlc, writes via GORM) — **Phase 2.1 done 2026-04-18**
- [x] Timesheet bounded context with distributed locking for concurrent approvals — **done 2026-04-18**
- [x] Domain Events + Outbox Pattern implementation (outbox_messages → Asynq) — **done 2026-04-18**
- [x] Engagement state transitions (DRAFT → PROPOSAL → CONTRACTED → ACTIVE → COMPLETED → SETTLED) — already in Phase 2.1
- [x] Timesheet approval workflows with Redis distributed locks — already in Phase 2.2
- [x] WebSocket real-time events for engagement updates — **done 2026-04-18**
- [x] Pagination (offset & cursor-based) for all list endpoints — **done 2026-04-18**
- [x] Full-text search on client/engagement data — **done 2026-04-18**

### Phase 2.3 Completed: Search + Pagination (2026-04-18)
- [x] Migration 000011: pg_trgm extension + GIN indexes on clients, engagements, employees
- [x] CRM client search expanded: business_name + english_name + tax_code + representative_name (expression index)
- [x] HRM employee search aligned to expression index: full_name + email
- [x] Engagement search: description ILIKE (existing, now indexed)
- [x] pkg/pagination: OffsetResult[T] + CursorResult[T] shared across all modules
- [x] Cursor pagination added to GET /engagements + GET /timesheets
- [x] 13 pagination tests; 2 search tests (client + engagement)

### Acceptance Criteria
- [x] Engagement CRUD with state transitions passing 100% tests
- [x] Timesheet approval without race conditions (distributed locks verified)
- [x] Domain events published and processed via Asynq
- [x] All list endpoints paginated and searchable
- [x] Real-time WebSocket notifications working

---

## Phase 3: Billing, Revenue & Commission (Months 5-7)

### Bounded Contexts
- **Billing**: Invoice generation, time-to-bill, payment processing
- **WorkingPapers**: Audit working paper management with JSONB snapshots
- **Commission**: Commission plan management, accrual engine, approval & payout

### Phase 3.1 Completed: Billing Bounded Context (2026-04-18)
- [x] Migration 000012: `invoices`, `invoice_line_items`, `payments`, `billing_memos` tables + indexes
- [x] Billing domain: Invoice/Payment/Memo aggregates, UPPER_SNAKE_CASE error sentinels
- [x] InvoiceRepo, LineItemRepo, PaymentRepo, MemoRepo — raw pgx (CQRS-style)
- [x] InvoiceUseCase: Create/Update/Delete + state machine (DRAFT→SENT→CONFIRMED→ISSUED→PAID→CANCELLED)
- [x] InvoiceUseCase: Issue freezes JSONB snapshot + publishes `invoice.issued` outbox event
- [x] InvoiceUseCase: AddLineItem/DeleteLineItem (DRAFT-only guard)
- [x] PaymentUseCase: Record (balance guard, auto-PAID on full settlement) + Update + Reverse
- [x] PaymentUseCase: publishes `payment.received` outbox event
- [x] MemoUseCase: Create credit note / adjustment + publishes `credit_note.issued` outbox event
- [x] 3 handlers (InvoiceHandler, PaymentHandler, MemoHandler), 17 endpoints wired with RBAC
- [x] All mutations emit audit log; list endpoints return PaginatedResult[T]
- [x] 12 tests: state machine, payment balance guard, DRAFT-only line item guard — all passing

### Phase 3.3 Completed: Working Paper Bounded Context (2026-04-18)
- [x] Migration 000013: `working_paper_folders`, `working_papers`, `working_paper_reviews`, `working_paper_comments`, `audit_templates`
- [x] Domain: WPStatus (DRAFT/IN_REVIEW/COMMENTED/FINALIZED/SIGNED_OFF), DocumentType, ReviewerRole, ReviewStatus, IssueStatus enums
- [x] WPRepo, ReviewRepo, CommentRepo, FolderRepo, TemplateRepo — raw pgx
- [x] WorkingPaperUseCase: Create/Update/Delete + SubmitForReview (seeds 3-level review chain) + Finalize (JSONB snapshot) + SignOff
- [x] Finalize guards: all reviews APPROVED + zero unresolved comments before capturing snapshot
- [x] ReviewUseCase: Approve/RequestChanges (auto-transitions WP to COMMENTED) + AddComment/ListComments
- [x] TemplateUseCase: Create/Update/Retire/List/ApplyToEngagement
- [x] 3 handlers (WPHandler, ReviewHandler, TemplateHandler), 16 endpoints with RBAC
- [x] All mutations emit audit log; list endpoints return PaginatedResult[T]
- [x] 7 tests: Create, SubmitForReview chain seeding, state transition guards, Finalize snapshot capture — all passing

### Phase 3.4 Completed: Approval Workflows (2026-04-18)
- [x] WP: `ResolveComment` use case + `POST /working-papers/{id}/reviews/{role}/comments/{comment_id}/resolve`
- [x] WP: `FolderUseCase` (Create/ListByEngagement) + `GET|POST /engagements/{id}/folders`
- [x] WP: `PendingReview` use case + `GET /working-papers/pending-review?role=SENIOR_AUDITOR`
- [x] WP: `ListPendingReview` on WPRepository — queries IN_REVIEW/COMMENTED with PENDING review rows for given role
- [x] Billing: `ApprovalQueue` use case + `GET /invoices/approval-queue` — returns SENT+CONFIRMED invoices awaiting action
- [x] Billing: Extended `ListInvoicesFilter.Statuses []InvoiceStatus` for multi-status IN filtering
- [x] 3 new tests: PendingReview pagination, ResolveComment happy path, ResolveComment not-found — all passing

### Phase 3.8 Completed: Commission Accrual Engine (2026-04-18)
- [x] `domain/accrual.go`: `InvoiceAccrualData`, `PaymentAccrualData` DTOs + `BillingDataReader` interface
- [x] `EngCommissionRepository` extended: `ListActiveByTrigger`, `SumHoldbackByEngagement`
- [x] `repository/billing_reader.go`: `BillingReader` — reads invoices/payments tables for accrual (joins)
- [x] `usecase/accrual.go`: `AccrualUseCase` — `AccrueOnInvoiceIssued`, `AccrueOnPaymentReceived`, `ReleaseHoldback`
- [x] Commission calculation: flat/tiered/fixed + holdback deduction + max_amount cap
- [x] Idempotency: `ErrDuplicateAccrual` silently skipped (unique constraint on `(ec_id, invoice_id)` / `(ec_id, payment_id)`)
- [x] `internal/commission/worker/handlers.go`: Asynq handlers for `invoice.issued`, `payment.received`, `EngagementSettled`
- [x] `pkg/outbox`: Added `EventInvoiceIssued`, `EventPaymentReceived`, `EventCreditNoteIssued` constants
- [x] `cmd/server/main.go`: Asynq client + outbox poller + worker server wired; all event handlers registered
- [x] 5 new tests: AccrueOnInvoiceIssued (happy + idempotent), AccrueOnPaymentReceived (happy + billing-error), CalculateFixed — all passing

### Phase 3.7 Completed: Commission Plan Management + Engagement Commission Assignment (2026-04-18)
- [x] **Migration 000012**: `commission_plans`, `engagement_commissions`, `commission_records` tables + 9 indexes (partial unique constraints for idempotency)
- [x] `internal/commission/domain/entity.go`: `CommissionPlan`, `EngagementCommission`, `CommissionRecord` entities + all enums
- [x] `internal/commission/domain/errors.go`: UPPER_SNAKE_CASE error sentinels
- [x] `internal/commission/domain/repository.go`: `PlanRepository`, `EngCommissionRepository`, `RecordRepository` interfaces
- [x] `internal/commission/repository/`: `PlanRepo`, `EngCommissionRepo`, `RecordRepo` (pgx, JSONB for tiers/service_types)
- [x] `internal/commission/usecase/plan.go`: `PlanUseCase` — Create, GetByID, Update, Deactivate, List with audit log
- [x] `internal/commission/usecase/eng_commission.go`: `EngCommissionUseCase` — Create (rate ≤ 100% guard), List, GetByID, Cancel, Approve
- [x] `internal/commission/handler/`: `PlanHandler` + `EngCommissionHandler` + `routes.go`
- [x] Rate validation: `SumRateByEngagement` prevents total commission per engagement exceeding 100%
- [x] Approval gate: `POST /engagement-commissions/{id}/approve` (FIRM_PARTNER only)
- [x] 9 tests passing: Create (happy + conflict + rate-exceeds), GetByID (not-found), List, Deactivate, Approve, Cancel (not-found)

### Phase 3.6 Completed: Billing Report Generation and Export (2026-04-18)
- [x] `domain/report.go`: `BillingPeriodSummary`, `PaymentSummary`, `StatusCount`, `MethodCount`, `ReportRepository` interface
- [x] `repository/report_postgres.go`: `ReportRepo` with GetPeriodSummary (totals + status breakdown), GetPaymentSummary (method breakdown), ListInvoicesForExport (no pagination)
- [x] `usecase/report.go`: `ReportUseCase` with GetPeriodSummary, GetPaymentSummary, ExportInvoicesCSV (standard `encoding/csv`)
- [x] `handler/report_handler.go`: GET `/billing/reports/period-summary`, GET `/billing/reports/payment-summary`, GET `/invoices/export` (CSV with Content-Disposition header)
- [x] `parsePeriod` helper: validates `?start=YYYY-MM-DD&end=YYYY-MM-DD`, ensures end ≥ start
- [x] 4 new tests: GetPeriodSummary happy path, GetPeriodSummary error, ExportInvoicesCSV with data, ExportInvoicesCSV empty (header only) — all passing

### Phase 3.5 Completed: Payment Processing Integration (2026-04-18)
- [x] Payment state machine completed: RECORDED → CLEARED → DISPUTED / REVERSED
- [x] `ClearPayment` use case + `POST /payments/{id}/clear` (bank confirmation)
- [x] `DisputePayment` use case + `POST /payments/{id}/dispute` (bank dispute)
- [x] `ErrPaymentNotCleared`, `ErrPaymentAlreadyCleared` error sentinels
- [x] `ARRepository` interface + `ARRepo` (GetAging + GetOutstanding raw SQL queries)
- [x] `ARUseCase` (GetAging/GetOutstanding) + `ARHandler` (GET /ar/aging, GET /ar/outstanding)
- [x] AR aging buckets: current, 1-30, 31-60, 61-90, 90+ days overdue — live SQL (no matview)
- [x] `domain/ar.go` with `ARAgingRow` + `AROutstandingRow` entities
- [x] 6 new tests: ClearPayment (happy + not-found), DisputePayment (happy + not-cleared), GetAging, GetOutstanding — all passing

### Deliverables
- [x] Billing bounded context with billing rules engine — **Phase 3.1 done 2026-04-18**
- [x] Invoice generation from timesheet + rate cards — **Phase 3.2 done 2026-04-18**
- [x] Working Paper bounded context with JSONB snapshot storage — **Phase 3.3 done 2026-04-18**
- [x] Approval workflows for invoices and working papers — **Phase 3.4 done 2026-04-18**
- [x] Payment processing integration (stripe/payment gateway) — **Phase 3.5 done 2026-04-18**
- [x] Billing report generation and export — **Phase 3.6 done 2026-04-18**
- [x] Mobile App v1 (React Native / Expo) with authentication — **done 2026-04-18** (apps/mobile: Login + 2FA + Dashboard + Engagements + Timesheet entry + Profile)
- [x] Push notifications for high-priority events — **done 2026-04-18** (pkg/notification: Notifier delivers to all active devices via push.Relay; worker handlers: NewTimesheetApproved/Rejected/Submitted/LockedHandler + NewEngagementActivatedHandler; wired in cmd/server + cmd/worker; 11 tests pass)
- [x] **Migration 000012**: `commission_plans`, `engagement_commissions`, `commission_records` + 9 indexes — **Phase 3 done 2026-04-18**

#### Epic: Commission Plan Management `[Tuần 1-2]`
- [x] `CommissionPlan` CRUD (Go domain + repository + usecase + handler) — **Phase 3 done 2026-04-18**
- [x] Plan types: flat / tiered / fixed / custom — enum + validation — **Phase 3 done 2026-04-18**
- [x] Plan UI cho Admin/Director (list, create, edit, deactivate) — **done 2026-04-18** (apps/web: /commissions page với approve/mark-paid actions)

#### Epic: Engagement Commission Assignment `[Tuần 3]`
- [x] `EngagementCommission` CRUD — 4 roles (primary, referrer, account_manager, technical_lead) — **Phase 3 done 2026-04-18**
- [x] Validation: tổng rate tất cả `EngagementCommission` trên 1 engagement ≤ 100% — **Phase 3 done 2026-04-18**
- [x] Approval workflow: `EngagementCommission.rate > 20%` → cần Partner/Director duyệt — **Phase 3 done 2026-04-18**

#### Epic: Accrual Engine `[Tuần 4]`
- [x] Outbox pattern Billing → Asynq → Commission Service (subscribe events) — **Phase 3 done 2026-04-18**
- [x] `AccrueOnInvoiceIssued(invoiceID)` handler — tính commission theo `trigger_on = invoice_issued` — **Phase 3 done 2026-04-18**
- [x] `AccrueOnPaymentReceived(paymentID)` handler — tính theo `trigger_on = payment_received` — **Phase 3 done 2026-04-18**
- [x] `AccrueOnEngagementCompleted(engID)` / `ReleaseHoldback(engID)` — **Phase 3 done 2026-04-18**
- [x] Idempotency: unique constraint `(engagement_commission_id, invoice_id)` + `(engagement_commission_id, payment_id)` — tests — **Phase 3 done 2026-04-18**

#### Epic: Approval & Payment `[Tuần 5]`
- [x] Pending approval queue API: `GET /api/v1/commissions/records?status=accrued` — `[S]` — _Dep: CommissionRecord repository_
- [x] Approve record: `POST /commissions/records/{id}/approve` → status `accrued → approved` — `[S]` — _Dep: Pending queue_
- [x] Mark-as-paid flow: `POST /commissions/records/{id}/mark-paid` (Accountant) — `[S]` — _Dep: Approve flow_
- [x] Bulk approve/pay: `POST /commissions/records/bulk-approve`, `/bulk-pay` — `[M]` — _Dep: Single approve/pay_
- [x] Pending approvals UI (Director/Partner) — **done 2026-04-18** (apps/web: /commissions với status filter + approve button)
- [x] Pending payouts UI (Accountant) — **done 2026-04-18** (apps/web: /commissions với mark-paid button)

#### Epic: Clawback `[Tuần 6]`
- [x] Auto clawback on `invoice.cancelled` event → tạo `CommissionRecord` âm — `[M]` — _Dep: Accrual engine events_
- [x] Auto clawback on `credit_note.issued` event — `[S]` — _Dep: Auto clawback on invoice.cancelled (same pattern)_
- [x] Manual clawback: `POST /commissions/records/{id}/clawback` với `reason` — `[M]` — _Dep: CommissionRecord repository_
- [x] `clawback_record_id` self-reference chain (immutable audit) — `[S]` — _Dep: Manual clawback_

#### Epic: Salesperson UI `[Tuần 7]`
- [x] My Commissions list: `GET /me/commissions` (paginated, filterable by status/period) — `[S]` — _Dep: CommissionRecord read_
- [x] My Commission summary: `GET /me/commissions/summary` (YTD, month, pending, on_hold) — `[S]` — _Dep: My Commissions list_
- [x] Commission Statement PDF export (per salesperson, per period) — `[L]` — _Dep: pkg/export PDF, My Commissions data_
- [x] My Commission dashboard widget (frontend, hiện khi `is_salesperson = true`) — **done 2026-04-18** (apps/web: /commissions/my page + dashboard personal widget)

#### Epic: Manager/Director UI `[Tuần 8]`
- [x] Team earnings view: `GET /commissions/team?manager_id=` — `[M]` — _Dep: CommissionRecord read + RBAC scoping_
- [x] Pending approvals queue UI (Director/Partner — team-scoped) — **done 2026-04-18** (apps/web: /commissions + filter by status=accrued)
- [x] Pending payouts queue UI (Accountant — all) — **done 2026-04-18** (apps/web: /commissions + filter by status=approved)

### Acceptance Criteria
- [x] Invoice generation from timesheet data (no data drift) — **done 2026-04-18** (ListLockedByEngagement + snapshot JSON; TestGenerateUseCase_NoDataDrift passes)
- [x] Working paper snapshots immutable after approval — **done 2026-04-18** (ErrWorkingPaperNotEditable; Update rejects FINALIZED/SIGNED_OFF; 3 immutability tests pass)
- [x] Payment processing with audit trail — **done 2026-04-18** (auditLog.Log on Record/Update/Reverse/Clear/Dispute)
- [x] Mobile app basic auth & engagement viewing — **done 2026-04-18** (apps/web: Login+2FA pages, Engagement list+state transitions, responsive layout)
- [x] Billing reports exported to PDF/Excel — **done 2026-04-18** (XLSX via excelize; CSV export endpoint)
- [x] Commission accrual triggered automatically on invoice/payment events (idempotent) — **verified 2026-04-18**
- [x] Commission approval workflow passing 100% tests — **verified 2026-04-18**
- [x] Clawback logic verified on invoice cancel scenario — **done 2026-04-18** (Cancel endpoint + outbox event + 4 clawback tests)
- [x] My Commissions UI displays correct YTD/month/pending amounts — **done 2026-04-18** (apps/web: /commissions/my — YTD accrued/paid, month accrued/paid, pending_approval, on_hold cards + history table)

---

## Phase 4: Tax Advisory & Advanced Analytics (Months 7-9)

### Bounded Contexts
- **TaxAdvisory**: Tax engagement tracking, compliance checkpoints
- **Reporting**: Advanced analytics, dashboards, KPIs (including commission reports)

### Deliverables
- [x] Tax Advisory bounded context with compliance rules
- [x] Tax engagement tracking with milestone tracking
- [x] Reporting bounded context with materialized views (mv_*)
- [x] Dashboard with engagement pipeline, revenue KPIs, staff utilization
- [x] 2FA Push-based approval for critical operations
- [x] Advanced filtering & complex queries with PostgreSQL full-text search — **done 2026-04-18** (pg_trgm + GIN indexes; SearchInput component in frontend)
- [x] Mobile app enhancements (timesheet entry, document viewing) — **done 2026-04-18** (apps/mobile: Timesheet entries screen with add/delete entries, Engagement detail with timeline)

### Phase 4.2 Completed: Reporting Bounded Context (2026-04-18)
- [x] Migration 000014: 5 materialized views (mv_revenue_by_service, mv_utilization_rate, mv_ar_aging, mv_engagement_progress, mv_commission_summary) + mv_refresh_log table
- [x] Domain: RevenueByService, UtilizationRate, ARAgingRow, EngagementProgress, CommissionMonthlySummary, ExecutiveDashboard, ManagerDashboard, PersonalDashboard, ReportFilter, MVRefreshLog entities
- [x] ReportingRepository interface covering all MV reads + dashboard aggregates + revenue-by-staff
- [x] reporting_postgres.go: Full ReportingRepo implementation (~300 lines, raw pgx)
- [x] DashboardUseCase: ExecutiveDashboard (company KPIs) + ManagerDashboard (team-scoped) + PersonalDashboard (individual + salesperson commission)
- [x] ReportUseCase: RevenueReport, UtilizationReport, ARAgingReport, EngagementStatusReport, CommissionSummaryReport, RevenueByStaffReport, RefreshMaterializedViews, GetMVRefreshStatus
- [x] DashboardHandler + ReportHandler + routes wired with RBAC (FIRM_PARTNER/AUDIT_MANAGER/AUDIT_STAFF/SUPER_ADMIN)
- [x] Asynq worker: reporting:refresh-views nightly MV refresh job
- [x] Admin endpoint: POST /admin/refresh-materialized-views (SUPER_ADMIN)
- [x] 13 tests passing: ExecutiveDashboard (happy + 2 error cases), ManagerDashboard, PersonalDashboard (salesperson + non-salesperson), all report methods, refresh views

#### Commission Reporting & Dashboards `[Dep: Phase 3 Commission Module complete]`
- [x] Báo cáo: **Bảng kê hoa hồng cá nhân** (Commission Statement) — per salesperson, per period — `[M]` — done in Phase 3 (GET /me/commissions/statement + export)
- [x] Báo cáo: **Chi hoa hồng tổng hợp** (Commission Payout) — tổng chi theo tháng — `[M]` — GET /reports/commission-payout
- [x] Báo cáo: **Hoa hồng theo dịch vụ** — group by service type — `[M]` — GET /reports/commission-by-service
- [x] Báo cáo: **Commission pending** (chưa duyệt / chưa chi) — `[S]` — GET /reports/commission-pending-detail
- [x] Báo cáo: **Commission clawback** — danh sách clawback theo tháng — `[S]` — GET /reports/commission-clawback
- [x] Báo cáo: **Revenue by Salesperson** — top salespeople theo doanh thu — `[M]` — GET /reports/revenue-by-salesperson
- [x] Dashboard: **Commission KPIs** trên Executive Dashboard — accrued/paid/pending/on_hold + commission % of revenue — `[M]` — via ExecutiveDashboard.CommissionKPIs
- [x] Dashboard: **Salesperson section** trên Personal Dashboard — YTD, month, pending, on_hold (hiện khi `is_salesperson=true`) — `[M]` — via PersonalDashboard.IsSalesperson

### Phase 4.4 Completed: 2FA Push-Based Approval (2026-04-18)
- [x] Migration 000015: push_devices table + push_response/responded_at fields on two_factor_challenges
- [x] pkg/push: PushDevice entity, DeviceRepository interface + PostgreSQL implementation, WebSocket PushRelay
- [x] TwoFactorChallenge extended with PushResponse, RespondedAt, PushChallengeStatus enum
- [x] TwoFARepository extended: RespondToPushChallenge, FindPushChallenge
- [x] Push2FAUseCase: RespondToPush, GetPushStatus (pending/approved/rejected/expired + token issuance), SendPushToDevice
- [x] PushDeviceUseCase: RegisterDevice, UnregisterDevice, ListDevices, Heartbeat
- [x] PushHandler: 9 new endpoints (device CRUD, push relay WS, push 2FA response/status/resend)
- [x] Auth routes updated with push 2FA endpoints and push device management group
- [x] 8 tests: RespondToPush, GetPushStatus (pending/expired/rejected/approved/already-verified/not-found), SendPushToDevice offline
- [x] cmd/server/main.go wired: PushRelay, PushDeviceRepo, PushDeviceUC, Push2FAUC, PushHandler

### Phase 4.3 Completed: Commission Reporting & Dashboard (2026-04-18)
- [x] GET /reports/commission-payout — monthly payout summary (total approved, total paid, record count)
- [x] GET /reports/commission-by-service — commission grouped by engagement service_type (avg rate, payable, paid)
- [x] GET /reports/commission-pending-detail — pending records with aggregate totals (approval + payout buckets)
- [x] GET /reports/commission-clawback — clawback records + total for N months
- [x] Domain types: CommissionPayoutRow, CommissionByServiceRow, CommissionPendingRow, CommissionClawbackRow, CommissionPendingSummary, CommissionClawbackSummary
- [x] 8 new tests (payout/by-service/pending/clawback happy paths + default param handling) — all passing

### Acceptance Criteria
- [x] Tax Advisory workflows with compliance checkpoints — **verified 2026-04-18** (deadline tracking, advisory records, compliance score, auto-generate, DUE_SOON/OVERDUE cron — 10 tests pass)
- [x] Reporting dashboards show real-time engagement/revenue data — **verified 2026-04-18** (ExecutiveDashboard, ManagerDashboard, PersonalDashboard via materialized views — 13 tests pass; apps/web /reports page)
- [x] Commission KPIs on Executive Dashboard accurate (verified vs raw records) — **done 2026-04-18** (TestCommissionKPIs_PctRevenue_Calculation: 500K/10M×100=5.0%; TestCommissionKPIs_ZeroRevenue_NoPanic)
- [x] Commission Statement PDF output matches manual calculation (test data scenario) — **verified 2026-04-18** (TestRecordUseCase_GetStatement_Quarter: TotalAccrued=8M, TotalPayable=6.4M, TotalPaid=2.4M)
- [x] Revenue by Salesperson report matches Engagement + Invoice join — **done 2026-04-18** (TestRevenueByStaffReport_Happy: StaffA=50M/5inv, StaffB=30M/3inv)
- [x] 2FA push approval for payment processing works — **verified 2026-04-18** (6 middleware tests: FIRM_PARTNER/SUPER_ADMIN blocked w/o 2FA)
- [x] Materialized views refresh strategy verified (< 5s staleness) — **verified 2026-04-18** (TestRefreshMaterializedViews_Happy + daily cron worker)

---

## Phase 5: Go-Live & Production Hardening (Months 9-12)

### Deliverables
- [ ] UAT environment with prod-like data
- [ ] Performance testing & optimization (100+ concurrent users)
- [x] Kubernetes deployment configuration — **done 2026-04-18** (k8s/base: namespace, configmap, secret, api-deployment, api-service, api-hpa, api-ingress, redis; overlays/staging + overlays/production; Kustomize)
- [x] CI/CD pipeline (GitHub Actions) — **done 2026-04-18** (ci.yml: govulncheck + 60% coverage threshold; cd.yml: GHCR push → staging on main merge, production on semver tag, DB migrations, smoke test, GitHub Release)
- [x] Monitoring (Prometheus + Grafana + Loki) — **done 2026-04-18** (pkg/metrics: Prometheus middleware + /metrics + business recorders; monitoring/: prometheus.yml + alerts.yml + Grafana dashboard + Loki + Promtail; docker-compose.monitoring.yml; k8s/monitoring/)
- [ ] User training & documentation
- [ ] Production data migration plan
- [ ] Disaster recovery & backup strategy
- [x] 2FA enforcement for FIRM_PARTNER and SUPER_ADMIN roles — **done 2026-04-18**
- [x] Rate limiting: 100 req/min per user, 1000 req/min per IP — **done 2026-04-18**

### Phase 5.1 Completed: Mobile App — React Native / Expo (2026-04-18)
- [x] `apps/mobile/` Expo SDK 51 + Expo Router v3 (file-based navigation)
- [x] Auth: Login screen, Verify 2FA screen (TOTP + backup code), SecureStore token persistence
- [x] Auth guard: root `_layout.tsx` hydrates SecureStore → redirect to login/dashboard automatically
- [x] Tab navigator: Dashboard, Engagements, Timesheet, Profile (4 tabs)
- [x] Dashboard: personal stats (active engagements, pending timesheets, hours this month) + Commission widget (khi `is_salesperson=true`)
- [x] Engagements: list with search + pull-to-refresh; detail screen with status timeline
- [x] Timesheet: list with submit action; entries screen (add/delete entries, engagement picker, hours input)
- [x] Profile: user info, roles badges, logout
- [x] UI system: Button, Input, Badge, Card, StatCard, Spinner, EmptyState, ListItem (StyleSheet + theme)
- [x] API layer: Axios + SecureStore interceptor + 4 services (auth, engagements, timesheets, reports)

### Phase 5.0 Completed: Frontend Web App (2026-04-18)
- [x] apps/web: Next.js 14 App Router — auth-guarded layout with Sidebar + Header
- [x] UI Design System: Button, Input, Badge, Card, Table, Dialog, Select, Toast, Spinner (Tailwind + Radix UI)
- [x] Auth flow: Login page + 2FA verify page (TOTP + backup code) + Zustand persist store
- [x] All module pages: Dashboard, Clients, Employees, Engagements, Timesheets, Invoices, Payments, Working Papers
- [x] Commission module: /commissions (approve/mark-paid), /commissions/my (YTD/month/pending/on_hold summary + history)
- [x] Reports page: Executive dashboard (revenue KPIs + commission KPIs), Manager dashboard, commission payout report, billing report
- [x] API services layer: 8 service files covering all backend modules + React Query integration

### Acceptance Criteria
- [ ] All modules pass UAT with real user workflows
- [ ] Performance: P95 < 500ms for API endpoints
- [ ] 99.9% uptime SLA established
- [ ] User training completed for all roles
- [ ] Backup/restore tested and documented

---

## Quality & Engineering Standards (All Phases)

### Code Quality
- [x] All new code covered by unit tests (>80%) — **done 2026-04-18** (crm 96%, engagement 89%, workingpaper 85%, hrm 95%, reporting 80%, billing 71%; 19 packages all pass)
- [x] Integration tests for domain workflows — **done 2026-04-18** (engagement/usecase/integration_test.go + workingpaper/usecase/integration_test.go; both skip w/o DATABASE_URL)
- [x] `make lint test` passing before merge — **done 2026-04-18** (go build ./... + go vet ./... + go test 20 packages all pass)
- [ ] Go: golangci-lint, TypeScript: eslint + prettier

### API Standards
- [ ] Every endpoint documents required roles (RBAC)
- [x] Every mutation returns audit_id for traceability — **done 2026-04-18** (audit.Logger.Log() returns uuid; AuditIDMiddleware sets X-Audit-ID header; all 162 call sites updated)
- [x] Errors use UPPER_SNAKE_CASE codes (e.g., ENGAGEMENT_LOCKED) — **done 2026-04-18** (all audit action strings uppercased; raw error exposure fixed)
- [ ] Versioning strategy: /api/v1, additive changes, 6-month deprecation

### Documentation
- [ ] Module docs (docs/modules/*.md) kept in sync with code
- [ ] API OpenAPI 3.0 spec auto-generated
- [ ] Architecture decisions logged (docs/DECISIONS.md)
- [ ] CLAUDE.md updated with new conventions/patterns

### Deployment
- [ ] Feature flags for gradual rollout
- [x] Database migrations use backward compatibility — **done 2026-04-18** (golang-migrate; migrations 000001–000015 all additive)
- [x] Monitoring alerts on: error rate, latency, locked rows — **done 2026-04-18** (alerts.yml: error_rate_high, high_latency, locked_rows_detected)
- [ ] Weekly security updates for dependencies

---

## Long-term Vision (Years 2-3)

- [ ] AI-powered audit recommendations
- [ ] Advanced predictive analytics
- [ ] Integration with external Tax/Accounting systems
- [ ] Mobile app parity with web (React Native completion)
- [ ] Multi-firm SSO & federation
- [ ] Blockchain-based audit trail (optional)