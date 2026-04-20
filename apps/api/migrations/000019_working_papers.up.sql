-- Migration 000019: Working Paper bounded context

CREATE TABLE working_paper_folders (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    engagement_id UUID        NOT NULL REFERENCES engagements(id),
    folder_name   VARCHAR(200) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by    UUID        NOT NULL
);

CREATE INDEX idx_wp_folders_engagement ON working_paper_folders (engagement_id);

CREATE TABLE working_papers (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    engagement_id   UUID         NOT NULL REFERENCES engagements(id),
    folder_id       UUID         REFERENCES working_paper_folders(id),
    document_type   VARCHAR(30)  NOT NULL
                    CHECK (document_type IN ('PROCEDURES','EVIDENCE','ANALYSIS','CONCLUSION','MANAGEMENT_LETTER')),
    title           VARCHAR(300) NOT NULL,
    status          VARCHAR(20)  NOT NULL DEFAULT 'DRAFT'
                    CHECK (status IN ('DRAFT','IN_REVIEW','COMMENTED','FINALIZED','SIGNED_OFF')),
    file_id         UUID,
    snapshot_data   JSONB        NOT NULL DEFAULT '{}',
    is_deleted      BOOLEAN      NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by      UUID         NOT NULL,
    updated_by      UUID         NOT NULL
);

CREATE INDEX idx_working_papers_engagement ON working_papers (engagement_id, status);
CREATE INDEX idx_working_papers_folder     ON working_papers (folder_id) WHERE folder_id IS NOT NULL;
CREATE INDEX idx_working_papers_file       ON working_papers (file_id)   WHERE file_id IS NOT NULL;

CREATE TABLE working_paper_reviews (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    working_paper_id UUID        NOT NULL REFERENCES working_papers(id) ON DELETE CASCADE,
    reviewer_role    VARCHAR(30) NOT NULL
                     CHECK (reviewer_role IN ('AUDITOR','SENIOR_AUDITOR','MANAGER','PARTNER')),
    review_status    VARCHAR(20) NOT NULL DEFAULT 'PENDING'
                     CHECK (review_status IN ('PENDING','REVIEWED','APPROVED','REJECTED')),
    review_date      TIMESTAMPTZ,
    reviewed_by      UUID,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wp_reviews_wp_id ON working_paper_reviews (working_paper_id, reviewer_role);

CREATE TABLE working_paper_comments (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id        UUID        NOT NULL REFERENCES working_paper_reviews(id) ON DELETE CASCADE,
    comment_text     TEXT        NOT NULL,
    issue_status     VARCHAR(20) NOT NULL DEFAULT 'OPEN'
                     CHECK (issue_status IN ('OPEN','RESOLVED')),
    raised_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at      TIMESTAMPTZ,
    created_by       UUID        NOT NULL
);

CREATE INDEX idx_wp_comments_review ON working_paper_comments (review_id);

CREATE TABLE audit_templates (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    template_type VARCHAR(50)  NOT NULL,
    title         VARCHAR(200) NOT NULL,
    version       VARCHAR(20)  NOT NULL DEFAULT '1.0',
    content       JSONB        NOT NULL DEFAULT '{}',
    vsa_compliant BOOLEAN      NOT NULL DEFAULT false,
    is_active     BOOLEAN      NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by    UUID         NOT NULL,
    updated_by    UUID         NOT NULL
);

CREATE INDEX idx_audit_templates_active ON audit_templates (is_active);
