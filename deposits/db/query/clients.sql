-- name: CreateClient :one
INSERT INTO clients (client_name, email, phone_number)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetClientByID :one
SELECT * FROM clients WHERE id = $1;