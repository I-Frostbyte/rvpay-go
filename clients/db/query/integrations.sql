-- name: CreateIntegration :one
INSERT INTO integrations (client_id, platform_id, external_account_id, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetIntegrationByID :one
SELECT id, client_id, platform_id, external_account_id, status, installed_at, last_sync_at, created_at, updated_at
FROM integrations
WHERE id = $1;

-- name: GetIntegrationByClientAndPlatform :one
SELECT id, client_id, platform_id, external_account_id, status, installed_at, last_sync_at, created_at, updated_at
FROM integrations
WHERE client_id = $1 AND platform_id = $2;

-- name: GetIntegrationByExternalAccountID :one
SELECT id, client_id, platform_id, external_account_id, status, installed_at, last_sync_at, created_at, updated_at
FROM integrations
WHERE external_account_id = $1;

-- name: ListIntegrationsByClient :many
SELECT id, client_id, platform_id, external_account_id, status, installed_at, last_sync_at, created_at, updated_at
FROM integrations
WHERE client_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListIntegrationsByPlatform :many
SELECT id, client_id, platform_id, external_account_id, status, installed_at, last_sync_at, created_at, updated_at
FROM integrations
WHERE platform_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActiveIntegrationsByClient :many
SELECT id, client_id, platform_id, external_account_id, status, installed_at, last_sync_at, created_at, updated_at
FROM integrations
WHERE client_id = $1 AND status = 'ACTIVE'
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountIntegrationsByClient :one
SELECT COUNT(*) FROM integrations WHERE client_id = $1;

-- name: IntegrationExistsByClientAndPlatform :one
SELECT EXISTS(
    SELECT 1 FROM integrations WHERE client_id = $1 AND platform_id = $2
);

-- name: UpdateIntegrationStatus :one
UPDATE integrations
SET status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, client_id, platform_id, external_account_id, status, installed_at, last_sync_at, created_at, updated_at;

-- name: UpdateIntegrationLastSyncAt :one
UPDATE integrations
SET last_sync_at = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, client_id, platform_id, external_account_id, status, installed_at, last_sync_at, created_at, updated_at;

-- name: DeleteIntegration :execrows
DELETE FROM integrations WHERE id = $1;