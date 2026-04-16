package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
	"github.com/mdh/erp-audit/api/internal/global/auth/usecase"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
)

// UserHandler handles /api/v1/users/* endpoints.
type UserHandler struct {
	createUser *usecase.CreateUserUseCase
	assignRole *usecase.AssignRoleUseCase
	listUsers  *usecase.ListUsersUseCase
	updateUser *usecase.UpdateUserUseCase
	deleteUser *usecase.DeleteUserUseCase
}

// NewUserHandler constructs a UserHandler.
func NewUserHandler(
	createUser *usecase.CreateUserUseCase,
	assignRole *usecase.AssignRoleUseCase,
	listUsers *usecase.ListUsersUseCase,
	updateUser *usecase.UpdateUserUseCase,
	deleteUser *usecase.DeleteUserUseCase,
) *UserHandler {
	return &UserHandler{
		createUser: createUser,
		assignRole: assignRole,
		listUsers:  listUsers,
		updateUser: updateUser,
		deleteUser: deleteUser,
	}
}

// ListUsers handles GET /api/v1/users.
func (h *UserHandler) ListUsers(c *gin.Context) {
	var req usecase.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", err.Error()))
		return
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 20
	}

	result, err := h.listUsers.Execute(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetUser handles GET /api/v1/users/:id.
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_ID", "Invalid user ID"))
		return
	}

	u, err := h.listUsers.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleUserError(c, err)
		return
	}
	c.JSON(http.StatusOK, u)
}

// CreateUser handles POST /api/v1/users.
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req usecase.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	callerID := callerUserID(c)
	resp, err := h.createUser.Execute(c.Request.Context(), req, callerID, c.ClientIP())
	if err != nil {
		h.handleUserError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// UpdateUser handles PUT /api/v1/users/:id.
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_ID", "Invalid user ID"))
		return
	}

	var req usecase.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	if err := h.updateUser.Execute(c.Request.Context(), id, req, callerUserID(c), c.ClientIP()); err != nil {
		h.handleUserError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user updated"})
}

// DeleteUser handles DELETE /api/v1/users/:id.
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_ID", "Invalid user ID"))
		return
	}

	if err := h.deleteUser.Execute(c.Request.Context(), id, callerUserID(c), c.ClientIP()); err != nil {
		h.handleUserError(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// AssignRole handles POST /api/v1/users/:id/roles.
func (h *UserHandler) AssignRole(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_ID", "Invalid user ID"))
		return
	}

	var req usecase.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	callerID := callerUserID(c)
	if err := h.assignRole.Execute(c.Request.Context(), userID, req.RoleCode, callerID, c.ClientIP()); err != nil {
		h.handleUserError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role assigned"})
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func (h *UserHandler) handleUserError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		c.JSON(http.StatusNotFound, errorResponse("USER_NOT_FOUND", "User not found"))
	case errors.Is(err, domain.ErrUserAlreadyExists):
		c.JSON(http.StatusConflict, errorResponse("USER_ALREADY_EXISTS", "Email already registered"))
	case errors.Is(err, domain.ErrRoleNotFound):
		c.JSON(http.StatusBadRequest, errorResponse("ROLE_NOT_FOUND", "Role does not exist"))
	default:
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "An internal error occurred"))
	}
}

// callerUserID extracts the authenticated user's ID from gin context (set by AuthMiddleware).
// Returns nil if the endpoint is public.
func callerUserID(c *gin.Context) *uuid.UUID {
	raw, exists := c.Get(pkgauth.CtxUserID)
	if !exists {
		return nil
	}
	id, ok := raw.(uuid.UUID)
	if !ok {
		return nil
	}
	return &id
}
