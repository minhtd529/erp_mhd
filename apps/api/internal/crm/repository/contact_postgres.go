package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/crm/domain"
)

// ContactRepo implements domain.ContactRepository using pgxpool.
type ContactRepo struct {
	pool *pgxpool.Pool
}

// NewContactRepo creates a new ContactRepo.
func NewContactRepo(pool *pgxpool.Pool) *ContactRepo {
	return &ContactRepo{pool: pool}
}

// Create inserts a new client contact.
func (r *ContactRepo) Create(ctx context.Context, p domain.CreateContactParams) (*domain.ClientContact, error) {
	const q = `
		INSERT INTO client_contacts (client_id, full_name, title, phone, email, is_primary, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		RETURNING id, client_id, full_name, title, phone, email, is_primary,
		          is_deleted, created_at, updated_at, created_by, updated_by`

	cc, err := scanContact(r.pool.QueryRow(ctx, q,
		p.ClientID, p.FullName, p.Title, p.Phone, p.Email, p.IsPrimary, p.CreatedBy,
	))
	if err != nil {
		return nil, fmt.Errorf("contact.Create: %w", err)
	}
	return cc, nil
}

// FindByID returns a single non-deleted contact.
func (r *ContactRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.ClientContact, error) {
	const q = `
		SELECT id, client_id, full_name, title, phone, email, is_primary,
		       is_deleted, created_at, updated_at, created_by, updated_by
		FROM client_contacts WHERE id = $1 AND is_deleted = false`

	cc, err := scanContact(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrContactNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("contact.FindByID: %w", err)
	}
	return cc, nil
}

// Update patches a contact and returns the refreshed entity.
// clientID is used to scope the update so a contact cannot be moved to another client.
func (r *ContactRepo) Update(ctx context.Context, p domain.UpdateContactParams) (*domain.ClientContact, error) {
	const q = `
		UPDATE client_contacts
		SET full_name  = $3,
		    title      = $4,
		    phone      = $5,
		    email      = $6,
		    is_primary = $7,
		    updated_by = $8,
		    updated_at = NOW()
		WHERE id = $1 AND client_id = $2 AND is_deleted = false
		RETURNING id, client_id, full_name, title, phone, email, is_primary,
		          is_deleted, created_at, updated_at, created_by, updated_by`

	cc, err := scanContact(r.pool.QueryRow(ctx, q,
		p.ID, p.ClientID, p.FullName, p.Title, p.Phone, p.Email, p.IsPrimary, p.UpdatedBy,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrContactNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("contact.Update: %w", err)
	}
	return cc, nil
}

// SoftDelete marks a contact as deleted.
// clientID scopes the delete to prevent cross-client tampering.
func (r *ContactRepo) SoftDelete(ctx context.Context, id uuid.UUID, clientID uuid.UUID, deletedBy *uuid.UUID) error {
	const q = `
		UPDATE client_contacts SET is_deleted = true, updated_by = $3, updated_at = NOW()
		WHERE id = $1 AND client_id = $2 AND is_deleted = false`

	tag, err := r.pool.Exec(ctx, q, id, clientID, deletedBy)
	if err != nil {
		return fmt.Errorf("contact.SoftDelete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrContactNotFound
	}
	return nil
}

// ListByClient returns all non-deleted contacts for a client, primary first.
func (r *ContactRepo) ListByClient(ctx context.Context, clientID uuid.UUID) ([]*domain.ClientContact, error) {
	const q = `
		SELECT id, client_id, full_name, title, phone, email, is_primary,
		       is_deleted, created_at, updated_at, created_by, updated_by
		FROM client_contacts
		WHERE client_id = $1 AND is_deleted = false
		ORDER BY is_primary DESC, created_at ASC`

	rows, err := r.pool.Query(ctx, q, clientID)
	if err != nil {
		return nil, fmt.Errorf("contact.ListByClient: %w", err)
	}
	defer rows.Close()

	var contacts []*domain.ClientContact
	for rows.Next() {
		cc, err := scanContact(rows)
		if err != nil {
			return nil, fmt.Errorf("contact.ListByClient scan: %w", err)
		}
		contacts = append(contacts, cc)
	}
	if contacts == nil {
		contacts = []*domain.ClientContact{}
	}
	return contacts, rows.Err()
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func scanContact(row scanner) (*domain.ClientContact, error) {
	var cc domain.ClientContact
	err := row.Scan(
		&cc.ID, &cc.ClientID, &cc.FullName, &cc.Title, &cc.Phone, &cc.Email, &cc.IsPrimary,
		&cc.IsDeleted, &cc.CreatedAt, &cc.UpdatedAt, &cc.CreatedBy, &cc.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	return &cc, nil
}
