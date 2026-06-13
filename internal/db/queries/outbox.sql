-- name: CreateOutboxEntry :one
INSERT INTO outbox (event_id, webhook_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetPendingOutboxEntries :many
SELECT * FROM outbox
WHERE status = 'pending'
ORDER BY created_at ASC;

-- name: UpdateOutboxStatus :exec
UPDATE outbox
SET status = $2, updated_at = NOW()
WHERE id = $1;