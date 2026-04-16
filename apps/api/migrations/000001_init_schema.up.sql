-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- Global: Organizations / Branches
-- ============================================================
CREATE TABLE branches (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(20)  NOT NULL UNIQUE,
    name        VARCHAR(200) NOT NULL,
    address     TEXT,
    phone       VARCHAR(20),
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  UUID,
    updated_by  UUID
);

CREATE TABLE departments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id   UUID REFERENCES branches(id),
    code        VARCHAR(20)  NOT NULL UNIQUE,
    name        VARCHAR(200) NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  UUID,
    updated_by  UUID
);

-- ============================================================
-- Global: Roles & Permissions
-- ============================================================
CREATE TABLE roles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(50)  NOT NULL UNIQUE,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    level       INT NOT NULL DEFAULT 0,
    is_system   BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed default roles
INSERT INTO roles (code, name, description, level, is_system) VALUES
    ('SUPER_ADMIN',    'System Admin',          'Full system access',              1, true),
    ('FIRM_PARTNER',   'Firm Partner',           'Partner-level access',            2, true),
    ('AUDIT_MANAGER',  'Audit Manager',          'Manage audit engagements',        3, true),
    ('AUDIT_STAFF',    'Audit Staff',            'Execute audit tasks',             4, true),
    ('CLIENT_ADMIN',   'Client Administrator',   'Client portal admin',             5, true),
    ('CLIENT_USER',    'Client User',            'Client portal read-only access',  6, true);

CREATE TABLE permissions (
    id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    module   VARCHAR(50) NOT NULL,
    resource VARCHAR(50) NOT NULL,
    action   VARCHAR(50) NOT NULL,
    UNIQUE(module, resource, action)
);

CREATE TABLE role_permissions (
    role_id       UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    scope         VARCHAR(20) NOT NULL DEFAULT 'all'
        CHECK (scope IN ('all', 'branch', 'department', 'own')),
    PRIMARY KEY (role_id, permission_id)
);

-- ============================================================
-- Global: Users
-- ============================================================
CREATE TABLE users (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email                   VARCHAR(255) NOT NULL UNIQUE,
    hashed_password         VARCHAR(255) NOT NULL,
    full_name               VARCHAR(200) NOT NULL,
    employee_id             UUID,
    branch_id               UUID REFERENCES branches(id),
    department_id           UUID REFERENCES departments(id),
    status                  VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'inactive', 'locked')),
    last_login_at           TIMESTAMPTZ,

    -- 2FA
    two_factor_enabled      BOOLEAN NOT NULL DEFAULT false,
    two_factor_method       VARCHAR(20) DEFAULT 'totp'
        CHECK (two_factor_method IN ('totp', 'push')),
    two_factor_secret       TEXT,
    two_factor_verified_at  TIMESTAMPTZ,
    backup_codes_hash       TEXT,
    trusted_devices         JSONB NOT NULL DEFAULT '[]',

    -- Push notifications
    push_subscriptions      JSONB NOT NULL DEFAULT '[]',

    is_deleted              BOOLEAN NOT NULL DEFAULT false,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by              UUID REFERENCES users(id),
    updated_by              UUID REFERENCES users(id)
);

CREATE INDEX idx_users_email ON users(email) WHERE is_deleted = false;
CREATE INDEX idx_users_branch ON users(branch_id) WHERE is_deleted = false;

CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- ============================================================
-- Global: Audit Trail (immutable)
-- ============================================================
CREATE TABLE audit_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID REFERENCES users(id),
    module      VARCHAR(50)  NOT NULL,
    resource    VARCHAR(50)  NOT NULL,
    resource_id UUID,
    action      VARCHAR(50)  NOT NULL,
    old_value   JSONB,
    new_value   JSONB,
    ip_address  VARCHAR(45),
    user_agent  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_resource ON audit_logs(module, resource, resource_id);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- ============================================================
-- Global: Outbox Pattern (domain events)
-- ============================================================
CREATE TABLE outbox_messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type  VARCHAR(50)  NOT NULL,
    aggregate_id    UUID         NOT NULL,
    event_type      VARCHAR(100) NOT NULL,
    payload         JSONB        NOT NULL,
    status          VARCHAR(20)  NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'processing', 'sent', 'failed')),
    retry_count     INT NOT NULL DEFAULT 0,
    error_message   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at    TIMESTAMPTZ
);

CREATE INDEX idx_outbox_status ON outbox_messages(status, created_at) WHERE status IN ('pending', 'failed');
