<!-- spec-version: 1.2 | last-sync: 2026-04-16 | changes: none (no changes in v1.2) -->
> **Spec version**: 1.2 — Last sync: 2026-04-16 — No changes in v1.2

# Module 5: WorkingPapers - Audit Documentation

## Overview
Manages audit working papers with multi-level review workflows, templates, and JSONB snapshot storage for immutability.

## Bounded Context: WorkingPapers

### Responsibilities
- Working paper document lifecycle (DRAFT → IN_REVIEW → FINALIZED → SIGNED_OFF)
- Multi-level review approval (Auditor → Senior → Manager → Partner)
- Template management (VSA-compliant)
- File storage & versioning (MinIO)
- Review comments & issue tracking

## Key Features

### 1. Working Paper Document
**Entity**: WorkingPaper aggregate
- Folder-based organization (by engagement)
- Document type: `PROCEDURES`, `EVIDENCE`, `ANALYSIS`, `CONCLUSION`, `MANAGEMENT_LETTER`
- Status: `DRAFT`, `IN_REVIEW`, `COMMENTED`, `FINALIZED`, `SIGNED_OFF`
- Content: File reference (from MinIO) + JSONB snapshot (after finalization, immutable)
- Review chain: Created by Auditor → Senior review → Manager review → Partner sign-off

### 2. Review Workflow
**Entity**: WorkingPaperReview aggregate
- Reviewer role: `AUDITOR`, `SENIOR_AUDITOR`, `MANAGER`, `PARTNER`
- Review status: `PENDING`, `REVIEWED`, `REJECTED`
- Comments (optional): JSONB list of issues
- Review decision: `APPROVE`, `REQUEST_CHANGES`, `REJECT`
- Approval required at each level before next level can review

### 3. Templates
**Entity**: AuditTemplate aggregate
- Template type, version, content (JSONB or file reference)
- VSA compliance flags
- Bulk application to engagement (creates draft WP per engagement)
- Versioning: Can retire old templates (not delete)

### 4. Distributed Locking
**Critical Operations** (require Redis locks):
- Finalize working paper (prevent concurrent finalization)
- Partner sign-off (exclusive approval)

**Lock Key Pattern**: `working_paper:{id}:finalize`, `working_paper:{id}:sign_off`

## Code Structure

### Go Package Layout
```
modules/working_papers/
  ├── domain/
  │   ├── working_paper.go          (WorkingPaper aggregate root)
  │   ├── review.go                 (Review value object)
  │   ├── audit_template.go         (Template aggregate)
  │   └── working_paper_events.go   (WPCreated, WPFinalized, WPSignedOff, etc.)
  ├── application/
  │   ├── working_paper_service.go  (WorkingPaperService)
  │   ├── review_service.go         (ReviewService with approval workflow)
  │   ├── template_service.go       (TemplateService)
  │   └── finalization_service.go   (FinalizeService with distributed lock)
  ├── infrastructure/
  │   ├── postgres/
  │   │   ├── working_paper_repository.go (CQRS: sqlc reads)
  │   │   ├── review_repository.go
  │   │   └── template_repository.go
  │   └── redis/
  │       └── distributed_lock.go
  └── interfaces/
      └── rest/
          ├── working_paper_handler.go   (WorkingPaperHandler)
          ├── review_handler.go          (ReviewHandler)
          └── template_handler.go        (TemplateHandler)
```

## API Endpoints

**Authorization**: AUDIT_STAFF, AUDIT_MANAGER, FIRM_PARTNER (engagement-scoped)

### Working Papers
| Method | Path | Description | Auth | Audit | Lock |
|--------|------|-------------|------|-------|------|
| GET | `/api/v1/engagements/{engagement_id}/working-papers` | List WPs | AUDIT_STAFF | No | No |
| POST | `/api/v1/engagements/{engagement_id}/working-papers` | Create WP (status=DRAFT) | AUDIT_STAFF | CREATE | No |
| GET | `/api/v1/working-papers/{id}` | Get WP details with review chain | AUDIT_STAFF | No | No |
| PUT | `/api/v1/working-papers/{id}` | Update WP (DRAFT only) | AUDIT_STAFF | UPDATE | No |
| POST | `/api/v1/working-papers/{id}/submit-for-review` | Submit (status→IN_REVIEW) | AUDIT_STAFF | STATE_TRANSITION | No |
| POST | `/api/v1/working-papers/{id}/finalize` | Finalize (JSONB snapshot) | AUDIT_MANAGER | STATE_TRANSITION | working_paper:{id}:finalize |
| DELETE | `/api/v1/working-papers/{id}` | Soft delete (DRAFT only) | AUDIT_STAFF | DELETE | No |

**Fields** (snake_case): `id`, `engagement_id`, `folder_id`, `document_type`, `title`, `status`, `file_id` (MinIO reference), `snapshot_data` (JSONB, after finalize), `created_at`, `created_by`, `updated_at`, `updated_by`

### Review
| Method | Path | Description | Auth | Audit | Lock |
|--------|------|-------------|------|-------|------|
| GET | `/api/v1/working-papers/{id}/reviews` | List reviews in chain | AUDIT_STAFF | No | No |
| POST | `/api/v1/working-papers/{id}/reviews/{reviewer_role}/approve` | Approve at level | AUDIENCE_ROLE | APPROVE | No |
| POST | `/api/v1/working-papers/{id}/reviews/{reviewer_role}/request-changes` | Request changes | AUDIENCE_ROLE | REJECT | No |
| GET | `/api/v1/working-papers/{id}/reviews/{reviewer_role}/comments` | Get review comments | AUDIT_STAFF | No | No |
| POST | `/api/v1/working-papers/{id}/reviews/{reviewer_role}/comment` | Add comment | AUDIENCE_ROLE | CREATE | No |
| POST | `/api/v1/working-papers/{id}/sign-off` | Partner final sign-off | FIRM_PARTNER | APPROVE | working_paper:{id}:sign_off |

**Review Statuses**: `PENDING`, `REVIEWED`, `REJECTED`, `APPROVED`

### Templates
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| GET | `/api/v1/audit-templates` | List active templates | AUDIT_STAFF | No |
| POST | `/api/v1/audit-templates` | Create template | FIRM_PARTNER | CREATE |
| POST | `/api/v1/audit-templates/{id}/apply-to-engagement` | Bulk apply to engagement | FIRM_PARTNER | CREATE |
| PUT | `/api/v1/audit-templates/{id}` | Update (creates new version) | FIRM_PARTNER | UPDATE |
| DELETE | `/api/v1/audit-templates/{id}` | Retire template | FIRM_PARTNER | DELETE |

## Database Tables

### Core Tables
- `working_papers` (id UUID, engagement_id, folder_id, document_type ENUM, title, status ENUM, file_id UUID (refs file_metadata), snapshot_data JSONB (after finalize), created_at, created_by, updated_at, created_by, is_deleted)
- `working_paper_reviews` (id UUID, working_paper_id, reviewer_role ENUM, review_status ENUM, review_date, reviewed_by, comments JSONB list, created_at)
- `working_paper_comments` (id UUID, review_id, comment_text, issue_status ENUM (OPEN|RESOLVED), raised_at, resolved_at, created_by)
- `audit_templates` (id UUID, template_type ENUM, title, version, content JSONB, vsa_compliant BOOLEAN, is_active, created_at)
- `working_paper_folders` (id UUID, engagement_id, folder_name, created_at)
- `outbox_messages` (for WorkingPaperCreated, WorkingPaperFinalized, WorkingPaperSignedOff, etc.)

### Indexes
- `idx_working_papers_engagement_id` on (engagement_id, status)
- `idx_working_paper_reviews_working_paper_id` on (working_paper_id, reviewer_role)
- `idx_audit_templates_is_active` on (is_active)
- `idx_working_papers_file_id` on (file_id)

## CQRS
**Writes**: GORM for WP/review mutations, finalization captures snapshot
**Reads**: sqlc for WP list, review chain status checks
**Events**: WorkingPaperCreated, WorkingPaperSubmitted, WorkingPaperFinalized, WorkingPaperSignedOff

## Snapshot Strategy
Once WP is FINALIZED:
1. All current content (file + metadata) captured to `snapshot_data` JSONB
2. File marked immutable (no overwrites allowed)
3. Future edits create new WP revision (new ID)
4. Historical snapshots preserved for audit trail

## Error Codes
`WORKING_PAPER_NOT_FOUND`
`WORKING_PAPER_LOCKED` - Cannot edit finalized WP
`REVIEW_CHAIN_INCOMPLETE` - Cannot finalize until reviewers approve
`INVALID_REVIEW_SEQUENCE` - Wrong reviewer role for level
`COMMENTS_NOT_RESOLVED` - Unresolved issues in review
`TEMPLATE_NOT_FOUND`
`INVALID_STATE_TRANSITION`