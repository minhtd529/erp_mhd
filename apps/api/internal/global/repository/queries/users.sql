-- name: GetUserByID :one
SELECT id, email, hashed_password, full_name, employee_id, branch_id, department_id,
       status, last_login_at, two_factor_enabled, two_factor_method, two_factor_secret,
       two_factor_verified_at, backup_codes_hash, is_deleted, created_at, updated_at,
       created_by, updated_by
FROM users
WHERE id = $1 AND is_deleted = false;

-- name: GetUserByEmail :one
SELECT id, email, hashed_password, full_name, employee_id, branch_id, department_id,
       status, last_login_at, two_factor_enabled, two_factor_method, two_factor_secret,
       two_factor_verified_at, backup_codes_hash, is_deleted, created_at, updated_at,
       created_by, updated_by
FROM users
WHERE email = $1 AND is_deleted = false;

-- name: CreateUser :one
INSERT INTO users (
    email, hashed_password, full_name, branch_id, department_id,
    status, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateUserStatus :exec
UPDATE users
SET status = $2, updated_at = NOW(), updated_by = $3
WHERE id = $1 AND is_deleted = false;

-- name: SoftDeleteUser :exec
UPDATE users
SET is_deleted = true, updated_at = NOW(), updated_by = $2
WHERE id = $1;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: ListUsers :many
SELECT id, email, full_name, branch_id, department_id, status, last_login_at,
       two_factor_enabled, created_at, updated_at
FROM users
WHERE is_deleted = false
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users WHERE is_deleted = false;
