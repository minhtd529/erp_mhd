-- Reverse migration 000010: restore outbox_messages to migration 000001 schema.
-- Note: narrowing aggregate_type back to VARCHAR(50) will fail if any row has a value longer
-- than 50 characters; truncate/clean data before running this down migration if needed.

DROP INDEX IF EXISTS idx_outbox_pending;

ALTER TABLE outbox_messages RENAME COLUMN last_error TO error_message;
ALTER TABLE outbox_messages RENAME COLUMN attempts TO retry_count;

ALTER TABLE outbox_messages DROP CONSTRAINT IF EXISTS outbox_messages_status_check;
UPDATE outbox_messages SET status = 'sent' WHERE status = 'PROCESSED';
UPDATE outbox_messages SET status = LOWER(status) WHERE status IN ('PENDING', 'PROCESSING', 'FAILED');
ALTER TABLE outbox_messages ADD CONSTRAINT outbox_messages_status_check
    CHECK (status IN ('pending', 'processing', 'sent', 'failed'));
ALTER TABLE outbox_messages ALTER COLUMN status SET DEFAULT 'pending';

ALTER TABLE outbox_messages ALTER COLUMN payload DROP DEFAULT;

ALTER TABLE outbox_messages ALTER COLUMN aggregate_type TYPE VARCHAR(50);

CREATE INDEX IF NOT EXISTS idx_outbox_status ON outbox_messages (status, created_at)
    WHERE status IN ('pending', 'failed');
