-- Reporting: Materialized Views for analytics dashboards

CREATE MATERIALIZED VIEW mv_revenue_by_service AS
SELECT
    COALESCE(engagement_type, 'UNKNOWN') AS service_type,
    COUNT(DISTINCT i.id)                  AS invoice_count,
    COALESCE(SUM(i.total_amount), 0)      AS total_revenue,
    COALESCE(SUM(i.tax_amount), 0)        AS total_tax
FROM invoices i
LEFT JOIN engagements e ON e.id = i.engagement_id
WHERE i.status IN ('PAID', 'PARTIALLY_PAID')
  AND i.is_void = FALSE
GROUP BY COALESCE(engagement_type, 'UNKNOWN');

CREATE UNIQUE INDEX ON mv_revenue_by_service (service_type);

CREATE MATERIALIZED VIEW mv_utilization_rate AS
SELECT
    ts.staff_id,
    DATE_TRUNC('month', te.entry_date)                           AS month,
    COALESCE(SUM(te.hours_worked), 0)                            AS total_hours,
    COALESCE(SUM(te.hours_worked) / NULLIF(160.0, 0) * 100, 0)  AS utilization_percent
FROM timesheet_entries te
JOIN timesheets ts ON ts.id = te.timesheet_id
WHERE ts.status IN ('APPROVED', 'LOCKED')
GROUP BY ts.staff_id, DATE_TRUNC('month', te.entry_date);

CREATE UNIQUE INDEX ON mv_utilization_rate (staff_id, month);

CREATE MATERIALIZED VIEW mv_ar_aging AS
SELECT
    client_id,
    COALESCE(SUM(outstanding_amount) FILTER (WHERE days_overdue = 0),  0) AS current_amount,
    COALESCE(SUM(outstanding_amount) FILTER (WHERE days_overdue BETWEEN 1  AND 30),  0) AS days_1_30,
    COALESCE(SUM(outstanding_amount) FILTER (WHERE days_overdue BETWEEN 31 AND 60),  0) AS days_31_60,
    COALESCE(SUM(outstanding_amount) FILTER (WHERE days_overdue BETWEEN 61 AND 90),  0) AS days_61_90,
    COALESCE(SUM(outstanding_amount) FILTER (WHERE days_overdue > 90), 0) AS days_over_90,
    COALESCE(SUM(outstanding_amount), 0)                                   AS total_outstanding
FROM ar_balances
GROUP BY client_id;

CREATE UNIQUE INDEX ON mv_ar_aging (client_id);

CREATE MATERIALIZED VIEW mv_engagement_progress AS
SELECT
    e.id                                                                          AS engagement_id,
    e.client_id,
    e.status,
    e.budgeted_hours,
    COALESCE(SUM(te.hours_worked), 0)                                             AS hours_logged,
    CASE
        WHEN COALESCE(e.budgeted_hours, 0) = 0 THEN 0
        ELSE ROUND(100.0 * COALESCE(SUM(te.hours_worked), 0) / e.budgeted_hours)
    END                                                                           AS completion_percent
FROM engagements e
LEFT JOIN timesheets ts ON ts.engagement_id = e.id AND ts.status IN ('APPROVED', 'LOCKED')
LEFT JOIN timesheet_entries te ON te.timesheet_id = ts.id
WHERE e.status NOT IN ('CANCELLED')
GROUP BY e.id, e.client_id, e.status, e.budgeted_hours;

CREATE UNIQUE INDEX ON mv_engagement_progress (engagement_id);

CREATE MATERIALIZED VIEW mv_commission_summary AS
SELECT
    DATE_TRUNC('month', accrued_at)                      AS month,
    COALESCE(SUM(calculated_amount), 0)                  AS total_accrued,
    COALESCE(SUM(calculated_amount) FILTER (WHERE status = 'approved'), 0) AS total_approved,
    COALESCE(SUM(payable_amount)    FILTER (WHERE status = 'paid'),     0) AS total_paid,
    COALESCE(SUM(payable_amount)    FILTER (WHERE status IN ('accrued','on_hold')), 0) AS total_pending,
    COALESCE(SUM(holdback_amount)   FILTER (WHERE status = 'on_hold'),  0) AS total_on_hold,
    COUNT(*) FILTER (WHERE is_clawback = TRUE)                             AS clawback_count
FROM commission_records
WHERE is_clawback = FALSE
GROUP BY DATE_TRUNC('month', accrued_at);

CREATE UNIQUE INDEX ON mv_commission_summary (month);

-- Track last MV refresh time
CREATE TABLE mv_refresh_log (
    view_name   VARCHAR(100) PRIMARY KEY,
    refreshed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    duration_ms  INT,
    success      BOOLEAN NOT NULL DEFAULT TRUE,
    error_msg    TEXT
);
