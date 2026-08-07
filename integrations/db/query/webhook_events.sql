-- name: CreateWebhookEvent :one
INSERT INTO webhook_events (provider, event_type, payload, processed)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetWebhookEventByID :one
SELECT * FROM webhook_events WHERE id = $1;