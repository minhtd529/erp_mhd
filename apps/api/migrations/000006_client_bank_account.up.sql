-- Phase 1.6: Bank account fields for clients
-- ALTER only — no new tables.

ALTER TABLE clients
    ADD COLUMN IF NOT EXISTS bank_name           VARCHAR(100),
    ADD COLUMN IF NOT EXISTS bank_account_number VARCHAR(50),
    ADD COLUMN IF NOT EXISTS bank_account_name   VARCHAR(200);
