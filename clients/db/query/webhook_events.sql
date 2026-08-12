-- name: CreateWebhookEvent :one
INSERT INTO webhook_events (integration_id, provider_event_id, event_type, payload)
VALUES ($1, $2, $3, $4)
ON CONFLICT (integration_id, provider_event_id) DO NOTHING
RETURNING *;

-- name: GetWebhookEventByIntegrationAndProvider :one
SELECT id, integration_id, provider_event_id, event_type, payload, received_at, created_at
FROM webhook_events
WHERE integration_id = $1 AND provider_event_id = $2;