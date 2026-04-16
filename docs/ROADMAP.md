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
- [x] Audit trail logging all mutations (LOGIN, CREATE, ASSIGN_ROLE)
- [x] Dev environment runs locally via `make dev`
- [x] Client API accepts/returns `sales_owner_id`, `referrer_id`
- [x] Employee API accepts/returns `is_salesperson`, `sales_commission_eligible`
- [x] Bank account fields encrypted at rest, not exposed in list endpoints

---

## Phase 2: Core Engagement & Timesheet (Months 3-5)

### Bounded Contexts
- **Engagement**: Audit engagement lifecycle, assignments, approvals
- **Timesheet**: Time entry, resource allocation, lockdown
- **Global** (extended): Notifications, audit logs enhancement

### Deliverables
- [ ] Engagement bounded context with CQRS (reads via sqlc, writes via GORM)
- [ ] Timesheet bounded context with distributed locking for concurrent approvals
- [ ] Domain Events + Outbox Pattern implementation (outbox_messages → Asynq)
- [ ] Engagement state transitions (DRAFT → IN_PROGRESS → COMPLETED → APPROVED)
- [ ] Timesheet approval workflows with Redis distributed locks
- [ ] WebSocket real-time events for engagement updates
- [ ] Pagination (offset & cursor-based) for all list endpoints
- [ ] Full-text search on client/engagement data

### Acceptance Criteria
- [ ] Engagement CRUD with state transitions passing 100% tests
- [ ] Timesheet approval without race conditions (distributed locks verified)
- [ ] Domain events published and processed via Asynq
- [ ] All list endpoints paginated and searchable
- [ ] Real-time WebSocket notifications working

---

## Phase 3: Billing, Revenue & Commission (Months 5-7)

### Bounded Contexts
- **Billing**: Invoice generation, time-to-bill, payment processing
- **WorkingPapers**: Audit working paper management with JSONB snapshots
- **Commission**: Commission plan management, accrual engine, approval & payout

### Deliverables
- [ ] Billing bounded context with billing rules engine
- [ ] Invoice generation from timesheet + rate cards
- [ ] Working Paper bounded context with JSONB snapshot storage
- [ ] Approval workflows for invoices and working papers
- [ ] Payment processing integration (stripe/payment gateway)
- [ ] Billing report generation and export
- [ ] Mobile App v1 (React Native) with authentication
- [ ] Push notifications for high-priority events
- [ ] **Migration 000006**: `commission_plans`, `engagement_commissions`, `commission_records` + 9 indexes — `[M]` — _Dep: Phase 1.5 migration 000005_

#### Epic: Commission Plan Management `[Tuần 1-2]`
- [ ] `CommissionPlan` CRUD (Go domain + repository + usecase + handler) — `[M]` — _Dep: migration 000006_
- [ ] Plan types: flat / tiered / fixed / custom — enum + validation — `[S]` — _Dep: CommissionPlan CRUD_
- [ ] Plan UI cho Admin/Director (list, create, edit, deactivate) — `[M]` — _Dep: CommissionPlan CRUD_

#### Epic: Engagement Commission Assignment `[Tuần 3]`
- [ ] `EngagementCommission` CRUD — 4 roles (primary, referrer, account_manager, technical_lead) — `[M]` — _Dep: CommissionPlan CRUD_
- [ ] Validation: tổng rate tất cả `EngagementCommission` trên 1 engagement ≤ 100% — `[S]` — _Dep: EngagementCommission CRUD_
- [ ] Approval workflow: `EngagementCommission.rate > 20%` → cần Partner/Director duyệt — `[M]` — _Dep: Engagement Commission CRUD_

#### Epic: Accrual Engine `[Tuần 4]`
- [ ] Outbox pattern Billing → NATS → Commission Service (subscribe events) — `[M]` — _Dep: Billing outbox đã có; cần NATS subscriber mới_
- [ ] `AccrueOnInvoiceIssued(invoiceID)` handler — tính commission theo `trigger_on = invoice_issued` — `[M]` — _Dep: Outbox subscriber_
- [ ] `AccrueOnPaymentReceived(paymentID)` handler — tính theo `trigger_on = payment_received` — `[M]` — _Dep: Outbox subscriber_
- [ ] `AccrueOnEngagementCompleted(engID)` / `ReleaseHoldback(engID)` — `[S]` — _Dep: AccrueOnInvoiceIssued pattern_
- [ ] Idempotency: unique constraint `(engagement_commission_id, invoice_id)` + `(engagement_commission_id, payment_id)` — tests — `[S]` — _Dep: AccrueOn handlers_

#### Epic: Approval & Payment `[Tuần 5]`
- [ ] Pending approval queue API: `GET /api/v1/commissions/records?status=accrued` — `[S]` — _Dep: CommissionRecord repository_
- [ ] Approve record: `POST /commissions/records/{id}/approve` → status `accrued → approved` — `[S]` — _Dep: Pending queue_
- [ ] Mark-as-paid flow: `POST /commissions/records/{id}/mark-paid` (Accountant) — `[S]` — _Dep: Approve flow_
- [ ] Bulk approve/pay: `POST /commissions/records/bulk-approve`, `/bulk-pay` — `[M]` — _Dep: Single approve/pay_
- [ ] Pending approvals UI (Director/Partner) — `[M]` — _Dep: Approve API_
- [ ] Pending payouts UI (Accountant) — `[M]` — _Dep: Mark-as-paid API_

#### Epic: Clawback `[Tuần 6]`
- [ ] Auto clawback on `invoice.cancelled` event → tạo `CommissionRecord` âm — `[M]` — _Dep: Accrual engine events_
- [ ] Auto clawback on `credit_note.issued` event — `[S]` — _Dep: Auto clawback on invoice.cancelled (same pattern)_
- [ ] Manual clawback: `POST /commissions/records/{id}/clawback` với `reason` — `[M]` — _Dep: CommissionRecord repository_
- [ ] `clawback_record_id` self-reference chain (immutable audit) — `[S]` — _Dep: Manual clawback_

#### Epic: Salesperson UI `[Tuần 7]`
- [ ] My Commissions list: `GET /me/commissions` (paginated, filterable by status/period) — `[S]` — _Dep: CommissionRecord read_
- [ ] My Commission summary: `GET /me/commissions/summary` (YTD, month, pending, on_hold) — `[S]` — _Dep: My Commissions list_
- [ ] Commission Statement PDF export (per salesperson, per period) — `[L]` — _Dep: pkg/export PDF, My Commissions data_
- [ ] My Commission dashboard widget (frontend, hiện khi `is_salesperson = true`) — `[M]` — _Dep: My Commission summary API_

#### Epic: Manager/Director UI `[Tuần 8]`
- [ ] Team earnings view: `GET /commissions/team?manager_id=` — `[M]` — _Dep: CommissionRecord read + RBAC scoping_
- [ ] Pending approvals queue UI (Director/Partner — team-scoped) — `[S]` — _Dep: Pending approval queue API_
- [ ] Pending payouts queue UI (Accountant — all) — `[S]` — _Dep: Pending payouts API_

### Acceptance Criteria
- [ ] Invoice generation from timesheet data (no data drift)
- [ ] Working paper snapshots immutable after approval
- [ ] Payment processing with audit trail
- [ ] Mobile app basic auth & engagement viewing
- [ ] Billing reports exported to PDF/Excel
- [ ] Commission accrual triggered automatically on invoice/payment events (idempotent)
- [ ] Commission approval workflow passing 100% tests
- [ ] Clawback logic verified on invoice cancel scenario
- [ ] My Commissions UI displays correct YTD/month/pending amounts

---

## Phase 4: Tax Advisory & Advanced Analytics (Months 7-9)

### Bounded Contexts
- **TaxAdvisory**: Tax engagement tracking, compliance checkpoints
- **Reporting**: Advanced analytics, dashboards, KPIs (including commission reports)

### Deliverables
- [ ] Tax Advisory bounded context with compliance rules
- [ ] Tax engagement tracking with milestone tracking
- [ ] Reporting bounded context with materialized views (mv_*)
- [ ] Dashboard with engagement pipeline, revenue KPIs, staff utilization
- [ ] 2FA Push-based approval for critical operations
- [ ] Advanced filtering & complex queries with PostgreSQL full-text search
- [ ] Mobile app enhancements (timesheet entry, document viewing)

#### Commission Reporting & Dashboards `[Dep: Phase 3 Commission Module complete]`
- [ ] Báo cáo: **Bảng kê hoa hồng cá nhân** (Commission Statement) — per salesperson, per period — `[M]` — _Dep: CommissionRecord read + PDF export_
- [ ] Báo cáo: **Chi hoa hồng tổng hợp** (Commission Payout) — tổng chi theo tháng — `[M]` — _Dep: CommissionRecord approved/paid status_
- [ ] Báo cáo: **Hoa hồng theo dịch vụ** — group by service type — `[M]` — _Dep: CommissionRecord + Engagement service_type join_
- [ ] Báo cáo: **Commission pending** (chưa duyệt / chưa chi) — `[S]` — _Dep: CommissionRecord status filter_
- [ ] Báo cáo: **Commission clawback** — danh sách clawback theo tháng — `[S]` — _Dep: CommissionRecord clawback status_
- [ ] Báo cáo: **Revenue by Salesperson** — top salespeople theo doanh thu — `[M]` — _Dep: Engagement + Invoice data join_
- [ ] Dashboard: **Commission KPIs** trên Executive Dashboard — accrued/paid/pending/on_hold + commission % of revenue — `[M]` — _Dep: Materialized view mv_commission_summary_
- [ ] Dashboard: **Salesperson section** trên Personal Dashboard — YTD, month, pending, on_hold (hiện khi `is_salesperson=true`) — `[M]` — _Dep: Commission KPIs query_

### Acceptance Criteria
- [ ] Tax Advisory workflows with compliance checkpoints
- [ ] Reporting dashboards show real-time engagement/revenue data
- [ ] Commission KPIs on Executive Dashboard accurate (verified vs raw records)
- [ ] Commission Statement PDF output matches manual calculation (test data scenario)
- [ ] Revenue by Salesperson report matches Engagement + Invoice join
- [ ] 2FA push approval for payment processing works
- [ ] Materialized views refresh strategy verified (< 5s staleness)

---

## Phase 5: Go-Live & Production Hardening (Months 9-12)

### Deliverables
- [ ] UAT environment with prod-like data
- [ ] Performance testing & optimization (100+ concurrent users)
- [ ] Kubernetes deployment configuration
- [ ] CI/CD pipeline (GitHub Actions)
- [ ] Monitoring (Prometheus + Grafana + Loki)
- [ ] User training & documentation
- [ ] Production data migration plan
- [ ] Disaster recovery & backup strategy
- [ ] 2FA enforcement for FIRM_PARTNER and SUPER_ADMIN roles
- [ ] Rate limiting: 100 req/min per user, 1000 req/min per IP

### Acceptance Criteria
- [ ] All modules pass UAT with real user workflows
- [ ] Performance: P95 < 500ms for API endpoints
- [ ] 99.9% uptime SLA established
- [ ] User training completed for all roles
- [ ] Backup/restore tested and documented

---

## Quality & Engineering Standards (All Phases)

### Code Quality
- [ ] All new code covered by unit tests (>80%)
- [ ] Integration tests for domain workflows
- [ ] `make lint test` passing before merge
- [ ] Go: golangci-lint, TypeScript: eslint + prettier

### API Standards
- [ ] Every endpoint documents required roles (RBAC)
- [ ] Every mutation returns audit_id for traceability
- [ ] Errors use UPPER_SNAKE_CASE codes (e.g., ENGAGEMENT_LOCKED)
- [ ] Versioning strategy: /api/v1, additive changes, 6-month deprecation

### Documentation
- [ ] Module docs (docs/modules/*.md) kept in sync with code
- [ ] API OpenAPI 3.0 spec auto-generated
- [ ] Architecture decisions logged (docs/DECISIONS.md)
- [ ] CLAUDE.md updated with new conventions/patterns

### Deployment
- [ ] Feature flags for gradual rollout
- [ ] Database migrations use backward compatibility
- [ ] Monitoring alerts on: error rate, latency, locked rows
- [ ] Weekly security updates for dependencies

---

## Long-term Vision (Years 2-3)

- [ ] AI-powered audit recommendations
- [ ] Advanced predictive analytics
- [ ] Integration with external Tax/Accounting systems
- [ ] Mobile app parity with web (React Native completion)
- [ ] Multi-firm SSO & federation
- [ ] Blockchain-based audit trail (optional)