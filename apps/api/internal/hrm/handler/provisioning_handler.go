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

// ProvisioningHandler handles /api/v1/hrm/user-provisioning-requests/* endpoints.
type ProvisioningHandler struct {
	uc *usecase.ProvisioningUseCase
}

func NewProvisioningHandler(uc *usecase.ProvisioningUseCase) *ProvisioningHandler {
	return &ProvisioningHandler{uc: uc}
}

// ListProvisioningRequests handles GET /api/v1/hrm/user-provisioning-requests.
func (h *ProvisioningHandler) ListProvisioningRequests(c *gin.Context) {
	var req usecase.ListProvisioningRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	result, err := h.uc.ListRequests(c.Request.Context(), req)
	if err != nil {
		log.Printf("ERROR hrm.ListProvisioningRequests: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

// CreateProvisioningRequest handles POST /api/v1/hrm/user-provisioning-requests.
func (h *ProvisioningHandler) CreateProvisioningRequest(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}

	var req usecase.CreateProvisioningRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.uc.CreateRequest(c.Request.Context(), req, callerID, c.ClientIP())
	if err != nil {
		h.handleProvisioningErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": resp})
}

// GetProvisioningRequest handles GET /api/v1/hrm/user-provisioning-requests/:id.
func (h *ProvisioningHandler) GetProvisioningRequest(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid request ID"))
		return
	}

	resp, err := h.uc.GetRequest(c.Request.Context(), id)
	if err != nil {
		h.handleProvisioningErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// BranchApprove handles POST /api/v1/hrm/user-provisioning-requests/:id/branch-approve.
func (h *ProvisioningHandler) BranchApprove(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid request ID"))
		return
	}

	branchID := callerBranchID(c)
	resp, err := h.uc.BranchApprove(c.Request.Context(), id, callerID, branchID, c.ClientIP())
	if err != nil {
		h.handleProvisioningErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// BranchReject handles POST /api/v1/hrm/user-provisioning-requests/:id/branch-reject.
func (h *ProvisioningHandler) BranchReject(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid request ID"))
		return
	}

	var req usecase.RejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	branchID := callerBranchID(c)
	resp, err := h.uc.BranchReject(c.Request.Context(), id, callerID, branchID, req.Reason, c.ClientIP())
	if err != nil {
		h.handleProvisioningErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// HRApprove handles POST /api/v1/hrm/user-provisioning-requests/:id/hr-approve.
func (h *ProvisioningHandler) HRApprove(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid request ID"))
		return
	}

	resp, err := h.uc.HRApprove(c.Request.Context(), id, callerID, c.ClientIP())
	if err != nil {
		h.handleProvisioningErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// HRReject handles POST /api/v1/hrm/user-provisioning-requests/:id/hr-reject.
func (h *ProvisioningHandler) HRReject(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid request ID"))
		return
	}

	var req usecase.RejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.uc.HRReject(c.Request.Context(), id, callerID, req.Reason, c.ClientIP())
	if err != nil {
		h.handleProvisioningErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// ExecuteProvisioning handles POST /api/v1/hrm/user-provisioning-requests/:id/execute.
// The employee's full_name and email must be provided in the request body for user creation.
func (h *ProvisioningHandler) ExecuteProvisioning(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid request ID"))
		return
	}

	var body struct {
		FullName string `json:"full_name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.uc.Execute(c.Request.Context(), id, callerID, body.FullName, body.Email, c.ClientIP())
	if err != nil {
		h.handleProvisioningErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// CancelProvisioningRequest handles POST /api/v1/hrm/user-provisioning-requests/:id/cancel.
func (h *ProvisioningHandler) CancelProvisioningRequest(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid request ID"))
		return
	}

	roles := callerRoles(c)
	resp, err := h.uc.CancelRequest(c.Request.Context(), id, callerID, roles, c.ClientIP())
	if err != nil {
		h.handleProvisioningErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// ─── Offboarding endpoints ────────────────────────────────────────────────────

// ListOffboarding handles GET /api/v1/hrm/offboarding.
func (h *ProvisioningHandler) ListOffboarding(c *gin.Context) {
	var req usecase.ListOffboardingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	result, err := h.uc.ListOffboarding(c.Request.Context(), req)
	if err != nil {
		log.Printf("ERROR hrm.ListOffboarding: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

// CreateOffboarding handles POST /api/v1/hrm/offboarding.
func (h *ProvisioningHandler) CreateOffboarding(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}

	var req usecase.CreateOffboardingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.uc.CreateOffboarding(c.Request.Context(), req, callerID, c.ClientIP())
	if err != nil {
		h.handleOffboardingErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": resp})
}

// GetOffboarding handles GET /api/v1/hrm/offboarding/:id.
func (h *ProvisioningHandler) GetOffboarding(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid checklist ID"))
		return
	}

	resp, err := h.uc.GetOffboarding(c.Request.Context(), id)
	if err != nil {
		h.handleOffboardingErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// UpdateOffboardingItem handles PUT /api/v1/hrm/offboarding/:id/items/:key.
func (h *ProvisioningHandler) UpdateOffboardingItem(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid checklist ID"))
		return
	}
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, errResp("INVALID_KEY", "Item key is required"))
		return
	}

	var req usecase.UpdateOffboardingItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.uc.UpdateOffboardingItem(c.Request.Context(), id, key, req, callerID, c.ClientIP())
	if err != nil {
		h.handleOffboardingErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// CompleteOffboarding handles POST /api/v1/hrm/offboarding/:id/complete.
func (h *ProvisioningHandler) CompleteOffboarding(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid checklist ID"))
		return
	}

	resp, err := h.uc.CompleteOffboarding(c.Request.Context(), id, callerID, c.ClientIP())
	if err != nil {
		h.handleOffboardingErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// ─── error mappers ────────────────────────────────────────────────────────────

func (h *ProvisioningHandler) handleProvisioningErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrRequestNotFound):
		c.JSON(http.StatusNotFound, errResp("REQUEST_NOT_FOUND", "Provisioning request not found"))
	case errors.Is(err, domain.ErrDuplicatePendingRequest):
		c.JSON(http.StatusConflict, errResp("DUPLICATE_PENDING_REQUEST", "A pending request already exists for this employee"))
	case errors.Is(err, domain.ErrInvalidRoleForProvisioning):
		c.JSON(http.StatusBadRequest, errResp("INVALID_ROLE_FOR_PROVISIONING", "This role cannot be provisioned through this workflow"))
	case errors.Is(err, domain.ErrRequestExpired):
		c.JSON(http.StatusUnprocessableEntity, errResp("REQUEST_EXPIRED", "This provisioning request has expired"))
	case errors.Is(err, domain.ErrRequestAlreadyExecuted):
		c.JSON(http.StatusConflict, errResp("REQUEST_ALREADY_EXECUTED", "This request has already been executed"))
	case errors.Is(err, domain.ErrInvalidRequestStatus):
		c.JSON(http.StatusUnprocessableEntity, errResp("INVALID_REQUEST_STATUS", "Request is not in a valid state for this action"))
	case errors.Is(err, domain.ErrInsufficientPermission):
		c.JSON(http.StatusForbidden, errResp("INSUFFICIENT_PERMISSION", "You do not have permission to perform this action"))
	case errors.Is(err, domain.ErrValidation):
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", "Invalid input"))
	default:
		log.Printf("ERROR hrm.provisioning: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}

func (h *ProvisioningHandler) handleOffboardingErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrOffboardingNotFound):
		c.JSON(http.StatusNotFound, errResp("OFFBOARDING_NOT_FOUND", "Offboarding checklist not found"))
	case errors.Is(err, domain.ErrInvalidRequestStatus):
		c.JSON(http.StatusUnprocessableEntity, errResp("INVALID_STATUS", "Checklist is not in a valid state for this action"))
	case errors.Is(err, domain.ErrInsufficientPermission):
		c.JSON(http.StatusForbidden, errResp("INSUFFICIENT_PERMISSION", "You do not have permission to perform this action"))
	case errors.Is(err, domain.ErrValidation):
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", "Invalid input"))
	default:
		log.Printf("ERROR hrm.offboarding: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}
