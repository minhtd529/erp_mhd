-- Migration: 000020_hrm_organization (DOWN)
-- Reverses 000020_hrm_organization.up.sql in strict reverse order.
-- Rollback risk: LOW — drops only columns/indexes/tables added by this migration.
--   Data in new columns (is_head_office, city, etc.) is lost on rollback.
--   Pre-existing columns (address, phone, is_active, code, branch_id, etc.) are NOT touched.

-- ============================================================
-- 3. DROP branch_departments (reverse of CREATE)
-- ============================================================
DROP INDEX IF EXISTS idx_branch_depts_active;
DROP INDEX IF EXISTS idx_branch_depts_dept;
DROP INDEX IF EXISTS idx_branch_depts_branch;
DROP TABLE IF EXISTS branch_departments;

-- ============================================================
-- 2. Revert departments changes (reverse of ALTER)
-- ============================================================
DROP INDEX IF EXISTS idx_departments_active;

ALTER TABLE departments
    DROP CONSTRAINT IF EXISTS chk_departments_dept_type;

ALTER TABLE departments
    DROP COLUMN IF EXISTS is_deleted,
    DROP COLUMN IF EXISTS authorization_file_id,
    DROP COLUMN IF EXISTS authorization_date,
    DROP COLUMN IF EXISTS authorization_doc_number,
    DROP COLUMN IF EXISTS head_employee_id,
    DROP COLUMN IF EXISTS dept_type,
    DROP COLUMN IF EXISTS description;

-- ============================================================
-- 1. Revert branches changes (reverse of ALTER)
-- ============================================================
DROP INDEX IF EXISTS uidx_branches_head_office;

ALTER TABLE branches
    DROP COLUMN IF EXISTS authorization_file_id,
    DROP COLUMN IF EXISTS authorization_date,
    DROP COLUMN IF EXISTS authorization_doc_number,
    DROP COLUMN IF EXISTS tax_code,
    DROP COLUMN IF EXISTS head_of_branch_user_id,
    DROP COLUMN IF EXISTS established_date,
    DROP COLUMN IF EXISTS city,
    DROP COLUMN IF EXISTS is_head_office;
