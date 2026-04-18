package usecase

import (
	"context"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/tax/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

type TaxDeadlineUseCase struct {
	repo     domain.DeadlineRepository
	auditLog *audit.Logger
}

func NewTaxDeadlineUseCase(repo domain.DeadlineRepository, auditLog *audit.Logger) *TaxDeadlineUseCase {
	return &TaxDeadlineUseCase{repo: repo, auditLog: auditLog}
}

type CreateDeadlineRequest struct {
	DeadlineType           domain.DeadlineType `json:"deadline_type" binding:"required"`
	DeadlineName           string              `json:"deadline_name" binding:"required"`
	DueDate                time.Time           `json:"due_date" binding:"required"`
	ExpectedSubmissionDate *time.Time          `json:"expected_submission_date"`
	Notes                  string              `json:"notes"`
}

type UpdateDeadlineRequest struct {
	DeadlineName           string     `json:"deadline_name" binding:"required"`
	DueDate                time.Time  `json:"due_date" binding:"required"`
	ExpectedSubmissionDate *time.Time `json:"expected_submission_date"`
	Notes                  string     `json:"notes"`
}

type MarkCompletedRequest struct {
	ActualDate time.Time `json:"actual_date" binding:"required"`
}

func (uc *TaxDeadlineUseCase) Create(ctx context.Context, clientID uuid.UUID, req CreateDeadlineRequest, callerID uuid.UUID, ip net.IP) (*domain.TaxDeadline, error) {
	d, err := uc.repo.Create(ctx, domain.CreateDeadlineParams{
		ClientID:               clientID,
		DeadlineType:           req.DeadlineType,
		DeadlineName:           req.DeadlineName,
		DueDate:                req.DueDate,
		ExpectedSubmissionDate: req.ExpectedSubmissionDate,
		Notes:                  req.Notes,
		CreatedBy:              callerID,
	})
	if err != nil {
		return nil, err
	}
	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, IPAddress: ip.String(),
		Module: "tax", Resource: "tax_deadline", ResourceID: &d.ID,
		Action: "CREATE", NewValue: d,
	})
	return d, nil
}

func (uc *TaxDeadlineUseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.TaxDeadline, error) {
	return uc.repo.FindByID(ctx, id)
}

func (uc *TaxDeadlineUseCase) List(ctx context.Context, clientID uuid.UUID, status domain.DeadlineStatus, page, size int) (pagination.OffsetResult[domain.TaxDeadline], error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	f := domain.ListDeadlinesFilter{ClientID: &clientID, Status: status}
	list, total, err := uc.repo.List(ctx, f, page, size)
	if err != nil {
		return pagination.OffsetResult[domain.TaxDeadline]{}, err
	}
	items := make([]domain.TaxDeadline, len(list))
	for i, d := range list {
		items[i] = *d
	}
	return pagination.NewOffsetResult(items, total, page, size), nil
}

func (uc *TaxDeadlineUseCase) Update(ctx context.Context, id uuid.UUID, req UpdateDeadlineRequest, callerID uuid.UUID, ip net.IP) (*domain.TaxDeadline, error) {
	d, err := uc.repo.Update(ctx, domain.UpdateDeadlineParams{
		ID:                     id,
		DeadlineName:           req.DeadlineName,
		DueDate:                req.DueDate,
		ExpectedSubmissionDate: req.ExpectedSubmissionDate,
		Notes:                  req.Notes,
		UpdatedBy:              callerID,
	})
	if err != nil {
		return nil, err
	}
	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, IPAddress: ip.String(),
		Module: "tax", Resource: "tax_deadline", ResourceID: &d.ID,
		Action: "UPDATE", NewValue: d,
	})
	return d, nil
}

func (uc *TaxDeadlineUseCase) MarkCompleted(ctx context.Context, id uuid.UUID, req MarkCompletedRequest, callerID uuid.UUID, ip net.IP) (*domain.TaxDeadline, error) {
	existing, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing.Status == domain.DeadlineStatusCompleted {
		return nil, domain.ErrInvalidStateTransition
	}

	d, err := uc.repo.MarkCompleted(ctx, id, req.ActualDate, callerID)
	if err != nil {
		return nil, err
	}
	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, IPAddress: ip.String(),
		Module: "tax", Resource: "tax_deadline", ResourceID: &d.ID,
		Action: "MARK_COMPLETED", NewValue: map[string]string{"status": "COMPLETED"},
	})
	return d, nil
}

// AutoGenerate creates standard annual tax deadlines for a client based on their fiscal year end.
// fiscalYearEnd is a "MM-DD" string (e.g. "12-31").
func (uc *TaxDeadlineUseCase) AutoGenerate(ctx context.Context, clientID uuid.UUID, fiscalYearEnd string, year int, callerID uuid.UUID, ip net.IP) ([]*domain.TaxDeadline, error) {
	if fiscalYearEnd == "" {
		return nil, domain.ErrFiscalYearNotConfigured
	}

	// Standard Vietnamese tax deadlines based on fiscal year
	type deadlineDef struct {
		typ  domain.DeadlineType
		name string
		due  time.Time
	}
	defs := []deadlineDef{
		{domain.DeadlineVATFiling, "VAT Q1 Filing", time.Date(year, 4, 30, 0, 0, 0, 0, time.UTC)},
		{domain.DeadlineVATFiling, "VAT Q2 Filing", time.Date(year, 7, 31, 0, 0, 0, 0, time.UTC)},
		{domain.DeadlineVATFiling, "VAT Q3 Filing", time.Date(year, 10, 31, 0, 0, 0, 0, time.UTC)},
		{domain.DeadlineVATFiling, "VAT Q4 Filing", time.Date(year+1, 1, 31, 0, 0, 0, 0, time.UTC)},
		{domain.DeadlineCorporateTax, "Corporate Income Tax Annual", time.Date(year+1, 3, 31, 0, 0, 0, 0, time.UTC)},
		{domain.DeadlineComplianceReporting, "Annual Financial Statement", time.Date(year+1, 3, 31, 0, 0, 0, 0, time.UTC)},
	}

	var created []*domain.TaxDeadline
	for _, def := range defs {
		d, err := uc.repo.Create(ctx, domain.CreateDeadlineParams{
			ClientID:     clientID,
			DeadlineType: def.typ,
			DeadlineName: def.name,
			DueDate:      def.due,
			CreatedBy:    callerID,
		})
		if err != nil {
			return nil, err
		}
		created = append(created, d)
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, IPAddress: ip.String(),
		Module: "tax", Resource: "tax_deadline",
		Action:   "AUTO_GENERATE",
		NewValue: map[string]any{"client_id": clientID, "year": year, "count": len(created)},
	})
	return created, nil
}

// RefreshOverdue is called by the daily cron job to mark past-due deadlines as OVERDUE.
func (uc *TaxDeadlineUseCase) RefreshOverdue(ctx context.Context) error {
	overdue, err := uc.repo.ListOverdue(ctx)
	if err != nil {
		return err
	}
	for _, d := range overdue {
		if err := uc.repo.UpdateStatus(ctx, d.ID, domain.DeadlineStatusOverdue); err != nil {
			return err
		}
	}
	return nil
}

// RefreshDueSoon marks deadlines within 7 days as DUE_SOON.
func (uc *TaxDeadlineUseCase) RefreshDueSoon(ctx context.Context) error {
	soon := time.Now().UTC().AddDate(0, 0, 7)
	deadlines, err := uc.repo.ListDueSoon(ctx, soon)
	if err != nil {
		return err
	}
	for _, d := range deadlines {
		if err := uc.repo.UpdateStatus(ctx, d.ID, domain.DeadlineStatusDueSoon); err != nil {
			return err
		}
	}
	return nil
}
