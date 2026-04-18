package domain

import "errors"

var (
	ErrPlanNotFound              = errors.New("PLAN_NOT_FOUND")
	ErrPlanCodeConflict          = errors.New("PLAN_CODE_CONFLICT")
	ErrPlanInactive              = errors.New("PLAN_INACTIVE")
	ErrEngCommissionNotFound     = errors.New("ENG_COMMISSION_NOT_FOUND")
	ErrEngCommissionRateExceeds  = errors.New("ENG_COMMISSION_RATE_EXCEEDS_100")
	ErrCommissionRecordNotFound  = errors.New("COMMISSION_RECORD_NOT_FOUND")
	ErrCommissionRecordImmutable = errors.New("COMMISSION_RECORD_IMMUTABLE")
	ErrRecordNotApprovable       = errors.New("RECORD_NOT_APPROVABLE")
	ErrRecordNotPayable          = errors.New("RECORD_NOT_PAYABLE")
	ErrDuplicateAccrual          = errors.New("DUPLICATE_ACCRUAL")
)
