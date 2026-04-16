package auth

import "github.com/google/uuid"

// TokenClaims are the JWT payload fields embedded in every access token.
// Keep small: only data needed for routing decisions without a DB round-trip.
type TokenClaims struct {
	UserID       uuid.UUID  `json:"uid"`
	Email        string     `json:"email"`
	Roles        []string   `json:"roles"`        // e.g. ["AUDIT_MANAGER","AUDIT_STAFF"]
	Permissions  []string   `json:"perms"`        // e.g. ["crm:client:read"]
	BranchID     *uuid.UUID `json:"branch_id,omitempty"`
	DepartmentID *uuid.UUID `json:"dept_id,omitempty"`
}
