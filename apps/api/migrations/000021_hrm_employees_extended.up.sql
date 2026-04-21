-- Migration: 000021_hrm_employees_extended
-- Purpose: Extend employees with ~50 HRM columns; create employee_dependents,
--          insurance_rate_config, employee_salary_history, employment_contracts;
--          wire deferred FKs from migration 000020.
-- SPEC: HRM_SPEC_v1.4.md §4, §5.2–5.5, §11.4–11.9
-- Depends on: 000020 (hrm_organization), 000004 (employees base), 000001 (users)
-- Rollback risk: MEDIUM — down migration re-adds old CHECK constraints; any rows
--               with grade=EXECUTIVE/SUPPORT or status=INACTIVE/TERMINATED must be
--               cleaned before rollback succeeds.
-- Next migration: 000022 (hrm_professional_development)

-- ============================================================
-- 1. DROP old inline CHECK constraints on employees
--    grade old values: INTERN/JUNIOR/SENIOR/MANAGER/DIRECTOR/PARTNER
--    status old values: ACTIVE/ON_LEAVE/RESIGNED/RETIRED
--    Both mismatch SPEC §3.6 and §4.2 — replaced in section 3.
-- ============================================================

ALTER TABLE employees
    DROP CONSTRAINT IF EXISTS employees_grade_check,
    DROP CONSTRAINT IF EXISTS employees_status_check;

-- ============================================================
-- 2. ALTER employees — add all HRM columns
--    IF NOT EXISTS guards against re-runs; pre-existing columns
--    (manager_id, date_of_birth, is_deleted) are no-ops.
-- ============================================================

ALTER TABLE employees
    -- § Basic identity
    ADD COLUMN IF NOT EXISTS employee_code              VARCHAR(12),
    ADD COLUMN IF NOT EXISTS user_id                    UUID         REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS branch_id                  UUID         REFERENCES branches(id),
    ADD COLUMN IF NOT EXISTS department_id              UUID         REFERENCES departments(id),
    ADD COLUMN IF NOT EXISTS position_title             VARCHAR(200),
    ADD COLUMN IF NOT EXISTS employment_type            VARCHAR(20)  NOT NULL DEFAULT 'FULL_TIME',
    ADD COLUMN IF NOT EXISTS hired_date                 DATE,
    ADD COLUMN IF NOT EXISTS probation_end_date         DATE,
    ADD COLUMN IF NOT EXISTS termination_date           DATE,
    ADD COLUMN IF NOT EXISTS termination_reason         TEXT,
    ADD COLUMN IF NOT EXISTS current_contract_id        UUID,
    -- § Personal / PII
    ADD COLUMN IF NOT EXISTS display_name               VARCHAR(100),
    ADD COLUMN IF NOT EXISTS gender                     VARCHAR(10),
    ADD COLUMN IF NOT EXISTS place_of_birth             VARCHAR(200),
    ADD COLUMN IF NOT EXISTS nationality                VARCHAR(50)  DEFAULT 'Vietnamese',
    ADD COLUMN IF NOT EXISTS ethnicity                  VARCHAR(50),
    ADD COLUMN IF NOT EXISTS personal_email             VARCHAR(200),
    ADD COLUMN IF NOT EXISTS personal_phone             VARCHAR(20),
    ADD COLUMN IF NOT EXISTS work_phone                 VARCHAR(20),
    ADD COLUMN IF NOT EXISTS current_address            TEXT,
    ADD COLUMN IF NOT EXISTS permanent_address          TEXT,
    ADD COLUMN IF NOT EXISTS cccd_encrypted             TEXT,
    ADD COLUMN IF NOT EXISTS cccd_issued_date           DATE,
    ADD COLUMN IF NOT EXISTS cccd_issued_place          VARCHAR(200),
    ADD COLUMN IF NOT EXISTS passport_number            VARCHAR(50),
    ADD COLUMN IF NOT EXISTS passport_expiry            DATE,
    -- § Employment details
    ADD COLUMN IF NOT EXISTS hired_source               VARCHAR(50),
    ADD COLUMN IF NOT EXISTS referrer_employee_id       UUID         REFERENCES employees(id),
    ADD COLUMN IF NOT EXISTS probation_salary_pct       NUMERIC(5,2) DEFAULT 85.00,
    ADD COLUMN IF NOT EXISTS work_location              VARCHAR(20)  NOT NULL DEFAULT 'OFFICE',
    ADD COLUMN IF NOT EXISTS remote_days_per_week       SMALLINT     DEFAULT 0,
    -- § Qualifications
    ADD COLUMN IF NOT EXISTS education_level            VARCHAR(30),
    ADD COLUMN IF NOT EXISTS education_major            VARCHAR(200),
    ADD COLUMN IF NOT EXISTS education_school           VARCHAR(200),
    ADD COLUMN IF NOT EXISTS education_graduation_year  SMALLINT,
    ADD COLUMN IF NOT EXISTS vn_cpa_number              VARCHAR(50),
    ADD COLUMN IF NOT EXISTS vn_cpa_issued_date         DATE,
    ADD COLUMN IF NOT EXISTS vn_cpa_expiry_date         DATE,
    ADD COLUMN IF NOT EXISTS practicing_certificate_number VARCHAR(50),
    ADD COLUMN IF NOT EXISTS practicing_certificate_expiry DATE,
    -- § Salary / Bank — encrypted at app layer (AES-256-GCM)
    ADD COLUMN IF NOT EXISTS base_salary                NUMERIC(15,2),
    ADD COLUMN IF NOT EXISTS salary_currency            VARCHAR(3)   DEFAULT 'VND',
    ADD COLUMN IF NOT EXISTS salary_effective_date      DATE,
    ADD COLUMN IF NOT EXISTS bank_account_encrypted     TEXT,
    ADD COLUMN IF NOT EXISTS bank_name                  VARCHAR(100),
    ADD COLUMN IF NOT EXISTS bank_branch                VARCHAR(200),
    ADD COLUMN IF NOT EXISTS mst_ca_nhan_encrypted      TEXT,
    -- § Commission / Sales
    ADD COLUMN IF NOT EXISTS commission_rate            NUMERIC(5,2),
    ADD COLUMN IF NOT EXISTS commission_type            VARCHAR(20)  NOT NULL DEFAULT 'NONE',
    ADD COLUMN IF NOT EXISTS sales_target_yearly        NUMERIC(15,2),
    ADD COLUMN IF NOT EXISTS biz_dev_region             VARCHAR(100),
    -- § BHXH / Insurance
    ADD COLUMN IF NOT EXISTS so_bhxh_encrypted          TEXT,
    ADD COLUMN IF NOT EXISTS bhxh_registered_date       DATE,
    ADD COLUMN IF NOT EXISTS bhxh_province_code         VARCHAR(10),
    ADD COLUMN IF NOT EXISTS bhyt_card_number           VARCHAR(20),
    ADD COLUMN IF NOT EXISTS bhyt_expiry_date           DATE,
    ADD COLUMN IF NOT EXISTS bhyt_registered_hospital_code  VARCHAR(20),
    ADD COLUMN IF NOT EXISTS bhyt_registered_hospital_name  VARCHAR(200),
    ADD COLUMN IF NOT EXISTS tncn_registered            BOOLEAN      NOT NULL DEFAULT false;

-- ============================================================
-- 3. ADD new named CHECK constraints (named for clean DROP in down)
-- ============================================================

ALTER TABLE employees
    ADD CONSTRAINT chk_employees_grade
        CHECK (grade IN ('EXECUTIVE','PARTNER','DIRECTOR','MANAGER','SENIOR','JUNIOR','INTERN','SUPPORT')),
    ADD CONSTRAINT chk_employees_status
        CHECK (status IN ('ACTIVE','INACTIVE','ON_LEAVE','TERMINATED')),
    ADD CONSTRAINT chk_employees_employment_type
        CHECK (employment_type IN ('FULL_TIME','PART_TIME','INTERN')),
    ADD CONSTRAINT chk_employees_gender
        CHECK (gender IS NULL OR gender IN ('MALE','FEMALE','OTHER')),
    ADD CONSTRAINT chk_employees_hired_source
        CHECK (hired_source IS NULL OR hired_source IN ('REFERRAL','PORTAL','DIRECT','AGENCY')),
    ADD CONSTRAINT chk_employees_work_location
        CHECK (work_location IN ('OFFICE','REMOTE','HYBRID')),
    ADD CONSTRAINT chk_employees_education_level
        CHECK (education_level IS NULL OR education_level IN ('BACHELOR','MASTER','PHD','COLLEGE','OTHER')),
    ADD CONSTRAINT chk_employees_commission_type
        CHECK (commission_type IN ('FIXED','TIERED','NONE'));

ALTER TABLE employees
    ADD CONSTRAINT uq_employees_employee_code UNIQUE (employee_code);

-- ============================================================
-- 4. Trigger: auto-generate employee_code (SPEC §11.5)
--    Format: NV{YY}-{SEQ4}, e.g. NV26-0001
--    YY from hired_date (or CURRENT_DATE if null), SEQ resets each year.
-- ============================================================

CREATE OR REPLACE FUNCTION fn_employees_set_code()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
    v_year TEXT;
    v_seq  INT;
    v_code TEXT;
BEGIN
    IF NEW.employee_code IS NOT NULL THEN RETURN NEW; END IF;
    v_year := TO_CHAR(COALESCE(NEW.hired_date, CURRENT_DATE), 'YY');
    SELECT COUNT(*) + 1 INTO v_seq
    FROM employees WHERE employee_code LIKE 'NV' || v_year || '-%';
    v_code := 'NV' || v_year || '-' || LPAD(v_seq::TEXT, 4, '0');
    WHILE EXISTS (SELECT 1 FROM employees WHERE employee_code = v_code) LOOP
        v_seq  := v_seq + 1;
        v_code := 'NV' || v_year || '-' || LPAD(v_seq::TEXT, 4, '0');
    END LOOP;
    NEW.employee_code := v_code;
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_employees_set_code
BEFORE INSERT ON employees
FOR EACH ROW EXECUTE FUNCTION fn_employees_set_code();

-- ============================================================
-- 5. CREATE employee_dependents (SPEC §11.6)
--    Tracks dependents for TNCN tax-deduction purposes.
-- ============================================================

CREATE TABLE IF NOT EXISTS employee_dependents (
    id                       UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id              UUID         NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    full_name                VARCHAR(200) NOT NULL,
    relationship             VARCHAR(30)  NOT NULL
        CHECK (relationship IN ('SPOUSE','CHILD','PARENT','SIBLING','OTHER')),
    date_of_birth            DATE,
    cccd_or_birth_cert       VARCHAR(50),
    tax_deduction_registered BOOLEAN      NOT NULL DEFAULT false,
    tax_deduction_from       DATE,
    tax_deduction_to         DATE,
    notes                    TEXT,
    created_at               TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at               TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_employee_dependents_employee
    ON employee_dependents(employee_id);

-- ============================================================
-- 6. CREATE insurance_rate_config (SPEC §11.7)
--    Stores BHXH/BHYT/BHTN rates; only 1 row may be active
--    (effective_to IS NULL) per effective_from date.
-- ============================================================

CREATE TABLE IF NOT EXISTS insurance_rate_config (
    id                UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    effective_from    DATE         NOT NULL,
    effective_to      DATE,
    bhxh_employee_pct NUMERIC(5,2) NOT NULL DEFAULT 8.00,
    bhxh_employer_pct NUMERIC(5,2) NOT NULL DEFAULT 17.50,
    bhyt_employee_pct NUMERIC(5,2) NOT NULL DEFAULT 1.50,
    bhyt_employer_pct NUMERIC(5,2) NOT NULL DEFAULT 3.00,
    bhtn_employee_pct NUMERIC(5,2) NOT NULL DEFAULT 1.00,
    bhtn_employer_pct NUMERIC(5,2) NOT NULL DEFAULT 1.00,
    kpcd_employer_pct NUMERIC(5,2) NOT NULL DEFAULT 2.00,
    salary_base_bhxh  NUMERIC(15,2),
    max_bhxh_salary   NUMERIC(15,2),
    notes             TEXT,
    created_by        UUID         REFERENCES users(id),
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_insurance_rate_dates
        CHECK (effective_to IS NULL OR effective_to > effective_from)
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_insurance_rate_active
    ON insurance_rate_config(effective_from) WHERE effective_to IS NULL;

-- ============================================================
-- 7. CREATE employee_salary_history — immutable (SPEC §11.8)
--    PostgreSQL RULEs block all UPDATE and DELETE on this table.
-- ============================================================

CREATE TABLE IF NOT EXISTS employee_salary_history (
    id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id      UUID          NOT NULL REFERENCES employees(id),
    effective_date   DATE          NOT NULL,
    base_salary      NUMERIC(15,2) NOT NULL,
    allowances_total NUMERIC(15,2) DEFAULT 0,
    salary_note      TEXT,
    change_type      VARCHAR(30)   NOT NULL DEFAULT 'INITIAL'
        CHECK (change_type IN ('INITIAL','INCREASE','DECREASE','PROMOTION','ADJUSTMENT')),
    approved_by      UUID          REFERENCES users(id),
    created_by       UUID          REFERENCES users(id),
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE RULE no_update_salary_history
    AS ON UPDATE TO employee_salary_history DO INSTEAD NOTHING;
CREATE RULE no_delete_salary_history
    AS ON DELETE TO employee_salary_history DO INSTEAD NOTHING;

CREATE INDEX IF NOT EXISTS idx_salary_history_employee
    ON employee_salary_history(employee_id);
CREATE INDEX IF NOT EXISTS idx_salary_history_date
    ON employee_salary_history(effective_date);

-- ============================================================
-- 8. CREATE employment_contracts (SPEC §11.9)
--    employees.current_contract_id FK is added in section 9
--    (circular dependency: contracts ref employees, employees ref contracts).
-- ============================================================

CREATE TABLE IF NOT EXISTS employment_contracts (
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id         UUID         NOT NULL REFERENCES employees(id),
    contract_number     VARCHAR(50)  UNIQUE,
    contract_type       VARCHAR(20)  NOT NULL
        CHECK (contract_type IN ('PROBATION','DEFINITE_TERM','INDEFINITE','INTERN')),
    start_date          DATE         NOT NULL,
    end_date            DATE,
    signed_date         DATE,
    salary_at_signing   NUMERIC(15,2),
    position_at_signing VARCHAR(200),
    notes               TEXT,
    document_url        TEXT,
    is_current          BOOLEAN      NOT NULL DEFAULT false,
    created_by          UUID         REFERENCES users(id),
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_contract_dates CHECK (end_date IS NULL OR end_date > start_date)
);

CREATE INDEX IF NOT EXISTS idx_contracts_employee
    ON employment_contracts(employee_id);
CREATE INDEX IF NOT EXISTS idx_contracts_end_date
    ON employment_contracts(end_date) WHERE end_date IS NOT NULL;

-- ============================================================
-- 9. Wire deferred FKs
-- ============================================================

-- Circular dep resolved: employment_contracts now exists.
ALTER TABLE employees
    ADD CONSTRAINT fk_employees_current_contract
        FOREIGN KEY (current_contract_id) REFERENCES employment_contracts(id);

-- Deferred from migration 000020 (employees not yet extended then).
ALTER TABLE departments
    ADD CONSTRAINT fk_departments_head_employee
        FOREIGN KEY (head_employee_id) REFERENCES employees(id);

ALTER TABLE branch_departments
    ADD CONSTRAINT fk_branch_departments_head_employee
        FOREIGN KEY (head_employee_id) REFERENCES employees(id);

-- ============================================================
-- 10. Indexes on new employee columns (SPEC §11.4)
--     idx_employees_status pre-exists from migration 000004 — IF NOT EXISTS
--     is a no-op; it is intentionally not re-created here.
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_employees_branch
    ON employees(branch_id);
CREATE INDEX IF NOT EXISTS idx_employees_dept
    ON employees(department_id);
CREATE INDEX IF NOT EXISTS idx_employees_manager
    ON employees(manager_id);
CREATE INDEX IF NOT EXISTS idx_employees_grade
    ON employees(grade);
CREATE INDEX IF NOT EXISTS idx_employees_hired
    ON employees(hired_date);
