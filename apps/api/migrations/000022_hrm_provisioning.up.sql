-- Migration: 000022_hrm_provisioning
-- Purpose: Create user provisioning request workflow tables and offboarding checklists.
-- SPEC: HRM_SPEC_v1.4.md §8 (User Provisioning Workflow), §11.22–§11.23,
--       §17.2.3 (Audit Events: PROVISIONING_*), §17.2.7 (Offboarding events)
-- Depends on: 000021 (employees, employment_contracts), 000020 (branches),
--             000001/000002 (users table)
-- Rollback risk: LOW — new tables only; no alterations to existing tables.
--               Any in-flight provisioning requests and offboarding checklists
--               will be lost on rollback. Safe on dev/staging. Never rollback
--               on production with active provisioning data.
-- Ordering note: SPEC §12 originally plans 000022 = hrm_professional (certifications,
--               training, CPE). Sprint 2 Day 3 starts with provisioning workflow
--               as the higher-priority deliverable. hrm_professional tables
--               follow in migration 000023.
-- Next migration: 000023 (hrm_professional)

-- ============================================================
-- 1. CREATE user_provisioning_requests (SPEC §11.22)
--    Supports 3 flows (SPEC §8.3):
--      - HCM 2-step: HoB approve (step 1) → HR approve (step 2) → SA execute
--      - HO 1-step: HR/CEO/CHAIRMAN create → SA execute (no branch step)
--      - Emergency: is_emergency=true bypasses approval; requires emergency_reason
--    Business rule: max 1 PENDING per employee (SPEC §8.4) — enforced by
--    uidx_provisioning_pending partial unique index.
-- ============================================================

CREATE TABLE IF NOT EXISTS user_provisioning_requests (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id             UUID        NOT NULL REFERENCES employees(id),
    requested_by            UUID        NOT NULL REFERENCES users(id),
    requested_role          VARCHAR(50) NOT NULL,
    requested_branch_id     UUID        REFERENCES branches(id),
    status                  VARCHAR(20) NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING','APPROVED','REJECTED','EXECUTED','CANCELLED')),
    approval_level          SMALLINT    NOT NULL DEFAULT 1,
    -- Step 1: HoB approval (HCM flow only; NULL for HO/emergency flows)
    branch_approver_id      UUID        REFERENCES users(id),
    branch_approved_at      TIMESTAMPTZ,
    branch_rejection_reason TEXT,
    -- Step 2: HR Manager approval
    hr_approver_id          UUID        REFERENCES users(id),
    hr_approved_at          TIMESTAMPTZ,
    hr_rejection_reason     TEXT,
    -- Execution by SA (atomic: creates user + assigns role + links employee)
    executed_by             UUID        REFERENCES users(id),
    executed_at             TIMESTAMPTZ,
    -- Emergency bypass (SPEC §8.3 Emergency Flow)
    is_emergency            BOOLEAN     NOT NULL DEFAULT false,
    emergency_reason        TEXT,
    notes                   TEXT,
    -- Request auto-expires 30 days after creation if unprocessed (SPEC §8.4)
    expires_at              TIMESTAMPTZ NOT NULL DEFAULT (now() + INTERVAL '30 days'),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Enforces SPEC §8.4: "Không thể có 2 PENDING request cho cùng 1 employee_id"
CREATE UNIQUE INDEX IF NOT EXISTS uidx_provisioning_pending
    ON user_provisioning_requests(employee_id) WHERE status = 'PENDING';

-- SPEC §11.22 named indexes
CREATE INDEX IF NOT EXISTS idx_provisioning_employee
    ON user_provisioning_requests(employee_id);
CREATE INDEX IF NOT EXISTS idx_provisioning_status
    ON user_provisioning_requests(status);

-- Additional FK indexes (TECHNICAL_RULES §1.9: index all FK columns)
CREATE INDEX IF NOT EXISTS idx_provisioning_requested_by
    ON user_provisioning_requests(requested_by);
CREATE INDEX IF NOT EXISTS idx_provisioning_branch
    ON user_provisioning_requests(requested_branch_id)
    WHERE requested_branch_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_provisioning_branch_approver
    ON user_provisioning_requests(branch_approver_id)
    WHERE branch_approver_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_provisioning_hr_approver
    ON user_provisioning_requests(hr_approver_id)
    WHERE hr_approver_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_provisioning_executed_by
    ON user_provisioning_requests(executed_by)
    WHERE executed_by IS NOT NULL;

-- ============================================================
-- 2. CREATE offboarding_checklists (SPEC §11.23)
--    Tracks both onboarding (SPEC §9.1) and offboarding (SPEC §9.4)
--    checklists. Items stored as JSONB array — template populated
--    at application layer when checklist is initiated.
--    Audit events: OFFBOARDING_INITIATED, OFFBOARDING_ITEM_COMPLETED,
--                  OFFBOARDING_COMPLETED (SPEC §17.2.7)
-- ============================================================

CREATE TABLE IF NOT EXISTS offboarding_checklists (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id    UUID        NOT NULL REFERENCES employees(id),
    checklist_type VARCHAR(20) NOT NULL DEFAULT 'OFFBOARDING'
        CHECK (checklist_type IN ('ONBOARDING','OFFBOARDING')),
    initiated_by   UUID        NOT NULL REFERENCES users(id),
    target_date    DATE,
    items          JSONB       NOT NULL DEFAULT '{"items":[]}',
    status         VARCHAR(20) NOT NULL DEFAULT 'IN_PROGRESS'
        CHECK (status IN ('IN_PROGRESS','COMPLETED','CANCELLED')),
    completed_at   TIMESTAMPTZ,
    notes          TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- SPEC §11.23 named indexes
CREATE INDEX IF NOT EXISTS idx_offboarding_employee
    ON offboarding_checklists(employee_id);
CREATE INDEX IF NOT EXISTS idx_offboarding_status
    ON offboarding_checklists(status);

-- Additional FK index (TECHNICAL_RULES §1.9)
CREATE INDEX IF NOT EXISTS idx_offboarding_initiated_by
    ON offboarding_checklists(initiated_by);
