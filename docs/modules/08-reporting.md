<!-- spec-version: 1.2 | last-sync: 2026-04-16 | changes: added Commission KPIs (dashboards), 5 commission reports, 6 API endpoints -->
> **Spec version**: 1.2 — Last sync: 2026-04-16 — Updated in v1.2

# Module 8: Reporting - Analytics & Dashboards

## Overview
Provides role-based dashboards, on-demand reports, and analytics for business intelligence.

## Bounded Context: Reporting

### Responsibilities
- Role-based dashboards (FIRM_PARTNER, AUDIT_MANAGER, AUDIT_STAFF - personal only)
- Report generation (scheduled & on-demand)
- Materialized views for analytics (refresh strategy)
- Export capabilities (Excel, PDF)

## Key Features

### 1. Dashboards
**Entity**: Dashboard view configuration (not persisted, computed from materialized views)

**Executive Dashboard** (FIRM_PARTNER only):
- Company-wide KPIs:
  - Total revenue YTD
  - Revenue by service/partner
  - Average utilization rate
  - Outstanding receivables (total + aging)
  - Pipeline: Engagements by status
  - Team member count, utilization breakdown
  - Commission KPIs (v1.2): `total_commission_accrued_month`, `total_commission_paid_month`, `total_commission_pending` (approved chưa chi), `total_commission_on_hold` (holdback), `commission_percent_of_revenue`
  - Charts (v1.2): Revenue by Salesperson (top earners), Commission by Salesperson, Monthly Commission Trend

**Manager Dashboard** (AUDIT_MANAGER):
- Team-scoped metrics:
  - Team utilization rate
  - Team revenue contribution
  - Outstanding payments from team's engagements
  - Task completion rate
  - CPE compliance status (team)

**Personal Dashboard** (AUDIT_STAFF):
- Individual metrics:
  - Personal utilization rate
  - Hours logged vs. budgeted
  - Certifications & CPE status
  - Performance review status
  - My engagements & tasks
  - Salesperson section (v1.2, hiển thị khi `is_salesperson = true`): `my_commission_ytd`, `my_commission_month`, `my_commission_pending`, `my_commission_on_hold`, `my_active_engagements`, `my_top_engagements`

### 2. Reports
**Report Types**: On-demand or scheduled
- **Revenue Report**: By service type, partner, period (filtered by role)
- **Revenue by Salesperson**: Top salespeople by revenue (Director+)
- **Utilization Report**: Staff utilization % by period
- **AR Aging**: Outstanding balance, aging bucket analysis
- **Engagement Status**: Progress tracking, budget vs. actual
- **Tax Deadline**: Upcoming/overdue tax deadlines per client
- **CPE Compliance**: Staff CPE hours vs. requirement (3-year rolling)
- **Team Performance**: KPI rankings, top performers
- **Commission Statement** (v1.2): Bảng kê hoa hồng cá nhân — Tháng/Quý/Năm; Salesperson (own), Manager+ (team); Excel, PDF
- **Commission Payout** (v1.2): Báo cáo chi hoa hồng — Tháng; Accountant, Director+; Excel, PDF
- **Commission by Service** (v1.2): Hoa hồng theo dịch vụ — Quý/Năm; Director+, Partner; Excel, PDF
- **Commission Pending** (v1.2): Hoa hồng chưa duyệt/chưa chi — Tuần; Accountant, Director+; Excel
- **Commission Clawback** (v1.2): Clawback commission — Tháng; Director+; Excel

**Formats**: JSON (API), Excel, PDF, CSV

### 3. Materialized Views (for performance)
**Refresh Strategy**: Nightly batch at 23:00 or on-demand
- `mv_revenue_by_service` (service_type, total_revenue, invoice_count)
- `mv_utilization_rate` (staff_id, period_year_month, utilization_percent)
- `mv_ar_aging` (client_id, current, days_1_30, days_31_60, ...)
- `mv_engagement_progress` (engagement_id, total_hours_budgeted, hours_logged, completion_percent)
- `mv_tax_deadline_status` (client_id, total_deadlines, completed, overdue, compliance_score)
- `mv_staff_cpe_status` (staff_id, cpe_hours_3y, compliance_status)

### 4. Distributed Locking
No concurrent write operations; reporting is read-only (queries materialized views).

## Code Structure

### Go Package Layout
```
modules/reporting/
  ├── domain/
  │   ├── dashboard.go              (Dashboard aggregate - computed)
  │   ├── report.go                 (Report aggregate - metadata)
  │   └── reporting_events.go       (ReportGenerated, DashboardViewed, etc.)
  ├── application/
  │   ├── dashboard_service.go      (DashboardService - query MV)
  │   ├── report_service.go         (ReportService - query + export)
  │   └── mv_refresh_service.go     (MaterializedViewRefreshService - Asynq job)
  ├── infrastructure/
  │   ├── postgres/
  │   │   ├── dashboard_repository.go (sqlc queries on materialized views)
  │   │   ├── report_repository.go
  │   │   └── mv_manager.go         (Refresh materialized views)
  │   └── asynq/
  │       └── mv_refresh_job.go     (Scheduled job: nightly refresh)
  └── interfaces/
      └── rest/
          ├── dashboard_handler.go   (DashboardHandler)
          ├── report_handler.go      (ReportHandler)
          └── export_handler.go      (Export to Excel/PDF)
```

## API Endpoints

**Authorization**: FIRM_PARTNER (all), AUDIT_MANAGER (team-scoped), AUDIT_STAFF (personal only)

### Dashboards
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| GET | `/api/v1/dashboard/executive` | Executive dashboard (KPIs) | FIRM_PARTNER | No |
| GET | `/api/v1/dashboard/manager` | Manager team dashboard | AUDIT_MANAGER | No |
| GET | `/api/v1/dashboard/personal` | Personal dashboard (my metrics) | AUDIT_STAFF | No |

**Response**: JSON with dashboard_type, metrics (JSONB), last_refresh_time, chart_data (pre-computed)

### Reports
| Method | Path | Description | Auth | Audit |
|--------|------|-------------|------|-------|
| GET | `/api/v1/reports/revenue` | Revenue report (PDF/Excel) | FIRM_PARTNER | No |
| GET | `/api/v1/reports/revenue?format=pdf&year=2026` | Filter params | FIRM_PARTNER | No |
| GET | `/api/v1/reports/utilization` | Utilization by staff | FIRM_PARTNER | No |
| GET | `/api/v1/reports/ar-aging` | A/R aging analysis | FIRM_PARTNER | No |
| GET | `/api/v1/reports/engagement-status` | Engagement progress | FIRM_PARTNER | No |
| GET | `/api/v1/reports/tax-deadlines` | Tax deadline status | FIRM_PARTNER | No |
| GET | `/api/v1/reports/cpe-compliance` | CPE compliance | FIRM_PARTNER | No |
| GET | `/api/v1/reports/team-performance` | Team KPI ranking | AUDIT_MANAGER | No |
| POST | `/api/v1/reports/commission-statement` | Commission statement (cá nhân/team) | AUDIT_STAFF, AUDIT_MANAGER | No |
| POST | `/api/v1/reports/commission-payout` | Chi hoa hồng export | FIRM_PARTNER | No |
| POST | `/api/v1/reports/commission-by-service` | Hoa hồng theo dịch vụ | FIRM_PARTNER | No |
| GET | `/api/v1/reports/commission-pending` | Pending commission (chưa duyệt/chưa chi) | FIRM_PARTNER | No |
| POST | `/api/v1/reports/commission-clawback` | Clawback commission report | FIRM_PARTNER | No |
| GET | `/api/v1/reports/revenue-by-salesperson` | Revenue by salesperson | FIRM_PARTNER | No |

**Query Params**: `?format=pdf|excel|json&year=&month=&office_id=&staff_id=`

**Response Headers** (for PDF/Excel):
```
Content-Type: application/pdf | application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
Content-Disposition: attachment; filename="report_revenue_2026-04.pdf"
```

## Database Tables

### Materialized Views (non-materialized in PostgreSQL, but named `mv_*`)
```sql
CREATE MATERIALIZED VIEW mv_revenue_by_service AS
  SELECT service_type, COUNT(*) as invoice_count, SUM(total_amount) as total_revenue
  FROM invoices WHERE status='PAID' GROUP BY service_type;

CREATE MATERIALIZED VIEW mv_utilization_rate AS
  SELECT ts.staff_id, DATE_TRUNC('month', te.entry_date) as month,
         SUM(te.hours_worked) / (20 * 8) as utilization_rate
  FROM timesheet_entries te JOIN timesheets ts ON ts.id=te.timesheet_id
  GROUP BY ts.staff_id, DATE_TRUNC('month', te.entry_date);

-- Other MVs similar pattern
```

### Optional: Report Metadata Table (for scheduled reports)
- `report_schedules` (id UUID, report_type ENUM, schedule_cron, recipient_emails, format ENUM, last_run_at, next_run_at, created_by, created_at)

### Indexes
- All MV columns indexed for fast access

## CQRS
**Writes**: None (Reporting is read-only)
**Reads**: Direct queries on materialized views (pre-computed, fast)
**Events**: ReportGenerated (logged for audit), DashboardViewed (optional, for tracking)

## Materialized View Refresh Strategy

**Nightly Batch Job** (Asynq): `reporting:refresh-views`
- Scheduled 23:00 PM daily
- Refreshes all materialized views in transaction
- Log refresh completion to last_refresh_time
- On failure, retry 3x with exponential backoff
- Alert on refresh failure

**On-Demand Refresh** (via API):
- `/api/v1/admin/refresh-materialized-views` (SUPER_ADMIN only)
- Manual trigger if data is stale

## Export Strategy
**Excel**: Using `github.com/xuri/excelize/v2`
- Template-based with formatting
- Include company logo, date, generated_by user

**PDF**: Using `github.com/go-pdf/fpdf` or similar
- Professional layout with charts (image embeds)
- Numbered pages, footer with "Confidential"

## Error Codes
`DASHBOARD_DATA_STALE` - MV last refresh > 24h
`REPORT_NOT_FOUND`
`INVALID_REPORT_FORMAT`
`INSUFFICIENT_PERMISSIONS` - Cannot view manager dashboard as staff
`MV_REFRESH_FAILED` - Last refresh failed
`EXPORT_GENERATION_FAILED`