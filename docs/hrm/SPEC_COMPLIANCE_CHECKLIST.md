# HRM SPEC Compliance Checklist
## ERP System — MDH Audit Firm
**Version:** 1.0 | **Based on:** HRM_SPEC_v1.4.md | **Run:** Before declaring each sprint complete

---

> Run this checklist at sprint close. If any item fails, STOP — implementation has drifted from SPEC. Follow the drift resolution protocol (§Drift Resolution at bottom).

---

## §3 Organization Structure

### §3.1 Branches
- [ ] Exactly 2 branches exist in DB: HO (is_head_office=true) and HCM (is_head_office=false)
- [ ] HO branch: city = 'Hà Nội', is_active = true
- [ ] HCM branch: city = 'TP. Hồ Chí Minh', is_active = true
- [ ] Unique index enforces only 1 branch with is_head_office = true

```sql
-- Verify 2 branches with correct values
SELECT code, name, is_head_office, city, is_active FROM branches ORDER BY code;
-- Expected: HCM (false, TP. Hồ Chí Minh, true), HO (true, Hà Nội, true)

-- Verify uniqueness constraint
SELECT COUNT(*) FROM branches WHERE is_head_office = true;
-- Expected: 1
```

### §3.2 Departments
- [ ] Exactly 5 departments: AUDIT, TAX, HR, FIN, IT
- [ ] Correct dept_type: AUDIT=CORE, TAX=CORE, HR=SUPPORT, FIN=SUPPORT, IT=SUPPORT
- [ ] All departments is_active = true

```sql
SELECT code, name, dept_type, is_active FROM departments ORDER BY code;
-- Expected: 5 rows with correct values
```

### §3.3 Branch-Department Matrix
- [ ] HO branch has 5 departments: AUDIT, TAX, HR, FIN, IT
- [ ] HCM branch has 2 departments: AUDIT, TAX only

```sql
SELECT b.code, d.code
FROM branch_departments bd
JOIN branches b ON b.id = bd.branch_id
JOIN departments d ON d.id = bd.department_id
ORDER BY b.code, d.code;
-- Expected: HCM-AUDIT, HCM-TAX, HO-AUDIT, HO-FIN, HO-HR, HO-IT, HO-TAX (7 rows)

SELECT COUNT(*) FROM branch_departments WHERE branch_id = (SELECT id FROM branches WHERE code='HO');
-- Expected: 5

SELECT COUNT(*) FROM branch_departments WHERE branch_id = (SELECT id FROM branches WHERE code='HCM');
-- Expected: 2
```

### §3.5 Roles (11 roles)
- [ ] All 11 role codes exist and are used in middleware.RequireRole() calls
- [ ] No additional roles invented beyond the 11 defined in SPEC §3.5

```sql
-- If roles stored in DB table:
SELECT name FROM roles WHERE name IN (
  'SUPER_ADMIN','CHAIRMAN','CEO','HR_MANAGER','HR_STAFF',
  'HEAD_OF_BRANCH','PARTNER','AUDIT_MANAGER','SENIOR_AUDITOR','JUNIOR_AUDITOR','ACCOUNTANT'
);
-- Expected: 11 rows
```

```bash
# Check no invented roles in Go code
grep -rn "RequireRole" apps/api/internal/hrm/ | grep -v "SUPER_ADMIN\|CHAIRMAN\|CEO\|HR_MANAGER\|HR_STAFF\|HEAD_OF_BRANCH\|PARTNER\|AUDIT_MANAGER\|SENIOR_AUDITOR\|JUNIOR_AUDITOR\|ACCOUNTANT"
# Expected: no output (no invented roles)
```

---

## §4 Employee Entity

### §4.1–§4.7 Column Completeness
- [ ] employees table has ALL columns from §4.2 (Basic): employee_code, grade, position_title, manager_id, status, employment_type, hired_date, probation_end_date, termination_date, termination_reason, current_contract_id
- [ ] PII columns from §4.3: full_name, gender, date_of_birth, place_of_birth, nationality, personal_email, personal_phone, cccd_encrypted, cccd_issued_date, cccd_issued_place, passport_number
- [ ] Employment columns from §4.4: hired_source, probation_salary_pct, work_location, remote_days_per_week, referrer_employee_id
- [ ] Qualification columns from §4.5: education_level, vn_cpa_number, vn_cpa_issued_date, vn_cpa_expiry_date, practicing_certificate_number
- [ ] Salary/Bank columns from §4.6: base_salary, bank_account_encrypted, bank_name, bank_branch, mst_ca_nhan_encrypted
- [ ] BHXH columns from §5.3: so_bhxh_encrypted, bhxh_registered_date, bhyt_card_number, bhyt_expiry_date

```sql
-- Verify encrypted column names (suffix must be _encrypted, not _enc)
SELECT column_name FROM information_schema.columns
WHERE table_name = 'employees'
AND column_name LIKE '%encrypt%';
-- Expected: cccd_encrypted, bank_account_encrypted, mst_ca_nhan_encrypted, so_bhxh_encrypted
-- NOT: cccd_enc, bank_enc, mst_enc, bhxh_enc
```

### §4.8 Employee Code Trigger
- [ ] Trigger `trg_employees_set_code` exists on employees table
- [ ] Function `fn_employees_set_code()` exists
- [ ] Format generates NV{YY}-{SEQ4}: e.g., `NV26-0001` for first 2026 employee

```sql
-- Verify trigger exists
SELECT trigger_name FROM information_schema.triggers
WHERE event_object_table = 'employees' AND trigger_name = 'trg_employees_set_code';
-- Expected: 1 row

-- Verify function exists
SELECT routine_name FROM information_schema.routines
WHERE routine_name = 'fn_employees_set_code';
-- Expected: 1 row

-- Test format (integration test, not prod):
-- INSERT employee without employee_code → check auto-generated code
```

### §4.9 Employment Contracts
- [ ] employment_contracts table exists with all columns
- [ ] contract_type CHECK: PROBATION, DEFINITE_TERM, INDEFINITE, INTERN
- [ ] is_current column present (BOOLEAN)
- [ ] chk_contract_dates constraint: end_date > start_date

```sql
SELECT column_name FROM information_schema.columns
WHERE table_name = 'employment_contracts' ORDER BY ordinal_position;
-- Must include: id, employee_id, contract_number, contract_type, start_date, end_date,
--              signed_date, salary_at_signing, position_at_signing, notes, document_url,
--              is_current, created_by, created_at
```

---

## §5 BHXH & Tax TNCN

### §5.2 Insurance Rate Config
- [ ] insurance_rate_config table exists
- [ ] 2024 seed data present: bhxh_employee_pct=8.00, bhxh_employer_pct=17.50, bhyt_employee_pct=1.50, bhyt_employer_pct=3.00, bhtn_employee_pct=1.00, bhtn_employer_pct=1.00, kpcd_employer_pct=2.00
- [ ] salary_base_bhxh=1800000, max_bhxh_salary=36000000 for 2024

```sql
SELECT effective_from, bhxh_employee_pct, bhxh_employer_pct, kpcd_employer_pct,
       salary_base_bhxh, max_bhxh_salary
FROM insurance_rate_config WHERE effective_from = '2024-01-01';
-- Expected: 1 row with values matching SPEC §5.2
```

### §5.4 Employee Dependents
- [ ] employee_dependents table exists
- [ ] relationship CHECK: SPOUSE, CHILD, PARENT, SIBLING, OTHER
- [ ] tax_deduction_registered, tax_deduction_from, tax_deduction_to columns present

### §5.5 Salary History (Immutable)
- [ ] employee_salary_history table exists
- [ ] change_type CHECK: INITIAL, INCREASE, DECREASE, PROMOTION, ADJUSTMENT
- [ ] Immutability rules present: no_update_salary_history, no_delete_salary_history

```sql
SELECT rulename FROM pg_rules WHERE tablename = 'employee_salary_history';
-- Expected: no_update_salary_history, no_delete_salary_history
```

---

## §6 Professional Development

### §6.1 Certifications
- [ ] certifications table exists
- [ ] cert_type CHECK: VN_CPA, ACCA, CFA, CIA, CISA, CPA_AUS, OTHER
- [ ] status CHECK: ACTIVE, EXPIRED, SUSPENDED, SURRENDERED
- [ ] renewal_reminder_days default 60

### §6.2–§6.3 Training
- [ ] training_courses table with course_type CHECK: INTERNAL, EXTERNAL, ONLINE, CONFERENCE
- [ ] training_records table with status CHECK: REGISTERED, IN_PROGRESS, COMPLETED, CANCELLED, NO_SHOW

### §6.4 CPE Requirements Seed
- [ ] cpe_requirements_by_role table exists
- [ ] Seed records present for PARTNER, SENIOR_AUDITOR, JUNIOR_AUDITOR

```sql
SELECT role_code, cert_type, required_hours_per_year, regulatory_body
FROM cpe_requirements_by_role WHERE effective_from = '2024-01-01'
ORDER BY role_code, cert_type;
-- Expected: 7 rows matching SPEC §6.4 seed data
-- PARTNER/VN_CPA/40/VACPA, SENIOR_AUDITOR/VN_CPA/40/VACPA, JUNIOR_AUDITOR/VN_CPA/40/VACPA
-- PARTNER/ACCA/40/ACCA Global, SENIOR_AUDITOR/ACCA/40/ACCA Global
-- PARTNER/CIA/40/IIA, PARTNER/CFA/20/CFA Institute
```

### §6.5–§6.7 Performance + Independence
- [ ] performance_reviews table exists with uidx_review_employee_period unique constraint
- [ ] review_type CHECK: SELF, MANAGER, PEER, COMMITTEE
- [ ] independence_declarations table with uidx_independence_annual partial unique index
- [ ] declaration_type CHECK: ANNUAL, PER_ENGAGEMENT
- [ ] chk_annual_year and chk_engagement_ref constraints exist

```sql
SELECT indexname FROM pg_indexes WHERE tablename = 'independence_declarations';
-- Expected: includes uidx_independence_annual (partial WHERE declaration_type = 'ANNUAL')
```

---

## §7 Time & Leave

### §7.1 Holidays
- [ ] holidays table exists with generated year column
- [ ] 2026 national holidays seeded (minimum 12 from SPEC §7.1)
- [ ] type CHECK: NATIONAL, COMPANY, REGIONAL

```sql
SELECT COUNT(*) FROM holidays WHERE year = 2026 AND type = 'NATIONAL';
-- Expected: >= 12

SELECT holiday_date, name FROM holidays WHERE year = 2026 ORDER BY holiday_date;
-- Expected: Tết Dương lịch 2026-01-01 present, Tết Nguyên Đán dates present, etc.
```

### §7.2 Leave Types
- [ ] leave_balances.leave_type CHECK includes all 8: ANNUAL, SICK, MATERNITY, PATERNITY, PERSONAL, MARRIAGE, FUNERAL, UNPAID
- [ ] leave_requests.leave_type same 8 values

```sql
-- Verify CHECK constraint on leave_type
SELECT conname, consrc FROM pg_constraint
WHERE conrelid = 'leave_requests'::regclass AND contype = 'c' AND conname LIKE '%leave_type%';
-- Expected: constraint includes all 8 leave types
```

### §7.3–§7.5 Leave Balances, Requests, OT
- [ ] leave_balances: uidx_leave_balance UNIQUE (employee_id, leave_type, year)
- [ ] ot_requests: chk_ot_times (end_time > start_time), chk_ot_hours (0 < ot_hours <= 12)
- [ ] employee_ot_summary_year VIEW exists with 300h cap calculation

```sql
SELECT viewname FROM pg_views WHERE viewname = 'employee_ot_summary_year';
-- Expected: 1 row

-- Check view calculates correctly: 300.0 - approved_hours
\d employee_ot_summary_year
```

### §7.7 timesheets ALTER
- [ ] timesheets table has ot_hours, ot_approved, ot_request_id columns

```sql
SELECT column_name FROM information_schema.columns
WHERE table_name = 'timesheets' AND column_name IN ('ot_hours','ot_approved','ot_request_id');
-- Expected: 3 rows
```

---

## §8 User Provisioning

### §8.2 user_provisioning_requests Table
- [ ] Table exists with all columns including expires_at (now() + 30 days)
- [ ] status CHECK: PENDING, APPROVED, REJECTED, EXECUTED, CANCELLED
- [ ] uidx_provisioning_pending WHERE status = 'PENDING' exists

```sql
SELECT indexname FROM pg_indexes WHERE tablename = 'user_provisioning_requests';
-- Expected: includes uidx_provisioning_pending
```

### §8.3 Approval Flow
- [ ] HCM flow: branch_approve → hr_approve → execute (3 steps)
- [ ] HO flow: direct execute (no intermediate approvals required)
- [ ] Emergency flow: is_emergency = true skips approvals

---

## §9 Lifecycle Events

### §9.4 Offboarding Checklists
- [ ] offboarding_checklists table exists
- [ ] checklist_type CHECK: ONBOARDING, OFFBOARDING
- [ ] items column is JSONB with default empty structure

---

## §10 Expense Claims

### §10.1–§10.2 Expense Tables
- [ ] expense_claims: claim_number auto-generated (PC{YY}-{SEQ4}) via trigger
- [ ] status CHECK: DRAFT, SUBMITTED, MANAGER_APPROVED, HR_APPROVED, PAID, REJECTED
- [ ] expense_claim_items: category CHECK: FLIGHT, HOTEL, TAXI, MEAL, STATIONERY, COMMUNICATION, OTHER
- [ ] expense_claim_items: amount CHECK > 0

```sql
-- Verify trigger
SELECT trigger_name FROM information_schema.triggers
WHERE event_object_table = 'expense_claims';
-- Expected: trg_expense_claims_set_number
```

---

## §13 API Endpoints Catalog

### Path Compliance Checks
- [ ] All organization endpoints match SPEC §13.1 exactly (10 endpoints)
- [ ] All employee endpoints match SPEC §13.2 (8 endpoints + my-profile)
- [ ] Provisioning endpoints match SPEC §13.12 (9 endpoints)
- [ ] Leave endpoints match SPEC §13.8 (10 endpoints)
- [ ] OT endpoints match SPEC §13.9 (7 endpoints)
- [ ] Independence endpoints match SPEC §13.7 (6 endpoints)

```bash
# Grep all registered routes and compare to SPEC §13
grep -rn "r\.\(GET\|POST\|PUT\|DELETE\)" apps/api/internal/hrm/ | sort
# Manually compare against SPEC §13 endpoint list
```

### Response Format Compliance
- [ ] List responses: `{"data": [...], "meta": {"page": N, "size": N, "total": N}}`
- [ ] Single item responses: `{"data": {...}}`
- [ ] Error responses: `{"error": "ERROR_CODE", "message": "..."}`

---

## §14 UI Pages

### Route Compliance
- [ ] All admin pages match SPEC §14.1 paths
- [ ] All HRM module pages match SPEC §14.2 paths
- [ ] All self-service pages match SPEC §14.3 paths
- [ ] All team management pages match SPEC §14.4 paths

```bash
# Check Next.js app router files match SPEC §14
find apps/web/src/app -name "page.tsx" | sort
# Compare against SPEC §14 paths
```

---

## §15 Permission Matrix

### Branch Scope Verification

```bash
# Verify HoB query filtering in Go code
grep -rn "branch_id" apps/api/internal/hrm/usecase/ | grep -i "caller\|scope\|filter"
# Expected: branch scope applied in usecase layer for HEAD_OF_BRANCH and HR_STAFF
```

### Sensitive Field Masking

```bash
# Verify masking logic exists
grep -rn "cccd_encrypted.*\*\*\*\|bank_account.*\*\*\*" apps/api/internal/hrm/
# Expected: masking applied for unauthorized roles
```

---

## §16 Notifications

- [ ] All 25 notification events from SPEC §16.1 are wired
- [ ] Events 1 (contract expiry), 2 (cert expiry), 19 (CPE), 25 (BHYT) are scheduled jobs
- [ ] Email templates match SPEC §16.3 (8 templates)

```bash
# Check scheduled jobs registered
grep -rn "contract.*expir\|cert.*expir\|cpe.*remind\|bhyt.*expir" apps/api/
# Expected: scheduler entries for these 4 events
```

---

## §17 Audit Log Events

### Event Coverage

- [ ] EMPLOYEE_PII_ACCESSED written on every GET /sensitive call
- [ ] EMPLOYEE_SALARY_CHANGED written on POST salary-history
- [ ] PROVISIONING_EXECUTED written on execute action (with user_id and role in metadata)
- [ ] PROVISIONING_EMERGENCY written when is_emergency = true
- [ ] SALARY_REPORT_GENERATED written every time /hrm/reports/salary is accessed
- [ ] LEAVE_BALANCE_ADJUSTED written on manual balance change
- [ ] OT_CAP_WARNING written when approved total > 270h

```bash
# Verify audit log calls exist in code
grep -rn "auditLog\|audit\.Log\|AuditLog" apps/api/internal/hrm/ | grep -c "EMPLOYEE_PII_ACCESSED\|SALARY_REPORT\|PROVISIONING_EXECUTED"
# Expected: >= 3 matches
```

---

## §18 Security & Privacy

### Encryption
- [ ] Only AES-256-GCM used for encryption (no MD5, SHA, bcrypt for PII encryption)
- [ ] `HRM_ENCRYPTION_KEY` read from environment, never hardcoded
- [ ] Exactly 4 encrypted columns in employees table (see §2.3 TECHNICAL_RULES.md)

```bash
# Verify no hardcoded keys
grep -rn "HRM_ENCRYPTION_KEY\s*=" apps/api/internal/hrm/
# Expected: only os.Getenv() calls, no string literals
```

### Application Layer Decryption
```bash
# Verify decrypt not in SQL
grep -rn "decrypt\|pgp_sym_decrypt\|decrypt_aes" apps/api/migrations/
# Expected: no output (decryption only in Go application layer)
```

---

## §19 Reporting

- [ ] All 10 report endpoints exist at paths matching SPEC §19.1
- [ ] Salary report (Report 10): accessible only to HR_MANAGER, CEO, CHAIRMAN
- [ ] All reports support ?branch_id and ?dept_id filters where specified
- [ ] Export: Excel (.xlsx) and CSV available for all reports

```bash
# Verify all 10 report paths registered
grep -rn "/hrm/reports/" apps/api/ | grep -o "GET.*reports/[a-z-]*" | sort | uniq
# Expected: cpe-compliance, leave-usage, ot-summary, headcount, contract-renewal,
#           cert-expiry, independence-status, expense-summary, performance-distribution, salary
```

---

## §20 Testing Strategy

- [ ] All unit tests from SPEC §20.1.1 implemented:
  - TestCalculateLeaveDays, TestCalculateOTHours, TestOTCapCheck, TestLeaveBalanceCheck
  - TestEmployeeCodeFormat, TestInsuranceContribution, TestTNCNDeduction, TestCPEProgress
- [ ] All integration tests from SPEC §20.2 implemented:
  - TestEmployeeCreation, TestEmployeeSalaryChange (with immutability test)
  - TestLeaveApprovalFlow, TestLeaveRejectionFlow
  - TestOTCapEnforcement (3 scenarios: 15h fail, 10h pass, 1h fail after 300)
  - TestProvisioningHCMFlow, TestIndependenceCleanDeclaration, TestIndependenceConflictFlow

```bash
go test ./internal/hrm/... -v -run "TestCalculateOTHours|TestOTCapEnforcement|TestLeaveApprovalFlow|TestProvisioningHCMFlow|TestIndependenceConflictFlow"
# Expected: all PASS
```

---

## §21 Roadmap — Sprint Scope Adherence

For each sprint, verify no Sprint N+1 code was committed in Sprint N:

```bash
# Check git log for sprint boundary violations (example for Sprint 1)
git log --oneline sprint1-start..sprint1-end -- apps/api/migrations/000021* apps/api/migrations/000022* apps/api/migrations/000023* apps/api/migrations/000024* apps/api/migrations/000025*
# Expected: no output (Sprint 2+ migrations not committed in Sprint 1)
```

---

## §22 Bootstrap

- [ ] All items in SPEC §22.7 bootstrap checklist satisfied
- [ ] Migrations 000019–000026 applied successfully on target environment
- [ ] SUPER_ADMIN user created with TOTP setup
- [ ] 6 initial users (CHAIRMAN, CEO, HR_MANAGER, HoB HCM, Partner HO, Partner HCM) created
- [ ] All employees have at least 1 employment contract
- [ ] Annual independence declaration submitted for all active employees
- [ ] Leave balances initialized for all active employees

---

## Drift Detection Rules

**Red flags that indicate implementation has drifted from SPEC:**

1. **New role code found that is not in the 11 defined roles** → Bug: invented role
2. **Field named without `_encrypted` suffix for encrypted PII** → Bug: naming violation
3. **API path not matching SPEC §13** → Bug: invented endpoint
4. **Enum value in DB as lowercase or Vietnamese** → Bug: convention violation
5. **Decryption happening in SQL query** → Bug: security violation
6. **SUPER_ADMIN check in business logic** → Bug: middleware handles this
7. **Migration 000019–000026 out of sequence or with gaps** → Bug: numbering violation
8. **Seed data in migrations 000019–000025** → Bug: only 000026 can have seed data
9. **SELECT * in any production query** → Bug: performance/security violation
10. **TODO/FIXME in production code path** → Bug: incomplete implementation

---

## Drift Resolution Protocol

When drift is found, follow this protocol exactly:

1. **STOP implementation** — do not add more code on top of drifted code
2. **Identify the gap:**
   - Is the implementation wrong? → Fix code to match SPEC
   - Is the SPEC ambiguous or wrong? → Escalate to user for clarification
3. **Decide with user:**
   - Fix code to match SPEC (preferred)
   - Update SPEC to reflect new decision (requires explicit user approval)
4. **If SPEC is updated:** Document the change in SPEC changelog (top of HRM_SPEC_v1.4.md)
5. **Resume implementation** after drift is resolved and user approves
6. **Re-run this checklist** for the affected section before marking done
