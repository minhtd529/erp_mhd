-- Migration 000025: user notification inbox for HRM alerts
-- Types: CERT_EXPIRY, CPE_DEADLINE, PROVISIONING_EXPIRED, CONTRACT_EXPIRY

CREATE TABLE notifications (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        VARCHAR(50) NOT NULL,
    title       TEXT        NOT NULL,
    body        TEXT        NOT NULL,
    data        JSONB       NOT NULL DEFAULT '{}',
    source_ref  TEXT        NOT NULL DEFAULT '',
    is_read     BOOLEAN     NOT NULL DEFAULT false,
    read_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Fast lookup: all notifications for a user sorted newest-first.
CREATE INDEX idx_notifications_user_created ON notifications (user_id, created_at DESC);

-- Deduplication: prevent the same alert (identified by source_ref) from being
-- inserted more than once per user. ON CONFLICT DO NOTHING is used by the job.
CREATE UNIQUE INDEX idx_notifications_source ON notifications (user_id, source_ref)
    WHERE source_ref != '';
