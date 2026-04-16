-- Phase 1.6: Bank account fields for clients
-- ALTER only — no new tables.

ALTER TABLE clients
    ADD COLUMN IF NOT EXISTS bank_name           VARCHAR(100),
    ADD COLUMN IF NOT EXISTS bank_account_number VARCHAR(50),
    ADD COLUMN IF NOT EXISTS bank_account_name   VARCHAR(200),
    ADD COLUMN IF NOT EXISTS address             VARCHAR(500) NOT NULL DEFAULT '';

ALTER TABLE clients ALTER COLUMN office_id   SET NOT NULL;
ALTER TABLE clients ALTER COLUMN created_by  SET NOT NULL;
ALTER TABLE clients ALTER COLUMN updated_by  SET NOT NULL;
