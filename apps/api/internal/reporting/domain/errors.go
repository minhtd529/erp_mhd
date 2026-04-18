package domain

import "errors"

var (
	ErrDashboardDataStale      = errors.New("DASHBOARD_DATA_STALE")
	ErrInvalidReportFormat     = errors.New("INVALID_REPORT_FORMAT")
	ErrMVRefreshFailed         = errors.New("MV_REFRESH_FAILED")
	ErrExportGenerationFailed  = errors.New("EXPORT_GENERATION_FAILED")
)
