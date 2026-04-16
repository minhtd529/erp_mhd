<!-- spec-version: 1.2 | last-sync: 2026-04-16 | changes: added primary_salesperson_id field -->
> **Spec version**: 1.2 — Last sync: 2026-04-16 — Updated in v1.2

# Module 2: Engagement - Service Engagement Lifecycle

## Overview
Manages the full lifecycle of audit and advisory engagements from proposal through settlement.

## Bounded Context: Engagement

### Responsibilities
- Engagement master data (client, service type, fee)
- Engagement state machine (DRAFT → ACTIVE → COMPLETED → SETTLED)
- Team member assignment with role-based allocation
- Task management (Planning, Fieldwork, Reporting phases)
- Direct cost tracking & approval

## Key Features

### 1. Engagement Lifecycle
**Entity**: Engagement aggregate
- State: `DRAFT`, `PROPOSAL`, `CONTRACTED`, `ACTIVE`, `COMPLETED`, `SETTLED`
- Service types: `AUDIT`, `REVIEW`, `COMPILATION`, `TAX_ADVISORY`, `BUSINESS_ADVISORY`
- Fee structure: Fixed, Time & Material (T&M), Retainer, Success
- Timeline: Planned vs Actual (start_date, end_date)
- Partner lead: Mandatory before CONTRACTED state

### 2. Team Management
**Entity**: EngagementMember aggregate
- Member role: `PARTNER`, `MANAGER`, `SENIOR_AUDITOR`, `AUDITOR`, `INTERN`
- Hourly rate (overrides staff default)
- Allocation %: 0-100% FTE
- Assignment period
- Status: `ASSIGNED`, `ACTIVE`, `COMPLETED`

### 3. Phase & Task Management
**Entities**: EngagementPhase, EngagementTask
- Phases: Planning, Fieldwork, Reporting (sequential)
- Tasks assigned to team members
- Progress tracking: `NOT_STARTED`, `IN_PROGRESS`, `COMPLETED`

### 4. Cost Tracking
**Entity**: DirectCost aggregate
- Types: Travel, Accommodation, Meals, Materials, Other
- Amount recorded with receipt/evidence
- Approval workflow: Staff → Manager → Partner
- Status: `DRAFT`, `SUBMITTED`, `APPROVED`, `REJECTED`

### 5. Distributed Locking
**Critical Operations** (require Redis locks):
- Edit engagement state (`ACTIVE` → `COMPLETED`)
- Assign team member (allocations must sum ≤ 100%)
- Approve direct cost (prevent concurrent approval)

**Lock Key Pattern**: `engagement:{engagement_id}:edit`, `engagement:{engagement_id}:member_assign`, `engagement:{engagement_id}:cost_approve`

## Code Structure

### Go Package Layout
```
modules/engagement/
  ├── domain/
  │   ├── engagement.go            (Engagement aggregate root)
  │   ├── team_member.go           (TeamMember value object)
  │   ├── engagement_task.go       (Task aggregate)
  │   ├── direct_cost.go           (DirectCost aggregate)
  │   └── engagement_events.go     (EngagementCreated, EngagementCompleted, etc.)
  ├── application/
  │   ├── engagement_service.go    (EngagementService - use cases)
  │   ├── team_service.go          (TeamService with distributed lock)
  │   ├── cost_service.go          (CostService with approval workflow)
  │   └── cost_approval_service.go (Orchestrates approval via Asynq)
  ├── infrastructure/
  │   ├── postgres/
  │   │   ├── engagement_repository.go (CQRS: reads via sqlc, writes via GORM)
  │   │   ├── team_repository.go
  │   │   └── cost_repository.go
  │   └── redis/
  │       └── distributed_lock.go  (Redis locks for concurrent operations)
  └── interfaces/
      └── rest/
          ├── engagement_handler.go       (EngagementHandler)
          ├── team_handler.go             (TeamHandler)
          ├── task_handler.go             (TaskHandler)
          └── cost_handler.go             (CostHandler)
```

## API Endpoints

**Authorization**: FIRM_PARTNER, AUDIT_MANAGER, AUDIT_STAFF (with engagement-scoped ABAC)

### Engagement CRUD
| Method | Path | Description | Auth | Audit | Lock |
|--------|------|-------------|------|-------|------|
| GET | `/api/v1/engagements` | List (paginated, filtered by office) | AUDIT_STAFF | No | No |
| POST | `/api/v1/engagements` | Create (status=DRAFT) | FIRM_PARTNER | CREATE | No |
| GET | `/api/v1/engagements/{id}` | Get details with team, costs | AUDIT_STAFF | No | No |
| PUT | `/api/v1/engagements/{id}` | Update engagement | FIRM_PARTNER | UPDATE | No |
| POST | `/api/v1/engagements/{id}/activate` | State → ACTIVE | FIRM_PARTNER | STATE_TRANSITION | engagement:{id}:edit |
| POST | `/api/v1/engagements/{id}/complete` | State → COMPLETED | FIRM_PARTNER | STATE_TRANSITION | engagement:{id}:edit |

**Fields** (snake_case): `id`, `client_id`, `service_type`, `fee_type`, `fee_amount`, `status`, `partner_id`, `primary_salesperson_id`, `start_date`, `end_date`, `created_at`, `created_by`, `updated_at`, `updated_by`, `is_deleted`

### Team Management
| Method | Path | Description | Auth | Audit | Lock |
|--------|------|-------------|------|-------|------|
| GET | `/api/v1/engagements/{id}/members` | List team members | AUDIT_STAFF | No | No |
| POST | `/api/v1/engagements/{id}/members` | Assign member | AUDIT_MANAGER | CREATE | engagement:{id}:member_assign |
| PUT | `/api/v1/engagements/{id}/members/{member_id}` | Update allocation % | AUDIT_MANAGER | UPDATE | engagement:{id}:member_assign |
| DELETE | `/api/v1/engagements/{id}/members/{member_id}` | Unassign member | AUDIT_MANAGER | DELETE | engagement:{id}:member_assign |

**Validation**: Sum of allocation % ≤ 100% (checked in lock)

### Tasks
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| GET | `/api/v1/engagements/{id}/tasks` | List tasks by phase | AUDIT_STAFF | No |
| POST | `/api/v1/engagements/{id}/tasks` | Create task | AUDIT_MANAGER | CREATE |
| PUT | `/api/v1/engagements/{id}/tasks/{task_id}` | Update task status | AUDIT_STAFF | UPDATE |

### Direct Costs
| Method | Path | Description | Auth | Audit | Lock |
|--------|------|-------------|------|-------|------|
| GET | `/api/v1/engagements/{id}/costs` | List costs | AUDIT_STAFF | No | No |
| POST | `/api/v1/engagements/{id}/costs` | Record cost (status=DRAFT) | AUDIT_STAFF | CREATE | No |
| POST | `/api/v1/engagements/{id}/costs/{cost_id}/submit` | Submit for approval | AUDIT_STAFF | STATE_TRANSITION | No |
| POST | `/api/v1/engagements/{id}/costs/{cost_id}/approve` | Manager approves | AUDIT_MANAGER | APPROVE | engagement:{id}:cost_approve |
| POST | `/api/v1/engagements/{id}/costs/{cost_id}/reject` | Manager rejects | AUDIT_MANAGER | REJECT | engagement:{id}:cost_approve |

## Database Tables

### Core Tables
- `engagements` (id UUID, client_id, service_type ENUM, fee_type ENUM, fee_amount DECIMAL, status ENUM, partner_id, primary_salesperson_id UUID REFERENCES employees(id), start_date, end_date, created_at, created_by, updated_at, updated_by, is_deleted)
- `engagement_members` (id, engagement_id, staff_id, role ENUM, hourly_rate DECIMAL, allocation_percent INT, status ENUM, created_at)
- `engagement_phases` (id, engagement_id, phase ENUM, start_date, end_date, created_at)
- `engagement_tasks` (id, engagement_id, phase_id, title, assigned_to, status ENUM, created_at)
- `direct_costs` (id, engagement_id, type ENUM, description, amount DECIMAL, status ENUM, receipt_file_id, submitted_at, submitted_by, approved_at, approved_by, created_at)
- `outbox_messages` (for EngagementCreated, EngagementActivated, EngagementCompleted, etc.)

### Indexes
- `idx_engagements_client_id_status` on (client_id, status)
- `idx_engagements_partner_id` on (partner_id)
- `idx_engagement_members_staff_id` on (staff_id)
- `idx_direct_costs_engagement_id_status` on (engagement_id, status)

## CQRS
**Writes**: GORM for engagement/team/cost mutations, publish to outbox
**Reads**: sqlc for engagement list, team member allocation checks (real-time)
**Events**: EngagementCreated, EngagementActivated, EngagementCompleted, TeamMemberAssigned, CostApproved

## Distributed Locking Strategy
All critical operations use `RedisDistributedLock` with 30-second timeout:
1. Acquire lock with transaction context
2. Perform state check + update in transaction
3. Release lock
4. Publish domain event to outbox

## Error Codes
`ENGAGEMENT_NOT_FOUND`
`INVALID_STATE_TRANSITION`
`ENGAGEMENT_LOCKED` - Cannot perform concurrent mutation
`TEAM_ALLOCATION_EXCEEDS_100` - Sum > 100%
`PARTNER_NOT_ASSIGNED`
`COST_APPROVAL_REQUIRED`