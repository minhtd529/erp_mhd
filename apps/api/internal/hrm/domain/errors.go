package domain

import "errors"

var (
	ErrEmployeeNotFound     = errors.New("EMPLOYEE_NOT_FOUND")
	ErrDuplicateEmail       = errors.New("DUPLICATE_EMAIL")
	ErrInvalidStateTransition = errors.New("INVALID_STATE_TRANSITION")
)
