package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("USER_NOT_FOUND")
	ErrUserAlreadyExists  = errors.New("USER_ALREADY_EXISTS")
	ErrInvalidCredentials = errors.New("INVALID_CREDENTIALS")
	ErrUserLocked         = errors.New("USER_LOCKED")
	ErrInvalidOTP         = errors.New("INVALID_OTP")
	ErrTwoFARequired      = errors.New("TWO_FA_REQUIRED")
)
