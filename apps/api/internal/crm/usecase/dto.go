package usecase

import (
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/crm/domain"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

// ClientCreateRequest is the body for POST /api/v1/clients.
type ClientCreateRequest struct {
	TaxCode             string     `json:"tax_code"             binding:"required,min=10,max=14"`
	BusinessName        string     `json:"business_name"        binding:"required"`
	EnglishName         *string    `json:"english_name"`
	Industry            *string    `json:"industry"`
	OfficeID            uuid.UUID  `json:"office_id"            binding:"required"`
	SalesOwnerID        *uuid.UUID `json:"sales_owner_id"`
	ReferrerID          *uuid.UUID `json:"referrer_id"`
	Address             string     `json:"address"              binding:"required"`
	BankName            *string    `json:"bank_name"`
	BankAccountNumber   *string    `json:"bank_account_number"`
	BankAccountName     *string    `json:"bank_account_name"`
	RepresentativeName  *string    `json:"representative_name"`
	RepresentativeTitle *string    `json:"representative_title"`
	RepresentativePhone *string    `json:"representative_phone"`
}

// ClientUpdateRequest is the body for PUT /api/v1/clients/:id.
type ClientUpdateRequest struct {
	BusinessName        string     `json:"business_name"        binding:"required"`
	EnglishName         *string    `json:"english_name"`
	Industry            *string    `json:"industry"`
	OfficeID            uuid.UUID  `json:"office_id"            binding:"required"`
	SalesOwnerID        *uuid.UUID `json:"sales_owner_id"`
	ReferrerID          *uuid.UUID `json:"referrer_id"`
	Address             string     `json:"address"              binding:"required"`
	BankName            *string    `json:"bank_name"`
	BankAccountNumber   *string    `json:"bank_account_number"`
	BankAccountName     *string    `json:"bank_account_name"`
	RepresentativeName  *string    `json:"representative_name"`
	RepresentativeTitle *string    `json:"representative_title"`
	RepresentativePhone *string    `json:"representative_phone"`
}

// ClientListRequest holds validated query parameters for GET /api/v1/clients.
type ClientListRequest struct {
	Page         int                 `form:"page,default=1"    binding:"min=1"`
	Size         int                 `form:"size,default=20"   binding:"min=1,max=100"`
	Status       domain.ClientStatus `form:"status"`
	Q            string              `form:"q"`
	SalesOwnerID *uuid.UUID          `form:"sales_owner_id"`
	Industry     *string             `form:"industry"`
	OfficeID     *uuid.UUID          `form:"office_id"`
}

// ClientResponse is the JSON representation of a client.
type ClientResponse struct {
	ID                  uuid.UUID           `json:"id"`
	TaxCode             string              `json:"tax_code"`
	BusinessName        string              `json:"business_name"`
	EnglishName         *string             `json:"english_name"`
	Industry            *string             `json:"industry"`
	Status              domain.ClientStatus `json:"status"`
	OfficeID            uuid.UUID           `json:"office_id"`
	SalesOwnerID        *uuid.UUID          `json:"sales_owner_id"`
	ReferrerID          *uuid.UUID          `json:"referrer_id"`
	Address             string              `json:"address"`
	BankName            *string             `json:"bank_name"`
	BankAccountNumber   *string             `json:"bank_account_number"`
	BankAccountName     *string             `json:"bank_account_name"`
	RepresentativeName  *string             `json:"representative_name"`
	RepresentativeTitle *string             `json:"representative_title"`
	RepresentativePhone *string             `json:"representative_phone"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
	CreatedBy           uuid.UUID           `json:"created_by"`
	UpdatedBy           uuid.UUID           `json:"updated_by"`
}

// ContactCreateRequest is the body for POST /api/v1/clients/:id/contacts.
type ContactCreateRequest struct {
	FullName  string  `json:"full_name" binding:"required"`
	Title     *string `json:"title"`
	Phone     *string `json:"phone"`
	Email     *string `json:"email"     binding:"omitempty,email"`
	IsPrimary bool    `json:"is_primary"`
}

// ContactUpdateRequest is the body for PUT /api/v1/clients/:id/contacts/:cid.
type ContactUpdateRequest struct {
	FullName  string  `json:"full_name" binding:"required"`
	Title     *string `json:"title"`
	Phone     *string `json:"phone"`
	Email     *string `json:"email"     binding:"omitempty,email"`
	IsPrimary bool    `json:"is_primary"`
}

// ContactResponse is the JSON representation of a client contact.
type ContactResponse struct {
	ID        uuid.UUID  `json:"id"`
	ClientID  uuid.UUID  `json:"client_id"`
	FullName  string     `json:"full_name"`
	Title     *string    `json:"title"`
	Phone     *string    `json:"phone"`
	Email     *string    `json:"email"`
	IsPrimary bool      `json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy uuid.UUID `json:"created_by"`
	UpdatedBy uuid.UUID `json:"updated_by"`
}

// PaginatedResult is the shared offset pagination type.
type PaginatedResult[T any] = pagination.OffsetResult[T]

func newPaginatedResult[T any](data []T, total int64, page, size int) PaginatedResult[T] {
	return pagination.NewOffsetResult(data, total, page, size)
}

func toClientResponse(c *domain.Client) ClientResponse {
	return ClientResponse{
		ID: c.ID, TaxCode: c.TaxCode, BusinessName: c.BusinessName,
		EnglishName: c.EnglishName, Industry: c.Industry, Status: c.Status,
		OfficeID: c.OfficeID, SalesOwnerID: c.SalesOwnerID, ReferrerID: c.ReferrerID,
		Address: c.Address,
		BankName: c.BankName, BankAccountNumber: c.BankAccountNumber, BankAccountName: c.BankAccountName,
		RepresentativeName: c.RepresentativeName, RepresentativeTitle: c.RepresentativeTitle,
		RepresentativePhone: c.RepresentativePhone,
		CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
		CreatedBy: c.CreatedBy, UpdatedBy: c.UpdatedBy,
	}
}

func toContactResponse(cc *domain.ClientContact) ContactResponse {
	return ContactResponse{
		ID: cc.ID, ClientID: cc.ClientID, FullName: cc.FullName,
		Title: cc.Title, Phone: cc.Phone, Email: cc.Email, IsPrimary: cc.IsPrimary,
		CreatedAt: cc.CreatedAt, UpdatedAt: cc.UpdatedAt,
		CreatedBy: cc.CreatedBy, UpdatedBy: cc.UpdatedBy,
	}
}
