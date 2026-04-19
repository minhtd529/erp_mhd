package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
	"go.uber.org/zap"
)

// RequestLogger logs each HTTP request using zap.
func RequestLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		log.Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("ip", c.ClientIP()),
		)
	}
}

// CORS sets CORS headers. Tighten allowedOrigins in production.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		for _, o := range allowedOrigins {
			if o == "*" || o == origin {
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,Accept")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// AuthMiddleware validates the Bearer JWT and populates gin context with claims.
// Depends on a pkgauth.JWTService to validate tokens.
func AuthMiddleware(jwtSvc *pkgauth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "Missing or invalid Authorization header",
			})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := jwtSvc.ValidateAccessToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "TOKEN_INVALID",
				"message": "Token is invalid or expired",
			})
			return
		}

		// Populate gin context for downstream handlers
		c.Set(pkgauth.CtxClaims, claims)
		c.Set(pkgauth.CtxUserID, claims.UserID)
		c.Set(pkgauth.CtxEmail, claims.Email)
		c.Set(pkgauth.CtxRoles, claims.Roles)
		c.Set(pkgauth.CtxPerms, claims.Permissions)
		c.Set(pkgauth.CtxBranchID, claims.BranchID)
		c.Set(pkgauth.CtxDeptID, claims.DepartmentID)
		c.Set(pkgauth.CtxTwoFAVerified, claims.TwoFAVerified)

		c.Next()
	}
}

// RequireRole aborts with 403 if the caller does not have at least one of the given roles.
// Must be placed after AuthMiddleware in the handler chain.
// SUPER_ADMIN bypasses all role checks.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		userRoles, _ := c.Get(pkgauth.CtxRoles)
		for _, r := range toStringSlice(userRoles) {
			if r == "SUPER_ADMIN" || allowed[r] {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":   "INSUFFICIENT_PERMISSIONS",
			"message": "You do not have permission to perform this action",
		})
	}
}

// RequirePermission aborts with 403 if the caller does not have the given permission.
// Permission format: "module:resource:action" (e.g. "crm:client:read").
func RequirePermission(module, resource, action string) gin.HandlerFunc {
	required := module + ":" + resource + ":" + action
	return func(c *gin.Context) {
		userPerms, _ := c.Get(pkgauth.CtxPerms)
		for _, p := range toStringSlice(userPerms) {
			if p == required {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":   "INSUFFICIENT_PERMISSIONS",
			"message": "You do not have the required permission: " + required,
		})
	}
}

// Require2FA aborts with 403 if the caller holds a privileged role (FIRM_PARTNER or
// SUPER_ADMIN) but has not completed the 2FA verification step. This enforces the
// policy that high-privilege accounts must always authenticate via TOTP or push 2FA.
// Must be placed after AuthMiddleware in the handler chain.
func Require2FA() gin.HandlerFunc {
	const (
		roleFirmPartner = "FIRM_PARTNER"
		roleSuperAdmin  = "SUPER_ADMIN"
	)
	return func(c *gin.Context) {
		userRoles, _ := c.Get(pkgauth.CtxRoles)
		for _, r := range toStringSlice(userRoles) {
			if r == roleFirmPartner || r == roleSuperAdmin {
				verified, _ := c.Get(pkgauth.CtxTwoFAVerified)
				if ok, _ := verified.(bool); !ok {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
						"error":   "TWO_FA_REQUIRED",
						"message": "This endpoint requires two-factor authentication for your role",
					})
					return
				}
				break
			}
		}
		c.Next()
	}
}

// toStringSlice safely casts an any value to []string.
func toStringSlice(v any) []string {
	if s, ok := v.([]string); ok {
		return s
	}
	return nil
}
