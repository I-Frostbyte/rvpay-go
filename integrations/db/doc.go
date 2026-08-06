// Package db contains the database related code.
package db

//go:generate go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate
//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination sqlc/mocks/querier.go -package mocks ./sqlc Querier
//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination repo/mocks/repo.go -package mocks ./repo IntegrationsRepo
