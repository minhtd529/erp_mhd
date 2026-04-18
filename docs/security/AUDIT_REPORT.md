# Security Audit Report — Phase 1 Baseline
**Date:** 2026-04-18  
**Scope:** Phase 1 codebase (`apps/api`) vs SPEC.md §12.3  
**Standard:** OWASP ASVS Level 2  
**Status:** DRAFT — Awaiting approval before Phase B execution

---

## 1. Controls Implemented in Phase 1

### 1.1 Authentication & Session Management

| Control | Spec Requirement | Implementation | Status |
|---|---|---|---|
| JWT access token TTL | 15 min | `pkg/auth/jwt.go` — `accessTokenTTL` configurable, default 15 min | ✅ Done |
| JWT refresh token TTL | 7 days | Opaque UUID, SHA-256 hashed in DB, default 7 days | ✅ Done |
| Token rotation on refresh | Required | `usecase/refresh.go` — old token deleted on use | ✅ Done |
| Password hashing | bcrypt cost=12 | `pkg/auth/password.go` — `bcryptCost = 12` | ✅ Done |
| Brute-force lock | 5 fails → 15 min | `usecase/login.go` — `maxLoginAttempts=5`, `lockDuration=15m` | ✅ Done |
| Credential masking | Don't reveal which field | `login.go:57` — `ErrUserNotFound` mapped to `ErrInvalidCredentials` | ✅ Done |
| Refresh token hash | Store hash only | `HashRefreshToken` (SHA-256), raw transmitted only | ✅ Done |
| JWT algorithm validation | Reject non-HMAC | `jwt.go:61` — method type assertion on parse | ✅ Done |

### 1.2 Two-Factor Authentication (TOTP)

| Control | Spec Requirement | Implementation | Status |
|---|---|---|---|
| TOTP generation | RFC 6238, SHA-1, 6 digits, 30s | `pkg/auth/totp.go` — `pquerna/otp` | ✅ Done |
| TOTP secret at rest | AES-256-GCM, key via env | `pkg/crypto/aes.go` — AES-256-GCM with random nonce | ✅ Done |
| Challenge TTL | 5 min | `login.go:17` — `challengeTTL = 5 * time.Minute` | ✅ Done |
| Backup codes | 10 codes, bcrypt, single-use | `pkg/auth/backup_codes.go` — 10×8 chars, bcrypt cost=10 | ✅ Done |
| Trusted device | SHA-256 fingerprint, 30 days, max 5 | `backup_codes.go:48` — `HashDeviceFingerprint` | ✅ Done |
| 2FA challenge attempts | Max 5 → invalidate | `twofa_postgres.go` — `attempt_count` + `invalidated_at` | ✅ Done |
| QR code self-hosted | No external API | `skip2/go-qrcode` — renders PNG locally | ✅ Done |

### 1.3 Authorization

| Control | Spec Requirement | Implementation | Status |
|---|---|---|---|
| RBAC | Role-based access | `middleware/middleware.go` — `RequireRole(...)` | ✅ Done |
| ABAC | Branch/dept scoping | Claims carry `branch_id`, `department_id` in JWT | ⚠️ Partial — in JWT but not enforced at DB query level |
| Permission check | module:resource:action | `RequirePermission(module, resource, action)` | ✅ Done |

### 1.4 Data Protection

| Control | Spec Requirement | Implementation | Status |
|---|---|---|---|
| TOTP secret encrypted | AES-256-GCM | `pkg/crypto/aes.go` — used in enable_2fa.go | ✅ Done |
| Bank account encrypted | Sensitive PII | `HRMConfig.BankEncryptionKey` — AES-256-GCM in hrm usecase | ✅ Done |
| Soft delete | No hard deletes | `is_deleted` flag, all handlers | ✅ Done |

### 1.5 Audit Trail

| Control | Spec Requirement | Implementation | Status |
|---|---|---|---|
| Immutable audit log | Append-only | `pkg/audit/audit.go` — INSERT only, no UPDATE/DELETE | ✅ Done |
| Login audit | Log successful logins | `login.go:158` — `LOGIN` action logged | ✅ Done |
| Mutation audit | Log CREATE/UPDATE/DELETE | CRM, HRM, user mutations all emit audit entries | ✅ Done |

### 1.6 Infrastructure

| Control | Spec Requirement | Implementation | Status |
|---|---|---|---|
| API versioning | /api/v1 prefix | All routes under `/api/v1` | ✅ Done |
| CORS | Whitelist origins | `middleware.CORS(allowedOrigins)` | ⚠️ Wildcard `*` in main.go |
| Request logging | Log all requests | `middleware.RequestLogger(zap)` | ✅ Done |
| Graceful shutdown | Clean shutdown | `main.go` — SIGTERM handling with 30s timeout | ✅ Done |
| CI pipeline | lint + test | `.github/workflows/ci.yml` | ✅ Done |

---

## 2. Missing Controls (Not Yet Implemented)

### 2.1 Critical Gaps

| # | Control | Spec Requirement | Current State | Risk |
|---|---|---|---|---|
| C1 | **Rate limiting** | 100 req/min/user, 1000 req/min/IP | No rate limiting middleware at all | Brute-force, DDoS amplification |
| C2 | **Security HTTP headers** | TLS 1.3, HSTS, CSP, X-Frame-Options | No headers set beyond CORS | Clickjacking, MIME-sniffing, XSS |
| C3 | **Config startup validation** | Secrets must be non-default in prod | `JWT_SECRET="change-me-in-production"` and `TOTP_ENCRYPTION_KEY="0000...000"` are valid defaults | Key leakage if misconfigured |
| C4 | **TLS enforcement** | TLS 1.3, sslmode=require | `sslmode=disable` default in `.env` and config defaults | DB traffic unencrypted |
| C5 | **Secrets scanning in CI** | No secrets in git | No gitleaks or detect-secrets in CI or pre-commit | Secret leakage via git history |

### 2.2 High Gaps

| # | Control | Spec Requirement | Current State | Risk |
|---|---|---|---|---|
| H1 | **CSRF protection** | SameSite cookies + CSRF token | Tokens are Bearer (partially mitigates), but no CSRF middleware | CSRF on state-changing endpoints |
| H2 | **Failed login audit log** | Audit trail for all operations | Failed logins increment counter but do NOT write to `audit_logs` table | No forensic trail for brute-force attacks |
| H3 | **Single active session** | Single active session per user (configurable) | Multiple refresh tokens allowed; no per-user session limit enforcement | Token accumulation, session hijack risk |
| H4 | **ABAC DB enforcement** | Branch/dept scoping at data level | Branch/dept IDs in JWT claims but queries don't filter by them | Cross-branch data access by valid users |
| H5 | **Input validation completeness** | go-playground/validator on all endpoints | Validator in go.mod as indirect; DTO struct tags inconsistent — binding only, no validate tags | Input injection, data integrity |
| H6 | **WebSocket token in URL** | Spec acknowledges limitation | `?token=<JWT>` logged by `RequestLogger` in access log | JWT leakage via log aggregation |

### 2.3 Medium Gaps

| # | Control | Spec Requirement | Current State | Risk |
|---|---|---|---|---|
| M1 | **Column-level PII encryption** | PII column encryption | Only TOTP secret + bank account encrypted; email, full_name, phone stored plaintext | PII exposure on DB breach |
| M2 | **Push 2FA** | Push-based 2FA support | Not implemented | User friction; only TOTP available |
| M3 | **Mandatory 2FA enforcement** | Admin enforce for Partner/Director | No admin enforcement endpoint or policy check | High-privilege accounts bypass 2FA |
| M4 | **File upload security** | Type validation, 50MB limit, ClamAV | MinIO configured but no upload handlers, no virus scan | Malware upload, path traversal |
| M5 | **Trusted device server-side validation** | Max 5 devices, 30-day TTL | Max 5 and TTL exist in config but cleanup cron/expiry not verified in DB | Device list grows unbounded |
| M6 | **Active sessions API** | List/revoke active sessions | `/auth/trusted-devices` exists but no active sessions endpoint | Users can't see/revoke active sessions |
| M7 | **Govulncheck in CI** | Dependency vulnerability scanning | Only `golangci-lint` + tests; no `govulncheck` | Known CVEs undetected |
| M8 | **npm audit in CI** | Frontend dependency audit | No `npm audit` in CI workflow | Vulnerable JS packages |

### 2.4 Low Gaps

| # | Control | Spec Requirement | Current State | Risk |
|---|---|---|---|---|
| L1 | **Data retention policy** | 10 years for audit records | Not enforced in code | Compliance gap |
| L2 | **Password complexity** | Min 8 chars + complexity | Only bcrypt hashing; no length/complexity validator on create-user | Weak passwords accepted |
| L3 | **KMS integration** | Keys managed by KMS | Encryption keys in env vars only | Key management risk at scale |
| L4 | **Immutability enforcement for audit_logs** | No delete/update | INSERT-only code, but no DB-level constraint (trigger/policy) | DBA can alter audit records |
| L5 | **OpenAPI security scheme** | API spec with auth docs | No OpenAPI spec generated | Developer misuse of unauthenticated endpoints |
| L6 | **Security.txt / Disclosure policy** | Not in spec but ASVS V14.5 | No `.well-known/security.txt` | No responsible disclosure path |

---

## 3. Architectural Vulnerabilities (OWASP ASVS Level 2)

### 3.1 ASVS V2 — Authentication

**V2.1 Password Security**  
- `pkg/auth/password.go` uses bcrypt cost=12 ✅  
- **GAP**: No password complexity/length enforcement at `create_user.go` or `update_password` path. The validator library is imported as indirect (via gin) but not applied. An 8-character minimum as specified is NOT enforced.

**V2.2 General Authenticator Security**  
- Brute-force protection exists ✅  
- **GAP**: TOTP challenge attempts (5 max) are tracked but the counter is not rate-limited at the HTTP layer — an attacker can create unlimited new challenges (via /login) after each 5-attempt lockout.

### 3.2 ASVS V3 — Session Management

**V3.2 Session Binding**  
- Refresh tokens hashed and stored with device_id and IP ✅  
- **GAP**: No absolute maximum session count per user. A compromised account can create tokens indefinitely.

**V3.7 Logout**  
- `logout.go` deletes single refresh token ✅  
- `logoutAll` revokes all tokens ✅  
- **GAP**: JWT access token not invalidated on logout (no token blocklist/Redis revocation). A stolen access token remains valid for up to 15 minutes after logout.

### 3.3 ASVS V4 — Access Control

**V4.1 General Access Control**  
- RBAC via `RequireRole` and `RequirePermission` ✅  
- **GAP**: Branch/department ABAC is only in JWT claims. The CRM and HRM repositories do not filter by `branch_id` or `department_id`. A user from Branch A with a valid JWT can read/modify Branch B records.

### 3.4 ASVS V5 — Validation and Encoding

**V5.1 Input Validation**  
- Gin's `ShouldBindJSON` validates basic types ✅  
- `go-playground/validator` present as transitive dep  
- **GAP**: DTO structs (`usecase/dto.go`) lack `validate:` struct tags. No explicit `c.ShouldBindJSON` followed by `validate.Struct(req)`. Server-side field-level validation (min length, regex, required fields) is absent beyond JSON parsing.

### 3.5 ASVS V7 — Error Handling and Logging

**V7.1 Log Content**  
- Zap structured logging ✅  
- Audit log on success ✅  
- **GAP**: Failed authentication attempts (wrong password, invalid OTP, backup code failure) are NOT written to `audit_logs`. Only the attempt counter is incremented. Security investigations have no forensic trail of failed attempts.

**V7.2 Log Protection**  
- **GAP**: Access logs include WebSocket JWT token in query string (`/events/stream?token=...`). Anyone with access to application logs can replay tokens for up to 15 minutes.

### 3.6 ASVS V8 — Data Protection

**V8.1 General Data Protection**  
- Sensitive data at rest encrypted for TOTP + bank account ✅  
- **GAP**: Broad PII (email, full_name, phone_number, national_id) stored as plaintext PostgreSQL columns. A DB credential leak or unintended SELECT exposes all PII without additional barriers.

### 3.7 ASVS V9 — Communication Security

**V9.1 Client Communication Security**  
- **GAP**: No TLS enforcement in API server. `srv.ListenAndServe()` without TLS. Production relies on reverse proxy, but there is no code-level guard. DB connection defaults to `sslmode=disable`.

### 3.8 ASVS V10 — Malicious Code

**V10.3 Deployed Application Integrity**  
- **GAP**: No `govulncheck` in CI. No `gitleaks` pre-commit hook. No `gosec` static analysis. Supply chain and secret leakage undetected until manual review.

### 3.9 ASVS V12 — Files and Resources

**V12.1 File Upload**  
- MinIO configured, but no upload handlers exist yet  
- **GAP** (future): When upload endpoints are added, there is no established framework for type validation, size limits, or ClamAV scanning.

### 3.10 ASVS V13 — API and Web Service

**V13.1 Generic Web Service Security**  
- **GAP**: No security headers middleware. Missing:
  - `Strict-Transport-Security` (HSTS)
  - `Content-Security-Policy`
  - `X-Content-Type-Options: nosniff`
  - `X-Frame-Options: DENY`
  - `Referrer-Policy: strict-origin-when-cross-origin`
  - `Permissions-Policy`

**V13.2 RESTful Web Service**  
- **GAP**: No API-level rate limiting. Redis is configured but not used for rate limiting. Rate limiting is deferred to Phase 5.

---

## 4. Priority Matrix

### CRITICAL — Must fix before Phase 2 starts

| ID | Finding | ASVS | Effort |
|---|---|---|---|
| C3 | Config startup validation — reject zero/default secrets in production | V2.10 | S |
| C5 | Add gitleaks to CI (scan for committed secrets) | V10.3 | S |
| C2 | Security HTTP headers middleware (HSTS, CSP, X-Frame, nosniff) | V13.1 | S |
| H2 | Write `LOGIN_FAILED` audit log entry on authentication failures | V7.1 | S |
| H5 | Add `validate:` struct tags + explicit validation on all auth DTOs | V5.1 | M |

### HIGH — Fix before Phase 2 go-live

| ID | Finding | ASVS | Effort |
|---|---|---|---|
| C1 | Rate limiting middleware (Redis token bucket, 100/min user, 1000/min IP) | V13.2 | M |
| C4 | TLS: `sslmode=require` default; startup warning if `ENV=production` and no TLS | V9.1 | S |
| H3 | Single active session: enforce max refresh token count per user | V3.2 | M |
| H4 | ABAC enforcement: filter DB queries by `branch_id` in CRM and HRM | V4.1 | M |
| H6 | WebSocket auth: move token to `Authorization` header or short-lived ticket, stop logging URL tokens | V3.7 | M |
| M3 | Mandatory 2FA enforcement API for high-privilege roles | V2.2 | M |

### MEDIUM — Fix before Phase 3 go-live

| ID | Finding | ASVS | Effort |
|---|---|---|---|
| L2 | Password complexity enforcement: min 8 chars, uppercase, digit, special | V2.1 | S |
| M1 | Column-level encryption for PII (email, phone) or pseudonymisation strategy | V8.1 | L |
| M7 | `govulncheck` in CI on every PR | V10.3 | S |
| M8 | `npm audit --audit-level=high` in CI | V10.3 | S |
| H1 | CSRF protection: SameSite=Strict on session cookies (if any), document Bearer-only rationale | V4.10 | S |
| M4 | File upload security framework (type check, size limit, ClamAV integration stub) | V12.1 | M |
| M5 | Expired trusted device cleanup (cron or DB expiry enforcement) | V3.2 | S |

### LOW — Track for Phase 5 hardening

| ID | Finding | ASVS | Effort |
|---|---|---|---|
| L4 | DB-level audit log immutability (PostgreSQL row-level security or write-only role) | V7.2 | M |
| L3 | KMS integration for encryption key management | V6.4 | L |
| L1 | Data retention policy: scheduled purge/archive of non-audit data | V8.3 | M |
| M6 | Active sessions list/revoke API | V3.7 | S |
| L5 | OpenAPI spec with security scheme documentation | V13.2 | M |

---

## 5. Summary Scorecard

| Category | Score | Notes |
|---|---|---|
| Authentication | 7/10 | JWT+bcrypt+2FA solid; missing rate limit, password complexity, config guard |
| Session Management | 5/10 | Rotation ✅; no session limit, access token not revocable on logout |
| Access Control | 5/10 | RBAC solid; ABAC claims-only, not DB-enforced |
| Input Validation | 4/10 | Type binding only; no field-level validation rules |
| Cryptography | 8/10 | AES-256-GCM, bcrypt, SHA-256 — good primitives; no KMS |
| Error Handling / Logging | 5/10 | Success audit trail good; failed attempts not logged |
| Data Protection | 5/10 | TOTP + bank encrypted; broad PII plaintext |
| Infrastructure Security | 3/10 | No TLS guard, no security headers, no rate limiting, no secrets scanning |
| **Overall Phase 1** | **5.2/10** | Foundation solid; infrastructure security layer not yet built |

---

_This report covers only what is visible in the current codebase. Review should be repeated after each phase. Approve Phase B to begin remediation._
