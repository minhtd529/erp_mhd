# ERP Audit System — Architectural Decisions

## Overview
This document records key architectural decisions made during the development of the ERP Audit System.

## Technology Stack Decisions

### Backend Framework
**Decision**: Golang with Gin/Chi framework
**Rationale**: High performance, type safety, concurrency native, compile-time checks
**Alternatives Considered**: Node.js (Express), Python (FastAPI), Java (Spring)
**Date**: 2026-04-12

### Frontend Framework
**Decision**: Next.js 14 with App Router
**Rationale**: SSR/SSG, SEO, file-based routing, middleware support
**Alternatives Considered**: React SPA, Vue.js, Angular
**Date**: 2026-04-12

### Database
**Decision**: PostgreSQL 16
**Rationale**: ACID compliance, JSON support, full-text search, partitioning
**Alternatives Considered**: MySQL, MongoDB, SQL Server
**Date**: 2026-04-12

### Authentication
**Decision**: JWT with RBAC + ABAC
**Rationale**: Stateless, scalable, fine-grained permissions
**Alternatives Considered**: Session-based auth, OAuth2 only
**Date**: 2026-04-12

### 2FA Implementation
**Decision**: Self-hosted TOTP + Push-based 2FA
**Rationale**: No dependency on third-party services, full control
**Alternatives Considered**: Twilio/Authy API, Firebase Auth
**Date**: 2026-04-12

### Push Notifications & Real-time
**Decision**: WebSocket for real-time updates, self-hosted Web Push relay
**Rationale**: No third-party dependency, real-time delivery, full control
**Channels**: Per-domain subscriptions (engagement, timesheet, billing, etc.)
**Alternatives Considered**: Firebase Cloud Messaging, Pusher, Socket.io
**Date**: 2026-04-12

### File Storage
**Decision**: MinIO for object storage (audit file tracking)
**Rationale**: S3-compatible, self-hosted, audit trail support
**Compliance**: All file operations logged in audit trail
**Date**: 2026-04-12

## Architecture Decisions

### Clean Architecture
**Decision**: Hexagonal Architecture (domain → application → infrastructure → interfaces)
**Rationale**: Testable, maintainable, technology-agnostic, domain-first approach
**Date**: 2026-04-12

### Domain-Driven Design (DDD)
**Decision**: 9 Bounded Contexts (Global, CRM, Engagement, Timesheet, Billing, WorkingPapers, TaxAdvisory, HRM, Reporting)
**Rationale**: Clear domain boundaries, independent scaling, service decomposition
**Date**: 2026-04-12

### CQRS Pattern
**Decision**: Separate read and write models: GORM for writes, sqlc/pgx for reads
**Rationale**: Optimized queries, better performance, scalable read replicas
**Date**: 2026-04-12

### Event-Driven Architecture
**Decision**: Domain Events + Outbox Pattern (outbox_messages table → Asynq worker)
**Rationale**: Eventual consistency, reliable event distribution, async processing
**Alternatives Considered**: Kafka, RabbitMQ (overkill for current scale)
**Date**: 2026-04-12

### Async Task Processing
**Decision**: Asynq (Redis-backed) for background jobs
**Rationale**: Built on Redis, simple setup, reliable retry, monorepo-friendly
**Alternatives Considered**: Bull, Celery, NATS
**Date**: 2026-04-12

### Distributed Locking
**Decision**: Redis Distributed Locks for concurrent operations
**Rationale**: Prevent race conditions on critical operations (edit engagement, approve timesheet, process payment)
**Locked Operations**: Edit engagement, assign resources, approve timesheet, sign-off working paper, process payment
**Date**: 2026-04-12

### Monorepo Structure
**Decision**: Single repository with apps/ (api, web) and packages/ (shared code)
**Rationale**: Atomic changes, shared code management, easier refactoring
**Date**: 2026-04-12

### API Design
**Decision**: RESTful JSON with OpenAPI 3.0, kebab-case paths, snake_case fields
**Additional Rules**: UUID IDs, ISO 8601 timestamps, UPPER_SNAKE_CASE enums, soft deletes
**Versioning**: URL-based (/api/v1), only bump for breaking changes
**Date**: 2026-04-12

### State Management (Frontend)
**Decision**: Zustand + React Query for client and server state
**Rationale**: Simple, performant, clear separation of concerns
**Alternatives Considered**: Redux, Context API
**Date**: 2026-04-12

## Security Decisions

### Password Hashing
**Decision**: bcrypt with cost=12
**Rationale**: Industry standard, configurable work factor
**Date**: 2026-04-12

### Data Encryption
**Decision**: AES-256-GCM for sensitive files
**Rationale**: Strong encryption, authenticated encryption
**Date**: 2026-04-12

### Rate Limiting
**Decision**: 100 req/min per user, 1000 req/min per IP
**Rationale**: DDoS protection, fair usage
**Date**: 2026-04-12

## Deployment Decisions

### Containerization
**Decision**: Docker + Docker Compose for dev, Kubernetes for prod
**Rationale**: Consistency across environments, scalability
**Date**: 2026-04-12

### CI/CD
**Decision**: GitHub Actions
**Rationale**: Integrated with GitHub, cost-effective
**Alternatives Considered**: Jenkins, GitLab CI
**Date**: 2026-04-12

### Monitoring
**Decision**: Prometheus + Grafana + Loki
**Rationale**: CNCF ecosystem, comprehensive monitoring
**Date**: 2026-04-12

## Data & Query Decisions

### Audit Trail
**Decision**: Immutable audit logs via pkg/audit (every mutation logged)
**Rationale**: Compliance, debugging, user accountability
**Fields**: created_at, created_by, updated_at, updated_by (all tables)
**Date**: 2026-04-12

### Snapshot Data
**Decision**: Store JSONB snapshots, immutable after confirmation
**Rationale**: Historical accuracy, avoid data drift
**Example**: Working paper content, billing invoice snapshot
**Date**: 2026-04-12

### Full-text Search
**Decision**: PostgreSQL pg_trgm extension with GIN index
**Rationale**: No external dependency, sufficient for audit data volume
**Date**: 2026-04-12

### Database Soft Deletes
**Decision**: Use is_deleted/is_void/is_active flags (no hard deletes)
**Rationale**: Audit compliance, data recovery, referential integrity
**Date**: 2026-04-12

## Design & UI Decisions

### UI Theme
**Decision**: "Professional Audit" aesthetic (dark navy, gold accents, clean typography)
**Rationale**: Trust, authority, clarity for professional users
**Color Primary**: #1F3A70 (deep navy)
**Color Secondary**: #D4A574 (gold)
**Typography**: Inter (headings & body), Fira Code (monospace)
**Date**: 2026-04-12

### API Naming Conventions
**Decision**: kebab-case paths, snake_case fields, UPPER_SNAKE_CASE enums
**Rationale**: Consistency, RESTful standard, JSON convention
**Date**: 2026-04-12

### Database Naming Conventions
**Decision**: snake_case tables/columns, {} plural nouns, UUID primary keys
**Rationale**: SQL standard, consistency, universally unique identifiers
**Date**: 2026-04-12

### Go Code Naming
**Decision**: {Entity}Handler, {Domain}Service, {Entity}Repository patterns
**Rationale**: Clear responsibility, domain-driven structure
**Date**: 2026-04-12

## Future Decisions
- [ ] Search Engine: Elasticsearch vs PostgreSQL full-text (scale to 1M+ engagements)
- [ ] Distributed Tracing: Jaeger vs Datadog
- [ ] Message Queue: Switch from Asynq to NATS if scale requires
- [ ] Mobile App: React Native vs Flutter (post Phase 3)