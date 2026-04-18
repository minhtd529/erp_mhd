package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdh/erp-audit/api/internal/reporting/usecase"
)

type DashboardHandler struct {
	uc *usecase.DashboardUseCase
}

func NewDashboardHandler(uc *usecase.DashboardUseCase) *DashboardHandler {
	return &DashboardHandler{uc: uc}
}

// Executive handles GET /dashboard/executive
func (h *DashboardHandler) Executive(c *gin.Context) {
	dash, err := h.uc.ExecutiveDashboard(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "Failed to build executive dashboard"))
		return
	}
	c.JSON(http.StatusOK, dash)
}

// Manager handles GET /dashboard/manager
func (h *DashboardHandler) Manager(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	dash, err := h.uc.ManagerDashboard(c.Request.Context(), callerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "Failed to build manager dashboard"))
		return
	}
	c.JSON(http.StatusOK, dash)
}

// Personal handles GET /dashboard/personal
func (h *DashboardHandler) Personal(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	dash, err := h.uc.PersonalDashboard(c.Request.Context(), callerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "Failed to build personal dashboard"))
		return
	}
	c.JSON(http.StatusOK, dash)
}
