# Claude Code Integration Guide

## Overview
This document provides guidance for integrating Claude Code with the ERP Audit System repository.

## Repository Structure
```
erp-audit/
├── docs/                    # Documentation
│   ├── SPEC.md             # Full technical specification
│   ├── ROADMAP.md          # Development roadmap
│   ├── DECISIONS.md        # Architectural decisions
│   └── modules/            # Module-specific documentation
│       ├── 00-global.md
│       ├── 01-crm.md
│       ├── 02-engagement.md
│       ├── 03-timesheet.md
│       ├── 04-billing.md
│       ├── 05-working-papers.md
│       ├── 06-tax-advisory.md
│       ├── 07-hrm.md
│       └── 08-reporting.md
├── .claude/                # Claude-specific configuration
│   └── commands/           # Custom Claude commands
├── apps/                   # Application code
│   ├── web/               # Next.js frontend
│   ├── mobile/            # React Native mobile app
│   └── api/               # Golang backend
├── packages/              # Shared packages
└── docker-compose.yml     # Development environment
```

## Claude Commands
Custom commands are available in the `.claude/commands/` directory for common development tasks.

## Development Workflow
1. Read the SPEC.md for complete system requirements
2. Check ROADMAP.md for current development phase
3. Review DECISIONS.md for architectural choices
4. Use module-specific docs in docs/modules/ for detailed requirements
5. Follow the established patterns and conventions

## Key Integration Points
- **Authentication**: JWT with 2FA support
- **Real-time**: WebSocket for notifications, self-hosted push relay
- **File Management**: MinIO for object storage
- **Database**: PostgreSQL with sqlc for type-safe queries
- **Frontend**: Next.js with Zustand + React Query
- **Backend**: Golang with Clean Architecture

## Architecture Patterns

- **Hexagonal Architecture**: domain → application → infrastructure → interfaces
- **CQRS**: GORM cho write, sqlc/pgx cho read
- **Domain Events + Outbox Pattern**: `outbox_messages` table → Asynq worker xử lý
- **DDD Bounded Contexts**: Global, CRM, Engagement, Timesheet, Billing, WorkingPapers, TaxAdvisory, HRM, Reporting
- **Distributed Locking**: Redis locks cho concurrent operations (edit engagement, assign resources, approve timesheet, sign-off working paper, process payment)

## Getting Started
1. Set up the development environment using docker-compose.yml
2. Review the SPEC.md document thoroughly
3. Start with Phase 1 modules (Global, CRM, HRM)
4. Use the provided API conventions and patterns

## Key References
- Full spec: `docs/SPEC.md`
- Current roadmap: `docs/ROADMAP.md`
- Module specs: `docs/modules/*.md`
- Commission business rules: `docs/SPEC.md` section 4.2.1
- Commission flow diagram: `docs/SPEC.md` section 4.2.1 (Commission Lifecycle)
- Commission calculation example: `docs/SPEC.md` section 4.2.1 (Ví dụ tính toán)

## Rules
1. ALWAYS read `docs/ROADMAP.md` first to know current phase
2. Follow module structure: domain → repository → usecase → handler
3. Every mutation MUST emit audit log via pkg/audit
4. Every list endpoint MUST use PaginatedResult[T]
5. Run `make lint test` before marking any task complete
6. Update `docs/ROADMAP.md` checkboxes as you progress
7. `CommissionRecord` là IMMUTABLE sau khi `status = approved` — KHÔNG UPDATE record đó; tạo clawback record mới (âm amount) để offset
8. Mọi trigger commission (invoice issued, payment received) phải IDEMPOTENT — dùng unique constraint `(engagement_commission_id, invoice_id)` hoặc `(engagement_commission_id, payment_id)`; kiểm tra trước khi insert
9. Event Billing → Commission đi qua OUTBOX PATTERN (ghi `outbox_messages` trong cùng DB transaction với invoice/payment mutation), không gọi CommissionService trực tiếp để tránh tight coupling và mất event khi crash

## Commands
- `make dev` — start dev environment
- `make migrate-up` — run migrations
- `make test` — run all tests
- `make lint` — golangci-lint + eslint

---

## API Conventions

### Naming
- **Endpoint paths**: `kebab-case` (e.g., `/list-clients`, `/my-engagements`)
- **JSON fields**: `snake_case`
- **Enum values**: `UPPER_SNAKE_CASE` (e.g., `DRAFT`, `APPROVED`, `IN_PROGRESS`)
- **IDs**: UUID format
- **Timestamps**: ISO 8601 (e.g., `"2026-04-16T10:00:00Z"`)

### HTTP Methods
- `GET` – read only
- `POST` – create new & state transitions (`/approve`, `/submit`)
- `PUT` – full update
- `DELETE` – soft delete (set `is_deleted=true`)
- `PATCH` – **không dùng**

### URL Patterns
- Collections: plural nouns `/clients`, `/engagements`, `/timesheets`
- State transitions: `/resource/{id}/action` (e.g., `/engagements/{id}/approve`)
- Authenticated user data: `/me/resource` (e.g., `/me/profile`, `/me/engagements`)
- Search param: `?q=...` (KHÔNG dùng `query`)

### Pagination
- Offset (default): `?page=1&size=20` (max 100)
- Cursor (large lists): `?cursor=...&size=20`

### Error Format
```json
{
  "error": "ERROR_CODE",
  "message": "Error message in English"
}
```
- Error codes: `UPPER_SNAKE_CASE` (e.g., `INVALID_CREDENTIALS`, `ENGAGEMENT_LOCKED`)
- Messages: English, user-friendly

### Versioning
- URL-based: `/api/v1`, `/api/v2`
- Bump version chỉ khi có **breaking changes** (xóa field, đổi type, đổi endpoint)
- Additive changes (thêm field, endpoint mới) → không bump version
- Deprecated endpoint giữ tối thiểu 6 tháng sau release mới

### WebSocket (Real-time)
- Endpoint: `GET /events/stream?token=<access_token>&channels=engagement,timesheet`
- Token trong query param (WebSocket không hỗ trợ custom headers)
- Timeout: 30 phút, client phải reconnect

### RBAC Roles
| Role | Code | Auth |
|------|------|------|
| System Admin | `SUPER_ADMIN` | Password 12+ chars + TOTP |
| Firm Partner | `FIRM_PARTNER` | Password 12+ chars + TOTP |
| Audit Manager | `AUDIT_MANAGER` | Password 12+ chars + 2FA |
| Audit Staff | `AUDIT_STAFF` | Password 12+ chars + 2FA |
| Client Administrator | `CLIENT_ADMIN` | Password + 2FA or SSO |
| Client User | `CLIENT_USER` | Password + 2FA or SSO |

---

## Database Conventions

- **Tables**: snake_case, số nhiều (e.g., `clients`, `engagements`, `timesheets`)
- **Columns**: snake_case (e.g., `client_name`, `created_at`, `is_deleted`)
- **Foreign Keys**: `{table_singular}_id` (e.g., `client_id`, `user_id`)
- **Indexes**: `idx_{table}_{columns}`
- **Unique Indexes**: `uidx_{table}_{columns}`
- **Materialized Views**: `mv_{purpose}` (e.g., `mv_billing_summary`)
- **Views**: `v_{purpose}`

### Key Rules
- **Deletion patterns**: Không có một pattern phù hợp cho tất cả — xem **SPEC.md §13.3 Deletion & Retention Conventions** để chọn đúng pattern (status-based, is_active, immutable/clawback, hard-delete+gate, hoặc is_deleted)
- **Primary Keys**: UUID cho tất cả bảng
- **Audit fields**: `created_at`, `created_by`, `updated_at`, `updated_by`
- **Enums**: Dùng PostgreSQL CHECK constraints (KHÔNG dùng ENUM type)
- **Full-text search**: `pg_trgm` extension + GIN index
- **Snapshot data**: Lưu JSONB, bất biến sau khi confirm – KHÔNG đọc lại từ live data
- **Distributed locks**: Redis locks cho concurrent write operations

---

## Go Code Naming Conventions

| Type | Pattern | Example |
|------|---------|---------|
| Handlers | `{Entity}Handler` | `EngagementHandler`, `TimesheetHandler` |
| Use Cases | `{Entity}UseCase` (bundled) or `{Action}{Entity}UseCase` (per-action) | `ClientUseCase`, `LoginUseCase` |
| Repositories | `{Entity}Repository` | `ClientRepository`, `TimesheetRepository` |
| Request DTOs | `{Entity}{Op}Request` | `EngagementCreateRequest` |
| Response DTOs | `{Entity}Response` (single shape) or `{Entity}{Op}Response` (multiple shapes) | `PlanResponse`, `ClientDetailResponse` |

### Use Case Struct Naming

**Pattern 1 — Bundled (default):** One struct per entity, methods per action. Use when all CRUD methods share the same dependencies.

```go
type ClientUseCase struct { repo ClientRepository; auditLog AuditLogger }
func (uc *ClientUseCase) Create(...) (*Client, error)
func (uc *ClientUseCase) Update(...) (*Client, error)
```

**Pattern 2 — Per-action (when deps diverge):** Separate struct per action. Use when operations have significantly different dependency sets.

```go
type LoginUseCase struct { userRepo UserRepository; tokenSvc TokenService; rateLimiter RateLimiter }
type CreateUserUseCase struct { userRepo UserRepository; hasher PasswordHasher }
```

Example: the `auth` module uses per-action because `Login` needs rate limiter + token service while `CreateUser` needs only a hasher. Bundling would force every call site to construct unused dependencies.

**Rule:** Don't mix patterns within a single module. Pick one per module.

### Response DTO Naming

**Single shape** — use plain `{Entity}Response` when one DTO serves all endpoints (list, detail, embed):

```go
type PlanResponse struct { ... }      // used by GET /plans/:id, GET /plans, embedded in EngCommissionResponse
type RecordResponse struct { ... }    // used by all commission record endpoints
```

**Multiple shapes** — add suffix only when you have 2+ genuinely different response shapes:

```go
type ClientListItem struct { ... }       // lean, for GET /clients
type ClientDetailResponse struct { ... } // includes relations, for GET /clients/:id
```

**Rule:** Don't introduce `*DetailResponse` suffix if there is no matching `*ListItem` or `*Summary`. Single-shape entities stay plain `{Entity}Response`.

---

## Frontend (React/Next.js) Conventions

- **Components**: PascalCase, 1 file mỗi component (e.g., `EngagementForm.tsx`)
- **Hooks**: camelCase, prefix `use` (e.g., `useEngagement`, `useTimesheet`)
- **Services**: Organize by domain in `src/services/` (e.g., `engagementService.ts`)
- **Pages**: App Router in `src/app/`

---

## UI / Style Guide – "Professional Audit"

### Philosophy
Modern Professional + Dark Audit Aesthetic – Trust, Authority, Clarity

### Color Palette
| Name | Hex | Usage |
|------|-----|-------|
| Primary | `#1F3A70` | Deep navy – headers, primary actions |
| Secondary | `#D4A574` | Gold – accents, secondary elements |
| Background | `#F5F5F5` | Light gray – page background |
| Action/CTA | `#2E5090` | Soft blue – main buttons |
| Success | `#2D6A4F` | Green – approval, confirmation |
| Danger | `#9B2226` | Red – critical actions, errors |
| Surface-Paper | `#FFFFFF` | Cards, tables, panels |
| Border | `#E0E0E0` | Subtle borders |
| Text-Primary | `#1A1A1A` | Dark – main text |
| Text-Secondary | `#5A5A5A` | Gray – supporting text, hints |

### Typography
- **Headings**: Inter or Segoe UI (professional, clean)
- **Body**: Inter (excellent readability, international support)
- **Monospace**: Fira Code (code snippets, data)

### Components
- **Button Primary**: Navy bg, white text, border-radius 6px, 2px solid border
- **Button Secondary**: Transparent, navy border & text
- **Button Danger**: Red bg, white text
- **Input**: White bg, light gray border. Focus → navy border + subtle shadow
- **Table header**: Light gray bg; rows alternating white/off-white; zebra striping
- **Card**: White bg, subtle shadow (1px), border-radius 8px
- **Modal**: Dark overlay (40% opacity), white card, close button top-right

### Decorative
- **Dashboard watermark**: Firm/project logo, ~2% opacity bottom-right
- **Divider**: Thin line, simple style (no decorative elements)
- **Loading**: Subtle spinner or progress bar (no animation overkill)
- **Toast Success**: Light green bg, dark green checkmark, auto-dismiss 5s
- **Toast Error**: Light red bg, dark red error icon, dismissible

## Migration Rules (CRITICAL — read before touching apps/api/migrations/)

### Immutability

Once a migration is committed to `main`, it is **immutable forever**:
- NEVER rename an existing migration file
- NEVER change the version number
- NEVER modify the SQL content of a committed migration
- NEVER delete a committed migration

If committed schema is wrong → create a NEW migration with higher version that corrects it.
The GitHub Actions `migration-lint` workflow blocks PRs that violate these rules.

### Before writing a new migration

1. Use the scaffold tool (don't hand-pick version numbers):
```bash
make migrate-create NAME=add_users_status_column
```

2. Check for duplicate table definitions before adding `CREATE TABLE`:
```bash
grep -ril "CREATE TABLE.*<table_name>" apps/api/migrations/*.up.sql
```
If already exists → use `ALTER TABLE` in new migration, not `CREATE`.

3. Write the down migration symmetrically:
   - `CREATE TABLE` in up → `DROP TABLE` in down
   - `ADD COLUMN` in up → `DROP COLUMN` in down
   - Test round-trip: `make migrate-up && make migrate-down && make migrate-up` must all succeed

4. Validate before commit:
```bash
make migrate-lint
```

### What the CI lint checks (`scripts/migration-lint.sh`)

- Every `.up.sql` has a matching `.down.sql`
- Version sequence has no gaps (1, 2, 3 … no missing numbers)
- No two migrations create the same table (`CREATE TABLE` dedup check)
- Filenames follow `000NNN_snake_case.{up,down}.sql` format
- No modification or deletion of migrations that exist in `origin/main` (PR check only)

### Emergency recovery

**If migration was accidentally renumbered/modified and already pushed:**
- DO NOT force-push to fix — other DBs will break
- Create a new migration that brings schema to the correct state
- Document the incident in CHANGELOG.md

**If it's still local only (not pushed):**
```bash
git restore apps/api/migrations/   # undo changes
docker compose down -v && make migrate-up  # reset local DB
```