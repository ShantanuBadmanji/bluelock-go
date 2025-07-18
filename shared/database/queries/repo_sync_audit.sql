-- name: ListActiveRepoSyncAuditOrderBySuccessfulSyncTimeCreatedAt :many
SELECT *
FROM repository_sync_audit
WHERE active = TRUE
ORDER BY successful_sync_time ASC, created_at ASC
LIMIT :limit OFFSET :offset;


-- name: GetRepoSyncAuditByID :one
SELECT *
FROM repository_sync_audit
WHERE id = :id;


-- name: CreateRepoSyncAudit :one
INSERT INTO repository_sync_audit (id, repo_name, successful_sync_time, success, error_context)
VALUES (:id, :repo_name, :successful_sync_time, :success, :error_context)
RETURNING *;


-- name: UpdateRepoSyncAudit :one
UPDATE repository_sync_audit
SET repo_name = :repo_name,
    successful_sync_time = :successful_sync_time,
    success = :success,
    error_context = :error_context,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id
RETURNING *;

-- name: UpdateRepoSyncAuditActiveStatus :one
UPDATE repository_sync_audit
SET active = :active,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id
RETURNING *;


-- name: DeleteInactiveRepoSyncAudit :one
DELETE FROM repository_sync_audit
WHERE id = :id AND active = FALSE
RETURNING *;