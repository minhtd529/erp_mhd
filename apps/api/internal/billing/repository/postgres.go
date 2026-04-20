// Package repository provides the PostgreSQL implementation of the Billing domain.
package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/billing/domain"
)

// ─── Invoice Repository ───────────────────────────────────────────────────────

type InvoiceRepo struct{ pool *pgxpool.Pool }

func NewInvoiceRepo(pool *pgxpool.Pool) *InvoiceRepo { return &InvoiceRepo{pool: pool} }

const invoiceCols = `
	id, invoice_number, client_id, engagement_id, invoice_type,
	status, issue_date, due_date, total_amount, tax_amount,
	snapshot_data, notes, is_deleted, created_at, updated_at, created_by, updated_by`

func (r *InvoiceRepo) Create(ctx context.Context, p domain.CreateInvoiceParams) (*domain.Invoice, error) {
	const q = `
		INSERT INTO invoices
			(invoice_number, client_id, engagement_id, invoice_type,
			 issue_date, due_date, total_amount, tax_amount, notes, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$10)
		RETURNING ` + invoiceCols

	row := r.pool.QueryRow(ctx, q,
		p.InvoiceNumber, p.ClientID, p.EngagementID, string(p.InvoiceType),
		p.IssueDate, p.DueDate, p.TotalAmount, p.TaxAmount, p.Notes, p.CreatedBy,
	)
	inv, err := scanInvoice(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrInvoiceNumberDuplicate
		}
		return nil, fmt.Errorf("billing.Create: %w", err)
	}
	return inv, nil
}

func (r *InvoiceRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	const q = `SELECT ` + invoiceCols + ` FROM invoices WHERE id = $1 AND is_deleted = false`
	inv, err := scanInvoice(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrInvoiceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("billing.FindByID: %w", err)
	}
	return inv, nil
}

func (r *InvoiceRepo) Update(ctx context.Context, p domain.UpdateInvoiceParams) (*domain.Invoice, error) {
	const q = `
		UPDATE invoices SET
			issue_date   = $2,
			due_date     = $3,
			total_amount = $4,
			tax_amount   = $5,
			notes        = $6,
			updated_by   = $7,
			updated_at   = NOW()
		WHERE id = $1 AND is_deleted = false AND status = 'DRAFT'
		RETURNING ` + invoiceCols

	inv, err := scanInvoice(r.pool.QueryRow(ctx, q,
		p.ID, p.IssueDate, p.DueDate, p.TotalAmount, p.TaxAmount, p.Notes, p.UpdatedBy,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrInvoiceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("billing.Update: %w", err)
	}
	return inv, nil
}

func (r *InvoiceRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.InvoiceStatus, updatedBy uuid.UUID) (*domain.Invoice, error) {
	const q = `
		UPDATE invoices SET status = $2, updated_by = $3, updated_at = NOW()
		WHERE id = $1 AND is_deleted = false
		RETURNING ` + invoiceCols

	inv, err := scanInvoice(r.pool.QueryRow(ctx, q, id, string(status), updatedBy))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrInvoiceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("billing.UpdateStatus: %w", err)
	}
	return inv, nil
}

func (r *InvoiceRepo) UpdateSnapshot(ctx context.Context, id uuid.UUID, snapshot []byte, updatedBy uuid.UUID) (*domain.Invoice, error) {
	const q = `
		UPDATE invoices SET snapshot_data = $2, updated_by = $3, updated_at = NOW()
		WHERE id = $1 AND is_deleted = false
		RETURNING ` + invoiceCols

	inv, err := scanInvoice(r.pool.QueryRow(ctx, q, id, snapshot, updatedBy))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrInvoiceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("billing.UpdateSnapshot: %w", err)
	}
	return inv, nil
}

func (r *InvoiceRepo) SoftDelete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	const q = `
		UPDATE invoices SET is_deleted = true, updated_by = $2, updated_at = NOW()
		WHERE id = $1 AND is_deleted = false AND status = 'DRAFT'`

	tag, err := r.pool.Exec(ctx, q, id, deletedBy)
	if err != nil {
		return fmt.Errorf("billing.SoftDelete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrInvoiceNotFound
	}
	return nil
}

func (r *InvoiceRepo) List(ctx context.Context, f domain.ListInvoicesFilter) ([]*domain.Invoice, int64, error) {
	offset := (f.Page - 1) * f.Size
	args := []any{}
	where := "WHERE is_deleted = false"
	idx := 1

	if f.ClientID != nil {
		where += fmt.Sprintf(" AND client_id = $%d", idx)
		args = append(args, f.ClientID)
		idx++
	}
	if f.EngagementID != nil {
		where += fmt.Sprintf(" AND engagement_id = $%d", idx)
		args = append(args, f.EngagementID)
		idx++
	}
	if len(f.Statuses) > 0 {
		placeholders := ""
		for i, s := range f.Statuses {
			if i > 0 {
				placeholders += ","
			}
			placeholders += fmt.Sprintf("$%d", idx)
			args = append(args, string(s))
			idx++
		}
		where += " AND status IN (" + placeholders + ")"
	} else if f.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, string(f.Status))
		idx++
	}
	if f.Q != "" {
		where += fmt.Sprintf(" AND invoice_number ILIKE $%d", idx)
		args = append(args, "%"+f.Q+"%")
		idx++
	}

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM invoices "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("billing.List count: %w", err)
	}

	args = append(args, f.Size, offset)
	dataQ := fmt.Sprintf(
		`SELECT `+invoiceCols+` FROM invoices %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, idx, idx+1,
	)
	rows, err := r.pool.Query(ctx, dataQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("billing.List query: %w", err)
	}
	defer rows.Close()

	var invoices []*domain.Invoice
	for rows.Next() {
		inv, err := scanInvoice(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("billing.List scan: %w", err)
		}
		invoices = append(invoices, inv)
	}
	if invoices == nil {
		invoices = []*domain.Invoice{}
	}
	return invoices, total, rows.Err()
}

// ─── Line Item Repository ─────────────────────────────────────────────────────

type LineItemRepo struct{ pool *pgxpool.Pool }

func NewLineItemRepo(pool *pgxpool.Pool) *LineItemRepo { return &LineItemRepo{pool: pool} }

func (r *LineItemRepo) Add(ctx context.Context, p domain.AddLineItemParams) (*domain.InvoiceLineItem, error) {
	const q = `
		INSERT INTO invoice_line_items
			(invoice_id, description, quantity, unit_price, tax_amount, total_amount, source_type, snapshot_data)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, invoice_id, description, quantity, unit_price, tax_amount, total_amount, source_type, snapshot_data, created_at`

	snap := p.SnapshotData
	if snap == nil {
		snap = json.RawMessage("{}")
	}
	var item domain.InvoiceLineItem
	var srcType string
	err := r.pool.QueryRow(ctx, q,
		p.InvoiceID, p.Description, p.Quantity, p.UnitPrice, p.TaxAmount, p.TotalAmount,
		string(p.SourceType), snap,
	).Scan(
		&item.ID, &item.InvoiceID, &item.Description, &item.Quantity, &item.UnitPrice,
		&item.TaxAmount, &item.TotalAmount, &srcType, &item.SnapshotData, &item.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("billing.AddLineItem: %w", err)
	}
	item.SourceType = domain.LineItemSourceType(srcType)
	return &item, nil
}

func (r *LineItemRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.InvoiceLineItem, error) {
	const q = `SELECT id, invoice_id, description, quantity, unit_price, tax_amount, total_amount, source_type, snapshot_data, created_at
		FROM invoice_line_items WHERE id = $1`
	var item domain.InvoiceLineItem
	var srcType string
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&item.ID, &item.InvoiceID, &item.Description, &item.Quantity, &item.UnitPrice,
		&item.TaxAmount, &item.TotalAmount, &srcType, &item.SnapshotData, &item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrLineItemNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("billing.FindLineItemByID: %w", err)
	}
	item.SourceType = domain.LineItemSourceType(srcType)
	return &item, nil
}

func (r *LineItemRepo) ListByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]*domain.InvoiceLineItem, error) {
	const q = `SELECT id, invoice_id, description, quantity, unit_price, tax_amount, total_amount, source_type, snapshot_data, created_at
		FROM invoice_line_items WHERE invoice_id = $1 ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, q, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("billing.ListLineItems: %w", err)
	}
	defer rows.Close()
	var items []*domain.InvoiceLineItem
	for rows.Next() {
		var item domain.InvoiceLineItem
		var srcType string
		if err := rows.Scan(&item.ID, &item.InvoiceID, &item.Description, &item.Quantity, &item.UnitPrice,
			&item.TaxAmount, &item.TotalAmount, &srcType, &item.SnapshotData, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("billing.ListLineItems scan: %w", err)
		}
		item.SourceType = domain.LineItemSourceType(srcType)
		items = append(items, &item)
	}
	if items == nil {
		items = []*domain.InvoiceLineItem{}
	}
	return items, rows.Err()
}

func (r *LineItemRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM invoice_line_items WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("billing.DeleteLineItem: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrLineItemNotFound
	}
	return nil
}

// ─── Payment Repository ───────────────────────────────────────────────────────

type PaymentRepo struct{ pool *pgxpool.Pool }

func NewPaymentRepo(pool *pgxpool.Pool) *PaymentRepo { return &PaymentRepo{pool: pool} }

const paymentCols = `id, invoice_id, payment_method, amount, payment_date,
	reference_number, status, notes, recorded_by, recorded_at, cleared_at, created_at, updated_at`

func (r *PaymentRepo) Record(ctx context.Context, p domain.RecordPaymentParams) (*domain.Payment, error) {
	const q = `
		INSERT INTO payments
			(invoice_id, payment_method, amount, payment_date, reference_number, notes, recorded_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING ` + paymentCols

	pay, err := scanPayment(r.pool.QueryRow(ctx, q,
		p.InvoiceID, string(p.PaymentMethod), p.Amount, p.PaymentDate,
		p.ReferenceNumber, p.Notes, p.RecordedBy,
	))
	if err != nil {
		return nil, fmt.Errorf("billing.RecordPayment: %w", err)
	}
	return pay, nil
}

func (r *PaymentRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	q := `SELECT ` + paymentCols + ` FROM payments WHERE id = $1`
	pay, err := scanPayment(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrPaymentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("billing.FindPaymentByID: %w", err)
	}
	return pay, nil
}

func (r *PaymentRepo) Update(ctx context.Context, p domain.UpdatePaymentParams) (*domain.Payment, error) {
	const q = `
		UPDATE payments SET
			payment_method   = $2,
			amount           = $3,
			payment_date     = $4,
			reference_number = $5,
			notes            = $6,
			updated_at       = NOW()
		WHERE id = $1 AND status = 'RECORDED'
		RETURNING ` + paymentCols

	pay, err := scanPayment(r.pool.QueryRow(ctx, q,
		p.ID, string(p.PaymentMethod), p.Amount, p.PaymentDate, p.ReferenceNumber, p.Notes,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrPaymentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("billing.UpdatePayment: %w", err)
	}
	return pay, nil
}

func (r *PaymentRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus) (*domain.Payment, error) {
	const q = `
		UPDATE payments SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING ` + paymentCols

	pay, err := scanPayment(r.pool.QueryRow(ctx, q, id, string(status)))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrPaymentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("billing.UpdatePaymentStatus: %w", err)
	}
	return pay, nil
}

func (r *PaymentRepo) Clear(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	const q = `
		UPDATE payments SET status='CLEARED', cleared_at=NOW(), updated_at=NOW()
		WHERE id=$1 AND status='RECORDED'
		RETURNING ` + paymentCols

	pay, err := scanPayment(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrPaymentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("billing.ClearPayment: %w", err)
	}
	return pay, nil
}

func (r *PaymentRepo) ListByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]*domain.Payment, error) {
	q := `SELECT ` + paymentCols + ` FROM payments WHERE invoice_id = $1 ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, q, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("billing.ListPayments: %w", err)
	}
	defer rows.Close()
	var payments []*domain.Payment
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, fmt.Errorf("billing.ListPayments scan: %w", err)
		}
		payments = append(payments, p)
	}
	if payments == nil {
		payments = []*domain.Payment{}
	}
	return payments, rows.Err()
}

func (r *PaymentRepo) ListAll(ctx context.Context, page, size int) ([]*domain.Payment, int64, error) {
	if size <= 0 {
		size = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * size

	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM payments`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("billing.ListAllPayments count: %w", err)
	}

	q := `SELECT ` + paymentCols + ` FROM payments ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, q, size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("billing.ListAllPayments: %w", err)
	}
	defer rows.Close()
	var payments []*domain.Payment
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("billing.ListAllPayments scan: %w", err)
		}
		payments = append(payments, p)
	}
	if payments == nil {
		payments = []*domain.Payment{}
	}
	return payments, total, rows.Err()
}

func (r *PaymentRepo) SumPaidByInvoice(ctx context.Context, invoiceID uuid.UUID) (float64, error) {
	var sum float64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount),0) FROM payments WHERE invoice_id = $1 AND status NOT IN ('REVERSED')`,
		invoiceID,
	).Scan(&sum)
	if err != nil {
		return 0, fmt.Errorf("billing.SumPaid: %w", err)
	}
	return sum, nil
}

// ─── Memo Repository ──────────────────────────────────────────────────────────

type MemoRepo struct{ pool *pgxpool.Pool }

func NewMemoRepo(pool *pgxpool.Pool) *MemoRepo { return &MemoRepo{pool: pool} }

func (r *MemoRepo) Create(ctx context.Context, p domain.CreateMemoParams) (*domain.BillingMemo, error) {
	const q = `
		INSERT INTO billing_memos
			(related_invoice_id, memo_type, memo_number, amount, reason, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$6)
		RETURNING id, related_invoice_id, memo_type, memo_number, amount, reason, status,
		          is_deleted, created_at, updated_at, created_by, updated_by`

	var m domain.BillingMemo
	var memoType, status string
	err := r.pool.QueryRow(ctx, q,
		p.RelatedInvoiceID, string(p.MemoType), p.MemoNumber, p.Amount, p.Reason, p.CreatedBy,
	).Scan(
		&m.ID, &m.RelatedInvoiceID, &memoType, &m.MemoNumber, &m.Amount, &m.Reason, &status,
		&m.IsDeleted, &m.CreatedAt, &m.UpdatedAt, &m.CreatedBy, &m.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("billing.CreateMemo: %w", err)
	}
	m.MemoType = domain.MemoType(memoType)
	m.Status = domain.MemoStatus(status)
	return &m, nil
}

func (r *MemoRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.BillingMemo, error) {
	const q = `SELECT id, related_invoice_id, memo_type, memo_number, amount, reason, status,
		is_deleted, created_at, updated_at, created_by, updated_by
		FROM billing_memos WHERE id = $1 AND is_deleted = false`
	var m domain.BillingMemo
	var memoType, status string
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&m.ID, &m.RelatedInvoiceID, &memoType, &m.MemoNumber, &m.Amount, &m.Reason, &status,
		&m.IsDeleted, &m.CreatedAt, &m.UpdatedAt, &m.CreatedBy, &m.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrMemoNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("billing.FindMemoByID: %w", err)
	}
	m.MemoType = domain.MemoType(memoType)
	m.Status = domain.MemoStatus(status)
	return &m, nil
}

func (r *MemoRepo) List(ctx context.Context, page, size int) ([]*domain.BillingMemo, int64, error) {
	offset := (page - 1) * size
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM billing_memos WHERE is_deleted = false`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("billing.ListMemos count: %w", err)
	}
	const q = `SELECT id, related_invoice_id, memo_type, memo_number, amount, reason, status,
		is_deleted, created_at, updated_at, created_by, updated_by
		FROM billing_memos WHERE is_deleted = false ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, q, size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("billing.ListMemos: %w", err)
	}
	defer rows.Close()
	var memos []*domain.BillingMemo
	for rows.Next() {
		var m domain.BillingMemo
		var memoType, status string
		if err := rows.Scan(&m.ID, &m.RelatedInvoiceID, &memoType, &m.MemoNumber, &m.Amount, &m.Reason, &status,
			&m.IsDeleted, &m.CreatedAt, &m.UpdatedAt, &m.CreatedBy, &m.UpdatedBy); err != nil {
			return nil, 0, fmt.Errorf("billing.ListMemos scan: %w", err)
		}
		m.MemoType = domain.MemoType(memoType)
		m.Status = domain.MemoStatus(status)
		memos = append(memos, &m)
	}
	if memos == nil {
		memos = []*domain.BillingMemo{}
	}
	return memos, total, rows.Err()
}

// ─── helpers ──────────────────────────────────────────────────────────────────

type scanner interface{ Scan(dest ...any) error }

func scanInvoice(row scanner) (*domain.Invoice, error) {
	var inv domain.Invoice
	var invType, status string
	err := row.Scan(
		&inv.ID, &inv.InvoiceNumber, &inv.ClientID, &inv.EngagementID,
		&invType, &status,
		&inv.IssueDate, &inv.DueDate, &inv.TotalAmount, &inv.TaxAmount,
		&inv.SnapshotData, &inv.Notes,
		&inv.IsDeleted, &inv.CreatedAt, &inv.UpdatedAt, &inv.CreatedBy, &inv.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	inv.InvoiceType = domain.InvoiceType(invType)
	inv.Status = domain.InvoiceStatus(status)
	return &inv, nil
}

func scanPayment(row scanner) (*domain.Payment, error) {
	var p domain.Payment
	var method, status string
	err := row.Scan(
		&p.ID, &p.InvoiceID, &method, &p.Amount, &p.PaymentDate,
		&p.ReferenceNumber, &status, &p.Notes,
		&p.RecordedBy, &p.RecordedAt, &p.ClearedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	p.PaymentMethod = domain.PaymentMethod(method)
	p.Status = domain.PaymentStatus(status)
	return &p, nil
}
