package handler

import (
	"net"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
)

func errResp(code, msg string) gin.H { return gin.H{"error": code, "message": msg} }

func callerID(c *gin.Context) *uuid.UUID {
	if v, ok := c.Get(string(pkgauth.CtxUserID)); ok {
		if id, ok := v.(uuid.UUID); ok {
			return &id
		}
	}
	return nil
}

func mustCallerID(c *gin.Context) (uuid.UUID, bool) {
	id := callerID(c)
	if id == nil {
		c.JSON(401, errResp("UNAUTHORIZED", "Authentication required"))
		return uuid.UUID{}, false
	}
	return *id, true
}

func clientIP(c *gin.Context) net.IP {
	return net.ParseIP(c.ClientIP())
}
