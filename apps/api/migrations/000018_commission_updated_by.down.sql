DROP INDEX IF EXISTS idx_commission_records_updated_by;
DROP INDEX IF EXISTS idx_commission_plans_updated_by;

ALTER TABLE commission_records DROP COLUMN IF EXISTS updated_by;
ALTER TABLE commission_plans DROP COLUMN IF EXISTS updated_by;
