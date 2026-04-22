-- Migration: 000022_hrm_provisioning (DOWN)
-- Reverses 000022 up in strict reverse order (offboarding first, then provisioning).
-- Indexes are dropped automatically with their tables in PostgreSQL.
-- WARNING: All provisioning requests and offboarding checklists will be
--          permanently deleted. Never rollback on production with active data.

-- ============================================================
-- 1. DROP offboarding_checklists
--    (no FK references from user_provisioning_requests — safe to drop first)
-- ============================================================

DROP TABLE IF EXISTS offboarding_checklists;

-- ============================================================
-- 2. DROP user_provisioning_requests
-- ============================================================

DROP TABLE IF EXISTS user_provisioning_requests;
