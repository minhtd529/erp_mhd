.PHONY: dev dev-infra dev-api dev-web stop \
        test test-api test-web \
        lint lint-api lint-web \
        build build-api build-web \
        migrate-up migrate-down migrate-create \
        generate sqlc \
        clean help

# ── Variables ──────────────────────────────────────────────────────────────────
API_DIR    := apps/api
WEB_DIR    := apps/web
MIGRATIONS := apps/api/migrations
DB_URL     ?= postgres://erp:erp@localhost:5433/erp_audit?sslmode=disable

ifeq ($(OS),Windows_NT)
  AIR_CONFIG := .air.toml
  BASH := C:/Program Files/Git/bin/bash.exe
  SHELL := C:/Program Files/Git/bin/bash.exe
else
  AIR_CONFIG := .air.unix.toml
  BASH := bash
endif

# ── Dev ───────────────────────────────────────────────────────────────────────
## dev: Start full development environment (infra + api + web)
dev: dev-infra
	@echo ">>> Infra up. Starting API and Web..."
	@$(MAKE) -j2 dev-api dev-web

## dev-infra: Start only infrastructure services (postgres, redis, minio, nats)
dev-infra:
	@echo ">>> Starting infrastructure services..."
	docker compose up -d postgres redis minio nats
	@echo ">>> Waiting for services to be healthy..."
	@docker compose run --rm minio-init 2>/dev/null || true
	@echo ">>> Infrastructure ready"

## dev-api: Start Go API with hot reload (requires air)
dev-api: check-tools
	@echo ">>> Starting API on :8080..."
	cd $(API_DIR) && air -c $(AIR_CONFIG)

## dev-web: Start Next.js dev server
dev-web: check-tools
	@echo ">>> Starting Web on :3000..."
	cd $(WEB_DIR) && pnpm dev

## stop: Stop all Docker services
stop:
	docker compose down

# ── Test ──────────────────────────────────────────────────────────────────────
## test: Run all tests (API + Web)
test: test-api test-web

## test-api: Run Go tests with race detector and coverage
test-api:
	@echo ">>> Running Go tests..."
	cd $(API_DIR) && go test -race -coverprofile=coverage.out ./...
	cd $(API_DIR) && go tool cover -func=coverage.out | tail -1

## test-web: Run Next.js / Vitest tests
test-web:
	@echo ">>> Running Web tests..."
	cd $(WEB_DIR) && pnpm test

# ── Lint ──────────────────────────────────────────────────────────────────────
## lint: Run all linters
lint: lint-api lint-web

## lint-api: Run golangci-lint
lint-api:
	@echo ">>> Linting Go code..."
	cd $(API_DIR) && golangci-lint run ./...

## lint-web: Run ESLint
lint-web:
	@echo ">>> Linting Web code..."
	cd $(WEB_DIR) && pnpm lint

# ── Build ─────────────────────────────────────────────────────────────────────
## build: Build all apps
build: build-api build-web

## build-api: Build Go binary
build-api:
	@echo ">>> Building API binary..."
	cd $(API_DIR) && CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/server ./cmd/server

## build-web: Build Next.js production bundle
build-web:
	@echo ">>> Building Web..."
	cd $(WEB_DIR) && pnpm build

# ── Database / Migrations ────────────────────────────────────────────────────
## migrate-lint: Validate migrations (no duplicates, no gaps, pairs matched)
.PHONY: migrate-lint
migrate-lint:
	@$(BASH) scripts/migration-lint.sh

## migrate-up: Apply all pending migrations (lints first)
migrate-up: migrate-lint
	@echo ">>> Running migrations UP..."
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" up

## migrate-down: Roll back the last migration
migrate-down:
	@echo ">>> Rolling back last migration..."
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" down 1

## migrate-create NAME=xxx: Scaffold new migration with next version number
.PHONY: migrate-create
migrate-create:
ifndef NAME
	$(error NAME is required. Usage: make migrate-create NAME=add_users_status)
endif
	@$(BASH) scripts/migration-create.sh "$(NAME)"

## migrate-status: Show current migration version
migrate-status:
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" version

# ── Code Generation ───────────────────────────────────────────────────────────
## sqlc: Generate type-safe Go code from SQL queries
sqlc:
	@echo ">>> Running sqlc generate..."
	cd $(API_DIR) && sqlc generate

## generate: Run all code generators (sqlc + go generate)
generate: sqlc
	cd $(API_DIR) && go generate ./...

# ── Utilities ─────────────────────────────────────────────────────────────────
## clean: Remove build artifacts
clean:
	cd $(API_DIR) && rm -rf bin/ tmp/ coverage.out
	cd $(WEB_DIR) && rm -rf .next/ out/
	@echo ">>> Cleaned"

## install: Install all dependencies
install:
	pnpm install
	cd $(API_DIR) && go mod download

## help: Show this help message
help:
	@echo "Available targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /'

.PHONY: check-tools
check-tools:
	@which air > /dev/null || (echo "ERROR: 'air' not installed. Run: go install github.com/air-verse/air@latest" && exit 1)
	@which pnpm > /dev/null || (echo "ERROR: 'pnpm' not installed. Install from https://pnpm.io" && exit 1)
	@test -d apps/web/node_modules || (echo "ERROR: node_modules missing. Run: pnpm install" && exit 1)
