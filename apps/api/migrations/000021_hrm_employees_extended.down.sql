-- Migration: 000021_hrm_employees_extended (DOWN)
-- Reverses everything in 000021 up — strict reverse order.
-- WARNING: If any employees row has grade IN ('EXECUTIVE','SUPPORT') or
--          status IN ('INACTIVE','TERMINATED'), the re-added old CHECK constraints
--          at the end of this script will fail. Clean data first.

-- ============================================================
-- 1. Drop deferred FKs (must precede table drops)
-- ============================================================

ALTER TABLE branch_departments
    DROP CONSTRAINT IF EXISTS fk_branch_departments_head_employee;

ALTER TABLE departments
    DROP CONSTRAINT IF EXISTS fk_departments_head_employee;

ALTER TABLE employees
    DROP CONSTRAINT IF EXISTS fk_employees_current_contract;

-- ============================================================
-- 2. Drop tables (reverse order: contracts, salary history,
--    insurance config, dependents)
-- ============================================================

DROP TABLE IF EXISTS employment_contracts;

-- Rules are auto-dropped with the table, but explicit for clarity.
DROP RULE IF EXISTS no_delete_salary_history ON employee_salary_history;
DROP RULE IF EXISTS no_update_salary_history ON employee_salary_history;
DROP TABLE IF EXISTS employee_salary_history;

DROP TABLE IF EXISTS insurance_rate_config;
DROP TABLE IF EXISTS employee_dependents;

-- ============================================================
-- 3. Drop trigger and function
-- ============================================================

DROP TRIGGER IF EXISTS trg_employees_set_code ON employees;
DROP FUNCTION IF EXISTS fn_employees_set_code();

-- ============================================================
-- 4. Drop indexes added by this migration
--    (idx_employees_status was NOT created here — do not drop it)
-- ============================================================

DROP INDEX IF EXISTS idx_employees_hired;
DROP INDEX IF EXISTS idx_employees_grade;
DROP INDEX IF EXISTS idx_employees_manager;
DROP INDEX IF EXISTS idx_employees_dept;
DROP INDEX IF EXISTS idx_employees_branch;

-- ============================================================
-- 5. Drop employee_code unique constraint
-- ============================================================

ALTER TABLE employees
    DROP CONSTRAINT IF EXISTS uq_employees_employee_code;

-- ============================================================
-- 6. Drop new named CHECK constraints
-- ============================================================

ALTER TABLE employees
    DROP CONSTRAINT IF EXISTS chk_employees_commission_type,
    DROP CONSTRAINT IF EXISTS chk_employees_education_level,
    DROP CONSTRAINT IF EXISTS chk_employees_work_location,
    DROP CONSTRAINT IF EXISTS chk_employees_hired_source,
    DROP CONSTRAINT IF EXISTS chk_employees_gender,
    DROP CONSTRAINT IF EXISTS chk_employees_employment_type,
    DROP CONSTRAINT IF EXISTS chk_employees_status,
    DROP CONSTRAINT IF EXISTS chk_employees_grade;

-- ============================================================
-- 7. Drop all columns added in section 2 of the up migration
-- ============================================================

ALTER TABLE employees
    -- BHXH / Insurance
    DROP COLUMN IF EXISTS tncn_registered,
    DROP COLUMN IF EXISTS bhyt_registered_hospital_name,
    DROP COLUMN IF EXISTS bhyt_registered_hospital_code,
    DROP COLUMN IF EXISTS bhyt_expiry_date,
    DROP COLUMN IF EXISTS bhyt_card_number,
    DROP COLUMN IF EXISTS bhxh_province_code,
    DROP COLUMN IF EXISTS bhxh_registered_date,
    DROP COLUMN IF EXISTS so_bhxh_encrypted,
    -- Commission / Sales
    DROP COLUMN IF EXISTS biz_dev_region,
    DROP COLUMN IF EXISTS sales_target_yearly,
    DROP COLUMN IF EXISTS commission_type,
    DROP COLUMN IF EXISTS commission_rate,
    -- Salary / Bank
    DROP COLUMN IF EXISTS mst_ca_nhan_encrypted,
    DROP COLUMN IF EXISTS bank_branch,
    DROP COLUMN IF EXISTS bank_name,
    DROP COLUMN IF EXISTS bank_account_encrypted,
    DROP COLUMN IF EXISTS salary_effective_date,
    DROP COLUMN IF EXISTS salary_currency,
    DROP COLUMN IF EXISTS base_salary,
    -- Qualifications
    DROP COLUMN IF EXISTS practicing_certificate_expiry,
    DROP COLUMN IF EXISTS practicing_certificate_number,
    DROP COLUMN IF EXISTS vn_cpa_expiry_date,
    DROP COLUMN IF EXISTS vn_cpa_issued_date,
    DROP COLUMN IF EXISTS vn_cpa_number,
    DROP COLUMN IF EXISTS education_graduation_year,
    DROP COLUMN IF EXISTS education_school,
    DROP COLUMN IF EXISTS education_major,
    DROP COLUMN IF EXISTS education_level,
    -- Employment details
    DROP COLUMN IF EXISTS remote_days_per_week,
    DROP COLUMN IF EXISTS work_location,
    DROP COLUMN IF EXISTS probation_salary_pct,
    DROP COLUMN IF EXISTS referrer_employee_id,
    DROP COLUMN IF EXISTS hired_source,
    -- Personal / PII
    DROP COLUMN IF EXISTS passport_expiry,
    DROP COLUMN IF EXISTS passport_number,
    DROP COLUMN IF EXISTS cccd_issued_place,
    DROP COLUMN IF EXISTS cccd_issued_date,
    DROP COLUMN IF EXISTS cccd_encrypted,
    DROP COLUMN IF EXISTS permanent_address,
    DROP COLUMN IF EXISTS current_address,
    DROP COLUMN IF EXISTS work_phone,
    DROP COLUMN IF EXISTS personal_phone,
    DROP COLUMN IF EXISTS personal_email,
    DROP COLUMN IF EXISTS ethnicity,
    DROP COLUMN IF EXISTS nationality,
    DROP COLUMN IF EXISTS place_of_birth,
    DROP COLUMN IF EXISTS gender,
    DROP COLUMN IF EXISTS display_name,
    -- Basic identity
    DROP COLUMN IF EXISTS current_contract_id,
    DROP COLUMN IF EXISTS termination_reason,
    DROP COLUMN IF EXISTS termination_date,
    DROP COLUMN IF EXISTS probation_end_date,
    DROP COLUMN IF EXISTS hired_date,
    DROP COLUMN IF EXISTS employment_type,
    DROP COLUMN IF EXISTS position_title,
    DROP COLUMN IF EXISTS department_id,
    DROP COLUMN IF EXISTS branch_id,
    DROP COLUMN IF EXISTS user_id,
    DROP COLUMN IF EXISTS employee_code;

-- ============================================================
-- 8. Restore original inline CHECK constraints (migration 000004)
-- ============================================================

ALTER TABLE employees
    ADD CONSTRAINT employees_grade_check
        CHECK (grade IN ('INTERN','JUNIOR','SENIOR','MANAGER','DIRECTOR','PARTNER')),
    ADD CONSTRAINT employees_status_check
        CHECK (status IN ('ACTIVE','ON_LEAVE','RESIGNED','RETIRED'));
