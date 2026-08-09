-- name: CreateCustomer :one
INSERT INTO customers (client_id, merchant_id, phone_number, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCustomerByID :one
SELECT * FROM customers WHERE id = $1;

-- name: GetCustomerByClientAndMerchantAndPhone :one
SELECT * FROM customers
WHERE client_id = $1 AND merchant_id = $2 AND phone_number = $3;

-- name: ListCustomersByClient :many
SELECT * FROM customers
WHERE client_id = $1
ORDER BY created_at;

-- name: ListCustomersByMerchant :many
SELECT * FROM customers
WHERE merchant_id = $1
ORDER BY created_at;

-- name: UpdateCustomerStatus :one
UPDATE customers
SET status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;