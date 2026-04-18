-- outbox_messages: transactional outbox for domain events.
-- Writers insert in the same DB transaction as the domain mutation.
-- The worker polls PENDING rows and dispatches to Asynq.
CREATE TABLE outbox_messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type  VARCHAR(100)  NOT NULL,
    aggregate_id    UUID          NOT NULL,
    event_type      VARCHAR(100)  NOT NULL,
    payload         JSONB         NOT NULL DEFAULT '{}',
    status          VARCHAR(20)   NOT NULL DEFAULT 'PENDING'
                        CHECK (status IN ('PENDING', 'PROCESSING', 'PROCESSED', 'FAILED')),
    attempts        INT           NOT NULL DEFAULT 0,
    last_error      TEXT,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    processed_at    TIMESTAMPTZ
);

-- Poller only scans rows with status=PENDING; partial index keeps it fast.
CREATE INDEX idx_outbox_pending ON outbox_messages (created_at ASC)
    WHERE status = 'PENDING';
