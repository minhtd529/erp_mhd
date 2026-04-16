> ⚠️ DEPRECATED — Xem docs/SPEC.md cho phiên bản mới nhất

# 📋 ERP SYSTEM — TECHNICAL SPECIFICATION DOCUMENT
## Hệ thống ERP cho Công ty Kiểm toán – Tư vấn Tài chính – Thuế

**Version:** 1.1 (Updated: 2026-04-12 — Bổ sung Push Notification Self-Hosted + 2FA/MFA)  
**Date:** 2026-04-12  
**Author:** Senior Architect / PM  
**Tech Stack:** Golang (Backend) · React / Next.js (Frontend) · PostgreSQL (Database)

---

# MỤC LỤC

1. [TỔNG QUAN HỆ THỐNG](#1-tổng-quan-hệ-thống)
2. [KIẾN TRÚC TỔNG THỂ](#2-kiến-trúc-tổng-thể)
3. [MODULE 0: GLOBAL SHARED](#3-module-0-global-shared)
4. [MODULE 1: CRM – QUẢN LÝ KHÁCH HÀNG](#4-module-1-crm)
5. [MODULE 2: ENGAGEMENT – QUẢN LÝ HỢP ĐỒNG & DỰ ÁN](#5-module-2-engagement)
6. [MODULE 3: TIMESHEET & RESOURCE – QUẢN LÝ THỜI GIAN & NGUỒN LỰC](#6-module-3-timesheet)
7. [MODULE 4: BILLING & INVOICING – QUẢN LÝ PHÍ & THANH TOÁN](#7-module-4-billing)
8. [MODULE 5: WORKING PAPERS – QUẢN LÝ HỒ SƠ KIỂM TOÁN](#8-module-5-working-papers)
9. [MODULE 6: TAX & ADVISORY – QUẢN LÝ THUẾ & TƯ VẤN](#9-module-6-tax-advisory)
10. [MODULE 7: HRM – QUẢN LÝ NHÂN SỰ & NĂNG LỰC](#10-module-7-hrm)
11. [MODULE 8: REPORTING & ANALYTICS](#11-module-8-reporting)
12. [NON-FUNCTIONAL REQUIREMENTS](#12-non-functional-requirements)
13. [DATABASE SCHEMA TỔNG QUAN](#13-database-schema)
14. [API DESIGN CONVENTIONS](#14-api-design-conventions)
15. [DEPLOYMENT & INFRASTRUCTURE](#15-deployment)

---

# 1. TỔNG QUAN HỆ THỐNG

## 1.1 Bối cảnh doanh nghiệp

| Thông số | Giá trị |
|---|---|
| Quy mô hiện tại | 100 nhân sự |
| Quy mô mục tiêu (3-5 năm) | 200 nhân sự |
| Concurrent users | 300 |
| Cơ cấu tổ chức | Hội đồng thành viên → Ban TGĐ → 4 Phòng KT-TVTC + Phòng KT XDCB + Phòng HC-KT + Chi nhánh MN |
| Chi nhánh | Hoạt động độc lập, công ty kiểm soát chất lượng dịch vụ |
| Dịch vụ chính | Kiểm toán BCTC, Kiểm toán nội bộ, Kiểm toán XDCB, Tư vấn thuế, Tư vấn tài chính, Soát xét |
| Hệ thống hiện tại | Misa (kế toán), Excel (thủ công) |
| Timeline | 9–12 tháng |
| Deploy | Cloud |
| Đồng tiền | VND (đơn tệ) |
| Tuân thủ | Luật Kiểm toán độc lập, VSA |
| Lưu trữ hồ sơ | 10 năm, yêu cầu bảo mật |

## 1.2 Pain Points cần giải quyết

1. **Thủ công quá nhiều** — quy trình chấm công, theo dõi công nợ, quản lý hồ sơ bằng Excel/giấy
2. **Quản lý rời rạc** — không có nguồn dữ liệu thống nhất (single source of truth)
3. **Thiếu kiểm soát chất lượng dịch vụ** — không có workflow review/sign-off chuẩn hóa
4. **Thiếu thông tin nhân sự thống nhất** — hồ sơ, chứng chỉ, năng lực phân tán
5. **Template lộn xộn** — chưa chuẩn hóa mẫu biểu, hồ sơ kiểm toán

## 1.3 Kỳ vọng ROI

- Tiết kiệm thời gian thao tác thủ công ≥ 40%
- Giảm sai sót do nhập liệu thủ công ≥ 60%
- Tăng khả năng kiểm soát chất lượng dịch vụ toàn diện

---

# 2. KIẾN TRÚC TỔNG THỂ

## 2.1 High-Level Architecture

```
┌──────────────────────────────────────────────────────────┐
│                    FRONTEND (Next.js)                     │
│  ┌─────────┐ ┌────────┐ ┌──────────┐ ┌───────────────┐  │
│  │  Pages   │ │ Hooks  │ │  Store   │ │  Components   │  │
│  │ (SSR/CSR)│ │(React Q)│ │ (Zustand)│ │  (Shadcn/UI)  │  │
│  └────┬─────┘ └───┬────┘ └────┬─────┘ └──────┬────────┘  │
│       └───────────┴──────────┴───────────────┘            │
│                         │ HTTP/REST + WebSocket            │
└─────────────────────────┼────────────────────────────────┘
                          │
┌─────────────────────────┼────────────────────────────────┐
│                    API GATEWAY (Golang)                    │
│  ┌──────────┐ ┌─────────┐ ┌──────────┐ ┌─────────────┐  │
│  │   Auth   │ │  Rate   │ │  CORS    │ │  Request    │  │
│  │Middleware│ │ Limiter │ │ Handler  │ │  Logger     │  │
│  └──────────┘ └─────────┘ └──────────┘ └─────────────┘  │
└─────────────────────────┼────────────────────────────────┘
                          │
┌─────────────────────────┼────────────────────────────────┐
│              BACKEND SERVICES (Golang)                     │
│                                                            │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐    │
│  │  Global  │ │   CRM    │ │Engagement│ │Timesheet │    │
│  │ Service  │ │ Service  │ │ Service  │ │ Service  │    │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘    │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐    │
│  │ Billing  │ │ Working  │ │   Tax    │ │   HRM    │    │
│  │ Service  │ │ Papers   │ │ Advisory │ │ Service  │    │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘    │
│  ┌──────────┐                                            │
│  │Reporting │                                            │
│  │ Service  │                                            │
│  └──────────┘                                            │
└─────────────────────────┼────────────────────────────────┘
                          │
┌─────────────────────────┼────────────────────────────────┐
│                   DATA LAYER                               │
│  ┌──────────┐ ┌─────────┐ ┌──────────┐ ┌─────────────┐  │
│  │PostgreSQL│ │  Redis  │ │  MinIO   │ │Elasticsearch│  │
│  │(Primary) │ │ (Cache) │ │(File/Obj)│ │(Full-text)  │  │
│  └──────────┘ └─────────┘ └──────────┘ └─────────────┘  │
└──────────────────────────────────────────────────────────┘
```

## 2.2 Tech Stack chi tiết

| Layer | Technology | Lý do chọn |
|---|---|---|
| **Frontend Framework** | Next.js 14 (App Router) | SSR/SSG, SEO, file-based routing, middleware |
| **UI Library** | React 18 + Shadcn/UI + Tailwind CSS | Component library mạnh, theming, accessibility |
| **State Management** | Zustand + React Query (TanStack) | Zustand cho client state, React Query cho server state & caching |
| **Backend Language** | Go 1.22+ | Performance cao, concurrency native, type safety, compile-time checks |
| **HTTP Framework** | Gin hoặc Chi | Lightweight, high-performance router |
| **ORM / DB** | sqlc + pgx | Type-safe SQL, zero-reflection, compile-time query validation |
| **Database** | PostgreSQL 16 | ACID, JSON support, full-text search, partitioning |
| **Cache** | Redis 7 | Session, cache, pub/sub cho real-time notifications |
| **Object Storage** | MinIO (S3-compatible) | Lưu file hồ sơ kiểm toán, attachments |
| **Search Engine** | Elasticsearch 8 (optional phase 2) | Full-text search hồ sơ, client, engagement |
| **Message Queue** | NATS (hoặc RabbitMQ) | Async processing: email, notification, report generation |
| **Auth** | JWT + RBAC + ABAC | Access token + Refresh token, role-based + attribute-based |
| **API Protocol** | RESTful JSON + WebSocket | REST cho CRUD, WebSocket cho real-time dashboard & notifications |
| **Push Notification** | Self-hosted Push Relay (WebSocket) + W3C Web Push (VAPID) | Không phụ thuộc Firebase/OneSignal/Pusher |
| **Mobile App** | React Native (Expo) | Push notification, 2FA approve, task view |
| **2FA/MFA** | TOTP (pquerna/otp) + Push-based (self-hosted) | Self-hosted, không dùng Twilio/Authy API |
| **QR Code** | skip2/go-qrcode | Self-render QR cho TOTP setup |
| **Containerization** | Docker + Docker Compose | Dev/staging/prod consistency |
| **Orchestration** | Kubernetes (production) | Auto-scaling, self-healing, rolling updates |
| **CI/CD** | GitHub Actions | Automated testing, building, deployment |
| **Monitoring** | Prometheus + Grafana + Loki | Metrics, dashboards, log aggregation |

## 2.3 Monorepo Structure

```
erp-audit/
├── apps/
│   ├── web/                    # Next.js frontend
│   │   ├── src/
│   │   │   ├── app/            # App Router pages
│   │   │   ├── components/     # Shared UI components
│   │   │   ├── hooks/          # Custom hooks
│   │   │   ├── lib/            # Utilities, API client
│   │   │   ├── store/          # Zustand stores
│   │   │   └── types/          # TypeScript types
│   │   ├── public/
│   │   │   └── sw.js           # Service Worker cho Web Push
│   │   └── package.json
│   ├── mobile/                 # React Native mobile app (self-hosted push)
│   │   ├── src/
│   │   │   ├── screens/        # Login, Notifications, TaskDetail, 2FA, Settings
│   │   │   ├── services/
│   │   │   │   ├── push-connection.ts    # Persistent WebSocket tới push relay
│   │   │   │   ├── push-handler.ts       # Xử lý push message → local notification
│   │   │   │   ├── background-sync.ts    # Background fetch (iOS/Android)
│   │   │   │   └── two-factor.ts         # 2FA push approve/reject
│   │   │   ├── hooks/
│   │   │   │   ├── use-push.ts           # Push notification hook
│   │   │   │   └── use-auth.ts           # Auth + 2FA hook
│   │   │   ├── navigation/
│   │   │   │   └── deep-linking.ts       # Deep link notification → screen
│   │   │   └── utils/
│   │   │       └── device-info.ts        # Device fingerprint, token
│   │   ├── android/
│   │   │   └── .../PushConnectionService.java  # Foreground service giữ WS connection
│   │   ├── ios/
│   │   │   └── BackgroundTaskManager.swift     # Background fetch scheduling
│   │   └── package.json
│   └── api/                    # Golang backend
│       ├── cmd/
│       │   └── server/         # Entry point
│       ├── internal/
│       │   ├── global/         # Module 0: shared services
│       │   ├── crm/            # Module 1
│       │   ├── engagement/     # Module 2
│       │   ├── timesheet/      # Module 3
│       │   ├── billing/        # Module 4
│       │   ├── workingpaper/   # Module 5
│       │   ├── tax/            # Module 6
│       │   ├── hrm/            # Module 7
│       │   └── reporting/      # Module 8
│       ├── pkg/                # Shared packages
│       │   ├── auth/
│       │   ├── middleware/
│       │   ├── database/
│       │   ├── storage/
│       │   ├── notification/
│       │   ├── audit/          # Audit trail
│       │   ├── export/         # Excel/PDF/Word export
│       │   └── validator/
│       ├── migrations/         # SQL migrations
│       └── go.mod
├── packages/
│   └── shared-types/           # Shared TypeScript types
├── docker-compose.yml
└── Makefile
```

## 2.4 Cấu trúc mỗi Module Backend (Golang)

Mỗi module tuân theo **Clean Architecture / Hexagonal Architecture**:

```
internal/<module>/
├── domain/
│   ├── entity.go          # Domain entities (structs)
│   ├── repository.go      # Repository interfaces
│   ├── service.go         # Domain service interfaces
│   └── errors.go          # Domain-specific errors
├── usecase/
│   ├── create.go          # Use case implementations
│   ├── update.go
│   ├── delete.go
│   ├── list.go
│   └── dto.go             # Request/Response DTOs
├── repository/
│   ├── postgres.go        # PostgreSQL implementation
│   └── queries/           # sqlc query files (.sql)
├── handler/
│   ├── http.go            # HTTP handlers (Gin/Chi)
│   └── routes.go          # Route registration
└── wire.go                # Dependency injection setup
```

---

# 3. MODULE 0: GLOBAL SHARED

Module này chứa tất cả các service, entity, function dùng chung cho toàn bộ hệ thống.

## 3.1 Authentication & Authorization Service

### 3.1.1 Entities

```go
// pkg/auth/entity.go

type User struct {
    ID             uuid.UUID       `json:"id" db:"id"`
    Email          string          `json:"email" db:"email"`
    HashedPassword string          `json:"-" db:"hashed_password"`
    FullName       string          `json:"full_name" db:"full_name"`
    EmployeeID     *uuid.UUID      `json:"employee_id" db:"employee_id"`     // FK → HRM.Employee
    BranchID       *uuid.UUID      `json:"branch_id" db:"branch_id"`         // FK → Branch
    DepartmentID   *uuid.UUID      `json:"department_id" db:"department_id"` // FK → Department
    Status         UserStatus      `json:"status" db:"status"`               // active, inactive, locked
    LastLoginAt    *time.Time      `json:"last_login_at" db:"last_login_at"`

    // Two-Factor Authentication (2FA/MFA)
    TwoFactorEnabled    bool       `json:"two_factor_enabled" db:"two_factor_enabled"`
    TwoFactorMethod     TwoFAMethod `json:"two_factor_method" db:"two_factor_method"` // totp, push
    TwoFactorSecret     string     `json:"-" db:"two_factor_secret"`                  // TOTP secret (encrypted)
    TwoFactorVerifiedAt *time.Time `json:"two_factor_verified_at" db:"two_factor_verified_at"`
    BackupCodesHash     string     `json:"-" db:"backup_codes_hash"`                  // Hashed backup codes
    TrustedDevices      []string   `json:"-" db:"trusted_devices"`                    // Device fingerprints (skip 2FA for 30 days)

    // Push Notification
    PushSubscriptions   []string   `json:"-" db:"push_subscriptions"`                 // JSON array of push endpoints

    CreatedAt      time.Time       `json:"created_at" db:"created_at"`
    UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

type TwoFAMethod string
const (
    TwoFAMethodTOTP TwoFAMethod = "totp"  // Time-based OTP (Google Authenticator, Authy...)
    TwoFAMethodPush TwoFAMethod = "push"  // Push notification to mobile app
)

type UserStatus string
const (
    UserStatusActive   UserStatus = "active"
    UserStatusInactive UserStatus = "inactive"
    UserStatusLocked   UserStatus = "locked"
)

type Role struct {
    ID          uuid.UUID  `json:"id" db:"id"`
    Code        string     `json:"code" db:"code"`        // e.g., "admin", "partner", "manager", "senior", "junior", "intern"
    Name        string     `json:"name" db:"name"`
    Description string     `json:"description" db:"description"`
    Level       int        `json:"level" db:"level"`      // Hierarchy level: 1=Chairman, 2=Partner...
    IsSystem    bool       `json:"is_system" db:"is_system"` // Built-in roles cannot be deleted
    CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

type Permission struct {
    ID       uuid.UUID `json:"id" db:"id"`
    Module   string    `json:"module" db:"module"`     // "crm", "engagement", "billing"...
    Resource string    `json:"resource" db:"resource"` // "client", "contract", "invoice"...
    Action   string    `json:"action" db:"action"`     // "create", "read", "update", "delete", "approve", "export"
}

type RolePermission struct {
    RoleID       uuid.UUID `db:"role_id"`
    PermissionID uuid.UUID `db:"permission_id"`
    Scope        string    `db:"scope"` // "all", "branch", "department", "own"
}

type UserRole struct {
    UserID uuid.UUID `db:"user_id"`
    RoleID uuid.UUID `db:"role_id"`
}

// Session / Token
type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`
}

type TokenClaims struct {
    UserID       uuid.UUID   `json:"user_id"`
    Email        string      `json:"email"`
    Roles        []string    `json:"roles"`
    BranchID     *uuid.UUID  `json:"branch_id"`
    DepartmentID *uuid.UUID  `json:"department_id"`
    Permissions  []string    `json:"permissions"` // flattened: "crm:client:read"
}

// ============================================================
// Two-Factor Authentication (2FA) Entities
// ============================================================

// TOTP Setup — trả về khi user bật 2FA
type TwoFactorSetup struct {
    Secret     string `json:"secret"`      // Base32-encoded TOTP secret
    QRCodeURL  string `json:"qr_code_url"` // otpauth:// URI → render QR trên frontend
    QRCodeImage []byte `json:"-"`           // QR code PNG bytes (self-rendered, không dùng API bên thứ 3)
    BackupCodes []string `json:"backup_codes"` // 10 mã backup, mỗi mã 8 ký tự
}

// Login Challenge — khi user có 2FA, login trả về challenge thay vì token
type LoginChallenge struct {
    ChallengeID    string      `json:"challenge_id"`        // Temporary ID, TTL 5 phút
    ChallengeType  TwoFAMethod `json:"challenge_type"`      // "totp" hoặc "push"
    ExpiresAt      time.Time   `json:"expires_at"`
    Message        string      `json:"message"`             // "Nhập mã OTP từ ứng dụng xác thực"
}

// Login Response — unified response cho cả có và không 2FA
type LoginResponse struct {
    RequiresTwoFactor bool             `json:"requires_two_factor"`
    Challenge         *LoginChallenge  `json:"challenge,omitempty"`    // != nil khi cần 2FA
    Token             *TokenPair       `json:"token,omitempty"`        // != nil khi login thành công
}

// Trusted Device — cho phép skip 2FA trên thiết bị đã xác minh
type TrustedDevice struct {
    ID            uuid.UUID `json:"id" db:"id"`
    UserID        uuid.UUID `json:"user_id" db:"user_id"`
    DeviceFingerprint string `json:"-" db:"device_fingerprint"` // SHA-256 hash
    DeviceName    string    `json:"device_name" db:"device_name"` // "Chrome on Windows", "ERP Mobile App"
    IPAddress     string    `json:"ip_address" db:"ip_address"`
    TrustedUntil  time.Time `json:"trusted_until" db:"trusted_until"` // 30 ngày
    CreatedAt     time.Time `json:"created_at" db:"created_at"`
}
```

### 3.1.2 Predefined Roles (seed data)

| Code | Name | Level | Mô tả |
|---|---|---|---|
| `chairman` | Chủ tịch HĐTV | 1 | Full quyền hệ thống, phê duyệt cao nhất |
| `partner` | Partner | 2 | Ký hợp đồng, phê duyệt báo cáo kiểm toán |
| `director` | Tổng Giám đốc / Phó TGĐ | 2 | Phân công nhân sự, phê duyệt hợp đồng |
| `branch_director` | Giám đốc Chi nhánh | 3 | Quản lý chi nhánh, phân công nhân sự chi nhánh |
| `manager` | Trưởng phòng / Manager | 4 | Review hồ sơ, quản lý team |
| `senior` | Senior Auditor | 5 | Thực hiện & kiểm tra hồ sơ |
| `junior` | Junior Auditor | 6 | Thực hiện công việc |
| `intern` | Thực tập sinh | 7 | Hỗ trợ, hạn chế quyền |
| `accountant` | Kế toán | 5 | Billing, invoice, công nợ |
| `admin` | System Admin | 1 | Quản trị hệ thống |
| `hr` | Nhân sự | 5 | Quản lý hồ sơ nhân sự |

### 3.1.3 Service Interface

```go
// pkg/auth/service.go

type AuthService interface {
    // Authentication
    Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)   // Returns challenge if 2FA enabled
    Logout(ctx context.Context, userID uuid.UUID) error
    RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
    ChangePassword(ctx context.Context, userID uuid.UUID, req ChangePasswordRequest) error
    ResetPassword(ctx context.Context, email string) error

    // Two-Factor Authentication (2FA)
    Enable2FA(ctx context.Context, userID uuid.UUID, method TwoFAMethod) (*TwoFactorSetup, error)
    Verify2FASetup(ctx context.Context, userID uuid.UUID, code string) error        // Xác minh lần đầu khi bật 2FA
    Disable2FA(ctx context.Context, userID uuid.UUID, password string) error         // Cần nhập lại password để tắt
    Verify2FALogin(ctx context.Context, challengeID string, code string, trustDevice bool, deviceInfo DeviceInfo) (*TokenPair, error) // Xác minh OTP khi login
    RegenerateBackupCodes(ctx context.Context, userID uuid.UUID, password string) ([]string, error)
    VerifyBackupCode(ctx context.Context, challengeID string, backupCode string) (*TokenPair, error) // Dùng backup code khi mất device

    // Trusted Device Management
    ListTrustedDevices(ctx context.Context, userID uuid.UUID) ([]TrustedDevice, error)
    RevokeTrustedDevice(ctx context.Context, userID uuid.UUID, deviceID uuid.UUID) error
    RevokeAllTrustedDevices(ctx context.Context, userID uuid.UUID) error

    // Push-based 2FA (qua mobile app tự build)
    Send2FAPushChallenge(ctx context.Context, userID uuid.UUID, challengeID string) error          // Gửi push tới app
    Check2FAPushResponse(ctx context.Context, challengeID string) (*TokenPair, bool, error)        // Polling: user đã approve trên app chưa?
    Respond2FAPush(ctx context.Context, challengeID string, approved bool, deviceToken string) error // Mobile app gọi khi user tap approve/reject

    // Authorization
    HasPermission(ctx context.Context, userID uuid.UUID, module, resource, action string) (bool, error)
    HasRole(ctx context.Context, userID uuid.UUID, roleCode string) (bool, error)
    GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]Permission, error)

    // User Management
    CreateUser(ctx context.Context, req CreateUserRequest) (*User, error)
    UpdateUser(ctx context.Context, id uuid.UUID, req UpdateUserRequest) (*User, error)
    DeactivateUser(ctx context.Context, id uuid.UUID) error
    ListUsers(ctx context.Context, filter UserFilter) (*PaginatedResult[User], error)
    AssignRole(ctx context.Context, userID, roleID uuid.UUID) error
    RevokeRole(ctx context.Context, userID, roleID uuid.UUID) error

    // 2FA Admin — Admin/Chairman bắt buộc 2FA cho role cụ thể
    Enforce2FAForRole(ctx context.Context, roleCode string) error
    Get2FAStatus(ctx context.Context, userID uuid.UUID) (*TwoFAStatusResponse, error)
    GetSystem2FAPolicy(ctx context.Context) (*TwoFAPolicyResponse, error)
}

type LoginRequest struct {
    Email           string `json:"email" validate:"required,email"`
    Password        string `json:"password" validate:"required,min=8"`
    DeviceFingerprint string `json:"device_fingerprint"` // Để check trusted device, skip 2FA
}

type DeviceInfo struct {
    Fingerprint string `json:"fingerprint"`
    Name        string `json:"name"`       // "Chrome on Windows 11"
    IPAddress   string `json:"ip_address"` // Auto-filled from request
}

type TwoFAStatusResponse struct {
    Enabled        bool        `json:"enabled"`
    Method         TwoFAMethod `json:"method"`
    TrustedDevices int         `json:"trusted_devices_count"`
    BackupCodesRemaining int   `json:"backup_codes_remaining"`
    EnforcedByPolicy bool      `json:"enforced_by_policy"` // Admin bắt buộc?
}

type TwoFAPolicyResponse struct {
    EnforcedRoles  []string `json:"enforced_roles"`    // Roles bắt buộc 2FA
    AllowedMethods []TwoFAMethod `json:"allowed_methods"`
    TrustDeviceDays int     `json:"trust_device_days"` // Default 30
}

type ChangePasswordRequest struct {
    OldPassword string `json:"old_password" validate:"required"`
    NewPassword string `json:"new_password" validate:"required,min=8,containsany=!@#$%"`
}
```

### 3.1.4 Middleware

```go
// pkg/middleware/auth.go

func AuthMiddleware(authSvc auth.AuthService) gin.HandlerFunc
func RequirePermission(module, resource, action string) gin.HandlerFunc
func RequireRole(roles ...string) gin.HandlerFunc
func RequireAnyRole(roles ...string) gin.HandlerFunc

// Data scoping middleware — tự động filter data theo branch/department
func BranchScopeMiddleware() gin.HandlerFunc   // User chỉ thấy data của branch mình
func EngagementScopeMiddleware() gin.HandlerFunc // User chỉ thấy engagement được assign
```

### 3.1.5 Frontend: Auth Store & Hook

```typescript
// apps/web/src/store/auth-store.ts
interface AuthState {
  user: User | null;
  token: string | null;
  permissions: string[];
  isAuthenticated: boolean;

  // 2FA state
  twoFactorChallenge: LoginChallenge | null;  // != null khi đang chờ 2FA
  twoFactorEnabled: boolean;

  // Actions
  login: (email: string, password: string, deviceFingerprint?: string) => Promise<LoginResponse>;
  verify2FA: (challengeID: string, code: string, trustDevice?: boolean) => Promise<void>;
  verifyBackupCode: (challengeID: string, code: string) => Promise<void>;
  logout: () => void;
  hasPermission: (module: string, resource: string, action: string) => boolean;
  hasRole: (role: string) => boolean;
}

// apps/web/src/hooks/use-permission.ts
function usePermission(module: string, resource: string, action: string): boolean;
function useRole(...roles: string[]): boolean;

// apps/web/src/hooks/use-two-factor.ts
interface UseTwoFactor {
  isEnabled: boolean;
  method: TwoFAMethod | null;
  enable2FA: (method: TwoFAMethod) => Promise<TwoFactorSetup>;
  verifySetup: (code: string) => Promise<void>;
  disable2FA: (password: string) => Promise<void>;
  regenerateBackupCodes: (password: string) => Promise<string[]>;
  trustedDevices: TrustedDevice[];
  revokeTrustedDevice: (deviceId: string) => Promise<void>;
  revokeAllDevices: () => Promise<void>;
}
function useTwoFactor(): UseTwoFactor;

// apps/web/src/components/auth/permission-guard.tsx
// Conditional rendering based on permission
<PermissionGuard module="crm" resource="client" action="create">
  <CreateClientButton />
</PermissionGuard>

// apps/web/src/components/auth/two-factor-dialog.tsx
// Modal hiển thị khi login cần 2FA
<TwoFactorDialog
  challenge={challenge}
  onVerify={(code, trustDevice) => verify2FA(challenge.id, code, trustDevice)}
  onUseBackupCode={(code) => verifyBackupCode(challenge.id, code)}
  onResendPush={() => resend2FAPush(challenge.id)}
/>

// apps/web/src/components/auth/two-factor-setup.tsx
// Component setup 2FA trong Settings: hiển thị QR code, nhập OTP xác minh
<TwoFactorSetup
  onEnable={(method) => enable2FA(method)}
  onVerify={(code) => verifySetup(code)}
  qrCode={setup?.qrCodeURL}
  backupCodes={setup?.backupCodes}
/>
```

---

## 3.2 Organization Service (Tổ chức)

### 3.2.1 Entities

```go
// internal/global/domain/organization.go

type Branch struct {
    ID        uuid.UUID `json:"id" db:"id"`
    Code      string    `json:"code" db:"code"`       // "HN", "HCM"
    Name      string    `json:"name" db:"name"`       // "Trụ sở chính", "Chi nhánh Miền Nam"
    Address   string    `json:"address" db:"address"`
    Phone     string    `json:"phone" db:"phone"`
    IsHQ      bool      `json:"is_hq" db:"is_hq"`    // Trụ sở chính?
    ManagerID *uuid.UUID `json:"manager_id" db:"manager_id"` // Giám đốc chi nhánh
    Status    string    `json:"status" db:"status"`   // active, inactive
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Department struct {
    ID        uuid.UUID  `json:"id" db:"id"`
    BranchID  uuid.UUID  `json:"branch_id" db:"branch_id"`
    Code      string     `json:"code" db:"code"`      // "KT1", "KT2", "KT_XDCB", "HC_KT"
    Name      string     `json:"name" db:"name"`      // "Phòng Kiểm toán - TVTC 1"
    Type      DeptType   `json:"type" db:"type"`      // audit, tax, admin, construction_audit
    HeadID    *uuid.UUID `json:"head_id" db:"head_id"`
    Status    string     `json:"status" db:"status"`
    CreatedAt time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

type DeptType string
const (
    DeptTypeAudit             DeptType = "audit"
    DeptTypeTax               DeptType = "tax"
    DeptTypeAdmin             DeptType = "admin"
    DeptTypeConstructionAudit DeptType = "construction_audit"
    DeptTypeFinancialAdvisory DeptType = "financial_advisory"
)
```

### 3.2.2 Service Interface

```go
type OrganizationService interface {
    // Branch
    CreateBranch(ctx context.Context, req CreateBranchRequest) (*Branch, error)
    UpdateBranch(ctx context.Context, id uuid.UUID, req UpdateBranchRequest) (*Branch, error)
    ListBranches(ctx context.Context) ([]Branch, error)
    GetBranch(ctx context.Context, id uuid.UUID) (*Branch, error)

    // Department
    CreateDepartment(ctx context.Context, req CreateDepartmentRequest) (*Department, error)
    UpdateDepartment(ctx context.Context, id uuid.UUID, req UpdateDepartmentRequest) (*Department, error)
    ListDepartments(ctx context.Context, branchID *uuid.UUID) ([]Department, error)
    GetDepartment(ctx context.Context, id uuid.UUID) (*Department, error)

    // Org chart
    GetOrgChart(ctx context.Context) (*OrgChartNode, error)
}
```

---

## 3.3 Audit Trail Service

Ghi nhận mọi thay đổi trong hệ thống — **bắt buộc** theo Luật Kiểm toán độc lập.

### 3.3.1 Entity

```go
// pkg/audit/entity.go

type AuditLog struct {
    ID         uuid.UUID       `json:"id" db:"id"`
    UserID     uuid.UUID       `json:"user_id" db:"user_id"`
    UserEmail  string          `json:"user_email" db:"user_email"`
    Module     string          `json:"module" db:"module"`         // "crm", "engagement"...
    Resource   string          `json:"resource" db:"resource"`     // "client", "contract"...
    ResourceID uuid.UUID       `json:"resource_id" db:"resource_id"`
    Action     string          `json:"action" db:"action"`         // "create", "update", "delete", "approve", "sign_off"
    Changes    json.RawMessage `json:"changes" db:"changes"`       // JSON diff {field: {old, new}}
    IPAddress  string          `json:"ip_address" db:"ip_address"`
    UserAgent  string          `json:"user_agent" db:"user_agent"`
    CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}
```

### 3.3.2 Service Interface

```go
type AuditService interface {
    Log(ctx context.Context, entry AuditLog) error
    LogFromContext(ctx context.Context, module, resource string, resourceID uuid.UUID, action string, changes interface{}) error
    GetLogs(ctx context.Context, filter AuditFilter) (*PaginatedResult[AuditLog], error)
    GetResourceHistory(ctx context.Context, module, resource string, resourceID uuid.UUID) ([]AuditLog, error)
}

// Middleware tự động capture audit log cho mọi mutation
func AuditMiddleware(auditSvc AuditService) gin.HandlerFunc
```

---

## 3.4 Notification Service

### 3.4.1 Entity

```go
// pkg/notification/entity.go

type Notification struct {
    ID          uuid.UUID          `json:"id" db:"id"`
    UserID      uuid.UUID          `json:"user_id" db:"user_id"`
    Type        NotificationType   `json:"type" db:"type"`
    Title       string             `json:"title" db:"title"`
    Message     string             `json:"message" db:"message"`
    Module      string             `json:"module" db:"module"`
    ResourceID  *uuid.UUID         `json:"resource_id" db:"resource_id"`
    ResourceURL string             `json:"resource_url" db:"resource_url"` // Deep link (web & mobile)
    Priority    NotificationPrio   `json:"priority" db:"priority"`
    Channels    []NotifChannel     `json:"channels" db:"channels"`         // Đã gửi qua kênh nào
    IsRead      bool               `json:"is_read" db:"is_read"`
    ReadAt      *time.Time         `json:"read_at" db:"read_at"`
    CreatedAt   time.Time          `json:"created_at" db:"created_at"`
}

// Kênh thông báo
type NotifChannel string
const (
    NotifChannelInApp  NotifChannel = "in_app"   // WebSocket real-time trong web app
    NotifChannelPush   NotifChannel = "push"     // Push notification tới mobile app (self-hosted)
    NotifChannelEmail  NotifChannel = "email"    // Email
)

type NotificationType string
const (
    NotifTypeApprovalRequired  NotificationType = "approval_required"
    NotifTypeDeadlineReminder  NotificationType = "deadline_reminder"
    NotifTypeAssignment        NotificationType = "assignment"        // *** Thông báo khi assign công việc ***
    NotifTypeTaskUpdated       NotificationType = "task_updated"      // Task thay đổi status
    NotifTypeReviewRequired    NotificationType = "review_required"
    NotifTypeTaxDeadline       NotificationType = "tax_deadline"
    NotifTypePaymentOverdue    NotificationType = "payment_overdue"
    NotifTypeCertExpiry        NotificationType = "certificate_expiry"
    NotifTypeSystemAlert       NotificationType = "system_alert"
    NotifType2FAChallenge      NotificationType = "two_fa_challenge"  // Push 2FA challenge
)

type NotificationPrio string
const (
    NotifPrioLow    NotificationPrio = "low"
    NotifPrioMedium NotificationPrio = "medium"
    NotifPrioHigh   NotificationPrio = "high"
    NotifPrioCritical NotificationPrio = "critical"
)
```

### 3.4.2 Service Interface

```go
type NotificationService interface {
    // Core — gửi thông báo qua TẤT CẢ kênh phù hợp (in-app + push + email tùy config)
    Send(ctx context.Context, notif Notification) error
    SendBulk(ctx context.Context, userIDs []uuid.UUID, notif Notification) error
    SendToRole(ctx context.Context, roleCode string, notif Notification) error
    MarkAsRead(ctx context.Context, userID, notifID uuid.UUID) error
    MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
    GetUnread(ctx context.Context, userID uuid.UUID) ([]Notification, error)
    GetAll(ctx context.Context, userID uuid.UUID, filter NotifFilter) (*PaginatedResult[Notification], error)

    // Assignment-specific — thông báo khi assign công việc (trigger push + in-app)
    NotifyAssignment(ctx context.Context, req AssignmentNotification) error
    NotifyTaskUpdate(ctx context.Context, req TaskUpdateNotification) error

    // Push Notification (self-hosted, không phụ thuộc bên thứ 3)
    RegisterPushDevice(ctx context.Context, userID uuid.UUID, req RegisterDeviceRequest) error
    UnregisterPushDevice(ctx context.Context, userID uuid.UUID, deviceToken string) error
    SendPush(ctx context.Context, userID uuid.UUID, payload PushPayload) error
    SendPushBulk(ctx context.Context, userIDs []uuid.UUID, payload PushPayload) error
    GetRegisteredDevices(ctx context.Context, userID uuid.UUID) ([]PushDevice, error)

    // Email notification
    SendEmail(ctx context.Context, to string, subject string, body string) error

    // User notification preferences
    GetPreferences(ctx context.Context, userID uuid.UUID) (*NotificationPreferences, error)
    UpdatePreferences(ctx context.Context, userID uuid.UUID, req UpdatePreferencesRequest) error
}

// Thông báo khi assign công việc
type AssignmentNotification struct {
    AssigneeID    uuid.UUID `json:"assignee_id"`
    AssignerID    uuid.UUID `json:"assigner_id"`
    EngagementID  uuid.UUID `json:"engagement_id"`
    TaskID        *uuid.UUID `json:"task_id"`
    Title         string    `json:"title"`          // "Bạn được phân công vào engagement KT-2026-0001"
    Description   string    `json:"description"`
    DueDate       *time.Time `json:"due_date"`
}

// Thông báo khi task thay đổi
type TaskUpdateNotification struct {
    TaskID       uuid.UUID `json:"task_id"`
    UpdatedBy    uuid.UUID `json:"updated_by"`
    NotifyUserIDs []uuid.UUID `json:"notify_user_ids"` // Team members
    ChangeType   string    `json:"change_type"`        // "status_changed", "reassigned", "due_date_changed"
    OldValue     string    `json:"old_value"`
    NewValue     string    `json:"new_value"`
}

// User notification preferences — user tự chọn kênh nhận thông báo
type NotificationPreferences struct {
    UserID                uuid.UUID `json:"user_id" db:"user_id"`
    AssignmentInApp       bool      `json:"assignment_in_app" db:"assignment_in_app"`           // Default: true
    AssignmentPush        bool      `json:"assignment_push" db:"assignment_push"`               // Default: true
    AssignmentEmail       bool      `json:"assignment_email" db:"assignment_email"`             // Default: false
    DeadlineReminderInApp bool      `json:"deadline_reminder_in_app" db:"deadline_reminder_in_app"` // Default: true
    DeadlineReminderPush  bool      `json:"deadline_reminder_push" db:"deadline_reminder_push"`     // Default: true
    DeadlineReminderEmail bool      `json:"deadline_reminder_email" db:"deadline_reminder_email"`   // Default: true
    ApprovalInApp         bool      `json:"approval_in_app" db:"approval_in_app"`               // Default: true
    ApprovalPush          bool      `json:"approval_push" db:"approval_push"`                   // Default: true
    ReviewInApp           bool      `json:"review_in_app" db:"review_in_app"`                   // Default: true
    ReviewPush            bool      `json:"review_push" db:"review_push"`                       // Default: true
    QuietHoursStart       *string   `json:"quiet_hours_start" db:"quiet_hours_start"`           // "22:00"
    QuietHoursEnd         *string   `json:"quiet_hours_end" db:"quiet_hours_end"`               // "07:00"
}
```

### 3.4.3 WebSocket cho Real-time

```go
// pkg/notification/websocket.go

type WSHub struct {
    clients    map[uuid.UUID]*WSClient
    register   chan *WSClient
    unregister chan *WSClient
    broadcast  chan *WSMessage
}

type WSMessage struct {
    UserID uuid.UUID       `json:"user_id"`
    Type   string          `json:"type"`   // "notification", "dashboard_update"
    Data   json.RawMessage `json:"data"`
}
```

```typescript
// Frontend: apps/web/src/hooks/use-notifications.ts
function useNotifications() {
  // WebSocket connection for real-time notifications
  // Returns: notifications[], unreadCount, markAsRead()
}
```

### 3.4.4 Self-Hosted Push Notification Service (Không phụ thuộc bên thứ 3)

**Yêu cầu từ stakeholder:** Hệ thống thông báo push khi assign công việc cho user qua mobile app, không phụ thuộc dịch vụ bên thứ 3 (không dùng Firebase Cloud Messaging, OneSignal, Pusher...).

#### Kiến trúc Push Notification Self-Hosted

```
┌─────────────────────────────────────────────────────────────────────┐
│                    PUSH NOTIFICATION FLOW                           │
│                                                                     │
│  [Manager assign task]                                              │
│         │                                                           │
│         ▼                                                           │
│  ┌──────────────┐    ┌───────────────────┐    ┌──────────────────┐  │
│  │ Engagement   │───▶│ Notification      │───▶│  Push Delivery   │  │
│  │ Service      │    │ Service           │    │  Service         │  │
│  │ (create task,│    │ (route to channels│    │  (self-hosted)   │  │
│  │  assign user)│    │  in-app + push)   │    │                  │  │
│  └──────────────┘    └───────────────────┘    └───────┬──────────┘  │
│                                                       │             │
│                              ┌────────────────────────┼─────┐       │
│                              │                        │     │       │
│                              ▼                        ▼     ▼       │
│                        ┌──────────┐  ┌──────────┐ ┌──────────────┐  │
│                        │ WebSocket│  │ Web Push │ │ Mobile Push  │  │
│                        │ (in-app) │  │ (browser)│ │ (SSE/WS      │  │
│                        │          │  │ W3C API  │ │  persistent) │  │
│                        └──────────┘  └──────────┘ └──────────────┘  │
│                              │            │              │          │
│                              ▼            ▼              ▼          │
│                        ┌──────────┐ ┌──────────┐  ┌──────────────┐  │
│                        │ Web App  │ │ Browser  │  │ Mobile App   │  │
│                        │ (Next.js)│ │ Notif Bar│  │ (React Native│  │
│                        │          │  │          │  │  self-built) │  │
│                        └──────────┘ └──────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

#### Entities

```go
// pkg/notification/push.go

// Push Device — thiết bị đã đăng ký nhận push
type PushDevice struct {
    ID            uuid.UUID      `json:"id" db:"id"`
    UserID        uuid.UUID      `json:"user_id" db:"user_id"`
    DeviceToken   string         `json:"-" db:"device_token"`       // Unique token của device
    Platform      DevicePlatform `json:"platform" db:"platform"`
    DeviceName    string         `json:"device_name" db:"device_name"` // "iPhone 15 Pro", "Samsung Galaxy S24"
    AppVersion    string         `json:"app_version" db:"app_version"`
    OSVersion     string         `json:"os_version" db:"os_version"`
    IsActive      bool           `json:"is_active" db:"is_active"`
    LastActiveAt  time.Time      `json:"last_active_at" db:"last_active_at"`
    CreatedAt     time.Time      `json:"created_at" db:"created_at"`
}

type DevicePlatform string
const (
    PlatformIOS     DevicePlatform = "ios"
    PlatformAndroid DevicePlatform = "android"
    PlatformWebPush DevicePlatform = "web_push"  // Browser Web Push API
)

// Push Payload — nội dung gửi tới device
type PushPayload struct {
    Title       string            `json:"title"`
    Body        string            `json:"body"`
    Icon        string            `json:"icon,omitempty"`
    Badge       int               `json:"badge,omitempty"`       // Số badge trên app icon
    Sound       string            `json:"sound,omitempty"`       // "default", "urgent"
    Data        map[string]string `json:"data,omitempty"`        // Custom data: {module, resource_id, action, deep_link}
    Priority    string            `json:"priority"`              // "high", "normal"
    TTL         int               `json:"ttl"`                   // Time to live (seconds)
    CollapseKey string            `json:"collapse_key,omitempty"` // Gộp nhiều notif cùng loại
}

// Push Delivery Log — log mỗi lần gửi push (để debug & retry)
type PushDeliveryLog struct {
    ID            uuid.UUID `json:"id" db:"id"`
    NotificationID uuid.UUID `json:"notification_id" db:"notification_id"`
    DeviceID      uuid.UUID `json:"device_id" db:"device_id"`
    Status        string    `json:"status" db:"status"`       // "sent", "delivered", "failed", "expired"
    ErrorMessage  string    `json:"error_message" db:"error_message"`
    SentAt        time.Time `json:"sent_at" db:"sent_at"`
    DeliveredAt   *time.Time `json:"delivered_at" db:"delivered_at"`
    RetryCount    int       `json:"retry_count" db:"retry_count"`
}

// Register Device Request — mobile app gọi khi đăng nhập
type RegisterDeviceRequest struct {
    DeviceToken  string         `json:"device_token" validate:"required"`
    Platform     DevicePlatform `json:"platform" validate:"required,oneof=ios android web_push"`
    DeviceName   string         `json:"device_name" validate:"required"`
    AppVersion   string         `json:"app_version"`
    OSVersion    string         `json:"os_version"`
}
```

#### Push Delivery Service (Self-Hosted)

```go
// pkg/notification/push_delivery.go

// PushDeliveryService — engine gửi push notification, KHÔNG dùng Firebase/OneSignal
// Cơ chế: Mobile app duy trì persistent connection (WebSocket hoặc SSE) tới server
// Server push message qua connection đó. Nếu device offline → queue & retry khi reconnect.

type PushDeliveryService interface {
    // Device registration
    RegisterDevice(ctx context.Context, userID uuid.UUID, req RegisterDeviceRequest) (*PushDevice, error)
    UnregisterDevice(ctx context.Context, userID uuid.UUID, deviceToken string) error
    HeartBeat(ctx context.Context, deviceToken string) error                   // Mobile app gọi định kỳ (5 phút)
    MarkDeviceInactive(ctx context.Context, deviceToken string) error

    // Push delivery
    DeliverToUser(ctx context.Context, userID uuid.UUID, payload PushPayload) error
    DeliverToDevice(ctx context.Context, deviceID uuid.UUID, payload PushPayload) error
    DeliverBulk(ctx context.Context, userIDs []uuid.UUID, payload PushPayload) error

    // Delivery tracking
    AcknowledgeDelivery(ctx context.Context, deliveryLogID uuid.UUID) error   // Device xác nhận đã nhận
    RetryFailedDeliveries(ctx context.Context) error                          // Cron job retry

    // Web Push (W3C Push API — chạy trên browser, KHÔNG cần bên thứ 3)
    GetVAPIDPublicKey(ctx context.Context) (string, error)                    // VAPID key cho Web Push
    SubscribeWebPush(ctx context.Context, userID uuid.UUID, subscription WebPushSubscription) error
    SendWebPush(ctx context.Context, subscription WebPushSubscription, payload PushPayload) error
}

// WebPushSubscription — từ browser Push API
type WebPushSubscription struct {
    Endpoint string `json:"endpoint"`
    Keys     struct {
        P256dh string `json:"p256dh"`
        Auth   string `json:"auth"`
    } `json:"keys"`
}
```

#### Cơ chế Self-Hosted Push cho Mobile App

```go
// pkg/notification/push_relay.go

// PushRelayServer — server quản lý persistent connections từ mobile app
// Mobile app kết nối WebSocket tới server, server push message qua connection
// Khi app ở background (Android/iOS) → dùng OS-level long-polling hoặc periodic sync

type PushRelayServer struct {
    connections  sync.Map                    // deviceToken → *PushConnection
    messageQueue chan PushQueueItem          // Queue cho messages chờ gửi
    offlineQueue *OfflineMessageStore        // Redis-backed queue cho device offline
}

type PushConnection struct {
    DeviceToken string
    UserID      uuid.UUID
    Platform    DevicePlatform
    Conn        *websocket.Conn             // Persistent WebSocket
    LastPing    time.Time
    IsAlive     bool
}

type PushQueueItem struct {
    DeviceToken string
    Payload     PushPayload
    CreatedAt   time.Time
    RetryCount  int
    MaxRetry    int                          // Default: 3
    TTL         time.Duration                // Default: 24h
}

// OfflineMessageStore — lưu message khi device offline, gửi khi reconnect
type OfflineMessageStore interface {
    Enqueue(ctx context.Context, deviceToken string, payload PushPayload, ttl time.Duration) error
    Dequeue(ctx context.Context, deviceToken string) ([]PushPayload, error)  // Lấy tất cả pending messages
    Flush(ctx context.Context, deviceToken string) error                     // Xóa sau khi gửi thành công
    GetQueueSize(ctx context.Context, deviceToken string) (int, error)
}
```

#### Web Push API (W3C Standard — Browser, không cần bên thứ 3)

```go
// pkg/notification/web_push.go
// Sử dụng W3C Web Push Protocol + VAPID authentication
// Thư viện Go: github.com/SherClockHolmes/webpush-go

type WebPushService struct {
    vapidPublicKey  string    // VAPID public key (tự generate)
    vapidPrivateKey string    // VAPID private key (tự generate)
    vapidContact    string    // "mailto:admin@company.vn"
}
```

#### Mobile App Architecture (React Native — Self-Built)

```
erp-mobile/
├── src/
│   ├── services/
│   │   ├── push-connection.ts       # Persistent WebSocket tới push relay server
│   │   ├── push-handler.ts          # Xử lý notification khi nhận
│   │   ├── background-sync.ts       # Periodic sync khi app ở background
│   │   └── two-factor.ts            # 2FA push approve/reject handler
│   ├── screens/
│   │   ├── LoginScreen.tsx           # Login + 2FA
│   │   ├── NotificationsScreen.tsx   # Danh sách thông báo
│   │   ├── TaskDetailScreen.tsx      # Xem task được assign
│   │   ├── TwoFactorScreen.tsx       # Approve/Reject 2FA push
│   │   └── SettingsScreen.tsx        # Notification preferences
│   ├── hooks/
│   │   ├── use-push-notifications.ts # Hook quản lý push connection
│   │   └── use-auth.ts               # Auth + 2FA hook
│   ├── navigation/
│   │   └── deep-linking.ts           # Deep link từ notification → màn hình cụ thể
│   └── utils/
│       └── device-info.ts            # Device fingerprint, token
├── android/                          # Android-specific config
│   └── app/src/main/java/.../
│       └── PushConnectionService.java # Foreground service giữ connection
├── ios/                              # iOS-specific config
│   └── PushConnection/
│       └── BackgroundTaskManager.swift # Background fetch scheduling
└── package.json
```

```typescript
// erp-mobile/src/services/push-connection.ts

class PushConnectionService {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private reconnectDelay = 1000; // Exponential backoff

  // Kết nối persistent WebSocket tới push relay server
  connect(serverURL: string, deviceToken: string, authToken: string): void {
    this.ws = new WebSocket(`${serverURL}/ws/push?token=${deviceToken}`);

    this.ws.onmessage = (event) => {
      const message: PushMessage = JSON.parse(event.data);
      this.handlePushMessage(message);
    };

    this.ws.onclose = () => {
      this.scheduleReconnect(); // Auto-reconnect với exponential backoff
    };

    // Heartbeat mỗi 30 giây để giữ connection alive
    this.startHeartbeat();
  }

  private handlePushMessage(message: PushMessage): void {
    switch (message.type) {
      case 'assignment':
        // Hiển thị local notification trên device
        this.showLocalNotification(message.payload);
        // Deep link tới task detail
        break;
      case 'two_fa_challenge':
        // Hiển thị 2FA approve/reject dialog
        this.show2FAChallenge(message.payload);
        break;
      case 'approval_required':
        this.showLocalNotification(message.payload);
        break;
      default:
        this.showLocalNotification(message.payload);
    }
  }

  // Acknowledge — báo server đã nhận thành công
  acknowledge(deliveryLogID: string): void {
    this.ws?.send(JSON.stringify({ type: 'ack', delivery_log_id: deliveryLogID }));
  }
}
```

#### API Endpoints — Push Notification & Device Management

```
# Device Registration (mobile app gọi sau khi login)
POST   /api/v1/push/devices                       # Đăng ký device
DELETE /api/v1/push/devices/:deviceToken           # Hủy đăng ký
GET    /api/v1/push/devices                        # Danh sách device của user
POST   /api/v1/push/devices/heartbeat              # Heartbeat từ mobile app

# Web Push (browser)
GET    /api/v1/push/vapid-key                      # Lấy VAPID public key
POST   /api/v1/push/web-subscribe                  # Đăng ký Web Push subscription

# Push Relay (WebSocket endpoint cho mobile app)
WS     /ws/push?token={deviceToken}                # Persistent connection

# Notification Preferences
GET    /api/v1/notifications/preferences           # Lấy preferences
PUT    /api/v1/notifications/preferences           # Cập nhật preferences

# Delivery logs (admin)
GET    /api/v1/admin/push/delivery-logs            # Log gửi push (debug)
POST   /api/v1/admin/push/retry-failed             # Retry failed deliveries

# 2FA Push Response (mobile app gọi)
POST   /api/v1/auth/2fa/push-response              # Approve/Reject 2FA từ mobile app
```

#### Trigger Points — Khi nào gửi Push Notification

| Sự kiện | Push? | In-app? | Email? | Mô tả |
|---|---|---|---|---|
| Assign vào engagement | ✅ | ✅ | ❌ | "Bạn được phân công vào KT-2026-0001" |
| Assign task mới | ✅ | ✅ | ❌ | "Task mới: Kiểm tra khoản phải thu" |
| Task due date sắp đến (1 ngày) | ✅ | ✅ | ✅ | "Task XYZ sẽ đến hạn ngày mai" |
| Yêu cầu review/approve | ✅ | ✅ | ❌ | "Có hồ sơ chờ bạn review" |
| Hồ sơ bị reject | ✅ | ✅ | ❌ | "Hồ sơ A1010 bị từ chối, cần sửa" |
| Tax deadline ≤ 7 ngày | ✅ | ✅ | ✅ | "Tờ khai VAT Q1 đến hạn ngày..." |
| Invoice quá hạn | ❌ | ✅ | ✅ | "Hóa đơn INV-001 quá hạn 15 ngày" |
| Chứng chỉ sắp hết hạn (90 ngày) | ❌ | ✅ | ✅ | "CPA cert hết hạn ngày..." |
| 2FA login challenge | ✅ (priority) | ❌ | ❌ | "Xác nhận đăng nhập từ Chrome..." |

---

## 3.5 File Storage Service

### 3.5.1 Entity

```go
// pkg/storage/entity.go

type FileMetadata struct {
    ID           uuid.UUID `json:"id" db:"id"`
    OriginalName string    `json:"original_name" db:"original_name"`
    StoragePath  string    `json:"-" db:"storage_path"`        // Internal MinIO path
    MimeType     string    `json:"mime_type" db:"mime_type"`
    Size         int64     `json:"size" db:"size"`             // bytes
    Module       string    `json:"module" db:"module"`         // "working_paper", "engagement"
    ResourceID   uuid.UUID `json:"resource_id" db:"resource_id"`
    UploadedBy   uuid.UUID `json:"uploaded_by" db:"uploaded_by"`
    Version      int       `json:"version" db:"version"`       // File versioning
    Checksum     string    `json:"checksum" db:"checksum"`     // SHA-256 integrity check
    IsEncrypted  bool      `json:"is_encrypted" db:"is_encrypted"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
```

### 3.5.2 Service Interface

```go
type StorageService interface {
    Upload(ctx context.Context, req UploadRequest) (*FileMetadata, error)
    Download(ctx context.Context, fileID uuid.UUID) (io.ReadCloser, *FileMetadata, error)
    GetPresignedURL(ctx context.Context, fileID uuid.UUID, expiry time.Duration) (string, error)
    Delete(ctx context.Context, fileID uuid.UUID) error
    ListByResource(ctx context.Context, module string, resourceID uuid.UUID) ([]FileMetadata, error)
    GetVersionHistory(ctx context.Context, fileID uuid.UUID) ([]FileMetadata, error)

    // Bulk operations
    UploadBulk(ctx context.Context, files []UploadRequest) ([]FileMetadata, error)
    DownloadAsZip(ctx context.Context, fileIDs []uuid.UUID) (io.ReadCloser, error)
}

type UploadRequest struct {
    File       io.Reader
    FileName   string
    MimeType   string
    Module     string
    ResourceID uuid.UUID
    Encrypt    bool
}
```

---

## 3.6 Export Service (Excel, PDF, Word)

```go
// pkg/export/service.go

type ExportService interface {
    ExportToExcel(ctx context.Context, req ExcelExportRequest) ([]byte, error)
    ExportToPDF(ctx context.Context, req PDFExportRequest) ([]byte, error)
    ExportToWord(ctx context.Context, req WordExportRequest) ([]byte, error)

    // Template-based export
    RenderTemplate(ctx context.Context, templateID uuid.UUID, data interface{}) ([]byte, error)
}

type ExcelExportRequest struct {
    SheetName string
    Headers   []string
    Rows      [][]interface{}
    Title     string
}

type PDFExportRequest struct {
    TemplateName string
    Data         interface{}
    Orientation  string // "portrait", "landscape"
}
```

---

## 3.7 Approval Workflow Engine

Dùng chung cho: phê duyệt hợp đồng, phê duyệt timesheet, phê duyệt hóa đơn, sign-off hồ sơ kiểm toán.

### 3.7.1 Entities

```go
// pkg/workflow/entity.go

type WorkflowDefinition struct {
    ID          uuid.UUID        `json:"id" db:"id"`
    Code        string           `json:"code" db:"code"`        // "contract_approval", "invoice_approval", "audit_signoff"
    Name        string           `json:"name" db:"name"`
    Module      string           `json:"module" db:"module"`
    Steps       []WorkflowStep   `json:"steps"`
    IsActive    bool             `json:"is_active" db:"is_active"`
    CreatedAt   time.Time        `json:"created_at" db:"created_at"`
}

type WorkflowStep struct {
    Order       int       `json:"order"`
    Name        string    `json:"name"`       // "Senior Review", "Manager Approval", "Partner Sign-off"
    RoleCode    string    `json:"role_code"`   // Role required to approve
    MinLevel    int       `json:"min_level"`   // Minimum hierarchy level
    IsOptional  bool      `json:"is_optional"`
    TimeoutDays int       `json:"timeout_days"` // Auto-escalate after N days
}

type WorkflowInstance struct {
    ID              uuid.UUID         `json:"id" db:"id"`
    DefinitionID    uuid.UUID         `json:"definition_id" db:"definition_id"`
    Module          string            `json:"module" db:"module"`
    ResourceID      uuid.UUID         `json:"resource_id" db:"resource_id"`
    CurrentStep     int               `json:"current_step" db:"current_step"`
    Status          WorkflowStatus    `json:"status" db:"status"`
    InitiatedBy     uuid.UUID         `json:"initiated_by" db:"initiated_by"`
    CreatedAt       time.Time         `json:"created_at" db:"created_at"`
    CompletedAt     *time.Time        `json:"completed_at" db:"completed_at"`
}

type WorkflowStatus string
const (
    WFStatusPending   WorkflowStatus = "pending"
    WFStatusInProgress WorkflowStatus = "in_progress"
    WFStatusApproved  WorkflowStatus = "approved"
    WFStatusRejected  WorkflowStatus = "rejected"
    WFStatusCancelled WorkflowStatus = "cancelled"
)

type WorkflowAction struct {
    ID             uuid.UUID `json:"id" db:"id"`
    InstanceID     uuid.UUID `json:"instance_id" db:"instance_id"`
    StepOrder      int       `json:"step_order" db:"step_order"`
    ActorID        uuid.UUID `json:"actor_id" db:"actor_id"`
    Action         string    `json:"action" db:"action"`         // "approve", "reject", "request_changes"
    Comment        string    `json:"comment" db:"comment"`
    CreatedAt      time.Time `json:"created_at" db:"created_at"`
}
```

### 3.7.2 Service Interface

```go
type WorkflowService interface {
    // Definition management
    CreateDefinition(ctx context.Context, req CreateWorkflowDefRequest) (*WorkflowDefinition, error)
    GetDefinition(ctx context.Context, code string) (*WorkflowDefinition, error)

    // Instance management
    StartWorkflow(ctx context.Context, definitionCode string, module string, resourceID uuid.UUID) (*WorkflowInstance, error)
    ApproveStep(ctx context.Context, instanceID uuid.UUID, comment string) (*WorkflowInstance, error)
    RejectStep(ctx context.Context, instanceID uuid.UUID, comment string) (*WorkflowInstance, error)
    RequestChanges(ctx context.Context, instanceID uuid.UUID, comment string) (*WorkflowInstance, error)
    CancelWorkflow(ctx context.Context, instanceID uuid.UUID) error

    // Query
    GetPendingApprovals(ctx context.Context, userID uuid.UUID) ([]WorkflowInstance, error)
    GetInstanceHistory(ctx context.Context, instanceID uuid.UUID) ([]WorkflowAction, error)
    GetInstanceByResource(ctx context.Context, module string, resourceID uuid.UUID) (*WorkflowInstance, error)
}
```

### 3.7.3 Predefined Workflows

| Code | Steps | Áp dụng cho |
|---|---|---|
| `contract_approval` | Staff → Manager → Director/Partner → Seal(Admin) | Phê duyệt hợp đồng |
| `invoice_approval` | Accountant → Manager → Director | Phê duyệt hóa đơn |
| `audit_signoff` | Auditor → Senior → Manager → Partner | Sign-off hồ sơ kiểm toán |
| `timesheet_approval` | Employee → Manager | Phê duyệt timesheet |
| `client_acceptance` | Staff → Manager → Partner | Chấp nhận khách hàng mới |

---

## 3.8 Shared Pagination, Filtering, Sorting

```go
// pkg/database/pagination.go

type PaginationParams struct {
    Page     int    `form:"page" json:"page" validate:"min=1"`
    PageSize int    `form:"page_size" json:"page_size" validate:"min=1,max=100"`
    SortBy   string `form:"sort_by" json:"sort_by"`
    SortDir  string `form:"sort_dir" json:"sort_dir" validate:"oneof=asc desc"`
}

type PaginatedResult[T any] struct {
    Data       []T   `json:"data"`
    Total      int64 `json:"total"`
    Page       int   `json:"page"`
    PageSize   int   `json:"page_size"`
    TotalPages int   `json:"total_pages"`
}

type FilterOperator string
const (
    FilterEq       FilterOperator = "eq"
    FilterNeq      FilterOperator = "neq"
    FilterGt       FilterOperator = "gt"
    FilterGte      FilterOperator = "gte"
    FilterLt       FilterOperator = "lt"
    FilterLte      FilterOperator = "lte"
    FilterIn       FilterOperator = "in"
    FilterLike     FilterOperator = "like"
    FilterBetween  FilterOperator = "between"
    FilterIsNull   FilterOperator = "is_null"
)

type FilterCondition struct {
    Field    string         `json:"field"`
    Operator FilterOperator `json:"operator"`
    Value    interface{}    `json:"value"`
}
```

```typescript
// Frontend: apps/web/src/hooks/use-pagination.ts
interface UsePaginationOptions {
  initialPage?: number;
  initialPageSize?: number;
  initialSortBy?: string;
  initialSortDir?: 'asc' | 'desc';
}

function usePagination<T>(
  queryKey: string[],
  fetchFn: (params: PaginationParams) => Promise<PaginatedResult<T>>,
  options?: UsePaginationOptions
): {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
  isLoading: boolean;
  setPage: (page: number) => void;
  setPageSize: (size: number) => void;
  setSortBy: (field: string) => void;
  toggleSortDir: () => void;
  refetch: () => void;
};
```

---

## 3.9 Common Enums & Constants

```go
// pkg/common/enums.go

// Service Types
type ServiceType string
const (
    ServiceFinancialAudit     ServiceType = "financial_audit"      // Kiểm toán BCTC
    ServiceInternalAudit      ServiceType = "internal_audit"       // Kiểm toán nội bộ
    ServiceConstructionAudit  ServiceType = "construction_audit"   // Kiểm toán XDCB
    ServiceTaxConsulting      ServiceType = "tax_consulting"       // Tư vấn thuế
    ServiceFinancialAdvisory  ServiceType = "financial_advisory"   // Tư vấn tài chính
    ServiceReview             ServiceType = "review"               // Soát xét
)

// Engagement Status
type EngagementStatus string
const (
    EngStatusDraft       EngagementStatus = "draft"
    EngStatusProposal    EngagementStatus = "proposal"      // Chào phí
    EngStatusContracted  EngagementStatus = "contracted"     // Đã ký HĐ
    EngStatusInProgress  EngagementStatus = "in_progress"    // Đang thực hiện
    EngStatusReview      EngagementStatus = "review"         // Đang review
    EngStatusCompleted   EngagementStatus = "completed"      // Hoàn thành
    EngStatusSettled     EngagementStatus = "settled"        // Thanh lý
    EngStatusCancelled   EngagementStatus = "cancelled"
)

// Fiscal Year End
type FiscalYearEnd string
const (
    FYEnd1231 FiscalYearEnd = "12-31"  // 31/12
    FYEnd0630 FiscalYearEnd = "06-30"  // 30/6
    FYEnd0331 FiscalYearEnd = "03-31"  // 31/3
)

// Employee Grade (cấp bậc)
type EmployeeGrade string
const (
    GradeIntern    EmployeeGrade = "intern"
    GradeJunior    EmployeeGrade = "junior"
    GradeSenior    EmployeeGrade = "senior"
    GradeManager   EmployeeGrade = "manager"
    GradeDirector  EmployeeGrade = "director"
    GradePartner   EmployeeGrade = "partner"
)
```

---

# 4. MODULE 1: CRM – QUẢN LÝ KHÁCH HÀNG

## 4.1 Mục tiêu

- Quản lý thông tin khách hàng tập trung (single source of truth)
- Đánh giá rủi ro khách hàng (Client Risk Assessment) trước khi nhận hợp đồng
- Kiểm tra xung đột lợi ích (Conflict of Interest) tự động
- Theo dõi lịch sử toàn bộ dịch vụ đã cung cấp cho từng khách hàng

## 4.2 Domain Entities

```go
// internal/crm/domain/entity.go

type Client struct {
    ID                uuid.UUID       `json:"id" db:"id"`
    Code              string          `json:"code" db:"code"`               // Auto-gen: KH-0001
    CompanyName       string          `json:"company_name" db:"company_name"`
    TaxCode           string          `json:"tax_code" db:"tax_code"`       // MST — unique
    BusinessType      BusinessType    `json:"business_type" db:"business_type"` // TNHH, CP, DNTN...
    Industry          string          `json:"industry" db:"industry"`       // Ngành nghề
    Address           string          `json:"address" db:"address"`
    Phone             string          `json:"phone" db:"phone"`
    Email             string          `json:"email" db:"email"`
    Website           string          `json:"website" db:"website"`
    FiscalYearEnd     FiscalYearEnd   `json:"fiscal_year_end" db:"fiscal_year_end"` // 31/12, 30/6...
    RepresentativeName string         `json:"representative_name" db:"representative_name"` // Người đại diện
    RepresentativeTitle string        `json:"representative_title" db:"representative_title"`
    RepresentativePhone string        `json:"representative_phone" db:"representative_phone"`
    RepresentativeEmail string        `json:"representative_email" db:"representative_email"`
    RiskLevel         RiskLevel       `json:"risk_level" db:"risk_level"`   // low, medium, high, critical
    RiskAssessmentID  *uuid.UUID      `json:"risk_assessment_id" db:"risk_assessment_id"`
    Status            ClientStatus    `json:"status" db:"status"`           // prospect, active, inactive, blacklisted
    Source            string          `json:"source" db:"source"`           // Nguồn giới thiệu
    AssignedPartnerID *uuid.UUID      `json:"assigned_partner_id" db:"assigned_partner_id"` // Partner phụ trách
    BranchID          uuid.UUID       `json:"branch_id" db:"branch_id"`
    Notes             string          `json:"notes" db:"notes"`
    CreatedBy         uuid.UUID       `json:"created_by" db:"created_by"`
    CreatedAt         time.Time       `json:"created_at" db:"created_at"`
    UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`
}

type BusinessType string
const (
    BizTypeTNHH BusinessType = "tnhh"         // TNHH
    BizTypeCP   BusinessType = "co_phan"       // Cổ phần
    BizTypeDNTN BusinessType = "dntn"          // DNTN
    BizTypeHKD  BusinessType = "ho_kinh_doanh" // Hộ kinh doanh
    BizTypeNN   BusinessType = "nha_nuoc"      // Nhà nước
    BizTypeFDI  BusinessType = "fdi"           // FDI
    BizTypeOther BusinessType = "other"
)

type ClientStatus string
const (
    ClientStatusProspect    ClientStatus = "prospect"    // Tiềm năng
    ClientStatusActive      ClientStatus = "active"      // Đang hoạt động
    ClientStatusInactive    ClientStatus = "inactive"    // Ngừng hợp tác
    ClientStatusBlacklisted ClientStatus = "blacklisted" // Từ chối
)

type RiskLevel string
const (
    RiskLow      RiskLevel = "low"
    RiskMedium   RiskLevel = "medium"
    RiskHigh     RiskLevel = "high"
    RiskCritical RiskLevel = "critical"
)

// Người liên hệ trực tiếp (có thể nhiều)
type ClientContact struct {
    ID          uuid.UUID `json:"id" db:"id"`
    ClientID    uuid.UUID `json:"client_id" db:"client_id"`
    FullName    string    `json:"full_name" db:"full_name"`
    Position    string    `json:"position" db:"position"`
    Phone       string    `json:"phone" db:"phone"`
    Email       string    `json:"email" db:"email"`
    IsPrimary   bool      `json:"is_primary" db:"is_primary"`
    HasInfluence bool     `json:"has_influence" db:"has_influence"` // Có ảnh hưởng đến HĐ?
    Notes       string    `json:"notes" db:"notes"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Đánh giá rủi ro khách hàng
type ClientRiskAssessment struct {
    ID                uuid.UUID       `json:"id" db:"id"`
    ClientID          uuid.UUID       `json:"client_id" db:"client_id"`
    AssessmentDate    time.Time       `json:"assessment_date" db:"assessment_date"`
    AssessorID        uuid.UUID       `json:"assessor_id" db:"assessor_id"`
    OverallRisk       RiskLevel       `json:"overall_risk" db:"overall_risk"`
    Criteria          []RiskCriterion `json:"criteria"`
    PublicInfoNotes   string          `json:"public_info_notes" db:"public_info_notes"` // Thông tin đại chúng
    ManagementIntegrity string       `json:"management_integrity" db:"management_integrity"` // Đánh giá ban lãnh đạo
    FinancialStability  string       `json:"financial_stability" db:"financial_stability"`
    IndustryRisk      string          `json:"industry_risk" db:"industry_risk"`
    LitigationHistory string          `json:"litigation_history" db:"litigation_history"` // Lịch sử kiện tụng
    Decision          string          `json:"decision" db:"decision"`       // accept, reject, conditional
    DecisionReason    string          `json:"decision_reason" db:"decision_reason"`
    ApprovedBy        *uuid.UUID      `json:"approved_by" db:"approved_by"` // Partner/Director phê duyệt
    ApprovedAt        *time.Time      `json:"approved_at" db:"approved_at"`
    Status            string          `json:"status" db:"status"`           // draft, submitted, approved, rejected
    CreatedAt         time.Time       `json:"created_at" db:"created_at"`
}

type RiskCriterion struct {
    Category    string    `json:"category"`    // "financial", "legal", "operational", "reputation"
    Question    string    `json:"question"`
    Score       int       `json:"score"`       // 1-5
    Weight      float64   `json:"weight"`      // 0.0-1.0
    Notes       string    `json:"notes"`
}

// Kiểm tra xung đột lợi ích
type ConflictCheck struct {
    ID              uuid.UUID `json:"id" db:"id"`
    ClientID        uuid.UUID `json:"client_id" db:"client_id"`
    CheckDate       time.Time `json:"check_date" db:"check_date"`
    CheckedBy       uuid.UUID `json:"checked_by" db:"checked_by"`
    ConflictFound   bool      `json:"conflict_found" db:"conflict_found"`
    ConflictDetails string    `json:"conflict_details" db:"conflict_details"`
    RelatedClients  []uuid.UUID `json:"related_clients"`                     // Các KH liên quan
    Resolution      string    `json:"resolution" db:"resolution"`
    Status          string    `json:"status" db:"status"`                   // clear, conflict_detected, resolved
    CreatedAt       time.Time `json:"created_at" db:"created_at"`
}
```

## 4.3 Repository Interface

```go
// internal/crm/domain/repository.go

type ClientRepository interface {
    Create(ctx context.Context, client *Client) error
    Update(ctx context.Context, client *Client) error
    GetByID(ctx context.Context, id uuid.UUID) (*Client, error)
    GetByTaxCode(ctx context.Context, taxCode string) (*Client, error)
    List(ctx context.Context, filter ClientFilter, pagination PaginationParams) (*PaginatedResult[Client], error)
    Search(ctx context.Context, query string, limit int) ([]Client, error)
    Delete(ctx context.Context, id uuid.UUID) error

    // Contacts
    CreateContact(ctx context.Context, contact *ClientContact) error
    UpdateContact(ctx context.Context, contact *ClientContact) error
    ListContacts(ctx context.Context, clientID uuid.UUID) ([]ClientContact, error)
    DeleteContact(ctx context.Context, id uuid.UUID) error

    // Risk Assessment
    CreateRiskAssessment(ctx context.Context, assessment *ClientRiskAssessment) error
    UpdateRiskAssessment(ctx context.Context, assessment *ClientRiskAssessment) error
    GetLatestRiskAssessment(ctx context.Context, clientID uuid.UUID) (*ClientRiskAssessment, error)
    ListRiskAssessments(ctx context.Context, clientID uuid.UUID) ([]ClientRiskAssessment, error)

    // Conflict Check
    CreateConflictCheck(ctx context.Context, check *ConflictCheck) error
    FindConflicts(ctx context.Context, clientID uuid.UUID) ([]ConflictCheck, error)
    CheckConflictByTaxCode(ctx context.Context, taxCode string, excludeClientID uuid.UUID) ([]Client, error)
    CheckConflictByName(ctx context.Context, name string, excludeClientID uuid.UUID) ([]Client, error)
    CheckConflictByRelatedParties(ctx context.Context, clientID uuid.UUID) ([]ConflictCheck, error)
}

type ClientFilter struct {
    Status        *ClientStatus  `form:"status"`
    RiskLevel     *RiskLevel     `form:"risk_level"`
    BranchID      *uuid.UUID     `form:"branch_id"`
    PartnerID     *uuid.UUID     `form:"partner_id"`
    BusinessType  *BusinessType  `form:"business_type"`
    Industry      *string        `form:"industry"`
    FiscalYearEnd *FiscalYearEnd `form:"fiscal_year_end"`
    SearchQuery   *string        `form:"q"` // Tìm theo tên, MST
    CreatedFrom   *time.Time     `form:"created_from"`
    CreatedTo     *time.Time     `form:"created_to"`
}
```

## 4.4 Use Cases / Service Interface

```go
// internal/crm/domain/service.go

type ClientService interface {
    // CRUD
    CreateClient(ctx context.Context, req CreateClientRequest) (*Client, error)
    UpdateClient(ctx context.Context, id uuid.UUID, req UpdateClientRequest) (*Client, error)
    GetClient(ctx context.Context, id uuid.UUID) (*ClientDetailResponse, error)
    ListClients(ctx context.Context, filter ClientFilter, pagination PaginationParams) (*PaginatedResult[ClientListItem], error)
    SearchClients(ctx context.Context, query string) ([]ClientSearchResult, error)

    // Client Acceptance (quy trình tiếp nhận KH mới)
    InitiateClientAcceptance(ctx context.Context, clientID uuid.UUID) error // Bắt đầu workflow chấp nhận KH
    PerformRiskAssessment(ctx context.Context, req CreateRiskAssessmentRequest) (*ClientRiskAssessment, error)
    ApproveClientAcceptance(ctx context.Context, clientID uuid.UUID, comment string) error
    RejectClientAcceptance(ctx context.Context, clientID uuid.UUID, reason string) error

    // Conflict of Interest
    RunConflictCheck(ctx context.Context, clientID uuid.UUID) (*ConflictCheckResult, error)
    ResolveConflict(ctx context.Context, conflictID uuid.UUID, resolution string) error

    // Contacts
    AddContact(ctx context.Context, clientID uuid.UUID, req CreateContactRequest) (*ClientContact, error)
    UpdateContact(ctx context.Context, contactID uuid.UUID, req UpdateContactRequest) (*ClientContact, error)
    RemoveContact(ctx context.Context, contactID uuid.UUID) error
    ListContacts(ctx context.Context, clientID uuid.UUID) ([]ClientContact, error)

    // Service History
    GetServiceHistory(ctx context.Context, clientID uuid.UUID) ([]EngagementSummary, error)

    // Export
    ExportClients(ctx context.Context, filter ClientFilter, format string) ([]byte, error) // "excel", "pdf"
}
```

## 4.5 DTOs

```go
// internal/crm/usecase/dto.go

type CreateClientRequest struct {
    CompanyName         string       `json:"company_name" validate:"required,min=2,max=255"`
    TaxCode             string       `json:"tax_code" validate:"required,len=10|len=13|len=14"` // MST 10 hoặc 13-14 ký tự
    BusinessType        BusinessType `json:"business_type" validate:"required"`
    Industry            string       `json:"industry" validate:"required"`
    Address             string       `json:"address" validate:"required"`
    Phone               string       `json:"phone"`
    Email               string       `json:"email" validate:"omitempty,email"`
    FiscalYearEnd       FiscalYearEnd `json:"fiscal_year_end" validate:"required"`
    RepresentativeName  string       `json:"representative_name" validate:"required"`
    RepresentativeTitle string       `json:"representative_title"`
    RepresentativePhone string       `json:"representative_phone"`
    RepresentativeEmail string       `json:"representative_email" validate:"omitempty,email"`
    Source              string       `json:"source"`
    BranchID            uuid.UUID    `json:"branch_id" validate:"required"`
    Notes               string       `json:"notes"`
    Contacts            []CreateContactRequest `json:"contacts"`
}

type CreateContactRequest struct {
    FullName     string `json:"full_name" validate:"required"`
    Position     string `json:"position"`
    Phone        string `json:"phone"`
    Email        string `json:"email" validate:"omitempty,email"`
    IsPrimary    bool   `json:"is_primary"`
    HasInfluence bool   `json:"has_influence"`
    Notes        string `json:"notes"`
}

type ClientDetailResponse struct {
    Client
    Contacts         []ClientContact            `json:"contacts"`
    LatestRisk       *ClientRiskAssessment       `json:"latest_risk_assessment"`
    ActiveEngagements []EngagementSummary        `json:"active_engagements"`
    TotalEngagements int                         `json:"total_engagements"`
    TotalRevenue     int64                       `json:"total_revenue"` // Tổng doanh thu từ KH
    OutstandingAR    int64                       `json:"outstanding_ar"` // Công nợ chưa thu
    LastConflictCheck *ConflictCheck             `json:"last_conflict_check"`
}

type ClientListItem struct {
    ID            uuid.UUID    `json:"id"`
    Code          string       `json:"code"`
    CompanyName   string       `json:"company_name"`
    TaxCode       string       `json:"tax_code"`
    BusinessType  BusinessType `json:"business_type"`
    RiskLevel     RiskLevel    `json:"risk_level"`
    Status        ClientStatus `json:"status"`
    ActiveEngCount int         `json:"active_engagement_count"`
    PartnerName   string       `json:"partner_name"`
    BranchName    string       `json:"branch_name"`
}

type ConflictCheckResult struct {
    HasConflict      bool              `json:"has_conflict"`
    ConflictsByTaxCode []Client        `json:"conflicts_by_tax_code"`
    ConflictsByName    []Client        `json:"conflicts_by_name"`
    RelatedPartyConflicts []ConflictCheck `json:"related_party_conflicts"`
    Recommendations  string            `json:"recommendations"`
}
```

## 4.6 API Endpoints

```
POST   /api/v1/clients                          # Tạo khách hàng
GET    /api/v1/clients                          # Danh sách KH (paginated, filtered)
GET    /api/v1/clients/:id                      # Chi tiết KH
PUT    /api/v1/clients/:id                      # Cập nhật KH
DELETE /api/v1/clients/:id                      # Xóa mềm KH

GET    /api/v1/clients/search?q=               # Tìm kiếm KH
POST   /api/v1/clients/export                   # Xuất danh sách KH

# Contacts
POST   /api/v1/clients/:id/contacts              # Thêm liên hệ
PUT    /api/v1/clients/:id/contacts/:contactId    # Sửa liên hệ
DELETE /api/v1/clients/:id/contacts/:contactId    # Xóa liên hệ

# Risk Assessment
POST   /api/v1/clients/:id/risk-assessments           # Tạo đánh giá rủi ro
GET    /api/v1/clients/:id/risk-assessments           # Lịch sử đánh giá
GET    /api/v1/clients/:id/risk-assessments/latest    # Đánh giá gần nhất

# Client Acceptance Workflow
POST   /api/v1/clients/:id/acceptance/initiate   # Bắt đầu quy trình chấp nhận KH
POST   /api/v1/clients/:id/acceptance/approve    # Phê duyệt
POST   /api/v1/clients/:id/acceptance/reject     # Từ chối

# Conflict of Interest
POST   /api/v1/clients/:id/conflict-check        # Chạy kiểm tra xung đột
GET    /api/v1/clients/:id/conflict-checks        # Lịch sử kiểm tra
POST   /api/v1/clients/:id/conflict-checks/:checkId/resolve  # Giải quyết xung đột

# Service History
GET    /api/v1/clients/:id/engagements            # Lịch sử dịch vụ
GET    /api/v1/clients/:id/invoices               # Lịch sử hóa đơn
GET    /api/v1/clients/:id/ar-summary             # Tổng hợp công nợ
```

## 4.7 Frontend Pages

```
/clients                          → ClientListPage (DataTable, filters, search)
/clients/new                      → CreateClientPage (multi-step form)
/clients/[id]                     → ClientDetailPage (tabs: info, contacts, risk, engagements, invoices)
/clients/[id]/edit                → EditClientPage
/clients/[id]/risk-assessment/new → RiskAssessmentFormPage (scoring form)
/clients/[id]/conflict-check      → ConflictCheckPage (results display)
```

---

# 5. MODULE 2: ENGAGEMENT – QUẢN LÝ HỢP ĐỒNG & DỰ ÁN

## 5.1 Mục tiêu

- Quản lý toàn bộ lifecycle: Chào phí → Ký HĐ → Thực hiện → Thanh lý
- Phân loại theo dịch vụ: Kiểm toán TC, KT XDCB, Tư vấn
- Theo dõi chi phí trực tiếp thực tế vs dự toán
- Quản lý kỳ tài chính khác nhau (31/12, 30/6...)
- Phân công nhân sự: TGĐ/Phó TGĐ phân công HQ, GĐ CN phân công chi nhánh

## 5.2 Domain Entities

```go
// internal/engagement/domain/entity.go

type Engagement struct {
    ID               uuid.UUID          `json:"id" db:"id"`
    Code             string             `json:"code" db:"code"`              // Auto: ENG-2026-0001
    ClientID         uuid.UUID          `json:"client_id" db:"client_id"`
    ServiceType      ServiceType        `json:"service_type" db:"service_type"`
    Title            string             `json:"title" db:"title"`
    Description      string             `json:"description" db:"description"`
    FiscalYearEnd    string             `json:"fiscal_year_end" db:"fiscal_year_end"` // "2025-12-31"
    AuditPeriodStart *time.Time         `json:"audit_period_start" db:"audit_period_start"`
    AuditPeriodEnd   *time.Time         `json:"audit_period_end" db:"audit_period_end"`
    Status           EngagementStatus   `json:"status" db:"status"`
    Priority         string             `json:"priority" db:"priority"` // low, medium, high, critical

    // Fee structure
    FeeType          FeeType            `json:"fee_type" db:"fee_type"`       // fixed, time_material, retainer, success
    FeeAmount        int64              `json:"fee_amount" db:"fee_amount"`   // VND
    EstimatedHours   float64            `json:"estimated_hours" db:"estimated_hours"`
    ActualCost       int64              `json:"actual_cost" db:"actual_cost"` // Chi phí trực tiếp thực tế

    // Assignment
    PartnerID        uuid.UUID          `json:"partner_id" db:"partner_id"`         // Partner phụ trách
    ManagerID        *uuid.UUID         `json:"manager_id" db:"manager_id"`         // Manager phụ trách
    LeadAuditorID    *uuid.UUID         `json:"lead_auditor_id" db:"lead_auditor_id"` // KTV chính
    AssignedBy       uuid.UUID          `json:"assigned_by" db:"assigned_by"`       // TGĐ/Phó TGĐ/GĐ CN phân công
    BranchID         uuid.UUID          `json:"branch_id" db:"branch_id"`
    DepartmentID     uuid.UUID          `json:"department_id" db:"department_id"`

    // Contract info
    ContractNumber   string             `json:"contract_number" db:"contract_number"`
    ContractDate     *time.Time         `json:"contract_date" db:"contract_date"`
    ContractSignedBy *uuid.UUID         `json:"contract_signed_by" db:"contract_signed_by"` // Chủ tịch/TGĐ/Phó TGĐ
    ContractFileID   *uuid.UUID         `json:"contract_file_id" db:"contract_file_id"`

    // Timeline
    PlannedStartDate *time.Time         `json:"planned_start_date" db:"planned_start_date"`
    PlannedEndDate   *time.Time         `json:"planned_end_date" db:"planned_end_date"`
    ActualStartDate  *time.Time         `json:"actual_start_date" db:"actual_start_date"`
    ActualEndDate    *time.Time         `json:"actual_end_date" db:"actual_end_date"`

    // Settlement (Thanh lý)
    SettlementDate   *time.Time         `json:"settlement_date" db:"settlement_date"`
    SettlementFileID *uuid.UUID         `json:"settlement_file_id" db:"settlement_file_id"`

    CreatedBy        uuid.UUID          `json:"created_by" db:"created_by"`
    CreatedAt        time.Time          `json:"created_at" db:"created_at"`
    UpdatedAt        time.Time          `json:"updated_at" db:"updated_at"`
}

type FeeType string
const (
    FeeTypeFixed        FeeType = "fixed"          // Phí cố định
    FeeTypeTimeMaterial FeeType = "time_material"  // Theo giờ công
    FeeTypeRetainer     FeeType = "retainer"       // Phí duy trì
    FeeTypeSuccess      FeeType = "success"        // Phí thành công
)

// Engagement Team Member
type EngagementMember struct {
    ID            uuid.UUID     `json:"id" db:"id"`
    EngagementID  uuid.UUID     `json:"engagement_id" db:"engagement_id"`
    EmployeeID    uuid.UUID     `json:"employee_id" db:"employee_id"`
    Role          string        `json:"role" db:"role"`         // "partner", "manager", "senior", "junior", "intern"
    PlannedHours  float64       `json:"planned_hours" db:"planned_hours"`
    ActualHours   float64       `json:"actual_hours" db:"actual_hours"` // Calculated from timesheet
    HourlyRate    int64         `json:"hourly_rate" db:"hourly_rate"`   // VND/hour
    StartDate     *time.Time    `json:"start_date" db:"start_date"`
    EndDate       *time.Time    `json:"end_date" db:"end_date"`
    Status        string        `json:"status" db:"status"`      // active, completed, removed
    AssignedBy    uuid.UUID     `json:"assigned_by" db:"assigned_by"`
    CreatedAt     time.Time     `json:"created_at" db:"created_at"`
}

// Engagement Phase (Phase → Task → Sub-task)
type EngagementPhase struct {
    ID            uuid.UUID `json:"id" db:"id"`
    EngagementID  uuid.UUID `json:"engagement_id" db:"engagement_id"`
    Name          string    `json:"name" db:"name"`          // "Planning", "Fieldwork", "Reporting"
    Order         int       `json:"order" db:"order"`
    Status        string    `json:"status" db:"status"`      // not_started, in_progress, completed
    PlannedStart  *time.Time `json:"planned_start" db:"planned_start"`
    PlannedEnd    *time.Time `json:"planned_end" db:"planned_end"`
    ActualStart   *time.Time `json:"actual_start" db:"actual_start"`
    ActualEnd     *time.Time `json:"actual_end" db:"actual_end"`
    Progress      int       `json:"progress" db:"progress"`  // 0-100%
}

type EngagementTask struct {
    ID            uuid.UUID  `json:"id" db:"id"`
    PhaseID       uuid.UUID  `json:"phase_id" db:"phase_id"`
    EngagementID  uuid.UUID  `json:"engagement_id" db:"engagement_id"`
    ParentTaskID  *uuid.UUID `json:"parent_task_id" db:"parent_task_id"` // For sub-tasks
    Title         string     `json:"title" db:"title"`
    Description   string     `json:"description" db:"description"`
    AssigneeID    *uuid.UUID `json:"assignee_id" db:"assignee_id"`
    Status        TaskStatus `json:"status" db:"status"`
    Priority      string     `json:"priority" db:"priority"`
    DueDate       *time.Time `json:"due_date" db:"due_date"`
    EstimatedHours float64   `json:"estimated_hours" db:"estimated_hours"`
    ActualHours   float64    `json:"actual_hours" db:"actual_hours"`
    Order         int        `json:"order" db:"order"`
    CreatedAt     time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type TaskStatus string
const (
    TaskStatusTodo       TaskStatus = "todo"
    TaskStatusInProgress TaskStatus = "in_progress"
    TaskStatusReview     TaskStatus = "review"
    TaskStatusDone       TaskStatus = "done"
    TaskStatusBlocked    TaskStatus = "blocked"
)

// Direct Cost tracking (chi phí trực tiếp)
type EngagementCost struct {
    ID            uuid.UUID `json:"id" db:"id"`
    EngagementID  uuid.UUID `json:"engagement_id" db:"engagement_id"`
    Category      string    `json:"category" db:"category"`     // "travel", "accommodation", "printing", "other"
    Description   string    `json:"description" db:"description"`
    Amount        int64     `json:"amount" db:"amount"`         // VND
    Date          time.Time `json:"date" db:"date"`
    ReceiptFileID *uuid.UUID `json:"receipt_file_id" db:"receipt_file_id"`
    CreatedBy     uuid.UUID `json:"created_by" db:"created_by"`
    ApprovedBy    *uuid.UUID `json:"approved_by" db:"approved_by"`
    Status        string    `json:"status" db:"status"`          // pending, approved, rejected
    CreatedAt     time.Time `json:"created_at" db:"created_at"`
}
```

## 5.3 Service Interface

```go
type EngagementService interface {
    // CRUD
    CreateEngagement(ctx context.Context, req CreateEngagementRequest) (*Engagement, error)
    UpdateEngagement(ctx context.Context, id uuid.UUID, req UpdateEngagementRequest) (*Engagement, error)
    GetEngagement(ctx context.Context, id uuid.UUID) (*EngagementDetailResponse, error)
    ListEngagements(ctx context.Context, filter EngagementFilter, pagination PaginationParams) (*PaginatedResult[EngagementListItem], error)

    // Status transitions
    SubmitProposal(ctx context.Context, id uuid.UUID) error        // draft → proposal
    SignContract(ctx context.Context, id uuid.UUID, req SignContractRequest) error  // proposal → contracted
    StartEngagement(ctx context.Context, id uuid.UUID) error       // contracted → in_progress
    SubmitForReview(ctx context.Context, id uuid.UUID) error       // in_progress → review
    CompleteEngagement(ctx context.Context, id uuid.UUID) error    // review → completed
    SettleEngagement(ctx context.Context, id uuid.UUID, req SettlementRequest) error // completed → settled

    // Team management
    AssignMember(ctx context.Context, engID uuid.UUID, req AssignMemberRequest) (*EngagementMember, error)
    RemoveMember(ctx context.Context, engID uuid.UUID, memberID uuid.UUID) error
    GetTeam(ctx context.Context, engID uuid.UUID) ([]EngagementMemberDetail, error)

    // Phase & Task management
    CreatePhase(ctx context.Context, engID uuid.UUID, req CreatePhaseRequest) (*EngagementPhase, error)
    UpdatePhaseStatus(ctx context.Context, phaseID uuid.UUID, status string) error
    CreateTask(ctx context.Context, req CreateTaskRequest) (*EngagementTask, error)
    UpdateTask(ctx context.Context, taskID uuid.UUID, req UpdateTaskRequest) (*EngagementTask, error)
    UpdateTaskStatus(ctx context.Context, taskID uuid.UUID, status TaskStatus) error
    GetTasksByEngagement(ctx context.Context, engID uuid.UUID) ([]EngagementTask, error)
    GetTasksByAssignee(ctx context.Context, employeeID uuid.UUID) ([]EngagementTask, error)

    // Cost tracking
    AddCost(ctx context.Context, engID uuid.UUID, req CreateCostRequest) (*EngagementCost, error)
    ApproveCost(ctx context.Context, costID uuid.UUID) error
    GetCostSummary(ctx context.Context, engID uuid.UUID) (*CostSummary, error)

    // Budget analysis
    GetBudgetVsActual(ctx context.Context, engID uuid.UUID) (*BudgetAnalysis, error)

    // Schedule conflict detection
    CheckScheduleConflict(ctx context.Context, employeeID uuid.UUID, startDate, endDate time.Time) ([]Engagement, error)

    // Export
    ExportEngagements(ctx context.Context, filter EngagementFilter, format string) ([]byte, error)
}
```

## 5.4 API Endpoints

```
POST   /api/v1/engagements                              # Tạo engagement
GET    /api/v1/engagements                              # Danh sách (filtered)
GET    /api/v1/engagements/:id                          # Chi tiết
PUT    /api/v1/engagements/:id                          # Cập nhật

# Status transitions
POST   /api/v1/engagements/:id/submit-proposal
POST   /api/v1/engagements/:id/sign-contract
POST   /api/v1/engagements/:id/start
POST   /api/v1/engagements/:id/submit-review
POST   /api/v1/engagements/:id/complete
POST   /api/v1/engagements/:id/settle

# Team
POST   /api/v1/engagements/:id/members
DELETE /api/v1/engagements/:id/members/:memberId
GET    /api/v1/engagements/:id/members

# Phases & Tasks
POST   /api/v1/engagements/:id/phases
PUT    /api/v1/engagements/:id/phases/:phaseId
POST   /api/v1/engagements/:id/tasks
PUT    /api/v1/tasks/:taskId
PATCH  /api/v1/tasks/:taskId/status

# Costs
POST   /api/v1/engagements/:id/costs
POST   /api/v1/costs/:costId/approve
GET    /api/v1/engagements/:id/costs/summary
GET    /api/v1/engagements/:id/budget-analysis

# My work
GET    /api/v1/my/engagements                            # Engagement của tôi
GET    /api/v1/my/tasks                                  # Task được assign cho tôi
GET    /api/v1/employees/:id/schedule-conflicts          # Kiểm tra xung đột lịch
```

## 5.5 Frontend Pages

```
/engagements                       → EngagementListPage (DataTable, Kanban view toggle)
/engagements/new                   → CreateEngagementPage
/engagements/[id]                  → EngagementDetailPage
    Tab: Overview                  → Thông tin chung, status timeline, budget summary
    Tab: Team                      → Team members, assignment
    Tab: Tasks                     → Phase/Task tree (Gantt-like view)
    Tab: Costs                     → Chi phí trực tiếp
    Tab: Documents                 → Hồ sơ liên quan
    Tab: Billing                   → Hóa đơn, thanh toán
    Tab: History                   → Audit trail
/engagements/[id]/edit             → EditEngagementPage
/my/tasks                          → MyTasksPage (personal task board)
```

---

# 6. MODULE 3: TIMESHEET & RESOURCE

## 6.1 Mục tiêu

- Chấm công theo ngày để tính lương (attendance)
- Theo dõi giờ thực hiện engagement để tính KPI
- Phân công nhân sự: TGĐ/Phó TGĐ phân công HQ, GĐ CN phân công chi nhánh
- Cảnh báo xung đột lịch (1 nhân viên assign nhiều dự án cùng lúc)
- Tính utilization rate (tương lai)

## 6.2 Domain Entities

```go
// internal/timesheet/domain/entity.go

// Attendance (Chấm công hàng ngày)
type Attendance struct {
    ID          uuid.UUID `json:"id" db:"id"`
    EmployeeID  uuid.UUID `json:"employee_id" db:"employee_id"`
    Date        time.Time `json:"date" db:"date"`
    CheckIn     *time.Time `json:"check_in" db:"check_in"`
    CheckOut    *time.Time `json:"check_out" db:"check_out"`
    Status      AttendanceStatus `json:"status" db:"status"` // present, absent, leave, holiday, wfh
    LeaveType   *string   `json:"leave_type" db:"leave_type"` // annual, sick, unpaid, maternity
    OTHours     float64   `json:"ot_hours" db:"ot_hours"`
    Notes       string    `json:"notes" db:"notes"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type AttendanceStatus string
const (
    AttPresent  AttendanceStatus = "present"
    AttAbsent   AttendanceStatus = "absent"
    AttLeave    AttendanceStatus = "leave"
    AttHoliday  AttendanceStatus = "holiday"
    AttWFH      AttendanceStatus = "wfh"
)

// Timesheet Entry (Giờ thực hiện engagement)
type TimesheetEntry struct {
    ID            uuid.UUID `json:"id" db:"id"`
    EmployeeID    uuid.UUID `json:"employee_id" db:"employee_id"`
    EngagementID  uuid.UUID `json:"engagement_id" db:"engagement_id"`
    TaskID        *uuid.UUID `json:"task_id" db:"task_id"`
    Date          time.Time `json:"date" db:"date"`
    Hours         float64   `json:"hours" db:"hours"` // 0.5 – 24
    Description   string    `json:"description" db:"description"`
    Status        TSStatus  `json:"status" db:"status"` // draft, submitted, approved, rejected
    ApprovedBy    *uuid.UUID `json:"approved_by" db:"approved_by"`
    ApprovedAt    *time.Time `json:"approved_at" db:"approved_at"`
    CreatedAt     time.Time `json:"created_at" db:"created_at"`
    UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type TSStatus string
const (
    TSDraft     TSStatus = "draft"
    TSSubmitted TSStatus = "submitted"
    TSApproved  TSStatus = "approved"
    TSRejected  TSStatus = "rejected"
)

// Resource Allocation (kế hoạch phân bổ nhân sự)
type ResourceAllocation struct {
    ID            uuid.UUID `json:"id" db:"id"`
    EmployeeID    uuid.UUID `json:"employee_id" db:"employee_id"`
    EngagementID  uuid.UUID `json:"engagement_id" db:"engagement_id"`
    StartDate     time.Time `json:"start_date" db:"start_date"`
    EndDate       time.Time `json:"end_date" db:"end_date"`
    Percentage    int       `json:"percentage" db:"percentage"`       // 0-100% allocation
    PlannedHours  float64   `json:"planned_hours" db:"planned_hours"`
    AllocatedBy   uuid.UUID `json:"allocated_by" db:"allocated_by"`  // TGĐ/Phó TGĐ/GĐ CN
    Status        string    `json:"status" db:"status"`               // planned, confirmed, completed
    CreatedAt     time.Time `json:"created_at" db:"created_at"`
}
```

## 6.3 Service Interface

```go
type TimesheetService interface {
    // Attendance
    CheckIn(ctx context.Context, employeeID uuid.UUID) (*Attendance, error)
    CheckOut(ctx context.Context, employeeID uuid.UUID) (*Attendance, error)
    RecordAttendance(ctx context.Context, req RecordAttendanceRequest) (*Attendance, error)
    GetMonthlyAttendance(ctx context.Context, employeeID uuid.UUID, year, month int) ([]Attendance, error)
    GetTeamAttendance(ctx context.Context, departmentID uuid.UUID, date time.Time) ([]Attendance, error)

    // Timesheet
    CreateEntry(ctx context.Context, req CreateTimesheetRequest) (*TimesheetEntry, error)
    UpdateEntry(ctx context.Context, id uuid.UUID, req UpdateTimesheetRequest) (*TimesheetEntry, error)
    DeleteEntry(ctx context.Context, id uuid.UUID) error
    SubmitWeekly(ctx context.Context, employeeID uuid.UUID, weekStart time.Time) error
    ApproveTimesheet(ctx context.Context, entryIDs []uuid.UUID, approverID uuid.UUID) error
    RejectTimesheet(ctx context.Context, entryIDs []uuid.UUID, reason string) error
    GetMyTimesheet(ctx context.Context, employeeID uuid.UUID, from, to time.Time) ([]TimesheetEntry, error)
    GetTimesheetByEngagement(ctx context.Context, engID uuid.UUID) ([]TimesheetEntry, error)
    GetPendingApprovals(ctx context.Context, managerID uuid.UUID) ([]TimesheetEntry, error)

    // Resource Allocation
    AllocateResource(ctx context.Context, req AllocateResourceRequest) (*ResourceAllocation, error)
    GetEmployeeAllocations(ctx context.Context, employeeID uuid.UUID, from, to time.Time) ([]ResourceAllocation, error)
    GetEngagementAllocations(ctx context.Context, engID uuid.UUID) ([]ResourceAllocation, error)
    CheckConflicts(ctx context.Context, employeeID uuid.UUID, from, to time.Time) ([]AllocationConflict, error)
    GetAvailableResources(ctx context.Context, from, to time.Time, departmentID *uuid.UUID) ([]ResourceAvailability, error)

    // Analytics
    GetUtilizationRate(ctx context.Context, employeeID uuid.UUID, from, to time.Time) (*UtilizationReport, error)
    GetTeamUtilization(ctx context.Context, departmentID uuid.UUID, from, to time.Time) ([]UtilizationReport, error)
}

type AllocationConflict struct {
    EmployeeID    uuid.UUID    `json:"employee_id"`
    EmployeeName  string       `json:"employee_name"`
    Engagements   []Engagement `json:"engagements"` // Overlapping engagements
    OverlapStart  time.Time    `json:"overlap_start"`
    OverlapEnd    time.Time    `json:"overlap_end"`
    TotalPercent  int          `json:"total_percent"` // >100% = overallocated
}
```

## 6.4 API Endpoints

```
# Attendance
POST   /api/v1/attendance/check-in
POST   /api/v1/attendance/check-out
POST   /api/v1/attendance
GET    /api/v1/attendance/monthly?employee_id=&year=&month=
GET    /api/v1/attendance/team?department_id=&date=

# Timesheet
POST   /api/v1/timesheets
PUT    /api/v1/timesheets/:id
DELETE /api/v1/timesheets/:id
POST   /api/v1/timesheets/submit-weekly
POST   /api/v1/timesheets/approve
POST   /api/v1/timesheets/reject
GET    /api/v1/my/timesheets?from=&to=
GET    /api/v1/timesheets/pending-approvals

# Resource Allocation
POST   /api/v1/resource-allocations
GET    /api/v1/employees/:id/allocations?from=&to=
GET    /api/v1/engagements/:id/allocations
GET    /api/v1/resource-allocations/conflicts?employee_id=&from=&to=
GET    /api/v1/resource-allocations/available?from=&to=&department_id=

# Analytics
GET    /api/v1/utilization/:employeeId?from=&to=
GET    /api/v1/utilization/team/:departmentId?from=&to=
```

---

# 7. MODULE 4: BILLING & INVOICING

## 7.1 Mục tiêu

- Hỗ trợ nhiều mô hình tính phí: Fixed, Time & Material, Retainer, Success fee
- Billing theo milestones (tiến độ) với gợi ý tự động
- Xuất hóa đơn nháp → KH xác nhận → Xuất hóa đơn điện tử
- Theo dõi công nợ phải thu (AR) và nhắc nhở tự động
- Quản lý credit note, điều chỉnh hóa đơn
- Đơn tệ VND

## 7.2 Domain Entities

```go
// internal/billing/domain/entity.go

type Invoice struct {
    ID              uuid.UUID      `json:"id" db:"id"`
    Code            string         `json:"code" db:"code"`            // INV-2026-0001
    EngagementID    uuid.UUID      `json:"engagement_id" db:"engagement_id"`
    ClientID        uuid.UUID      `json:"client_id" db:"client_id"`
    Type            InvoiceType    `json:"type" db:"type"`            // regular, credit_note, adjustment
    Status          InvoiceStatus  `json:"status" db:"status"`
    
    // Amounts
    SubTotal        int64          `json:"sub_total" db:"sub_total"`       // VND trước thuế
    VATRate         float64        `json:"vat_rate" db:"vat_rate"`         // 0.08, 0.10
    VATAmount       int64          `json:"vat_amount" db:"vat_amount"`
    TotalAmount     int64          `json:"total_amount" db:"total_amount"` // Sau thuế
    PaidAmount      int64          `json:"paid_amount" db:"paid_amount"`
    RemainingAmount int64          `json:"remaining_amount" db:"remaining_amount"`

    // Milestone billing
    MilestoneID     *uuid.UUID     `json:"milestone_id" db:"milestone_id"`
    MilestoneDesc   string         `json:"milestone_desc" db:"milestone_desc"`
    MilestonePercent float64       `json:"milestone_percent" db:"milestone_percent"` // % of total fee

    // Dates
    IssueDate       time.Time      `json:"issue_date" db:"issue_date"`
    DueDate         time.Time      `json:"due_date" db:"due_date"`
    SentToClientAt  *time.Time     `json:"sent_to_client_at" db:"sent_to_client_at"`
    ClientConfirmedAt *time.Time   `json:"client_confirmed_at" db:"client_confirmed_at"`

    // E-invoice (hóa đơn điện tử)
    EInvoiceNumber  string         `json:"einvoice_number" db:"einvoice_number"`
    EInvoiceDate    *time.Time     `json:"einvoice_date" db:"einvoice_date"`
    EInvoiceStatus  string         `json:"einvoice_status" db:"einvoice_status"` // draft, issued, cancelled

    // Refs
    ContractRef     string         `json:"contract_ref" db:"contract_ref"`
    Notes           string         `json:"notes" db:"notes"`
    CreatedBy       uuid.UUID      `json:"created_by" db:"created_by"`
    ApprovedBy      *uuid.UUID     `json:"approved_by" db:"approved_by"`
    CreatedAt       time.Time      `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
}

type InvoiceType string
const (
    InvTypeRegular    InvoiceType = "regular"
    InvTypeCreditNote InvoiceType = "credit_note"
    InvTypeAdjustment InvoiceType = "adjustment"
)

type InvoiceStatus string
const (
    InvStatusDraft         InvoiceStatus = "draft"           // Nháp
    InvStatusSentToClient  InvoiceStatus = "sent_to_client"  // Gửi KH xem
    InvStatusClientConfirmed InvoiceStatus = "client_confirmed" // KH đồng ý
    InvStatusIssued        InvoiceStatus = "issued"          // Đã xuất HĐĐT
    InvStatusPartialPaid   InvoiceStatus = "partial_paid"    // Thanh toán một phần
    InvStatusPaid          InvoiceStatus = "paid"            // Đã thanh toán
    InvStatusOverdue       InvoiceStatus = "overdue"         // Quá hạn
    InvStatusCancelled     InvoiceStatus = "cancelled"
)

type InvoiceLineItem struct {
    ID          uuid.UUID `json:"id" db:"id"`
    InvoiceID   uuid.UUID `json:"invoice_id" db:"invoice_id"`
    Description string    `json:"description" db:"description"`
    Quantity    float64   `json:"quantity" db:"quantity"`
    UnitPrice   int64     `json:"unit_price" db:"unit_price"`
    Amount      int64     `json:"amount" db:"amount"`
    Order       int       `json:"order" db:"order"`
}

// Payment (Thanh toán)
type Payment struct {
    ID          uuid.UUID     `json:"id" db:"id"`
    InvoiceID   uuid.UUID     `json:"invoice_id" db:"invoice_id"`
    Amount      int64         `json:"amount" db:"amount"` // VND
    PaymentDate time.Time     `json:"payment_date" db:"payment_date"`
    Method      string        `json:"method" db:"method"` // bank_transfer, cash, other
    Reference   string        `json:"reference" db:"reference"` // Số chứng từ
    Notes       string        `json:"notes" db:"notes"`
    RecordedBy  uuid.UUID     `json:"recorded_by" db:"recorded_by"`
    CreatedAt   time.Time     `json:"created_at" db:"created_at"`
}

// Billing Milestone
type BillingMilestone struct {
    ID            uuid.UUID  `json:"id" db:"id"`
    EngagementID  uuid.UUID  `json:"engagement_id" db:"engagement_id"`
    Name          string     `json:"name" db:"name"`          // "Tạm ứng 50%", "Hoàn thành báo cáo"
    Percentage    float64    `json:"percentage" db:"percentage"` // % of total fee
    Amount        int64      `json:"amount" db:"amount"`
    TriggerPhase  *uuid.UUID `json:"trigger_phase" db:"trigger_phase"` // Auto-suggest khi phase hoàn thành
    DueDate       *time.Time `json:"due_date" db:"due_date"`
    InvoiceID     *uuid.UUID `json:"invoice_id" db:"invoice_id"` // Link to generated invoice
    Status        string     `json:"status" db:"status"` // pending, ready, invoiced, paid
    CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// AR Aging (Phân tích tuổi nợ)
type ARAgingReport struct {
    ClientID       uuid.UUID `json:"client_id"`
    ClientName     string    `json:"client_name"`
    Current        int64     `json:"current"`          // Chưa đến hạn
    Days1to30      int64     `json:"days_1_to_30"`     // Quá hạn 1-30 ngày
    Days31to60     int64     `json:"days_31_to_60"`
    Days61to90     int64     `json:"days_61_to_90"`
    Over90Days     int64     `json:"over_90_days"`
    TotalOverdue   int64     `json:"total_overdue"`
}
```

## 7.3 Service Interface

```go
type BillingService interface {
    // Invoice
    CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*Invoice, error)
    UpdateInvoice(ctx context.Context, id uuid.UUID, req UpdateInvoiceRequest) (*Invoice, error)
    GetInvoice(ctx context.Context, id uuid.UUID) (*InvoiceDetailResponse, error)
    ListInvoices(ctx context.Context, filter InvoiceFilter, pagination PaginationParams) (*PaginatedResult[InvoiceListItem], error)

    // Invoice workflow
    SendToClient(ctx context.Context, id uuid.UUID) error
    ConfirmByClient(ctx context.Context, id uuid.UUID) error
    IssueEInvoice(ctx context.Context, id uuid.UUID) error
    CancelInvoice(ctx context.Context, id uuid.UUID, reason string) error
    CreateCreditNote(ctx context.Context, originalInvoiceID uuid.UUID, req CreateCreditNoteRequest) (*Invoice, error)

    // Payment
    RecordPayment(ctx context.Context, req RecordPaymentRequest) (*Payment, error)
    GetPaymentsByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]Payment, error)

    // Milestones
    CreateMilestone(ctx context.Context, req CreateMilestoneRequest) (*BillingMilestone, error)
    GetMilestonesByEngagement(ctx context.Context, engID uuid.UUID) ([]BillingMilestone, error)
    CheckMilestonesReady(ctx context.Context, engID uuid.UUID) ([]BillingMilestone, error) // Auto-suggest

    // AR Management
    GetARSummary(ctx context.Context, filter ARFilter) ([]ARAgingReport, error)
    GetOverdueInvoices(ctx context.Context) ([]Invoice, error)
    SendPaymentReminder(ctx context.Context, invoiceID uuid.UUID) error   // Email nhắc nợ
    GetARByClient(ctx context.Context, clientID uuid.UUID) (*ClientARDetail, error)

    // Revenue
    GetRevenueReport(ctx context.Context, from, to time.Time, groupBy string) (*RevenueReport, error)

    // Export
    ExportInvoicePDF(ctx context.Context, invoiceID uuid.UUID) ([]byte, error)
    ExportARReport(ctx context.Context, filter ARFilter, format string) ([]byte, error)
}
```

## 7.4 API Endpoints

```
# Invoices
POST   /api/v1/invoices
GET    /api/v1/invoices
GET    /api/v1/invoices/:id
PUT    /api/v1/invoices/:id
POST   /api/v1/invoices/:id/send-to-client
POST   /api/v1/invoices/:id/client-confirm
POST   /api/v1/invoices/:id/issue-einvoice
POST   /api/v1/invoices/:id/cancel
POST   /api/v1/invoices/:id/credit-note
GET    /api/v1/invoices/:id/pdf

# Payments
POST   /api/v1/payments
GET    /api/v1/invoices/:id/payments

# Milestones
POST   /api/v1/engagements/:id/billing-milestones
GET    /api/v1/engagements/:id/billing-milestones
GET    /api/v1/engagements/:id/billing-milestones/ready

# AR
GET    /api/v1/ar/summary
GET    /api/v1/ar/overdue
POST   /api/v1/ar/send-reminder/:invoiceId
GET    /api/v1/clients/:id/ar

# Revenue
GET    /api/v1/revenue?from=&to=&group_by=service|partner|branch|department
```

---

# 8. MODULE 5: WORKING PAPERS – QUẢN LÝ HỒ SƠ KIỂM TOÁN

## 8.1 Mục tiêu

- Chuyển từ giấy/file local sang hệ thống tập trung trên cloud
- Chuẩn hóa template theo VSA
- Workflow review & sign-off theo cấp bậc: Auditor → Senior → Manager → Partner
- Phân quyền truy cập theo dự án
- Lưu trữ 10 năm, bảo mật cao
- Versioning cho mọi file

## 8.2 Domain Entities

```go
// internal/workingpaper/domain/entity.go

type WorkingPaperFolder struct {
    ID            uuid.UUID `json:"id" db:"id"`
    EngagementID  uuid.UUID `json:"engagement_id" db:"engagement_id"`
    ParentID      *uuid.UUID `json:"parent_id" db:"parent_id"`
    Name          string    `json:"name" db:"name"`
    Code          string    `json:"code" db:"code"`        // "A1000", "B2000" — mã hồ sơ
    Type          string    `json:"type" db:"type"`         // "planning", "fieldwork", "completion", "permanent"
    Order         int       `json:"order" db:"order"`
    CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type WorkingPaper struct {
    ID              uuid.UUID         `json:"id" db:"id"`
    EngagementID    uuid.UUID         `json:"engagement_id" db:"engagement_id"`
    FolderID        uuid.UUID         `json:"folder_id" db:"folder_id"`
    TemplateID      *uuid.UUID        `json:"template_id" db:"template_id"`
    Code            string            `json:"code" db:"code"`              // "A1010" — mã giấy tờ
    Title           string            `json:"title" db:"title"`
    Description     string            `json:"description" db:"description"`
    FileID          *uuid.UUID        `json:"file_id" db:"file_id"`        // Link to storage
    Status          WPStatus          `json:"status" db:"status"`
    
    // Sign-off chain
    PreparedBy      *uuid.UUID        `json:"prepared_by" db:"prepared_by"`
    PreparedAt      *time.Time        `json:"prepared_at" db:"prepared_at"`
    ReviewedBy      *uuid.UUID        `json:"reviewed_by" db:"reviewed_by"`     // Senior
    ReviewedAt      *time.Time        `json:"reviewed_at" db:"reviewed_at"`
    ManagerApproval *uuid.UUID        `json:"manager_approval" db:"manager_approval"` // Manager
    ManagerApprovedAt *time.Time      `json:"manager_approved_at" db:"manager_approved_at"`
    PartnerSignOff  *uuid.UUID        `json:"partner_sign_off" db:"partner_sign_off"` // Partner
    PartnerSignedAt *time.Time        `json:"partner_signed_at" db:"partner_signed_at"`

    // Review notes
    ReviewNotes     string            `json:"review_notes" db:"review_notes"`
    
    Version         int               `json:"version" db:"version"`
    CreatedAt       time.Time         `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time         `json:"updated_at" db:"updated_at"`
}

type WPStatus string
const (
    WPStatusDraft          WPStatus = "draft"
    WPStatusInProgress     WPStatus = "in_progress"
    WPStatusPrepared       WPStatus = "prepared"         // Auditor hoàn thành
    WPStatusSeniorReviewed WPStatus = "senior_reviewed"  // Senior review xong
    WPStatusManagerApproved WPStatus = "manager_approved" // Manager duyệt
    WPStatusPartnerSignedOff WPStatus = "partner_signed_off" // Partner sign-off
    WPStatusFinalized      WPStatus = "finalized"        // Hoàn tất, lock
)

// Audit Template (mẫu chuẩn hóa theo VSA)
type AuditTemplate struct {
    ID          uuid.UUID `json:"id" db:"id"`
    Code        string    `json:"code" db:"code"`          // "VSA-200", "VSA-315"
    Name        string    `json:"name" db:"name"`
    Category    string    `json:"category" db:"category"`  // "planning", "risk_assessment", "substantive"
    ServiceType ServiceType `json:"service_type" db:"service_type"`
    FileID      uuid.UUID `json:"file_id" db:"file_id"`
    Version     string    `json:"version" db:"version"`
    IsActive    bool      `json:"is_active" db:"is_active"`
    CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Review Comment (nhận xét khi review)
type ReviewComment struct {
    ID              uuid.UUID `json:"id" db:"id"`
    WorkingPaperID  uuid.UUID `json:"working_paper_id" db:"working_paper_id"`
    AuthorID        uuid.UUID `json:"author_id" db:"author_id"`
    Content         string    `json:"content" db:"content"`
    Status          string    `json:"status" db:"status"` // open, resolved, wont_fix
    ResolvedBy      *uuid.UUID `json:"resolved_by" db:"resolved_by"`
    ResolvedAt      *time.Time `json:"resolved_at" db:"resolved_at"`
    CreatedAt       time.Time `json:"created_at" db:"created_at"`
}
```

## 8.3 Service Interface

```go
type WorkingPaperService interface {
    // Folder structure
    CreateFolder(ctx context.Context, req CreateFolderRequest) (*WorkingPaperFolder, error)
    GetFolderTree(ctx context.Context, engID uuid.UUID) ([]WorkingPaperFolderNode, error)

    // Working Papers
    CreateWorkingPaper(ctx context.Context, req CreateWPRequest) (*WorkingPaper, error)
    UpdateWorkingPaper(ctx context.Context, id uuid.UUID, req UpdateWPRequest) (*WorkingPaper, error)
    GetWorkingPaper(ctx context.Context, id uuid.UUID) (*WorkingPaperDetail, error)
    ListByEngagement(ctx context.Context, engID uuid.UUID, filter WPFilter) ([]WorkingPaper, error)
    UploadFile(ctx context.Context, wpID uuid.UUID, file io.Reader, fileName string) (*FileMetadata, error)
    GetVersionHistory(ctx context.Context, wpID uuid.UUID) ([]WorkingPaper, error)

    // Sign-off Workflow
    MarkAsPrepared(ctx context.Context, wpID uuid.UUID) error          // Auditor: draft → prepared
    SeniorReview(ctx context.Context, wpID uuid.UUID, approved bool, notes string) error
    ManagerApprove(ctx context.Context, wpID uuid.UUID, approved bool, notes string) error
    PartnerSignOff(ctx context.Context, wpID uuid.UUID, approved bool, notes string) error
    FinalizeWorkingPaper(ctx context.Context, wpID uuid.UUID) error    // Lock, no more changes
    GetPendingReviews(ctx context.Context, reviewerID uuid.UUID) ([]WorkingPaper, error)

    // Review Comments
    AddComment(ctx context.Context, wpID uuid.UUID, req AddCommentRequest) (*ReviewComment, error)
    ResolveComment(ctx context.Context, commentID uuid.UUID) error
    GetComments(ctx context.Context, wpID uuid.UUID) ([]ReviewComment, error)
    GetOpenComments(ctx context.Context, engID uuid.UUID) ([]ReviewComment, error)

    // Templates
    CreateTemplate(ctx context.Context, req CreateTemplateRequest) (*AuditTemplate, error)
    ListTemplates(ctx context.Context, serviceType *ServiceType) ([]AuditTemplate, error)
    InitializeFromTemplate(ctx context.Context, engID uuid.UUID, templateIDs []uuid.UUID) error // Tạo WP từ template

    // Engagement-level status
    GetEngagementWPSummary(ctx context.Context, engID uuid.UUID) (*WPSummary, error)

    // Access control (phân quyền theo dự án)
    CheckAccess(ctx context.Context, userID, wpID uuid.UUID) (bool, error)
}
```

## 8.4 API Endpoints

```
# Folders
POST   /api/v1/engagements/:id/wp-folders
GET    /api/v1/engagements/:id/wp-folders/tree

# Working Papers
POST   /api/v1/working-papers
PUT    /api/v1/working-papers/:id
GET    /api/v1/working-papers/:id
GET    /api/v1/engagements/:id/working-papers
POST   /api/v1/working-papers/:id/upload
GET    /api/v1/working-papers/:id/versions
GET    /api/v1/working-papers/:id/download

# Sign-off
POST   /api/v1/working-papers/:id/mark-prepared
POST   /api/v1/working-papers/:id/senior-review
POST   /api/v1/working-papers/:id/manager-approve
POST   /api/v1/working-papers/:id/partner-signoff
POST   /api/v1/working-papers/:id/finalize
GET    /api/v1/my/pending-reviews

# Comments
POST   /api/v1/working-papers/:id/comments
GET    /api/v1/working-papers/:id/comments
POST   /api/v1/comments/:id/resolve
GET    /api/v1/engagements/:id/open-comments

# Templates
POST   /api/v1/audit-templates
GET    /api/v1/audit-templates
POST   /api/v1/engagements/:id/initialize-from-templates

# Summary
GET    /api/v1/engagements/:id/wp-summary
```

---

# 9. MODULE 6: TAX & ADVISORY – QUẢN LÝ THUẾ & TƯ VẤN TÀI CHÍNH

## 9.1 Mục tiêu

- Theo dõi deadline nộp tờ khai thuế từng KH, cảnh báo tự động
- Quản lý các loại dịch vụ thuế: Kê khai, quyết toán, tax planning, hỗ trợ thanh tra
- Lưu trữ lịch sử tư vấn và khuyến nghị đã đưa ra
- Quản lý deliverables (báo cáo, mô hình tài chính)

## 9.2 Domain Entities

```go
// internal/tax/domain/entity.go

// Tax Filing Deadline tracking
type TaxDeadline struct {
    ID              uuid.UUID       `json:"id" db:"id"`
    ClientID        uuid.UUID       `json:"client_id" db:"client_id"`
    EngagementID    *uuid.UUID      `json:"engagement_id" db:"engagement_id"`
    TaxType         TaxType         `json:"tax_type" db:"tax_type"`
    Period          string          `json:"period" db:"period"`          // "2026-Q1", "2025-FY"
    DueDate         time.Time       `json:"due_date" db:"due_date"`
    ExtendedDueDate *time.Time      `json:"extended_due_date" db:"extended_due_date"`
    Status          TaxDeadlineStatus `json:"status" db:"status"`
    AssigneeID      *uuid.UUID      `json:"assignee_id" db:"assignee_id"`
    SubmittedDate   *time.Time      `json:"submitted_date" db:"submitted_date"`
    Notes           string          `json:"notes" db:"notes"`
    ReminderSent    bool            `json:"reminder_sent" db:"reminder_sent"`
    CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

type TaxType string
const (
    TaxVAT          TaxType = "vat"
    TaxCIT          TaxType = "cit"           // Thuế TNDN
    TaxPIT          TaxType = "pit"           // Thuế TNCN
    TaxFCT          TaxType = "fct"           // Thuế nhà thầu
    TaxSpecial      TaxType = "special"       // Thuế tiêu thụ đặc biệt
    TaxOther        TaxType = "other"
)

type TaxDeadlineStatus string
const (
    TDStatusUpcoming   TaxDeadlineStatus = "upcoming"
    TDStatusDueSoon    TaxDeadlineStatus = "due_soon"    // ≤ 7 ngày
    TDStatusOverdue    TaxDeadlineStatus = "overdue"
    TDStatusSubmitted  TaxDeadlineStatus = "submitted"
    TDStatusCompleted  TaxDeadlineStatus = "completed"
)

// Advisory Record (lịch sử tư vấn)
type AdvisoryRecord struct {
    ID            uuid.UUID `json:"id" db:"id"`
    ClientID      uuid.UUID `json:"client_id" db:"client_id"`
    EngagementID  *uuid.UUID `json:"engagement_id" db:"engagement_id"`
    Type          string    `json:"type" db:"type"`         // "tax_advice", "financial_analysis", "tax_planning", "inspection_support"
    Subject       string    `json:"subject" db:"subject"`
    Summary       string    `json:"summary" db:"summary"`
    Recommendations string  `json:"recommendations" db:"recommendations"`
    AdvisorID     uuid.UUID `json:"advisor_id" db:"advisor_id"`
    Date          time.Time `json:"date" db:"date"`
    Deliverables  []uuid.UUID `json:"deliverables"`  // FileMetadata IDs
    CreatedAt     time.Time `json:"created_at" db:"created_at"`
}
```

## 9.3 Service Interface

```go
type TaxAdvisoryService interface {
    // Tax Deadlines
    CreateDeadline(ctx context.Context, req CreateDeadlineRequest) (*TaxDeadline, error)
    UpdateDeadline(ctx context.Context, id uuid.UUID, req UpdateDeadlineRequest) (*TaxDeadline, error)
    ListDeadlines(ctx context.Context, filter TaxDeadlineFilter) ([]TaxDeadline, error)
    GetUpcomingDeadlines(ctx context.Context, days int) ([]TaxDeadline, error)
    MarkAsSubmitted(ctx context.Context, id uuid.UUID) error
    BulkGenerateDeadlines(ctx context.Context, clientID uuid.UUID, year int) ([]TaxDeadline, error) // Auto-gen từ lịch thuế

    // Scheduled Jobs
    CheckAndSendReminders(ctx context.Context) error  // Cron job: check deadlines & send alerts

    // Advisory Records
    CreateAdvisoryRecord(ctx context.Context, req CreateAdvisoryRequest) (*AdvisoryRecord, error)
    ListAdvisoryHistory(ctx context.Context, clientID uuid.UUID) ([]AdvisoryRecord, error)
    GetAdvisoryRecord(ctx context.Context, id uuid.UUID) (*AdvisoryRecord, error)
    AttachDeliverable(ctx context.Context, recordID uuid.UUID, file io.Reader, fileName string) (*FileMetadata, error)

    // Dashboard data
    GetTaxDashboard(ctx context.Context) (*TaxDashboard, error)
}

type TaxDashboard struct {
    OverdueCount    int             `json:"overdue_count"`
    DueSoonCount    int             `json:"due_soon_count"`
    UpcomingCount   int             `json:"upcoming_count"`
    CompletedMonth  int             `json:"completed_this_month"`
    OverdueClients  []TaxDeadline   `json:"overdue_clients"`
    DueSoonItems    []TaxDeadline   `json:"due_soon_items"`
}
```

## 9.4 API Endpoints

```
# Tax Deadlines
POST   /api/v1/tax-deadlines
PUT    /api/v1/tax-deadlines/:id
GET    /api/v1/tax-deadlines?status=&client_id=&from=&to=
GET    /api/v1/tax-deadlines/upcoming?days=30
POST   /api/v1/tax-deadlines/:id/submit
POST   /api/v1/tax-deadlines/bulk-generate

# Advisory Records
POST   /api/v1/advisory-records
GET    /api/v1/advisory-records?client_id=
GET    /api/v1/advisory-records/:id
POST   /api/v1/advisory-records/:id/deliverables

# Dashboard
GET    /api/v1/tax/dashboard
```

---

# 10. MODULE 7: HRM – QUẢN LÝ NHÂN SỰ & NĂNG LỰC

## 10.1 Mục tiêu

- Hồ sơ cá nhân, chứng chỉ, cấp bậc, đào tạo
- Cấu trúc cấp bậc: Intern → Junior → Senior → Manager → Director → Partner
- Theo dõi chứng chỉ chuyên môn + hạn hiệu lực (CPA, ACCA, KTV hành nghề...)
- Đánh giá hiệu suất (KPI) gắn với timesheet & engagement
- Quản lý CPE (Continuing Professional Education) theo yêu cầu Bộ TC/VACPA

## 10.2 Domain Entities

```go
// internal/hrm/domain/entity.go

type Employee struct {
    ID              uuid.UUID     `json:"id" db:"id"`
    Code            string        `json:"code" db:"code"`              // NV-0001
    UserID          *uuid.UUID    `json:"user_id" db:"user_id"`        // Link to auth.User
    FullName        string        `json:"full_name" db:"full_name"`
    DateOfBirth     *time.Time    `json:"date_of_birth" db:"date_of_birth"`
    Gender          string        `json:"gender" db:"gender"`
    IDNumber        string        `json:"id_number" db:"id_number"`    // CCCD
    Phone           string        `json:"phone" db:"phone"`
    PersonalEmail   string        `json:"personal_email" db:"personal_email"`
    WorkEmail       string        `json:"work_email" db:"work_email"`
    Address         string        `json:"address" db:"address"`
    
    // Employment
    Grade           EmployeeGrade `json:"grade" db:"grade"`
    Position        string        `json:"position" db:"position"`      // Chức danh
    BranchID        uuid.UUID     `json:"branch_id" db:"branch_id"`
    DepartmentID    uuid.UUID     `json:"department_id" db:"department_id"`
    DirectManagerID *uuid.UUID    `json:"direct_manager_id" db:"direct_manager_id"`
    JoinDate        time.Time     `json:"join_date" db:"join_date"`
    ContractType    string        `json:"contract_type" db:"contract_type"` // permanent, contract, probation, intern
    Status          string        `json:"status" db:"status"`          // active, resigned, terminated
    
    // Emergency contact
    EmergencyContact     string `json:"emergency_contact" db:"emergency_contact"`
    EmergencyContactPhone string `json:"emergency_contact_phone" db:"emergency_contact_phone"`
    
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Professional Certification
type Certification struct {
    ID              uuid.UUID `json:"id" db:"id"`
    EmployeeID      uuid.UUID `json:"employee_id" db:"employee_id"`
    Type            string    `json:"type" db:"type"`          // "cpa", "acca", "cfa", "ktv_hanh_nghe", "other"
    Name            string    `json:"name" db:"name"`
    IssuedBy        string    `json:"issued_by" db:"issued_by"`
    IssueDate       time.Time `json:"issue_date" db:"issue_date"`
    ExpiryDate      *time.Time `json:"expiry_date" db:"expiry_date"`
    CertNumber      string    `json:"cert_number" db:"cert_number"`
    Status          string    `json:"status" db:"status"`       // active, expired, revoked
    FileID          *uuid.UUID `json:"file_id" db:"file_id"`   // Scan file
    ReminderSentAt  *time.Time `json:"reminder_sent_at" db:"reminder_sent_at"`
    CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// Training / CPE
type TrainingRecord struct {
    ID            uuid.UUID `json:"id" db:"id"`
    EmployeeID    uuid.UUID `json:"employee_id" db:"employee_id"`
    Title         string    `json:"title" db:"title"`
    Provider      string    `json:"provider" db:"provider"`     // VACPA, internal, external
    Category      string    `json:"category" db:"category"`     // "audit", "tax", "accounting", "soft_skills"
    CPEHours      float64   `json:"cpe_hours" db:"cpe_hours"`   // Số giờ CPE
    Date          time.Time `json:"date" db:"date"`
    CertificateID *uuid.UUID `json:"certificate_id" db:"certificate_id"` // File chứng nhận
    Status        string    `json:"status" db:"status"`          // planned, completed, cancelled
    Year          int       `json:"year" db:"year"`
    CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// Performance Review (KPI)
type PerformanceReview struct {
    ID            uuid.UUID       `json:"id" db:"id"`
    EmployeeID    uuid.UUID       `json:"employee_id" db:"employee_id"`
    ReviewerID    uuid.UUID       `json:"reviewer_id" db:"reviewer_id"`
    Period        string          `json:"period" db:"period"`        // "2026-H1", "2026-FY"
    
    // Metrics from system
    TotalHours    float64         `json:"total_hours" db:"total_hours"`        // From timesheet
    EngagementCount int           `json:"engagement_count" db:"engagement_count"`
    UtilizationRate float64       `json:"utilization_rate" db:"utilization_rate"`
    OnTimeDelivery  float64       `json:"on_time_delivery" db:"on_time_delivery"` // %
    
    // Subjective assessment
    TechnicalScore   int          `json:"technical_score" db:"technical_score"`   // 1-5
    CommunicationScore int        `json:"communication_score" db:"communication_score"` // 1-5
    LeadershipScore  int          `json:"leadership_score" db:"leadership_score"` // 1-5
    OverallScore     float64      `json:"overall_score" db:"overall_score"`
    OverallRating    string       `json:"overall_rating" db:"overall_rating"` // A, B, C, D
    Comments         string       `json:"comments" db:"comments"`
    EmployeeComments string       `json:"employee_comments" db:"employee_comments"`
    
    Status    string    `json:"status" db:"status"` // draft, submitted, acknowledged
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}
```

## 10.3 Service Interface

```go
type HRMService interface {
    // Employee
    CreateEmployee(ctx context.Context, req CreateEmployeeRequest) (*Employee, error)
    UpdateEmployee(ctx context.Context, id uuid.UUID, req UpdateEmployeeRequest) (*Employee, error)
    GetEmployee(ctx context.Context, id uuid.UUID) (*EmployeeDetailResponse, error)
    ListEmployees(ctx context.Context, filter EmployeeFilter, pagination PaginationParams) (*PaginatedResult[EmployeeListItem], error)
    GetOrgChart(ctx context.Context, branchID *uuid.UUID) (*OrgChartNode, error)

    // Certifications
    AddCertification(ctx context.Context, req AddCertRequest) (*Certification, error)
    UpdateCertification(ctx context.Context, id uuid.UUID, req UpdateCertRequest) (*Certification, error)
    GetExpiringCertifications(ctx context.Context, withinDays int) ([]Certification, error)
    CheckAndNotifyExpiring(ctx context.Context) error // Cron job

    // Training / CPE
    RecordTraining(ctx context.Context, req RecordTrainingRequest) (*TrainingRecord, error)
    GetCPESummary(ctx context.Context, employeeID uuid.UUID, year int) (*CPESummary, error)
    GetTeamCPEStatus(ctx context.Context, departmentID uuid.UUID, year int) ([]EmployeeCPEStatus, error)
    GetCPEDeficient(ctx context.Context, year int) ([]EmployeeCPEStatus, error) // KTV thiếu giờ CPE

    // Performance Review
    CreateReview(ctx context.Context, req CreateReviewRequest) (*PerformanceReview, error)
    UpdateReview(ctx context.Context, id uuid.UUID, req UpdateReviewRequest) (*PerformanceReview, error)
    GetReviewsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]PerformanceReview, error)
    GetPendingReviews(ctx context.Context, reviewerID uuid.UUID) ([]PerformanceReview, error)
    GeneratePerformanceMetrics(ctx context.Context, employeeID uuid.UUID, period string) (*PerformanceMetrics, error)

    // Export
    ExportEmployeeList(ctx context.Context, filter EmployeeFilter, format string) ([]byte, error)
}

type CPESummary struct {
    EmployeeID      uuid.UUID         `json:"employee_id"`
    Year            int               `json:"year"`
    TotalCPEHours   float64           `json:"total_cpe_hours"`
    RequiredHours   float64           `json:"required_hours"`     // Theo VACPA
    RemainingHours  float64           `json:"remaining_hours"`
    IsCompliant     bool              `json:"is_compliant"`
    TrainingRecords []TrainingRecord  `json:"training_records"`
}
```

## 10.4 API Endpoints

```
# Employees
POST   /api/v1/employees
PUT    /api/v1/employees/:id
GET    /api/v1/employees/:id
GET    /api/v1/employees
GET    /api/v1/org-chart

# Certifications
POST   /api/v1/employees/:id/certifications
PUT    /api/v1/certifications/:id
GET    /api/v1/employees/:id/certifications
GET    /api/v1/certifications/expiring?days=90

# Training / CPE
POST   /api/v1/training-records
GET    /api/v1/employees/:id/cpe-summary?year=
GET    /api/v1/departments/:id/cpe-status?year=
GET    /api/v1/cpe/deficient?year=

# Performance Review
POST   /api/v1/performance-reviews
PUT    /api/v1/performance-reviews/:id
GET    /api/v1/employees/:id/performance-reviews
GET    /api/v1/my/pending-reviews
GET    /api/v1/employees/:id/performance-metrics?period=
```

---

# 11. MODULE 8: REPORTING & ANALYTICS

## 11.1 Mục tiêu

- Dashboard real-time + báo cáo định kỳ
- Phân quyền xem báo cáo theo level
- Xuất báo cáo Excel, PDF, Word
- KPI tự động tính từ dữ liệu hệ thống

## 11.2 Dashboard Types

### 11.2.1 Executive Dashboard (Chairman/Director/Partner)

```go
type ExecutiveDashboard struct {
    // Revenue
    RevenueThisMonth    int64   `json:"revenue_this_month"`
    RevenueThisQuarter  int64   `json:"revenue_this_quarter"`
    RevenueThisYear     int64   `json:"revenue_this_year"`
    RevenueGrowthRate   float64 `json:"revenue_growth_rate"`   // vs cùng kỳ

    // Engagements
    ActiveEngagements   int     `json:"active_engagements"`
    CompletedThisMonth  int     `json:"completed_this_month"`
    OverdueEngagements  int     `json:"overdue_engagements"`

    // AR
    TotalOutstandingAR  int64   `json:"total_outstanding_ar"`
    OverdueARAmount     int64   `json:"overdue_ar_amount"`

    // HR
    TotalEmployees      int     `json:"total_employees"`
    AvgUtilizationRate  float64 `json:"avg_utilization_rate"`

    // Charts data
    RevenueByService    []ChartDataPoint `json:"revenue_by_service"`
    RevenueByPartner    []ChartDataPoint `json:"revenue_by_partner"`
    RevenueByBranch     []ChartDataPoint `json:"revenue_by_branch"`
    MonthlyRevenueTrend []TimeSeriesPoint `json:"monthly_revenue_trend"`
    EngagementsByStatus []ChartDataPoint `json:"engagements_by_status"`
    ARAgingDistribution []ChartDataPoint `json:"ar_aging_distribution"`
}
```

### 11.2.2 Manager Dashboard

```go
type ManagerDashboard struct {
    MyTeamEngagements   []EngagementSummary `json:"my_team_engagements"`
    PendingApprovals    int                 `json:"pending_approvals"`
    PendingReviews      int                 `json:"pending_reviews"`
    TeamUtilization     []EmployeeUtil      `json:"team_utilization"`
    UpcomingDeadlines   []DeadlineItem      `json:"upcoming_deadlines"`
    OverdueItems        []OverdueItem       `json:"overdue_items"`
}
```

### 11.2.3 Personal Dashboard (All users)

```go
type PersonalDashboard struct {
    MyEngagements      []EngagementSummary `json:"my_engagements"`
    MyTasks            []TaskSummary       `json:"my_tasks"`
    MyTimesheetThisWeek []TimesheetEntry   `json:"my_timesheet_this_week"`
    HoursThisWeek      float64             `json:"hours_this_week"`
    HoursThisMonth     float64             `json:"hours_this_month"`
    PendingActions     []PendingAction     `json:"pending_actions"` // Approvals, reviews, deadlines
    Notifications      []Notification      `json:"notifications"`
}
```

## 11.3 Report Templates

| Báo cáo | Tần suất | Quyền xem | Format |
|---|---|---|---|
| Doanh thu theo dịch vụ | Tháng/Quý/Năm | Director+ | Excel, PDF |
| Doanh thu theo Partner | Tháng/Quý/Năm | Chairman, Partner | Excel, PDF |
| Doanh thu theo chi nhánh | Tháng/Quý/Năm | Director+ | Excel |
| Utilization Rate | Tháng | Manager+ | Excel |
| Công nợ phải thu (AR Aging) | Tuần/Tháng | Accountant, Director+ | Excel, PDF |
| Tiến độ engagement | Tuần | Manager+ | Excel |
| CPE Summary | Năm | HR, Director+ | Excel |
| Chứng chỉ sắp hết hạn | Tháng | HR, Director+ | Excel |
| KPI nhân viên | Quý/Năm | Manager+ | Excel, PDF |
| Tax deadline status | Tuần | Tax team, Director+ | Excel |

## 11.4 Service Interface

```go
type ReportingService interface {
    // Dashboards
    GetExecutiveDashboard(ctx context.Context) (*ExecutiveDashboard, error)
    GetManagerDashboard(ctx context.Context, userID uuid.UUID) (*ManagerDashboard, error)
    GetPersonalDashboard(ctx context.Context, userID uuid.UUID) (*PersonalDashboard, error)

    // Reports
    GenerateRevenueReport(ctx context.Context, req RevenueReportRequest) (*ReportOutput, error)
    GenerateUtilizationReport(ctx context.Context, req UtilizationReportRequest) (*ReportOutput, error)
    GenerateARReport(ctx context.Context, req ARReportRequest) (*ReportOutput, error)
    GenerateEngagementProgressReport(ctx context.Context, req EngProgressRequest) (*ReportOutput, error)
    GenerateCPEReport(ctx context.Context, year int) (*ReportOutput, error)
    GenerateCertExpiryReport(ctx context.Context, withinDays int) (*ReportOutput, error)
    GenerateKPIReport(ctx context.Context, req KPIReportRequest) (*ReportOutput, error)
    GenerateTaxDeadlineReport(ctx context.Context, req TaxDeadlineReportRequest) (*ReportOutput, error)

    // Scheduled reports
    RegisterScheduledReport(ctx context.Context, req ScheduleReportRequest) error
    RunScheduledReports(ctx context.Context) error // Cron job
}

type ReportOutput struct {
    FileName string `json:"file_name"`
    Format   string `json:"format"` // "xlsx", "pdf", "docx"
    Data     []byte `json:"-"`
    URL      string `json:"url"` // Presigned download URL
}
```

## 11.5 API Endpoints

```
# Dashboards
GET    /api/v1/dashboard/executive
GET    /api/v1/dashboard/manager
GET    /api/v1/dashboard/personal

# Reports
POST   /api/v1/reports/revenue
POST   /api/v1/reports/utilization
POST   /api/v1/reports/ar-aging
POST   /api/v1/reports/engagement-progress
POST   /api/v1/reports/cpe
POST   /api/v1/reports/cert-expiry
POST   /api/v1/reports/kpi
POST   /api/v1/reports/tax-deadlines

# Scheduled Reports
POST   /api/v1/reports/schedule
GET    /api/v1/reports/scheduled
DELETE /api/v1/reports/scheduled/:id
```

## 11.6 Frontend Pages

```
/dashboard                          → MainDashboard (role-based rendering)
/reports                            → ReportCatalogPage
/reports/revenue                    → RevenueReportPage (filters + chart + table)
/reports/utilization                → UtilizationReportPage
/reports/ar-aging                   → ARAgingReportPage
/reports/engagement-progress        → EngagementProgressPage
/reports/cpe                        → CPEReportPage
/reports/kpi                        → KPIReportPage
/reports/tax-deadlines              → TaxDeadlineReportPage

# Notification Center
/notifications                      → NotificationCenterPage (all notifications, filter by type/module)

# Settings
/settings/notifications             → NotificationPreferencesPage
    ├── Toggle per channel (in-app, push, email) per notification type
    ├── Quiet hours configuration
    └── Manage push devices (list, remove)

/settings/security                  → SecuritySettingsPage
    ├── Change password
    ├── Two-Factor Authentication
    │     ├── Enable/Disable 2FA (TOTP / Push)
    │     ├── QR code scan & verify setup
    │     └── Download/regenerate backup codes
    ├── Trusted Devices (list, revoke)
    └── Active Sessions (view, terminate)

# Admin — 2FA & Push Management
/admin/security/2fa-policy          → TwoFAPolicyPage (enforce 2FA per role)
/admin/security/2fa-compliance      → TwoFACompliancePage (users chưa bật 2FA)
/admin/push/delivery-logs           → PushDeliveryLogPage (debug push failures)
/admin/push/devices                 → PushDeviceAdminPage (all registered devices)
```

---

# 12. NON-FUNCTIONAL REQUIREMENTS

## 12.1 Performance

| Metric | Target | Cách đo |
|---|---|---|
| API response time (p95) | ≤ 200ms cho CRUD, ≤ 500ms cho report | Prometheus histogram |
| Page load time (TTI) | ≤ 2s (3G) | Lighthouse |
| Concurrent users | 300 | Load testing (k6) |
| Database query time | ≤ 50ms (p95) | pg_stat_statements |
| File upload (50MB) | ≤ 10s | End-to-end test |
| Report generation | ≤ 30s cho 10k records | Benchmark |
| WebSocket latency | ≤ 100ms | Ping measurement |
| **Push delivery latency** | **≤ 2s (device online), ≤ 30s (reconnect)** | **Push delivery log timestamps** |
| **Push relay concurrent connections** | **600 persistent connections** | **Goroutine + connection pool metrics** |
| **2FA TOTP verification** | **≤ 100ms** | **Prometheus histogram** |
| **2FA push challenge round-trip** | **≤ 10s (user response dependent)** | **Challenge created_at → responded_at** |
| **Web Push delivery** | **≤ 5s** | **VAPID push send → delivery ack** |

### Strategies

- **Database**: Connection pooling (pgbouncer), proper indexing, query optimization, read replicas (phase 2)
- **Caching**: Redis cho hot data (dashboard, user sessions, permissions, 2FA challenges TTL 5min)
- **Frontend**: React Query caching, code splitting, lazy loading, ISR/SSG cho static pages
- **API**: Pagination bắt buộc cho list endpoints, batch APIs, gzip compression
- **Background**: Heavy operations (report gen, email, notifications) via message queue
- **Push Relay**: Goroutine-per-connection, epoll/kqueue multiplexing (gobwas/ws for memory efficiency), sticky sessions via K8s headless service
- **2FA**: TOTP verification in-memory (no DB call), challenge state in Redis with TTL auto-expire

## 12.2 Reliability

| Metric | Target |
|---|---|
| Uptime | 99.5% (≤ 3.65h downtime/tháng) |
| RPO (Recovery Point Objective) | ≤ 1 hour |
| RTO (Recovery Time Objective) | ≤ 4 hours |
| Backup frequency | Database: hourly incremental, daily full; Files: daily sync |

### Strategies

- **Database**: PostgreSQL streaming replication, automated failover
- **Application**: Health check endpoints, graceful shutdown, circuit breaker pattern
- **Storage**: MinIO replication, cross-AZ backup
- **Monitoring**: Alert on error rate > 1%, latency p99 > 1s, disk > 80%
- **Push Relay**: Auto-reconnect with exponential backoff (client-side), offline message queue (Redis Streams) ensures no message loss, heartbeat detection marks stale connections within 60s
- **2FA**: Backup codes ensure access when primary device unavailable, fallback to email-based password reset as last resort

## 12.3 Security

### Authentication & Authorization

- JWT tokens: Access (15min), Refresh (7 days), rotation on refresh
- Password: bcrypt with cost=12, min 8 chars, complexity requirements
- RBAC + ABAC: Role-based + branch/department scoping
- Session management: single active session per user (configurable)
- Brute force protection: 5 failed attempts → lock 15 minutes

### Two-Factor Authentication (2FA/MFA) — Xác thực 2 lớp

**Yêu cầu từ stakeholder:** Bảo mật xác nhận 2 lớp, tự triển khai không phụ thuộc bên thứ 3.

#### Phương thức 2FA hỗ trợ

| Phương thức | Mô tả | Khi nào dùng |
|---|---|---|
| **TOTP** (Time-based OTP) | Mã 6 số thay đổi mỗi 30 giây, dùng app Google Authenticator/Authy | Phương thức chính, hoạt động offline |
| **Push-based** | Gửi push tới mobile app tự build, user tap Approve/Reject | Tiện lợi hơn TOTP, cần có mobile app |
| **Backup Codes** | 10 mã dùng 1 lần, mỗi mã 8 ký tự | Khi mất device hoặc không truy cập app |

#### Login Flow với 2FA

```
User nhập Email + Password
         │
         ▼
   ┌─────────────┐
   │ Verify      │──── Sai ──▶ [401 Invalid credentials]
   │ credentials │              (không tiết lộ field nào sai)
   └──────┬──────┘
          │ Đúng
          ▼
   ┌─────────────┐
   │ 2FA enabled?│──── Không ──▶ [200 + TokenPair] → Login thành công
   └──────┬──────┘
          │ Có
          ▼
   ┌─────────────────┐
   │ Trusted device?  │──── Có ──▶ [200 + TokenPair] → Skip 2FA
   │ (fingerprint +   │
   │  trong 30 ngày)  │
   └──────┬───────────┘
          │ Không
          ▼
   ┌──────────────────┐
   │ Return Challenge  │
   │ {challenge_id,    │
   │  challenge_type,  │──── method = "totp" ──▶ Frontend hiển thị OTP input
   │  expires_in: 300} │
   │                   │──── method = "push"  ──▶ Server gửi push tới mobile app
   │                   │                           Frontend polling check response
   └──────────────────┘
          │
          ▼
   ┌──────────────────┐
   │ User nhập OTP    │
   │ HOẶC tap Approve │──── Sai (3 lần) ──▶ [401 + lock 15 min]
   │ HOẶC backup code │
   └──────┬───────────┘
          │ Đúng
          ▼
   ┌──────────────────┐
   │ Trust device?     │──── Có ──▶ Lưu device fingerprint (30 ngày)
   └──────┬───────────┘
          │
          ▼
   [200 + TokenPair] → Login thành công
```

#### Bảo mật 2FA

- **TOTP Secret**: Mã hóa AES-256-GCM trước khi lưu vào DB, key quản lý bởi KMS
- **TOTP Algorithm**: SHA-1, 6 digits, 30-second period (RFC 6238 compliant)
- **Challenge TTL**: 5 phút (hết hạn tự động)
- **Rate limit**: Tối đa 5 lần verify sai mỗi challenge → challenge bị hủy
- **Backup codes**: Bcrypt hash, mỗi code dùng 1 lần rồi xóa
- **Trusted device**: SHA-256 fingerprint (User-Agent + Screen + Timezone + Canvas), TTL 30 ngày, tối đa 5 devices
- **Push 2FA**: Mã hóa end-to-end payload, challenge ID là UUID v4 + HMAC signature
- **Mandatory 2FA**: Admin có thể enforce 2FA cho role cụ thể (Partner, Director, Manager — bắt buộc)
- **Recovery**: Khi mất device → dùng backup code → regenerate TOTP secret

#### Thư viện Self-Hosted (không phụ thuộc SaaS)

| Component | Thư viện Go | Ghi chú |
|---|---|---|
| TOTP generation/verification | `github.com/pquerna/otp` | RFC 6238, self-contained |
| QR code generation | `github.com/skip2/go-qrcode` | Tự render PNG, không gọi API ngoài |
| Web Push (VAPID) | `github.com/SherClockHolmes/webpush-go` | W3C standard, self-hosted |
| WebSocket (push relay) | `github.com/gorilla/websocket` | Persistent connection mobile app |
| Backup code generation | `crypto/rand` (stdlib) | Cryptographically secure random |
| Device fingerprinting | Frontend: `@fingerprintjs/fingerprintjs` (open-source MIT) | Hoặc tự implement canvas + UA hashing |

#### 2FA Frontend Pages

```
/settings/security                    → SecuritySettingsPage
    ├── Enable/Disable 2FA             → TwoFactorSetupWizard
    │     Step 1: Chọn method (TOTP / Push)
    │     Step 2: Scan QR code (TOTP) hoặc cài mobile app (Push)
    │     Step 3: Nhập OTP xác minh
    │     Step 4: Lưu backup codes (bắt buộc xác nhận đã lưu)
    ├── Manage Trusted Devices          → TrustedDevicesManager
    ├── Regenerate Backup Codes         → BackupCodesRegenDialog
    └── Active Sessions                 → ActiveSessionsManager

/login                                → LoginPage (updated)
    ├── Email + Password form
    └── 2FA Challenge Dialog (conditional)
          ├── TOTP: OTP 6-digit input + countdown timer
          ├── Push: "Đã gửi thông báo tới app" + polling indicator
          ├── "Dùng backup code" link
          └── "Trust this device for 30 days" checkbox
```

#### 2FA API Endpoints

```
# 2FA Setup
POST   /api/v1/auth/2fa/enable                    # Bật 2FA, trả về QR + backup codes
POST   /api/v1/auth/2fa/verify-setup               # Xác minh OTP lần đầu khi bật 2FA
POST   /api/v1/auth/2fa/disable                    # Tắt 2FA (cần password)
POST   /api/v1/auth/2fa/regenerate-backup-codes    # Tạo lại backup codes

# 2FA Login Verification
POST   /api/v1/auth/2fa/verify                     # Xác minh OTP khi login
POST   /api/v1/auth/2fa/verify-backup              # Dùng backup code
POST   /api/v1/auth/2fa/push-response              # Mobile app gọi: approve/reject
GET    /api/v1/auth/2fa/push-status/:challengeId   # Frontend polling: đã approve chưa?
POST   /api/v1/auth/2fa/resend-push                # Gửi lại push challenge

# Trusted Devices
GET    /api/v1/auth/trusted-devices                # Danh sách trusted devices
DELETE /api/v1/auth/trusted-devices/:id            # Revoke 1 device
DELETE /api/v1/auth/trusted-devices                # Revoke tất cả

# 2FA Admin
GET    /api/v1/auth/2fa/status                     # Status 2FA của user hiện tại
POST   /api/v1/admin/2fa/enforce                   # Bắt buộc 2FA cho role
GET    /api/v1/admin/2fa/policy                    # Lấy policy hiện tại
GET    /api/v1/admin/2fa/compliance                # Danh sách user chưa bật 2FA (khi bị enforce)
```

### Data Security

- **At rest**: AES-256 encryption cho files nhạy cảm (BCTC chưa công bố)
- **In transit**: TLS 1.3, HSTS
- **Database**: Column-level encryption cho PII
- **Audit trail**: Immutable, append-only log (không được xóa/sửa)

### Application Security

- Input validation: go-playground/validator, tất cả endpoints
- SQL injection: sqlc (compile-time, parameterized queries)
- XSS: React auto-escaping, CSP headers
- CSRF: SameSite cookies + CSRF token
- Rate limiting: 100 req/min per user, 1000 req/min per IP
- CORS: Whitelist origins only
- File upload: type validation, size limits (50MB), virus scanning (ClamAV)
- API versioning: `/api/v1/` prefix

### Compliance

- Tuân thủ Luật Kiểm toán độc lập
- Audit trail cho mọi thao tác trên dữ liệu KH
- Data retention: 10 năm cho hồ sơ kiểm toán
- Phân quyền truy cập theo dự án (engagement-level ACL)
- Chỉ BGĐ phê duyệt mới được cung cấp thông tin KH ra bên ngoài

## 12.4 Maintainability

| Aspect | Standard |
|---|---|
| Code coverage | ≥ 80% unit tests |
| Linting | golangci-lint (Go), ESLint + Prettier (TS) |
| Documentation | OpenAPI 3.0 spec, auto-generated from code annotations |
| Migration | Sequential numbered SQL migrations, up/down |
| Logging | Structured JSON logs (zerolog), correlation IDs |
| Configuration | Environment variables, .env files, no hardcoded secrets |
| Error handling | Custom error types, error codes, user-friendly messages |

### Coding Standards (Go)

- Clean Architecture / Hexagonal Architecture
- Interface-driven design (dependency injection)
- Table-driven tests
- Context propagation for cancellation & tracing
- Structured error wrapping with `fmt.Errorf("...: %w", err)`

### Coding Standards (React/Next.js)

- TypeScript strict mode
- Component composition pattern
- Custom hooks for shared logic
- Colocation: test, style, types cùng component
- Accessibility: WCAG 2.1 AA

## 12.5 Scalability

| Dimension | Current | Target (3-5yr) |
|---|---|---|
| Users | 100 | 200 |
| Concurrent (web) | 300 | 600 |
| Push relay connections | 200 (mobile + web push) | 500 |
| Data volume | ~10GB/yr | ~30GB/yr |
| File storage | ~100GB/yr | ~500GB/yr |
| Push messages/day | ~2,000 | ~10,000 |

### Strategies

- Horizontal scaling via Kubernetes (stateless API servers)
- Database: partitioning by year cho audit logs, timesheets, push_delivery_logs
- File storage: MinIO cluster, lifecycle policies (cold storage after 2 years)
- CDN cho static assets
- Push Relay: Sticky sessions (K8s headless service), sharding by user_id hash, gobwas/ws for memory-efficient WebSocket (1 goroutine per N connections instead of 2 goroutines per connection)
- 2FA challenges: Redis with TTL (no DB pressure), auto-expire after 5 minutes
- Push offline queue: Redis Streams per device, auto-purge after 24h TTL

---

# 13. DATABASE SCHEMA TỔNG QUAN

## 13.1 Core Tables

```sql
-- Organization
CREATE TABLE branches (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), ...);
CREATE TABLE departments (id UUID PRIMARY KEY, branch_id UUID REFERENCES branches(id), ...);

-- Auth
CREATE TABLE users (id UUID PRIMARY KEY, email VARCHAR(255) UNIQUE NOT NULL, two_factor_enabled BOOLEAN DEFAULT FALSE, two_factor_method VARCHAR(10), two_factor_secret TEXT, ...);
CREATE TABLE roles (id UUID PRIMARY KEY, code VARCHAR(50) UNIQUE NOT NULL, ...);
CREATE TABLE permissions (id UUID PRIMARY KEY, module VARCHAR(50), resource VARCHAR(50), action VARCHAR(50), ...);
CREATE TABLE user_roles (user_id UUID REFERENCES users(id), role_id UUID REFERENCES roles(id), PRIMARY KEY (user_id, role_id));
CREATE TABLE role_permissions (role_id UUID REFERENCES roles(id), permission_id UUID REFERENCES permissions(id), scope VARCHAR(20), ...);

-- 2FA & Trusted Devices
CREATE TABLE two_factor_backup_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    code_hash VARCHAR(255) NOT NULL,           -- bcrypt hash
    is_used BOOLEAN DEFAULT FALSE,
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE trusted_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    device_fingerprint VARCHAR(255) NOT NULL,   -- SHA-256 hash
    device_name VARCHAR(255),
    ip_address VARCHAR(45),
    trusted_until TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE two_factor_challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    challenge_type VARCHAR(10) NOT NULL,        -- "totp", "push"
    status VARCHAR(20) DEFAULT 'pending',       -- "pending", "approved", "rejected", "expired"
    expires_at TIMESTAMP NOT NULL,
    responded_at TIMESTAMP,
    device_token VARCHAR(255),                  -- Push response device
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE two_factor_policy (
    role_code VARCHAR(50) PRIMARY KEY,
    enforced BOOLEAN DEFAULT FALSE,
    enforced_at TIMESTAMP,
    enforced_by UUID REFERENCES users(id)
);

-- Push Notification Devices
CREATE TABLE push_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    device_token VARCHAR(500) UNIQUE NOT NULL,
    platform VARCHAR(20) NOT NULL,              -- "ios", "android", "web_push"
    device_name VARCHAR(255),
    app_version VARCHAR(50),
    os_version VARCHAR(50),
    is_active BOOLEAN DEFAULT TRUE,
    last_active_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE web_push_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL UNIQUE,
    p256dh_key TEXT NOT NULL,
    auth_key TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE push_delivery_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID REFERENCES notifications(id),
    device_id UUID REFERENCES push_devices(id),
    status VARCHAR(20) DEFAULT 'sent',          -- "sent", "delivered", "failed", "expired"
    error_message TEXT,
    sent_at TIMESTAMP DEFAULT NOW(),
    delivered_at TIMESTAMP,
    retry_count INT DEFAULT 0
);
CREATE TABLE notification_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    assignment_in_app BOOLEAN DEFAULT TRUE,
    assignment_push BOOLEAN DEFAULT TRUE,
    assignment_email BOOLEAN DEFAULT FALSE,
    deadline_reminder_in_app BOOLEAN DEFAULT TRUE,
    deadline_reminder_push BOOLEAN DEFAULT TRUE,
    deadline_reminder_email BOOLEAN DEFAULT TRUE,
    approval_in_app BOOLEAN DEFAULT TRUE,
    approval_push BOOLEAN DEFAULT TRUE,
    review_in_app BOOLEAN DEFAULT TRUE,
    review_push BOOLEAN DEFAULT TRUE,
    quiet_hours_start TIME,
    quiet_hours_end TIME,
    updated_at TIMESTAMP DEFAULT NOW()
);

-- CRM
CREATE TABLE clients (id UUID PRIMARY KEY, code VARCHAR(20) UNIQUE, tax_code VARCHAR(14) UNIQUE, ...);
CREATE TABLE client_contacts (id UUID PRIMARY KEY, client_id UUID REFERENCES clients(id), ...);
CREATE TABLE client_risk_assessments (id UUID PRIMARY KEY, client_id UUID REFERENCES clients(id), ...);
CREATE TABLE conflict_checks (id UUID PRIMARY KEY, client_id UUID REFERENCES clients(id), ...);

-- Engagement
CREATE TABLE engagements (id UUID PRIMARY KEY, code VARCHAR(20) UNIQUE, client_id UUID REFERENCES clients(id), ...);
CREATE TABLE engagement_members (id UUID PRIMARY KEY, engagement_id UUID REFERENCES engagements(id), ...);
CREATE TABLE engagement_phases (id UUID PRIMARY KEY, engagement_id UUID REFERENCES engagements(id), ...);
CREATE TABLE engagement_tasks (id UUID PRIMARY KEY, phase_id UUID REFERENCES engagement_phases(id), ...);
CREATE TABLE engagement_costs (id UUID PRIMARY KEY, engagement_id UUID REFERENCES engagements(id), ...);

-- Timesheet
CREATE TABLE attendance (id UUID PRIMARY KEY, employee_id UUID, date DATE, ...);
CREATE TABLE timesheet_entries (id UUID PRIMARY KEY, employee_id UUID, engagement_id UUID REFERENCES engagements(id), ...);
CREATE TABLE resource_allocations (id UUID PRIMARY KEY, employee_id UUID, engagement_id UUID REFERENCES engagements(id), ...);

-- Billing
CREATE TABLE invoices (id UUID PRIMARY KEY, code VARCHAR(20) UNIQUE, engagement_id UUID REFERENCES engagements(id), ...);
CREATE TABLE invoice_line_items (id UUID PRIMARY KEY, invoice_id UUID REFERENCES invoices(id), ...);
CREATE TABLE payments (id UUID PRIMARY KEY, invoice_id UUID REFERENCES invoices(id), ...);
CREATE TABLE billing_milestones (id UUID PRIMARY KEY, engagement_id UUID REFERENCES engagements(id), ...);

-- Working Papers
CREATE TABLE wp_folders (id UUID PRIMARY KEY, engagement_id UUID REFERENCES engagements(id), ...);
CREATE TABLE working_papers (id UUID PRIMARY KEY, engagement_id UUID REFERENCES engagements(id), folder_id UUID REFERENCES wp_folders(id), ...);
CREATE TABLE audit_templates (id UUID PRIMARY KEY, ...);
CREATE TABLE review_comments (id UUID PRIMARY KEY, working_paper_id UUID REFERENCES working_papers(id), ...);

-- Tax & Advisory
CREATE TABLE tax_deadlines (id UUID PRIMARY KEY, client_id UUID REFERENCES clients(id), ...);
CREATE TABLE advisory_records (id UUID PRIMARY KEY, client_id UUID REFERENCES clients(id), ...);

-- HRM
CREATE TABLE employees (id UUID PRIMARY KEY, code VARCHAR(20) UNIQUE, user_id UUID REFERENCES users(id), ...);
CREATE TABLE certifications (id UUID PRIMARY KEY, employee_id UUID REFERENCES employees(id), ...);
CREATE TABLE training_records (id UUID PRIMARY KEY, employee_id UUID REFERENCES employees(id), ...);
CREATE TABLE performance_reviews (id UUID PRIMARY KEY, employee_id UUID REFERENCES employees(id), ...);

-- Global
CREATE TABLE audit_logs (id UUID PRIMARY KEY, user_id UUID, module VARCHAR(50), ...) PARTITION BY RANGE (created_at);
CREATE TABLE notifications (id UUID PRIMARY KEY, user_id UUID REFERENCES users(id), ...);
CREATE TABLE file_metadata (id UUID PRIMARY KEY, module VARCHAR(50), resource_id UUID, ...);
CREATE TABLE workflow_definitions (id UUID PRIMARY KEY, ...);
CREATE TABLE workflow_instances (id UUID PRIMARY KEY, definition_id UUID REFERENCES workflow_definitions(id), ...);
CREATE TABLE workflow_actions (id UUID PRIMARY KEY, instance_id UUID REFERENCES workflow_instances(id), ...);

-- Indexes (critical ones)
CREATE INDEX idx_clients_tax_code ON clients(tax_code);
CREATE INDEX idx_clients_branch ON clients(branch_id);
CREATE INDEX idx_engagements_client ON engagements(client_id);
CREATE INDEX idx_engagements_status ON engagements(status);
CREATE INDEX idx_engagements_branch ON engagements(branch_id);
CREATE INDEX idx_timesheet_employee_date ON timesheet_entries(employee_id, date);
CREATE INDEX idx_timesheet_engagement ON timesheet_entries(engagement_id);
CREATE INDEX idx_invoices_client ON invoices(client_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_due_date ON invoices(due_date);
CREATE INDEX idx_working_papers_engagement ON working_papers(engagement_id);
CREATE INDEX idx_audit_logs_module_resource ON audit_logs(module, resource, resource_id);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_notifications_user_read ON notifications(user_id, is_read);
CREATE INDEX idx_tax_deadlines_due ON tax_deadlines(due_date, status);
CREATE INDEX idx_certifications_expiry ON certifications(expiry_date, status);
CREATE INDEX idx_employees_branch_dept ON employees(branch_id, department_id);

-- 2FA & Push Indexes
CREATE INDEX idx_trusted_devices_user ON trusted_devices(user_id, trusted_until);
CREATE INDEX idx_trusted_devices_fingerprint ON trusted_devices(device_fingerprint);
CREATE INDEX idx_2fa_challenges_status ON two_factor_challenges(status, expires_at);
CREATE INDEX idx_push_devices_user ON push_devices(user_id, is_active);
CREATE INDEX idx_push_devices_token ON push_devices(device_token);
CREATE INDEX idx_push_delivery_status ON push_delivery_logs(status, sent_at);
CREATE INDEX idx_web_push_user ON web_push_subscriptions(user_id, is_active);
CREATE INDEX idx_2fa_backup_user ON two_factor_backup_codes(user_id, is_used);
```

---

# 14. API DESIGN CONVENTIONS

## 14.1 URL Patterns

```
BASE: /api/v1

# Nouns, plural, lowercase, kebab-case
GET    /api/v1/{resources}              → List (paginated)
POST   /api/v1/{resources}              → Create
GET    /api/v1/{resources}/{id}         → Get by ID
PUT    /api/v1/{resources}/{id}         → Full update
PATCH  /api/v1/{resources}/{id}         → Partial update
DELETE /api/v1/{resources}/{id}         → Soft delete

# Sub-resources
GET    /api/v1/{resources}/{id}/{sub-resources}

# Actions (non-CRUD)
POST   /api/v1/{resources}/{id}/{action}  → approve, reject, submit, etc.

# My resources
GET    /api/v1/my/{resources}             → Current user's resources
```

## 14.2 Response Format

```json
// Success
{
  "success": true,
  "data": { ... },
  "meta": {
    "page": 1,
    "page_size": 20,
    "total": 150,
    "total_pages": 8
  }
}

// Error
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Dữ liệu không hợp lệ",
    "details": [
      {"field": "tax_code", "message": "MST phải có 10 hoặc 13-14 ký tự"}
    ]
  }
}
```

## 14.3 Error Codes

| HTTP | Code | Mô tả |
|---|---|---|
| 400 | VALIDATION_ERROR | Input validation failed |
| 401 | UNAUTHORIZED | Missing/invalid token |
| 403 | FORBIDDEN | Insufficient permissions |
| 404 | NOT_FOUND | Resource not found |
| 409 | CONFLICT | Duplicate (e.g., tax_code) |
| 422 | BUSINESS_RULE_VIOLATION | Business logic error |
| 429 | RATE_LIMIT_EXCEEDED | Too many requests |
| 500 | INTERNAL_ERROR | Server error |

---

# 15. DEPLOYMENT & INFRASTRUCTURE

## 15.1 Environments

| Env | Purpose | URL |
|---|---|---|
| Local | Development | localhost:3000 / localhost:8080 |
| Staging | QA & UAT | staging.erp.company.vn |
| Production | Live | erp.company.vn |

## 15.2 Cloud Architecture (AWS / GCP / Azure)

```
Internet → CloudFlare (CDN + WAF)
         → Load Balancer
         → Kubernetes Cluster
              ├── Frontend Pods (Next.js)
              ├── API Pods (Go)
              ├── Worker Pods (Background jobs)
              ├── WebSocket Pods (in-app real-time)
              └── Push Relay Pods (persistent mobile connections, sticky sessions)
         → Managed PostgreSQL (RDS/Cloud SQL)
         → Redis Cluster (ElastiCache) — sessions, push offline queue, 2FA challenges
         → Object Storage (S3/GCS/MinIO)
         → NATS (Message Queue)
```

## 15.3 Cron Jobs

| Job | Schedule | Mô tả |
|---|---|---|
| Tax deadline reminder | Daily 8:00 AM | Check & send reminders for deadlines ≤ 7 days |
| Cert expiry reminder | Weekly Monday | Check certs expiring within 90 days |
| Invoice overdue check | Daily 9:00 AM | Mark overdue invoices, send reminders |
| Scheduled reports | Configurable | Generate & email scheduled reports |
| Audit log partition | Monthly 1st | Create new partition for audit_logs |
| Database backup | Hourly/Daily | Automated backup |
| File cold storage | Monthly | Move files > 2 years to cold storage |
| 2FA challenge cleanup | Every 10 min | Expire challenges older than 5 min |
| Trusted device cleanup | Daily midnight | Remove expired trusted devices |
| Push device cleanup | Weekly | Mark devices inactive if no heartbeat > 30 days |
| Push retry failed | Every 5 min | Retry failed push deliveries (max 3 attempts) |
| Push offline queue flush | Every 1 min | Send queued messages to reconnected devices |

## 15.4 Rollout Plan

| Phase | Timeline | Modules | Mục tiêu |
|---|---|---|---|
| **Phase 1** | Tháng 1-3 | Global (Auth + 2FA TOTP + Org) + CRM + HRM | Foundation, 2FA TOTP, user/client management |
| **Phase 2** | Tháng 3-5 | Engagement + Timesheet + Push Notification (Web Push) | Core business, Web Push notification khi assign |
| **Phase 3** | Tháng 5-7 | Billing + Working Papers + Mobile App v1 (React Native) | Revenue, document mgmt, mobile push notification |
| **Phase 4** | Tháng 7-9 | Tax & Advisory + Reporting + 2FA Push (via mobile app) | Full features, 2FA push approve trên app |
| **Phase 5** | Tháng 9-12 | UAT, Training, Go-live, 2FA enforcement | Testing, training (Intern & Junior), enforce 2FA cho Partner/Director |

---

*END OF SPEC — Version 1.1*