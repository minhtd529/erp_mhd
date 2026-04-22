-- Migration: 000024_hrm_professional (DOWN)
-- Reverses everything in 000024 up — strict reverse order.
-- WARNING: All certification, training course, training record, and CPE
--          requirement data will be permanently lost on rollback.
--          Verify no live data exists before running on staging/production.

-- ============================================================
-- 1. Drop tables (reverse creation order)
--    Indexes are dropped automatically with their tables.
-- ============================================================

DROP TABLE IF EXISTS cpe_requirements_by_role;
DROP TABLE IF EXISTS training_records;
DROP TABLE IF EXISTS training_courses;
DROP TABLE IF EXISTS certifications;
