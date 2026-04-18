-- Phase 2: Timesheet bounded context tables

CREATE TABLE IF NOT EXISTS timesheets (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    staff_id          UUID NOT NULL REFERENCES users(id),
    period_start_date DATE NOT NULL,           -- always a Monday
    status            VARCHAR(20) NOT NULL DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT','SUBMITTED','APPROVED','REJECTED','LOCKED')),
    total_hours       NUMERIC(6,2) NOT NULL DEFAULT 0,
    submitted_at      TIMESTAMPTZ,
    submitted_by      UUID REFERENCES users(id),
    approved_at       TIMESTAMPTZ,
    approved_by       UUID REFERENCES users(id),
    reject_reason     TEXT,
    locked_at         TIMESTAMPTZ,
    is_deleted        BOOLEAN NOT NULL DEFAULT false,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by        UUID NOT NULL REFERENCES users(id),
    updated_by        UUID NOT NULL REFERENCES users(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_timesheets_staff_week
    ON timesheets(staff_id, period_start_date)
    WHERE is_deleted = false;

CREATE INDEX IF NOT EXISTS idx_timesheets_staff_id_period_start
    ON timesheets(staff_id, period_start_date DESC)
    WHERE is_deleted = false;

CREATE INDEX IF NOT EXISTS idx_timesheets_status
    ON timesheets(status)
    WHERE is_deleted = false;

CREATE TABLE IF NOT EXISTS timesheet_entries (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timesheet_id  UUID NOT NULL REFERENCES timesheets(id) ON DELETE CASCADE,
    entry_date    DATE NOT NULL,
    engagement_id UUID NOT NULL REFERENCES engagements(id),
    task_id       UUID REFERENCES engagement_tasks(id),
    hours_worked  NUMERIC(4,2) NOT NULL CHECK (hours_worked > 0 AND hours_worked <= 24),
    description   TEXT,
    is_deleted    BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by    UUID NOT NULL REFERENCES users(id),
    updated_by    UUID NOT NULL REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_timesheet_entries_timesheet_id
    ON timesheet_entries(timesheet_id)
    WHERE is_deleted = false;

CREATE INDEX IF NOT EXISTS idx_timesheet_entries_engagement_id
    ON timesheet_entries(engagement_id)
    WHERE is_deleted = false;

CREATE TABLE IF NOT EXISTS attendance (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    staff_id       UUID NOT NULL REFERENCES users(id),
    check_in_time  TIMESTAMPTZ NOT NULL,
    check_out_time TIMESTAMPTZ,
    location       VARCHAR(20) NOT NULL DEFAULT 'ON_SITE'
        CHECK (location IN ('ON_SITE','REMOTE')),
    status         VARCHAR(20) NOT NULL DEFAULT 'PRESENT'
        CHECK (status IN ('PRESENT','LEAVE','HOLIDAY','ABSENT')),
    notes          TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_attendance_staff_id_check_in_time
    ON attendance(staff_id, check_in_time DESC);
