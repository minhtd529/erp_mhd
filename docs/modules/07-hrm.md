<!-- spec-version: 1.2 | last-sync: 2026-04-16 | changes: added salesperson fields (is_salesperson, sales_commission_eligible, bank info) -->
> **Spec version**: 1.2 — Last sync: 2026-04-16 — Updated in v1.2

# Module 7: HRM - Human Resources Management

## Overview
Manages employee profiles, certifications, training, and performance reviews for the firm's team.

## Bounded Context: HRM

### Responsibilities
- Employee master data (profile, grade, position, organizational assignment)
- Professional certification tracking (expiry alerts)
- Training & CPE (Continuing Professional Education) tracking
- Performance reviews (KPI-based, linked to timesheet utilization)

## Key Features

### 1. Employee Management
**Entity**: Employee aggregate
- Full name, email, phone, date of birth
- Grade/level: `INTERN`, `JUNIOR`, `SENIOR`, `MANAGER`, `DIRECTOR`, `PARTNER`
- Position: `AUDITOR`, `TAX_ADVISOR`, `BUSINESS_ADVISOR`, `MANAGER`, etc.
- Office assignment, department, manager_id
- Hourly rate (default for engagements)
- Status: `ACTIVE`, `ON_LEAVE`, `RESIGNED`, `RETIRED`
- Employment date, contract end date (if fixed-term)

**Salesperson fields** (added v1.2, used by Commission Module):
- `is_salesperson` (bool) — employee được phân loại là người bán hàng/khai thác KH
- `sales_commission_eligible` (bool) — đủ điều kiện nhận hoa hồng
- `default_commission_plan_id` (UUID, nullable) — plan mặc định khi assign engagement commission
- `bank_account_number` (string, encrypted) — tài khoản ngân hàng để chi hoa hồng
- `bank_account_name` (string, encrypted) — tên chủ tài khoản

### 2. Certification Tracking
**Entity**: Certification aggregate
- Certification type: `CIA`, `CPA`, `ACCA`, `VACPA`, `CA_VIETNAM`, `OTHER`
- Issue date, expiry date
- Status: `VALID`, `EXPIRING_SOON` (< 90 days), `EXPIRED`, `PENDING_RENEWAL`
- Renewal tracking: Last renewal date, next renewal date
- Alert: Reminder email 90 days before expiry (via Asynq)

### 3. Training & CPE
**Entity**: TrainingRecord aggregate
- Training type: `INTERNAL`, `EXTERNAL`, `ONLINE`, `CONFERENCE`
- CPE hours (Continuing Professional Education)
- Training date, completion status
- Provider, cost
- VACPA compliance tracking: 120 hours/3-year rolling window

### 4. Performance Review
**Entity**: PerformanceReview aggregate
- Review period: Quarterly or Annual
- Reviewer: Manager/Director
- KPIs: Utilization rate (from Timesheet), billable hours, client satisfaction (if tracked), project delivery
- Rating: `EXCEEDS`, `MEETS`, `BELOW`
- Comments, recommendations
- Status: `DRAFT`, `SUBMITTED`, `REVIEWED`, `FINALIZED`

### 5. Distributed Locking
No critical concurrent operations; performance reviews are sequential by employee_id + period.

## Code Structure

### Go Package Layout
```
modules/hrm/
  ├── domain/
  │   ├── employee.go               (Employee aggregate)
  │   ├── certification.go          (Certification aggregate)
  │   ├── training_record.go        (TrainingRecord aggregate)
  │   ├── performance_review.go     (PerformanceReview aggregate)
  │   └── hrm_events.go             (EmployeeCreated, CertificationExpiring, etc.)
  ├── application/
  │   ├── employee_service.go       (EmployeeService)
  │   ├── certification_service.go  (CertificationService)
  │   ├── training_service.go       (TrainingService with CPE calculations)
  │   ├── performance_review_service.go
  │   └── alert_service.go          (AlertService - Asynq cert reminder jobs)
  ├── infrastructure/
  │   ├── postgres/
  │   │   ├── employee_repository.go (sqlc for reads)
  │   │   ├── certification_repository.go
  │   │   ├── training_repository.go
  │   │   └── performance_review_repository.go
  │   └── asynq/
  │       └── certification_alert_job.go (Scheduled: cert expiry reminders)
  └── interfaces/
      └── rest/
          ├── employee_handler.go            (EmployeeHandler)
          ├── certification_handler.go       (CertificationHandler)
          ├── training_handler.go            (TrainingHandler)
          └── performance_review_handler.go  (PerformanceReviewHandler)
```

## API Endpoints

**Authorization**: FIRM_PARTNER, AUDIT_MANAGER, own profile (read)

### Employee
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| GET | `/api/v1/employees` | List employees (paginated) | AUDIT_MANAGER | No |
| POST | `/api/v1/employees` | Create employee | FIRM_PARTNER | CREATE |
| GET | `/api/v1/employees/{id}` | Get employee profile | AUDIT_MANAGER, self | No |
| PUT | `/api/v1/employees/{id}` | Update employee | FIRM_PARTNER | UPDATE |
| DELETE | `/api/v1/employees/{id}` | Soft delete (mark RESIGNED) | FIRM_PARTNER | DELETE |

**Fields** (snake_case): `id`, `full_name`, `email`, `phone`, `date_of_birth`, `grade`, `position`, `office_id`, `manager_id`, `hourly_rate`, `status`, `employment_date`, `contract_end_date`, `is_salesperson`, `sales_commission_eligible`, `default_commission_plan_id`, `created_at`, `updated_at`

> `bank_account_number` và `bank_account_name` không expose qua API thông thường — chỉ readable bởi Accountant/Director khi xử lý payout hoa hồng.

### Certifications
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| GET | `/api/v1/employees/{id}/certifications` | List certifications | AUDIT_MANAGER, self | No |
| POST | `/api/v1/employees/{id}/certifications` | Record certification | FIRM_PARTNER | CREATE |
| GET | `/api/v1/employees/{id}/certifications/{cert_id}` | Get certification detail | AUDIT_MANAGER, self | No |
| PUT | `/api/v1/employees/{id}/certifications/{cert_id}` | Update certification | FIRM_PARTNER | UPDATE |
| POST | `/api/v1/employees/{id}/certifications/{cert_id}/renew` | Record renewal | FIRM_PARTNER | CREATE |

**Fields**: `id`, `employee_id`, `certification_type`, `issue_date`, `expiry_date`, `status`, `last_renewal_date`, `created_at`

### Training & CPE
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| GET | `/api/v1/employees/{id}/training-records` | List training | AUDIT_MANAGER, self | No |
| POST | `/api/v1/employees/{id}/training-records` | Record training | AUDIT_STAFF | CREATE |
| GET | `/api/v1/employees/{id}/cpe-status` | Get CPE hours (rolling 3y) | AUDIT_MANAGER, self | No |

**Fields** (snake_case): `id`, `employee_id`, `training_type`, `training_date`, `cpe_hours`, `provider`, `cost`, `completion_status`, `created_at`

### Performance Reviews
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| GET | `/api/v1/employees/{id}/performance-reviews` | List reviews | AUDIT_MANAGER, self | No |
| POST | `/api/v1/employees/{id}/performance-reviews` | Create review (status=DRAFT) | AUDIT_MANAGER | CREATE |
| GET | `/api/v1/employees/{id}/performance-reviews/{review_id}` | Get review detail | AUDIT_MANAGER, self | No |
| PUT | `/api/v1/employees/{id}/performance-reviews/{review_id}` | Update review (DRAFT) | AUDIT_MANAGER | UPDATE |
| POST | `/api/v1/employees/{id}/performance-reviews/{review_id}/submit` | Submit (status→SUBMITTED) | AUDIT_MANAGER | STATE_TRANSITION |
| POST | `/api/v1/employees/{id}/performance-reviews/{review_id}/finalize` | Finalize (status→FINALIZED) | FIRM_PARTNER | STATE_TRANSITION |

**Fields** (snake_case): `id`, `employee_id`, `review_period` (enum: Q1/Q2/Q3/Q4/ANNUAL), `utilization_rate`, `billable_hours`, `kpi_scores` (JSONB), `rating`, `comments`, `status`, `created_at`

## Database Tables

### Core Tables
- `employees` (id UUID, full_name, email unique, phone, date_of_birth, grade ENUM, position, office_id, manager_id, hourly_rate DECIMAL, status ENUM, employment_date, contract_end_date, is_salesperson BOOLEAN DEFAULT FALSE, sales_commission_eligible BOOLEAN DEFAULT FALSE, default_commission_plan_id UUID, bank_account_number TEXT, bank_account_name TEXT, created_at, updated_at, is_deleted)
- `certifications` (id UUID, employee_id, certification_type ENUM, issue_date, expiry_date, status ENUM, last_renewal_date, created_at)
- `training_records` (id UUID, employee_id, training_type ENUM, training_date DATE, cpe_hours DECIMAL, provider, cost DECIMAL, completion_status ENUM, created_at)
- `performance_reviews` (id UUID, employee_id, review_period ENUM, utilization_rate DECIMAL, billable_hours DECIMAL, kpi_scores JSONB, rating ENUM, comments TEXT, status ENUM, reviewer_id, created_at, finalized_at)
- `outbox_messages` (for EmployeeCreated, CertificationExpiring, CPEThresholdWarning, etc.)

### Materialized Views
- `mv_employee_cpe_status` (employee_id, total_cpe_hours_3y, compliance_status, last_updated) - refreshed monthly

### Indexes
- `uidx_employees_email` on (email) where is_deleted=false
- `idx_employees_office_id` on (office_id)
- `idx_certifications_employee_id_expiry_date` on (employee_id, expiry_date)
- `idx_training_records_employee_id` on (employee_id, training_date DESC)
- `idx_performance_reviews_employee_id_period` on (employee_id, review_period)

## CQRS
**Writes**: GORM for employee/cert/training/review mutations
**Reads**: sqlc for employee list, materialized view for CPE status
**Events**: EmployeeCreated, CertificationExpiring, CPEComplianceWarning, PerformanceReviewFinalized

## Asynq Job Scheduling
**Job 1**: `hrm:certification-expiry-alert`
- Runs daily at 07:00 AM
- Finds certifications with expiry_date = today + 90 days
- Updates status → EXPIRING_SOON
- Publishes CertificationExpiring event (triggers email alert)

**Job 2**: `hrm:cpe-compliance-check`
- Runs monthly (1st of month)
- Calculates rolling 3-year CPE hours per employee
- Flags employees < 120 hours with status BELOW_THRESHOLD
- Publishes CPEComplianceWarning event

## Error Codes
`EMPLOYEE_NOT_FOUND`
`DUPLICATE_EMAIL`
`CERTIFICATION_NOT_FOUND`
`TRAINING_RECORD_NOT_FOUND`
`PERFORMANCE_REVIEW_NOT_FOUND`
`CPE_HOURS_INSUFFICIENT` - Below 120/3 years for VACPA
`INVALID_STATE_TRANSITION`
`COMMISSION_PLAN_NOT_FOUND` - default_commission_plan_id references invalid plan