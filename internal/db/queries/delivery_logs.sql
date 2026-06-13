-- name: CreatedDeliveryLog :one
INSERT INTO delivery_logs (outbox_id, attempt_number, status_code, response_body, error_message,success)
VALUES ($1,$2,$3,$4,$5,$6)
RETURNING *;

-- name: GetDeliveryLogsByOutbookId :many
SELECT * FROM delivery_logs 
WHERE outbox_id=$1
ORDER BY attempt_number ASC;