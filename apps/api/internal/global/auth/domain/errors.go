package domain

import "errors"

// Sentinel errors — codes match the API error spec (UPPER_SNAKE_CASE).
var (
	ErrInvalidCredentials = errors.New("INVALID_CREDENTIALS")
	ErrUserNotFound       = errors.New("USER_NOT_FOUND")
	ErrUserLocked         = errors.New("USER_LOCKED")
	ErrUserInactive       = errors.New("USER_INACTIVE")
	ErrUserAlreadyExists  = errors.New("USER_ALREADY_EXISTS")
	ErrRoleNotFound       = errors.New("ROLE_NOT_FOUND")
	ErrTokenInvalid       = errors.New("TOKEN_INVALID")
	ErrTokenExpired       = errors.New("TOKEN_EXPIRED")
	ErrInsufficientPerms  = errors.New("INSUFFICIENT_PERMISSIONS")

	// 2FA errors (Phase 1.2)
	ErrTwoFARequired        = errors.New("TWO_FA_REQUIRED")
	ErrTwoFAInvalid         = errors.New("TOTP_INVALID")
	ErrTwoFAAlreadyEnabled  = errors.New("TWO_FA_ALREADY_ENABLED")
	ErrTwoFANotEnabled      = errors.New("TWO_FA_NOT_ENABLED")
	ErrChallengeNotFound    = errors.New("CHALLENGE_NOT_FOUND")
	ErrChallengeExpired     = errors.New("CHALLENGE_EXPIRED")
	ErrChallengeInvalidated = errors.New("CHALLENGE_INVALIDATED")
	ErrTooManyAttempts      = errors.New("TOO_MANY_ATTEMPTS")
	ErrBackupCodeInvalid    = errors.New("BACKUP_CODE_INVALID")
	ErrAccountLocked        = errors.New("ACCOUNT_LOCKED")
)
