-- Migration: 000024_hrm_professional
-- Purpose: Create professional-development tables: certifications, training_courses,
--          training_records, cpe_requirements_by_role.
-- SPEC: HRM_SPEC_v1.4.md §11.9 (Certifications), §11.10 (Training Courses),
--       §11.11 (Training Records), §11.12 (CPE Requirements by Role)
-- Depends on: 000021 (employees), 000023 (roles seed)
-- Rollback risk: LOW — new tables only; no alterations to existing tables.
--               Cert / training data in flight will be lost on rollback.
--               Safe on dev/staging. Coordinate with HR before rolling back
--               on a DB with live certification records.
-- Next migration: 000025

-- ============================================================
-- 1. CREATE certifications (SPEC §11.9)
--    Tracks professional certifications held by employees
--    (VN CPA, ACCA, CPA Australia, IFRS, etc.).
--    Soft-deleted via is_deleted; hard deletes are not allowed
--    (audit trail required for lapsed certs).
-- ============================================================

CREATE TABLE IF NOT EXISTS certifications (
    id                UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id       UUID         NOT NULL REFERENCES employees(id),
    cert_type         VARCHAR(30)  NOT NULL
        CHECK (cert_type IN (
            'VN_CPA','ACCA','CPA_AUSTRALIA','CFA','CIA','CISA',
            'IFRS','ICAEW','CMA','OTHER'
        )),
    cert_name         VARCHAR(200) NOT NULL,
    cert_number       VARCHAR(100),
    issued_date       DATE,
    expiry_date       DATE,
    issuing_authority VARCHAR(200),
    status            VARCHAR(20)  NOT NULL DEFAULT 'ACTIVE'
        CHECK (status IN ('ACTIVE','EXPIRED','REVOKED','SUSPENDED')),
    document_url      TEXT,
    notes             TEXT,
    is_deleted        BOOLEAN      NOT NULL DEFAULT false,
    created_by        UUID         REFERENCES users(id),
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_cert_dates
        CHECK (expiry_date IS NULL OR issued_date IS NULL OR expiry_date > issued_date)
);

CREATE INDEX IF NOT EXISTS idx_certifications_employee
    ON certifications(employee_id);
CREATE INDEX IF NOT EXISTS idx_certifications_type
    ON certifications(cert_type);
CREATE INDEX IF NOT EXISTS idx_certifications_status
    ON certifications(status) WHERE is_deleted = false;
CREATE INDEX IF NOT EXISTS idx_certifications_expiry
    ON certifications(expiry_date) WHERE expiry_date IS NOT NULL AND is_deleted = false;

-- ============================================================
-- 2. CREATE training_courses (SPEC §11.10)
--    Master catalog of CPE / training courses offered or
--    recognised by the firm. course_code is unique.
--    cpe_hours: CPE credit hours awarded on completion.
-- ============================================================

CREATE TABLE IF NOT EXISTS training_courses (
    id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(30)   NOT NULL,
    name        VARCHAR(200)  NOT NULL,
    provider    VARCHAR(200),
    description TEXT,
    cpe_hours   NUMERIC(6,2)  NOT NULL DEFAULT 0
        CHECK (cpe_hours >= 0),
    course_type VARCHAR(20)   NOT NULL
        CHECK (course_type IN (
            'TECHNICAL','ETHICS','MANAGEMENT',
            'SOFT_SKILLS','COMPLIANCE','OTHER'
        )),
    is_active   BOOLEAN       NOT NULL DEFAULT true,
    notes       TEXT,
    created_by  UUID          REFERENCES users(id),
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    CONSTRAINT uq_training_courses_code UNIQUE (code)
);

CREATE INDEX IF NOT EXISTS idx_training_courses_type
    ON training_courses(course_type);
CREATE INDEX IF NOT EXISTS idx_training_courses_active
    ON training_courses(is_active) WHERE is_active = true;

-- ============================================================
-- 3. CREATE training_records (SPEC §11.11)
--    Records each employee's enrolment and completion of a
--    training course. cpe_hours_earned may differ from the
--    course default (e.g. partial completion).
--    Soft-deleted: is_deleted = true hides records without
--    destroying the CPE audit trail.
-- ============================================================

CREATE TABLE IF NOT EXISTS training_records (
    id                UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id       UUID          NOT NULL REFERENCES employees(id),
    course_id         UUID          NOT NULL REFERENCES training_courses(id),
    completion_date   DATE,
    cpe_hours_earned  NUMERIC(6,2)  NOT NULL DEFAULT 0
        CHECK (cpe_hours_earned >= 0),
    certificate_url   TEXT,
    status            VARCHAR(20)   NOT NULL DEFAULT 'ENROLLED'
        CHECK (status IN ('ENROLLED','IN_PROGRESS','COMPLETED','FAILED','CANCELLED')),
    notes             TEXT,
    is_deleted        BOOLEAN       NOT NULL DEFAULT false,
    created_by        UUID          REFERENCES users(id),
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT now(),
    CONSTRAINT chk_training_completion
        CHECK (completion_date IS NULL OR status IN ('COMPLETED','FAILED'))
);

CREATE INDEX IF NOT EXISTS idx_training_records_employee
    ON training_records(employee_id);
CREATE INDEX IF NOT EXISTS idx_training_records_course
    ON training_records(course_id);
CREATE INDEX IF NOT EXISTS idx_training_records_status
    ON training_records(status) WHERE is_deleted = false;
CREATE INDEX IF NOT EXISTS idx_training_records_completion
    ON training_records(completion_date) WHERE completion_date IS NOT NULL AND is_deleted = false;

-- ============================================================
-- 4. CREATE cpe_requirements_by_role (SPEC §11.12)
--    Defines annual CPE hour targets per role per year.
--    category_breakdown JSONB stores per-category minimums,
--    e.g. {"TECHNICAL":20,"ETHICS":4,"OTHER":6}.
--    One row per (role_code, year) — enforced by unique index.
-- ============================================================

CREATE TABLE IF NOT EXISTS cpe_requirements_by_role (
    id                   UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    role_code            VARCHAR(50)  NOT NULL,
    year                 SMALLINT     NOT NULL
        CHECK (year >= 2000 AND year <= 2100),
    required_hours       NUMERIC(6,2) NOT NULL
        CHECK (required_hours >= 0),
    category_breakdown   JSONB,
    notes                TEXT,
    created_by           UUID         REFERENCES users(id),
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uq_cpe_requirements_role_year UNIQUE (role_code, year)
);

CREATE INDEX IF NOT EXISTS idx_cpe_requirements_role
    ON cpe_requirements_by_role(role_code);
CREATE INDEX IF NOT EXISTS idx_cpe_requirements_year
    ON cpe_requirements_by_role(year);
