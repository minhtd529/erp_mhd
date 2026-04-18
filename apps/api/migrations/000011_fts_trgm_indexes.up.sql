-- Full-text search upgrade: pg_trgm extension + GIN indexes.
-- ILIKE '%term%' queries automatically use GIN trgm indexes,
-- giving sub-millisecond search on large tables without schema changes.

CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- clients: search across business_name + english_name + tax_code + representative_name
CREATE INDEX IF NOT EXISTS idx_clients_search_trgm
    ON clients USING GIN (
        (COALESCE(business_name, '') || ' ' ||
         COALESCE(english_name,  '') || ' ' ||
         COALESCE(tax_code,      '') || ' ' ||
         COALESCE(representative_name, ''))
        gin_trgm_ops
    );

-- engagements: search on description
CREATE INDEX IF NOT EXISTS idx_engagements_desc_trgm
    ON engagements USING GIN (COALESCE(description, '') gin_trgm_ops);

-- employees: search across full_name + email
CREATE INDEX IF NOT EXISTS idx_employees_search_trgm
    ON employees USING GIN (
        (COALESCE(full_name, '') || ' ' || COALESCE(email, ''))
        gin_trgm_ops
    );
