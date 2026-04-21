-- Preview seed for HRM Organization (Sprint 1 Phase 4 verification)
-- SPEC: HRM_SPEC_v1.4.md §3, §22.1 (bootstrap data)
-- NOTE: This is a DEV seed, not migration 000027. For dev E2E testing only.
-- Safe to re-run (ON CONFLICT DO UPDATE).

-- ============================================================
-- Branches (2 rows)
-- ============================================================

INSERT INTO branches (
    id, code, name, address, phone, city, tax_code,
    is_head_office, established_date,
    is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    'HO',
    'Trụ sở chính Hà Nội',
    'Tầng 10, Tòa nhà ABC, 123 Đường Nguyễn Trãi, Quận Thanh Xuân, Hà Nội',
    '024-3456-7890',
    'Hà Nội',
    '0123456789',
    TRUE,
    '2015-06-15',
    TRUE,
    NOW(),
    NOW()
) ON CONFLICT (code) DO UPDATE SET
    name              = EXCLUDED.name,
    address           = EXCLUDED.address,
    phone             = EXCLUDED.phone,
    city              = EXCLUDED.city,
    tax_code          = EXCLUDED.tax_code,
    is_head_office    = EXCLUDED.is_head_office,
    established_date  = EXCLUDED.established_date,
    is_active         = EXCLUDED.is_active,
    updated_at        = NOW();

INSERT INTO branches (
    id, code, name, address, phone, city, tax_code,
    is_head_office, established_date,
    is_active, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    'HCM',
    'Chi nhánh Hồ Chí Minh',
    'Tầng 5, Tòa nhà XYZ, 456 Đường Nguyễn Huệ, Quận 1, TP. Hồ Chí Minh',
    '028-9876-5432',
    'TP. Hồ Chí Minh',
    '0123456789-001',
    FALSE,
    '2019-03-20',
    TRUE,
    NOW(),
    NOW()
) ON CONFLICT (code) DO UPDATE SET
    name              = EXCLUDED.name,
    address           = EXCLUDED.address,
    phone             = EXCLUDED.phone,
    city              = EXCLUDED.city,
    tax_code          = EXCLUDED.tax_code,
    is_head_office    = EXCLUDED.is_head_office,
    established_date  = EXCLUDED.established_date,
    is_active         = EXCLUDED.is_active,
    updated_at        = NOW();

-- ============================================================
-- Departments (5 rows)
-- ============================================================

INSERT INTO departments (
    id, code, name, description, dept_type,
    is_active, created_at, updated_at
) VALUES
    (gen_random_uuid(), 'AUDIT', 'Phòng Kiểm toán',          'Thực hiện dịch vụ kiểm toán báo cáo tài chính',      'CORE',    TRUE, NOW(), NOW()),
    (gen_random_uuid(), 'TAX',   'Phòng Thuế',                'Tư vấn thuế và lập báo cáo thuế',                   'CORE',    TRUE, NOW(), NOW()),
    (gen_random_uuid(), 'HR',    'Phòng Nhân sự',             'Quản lý nhân sự và tuyển dụng',                     'SUPPORT', TRUE, NOW(), NOW()),
    (gen_random_uuid(), 'FIN',   'Phòng Tài chính',           'Quản lý tài chính nội bộ và kế toán công ty',       'SUPPORT', TRUE, NOW(), NOW()),
    (gen_random_uuid(), 'IT',    'Phòng Công nghệ thông tin', 'Hỗ trợ CNTT và phát triển hệ thống nội bộ',         'SUPPORT', TRUE, NOW(), NOW())
ON CONFLICT (code) DO UPDATE SET
    name        = EXCLUDED.name,
    description = EXCLUDED.description,
    dept_type   = EXCLUDED.dept_type,
    is_active   = EXCLUDED.is_active,
    updated_at  = NOW();

-- ============================================================
-- Branch × Department Matrix (7 rows)
-- HO  → all 5 departments (AUDIT, TAX, HR, FIN, IT)
-- HCM → CORE only       (AUDIT, TAX)
-- Per SPEC §3.3
-- ============================================================

INSERT INTO branch_departments (branch_id, department_id, is_active, created_at)
SELECT b.id, d.id, TRUE, NOW()
FROM branches b
CROSS JOIN departments d
WHERE (b.code = 'HO'  AND d.code IN ('AUDIT', 'TAX', 'HR', 'FIN', 'IT'))
   OR (b.code = 'HCM' AND d.code IN ('AUDIT', 'TAX'))
ON CONFLICT (branch_id, department_id) DO UPDATE SET
    is_active = EXCLUDED.is_active;

-- ============================================================
-- Verification queries
-- ============================================================

SELECT 'branches' AS entity, COUNT(*)::text AS count FROM branches;
SELECT 'departments' AS entity, COUNT(*)::text AS count FROM departments WHERE is_deleted = FALSE;
SELECT 'branch_departments' AS entity, COUNT(*)::text AS count FROM branch_departments WHERE is_active = TRUE;

SELECT b.code AS branch, d.code AS dept, d.dept_type
FROM branch_departments bd
JOIN branches b ON bd.branch_id = b.id
JOIN departments d ON bd.department_id = d.id
WHERE bd.is_active = TRUE
ORDER BY b.code, d.code;
