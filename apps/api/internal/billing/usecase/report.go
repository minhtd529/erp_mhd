package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/mdh/erp-audit/api/internal/billing/domain"
	"github.com/xuri/excelize/v2"
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

// ExportInvoicesXLSX returns an XLSX workbook containing invoices matching the filter.
// The sheet has a styled header row and one data row per invoice.
func (uc *ReportUseCase) ExportInvoicesXLSX(ctx context.Context, f domain.ListInvoicesFilter) ([]byte, error) {
	invoices, err := uc.reportRepo.ListInvoicesForExport(ctx, f)
	if err != nil {
		return nil, err
	}

	xl := excelize.NewFile()
	defer xl.Close()

	sheet := "Invoices"
	xl.SetSheetName("Sheet1", sheet)

	headers := []string{
		"ID", "Invoice Number", "Client ID", "Engagement ID",
		"Type", "Status", "Issue Date", "Due Date",
		"Total Amount", "Tax Amount", "Created At", "Created By",
	}
	for col, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		xl.SetCellValue(sheet, cell, h)
	}

	// Bold header style
	style, _ := xl.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#D4E6F1"}, Pattern: 1},
	})
	lastHeaderCell, _ := excelize.CoordinatesToCellName(len(headers), 1)
	xl.SetCellStyle(sheet, "A1", lastHeaderCell, style)

	for i, inv := range invoices {
		row := i + 2
		issueDate, dueDate, engID := "", "", ""
		if inv.IssueDate != nil {
			issueDate = inv.IssueDate.Format("2006-01-02")
		}
		if inv.DueDate != nil {
			dueDate = inv.DueDate.Format("2006-01-02")
		}
		if inv.EngagementID != nil {
			engID = inv.EngagementID.String()
		}

		rowData := []any{
			inv.ID.String(),
			inv.InvoiceNumber,
			inv.ClientID.String(),
			engID,
			string(inv.InvoiceType),
			string(inv.Status),
			issueDate,
			dueDate,
			inv.TotalAmount,
			inv.TaxAmount,
			inv.CreatedAt.Format(time.RFC3339),
			inv.CreatedBy.String(),
		}
		for col, val := range rowData {
			cell, _ := excelize.CoordinatesToCellName(col+1, row)
			xl.SetCellValue(sheet, cell, val)
		}
	}

	// Auto-fit columns by setting a reasonable width
	for col := 1; col <= len(headers); col++ {
		colName, _ := excelize.ColumnNumberToName(col)
		xl.SetColWidth(sheet, colName, colName, 18)
	}

	var buf bytes.Buffer
	if err := xl.Write(&buf); err != nil {
		return nil, fmt.Errorf("xlsx write: %w", err)
	}
	return buf.Bytes(), nil
}

// ExportPeriodSummaryXLSX returns an XLSX workbook with the billing period summary report.
func (uc *ReportUseCase) ExportPeriodSummaryXLSX(ctx context.Context, req PeriodSummaryRequest) ([]byte, error) {
	end := req.End
	if end.Hour() == 0 && end.Minute() == 0 && end.Second() == 0 {
		end = end.Add(24 * time.Hour)
	}
	summary, err := uc.reportRepo.GetPeriodSummary(ctx, req.Start, end)
	if err != nil {
		return nil, err
	}
	paymentSummary, err := uc.reportRepo.GetPaymentSummary(ctx, req.Start, end)
	if err != nil {
		return nil, err
	}

	xl := excelize.NewFile()
	defer xl.Close()

	// ─── Sheet 1: Period Summary ────────────────────────────────────────────────
	sheet1 := "Period Summary"
	xl.SetSheetName("Sheet1", sheet1)

	titleStyle, _ := xl.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 13},
	})
	headerStyle, _ := xl.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#D4E6F1"}, Pattern: 1},
	})

	xl.SetCellValue(sheet1, "A1", "Billing Period Summary Report")
	xl.SetCellStyle(sheet1, "A1", "A1", titleStyle)
	xl.SetCellValue(sheet1, "A2", fmt.Sprintf("Period: %s → %s", req.Start.Format("2006-01-02"), req.End.Format("2006-01-02")))
	xl.SetCellValue(sheet1, "A3", fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05")))

	// KPI section
	kpiStart := 5
	kpiHeaders := []string{"Metric", "Value"}
	for i, h := range kpiHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, kpiStart)
		xl.SetCellValue(sheet1, cell, h)
		xl.SetCellStyle(sheet1, cell, cell, headerStyle)
	}

	kpis := [][]any{
		{"Total Invoiced", summary.TotalInvoiced},
		{"Total Paid", summary.TotalPaid},
		{"Total Outstanding", summary.TotalOutstanding},
		{"Invoice Count", summary.InvoiceCount},
		{"Paid Count", summary.PaidCount},
		{"Overdue Count", summary.OverdueCount},
	}
	for i, row := range kpis {
		r := kpiStart + 1 + i
		for col, val := range row {
			cell, _ := excelize.CoordinatesToCellName(col+1, r)
			xl.SetCellValue(sheet1, cell, val)
		}
	}

	// Status breakdown
	statusStart := kpiStart + len(kpis) + 3
	xl.SetCellValue(sheet1, fmt.Sprintf("A%d", statusStart), "Invoice Status Breakdown")
	xl.SetCellStyle(sheet1, fmt.Sprintf("A%d", statusStart), fmt.Sprintf("A%d", statusStart), titleStyle)
	statusHeaders := []string{"Status", "Count", "Amount"}
	for i, h := range statusHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, statusStart+1)
		xl.SetCellValue(sheet1, cell, h)
		xl.SetCellStyle(sheet1, cell, cell, headerStyle)
	}
	for i, sc := range summary.ByStatus {
		r := statusStart + 2 + i
		xl.SetCellValue(sheet1, fmt.Sprintf("A%d", r), string(sc.Status))
		xl.SetCellValue(sheet1, fmt.Sprintf("B%d", r), sc.Count)
		xl.SetCellValue(sheet1, fmt.Sprintf("C%d", r), sc.Amount)
	}

	// ─── Sheet 2: Payment Summary ────────────────────────────────────────────────
	sheet2 := "Payment Summary"
	xl.NewSheet(sheet2)

	xl.SetCellValue(sheet2, "A1", "Payment Summary Report")
	xl.SetCellStyle(sheet2, "A1", "A1", titleStyle)
	xl.SetCellValue(sheet2, "A2", fmt.Sprintf("Period: %s → %s", req.Start.Format("2006-01-02"), req.End.Format("2006-01-02")))

	payKPIStart := 4
	payKPIs := [][]any{
		{"Total Received", paymentSummary.TotalReceived},
		{"Payment Count", paymentSummary.PaymentCount},
	}
	payHeaders := []string{"Metric", "Value"}
	for i, h := range payHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, payKPIStart)
		xl.SetCellValue(sheet2, cell, h)
		xl.SetCellStyle(sheet2, cell, cell, headerStyle)
	}
	for i, row := range payKPIs {
		r := payKPIStart + 1 + i
		for col, val := range row {
			cell, _ := excelize.CoordinatesToCellName(col+1, r)
			xl.SetCellValue(sheet2, cell, val)
		}
	}

	// Payment method breakdown
	methodStart := payKPIStart + len(payKPIs) + 3
	xl.SetCellValue(sheet2, fmt.Sprintf("A%d", methodStart), "Payment Method Breakdown")
	xl.SetCellStyle(sheet2, fmt.Sprintf("A%d", methodStart), fmt.Sprintf("A%d", methodStart), titleStyle)
	methodHeaders := []string{"Method", "Count", "Amount"}
	for i, h := range methodHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, methodStart+1)
		xl.SetCellValue(sheet2, cell, h)
		xl.SetCellStyle(sheet2, cell, cell, headerStyle)
	}
	for i, mc := range paymentSummary.ByMethod {
		r := methodStart + 2 + i
		xl.SetCellValue(sheet2, fmt.Sprintf("A%d", r), string(mc.Method))
		xl.SetCellValue(sheet2, fmt.Sprintf("B%d", r), mc.Count)
		xl.SetCellValue(sheet2, fmt.Sprintf("C%d", r), mc.Amount)
	}

	xl.SetColWidth(sheet1, "A", "C", 22)
	xl.SetColWidth(sheet2, "A", "C", 22)

	var buf bytes.Buffer
	if err := xl.Write(&buf); err != nil {
		return nil, fmt.Errorf("xlsx write: %w", err)
	}
	return buf.Bytes(), nil
}
