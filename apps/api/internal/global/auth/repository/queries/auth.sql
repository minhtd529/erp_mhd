-- name: GetUserForAuth :one
-- Returns all auth-relevant columns for a user by email.
SELECT
    u.id,
    u.email,
    u.hashed_password,
    u.full_name,
    u.branch_id,
    u.department_id,
    u.status,
    u.two_factor_enabled,
    u.two_factor_method
FROM users u
WHERE u.email = $1
  AND u.is_deleted = false;

-- name: GetUserForAuthByID :one
SELECT
    u.id,
    u.email,
    u.hashed_password,
    u.full_name,
    u.branch_id,
    u.department_id,
    u.status,
    u.two_factor_enabled,
    u.two_factor_method
FROM users u
WHERE u.id = $1
  AND u.is_deleted = false;

-- name: CreateUser :one
INSERT INTO users (
    email, hashed_password, full_name,
    branch_id, department_id, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING id;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: GetUserRoleCodes :many
-- Returns role codes for a user (e.g. ["AUDIT_MANAGER"]).
SELECT r.code
FROM roles r
JOIN user_roles ur ON ur.role_id = r.id
WHERE ur.user_id = $1;

-- name: GetUserPermissions :many
-- Returns flattened "module:resource:action" strings for a user.
SELECT p.module || ':' || p.resource || ':' || p.action AS permission
FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.id
JOIN user_roles ur       ON ur.role_id = rp.role_id
WHERE ur.user_id = $1;

-- name: GetRoleByCode :one
SELECT id, code, name
FROM roles
WHERE code = $1;

-- name: UpsertUserRole :exec
INSERT INTO user_roles (user_id, role_id)
VALUES ($1, $2)
ON CONFLICT (user_id, role_id) DO NOTHING;

-- name: ListUsers :many
SELECT
    id, email, full_name, branch_id, department_id,
    status, two_factor_enabled, created_at, updated_at
FROM users
WHERE is_deleted = false
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users WHERE is_deleted = false;
