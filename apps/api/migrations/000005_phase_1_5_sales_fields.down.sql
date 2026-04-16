-- Rollback Phase 1.5 sales fields

-- ── engagements (if exists) ───────────────────────────────────────────────────
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name   = 'engagements'
          AND column_name  = 'primary_salesperson_id'
    ) THEN
        DROP INDEX IF EXISTS idx_engagements_primary_salesperson_id;
        ALTER TABLE engagements DROP COLUMN primary_salesperson_id;
    END IF;
END;
$$;

-- ── employees ────────────────────────────────────────────────────────────────
DROP INDEX IF EXISTS idx_employees_is_salesperson;
ALTER TABLE employees
    DROP COLUMN IF EXISTS bank_account_name,
    DROP COLUMN IF EXISTS bank_account_number_enc,
    DROP COLUMN IF EXISTS sales_commission_eligible,
    DROP COLUMN IF EXISTS is_salesperson;

-- ── clients ──────────────────────────────────────────────────────────────────
DROP INDEX IF EXISTS idx_clients_referrer_id;
DROP INDEX IF EXISTS idx_clients_sales_owner_id;
ALTER TABLE clients
    DROP COLUMN IF EXISTS referrer_id,
    DROP COLUMN IF EXISTS sales_owner_id;
