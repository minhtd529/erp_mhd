package handler

import (
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
	"github.com/mdh/erp-audit/api/internal/global/auth/usecase"
)

// TwoFAHandler handles all TOTP 2FA endpoints.
type TwoFAHandler struct {
	enable       *usecase.Enable2FAUseCase
	verifySetup  *usecase.VerifySetupUseCase
	disable      *usecase.Disable2FAUseCase
	verifyLogin  *usecase.Verify2FALoginUseCase
	verifyBackup *usecase.VerifyBackupCodeUseCase
	regenBackup  *usecase.RegenBackupCodesUseCase
}

func NewTwoFAHandler(
	enable *usecase.Enable2FAUseCase,
	verifySetup *usecase.VerifySetupUseCase,
	disable *usecase.Disable2FAUseCase,
	verifyLogin *usecase.Verify2FALoginUseCase,
	verifyBackup *usecase.VerifyBackupCodeUseCase,
	regenBackup *usecase.RegenBackupCodesUseCase,
) *TwoFAHandler {
	return &TwoFAHandler{
		enable:       enable,
		verifySetup:  verifySetup,
		disable:      disable,
		verifyLogin:  verifyLogin,
		verifyBackup: verifyBackup,
		regenBackup:  regenBackup,
	}
}

// Setup starts the 2FA enrollment process: generates a TOTP key + backup codes.
// GET /api/v1/auth/2fa/setup
func (h *TwoFAHandler) Setup(c *gin.Context) {
	userID := callerUserID(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED", "message": "authentication required"})
		return
	}

	resp, err := h.enable.Execute(c.Request.Context(), *userID)
	if err != nil {
		handle2FAError(c, err)
		return
	}

	// QR code PNG is returned as base64 string for JSON compatibility
	c.JSON(http.StatusOK, gin.H{
		"secret":          resp.Secret,
		"qr_code_png":     base64.StdEncoding.EncodeToString(resp.QRCodePNG),
		"backup_codes":    resp.BackupCodes,
		"remaining_codes": resp.RemainingCodes,
	})
}

// Confirm verifies the user's first TOTP code to activate 2FA.
// POST /api/v1/auth/2fa/confirm
func (h *TwoFAHandler) Confirm(c *gin.Context) {
	userID := callerUserID(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED", "message": "authentication required"})
		return
	}

	var req usecase.Verify2FASetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	if err := h.verifySetup.Execute(c.Request.Context(), *userID, req.Code, c.ClientIP()); err != nil {
		handle2FAError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA enabled successfully"})
}

// Disable removes 2FA after verifying the user's password.
// DELETE /api/v1/auth/2fa
func (h *TwoFAHandler) Disable(c *gin.Context) {
	userID := callerUserID(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED", "message": "authentication required"})
		return
	}

	var req usecase.Disable2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	if err := h.disable.Execute(c.Request.Context(), *userID, req.Password, c.ClientIP()); err != nil {
		handle2FAError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA disabled successfully"})
}

// VerifyLogin completes login by validating a TOTP code against a challenge.
// POST /api/v1/auth/2fa/verify
func (h *TwoFAHandler) VerifyLogin(c *gin.Context) {
	var req usecase.Verify2FALoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	resp, err := h.verifyLogin.Execute(
		c.Request.Context(),
		req,
		c.ClientIP(),
		c.Request.UserAgent(),
	)
	if err != nil {
		handle2FAError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// VerifyBackup completes login using a one-time backup code.
// POST /api/v1/auth/2fa/backup
func (h *TwoFAHandler) VerifyBackup(c *gin.Context) {
	var req usecase.VerifyBackupCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	resp, err := h.verifyBackup.Execute(c.Request.Context(), req, c.ClientIP())
	if err != nil {
		handle2FAError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// RegenBackupCodes replaces all backup codes after verifying the user's password.
// POST /api/v1/auth/2fa/backup-codes/regenerate
func (h *TwoFAHandler) RegenBackupCodes(c *gin.Context) {
	userID := callerUserID(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED", "message": "authentication required"})
		return
	}

	var req usecase.RegenBackupCodesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	resp, err := h.regenBackup.Execute(c.Request.Context(), *userID, req.Password, c.ClientIP())
	if err != nil {
		handle2FAError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func handle2FAError(c *gin.Context, err error) {
	type errMap struct {
		code   string
		status int
	}
	m := map[error]errMap{
		domain.ErrTwoFAAlreadyEnabled:  {"TWO_FA_ALREADY_ENABLED", http.StatusConflict},
		domain.ErrTwoFANotEnabled:      {"TWO_FA_NOT_ENABLED", http.StatusBadRequest},
		domain.ErrTwoFAInvalid:         {"TOTP_INVALID", http.StatusUnauthorized},
		domain.ErrChallengeNotFound:    {"CHALLENGE_NOT_FOUND", http.StatusNotFound},
		domain.ErrChallengeExpired:     {"CHALLENGE_EXPIRED", http.StatusUnauthorized},
		domain.ErrChallengeInvalidated: {"CHALLENGE_INVALIDATED", http.StatusUnauthorized},
		domain.ErrTooManyAttempts:      {"TOO_MANY_ATTEMPTS", http.StatusTooManyRequests},
		domain.ErrBackupCodeInvalid:    {"BACKUP_CODE_INVALID", http.StatusUnauthorized},
		domain.ErrInvalidCredentials:   {"INVALID_CREDENTIALS", http.StatusUnauthorized},
		domain.ErrUserNotFound:         {"USER_NOT_FOUND", http.StatusNotFound},
	}

	for sentinel, v := range m {
		if errors.Is(err, sentinel) {
			c.JSON(v.status, gin.H{"error": v.code, "message": err.Error()})
			return
		}
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR", "message": "an unexpected error occurred"})
}
