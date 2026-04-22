-- Migration: 000023_hrm_roles_seed
-- Purpose: Seed HRM-specific roles defined in SPEC §3.5 that are absent from
--          the initial schema migration (000001_init_schema).
-- SPEC: HRM_SPEC_v1.4.md §3.5 — Roles (11 roles, all Phase 1 active)
-- Depends on: 000001 (roles table with UNIQUE(code) constraint)
-- Rollback risk: LOW — pure data seed; no schema changes.
--               Rolling back removes these role rows. Any users assigned to
--               these roles will have orphaned user_roles records (FK cascade
--               ON DELETE CASCADE on role_id will clean them). Safe on
--               dev/staging only. Never rollback on production with live users.
-- Note on 000001 overlap: AUDIT_MANAGER is already seeded in 000001.
--   The INSERT below includes it with ON CONFLICT DO NOTHING for completeness;
--   it will be a no-op on any DB that has run 000001.

-- ============================================================
-- Seed HRM roles (SPEC §3.5)
-- ============================================================

INSERT INTO roles (code, name, description, level, is_system) VALUES
    -- Executive tier (level 1)
    ('CHAIRMAN',        'Chairman',           'Board of Directors Chairman — full authority across all branches', 1, true),
    ('CEO',             'CEO',                'Chief Executive Officer — full operational authority',            1, true),
    -- Senior professional tier (level 2)
    ('PARTNER',         'Audit Partner',      'Audit partner — oversight of own department and branch',         2, true),
    -- Management tier (level 3)
    ('HR_MANAGER',      'HR Manager',         'Human Resources Manager — manage HR across all branches',        3, true),
    ('HEAD_OF_BRANCH',  'Head of Branch',     'Head of Branch — manage own branch operations',                  3, true),
    ('SENIOR_AUDITOR',  'Senior Auditor',     'Senior Auditor — audit execution within own department',         3, true),
    -- Staff tier (level 4)
    ('HR_STAFF',        'HR Staff',           'HR Staff — human resources operations at assigned branch',       4, true),
    ('JUNIOR_AUDITOR',  'Junior Auditor',     'Junior Auditor — audit task execution, self scope only',         4, true),
    ('ACCOUNTANT',      'Accountant',         'Accountant — finance and accounting operations, self scope',     4, true),
    -- Already in 000001; included for completeness — no-op on existing DBs
    ('AUDIT_MANAGER',   'Audit Manager',      'Audit Manager — manage engagements within own department',       3, true)
ON CONFLICT (code) DO NOTHING;
