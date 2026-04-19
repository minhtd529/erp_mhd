-- Migration 000012: Billing bounded context
-- invoices, invoice_line_items, payments, billing_memos

CREATE TABLE invoices (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number   VARCHAR(50) NOT NULL,
    client_id        UUID        NOT NULL REFERENCES clients(id),
    engagement_id    UUID        REFERENCES engagements(id),
    invoice_type     VARCHAR(30) NOT NULL
                     CHECK (invoice_type IN ('TIME_AND_MATERIAL','FIXED_FEE','RETAINER','CREDIT_NOTE')),
    status           VARCHAR(20) NOT NULL DEFAULT 'DRAFT'
                     CHECK (status IN ('DRAFT','SENT','CONFIRMED','ISSUED','PAID','CANCELLED')),
    issue_date       DATE,
    due_date         DATE,
    total_amount     NUMERIC(18,2) NOT NULL DEFAULT 0,
    tax_amount       NUMERIC(18,2) NOT NULL DEFAULT 0,
    snapshot_data    JSONB        NOT NULL DEFAULT '{}',
    notes            TEXT,
    is_deleted       BOOLEAN     NOT NULL DEFAULT false,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by       UUID        NOT NULL,
    updated_by       UUID        NOT NULL
);

CREATE UNIQUE INDEX uidx_invoices_invoice_number ON invoices (invoice_number) WHERE is_deleted = false;
CREATE INDEX idx_invoices_client_id_status    ON invoices (client_id, status);
CREATE INDEX idx_invoices_engagement_id       ON invoices (engagement_id) WHERE engagement_id IS NOT NULL;
CREATE INDEX idx_invoices_status_created      ON invoices (status, created_at DESC);

CREATE TABLE invoice_line_items (
    id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id    UUID          NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    description   TEXT          NOT NULL,
    quantity      NUMERIC(10,4) NOT NULL DEFAULT 1,
    unit_price    NUMERIC(18,2) NOT NULL DEFAULT 0,
    tax_amount    NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_amount  NUMERIC(18,2) NOT NULL DEFAULT 0,
    source_type   VARCHAR(30)   NOT NULL DEFAULT 'MANUAL'
                  CHECK (source_type IN ('ENGAGEMENT_FEE','TIMESHEET_HOURS','DIRECT_COST','MANUAL')),
    snapshot_data JSONB         NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_line_items_invoice_id ON invoice_line_items (invoice_id);

CREATE TABLE payments (
    id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id       UUID          NOT NULL REFERENCES invoices(id),
    payment_method   VARCHAR(20)   NOT NULL
                     CHECK (payment_method IN ('BANK_TRANSFER','CHEQUE','CASH','CREDIT_CARD')),
    amount           NUMERIC(18,2) NOT NULL,
    payment_date     DATE          NOT NULL,
    reference_number VARCHAR(100),
    status           VARCHAR(20)   NOT NULL DEFAULT 'RECORDED'
                     CHECK (status IN ('RECORDED','CLEARED','DISPUTED','REVERSED')),
    notes            TEXT,
    recorded_by      UUID          NOT NULL,
    recorded_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    cleared_at       TIMESTAMPTZ,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_invoice_id    ON payments (invoice_id);
CREATE INDEX idx_payments_payment_date  ON payments (payment_date);
CREATE INDEX idx_payments_status        ON payments (status);

CREATE TABLE billing_memos (
    id                  UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    related_invoice_id  UUID          REFERENCES invoices(id),
    memo_type           VARCHAR(20)   NOT NULL
                        CHECK (memo_type IN ('CREDIT_NOTE','ADJUSTMENT')),
    memo_number         VARCHAR(50)   NOT NULL,
    amount              NUMERIC(18,2) NOT NULL,
    reason              TEXT          NOT NULL,
    status              VARCHAR(20)   NOT NULL DEFAULT 'DRAFT'
                        CHECK (status IN ('DRAFT','ISSUED','REVERSED')),
    is_deleted          BOOLEAN       NOT NULL DEFAULT false,
    created_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    created_by          UUID          NOT NULL,
    updated_by          UUID          NOT NULL
);

CREATE UNIQUE INDEX uidx_billing_memos_number ON billing_memos (memo_number) WHERE is_deleted = false;
CREATE INDEX idx_billing_memos_invoice_id ON billing_memos (related_invoice_id) WHERE related_invoice_id IS NOT NULL;
