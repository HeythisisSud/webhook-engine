-- name: CreateEvents :one
INSERT INTO events (client_id,event_type, payload)
VALUES ($1,$2,$3)
RETURNING *;

-- name: GetEvent :one
SELECT * FROM events
WHERE id=$1;