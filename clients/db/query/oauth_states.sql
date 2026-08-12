-- name: CreateOAuthState :one
INSERT INTO oauth_states (state, client_id, platform_id, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetOAuthStateByState :one
SELECT id, state, client_id, platform_id, expires_at, consumed_at, created_at, updated_at
FROM oauth_states
WHERE state = $1;

-- name: ConsumeOAuthState :one
UPDATE oauth_states
SET consumed_at = NOW(),
    updated_at = NOW()
WHERE state = $1
  AND consumed_at IS NULL
  AND expires_at > NOW()
RETURNING id, state, client_id, platform_id, expires_at, consumed_at, created_at, updated_at;

-- name: DeleteExpiredOAuthStates :execrows
DELETE FROM oauth_states WHERE expires_at <= NOW();