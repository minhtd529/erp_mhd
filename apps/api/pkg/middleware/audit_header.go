package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

// AuditIDMiddleware injects a mutable audit-ID slot into every request context
// and, after the handler returns, sets the X-Audit-ID response header when a
// mutation called audit.Logger.Log() during that request.
func AuditIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := audit.WithAuditSlot(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		if id := audit.GetID(ctx); id != uuid.Nil {
			c.Header("X-Audit-ID", id.String())
		}
	}
}
