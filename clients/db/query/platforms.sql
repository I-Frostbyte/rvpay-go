-- name: CreatePlatform :one
INSERT INTO platforms (name, display_name, slug, enabled, oauth_capable, webhook_capable)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetPlatformByID :one
SELECT id, name, display_name, slug, enabled, oauth_capable, webhook_capable, created_at, updated_at
FROM platforms
WHERE id = $1;

-- name: GetPlatformByName :one
SELECT id, name, display_name, slug, enabled, oauth_capable, webhook_capable, created_at, updated_at
FROM platforms
WHERE name = $1;

-- name: GetPlatformBySlug :one
SELECT id, name, display_name, slug, enabled, oauth_capable, webhook_capable, created_at, updated_at
FROM platforms
WHERE slug = $1;

-- name: ListPlatforms :many
SELECT id, name, display_name, slug, enabled, oauth_capable, webhook_capable, created_at, updated_at
FROM platforms
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListEnabledPlatforms :many
SELECT id, name, display_name, slug, enabled, oauth_capable, webhook_capable, created_at, updated_at
FROM platforms
WHERE enabled = TRUE
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountPlatforms :one
SELECT COUNT(*) FROM platforms;

-- name: PlatformExistsBySlug :one
SELECT EXISTS(
    SELECT 1 FROM platforms WHERE slug = $1
);

-- name: UpdatePlatform :one
UPDATE platforms
SET name = $2,
    display_name = $3,
    slug = $4,
    enabled = $5,
    oauth_capable = $6,
    webhook_capable = $7,
    updated_at = NOW()
WHERE id = $1
RETURNING id, name, display_name, slug, enabled, oauth_capable, webhook_capable, created_at, updated_at;

-- name: DeletePlatform :execrows
DELETE FROM platforms WHERE id = $1;