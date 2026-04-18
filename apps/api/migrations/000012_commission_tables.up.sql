-- ============================================================
-- Commission Module: plans, engagement assignments, records
-- ============================================================

CREATE TABLE commission_plans (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    code         VARCHAR(100) NOT NULL UNIQUE,
    name         VARCHAR(200) NOT NULL,
    description  TEXT         NOT NULL DEFAULT '',
    type         VARCHAR(20)  NOT NULL CHECK (type IN ('flat','tiered','fixed','custom')),
    default_rate NUMERIC(8,6) NOT NULL DEFAULT 0,  -- 0.05 = 5%
    tiers        JSONB        NOT NULL DEFAULT '[]',
    apply_base   VARCHAR(30)  NOT NULL CHECK (apply_base IN ('fee_contracted','fee_invoiced','fee_paid','gross_margin')),
    trigger_on   VARCHAR(30)  NOT NULL CHECK (trigger_on IN ('contract_signed','invoice_issued','payment_received','eng_completed')),
    service_types JSONB       NOT NULL DEFAULT '[]',
    is_active    BOOLEAN      NOT NULL DEFAULT true,
    created_by   UUID         NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_commission_plans_is_active ON commission_plans (is_active);
CREATE INDEX idx_commission_plans_type      ON commission_plans (type);

CREATE TABLE engagement_commissions (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    engagement_id    UUID        NOT NULL REFERENCES engagements(id),
    salesperson_id   UUID        NOT NULL REFERENCES employees(id),
    role             VARCHAR(30) NOT NULL CHECK (role IN ('primary','referrer','account_manager','technical_lead')),
    plan_id          UUID        REFERENCES commission_plans(id),

    -- Calculation config (overrides plan when set)
    rate_type        VARCHAR(20) NOT NULL CHECK (rate_type IN ('flat','tiered','fixed','custom')),
    rate             NUMERIC(8,6) NOT NULL DEFAULT 0,
    fixed_amount     BIGINT,
    tiers            JSONB        NOT NULL DEFAULT '[]',
    apply_base       VARCHAR(30) NOT NULL CHECK (apply_base IN ('fee_contracted','fee_invoiced','fee_paid','gross_margin')),
    trigger_on       VARCHAR(30) NOT NULL CHECK (trigger_on IN ('contract_signed','invoice_issued','payment_received','eng_completed')),

    -- Caps & holdback
    max_amount       BIGINT,
    holdback_pct     NUMERIC(5,4) NOT NULL DEFAULT 0, -- 0.20 = 20%

    status           VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active','cancelled')),
    notes            TEXT         NOT NULL DEFAULT '',
    approved_by      UUID         REFERENCES users(id),
    approved_at      TIMESTAMPTZ,
    created_by       UUID         NOT NULL,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_eng_commissions_engagement ON engagement_commissions (engagement_id);
CREATE INDEX idx_eng_commissions_salesperson ON engagement_commissions (salesperson_id);
CREATE INDEX idx_eng_commissions_status      ON engagement_commissions (status);

CREATE TABLE commission_records (
    id                       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    engagement_commission_id UUID        NOT NULL REFERENCES engagement_commissions(id),
    engagement_id            UUID        NOT NULL REFERENCES engagements(id),
    salesperson_id           UUID        NOT NULL REFERENCES employees(id),

    -- Source trigger (one of two)
    invoice_id               UUID        REFERENCES invoices(id),
    payment_id               UUID        REFERENCES payments(id),

    -- Calculation snapshot (immutable)
    base_amount              BIGINT      NOT NULL,
    rate                     NUMERIC(8,6) NOT NULL,
    calculated_amount        BIGINT      NOT NULL,
    holdback_amount          BIGINT      NOT NULL DEFAULT 0,
    payable_amount           BIGINT      NOT NULL,

    -- Lifecycle
    status                   VARCHAR(20) NOT NULL DEFAULT 'accrued'
                                 CHECK (status IN ('accrued','approved','on_hold','paid','clawback','cancelled')),
    accrued_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_by              UUID        REFERENCES users(id),
    approved_at              TIMESTAMPTZ,
    paid_at                  TIMESTAMPTZ,
    paid_by_payroll_id       UUID,
    payout_reference         VARCHAR(200) NOT NULL DEFAULT '',

    -- Clawback
    clawback_record_id       UUID        REFERENCES commission_records(id),
    is_clawback              BOOLEAN     NOT NULL DEFAULT false,
    clawback_reason          TEXT        NOT NULL DEFAULT '',

    notes                    TEXT        NOT NULL DEFAULT '',
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Idempotency: one record per (engagement_commission, invoice)
    CONSTRAINT uq_comm_record_invoice  UNIQUE NULLS NOT DISTINCT (engagement_commission_id, invoice_id),
    -- Idempotency: one record per (engagement_commission, payment)
    CONSTRAINT uq_comm_record_payment  UNIQUE NULLS NOT DISTINCT (engagement_commission_id, payment_id)
);

CREATE INDEX idx_comm_records_eng_commission ON commission_records (engagement_commission_id);
CREATE INDEX idx_comm_records_engagement     ON commission_records (engagement_id);
CREATE INDEX idx_comm_records_salesperson    ON commission_records (salesperson_id);
CREATE INDEX idx_comm_records_status         ON commission_records (status);
CREATE INDEX idx_comm_records_invoice        ON commission_records (invoice_id) WHERE invoice_id IS NOT NULL;
CREATE INDEX idx_comm_records_payment        ON commission_records (payment_id) WHERE payment_id IS NOT NULL;
