package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
)

func errResp(code, msg string) gin.H { return gin.H{"error": code, "message": msg} }

func mustCallerID(c *gin.Context) (uuid.UUID, bool) {
	if v, ok := c.Get(string(pkgauth.CtxUserID)); ok {
		if id, ok := v.(uuid.UUID); ok {
			return id, true
		}
	}
	c.JSON(http.StatusUnauthorized, errResp("UNAUTHORIZED", "Authentication required"))
	return uuid.Nil, false
}
