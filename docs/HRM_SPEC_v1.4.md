# HRM Module Specification v1.4
## ERP System — Audit Firm (MDH)

**Version:** 1.4  
**Date:** 2026-04-20  
**Status:** Draft — Pending Review  
**Author:** minhtd529@gmail.com  

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Scope & Goals](#2-scope--goals)
3. [Organization Structure](#3-organization-structure)
4. [Employee Entity](#4-employee-entity)
5. [BHXH & Tax TNCN](#5-bhxh--tax-tncn)
6. [Professional Development](#6-professional-development)
7. [Time & Leave](#7-time--leave)
8. [User Provisioning Workflow](#8-user-provisioning-workflow)
9. [Lifecycle Events](#9-lifecycle-events)
10. [Expense Claims](#10-expense-claims)
11. [Database Schema](#11-database-schema)
12. [Migration Plan](#12-migration-plan)
13. [API Endpoints Catalog](#13-api-endpoints-catalog)
14. [UI Pages](#14-ui-pages)
15. [Permission Matrix](#15-permission-matrix)
16. [Notifications & Alerts](#16-notifications--alerts)
17. [Audit Log Events](#17-audit-log-events)
18. [Security & Privacy](#18-security--privacy)
19. [Reporting](#19-reporting)
20. [Testing Strategy](#20-testing-strategy)
21. [Roadmap — 5 Sprints](#21-roadmap--5-sprints)
22. [Bootstrap Initial Setup](#22-bootstrap-initial-setup)

---

## 1. Executive Summary

### 1.1 Overview

Hệ thống HRM (Human Resource Management) là module quản lý nhân sự tích hợp trong nền tảng ERP dành riêng cho công ty kiểm toán MDH. Module này quản lý toàn bộ vòng đời nhân viên từ tuyển dụng đến nghỉ việc, bao gồm: quản lý tổ chức, hồ sơ nhân viên, BHXH & thuế TNCN, phát triển chuyên môn, chấm công & nghỉ phép, quy trình cấp tài khoản, sự kiện vòng đời, và thanh toán chi phí.

Công ty kiểm toán có những yêu cầu đặc thù so với HRM thông thường:
- **Độc lập kiểm toán (Independence):** Nhân viên phải khai báo độc lập hàng năm và theo từng hợp đồng kiểm toán
- **CPE (Continuing Professional Education):** VACPA yêu cầu 40 giờ/năm cho CPA
- **Chứng chỉ nghề nghiệp:** CPA Việt Nam, ACCA, CFA, CIA — theo dõi hạn hàng năm
- **Ma trận chi nhánh-phòng ban:** Nhân viên có thể kiêm nhiệm (concurrent assignment) ở nhiều phòng
- **Phân quyền theo chi nhánh:** Head of Branch HCM chỉ thấy dữ liệu HCM

### 1.2 Scope Summary

| Hạng mục | Phase 1 (v1.4) | Deferred |
|---|---|---|
| Quản lý tổ chức | ✅ 2 chi nhánh, 5 phòng ban | Thêm chi nhánh mới |
| Hồ sơ nhân viên | ✅ 40+ trường, mã hóa PII | Ảnh chân dung |
| BHXH & thuế TNCN | ✅ Đăng ký, người phụ thuộc, cấu hình tỷ lệ | Tích hợp cổng BHXH |
| Chứng chỉ & CPE | ✅ Tracking, cảnh báo hết hạn | Tích hợp VACPA portal |
| Đào tạo | ✅ Khóa học, hồ sơ, CPE hours | LMS integration |
| Đánh giá hiệu suất | ✅ KPI period, peer review | 360 tự động hóa |
| Khai báo độc lập | ✅ Annual + per-engagement | Conflict screening API |
| Nghỉ phép & OT | ✅ 8 loại phép, OT cap 300h | Tích hợp timekeeping HW |
| Cấp tài khoản | ✅ Workflow 2 cấp (HCM) | SSO / AD sync |
| Onboarding/Offboarding | ✅ Checklist JSONB | Tích hợp IT ticketing |
| Hợp đồng lao động | ✅ Lịch sử, cảnh báo | Chữ ký điện tử |
| Chi phí | ✅ Claims + items + approval | Tích hợp thẻ công ty |
| Lương | ✅ Lịch sử, payroll snapshot | Tích hợp bank transfer |
| Báo cáo HRM | ✅ 10 báo cáo chuẩn | Custom report builder |

### 1.3 Technical Scope

- **22 bảng mới/thay đổi** trong database PostgreSQL
- **8 migrations** (000019 → 000026)
- **80+ API endpoints** RESTful JSON
- **Mã hóa AES-256-GCM** cho các trường PII nhạy cảm
- **Ước tính:** 4–6 tuần (5 sprints)

### 1.4 Key Design Decisions

| Quyết định | Lý do |
|---|---|
| PostgreSQL triggers cho mã NV | Đảm bảo tính tuần tự không bị race condition |
| JSONB cho offboarding checklist | Linh hoạt, không cần schema migration khi thêm bước |
| Separate table cho salary history | Không ghi đè — audit trail đầy đủ |
| AES-256-GCM tại application layer | Key rotation dễ hơn transparent encryption |
| Branch-scoped views cho HoB | Row-level logic trong API, không dùng RLS (đơn giản hơn) |

---

## 2. Scope & Goals

### 2.1 Phase 1 Goals

1. **Hồ sơ nhân viên đầy đủ:** Tất cả thông tin cần thiết cho audit firm được lưu trữ an toàn
2. **Tuân thủ pháp luật lao động VN:** BHXH, BHYT, thuế TNCN, nghỉ phép theo Bộ luật Lao động 2019
3. **Quản lý chứng chỉ nghề nghiệp:** Không bao giờ để CPA hết hạn mà không biết
4. **Khai báo độc lập:** Đảm bảo compliance với chuẩn mực kiểm toán VSA/ISA
5. **Self-service cho nhân viên:** Nhân viên có thể xem hồ sơ, nộp đơn nghỉ phép, khai báo chi phí
6. **Quy trình phê duyệt rõ ràng:** Mọi thay đổi quan trọng đều có approval trail

### 2.2 What's In Phase 1

```
✅ Quản lý tổ chức (branches, departments, matrix)
✅ Hồ sơ nhân viên mở rộng (40+ fields)
✅ BHXH, BHYT, Thuế TNCN
✅ Người phụ thuộc (dependents)
✅ Hợp đồng lao động (contracts)
✅ Lịch sử lương (salary history)
✅ Chứng chỉ nghề nghiệp (certifications)
✅ Đào tạo & CPE (training, CPE hours)
✅ Đánh giá hiệu suất (performance reviews)
✅ Peer review theo engagement
✅ Khai báo độc lập (independence declarations)
✅ Ngày lễ quốc gia (holidays, seeded 2026-2030)
✅ Cân đối nghỉ phép (leave balances)
✅ Đơn nghỉ phép (leave requests)
✅ Theo dõi OT (overtime requests, 300h annual cap)
✅ Quy trình cấp tài khoản (user provisioning)
✅ Offboarding checklist
✅ Chi phí (expense claims + items)
✅ 10 báo cáo HRM chuẩn
```

### 2.3 Deferred to Later Phases

```
⏸ Phase 2: Tích hợp cổng BHXH điện tử
⏸ Phase 2: Chữ ký điện tử hợp đồng lao động
⏸ Phase 3: Tích hợp phần mềm chấm công phần cứng
⏸ Phase 3: Tính lương tự động (payroll engine)
⏸ Phase 4: SSO / Active Directory sync
⏸ Phase 4: Tích hợp LMS đào tạo trực tuyến
⏸ Phase 4: Custom report builder
⏸ Phase 5: Tích hợp ngân hàng (chuyển lương)
```

### 2.4 Success Criteria

| Tiêu chí | Đo lường |
|---|---|
| Hồ sơ nhân viên đầy đủ | 100% nhân viên active có đủ 6 nhóm field |
| CPE compliance | Không có CPA nào quá hạn >30 ngày mà không có cảnh báo |
| Khai báo độc lập | 100% nhân viên tham gia engagement đã khai báo trước khi bắt đầu |
| Đơn nghỉ phép | Thời gian phê duyệt < 24 giờ (working hours) |
| Chi phí | Thời gian hoàn ứng < 5 ngày làm việc |
| API response time | p95 < 500ms |
| PII encryption | 100% CCCD, MST, BHXH, tài khoản ngân hàng được mã hóa |

---

## 3. Organization Structure

### 3.1 Chi Nhánh (Branches)

Công ty có 2 chi nhánh trong Phase 1:

| Field | HO | HCM |
|---|---|---|
| `code` | `HO` | `HCM` |
| `name` | Trụ sở chính (Hà Nội) | Chi nhánh TP.HCM |
| `is_head_office` | `true` | `false` |
| `city` | Hà Nội | TP. Hồ Chí Minh |
| `address` | (địa chỉ HO) | (địa chỉ HCM) |
| `phone` | (số HO) | (số HCM) |
| `established_date` | (ngày QĐ) | (ngày QĐ) |

**Quy tắc:**
- Chỉ có 1 chi nhánh `is_head_office = true` tại một thời điểm
- Chi nhánh HO chứa toàn bộ Ban Giám đốc (Chairman, CEO...)
- Mỗi chi nhánh có 1 `head_of_branch_user_id` (FK → users)

### 3.2 Phòng Ban (Departments)

5 phòng ban trong Phase 1:

| Code | Tên | Loại | Chi nhánh |
|---|---|---|---|
| `AUDIT` | Phòng Kiểm toán | Core | HO + HCM |
| `TAX` | Phòng Thuế | Core | HO + HCM |
| `HR` | Phòng Nhân sự | Support | HO only |
| `FIN` | Phòng Tài chính — Kế toán | Support | HO only |
| `IT` | Phòng Công nghệ Thông tin | Support | HO only |

**Phòng ban mở rộng trong tương lai:** LEGAL, CONSULTING, ADVISORY (Phase 3+)

### 3.3 Ma Trận Chi Nhánh — Phòng Ban (branch_departments)

Bảng junction `branch_departments` thể hiện phòng nào hoạt động ở chi nhánh nào:

```
HO:  AUDIT, TAX, HR, FIN, IT     (5 phòng)
HCM: AUDIT, TAX                  (2 phòng)
```

Một nhân viên thuộc về 1 `(branch_id, department_id)` chính. Kiêm nhiệm (concurrent) được theo dõi qua bảng `employee_concurrent_assignments` (Phase 2 — deferred).

### 3.4 Cơ Cấu Phân Quyền Tổ Chức

```
Chủ tịch HĐQT (Chairman)
    └── Tổng Giám đốc (CEO)
            ├── Trưởng Chi nhánh HCM (Head of Branch HCM)
            │       ├── Partner(s) HCM
            │       ├── Kiểm toán viên Senior HCM
            │       └── Kiểm toán viên Junior HCM
            ├── Partner(s) HO
            ├── Phòng HR (HR Manager → HR Staff)
            ├── Phòng FIN (Accountant)
            └── Phòng IT (IT Staff)
```

### 3.5 Roles (11 roles, all Phase 1 active)

| Role Code | Tên | Phase | Phạm vi |
|---|---|---|---|
| `SUPER_ADMIN` | Quản trị viên Hệ thống | 1 | System |
| `CHAIRMAN` | Chủ tịch HĐQT | 1 | All branches |
| `CEO` | Tổng Giám đốc | 1 | All branches |
| `HR_MANAGER` | Trưởng phòng Nhân sự | 1 | All branches (read) |
| `HR_STAFF` | Nhân viên Nhân sự | 1 | Assigned branch |
| `HEAD_OF_BRANCH` | Trưởng Chi nhánh | 1 | Own branch |
| `PARTNER` | Partner Kiểm toán | 1 | Own dept/branch |
| `AUDIT_MANAGER` | Quản lý Kiểm toán | 1 | Own dept + team |
| `SENIOR_AUDITOR` | Kiểm toán viên Cao cấp | 1 | Own dept |
| `JUNIOR_AUDITOR` | Kiểm toán viên | 1 | Self only |
| `ACCOUNTANT` | Kế toán | 1 | Self only |

### 3.6 Employee Grades (Bậc Nhân Viên)

8 bậc được dùng xuyên suốt hệ thống:

| Grade | Tên | Typical Roles |
|---|---|---|
| `EXECUTIVE` | Ban điều hành | CHAIRMAN, CEO |
| `PARTNER` | Partner | PARTNER |
| `DIRECTOR` | Giám đốc / Trưởng chi nhánh | HEAD_OF_BRANCH |
| `MANAGER` | Quản lý | HR_MANAGER, AUDIT_MANAGER |
| `SENIOR` | Cao cấp | SENIOR_AUDITOR |
| `JUNIOR` | Nhân viên | JUNIOR_AUDITOR, ACCOUNTANT, HR_STAFF |
| `INTERN` | Thực tập sinh | — |
| `SUPPORT` | Hỗ trợ | Lễ tân, tạp vụ |

---

## 4. Employee Entity

### 4.1 Overview

Bảng `employees` là trung tâm của module HRM. Trong Phase 1, ~40 cột mới được thêm vào bảng hiện tại qua migration 000020.

**Mã nhân viên:** Format `NV{YY}-{SEQ4}` ví dụ `NV26-0001`
- Generated bằng PostgreSQL trigger `trg_employees_set_code`
- `YY` = 2 chữ số năm tuyển dụng
- `SEQ4` = số thứ tự 4 chữ số trong năm đó, reset mỗi năm

### 4.2 Nhóm Field 1: Thông Tin Cơ Bản (Basic)

| Column | Type | Mô tả |
|---|---|---|
| `id` | UUID PK | Generated |
| `employee_code` | VARCHAR(12) UNIQUE | NV26-0001, auto-generated |
| `user_id` | UUID FK → users | Tài khoản đăng nhập (nullable — trước khi có account) |
| `branch_id` | UUID FK → branches | Chi nhánh chính |
| `department_id` | UUID FK → departments | Phòng ban chính |
| `grade` | VARCHAR(20) | EXECUTIVE/PARTNER/DIRECTOR/MANAGER/SENIOR/JUNIOR/INTERN/SUPPORT |
| `position_title` | VARCHAR(200) | Chức danh (ví dụ: "Kiểm toán viên Cao cấp") |
| `manager_id` | UUID FK → employees | Quản lý trực tiếp |
| `status` | VARCHAR(20) | ACTIVE/INACTIVE/ON_LEAVE/TERMINATED |
| `employment_type` | VARCHAR(20) | FULL_TIME/PART_TIME/INTERN |
| `hired_date` | DATE | Ngày ký hợp đồng đầu tiên |
| `probation_end_date` | DATE | Ngày kết thúc thử việc |
| `termination_date` | DATE | Ngày nghỉ việc (nullable) |
| `termination_reason` | TEXT | Lý do nghỉ việc |

### 4.3 Nhóm Field 2: Thông Tin Cá Nhân / PII

| Column | Type | Mô tả |
|---|---|---|
| `full_name` | VARCHAR(200) NOT NULL | Họ và tên đầy đủ |
| `display_name` | VARCHAR(100) | Tên hiển thị ngắn |
| `gender` | VARCHAR(10) | MALE/FEMALE/OTHER |
| `date_of_birth` | DATE | Ngày sinh |
| `place_of_birth` | VARCHAR(200) | Nơi sinh |
| `nationality` | VARCHAR(50) | DEFAULT 'Vietnamese' |
| `ethnicity` | VARCHAR(50) | Dân tộc |
| `personal_email` | VARCHAR(200) | Email cá nhân |
| `personal_phone` | VARCHAR(20) | Điện thoại cá nhân |
| `work_phone` | VARCHAR(20) | Điện thoại công ty |
| `current_address` | TEXT | Địa chỉ thường trú hiện tại |
| `permanent_address` | TEXT | Địa chỉ hộ khẩu |
| `cccd_encrypted` | TEXT | CCCD/CMND — AES-256-GCM |
| `cccd_issued_date` | DATE | Ngày cấp CCCD |
| `cccd_issued_place` | VARCHAR(200) | Nơi cấp CCCD |
| `passport_number` | VARCHAR(50) | Số hộ chiếu (nullable) |
| `passport_expiry` | DATE | Hạn hộ chiếu |

### 4.4 Nhóm Field 3: Thông Tin Tuyển Dụng (Employment)

| Column | Type | Mô tả |
|---|---|---|
| `hired_source` | VARCHAR(50) | REFERRAL/PORTAL/DIRECT/AGENCY |
| `referrer_employee_id` | UUID FK → employees | Người giới thiệu (nếu REFERRAL) |
| `probation_salary_pct` | NUMERIC(5,2) | % lương thử việc (default 85.00) |
| `current_contract_id` | UUID FK → employment_contracts | Hợp đồng hiện tại |
| `work_location` | VARCHAR(20) | OFFICE/REMOTE/HYBRID |
| `remote_days_per_week` | SMALLINT | Số ngày WFH / tuần |

### 4.5 Nhóm Field 4: Trình Độ Học Vấn (Qualifications)

| Column | Type | Mô tả |
|---|---|---|
| `education_level` | VARCHAR(30) | BACHELOR/MASTER/PHD/COLLEGE/OTHER |
| `education_major` | VARCHAR(200) | Chuyên ngành đại học |
| `education_school` | VARCHAR(200) | Trường đại học |
| `education_graduation_year` | SMALLINT | Năm tốt nghiệp |
| `vn_cpa_number` | VARCHAR(50) | Số CPA Việt Nam (nullable) |
| `vn_cpa_issued_date` | DATE | Ngày cấp CPA |
| `vn_cpa_expiry_date` | DATE | Ngày hết hạn CPA (nếu có) |
| `practicing_certificate_number` | VARCHAR(50) | Số chứng chỉ hành nghề kiểm toán |
| `practicing_certificate_expiry` | DATE | Hạn chứng chỉ hành nghề |

### 4.6 Nhóm Field 5: Lương & Ngân Hàng (Salary/Bank) — Sensitive

> Các field này chỉ HR_MANAGER, CEO, CHAIRMAN được xem. Được mask trong API response theo role.

| Column | Type | Mô tả |
|---|---|---|
| `base_salary` | NUMERIC(15,2) | Lương cơ bản hiện tại (VNĐ) |
| `salary_currency` | VARCHAR(3) | DEFAULT 'VND' |
| `salary_effective_date` | DATE | Ngày áp dụng lương hiện tại |
| `bank_account_encrypted` | TEXT | Số tài khoản ngân hàng — AES-256-GCM |
| `bank_name` | VARCHAR(100) | Tên ngân hàng |
| `bank_branch` | VARCHAR(200) | Chi nhánh ngân hàng |
| `mst_ca_nhan_encrypted` | TEXT | Mã số thuế cá nhân — AES-256-GCM |

### 4.7 Nhóm Field 6: Hoa Hồng & Kinh Doanh (Sales/Commission)

| Column | Type | Mô tả |
|---|---|---|
| `commission_rate` | NUMERIC(5,2) | Tỷ lệ hoa hồng % (nullable) |
| `commission_type` | VARCHAR(20) | FIXED/TIERED/NONE |
| `sales_target_yearly` | NUMERIC(15,2) | Chỉ tiêu doanh thu năm (VNĐ) |
| `biz_dev_region` | VARCHAR(100) | Vùng phát triển kinh doanh |

### 4.8 Mã Nhân Viên — Trigger Logic

```sql
-- Trigger tự động sinh mã NV{YY}-{SEQ}
-- Chạy BEFORE INSERT trên bảng employees
-- Logic:
--   1. Lấy YY từ hired_date (hoặc CURRENT_DATE nếu null)
--   2. Đếm số nhân viên có mã bắt đầu bằng 'NV{YY}-' + 1
--   3. Format: LPAD(seq::text, 4, '0')
--   4. Gán NEW.employee_code = 'NV' || YY || '-' || seq_padded
-- Đảm bảo UNIQUE bằng constraint + retry trong trigger
```

### 4.9 Employment Contracts (Hợp Đồng Lao Động)

Bảng `employment_contracts` lưu lịch sử hợp đồng:

| Column | Type | Mô tả |
|---|---|---|
| `id` | UUID PK | |
| `employee_id` | UUID FK | |
| `contract_number` | VARCHAR(50) | Số hợp đồng |
| `contract_type` | VARCHAR(20) | PROBATION/DEFINITE_TERM/INDEFINITE/INTERN |
| `start_date` | DATE NOT NULL | |
| `end_date` | DATE | NULL = hợp đồng không thời hạn |
| `signed_date` | DATE | |
| `salary_at_signing` | NUMERIC(15,2) | Lương ghi trong hợp đồng |
| `position_at_signing` | VARCHAR(200) | Chức danh trong hợp đồng |
| `notes` | TEXT | |
| `document_url` | TEXT | Link tài liệu scan |
| `is_current` | BOOLEAN DEFAULT false | |
| `created_by` | UUID FK → users | |
| `created_at` | TIMESTAMPTZ | |

**Cảnh báo hợp đồng:** Hệ thống tự động tạo notification 30 ngày trước `end_date` cho HR_MANAGER và CEO.

---

## 5. BHXH & Tax TNCN

### 5.1 Tổng Quan Bảo Hiểm Xã Hội Việt Nam

Theo quy định hiện hành (2024), tỷ lệ đóng:

| Khoản | NV đóng | Công ty đóng |
|---|---|---|
| BHXH (Xã hội) | 8% | 17.5% |
| BHYT (Y tế) | 1.5% | 3% |
| BHTN (Thất nghiệp) | 1% | 1% |
| KPCĐ (Công đoàn) | 0% | 2% |
| **Tổng** | **10.5%** | **23.5%** |

> **Ghi chú KPCĐ:** Theo Luật Công đoàn 2012, công ty có từ 10 lao động trở lên phải đóng Kinh phí Công đoàn 2% trên quỹ lương đóng BHXH. Đóng cùng thời điểm với BHXH hàng tháng.

### 5.2 Bảng insurance_rate_config

Lưu lịch sử tỷ lệ đóng, cho phép thay đổi khi Nhà nước điều chỉnh:

| Column | Type | Mô tả |
|---|---|---|
| `id` | UUID PK | |
| `effective_from` | DATE NOT NULL | Ngày áp dụng |
| `effective_to` | DATE | NULL = đang áp dụng |
| `bhxh_employee_pct` | NUMERIC(5,2) | Default 8.00 |
| `bhxh_employer_pct` | NUMERIC(5,2) | Default 17.50 |
| `bhyt_employee_pct` | NUMERIC(5,2) | Default 1.50 |
| `bhyt_employer_pct` | NUMERIC(5,2) | Default 3.00 |
| `bhtn_employee_pct` | NUMERIC(5,2) | Default 1.00 |
| `bhtn_employer_pct` | NUMERIC(5,2) | Default 1.00 |
| `kpcd_employer_pct` | NUMERIC(5,2) | Default 2.00 — Kinh phí công đoàn (công ty đóng) |
| `salary_base_bhxh` | NUMERIC(15,2) | Mức lương tối thiểu vùng áp dụng |
| `max_bhxh_salary` | NUMERIC(15,2) | Mức tối đa tính BHXH (20 × lương cơ sở) |
| `notes` | TEXT | |
| `created_by` | UUID FK | |
| `created_at` | TIMESTAMPTZ | |

**Seed data 2024:**
```sql
INSERT INTO insurance_rate_config (effective_from, bhxh_employee_pct, bhxh_employer_pct,
  bhyt_employee_pct, bhyt_employer_pct, bhtn_employee_pct, bhtn_employer_pct,
  kpcd_employer_pct, salary_base_bhxh, max_bhxh_salary, notes)
VALUES ('2024-01-01', 8.00, 17.50, 1.50, 3.00, 1.00, 1.00, 2.00,
  1800000, 36000000, 'Tỷ lệ áp dụng từ 01/01/2024 theo QĐ số ...');
```

### 5.3 Thông Tin BHXH Trên Hồ Sơ Nhân Viên

Các field bổ sung vào bảng `employees`:

| Column | Type | Mô tả |
|---|---|---|
| `so_bhxh_encrypted` | TEXT | Số sổ BHXH — AES-256-GCM |
| `bhxh_registered_date` | DATE | Ngày đăng ký tham gia BHXH |
| `bhxh_province_code` | VARCHAR(10) | Mã tỉnh đăng ký BHXH |
| `bhyt_card_number` | VARCHAR(20) | Số thẻ BHYT (dạng hiển thị, không mã hóa) |
| `bhyt_expiry_date` | DATE | Ngày hết hạn thẻ BHYT |
| `bhyt_registered_hospital_code` | VARCHAR(20) | Mã cơ sở KCB ban đầu |
| `bhyt_registered_hospital_name` | VARCHAR(200) | Tên cơ sở KCB ban đầu |
| `tncn_registered` | BOOLEAN DEFAULT false | Đã đăng ký thuế TNCN chưa |

### 5.4 Người Phụ Thuộc (employee_dependents)

Bảng `employee_dependents` theo dõi người phụ thuộc để tính giảm trừ gia cảnh thuế TNCN:

| Column | Type | Mô tả |
|---|---|---|
| `id` | UUID PK | |
| `employee_id` | UUID FK → employees | |
| `full_name` | VARCHAR(200) NOT NULL | |
| `relationship` | VARCHAR(30) | SPOUSE/CHILD/PARENT/SIBLING/OTHER |
| `date_of_birth` | DATE | |
| `cccd_or_birth_cert` | VARCHAR(50) | CCCD hoặc số giấy khai sinh |
| `tax_deduction_registered` | BOOLEAN DEFAULT false | Đã đăng ký giảm trừ gia cảnh |
| `tax_deduction_from` | DATE | Từ tháng nào tính giảm trừ |
| `tax_deduction_to` | DATE | Đến tháng nào (NULL = còn hiệu lực) |
| `notes` | TEXT | |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

**Mức giảm trừ gia cảnh (2024):**
- Bản thân: 11,000,000 VNĐ/tháng
- Người phụ thuộc: 4,400,000 VNĐ/tháng/người

### 5.5 Lịch Sử Lương (employee_salary_history)

Bảng immutable — chỉ INSERT, không UPDATE/DELETE:

| Column | Type | Mô tả |
|---|---|---|
| `id` | UUID PK | |
| `employee_id` | UUID FK | |
| `effective_date` | DATE NOT NULL | Ngày áp dụng mức lương mới |
| `base_salary` | NUMERIC(15,2) NOT NULL | Lương cơ bản |
| `allowances_total` | NUMERIC(15,2) | Tổng phụ cấp |
| `salary_note` | TEXT | Lý do thay đổi (tăng lương, lên chức...) |
| `change_type` | VARCHAR(30) | INITIAL/INCREASE/DECREASE/PROMOTION/ADJUSTMENT |
| `approved_by` | UUID FK → users | Người phê duyệt |
| `created_by` | UUID FK → users | |
| `created_at` | TIMESTAMPTZ | |

---

## 6. Professional Development

### 6.1 Chứng Chỉ Nghề Nghiệp (certifications)

| Column | Type | Mô tả |
|---|---|---|
| `id` | UUID PK | |
| `employee_id` | UUID FK → employees | |
| `cert_type` | VARCHAR(30) | VN_CPA/ACCA/CFA/CIA/CISA/CPA_AUS/OTHER |
| `cert_name` | VARCHAR(200) | Tên đầy đủ |
| `cert_number` | VARCHAR(100) | Số chứng chỉ |
| `issued_date` | DATE | |
| `expiry_date` | DATE | NULL = không hết hạn |
| `issued_by` | VARCHAR(200) | Cơ quan cấp |
| `status` | VARCHAR(20) | ACTIVE/EXPIRED/SUSPENDED/SURRENDERED |
| `renewal_reminder_days` | INT DEFAULT 60 | Cảnh báo trước bao nhiêu ngày |
| `document_url` | TEXT | Link scan chứng chỉ |
| `notes` | TEXT | |
| `created_by` | UUID FK | |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

**Loại chứng chỉ phổ biến tại công ty kiểm toán:**

| Code | Tên đầy đủ | Cơ quan cấp | Gia hạn |
|---|---|---|---|
| `VN_CPA` | Chứng chỉ CPA Việt Nam | VACPA | Hàng năm (40 CPE) |
| `ACCA` | Association of Chartered Certified Accountants | ACCA Global | 40 CPD/năm |
| `CFA` | Chartered Financial Analyst | CFA Institute | 20 CPD/năm |
| `CIA` | Certified Internal Auditor | IIA | 40 CPE/năm |
| `CISA` | Certified Information Systems Auditor | ISACA | 20 CPE/năm |
| `CPA_AUS` | CPA Australia | CPA Australia | 120 CPD/3 năm |
| `OTHER` | Khác | — | — |

### 6.2 Khóa Đào Tạo (training_courses)

Catalog khóa đào tạo nội bộ và bên ngoài:

| Column | Type | Mô tả |
|---|---|---|
| `id` | UUID PK | |
| `course_code` | VARCHAR(50) UNIQUE | |
| `course_name` | VARCHAR(300) NOT NULL | |
| `provider` | VARCHAR(200) | Đơn vị đào tạo |
| `course_type` | VARCHAR(30) | INTERNAL/EXTERNAL/ONLINE/CONFERENCE |
| `category` | VARCHAR(50) | AUDIT/TAX/ACCOUNTING/SOFT_SKILL/COMPLIANCE/IT/OTHER |
| `cpe_hours` | NUMERIC(5,1) | Số giờ CPE được công nhận |
| `duration_days` | NUMERIC(5,1) | Thời lượng khóa học |
| `is_mandatory` | BOOLEAN | |
| `applicable_grades` | TEXT[] | Bậc nhân viên áp dụng |
| `description` | TEXT | |
| `is_active` | BOOLEAN DEFAULT true | |
| `created_at` | TIMESTAMPTZ | |

### 6.3 Hồ Sơ Đào Tạo (training_records)

| Column | Type | Mô tả |
|---|---|---|
| `id` | UUID PK | |
| `employee_id` | UUID FK → employees | |
| `course_id` | UUID FK → training_courses | NULL nếu khóa ad-hoc |
| `course_name_override` | VARCHAR(300) | Tên nếu không có course_id |
| `training_year` | SMALLINT NOT NULL | Năm tính CPE |
| `start_date` | DATE | |
| `end_date` | DATE | |
| `completion_date` | DATE | |
| `status` | VARCHAR(20) | REGISTERED/IN_PROGRESS/COMPLETED/CANCELLED/NO_SHOW |
| `cpe_hours_claimed` | NUMERIC(5,1) | Giờ CPE thực tế được công nhận |
| `result` | VARCHAR(20) | PASS/FAIL/ATTEND (không thi) |
| `score` | NUMERIC(5,2) | Điểm (nếu có thi) |
| `certificate_url` | TEXT | Link chứng nhận |
| `cost` | NUMERIC(12,2) | Chi phí đào tạo (VNĐ) |
| `cost_borne_by` | VARCHAR(20) | COMPANY/EMPLOYEE/SPLIT |
| `approved_by` | UUID FK → users | |
| `notes` | TEXT | |
| `created_at` | TIMESTAMPTZ | |

### 6.4 Yêu Cầu CPE Theo Role (cpe_requirements_by_role)

| Column | Type | Mô tả |
|---|---|---|
| `id` | UUID PK | |
| `role_code` | VARCHAR(50) NOT NULL | Role áp dụng |
| `cert_type` | VARCHAR(30) | NULL = áp dụng chung |
| `required_hours_per_year` | NUMERIC(5,1) NOT NULL | |
| `regulatory_body` | VARCHAR(100) | VACPA/IIA/ACCA/... |
| `notes` | TEXT | |
| `effective_from` | DATE | |
| `effective_to` | DATE | |

**Seed data CPE requirements:**

```sql
INSERT INTO cpe_requirements_by_role (role_code, cert_type, required_hours_per_year, regulatory_body, effective_from)
VALUES
  ('PARTNER',         'VN_CPA', 40.0, 'VACPA',       '2024-01-01'),
  ('SENIOR_AUDITOR',  'VN_CPA', 40.0, 'VACPA',       '2024-01-01'),
  ('JUNIOR_AUDITOR',  'VN_CPA', 40.0, 'VACPA',       '2024-01-01'),
  ('PARTNER',         'ACCA',   40.0, 'ACCA Global', '2024-01-01'),
  ('SENIOR_AUDITOR',  'ACCA',   40.0, 'ACCA Global', '2024-01-01'),
  ('PARTNER',         'CIA',    40.0, 'IIA',          '2024-01-01');
```

### 6.5 Đánh Giá Hiệu Suất (performance_reviews)

| Column | Type | Mô tả |
|---|---|---|
| `id` | UUID PK | |
| `employee_id` | UUID FK → employees | Người được đánh giá |
| `reviewer_id` | UUID FK → employees | Người đánh giá (manager) |
| `review_period` | VARCHAR(20) NOT NULL | Ví dụ: '2026-H1', '2026-FY', '2026-Q1' |
| `review_type` | VARCHAR(20) | SELF/MANAGER/PEER/COMMITTEE |
| `status` | VARCHAR(20) | DRAFT/SUBMITTED/ACKNOWLEDGED/FINAL |
| `overall_rating` | NUMERIC(3,1) | 1.0 – 5.0 |
| `kpi_scores` | JSONB | { "kpi_code": score } |
| `strengths` | TEXT | |
| `areas_for_improvement` | TEXT | |
| `development_plan` | TEXT | |
| `reviewer_comments` | TEXT | |
| `employee_comments` | TEXT | |
| `employee_acknowledged_at` | TIMESTAMPTZ | |
| `submitted_at` | TIMESTAMPTZ | |
| `finalized_at` | TIMESTAMPTZ | |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

**Chu kỳ đánh giá:**
- `2026-H1`: Giữa năm (tháng 6–7)
- `2026-FY`: Cuối năm (tháng 12 – tháng 1 năm sau)

### 6.6 Peer Review Theo Engagement (engagement_peer_reviews)

| Column | Type | Mô tả |
|---|---|---|
| `id` | UUID PK | |
| `engagement_id` | UUID FK → engagements | |
| `reviewer_id` | UUID FK → employees | Người review |
| `reviewee_id` | UUID FK → employees | Người được review |
| `review_period` | VARCHAR(20) | |
| `technical_rating` | SMALLINT | 1–5 |
| `teamwork_rating` | SMALLINT | 1–5 |
| `communication_rating` | SMALLINT | 1–5 |
| `comments` | TEXT | |
| `is_anonymous` | BOOLEAN DEFAULT true | |
| `submitted_at` | TIMESTAMPTZ | |
| `created_at` | TIMESTAMPTZ | |

### 6.7 Khai Báo Độc Lập (independence_declarations)

Đây là yêu cầu **bắt buộc** theo chuẩn mực kiểm toán VSA 220 / ISA 220.

| Column | Type | Mô tả |
|---|---|---|
| `id` | UUID PK | |
| `employee_id` | UUID FK → employees | |
| `declaration_type` | VARCHAR(20) | ANNUAL/PER_ENGAGEMENT |
| `engagement_id` | UUID FK → engagements | NULL nếu ANNUAL |
| `declaration_year` | SMALLINT | Năm khai báo (nếu ANNUAL) |
| `declared_at` | TIMESTAMPTZ NOT NULL | |
| `has_conflict` | BOOLEAN NOT NULL | Có xung đột lợi ích không |
| `conflict_description` | TEXT | Mô tả xung đột (nếu có) |
| `resolution_action` | TEXT | Biện pháp xử lý xung đột |
| `acknowledged_by_partner` | UUID FK → employees | Partner phụ trách xác nhận |
| `acknowledged_at` | TIMESTAMPTZ | |
| `status` | VARCHAR(20) | PENDING/CLEAN/CONFLICT_RESOLVED/WITHDRAWN |
| `ip_address` | INET | IP khi khai báo |
| `created_at` | TIMESTAMPTZ | |

**Quy tắc:**
- Mỗi năm, **tất cả nhân viên** tham gia engagement phải có ít nhất 1 `ANNUAL` declaration
- Trước khi bắt đầu mỗi engagement, **tất cả thành viên team** phải khai báo `PER_ENGAGEMENT`
- Nếu `has_conflict = true`, Partner phụ trách phải `acknowledged` trước khi engagement được bắt đầu

---

## 7. Time & Leave

### 7.1 Ngày Lễ Quốc Gia (holidays)

| Column | Type | Mô tả |
|---|---|---|
| `id` | UUID PK | |
| `holiday_date` | DATE UNIQUE NOT NULL | |
| `name` | VARCHAR(200) NOT NULL | |
| `name_en` | VARCHAR(200) | |
| `type` | VARCHAR(20) | NATIONAL/COMPANY/REGIONAL |
| `applies_to_branches` | TEXT[] | NULL = tất cả chi nhánh |
| `is_compensated` | BOOLEAN DEFAULT true | Nghỉ bù nếu trùng cuối tuần |
| `year` | SMALLINT GENERATED | Extracted từ holiday_date |
| `notes` | TEXT | |

**Seed data ngày lễ 2026:**

```sql
INSERT INTO holidays (holiday_date, name, name_en, type) VALUES
  ('2026-01-01', 'Tết Dương lịch',           'New Year''s Day',              'NATIONAL'),
  ('2026-01-28', 'Tết Nguyên Đán (28 tháng Chạp)', 'Lunar New Year Eve',     'NATIONAL'),
  ('2026-01-29', 'Tết Nguyên Đán (Mùng 1)',   'Lunar New Year Day 1',        'NATIONAL'),
  ('2026-01-30', 'Tết Nguyên Đán (Mùng 2)',   'Lunar New Year Day 2',        'NATIONAL'),
  ('2026-01-31', 'Tết Nguyên Đán (Mùng 3)',   'Lunar New Year Day 3',        'NATIONAL'),
  ('2026-02-01', 'Tết Nguyên Đán (Mùng 4)',   'Lunar New Year Day 4',        'NATIONAL'),
  ('2026-02-02', 'Tết Nguyên Đán (Mùng 5)',   'Lunar New Year Day 5',        'NATIONAL'),
  ('2026-04-07', 'Giỗ Tổ Hùng Vương',         'Hung King''s Commemoration',  'NATIONAL'),
  ('2026-04-30', 'Ngày Giải phóng Miền Nam',  'Liberation Day',              'NATIONAL'),
  ('2026-05-01', 'Ngày Quốc tế Lao động',     'International Labor Day',     'NATIONAL'),
  ('2026-09-02', 'Ngày Quốc khánh',           'National Day',                'NATIONAL'),
  ('2026-09-03', 'Ngày Quốc khánh (nghỉ bù)', 'National Day (compensated)', 'NATIONAL');
-- Tương tự cho 2027, 2028, 2029, 2030
```

### 7.2 Loại Nghỉ Phép

| Code | Tên | Có lương | Số ngày tối đa | Ghi chú |
|---|---|---|---|---|
| `ANNUAL` | Nghỉ phép năm | ✅ | 12–16 ngày/năm | Theo thâm niên |
| `SICK` | Nghỉ ốm | ✅ (BHXH trả) | 30 ngày/năm | Cần giấy bệnh viện nếu >2 ngày |
| `MATERNITY` | Thai sản (nữ) | ✅ (BHXH trả) | 180 ngày | Bộ luật LĐ 2019 |
| `PATERNITY` | Thai sản (nam) | ✅ (BHXH trả) | 5–14 ngày | Tùy số con |
| `PERSONAL` | Nghỉ việc riêng | ❌ | Không giới hạn | Không hưởng lương |
| `MARRIAGE` | Nghỉ kết hôn | ✅ | 3 ngày | |
| `FUNERAL` | Nghỉ tang | ✅ | 3 ngày | Người thân ruột thịt |
| `UNPAID` | Nghỉ không lương | ❌ | Thỏa thuận | Cần CEO/HR duyệt |

### 7.3 Cân Đối Nghỉ Phép (leave_balances)

| Column | Type | Mô tả |
|---|---|---|
| `id` | UUID PK | |
| `employee_id` | UUID FK → employees | |
| `leave_type` | VARCHAR(20) | ANNUAL/SICK/... |
| `year` | SMALLINT NOT NULL | |
| `entitled_days` | NUMERIC(5,1) NOT NULL | Số ngày được hưởng |
| `carried_over_days` | NUMERIC(5,1) DEFAULT 0 | Ngày chuyển từ năm trước |
| `used_days` | NUMERIC(5,1) DEFAULT 0 | Đã dùng (tự động update) |
| `pending_days` | NUMERIC(5,1) DEFAULT 0 | Đang chờ duyệt |
| `remaining_days` | NUMERIC(5,1) GENERATED | entitled + carried - used - pending |
| `notes` | TEXT | |
| `updated_at` | TIMESTAMPTZ | |

**Chính sách phép năm theo thâm niên:**

| Thâm niên | Số ngày phép/năm |
|---|---|
| < 1 năm | Pro-rated (1 ngày/tháng) |
| 1 – 5 năm | 12 ngày |
| 5 – 10 năm | 14 ngày |
| > 10 năm | 16 ngày |

### 7.4 Đơn Nghỉ Phép (leave_requests)

| Column | Type | Mô tả |
|---|---|---|
| `id` | UUID PK | |
| `employee_id` | UUID FK → employees | |
| `leave_type` | VARCHAR(20) NOT NULL | |
| `start_date` | DATE NOT NULL | |
| `end_date` | DATE NOT NULL | |
| `total_days` | NUMERIC(5,1) NOT NULL | |
| `reason` | TEXT | |
| `supporting_doc_url` | TEXT | Link giấy tờ |
| `status` | VARCHAR(20) | PENDING/APPROVED/REJECTED/CANCELLED |
| `approved_by` | UUID FK → users | |
| `approved_at` | TIMESTAMPTZ | |
| `rejection_reason` | TEXT | |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

**Quy trình duyệt:**
1. Nhân viên submit → Manager nhận notification
2. Manager APPROVE/REJECT → Nhân viên nhận notification
3. Nếu `UNPAID`: cần HR_MANAGER confirm thêm

### 7.5 Tăng Ca / Overtime (ot_requests)

| Column | Type | Mô tả |
|---|---|---|
| `id` | UUID PK | |
| `employee_id` | UUID FK → employees | |
| `ot_date` | DATE NOT NULL | |
| `start_time` | TIME NOT NULL | |
| `end_time` | TIME NOT NULL | |
| `ot_hours` | NUMERIC(4,1) NOT NULL | |
| `ot_type` | VARCHAR(20) | WEEKDAY/WEEKEND/HOLIDAY |
| `reason` | TEXT NOT NULL | |
| `engagement_id` | UUID FK → engagements | Liên kết dự án (nullable) |
| `status` | VARCHAR(20) | PENDING/APPROVED/REJECTED |
| `approved_by` | UUID FK → users | |
| `approved_at` | TIMESTAMPTZ | |
| `is_compensated_leave` | BOOLEAN DEFAULT false | Nghỉ bù thay tiền |
| `created_at` | TIMESTAMPTZ | |

**OT rate theo Bộ luật Lao động 2019:**

| Loại OT | Hệ số lương |
|---|---|
| Ngày thường | 150% |
| Ngày nghỉ cuối tuần | 200% |
| Ngày lễ quốc gia | 300% |
| Ban đêm (22h–6h) | +30% thêm |

**Giới hạn OT:**
- Tối đa **40 giờ/tháng**
- Tối đa **300 giờ/năm** (theo Điều 107 BLLĐ 2019)
- Một số ngành đặc biệt có thể được lên đến 400h/năm — nhưng công ty áp dụng 300h

### 7.6 OT Summary View

```sql
CREATE VIEW employee_ot_summary_year AS
SELECT
  employee_id,
  EXTRACT(YEAR FROM ot_date)::SMALLINT AS year,
  SUM(ot_hours) FILTER (WHERE status = 'APPROVED') AS approved_hours,
  SUM(ot_hours) FILTER (WHERE status = 'PENDING')  AS pending_hours,
  300.0 - SUM(ot_hours) FILTER (WHERE status = 'APPROVED') AS remaining_cap
FROM ot_requests
GROUP BY employee_id, EXTRACT(YEAR FROM ot_date);
```

### 7.7 Cột OT Bổ Sung Trên timesheets

Migration 000023 ALTER bảng `timesheets` hiện tại:

```sql
ALTER TABLE timesheets
  ADD COLUMN IF NOT EXISTS ot_hours       NUMERIC(4,1) DEFAULT 0,
  ADD COLUMN IF NOT EXISTS ot_approved    BOOLEAN DEFAULT false,
  ADD COLUMN IF NOT EXISTS ot_request_id  UUID REFERENCES ot_requests(id);
```

---

## 8. User Provisioning Workflow

### 8.1 Vấn Đề

Khi tuyển nhân viên mới, cần cấp tài khoản hệ thống. Quy trình phải:
1. Có approval để tránh cấp quyền sai
2. Phân biệt rõ ai có quyền request, ai approve, ai execute
3. Có trail audit đầy đủ

### 8.2 Bảng user_provisioning_requests

| Column | Type | Mô tả |
|---|---|---|
| `id` | UUID PK | |
| `employee_id` | UUID FK → employees | Nhân viên cần cấp account |
| `requested_by` | UUID FK → users | Người yêu cầu |
| `requested_role` | VARCHAR(50) NOT NULL | Role cần gán |
| `requested_branch_id` | UUID FK → branches | |
| `status` | VARCHAR(20) | PENDING/APPROVED/REJECTED/EXECUTED/CANCELLED |
| `approval_level` | SMALLINT | 1 = HoB, 2 = HR |
| `branch_approver_id` | UUID FK → users | HoB HCM (bước 1) |
| `branch_approved_at` | TIMESTAMPTZ | |
| `branch_rejection_reason` | TEXT | |
| `hr_approver_id` | UUID FK → users | HR Manager (bước 2) |
| `hr_approved_at` | TIMESTAMPTZ | |
| `hr_rejection_reason` | TEXT | |
| `executed_by` | UUID FK → users | SA thực thi |
| `executed_at` | TIMESTAMPTZ | |
| `is_emergency` | BOOLEAN DEFAULT false | Cấp khẩn, skip approval |
| `emergency_reason` | TEXT | |
| `notes` | TEXT | |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

### 8.3 Luồng Phê Duyệt

#### Flow HCM (Chi nhánh):

```
Trigger: Nhân viên HCM cần account
   │
   ▼
[HoB HCM] nhận notification
   │
   ├── REJECT → Requester nhận notification
   │
   └── APPROVE (step 1) →
         │
         ▼
      [HR Manager (HO)] nhận notification
         │
         ├── REJECT → Requester + HoB nhận notification
         │
         └── APPROVE (step 2) →
               │
               ▼
            [SUPER_ADMIN] nhận notification
               │
               └── EXECUTE (atomic):
                     1. Tạo user account
                     2. Gán role
                     3. Link user_id → employee
                     4. Gửi welcome email
                     5. Update request status = EXECUTED
```

#### Flow HO (Trụ sở):

```
Trigger: Nhân viên HO cần account
   │
   ▼
[CEO hoặc HR Manager] tạo trực tiếp
   │
   └── SA execute ngay (không cần approval riêng)
```

#### Emergency Flow:

```
SA hoặc CEO tạo với is_emergency = true
   │
   └── Account tạo ngay lập tức
         └── Audit log ghi nhận emergency reason
```

### 8.4 Ràng Buộc Nghiệp Vụ

- Không thể có 2 `PENDING` request cho cùng 1 `employee_id`
- Role `SUPER_ADMIN` và `CHAIRMAN` chỉ SA mới có thể gán (không qua flow này)
- Sau khi EXECUTED, không thể cancel
- Request hết hạn sau 30 ngày nếu không được xử lý (scheduled job)

---

## 9. Lifecycle Events

### 9.1 Onboarding Process

Khi `employment_contracts` mới được tạo cho nhân viên:

**Checklist onboarding (stored as JSONB trong offboarding_checklists với type='ONBOARDING'):**

```json
{
  "items": [
    { "key": "contract_signed",       "label": "Ký hợp đồng lao động",         "owner": "HR",       "done": false },
    { "key": "cccd_copy",             "label": "Nộp bản sao CCCD",              "owner": "HR",       "done": false },
    { "key": "bhxh_registered",       "label": "Đăng ký BHXH",                  "owner": "HR",       "done": false },
    { "key": "bank_account_provided", "label": "Cung cấp tài khoản ngân hàng",  "owner": "Employee", "done": false },
    { "key": "email_setup",           "label": "Cài đặt email công ty",          "owner": "IT",       "done": false },
    { "key": "system_access",         "label": "Cấp quyền hệ thống",            "owner": "IT",       "done": false },
    { "key": "independence_declared", "label": "Khai báo độc lập lần đầu",      "owner": "Employee", "done": false },
    { "key": "orientation_done",      "label": "Hoàn thành buổi định hướng",    "owner": "HR",       "done": false }
  ]
}
```

### 9.2 Thăng Chức / Thuyên Chuyển (Promotion/Transfer)

Workflow:
1. Manager/HR tạo `promotion_event` (bảng lifecycle_events — Phase 2 có thể tách riêng)
2. Trong Phase 1: ghi nhận qua `employee_salary_history` (change_type = 'PROMOTION') + cập nhật `employees.grade`, `position_title`, `department_id`
3. Tạo hợp đồng mới hoặc phụ lục hợp đồng
4. Audit log ghi nhận before/after

### 9.3 Cảnh Báo Hợp Đồng Sắp Hết Hạn

Scheduled job chạy hàng ngày lúc 8:00 SA:
- Tìm tất cả `employment_contracts` có `end_date` trong **30 ngày tới** và `is_current = true`
- Gửi notification cho HR_MANAGER và manager trực tiếp
- Gửi in-app + email

### 9.4 Offboarding Checklist (offboarding_checklists)

| Column | Type | Mô tả |
|---|---|---|
| `id` | UUID PK | |
| `employee_id` | UUID FK → employees | |
| `checklist_type` | VARCHAR(20) | ONBOARDING/OFFBOARDING |
| `initiated_by` | UUID FK → users | |
| `target_date` | DATE | Ngày dự kiến hoàn thành |
| `items` | JSONB NOT NULL | Array các checklist item |
| `status` | VARCHAR(20) | IN_PROGRESS/COMPLETED/CANCELLED |
| `completed_at` | TIMESTAMPTZ | |
| `notes` | TEXT | |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

**Offboarding checklist template:**

```json
{
  "items": [
    { "key": "resignation_letter",    "label": "Nhận đơn xin thôi việc",        "owner": "HR",       "done": false },
    { "key": "handover_doc",          "label": "Bàn giao tài liệu công việc",   "owner": "Manager",  "done": false },
    { "key": "equipment_return",      "label": "Thu hồi thiết bị công ty",      "owner": "IT",       "done": false },
    { "key": "access_revoked",        "label": "Thu hồi quyền truy cập hệ thống", "owner": "IT",     "done": false },
    { "key": "final_settlement",      "label": "Thanh toán lương cuối + BHXH",  "owner": "HR",       "done": false },
    { "key": "bhxh_transfer",         "label": "Chốt sổ BHXH",                 "owner": "HR",       "done": false },
    { "key": "tax_finalization",      "label": "Quyết toán thuế TNCN",          "owner": "FIN",      "done": false },
    { "key": "exit_interview",        "label": "Phỏng vấn thôi việc",           "owner": "HR",       "done": false }
  ]
}
```

---

## 10. Expense Claims

### 10.1 Chi Phí Công Tác (expense_claims)

| Column | Type | Mô tả |
|---|---|---|
| `id` | UUID PK | |
| `claim_number` | VARCHAR(20) UNIQUE | PC{YY}-{SEQ4}, auto-generated |
| `employee_id` | UUID FK → employees | |
| `engagement_id` | UUID FK → engagements | NULL nếu không liên quan dự án |
| `claim_period_from` | DATE | |
| `claim_period_to` | DATE | |
| `total_amount` | NUMERIC(15,2) | Tổng tiền (tự tính từ items) |
| `currency` | VARCHAR(3) DEFAULT 'VND' | |
| `description` | TEXT | |
| `status` | VARCHAR(20) | DRAFT/SUBMITTED/MANAGER_APPROVED/HR_APPROVED/PAID/REJECTED |
| `submitted_at` | TIMESTAMPTZ | |
| `manager_approver_id` | UUID FK → users | |
| `manager_approved_at` | TIMESTAMPTZ | |
| `manager_rejection_reason` | TEXT | |
| `hr_approver_id` | UUID FK → users | |
| `hr_approved_at` | TIMESTAMPTZ | |
| `hr_rejection_reason` | TEXT | |
| `paid_at` | TIMESTAMPTZ | |
| `payment_reference` | VARCHAR(100) | Số chứng từ thanh toán |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

### 10.2 Chi Tiết Chi Phí (expense_claim_items)

| Column | Type | Mô tả |
|---|---|---|
| `id` | UUID PK | |
| `claim_id` | UUID FK → expense_claims | |
| `category` | VARCHAR(30) | FLIGHT/HOTEL/TAXI/MEAL/STATIONERY/COMMUNICATION/OTHER |
| `expense_date` | DATE NOT NULL | |
| `description` | VARCHAR(500) | |
| `amount` | NUMERIC(12,2) NOT NULL | |
| `currency` | VARCHAR(3) DEFAULT 'VND' | |
| `is_billable` | BOOLEAN DEFAULT false | Có tính vào chi phí khách hàng không |
| `receipt_url` | TEXT | Link hóa đơn scan |
| `notes` | TEXT | |
| `created_at` | TIMESTAMPTZ | |

### 10.3 Danh Mục Chi Phí

| Category | Tên | Yêu cầu hóa đơn | Mức tối đa/lần |
|---|---|---|---|
| `FLIGHT` | Vé máy bay | ✅ Bắt buộc | Không giới hạn (theo policy) |
| `HOTEL` | Khách sạn | ✅ Bắt buộc | Policy tùy cấp bậc |
| `TAXI` | Taxi / grab | ✅ Nếu > 100k | 500k/ngày |
| `MEAL` | Ăn uống | ✅ Nếu > 200k | 300k/ngày |
| `STATIONERY` | Văn phòng phẩm | ✅ Bắt buộc | 500k/lần |
| `COMMUNICATION` | Điện thoại / internet | ✅ Bắt buộc | 300k/tháng |
| `OTHER` | Khác | ✅ Bắt buộc | Cần ghi rõ |

### 10.4 Luồng Phê Duyệt Chi Phí

```
DRAFT → SUBMITTED → MANAGER_APPROVED → HR_APPROVED → PAID
                  ↘                  ↘
               REJECTED           REJECTED
```

1. Nhân viên tạo claim ở DRAFT, thêm items
2. Submit → Manager trực tiếp nhận notification
3. Manager approve → HR nhận notification
4. HR approve → FIN/kế toán thanh toán
5. FIN đánh dấu PAID với payment_reference

---


---

## 11. Database Schema

### 11.1 ALTER branches

```sql
-- Migration 000019
ALTER TABLE branches
  ADD COLUMN IF NOT EXISTS is_head_office         BOOLEAN     NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS city                   VARCHAR(100),
  ADD COLUMN IF NOT EXISTS address                TEXT,
  ADD COLUMN IF NOT EXISTS phone                  VARCHAR(20),
  ADD COLUMN IF NOT EXISTS established_date       DATE,
  ADD COLUMN IF NOT EXISTS head_of_branch_user_id UUID        REFERENCES users(id),
  ADD COLUMN IF NOT EXISTS is_active              BOOLEAN     NOT NULL DEFAULT true;

CREATE UNIQUE INDEX IF NOT EXISTS uidx_branches_head_office
  ON branches (is_head_office) WHERE is_head_office = true;
```

### 11.2 ALTER departments

```sql
ALTER TABLE departments
  ADD COLUMN IF NOT EXISTS code        VARCHAR(20) UNIQUE,
  ADD COLUMN IF NOT EXISTS dept_type   VARCHAR(20) DEFAULT 'CORE'
    CHECK (dept_type IN ('CORE','SUPPORT','MANAGEMENT')),
  ADD COLUMN IF NOT EXISTS description TEXT,
  ADD COLUMN IF NOT EXISTS is_active   BOOLEAN NOT NULL DEFAULT true;
```

### 11.3 CREATE branch_departments

```sql
CREATE TABLE IF NOT EXISTS branch_departments (
  id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  branch_id     UUID        NOT NULL REFERENCES branches(id),
  department_id UUID        NOT NULL REFERENCES departments(id),
  is_active     BOOLEAN     NOT NULL DEFAULT true,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT uidx_branch_department UNIQUE (branch_id, department_id)
);
CREATE INDEX idx_branch_departments_branch ON branch_departments(branch_id);
CREATE INDEX idx_branch_departments_dept   ON branch_departments(department_id);
```

### 11.4 ALTER employees (Extended Fields)

```sql
ALTER TABLE employees
  -- Basic
  ADD COLUMN IF NOT EXISTS employee_code              VARCHAR(12)  UNIQUE,
  ADD COLUMN IF NOT EXISTS grade                      VARCHAR(20)
    CHECK (grade IN ('EXECUTIVE','PARTNER','DIRECTOR','MANAGER','SENIOR','JUNIOR','INTERN','SUPPORT')),
  ADD COLUMN IF NOT EXISTS position_title             VARCHAR(200),
  ADD COLUMN IF NOT EXISTS manager_id                 UUID         REFERENCES employees(id),
  ADD COLUMN IF NOT EXISTS employment_type            VARCHAR(20)  DEFAULT 'FULL_TIME'
    CHECK (employment_type IN ('FULL_TIME','PART_TIME','INTERN')),
  ADD COLUMN IF NOT EXISTS status                     VARCHAR(20)  NOT NULL DEFAULT 'ACTIVE'
    CHECK (status IN ('ACTIVE','INACTIVE','ON_LEAVE','TERMINATED')),
  ADD COLUMN IF NOT EXISTS hired_date                 DATE,
  ADD COLUMN IF NOT EXISTS probation_end_date         DATE,
  ADD COLUMN IF NOT EXISTS termination_date           DATE,
  ADD COLUMN IF NOT EXISTS termination_reason         TEXT,
  ADD COLUMN IF NOT EXISTS current_contract_id        UUID,
  -- Personal / PII
  ADD COLUMN IF NOT EXISTS gender                     VARCHAR(10)
    CHECK (gender IN ('MALE','FEMALE','OTHER')),
  ADD COLUMN IF NOT EXISTS date_of_birth              DATE,
  ADD COLUMN IF NOT EXISTS place_of_birth             VARCHAR(200),
  ADD COLUMN IF NOT EXISTS nationality                VARCHAR(50)  DEFAULT 'Vietnamese',
  ADD COLUMN IF NOT EXISTS ethnicity                  VARCHAR(50),
  ADD COLUMN IF NOT EXISTS personal_email             VARCHAR(200),
  ADD COLUMN IF NOT EXISTS personal_phone             VARCHAR(20),
  ADD COLUMN IF NOT EXISTS work_phone                 VARCHAR(20),
  ADD COLUMN IF NOT EXISTS current_address            TEXT,
  ADD COLUMN IF NOT EXISTS permanent_address          TEXT,
  ADD COLUMN IF NOT EXISTS cccd_encrypted             TEXT,
  ADD COLUMN IF NOT EXISTS cccd_issued_date           DATE,
  ADD COLUMN IF NOT EXISTS cccd_issued_place          VARCHAR(200),
  ADD COLUMN IF NOT EXISTS passport_number            VARCHAR(50),
  ADD COLUMN IF NOT EXISTS passport_expiry            DATE,
  -- Employment
  ADD COLUMN IF NOT EXISTS hired_source               VARCHAR(50)
    CHECK (hired_source IN ('REFERRAL','PORTAL','DIRECT','AGENCY')),
  ADD COLUMN IF NOT EXISTS referrer_employee_id       UUID         REFERENCES employees(id),
  ADD COLUMN IF NOT EXISTS probation_salary_pct       NUMERIC(5,2) DEFAULT 85.00,
  ADD COLUMN IF NOT EXISTS work_location              VARCHAR(20)  DEFAULT 'OFFICE'
    CHECK (work_location IN ('OFFICE','REMOTE','HYBRID')),
  ADD COLUMN IF NOT EXISTS remote_days_per_week       SMALLINT     DEFAULT 0,
  -- Qualifications
  ADD COLUMN IF NOT EXISTS education_level            VARCHAR(30)
    CHECK (education_level IN ('BACHELOR','MASTER','PHD','COLLEGE','OTHER')),
  ADD COLUMN IF NOT EXISTS education_major            VARCHAR(200),
  ADD COLUMN IF NOT EXISTS education_school           VARCHAR(200),
  ADD COLUMN IF NOT EXISTS education_graduation_year  SMALLINT,
  ADD COLUMN IF NOT EXISTS vn_cpa_number              VARCHAR(50),
  ADD COLUMN IF NOT EXISTS vn_cpa_issued_date         DATE,
  ADD COLUMN IF NOT EXISTS vn_cpa_expiry_date         DATE,
  ADD COLUMN IF NOT EXISTS practicing_certificate_number VARCHAR(50),
  ADD COLUMN IF NOT EXISTS practicing_certificate_expiry DATE,
  -- Salary / Bank (sensitive)
  ADD COLUMN IF NOT EXISTS base_salary                NUMERIC(15,2),
  ADD COLUMN IF NOT EXISTS salary_currency            VARCHAR(3)   DEFAULT 'VND',
  ADD COLUMN IF NOT EXISTS salary_effective_date      DATE,
  ADD COLUMN IF NOT EXISTS bank_account_encrypted     TEXT,
  ADD COLUMN IF NOT EXISTS bank_name                  VARCHAR(100),
  ADD COLUMN IF NOT EXISTS bank_branch                VARCHAR(200),
  ADD COLUMN IF NOT EXISTS mst_ca_nhan_encrypted      TEXT,
  -- Commission
  ADD COLUMN IF NOT EXISTS commission_rate            NUMERIC(5,2),
  ADD COLUMN IF NOT EXISTS commission_type            VARCHAR(20)  DEFAULT 'NONE'
    CHECK (commission_type IN ('FIXED','TIERED','NONE')),
  ADD COLUMN IF NOT EXISTS sales_target_yearly        NUMERIC(15,2),
  ADD COLUMN IF NOT EXISTS biz_dev_region             VARCHAR(100),
  -- BHXH / Insurance
  ADD COLUMN IF NOT EXISTS so_bhxh_encrypted          TEXT,
  ADD COLUMN IF NOT EXISTS bhxh_registered_date       DATE,
  ADD COLUMN IF NOT EXISTS bhxh_province_code         VARCHAR(10),
  ADD COLUMN IF NOT EXISTS bhyt_card_number           VARCHAR(20),
  ADD COLUMN IF NOT EXISTS bhyt_expiry_date           DATE,
  ADD COLUMN IF NOT EXISTS bhyt_registered_hospital_code  VARCHAR(20),
  ADD COLUMN IF NOT EXISTS bhyt_registered_hospital_name  VARCHAR(200),
  ADD COLUMN IF NOT EXISTS tncn_registered            BOOLEAN      DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_employees_branch   ON employees(branch_id);
CREATE INDEX IF NOT EXISTS idx_employees_dept     ON employees(department_id);
CREATE INDEX IF NOT EXISTS idx_employees_manager  ON employees(manager_id);
CREATE INDEX IF NOT EXISTS idx_employees_status   ON employees(status);
CREATE INDEX IF NOT EXISTS idx_employees_grade    ON employees(grade);
CREATE INDEX IF NOT EXISTS idx_employees_hired    ON employees(hired_date);
```

### 11.5 Trigger: Auto-Generate Employee Code

```sql
CREATE OR REPLACE FUNCTION fn_employees_set_code()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
  v_year TEXT;
  v_seq  INT;
  v_code TEXT;
BEGIN
  IF NEW.employee_code IS NOT NULL THEN RETURN NEW; END IF;
  v_year := TO_CHAR(COALESCE(NEW.hired_date, CURRENT_DATE), 'YY');
  SELECT COUNT(*) + 1 INTO v_seq
  FROM employees WHERE employee_code LIKE 'NV' || v_year || '-%';
  v_code := 'NV' || v_year || '-' || LPAD(v_seq::TEXT, 4, '0');
  WHILE EXISTS (SELECT 1 FROM employees WHERE employee_code = v_code) LOOP
    v_seq := v_seq + 1;
    v_code := 'NV' || v_year || '-' || LPAD(v_seq::TEXT, 4, '0');
  END LOOP;
  NEW.employee_code := v_code;
  RETURN NEW;
END;
$$;

CREATE TRIGGER trg_employees_set_code
BEFORE INSERT ON employees
FOR EACH ROW EXECUTE FUNCTION fn_employees_set_code();
```

### 11.6 CREATE employee_dependents

```sql
CREATE TABLE IF NOT EXISTS employee_dependents (
  id                        UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id               UUID        NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  full_name                 VARCHAR(200) NOT NULL,
  relationship              VARCHAR(30) NOT NULL
    CHECK (relationship IN ('SPOUSE','CHILD','PARENT','SIBLING','OTHER')),
  date_of_birth             DATE,
  cccd_or_birth_cert        VARCHAR(50),
  tax_deduction_registered  BOOLEAN     NOT NULL DEFAULT false,
  tax_deduction_from        DATE,
  tax_deduction_to          DATE,
  notes                     TEXT,
  created_at                TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at                TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_employee_dependents_employee ON employee_dependents(employee_id);
```

### 11.7 CREATE insurance_rate_config

```sql
CREATE TABLE IF NOT EXISTS insurance_rate_config (
  id                 UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  effective_from     DATE        NOT NULL,
  effective_to       DATE,
  bhxh_employee_pct  NUMERIC(5,2) NOT NULL DEFAULT 8.00,
  bhxh_employer_pct  NUMERIC(5,2) NOT NULL DEFAULT 17.50,
  bhyt_employee_pct  NUMERIC(5,2) NOT NULL DEFAULT 1.50,
  bhyt_employer_pct  NUMERIC(5,2) NOT NULL DEFAULT 3.00,
  bhtn_employee_pct  NUMERIC(5,2) NOT NULL DEFAULT 1.00,
  bhtn_employer_pct  NUMERIC(5,2) NOT NULL DEFAULT 1.00,
  kpcd_employer_pct  NUMERIC(5,2) NOT NULL DEFAULT 2.00,
  salary_base_bhxh   NUMERIC(15,2),
  max_bhxh_salary    NUMERIC(15,2),
  notes              TEXT,
  created_by         UUID        REFERENCES users(id),
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_rate_dates CHECK (effective_to IS NULL OR effective_to > effective_from)
);
CREATE UNIQUE INDEX uidx_insurance_rate_active
  ON insurance_rate_config (effective_from) WHERE effective_to IS NULL;
```

### 11.8 CREATE employee_salary_history (Immutable)

```sql
CREATE TABLE IF NOT EXISTS employee_salary_history (
  id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id      UUID        NOT NULL REFERENCES employees(id),
  effective_date   DATE        NOT NULL,
  base_salary      NUMERIC(15,2) NOT NULL,
  allowances_total NUMERIC(15,2) DEFAULT 0,
  salary_note      TEXT,
  change_type      VARCHAR(30) NOT NULL DEFAULT 'INITIAL'
    CHECK (change_type IN ('INITIAL','INCREASE','DECREASE','PROMOTION','ADJUSTMENT')),
  approved_by      UUID        REFERENCES users(id),
  created_by       UUID        REFERENCES users(id),
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- Immutable: block UPDATE and DELETE
CREATE RULE no_update_salary_history AS ON UPDATE TO employee_salary_history DO INSTEAD NOTHING;
CREATE RULE no_delete_salary_history AS ON DELETE TO employee_salary_history DO INSTEAD NOTHING;

CREATE INDEX idx_salary_history_employee ON employee_salary_history(employee_id);
CREATE INDEX idx_salary_history_date     ON employee_salary_history(effective_date);
```

### 11.9 CREATE employment_contracts

```sql
CREATE TABLE IF NOT EXISTS employment_contracts (
  id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id         UUID        NOT NULL REFERENCES employees(id),
  contract_number     VARCHAR(50) UNIQUE,
  contract_type       VARCHAR(20) NOT NULL
    CHECK (contract_type IN ('PROBATION','DEFINITE_TERM','INDEFINITE','INTERN')),
  start_date          DATE        NOT NULL,
  end_date            DATE,
  signed_date         DATE,
  salary_at_signing   NUMERIC(15,2),
  position_at_signing VARCHAR(200),
  notes               TEXT,
  document_url        TEXT,
  is_current          BOOLEAN     NOT NULL DEFAULT false,
  created_by          UUID        REFERENCES users(id),
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_contract_dates CHECK (end_date IS NULL OR end_date > start_date)
);
CREATE INDEX idx_contracts_employee ON employment_contracts(employee_id);
CREATE INDEX idx_contracts_end_date ON employment_contracts(end_date) WHERE end_date IS NOT NULL;
```

### 11.10 CREATE certifications

```sql
CREATE TABLE IF NOT EXISTS certifications (
  id                    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id           UUID        NOT NULL REFERENCES employees(id),
  cert_type             VARCHAR(30) NOT NULL
    CHECK (cert_type IN ('VN_CPA','ACCA','CFA','CIA','CISA','CPA_AUS','OTHER')),
  cert_name             VARCHAR(200) NOT NULL,
  cert_number           VARCHAR(100),
  issued_date           DATE,
  expiry_date           DATE,
  issued_by             VARCHAR(200),
  status                VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
    CHECK (status IN ('ACTIVE','EXPIRED','SUSPENDED','SURRENDERED')),
  renewal_reminder_days INT         NOT NULL DEFAULT 60,
  document_url          TEXT,
  notes                 TEXT,
  created_by            UUID        REFERENCES users(id),
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_certifications_employee ON certifications(employee_id);
CREATE INDEX idx_certifications_type     ON certifications(cert_type);
CREATE INDEX idx_certifications_expiry   ON certifications(expiry_date) WHERE expiry_date IS NOT NULL;
```

### 11.11 CREATE training_courses

```sql
CREATE TABLE IF NOT EXISTS training_courses (
  id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  course_code       VARCHAR(50) UNIQUE NOT NULL,
  course_name       VARCHAR(300) NOT NULL,
  provider          VARCHAR(200),
  course_type       VARCHAR(30) NOT NULL DEFAULT 'EXTERNAL'
    CHECK (course_type IN ('INTERNAL','EXTERNAL','ONLINE','CONFERENCE')),
  category          VARCHAR(50) NOT NULL DEFAULT 'OTHER'
    CHECK (category IN ('AUDIT','TAX','ACCOUNTING','SOFT_SKILL','COMPLIANCE','IT','OTHER')),
  cpe_hours         NUMERIC(5,1) DEFAULT 0,
  duration_days     NUMERIC(5,1),
  is_mandatory      BOOLEAN     NOT NULL DEFAULT false,
  applicable_grades TEXT[],
  description       TEXT,
  is_active         BOOLEAN     NOT NULL DEFAULT true,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### 11.12 CREATE training_records

```sql
CREATE TABLE IF NOT EXISTS training_records (
  id                   UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id          UUID        NOT NULL REFERENCES employees(id),
  course_id            UUID        REFERENCES training_courses(id),
  course_name_override VARCHAR(300),
  training_year        SMALLINT    NOT NULL,
  start_date           DATE,
  end_date             DATE,
  completion_date      DATE,
  status               VARCHAR(20) NOT NULL DEFAULT 'REGISTERED'
    CHECK (status IN ('REGISTERED','IN_PROGRESS','COMPLETED','CANCELLED','NO_SHOW')),
  cpe_hours_claimed    NUMERIC(5,1) DEFAULT 0,
  result               VARCHAR(20)
    CHECK (result IN ('PASS','FAIL','ATTEND')),
  score                NUMERIC(5,2),
  certificate_url      TEXT,
  cost                 NUMERIC(12,2),
  cost_borne_by        VARCHAR(20) DEFAULT 'COMPANY'
    CHECK (cost_borne_by IN ('COMPANY','EMPLOYEE','SPLIT')),
  approved_by          UUID        REFERENCES users(id),
  notes                TEXT,
  created_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_training_records_employee ON training_records(employee_id);
CREATE INDEX idx_training_records_year     ON training_records(training_year);
CREATE INDEX idx_training_records_status   ON training_records(status);
```

### 11.13 CREATE cpe_requirements_by_role

```sql
CREATE TABLE IF NOT EXISTS cpe_requirements_by_role (
  id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  role_code               VARCHAR(50) NOT NULL,
  cert_type               VARCHAR(30),
  required_hours_per_year NUMERIC(5,1) NOT NULL,
  regulatory_body         VARCHAR(100),
  notes                   TEXT,
  effective_from          DATE,
  effective_to            DATE
);
```

### 11.14 CREATE performance_reviews

```sql
CREATE TABLE IF NOT EXISTS performance_reviews (
  id                       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id              UUID        NOT NULL REFERENCES employees(id),
  reviewer_id              UUID        REFERENCES employees(id),
  review_period            VARCHAR(20) NOT NULL,
  review_type              VARCHAR(20) NOT NULL DEFAULT 'MANAGER'
    CHECK (review_type IN ('SELF','MANAGER','PEER','COMMITTEE')),
  status                   VARCHAR(20) NOT NULL DEFAULT 'DRAFT'
    CHECK (status IN ('DRAFT','SUBMITTED','ACKNOWLEDGED','FINAL')),
  overall_rating           NUMERIC(3,1) CHECK (overall_rating BETWEEN 1.0 AND 5.0),
  kpi_scores               JSONB,
  strengths                TEXT,
  areas_for_improvement    TEXT,
  development_plan         TEXT,
  reviewer_comments        TEXT,
  employee_comments        TEXT,
  employee_acknowledged_at TIMESTAMPTZ,
  submitted_at             TIMESTAMPTZ,
  finalized_at             TIMESTAMPTZ,
  created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT uidx_review_employee_period UNIQUE (employee_id, review_period, review_type)
);
CREATE INDEX idx_perf_reviews_employee ON performance_reviews(employee_id);
CREATE INDEX idx_perf_reviews_period   ON performance_reviews(review_period);
CREATE INDEX idx_perf_reviews_status   ON performance_reviews(status);
```

### 11.15 CREATE engagement_peer_reviews

```sql
CREATE TABLE IF NOT EXISTS engagement_peer_reviews (
  id                   UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  engagement_id        UUID        NOT NULL REFERENCES engagements(id),
  reviewer_id          UUID        NOT NULL REFERENCES employees(id),
  reviewee_id          UUID        NOT NULL REFERENCES employees(id),
  review_period        VARCHAR(20),
  technical_rating     SMALLINT    CHECK (technical_rating BETWEEN 1 AND 5),
  teamwork_rating      SMALLINT    CHECK (teamwork_rating BETWEEN 1 AND 5),
  communication_rating SMALLINT    CHECK (communication_rating BETWEEN 1 AND 5),
  comments             TEXT,
  is_anonymous         BOOLEAN     NOT NULL DEFAULT true,
  submitted_at         TIMESTAMPTZ,
  created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_peer_review_self CHECK (reviewer_id <> reviewee_id),
  CONSTRAINT uidx_peer_review UNIQUE (engagement_id, reviewer_id, reviewee_id)
);
```

### 11.16 CREATE independence_declarations

```sql
CREATE TABLE IF NOT EXISTS independence_declarations (
  id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id             UUID        NOT NULL REFERENCES employees(id),
  declaration_type        VARCHAR(20) NOT NULL
    CHECK (declaration_type IN ('ANNUAL','PER_ENGAGEMENT')),
  engagement_id           UUID        REFERENCES engagements(id),
  declaration_year        SMALLINT,
  declared_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
  has_conflict            BOOLEAN     NOT NULL,
  conflict_description    TEXT,
  resolution_action       TEXT,
  acknowledged_by_partner UUID        REFERENCES employees(id),
  acknowledged_at         TIMESTAMPTZ,
  status                  VARCHAR(20) NOT NULL DEFAULT 'PENDING'
    CHECK (status IN ('PENDING','CLEAN','CONFLICT_RESOLVED','WITHDRAWN')),
  ip_address              INET,
  created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_annual_year    CHECK (declaration_type <> 'ANNUAL' OR declaration_year IS NOT NULL),
  CONSTRAINT chk_engagement_ref CHECK (declaration_type <> 'PER_ENGAGEMENT' OR engagement_id IS NOT NULL)
);
CREATE INDEX idx_independence_employee   ON independence_declarations(employee_id);
CREATE INDEX idx_independence_engagement ON independence_declarations(engagement_id);
CREATE UNIQUE INDEX uidx_independence_annual
  ON independence_declarations(employee_id, declaration_year)
  WHERE declaration_type = 'ANNUAL';
```

### 11.17 CREATE holidays

```sql
CREATE TABLE IF NOT EXISTS holidays (
  id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  holiday_date        DATE        UNIQUE NOT NULL,
  name                VARCHAR(200) NOT NULL,
  name_en             VARCHAR(200),
  type                VARCHAR(20) NOT NULL DEFAULT 'NATIONAL'
    CHECK (type IN ('NATIONAL','COMPANY','REGIONAL')),
  applies_to_branches TEXT[],
  is_compensated      BOOLEAN     NOT NULL DEFAULT true,
  year                SMALLINT    GENERATED ALWAYS AS (EXTRACT(YEAR FROM holiday_date)::SMALLINT) STORED,
  notes               TEXT
);
CREATE INDEX idx_holidays_year ON holidays(year);
```

### 11.18 CREATE leave_balances

```sql
CREATE TABLE IF NOT EXISTS leave_balances (
  id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id       UUID        NOT NULL REFERENCES employees(id),
  leave_type        VARCHAR(20) NOT NULL
    CHECK (leave_type IN ('ANNUAL','SICK','MATERNITY','PATERNITY','PERSONAL','MARRIAGE','FUNERAL','UNPAID')),
  year              SMALLINT    NOT NULL,
  entitled_days     NUMERIC(5,1) NOT NULL DEFAULT 0,
  carried_over_days NUMERIC(5,1) NOT NULL DEFAULT 0,
  used_days         NUMERIC(5,1) NOT NULL DEFAULT 0,
  pending_days      NUMERIC(5,1) NOT NULL DEFAULT 0,
  notes             TEXT,
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT uidx_leave_balance UNIQUE (employee_id, leave_type, year),
  CONSTRAINT chk_days_non_negative CHECK (
    entitled_days >= 0 AND carried_over_days >= 0 AND
    used_days >= 0 AND pending_days >= 0
  )
);
CREATE INDEX idx_leave_balances_employee ON leave_balances(employee_id, year);
```

### 11.19 CREATE leave_requests

```sql
CREATE TABLE IF NOT EXISTS leave_requests (
  id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id       UUID        NOT NULL REFERENCES employees(id),
  leave_type        VARCHAR(20) NOT NULL
    CHECK (leave_type IN ('ANNUAL','SICK','MATERNITY','PATERNITY','PERSONAL','MARRIAGE','FUNERAL','UNPAID')),
  start_date        DATE        NOT NULL,
  end_date          DATE        NOT NULL,
  total_days        NUMERIC(5,1) NOT NULL,
  reason            TEXT,
  supporting_doc_url TEXT,
  status            VARCHAR(20) NOT NULL DEFAULT 'PENDING'
    CHECK (status IN ('PENDING','APPROVED','REJECTED','CANCELLED')),
  approved_by       UUID        REFERENCES users(id),
  approved_at       TIMESTAMPTZ,
  rejection_reason  TEXT,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_leave_dates CHECK (end_date >= start_date),
  CONSTRAINT chk_leave_days  CHECK (total_days > 0)
);
CREATE INDEX idx_leave_requests_employee ON leave_requests(employee_id);
CREATE INDEX idx_leave_requests_status   ON leave_requests(status);
CREATE INDEX idx_leave_requests_dates    ON leave_requests(start_date, end_date);
```

### 11.20 CREATE ot_requests

```sql
CREATE TABLE IF NOT EXISTS ot_requests (
  id                   UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id          UUID        NOT NULL REFERENCES employees(id),
  ot_date              DATE        NOT NULL,
  start_time           TIME        NOT NULL,
  end_time             TIME        NOT NULL,
  ot_hours             NUMERIC(4,1) NOT NULL,
  ot_type              VARCHAR(20) NOT NULL DEFAULT 'WEEKDAY'
    CHECK (ot_type IN ('WEEKDAY','WEEKEND','HOLIDAY')),
  reason               TEXT        NOT NULL,
  engagement_id        UUID        REFERENCES engagements(id),
  status               VARCHAR(20) NOT NULL DEFAULT 'PENDING'
    CHECK (status IN ('PENDING','APPROVED','REJECTED')),
  approved_by          UUID        REFERENCES users(id),
  approved_at          TIMESTAMPTZ,
  is_compensated_leave BOOLEAN     NOT NULL DEFAULT false,
  created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_ot_times CHECK (end_time > start_time),
  CONSTRAINT chk_ot_hours CHECK (ot_hours > 0 AND ot_hours <= 12)
);
CREATE INDEX idx_ot_requests_employee ON ot_requests(employee_id);
CREATE INDEX idx_ot_requests_date     ON ot_requests(ot_date);
CREATE INDEX idx_ot_requests_status   ON ot_requests(status);
```

### 11.21 CREATE VIEW employee_ot_summary_year

```sql
CREATE OR REPLACE VIEW employee_ot_summary_year AS
SELECT
  employee_id,
  EXTRACT(YEAR FROM ot_date)::SMALLINT AS year,
  SUM(ot_hours) FILTER (WHERE status = 'APPROVED') AS approved_hours,
  SUM(ot_hours) FILTER (WHERE status = 'PENDING')  AS pending_hours,
  300.0 - COALESCE(SUM(ot_hours) FILTER (WHERE status = 'APPROVED'), 0) AS remaining_cap
FROM ot_requests
GROUP BY employee_id, EXTRACT(YEAR FROM ot_date);
```

### 11.22 CREATE user_provisioning_requests

```sql
CREATE TABLE IF NOT EXISTS user_provisioning_requests (
  id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id             UUID        NOT NULL REFERENCES employees(id),
  requested_by            UUID        NOT NULL REFERENCES users(id),
  requested_role          VARCHAR(50) NOT NULL,
  requested_branch_id     UUID        REFERENCES branches(id),
  status                  VARCHAR(20) NOT NULL DEFAULT 'PENDING'
    CHECK (status IN ('PENDING','APPROVED','REJECTED','EXECUTED','CANCELLED')),
  approval_level          SMALLINT    NOT NULL DEFAULT 1,
  branch_approver_id      UUID        REFERENCES users(id),
  branch_approved_at      TIMESTAMPTZ,
  branch_rejection_reason TEXT,
  hr_approver_id          UUID        REFERENCES users(id),
  hr_approved_at          TIMESTAMPTZ,
  hr_rejection_reason     TEXT,
  executed_by             UUID        REFERENCES users(id),
  executed_at             TIMESTAMPTZ,
  is_emergency            BOOLEAN     NOT NULL DEFAULT false,
  emergency_reason        TEXT,
  notes                   TEXT,
  expires_at              TIMESTAMPTZ NOT NULL DEFAULT (now() + INTERVAL '30 days'),
  created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX uidx_provisioning_pending
  ON user_provisioning_requests(employee_id) WHERE status = 'PENDING';
CREATE INDEX idx_provisioning_status   ON user_provisioning_requests(status);
CREATE INDEX idx_provisioning_employee ON user_provisioning_requests(employee_id);
```

### 11.23 CREATE offboarding_checklists

```sql
CREATE TABLE IF NOT EXISTS offboarding_checklists (
  id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id    UUID        NOT NULL REFERENCES employees(id),
  checklist_type VARCHAR(20) NOT NULL DEFAULT 'OFFBOARDING'
    CHECK (checklist_type IN ('ONBOARDING','OFFBOARDING')),
  initiated_by   UUID        NOT NULL REFERENCES users(id),
  target_date    DATE,
  items          JSONB       NOT NULL DEFAULT '{"items":[]}',
  status         VARCHAR(20) NOT NULL DEFAULT 'IN_PROGRESS'
    CHECK (status IN ('IN_PROGRESS','COMPLETED','CANCELLED')),
  completed_at   TIMESTAMPTZ,
  notes          TEXT,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_offboarding_employee ON offboarding_checklists(employee_id);
CREATE INDEX idx_offboarding_status   ON offboarding_checklists(status);
```

### 11.24 CREATE expense_claims

```sql
CREATE TABLE IF NOT EXISTS expense_claims (
  id                       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  claim_number             VARCHAR(20) UNIQUE,
  employee_id              UUID        NOT NULL REFERENCES employees(id),
  engagement_id            UUID        REFERENCES engagements(id),
  claim_period_from        DATE,
  claim_period_to          DATE,
  total_amount             NUMERIC(15,2) NOT NULL DEFAULT 0,
  currency                 VARCHAR(3)  NOT NULL DEFAULT 'VND',
  description              TEXT,
  status                   VARCHAR(20) NOT NULL DEFAULT 'DRAFT'
    CHECK (status IN ('DRAFT','SUBMITTED','MANAGER_APPROVED','HR_APPROVED','PAID','REJECTED')),
  submitted_at             TIMESTAMPTZ,
  manager_approver_id      UUID        REFERENCES users(id),
  manager_approved_at      TIMESTAMPTZ,
  manager_rejection_reason TEXT,
  hr_approver_id           UUID        REFERENCES users(id),
  hr_approved_at           TIMESTAMPTZ,
  hr_rejection_reason      TEXT,
  paid_at                  TIMESTAMPTZ,
  payment_reference        VARCHAR(100),
  created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at               TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION fn_expense_claims_set_number()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE v_year TEXT; v_seq INT; BEGIN
  IF NEW.claim_number IS NOT NULL THEN RETURN NEW; END IF;
  v_year := TO_CHAR(now(), 'YY');
  SELECT COUNT(*) + 1 INTO v_seq FROM expense_claims
    WHERE claim_number LIKE 'PC' || v_year || '-%';
  NEW.claim_number := 'PC' || v_year || '-' || LPAD(v_seq::TEXT, 4, '0');
  RETURN NEW;
END; $$;
CREATE TRIGGER trg_expense_claims_set_number
  BEFORE INSERT ON expense_claims FOR EACH ROW EXECUTE FUNCTION fn_expense_claims_set_number();

CREATE INDEX idx_expense_claims_employee   ON expense_claims(employee_id);
CREATE INDEX idx_expense_claims_status     ON expense_claims(status);
CREATE INDEX idx_expense_claims_engagement ON expense_claims(engagement_id) WHERE engagement_id IS NOT NULL;
```

### 11.25 CREATE expense_claim_items

```sql
CREATE TABLE IF NOT EXISTS expense_claim_items (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  claim_id     UUID        NOT NULL REFERENCES expense_claims(id) ON DELETE CASCADE,
  category     VARCHAR(30) NOT NULL
    CHECK (category IN ('FLIGHT','HOTEL','TAXI','MEAL','STATIONERY','COMMUNICATION','OTHER')),
  expense_date DATE        NOT NULL,
  description  VARCHAR(500),
  amount       NUMERIC(12,2) NOT NULL CHECK (amount > 0),
  currency     VARCHAR(3)  NOT NULL DEFAULT 'VND',
  is_billable  BOOLEAN     NOT NULL DEFAULT false,
  receipt_url  TEXT,
  notes        TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_expense_items_claim ON expense_claim_items(claim_id);
```

### 11.26 Seed Data

```sql
-- =============================================
-- Branches
-- =============================================
INSERT INTO branches (code, name, is_head_office, city, is_active) VALUES
  ('HO',  'Trụ sở chính (Hà Nội)',  true,  'Hà Nội',          true),
  ('HCM', 'Chi nhánh TP.HCM',        false, 'TP. Hồ Chí Minh', true)
ON CONFLICT (code) DO NOTHING;

-- =============================================
-- Departments
-- =============================================
INSERT INTO departments (code, name, dept_type, is_active) VALUES
  ('AUDIT', 'Phòng Kiểm toán',             'CORE',    true),
  ('TAX',   'Phòng Thuế',                  'CORE',    true),
  ('HR',    'Phòng Nhân sự',               'SUPPORT', true),
  ('FIN',   'Phòng Tài chính - Kế toán',   'SUPPORT', true),
  ('IT',    'Phòng Công nghệ Thông tin',    'SUPPORT', true)
ON CONFLICT (code) DO NOTHING;

-- =============================================
-- branch_departments matrix
-- =============================================
INSERT INTO branch_departments (branch_id, department_id)
SELECT b.id, d.id FROM branches b, departments d
WHERE (b.code = 'HO')
   OR (b.code = 'HCM' AND d.code IN ('AUDIT', 'TAX'))
ON CONFLICT DO NOTHING;

-- =============================================
-- Insurance rates 2024
-- =============================================
INSERT INTO insurance_rate_config (
  effective_from, bhxh_employee_pct, bhxh_employer_pct,
  bhyt_employee_pct, bhyt_employer_pct, bhtn_employee_pct, bhtn_employer_pct,
  kpcd_employer_pct, salary_base_bhxh, max_bhxh_salary, notes
) VALUES (
  '2024-01-01', 8.00, 17.50, 1.50, 3.00, 1.00, 1.00, 2.00,
  1800000, 36000000, 'Tỷ lệ BHXH + KPCĐ áp dụng từ 01/01/2024'
) ON CONFLICT DO NOTHING;

-- =============================================
-- CPE requirements seed
-- =============================================
INSERT INTO cpe_requirements_by_role (role_code, cert_type, required_hours_per_year, regulatory_body, effective_from)
VALUES
  ('PARTNER',        'VN_CPA', 40.0, 'VACPA',       '2024-01-01'),
  ('SENIOR_AUDITOR', 'VN_CPA', 40.0, 'VACPA',       '2024-01-01'),
  ('JUNIOR_AUDITOR', 'VN_CPA', 40.0, 'VACPA',       '2024-01-01'),
  ('PARTNER',        'ACCA',   40.0, 'ACCA Global', '2024-01-01'),
  ('SENIOR_AUDITOR', 'ACCA',   40.0, 'ACCA Global', '2024-01-01'),
  ('PARTNER',        'CIA',    40.0, 'IIA',          '2024-01-01'),
  ('PARTNER',        'CFA',    20.0, 'CFA Institute','2024-01-01')
ON CONFLICT DO NOTHING;

-- =============================================
-- Holidays 2026
-- =============================================
INSERT INTO holidays (holiday_date, name, name_en, type) VALUES
  ('2026-01-01', 'Tết Dương lịch',                  'New Year''s Day',              'NATIONAL'),
  ('2026-01-28', 'Tết Nguyên Đán (28 tháng Chạp)',  'Lunar New Year Eve',           'NATIONAL'),
  ('2026-01-29', 'Tết Nguyên Đán (Mùng 1)',          'Lunar New Year Day 1',         'NATIONAL'),
  ('2026-01-30', 'Tết Nguyên Đán (Mùng 2)',          'Lunar New Year Day 2',         'NATIONAL'),
  ('2026-01-31', 'Tết Nguyên Đán (Mùng 3)',          'Lunar New Year Day 3',         'NATIONAL'),
  ('2026-02-01', 'Tết Nguyên Đán (Mùng 4)',          'Lunar New Year Day 4',         'NATIONAL'),
  ('2026-02-02', 'Tết Nguyên Đán (Mùng 5)',          'Lunar New Year Day 5',         'NATIONAL'),
  ('2026-04-07', 'Giỗ Tổ Hùng Vương',                'Hung King''s Commemoration',   'NATIONAL'),
  ('2026-04-30', 'Ngày Giải phóng Miền Nam',          'Liberation Day',               'NATIONAL'),
  ('2026-05-01', 'Ngày Quốc tế Lao động',             'International Labor Day',      'NATIONAL'),
  ('2026-09-02', 'Ngày Quốc khánh',                   'National Day',                 'NATIONAL'),
  ('2026-09-03', 'Ngày Quốc khánh (nghỉ bù)',         'National Day (compensated)',   'NATIONAL')
ON CONFLICT (holiday_date) DO NOTHING;
```

---

## 12. Migration Plan

### 12.1 Overview

| Migration | Tên | Nội dung |
|---|---|---|
| 000019 | hrm_organization | ALTER branches, departments; CREATE branch_departments |
| 000020 | hrm_employees_extended | ALTER employees (40+ cols); triggers; CREATE dependents, salary_history, contracts, insurance_rates |
| 000021 | hrm_professional | CREATE certifications, training_courses, training_records, cpe_requirements |
| 000022 | hrm_performance | CREATE performance_reviews, peer_reviews, independence_declarations |
| 000023 | hrm_time_leave | CREATE holidays, leave_balances, leave_requests; ALTER timesheets; CREATE ot_requests, VIEW |
| 000024 | hrm_provisioning | CREATE user_provisioning_requests, offboarding_checklists |
| 000025 | hrm_expenses | CREATE expense_claims (+ trigger), expense_claim_items |
| 000026 | hrm_seed_data | Seed branches, departments, matrix, insurance rates, CPE reqs, holidays 2026–2030 |

### 12.2 Migration 000019: hrm_organization

**File:** `000019_hrm_organization.up.sql` / `000019_hrm_organization.down.sql`

**Up:** ALTER branches (add HRM fields), ALTER departments (add code/dept_type), CREATE branch_departments, uidx_branches_head_office.

**Down:**
```sql
DROP TABLE IF EXISTS branch_departments;
ALTER TABLE departments
  DROP COLUMN IF EXISTS code, DROP COLUMN IF EXISTS dept_type,
  DROP COLUMN IF EXISTS description, DROP COLUMN IF EXISTS is_active;
ALTER TABLE branches
  DROP COLUMN IF EXISTS is_head_office, DROP COLUMN IF EXISTS city,
  DROP COLUMN IF EXISTS address, DROP COLUMN IF EXISTS phone,
  DROP COLUMN IF EXISTS established_date, DROP COLUMN IF EXISTS head_of_branch_user_id,
  DROP COLUMN IF EXISTS is_active;
```

**Dependencies:** branches, departments (existing)
**Rollback risk:** Low — no FK references yet. Seed data (code values) lost on down.

### 12.3 Migration 000020: hrm_employees_extended

**File:** `000020_hrm_employees_extended.up.sql` / `000020_hrm_employees_extended.down.sql`

**Up:** All employee column additions, fn_employees_set_code trigger, CREATE employee_dependents, insurance_rate_config, employee_salary_history (with immutable rules), employment_contracts. All indexes.

**Down:**
```sql
DROP TRIGGER IF EXISTS trg_employees_set_code ON employees;
DROP FUNCTION IF EXISTS fn_employees_set_code();
DROP TABLE IF EXISTS employment_contracts;
DROP TABLE IF EXISTS employee_salary_history;
DROP TABLE IF EXISTS insurance_rate_config;
DROP TABLE IF EXISTS employee_dependents;
-- DROP all added columns from employees (all 40+ columns listed explicitly)
```

**Dependencies:** 000019
**Rollback risk:** High — employee data loss. Never rollback on production.

### 12.4 Migration 000021: hrm_professional

**Up:** CREATE certifications, training_courses, training_records, cpe_requirements_by_role.

**Down:**
```sql
DROP TABLE IF EXISTS cpe_requirements_by_role;
DROP TABLE IF EXISTS training_records;
DROP TABLE IF EXISTS training_courses;
DROP TABLE IF EXISTS certifications;
```

**Dependencies:** 000020 (employees)

### 12.5 Migration 000022: hrm_performance

**Up:** CREATE performance_reviews, engagement_peer_reviews, independence_declarations.

**Down:**
```sql
DROP TABLE IF EXISTS independence_declarations;
DROP TABLE IF EXISTS engagement_peer_reviews;
DROP TABLE IF EXISTS performance_reviews;
```

**Dependencies:** 000020 (employees), engagements (existing)

### 12.6 Migration 000023: hrm_time_leave

**Up:** CREATE holidays, leave_balances, leave_requests; ALTER timesheets (add ot_hours, ot_approved, ot_request_id); CREATE ot_requests; CREATE VIEW employee_ot_summary_year.

**Down:**
```sql
DROP VIEW IF EXISTS employee_ot_summary_year;
DROP TABLE IF EXISTS ot_requests;
ALTER TABLE timesheets
  DROP COLUMN IF EXISTS ot_hours,
  DROP COLUMN IF EXISTS ot_approved,
  DROP COLUMN IF EXISTS ot_request_id;
DROP TABLE IF EXISTS leave_requests;
DROP TABLE IF EXISTS leave_balances;
DROP TABLE IF EXISTS holidays;
```

**Dependencies:** 000020 (employees), timesheets (existing)
**Rollback risk:** Holiday seed data lost. Re-seed required after re-up.

### 12.7 Migration 000024: hrm_provisioning

**Up:** CREATE user_provisioning_requests, offboarding_checklists.

**Down:**
```sql
DROP TABLE IF EXISTS offboarding_checklists;
DROP TABLE IF EXISTS user_provisioning_requests;
```

**Dependencies:** 000020, users (existing)

### 12.8 Migration 000025: hrm_expenses

**Up:** CREATE expense_claims + trigger fn_expense_claims_set_number, CREATE expense_claim_items.

**Down:**
```sql
DROP TABLE IF EXISTS expense_claim_items;
DROP TRIGGER IF EXISTS trg_expense_claims_set_number ON expense_claims;
DROP FUNCTION IF EXISTS fn_expense_claims_set_number();
DROP TABLE IF EXISTS expense_claims;
```

**Dependencies:** 000020, engagements (existing)

### 12.9 Migration 000026: hrm_seed_data

**Up:** INSERT branches (ON CONFLICT DO NOTHING), INSERT departments, INSERT branch_departments matrix, INSERT insurance_rate_config, INSERT cpe_requirements_by_role, INSERT holidays 2026–2030.

**Down:**
```sql
DELETE FROM holidays WHERE year BETWEEN 2026 AND 2030 AND type = 'NATIONAL';
DELETE FROM cpe_requirements_by_role WHERE effective_from = '2024-01-01';
-- Leave branches/departments intact (FK risk)
```

**Dependencies:** 000019, 000021, 000023
**Note:** Use ON CONFLICT DO NOTHING for idempotency — safe to re-run.

---

## 13. API Endpoints Catalog

### Conventions

- Base path: `/api/v1`
- Auth: JWT Bearer (all endpoints unless noted)
- Collections: `{ "data": [...], "meta": { "page": 1, "size": 20, "total": 100 } }`
- Single items: `{ "data": {...} }`
- Errors: `{ "error": "ERROR_CODE", "message": "..." }`

### 13.1 Organization

| Method | Path | Role | Mô tả |
|---|---|---|---|
| GET | /hrm/organization/branches | ALL | Danh sách chi nhánh |
| GET | /hrm/organization/branches/:id | ALL | Chi tiết chi nhánh |
| PUT | /hrm/organization/branches/:id | SA, CHAIRMAN | Cập nhật chi nhánh |
| GET | /hrm/organization/departments | ALL | Danh sách phòng ban |
| GET | /hrm/organization/departments/:id | ALL | Chi tiết phòng ban |
| PUT | /hrm/organization/departments/:id | SA, CHAIRMAN, CEO | Cập nhật phòng ban |
| GET | /hrm/organization/branch-departments | ALL | Ma trận chi nhánh-phòng |
| POST | /hrm/organization/branch-departments | SA, CHAIRMAN | Thêm phòng vào chi nhánh |
| DELETE | /hrm/organization/branch-departments/:id | SA, CHAIRMAN | Gỡ phòng khỏi chi nhánh |
| GET | /hrm/organization/org-chart | ALL | Sơ đồ tổ chức (tree) |

### 13.2 Employees

| Method | Path | Role | Mô tả |
|---|---|---|---|
| GET | /hrm/employees | HR, CEO, CHAIRMAN, SA, HoB | Danh sách nhân viên |
| POST | /hrm/employees | HR_MANAGER, CEO, SA | Tạo hồ sơ mới |
| GET | /hrm/employees/:id | HR, CEO, CHAIRMAN, SA, HoB, PARTNER(dept), self | Chi tiết |
| PUT | /hrm/employees/:id | HR_MANAGER, CEO, SA | Cập nhật |
| DELETE | /hrm/employees/:id | SA | Soft delete |
| GET | /hrm/employees/:id/sensitive | HR_MANAGER, CEO, CHAIRMAN | PII fields (always audit logged) |
| POST | /hrm/employees/:id/terminate | HR_MANAGER, CEO | Nghỉ việc |
| GET | /my-profile | ALL | Hồ sơ cá nhân |
| PUT | /my-profile | ALL | Cập nhật thông tin cá nhân (limited fields) |

**Errors:** `EMPLOYEE_NOT_FOUND`, `DUPLICATE_EMPLOYEE_CODE`, `BRANCH_NOT_FOUND`, `DEPARTMENT_NOT_FOUND`, `INSUFFICIENT_PERMISSION`

### 13.3 Dependents

| Method | Path | Role | Mô tả |
|---|---|---|---|
| GET | /hrm/employees/:id/dependents | HR, CEO, CHAIRMAN, self | Danh sách |
| POST | /hrm/employees/:id/dependents | HR_MANAGER, self | Thêm |
| PUT | /hrm/employees/:id/dependents/:dep_id | HR_MANAGER, self | Cập nhật |
| DELETE | /hrm/employees/:id/dependents/:dep_id | HR_MANAGER | Xóa |

### 13.4 Certifications

| Method | Path | Role | Mô tả |
|---|---|---|---|
| GET | /hrm/employees/:id/certifications | HR, CEO, PARTNER, HoB, self | Danh sách |
| POST | /hrm/employees/:id/certifications | HR_MANAGER, self | Thêm |
| PUT | /hrm/employees/:id/certifications/:cert_id | HR_MANAGER, self | Cập nhật |
| DELETE | /hrm/employees/:id/certifications/:cert_id | HR_MANAGER | Xóa |
| GET | /hrm/certifications/expiring | HR, CEO, PARTNER | Sắp hết hạn (default 60d) |

**Query params:** `?days=60&branch_id=uuid&cert_type=VN_CPA`

### 13.5 Training & CPE

| Method | Path | Role | Mô tả |
|---|---|---|---|
| GET | /hrm/training/courses | ALL | Catalog khóa đào tạo |
| POST | /hrm/training/courses | HR_MANAGER, SA | Tạo khóa |
| PUT | /hrm/training/courses/:id | HR_MANAGER | Cập nhật |
| GET | /hrm/employees/:id/training | HR, CEO, PARTNER, self | Hồ sơ đào tạo |
| POST | /hrm/employees/:id/training | HR_MANAGER, self | Đăng ký |
| PUT | /hrm/employees/:id/training/:rec_id | HR_MANAGER | Cập nhật kết quả |
| GET | /hrm/employees/:id/cpe-summary | HR, CEO, PARTNER, self | CPE của 1 nhân viên |
| GET | /hrm/training/cpe-summary | HR, CEO, CHAIRMAN | CPE toàn công ty |

### 13.6 Performance Reviews

| Method | Path | Role | Mô tả |
|---|---|---|---|
| GET | /hrm/performance/reviews | HR, CEO, CHAIRMAN | Tất cả (filter: period, status) |
| GET | /hrm/employees/:id/performance | HR, CEO, PARTNER, HoB, self | Reviews của nhân viên |
| POST | /hrm/employees/:id/performance | HR_MANAGER, PARTNER, AUDIT_MANAGER | Tạo review |
| PUT | /hrm/performance/reviews/:id | reviewer | Cập nhật draft |
| POST | /hrm/performance/reviews/:id/submit | reviewer | Submit |
| POST | /hrm/performance/reviews/:id/acknowledge | employee | Xác nhận |
| GET | /hrm/performance/peer-reviews | HR, CEO, PARTNER | Peer reviews |
| POST | /hrm/performance/peer-reviews | PARTNER, SENIOR, JUNIOR | Tạo peer review |

### 13.7 Independence Declarations

| Method | Path | Role | Mô tả |
|---|---|---|---|
| GET | /hrm/independence | HR, CEO, CHAIRMAN, PARTNER | Tất cả |
| GET | /hrm/employees/:id/independence | HR, CEO, PARTNER, self | Của nhân viên |
| POST | /hrm/independence | ALL | Tạo khai báo |
| GET | /hrm/independence/:id | HR, CEO, CHAIRMAN, PARTNER, owner | Chi tiết |
| POST | /hrm/independence/:id/acknowledge | PARTNER | Xác nhận conflict |
| GET | /hrm/independence/annual-status | HR, CEO, CHAIRMAN | Tổng hợp năm hiện tại |

### 13.8 Leave

| Method | Path | Role | Mô tả |
|---|---|---|---|
| GET | /hrm/leave/balances | HR, CEO, HoB | Tất cả cân đối |
| GET | /hrm/employees/:id/leave-balance | HR, CEO, HoB, manager, self | Cân đối của nhân viên |
| PUT | /hrm/employees/:id/leave-balance | HR_MANAGER | Điều chỉnh cân đối |
| GET | /hrm/leave/requests | HR, CEO, HoB | Tất cả đơn |
| GET | /my-leave/requests | ALL | Đơn của tôi |
| POST | /my-leave/requests | ALL | Nộp đơn |
| POST | /hrm/leave/requests/:id/approve | manager, HR_MANAGER | Duyệt |
| POST | /hrm/leave/requests/:id/reject | manager, HR_MANAGER | Từ chối |
| POST | /my-leave/requests/:id/cancel | owner | Hủy (PENDING only) |
| GET | /my-team/leave | PARTNER, HoB, CEO | Lịch nghỉ team |
| GET | /hrm/leave/calendar | ALL | Lịch nghỉ toàn công ty |

### 13.9 Overtime

| Method | Path | Role | Mô tả |
|---|---|---|---|
| GET | /hrm/overtime/requests | HR, CEO, HoB | Tất cả |
| GET | /my-overtime/requests | ALL | Của tôi |
| POST | /my-overtime/requests | ALL | Đăng ký OT |
| POST | /hrm/overtime/requests/:id/approve | manager, HR_MANAGER | Duyệt |
| POST | /hrm/overtime/requests/:id/reject | manager, HR_MANAGER | Từ chối |
| GET | /hrm/overtime/summary | HR, CEO, CHAIRMAN | Tổng hợp theo năm |
| GET | /my-overtime/summary | ALL | Summary của tôi |

**Error:** `OT_ANNUAL_CAP_EXCEEDED` (khi approved_hours + ot_hours > 300)

### 13.10 Holidays

| Method | Path | Role | Mô tả |
|---|---|---|---|
| GET | /hrm/holidays | ALL | Danh sách (?year=2026) |
| POST | /hrm/holidays | SA, HR_MANAGER | Thêm |
| PUT | /hrm/holidays/:id | SA, HR_MANAGER | Cập nhật |
| DELETE | /hrm/holidays/:id | SA | Xóa |

### 13.11 Contracts

| Method | Path | Role | Mô tả |
|---|---|---|---|
| GET | /hrm/employees/:id/contracts | HR, CEO, CHAIRMAN, self | Lịch sử hợp đồng |
| POST | /hrm/employees/:id/contracts | HR_MANAGER | Tạo mới |
| PUT | /hrm/employees/:id/contracts/:cid | HR_MANAGER | Cập nhật |
| POST | /hrm/employees/:id/contracts/:cid/set-current | HR_MANAGER | Đặt làm hợp đồng hiện tại |
| GET | /hrm/contracts/expiring | HR, CEO, CHAIRMAN | Sắp hết hạn (?days=30) |

### 13.12 User Provisioning Requests

| Method | Path | Role | Mô tả |
|---|---|---|---|
| GET | /hrm/user-provisioning-requests | SA, HR_MANAGER, HoB | Danh sách |
| POST | /hrm/user-provisioning-requests | HR_MANAGER, HoB, CEO | Tạo request |
| GET | /hrm/user-provisioning-requests/:id | SA, HR_MANAGER, HoB | Chi tiết |
| POST | /hrm/user-provisioning-requests/:id/branch-approve | HoB | Approve bước 1 |
| POST | /hrm/user-provisioning-requests/:id/branch-reject | HoB | Reject bước 1 |
| POST | /hrm/user-provisioning-requests/:id/hr-approve | HR_MANAGER | Approve bước 2 |
| POST | /hrm/user-provisioning-requests/:id/hr-reject | HR_MANAGER | Reject bước 2 |
| POST | /hrm/user-provisioning-requests/:id/execute | SA | Thực thi tạo account |
| POST | /hrm/user-provisioning-requests/:id/cancel | requester, SA | Hủy |

**Errors:** `DUPLICATE_PENDING_REQUEST`, `INVALID_ROLE_FOR_PROVISIONING`, `REQUEST_EXPIRED`

### 13.13 Offboarding

| Method | Path | Role | Mô tả |
|---|---|---|---|
| GET | /hrm/offboarding | HR, CEO, SA | Danh sách |
| POST | /hrm/offboarding | HR_MANAGER | Khởi tạo |
| GET | /hrm/offboarding/:id | HR, CEO, SA | Chi tiết |
| PUT | /hrm/offboarding/:id/items/:key | HR, IT, FIN | Cập nhật item |
| POST | /hrm/offboarding/:id/complete | HR_MANAGER | Đánh dấu hoàn thành |

### 13.14 Expenses

| Method | Path | Role | Mô tả |
|---|---|---|---|
| GET | /hrm/expenses | HR, CEO, CHAIRMAN | Tất cả claims |
| GET | /my-expenses | ALL | Claims của tôi |
| POST | /my-expenses | ALL | Tạo claim (DRAFT) |
| PUT | /my-expenses/:id | owner (DRAFT) | Cập nhật |
| POST | /my-expenses/:id/items | owner (DRAFT) | Thêm item |
| DELETE | /my-expenses/:id/items/:item_id | owner (DRAFT) | Xóa item |
| POST | /my-expenses/:id/submit | owner | Submit |
| POST | /hrm/expenses/:id/manager-approve | manager | Duyệt cấp 1 |
| POST | /hrm/expenses/:id/manager-reject | manager | Từ chối cấp 1 |
| POST | /hrm/expenses/:id/hr-approve | HR_MANAGER | Duyệt cấp 2 |
| POST | /hrm/expenses/:id/hr-reject | HR_MANAGER | Từ chối cấp 2 |
| POST | /hrm/expenses/:id/mark-paid | HR_MANAGER, ACCOUNTANT | Đánh dấu thanh toán |
| GET | /hrm/expenses/summary | HR, CEO, CHAIRMAN | Tổng hợp |

### 13.15 Salary History

| Method | Path | Role | Mô tả |
|---|---|---|---|
| GET | /hrm/employees/:id/salary-history | HR_MANAGER, CEO, CHAIRMAN | Lịch sử lương |
| POST | /hrm/employees/:id/salary-history | HR_MANAGER, CEO | Ghi nhận thay đổi lương |

### 13.16 Insurance Config

| Method | Path | Role | Mô tả |
|---|---|---|---|
| GET | /hrm/insurance-config | HR, CEO, CHAIRMAN | Danh sách cấu hình |
| GET | /hrm/insurance-config/current | HR, CEO, CHAIRMAN | Cấu hình hiện tại |
| POST | /hrm/insurance-config | SA, HR_MANAGER | Thêm cấu hình mới |

---

## 14. UI Pages

### 14.1 Admin Pages

| Path | Tên | Role | Mô tả |
|---|---|---|---|
| /admin/hrm/organization | Cơ cấu Tổ chức | SA, CHAIRMAN, CEO | Sơ đồ org chart, CRUD branches/depts |
| /admin/hrm/employees | Quản lý Nhân viên | SA, CEO, HR_MGR | Danh sách + filter + export |
| /admin/hrm/employees/new | Thêm Nhân viên | SA, CEO, HR_MGR | Form tạo mới |
| /admin/hrm/employees/:id | Hồ sơ Nhân viên | SA, CEO, HR_MGR | Tabs: Basic, PII, Salary, Contract, Insurance |
| /admin/hrm/employees/:id/edit | Sửa Hồ sơ | SA, CEO, HR_MGR | Form chỉnh sửa |
| /admin/hrm/provisioning | Cấp Tài khoản | SA, HR_MGR | Danh sách + execute |
| /admin/hrm/insurance-config | Cấu hình BHXH | SA, HR_MGR | Lịch sử tỷ lệ |
| /admin/hrm/holidays | Ngày Lễ | SA, HR_MGR | Calendar CRUD |
| /admin/hrm/reports | Báo cáo HRM | SA, CEO, CHAIRMAN | Dashboard 10 báo cáo |
| /admin/hrm/offboarding | Offboarding | SA, HR_MGR | Active checklists |

### 14.2 HRM Module Pages

| Path | Tên | Role | Mô tả |
|---|---|---|---|
| /hrm/employees | Nhân viên | HR, HoB, PARTNER | Danh sách (branch-scoped cho HoB) |
| /hrm/employees/:id | Hồ sơ | HR, HoB, PARTNER | Chi tiết (fields theo role) |
| /hrm/leave/requests | Đơn Nghỉ phép | HR, HoB, PARTNER | Cần duyệt |
| /hrm/leave/calendar | Lịch Nghỉ | HR, HoB | Calendar toàn chi nhánh |
| /hrm/overtime/requests | Đơn Tăng ca | HR, HoB, PARTNER | Cần duyệt |
| /hrm/performance | Đánh giá Hiệu suất | HR, CEO | Quản lý theo period |
| /hrm/certifications | Chứng chỉ | HR, CEO, PARTNER | Tổng hợp + expiring alerts |
| /hrm/training | Đào tạo & CPE | HR, CEO | Khóa học + CPE dashboard |
| /hrm/independence | Khai báo Độc lập | HR, CEO, CHAIRMAN | Compliance tracking |
| /hrm/expenses | Chi phí | HR, CEO | Cần duyệt |
| /hrm/contracts/expiring | HĐ Sắp Hết Hạn | HR, CEO | Alert list |
| /hrm/provisioning | Cấp Tài khoản | HR, HoB | Requests flow |

### 14.3 Self-Service Pages

| Path | Tên | Mô tả |
|---|---|---|
| /my-profile | Hồ sơ Cá nhân | Xem + cập nhật limited fields |
| /my-profile/certifications | Chứng chỉ Cá nhân | Xem + thêm chứng chỉ |
| /my-profile/training | Đào tạo Cá nhân | Lịch sử + CPE progress |
| /my-profile/contracts | Hợp đồng Cá nhân | Xem lịch sử |
| /my-dependents | Người Phụ thuộc | Quản lý người PT thuế |
| /my-leave | Nghỉ phép | Cân đối + lịch sử đơn |
| /my-leave/new | Xin Nghỉ phép | Form nộp đơn |
| /my-overtime | Tăng ca | OT summary + lịch sử |
| /my-overtime/new | Đăng ký OT | Form đăng ký |
| /my-expenses | Chi phí Công tác | Danh sách claims |
| /my-expenses/new | Khai báo Chi phí | Form + items |
| /my-independence | Khai báo Độc lập | Danh sách + form |
| /my-performance | Đánh giá | Reviews của mình |

### 14.4 Team Management Pages

| Path | Tên | Role | Mô tả |
|---|---|---|---|
| /my-team | Nhóm của Tôi | PARTNER, HoB, CEO | Danh sách subordinates |
| /my-team/leave | Lịch Nghỉ Nhóm | PARTNER, HoB | Calendar |
| /my-team/performance | Đánh giá Nhóm | PARTNER, HoB, CEO | Reviews tôi cần làm |
| /my-team/overtime | OT Nhóm | PARTNER, HoB | Requests cần duyệt |
| /my-team/expenses | Chi phí Nhóm | PARTNER, HoB | Claims cần duyệt |
| /my-team/independence | Độc lập Nhóm | PARTNER | Status khai báo |

---

## 15. Permission Matrix

### 15.1 Ký Hiệu

| Ký hiệu | Nghĩa |
|---|---|
| ✅ | Full CRUD access |
| 🔍 | Read only |
| 📝 | Create/Update own records |
| 👁 | Read own only |
| ❌ | Không có quyền |
| 🏢 | Scoped by branch |
| 👥 | Scoped by department |

### 15.2 Ma Trận Quyền Chính

| Entity | SA | CHAIRMAN | CEO | HR_MGR | HR_STAFF | HoB | PARTNER | AUDIT_MGR | SENIOR | JUNIOR | ACCOUNTANT |
|---|---|---|---|---|---|---|---|---|---|---|---|
| **Branches (write)** | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Branches (read)** | ✅ | ✅ | 🔍 | 🔍 | 🔍 | 🔍 | 🔍 | 🔍 | 🔍 | 🔍 | 🔍 |
| **Depts (write)** | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Employee create** | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Employee read** | ✅ | ✅ | ✅ | ✅ | 🏢 | 🏢 | 👥 | 👥 | 👁 | 👁 | 👁 |
| **Employee update** | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Salary fields** | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **PII (CCCD,MST,bank)** | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **BHXH data** | ✅ | 🔍 | 🔍 | ✅ | 🏢 | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Dependents read** | ✅ | ✅ | ✅ | ✅ | 🏢 | ❌ | ❌ | ❌ | ❌ | 👁 | 👁 |
| **Dependents write** | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | 📝 | 📝 |
| **Certifications read** | ✅ | ✅ | ✅ | ✅ | 🏢 | 🏢 | 👥 | 👥 | 👁 | 👁 | 👁 |
| **Certifications write** | ✅ | ❌ | ❌ | ✅ | 🏢 | ❌ | ❌ | ❌ | 📝 | 📝 | 📝 |
| **Training read** | ✅ | ✅ | ✅ | ✅ | 🏢 | 🏢 | 👥 | 👥 | 👁 | 👁 | 👁 |
| **Training write** | ✅ | ❌ | ❌ | ✅ | 🏢 | ❌ | ❌ | ❌ | 📝 | 📝 | 📝 |
| **Performance create** | ✅ | ❌ | ✅ | ✅ | ❌ | 🏢 | 👥 | 👥 | ❌ | ❌ | ❌ |
| **Performance read** | ✅ | ✅ | ✅ | ✅ | ❌ | 🏢 | 👥 | 👥 | 👁 | 👁 | 👁 |
| **Independence create** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Independence read** | ✅ | ✅ | ✅ | ✅ | ❌ | 🏢 | 👥 | 👥 | 👁 | 👁 | 👁 |
| **Leave create** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Leave approve** | ✅ | ❌ | ✅ | ✅ | ❌ | 🏢 | 👥 | 👥 | ❌ | ❌ | ❌ |
| **OT create** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **OT approve** | ✅ | ❌ | ✅ | ✅ | ❌ | 🏢 | 👥 | 👥 | ❌ | ❌ | ❌ |
| **Provisioning request** | ✅ | ❌ | ✅ | ✅ | ❌ | 🏢 | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Provisioning approve** | ✅ | ❌ | ✅ | ✅ | ❌ | 🏢 (step1) | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Provisioning execute** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Offboarding** | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Expenses create** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Expenses approve L1** | ✅ | ❌ | ✅ | ✅ | ❌ | 🏢 | 👥 | 👥 | ❌ | ❌ | ❌ |
| **Expenses approve L2** | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Expenses mark paid** | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| **Salary history read** | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Salary history write** | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Insurance config** | ✅ | 🔍 | 🔍 | ✅ | 🔍 | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Holidays write** | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **HRM Reports** | ✅ | ✅ | ✅ | ✅ | 🏢 | 🏢 | ❌ | 👥 | ❌ | ❌ | ❌ |

### 15.3 Branch Scope Enforcement

**Rule:** API middleware checks `employee.branch_id == caller.branch_id` for HoB and HR_STAFF roles.

```
HoB HCM:
  - GET /hrm/employees → WHERE branch_id = HCM_ID
  - POST /hrm/leave/requests/:id/approve → 403 nếu employee.branch_id != HCM_ID
  - GET /hrm/organization/branches → returns only HCM branch
  - CANNOT see HO employees' salary, PII, or contracts
```

### 15.4 Sensitive Field Masking

Khi caller không có quyền xem sensitive fields, API trả về:

```json
{
  "base_salary": null,
  "bank_account_encrypted": "***",
  "mst_ca_nhan_encrypted": "***",
  "cccd_encrypted": "***",
  "so_bhxh_encrypted": "***"
}
```

GET `/hrm/employees/:id/sensitive` (HR_MANAGER, CEO, CHAIRMAN only):
- Decrypt và trả về plain text
- **LUÔN ghi audit log** kể cả khi không có mutation
- Log: `{ action: "PII_ACCESSED", fields: ["cccd", "mst", "bank_account"], employee_id, accessor_id, ip }`


---

## 16. Notifications & Alerts

### 16.1 Event Catalog

| # | Event | Trigger | Recipients | Channel |
|---|---|---|---|---|
| 1 | Hợp đồng sắp hết hạn (30 ngày) | Daily cron | HR_MANAGER, manager trực tiếp | Email + In-app |
| 2 | Chứng chỉ sắp hết hạn | Daily cron (60d/30d/7d trước) | Employee, HR_MANAGER | Email + In-app |
| 3 | Đơn nghỉ phép mới cần duyệt | Leave request submitted | Manager trực tiếp | In-app + Email |
| 4 | Đơn nghỉ phép được duyệt | Leave approved | Employee | In-app + Email |
| 5 | Đơn nghỉ phép bị từ chối | Leave rejected | Employee | In-app + Email |
| 6 | Đơn OT mới cần duyệt | OT request submitted | Manager trực tiếp | In-app |
| 7 | Đơn OT được duyệt/từ chối | OT status change | Employee | In-app |
| 8 | Cảnh báo OT gần đạt 300h/năm | OT approved (>250h) | Employee, HR_MANAGER | In-app + Email |
| 9 | Provisioning request cần duyệt | Provisioning created | HoB (HCM flow) / HR_MANAGER (HO) | In-app + Email |
| 10 | Provisioning approved bước 1 | Branch approve | HR_MANAGER | In-app + Email |
| 11 | Provisioning approved bước 2 | HR approve | SA | In-app + Email |
| 12 | Provisioning rejected | Any rejection | Requester, relevant approver | In-app + Email |
| 13 | Account đã được tạo | Provisioning executed | Employee (welcome email) | Email |
| 14 | Chi phí cần duyệt | Expense submitted | Manager trực tiếp | In-app + Email |
| 15 | Chi phí được duyệt/từ chối | Expense status change | Employee | In-app + Email |
| 16 | Chi phí đã thanh toán | Expense marked paid | Employee | In-app + Email |
| 17 | Khai báo độc lập có conflict | Independence submitted với has_conflict=true | PARTNER phụ trách | In-app + Email |
| 18 | Nhắc khai báo độc lập năm | Annual (tháng 1, nhắc lại tháng 3 nếu chưa khai) | Tất cả employees | Email |
| 19 | CPE thiếu (31/10 hàng năm) | Annual cron | Employee thiếu giờ, HR_MANAGER | Email |
| 20 | Performance review cần submit | Review period end - 14 ngày | Reviewer | In-app + Email |
| 21 | Performance review cần acknowledge | Review submitted | Employee được đánh giá | In-app + Email |
| 22 | Offboarding checklist còn hạng mục | Daily cron (target_date - 3 ngày) | Assigned owner của item | In-app |
| 23 | Nhân viên mới (onboarding started) | Employee created | HR_MANAGER | In-app + Email |
| 24 | Lương được thay đổi | Salary history added | Employee | Email (không hiển thị số tiền) |
| 25 | BHYT sắp hết hạn (30 ngày) | Daily cron | Employee, HR_STAFF | In-app + Email |

### 16.2 Notification Structure

Mỗi notification trong database `notifications` table:

```json
{
  "id": "uuid",
  "user_id": "uuid",
  "type": "CONTRACT_EXPIRING",
  "title": "Hợp đồng sắp hết hạn",
  "body": "Hợp đồng của Nguyễn Văn A sẽ hết hạn vào 15/06/2026",
  "entity_type": "employment_contract",
  "entity_id": "uuid",
  "action_url": "/hrm/employees/uuid/contracts",
  "is_read": false,
  "created_at": "2026-05-16T08:00:00Z"
}
```

### 16.3 Email Templates

| Template | Subject | Language |
|---|---|---|
| `welcome_employee` | "Chào mừng bạn đến với [Company] — Thông tin tài khoản" | Vietnamese |
| `contract_expiring` | "Cảnh báo: Hợp đồng lao động sắp hết hạn" | Vietnamese |
| `cert_expiring` | "Nhắc nhở: Chứng chỉ [cert_name] sắp hết hạn" | Vietnamese |
| `leave_approved` | "Đơn nghỉ phép đã được duyệt" | Vietnamese |
| `leave_rejected` | "Đơn nghỉ phép bị từ chối" | Vietnamese |
| `expense_paid` | "Chi phí công tác đã được thanh toán" | Vietnamese |
| `cpe_reminder` | "Nhắc nhở: CPE hours chưa đủ — [X/40h]" | Vietnamese |
| `independence_annual` | "Nhắc nhở: Khai báo độc lập năm [YEAR] chưa hoàn thành" | Vietnamese |

---

## 17. Audit Log Events

### 17.1 Overview

Tất cả các sự kiện sau phải được ghi vào bảng `audit_logs` hiện tại của hệ thống.

Format chuẩn của audit log:

```json
{
  "id": "uuid",
  "timestamp": "2026-04-20T10:30:00Z",
  "actor_id": "uuid",
  "actor_role": "HR_MANAGER",
  "action": "EMPLOYEE_PII_ACCESSED",
  "entity_type": "employee",
  "entity_id": "uuid",
  "before": null,
  "after": null,
  "metadata": {
    "fields_accessed": ["cccd", "mst_ca_nhan", "bank_account"],
    "reason": "onboarding"
  },
  "ip_address": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "branch_id": "uuid"
}
```

**Lưu ý bảo mật:** `before` và `after` cho các trường PII KHÔNG lưu plain text — chỉ lưu `"[ENCRYPTED]"` hoặc indicator rằng field đã thay đổi.

### 17.2 Events Phải Ghi Log

#### 17.2.1 PII & Sensitive Data

| Action Code | Trigger | Fields |
|---|---|---|
| `EMPLOYEE_PII_ACCESSED` | GET /hrm/employees/:id/sensitive | fields_accessed |
| `EMPLOYEE_SALARY_VIEWED` | Xem salary fields bởi authorized user | — |
| `EMPLOYEE_SALARY_CHANGED` | Thêm vào salary_history | before_salary, after_salary (masked) |
| `EMPLOYEE_BANK_UPDATED` | Cập nhật bank_account_encrypted | — (không log giá trị) |
| `EMPLOYEE_CCCD_UPDATED` | Cập nhật cccd_encrypted | — (không log giá trị) |
| `EMPLOYEE_MST_UPDATED` | Cập nhật mst_ca_nhan_encrypted | — |

#### 17.2.2 Employee Lifecycle

| Action Code | Trigger |
|---|---|
| `EMPLOYEE_CREATED` | POST /hrm/employees |
| `EMPLOYEE_UPDATED` | PUT /hrm/employees/:id |
| `EMPLOYEE_TERMINATED` | POST /hrm/employees/:id/terminate |
| `EMPLOYEE_STATUS_CHANGED` | Bất kỳ thay đổi status |
| `EMPLOYEE_GRADE_CHANGED` | Thay đổi grade/position |
| `EMPLOYEE_BRANCH_TRANSFERRED` | Thuyên chuyển chi nhánh |
| `CONTRACT_CREATED` | Tạo hợp đồng mới |
| `CONTRACT_SET_CURRENT` | Đặt làm hợp đồng hiện tại |

#### 17.2.3 Provisioning

| Action Code | Trigger |
|---|---|
| `PROVISIONING_REQUESTED` | Tạo request |
| `PROVISIONING_BRANCH_APPROVED` | HoB approve bước 1 |
| `PROVISIONING_BRANCH_REJECTED` | HoB reject |
| `PROVISIONING_HR_APPROVED` | HR approve bước 2 |
| `PROVISIONING_HR_REJECTED` | HR reject |
| `PROVISIONING_EXECUTED` | SA thực thi — ghi rõ user_id được tạo, role được gán |
| `PROVISIONING_EMERGENCY` | Emergency create với is_emergency=true |

#### 17.2.4 Leave & OT

| Action Code | Trigger |
|---|---|
| `LEAVE_REQUEST_SUBMITTED` | Submit leave |
| `LEAVE_REQUEST_APPROVED` | Approve |
| `LEAVE_REQUEST_REJECTED` | Reject |
| `LEAVE_BALANCE_ADJUSTED` | HR điều chỉnh balance thủ công |
| `OT_REQUEST_SUBMITTED` | Submit OT |
| `OT_REQUEST_APPROVED` | Approve |
| `OT_REQUEST_REJECTED` | Reject |
| `OT_CAP_WARNING` | Khi approved_hours > 270 (90% of 300) |

#### 17.2.5 Expense Claims

| Action Code | Trigger |
|---|---|
| `EXPENSE_SUBMITTED` | Submit claim |
| `EXPENSE_MANAGER_APPROVED` | Duyệt cấp 1 |
| `EXPENSE_MANAGER_REJECTED` | Từ chối cấp 1 |
| `EXPENSE_HR_APPROVED` | Duyệt cấp 2 |
| `EXPENSE_HR_REJECTED` | Từ chối cấp 2 |
| `EXPENSE_PAID` | Mark paid — ghi payment_reference |

#### 17.2.6 Independence & Performance

| Action Code | Trigger |
|---|---|
| `INDEPENDENCE_DECLARED` | Tạo declaration — ghi has_conflict |
| `INDEPENDENCE_CONFLICT_ACKNOWLEDGED` | Partner acknowledge conflict |
| `PERFORMANCE_REVIEW_SUBMITTED` | Submit review |
| `PERFORMANCE_REVIEW_ACKNOWLEDGED` | Employee acknowledge |

#### 17.2.7 Offboarding

| Action Code | Trigger |
|---|---|
| `OFFBOARDING_INITIATED` | Tạo checklist |
| `OFFBOARDING_ITEM_COMPLETED` | Check item done |
| `OFFBOARDING_COMPLETED` | Mark complete |
| `ACCESS_REVOKED` | Item key = 'access_revoked' done |

---

## 18. Security & Privacy

### 18.1 Encryption Strategy

**Algorithm:** AES-256-GCM (Authenticated Encryption)

**Encrypted fields:**
- `employees.cccd_encrypted` — CCCD/CMND number
- `employees.mst_ca_nhan_encrypted` — Mã số thuế cá nhân
- `employees.so_bhxh_encrypted` — Số sổ BHXH
- `employees.bank_account_encrypted` — Số tài khoản ngân hàng

**Key Management:**
- Encryption key lưu trong environment variable `HRM_ENCRYPTION_KEY` (32 bytes base64)
- Key rotation: Khi rotate key, cần re-encrypt tất cả existing records (scheduled maintenance job)
- Backup key lưu trong secure vault (không trong database)
- Mỗi ciphertext bao gồm: IV (12 bytes) + ciphertext + auth tag (16 bytes), base64-encoded

**Encrypt/Decrypt tại application layer (Go):**
```
plaintext → AES-256-GCM encrypt (key, IV random) → base64(IV + ciphertext + tag)
base64 string → decode → AES-256-GCM decrypt (key, IV, tag verify) → plaintext
```

### 18.2 Access Control cho PII

| Layer | Biện pháp |
|---|---|
| API | Role check trong middleware trước khi decrypt |
| Response | Mask sensitive fields bằng `***` cho unauthorized roles |
| Audit | Ghi audit log mỗi lần decrypt và trả về |
| Database | Encrypted at rest — không ai đọc DB thấy plain text |
| Transport | HTTPS/TLS bắt buộc (enforce trong load balancer) |
| Logs | Application logs không được log giá trị PII đã decrypt |

### 18.3 Data Retention

Theo Bộ luật Lao động 2019 và quy định về lưu trữ hồ sơ:

| Loại dữ liệu | Thời gian lưu | Ghi chú |
|---|---|---|
| Hồ sơ nhân viên (terminated) | 10 năm | Từ ngày nghỉ việc |
| Hợp đồng lao động | 10 năm | Từ ngày hết hiệu lực |
| Lịch sử lương | 10 năm | — |
| BHXH records | Vĩnh viễn | Yêu cầu của cơ quan BHXH |
| Audit logs | 7 năm | Theo quy định kế toán |
| Expense claims | 5 năm | Sau khi thanh toán |
| Training records | 5 năm | — |
| Performance reviews | 3 năm | Sau khi nhân viên nghỉ |
| Independence declarations | 7 năm | Audit compliance |

### 18.4 GDPR / Privacy Considerations

Mặc dù công ty hoạt động chủ yếu theo luật Việt Nam, áp dụng nguyên tắc:
- **Minimization:** Chỉ thu thập dữ liệu thực sự cần thiết
- **Purpose limitation:** PII chỉ dùng cho mục đích HR/payroll/compliance
- **Access control:** Principle of least privilege (phân quyền tối thiểu)
- **Right to access:** Nhân viên có thể xem dữ liệu của chính mình qua self-service
- **Security:** Mã hóa, audit trail, access logging

### 18.5 API Security

- **Authentication:** JWT (RS256) với expiry 15 phút, refresh token 7 ngày
- **Rate limiting:** 100 req/min cho regular endpoints; 10 req/min cho sensitive endpoints
- **Input validation:** Validate tất cả input; reject SQL injection patterns
- **CORS:** Chỉ allow từ whitelisted origins
- **Sensitive endpoints:** Thêm re-authentication yêu cầu (confirm password) khi xem/sửa PII (Phase 2 consideration)

---

## 19. Reporting

### 19.1 Danh Sách 10 Báo Cáo Chuẩn

#### Report 1: CPE Compliance Report

**Path:** `/hrm/reports/cpe-compliance`
**Mô tả:** Tỷ lệ tuân thủ CPE của tất cả nhân viên có chứng chỉ

**Output columns:**
- Employee code, tên, phòng, chi nhánh
- Cert type
- Required hours (năm hiện tại)
- Completed hours
- Remaining hours
- % completion
- Status: ON_TRACK / AT_RISK / NON_COMPLIANT

**Filters:** year, branch_id, dept_id, cert_type, status

---

#### Report 2: Leave Usage Report

**Path:** `/hrm/reports/leave-usage`
**Mô tả:** Tổng hợp sử dụng phép theo loại, theo nhân viên

**Output columns:**
- Employee code, tên, phòng, chi nhánh
- Leave type
- Entitled, used, pending, remaining (days)
- % utilization

**Filters:** year, branch_id, dept_id, leave_type

---

#### Report 3: OT Summary Report

**Path:** `/hrm/reports/ot-summary`
**Mô tả:** Tổng hợp giờ làm thêm theo nhân viên

**Output columns:**
- Employee code, tên, phòng
- Month
- OT weekday hours, OT weekend hours, OT holiday hours
- Total OT hours
- YTD total
- % of annual cap (300h)

**Alerts:** Highlight nhân viên > 250h (>83% cap)

---

#### Report 4: Headcount Report

**Path:** `/hrm/reports/headcount`
**Mô tả:** Số lượng nhân viên theo chi nhánh, phòng ban, grade, loại HĐ

**Output:** Pivot table hoặc summary counts

**Dimensions:** branch, department, grade, employment_type, status

---

#### Report 5: Contract Renewal Alert Report

**Path:** `/hrm/reports/contract-renewal`
**Mô tả:** Danh sách hợp đồng sắp hết hạn

**Output columns:**
- Employee code, tên
- Contract number, type
- End date
- Days remaining
- Manager name

**Filters:** days_ahead (default 60)

---

#### Report 6: Certification Expiry Report

**Path:** `/hrm/reports/cert-expiry`
**Mô tả:** Danh sách chứng chỉ sắp hết hạn hoặc đã hết hạn

**Output columns:**
- Employee, cert type, cert name, cert number
- Expiry date
- Days remaining (negative = đã hết hạn)
- Status

**Filters:** days_ahead, cert_type, branch_id

---

#### Report 7: Independence Status Report

**Path:** `/hrm/reports/independence-status`
**Mô tả:** Trạng thái khai báo độc lập năm hiện tại

**Output columns:**
- Employee code, tên, phòng
- Annual declaration: DECLARED / NOT_DECLARED / HAS_CONFLICT
- Declared date
- Outstanding per-engagement declarations

**Alert:** Highlight nhân viên tham gia engagement nhưng chưa khai báo

---

#### Report 8: Expense Summary Report

**Path:** `/hrm/reports/expense-summary`
**Mô tả:** Tổng hợp chi phí công tác theo kỳ

**Output columns:**
- Employee, phòng, chi nhánh
- Period
- Claims count
- Total amount by category (FLIGHT, HOTEL, TAXI, MEAL, etc.)
- Grand total
- Status breakdown

**Filters:** period_from, period_to, branch_id, dept_id, status, engagement_id

---

#### Report 9: Performance Distribution Report

**Path:** `/hrm/reports/performance-distribution`
**Mô tả:** Phân bố điểm đánh giá hiệu suất theo period

**Output:**
- Histogram: số nhân viên theo dải điểm (1-2, 2-3, 3-4, 4-5)
- Top/bottom performers
- Average by department, branch

**Filters:** review_period, branch_id, dept_id

---

#### Report 10: Salary Report

**Path:** `/hrm/reports/salary`
**Access:** HR_MANAGER, CEO, CHAIRMAN only

**Mô tả:** Tổng hợp lương toàn công ty (snapshot)

**Output columns:**
- Employee code, tên, phòng, chi nhánh, grade
- Base salary
- Date salary effective
- BHXH employee contribution
- BHXH employer contribution
- Total employer cost

**Filters:** branch_id, dept_id, grade

**Security:** Mỗi lần chạy report này, ghi audit log `SALARY_REPORT_GENERATED`.

### 19.2 Export Formats

Tất cả reports hỗ trợ export:
- **Excel (.xlsx):** Recommended cho HR use
- **CSV:** Cho data integration
- **PDF:** Cho management reporting (basic table layout)

---

## 20. Testing Strategy

### 20.1 Unit Tests

**Mục tiêu:** Test các hàm business logic thuần túy, không cần database.

#### 20.1.1 Calculations

| Test | Input | Expected |
|---|---|---|
| `TestCalculateLeaveDays` | start_date, end_date, holidays list | Số ngày làm việc (trừ holidays + weekends) |
| `TestCalculateOTHours` | start_time, end_time | Giờ OT chính xác đến 0.5h |
| `TestOTCapCheck` | current_approved, new_request | PASS/EXCEED khi > 300h/năm |
| `TestLeaveBalanceCheck` | balance, requested_days | SUFFICIENT/INSUFFICIENT |
| `TestEmployeeCodeFormat` | year=26, seq=1 | "NV26-0001" |
| `TestInsuranceContribution` | salary, rate_config | BHXH employee/employer amounts |
| `TestTNCNDeduction` | salary, dependents_count | Tax deduction amount |
| `TestCPEProgress` | records, required | hours_completed, remaining, percentage |

#### 20.1.2 Validators

| Test | Rule |
|---|---|
| `TestValidateEmployeeCode` | Phải match `NV\d{2}-\d{4}` |
| `TestValidateCCCD` | 12 chữ số (CCCD mới) hoặc 9 chữ số (CMND cũ) |
| `TestValidateGrade` | Chỉ chấp nhận 8 values |
| `TestValidateLeaveType` | Chỉ chấp nhận 8 types |
| `TestValidateOTDate` | Không được là ngày tương lai > 7 ngày |
| `TestValidateContractDates` | end_date > start_date |

### 20.2 Integration Tests

**Mục tiêu:** Test workflows end-to-end với database thật (test DB, không mock).

#### 20.2.1 Employee Lifecycle

```
TestEmployeeCreation:
  1. POST /hrm/employees → 201
  2. Assert employee_code = "NV26-0001"
  3. GET /hrm/employees/:id → 200, correct data

TestEmployeeSalaryChange:
  1. POST /hrm/employees/:id/salary-history
  2. Assert record created in salary_history
  3. Assert UPDATE rule blocks update attempt
  4. Assert DELETE rule blocks delete attempt
```

#### 20.2.2 Leave Workflow

```
TestLeaveApprovalFlow:
  1. Employee: POST /my-leave/requests (PENDING)
  2. Assert leave_balance.pending_days += total_days
  3. Manager: POST /hrm/leave/requests/:id/approve
  4. Assert status = APPROVED
  5. Assert leave_balance.used_days += total_days
  6. Assert leave_balance.pending_days -= total_days

TestLeaveRejectionFlow:
  1. Employee: POST /my-leave/requests (PENDING)
  2. Manager: POST /hrm/leave/requests/:id/reject
  3. Assert status = REJECTED
  4. Assert leave_balance.pending_days -= total_days (rollback)
  5. Assert leave_balance.used_days unchanged
```

#### 20.2.3 OT Cap Enforcement

```
TestOTCapEnforcement:
  1. Seed: employee với 290 approved OT hours
  2. POST /my-overtime/requests với ot_hours = 15
  3. Manager: POST approve → 422 OT_ANNUAL_CAP_EXCEEDED
  4. POST /my-overtime/requests với ot_hours = 10
  5. Manager: POST approve → 200 (290 + 10 = 300, exactly at cap)
  6. POST /my-overtime/requests với ot_hours = 1
  7. Manager: POST approve → 422 OT_ANNUAL_CAP_EXCEEDED
```

#### 20.2.4 Provisioning Workflow (HCM)

```
TestProvisioningHCMFlow:
  1. HR: POST /hrm/user-provisioning-requests (HCM employee)
  2. Assert status = PENDING, approval_level = 1
  3. HoB HCM: POST /branch-approve → status unchanged (moves to step 2)
  4. HR_MANAGER: POST /hr-approve → status = APPROVED
  5. SA: POST /execute → atomic:
       - user created
       - role assigned
       - employee.user_id linked
       - status = EXECUTED
  6. Assert welcome email sent
  7. Assert audit log PROVISIONING_EXECUTED recorded
```

#### 20.2.5 Independence Declaration

```
TestIndependenceCleanDeclaration:
  1. POST /hrm/independence { annual, has_conflict: false }
  2. Assert status = CLEAN
  3. Assert no notification sent to Partner

TestIndependenceConflictFlow:
  1. POST /hrm/independence { per_engagement, has_conflict: true, conflict_desc: "..." }
  2. Assert status = PENDING
  3. Assert notification sent to Partner
  4. Partner: POST /acknowledge
  5. Assert status = CONFLICT_RESOLVED
```

### 20.3 E2E Tests (Browser)

**Tool:** Playwright (TypeScript)
**Environment:** Staging environment với seeded data

| User Journey | Steps |
|---|---|
| New employee onboarding | SA login → create employee → HR login → add contract → BHXH info → certifications |
| Self-service leave | Employee login → check balance → submit leave → manager login → approve → employee sees approved |
| Expense claim flow | Employee → create claim → add items (FLIGHT + HOTEL) → submit → manager approve → HR approve → mark paid |
| Independence declaration | Employee login → /my-independence → khai báo năm 2026 → no conflict → see CLEAN status |
| CPE tracking | Employee login → /my-profile/training → add training record (CPE 8h) → check progress (8/40h) |
| Provisioning HCM | HR → request for HCM employee → HoB HCM approve → HR approve → SA execute → welcome email sent |
| Performance review | Manager login → create review for subordinate → submit → employee login → acknowledge |

### 20.4 Test Data Factories

```go
// Go test helpers (không phải production code)
// factory.go trong test packages

func NewEmployee(overrides ...EmployeeOption) Employee
func NewLeaveRequest(employeeID UUID, opts ...LeaveOption) LeaveRequest
func NewOTRequest(employeeID UUID, hours float64) OTRequest
func NewExpenseClaim(employeeID UUID, items ...ExpenseItem) ExpenseClaim
func NewProvisioningRequest(employeeID UUID, role string) ProvisioningRequest
```

---

## 21. Roadmap — 5 Sprints

### Tổng Quan

| Sprint | Tên | Duration | Story Points |
|---|---|---|---|
| Sprint 1 | Organization + Employees | 2 tuần | ~40 SP |
| Sprint 2 | Provisioning + Certifications/Training | 2 tuần | ~35 SP |
| Sprint 3 | Performance Reviews + Independence | 1 tuần | ~20 SP |
| Sprint 4 | Leave + OT + Holidays | 1 tuần | ~25 SP |
| Sprint 5 | Expenses + Reports + Polish | 1 tuần | ~25 SP |

**Tổng:** 4–6 tuần (tùy team size)

---

### Sprint 1: Organization + Employees (Tuần 1–2)

**Deliverables:**
- [ ] Migration 000019: hrm_organization
- [ ] Migration 000020: hrm_employees_extended
- [ ] Migration 000026 (partial): seed branches, departments, matrix
- [ ] API: /hrm/organization/* (10 endpoints)
- [ ] API: /hrm/employees (CRUD, 8 endpoints)
- [ ] API: /my-profile (2 endpoints)
- [ ] API: /hrm/employees/:id/sensitive (PII decrypt + audit)
- [ ] API: /hrm/employees/:id/dependents (4 endpoints)
- [ ] API: /hrm/employees/:id/salary-history (2 endpoints)
- [ ] API: /hrm/employees/:id/contracts (5 endpoints)
- [ ] UI: /admin/hrm/organization
- [ ] UI: /admin/hrm/employees (list + detail)
- [ ] UI: /my-profile
- [ ] Unit tests: employee code generation, salary history immutability
- [ ] Integration tests: employee lifecycle

**Definition of Done:**
- Tất cả endpoints return correct response theo role
- PII fields encrypted at rest, masked in non-sensitive responses
- Audit log ghi khi PII accessed
- Branch scope enforced cho HoB

---

### Sprint 2: Provisioning + Certifications/Training (Tuần 3–4)

**Deliverables:**
- [ ] Migration 000021: hrm_professional
- [ ] Migration 000024: hrm_provisioning
- [ ] API: /hrm/user-provisioning-requests (9 endpoints)
- [ ] API: /hrm/employees/:id/certifications (5 endpoints)
- [ ] API: /hrm/certifications/expiring
- [ ] API: /hrm/training/courses (3 endpoints)
- [ ] API: /hrm/employees/:id/training (3 endpoints)
- [ ] API: /hrm/employees/:id/cpe-summary
- [ ] API: /hrm/training/cpe-summary
- [ ] API: /hrm/offboarding (5 endpoints)
- [ ] UI: /admin/hrm/provisioning
- [ ] UI: /hrm/provisioning
- [ ] UI: /hrm/certifications
- [ ] UI: /hrm/training
- [ ] UI: /my-profile/certifications
- [ ] UI: /my-profile/training
- [ ] Notification: provisioning flow (events 9–13)
- [ ] Notification: cert expiry (event 2)
- [ ] Integration tests: provisioning HCM flow, provisioning HO flow, emergency flow

**Definition of Done:**
- Provisioning flow hoàn chỉnh (HCM 2-step, HO direct)
- Account creation atomic (user + role + employee link)
- CPE tracking chính xác theo năm

---

### Sprint 3: Performance Reviews + Independence (Tuần 5)

**Deliverables:**
- [ ] Migration 000022: hrm_performance
- [ ] API: /hrm/performance/reviews (6 endpoints)
- [ ] API: /hrm/performance/peer-reviews (2 endpoints)
- [ ] API: /hrm/independence (6 endpoints)
- [ ] UI: /hrm/performance
- [ ] UI: /hrm/independence
- [ ] UI: /my-performance
- [ ] UI: /my-independence
- [ ] UI: /my-team/performance
- [ ] UI: /my-team/independence
- [ ] Notification: independence conflict (event 17), annual reminder (event 18)
- [ ] Notification: performance review (events 20–21)
- [ ] Integration tests: independence declaration flows (clean + conflict)

**Definition of Done:**
- ANNUAL unique constraint enforced (1 declaration/năm/nhân viên)
- Conflict flow requires Partner acknowledgment
- Per-engagement declaration required trước khi engagement start (validation hook)

---

### Sprint 4: Leave + OT + Holidays (Tuần 6)

**Deliverables:**
- [ ] Migration 000023: hrm_time_leave
- [ ] Migration 000026 (remaining): seed holidays 2026–2030
- [ ] API: /hrm/holidays (4 endpoints)
- [ ] API: /hrm/leave/balances, requests (10 endpoints)
- [ ] API: /my-leave (4 endpoints)
- [ ] API: /hrm/overtime (5 endpoints)
- [ ] API: /my-overtime (4 endpoints)
- [ ] VIEW: employee_ot_summary_year
- [ ] UI: /hrm/leave/requests, /hrm/leave/calendar
- [ ] UI: /hrm/overtime/requests
- [ ] UI: /my-leave (+ /my-leave/new)
- [ ] UI: /my-overtime (+ /my-overtime/new)
- [ ] UI: /admin/hrm/holidays
- [ ] Notification: leave events (events 3–5), OT events (6–8)
- [ ] Integration tests: leave approval flow, OT cap enforcement
- [ ] Scheduled job: contract expiry alert (event 1)
- [ ] Scheduled job: BHYT expiry alert (event 25)
- [ ] Scheduled job: CPE reminder (event 19)

**Definition of Done:**
- OT cap 300h/năm enforced tại API level
- Leave balance updated atomically khi approve/reject
- Holidays seeded và calendar hiển thị đúng

---

### Sprint 5: Expenses + Reports + Polish (Tuần 7)

**Deliverables:**
- [ ] Migration 000025: hrm_expenses
- [ ] API: /my-expenses (8 endpoints)
- [ ] API: /hrm/expenses (5 endpoints)
- [ ] API: /hrm/reports/* (10 report endpoints)
- [ ] API: /hrm/insurance-config (3 endpoints)
- [ ] UI: /my-expenses (+ /my-expenses/new)
- [ ] UI: /hrm/expenses
- [ ] UI: /admin/hrm/reports
- [ ] UI: /admin/hrm/insurance-config
- [ ] Report: CPE Compliance, Leave Usage, OT Summary
- [ ] Report: Headcount, Contract Renewal, Cert Expiry
- [ ] Report: Independence Status, Expense Summary, Performance Distribution, Salary Report
- [ ] Export: Excel + CSV cho tất cả reports
- [ ] Notification: expense events (14–16)
- [ ] Integration tests: expense claim flow
- [ ] E2E tests: tất cả major user journeys
- [ ] Security review: PII access, encryption keys, audit trails
- [ ] Performance: verify p95 < 500ms cho common queries

**Definition of Done:**
- Tất cả 10 reports hoạt động với real data
- Salary report có audit log
- Export Excel/CSV hoạt động
- p95 API latency < 500ms
- All integration + E2E tests pass

---

## 22. Bootstrap Initial Setup

### 22.1 Tổng Quan

Quy trình bootstrap là sequence các bước thủ công + tự động để khởi tạo hệ thống từ đầu với dữ liệu thực tế.

**Điều kiện tiên quyết:**
- Database migration 000001–000026 đã chạy thành công
- Seed data cơ bản đã có (branches, departments, matrix, insurance rates)

### 22.2 Bước 1: Tạo SUPER_ADMIN User (Thủ công)

SA là tài khoản đầu tiên — không thể tạo qua UI vì chưa có ai có quyền.

```bash
# SA chạy lệnh trực tiếp từ server
make create-super-admin EMAIL=admin@mdh.vn PASSWORD=<secure_password>
```

Hoặc qua migration seed:
```sql
INSERT INTO users (email, password_hash, role, is_active, email_verified)
VALUES ('admin@mdh.vn', <bcrypt_hash>, 'SUPER_ADMIN', true, true);
```

**Lưu ý:** Password phải 12+ ký tự, TOTP bắt buộc. SA setup TOTP ngay sau khi đăng nhập lần đầu.

### 22.3 Bước 2: Xác Nhận Chi Nhánh (Offline)

**Ngoài hệ thống:** Chủ tịch HĐQT ký Quyết định thành lập chi nhánh HCM (giấy tờ pháp lý thực tế).

SA cập nhật thông tin chi nhánh vào hệ thống:
```sql
UPDATE branches SET
  address = '123 Lê Lợi, Q1, TP.HCM',
  phone = '028-XXXX-XXXX',
  established_date = '2020-01-01'
WHERE code = 'HCM';

UPDATE branches SET
  address = '456 Nguyễn Chí Thanh, Đống Đa, Hà Nội',
  phone = '024-XXXX-XXXX',
  established_date = '2018-03-01'
WHERE code = 'HO';
```

### 22.4 Bước 3: Tạo Hồ Sơ + Tài Khoản 6 User Đầu Tiên

SA thực hiện qua UI `/admin/hrm/employees` + `/admin/hrm/provisioning`:

| STT | Vai trò | Chi nhánh | Role | Flow |
|---|---|---|---|---|
| 1 | Chủ tịch HĐQT | HO | CHAIRMAN | SA tạo trực tiếp (emergency) |
| 2 | Tổng Giám đốc (CEO) | HO | CEO | SA tạo trực tiếp |
| 3 | Trưởng phòng Nhân sự (HR Manager) | HO | HR_MANAGER | SA tạo trực tiếp |
| 4 | Trưởng Chi nhánh HCM (HoB) | HCM | HEAD_OF_BRANCH | SA tạo trực tiếp |
| 5 | Partner HO | HO | PARTNER | CEO/HR tạo qua flow thông thường |
| 6 | Partner HCM | HCM | PARTNER | HoB HCM request → HR approve → SA execute |

**Sequence thực tế:**

```
Ngày 1 (SA):
  1. Tạo hồ sơ CHAIRMAN (employee record)
  2. Tạo account CHAIRMAN (is_emergency=true, reason="Initial bootstrap")
  3. Tạo hồ sơ CEO
  4. Tạo account CEO (is_emergency=true)
  5. Tạo hồ sơ HR Manager
  6. Tạo account HR Manager (is_emergency=true)
  7. Tạo hồ sơ HoB HCM
  8. Tạo account HoB HCM (is_emergency=true)

Ngày 2 (CEO / HR Manager đã có account):
  9.  HR tạo hồ sơ Partner HO
  10. CEO approve → SA execute
  11. HR request account cho Partner HCM
  12. HoB HCM approve (step 1) → HR approve (step 2) → SA execute

Kết quả: 6 users active, hệ thống sẵn sàng vận hành thông thường
```

### 22.5 Bước 4: Thiết Lập Dữ Liệu Ban Đầu

HR Manager thực hiện sau khi có account:

```
1. Xác nhận ngày lễ 2026 đúng (Admin → Holidays)
2. Xác nhận tỷ lệ BHXH 2024 đúng (Admin → Insurance Config)
3. Tạo các khóa đào tạo nội bộ (Training → Courses)
4. Setup leave balances cho 6 users đầu tiên (HR → Leave Balances)
5. Import/nhập hồ sơ các nhân viên còn lại theo batch
```

### 22.6 Bước 5: Khai Báo Độc Lập Năm Đầu Tiên

Sau khi tất cả nhân viên có account:

```
Tháng 1 (hoặc tháng bắt đầu hệ thống):
  - HR gửi announcement: Tất cả nhân viên cần khai báo độc lập năm YYYY
  - Nhân viên login → /my-independence → Khai báo ANNUAL
  - HR theo dõi trên /hrm/independence/annual-status
  - Nhắc lại sau 2 tuần nếu chưa khai báo
```

### 22.7 Checklist Bootstrap Hoàn Tất

```
[ ] Migration 000019–000026 đã run thành công
[ ] SUPER_ADMIN user đã tạo và setup TOTP
[ ] Branches (HO, HCM) đã có đầy đủ thông tin địa chỉ
[ ] Departments (5) đã có code đúng
[ ] branch_departments matrix đúng (HO: 5 phòng, HCM: 2 phòng)
[ ] Insurance rates 2024 seeded
[ ] Holidays 2026–2030 seeded
[ ] CPE requirements seeded
[ ] CHAIRMAN user đã tạo và đăng nhập được
[ ] CEO user đã tạo và đăng nhập được
[ ] HR_MANAGER user đã tạo và đăng nhập được
[ ] HoB HCM user đã tạo và đăng nhập được
[ ] Partner HO, Partner HCM đã có account
[ ] Leave balances năm đầu tiên đã khởi tạo
[ ] Tất cả nhân viên đã khai báo độc lập ANNUAL năm đầu tiên
[ ] Tất cả nhân viên đã có ít nhất 1 hợp đồng lao động trong hệ thống
[ ] Onboarding checklist đã tạo cho nhân viên còn trong thời gian thử việc
[ ] HR review báo cáo Headcount → confirm số lượng đúng
```

### 22.8 Vận Hành Thường Xuyên Sau Bootstrap

Sau khi bootstrap hoàn tất, các quy trình vận hành chạy theo cơ chế thông thường (phi tập trung):

| Quy trình | Người thực hiện | Tần suất |
|---|---|---|
| Tuyển nhân viên mới (HCM) | HoB HCM request → HR approve → SA execute | Per hire |
| Tuyển nhân viên mới (HO) | HR create → SA execute | Per hire |
| Tăng lương | HR + CEO → ghi salary_history | Khi có quyết định |
| Gia hạn hợp đồng | HR tạo contract mới → set-current | 30 ngày trước hết hạn |
| Khai báo độc lập năm | Tất cả nhân viên | Tháng 1 hàng năm |
| Đăng ký CPE | Nhân viên tự đăng ký + HR confirm | Liên tục |
| Đánh giá hiệu suất | Manager → HR consolidate | 2 lần/năm (H1, FY) |
| Cập nhật ngày lễ năm mới | HR | Tháng 11 năm trước |
| Cập nhật tỷ lệ BHXH | HR (khi Nhà nước điều chỉnh) | Khi có thay đổi |
| Report hàng tháng | HR gửi cho CEO/CHAIRMAN | Đầu tháng |

---

## Appendix A: Glossary

| Thuật ngữ | Định nghĩa |
|---|---|
| BHXH | Bảo hiểm Xã hội |
| BHYT | Bảo hiểm Y tế |
| BHTN | Bảo hiểm Thất nghiệp |
| TNCN | Thu nhập Cá nhân (thuế) |
| CCCD | Căn cước Công dân (ID card mới) |
| CMND | Chứng minh Nhân dân (ID card cũ) |
| MST | Mã Số Thuế |
| CPA | Certified Public Accountant |
| CPE | Continuing Professional Education |
| VACPA | Hội Kiểm toán viên Hành nghề Việt Nam |
| VSA | Chuẩn mực Kiểm toán Việt Nam |
| ISA | International Standards on Auditing |
| HoB | Head of Branch (Trưởng Chi nhánh) |
| SA | Super Admin (Quản trị viên Hệ thống) |
| OT | Overtime (Làm thêm giờ) |
| PII | Personally Identifiable Information |
| NV | Nhân viên (prefix mã nhân viên) |
| PC | Phiếu Chi (prefix mã claim chi phí) |
| Kiêm nhiệm | Concurrent role assignment |
| Hoa hồng | Sales commission |

---

## Appendix B: Related Documents

- `docs/SPEC.md` — Full system specification
- `docs/ROADMAP.md` — Development roadmap
- `docs/modules/07-hrm.md` — HRM module overview
- `docs/DECISIONS.md` — Architectural decisions
- `apps/api/migrations/` — SQL migration files

---

*Document version 1.4 — Last updated 2026-04-20*
*For questions: contact minhtd529@gmail.com*
