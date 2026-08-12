// Package db contains the database related code.
package db

//go:generate go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate
//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination sqlc/mocks/querier.go -package mocks ./sqlc Querier
//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination repo/mocks/repo.go -package mocks ./repo ClientsRepo
//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination repo/mocks/client_repo.go -package mocks ./repo ClientRepo
//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination repo/mocks/platform_repo.go -package mocks ./repo PlatformRepo
//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination repo/mocks/integration_repo.go -package mocks ./repo IntegrationRepo
//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination repo/mocks/oauth_token_repo.go -package mocks ./repo OAuthTokenRepo
//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination repo/mocks/webhook_subscription_repo.go -package mocks ./repo WebhookSubscriptionRepo
//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination repo/mocks/oauth_state_repo.go -package mocks ./repo OAuthStateRepo
//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination repo/mocks/webhook_event_repo.go -package mocks ./repo WebhookEventRepo
