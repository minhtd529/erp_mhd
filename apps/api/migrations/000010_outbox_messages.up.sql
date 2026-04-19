-- outbox_messages was created in migration 000001 with lowercase status values,
-- VARCHAR(50) aggregate_type, and columns named retry_count/error_message.
-- This migration aligns the existing table to the spec:
--   aggregate_type VARCHAR(50) → VARCHAR(100)
--   payload gains DEFAULT '{}'
--   status values: lowercase → uppercase, 'sent' → 'PROCESSED'
--   retry_count → attempts
--   error_message → last_error
--   idx_outbox_status replaced by idx_outbox_pending

ALTER TABLE outbox_messages ALTER COLUMN aggregate_type TYPE VARCHAR(100);

ALTER TABLE outbox_messages ALTER COLUMN payload SET DEFAULT '{}';

ALTER TABLE outbox_messages ALTER COLUMN status SET DEFAULT 'PENDING';
ALTER TABLE outbox_messages DROP CONSTRAINT IF EXISTS outbox_messages_status_check;
UPDATE outbox_messages SET status = 'PROCESSED' WHERE status = 'sent';
UPDATE outbox_messages SET status = UPPER(status) WHERE status IN ('pending', 'processing', 'failed');
ALTER TABLE outbox_messages ADD CONSTRAINT outbox_messages_status_check
    CHECK (status IN ('PENDING', 'PROCESSING', 'PROCESSED', 'FAILED'));

ALTER TABLE outbox_messages RENAME COLUMN retry_count TO attempts;
ALTER TABLE outbox_messages RENAME COLUMN error_message TO last_error;

DROP INDEX IF EXISTS idx_outbox_status;
CREATE INDEX IF NOT EXISTS idx_outbox_pending ON outbox_messages (created_at ASC)
    WHERE status = 'PENDING';
