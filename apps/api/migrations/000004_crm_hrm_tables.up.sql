-- ============================================================
-- CRM: Clients
-- ============================================================
CREATE TABLE clients (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tax_code      VARCHAR(14)  NOT NULL,
    business_name VARCHAR(200) NOT NULL,
    english_name  VARCHAR(200),
    industry      VARCHAR(100),
    status        VARCHAR(30)  NOT NULL DEFAULT 'PROSPECT'
        CHECK (status IN ('PROSPECT','ASSESSMENT','ACCEPTED','INACTIVE')),
    office_id     UUID,
    is_deleted    BOOLEAN      NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by    UUID REFERENCES users(id),
    updated_by    UUID REFERENCES users(id)
);

CREATE UNIQUE INDEX uidx_clients_tax_code ON clients(tax_code) WHERE is_deleted = false;
CREATE INDEX idx_clients_status    ON clients(status)    WHERE is_deleted = false;
CREATE INDEX idx_clients_office_id ON clients(office_id) WHERE is_deleted = false;

-- ============================================================
-- HRM: Employees
-- ============================================================
CREATE TABLE employees (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name         VARCHAR(200) NOT NULL,
    email             VARCHAR(255) NOT NULL,
    phone             VARCHAR(20),
    date_of_birth     DATE,
    grade             VARCHAR(20)  NOT NULL DEFAULT 'JUNIOR'
        CHECK (grade IN ('INTERN','JUNIOR','SENIOR','MANAGER','DIRECTOR','PARTNER')),
    position          VARCHAR(50),
    office_id         UUID,
    manager_id        UUID REFERENCES employees(id),
    hourly_rate       DECIMAL(10,2),
    status            VARCHAR(20)  NOT NULL DEFAULT 'ACTIVE'
        CHECK (status IN ('ACTIVE','ON_LEAVE','RESIGNED','RETIRED')),
    employment_date   DATE,
    contract_end_date DATE,
    is_deleted        BOOLEAN      NOT NULL DEFAULT false,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by        UUID REFERENCES users(id),
    updated_by        UUID REFERENCES users(id)
);

CREATE UNIQUE INDEX uidx_employees_email ON employees(email) WHERE is_deleted = false;
CREATE INDEX idx_employees_office_id  ON employees(office_id)  WHERE is_deleted = false;
CREATE INDEX idx_employees_manager_id ON employees(manager_id) WHERE is_deleted = false;
CREATE INDEX idx_employees_status     ON employees(status)     WHERE is_deleted = false;
