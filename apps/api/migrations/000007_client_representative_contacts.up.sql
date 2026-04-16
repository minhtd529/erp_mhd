-- Phase 1.7: Legal representative fields + client contacts table

-- ── A: Người đại diện pháp lý (on clients) ───────────────────────────────────
ALTER TABLE clients
    ADD COLUMN IF NOT EXISTS representative_name  VARCHAR(200),
    ADD COLUMN IF NOT EXISTS representative_title VARCHAR(100),
    ADD COLUMN IF NOT EXISTS representative_phone VARCHAR(20);

-- ── B: Người liên hệ đầu mối ─────────────────────────────────────────────────
CREATE TABLE client_contacts (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id  UUID        NOT NULL REFERENCES clients(id),
    full_name  VARCHAR(200) NOT NULL,
    title      VARCHAR(100),
    phone      VARCHAR(20),
    email      VARCHAR(255),
    is_primary BOOLEAN     NOT NULL DEFAULT false,
    is_deleted BOOLEAN     NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id)
);

CREATE INDEX idx_client_contacts_client_id ON client_contacts(client_id) WHERE is_deleted = false;
