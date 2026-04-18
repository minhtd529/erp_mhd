package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/workingpaper/domain"
	"github.com/mdh/erp-audit/api/internal/workingpaper/usecase"
)

// ReviewHandler handles working paper review endpoints.
type ReviewHandler struct {
	uc *usecase.ReviewUseCase
}

// NewReviewHandler constructs a ReviewHandler.
func NewReviewHandler(uc *usecase.ReviewUseCase) *ReviewHandler {
	return &ReviewHandler{uc: uc}
}

func (h *ReviewHandler) ListReviews(c *gin.Context) {
	wpID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid working paper ID"))
		return
	}
	reviews, err := h.uc.ListReviews(c.Request.Context(), wpID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, reviews)
}

func (h *ReviewHandler) Approve(c *gin.Context) {
	wpID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid working paper ID"))
		return
	}
	role := domain.ReviewerRole(c.Param("role"))
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.Approve(c.Request.Context(), wpID, role, caller, c.ClientIP())
	if err != nil {
		mapReviewErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ReviewHandler) RequestChanges(c *gin.Context) {
	wpID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid working paper ID"))
		return
	}
	role := domain.ReviewerRole(c.Param("role"))
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.RequestChanges(c.Request.Context(), wpID, role, caller, c.ClientIP())
	if err != nil {
		mapReviewErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ReviewHandler) AddComment(c *gin.Context) {
	wpID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid working paper ID"))
		return
	}
	role := domain.ReviewerRole(c.Param("role"))
	var req usecase.CommentAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.AddComment(c.Request.Context(), wpID, role, req, caller, c.ClientIP())
	if err != nil {
		mapReviewErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *ReviewHandler) ListComments(c *gin.Context) {
	wpID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid working paper ID"))
		return
	}
	role := domain.ReviewerRole(c.Param("role"))
	comments, err := h.uc.ListComments(c.Request.Context(), wpID, role)
	if err != nil {
		mapReviewErr(c, err)
		return
	}
	c.JSON(http.StatusOK, comments)
}

func (h *ReviewHandler) ResolveComment(c *gin.Context) {
	commentID, err := uuid.Parse(c.Param("comment_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid comment ID"))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.ResolveComment(c.Request.Context(), commentID, caller, c.ClientIP())
	if err != nil {
		mapReviewErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func mapReviewErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrReviewNotFound):
		c.JSON(http.StatusNotFound, errResp("REVIEW_NOT_FOUND", "Review not found"))
	case errors.Is(err, domain.ErrCommentNotFound):
		c.JSON(http.StatusNotFound, errResp("COMMENT_NOT_FOUND", "Comment not found"))
	case errors.Is(err, domain.ErrWorkingPaperNotFound):
		c.JSON(http.StatusNotFound, errResp("WORKING_PAPER_NOT_FOUND", "Working paper not found"))
	case errors.Is(err, domain.ErrInvalidStateTransition):
		c.JSON(http.StatusUnprocessableEntity, errResp("INVALID_STATE_TRANSITION", "Working paper is not in review"))
	default:
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}
