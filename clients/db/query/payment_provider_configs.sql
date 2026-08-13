-- name: CreatePaymentProviderConfig :one
INSERT INTO payment_provider_configs (
    integration_id,
    provider_name,
    provider_description,
    provider_image_url,
    location_id,
    query_url,
    payments_url,
    supports_subscription_schedule,
    provider_api_key
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetPaymentProviderConfigByIntegrationID :one
SELECT id, integration_id, provider_name, provider_description, provider_image_url, location_id, query_url, payments_url, supports_subscription_schedule, provider_api_key, created_at, updated_at
FROM payment_provider_configs
WHERE integration_id = $1;

-- name: GetPaymentProviderConfigByLocationID :one
SELECT id, integration_id, provider_name, provider_description, provider_image_url, location_id, query_url, payments_url, supports_subscription_schedule, provider_api_key, created_at, updated_at
FROM payment_provider_configs
WHERE location_id = $1;

-- name: GetPaymentProviderConfigByAPIKey :one
SELECT id, integration_id, provider_name, provider_description, provider_image_url, location_id, query_url, payments_url, supports_subscription_schedule, provider_api_key, created_at, updated_at
FROM payment_provider_configs
WHERE provider_api_key = $1;

-- name: UpdatePaymentProviderConfig :one
UPDATE payment_provider_configs
SET provider_name = $2,
    provider_description = $3,
    provider_image_url = $4,
    location_id = $5,
    query_url = $6,
    payments_url = $7,
    supports_subscription_schedule = $8,
    provider_api_key = $9,
    updated_at = NOW()
WHERE integration_id = $1
RETURNING id, integration_id, provider_name, provider_description, provider_image_url, location_id, query_url, payments_url, supports_subscription_schedule, provider_api_key, created_at, updated_at;

-- name: DeletePaymentProviderConfig :execrows
DELETE FROM payment_provider_configs WHERE integration_id = $1;