package domain

import "errors"

var (
	ErrClientNotFound       = errors.New("CLIENT_NOT_FOUND")
	ErrDuplicateTaxCode     = errors.New("DUPLICATE_TAX_CODE")
	ErrInvalidStateTransition = errors.New("INVALID_STATE_TRANSITION")
)
