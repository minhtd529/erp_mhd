-- Rollback Phase 1.6 client bank account fields

ALTER TABLE clients ALTER COLUMN updated_by  DROP NOT NULL;
ALTER TABLE clients ALTER COLUMN created_by  DROP NOT NULL;
ALTER TABLE clients ALTER COLUMN office_id   DROP NOT NULL;

ALTER TABLE clients
    DROP COLUMN IF EXISTS address,
    DROP COLUMN IF EXISTS bank_account_name,
    DROP COLUMN IF EXISTS bank_account_number,
    DROP COLUMN IF EXISTS bank_name;
