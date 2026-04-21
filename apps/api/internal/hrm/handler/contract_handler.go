package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/hrm/domain"
	"github.com/mdh/erp-audit/api/internal/hrm/usecase"
)

// ContractHandler handles /api/v1/hrm/employees/:id/contracts endpoints.
type ContractHandler struct {
	uc *usecase.ContractUseCase
}

func NewContractHandler(uc *usecase.ContractUseCase) *ContractHandler {
	return &ContractHandler{uc: uc}
}

// ListContracts handles GET /api/v1/hrm/employees/:id/contracts.
func (h *ContractHandler) ListContracts(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid employee ID"))
		return
	}

	items, err := h.uc.ListContracts(c.Request.Context(), employeeID)
	if err != nil {
		log.Printf("ERROR hrm.ListContracts: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// CreateContract handles POST /api/v1/hrm/employees/:id/contracts.
func (h *ContractHandler) CreateContract(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid employee ID"))
		return
	}

	var req usecase.CreateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.uc.CreateContract(c.Request.Context(), employeeID, req, &caller, c.ClientIP())
	if err != nil {
		h.handleContractErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": resp})
}

// UpdateContract handles PUT /api/v1/hrm/employees/:id/contracts/:cid.
func (h *ContractHandler) UpdateContract(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid employee ID"))
		return
	}
	cid, err := uuid.Parse(c.Param("cid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid contract ID"))
		return
	}

	var req usecase.UpdateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.uc.UpdateContract(c.Request.Context(), cid, employeeID, req, &caller, c.ClientIP())
	if err != nil {
		h.handleContractErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// TerminateContract handles POST /api/v1/hrm/employees/:id/contracts/:cid/terminate.
func (h *ContractHandler) TerminateContract(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid employee ID"))
		return
	}
	cid, err := uuid.Parse(c.Param("cid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid contract ID"))
		return
	}

	if err := h.uc.TerminateContract(c.Request.Context(), cid, employeeID, &caller, c.ClientIP()); err != nil {
		h.handleContractErr(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *ContractHandler) handleContractErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrContractNotFound):
		c.JSON(http.StatusNotFound, errResp("CONTRACT_NOT_FOUND", "Contract not found"))
	case errors.Is(err, domain.ErrContractActiveExists):
		c.JSON(http.StatusConflict, errResp("CONTRACT_ACTIVE_EXISTS", "An active contract already exists for this employee"))
	default:
		log.Printf("ERROR hrm.contract: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}
