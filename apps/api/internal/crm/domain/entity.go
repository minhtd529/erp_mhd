package domain

import (
	"time"

	"github.com/google/uuid"
)

// ClientStatus represents the lifecycle stage of a client.
type ClientStatus string

const (
	ClientStatusProspect   ClientStatus = "PROSPECT"
	ClientStatusAssessment ClientStatus = "ASSESSMENT"
	ClientStatusAccepted   ClientStatus = "ACCEPTED"
	ClientStatusInactive   ClientStatus = "INACTIVE"
)

// Client is the CRM aggregate root.
type Client struct {
	ID                 uuid.UUID    `json:"id"                   db:"id"`
	TaxCode            string       `json:"tax_code"             db:"tax_code"`
	BusinessName       string       `json:"business_name"        db:"business_name"`
	EnglishName        *string      `json:"english_name"         db:"english_name"`
	Industry           *string      `json:"industry"             db:"industry"`
	Status             ClientStatus `json:"status"               db:"status"`
	OfficeID           *uuid.UUID   `json:"office_id"            db:"office_id"`
	SalesOwnerID       *uuid.UUID   `json:"sales_owner_id"       db:"sales_owner_id"`
	ReferrerID         *uuid.UUID   `json:"referrer_id"          db:"referrer_id"`
	BankName           *string      `json:"bank_name"            db:"bank_name"`
	BankAccountNumber  *string      `json:"bank_account_number"  db:"bank_account_number"`
	BankAccountName    *string      `json:"bank_account_name"    db:"bank_account_name"`
	IsDeleted          bool         `json:"is_deleted"           db:"is_deleted"`
	CreatedAt          time.Time    `json:"created_at"           db:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at"           db:"updated_at"`
	CreatedBy          *uuid.UUID   `json:"created_by"           db:"created_by"`
	UpdatedBy          *uuid.UUID   `json:"updated_by"           db:"updated_by"`
}

// CreateClientParams holds the parameters needed to create a new client.
type CreateClientParams struct {
	TaxCode            string
	BusinessName       string
	EnglishName        *string
	Industry           *string
	OfficeID           *uuid.UUID
	SalesOwnerID       *uuid.UUID
	ReferrerID         *uuid.UUID
	BankName           *string
	BankAccountNumber  *string
	BankAccountName    *string
	CreatedBy          *uuid.UUID
}

// UpdateClientParams holds the fields that can be changed after creation.
type UpdateClientParams struct {
	ID                 uuid.UUID
	BusinessName       string
	EnglishName        *string
	Industry           *string
	OfficeID           *uuid.UUID
	SalesOwnerID       *uuid.UUID
	ReferrerID         *uuid.UUID
	BankName           *string
	BankAccountNumber  *string
	BankAccountName    *string
	UpdatedBy          *uuid.UUID
}

// ListClientsFilter controls pagination and filtering for the list query.
type ListClientsFilter struct {
	Page   int
	Size   int
	Status ClientStatus
	Q      string // free-text search on business_name / tax_code
}
