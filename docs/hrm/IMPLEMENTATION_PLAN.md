# HRM Implementation Plan
## ERP System — MDH Audit Firm
**Version:** 1.2 | **Based on:** HRM_SPEC_v1.4.md | **Date:** 2026-04-22

> **Sprint 1 Status (2026-04-22):**
> - Day 1–7 backend fully implemented and committed (migration 000020, 000021, all API endpoints)
> - Phase 3 (Day 2 supplement): AES-256-GCM cipher (`pkg/crypto/aesgcm.go`) + sensitive PII endpoints (`GET/PUT /hrm/employees/:id/sensitive`) + salary history endpoints (`GET/POST /hrm/employees/:id/salary-history`) — committed `dac848f`
> - Phase 4 (Day 8–10 equivalent): Full Employee UI implemented 2026-04-22 — see §Phase 4 Deviations below
> - **Deviation from SPEC:** `salary_at_signing` on contracts is NUMERIC (plaintext), not encrypted — SPEC §11.9 shows TEXT for encryption but migration used NUMERIC. Encryption deferred to future migration.
> - **Deviation from SPEC:** `so_bhxh_encrypted`, `mst_ca_nhan_encrypted` columns present; `bank_account_encrypted` uses separate `HRM_BANK_ENCRYPTION_KEY` (hex-encoded 32 bytes). PII columns (`cccd_encrypted`, `mst_ca_nhan_encrypted`, `so_bhxh_encrypted`) use `HRM_ENCRYPTION_KEY` (base64-encoded 32 bytes, AES-256-GCM).
> - **Test result:** `go test ./pkg/crypto/... ./internal/hrm/...` — 14 crypto tests PASS, all employee/usecase tests PASS.

> **Migration numbering note (2026-04-21):** SPEC v1.4 originally planned HRM migrations 000019–000026. After implementation started, migration 000019 was already occupied by `000019_working_papers`. All HRM migrations shifted +1: they now occupy 000020–000027. All references in this document have been updated accordingly.

---

## Executive Roadmap (Gantt-Style Timeline)

```
Week 1  │ Sprint 1 ████████████████████████████████████
Week 2  │ Sprint 1 ████████████████████████████████████
Week 3  │ Sprint 2 ████████████████████████████████████
Week 4  │ Sprint 2 ████████████████████████████████████
Week 5  │ Sprint 3 ██████████████████
Week 6  │ Sprint 4 ██████████████████
Week 7  │ Sprint 5 ██████████████████
```

| Sprint | Focus | Weeks | Story Points | Migrations |
|---|---|---|---|---|
| Sprint 1 | Organization + Employees | 1–2 | ~40 SP | 000020, 000021, 000027 (partial) |
| Sprint 2 | Provisioning + Certifications/Training | 3–4 | ~35 SP | 000022, 000025 |
| Sprint 3 | Performance Reviews + Independence | 5 | ~20 SP | 000023 |
| Sprint 4 | Leave + OT + Holidays | 6 | ~25 SP | 000024, 000027 (remaining) |
| Sprint 5 | Expenses + Reports + Polish | 7 | ~25 SP | 000026 |

**Total: 4–6 weeks, 8 migrations (000020–000027), 80+ endpoints, 30+ UI pages**

> **Sprint boundary rule:** No code from Sprint N+1 may be committed in Sprint N.
> If a dependency is discovered, escalate immediately (see CLAUDE_CODE_PLAYBOOK.md §Escalation).

---

## Sprint 1: Organization + Employees (Weeks 1–2, ~40 SP)

### Goals
Establish the organizational foundation and full employee record management. After Sprint 1, HR can create employee records with all 40+ fields, encrypted PII is stored and accessible only to authorized roles, and the branch/department matrix is operational.

### Scope
**Migrations:** 000020, 000021, 000027 (branches + departments seed only)
**Tables created/altered:** branches, departments, branch_departments, employees (extended), employee_dependents, insurance_rate_config, employee_salary_history, employment_contracts
**API groups:** organization (10 endpoints), employees (8), my-profile (2), sensitive PII (1), dependents (4), salary-history (2), contracts (5) — See SPEC §13.1, §13.2, §13.3, §13.11, §13.15, §13.16
**UI pages:** /admin/hrm/organization, /admin/hrm/employees, /admin/hrm/employees/new, /admin/hrm/employees/:id, /my-profile

### Dependencies
- Existing: `branches`, `departments`, `users`, `engagements` tables must exist
- Migration 000001–000018 must be applied successfully
- `HRM_ENCRYPTION_KEY` environment variable must be set (32 bytes, base64) — See SPEC §18.1
- `make migrate-lint` script at `scripts/migration-lint.sh` must pass

### Daily Task Breakdown

#### Day 1: Migration 000020 — Organization Schema
**Task 1.1:** Write `000020_hrm_organization.up.sql` — See SPEC §11.1, §11.2, §11.3
- ALTER branches: add is_head_office, city, address, phone, established_date, head_of_branch_user_id, is_active
- ALTER departments: add code (UNIQUE), dept_type (CHECK), description, is_active
- CREATE TABLE branch_departments with uidx_branch_department
- CREATE UNIQUE INDEX uidx_branches_head_office WHERE is_head_office = true

**Task 1.2:** Write `000020_hrm_organization.down.sql` — See SPEC §12.2
- DROP TABLE branch_departments
- ALTER departments DROP COLUMN code, dept_type, description, is_active
- ALTER branches DROP COLUMN is_head_office, city, address, phone, established_date, head_of_branch_user_id, is_active

**Task 1.3:** Run `make migrate-lint` and `make migrate-up && make migrate-down && make migrate-up`
**Task 1.4:** Implement Go repository + handler for `/hrm/organization/branches` (GET list, GET by id) — See SPEC §13.1

#### Day 2: Migration 000021 Part 1 — Employee Extended Columns
**Task 2.1:** Write first half of `000021_hrm_employees_extended.up.sql` — See SPEC §11.4
- ADD all Basic columns: employee_code, grade (CHECK), position_title, manager_id, employment_type, status, hired_date, probation_end_date, termination_date, termination_reason, current_contract_id
- ADD Personal/PII columns: gender, date_of_birth, place_of_birth, nationality, ethnicity, personal_email, personal_phone, work_phone, current_address, permanent_address, cccd_encrypted, cccd_issued_date, cccd_issued_place, passport_number, passport_expiry
- ADD Employment columns: hired_source, referrer_employee_id, probation_salary_pct, work_location, remote_days_per_week
- ADD Qualification columns: education_level, education_major, education_school, education_graduation_year, vn_cpa_number, vn_cpa_issued_date, vn_cpa_expiry_date, practicing_certificate_number, practicing_certificate_expiry

**Task 2.2:** Write `fn_employees_set_code()` trigger — See SPEC §11.5
- Format: `NV{YY}-{SEQ4}` with year from hired_date, padded 4-digit sequence
- BEFORE INSERT trigger with collision-retry loop

#### Day 3: Migration 000021 Part 2 — Related Tables
**Task 3.1:** Continue `000021_hrm_employees_extended.up.sql` — See SPEC §11.4
- ADD Salary/Bank columns: base_salary, salary_currency, salary_effective_date, bank_account_encrypted, bank_name, bank_branch, mst_ca_nhan_encrypted
- ADD Commission columns: commission_rate, commission_type, sales_target_yearly, biz_dev_region
- ADD BHXH columns: so_bhxh_encrypted, bhxh_registered_date, bhxh_province_code, bhyt_card_number, bhyt_expiry_date, bhyt_registered_hospital_code, bhyt_registered_hospital_name, tncn_registered
- CREATE all indexes: idx_employees_branch, idx_employees_dept, idx_employees_manager, idx_employees_status, idx_employees_grade, idx_employees_hired

**Task 3.2:** CREATE TABLE employee_dependents — See SPEC §11.6
**Task 3.3:** CREATE TABLE insurance_rate_config + unique index — See SPEC §11.7
**Task 3.4:** CREATE TABLE employee_salary_history + immutability rules — See SPEC §11.8
- CREATE RULE no_update_salary_history, CREATE RULE no_delete_salary_history

**Task 3.5:** CREATE TABLE employment_contracts — See SPEC §11.9

#### Day 4: Write Migration 000021 Down + 000027 Seed (Partial)
**Task 4.1:** Write `000021_hrm_employees_extended.down.sql` — See SPEC §12.3
- DROP TRIGGER trg_employees_set_code
- DROP FUNCTION fn_employees_set_code()
- DROP TABLE employment_contracts, employee_salary_history, insurance_rate_config, employee_dependents
- DROP COLUMN all 40+ columns explicitly (enumerate each)
- Risk note: HIGH — never rollback 000021 on production with data

**Task 4.2:** Write partial `000027_hrm_seed_data.up.sql` — See SPEC §11.26
- INSERT branches: HO (is_head_office=true, Hà Nội), HCM (false, TP.HCM)
- INSERT departments: AUDIT, TAX, HR, FIN, IT with correct dept_type
- INSERT branch_departments matrix: HO → all 5, HCM → AUDIT + TAX only
- INSERT insurance_rate_config: 2024 seed with KPCĐ 2%

**Task 4.3:** Run full migration test: 000001 through 000021 + 000027 partial on clean DB

#### Day 5: Employee CRUD API (Backend)
**Task 5.1:** Implement `EmployeeRepository` with sqlc queries — GET list with branch filter, GET by id, INSERT, UPDATE, soft delete
**Task 5.2:** Implement `EmployeeUseCase` with business logic:
- Branch scope enforcement for HoB/HR_STAFF roles
- Sensitive field masking (return `***` for unauthorized roles) — See SPEC §15.4
- Audit log on any mutation: EMPLOYEE_CREATED, EMPLOYEE_UPDATED, EMPLOYEE_TERMINATED
**Task 5.3:** Implement `EmployeeHandler` — bind to Gin routes
- GET /hrm/employees, POST /hrm/employees, GET /hrm/employees/:id, PUT /hrm/employees/:id
- DELETE /hrm/employees/:id (soft delete, SA only)
- POST /hrm/employees/:id/terminate (HR_MANAGER, CEO)
**Task 5.4:** GET /hrm/employees/:id/sensitive — decrypt + always audit log EMPLOYEE_PII_ACCESSED — See SPEC §15.4, §17.2.1

#### Day 6: My Profile + Organization API
**Task 6.1:** GET /my-profile, PUT /my-profile (limited fields only — see SPEC §13.2)
**Task 6.2:** Remaining organization endpoints — See SPEC §13.1:
- PUT /hrm/organization/branches/:id (SA, CHAIRMAN only)
- GET /hrm/organization/departments, GET :id, PUT :id
- GET /hrm/organization/branch-departments, POST, DELETE /:id
- GET /hrm/organization/org-chart (tree structure)

#### Day 7: Dependents, Salary History, Contracts API
**Task 7.1:** GET/POST/PUT/DELETE /hrm/employees/:id/dependents — See SPEC §13.3
**Task 7.2:** GET/POST /hrm/employees/:id/salary-history — See SPEC §13.15
- Enforce immutability at application layer (no update/delete endpoints)
- Audit log EMPLOYEE_SALARY_CHANGED on POST
**Task 7.3:** GET/POST/PUT /hrm/employees/:id/contracts — See SPEC §13.11
- POST /hrm/employees/:id/contracts/:cid/set-current
- GET /hrm/contracts/expiring?days=30
**Task 7.4:** GET/POST /hrm/insurance-config, GET /hrm/insurance-config/current — See SPEC §13.16

#### Day 8: Admin UI — Organization Page
**Task 8.1:** `/admin/hrm/organization` page — See SPEC §14.1
- Org chart tree view (branches → departments → roles)
- CRUD forms for branches (SA/CHAIRMAN gated)
- CRUD forms for departments
- branch_departments matrix editor

#### Day 9: Admin UI — Employee List + Detail
**Task 9.1:** `/admin/hrm/employees` — See SPEC §14.1
- DataTable with filter: branch, department, grade, status, search by name/code
- Export button (CSV at minimum)
- Role-gated create button
**Task 9.2:** `/admin/hrm/employees/new` — create form
**Task 9.3:** `/admin/hrm/employees/:id` — detail with tabs:
- Tab 1: Basic Info (employee code, branch, dept, grade, status, hired_date)
- Tab 2: Personal/PII (masked for unauthorized roles)
- Tab 3: Salary (visible only to HR_MANAGER, CEO, CHAIRMAN)
- Tab 4: Contracts (list with set-current action)
- Tab 5: Insurance/BHXH

#### Day 10: Self-Service Profile UI + Unit Tests
**Task 10.1:** `/my-profile` — See SPEC §14.3
- Display own employee data (non-sensitive fields)
- Edit form for limited fields: personal_email, personal_phone, current_address
**Task 10.2:** Unit tests — See SPEC §20.1:
- `TestEmployeeCodeFormat`: NV26-0001 format
- `TestInsuranceContribution`: BHXH calculation from rate_config
- `TestValidateGrade`: only 8 valid values
- `TestValidateCCCD`: 12-digit CCCD or 9-digit CMND

### Exit Criteria (Sprint 1 Complete)

- [x] Migrations 000020, 000021, 000027 (partial) pass lint + round-trip test — **DONE 2026-04-21**
- [x] All 32 endpoints from scope respond correctly per role — **DONE 2026-04-21 (Day 2: 14 org endpoints; Day 5-7: employee CRUD, sensitive PII, salary history, dependents, contracts, profile)**
- [x] POST /hrm/employees generates employee_code NV{YY}-{SEQ4} via trigger — **DONE (trigger fn_employees_set_code in 000021)**
- [x] PII fields (cccd_encrypted, mst_ca_nhan_encrypted, so_bhxh_encrypted, bank_account_encrypted) encrypted at rest, returned as `***` to unauthorized roles — **DONE 2026-04-22 (pkg/crypto/aesgcm.go, usecase/sensitive.go)**
- [x] GET /hrm/employees/:id/sensitive always writes EMPLOYEE_PII_ACCESSED to audit_logs — **DONE 2026-04-22 (fail-closed implementation)**
- [x] HoB HCM cannot see HO employees (branch scope enforced) — **DONE (empScopeBranch in usecase/employee.go)**
- [x] salary_history: INSERT works, UPDATE returns nothing (immutability rule), DELETE returns nothing — **DONE (PostgreSQL RULE no_update_salary_history + no_delete_salary_history)**
- [x] /admin/hrm/employees list renders, filters work, tabs load — **DONE 2026-04-22 (Phase 4 frontend)**
- [x] /my-profile loads and limited-field edit works — **DONE 2026-04-22 (Phase 4 frontend)**
- [ ] `make lint test` passes (no compile errors, no test failures) — *pending full backend + frontend integration test run*

### Phase 4 Frontend Implementation (2026-04-22)

**Files created:**
- `apps/web/src/services/hrm/employee.ts` — Full TypeScript service layer (all employee/profile/sensitive/salary/dependent/contract APIs)
- `apps/web/src/app/(dashboard)/admin/hrm/employees/page.tsx` — Employee list with search, pagination, role-gated delete
- `apps/web/src/app/(dashboard)/admin/hrm/employees/new/page.tsx` — Create employee form page
- `apps/web/src/app/(dashboard)/admin/hrm/employees/[id]/page.tsx` — Employee detail with 4 tabs (Info, Dependents, Contracts, Salary History)
- `apps/web/src/app/(dashboard)/my-profile/page.tsx` — Self-service profile page with limited-field edit
- `apps/web/src/components/hrm/employee-form.tsx` — `CreateEmployeeForm` + `UpdateEmployeeForm` with react-hook-form + zod
- `apps/web/src/components/hrm/sensitive-modal.tsx` — PII access confirmation dialog + decrypt view + update form
- `apps/web/src/components/hrm/dependent-section.tsx` — Dependent CRUD with inline dialogs
- `apps/web/src/components/hrm/contract-section.tsx` — Contract CRUD + terminate with dialogs
- `apps/web/src/components/hrm/salary-history-section.tsx` — Immutable salary history with create form (write roles: SA, CEO, HR_MANAGER)

**TypeScript check:** 0 errors in HRM files (`npx tsc --noEmit` — pre-existing errors in commissions/my/page.tsx unrelated to HRM work)

### Risk Items + Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| AES-256-GCM key not set in env | Medium | High | Fail-fast on startup: `panic("HRM_ENCRYPTION_KEY not set")` |
| employee_code trigger race condition | Low | Medium | Retry loop in trigger already handles it (See SPEC §4.8) |
| 40+ columns ALTER TABLE on existing employees | Medium | Medium | Use `IF NOT EXISTS` on all columns; test on copy of prod first |
| Branch scope bypass | Low | High | Integration test: HoB HCM token → GET /hrm/employees?branch_id=HO → 403 |

### Rollback Strategy

```bash
# Rollback Sprint 1 (dev/staging only — never on production with data)
make migrate-down VERSION=000018   # reverts 000027, 000021, 000020
git revert HEAD~N                  # revert all Sprint 1 commits
```

### Integration Test Plan

```
POST /hrm/employees as HR_MANAGER → 201, employee_code = "NV26-XXXX"
GET  /hrm/employees/:id as HR_MANAGER → 200, all fields visible
GET  /hrm/employees/:id as HoB HCM (own branch) → 200
GET  /hrm/employees/:id as HoB HCM (other branch) → 403
GET  /hrm/employees/:id/sensitive as CEO → 200, audit log written
GET  /hrm/employees/:id/sensitive as PARTNER → 403
POST /hrm/employees/:id/salary-history → 201, immutability rule fires on UPDATE attempt
```

---

## Sprint 2: Provisioning + Certifications/Training (Weeks 3–4, ~35 SP)

### Goals
Full user provisioning workflow operational (HCM 2-step, HO direct, emergency). Certifications and CPE tracking functional. Offboarding checklists created. Notifications wired for provisioning and cert expiry.

### Scope
**Migrations:** 000022, 000025
**Tables:** certifications, training_courses, training_records, cpe_requirements_by_role, user_provisioning_requests, offboarding_checklists
**API groups:** certifications (5), training/CPE (8), provisioning (9), offboarding (5) — See SPEC §13.4, §13.5, §13.12, §13.13
**UI pages:** /admin/hrm/provisioning, /hrm/provisioning, /hrm/certifications, /hrm/training, /my-profile/certifications, /my-profile/training
**Notifications:** Events 2, 9, 10, 11, 12, 13 — See SPEC §16.1

### Dependencies
- Sprint 1 complete and merged to main
- Migrations 000020, 000021, 000027 (partial) applied
- Notification infrastructure (outbox pattern) available — See CLAUDE.md §Architecture Patterns

### Daily Task Breakdown

#### Day 11: Migration 000022 — Professional Tables
**Task 11.1:** Write `000022_hrm_professional.up.sql` — See SPEC §11.10–§11.13
- CREATE TABLE certifications with all indexes
- CREATE TABLE training_courses (course_code UNIQUE)
- CREATE TABLE training_records with status CHECK, indexes
- CREATE TABLE cpe_requirements_by_role

**Task 11.2:** Write `000022_hrm_professional.down.sql` — See SPEC §12.4
- DROP TABLE cpe_requirements_by_role, training_records, training_courses, certifications

**Task 11.3:** Add CPE requirements seed to 000027 — See SPEC §11.26, §6.4
- PARTNER/SENIOR_AUDITOR/JUNIOR_AUDITOR: 40h/year VN_CPA via VACPA
- PARTNER: 40h ACCA, 40h CIA, 20h CFA

#### Day 12: Migration 000025 — Provisioning Tables
**Task 12.1:** Write `000025_hrm_provisioning.up.sql` — See SPEC §11.22, §11.23
- CREATE TABLE user_provisioning_requests with all columns and expires_at
- CREATE UNIQUE INDEX uidx_provisioning_pending WHERE status = 'PENDING'
- CREATE TABLE offboarding_checklists with JSONB items

**Task 12.2:** Write `000025_hrm_provisioning.down.sql` — See SPEC §12.7
**Task 12.3:** Run lint + round-trip test on migrations 000022 and 000025

#### Day 13: Certifications API
**Task 13.1:** Implement certifications repository + use case + handler — See SPEC §13.4
- GET /hrm/employees/:id/certifications (HR, CEO, PARTNER, HoB, self)
- POST /hrm/employees/:id/certifications (HR_MANAGER, self) — audit log CERT_ADDED
- PUT /hrm/employees/:id/certifications/:cert_id (HR_MANAGER, self)
- DELETE /hrm/employees/:id/certifications/:cert_id (HR_MANAGER only)
- GET /hrm/certifications/expiring?days=60&cert_type=VN_CPA (HR, CEO, PARTNER)

#### Day 14: Training & CPE API
**Task 14.1:** Training courses — See SPEC §13.5
- GET /hrm/training/courses (ALL), POST (HR_MANAGER, SA), PUT/:id (HR_MANAGER)
**Task 14.2:** Training records
- GET /hrm/employees/:id/training, POST, PUT/:rec_id
- GET /hrm/employees/:id/cpe-summary — calculate hours vs cpe_requirements_by_role
- GET /hrm/training/cpe-summary — aggregate for all employees (HR, CEO, CHAIRMAN)

#### Day 15: Provisioning API — Core Flow
**Task 15.1:** Implement provisioning repository — See SPEC §13.12, §8.2
- GET /hrm/user-provisioning-requests (SA, HR_MANAGER, HoB)
- POST /hrm/user-provisioning-requests — validate no duplicate PENDING per employee_id
- GET /hrm/user-provisioning-requests/:id

**Task 15.2:** Implement approval steps — See SPEC §8.3
- POST /:id/branch-approve (HoB only) → update approval_level = 2, send notification to HR_MANAGER
- POST /:id/branch-reject (HoB only) → status = REJECTED, notify requester
- POST /:id/hr-approve (HR_MANAGER) → status = APPROVED, notify SA
- POST /:id/hr-reject (HR_MANAGER) → status = REJECTED, notify requester + HoB
- POST /:id/cancel (requester, SA) → status = CANCELLED (not EXECUTED)

#### Day 16: Provisioning Execute + Atomic Account Creation
**Task 16.1:** POST /:id/execute (SA only) — atomic transaction:
1. Validate status = APPROVED
2. Create user account in users table
3. Assign requested_role
4. Link employee.user_id → new user.id
5. Update request status = EXECUTED, executed_by, executed_at
6. Trigger welcome email notification (event 13)
7. Write audit log PROVISIONING_EXECUTED with user_id created, role assigned — See SPEC §17.2.3

**Task 16.2:** Emergency flow: if is_emergency = true, skip approval steps, execute immediately
**Task 16.3:** Business rule enforcement: SUPER_ADMIN and CHAIRMAN roles cannot be provisioned via this flow

#### Day 17: Offboarding API
**Task 17.1:** Implement offboarding endpoints — See SPEC §13.13
- GET /hrm/offboarding (HR, CEO, SA)
- POST /hrm/offboarding — initiate with JSONB template from SPEC §9.4
- GET /hrm/offboarding/:id
- PUT /hrm/offboarding/:id/items/:key — update individual checklist item
- POST /hrm/offboarding/:id/complete (HR_MANAGER) — write OFFBOARDING_COMPLETED audit log

#### Day 18: Admin Provisioning UI
**Task 18.1:** `/admin/hrm/provisioning` — See SPEC §14.1
- List all requests with status filter
- Detail view with approval history
- Execute button for SA
**Task 18.2:** `/hrm/provisioning` — See SPEC §14.2
- HoB view: branch-scoped requests, approve/reject buttons

#### Day 19: Certifications + Training UI
**Task 19.1:** `/hrm/certifications` — company-wide cert list, expiring alerts banner — See SPEC §14.2
**Task 19.2:** `/hrm/training` — CPE dashboard, course catalog — See SPEC §14.2
**Task 19.3:** `/my-profile/certifications` — employee's own certs, add form — See SPEC §14.3
**Task 19.4:** `/my-profile/training` — training history, CPE progress bar (X/40h)

#### Day 20: Notifications + Integration Tests
**Task 20.1:** Wire notifications for provisioning flow — See SPEC §16.1 events 9–13
**Task 20.2:** Wire cert expiry notification — event 2 (daily cron: 60d/30d/7d before expiry)
**Task 20.3:** Integration tests — See SPEC §20.2.4:
- TestProvisioningHCMFlow: full 5-step flow
- TestProvisioningHOFlow: direct SA execute
- TestProvisioningEmergency: is_emergency=true skips approval
- TestProvisioningDuplicatePending: second PENDING → 409 DUPLICATE_PENDING_REQUEST
- TestCPEProgress: add training record → CPE summary shows correct hours

### Exit Criteria (Sprint 2 Complete)

- [ ] Migrations 000022, 000025 pass lint + round-trip
- [ ] Provisioning HCM 2-step flow works end-to-end (HoB approve → HR approve → SA execute)
- [ ] Account creation is atomic (user + role + employee link in single transaction)
- [ ] SUPER_ADMIN / CHAIRMAN roles blocked from provisioning flow
- [ ] CPE hours calculated correctly per year and per role requirement
- [ ] Cert expiry endpoint returns items sorted by days remaining
- [ ] All 5 integration tests pass
- [ ] `make lint test` passes

### Risk Items + Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Atomic user creation fails mid-transaction | Low | High | Wrap in DB transaction; rollback on any step failure |
| Duplicate PENDING not caught | Low | High | UNIQUE INDEX uidx_provisioning_pending enforces at DB level |
| Request expires (30 days) while PENDING | Medium | Low | Scheduled job: daily scan expires_at < now() → status = CANCELLED |

---

## Sprint 3: Performance Reviews + Independence (Week 5, ~20 SP)

### Goals
Performance review workflow (manager → employee acknowledge). Independence declarations for annual and per-engagement. Compliance tracking for VSA 220/ISA 220 requirements.

### Scope
**Migrations:** 000023
**Tables:** performance_reviews, engagement_peer_reviews, independence_declarations
**API groups:** performance (8), independence (6) — See SPEC §13.6, §13.7
**UI pages:** /hrm/performance, /hrm/independence, /my-performance, /my-independence, /my-team/performance, /my-team/independence
**Notifications:** Events 17, 18, 20, 21 — See SPEC §16.1

### Dependencies
- Sprint 1 + 2 complete
- engagements table exists (for per-engagement declarations)
- PARTNER role users exist for acknowledgment flow

### Daily Task Breakdown

#### Day 21: Migration 000023
**Task 21.1:** Write `000023_hrm_performance.up.sql` — See SPEC §11.14–§11.16
- CREATE TABLE performance_reviews with uidx_review_employee_period UNIQUE
- CREATE TABLE engagement_peer_reviews with chk_peer_review_self CHECK
- CREATE TABLE independence_declarations with uidx_independence_annual, chk_annual_year, chk_engagement_ref

**Task 21.2:** Write `000023_hrm_performance.down.sql` — See SPEC §12.5

#### Day 22: Independence API
**Task 22.1:** Independence declarations — See SPEC §13.7
- GET /hrm/independence (HR, CEO, CHAIRMAN, PARTNER — with scoping)
- GET /hrm/employees/:id/independence (HR, CEO, PARTNER, self)
- POST /hrm/independence — creates ANNUAL or PER_ENGAGEMENT declaration
  - On has_conflict=true: status = PENDING, notify PARTNER (event 17)
  - On has_conflict=false: status = CLEAN immediately
- GET /hrm/independence/:id
- POST /hrm/independence/:id/acknowledge (PARTNER only) → status = CONFLICT_RESOLVED
- GET /hrm/independence/annual-status — See SPEC §13.7

#### Day 23: Performance Review API
**Task 23.1:** Performance reviews — See SPEC §13.6
- GET /hrm/performance/reviews (HR, CEO, CHAIRMAN — all)
- GET /hrm/employees/:id/performance (HR, CEO, PARTNER, HoB, self)
- POST /hrm/employees/:id/performance (HR_MANAGER, PARTNER, AUDIT_MANAGER) → creates DRAFT
- PUT /hrm/performance/reviews/:id (reviewer only, DRAFT status)
- POST /hrm/performance/reviews/:id/submit (reviewer) → SUBMITTED, notify employee (event 21)
- POST /hrm/performance/reviews/:id/acknowledge (employee) → ACKNOWLEDGED
**Task 23.2:** Peer reviews
- GET /hrm/performance/peer-reviews (HR, CEO, PARTNER)
- POST /hrm/performance/peer-reviews (PARTNER, SENIOR, JUNIOR) — is_anonymous=true default

#### Day 24: Performance + Independence UI
**Task 24.1:** `/hrm/performance` — period selector, review status board — See SPEC §14.2
**Task 24.2:** `/my-performance` — employee's own reviews — See SPEC §14.3
**Task 24.3:** `/hrm/independence` — compliance tracker (annual-status endpoint)
**Task 24.4:** `/my-independence` — own declarations + new declaration form
**Task 24.5:** `/my-team/performance` — reviews I need to complete
**Task 24.6:** `/my-team/independence` — team declaration status (PARTNER view)

#### Day 25: Notifications + Integration Tests
**Task 25.1:** Wire notifications events 17, 18, 20, 21 — See SPEC §16.1
**Task 25.2:** Integration tests — See SPEC §20.2.5:
- TestIndependenceCleanDeclaration: no conflict → CLEAN, no notification
- TestIndependenceConflictFlow: has_conflict=true → PENDING → Partner acknowledges → CONFLICT_RESOLVED
- TestAnnualDeclarationUnique: second ANNUAL same year → 409

### Exit Criteria (Sprint 3 Complete)

- [ ] Migration 000023 passes lint + round-trip
- [ ] ANNUAL unique constraint: one declaration per employee per year
- [ ] Conflict flow: has_conflict=true → Partner notification → acknowledgment required
- [ ] Per-engagement declaration has engagement_id FK (enforced by chk_engagement_ref)
- [ ] Performance review state machine: DRAFT → SUBMITTED → ACKNOWLEDGED → FINAL
- [ ] Peer review self-check constraint fires (reviewer_id ≠ reviewee_id)
- [ ] `make lint test` passes

---

## Sprint 4: Leave + OT + Holidays (Week 6, ~25 SP)

### Goals
Leave management with balance tracking. OT tracking with 300h/year hard cap. Holiday calendar seeded 2026–2030. All time-related approvals with notifications.

### Scope
**Migrations:** 000024, 000027 (remaining: holidays + CPE reqs)
**Tables:** holidays, leave_balances, leave_requests, ot_requests, VIEW employee_ot_summary_year; ALTER timesheets
**API groups:** holidays (4), leave (10), OT (7) — See SPEC §13.8, §13.9, §13.10
**UI pages:** /hrm/leave/requests, /hrm/leave/calendar, /hrm/overtime/requests, /my-leave, /my-overtime, /admin/hrm/holidays
**Notifications:** Events 3–8, 25 — See SPEC §16.1
**Scheduled jobs:** Contract expiry alert (event 1), BHYT expiry (event 25), CPE reminder (event 19)

### Dependencies
- Sprint 1–3 complete
- timesheets table exists (ALTER)

### Daily Task Breakdown

#### Day 26: Migration 000024
**Task 26.1:** Write `000024_hrm_time_leave.up.sql` — See SPEC §11.17–§11.21
- CREATE TABLE holidays with generated year column
- CREATE TABLE leave_balances with uidx_leave_balance UNIQUE, chk_days_non_negative
- CREATE TABLE leave_requests with all status CHECKs
- ALTER TABLE timesheets ADD ot_hours, ot_approved, ot_request_id
- CREATE TABLE ot_requests with chk_ot_times, chk_ot_hours
- CREATE VIEW employee_ot_summary_year (300h cap calculation)

**Task 26.2:** Write `000024_hrm_time_leave.down.sql` — See SPEC §12.6
**Task 26.3:** Complete `000027_hrm_seed_data.up.sql` — See SPEC §11.26
- INSERT holidays 2026 (12 national holidays) + 2027–2030 estimated dates

#### Day 27: Leave API
**Task 27.1:** Leave balance endpoints — See SPEC §13.8
- GET /hrm/leave/balances (HR, CEO, HoB)
- GET /hrm/employees/:id/leave-balance (HR, CEO, HoB, manager, self)
- PUT /hrm/employees/:id/leave-balance (HR_MANAGER only) — audit log LEAVE_BALANCE_ADJUSTED
**Task 27.2:** Leave request flow
- GET /hrm/leave/requests (HR, CEO, HoB — with branch scope)
- GET /my-leave/requests (own only), POST /my-leave/requests
- On POST: auto-increment leave_balance.pending_days
- POST /hrm/leave/requests/:id/approve — increment used_days, decrement pending_days (atomic)
- POST /hrm/leave/requests/:id/reject — decrement pending_days only
- POST /my-leave/requests/:id/cancel (PENDING only)
- GET /my-team/leave (PARTNER, HoB, CEO), GET /hrm/leave/calendar (ALL)

#### Day 28: OT API
**Task 28.1:** OT endpoint implementations — See SPEC §13.9
- GET /hrm/overtime/requests (HR, CEO, HoB), GET /my-overtime/requests
- POST /my-overtime/requests — pre-validate: annual cap would not exceed 300h if approved
- POST /hrm/overtime/requests/:id/approve — enforce 300h cap; return OT_ANNUAL_CAP_EXCEEDED if exceeded
  - On approve: write OT_CAP_WARNING audit log if total > 270h
- POST /hrm/overtime/requests/:id/reject
- GET /hrm/overtime/summary (HR, CEO, CHAIRMAN), GET /my-overtime/summary
**Task 28.2:** Holidays endpoints — See SPEC §13.10
- GET /hrm/holidays?year=2026 (ALL), POST (SA, HR_MANAGER), PUT/:id, DELETE/:id (SA)

#### Day 29: Leave + OT UI
**Task 29.1:** `/my-leave` — balance summary cards + leave request history — See SPEC §14.3
**Task 29.2:** `/my-leave/new` — leave request form with date picker, type selector
**Task 29.3:** `/my-overtime` — OT summary (used/300h progress bar) + history
**Task 29.4:** `/my-overtime/new` — OT registration form
**Task 29.5:** `/hrm/leave/requests` — manager/HR approval queue with approve/reject buttons
**Task 29.6:** `/hrm/leave/calendar` — month view calendar with leave markers
**Task 29.7:** `/admin/hrm/holidays` — calendar CRUD, year switcher

#### Day 30: Notifications + Scheduled Jobs + Integration Tests
**Task 30.1:** Wire leave notifications (events 3–5) — See SPEC §16.1
**Task 30.2:** Wire OT notifications (events 6–8 including 250h cap warning)
**Task 30.3:** Scheduled jobs — See SPEC §9.3, §16.1:
- Daily 8:00 AM: scan employment_contracts end_date within 30 days → notification event 1
- Daily: BHYT expiry scan → event 25
- October 31 annual: CPE deficit scan → event 19
**Task 30.4:** Integration tests — See SPEC §20.2.2, §20.2.3:
- TestLeaveApprovalFlow: balance updated atomically
- TestLeaveRejectionFlow: pending_days rolled back
- TestOTCapEnforcement: 290h + 15h → 422; 290h + 10h → 200; 300h + 1h → 422

### Exit Criteria (Sprint 4 Complete)

- [ ] Migration 000024 + 000027 (complete) pass lint + round-trip
- [ ] OT cap 300h/year enforced at API level (not just DB)
- [ ] Leave balance updates are atomic (approve and reject)
- [ ] Holidays 2026–2030 seeded and calendar renders correctly
- [ ] All 3 integration tests pass (leave approval, leave rejection, OT cap)
- [ ] Scheduled jobs registered and fire at correct times
- [ ] `make lint test` passes

---

## Sprint 5: Expenses + Reports + Polish (Week 7, ~25 SP)

### Goals
Expense claim workflow end-to-end. All 10 HRM standard reports with Excel/CSV export. Security review, performance validation, E2E tests.

### Scope
**Migrations:** 000026
**Tables:** expense_claims (+ trigger), expense_claim_items
**API groups:** expenses (13), reports (10), insurance config (3) — See SPEC §13.14, §13.16, §19.1
**UI pages:** /my-expenses, /hrm/expenses, /admin/hrm/reports, /admin/hrm/insurance-config
**Notifications:** Events 14–16 — See SPEC §16.1
**Non-functional:** p95 < 500ms, security review

### Dependencies
- Sprint 1–4 complete
- Leave/OT data exists for reports

### Daily Task Breakdown

#### Day 31: Migration 000026
**Task 31.1:** Write `000026_hrm_expenses.up.sql` — See SPEC §11.24, §11.25
- CREATE TABLE expense_claims with fn_expense_claims_set_number trigger (PC{YY}-{SEQ4})
- CREATE TABLE expense_claim_items

**Task 31.2:** Write `000026_hrm_expenses.down.sql` — See SPEC §12.8

#### Day 32: Expenses API
**Task 32.1:** Self-service expense endpoints — See SPEC §13.14
- GET /my-expenses, POST /my-expenses (DRAFT)
- PUT /my-expenses/:id (owner, DRAFT only), DELETE item, POST item, POST submit
**Task 32.2:** Approval flow
- POST /hrm/expenses/:id/manager-approve → MANAGER_APPROVED, notify HR (event 14)
- POST /hrm/expenses/:id/manager-reject, POST /hrm/expenses/:id/hr-approve
- POST /hrm/expenses/:id/hr-reject, POST /hrm/expenses/:id/mark-paid → PAID, notify employee (event 16)
- GET /hrm/expenses (HR, CEO, CHAIRMAN), GET /hrm/expenses/summary

#### Day 33: Expense UI + Notifications
**Task 33.1:** `/my-expenses` — claim list with status badges — See SPEC §14.3
**Task 33.2:** `/my-expenses/new` — multi-step form: claim info → line items (with receipt upload) → submit
**Task 33.3:** `/hrm/expenses` — approval queue (See SPEC §14.2)
**Task 33.4:** Wire expense notifications events 14–16 — See SPEC §16.1

#### Day 34: Reports API (5 reports)
**Task 34.1:** CPE Compliance Report `/hrm/reports/cpe-compliance` — See SPEC §19.1 Report 1
**Task 34.2:** Leave Usage Report `/hrm/reports/leave-usage` — Report 2
**Task 34.3:** OT Summary Report `/hrm/reports/ot-summary` — Report 3
**Task 34.4:** Headcount Report `/hrm/reports/headcount` — Report 4
**Task 34.5:** Contract Renewal Alert `/hrm/reports/contract-renewal` — Report 5

#### Day 35: Reports API (5 reports) + Export
**Task 35.1:** Cert Expiry Report `/hrm/reports/cert-expiry` — See SPEC §19.1 Report 6
**Task 35.2:** Independence Status `/hrm/reports/independence-status` — Report 7
**Task 35.3:** Expense Summary `/hrm/reports/expense-summary` — Report 8
**Task 35.4:** Performance Distribution `/hrm/reports/performance-distribution` — Report 9
**Task 35.5:** Salary Report `/hrm/reports/salary` (HR_MANAGER, CEO, CHAIRMAN) — Report 10
- Always write audit log SALARY_REPORT_GENERATED
**Task 35.6:** Excel + CSV export for all 10 reports — See SPEC §19.2

#### Day 36: Admin Reports UI + Insurance Config UI
**Task 36.1:** `/admin/hrm/reports` — dashboard with 10 report tiles, filter panels, export buttons — See SPEC §14.1
**Task 36.2:** `/admin/hrm/insurance-config` — insurance rate history, add new config form

#### Day 37: Security Review + Performance
**Task 37.1:** Security review checklist — See SPEC §18:
- All 4 encrypted fields use AES-256-GCM, key from env only
- Audit logs written for every PII access
- Salary report audit log fires
- Rate limiting on sensitive endpoints (100 req/min standard, 10 req/min sensitive)
**Task 37.2:** Performance test: verify p95 < 500ms for:
- GET /hrm/employees (paginated 20/page)
- GET /hrm/reports/headcount
- GET /hrm/leave/calendar

#### Day 38: E2E Tests + Polish
**Task 38.1:** Playwright E2E tests — See SPEC §20.3:
- New employee onboarding journey
- Self-service leave flow
- Expense claim flow
- Independence declaration
- CPE tracking
- Provisioning HCM
- Performance review
**Task 38.2:** Smoke test existing modules (dashboard, login, CRM) — no regressions
**Task 38.3:** Zero TODO comments in production code paths
**Task 38.4:** Zero unused imports audit (`go vet ./...`)

### Exit Criteria (Sprint 5 Complete)

- [ ] Migration 000026 passes lint + round-trip
- [ ] All 10 reports return correct data with filters
- [ ] Salary report writes SALARY_REPORT_GENERATED audit log every time
- [ ] Excel + CSV export works for all 10 reports
- [ ] All 7 E2E test journeys pass on staging
- [ ] No regression in existing modules
- [ ] p95 API latency < 500ms for common queries (measured)
- [ ] Security review signed off: encryption, audit trail, PII masking
- [ ] `make lint test` passes (all suites)

---

## Weekly Review Cadence

| Day | Activity |
|---|---|
| Monday (start of sprint week) | Read SPEC §21 sprint goals, run `git log --oneline -20` to orient |
| Wednesday | Mid-sprint check: are exit criteria on track? Any SPEC ambiguities? |
| Friday | Sprint close: run full test suite, verify exit criteria checklist |

## Go/No-Go Decision Points

| Checkpoint | Condition to proceed |
|---|---|
| Before Sprint 2 | Sprint 1 exit criteria 100% complete, all tests green |
| Before Sprint 3 | Sprint 2 exit criteria 100% complete |
| Before Sprint 4 | Sprint 3 exit criteria 100% complete |
| Before Sprint 5 | Sprint 4 exit criteria 100% complete |
| Before production deploy | Sprint 5 exit criteria 100% complete + security review signed off |

> **If exit criteria not met:** Do not advance sprint. Fix blockers. Re-run checklist. Only proceed when all boxes checked.
