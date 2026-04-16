-- Rollback Phase 1.6 client bank account fields

ALTER TABLE clients
    DROP COLUMN IF EXISTS bank_account_name,
    DROP COLUMN IF EXISTS bank_account_number,
    DROP COLUMN IF EXISTS bank_name;
