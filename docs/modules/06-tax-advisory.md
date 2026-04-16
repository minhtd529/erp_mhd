<!-- spec-version: 1.2 | last-sync: 2026-04-16 | changes: none (no changes in v1.2) -->
> **Spec version**: 1.2 — Last sync: 2026-04-16 — No changes in v1.2

# Module 6: TaxAdvisory - Tax Compliance & Advisory Services

## Overview
Manages tax deadlines, advisory engagements, and tax-related deliverables for clients.

## Bounded Context: TaxAdvisory

### Responsibilities
- Tax deadline tracking (client fiscal year + regulatory deadlines)
- Advisory service engagement management
- Tax compliance monitoring & alerts
- Advisory recommendation documentation
- Deadline reminder scheduling (via Asynq)

## Key Features

### 1. Tax Deadline Tracking
**Entity**: TaxDeadline aggregate
- Client fiscal year configuration (auto-calculate deadlines)
- Deadline types: `VAT_FILING`, `CORPORATE_TAX`, `PERSONAL_TAX`, `COMPLIANCE_REPORTING`, `CUSTOM`
- Status: `NOT_DUE`, `DUE_SOON` (< 7 days), `OVERDUE`, `COMPLETED`
- Submission tracking: Expected date, actual submission date, submission status
- Reminder scheduling: Auto-trigger via Asynq 7 days before

### 2. Advisory Records
**Entity**: AdvisoryRecord aggregate
- Advisory type: `TAX_CONSULTATION`, `BUSINESS_ADVISORY`, `COMPLIANCE_REVIEW`
- Engagement link (optional, if part of paid engagement)
- Recommendation & findings (TEXT)
- Deliverables: File references to MinIO
- Status: `DRAFTED`, `DELIVERED`, `ACTED_ON`
- Response tracking: Client action feedback

### 3. Compliance Monitoring
**Computed View**: Compliance dashboard
- Tax deadline calendar (client-scoped)
- Overdue alerts
- Compliance score per client
- Regulatory requirement tracking

### 4. Distributed Locking
No critical concurrent operations; deadline deadlines serialized by deadline_id index.

## Code Structure

### Go Package Layout
```
modules/tax_advisory/
  ├── domain/
  │   ├── tax_deadline.go           (TaxDeadline aggregate)
  │   ├── advisory_record.go        (AdvisoryRecord aggregate)
  │   ├── compliance_rule.go        (ComplianceRule aggregate)
  │   └── tax_advisory_events.go    (TaxDeadlineApproaching, AdvisoryDelivered, etc.)
  ├── application/
  │   ├── tax_deadline_service.go   (TaxDeadlineService)
  │   ├── advisory_service.go       (AdvisoryService)
  │   ├── compliance_service.go     (ComplianceService)
  │   └── reminder_service.go       (ReminderService - Asynq scheduled jobs)
  ├── infrastructure/
  │   ├── postgres/
  │   │   ├── tax_deadline_repository.go (sqlc for reads)
  │   │   └── advisory_repository.go
  │   └── asynq/
  │       └── reminder_job.go       (Scheduled job: 7-day pre-deadline)
  └── interfaces/
      └── rest/
          ├── tax_deadline_handler.go    (TaxDeadlineHandler)
          ├── advisory_handler.go        (AdvisoryHandler)
          └── compliance_handler.go      (ComplianceHandler)
```

## API Endpoints

**Authorization**: FIRM_PARTNER, AUDIT_MANAGER, AUDIT_STAFF (read-only)

### Tax Deadlines
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| GET | `/api/v1/clients/{client_id}/tax-deadlines` | List deadlines (paginated) | AUDIT_STAFF | No |
| POST | `/api/v1/clients/{client_id}/tax-deadlines` | Create deadline (manual) | AUDIT_MANAGER | CREATE |
| GET | `/api/v1/clients/{client_id}/tax-deadlines/{id}` | Get deadline detail | AUDIT_STAFF | No |
| PUT | `/api/v1/clients/{client_id}/tax-deadlines/{id}` | Update deadline | AUDIT_MANAGER | UPDATE |
| POST | `/api/v1/clients/{client_id}/tax-deadlines/{id}/mark-completed` | Mark submitted | AUDIT_STAFF | STATE_TRANSITION |
| POST | `/api/v1/clients/{client_id}/tax-deadlines/auto-generate` | Generate from fiscal year | FIRM_PARTNER | CREATE |

**Fields** (snake_case): `id`, `client_id`, `deadline_type`, `deadline_name`, `due_date`, `status`, `expected_submission_date`, `actual_submission_date`, `submission_status`, `created_at`, `updated_at`

### Advisory Records
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| GET | `/api/v1/clients/{client_id}/advisory-records` | List advisory records | AUDIT_STAFF | No |
| POST | `/api/v1/clients/{client_id}/advisory-records` | Create advisory record | AUDIT_MANAGER | CREATE |
| GET | `/api/v1/advisory-records/{id}` | Get advisory detail | AUDIT_STAFF | No |
| PUT | `/api/v1/advisory-records/{id}` | Update advisory (DRAFTED) | AUDIT_MANAGER | UPDATE |
| POST | `/api/v1/advisory-records/{id}/deliver` | Mark delivered | AUDIT_MANAGER | STATE_TRANSITION |
| POST | `/api/v1/advisory-records/{id}/attach-file` | Attach deliverable file | AUDIT_MANAGER | CREATE |

**Fields** (snake_case): `id`, `client_id`, `engagement_id` (optional), `advisory_type`, `recommendation` (TEXT), `status`, `delivered_date`, `created_at`, `created_by`, `updated_at`

### Compliance Dashboard
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| GET | `/api/v1/clients/{client_id}/tax/compliance-status` | Get client compliance score | AUDIT_STAFF | No |
| GET | `/api/v1/tax/dashboard` | Tax deadline calendar (all clients) | FIRM_PARTNER | No |
| GET | `/api/v1/tax/overdue-alerts` | List overdue deadlines | FIRM_PARTNER | No |

## Database Tables

### Core Tables
- `tax_deadlines` (id UUID, client_id, deadline_type ENUM, deadline_name, due_date DATE, status ENUM, expected_submission_date, actual_submission_date, submission_status ENUM, created_at, updated_at)
- `advisory_records` (id UUID, client_id, engagement_id (nullable), advisory_type ENUM, recommendation TEXT, status ENUM, delivered_date, created_at, created_by, updated_at)
- `advisory_files` (id UUID, advisory_id, file_id UUID (refs file_metadata), created_at)
- `tax_compliance_rules` (id UUID, rule_type, description, check_query TEXT, severity ENUM (LOW|MEDIUM|HIGH), is_active, created_at)
- `outbox_messages` (for TaxDeadlineApproaching, AdvisoryDelivered, etc.)

### Materialized Views
- `mv_tax_compliance_status` (client_id, total_deadlines, completed, overdue, compliance_score) - refreshed daily

### Indexes
- `idx_tax_deadlines_client_id_due_date` on (client_id, due_date)
- `idx_tax_deadlines_status` on (status) where status IN ('NOT_DUE', 'DUE_SOON')
- `idx_advisory_records_client_id` on (client_id)
- `idx_advisory_files_advisory_id` on (advisory_id)

## CQRS
**Writes**: GORM for deadline/advisory mutations
**Reads**: sqlc for deadline list, materialized view for compliance dashboard
**Events**: TaxDeadlineCreated, TaxDeadlineApproaching (triggered by Asynq reminder job), DeadlineCompleted

## Asynq Job Scheduling
**Job**: `tax:deadline-reminder`
- Scheduled daily at 08:00 AM
- Finds deadlines with due_date = today + 7 days
- Updates status → DUE_SOON
- Publishes TaxDeadlineApproaching event (triggers WebSocket notification)

## Error Codes
`CLIENT_NOT_FOUND`
`TAX_DEADLINE_NOT_FOUND`
`ADVISORY_RECORD_NOT_FOUND`
`INVALID_STATE_TRANSITION` - Cannot mark completed if already overdue
`FISCAL_YEAR_NOT_CONFIGURED` - Cannot auto-generate without fiscal year