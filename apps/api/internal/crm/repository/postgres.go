// Package repository provides the PostgreSQL implementation of the CRM domain
// repository interfaces.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/crm/domain"
)

// Repo implements domain.ClientRepository using pgxpool.
type Repo struct {
	pool *pgxpool.Pool
}

// New creates a new CRM Repo.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

// Create inserts a new client row and returns the full entity.
func (r *Repo) Create(ctx context.Context, p domain.CreateClientParams) (*domain.Client, error) {
	const q = `
		INSERT INTO clients (tax_code, business_name, english_name, industry, office_id,
		                     sales_owner_id, referrer_id,
		                     address,
		                     bank_name, bank_account_number, bank_account_name,
		                     representative_name, representative_title, representative_phone,
		                     created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $15)
		RETURNING id, tax_code, business_name, english_name, industry, status, office_id,
		          sales_owner_id, referrer_id,
		          address,
		          bank_name, bank_account_number, bank_account_name,
		          representative_name, representative_title, representative_phone,
		          is_deleted, created_at, updated_at, created_by, updated_by`

	row := r.pool.QueryRow(ctx, q,
		p.TaxCode, p.BusinessName, p.EnglishName, p.Industry, p.OfficeID,
		p.SalesOwnerID, p.ReferrerID,
		p.Address,
		p.BankName, p.BankAccountNumber, p.BankAccountName,
		p.RepresentativeName, p.RepresentativeTitle, p.RepresentativePhone,
		p.CreatedBy,
	)
	c, err := scanClient(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrDuplicateTaxCode
		}
		return nil, fmt.Errorf("crm.Create: %w", err)
	}
	return c, nil
}

// FindByID returns a single non-deleted client.
func (r *Repo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Client, error) {
	const q = `
		SELECT id, tax_code, business_name, english_name, industry, status, office_id,
		       sales_owner_id, referrer_id,
		       address,
		       bank_name, bank_account_number, bank_account_name,
		       representative_name, representative_title, representative_phone,
		       is_deleted, created_at, updated_at, created_by, updated_by
		FROM clients WHERE id = $1 AND is_deleted = false`

	c, err := scanClient(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrClientNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("crm.FindByID: %w", err)
	}
	return c, nil
}

// Update patches mutable fields and returns the refreshed entity.
func (r *Repo) Update(ctx context.Context, p domain.UpdateClientParams) (*domain.Client, error) {
	const q = `
		UPDATE clients
		SET business_name        = $2,
		    english_name         = $3,
		    industry             = $4,
		    office_id            = $5,
		    sales_owner_id       = $6,
		    referrer_id          = $7,
		    address              = $8,
		    bank_name            = $9,
		    bank_account_number  = $10,
		    bank_account_name    = $11,
		    representative_name  = $12,
		    representative_title = $13,
		    representative_phone = $14,
		    updated_by           = $15,
		    updated_at           = NOW()
		WHERE id = $1 AND is_deleted = false
		RETURNING id, tax_code, business_name, english_name, industry, status, office_id,
		          sales_owner_id, referrer_id,
		          address,
		          bank_name, bank_account_number, bank_account_name,
		          representative_name, representative_title, representative_phone,
		          is_deleted, created_at, updated_at, created_by, updated_by`

	c, err := scanClient(r.pool.QueryRow(ctx, q,
		p.ID, p.BusinessName, p.EnglishName, p.Industry, p.OfficeID,
		p.SalesOwnerID, p.ReferrerID,
		p.Address,
		p.BankName, p.BankAccountNumber, p.BankAccountName,
		p.RepresentativeName, p.RepresentativeTitle, p.RepresentativePhone,
		p.UpdatedBy,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrClientNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("crm.Update: %w", err)
	}
	return c, nil
}

// SoftDelete marks a client as deleted.
func (r *Repo) SoftDelete(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	const q = `
		UPDATE clients SET is_deleted = true, updated_by = $2, updated_at = NOW()
		WHERE id = $1 AND is_deleted = false`

	tag, err := r.pool.Exec(ctx, q, id, deletedBy)
	if err != nil {
		return fmt.Errorf("crm.SoftDelete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrClientNotFound
	}
	return nil
}

// List returns a paginated list of non-deleted clients with optional filtering.
func (r *Repo) List(ctx context.Context, f domain.ListClientsFilter) ([]*domain.Client, int64, error) {
	offset := (f.Page - 1) * f.Size

	// Build dynamic WHERE clause
	args := []any{}
	where := "WHERE is_deleted = false"
	idx := 1

	if f.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, string(f.Status))
		idx++
	}
	if f.Q != "" {
		// Searches the combined trgm expression index: business_name + english_name + tax_code + representative_name
		where += fmt.Sprintf(` AND (
			COALESCE(business_name,'') || ' ' ||
			COALESCE(english_name, '') || ' ' ||
			COALESCE(tax_code,     '') || ' ' ||
			COALESCE(representative_name,'')
		) ILIKE $%d`, idx)
		args = append(args, "%"+f.Q+"%")
		idx++
	}

	// Count query
	var total int64
	countQ := "SELECT COUNT(*) FROM clients " + where
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("crm.List count: %w", err)
	}

	// Data query
	args = append(args, f.Size, offset)
	dataQ := fmt.Sprintf(`
		SELECT id, tax_code, business_name, english_name, industry, status, office_id,
		       sales_owner_id, referrer_id,
		       address,
		       bank_name, bank_account_number, bank_account_name,
		       representative_name, representative_title, representative_phone,
		       is_deleted, created_at, updated_at, created_by, updated_by
		FROM clients %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, where, idx, idx+1)

	rows, err := r.pool.Query(ctx, dataQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("crm.List query: %w", err)
	}
	defer rows.Close()

	var clients []*domain.Client
	for rows.Next() {
		c, err := scanClient(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("crm.List scan: %w", err)
		}
		clients = append(clients, c)
	}
	if clients == nil {
		clients = []*domain.Client{}
	}
	return clients, total, rows.Err()
}

// ─── helpers ─────────────────────────────────────────────────────────────────

type scanner interface {
	Scan(dest ...any) error
}

func scanClient(row scanner) (*domain.Client, error) {
	var c domain.Client
	err := row.Scan(
		&c.ID, &c.TaxCode, &c.BusinessName, &c.EnglishName, &c.Industry,
		&c.Status, &c.OfficeID, &c.SalesOwnerID, &c.ReferrerID,
		&c.Address, // NOT NULL — scans directly into string
		&c.BankName, &c.BankAccountNumber, &c.BankAccountName,
		&c.RepresentativeName, &c.RepresentativeTitle, &c.RepresentativePhone,
		&c.IsDeleted, &c.CreatedAt, &c.UpdatedAt, &c.CreatedBy, &c.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
