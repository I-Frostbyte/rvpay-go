-- name: CreateMerchant :one
INSERT INTO merchants (name, slug, status)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetMerchantByID :one
SELECT * FROM merchants WHERE id = $1;

-- name: GetMerchantBySlug :one
SELECT * FROM merchants WHERE slug = $1;

-- name: ListMerchants :many
SELECT * FROM merchants
ORDER BY created_at
LIMIT $1 OFFSET $2;

-- name: CountMerchants :one
SELECT COUNT(*) FROM merchants;

-- name: UpdateMerchantStatus :one
UPDATE merchants
SET status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;