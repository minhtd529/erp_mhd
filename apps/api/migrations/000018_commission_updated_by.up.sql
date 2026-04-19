-- Add updated_by audit column to commission tables
-- Tracks which user last modified the record
-- Nullable until backfilled; new updates MUST set it

ALTER TABLE commission_plans
    ADD COLUMN updated_by UUID REFERENCES users(id);

ALTER TABLE commission_records
    ADD COLUMN updated_by UUID REFERENCES users(id);

-- Partial indexes — only index populated rows to save space
CREATE INDEX idx_commission_plans_updated_by
    ON commission_plans(updated_by)
    WHERE updated_by IS NOT NULL;

CREATE INDEX idx_commission_records_updated_by
    ON commission_records(updated_by)
    WHERE updated_by IS NOT NULL;
