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
	ID                  uuid.UUID    `json:"id"                    db:"id"`
	TaxCode             string       `json:"tax_code"              db:"tax_code"`
	BusinessName        string       `json:"business_name"         db:"business_name"`
	EnglishName         *string      `json:"english_name"          db:"english_name"`
	Industry            *string      `json:"industry"              db:"industry"`
	Status              ClientStatus `json:"status"                db:"status"`
	OfficeID            uuid.UUID    `json:"office_id"             db:"office_id"`
	SalesOwnerID        *uuid.UUID   `json:"sales_owner_id"        db:"sales_owner_id"`
	ReferrerID          *uuid.UUID   `json:"referrer_id"           db:"referrer_id"`
	Address             string       `json:"address"               db:"address"`
	BankName            *string      `json:"bank_name"             db:"bank_name"`
	BankAccountNumber   *string      `json:"bank_account_number"   db:"bank_account_number"`
	BankAccountName     *string      `json:"bank_account_name"     db:"bank_account_name"`
	RepresentativeName  *string      `json:"representative_name"   db:"representative_name"`
	RepresentativeTitle *string      `json:"representative_title"  db:"representative_title"`
	RepresentativePhone *string      `json:"representative_phone"  db:"representative_phone"`
	IsDeleted           bool         `json:"is_deleted"            db:"is_deleted"`
	CreatedAt           time.Time    `json:"created_at"            db:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at"            db:"updated_at"`
	CreatedBy           uuid.UUID    `json:"created_by"            db:"created_by"`
	UpdatedBy           uuid.UUID    `json:"updated_by"            db:"updated_by"`
}

// ClientContact is an external contact person for a client.
type ClientContact struct {
	ID        uuid.UUID  `json:"id"         db:"id"`
	ClientID  uuid.UUID  `json:"client_id"  db:"client_id"`
	FullName  string     `json:"full_name"  db:"full_name"`
	Title     *string    `json:"title"      db:"title"`
	Phone     *string    `json:"phone"      db:"phone"`
	Email     *string    `json:"email"      db:"email"`
	IsPrimary bool       `json:"is_primary" db:"is_primary"`
	IsDeleted bool       `json:"is_deleted" db:"is_deleted"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy uuid.UUID  `json:"created_by" db:"created_by"`
	UpdatedBy uuid.UUID  `json:"updated_by" db:"updated_by"`
}

// CreateContactParams holds parameters for creating a client contact.
type CreateContactParams struct {
	ClientID  uuid.UUID
	FullName  string
	Title     *string
	Phone     *string
	Email     *string
	IsPrimary bool
	CreatedBy uuid.UUID
}

// UpdateContactParams holds parameters for updating a client contact.
type UpdateContactParams struct {
	ID        uuid.UUID
	ClientID  uuid.UUID
	FullName  string
	Title     *string
	Phone     *string
	Email     *string
	IsPrimary bool
	UpdatedBy uuid.UUID
}

// CreateClientParams holds the parameters needed to create a new client.
type CreateClientParams struct {
	TaxCode             string
	BusinessName        string
	EnglishName         *string
	Industry            *string
	OfficeID            uuid.UUID
	SalesOwnerID        *uuid.UUID
	ReferrerID          *uuid.UUID
	Address             string
	BankName            *string
	BankAccountNumber   *string
	BankAccountName     *string
	RepresentativeName  *string
	RepresentativeTitle *string
	RepresentativePhone *string
	CreatedBy           uuid.UUID
}

// UpdateClientParams holds the fields that can be changed after creation.
type UpdateClientParams struct {
	ID                  uuid.UUID
	BusinessName        string
	EnglishName         *string
	Industry            *string
	OfficeID            uuid.UUID
	SalesOwnerID        *uuid.UUID
	ReferrerID          *uuid.UUID
	Address             string
	BankName            *string
	BankAccountNumber   *string
	BankAccountName     *string
	RepresentativeName  *string
	RepresentativeTitle *string
	RepresentativePhone *string
	UpdatedBy           uuid.UUID
}

// ListClientsFilter controls pagination and filtering for the list query.
type ListClientsFilter struct {
	Page   int
	Size   int
	Status ClientStatus
	Q      string // free-text search on business_name / tax_code
}
