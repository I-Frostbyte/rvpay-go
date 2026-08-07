-- name: CreateIntegration :one
INSERT INTO integrations (provider, location_id, access_token, refresh_token, token_expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetIntegrationByID :one
SELECT * FROM integrations WHERE id = $1;

-- name: DeleteIntegration :execrows
DELETE FROM integrations WHERE id = $1;
