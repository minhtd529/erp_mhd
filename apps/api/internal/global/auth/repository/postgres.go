// Package repository provides PostgreSQL implementations of the auth domain
// repository interfaces.  The SQL used here mirrors the sqlc query files in
// ./queries/ — run `make sqlc` to regenerate type-safe wrappers when needed.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
)

// Repo implements UserRepository, RoleRepository, and RefreshTokenRepository
// using a pgxpool.Pool.
type Repo struct {
	pool *pgxpool.Pool
}

// New creates a new Repo.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

// ─── UserRepository ──────────────────────────────────────────────────────────

// FindByEmail looks up a user by email and returns auth-relevant fields.
func (r *Repo) FindByEmail(ctx context.Context, email string) (*domain.UserForAuth, error) {
	const q = `
		SELECT id, email, hashed_password, full_name,
		       branch_id, department_id, status,
		       two_factor_enabled, two_factor_method, two_factor_secret,
		       login_attempt_count, login_locked_until
		FROM users
		WHERE email = $1 AND is_deleted = false`

	u, err := r.scanUser(ctx, q, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("FindByEmail: %w", err)
	}
	if err := r.loadRolesAndPerms(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// FindByID looks up a user by primary key.
func (r *Repo) FindByID(ctx context.Context, id uuid.UUID) (*domain.UserForAuth, error) {
	const q = `
		SELECT id, email, hashed_password, full_name,
		       branch_id, department_id, status,
		       two_factor_enabled, two_factor_method, two_factor_secret,
		       login_attempt_count, login_locked_until
		FROM users
		WHERE id = $1 AND is_deleted = false`

	u, err := r.scanUser(ctx, q, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("FindByID: %w", err)
	}
	if err := r.loadRolesAndPerms(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// CreateUser inserts a new user row and returns the generated UUID.
func (r *Repo) CreateUser(ctx context.Context, p domain.CreateUserParams) (uuid.UUID, error) {
	const q = `
		INSERT INTO users (email, hashed_password, full_name, branch_id, department_id, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	var id uuid.UUID
	err := r.pool.QueryRow(ctx, q,
		p.Email, p.HashedPassword, p.FullName,
		p.BranchID, p.DepartmentID, p.CreatedBy,
	).Scan(&id)
	if err != nil {
		// Postgres unique-violation code = "23505"
		if isUniqueViolation(err) {
			return uuid.Nil, domain.ErrUserAlreadyExists
		}
		return uuid.Nil, fmt.Errorf("CreateUser: %w", err)
	}
	return id, nil
}

// UpdateLastLogin sets last_login_at to NOW() for a user.
func (r *Repo) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE users SET last_login_at = NOW(), updated_at = NOW() WHERE id = $1`
	if _, err := r.pool.Exec(ctx, q, id); err != nil {
		return fmt.Errorf("UpdateLastLogin: %w", err)
	}
	return nil
}

// ─── RoleRepository ──────────────────────────────────────────────────────────

// FindByCode returns a role by its code (e.g. "AUDIT_MANAGER").
func (r *Repo) FindByCode(ctx context.Context, code string) (*domain.Role, error) {
	const q = `SELECT id, code, name FROM roles WHERE code = $1`
	var role domain.Role
	err := r.pool.QueryRow(ctx, q, code).Scan(&role.ID, &role.Code, &role.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrRoleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("FindByCode: %w", err)
	}
	return &role, nil
}

// AssignToUser upserts a user_roles record.
func (r *Repo) AssignToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	const q = `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, role_id) DO NOTHING`
	if _, err := r.pool.Exec(ctx, q, userID, roleID); err != nil {
		return fmt.Errorf("AssignToUser: %w", err)
	}
	return nil
}

// GetUserRoles returns the role codes for a user.
func (r *Repo) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	const q = `
		SELECT r.code
		FROM roles r JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("GetUserRoles: %w", err)
	}
	defer rows.Close()

	var codes []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		codes = append(codes, c)
	}
	return codes, rows.Err()
}

// GetUserPermissions returns flattened "module:resource:action" strings for a user.
func (r *Repo) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	const q = `
		SELECT p.module || ':' || p.resource || ':' || p.action
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		JOIN user_roles       ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = $1`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("GetUserPermissions: %w", err)
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, rows.Err()
}

// ─── RefreshTokenRepository ──────────────────────────────────────────────────

// CreateRefreshToken persists a new refresh token record.
func (r *Repo) CreateRefreshToken(ctx context.Context, t domain.RefreshToken) error {
	const q = `
		INSERT INTO refresh_tokens (user_id, token_hash, device_id, ip_address, user_agent, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	if _, err := r.pool.Exec(ctx, q,
		t.UserID, t.TokenHash, t.DeviceID, t.IPAddress, t.UserAgent, t.ExpiresAt,
	); err != nil {
		return fmt.Errorf("CreateRefreshToken: %w", err)
	}
	return nil
}

// FindByHash retrieves a refresh token by its hash.
func (r *Repo) FindByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	const q = `
		SELECT id, user_id, token_hash, device_id, ip_address, user_agent,
		       expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1`

	var t domain.RefreshToken
	err := r.pool.QueryRow(ctx, q, tokenHash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.DeviceID,
		&t.IPAddress, &t.UserAgent, &t.ExpiresAt, &t.RevokedAt, &t.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTokenInvalid
	}
	if err != nil {
		return nil, fmt.Errorf("FindByHash: %w", err)
	}
	return &t, nil
}

// Revoke sets revoked_at on a refresh token identified by its hash.
func (r *Repo) Revoke(ctx context.Context, tokenHash string, at time.Time) error {
	const q = `
		UPDATE refresh_tokens
		SET revoked_at = $2
		WHERE token_hash = $1 AND revoked_at IS NULL`
	if _, err := r.pool.Exec(ctx, q, tokenHash, at); err != nil {
		return fmt.Errorf("Revoke: %w", err)
	}
	return nil
}

// RevokeAllForUser invalidates all active refresh tokens for a user (e.g. on logout).
func (r *Repo) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	const q = `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL`
	if _, err := r.pool.Exec(ctx, q, userID); err != nil {
		return fmt.Errorf("RevokeAllForUser: %w", err)
	}
	return nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func (r *Repo) scanUser(ctx context.Context, q string, arg any) (*domain.UserForAuth, error) {
	var u domain.UserForAuth
	var twoFactorSecret *string
	err := r.pool.QueryRow(ctx, q, arg).Scan(
		&u.ID, &u.Email, &u.HashedPassword, &u.FullName,
		&u.BranchID, &u.DepartmentID, &u.Status,
		&u.TwoFactorEnabled, &u.TwoFactorMethod, &twoFactorSecret,
		&u.LoginAttemptCount, &u.LoginLockedUntil,
	)
	if err != nil {
		return nil, err
	}
	if twoFactorSecret != nil {
		u.TwoFactorSecret = *twoFactorSecret
	}
	return &u, nil
}

func (r *Repo) loadRolesAndPerms(ctx context.Context, u *domain.UserForAuth) error {
	roles, err := r.GetUserRoles(ctx, u.ID)
	if err != nil {
		return err
	}
	perms, err := r.GetUserPermissions(ctx, u.ID)
	if err != nil {
		return err
	}
	u.Roles = roles
	u.Permissions = perms
	return nil
}

// isUniqueViolation returns true when err is a Postgres unique-constraint violation (23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// ─── User management (non-auth) ──────────────────────────────────────────────

// UpdateUser patches mutable fields on a user row.
func (r *Repo) UpdateUser(ctx context.Context, p domain.UpdateUserParams) error {
	const q = `
		UPDATE users
		SET full_name     = $2,
		    branch_id     = $3,
		    department_id = $4,
		    status        = $5,
		    updated_by    = $6,
		    updated_at    = NOW()
		WHERE id = $1 AND is_deleted = false`

	tag, err := r.pool.Exec(ctx, q, p.ID, p.FullName, p.BranchID, p.DepartmentID, p.Status, p.UpdatedBy)
	if err != nil {
		return fmt.Errorf("UpdateUser: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// SoftDeleteUser marks a user as deleted.
func (r *Repo) SoftDeleteUser(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	const q = `
		UPDATE users SET is_deleted = true, updated_by = $2, updated_at = NOW()
		WHERE id = $1 AND is_deleted = false`
	tag, err := r.pool.Exec(ctx, q, id, deletedBy)
	if err != nil {
		return fmt.Errorf("SoftDeleteUser: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// ListUsers returns a paginated list of non-deleted users.
func (r *Repo) ListUsers(ctx context.Context, f domain.ListUsersFilter) ([]*domain.UserForAuth, int64, error) {
	offset := (f.Page - 1) * f.Size
	args := []any{}
	where := "WHERE is_deleted = false"
	idx := 1

	if f.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, f.Status)
		idx++
	}
	if f.Q != "" {
		where += fmt.Sprintf(" AND (full_name ILIKE $%d OR email ILIKE $%d)", idx, idx)
		args = append(args, "%"+f.Q+"%")
		idx++
	}

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("ListUsers count: %w", err)
	}

	args = append(args, f.Size, offset)
	dataQ := fmt.Sprintf(`
		SELECT id, email, hashed_password, full_name,
		       branch_id, department_id, status,
		       two_factor_enabled, two_factor_method, two_factor_secret,
		       login_attempt_count, login_locked_until
		FROM users %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, where, idx, idx+1)

	rows, err := r.pool.Query(ctx, dataQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("ListUsers query: %w", err)
	}
	defer rows.Close()

	var users []*domain.UserForAuth
	for rows.Next() {
		var u domain.UserForAuth
		var twoFactorSecret *string
		if err := rows.Scan(
			&u.ID, &u.Email, &u.HashedPassword, &u.FullName,
			&u.BranchID, &u.DepartmentID, &u.Status,
			&u.TwoFactorEnabled, &u.TwoFactorMethod, &twoFactorSecret,
			&u.LoginAttemptCount, &u.LoginLockedUntil,
		); err != nil {
			return nil, 0, fmt.Errorf("ListUsers scan: %w", err)
		}
		if twoFactorSecret != nil {
			u.TwoFactorSecret = *twoFactorSecret
		}
		users = append(users, &u)
	}
	if users == nil {
		users = []*domain.UserForAuth{}
	}
	return users, total, rows.Err()
}
