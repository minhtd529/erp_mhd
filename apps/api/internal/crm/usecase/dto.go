package usecase

import (
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/crm/domain"
)

// ClientCreateRequest is the body for POST /api/v1/clients.
type ClientCreateRequest struct {
	TaxCode           string     `json:"tax_code"            binding:"required,min=10,max=14"`
	BusinessName      string     `json:"business_name"       binding:"required"`
	EnglishName       *string    `json:"english_name"`
	Industry          *string    `json:"industry"`
	OfficeID          *uuid.UUID `json:"office_id"`
	SalesOwnerID      *uuid.UUID `json:"sales_owner_id"`
	ReferrerID        *uuid.UUID `json:"referrer_id"`
	BankName          *string    `json:"bank_name"`
	BankAccountNumber *string    `json:"bank_account_number"`
	BankAccountName   *string    `json:"bank_account_name"`
}

// ClientUpdateRequest is the body for PUT /api/v1/clients/:id.
type ClientUpdateRequest struct {
	BusinessName      string     `json:"business_name"       binding:"required"`
	EnglishName       *string    `json:"english_name"`
	Industry          *string    `json:"industry"`
	OfficeID          *uuid.UUID `json:"office_id"`
	SalesOwnerID      *uuid.UUID `json:"sales_owner_id"`
	ReferrerID        *uuid.UUID `json:"referrer_id"`
	BankName          *string    `json:"bank_name"`
	BankAccountNumber *string    `json:"bank_account_number"`
	BankAccountName   *string    `json:"bank_account_name"`
}

// ClientListRequest holds validated query parameters for GET /api/v1/clients.
type ClientListRequest struct {
	Page   int                  `form:"page,default=1"    binding:"min=1"`
	Size   int                  `form:"size,default=20"   binding:"min=1,max=100"`
	Status domain.ClientStatus  `form:"status"`
	Q      string               `form:"q"`
}

// ClientResponse is the JSON representation of a client.
type ClientResponse struct {
	ID                uuid.UUID           `json:"id"`
	TaxCode           string              `json:"tax_code"`
	BusinessName      string              `json:"business_name"`
	EnglishName       *string             `json:"english_name"`
	Industry          *string             `json:"industry"`
	Status            domain.ClientStatus `json:"status"`
	OfficeID          *uuid.UUID          `json:"office_id"`
	SalesOwnerID      *uuid.UUID          `json:"sales_owner_id"`
	ReferrerID        *uuid.UUID          `json:"referrer_id"`
	BankName          *string             `json:"bank_name"`
	BankAccountNumber *string             `json:"bank_account_number"`
	BankAccountName   *string             `json:"bank_account_name"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
	CreatedBy         *uuid.UUID          `json:"created_by"`
	UpdatedBy         *uuid.UUID          `json:"updated_by"`
}

// PaginatedResult is a generic paginated wrapper (local copy to avoid import cycle).
type PaginatedResult[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Size       int   `json:"size"`
	TotalPages int   `json:"total_pages"`
}

func newPaginatedResult[T any](data []T, total int64, page, size int) PaginatedResult[T] {
	tp := int(total) / size
	if int(total)%size != 0 {
		tp++
	}
	return PaginatedResult[T]{Data: data, Total: total, Page: page, Size: size, TotalPages: tp}
}

func toClientResponse(c *domain.Client) ClientResponse {
	return ClientResponse{
		ID: c.ID, TaxCode: c.TaxCode, BusinessName: c.BusinessName,
		EnglishName: c.EnglishName, Industry: c.Industry, Status: c.Status,
		OfficeID: c.OfficeID, SalesOwnerID: c.SalesOwnerID, ReferrerID: c.ReferrerID,
		BankName: c.BankName, BankAccountNumber: c.BankAccountNumber, BankAccountName: c.BankAccountName,
		CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
		CreatedBy: c.CreatedBy, UpdatedBy: c.UpdatedBy,
	}
}
