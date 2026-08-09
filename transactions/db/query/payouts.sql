-- name: CreatePayout :one
INSERT INTO payouts (
    client_id,
    merchant_id,
    amount,
    currency,
    provider,
    destination_reference,
    status,
    idempotency_key
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetPayoutByID :one
SELECT * FROM payouts WHERE id = $1;

-- name: GetPayoutByExternalReference :one
SELECT * FROM payouts WHERE external_reference = $1;

-- name: GetPayoutByIdempotencyKey :one
SELECT * FROM payouts WHERE idempotency_key = $1;

-- name: ListPayoutsByClient :many
SELECT * FROM payouts
WHERE client_id = $1
ORDER BY created_at DESC;

-- name: ListPayoutsByMerchant :many
SELECT * FROM payouts
WHERE merchant_id = $1
ORDER BY created_at DESC;

-- name: ListPayoutsByStatus :many
SELECT * FROM payouts
WHERE status = $1
ORDER BY created_at DESC;

-- name: UpdatePayoutStatus :one
UPDATE payouts
SET status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdatePayoutStatusAndCompletedAt :one
UPDATE payouts
SET status = $2,
    completed_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdatePayoutStatusAndFailedAt :one
UPDATE payouts
SET status = $2,
    failed_at = NOW(),
    failure_reason = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;