package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/hrm/domain"
	"github.com/mdh/erp-audit/api/internal/hrm/usecase"
)

// ProfessionalHandler handles certifications, training courses/records,
// and CPE requirements endpoints (SPEC §13.10, §13.11).
type ProfessionalHandler struct {
	uc *usecase.ProfessionalUseCase
}

func NewProfessionalHandler(uc *usecase.ProfessionalUseCase) *ProfessionalHandler {
	return &ProfessionalHandler{uc: uc}
}

// handleProfErr maps domain errors to HTTP responses.
func (h *ProfessionalHandler) handleProfErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrCertificationNotFound):
		c.JSON(http.StatusNotFound, errResp("CERTIFICATION_NOT_FOUND", "Certification not found"))
	case errors.Is(err, domain.ErrTrainingCourseNotFound):
		c.JSON(http.StatusNotFound, errResp("TRAINING_COURSE_NOT_FOUND", "Training course not found"))
	case errors.Is(err, domain.ErrTrainingRecordNotFound):
		c.JSON(http.StatusNotFound, errResp("TRAINING_RECORD_NOT_FOUND", "Training record not found"))
	case errors.Is(err, domain.ErrCPERequirementNotFound):
		c.JSON(http.StatusNotFound, errResp("CPE_REQUIREMENT_NOT_FOUND", "CPE requirement not found"))
	case errors.Is(err, domain.ErrDuplicateCourseCode):
		c.JSON(http.StatusConflict, errResp("DUPLICATE_COURSE_CODE", "A course with this code already exists"))
	case errors.Is(err, domain.ErrDuplicateCPERequirement):
		c.JSON(http.StatusConflict, errResp("DUPLICATE_CPE_REQUIREMENT", "CPE requirement for this role/year already exists"))
	case errors.Is(err, domain.ErrInsufficientPermission):
		c.JSON(http.StatusForbidden, errResp("FORBIDDEN", "Insufficient permission"))
	case errors.Is(err, domain.ErrEmployeeNotFound):
		c.JSON(http.StatusNotFound, errResp("EMPLOYEE_NOT_FOUND", "Employee not found"))
	case errors.Is(err, domain.ErrValidation):
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
	default:
		log.Printf("ERROR hrm.professional: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}

// ─── Certifications ────────────────────────────────────────────────────────────

// ListCertifications handles GET /hrm/employees/:id/certifications.
func (h *ProfessionalHandler) ListCertifications(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	empID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid employee ID"))
		return
	}
	roles := callerRoles(c)
	result, err := h.uc.ListCertifications(c.Request.Context(), empID, callerID, roles)
	if err != nil {
		h.handleProfErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// CreateCertification handles POST /hrm/employees/:id/certifications.
func (h *ProfessionalHandler) CreateCertification(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	empID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid employee ID"))
		return
	}
	var req usecase.CreateCertificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.CreateCertification(c.Request.Context(), empID, req, callerID, c.ClientIP())
	if err != nil {
		h.handleProfErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": resp})
}

// GetCertification handles GET /hrm/certifications/:id.
func (h *ProfessionalHandler) GetCertification(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid certification ID"))
		return
	}
	roles := callerRoles(c)
	resp, err := h.uc.GetCertification(c.Request.Context(), id, callerID, roles)
	if err != nil {
		h.handleProfErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// UpdateCertification handles PUT /hrm/certifications/:id.
func (h *ProfessionalHandler) UpdateCertification(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid certification ID"))
		return
	}
	var req usecase.UpdateCertificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.UpdateCertification(c.Request.Context(), id, req, callerID, c.ClientIP())
	if err != nil {
		h.handleProfErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// DeleteCertification handles DELETE /hrm/certifications/:id.
func (h *ProfessionalHandler) DeleteCertification(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid certification ID"))
		return
	}
	if err := h.uc.DeleteCertification(c.Request.Context(), id, callerID, c.ClientIP()); err != nil {
		h.handleProfErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Certification deleted"})
}

// ListExpiringCertifications handles GET /hrm/certifications/expiring.
func (h *ProfessionalHandler) ListExpiringCertifications(c *gin.Context) {
	days := 90
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 {
			days = v
		}
	}
	result, err := h.uc.ListExpiringCertifications(c.Request.Context(), days)
	if err != nil {
		log.Printf("ERROR hrm.ListExpiringCertifications: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result, "days": days})
}

// ─── Training courses ─────────────────────────────────────────────────────────

// ListTrainingCourses handles GET /hrm/training-courses.
func (h *ProfessionalHandler) ListTrainingCourses(c *gin.Context) {
	var req usecase.ListTrainingCoursesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	result, err := h.uc.ListTrainingCourses(c.Request.Context(), req)
	if err != nil {
		log.Printf("ERROR hrm.ListTrainingCourses: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

// CreateTrainingCourse handles POST /hrm/training-courses.
func (h *ProfessionalHandler) CreateTrainingCourse(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	var req usecase.CreateTrainingCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.CreateTrainingCourse(c.Request.Context(), req, callerID, c.ClientIP())
	if err != nil {
		h.handleProfErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": resp})
}

// GetTrainingCourse handles GET /hrm/training-courses/:id.
func (h *ProfessionalHandler) GetTrainingCourse(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid course ID"))
		return
	}
	resp, err := h.uc.GetTrainingCourse(c.Request.Context(), id)
	if err != nil {
		h.handleProfErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// UpdateTrainingCourse handles PUT /hrm/training-courses/:id.
func (h *ProfessionalHandler) UpdateTrainingCourse(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid course ID"))
		return
	}
	var req usecase.UpdateTrainingCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.UpdateTrainingCourse(c.Request.Context(), id, req, callerID, c.ClientIP())
	if err != nil {
		h.handleProfErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// DeleteTrainingCourse handles DELETE /hrm/training-courses/:id.
func (h *ProfessionalHandler) DeleteTrainingCourse(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid course ID"))
		return
	}
	if err := h.uc.DeleteTrainingCourse(c.Request.Context(), id, callerID, c.ClientIP()); err != nil {
		h.handleProfErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Training course deleted"})
}

// ─── Training records ─────────────────────────────────────────────────────────

// ListTrainingRecords handles GET /hrm/employees/:id/training-records.
func (h *ProfessionalHandler) ListTrainingRecords(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	empID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid employee ID"))
		return
	}
	roles := callerRoles(c)
	result, err := h.uc.ListTrainingRecords(c.Request.Context(), empID, callerID, roles)
	if err != nil {
		h.handleProfErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// CreateTrainingRecord handles POST /hrm/employees/:id/training-records.
func (h *ProfessionalHandler) CreateTrainingRecord(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	empID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid employee ID"))
		return
	}
	var req usecase.CreateTrainingRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.CreateTrainingRecord(c.Request.Context(), empID, req, callerID, c.ClientIP())
	if err != nil {
		h.handleProfErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": resp})
}

// GetTrainingRecord handles GET /hrm/training-records/:id.
func (h *ProfessionalHandler) GetTrainingRecord(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid record ID"))
		return
	}
	roles := callerRoles(c)
	resp, err := h.uc.GetTrainingRecord(c.Request.Context(), id, callerID, roles)
	if err != nil {
		h.handleProfErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// UpdateTrainingRecord handles PUT /hrm/training-records/:id.
func (h *ProfessionalHandler) UpdateTrainingRecord(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid record ID"))
		return
	}
	var req usecase.UpdateTrainingRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.UpdateTrainingRecord(c.Request.Context(), id, req, callerID, c.ClientIP())
	if err != nil {
		h.handleProfErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// DeleteTrainingRecord handles DELETE /hrm/training-records/:id.
func (h *ProfessionalHandler) DeleteTrainingRecord(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid record ID"))
		return
	}
	if err := h.uc.DeleteTrainingRecord(c.Request.Context(), id, callerID, c.ClientIP()); err != nil {
		h.handleProfErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Training record deleted"})
}

// GetCPESummary handles GET /hrm/employees/:id/cpe-summary.
func (h *ProfessionalHandler) GetCPESummary(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	empID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid employee ID"))
		return
	}
	year := 0
	if y := c.Query("year"); y != "" {
		if v, err := strconv.Atoi(y); err == nil {
			year = v
		}
	}
	roles := callerRoles(c)
	resp, err := h.uc.GetCPESummary(c.Request.Context(), empID, year, callerID, roles)
	if err != nil {
		h.handleProfErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// ─── CPE requirements ─────────────────────────────────────────────────────────

// ListCPERequirements handles GET /hrm/cpe-requirements.
func (h *ProfessionalHandler) ListCPERequirements(c *gin.Context) {
	var req usecase.ListCPERequirementsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	result, err := h.uc.ListCPERequirements(c.Request.Context(), req)
	if err != nil {
		log.Printf("ERROR hrm.ListCPERequirements: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// CreateCPERequirement handles POST /hrm/cpe-requirements.
func (h *ProfessionalHandler) CreateCPERequirement(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	var req usecase.CreateCPERequirementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.CreateCPERequirement(c.Request.Context(), req, callerID, c.ClientIP())
	if err != nil {
		h.handleProfErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": resp})
}

// UpdateCPERequirement handles PUT /hrm/cpe-requirements/:id.
func (h *ProfessionalHandler) UpdateCPERequirement(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid CPE requirement ID"))
		return
	}
	var req usecase.UpdateCPERequirementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.UpdateCPERequirement(c.Request.Context(), id, req, callerID, c.ClientIP())
	if err != nil {
		h.handleProfErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}
