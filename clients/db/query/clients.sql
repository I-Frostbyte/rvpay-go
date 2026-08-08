-- name: CreateClient :one
INSERT INTO clients (client_name, status)
VALUES ($1, $2)
RETURNING *;

-- name: GetClientByID :one
SELECT id, client_name, status, created_at, updated_at
FROM clients
WHERE id = $1;

-- name: GetClientByName :one
SELECT id, client_name, status, created_at, updated_at
FROM clients
WHERE client_name = $1;

-- name: ListClients :many
SELECT id, client_name, status, created_at, updated_at
FROM clients
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListActiveClients :many
SELECT id, client_name, status, created_at, updated_at
FROM clients
WHERE status = 'ACTIVE'
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountClients :one
SELECT COUNT(*) FROM clients;

-- name: ClientExistsByID :one
SELECT EXISTS(
    SELECT 1 FROM clients WHERE id = $1
);

-- name: UpdateClientStatus :one
UPDATE clients
SET status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, client_name, status, created_at, updated_at;

-- name: DeleteClient :execrows
DELETE FROM clients WHERE id = $1;