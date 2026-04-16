-- Rollback Phase 1.7

ALTER TABLE client_contacts ALTER COLUMN updated_by DROP NOT NULL;
ALTER TABLE client_contacts ALTER COLUMN created_by DROP NOT NULL;

DROP TABLE IF EXISTS client_contacts;

ALTER TABLE clients
    DROP COLUMN IF EXISTS representative_phone,
    DROP COLUMN IF EXISTS representative_title,
    DROP COLUMN IF EXISTS representative_name;
