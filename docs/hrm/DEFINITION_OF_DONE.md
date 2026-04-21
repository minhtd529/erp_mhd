# HRM Definition of Done
## ERP System — MDH Audit Firm
**Version:** 1.0 | **Based on:** HRM_SPEC_v1.4.md | **Rule:** No feature ships without satisfying all applicable DoD items.

> **Migration numbering note (2026-04-21):** SPEC v1.4 originally planned HRM migrations 000019–000026. After implementation started, migration 000019 was already occupied by `000019_working_papers`. All HRM migrations shifted +1: they now occupy 000020–000027. All references in this document have been updated accordingly.

---

> This document defines the minimum quality bar for every artifact produced during HRM implementation. Before marking any task DONE, run through the applicable checklist. Partial completion = not done.

---

## Per Migration Checklist

Apply to each of the 8 migrations: 000020, 000021, 000022, 000023, 000024, 000025, 000026, 000027.

- [ ] Both `.up.sql` and `.down.sql` files present in `apps/api/migrations/`
- [ ] Filename follows `000NNN_snake_case.{up,down}.sql` format
- [ ] `make migrate-lint` passes (no lint errors from `scripts/migration-lint.sh`)
- [ ] Round-trip test succeeds: `make migrate-up && make migrate-down && make migrate-up`
- [ ] No data loss on rollback (verified on dev DB with test data)
- [ ] Migration tested on a **clean DB** (migrate from 000001 to head, not just incremental)
- [ ] Seed data (if any) is ONLY in migration 000027, not in 000020–000026
- [ ] Indexes created for ALL foreign key columns
- [ ] CHECK constraints defined for ALL enum/status columns
- [ ] No modification to any previously committed migration (immutability rule)
- [ ] Rollback risk documented in migration file header comment
- [ ] `make migrate-lint` passes again after writing down.sql

**Migration-specific items:**

| Migration | Extra DoD Items |
|---|---|
| 000020 | uidx_branches_head_office WHERE is_head_office = true exists |
| 000021 | fn_employees_set_code() trigger fires on INSERT, generates NV{YY}-{SEQ4} |
| 000021 | salary_history immutability: UPDATE and DELETE rules return nothing |
| 000021 | All 4 encrypted columns use `_encrypted` suffix |
| 000022 | cpe_requirements_by_role table created |
| 000023 | uidx_independence_annual partial index exists (ANNUAL per employee per year) |
| 000023 | chk_peer_review_self constraint: reviewer_id ≠ reviewee_id |
| 000024 | employee_ot_summary_year VIEW created |
| 000024 | ALTER timesheets adds ot_hours, ot_approved, ot_request_id |
| 000025 | uidx_provisioning_pending WHERE status = 'PENDING' exists |
| 000026 | fn_expense_claims_set_number() trigger creates PC{YY}-{SEQ4} |
| 000027 | 2 branches (HO, HCM), 5 departments, 7-entry branch_departments matrix |
| 000027 | insurance_rate_config 2024 seed with kpcd_employer_pct = 2.00 |
| 000027 | Holidays 2026–2030 seeded (at minimum 2026: 12 national holidays) |
| 000027 | CPE requirements seeded for PARTNER, SENIOR_AUDITOR, JUNIOR_AUDITOR |

---

## Per API Endpoint Checklist

Apply to every endpoint in SPEC §13 (80+ total).

- [ ] HTTP method and path match SPEC §13 exactly (no invented paths)
- [ ] Permission check via `middleware.RequireRole(...)` — no inline role checks in handlers
- [ ] Request body validated — return 400 with specific field errors on bad input
- [ ] Path parameters validated — return 404 for non-existent IDs (not 500)
- [ ] Error handling wraps underlying errors: `fmt.Errorf("context: %w", err)`
- [ ] Error logged with `log.Printf` before returning any 500 response
- [ ] Audit log entry written for all mutations (CREATE, UPDATE, DELETE, state transitions)
- [ ] Integration test covers: happy path + at least 1 error path (e.g., 403, 404, 422)
- [ ] Response format matches convention: `{"data": ...}` or `{"data": [], "meta": {...}}`
- [ ] Correct HTTP status: 201 for POST creates, 204 for DELETE, 422 for business rule violations
- [ ] No `SELECT *` in any underlying query

**Endpoint-specific items:**

| Endpoint | Extra DoD Items |
|---|---|
| GET /hrm/employees | Branch scope enforced for HEAD_OF_BRANCH and HR_STAFF |
| GET /hrm/employees/:id | Sensitive fields masked (`***`) for unauthorized roles |
| GET /hrm/employees/:id/sensitive | ALWAYS writes EMPLOYEE_PII_ACCESSED to audit_logs; HR_MANAGER, CEO, CHAIRMAN only |
| POST /hrm/employees | employee_code auto-generated via DB trigger (not in request body) |
| POST /hrm/employees/:id/salary-history | Immutability enforced — no update/delete endpoints exist |
| POST /my-leave/requests | leave_balance.pending_days incremented atomically |
| POST /hrm/leave/requests/:id/approve | leave_balance.used_days +, pending_days - (atomic transaction) |
| POST /hrm/leave/requests/:id/reject | leave_balance.pending_days - (rollback atomic) |
| POST /hrm/overtime/requests/:id/approve | Returns 422 OT_ANNUAL_CAP_EXCEEDED if total > 300h |
| POST /hrm/user-provisioning-requests | Returns 409 DUPLICATE_PENDING_REQUEST if PENDING exists for employee |
| POST /hrm/user-provisioning-requests/:id/execute | Atomic: user created + role assigned + employee.user_id linked in one transaction |
| POST /hrm/independence | has_conflict=true → notifies PARTNER; has_conflict=false → status=CLEAN |
| GET /hrm/employees/:id/salary-history | HR_MANAGER, CEO, CHAIRMAN only |
| GET /hrm/reports/salary | Always writes SALARY_REPORT_GENERATED audit log |
| DELETE /hrm/employees/:id | Soft delete only (is_deleted=true or status=TERMINATED) — SA only |

---

## Per UI Page Checklist

Apply to every UI page in SPEC §14 (30+ pages).

- [ ] Page route matches SPEC §14 exactly
- [ ] **Loading state:** skeleton or spinner visible while data fetches
- [ ] **Empty state:** Vietnamese message + icon when no records found
- [ ] **Error state:** Vietnamese error message + retry button when API fails
- [ ] Responsive layout works at ≥ 375px width (test at mobile viewport)
- [ ] Keyboard navigation: Tab order is sensible, forms reachable without mouse
- [ ] Permission-gated: redirect to 403 page or hide unauthorized actions
- [ ] Forms use react-hook-form + zod schema validation
- [ ] Validation errors display per field (not just a generic top-level error)
- [ ] Success toast (Vietnamese) shown on successful mutation
- [ ] No console errors in browser DevTools after normal interactions
- [ ] Tailwind classes only — no inline styles
- [ ] shadcn/ui components used (no custom button/input reinventions)
- [ ] API calls through `services/*.ts` — no direct axios/fetch in components

**Page-specific items:**

| Page | Extra DoD Items |
|---|---|
| /admin/hrm/employees | Filter works (branch, dept, grade, status, name/code search) |
| /admin/hrm/employees/:id | All 5 tabs render (Basic, PII, Salary, Contract, Insurance) |
| /admin/hrm/employees/:id | Salary tab: hidden/empty for unauthorized roles |
| /admin/hrm/employees/:id | PII tab: shows `***` for encrypted fields to unauthorized roles |
| /admin/hrm/provisioning | Execute button visible only to SA; approval buttons visible to HoB/HR |
| /admin/hrm/reports | All 10 report tiles present; Excel + CSV export buttons work |
| /admin/hrm/holidays | Calendar view with CRUD, year switcher |
| /my-leave | Balance summary cards for each leave type visible |
| /my-leave/new | Date picker prevents past-date end before start |
| /my-overtime | Progress bar shows used/300h annual cap |
| /my-expenses/new | Multi-step form: claim info → line items → submit |
| /hrm/leave/requests | Approve/reject buttons visible only to managers and HR |
| /hrm/leave/calendar | Month view with leave markers per employee |
| /hrm/certifications | Expiring-soon alerts banner at top |
| /my-independence | Annual declaration status badge visible |
| /admin/hrm/organization | Org chart tree renders correctly |

---

## Per Feature (End-to-End) Checklist

Apply when an entire feature (migration + backend + frontend) is declared complete.

- [ ] Migration applied to dev DB without errors
- [ ] Backend endpoints tested via `curl` or HTTPie — correct responses confirmed
- [ ] Frontend connects successfully to backend (no CORS errors, no 401/403 on legitimate calls)
- [ ] Permission matrix respected — tested with at least 3 different roles:
  - Role that should have full access (e.g., HR_MANAGER)
  - Role that should have partial access (e.g., HEAD_OF_BRANCH — branch-scoped)
  - Role that should have no access (e.g., JUNIOR_AUDITOR)
- [ ] Encryption roundtrip works (if applicable): encrypt → store → retrieve → decrypt → matches
- [ ] Audit log entries appear in audit_logs for all mutations and PII access
- [ ] Notification sent (if applicable per SPEC §16.1) — verified in notification log/table
- [ ] No TODO comments left in production code paths (search: `grep -r "TODO\|FIXME\|HACK" apps/api/internal/hrm/ apps/web/src/app/*hrm*`)

**Feature-specific items:**

| Feature | Roles to Test | Extra DoD |
|---|---|---|
| Employee CRUD | HR_MANAGER (ALLOW), HoB own branch (ALLOW scoped), HoB other branch (DENY), PARTNER (DENY write) | employee_code trigger verified |
| PII access | HR_MANAGER (ALLOW + audit log), CEO (ALLOW + audit log), PARTNER (DENY 403) | Audit log always present |
| Salary history | HR_MANAGER (ALLOW write), CEO (ALLOW write), HR_STAFF (DENY) | Immutability: PUT/DELETE return nothing |
| Provisioning HCM | HoB approve (step 1), HR_MANAGER approve (step 2), SA execute | Atomic: user+role+link in one TX |
| Provisioning HO | CEO or HR direct, SA execute | Emergency flag skips approval |
| Leave approval | Manager APPROVE/REJECT, HR_MANAGER for UNPAID | Balance atomic update |
| OT approval | Manager APPROVE, 300h cap at 300h (not 301h) | OT_CAP_WARNING at 270h |
| Independence ANNUAL | ALL employees can create, only 1 per year | Partner notified on conflict |
| Independence PER_ENGAGEMENT | Must have engagement_id | Conflict blocks engagement start |
| Expense claim | Employee creates, Manager L1, HR L2, Finance/HR mark paid | Claim number PC{YY}-{SEQ4} |
| CPE tracking | Employee adds training record, CPE summary updates | Correct hours per role requirement |
| Certifications | Employee adds own cert, HR_MANAGER updates, expiry alert at 60/30/7 days | Expiry sorted ascending |

---

## Per Sprint Checklist

Apply before declaring a sprint complete (gate to next sprint).

- [ ] All features listed in IMPLEMENTATION_PLAN.md §Sprint N scope are shipped
- [ ] All migration DoD items passed for migrations in this sprint
- [ ] All API endpoint DoD items passed for endpoints in this sprint
- [ ] All UI page DoD items passed for pages in this sprint
- [ ] No regression in existing functionality — smoke test:
  - Login works (all roles)
  - Dashboard loads
  - CRM/Engagement pages load (if previously implemented)
- [ ] SPEC_COMPLIANCE_CHECKLIST.md sections relevant to this sprint audited
- [ ] Zero TODO/FIXME/HACK comments in production code paths
- [ ] Zero unused imports: `go vet ./...` and `eslint apps/web/src/`
- [ ] `make lint test` passes (Go lint + TypeScript lint + all tests)
- [ ] CHANGELOG.md updated with sprint deliverables
- [ ] No console errors in browser DevTools for new UI pages

**Per-sprint extra items:**

| Sprint | Extra DoD |
|---|---|
| Sprint 1 | AES-256-GCM encryption working: encrypt on write, decrypt on /sensitive endpoint |
| Sprint 1 | Branch scope: HoB HCM cannot list HO employees |
| Sprint 2 | Provisioning execute: atomic transaction verified (rollback if any step fails) |
| Sprint 2 | Emergency provisioning writes PROVISIONING_EMERGENCY audit log |
| Sprint 3 | ANNUAL independence unique per year per employee (try duplicate → 409) |
| Sprint 4 | OT cap: 300h enforced precisely (not 299.9, not 300.1) |
| Sprint 4 | Leave balance: atomic update on approve (used ↑, pending ↓) verified |
| Sprint 5 | All 10 reports return data from real DB (not hardcoded) |
| Sprint 5 | Excel export opens in Excel without errors |
| Sprint 5 | p95 API latency < 500ms measured on common queries |
| Sprint 5 | Security review completed and signed off |
| Sprint 5 | E2E Playwright tests: all 7 user journeys pass on staging |

---

## Quick Reference: DoD Summary

```
Before any commit:
  ✓ make migrate-lint (if migration changed)
  ✓ make lint (Go + TS)
  ✓ go test ./... (no failures)

Before Sprint N+1:
  ✓ Sprint N exit criteria 100% complete
  ✓ make test (full suite)
  ✓ Smoke test existing modules
  ✓ CHANGELOG entry added

Before production deploy (after Sprint 5):
  ✓ All sprints done
  ✓ Security review signed off
  ✓ E2E tests passing on staging
  ✓ p95 < 500ms verified
```
