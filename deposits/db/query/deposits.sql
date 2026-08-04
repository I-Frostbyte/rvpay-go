-- name: CreateDeposit :one
INSERT INTO deposits (amount, currency, payer_type, payer_phone_number, payer_provider, client_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetDepositByID :one
SELECT * FROM deposits WHERE id = $1;