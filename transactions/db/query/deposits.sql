-- name: CreateDeposit :one
INSERT INTO deposits (
    client_id,
    customer_id,
    merchant_id,
    amount,
    currency,
    payment_type,
    payer_phone_number,
    provider,
    status,
    idempotency_key
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetDepositByID :one
SELECT * FROM deposits WHERE id = $1;

-- name: GetDepositByExternalReference :one
SELECT * FROM deposits WHERE external_reference = $1;

-- name: GetDepositByIdempotencyKey :one
SELECT * FROM deposits WHERE idempotency_key = $1;

-- name: ListDepositsByClient :many
SELECT * FROM deposits
WHERE client_id = $1
ORDER BY created_at DESC;

-- name: ListDepositsByCustomer :many
SELECT * FROM deposits
WHERE customer_id = $1
ORDER BY created_at DESC;

-- name: ListDepositsByMerchant :many
SELECT * FROM deposits
WHERE merchant_id = $1
ORDER BY created_at DESC;

-- name: ListDepositsByStatus :many
SELECT * FROM deposits
WHERE status = $1
ORDER BY created_at DESC;

-- name: UpdateDepositStatus :one
UPDATE deposits
SET status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateDepositStatusAndCompletedAt :one
UPDATE deposits
SET status = $2,
    completed_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateDepositStatusAndFailedAt :one
UPDATE deposits
SET status = $2,
    failed_at = NOW(),
    failure_reason = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;