package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/crm/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

// ContactUseCase bundles CRUD operations for client contacts.
type ContactUseCase struct {
	repo     domain.ContactRepository
	auditLog *audit.Logger
}

// NewContactUseCase constructs a ContactUseCase.
func NewContactUseCase(repo domain.ContactRepository, auditLog *audit.Logger) *ContactUseCase {
	return &ContactUseCase{repo: repo, auditLog: auditLog}
}

// ListByClient returns all contacts for a client.
func (uc *ContactUseCase) ListByClient(ctx context.Context, clientID uuid.UUID) ([]ContactResponse, error) {
	contacts, err := uc.repo.ListByClient(ctx, clientID)
	if err != nil {
		return nil, err
	}
	items := make([]ContactResponse, len(contacts))
	for i, cc := range contacts {
		items[i] = toContactResponse(cc)
	}
	return items, nil
}

// Create adds a new contact to a client.
func (uc *ContactUseCase) Create(ctx context.Context, clientID uuid.UUID, req ContactCreateRequest, callerID *uuid.UUID, ip string) (*ContactResponse, error) {
	cc, err := uc.repo.Create(ctx, domain.CreateContactParams{
		ClientID:  clientID,
		FullName:  req.FullName,
		Title:     req.Title,
		Phone:     req.Phone,
		Email:     req.Email,
		IsPrimary: req.IsPrimary,
		CreatedBy: callerID,
	})
	if err != nil {
		return nil, err
	}

	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID:     callerID,
		Module:     "crm",
		Resource:   "client_contacts",
		ResourceID: &cc.ID,
		Action:     "CREATE",
		IPAddress:  ip,
	})

	resp := toContactResponse(cc)
	return &resp, nil
}

// Update mutates an existing contact.
func (uc *ContactUseCase) Update(ctx context.Context, clientID, contactID uuid.UUID, req ContactUpdateRequest, callerID *uuid.UUID, ip string) (*ContactResponse, error) {
	cc, err := uc.repo.Update(ctx, domain.UpdateContactParams{
		ID:        contactID,
		ClientID:  clientID,
		FullName:  req.FullName,
		Title:     req.Title,
		Phone:     req.Phone,
		Email:     req.Email,
		IsPrimary: req.IsPrimary,
		UpdatedBy: callerID,
	})
	if err != nil {
		return nil, err
	}

	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID:     callerID,
		Module:     "crm",
		Resource:   "client_contacts",
		ResourceID: &contactID,
		Action:     "UPDATE",
		IPAddress:  ip,
	})

	resp := toContactResponse(cc)
	return &resp, nil
}

// Delete soft-deletes a contact.
func (uc *ContactUseCase) Delete(ctx context.Context, clientID, contactID uuid.UUID, callerID *uuid.UUID, ip string) error {
	if err := uc.repo.SoftDelete(ctx, contactID, clientID, callerID); err != nil {
		return err
	}

	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID:     callerID,
		Module:     "crm",
		Resource:   "client_contacts",
		ResourceID: &contactID,
		Action:     "DELETE",
		IPAddress:  ip,
	})

	return nil
}
