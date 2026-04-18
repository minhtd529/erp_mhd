package auth

// Gin context keys used by AuthMiddleware and consumed by handlers / use cases.
// Using typed string constants prevents accidental key collisions.
const (
	CtxUserID       = "auth.user_id"       // uuid.UUID
	CtxEmail        = "auth.email"         // string
	CtxRoles        = "auth.roles"         // []string
	CtxPerms        = "auth.perms"         // []string
	CtxBranchID     = "auth.branch_id"     // *uuid.UUID
	CtxDeptID       = "auth.dept_id"       // *uuid.UUID
	CtxClaims       = "auth.claims"        // *auth.TokenClaims
	CtxTwoFAVerified = "auth.2fa_verified" // bool
)
