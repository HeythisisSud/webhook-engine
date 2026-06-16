-- name: CreateWebhook :one
INSERT INTO webhooks (client_id, target_url, secret, event_types)
VALUES ($1,$2,$3,$4)
RETURNING *;

-- name: GetWebhook :one
SELECT * FROM webhooks
WHERE id=$1;

-- name: GetWebhooksByClientID :many
SELECT * FROM webhooks
WHERE client_id= $1;

-- name: GetWebhooksByEventType :many
SELECT * FROM webhooks
WHERE event_types && ARRAY[$1::text] AND enabled = true;

-- name: DeleteWebhook :exec
DELETE FROM webhooks
WHERE id=$1;

-- name: UpdateWebhook :one
UPDATE webhooks
SET target_url= $2, event_types=$3, enabled=$4, updated_at=NOW()
WHERE id=$1
RETURNING *;

