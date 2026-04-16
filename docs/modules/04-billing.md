<!-- spec-version: 1.2 | last-sync: 2026-04-16 | changes: added Commission Integration (Event-Driven) section -->
> **Spec version**: 1.2 — Last sync: 2026-04-16 — Updated in v1.2

# Module 4: Billing - Invoicing & Payment Processing

## Overview
Handles invoice generation from locked timesheets, payment recording, and accounts receivable management.

## Bounded Context: Billing

### Responsibilities
- Invoice generation (from engagement fees, time & material, milestones)
- Invoice state management (DRAFT → ISSUED → PAID → SETTLED)
- Payment recording & reconciliation
- Billing memo/credit note management
- Accounts receivable aging

## Key Features

### 1. Invoice Generation
**Entity**: Invoice aggregate
- Invoice number (auto-generated, sequential)
- Issue date, due date (net 30 days default)
- Client_id, engagement_id, invoice_type (enum)
- Status: `DRAFT`, `SENT`, `CONFIRMED`, `ISSUED`, `PAID`, `CANCELLED`
- Currency: VND only
- JSONB snapshot of billing data (immutable after ISSUED)

**Invoice Types**:
- `TIME_AND_MATERIAL`: From locked timesheets + direct costs
- `FIXED_FEE`: Milestone-based or full engagement
- `RETAINER`: Monthly/periodic billing
- `CREDIT_NOTE`: Links to original invoice_id

### 2. Invoice Line Items
**Entity**: InvoiceLineItem aggregate
- Description, quantity, unit_price, tax_amount, total_amount
- Source (engagement fee, timesheet hours, direct cost)
- Snapshot JSONB: rate_card snapshot, hourly rates, cost amounts

### 3. Payment Processing
**Entity**: Payment aggregate
- Payment method: `BANK_TRANSFER`, `CHEQUE`, `CASH`, `CREDIT_CARD`
- Amount, payment_date, reference_number
- Status: `RECORDED`, `CLEARED`, `DISPUTED`, `REVERSED`
- Audit trail for all payment changes

### 4. Accounts Receivable
**Computed**: Aging analysis via query
- Current, 1-30, 31-60, 61-90, 90+ days overdue
- Outstanding balance per client
- Payment reminders via Asynq (scheduled jobs)

### 5. Commission Integration (Event-Driven)

Billing Service publishes events vào NATS. Commission Service subscribes và trigger accrual/clawback:

| Event | Publish khi | Commission Handler |
|---|---|---|
| `invoice.issued` | Invoice chuyển `status = ISSUED` | `AccrueOnInvoiceIssued(invoiceID)` |
| `payment.received` | Ghi nhận payment mới | `AccrueOnPaymentReceived(paymentID)` |
| `invoice.cancelled` | Invoice bị cancel | `AutoClawbackOnInvoiceCancel(invoiceID)` |
| `credit_note.issued` | Credit note được tạo | `AutoClawbackOnCreditNote(creditNoteID)` |
| `engagement.settled` | Engagement thanh lý | `ReleaseHoldback(engagementID)` |

**Pattern**: Outbox pattern — event được ghi vào `outbox_messages` trong cùng transaction với invoice/payment mutation, đảm bảo consistency giữa Billing và Commission.

> Implemented trong Phase 3 cùng với Commission Module.

### 6. Distributed Locking
**Critical Operations** (require Redis locks):
- Process payment (prevent duplicate recording)
- Issue invoice (capture client data snapshot)
- Generate credit note (validate original invoice)

**Lock Key Pattern**: `invoice:{id}:issue`, `payment:{id}:process`, `invoice:{id}:credit_note`

> Section 5 above (Commission Integration) added in SPEC v1.2.

## Code Structure

### Go Package Layout
```
modules/billing/
  ├── domain/
  │   ├── invoice.go                (Invoice aggregate root)
  │   ├── invoice_line_item.go      (LineItem value object)
  │   ├── payment.go                (Payment aggregate)
  │   ├── billing_memo.go           (Memo/credit note aggregate)
  │   └── billing_events.go         (InvoiceIssued, PaymentRecorded, etc.)
  ├── application/
  │   ├── invoice_service.go        (InvoiceService - use cases)
  │   ├── invoice_generation_service.go (Calculates invoice from engagement data)
  │   ├── payment_service.go        (PaymentService with distributed lock)
  │   └── ar_service.go             (Accounts Receivable analytics)
  ├── infrastructure/
  │   ├── postgres/
  │   │   ├── invoice_repository.go (CQRS: sqlc for list, GORM for mutations)
  │   │   ├── payment_repository.go
  │   │   └── ar_repository.go     (Materialized view mv_ar_aging)
  │   └── redis/
  │       └── distributed_lock.go
  └── interfaces/
      └── rest/
          ├── invoice_handler.go    (InvoiceHandler)
          ├── payment_handler.go    (PaymentHandler)
          └── ar_handler.go         (ARHandler)
```

## API Endpoints

**Authorization**: FIRM_PARTNER, AUDIT_MANAGER (AUDIT_STAFF read-only)

### Invoice CRUD
| Method | Path | Description | Auth | Audit | Lock |
|--------|------|-------------|------|-------|------|
| GET | `/api/v1/invoices` | List invoices (paginated) | AUDIT_STAFF | No | No |
| POST | `/api/v1/invoices/generate-from-engagement` | Create T&M invoice from locked timesheet | AUDIT_MANAGER | CREATE | invoice:{id}:issue |
| GET | `/api/v1/invoices/{id}` | Get invoice details | AUDIT_STAFF | No | No |
| PUT | `/api/v1/invoices/{id}` | Update draft invoice | FIRM_PARTNER | UPDATE | No |
| POST | `/api/v1/invoices/{id}/send` | Send to client (status→SENT) | FIRM_PARTNER | STATE_TRANSITION | No |
| POST | `/api/v1/invoices/{id}/confirm` | Client confirms (status→CONFIRMED) | CLIENT_ADMIN | STATE_TRANSITION | invoice:{id}:issue |
| POST | `/api/v1/invoices/{id}/issue` | Official issue (status→ISSUED, snapshot) | FIRM_PARTNER | STATE_TRANSITION | invoice:{id}:issue |
| DELETE | `/api/v1/invoices/{id}` | Soft delete (draft only) | FIRM_PARTNER | DELETE | No |

**Fields** (snake_case): `id`, `invoice_number`, `client_id`, `engagement_id`, `issue_date`, `due_date`, `status`, `total_amount` (DECIMAL), `tax_amount`, `snapshot_data` (JSONB), `created_at`, `created_by`, `updated_at`

### Line Items
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| GET | `/api/v1/invoices/{id}/line-items` | List items | AUDIT_STAFF | No |
| POST | `/api/v1/invoices/{id}/line-items` | Add item (draft only) | FIRM_PARTNER | CREATE |
| DELETE | `/api/v1/invoices/{id}/line-items/{item_id}` | Remove item (draft only) | FIRM_PARTNER | DELETE |

### Payments
| Method | Path | Description | Auth | Audit | Lock |
|--------|------|-------------|------|-------|------|
| GET | `/api/v1/invoices/{id}/payments` | List invoice payments | AUDIT_STAFF | No | No |
| POST | `/api/v1/invoices/{id}/record-payment` | Record payment received | FIRM_PARTNER | CREATE | payment:{id}:process |
| PUT | `/api/v1/payments/{id}` | Update payment (RECORDED only) | FIRM_PARTNER | UPDATE | payment:{id}:process |
| DELETE | `/api/v1/payments/{id}` | Reverse payment | FIRM_PARTNER | DELETE | payment:{id}:process |

### Credit Notes
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| POST | `/api/v1/invoices/{id}/credit-note` | Create credit note | FIRM_PARTNER | CREATE |
| GET | `/api/v1/credit-notes` | List credit notes | AUDIT_STAFF | No |

### Accounts Receivable
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| GET | `/api/v1/ar/aging` | AR aging by client (days overdue) | FIRM_PARTNER | No |
| GET | `/api/v1/ar/outstanding` | Outstanding balance by client | FIRM_PARTNER | No |
| GET | `/api/v1/ar/reminders` | Payment reminder schedule | FIRM_PARTNER | No |

## Database Tables

### Core Tables
- `invoices` (id UUID, invoice_number VARCHAR unique, client_id, engagement_id, issue_date, due_date, status ENUM, total_amount DECIMAL, tax_amount DECIMAL, snapshot_data JSONB, created_at, created_by, updated_at, updated_by, is_deleted)
- `invoice_line_items` (id UUID, invoice_id, description, quantity DECIMAL, unit_price DECIMAL, tax_amount DECIMAL, total_amount DECIMAL, source_type ENUM, snapshot_data JSONB, created_at)
- `payments` (id UUID, invoice_id, payment_method ENUM, amount DECIMAL, payment_date DATE, reference_number VARCHAR, status ENUM, recorded_by, recorded_at, cleared_at, created_at)
- `billing_memos` (id UUID, related_invoice_id, type ENUM (CREDIT_NOTE|ADJUSTMENT), memo_number, amount DECIMAL, Reason TEXT, status ENUM, created_at, created_by)
- `outbox_messages` (for InvoiceIssued, PaymentRecorded, etc.)

### Materialized Views
- `mv_ar_aging` (client_id, current, days_1_30, days_31_60, days_61_90, days_90plus, total_outstanding) - refreshed daily

### Indexes
- `uidx_invoices_invoice_number` on (invoice_number) where is_deleted=false
- `idx_invoices_client_id_status` on (client_id, status)
- `idx_invoices_engagement_id` on (engagement_id)
- `idx_payments_invoice_id` on (invoice_id)
- `idx_payments_payment_date` on (payment_date)

## CQRS
**Writes**: GORM for invoice/payment mutations, publish InvoiceIssued, PaymentRecorded
**Reads**: sqlc for invoice list, AR aging from materialized view
**Events**: InvoiceCreated, InvoiceIssued, PaymentRecorded, CreditNoteCreated, EngagementSettled (→ outbox_messages → NATS → Commission Service)

## Distributed Locking Strategy
1. Approve invoice issuance → acquire lock `invoice:{id}:issue`
2. Capture snapshot of engagement/cost data
3. Finalize invoice, release lock
4. Publish InvoiceIssued event (for AR tracking)

## Error Codes
`INVOICE_NOT_FOUND`
`INVOICE_LOCKED` - Cannot edit issued invoice
`INVOICE_NUMBER_MISMATCH` - Credit note original not found
`PAYMENT_EXCEEDS_BALANCE` - Payment > outstanding
`ENGAGEMENT_NOT_APPROVED_FOR_BILLING`
`DUPLICATE_PAYMENT_DETECTED` - Payment lock collision