package usecase

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/hrm/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

// ─── DTOs ─────────────────────────────────────────────────────────────────────

// SalaryHistoryItem is one row returned by GET /hrm/employees/:id/salary-history.
type SalaryHistoryItem struct {
	ID              string  `json:"id"`
	EmployeeID      string  `json:"employee_id"`
	EffectiveFrom   string  `json:"effective_from"`
	BaseSalary      string  `json:"base_salary"`           // decimal string e.g. "25000000.00"
	AllowancesTotal *string `json:"allowances_total,omitempty"`
	ChangeType      string  `json:"change_type"`
	Reason          *string `json:"reason,omitempty"`      // maps to salary_note
	CreatedByName   string  `json:"created_by_name"`
	CreatedAt       string  `json:"created_at"`
}

// CreateSalaryHistoryRequest is the body for POST /hrm/employees/:id/salary-history.
// base_salary is a decimal string (e.g. "25000000.00") so callers don't deal with
// floating-point rounding in JSON.
//
// change_type values are constrained by the DB CHECK: INITIAL, INCREASE, DECREASE,
// PROMOTION, ADJUSTMENT.
type CreateSalaryHistoryRequest struct {
	EffectiveFrom   string  `json:"effective_from"    binding:"required"`
	BaseSalary      string  `json:"base_salary"       binding:"required"`
	AllowancesTotal *string `json:"allowances_total,omitempty"`
	ChangeType      string  `json:"change_type"       binding:"required,oneof=INITIAL INCREASE DECREASE PROMOTION ADJUSTMENT"`
	Reason          *string `json:"reason,omitempty"`
}

// ─── UseCase ─────────────────────────────────────────────────────────────────

type SalaryHistoryUseCase struct {
	repo        domain.SalaryHistoryRepository
	employeeRepo domain.EmployeeRepository
	auditLog    *audit.Logger
}

func NewSalaryHistoryUseCase(
	repo domain.SalaryHistoryRepository,
	employeeRepo domain.EmployeeRepository,
	auditLog *audit.Logger,
) *SalaryHistoryUseCase {
	return &SalaryHistoryUseCase{repo: repo, employeeRepo: employeeRepo, auditLog: auditLog}
}

// hasSalaryReadPermission checks roles allowed to view salary history.
// Per permission matrix: SA, CHAIRMAN, CEO, HR_MANAGER.
func hasSalaryReadPermission(roles []string) bool {
	for _, r := range roles {
		switch r {
		case "SUPER_ADMIN", "CHAIRMAN", "CEO", "HR_MANAGER":
			return true
		}
	}
	return false
}

// hasSalaryWritePermission checks roles allowed to create salary history.
// Per permission matrix: SA, CEO, HR_MANAGER (CHAIRMAN is DENY on POST).
func hasSalaryWritePermission(roles []string) bool {
	for _, r := range roles {
		switch r {
		case "SUPER_ADMIN", "CEO", "HR_MANAGER":
			return true
		}
	}
	return false
}

// ListSalaryHistory returns all salary history records for an employee.
// Writes SALARY_HISTORY_ACCESSED audit log on every call.
func (uc *SalaryHistoryUseCase) ListSalaryHistory(
	ctx context.Context,
	employeeID uuid.UUID,
	callerID uuid.UUID,
	callerRoles []string,
	ip string,
) ([]SalaryHistoryItem, error) {
	if !hasSalaryReadPermission(callerRoles) {
		return nil, domain.ErrInsufficientPermission
	}

	// Verify employee exists.
	if _, err := uc.employeeRepo.FindByID(ctx, employeeID); err != nil {
		return nil, fmt.Errorf("salaryHistory.List: %w", err)
	}

	records, err := uc.repo.ListByEmployeeID(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("salaryHistory.List: %w", err)
	}

	if _, auditErr := uc.auditLog.Log(ctx, audit.Entry{
		UserID:     &callerID,
		Module:     "hrm",
		Resource:   "employee_salary_history",
		ResourceID: &employeeID,
		Action:     "SALARY_HISTORY_ACCESSED",
		IPAddress:  ip,
	}); auditErr != nil {
		log.Printf("ERROR salaryHistory.List audit SALARY_HISTORY_ACCESSED employee=%s caller=%s: %v",
			employeeID, callerID, auditErr)
	}

	items := make([]SalaryHistoryItem, len(records))
	for i, h := range records {
		items[i] = toSalaryHistoryItem(h)
	}
	return items, nil
}

// CreateSalaryHistory appends an immutable salary record.
// Records cannot be updated or deleted (enforced by PostgreSQL RULEs).
// Writes SALARY_HISTORY_CREATED audit log.
func (uc *SalaryHistoryUseCase) CreateSalaryHistory(
	ctx context.Context,
	employeeID uuid.UUID,
	req CreateSalaryHistoryRequest,
	callerID uuid.UUID,
	callerRoles []string,
	ip string,
) (*SalaryHistoryItem, error) {
	if !hasSalaryWritePermission(callerRoles) {
		return nil, domain.ErrInsufficientPermission
	}

	// Verify employee exists.
	if _, err := uc.employeeRepo.FindByID(ctx, employeeID); err != nil {
		return nil, fmt.Errorf("salaryHistory.Create: %w", err)
	}

	// Parse base_salary as decimal(15,2).
	baseSalaryFloat, err := strconv.ParseFloat(req.BaseSalary, 64)
	if err != nil || baseSalaryFloat <= 0 {
		return nil, fmt.Errorf("salaryHistory.Create: base_salary must be a positive decimal: %w", domain.ErrValidation)
	}

	// Parse effective_from date.
	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveFrom)
	if err != nil {
		return nil, fmt.Errorf("salaryHistory.Create: effective_from invalid date: %w", domain.ErrValidation)
	}

	// Parse optional allowances_total.
	var allowancesTotal float64
	if req.AllowancesTotal != nil {
		v, err := strconv.ParseFloat(*req.AllowancesTotal, 64)
		if err != nil {
			return nil, fmt.Errorf("salaryHistory.Create: allowances_total invalid: %w", domain.ErrValidation)
		}
		allowancesTotal = v
	}

	p := domain.CreateSalaryHistoryParams{
		EmployeeID:      employeeID,
		EffectiveDate:   effectiveDate,
		BaseSalary:      baseSalaryFloat,
		AllowancesTotal: allowancesTotal,
		SalaryNote:      req.Reason,
		ChangeType:      req.ChangeType,
		CreatedBy:       &callerID,
	}

	h, err := uc.repo.Create(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("salaryHistory.Create: %w", err)
	}

	if _, auditErr := uc.auditLog.Log(ctx, audit.Entry{
		UserID:     &callerID,
		Module:     "hrm",
		Resource:   "employee_salary_history",
		ResourceID: &h.ID,
		Action:     "SALARY_HISTORY_CREATED",
		NewValue: map[string]any{
			"employee_id":    employeeID.String(),
			"change_type":    req.ChangeType,
			"effective_from": req.EffectiveFrom,
		},
		IPAddress: ip,
	}); auditErr != nil {
		log.Printf("ERROR salaryHistory.Create audit SALARY_HISTORY_CREATED employee=%s: %v",
			employeeID, auditErr)
	}

	item := toSalaryHistoryItem(h)
	return &item, nil
}

// ─── Converters ───────────────────────────────────────────────────────────────

func formatDecimal(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}

func toSalaryHistoryItem(h *domain.SalaryHistory) SalaryHistoryItem {
	createdByName := ""
	if h.CreatedByName != nil {
		createdByName = *h.CreatedByName
	}

	var allowancesTotal *string
	if h.AllowancesTotal != 0 {
		s := formatDecimal(h.AllowancesTotal)
		allowancesTotal = &s
	}

	return SalaryHistoryItem{
		ID:              h.ID.String(),
		EmployeeID:      h.EmployeeID.String(),
		EffectiveFrom:   h.EffectiveDate.Format("2006-01-02"),
		BaseSalary:      formatDecimal(h.BaseSalary),
		AllowancesTotal: allowancesTotal,
		ChangeType:      h.ChangeType,
		Reason:          h.SalaryNote,
		CreatedByName:   createdByName,
		CreatedAt:       h.CreatedAt.Format(time.RFC3339),
	}
}
