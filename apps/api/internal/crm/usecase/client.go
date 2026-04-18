// Package usecase implements the CRM application layer.
package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/crm/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

// ClientUseCase bundles all client CRUD operations.
type ClientUseCase struct {
	repo     domain.ClientRepository
	auditLog *audit.Logger
}

// NewClientUseCase constructs a ClientUseCase.
func NewClientUseCase(repo domain.ClientRepository, auditLog *audit.Logger) *ClientUseCase {
	return &ClientUseCase{repo: repo, auditLog: auditLog}
}

// Create creates a new client in PROSPECT status.
func (uc *ClientUseCase) Create(ctx context.Context, req ClientCreateRequest, callerID uuid.UUID, ip string) (*ClientResponse, error) {
	c, err := uc.repo.Create(ctx, domain.CreateClientParams{
		TaxCode:             req.TaxCode,
		BusinessName:        req.BusinessName,
		EnglishName:         req.EnglishName,
		Industry:            req.Industry,
		OfficeID:            req.OfficeID,
		SalesOwnerID:        req.SalesOwnerID,
		ReferrerID:          req.ReferrerID,
		Address:             req.Address,
		BankName:            req.BankName,
		BankAccountNumber:   req.BankAccountNumber,
		BankAccountName:     req.BankAccountName,
		RepresentativeName:  req.RepresentativeName,
		RepresentativeTitle: req.RepresentativeTitle,
		RepresentativePhone: req.RepresentativePhone,
		CreatedBy:           callerID,
	})
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID:     &callerID,
		Module:     "crm",
		Resource:   "clients",
		ResourceID: &c.ID,
		Action:     "CREATE",
		IPAddress:  ip,
	})

	resp := toClientResponse(c)
	return &resp, nil
}

// GetByID retrieves a single client.
func (uc *ClientUseCase) GetByID(ctx context.Context, id uuid.UUID) (*ClientResponse, error) {
	c, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toClientResponse(c)
	return &resp, nil
}

// Update mutates allowed fields on an existing client.
func (uc *ClientUseCase) Update(ctx context.Context, id uuid.UUID, req ClientUpdateRequest, callerID uuid.UUID, ip string) (*ClientResponse, error) {
	c, err := uc.repo.Update(ctx, domain.UpdateClientParams{
		ID:                  id,
		BusinessName:        req.BusinessName,
		EnglishName:         req.EnglishName,
		Industry:            req.Industry,
		OfficeID:            req.OfficeID,
		SalesOwnerID:        req.SalesOwnerID,
		ReferrerID:          req.ReferrerID,
		Address:             req.Address,
		BankName:            req.BankName,
		BankAccountNumber:   req.BankAccountNumber,
		BankAccountName:     req.BankAccountName,
		RepresentativeName:  req.RepresentativeName,
		RepresentativeTitle: req.RepresentativeTitle,
		RepresentativePhone: req.RepresentativePhone,
		UpdatedBy:           callerID,
	})
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID:     &callerID,
		Module:     "crm",
		Resource:   "clients",
		ResourceID: &id,
		Action:     "UPDATE",
		IPAddress:  ip,
	})

	resp := toClientResponse(c)
	return &resp, nil
}

// Delete soft-deletes a client.
func (uc *ClientUseCase) Delete(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip string) error {
	if err := uc.repo.SoftDelete(ctx, id, &callerID); err != nil {
		return err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID:     &callerID,
		Module:     "crm",
		Resource:   "clients",
		ResourceID: &id,
		Action:     "DELETE",
		IPAddress:  ip,
	})

	return nil
}

// List returns a paginated list of clients.
func (uc *ClientUseCase) List(ctx context.Context, req ClientListRequest) (PaginatedResult[ClientResponse], error) {
	clients, total, err := uc.repo.List(ctx, domain.ListClientsFilter{
		Page:         req.Page,
		Size:         req.Size,
		Status:       req.Status,
		Q:            req.Q,
		SalesOwnerID: req.SalesOwnerID,
		Industry:     req.Industry,
		OfficeID:     req.OfficeID,
	})
	if err != nil {
		return PaginatedResult[ClientResponse]{}, fmt.Errorf("crm.List: %w", err)
	}

	items := make([]ClientResponse, len(clients))
	for i, c := range clients {
		items[i] = toClientResponse(c)
	}
	return newPaginatedResult(items, total, req.Page, req.Size), nil
}
