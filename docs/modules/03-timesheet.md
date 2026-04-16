<!-- spec-version: 1.2 | last-sync: 2026-04-16 | changes: none (no changes in v1.2) -->
> **Spec version**: 1.2 — Last sync: 2026-04-16 — No changes in v1.2

# Module 3: Timesheet - Time Tracking & Approval

## Overview
Tracks time entries against engagements and manages timesheet approval workflows with resource allocation constraints.

## Bounded Context: Timesheet

### Responsibilities
- Time entry recording (daily against engagement/task)
- Weekly timesheet submission & approval workflow
- Attendance tracking (check-in/check-out for on-site work)
- Utilization rate calculations (for reporting/resource planning)

## Key Features

### 1. Timesheet Entries
**Entity**: TimesheetEntry aggregate
- Entry date, hours, engagement_id, task_id, description
- Status: `DRAFT`, `SUBMITTED`, `APPROVED`, `REJECTED`
- Entries grouped by week (Mon-Sun)
- One entry = one task on one day (daily granularity)

### 2. Timesheet Submission & Approval
**Entity**: Timesheet aggregate (weekly)
- Period: Start date (Monday), End date (Sunday)
- Status: `DRAFT`, `SUBMITTED`, `APPROVED`, `LOCKED`
- Submission by staff → Manager approval
- Once locked (APPROVED), entries immutable
- Approval required for billing integration

### 3. Attendance Tracking
**Entity**: Attendance log (optional, for on-site work)
- Check-in/check-out times
- Location (on-site vs remote)
- Status: `PRESENT`, `LEAVE`, `HOLIDAY`, `ABSENT`

### 4. Distributed Locking
**Critical Operations** (require Redis locks):
- Approve timesheet (Manager concurrent approvals)
- Lock timesheet for billing (prevent further edits)

**Lock Key Pattern**: `timesheet:{week_id}:approve`, `timesheet:{week_id}:lock`

## Code Structure

### Go Package Layout
```
modules/timesheet/
  ├── domain/
  │   ├── timesheet.go              (Timesheet weekly aggregate)
  │   ├── timesheet_entry.go        (TimesheetEntry value object)
  │   ├── attendance.go             (Attendance aggregate)
  │   └── timesheet_events.go       (TimesheetSubmitted, TimesheetApproved, etc.)
  ├── application/
  │   ├── timesheet_service.go      (TimesheetService - use cases)
  │   ├── approval_service.go       (ApprovalService with distributed lock)
  │   └── entry_service.go          (EntryService)
  ├── infrastructure/
  │   ├── postgres/
  │   │   ├── timesheet_repository.go (CQRS: sqlc reads)
  │   │   ├── entry_repository.go
  │   │   └── attendance_repository.go
  │   └── redis/
  │       └── distributed_lock.go
  └── interfaces/
      └── rest/
          ├── timesheet_handler.go      (TimesheetHandler)
          ├── entry_handler.go          (TimesheetEntryHandler)
          └── attendance_handler.go     (AttendanceHandler)
```

## API Endpoints

**Authorization**: AUDIT_STAFF (own), AUDIT_MANAGER (team), FIRM_PARTNER (all)

### Timesheet Management
| Method | Path | Description | Auth | Audit | Lock |
|--------|------|-------------|------|-------|------|
| GET | `/api/v1/timesheets` | List user's timesheets | AUDIT_STAFF | No | No |
| GET | `/api/v1/timesheets/{week_id}` | Get timesheet for week | AUDIT_STAFF | No | No |
| POST | `/api/v1/timesheets/{week_id}/submit` | Submit for approval | AUDIT_STAFF | STATE_TRANSITION | No |
| POST | `/api/v1/timesheets/{week_id}/approve` | Manager approves | AUDIT_MANAGER | APPROVE | timesheet:{week_id}:approve |
| POST | `/api/v1/timesheets/{week_id}/reject` | Manager rejects | AUDIT_MANAGER | REJECT | timesheet:{week_id}:approve |
| POST | `/api/v1/timesheets/{week_id}/lock` | Lock for billing (immutable) | AUDIT_MANAGER | STATE_TRANSITION | timesheet:{week_id}:lock |

**Fields** (snake_case): `id` (week_id), `staff_id`, `period_start_date`, `status`, `total_hours`, `submitted_at`, `submitted_by`, `approved_at`, `approved_by`, `locked_at`

### Time Entries
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| GET | `/api/v1/timesheets/{week_id}/entries` | List entries for week | AUDIT_STAFF | No |
| POST | `/api/v1/timesheets/{week_id}/entries` | Record time entry | AUDIT_STAFF | CREATE |
| PUT | `/api/v1/timesheets/{week_id}/entries/{entry_id}` | Update entry (before submit) | AUDIT_STAFF | UPDATE |
| DELETE | `/api/v1/timesheets/{week_id}/entries/{entry_id}` | Delete entry (before submit) | AUDIT_STAFF | DELETE |

**Fields**: `id`, `timesheet_id`, `entry_date` (ISO 8601), `engagement_id`, `task_id`, `hours_worked` (DECIMAL), `description`, `created_at`

### Attendance
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| POST | `/api/v1/attendance/check-in` | Clock in | AUDIT_STAFF | CREATE |
| POST | `/api/v1/attendance/check-out` | Clock out | AUDIT_STAFF | UPDATE |
| GET | `/api/v1/attendance/my-records` | View attendance | AUDIT_STAFF | No |

## Database Tables

### Core Tables
- `timesheets` (id=week_id UUID, staff_id, period_start_date DATE, status ENUM, total_hours DECIMAL, submitted_at, submitted_by, approved_at, approved_by, locked_at, created_at)
- `timesheet_entries` (id UUID, timesheet_id, entry_date DATE, engagement_id, task_id, hours_worked DECIMAL, description TEXT, created_at)
- `attendance` (id UUID, staff_id, check_in_time TIMESTAMP, check_out_time TIMESTAMP, location ENUM, status ENUM (PRESENT|LEAVE|HOLIDAY), created_at)
- `outbox_messages` (for TimesheetSubmitted, TimesheetApproved, TimesheetLocked, etc.)

### Indexes
- `idx_timesheets_staff_id_period_start` on (staff_id, period_start_date DESC)
- `idx_timesheet_entries_timesheet_id` on (timesheet_id)
- `idx_timesheet_entries_engagement_id` on (engagement_id)
- `idx_attendance_staff_id_check_in_time` on (staff_id, check_in_time DESC)
- `uidx_timesheets_staff_week` on (staff_id, period_start_date) where is_deleted=false

## CQRS
**Writes**: GORM mutations + publish TimesheetApproved to outbox
**Reads**: sqlc for timesheet list, utilization rollups (materialized view `mv_timesheet_utilization`)
**Events**: TimesheetSubmitted, TimesheetApproved, TimesheetLocked, TimesheetRejected

## Distributed Locking Strategy
1. Manager approves timesheet → acquire lock `timesheet:{week_id}:approve`
2. Check all prior approvals + final validation
3. Update status to APPROVED, release lock
4. Publish TimesheetApproved event (triggers billing intake)

## Error Codes
`TIMESHEET_NOT_FOUND`
`TIMESHEET_LOCKED` - Cannot edit locked timesheet
`INVALID_STATE_TRANSITION` - e.g., cannot approve DRAFT (must be SUBMITTED)
`TOTAL_HOURS_EXCEEDS_LIMIT` - > 60 hours/week warning
`ENTRY_DATE_OUTSIDE_PERIOD` - Date not in timesheet week
`ENGAGEMENT_NOT_FOUND` - Referenced engagement deleted