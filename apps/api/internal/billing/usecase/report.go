package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/mdh/erp-audit/api/internal/billing/domain"
)

// ReportUseCase handles billing report generation and data export.
type ReportUseCase struct {
	reportRepo domain.ReportRepository
}

// NewReportUseCase constructs a ReportUseCase.
func NewReportUseCase(reportRepo domain.ReportRepository) *ReportUseCase {
	return &ReportUseCase{reportRepo: reportRepo}
}

// PeriodSummaryRequest carries date range parameters for billing summary reports.
type PeriodSummaryRequest struct {
	Start time.Time `form:"start" binding:"required"`
	End   time.Time `form:"end"   binding:"required"`
}

// ExportInvoicesRequest carries filter + format for invoice export.
type ExportInvoicesRequest struct {
	Status       domain.InvoiceStatus `form:"status"`
	EngagementID *string              `form:"engagement_id"` // string so binding works, parsed in handler
}

func (uc *ReportUseCase) GetPeriodSummary(ctx context.Context, req PeriodSummaryRequest) (*domain.BillingPeriodSummary, error) {
	end := req.End
	// Include the whole end day
	if end.Hour() == 0 && end.Minute() == 0 && end.Second() == 0 {
		end = end.Add(24 * time.Hour)
	}
	return uc.reportRepo.GetPeriodSummary(ctx, req.Start, end)
}

func (uc *ReportUseCase) GetPaymentSummary(ctx context.Context, req PeriodSummaryRequest) (*domain.PaymentSummary, error) {
	end := req.End
	if end.Hour() == 0 && end.Minute() == 0 && end.Second() == 0 {
		end = end.Add(24 * time.Hour)
	}
	return uc.reportRepo.GetPaymentSummary(ctx, req.Start, end)
}

// ExportInvoicesCSV returns a CSV-encoded byte slice of invoices matching the filter.
func (uc *ReportUseCase) ExportInvoicesCSV(ctx context.Context, f domain.ListInvoicesFilter) ([]byte, error) {
	invoices, err := uc.reportRepo.ListInvoicesForExport(ctx, f)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Header
	_ = w.Write([]string{
		"id", "invoice_number", "client_id", "engagement_id",
		"invoice_type", "status", "issue_date", "due_date",
		"total_amount", "tax_amount", "created_at", "created_by",
	})

	for _, inv := range invoices {
		issueDate := ""
		if inv.IssueDate != nil {
			issueDate = inv.IssueDate.Format("2006-01-02")
		}
		dueDate := ""
		if inv.DueDate != nil {
			dueDate = inv.DueDate.Format("2006-01-02")
		}
		engID := ""
		if inv.EngagementID != nil {
			engID = inv.EngagementID.String()
		}
		_ = w.Write([]string{
			inv.ID.String(),
			inv.InvoiceNumber,
			inv.ClientID.String(),
			engID,
			string(inv.InvoiceType),
			string(inv.Status),
			issueDate,
			dueDate,
			fmt.Sprintf("%.2f", inv.TotalAmount),
			fmt.Sprintf("%.2f", inv.TaxAmount),
			inv.CreatedAt.Format(time.RFC3339),
			inv.CreatedBy.String(),
		})
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("csv write: %w", err)
	}
	return buf.Bytes(), nil
}
