-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (user_id, token_hash, device_id, ip_address, user_agent, expires_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetRefreshTokenByHash :one
SELECT id, user_id, token_hash, device_id, ip_address, user_agent,
       expires_at, revoked_at, created_at
FROM refresh_tokens
WHERE token_hash = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = $2
WHERE token_hash = $1
  AND revoked_at IS NULL;

-- name: RevokeAllUserRefreshTokens :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE user_id = $1
  AND revoked_at IS NULL;
