package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
)

func errResp(code, msg string) gin.H {
	return gin.H{"error": code, "message": msg}
}

func callerID(c *gin.Context) *uuid.UUID {
	raw, ok := c.Get(pkgauth.CtxUserID)
	if !ok {
		return nil
	}
	id, ok := raw.(uuid.UUID)
	if !ok {
		return nil
	}
	return &id
}

func mustCallerID(c *gin.Context) (uuid.UUID, bool) {
	id := callerID(c)
	if id == nil {
		c.JSON(http.StatusUnauthorized, errResp("UNAUTHORIZED", "Authentication required"))
		return uuid.UUID{}, false
	}
	return *id, true
}

func parsePageSize(c *gin.Context) (page, size int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ = strconv.Atoi(c.DefaultQuery("size", "20"))
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}
	return
}
