-- Tax Advisory Module

CREATE TABLE tax_deadlines (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id               UUID NOT NULL REFERENCES clients(id),
    deadline_type           VARCHAR(50) NOT NULL CHECK (deadline_type IN ('VAT_FILING','CORPORATE_TAX','PERSONAL_TAX','COMPLIANCE_REPORTING','CUSTOM')),
    deadline_name           VARCHAR(255) NOT NULL,
    due_date                DATE NOT NULL,
    status                  VARCHAR(20) NOT NULL DEFAULT 'NOT_DUE' CHECK (status IN ('NOT_DUE','DUE_SOON','OVERDUE','COMPLETED')),
    expected_submission_date DATE,
    actual_submission_date  DATE,
    submission_status       VARCHAR(30) CHECK (submission_status IN ('PENDING','SUBMITTED','LATE','WAIVED')),
    notes                   TEXT,
    created_by              UUID NOT NULL,
    updated_by              UUID,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE advisory_records (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id       UUID NOT NULL REFERENCES clients(id),
    engagement_id   UUID REFERENCES engagements(id),
    advisory_type   VARCHAR(50) NOT NULL CHECK (advisory_type IN ('TAX_CONSULTATION','BUSINESS_ADVISORY','COMPLIANCE_REVIEW')),
    recommendation  TEXT NOT NULL,
    findings        TEXT,
    status          VARCHAR(20) NOT NULL DEFAULT 'DRAFTED' CHECK (status IN ('DRAFTED','DELIVERED','ACTED_ON')),
    delivered_date  DATE,
    client_feedback TEXT,
    created_by      UUID NOT NULL,
    updated_by      UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE advisory_files (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    advisory_id UUID NOT NULL REFERENCES advisory_records(id) ON DELETE CASCADE,
    file_name   VARCHAR(255) NOT NULL,
    file_path   TEXT NOT NULL,
    created_by  UUID NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE tax_compliance_rules (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_type   VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    check_query TEXT,
    severity    VARCHAR(10) NOT NULL DEFAULT 'MEDIUM' CHECK (severity IN ('LOW','MEDIUM','HIGH')),
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_tax_deadlines_client_id_due_date ON tax_deadlines (client_id, due_date);
CREATE INDEX idx_tax_deadlines_status ON tax_deadlines (status) WHERE status IN ('NOT_DUE','DUE_SOON');
CREATE INDEX idx_advisory_records_client_id ON advisory_records (client_id);
CREATE INDEX idx_advisory_files_advisory_id ON advisory_files (advisory_id);

-- Materialized view: compliance score per client
CREATE MATERIALIZED VIEW mv_tax_compliance_status AS
SELECT
    client_id,
    COUNT(*) AS total_deadlines,
    COUNT(*) FILTER (WHERE status = 'COMPLETED') AS completed,
    COUNT(*) FILTER (WHERE status = 'OVERDUE') AS overdue,
    COUNT(*) FILTER (WHERE status = 'DUE_SOON') AS due_soon,
    CASE
        WHEN COUNT(*) = 0 THEN 100
        ELSE ROUND(100.0 * COUNT(*) FILTER (WHERE status IN ('COMPLETED','NOT_DUE','DUE_SOON')) / COUNT(*))
    END AS compliance_score
FROM tax_deadlines
GROUP BY client_id;

CREATE UNIQUE INDEX ON mv_tax_compliance_status (client_id);
