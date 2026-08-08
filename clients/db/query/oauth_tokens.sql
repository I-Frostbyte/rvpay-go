-- name: CreateOAuthToken :one
INSERT INTO oauth_tokens (integration_id, access_token, refresh_token, expires_at, scope, token_type)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetOAuthTokenByID :one
SELECT id, integration_id, access_token, refresh_token, expires_at, scope, token_type, created_at, updated_at
FROM oauth_tokens
WHERE id = $1;

-- name: GetOAuthTokenByIntegrationID :one
SELECT id, integration_id, access_token, refresh_token, expires_at, scope, token_type, created_at, updated_at
FROM oauth_tokens
WHERE integration_id = $1;

-- name: OAuthTokenExistsByIntegrationID :one
SELECT EXISTS(
    SELECT 1 FROM oauth_tokens WHERE integration_id = $1
);

-- name: UpdateOAuthToken :one
UPDATE oauth_tokens
SET access_token = $2,
    refresh_token = $3,
    expires_at = $4,
    scope = $5,
    token_type = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING id, integration_id, access_token, refresh_token, expires_at, scope, token_type, created_at, updated_at;

-- name: DeleteOAuthToken :execrows
DELETE FROM oauth_tokens WHERE id = $1;

-- name: DeleteOAuthTokenByIntegrationID :execrows
DELETE FROM oauth_tokens WHERE integration_id = $1;