Read docs/SPEC.md and docs/CLAUDE.md, then audit the current codebase for compliance.

Check each category below and report findings as ✓ PASS, ✗ FAIL, or ⚠ PARTIAL.

---

## 1. API Conventions

- [ ] All endpoint paths use kebab-case
- [ ] All JSON fields use snake_case
- [ ] All enum values use UPPER_SNAKE_CASE
- [ ] All IDs are UUID format
- [ ] All timestamps are ISO 8601
- [ ] PATCH is not used anywhere — only GET / POST / PUT / DELETE
- [ ] Error responses follow `{"error": "UPPER_SNAKE", "message": "..."}` format
- [ ] Every list endpoint has pagination (`page`, `size` params, returns `PaginatedResult[T]`)
- [ ] No endpoint returns unbounded arrays

## 2. RBAC & Security

- [ ] Every handler that mutates data is protected by `AuthMiddleware`
- [ ] Admin-only endpoints use `RequireRole(...)` or `RequirePermission(...)`
- [ ] Passwords are bcrypt cost ≥ 12
- [ ] Refresh tokens are stored as SHA-256 hash, never plaintext
- [ ] TOTP secrets are AES-256-GCM encrypted at rest
- [ ] No secrets or credentials appear in source code (hardcoded keys, passwords)

## 3. Audit Trail

- [ ] Every CREATE mutation calls `auditLog.Log(...)` with action = "CREATE_*"
- [ ] Every UPDATE mutation calls `auditLog.Log(...)` with action = "UPDATE_*"
- [ ] Every DELETE mutation calls `auditLog.Log(...)` with action = "DELETE_*"
- [ ] Every state transition calls `auditLog.Log(...)` with the transition action
- [ ] `audit_logs` table is never updated or deleted (immutable)

## 4. Database Conventions

- [ ] All tables use UUID primary keys with `gen_random_uuid()`
- [ ] No hard deletes — soft deletes via `is_deleted` / `is_void` / `is_active`
- [ ] All tables have `created_at`, `updated_at` audit columns
- [ ] Foreign keys follow `{table_singular}_id` naming
- [ ] Indexes follow `idx_{table}_{columns}` naming
- [ ] No PostgreSQL ENUM types — CHECK constraints only

## 5. Clean Architecture

- [ ] Domain layer has zero imports from repository/usecase/handler layers
- [ ] Use cases depend only on domain interfaces, not concrete repository types
- [ ] Handlers depend only on use case structs, never on repository directly
- [ ] No business logic in handlers (only bind → call usecase → respond)
- [ ] No SQL in use cases or domain (SQL lives in repository only)

## 6. Error Handling

- [ ] `ErrUserNotFound` is never leaked to clients — masked as `ErrInvalidCredentials` in login
- [ ] All `pgx.ErrNoRows` paths return domain sentinel errors, not raw DB errors
- [ ] Unique constraint violations return domain-level errors (e.g., `ErrUserAlreadyExists`)
- [ ] Internal errors return `INTERNAL_ERROR` to client, never raw error messages

## 7. Code Naming

- [ ] Handlers named `{Entity}Handler`
- [ ] Use cases named `{Action}{Entity}UseCase`
- [ ] Repositories named `{Entity}Repository` (interfaces) / `Repo` (concrete)
- [ ] Request DTOs named `{Entity}{Op}Request`
- [ ] Response DTOs named `{Entity}{Op}Response` or `{Entity}DetailResponse`

---

## Output Format

For each category, list:
- Items that PASS (one line each)
- Items that FAIL with file:line reference and what needs fixing
- Items that are PARTIAL with explanation

End with a prioritised fix list: CRITICAL (security/data integrity) → HIGH (spec violations) → LOW (naming/style).
