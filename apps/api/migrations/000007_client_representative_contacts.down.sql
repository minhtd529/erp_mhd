-- Rollback Phase 1.7

DROP TABLE IF EXISTS client_contacts;

ALTER TABLE clients
    DROP COLUMN IF EXISTS representative_phone,
    DROP COLUMN IF EXISTS representative_title,
    DROP COLUMN IF EXISTS representative_name;
