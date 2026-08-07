-- name: CreateWebhookSubscription :one
INSERT INTO webhook_subscriptions (integration_id, endpoint, secret, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetWebhookSubscriptionByID :one
SELECT id, integration_id, endpoint, secret, status, last_delivery, created_at, updated_at
FROM webhook_subscriptions
WHERE id = $1;

-- name: GetWebhookSubscriptionByIntegrationIDAndEndpoint :one
SELECT id, integration_id, endpoint, secret, status, last_delivery, created_at, updated_at
FROM webhook_subscriptions
WHERE integration_id = $1 AND endpoint = $2;

-- name: ListWebhookSubscriptionsByIntegrationID :many
SELECT id, integration_id, endpoint, secret, status, last_delivery, created_at, updated_at
FROM webhook_subscriptions
WHERE integration_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActiveWebhookSubscriptionsByIntegrationID :many
SELECT id, integration_id, endpoint, secret, status, last_delivery, created_at, updated_at
FROM webhook_subscriptions
WHERE integration_id = $1 AND status = 'ACTIVE'
ORDER BY created_at DESC;

-- name: CountWebhookSubscriptionsByIntegrationID :one
SELECT COUNT(*) FROM webhook_subscriptions WHERE integration_id = $1;

-- name: WebhookSubscriptionExists :one
SELECT EXISTS(
    SELECT 1 FROM webhook_subscriptions WHERE integration_id = $1 AND endpoint = $2
);

-- name: UpdateWebhookSubscriptionStatus :one
UPDATE webhook_subscriptions
SET status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, integration_id, endpoint, secret, status, last_delivery, created_at, updated_at;

-- name: UpdateWebhookSubscriptionLastDelivery :one
UPDATE webhook_subscriptions
SET last_delivery = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, integration_id, endpoint, secret, status, last_delivery, created_at, updated_at;

-- name: DeleteWebhookSubscription :execrows
DELETE FROM webhook_subscriptions WHERE id = $1;