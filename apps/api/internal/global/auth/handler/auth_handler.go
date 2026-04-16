package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
	"github.com/mdh/erp-audit/api/internal/global/auth/usecase"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
)

// AuthHandler handles /api/v1/auth/* endpoints.
type AuthHandler struct {
	login   *usecase.LoginUseCase
	refresh *usecase.RefreshTokenUseCase
	logout  *usecase.LogoutUseCase
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(
	login *usecase.LoginUseCase,
	refresh *usecase.RefreshTokenUseCase,
	logout *usecase.LogoutUseCase,
) *AuthHandler {
	return &AuthHandler{login: login, refresh: refresh, logout: logout}
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req usecase.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.login.Execute(c.Request.Context(), req, c.ClientIP())
	if err != nil {
		// 2FA required — respond 202 Accepted with challenge details
		if errors.Is(err, domain.ErrTwoFARequired) && resp != nil {
			c.JSON(http.StatusAccepted, resp)
			return
		}
		h.handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Refresh handles POST /api/v1/auth/refresh.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req usecase.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.refresh.Execute(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Logout handles POST /api/v1/auth/logout.
func (h *AuthHandler) Logout(c *gin.Context) {
	var req usecase.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	if err := h.logout.Execute(c.Request.Context(), req.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "logout failed"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// Me handles GET /api/v1/me — returns the current user's token claims.
func (h *AuthHandler) Me(c *gin.Context) {
	claimsRaw, exists := c.Get(pkgauth.CtxClaims)
	if !exists {
		c.JSON(http.StatusUnauthorized, errorResponse("UNAUTHORIZED", "not authenticated"))
		return
	}
	claims, ok := claimsRaw.(*pkgauth.TokenClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "malformed claims"))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":            claims.UserID,
		"email":         claims.Email,
		"roles":         claims.Roles,
		"permissions":   claims.Permissions,
		"branch_id":     claims.BranchID,
		"department_id": claims.DepartmentID,
	})
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func (h *AuthHandler) handleAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, errorResponse("INVALID_CREDENTIALS", "Invalid email or password"))
	case errors.Is(err, domain.ErrUserLocked):
		c.JSON(http.StatusForbidden, errorResponse("USER_LOCKED", "Account is locked"))
	case errors.Is(err, domain.ErrUserInactive):
		c.JSON(http.StatusForbidden, errorResponse("USER_INACTIVE", "Account is inactive"))
	case errors.Is(err, domain.ErrAccountLocked):
		c.JSON(http.StatusTooManyRequests, errorResponse("ACCOUNT_LOCKED", "Account temporarily locked due to too many failed attempts"))
	case errors.Is(err, domain.ErrTokenInvalid):
		c.JSON(http.StatusUnauthorized, errorResponse("TOKEN_INVALID", "Token is invalid or has been revoked"))
	case errors.Is(err, domain.ErrTokenExpired):
		c.JSON(http.StatusUnauthorized, errorResponse("TOKEN_EXPIRED", "Token has expired"))
	default:
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "An internal error occurred"))
	}
}

func errorResponse(code, message string) gin.H {
	return gin.H{"error": code, "message": message}
}
