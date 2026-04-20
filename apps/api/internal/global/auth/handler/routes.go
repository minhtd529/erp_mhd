package handler

import (
	"github.com/gin-gonic/gin"
	mw "github.com/mdh/erp-audit/api/pkg/middleware"
)

// RegisterRoutes wires all auth + user management routes under the given /api/v1 group.
func RegisterRoutes(
	v1 *gin.RouterGroup,
	auth *AuthHandler,
	users *UserHandler,
	twoFA *TwoFAHandler,
	pushH *PushHandler,
	auditH *AuditHandler,
	authMW gin.HandlerFunc,
) {
	// ─── Public auth endpoints ────────────────────────────────────────────────
	a := v1.Group("/auth")
	{
		a.POST("/login", auth.Login)
		a.POST("/refresh", auth.Refresh)
		a.POST("/logout", authMW, auth.Logout)

		// 2FA — verify endpoints are public (challenge_id acts as the token)
		a.POST("/2fa/verify", twoFA.VerifyLogin)
		a.POST("/2fa/backup", twoFA.VerifyBackup)

		// 2FA management — requires an active session
		a.GET("/2fa/setup", authMW, twoFA.Setup)
		a.POST("/2fa/confirm", authMW, twoFA.Confirm)
		a.DELETE("/2fa", authMW, twoFA.Disable)
		a.POST("/2fa/backup-codes/regenerate", authMW, twoFA.RegenBackupCodes)

		// Push 2FA (mobile app response + browser polling)
		a.POST("/2fa/push-response", pushH.PushResponse)
		a.GET("/2fa/push-status/:challengeId", pushH.PushStatus)
		a.POST("/2fa/resend-push", pushH.ResendPush)
	}

	// Push device management (authenticated)
	pd := v1.Group("/push", authMW)
	{
		pd.POST("/devices/register", pushH.RegisterDevice)
		pd.DELETE("/devices", pushH.UnregisterDevice)
		pd.GET("/devices", pushH.ListDevices)
		pd.POST("/devices/heartbeat", pushH.Heartbeat)
	}

	// Push relay WebSocket (device-token authenticated, not JWT)
	v1.GET("/ws/push", pushH.PushRelayWS)

	// ─── Authenticated user endpoints ────────────────────────────────────────
	v1.GET("/me", authMW, auth.Me)

	// ─── User management (SUPER_ADMIN, FIRM_PARTNER) ─────────────────────────
	u := v1.Group("/users", authMW, mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER"))
	{
		u.GET("", users.ListUsers)
		u.POST("", users.CreateUser)
		u.GET("/:id", users.GetUser)
		u.PUT("/:id", users.UpdateUser)
		u.DELETE("/:id", users.DeleteUser)
		u.POST("/:id/roles", users.AssignRole)
	}

	// ─── Audit logs (SUPER_ADMIN only) ────────────────────────────────────────
	v1.GET("/audit-logs", authMW, mw.RequireRole("SUPER_ADMIN"), auditH.List)
}
