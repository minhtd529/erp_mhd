package domain

import "errors"

var (
	ErrTaxDeadlineNotFound      = errors.New("TAX_DEADLINE_NOT_FOUND")
	ErrAdvisoryRecordNotFound   = errors.New("ADVISORY_RECORD_NOT_FOUND")
	ErrInvalidStateTransition   = errors.New("INVALID_STATE_TRANSITION")
	ErrFiscalYearNotConfigured  = errors.New("FISCAL_YEAR_NOT_CONFIGURED")
	ErrAdvisoryNotDeliverable   = errors.New("ADVISORY_NOT_DELIVERABLE")
)
