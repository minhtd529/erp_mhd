# HRM Technical Rules
## ERP System — MDH Audit Firm
**Version:** 1.0 | **Based on:** HRM_SPEC_v1.4.md | **Enforcement:** Mandatory — no exceptions without user approval

---

> These rules exist to prevent drift, reinvention, and quality failures during the 4–6 week HRM implementation. Claude Code MUST follow every rule in this document. If a rule conflicts with an implementation requirement, STOP and escalate to the user before proceeding.

---

## 1. Migration Rules

Reference: `scripts/migration-lint.sh`, SPEC §12, CLAUDE.md §Migration Rules

### 1.1 Immutability
- **NEVER** modify a migration file after it has been committed to `main`
- **NEVER** rename an existing migration file
- **NEVER** delete a migration file
- If a committed migration has an error → create a NEW migration with a higher version number that corrects it
- Document the incident in CHANGELOG.md

### 1.2 File Format
All migration files must follow: `000NNN_snake_case_name.{up,down}.sql`
- Correct: `000019_hrm_organization.up.sql`
- Wrong: `19_hrm_org.sql`, `000019_HRMOrganization.up.sql`

### 1.3 Up/Down Pair
Every migration MUST have both `.up.sql` and `.down.sql`. No exceptions.
- `up.sql`: Creates or alters schema
- `down.sql`: Fully reverses `up.sql` with no data loss on dev/staging

### 1.4 Sequential Numbering
- HRM migrations: 000019 → 000026 (no gaps)
- Check existing sequence before writing: `ls apps/api/migrations/*.up.sql | sort`
- Use scaffold: `make migrate-create NAME=hrm_expenses`

### 1.5 Test Before Commit
```bash
make migrate-up && make migrate-down && make migrate-up
# All three must succeed without errors
make migrate-lint
# Must pass before any commit containing a migration
```

### 1.6 Seed Data in Separate Migration
- Business seed data (branches, holidays, CPE requirements) → migration 000026 only
- Schema migrations (000019–000025) MUST NOT contain INSERT statements (except for triggers/functions)
- Migration 000026 is the ONLY seed migration — See SPEC §12.9

### 1.7 Rollback Risk Classification
Document risk in migration comments:

```sql
-- Rollback risk: LOW — no FK references, safe to rollback on dev
-- Rollback risk: HIGH — contains employee data, never rollback on production
```

### 1.8 Check for Existing Tables
Before `CREATE TABLE`, verify table does not already exist:
```bash
grep -ril "CREATE TABLE.*table_name" apps/api/migrations/*.up.sql
```
If table exists → use `ALTER TABLE` in new migration, not `CREATE TABLE`.

### 1.9 Required Constraints
Every new table MUST have:
- `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
- Indexes for all foreign key columns
- CHECK constraints for all enum/status columns

---

## 2. Naming Conventions

Reference: SPEC §11 (all CREATE TABLE statements), CLAUDE.md §Database Conventions

### 2.1 Tables
- snake_case, use existing convention from migrations 000001–000018
- Plural form generally: `employees`, `certifications`, `leave_requests`
- Junction tables: `{table1}_{table2}`: `branch_departments`

### 2.2 Columns
- snake_case always: `employee_code`, `hired_date`, `is_head_office`
- Never camelCase or PascalCase

### 2.3 Encrypted Columns (CRITICAL)
Use `_encrypted` suffix exactly — no abbreviations:
- ✅ `cccd_encrypted`, `bank_account_encrypted`, `mst_ca_nhan_encrypted`, `so_bhxh_encrypted`
- ❌ `cccd_enc`, `bank_acct_enc`, `cccd_hash`, `cccd`

The 4 encrypted fields in HRM (See SPEC §18.1):
1. `employees.cccd_encrypted` — CCCD/CMND number
2. `employees.mst_ca_nhan_encrypted` — Personal tax number
3. `employees.so_bhxh_encrypted` — BHXH book number
4. `employees.bank_account_encrypted` — Bank account number

### 2.4 Foreign Keys
Pattern: `{singular_table_name}_id`
- `employee_id` references employees(id)
- `branch_id` references branches(id)
- `reviewer_id` references employees(id)
- `approved_by` references users(id) — exception: actor columns use verb form

### 2.5 Timestamps
- Created: `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
- Updated: `updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`
- ❌ Never: `created_time`, `update_time`, `createdAt`, `updatedAt`

### 2.6 Boolean Columns
- Prefix with `is_`: `is_active`, `is_deleted`, `is_head_office`, `is_emergency`, `is_current`, `is_billable`
- ❌ Never: `active`, `deleted`, `head_office`, `emergency`

### 2.7 Enum Values in VARCHAR Columns
Always UPPERCASE English in CHECK constraints:
```sql
CHECK (status IN ('PENDING','APPROVED','REJECTED','EXECUTED','CANCELLED'))
CHECK (grade IN ('EXECUTIVE','PARTNER','DIRECTOR','MANAGER','SENIOR','JUNIOR','INTERN','SUPPORT'))
```
- ❌ Never: lowercase `'pending'`, mixed case `'Pending'`, Vietnamese `'CHO_DUYET'`

### 2.8 Indexes
- Regular: `idx_{table}_{column(s)}`  — `idx_employees_branch`, `idx_leave_requests_status`
- Unique: `uidx_{table}_{column(s)}` — `uidx_leave_balance`, `uidx_branches_head_office`
- Partial: include WHERE clause as comment — `uidx_provisioning_pending WHERE status = 'PENDING'`

### 2.9 PostgreSQL Rules (Immutability)
Name rules as: `no_{action}_{table}` — `no_update_salary_history`, `no_delete_salary_history`

---

## 3. Go Code Patterns

Reference: `apps/api/internal/` existing structure, CLAUDE.md §Architecture Patterns, §Go Code Naming

### 3.1 Handler → UseCase → Repository Layering
**Strict separation — NEVER bypass layers:**
```
Handler (HTTP input/output) → UseCase (business logic) → Repository (DB queries)
```
- Business logic in UseCase, NOT in handlers
- DB queries in Repository, NOT in UseCases
- HTTP status codes only in handlers
- ❌ Never: business logic in handlers, SQL in use cases

### 3.2 sqlc for Queries
- All repository queries defined in `.sql` files, generated via sqlc
- ❌ Never: `SELECT *` in any query
- Always select explicit columns
- Raw `pgx` allowed for complex dynamic queries that sqlc cannot express

### 3.3 Gin for Routing
Match route registration to SPEC §13 exactly:
```go
hrm := r.Group("/api/v1/hrm")
hrm.GET("/employees", middleware.RequireRole("HR_MANAGER","CEO","CHAIRMAN","SUPER_ADMIN","HEAD_OF_BRANCH"), h.ListEmployees)
hrm.POST("/employees", middleware.RequireRole("HR_MANAGER","CEO","SUPER_ADMIN"), h.CreateEmployee)
```

### 3.4 Error Wrapping
Always wrap with context:
```go
return nil, fmt.Errorf("EmployeeUseCase.GetByID: %w", err)
return nil, fmt.Errorf("EmployeeRepository.FindByBranch branch=%s: %w", branchID, err)
```
❌ Never: `return nil, err` without wrapping

### 3.5 Error Logging Before 500
```go
if err != nil {
    log.Printf("EmployeeHandler.GetSensitive: %v", err)  // always log before returning 500
    return errResp(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
}
```
Lesson learned from dashboard debugging: silent 500s are impossible to diagnose.

### 3.6 errResp Helper
```go
// Use existing errResp pattern (do not invent new error response helpers)
return errResp(c, http.StatusForbidden, "INSUFFICIENT_PERMISSION", "Access denied")
return errResp(c, http.StatusNotFound, "EMPLOYEE_NOT_FOUND", "Employee not found")
```
Error codes from SPEC §13: `EMPLOYEE_NOT_FOUND`, `DUPLICATE_EMPLOYEE_CODE`, `OT_ANNUAL_CAP_EXCEEDED`, `DUPLICATE_PENDING_REQUEST`, `REQUEST_EXPIRED`, `INVALID_ROLE_FOR_PROVISIONING`

### 3.7 Audit Log on All Mutations
Every CREATE, UPDATE, DELETE must write to audit_logs via `pkg/audit`:
```go
// In UseCase, after successful mutation:
uc.auditLog.Log(ctx, audit.Event{
    Action:     "EMPLOYEE_CREATED",
    EntityType: "employee",
    EntityID:   employee.ID,
    ActorID:    callerID,
    ActorRole:  callerRole,
    After:      employee, // omit PII values
    IPAddress:  ipAddress,
})
```
See SPEC §17.2 for complete event catalog (EMPLOYEE_CREATED, EMPLOYEE_UPDATED, LEAVE_REQUEST_APPROVED, etc.)

### 3.8 Struct Naming
| Type | Pattern | Example |
|---|---|---|
| Handler | `{Entity}Handler` | `EmployeeHandler`, `LeaveHandler` |
| UseCase | `{Entity}UseCase` | `EmployeeUseCase`, `ProvisioningUseCase` |
| Repository | `{Entity}Repository` | `EmployeeRepository`, `CertificationRepository` |
| Request DTO | `{Entity}{Op}Request` | `EmployeeCreateRequest`, `LeaveApproveRequest` |
| Response DTO | `{Entity}Response` | `EmployeeResponse`, `LeaveBalanceResponse` |

### 3.9 Branch Scope Enforcement
For HoB and HR_STAFF roles, filter all list queries by caller's branch_id:
```go
func (uc *EmployeeUseCase) List(ctx context.Context, caller Caller, filter ListFilter) ([]*Employee, error) {
    if caller.Role == "HEAD_OF_BRANCH" || caller.Role == "HR_STAFF" {
        filter.BranchID = &caller.BranchID  // force branch scope
    }
    return uc.repo.List(ctx, filter)
}
```
See SPEC §15.3 for scope rules per role.

### 3.10 Sensitive Field Masking
```go
// In response serialization — never return raw values to unauthorized roles
if !caller.CanViewSensitive() {
    emp.CCCDEncrypted = "***"
    emp.BankAccountEncrypted = "***"
    emp.MstCaNhanEncrypted = "***"
    emp.SoBHXHEncrypted = "***"
    emp.BaseSalary = nil
}
```
See SPEC §15.4 for exact masking rules.

### 3.11 SUPER_ADMIN Bypass
`SUPER_ADMIN` bypasses all `RequireRole` middleware automatically. Already implemented in middleware.go — do NOT add special SUPER_ADMIN cases in business logic. The middleware handles it.

---

## 4. TypeScript/React Patterns

Reference: `apps/web/src/` existing structure, CLAUDE.md §Frontend Conventions

### 4.1 API Service Layer
All API calls through `services/*.ts` files:
```typescript
// services/hrm/employeeService.ts
export const employeeService = {
  list: (params: EmployeeListParams) =>
    axios.get<PaginatedResponse<Employee>>('/api/v1/hrm/employees', { params }),
  getById: (id: string) =>
    axios.get<SingleResponse<Employee>>(`/api/v1/hrm/employees/${id}`),
  create: (data: EmployeeCreateRequest) =>
    axios.post<SingleResponse<Employee>>('/api/v1/hrm/employees', data),
}
```
❌ Never: raw `fetch()` or `axios` calls inside components

### 4.2 Data Fetching with react-query
```typescript
const { data, isLoading, error } = useQuery({
  queryKey: ['hrm', 'employees', filters],
  queryFn: () => employeeService.list(filters),
})
```
Mutations:
```typescript
const mutation = useMutation({
  mutationFn: employeeService.create,
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['hrm', 'employees'] })
    toast.success('Nhân viên đã được tạo thành công')
  },
})
```

### 4.3 Global State with zustand
Global state only for: auth user/role, branch filter, UI preferences.
Do NOT use zustand for server data — that belongs in react-query.

### 4.4 UI Components (shadcn/ui only)
```typescript
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { DataTable } from '@/components/ui/data-table'
```
❌ Never: custom button/input components that duplicate shadcn ones

### 4.5 Forms (react-hook-form + zod)
```typescript
const schema = z.object({
  full_name: z.string().min(1, 'Họ tên không được để trống'),
  grade: z.enum(['EXECUTIVE','PARTNER','DIRECTOR','MANAGER','SENIOR','JUNIOR','INTERN','SUPPORT']),
  hired_date: z.string().refine(d => !isNaN(Date.parse(d)), 'Ngày không hợp lệ'),
})

const form = useForm<z.infer<typeof schema>>({
  resolver: zodResolver(schema),
})
```

### 4.6 Styling (Tailwind only)
- ❌ Never: inline styles (`style={{ color: 'red' }}`)
- ❌ Never: CSS modules for new components
- Use design system colors: `bg-[#1F3A70]`, `text-[#D4A574]` — See CLAUDE.md §Color Palette
- Loading: skeleton or spinner from shadcn
- Toast: success `toast.success()`, error `toast.error()`

### 4.7 Required UI States
Every list page MUST have:
- **Loading state:** `<Skeleton />` while fetching
- **Empty state:** Vietnamese message + icon when no data
- **Error state:** Vietnamese error message + retry button

Every form page MUST have:
- **Validation errors:** zod schema errors displayed per field
- **Submit loading:** button disabled + spinner while submitting
- **Success toast:** Vietnamese confirmation after mutation

### 4.8 Mobile Responsiveness
All pages must work at ≥ 375px width. Use responsive Tailwind classes:
```typescript
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
```

---

## 5. API Conventions

Reference: SPEC §13, CLAUDE.md §API Conventions

### 5.1 Base Path
All HRM endpoints: `/api/v1/hrm/...` and `/api/v1/my-*`

### 5.2 Paths Match SPEC §13 Exactly
Do NOT invent paths not in the spec. Examples:
- ✅ `/api/v1/hrm/employees` (SPEC §13.2)
- ❌ `/api/v1/hrm/staff` (invented, not in SPEC)
- ✅ `/api/v1/my-leave/requests` (SPEC §13.8)
- ❌ `/api/v1/leave/my-requests` (wrong structure)

### 5.3 Response Format
Collections:
```json
{ "data": [...], "meta": { "page": 1, "size": 20, "total": 100 } }
```
Singles:
```json
{ "data": { "id": "...", ... } }
```
Errors:
```json
{ "error": "ERROR_CODE", "message": "Human-readable message in English" }
```

### 5.4 HTTP Status Codes
| Code | When |
|---|---|
| 200 | Successful GET, PUT |
| 201 | Successful POST (created) |
| 204 | Successful DELETE |
| 400 | Bad request / validation failure |
| 401 | Not authenticated |
| 403 | Authenticated but not authorized |
| 404 | Resource not found |
| 409 | Conflict (e.g., DUPLICATE_PENDING_REQUEST) |
| 422 | Business rule violation (e.g., OT_ANNUAL_CAP_EXCEEDED) |
| 500 | Internal server error (always log before returning) |

### 5.5 Pagination
Default: `?page=1&size=20` (max 100)
Always include meta in list responses.

---

## 6. Permission Enforcement

Reference: SPEC §15, CLAUDE.md §RBAC

### 6.1 Middleware First
All role checks via Gin middleware — never inline role checks in handlers:
```go
// Correct:
hrm.GET("/employees", middleware.RequireRole("HR_MANAGER","CEO"), h.ListEmployees)

// Wrong:
func (h *EmployeeHandler) ListEmployees(c *gin.Context) {
    if caller.Role != "HR_MANAGER" { ... }  // DON'T do this in handlers
}
```

### 6.2 SUPER_ADMIN Bypass
SUPER_ADMIN already bypasses all RequireRole middleware (implemented in middleware.go). Do NOT add `|| role == "SUPER_ADMIN"` to RequireRole calls — it's redundant and pollutes the code.

### 6.3 Branch Scope (ABAC)
Branch scope filtering in UseCase layer (not middleware):
- `HEAD_OF_BRANCH`: sees only employees/requests in own branch_id
- `HR_STAFF`: sees only employees in assigned branch
- After middleware passes role check, UseCase applies branch filter

### 6.4 Sensitive Field Access
In application layer (not middleware), check before decrypting:
```go
func (uc *EmployeeUseCase) GetSensitive(ctx context.Context, caller Caller, empID string) (*EmployeeSensitive, error) {
    if !caller.CanViewPII() {  // only HR_MANAGER, CEO, CHAIRMAN
        return nil, ErrInsufficientPermission
    }
    // decrypt and return
    // ALWAYS log to audit_logs regardless
}
```

### 6.5 Role Definitions (11 roles — exact codes from SPEC §3.5)
```
SUPER_ADMIN, CHAIRMAN, CEO, HR_MANAGER, HR_STAFF,
HEAD_OF_BRANCH, PARTNER, AUDIT_MANAGER, SENIOR_AUDITOR, JUNIOR_AUDITOR, ACCOUNTANT
```
❌ Never use any role code not in this list (no HR_ADMIN, no MANAGER, no STAFF)

---

## 7. Encryption Rules

Reference: SPEC §18.1

### 7.1 Algorithm
AES-256-GCM (Authenticated Encryption). No other algorithm for HRM PII.

### 7.2 Key Source
ONLY from environment variable `HRM_ENCRYPTION_KEY` (32 bytes, base64-encoded).
```go
key := os.Getenv("HRM_ENCRYPTION_KEY")
if key == "" {
    panic("HRM_ENCRYPTION_KEY environment variable not set")
}
```
❌ Never: hardcode key, derive from password, store in database

### 7.3 Encrypted Column Naming
Suffix `_encrypted` on all 4 encrypted columns (See §2.3 above).

### 7.4 Decrypt at Service Layer Only
```go
// UseCase decrypts — Repository never decrypts
func (uc *EmployeeUseCase) GetSensitive(...) {
    raw, _ := uc.repo.GetEncryptedFields(ctx, empID)
    cccd, _ := uc.crypto.Decrypt(raw.CCCDEncrypted)  // decrypt here
    return &EmployeeSensitive{CCCD: cccd}, nil
}
```
❌ Never: SQL decrypt functions, decrypt in handler, decrypt in repository

### 7.5 Audit Every Decryption
```go
uc.auditLog.Log(ctx, audit.Event{
    Action:   "EMPLOYEE_PII_ACCESSED",
    Metadata: map[string]any{"fields_accessed": []string{"cccd", "mst_ca_nhan", "bank_account"}},
})
```
This is mandatory even for read-only access — See SPEC §17.2.1

### 7.6 Never Log PII Values
Application logs MUST NOT contain decrypted PII values:
```go
// Wrong:
log.Printf("Employee CCCD: %s", cccdPlaintext)

// Correct:
log.Printf("Employee PII accessed for employee_id=%s", empID)
```

---

## 8. Testing Requirements

Reference: SPEC §20

### 8.1 Unit Tests (Required)
Mandatory unit tests for all calculation functions — See SPEC §20.1:
- `TestCalculateOTHours` — OT hours from start/end time
- `TestOTCapCheck` — 300h annual cap logic
- `TestLeaveBalanceCheck` — sufficient balance check
- `TestEmployeeCodeFormat` — NV{YY}-{SEQ4} format
- `TestInsuranceContribution` — BHXH calculation from rate config
- `TestCPEProgress` — CPE hours progress calculation

### 8.2 Integration Tests (Required)
One integration test per workflow — See SPEC §20.2:
- TestEmployeeLifecycle (Sprint 1)
- TestProvisioningHCMFlow (Sprint 2)
- TestLeaveApprovalFlow (Sprint 4)
- TestOTCapEnforcement (Sprint 4)
- TestExpenseClaimFlow (Sprint 5)

### 8.3 Test Naming Convention
```
Test{Feature}_{Scenario}_{ExpectedBehavior}
TestOTCapCheck_ExceedsCap_Returns422
TestLeaveApproval_ManagerApproves_BalanceUpdated
TestEmployeeCode_NewEmployee_FormatsCorrectly
```

### 8.4 Coverage Target
Minimum 70% test coverage for all new HRM code:
```bash
go test ./internal/hrm/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total
```

### 8.5 Real Database (No Mocks)
Integration tests MUST use a real test database — not mocks.
Lesson: mocked tests passed while prod migration failed in Q4 last year.
```go
// Use TestMain to set up test DB
func TestMain(m *testing.M) {
    db = setupTestDB()  // real postgres, test schema
    os.Exit(m.Run())
}
```

---

## 9. Git Discipline

Reference: CLAUDE.md §Git rules

### 9.1 One Migration = One Commit
```bash
git add apps/api/migrations/000019_hrm_organization.{up,down}.sql
git commit -m "feat(hrm): add migration 000019 hrm_organization schema"
```

### 9.2 Commit Message Format
`{type}({scope}): {description}`

| Type | Use |
|---|---|
| `feat` | New feature |
| `fix` | Bug fix |
| `chore` | Tooling, config |
| `test` | Adding tests |
| `docs` | Documentation |

Examples:
- `feat(hrm): add employee CRUD endpoints`
- `feat(hrm): add provisioning HCM approval flow`
- `test(hrm): add OT cap enforcement integration test`
- `fix(hrm): correct leave balance rollback on rejection`

### 9.3 Never Commit
- `.env` files, credentials, API keys
- `HRM_ENCRYPTION_KEY` value
- Compiled binaries
- Test database dumps

### 9.4 PR Scope
One logical feature per PR. Examples:
- "Sprint 1: Migrations 000019 + 000020 + organization API + employee API"
- "Sprint 2: Provisioning workflow (HCM + HO + emergency)"
- NOT: "HRM everything" in one massive PR

---

## 10. Vietnamese / English Content Policy

Reference: SPEC overall language usage

### 10.1 UI Labels → Vietnamese
```typescript
<Button>Tạo Nhân viên</Button>
<Label>Họ và tên</Label>
<TableHeader>Chi nhánh</TableHeader>
```

### 10.2 Error Messages Shown to Users → English (per API convention)
```json
{ "error": "EMPLOYEE_NOT_FOUND", "message": "Employee not found" }
```
Internal error messages in application logs may be Vietnamese or English — be consistent per file.

### 10.3 DB Identifiers → English snake_case
Tables, columns, enum values: always English.
```sql
-- Column names: English
employee_code, hired_date, is_head_office

-- Enum values: English UPPERCASE
CHECK (status IN ('ACTIVE', 'INACTIVE', 'TERMINATED'))
-- NOT: ('HOAT_DONG', 'KHONG_HOAT_DONG')
```

### 10.4 Code Comments → English or Vietnamese (consistent per file)
If starting a file with English comments, keep all comments in that file English.
If starting with Vietnamese, keep Vietnamese.
Don't mix within a file.

---

## 11. Anti-Patterns (Things Claude Code Must NEVER Do)

These are hard stops. If tempted to do any of these, stop and re-read the rules above.

| Anti-Pattern | Correct Approach |
|---|---|
| `SELECT *` in any query | Select explicit columns |
| Business logic in handlers | Move to UseCase layer |
| DB queries in UseCases | Move to Repository layer |
| Inline style in React | Tailwind classes |
| Hardcode `HRM_ENCRYPTION_KEY` | Always from `os.Getenv()` |
| Decrypt in SQL | Decrypt in application (UseCase) |
| Log PII plaintext values | Log field names only |
| Skip audit log on PII access | Always log EMPLOYEE_PII_ACCESSED |
| Invent field/table names not in SPEC | Use exact names from SPEC §11 |
| Create role codes not in SPEC §3.5 | Use only the 11 defined roles |
| Bypass middleware with inline role check | Use middleware.RequireRole() |
| Modify committed migrations | Create new migration |
| mock DB in integration tests | Use real test database |
| Silence errors without logging | Always `log.Printf` before returning 500 |
| `return nil, err` without wrap | `return nil, fmt.Errorf("context: %w", err)` |
| Add API paths not in SPEC §13 | Only paths from SPEC |
