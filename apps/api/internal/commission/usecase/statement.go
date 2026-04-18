package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/commission/domain"
)

// CommissionStatement is the response for /me/commissions/statement.
type CommissionStatement struct {
	SalespersonID uuid.UUID        `json:"salesperson_id"`
	Period        string           `json:"period"`
	From          time.Time        `json:"from"`
	To            time.Time        `json:"to"`
	Records       []RecordResponse `json:"records"`
	TotalAccrued  int64            `json:"total_accrued"`
	TotalPayable  int64            `json:"total_payable"`
	TotalPaid     int64            `json:"total_paid"`
	GeneratedAt   time.Time        `json:"generated_at"`
}

// GetStatement returns commission records for a salesperson for the given period.
// period format: "2026-Q1", "2026-01" (month), or "2026" (full year).
func (uc *RecordUseCase) GetStatement(ctx context.Context, salespersonID uuid.UUID, period string) (*CommissionStatement, error) {
	from, to, err := parsePeriod(period)
	if err != nil {
		return nil, err
	}

	records, err := uc.recordRepo.ListForStatement(ctx, domain.StatementFilter{
		SalespersonID: salespersonID,
		From:          from.Format(time.RFC3339),
		To:            to.Format(time.RFC3339),
	})
	if err != nil {
		return nil, err
	}

	stmt := &CommissionStatement{
		SalespersonID: salespersonID,
		Period:        period,
		From:          from,
		To:            to,
		GeneratedAt:   time.Now().UTC(),
	}

	items := make([]RecordResponse, len(records))
	for i, r := range records {
		resp := toRecordResponse(r)
		items[i] = resp
		stmt.TotalAccrued += r.CalculatedAmount
		stmt.TotalPayable += r.PayableAmount
		if r.Status == domain.CommStatusPaid {
			stmt.TotalPaid += r.PayableAmount
		}
	}
	stmt.Records = items
	return stmt, nil
}

// ExportStatementCSV returns a CSV byte slice for the commission statement.
func (uc *RecordUseCase) ExportStatementCSV(ctx context.Context, salespersonID uuid.UUID, period string) ([]byte, error) {
	from, to, err := parsePeriod(period)
	if err != nil {
		return nil, err
	}

	records, err := uc.recordRepo.ListForStatement(ctx, domain.StatementFilter{
		SalespersonID: salespersonID,
		From:          from.Format(time.RFC3339),
		To:            to.Format(time.RFC3339),
	})
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	_ = w.Write([]string{
		"id", "engagement_id", "status",
		"base_amount", "rate", "calculated_amount", "holdback_amount", "payable_amount",
		"accrued_at", "approved_at", "paid_at", "payout_reference",
		"is_clawback", "clawback_reason",
	})

	for _, r := range records {
		approvedAt := ""
		if r.ApprovedAt != nil {
			approvedAt = r.ApprovedAt.Format("2006-01-02")
		}
		paidAt := ""
		if r.PaidAt != nil {
			paidAt = r.PaidAt.Format("2006-01-02")
		}
		_ = w.Write([]string{
			r.ID.String(),
			r.EngagementID.String(),
			string(r.Status),
			strconv.FormatInt(r.BaseAmount, 10),
			strconv.FormatFloat(r.Rate, 'f', 6, 64),
			strconv.FormatInt(r.CalculatedAmount, 10),
			strconv.FormatInt(r.HoldbackAmount, 10),
			strconv.FormatInt(r.PayableAmount, 10),
			r.AccruedAt.Format("2006-01-02"),
			approvedAt,
			paidAt,
			r.PayoutReference,
			strconv.FormatBool(r.IsClawback),
			r.ClawbackReason,
		})
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("csv write: %w", err)
	}
	return buf.Bytes(), nil
}

// parsePeriod converts a period string into a [from, to) time range.
// Supported formats: "2026-Q1", "2026-Q2", "2026-Q3", "2026-Q4", "2026-01".."2026-12", "2026".
func parsePeriod(period string) (from, to time.Time, err error) {
	period = strings.TrimSpace(period)

	// Year-Quarter: "2026-Q1"
	if len(period) == 7 && strings.Contains(period, "-Q") {
		parts := strings.SplitN(period, "-Q", 2)
		year, e1 := strconv.Atoi(parts[0])
		q, e2 := strconv.Atoi(parts[1])
		if e1 != nil || e2 != nil || q < 1 || q > 4 {
			err = fmt.Errorf("invalid period %q: expected YYYY-Q[1-4]", period)
			return
		}
		startMonth := time.Month((q-1)*3 + 1)
		from = time.Date(year, startMonth, 1, 0, 0, 0, 0, time.UTC)
		to = from.AddDate(0, 3, 0)
		return
	}

	// Year-Month: "2026-01"
	if len(period) == 7 && strings.Contains(period, "-") {
		parts := strings.SplitN(period, "-", 2)
		year, e1 := strconv.Atoi(parts[0])
		month, e2 := strconv.Atoi(parts[1])
		if e1 != nil || e2 != nil || month < 1 || month > 12 {
			err = fmt.Errorf("invalid period %q: expected YYYY-MM", period)
			return
		}
		from = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		to = from.AddDate(0, 1, 0)
		return
	}

	// Full year: "2026"
	if len(period) == 4 {
		year, e := strconv.Atoi(period)
		if e != nil {
			err = fmt.Errorf("invalid period %q: expected YYYY", period)
			return
		}
		from = time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
		to = from.AddDate(1, 0, 0)
		return
	}

	err = fmt.Errorf("invalid period %q: use YYYY-Q[1-4], YYYY-MM, or YYYY", period)
	return
}
