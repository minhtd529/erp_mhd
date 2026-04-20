<!-- spec-version: 1.4 | last-sync: 2026-04-20 | changes: Org bounded context (branches/departments), audit-logs read API, web UI pages -->
> **Spec version**: 1.4 — Last sync: 2026-04-20 — Added Org bounded context, audit-logs read API, Users/Branches/AuditLogs web pages

# Module 0: Global - Shared Organization & Infrastructure

## Overview
Shared services for authentication, organization management, audit trails, notifications, and file storage used across all bounded contexts.

## Bounded Context: Global

### Responsibilities
- User & team member authentication (JWT + 2FA)
- Organization hierarchy (firm, branches, departments)
- Immutable audit logging for all mutations
- Real-time notifications (WebSocket)
- File storage & versioning
- System configuration

## Key Features

### 1. Authentication & Authorization
**Pattern**: Stateless JWT with TOTP 2FA
- JWT access token (15 min) + refresh token (7 days)
- TOTP-based 2FA for password login
- Push notification 2FA for critical operations
- Roles: SUPER_ADMIN, FIRM_PARTNER, AUDIT_MANAGER, AUDIT_STAFF, CLIENT_ADMIN, CLIENT_USER
- No hard role limits; RBAC + ABAC via policies

**2FA Enum Values**:
```
TWO_FACTOR_METHOD: TOTP | PUSH | EMAIL
TWO_FACTOR_STATUS: ENABLED | DISABLED | PENDING_SETUP
```

### 2. Organization Structure
**Entities**: Organization → Office → Department → Team
- Support multi-office firm structure
- Department-level resource allocation tracking
- Team leadership assignment

### 3. Audit Trail (pkg/audit)
**Every mutation MUST log via `pkg/audit.Log()`**:
- Table: `audit_logs`
- Fields: `id` (UUID), `created_at`, `entity_type`, `entity_id`, `action`, `old_data` (JSONB), `new_data` (JSONB), `created_by` (UUID), `ip_address`
- Immutable (no updates/deletes)
- Indexed on `entity_type`, `entity_id`, `created_at`

**Audit Actions**: CREATE, UPDATE, DELETE, APPROVE, REJECT, STATE_TRANSITION, LOCK, UNLOCK

### 4. Notification System
**Real-time**: WebSocket via `/events/stream?token=<JWT>&channels=global,engagement,timesheet`
**Channels**: `global`, `engagement`, `timesheet`, `billing`, `tax_advisory`
**Outbox Pattern**: Domain events → `outbox_messages` → Asynq worker → WebSocket broadcast

### 5. File Storage
**MinIO S3-compatible Storage**
- Table: `file_metadata` (id, bucket, key, version, created_by, created_at, size_bytes, mime_type, is_deleted)
- All file operations audit-logged
- Presigned URLs for secure access (1-hour TTL)

### 6. Distributed Locking
No critical concurrent operations in Global context.

## Code Structure

### Go Package Layout
```
internal/global/auth/
  ├── domain/          (UserForAuth, RefreshToken, TwoFactorChallenge, repository interfaces)
  ├── repository/
  │   ├── postgres.go           (UserRepo, RoleRepo, RefreshTokenRepo)
  │   ├── twofa_postgres.go     (TwoFARepository)
  │   ├── audit_postgres.go     (AuditLogRepo — ListAuditLogs with JOIN users)
  │   └── queries/              (sqlc .sql files)
  ├── usecase/
  │   ├── login.go              (LoginUseCase)
  │   ├── create_user.go        (CreateUserUseCase)
  │   ├── list_users.go         (ListUsersUseCase)
  │   ├── list_audit_logs.go    (ListAuditLogsUseCase + AuditLogQuerier interface)
  │   ├── enable_2fa.go / verify_setup.go / disable_2fa.go / ...
  │   └── dto.go
  └── handler/
      ├── auth_handler.go       (AuthHandler: /login /refresh /logout /me)
      ├── user_handler.go       (UserHandler: /users CRUD)
      ├── twofa_handler.go      (TwoFAHandler: /auth/2fa/*)
      ├── push_handler.go       (PushHandler: /auth/push/*, /push-devices/*)
      ├── audit_handler.go      (AuditHandler: GET /audit-logs)
      └── routes.go

internal/org/
  ├── domain/
  │   ├── entity.go   (Branch, Department + Params/Filter structs)
  │   ├── repository.go
  │   └── errors.go
  ├── usecase/
  │   ├── branch.go   (BranchUseCase: Create/Update/List/GetByID)
  │   └── department.go (DepartmentUseCase: Create/Update/List/GetByID)
  ├── repository/
  │   └── postgres.go
  └── handler/
      ├── branch_handler.go
      ├── department_handler.go
      └── routes.go
```

### Go Naming Examples
- `AuthHandler` — handles `/api/v1/auth/*`
- `UserHandler` — handles `/api/v1/users/*`
- `AuditHandler` — handles `GET /api/v1/audit-logs`
- `BranchHandler` / `DepartmentHandler` — handles `/api/v1/branches`, `/api/v1/departments`
- `ListAuditLogsUseCase` / `BranchUseCase` / `DepartmentUseCase` — use cases
- `AuditLogRepo` / `AuditLogQuerier` (interface) — repository pattern

## API Endpoints

### Authentication
| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | `/api/v1/auth/login` | Login with credentials | None |
| POST | `/api/v1/auth/2fa/verify` | Verify TOTP/Push | JWT pending_mfa |
| POST | `/api/v1/auth/refresh` | Refresh access token | Refresh token |
| POST | `/api/v1/auth/logout` | Logout | JWT |
| GET | `/api/v1/auth/2fa/setup` | Start 2FA setup | JWT |
| POST | `/api/v1/auth/2fa/confirm` | Confirm TOTP setup | JWT |
| GET | `/api/v1/me` | Current user profile | JWT |

**Enum Status Values**: `DRAFT`, `CONFIRMED`, `VERIFIED`, `ACTIVE`, `DISABLED`

### User Management
| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/api/v1/users` | List users (paginated) | SUPER_ADMIN, FIRM_PARTNER |
| POST | `/api/v1/users` | Create user | SUPER_ADMIN, FIRM_PARTNER |
| GET | `/api/v1/users/{id}` | Get user details | SUPER_ADMIN, FIRM_PARTNER, self |
| PUT | `/api/v1/users/{id}` | Update user | SUPER_ADMIN, FIRM_PARTNER, self |
| DELETE | `/api/v1/users/{id}` | Soft delete user | SUPER_ADMIN, FIRM_PARTNER |

**Fields** (snake_case): `id`, `email`, `full_name`, `role`, `org_id`, `is_active`, `created_at`, `created_by`, `updated_at`, `updated_by`

### Organization
| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/api/v1/organizations/{id}` | Get org structure | JWT |
| GET | `/api/v1/organizations/{id}/offices` | List offices | JWT |
| POST | `/api/v1/organizations/{id}/offices` | Create office | SUPER_ADMIN, FIRM_PARTNER |

### Audit Trail
| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/api/v1/audit-logs` | List audit logs (paginated, immutable, with JOIN to users) | SUPER_ADMIN, AUDIT_MANAGER |

**Query Params**: `?page=1&size=50&module=&resource=&action=&user_id=&from=YYYY-MM-DD&to=YYYY-MM-DD`

**Response fields per entry**: `id`, `user_id`, `user_name` (joined from users), `module`, `resource`, `resource_id`, `action`, `ip_address`, `created_at`

### Files
| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | `/api/v1/files/upload` | Upload file to MinIO | JWT |
| GET | `/api/v1/files/{id}/download` | Download file (presigned URL) | JWT |
| GET | `/api/v1/files/{id}/versions` | List file versions | JWT |

### Organization — Branches & Departments (`internal/org/`)

**Entities**: `Branch` (id, code, name, address, phone, is_active), `Department` (id, branch_id, code, name, is_active)

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/api/v1/branches` | List branches (paginated, ?q=, ?is_active=) | JWT |
| POST | `/api/v1/branches` | Create branch | SUPER_ADMIN, FIRM_PARTNER |
| GET | `/api/v1/branches/{id}` | Get branch by ID | JWT |
| PUT | `/api/v1/branches/{id}` | Update branch | SUPER_ADMIN, FIRM_PARTNER |
| GET | `/api/v1/departments` | List departments (?branch_id=, ?q=, ?is_active=) | JWT |
| POST | `/api/v1/departments` | Create department | SUPER_ADMIN, FIRM_PARTNER |
| GET | `/api/v1/departments/{id}` | Get department by ID | JWT |
| PUT | `/api/v1/departments/{id}` | Update department | SUPER_ADMIN, FIRM_PARTNER |

**Note**: Org is a separate bounded context (`internal/org/`) with own domain/repository/usecase/handler. All mutations emit audit log.

## Database Tables

### Core Tables
- `users` (id, email, password_hash, full_name, role, is_active, is_deleted, created_at, created_by, updated_at, updated_by)
- `branches` (id, code, name, address, phone, is_active, created_at, created_by, updated_at, updated_by)
- `departments` (id, branch_id, code, name, is_active, created_at, created_by, updated_at, updated_by)
- `audit_logs` (id, entity_type, entity_id, action, old_data JSONB, new_data JSONB, created_at, created_by, ip_address)
- `file_metadata` (id, bucket, object_key, file_name, file_size_bytes, mime_type, created_by, created_at, is_deleted)
- `two_factor_secrets` (id, user_id, secret, backup_codes JSONB, method, is_verified, created_at)
- `outbox_messages` (id, bounded_context, event_type, payload JSONB, created_at, processed_at)

### Indexes
- `idx_audit_logs_entity_type_id` on (entity_type, entity_id, created_at)
- `idx_file_metadata_created_by` on (created_by, created_at)
- `uidx_users_email` on (email) where is_deleted=false
- `uidx_organizations_name` on (name)

## CQRS & Event Publishing

**Read Model**: Query tables directly via sqlc for performance
**Write Model**: GORM for user/org/file mutations, publish to outbox_messages
**Events**: UserCreated, UserUpdated, OrganizationUpdated, FileUploaded

## Frontend Pages (apps/web)

| Page | Route | Description |
|------|-------|-------------|
| Users | `/users` | CRUD + role assignment; filterable by name/email; create/edit/delete dialog |
| Branches & Departments | `/branches` | Tabbed: Branch CRUD + Department CRUD (scoped by branch) |
| Audit Logs | `/audit-logs` | Read-only log viewer; filter by module, action, date range; color-coded action badges |

**Services**: `src/services/users.ts`, `src/services/branches.ts`, `src/services/audit.ts`

**Audit action badge colors**: CREATE=green, DELETE=red, UPDATE=default, UPDATE_BANK_DETAILS/UPDATE_STATUS/LOCK=yellow, APPROVE=green, REJECT=red, DEACTIVATE=ghost

## Error Codes
`INVALID_CREDENTIALS` — Login failure
`USER_NOT_FOUND` — User lookup failed
`TOTP_INVALID` — 2FA code incorrect
`BRANCH_NOT_FOUND` — Branch lookup failed
`DEPARTMENT_NOT_FOUND` — Department lookup failed
`FILE_NOT_FOUND` — File lookup failed
`AUDIT_LOG_IMMUTABLE` — Attempt to modify audit log
`INSUFFICIENT_PERMISSIONS` — RBAC/ABAC policy denial