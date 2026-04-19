# ERP Audit System

Hệ thống ERP cho Công ty Kiểm toán – Tư vấn Tài chính – Thuế.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.22 + Gin + pgx/v5 + sqlc |
| Frontend | Next.js 14 (App Router) + Shadcn/UI + Tailwind CSS |
| Database | PostgreSQL 16 |
| Cache | Redis 7 |
| Object Storage | MinIO (S3-compatible) |
| Message Queue | NATS 2.10 |
| Monorepo | Turborepo + pnpm workspaces |

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.22+ | https://go.dev/dl/ |
| Node.js | 20+ | https://nodejs.org |
| pnpm | 10+ | `npm i -g pnpm` |
| Docker + Compose | 24+ | https://docs.docker.com/get-docker/ |
| golangci-lint | latest | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` |
| golang-migrate | latest | `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest` |
| sqlc | latest | `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` |
| air (hot reload) | latest | `go install github.com/air-verse/air@latest` |

## Quick Start

### 1. Clone & install dependencies

```bash
git clone <repo-url> erp-audit
cd erp-audit
pnpm install           # Install web dependencies
cd apps/api && go mod download && cd ../..
```

### 2. Configure environment

```bash
# API
cp apps/api/.env.example apps/api/.env

# Web
cp apps/web/.env.local.example apps/web/.env.local
```

### 3. Start infrastructure

```bash
make dev-infra
# Starts: PostgreSQL :5432, Redis :6379, MinIO :9000/:9001, NATS :4222
```

### 4. Run migrations

```bash
make migrate-up
# Or with custom DB URL:
DB_URL=postgres://erp:erp@localhost:5432/erp_audit?sslmode=disable make migrate-up
```

### 5. Start development servers

```bash
# Start everything (API + Web + infra)
make dev

# Or start separately:
make dev-api   # API on :8080 (hot reload via air)
make dev-web   # Web on :3000
```

> **Note — air config is OS-aware.**  
> `make dev-api` automatically selects `.air.toml` on Windows (builds `tmp/server.exe`)  
> and `.air.unix.toml` on Linux/macOS (builds `tmp/server`).  
> Do not pass `-c` manually; let the Makefile choose the right config.

### 6. Verify

```bash
curl http://localhost:8080/health
# → {"status":"ok","version":"v1","env":"development"}

open http://localhost:3000
```

## Available Commands

```bash
make dev            # Start full dev environment
make dev-infra      # Start only Docker services
make dev-api        # Start Go API with hot reload
make dev-web        # Start Next.js dev server
make stop           # Stop all Docker services

make test           # Run all tests
make test-api       # Go tests with race detector + coverage
make test-web       # Vitest tests

make lint           # Run all linters
make lint-api       # golangci-lint
make lint-web       # ESLint

make build          # Build all apps
make build-api      # Build Go binary → apps/api/bin/server
make build-web      # Build Next.js → apps/web/.next/

make migrate-up     # Apply pending migrations
make migrate-down   # Roll back last migration
make migrate-create name=add_clients_table   # New migration

make sqlc           # Generate type-safe Go from SQL
make generate       # All code generators
make clean          # Remove build artifacts
make help           # Show all targets
```

## Project Structure

```
erp-audit/
├── apps/
│   ├── api/                    # Go backend
│   │   ├── cmd/server/         # Entry point (main.go)
│   │   ├── internal/           # Business logic (per module)
│   │   │   ├── global/         # Auth, users, audit trail
│   │   │   ├── crm/            # Client management
│   │   │   ├── engagement/     # Audit engagements
│   │   │   ├── timesheet/      # Timesheets & resources
│   │   │   ├── billing/        # Invoicing & billing
│   │   │   ├── workingpaper/   # Audit working papers
│   │   │   ├── tax/            # Tax advisory
│   │   │   ├── hrm/            # HR management
│   │   │   └── reporting/      # Analytics & reports
│   │   ├── pkg/                # Shared packages
│   │   │   ├── audit/          # Immutable audit trail writer
│   │   │   ├── config/         # App configuration
│   │   │   ├── database/       # Connection pool + migrator
│   │   │   ├── logger/         # Zap logger
│   │   │   └── middleware/     # HTTP middleware (auth, cors, logger)
│   │   └── migrations/         # SQL migration files
│   └── web/                    # Next.js 14 frontend
│       └── src/
│           ├── app/            # App Router pages
│           ├── lib/            # API client, providers
│           └── types/          # TypeScript types
├── docker-compose.yml          # Dev infrastructure
├── Makefile                    # Developer commands
├── turbo.json                  # Turborepo config
└── docs/                       # Specs, roadmap, decisions
```

## Module Architecture (Clean Architecture)

Each backend module follows:

```
internal/<module>/
├── domain/         # Entities, repository interfaces, domain errors
├── usecase/        # Business logic + DTOs
├── repository/     # PostgreSQL implementation + sqlc queries
├── handler/        # HTTP handlers + route registration
└── wire.go         # Dependency injection
```

## Services & Ports

| Service | Port | UI |
|---------|------|----|
| API (Go) | 8080 | http://localhost:8080/health |
| Web (Next.js) | 3000 | http://localhost:3000 |
| PostgreSQL | 5432 | — |
| Redis | 6379 | — |
| MinIO API | 9000 | — |
| MinIO Console | 9001 | http://localhost:9001 |
| NATS | 4222 | — |
| NATS Monitor | 8222 | http://localhost:8222 |

MinIO Console credentials: `minioadmin` / `minioadmin`

## API Conventions

- **Base URL**: `http://localhost:8080/api/v1`
- **Auth**: `Authorization: Bearer <access_token>`
- **Pagination**: `?page=1&size=20`
- **Error format**: `{"error": "ERROR_CODE", "message": "..."}`
- **IDs**: UUID format
- **Timestamps**: ISO 8601

## Database

- **Migrations**: `apps/api/migrations/` (golang-migrate, sequential numbering)
- **Queries**: `apps/api/internal/*/repository/queries/*.sql` (sqlc generates Go code)
- **Generated code**: `apps/api/pkg/db/`
- **Conventions**: snake_case tables, UUIDs for PKs, soft deletes via `is_deleted`

## Contributing

1. Create feature branch from `develop`
2. Follow module structure: `domain → repository → usecase → handler`
3. Every mutation MUST emit audit log via `pkg/audit`
4. Run `make lint test` before opening PR
5. Update `docs/ROADMAP.md` checkbox when task is complete
