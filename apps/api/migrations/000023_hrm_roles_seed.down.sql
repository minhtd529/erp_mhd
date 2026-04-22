-- Migration: 000023_hrm_roles_seed (DOWN)
-- Removes only the roles introduced by this migration.
-- AUDIT_MANAGER is owned by 000001_init_schema and is NOT deleted here.
-- WARNING: Deletes role rows. Any user_roles rows referencing these roles
--          are removed by ON DELETE CASCADE. Never rollback on production.

DELETE FROM roles
WHERE code IN (
    'CHAIRMAN',
    'CEO',
    'PARTNER',
    'HR_MANAGER',
    'HEAD_OF_BRANCH',
    'SENIOR_AUDITOR',
    'HR_STAFF',
    'JUNIOR_AUDITOR',
    'ACCOUNTANT'
);
