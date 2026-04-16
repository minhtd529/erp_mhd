<!-- spec-version: 1.3 | last-sync: 2026-04-16 | changes: added address, bank account, representative, client_contacts fields -->
> **Spec version**: 1.3 — Last sync: 2026-04-16 — Updated in v1.3

# Module 1: CRM – Customer Relationship Management & Commission

## Overview
Manages client information, risk assessments, conflict checks, sales owner tracking, and commission management for audit and advisory engagements.

## Bounded Context: CRM

### Responsibilities
- Client master data (registration, contacts, classification)
- Risk assessment & scoring
- Conflict of interest detection
- Client acceptance workflow (DRAFT → APPROVED → ACCEPTED)
- Sales owner tracking (sales_owner_id, referrer_id) at client and engagement level
- Commission management: plan definition, per-engagement commission assignment, accrual, approval, payment, clawback

## Key Features

### 1. Client Management
**Entity**: Client aggregate
- Tax code (10 or 13-14 digits, unique)
- Business name, English name
- Industry classification, business type
- Address (full address text)
- Primary/secondary contact persons (table `client_contacts`)
- Legal representative: name, title, phone
- Bank account: bank name, account number, account holder name
- Office assignment (multi-office support)
- Status: `PROSPECT` → `ASSESSMENT` → `ACCEPTED` → `INACTIVE`

### 2. Risk Assessment
**Entity**: ClientRiskAssessment
- Risk level: `LOW`, `MEDIUM`, `HIGH`, `CRITICAL` (enum)
- Scoring: Criteria-based (0-100)
- Assessment date, assessor (user_id)
- Required before client acceptance

### 3. Conflict of Interest Checking
**Entity**: ConflictCheck
- Automatic check on client creation
- Related party analysis (tax code, business name)
- Conflict status: `CLEAR`, `PENDING_REVIEW`, `CONFLICT_FOUND`, `RESOLVED`
- Issued at creation, results immutable after Partner approval

### 4. Client Acceptance Workflow
**Pattern**: Domain event-driven state transition
- State progression: DRAFT → RISK_ASSESSED → CONFLICT_CHECKED → READY_FOR_APPROVAL → APPROVED → ACCEPTED
- Multi-step approval (AUDIT_MANAGER → FIRM_PARTNER)
- Notifications via outbox pattern + WebSocket broadcast on `channel=crm`

### 5. Distributed Locking
No concurrent critical operations; conflict checking serialized by tax code index.

## Code Structure

### Go Package Layout
```
modules/crm/
  ├── domain/
  │   ├── client.go                (Client aggregate root)
  │   ├── client_contact.go        (Contact value object)
  │   ├── risk_assessment.go       (RiskAssessment aggregate)
  │   ├── conflict_check.go        (ConflictCheck aggregate)
  │   └── client_events.go         (ClientCreated, ClientApproved events)
  ├── application/
  │   ├── client_service.go        (ClientService - use cases)
  │   ├── risk_assessment_service.go
  │   └── conflict_service.go
  ├── infrastructure/
  │   └── postgres/
  │       ├── client_repository.go (queries + writes via GORM)
  │       └── conflict_check_repository.go
  └── interfaces/
      └── rest/
          ├── client_handler.go            (ClientHandler)
          ├── risk_assessment_handler.go   (RiskAssessmentHandler)
          └── conflict_check_handler.go    (ConflictCheckHandler)
```

## API Endpoints

**Authorization**: FARM_PARTNER, AUDIT_MANAGER, AUDIT_STAFF (read-only for staff)

### Client CRUD
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| GET | `/api/v1/clients` | List clients (paginated) | AUDIT_MANAGER | No |
| POST | `/api/v1/clients` | Create client (status=DRAFT) | FIRM_PARTNER | CREATE |
| GET | `/api/v1/clients/{id}` | Get client detail | AUDIT_MANAGER | No |
| PUT | `/api/v1/clients/{id}` | Update client | FIRM_PARTNER | UPDATE |
| DELETE | `/api/v1/clients/{id}` | Soft delete (is_deleted=true) | FIRM_PARTNER | DELETE |

**Fields** (snake_case): `id`, `tax_code`, `business_name`, `english_name`, `industry`, `status`, `office_id`, `sales_owner_id`, `referrer_id`, `assigned_partner_id`, `address`, `bank_name`, `bank_account_number`, `bank_account_name`, `representative_name`, `representative_title`, `representative_phone`, `created_at`, `created_by`, `updated_at`, `updated_by`, `is_deleted`

### Client Contacts (Người liên hệ đầu mối)
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| GET | `/api/v1/clients/{id}/contacts` | List contacts (primary first) | AUDIT_MANAGER | No |
| POST | `/api/v1/clients/{id}/contacts` | Add contact | FIRM_PARTNER | CREATE |
| PUT | `/api/v1/clients/{id}/contacts/{cid}` | Update contact | FIRM_PARTNER | UPDATE |
| DELETE | `/api/v1/clients/{id}/contacts/{cid}` | Soft delete contact | FIRM_PARTNER | DELETE |

**Fields**: `id`, `client_id`, `full_name`, `title`, `phone`, `email`, `is_primary`, `created_at`, `created_by`, `updated_at`, `updated_by`

### Risk Assessment
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| POST | `/api/v1/clients/{client_id}/risk-assessment` | Create assessment | AUDIT_MANAGER | CREATE |
| GET | `/api/v1/clients/{client_id}/risk-assessment` | Get latest assessment | AUDIT_MANAGER | No |
| POST | `/api/v1/clients/{client_id}/risk-assessment/submit` | Submit for approval | AUDIT_MANAGER | STATE_TRANSITION |
| POST | `/api/v1/clients/{client_id}/risk-assessment/approve` | Manager approval | FIRM_PARTNER | APPROVE |

**Fields**: `id`, `client_id`, `risk_level` (enum), `score`, `criteria_scores` (JSONB), `status`, `assessor_id`, `created_at`

### Conflict of Interest Check
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| GET | `/api/v1/clients/{client_id}/conflict-check` | Get conflict check result | AUDIT_MANAGER | No |
| POST | `/api/v1/clients/{client_id}/conflict-check/approve` | Approve conflict findings | FIRM_PARTNER | APPROVE |

**Fields**: `id`, `client_id`, `status` (enum), `related_parties` (JSONB), `findings` (JSONB text), `checked_at`, `checked_by`

### Client Acceptance
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| POST | `/api/v1/clients/{id}/accept` | Accept client (state transition) | FIRM_PARTNER | STATE_TRANSITION |
| GET | `/api/v1/clients/{id}/status` | Get acceptance workflow status | AUDIT_MANAGER | No |

## Database Tables

### Core Tables
- `clients` (id UUID, tax_code VARCHAR(14) unique, business_name VARCHAR(200), english_name VARCHAR(200), industry VARCHAR(100), status ENUM, office_id UUID, sales_owner_id UUID REFERENCES users(id), referrer_id UUID REFERENCES users(id), assigned_partner_id UUID REFERENCES users(id), address VARCHAR(500), bank_name VARCHAR(100), bank_account_number VARCHAR(50), bank_account_name VARCHAR(200), representative_name VARCHAR(200), representative_title VARCHAR(100), representative_phone VARCHAR(20), created_at, created_by, updated_at, updated_by, is_deleted)
- `client_contacts` (id UUID, client_id UUID REFERENCES clients(id), full_name VARCHAR(200), title VARCHAR(100), phone VARCHAR(20), email VARCHAR(255), is_primary BOOLEAN DEFAULT false, is_deleted BOOLEAN DEFAULT false, created_at, created_by, updated_at, updated_by)
- `risk_assessments` (id, client_id, risk_level ENUM, score INT, criteria_scores JSONB, status ENUM, assessor_id, created_at)
- `conflict_checks` (id, client_id, status ENUM, related_parties JSONB, findings TEXT, checked_by, checked_at, approved_by, approved_at)

### Commission Tables (Phase 3)
- `commission_plans` (id UUID, code VARCHAR(50) UNIQUE, name, type ENUM (flat|tiered|fixed|custom), rate NUMERIC(5,4), fixed_amount BIGINT, tiers JSONB, apply_base VARCHAR(20), trigger_on VARCHAR(30), is_active BOOLEAN, created_at, created_by)
- `engagement_commissions` (id UUID, engagement_id UUID REFERENCES engagements(id), salesperson_id UUID REFERENCES employees(id), role VARCHAR(30) — primary|referrer|account_manager|technical_lead, plan_id UUID REFERENCES commission_plans(id), rate_type, rate, fixed_amount, tiers JSONB, apply_base, trigger_on, status ENUM (active|cancelled), created_at, created_by)
- `commission_records` (id UUID, engagement_commission_id UUID REFERENCES engagement_commissions(id), engagement_id, salesperson_id, invoice_id, payment_id, gross_amount BIGINT, net_amount BIGINT, status ENUM (accrued|approved|on_hold|paid|clawback|cancelled), clawback_record_id UUID REFERENCES commission_records(id), accrued_at, approved_at, paid_at, created_at)

### Indexes
- `uidx_clients_tax_code` on (tax_code) where is_deleted=false
- `idx_clients_office_id` on (office_id)
- `idx_clients_sales_owner_id` on (sales_owner_id)
- `idx_risk_assessments_client_id_created_at` on (client_id, created_at DESC)
- `idx_conflict_checks_client_id` on (client_id)
- `idx_eng_commissions_engagement` on (engagement_id, status)
- `idx_eng_commissions_salesperson` on (salesperson_id, status)
- `idx_commission_records_salesperson` on (salesperson_id, status, accrued_at)
- `idx_commission_records_engagement` on (engagement_id)
- `idx_commission_records_pending_payout` on (status) WHERE status = 'approved' AND paid_at IS NULL
- `idx_commission_plans_active` on (is_active, type)

## Commission API Endpoints (Phase 3)

**Authorization**: Commission Plans — FIRM_PARTNER/AUDIT_MANAGER (CRUD); EngagementCommission — AUDIT_MANAGER+; CommissionRecord approval — FIRM_PARTNER+

### Commission Plans
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| GET | `/api/v1/commission-plans` | List plans | AUDIT_MANAGER | No |
| POST | `/api/v1/commission-plans` | Create plan | FIRM_PARTNER | CREATE |
| GET | `/api/v1/commission-plans/{id}` | Get plan detail | AUDIT_MANAGER | No |
| PUT | `/api/v1/commission-plans/{id}` | Update plan | FIRM_PARTNER | UPDATE |
| DELETE | `/api/v1/commission-plans/{id}` | Deactivate plan | FIRM_PARTNER | DELETE |

### Engagement Commission Assignment
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| GET | `/api/v1/engagements/{id}/commissions` | List engagement commissions | AUDIT_MANAGER | No |
| POST | `/api/v1/engagements/{id}/commissions` | Assign commission to salesperson | AUDIT_MANAGER | CREATE |
| PUT | `/api/v1/engagements/{id}/commissions/{ec_id}` | Update commission config | FIRM_PARTNER | UPDATE |
| DELETE | `/api/v1/engagements/{id}/commissions/{ec_id}` | Cancel commission | FIRM_PARTNER | DELETE |

### Commission Records
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| GET | `/api/v1/commissions/records` | List records (filterable) | AUDIT_MANAGER | No |
| POST | `/api/v1/commissions/records/{id}/approve` | Approve record for payout | FIRM_PARTNER | APPROVE |
| POST | `/api/v1/commissions/records/{id}/mark-paid` | Mark as paid | FIRM_PARTNER | UPDATE |
| POST | `/api/v1/commissions/records/{id}/clawback` | Manual clawback | FIRM_PARTNER | CREATE |

### My Commissions
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| GET | `/api/v1/me/commissions` | My commission records | AUDIT_STAFF | No |
| GET | `/api/v1/me/commissions/summary` | My commission summary (YTD, pending, on_hold) | AUDIT_STAFF | No |

## CQRS
**Writes**: GORM for client/risk/conflict mutations; commission mutations via GORM
**Reads**: sqlc for client list (high-cardinality reads); sqlc for commission queries + salesperson earnings
**Events**: ClientCreated, RiskAssessmentCreated, ConflictCheckCompleted, ClientAccepted (→ outbox_messages)

## Error Codes
`CLIENT_NOT_FOUND`
`DUPLICATE_TAX_CODE`
`CONTACT_NOT_FOUND`
`RISK_ASSESSMENT_REQUIRED`
`CONFLICT_FOUND_CANNOT_ACCEPT`
`INVALID_STATE_TRANSITION`
`COMMISSION_PLAN_NOT_FOUND`
`COMMISSION_RATE_EXCEEDS_100` - Tổng rate trên 1 engagement vượt 100%
`ENGAGEMENT_COMMISSION_NOT_FOUND`
`COMMISSION_RECORD_IMMUTABLE` - Đã approved, không thể sửa; tạo clawback record thay thế
`DUPLICATE_COMMISSION_TRIGGER` - Invoice/payment đã trigger commission (idempotency violation)