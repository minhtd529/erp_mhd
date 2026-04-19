# CHANGELOG

All notable changes to the ERP Audit System specification and documentation are recorded here.  
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## [1.3.0] — 2026-04-19

### Security
- Added `ValidateProductionConfig` startup check that blocks insecure default values for `JWT_SECRET`, `TOTP_ENCRYPTION_KEY`, `HRM_BANK_ENCRYPTION_KEY`, and `MINIO_*` credentials in non-development environments.

### Fixed
- Gin routing conflicts in `workingpaper` and `tax` modules: nested routes under `/engagements/:X` and `/clients/:X` now use `:id` canonically (Gin disallows mixed param names on shared prefixes).
- Asynq Redis connection: use parsed `Addr` (`host:port`) instead of full `redis://` URL, which caused "too many colons in address" error at startup.
- Migration 000018: added `updated_by UUID` column to `commission_plans` and `commission_records` for full audit-column compliance.

### Documentation
- **SPEC §13.2 Audit Log Format**: documented normalized (`module` + `resource` + `action`) form as the canonical implementation; verbose `"CREATE_CLIENT"` form acknowledged as acceptable alternative. New modules must use normalized form.
- **SPEC §13.3 Deletion & Retention Conventions**: introduced 5 patterns (A–E) replacing one-size-fits-all `is_deleted`. Each entity type mapped to the pattern that fits its semantics (status-based, is_active flag, immutable/clawback, hard-delete+state-gate, is_deleted).
- **CLAUDE.md Use Case Struct Naming**: documented bundled `{Entity}UseCase` as the default; per-action split (`{Action}{Entity}UseCase`) reserved for modules with diverging dependencies (e.g., `auth`). Corrected "Services" row — codebase uses `UseCase` structs, not `Service` structs.
- **CLAUDE.md Response DTO Naming**: `{Entity}Response` default for single-shape entities; `{Entity}{Op}Response` suffix introduced only when 2+ shapes exist for the same entity.

### Unchanged (explicitly verified)
- `payments`, `engagement_commissions`: status-based lifecycle retained (Pattern A).
- `commission_plans`: `is_active` flag retained as soft-delete mechanism (Pattern B).
- `commission_records`: immutable append-only + clawback chain retained (Pattern C).
- `invoice_line_items`: hard delete + DRAFT state gate retained (Pattern D).

---

## [1.2.0] — 2026-04-16

### Added
- **Module 1 CRM**: Sales Owner tracking — `sales_owner_id`, `referrer_id` on `Client` entity
- **Module 1 CRM**: Commission Management sub-module
  - `CommissionPlan` entity (types: flat / tiered / fixed / custom)
  - `EngagementCommission` entity (roles: primary, referrer, account_manager, technical_lead)
  - `CommissionRecord` entity (lifecycle: accrued → approved → paid / clawback)
  - Business rules: immutable record, idempotency, total rate ≤ 100%, approval threshold, holdback, auto-clawback
  - Commission Lifecycle flow diagram + calculation example
  - `CommissionRepository` + `CommissionService` interfaces
  - Commission API endpoints (`/commission-plans`, `/engagements/{id}/commissions`, `/commissions/records`, `/me/commissions`)
- **Module 2 Engagement**: `primary_salesperson_id` field (denormalized reference to main salesperson)
- **Module 4 Billing**: Event-driven Commission integration via NATS outbox
  - 5 events: `invoice.issued`, `payment.received`, `invoice.cancelled`, `credit_note.issued`, `engagement.settled`
- **Module 7 HRM**: Salesperson fields on `Employee` entity
  - `is_salesperson`, `sales_commission_eligible`, `default_commission_plan_id`
  - `bank_account_number`, `bank_account_name` (AES-256-GCM encrypted)
- **Module 8 Reporting**: Commission KPIs on Executive Dashboard (accrued/paid/pending/on_hold, % of revenue)
- **Module 8 Reporting**: Salesperson section on Personal Dashboard (YTD, month, pending, on_hold)
- **Module 8 Reporting**: 5 new commission reports (Statement, Payout, By Service, Pending, Clawback) + Revenue by Salesperson
- **Database**: 3 new tables (`commission_plans`, `engagement_commissions`, `commission_records`) + 9 indexes
- **Rollout Plan**: Commission Module placed in Phase 3 (8-week detailed breakdown)
- **ROADMAP.md**: Commission epics in Phase 3 (7 epics, 25 tasks with effort/dependencies/week)
- **ROADMAP.md**: Commission reporting tasks in Phase 4
- **ROADMAP.md**: Phase 1.5 tasks for salesperson fields (migration 000005)
- **CLAUDE.md**: Commission business rules (Rules 7-9)
- **CLAUDE.md**: Commission spec reference links

### Changed
- Module 1 CRM title: "CRM – QUẢN LÝ KHÁCH HÀNG" → "CRM – QUẢN LÝ KHÁCH HÀNG & HOA HỒNG"
- Phase 3 title: "Billing & Revenue" → "Billing, Revenue & Commission"

### Archived
- `docs/SPEC.md` v1.1 → `docs/archive/SPEC_v1.1.md`

---

## [1.1.0] — 2026-04-12

### Added
- Push Notification Self-Hosted architecture (WebSocket relay + W3C Web Push / VAPID)
- 2FA/MFA: TOTP (pquerna/otp) + Push-based (self-hosted) — no Twilio/Authy dependency
- Mobile app push connection service (`PushConnectionService.java`, `BackgroundTaskManager.swift`)
- pkg/ws: WebSocket Hub + Client (channel-based fan-out, 30-min session timeout)

### Changed
- Auth flow: Login returns HTTP 202 + `challenge_id` when 2FA required
- Trusted device support: SHA-256 fingerprint, 30-day skip window

---

## [1.0.0] — 2026-04-12

### Added
- Initial specification: 9 DDD Bounded Contexts (Global, CRM, Engagement, Timesheet, Billing, WorkingPapers, TaxAdvisory, HRM, Reporting)
- Hexagonal Architecture + CQRS + Outbox Pattern conventions
- Full API Design Conventions (RESTful, RBAC, error codes, pagination)
- Database conventions (soft delete, audit fields, UUID PKs, CHECK constraints)
- Go naming conventions + Frontend conventions
- UI/Style Guide "Professional Audit" (navy + gold palette)
- Phase 1-5 roadmap (9-12 month timeline)
