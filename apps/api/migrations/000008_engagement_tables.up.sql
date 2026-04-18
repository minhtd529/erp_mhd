-- Phase 2: Engagement bounded context tables

CREATE TABLE IF NOT EXISTS engagements (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id               UUID NOT NULL REFERENCES clients(id),
    service_type            VARCHAR(30) NOT NULL
        CHECK (service_type IN ('AUDIT','REVIEW','COMPILATION','TAX_ADVISORY','BUSINESS_ADVISORY')),
    fee_type                VARCHAR(20) NOT NULL
        CHECK (fee_type IN ('FIXED','TIME_AND_MATERIAL','RETAINER','SUCCESS')),
    fee_amount              NUMERIC(18,2) NOT NULL DEFAULT 0,
    status                  VARCHAR(20) NOT NULL DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT','PROPOSAL','CONTRACTED','ACTIVE','COMPLETED','SETTLED')),
    partner_id              UUID REFERENCES users(id),
    primary_salesperson_id  UUID REFERENCES employees(id),
    start_date              DATE,
    end_date                DATE,
    description             TEXT,
    is_deleted              BOOLEAN NOT NULL DEFAULT false,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by              UUID NOT NULL REFERENCES users(id),
    updated_by              UUID NOT NULL REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_engagements_client_id_status  ON engagements(client_id, status) WHERE is_deleted = false;
CREATE INDEX IF NOT EXISTS idx_engagements_partner_id         ON engagements(partner_id)        WHERE is_deleted = false;
CREATE INDEX IF NOT EXISTS idx_engagements_status             ON engagements(status)             WHERE is_deleted = false;

CREATE TABLE IF NOT EXISTS engagement_members (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    engagement_id     UUID NOT NULL REFERENCES engagements(id) ON DELETE CASCADE,
    staff_id          UUID NOT NULL REFERENCES users(id),
    role              VARCHAR(20) NOT NULL
        CHECK (role IN ('PARTNER','MANAGER','SENIOR_AUDITOR','AUDITOR','INTERN')),
    hourly_rate       NUMERIC(10,2),
    allocation_percent INT NOT NULL DEFAULT 100
        CHECK (allocation_percent >= 0 AND allocation_percent <= 100),
    status            VARCHAR(20) NOT NULL DEFAULT 'ASSIGNED'
        CHECK (status IN ('ASSIGNED','ACTIVE','COMPLETED')),
    is_deleted        BOOLEAN NOT NULL DEFAULT false,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by        UUID NOT NULL REFERENCES users(id),
    updated_by        UUID NOT NULL REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_engagement_members_engagement_id ON engagement_members(engagement_id) WHERE is_deleted = false;
CREATE INDEX IF NOT EXISTS idx_engagement_members_staff_id      ON engagement_members(staff_id)      WHERE is_deleted = false;

CREATE TABLE IF NOT EXISTS engagement_tasks (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    engagement_id UUID NOT NULL REFERENCES engagements(id) ON DELETE CASCADE,
    phase         VARCHAR(20) NOT NULL
        CHECK (phase IN ('PLANNING','FIELDWORK','REPORTING')),
    title         VARCHAR(500) NOT NULL,
    assigned_to   UUID REFERENCES users(id),
    status        VARCHAR(20) NOT NULL DEFAULT 'NOT_STARTED'
        CHECK (status IN ('NOT_STARTED','IN_PROGRESS','COMPLETED')),
    due_date      DATE,
    is_deleted    BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by    UUID NOT NULL REFERENCES users(id),
    updated_by    UUID NOT NULL REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_engagement_tasks_engagement_id ON engagement_tasks(engagement_id) WHERE is_deleted = false;

CREATE TABLE IF NOT EXISTS direct_costs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    engagement_id UUID NOT NULL REFERENCES engagements(id) ON DELETE CASCADE,
    cost_type     VARCHAR(20) NOT NULL
        CHECK (cost_type IN ('TRAVEL','ACCOMMODATION','MEALS','MATERIALS','OTHER')),
    description   TEXT NOT NULL,
    amount        NUMERIC(18,2) NOT NULL,
    status        VARCHAR(20) NOT NULL DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT','SUBMITTED','APPROVED','REJECTED')),
    receipt_url   TEXT,
    submitted_at  TIMESTAMPTZ,
    submitted_by  UUID REFERENCES users(id),
    approved_at   TIMESTAMPTZ,
    approved_by   UUID REFERENCES users(id),
    reject_reason TEXT,
    is_deleted    BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by    UUID NOT NULL REFERENCES users(id),
    updated_by    UUID NOT NULL REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_direct_costs_engagement_id_status ON direct_costs(engagement_id, status) WHERE is_deleted = false;
