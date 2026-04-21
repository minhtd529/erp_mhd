-- Migration: 000020_hrm_organization
-- Purpose: HRM Organization schema foundation — branches + departments + branch_departments matrix
-- SPEC: HRM_SPEC_v1.4.md §3 (Organization), §11.1-11.3, §12.2
--       NOTE: Numbered 000020 (not 000019 as in SPEC v1.4) because 000019 is taken by working_papers.
-- Depends on: 000019 (working_papers), 000001 (users), existing branches + departments tables
-- Rollback risk: LOW — DDL only, no data migration
-- Known limitation: departments.branch_id (legacy single-branch FK) is kept as-is.
--   It is superseded by the branch_departments junction table but not dropped in this
--   migration to avoid breaking existing code paths. A future migration will clean this
--   up after auditing all code references.
-- Next migration: 000021 (hrm_employees_extended)

-- ============================================================
-- 1. ALTER branches — add HRM fields
-- ============================================================
-- Pre-existing columns (NOT re-added): id, code, name, address, phone, is_active,
--   created_at, updated_at, created_by, updated_by

ALTER TABLE branches
    ADD COLUMN IF NOT EXISTS is_head_office           BOOLEAN      NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS city                     VARCHAR(100),
    ADD COLUMN IF NOT EXISTS established_date         DATE,
    ADD COLUMN IF NOT EXISTS head_of_branch_user_id   UUID         REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS tax_code                 VARCHAR(20),
    ADD COLUMN IF NOT EXISTS authorization_doc_number VARCHAR(50),
    ADD COLUMN IF NOT EXISTS authorization_date       DATE,
    ADD COLUMN IF NOT EXISTS authorization_file_id    UUID;

-- Exactly one branch may be the head office at any time.
-- Note: branches has no is_deleted column; partial condition uses is_head_office = TRUE only.
CREATE UNIQUE INDEX IF NOT EXISTS uidx_branches_head_office
    ON branches(is_head_office)
    WHERE is_head_office = TRUE;

-- ============================================================
-- 2. ALTER departments — add HRM fields
-- ============================================================
-- Pre-existing columns (NOT re-added): id, branch_id, code, name, is_active,
--   created_at, updated_at, created_by, updated_by
-- branch_id (legacy single-branch FK) intentionally kept — see header note.

ALTER TABLE departments
    ADD COLUMN IF NOT EXISTS description              TEXT,
    ADD COLUMN IF NOT EXISTS dept_type                VARCHAR(20)  NOT NULL DEFAULT 'CORE',
    ADD COLUMN IF NOT EXISTS head_employee_id         UUID,
    ADD COLUMN IF NOT EXISTS authorization_doc_number VARCHAR(50),
    ADD COLUMN IF NOT EXISTS authorization_date       DATE,
    ADD COLUMN IF NOT EXISTS authorization_file_id    UUID,
    ADD COLUMN IF NOT EXISTS is_deleted               BOOLEAN      NOT NULL DEFAULT FALSE;

-- CHECK constraint added separately so it can be dropped cleanly in down migration.
ALTER TABLE departments
    ADD CONSTRAINT chk_departments_dept_type
    CHECK (dept_type IN ('CORE', 'SUPPORT'));

-- Soft-delete aware index for listing active departments.
CREATE INDEX IF NOT EXISTS idx_departments_active
    ON departments(is_active)
    WHERE is_deleted = FALSE;

-- ============================================================
-- 3. CREATE branch_departments — branch ↔ department matrix
-- ============================================================
-- head_employee_id has no FK to employees yet — employees not extended until migration 000021.
-- Composite PK (branch_id, department_id) enforces uniqueness of each pair.

CREATE TABLE IF NOT EXISTS branch_departments (
    branch_id        UUID        NOT NULL REFERENCES branches(id),
    department_id    UUID        NOT NULL REFERENCES departments(id),
    head_employee_id UUID,
    is_active        BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (branch_id, department_id)
);

CREATE INDEX IF NOT EXISTS idx_branch_depts_branch
    ON branch_departments(branch_id);

CREATE INDEX IF NOT EXISTS idx_branch_depts_dept
    ON branch_departments(department_id);

CREATE INDEX IF NOT EXISTS idx_branch_depts_active
    ON branch_departments(is_active)
    WHERE is_active = TRUE;
