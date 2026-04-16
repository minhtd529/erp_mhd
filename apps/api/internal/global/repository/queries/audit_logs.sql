-- name: CreateAuditLog :exec
INSERT INTO audit_logs (
    user_id, module, resource, resource_id, action,
    old_value, new_value, ip_address, user_agent
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
);

-- name: ListAuditLogs :many
SELECT id, user_id, module, resource, resource_id, action,
       old_value, new_value, ip_address, user_agent, created_at
FROM audit_logs
WHERE module = $1
  AND resource = $2
  AND resource_id = $3
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;
