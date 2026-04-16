-- Phase 1.5: Sales Owner & Salesperson fields
-- ALTER only — no new tables, no business logic.

-- ── clients ──────────────────────────────────────────────────────────────────
ALTER TABLE clients
    ADD COLUMN IF NOT EXISTS sales_owner_id UUID REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS referrer_id    UUID REFERENCES users(id);

CREATE INDEX IF NOT EXISTS idx_clients_sales_owner_id ON clients(sales_owner_id) WHERE sales_owner_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_clients_referrer_id    ON clients(referrer_id)    WHERE referrer_id    IS NOT NULL;

-- ── employees ────────────────────────────────────────────────────────────────
ALTER TABLE employees
    ADD COLUMN IF NOT EXISTS is_salesperson              BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS sales_commission_eligible   BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS bank_account_number_enc     TEXT,
    ADD COLUMN IF NOT EXISTS bank_account_name           TEXT;

CREATE INDEX IF NOT EXISTS idx_employees_is_salesperson ON employees(is_salesperson) WHERE is_salesperson = true;

-- ── engagements ───────────────────────────────────────────────────────────────
-- engagements table is created in a later migration (Phase 2); guard with a
-- DO block so the migration is idempotent if run out-of-order in testing.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'engagements'
    ) THEN
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public'
              AND table_name   = 'engagements'
              AND column_name  = 'primary_salesperson_id'
        ) THEN
            ALTER TABLE engagements
                ADD COLUMN primary_salesperson_id UUID REFERENCES employees(id);
            CREATE INDEX idx_engagements_primary_salesperson_id
                ON engagements(primary_salesperson_id)
                WHERE primary_salesperson_id IS NOT NULL;
        END IF;
    END IF;
END;
$$;
