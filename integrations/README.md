# Integrations service

The Integrations service manages third-party provider connections for
`rvpay-go`. It stores provider OAuth credentials, processes incoming webhook
events, and exposes a gRPC API (with a REST gateway) for managing integrations.

## Purpose

The Integrations service is the single place where external provider
connections are authorized, stored, and refreshed. It keeps provider tokens
encrypted at rest and records every webhook event so downstream services can
react to provider activity without coupling to a specific vendor.

## Responsibilities

- **OAuth management** — handles the HighLevel OAuth callback flow, exchanges
  authorization codes for tokens, and stores them encrypted.
- **Webhook processing** — receives HighLevel webhook events, persists them,
  and processes them asynchronously.
- **Token lifecycle** — stores access and refresh tokens with an expiry time
  and encrypts them with a service-level AES key.
- **Integration storage** — persists provider integrations in PostgreSQL and
  exposes create, get, and delete operations over gRPC/REST.
- **Provider abstraction** — the service is structured so additional providers
  can be added without changing the core integration and webhook flows.

## Runtime flow

```text
gRPC/REST client
    │ CreateIntegration / GetIntegration / DeleteIntegration
    ▼
IntegrationService
    ├── creates / reads / deletes integration records in PostgreSQL
    └── returns the Integration resource
```

```text
HighLevel
    │ OAuth callback (code)
    ▼
/oauth/callback
    └── exchanges code, encrypts tokens, stores integration
```

```text
HighLevel
    │ webhook POST
    ▼
/webhooks/highlevel
    └── stores webhook event, processes asynchronously
```

The process entry point is `cmd/grpc-service/main.go`. It loads `.env` from
its working directory when present, reads environment variables, connects and
pings PostgreSQL, applies migrations, registers the gRPC service with
reflection, health, and a unary panic-recovery interceptor, then listens on
`:$LISTEN_PORT`. It also starts an HTTP gateway on `:$PORT` (default `8080`)
that serves the REST endpoints, `/healthz`, `/oauth/callback`, and
`/webhooks/highlevel`. It stops gracefully on `SIGINT` or `SIGTERM`.

## Folder Layout

```text
integrations/
├── cmd/grpc-service/main.go      # gRPC + HTTP gateway server bootstrap and shutdown
├── config/model.go               # Config and DBConfig environment bindings
├── db/
│   ├── migrations/               # 000001 creates integrations and webhook_events
│   ├── query/                    # Create/Get/Delete integration and webhook SQL
│   ├── repo/                     # pgx pool adapter plus migration helpers
│   ├── sqlc/                     # sqlc-generated data access code
│   └── doc.go                    # go:generate directives
├── integrations/service.go       # IntegrationService gRPC implementation
├── oauth/                        # HighLevel OAuth callback handler and service
├── webhook/                      # HighLevel webhook handler and service
├── .env                          # local runtime configuration; do not commit
└── Makefile                      # local development tasks
```

## Local Development

### Prerequisites

- Go 1.26.5
- Docker (for `make rundb` and `make build`)
- `grpcurl` (optional, for testing the gRPC API)

### Environment variables

Copy the repository template and edit it:

```bash
cp .env.example .env
```

| Variable | Required | Purpose |
| --- | --- | --- |
| `HIGHLEVEL_CLIENT_ID` | No config-level validation | HighLevel OAuth client ID |
| `HIGHLEVEL_CLIENT_SECRET` | No config-level validation | HighLevel OAuth client secret |
| `HIGHLEVEL_REDIRECT_URL` | No config-level validation | HighLevel OAuth redirect URL |
| `HIGHLEVEL_SSO_KEY` | No config-level validation | HighLevel webhook verification key |
| `TOKEN_ENCRYPTION_KEY` | Yes | AES key used to encrypt provider tokens |
| `LOG_LEVEL` | No; defaults to `debug` | Zerolog level |
| `LISTEN_PORT` | Yes | gRPC TCP port |
| `PORT` | No; defaults to `8080` | HTTP gateway and webhook port |
| `MIGRATION_PATH` | Yes | Migration directory, typically `db/migrations` |
| `RUN_MIGRATIONS` | No; defaults to `true` | Apply migrations at startup |
| `DB_USER` | Yes | PostgreSQL user |
| `DB_PASSWORD` | Yes | PostgreSQL password |
| `DB_HOST` | Yes | PostgreSQL host |
| `DB_PORT` | Yes | PostgreSQL port |
| `DB_NAME` | Yes | PostgreSQL database |
| `DB_TLS_DISABLED` | No; defaults to `false` | Selects `sslmode=disable`; set `true` for local |
| `env` | No | Defined but not currently consumed |

### Makefile commands

```bash
make install-tools
make generate
make lint
make test
make run
make rundb
make create-migration name=descriptive_migration_name
```

### Running locally

Run all commands below from this directory.

```bash
make rundb
docker exec -it integrations-postgres psql -U postgres -d integrations -c 'CREATE EXTENSION IF NOT EXISTS pgcrypto;'
make run
```

`make rundb` starts a detached PostgreSQL 16 Alpine container named
`integrations-postgres`, exposes `DB_PORT`, and uses `DB_USER`, `DB_PASSWORD`,
and `DB_NAME` from `.env`. The service applies migrations automatically at
startup. The initial schema relies on `gen_random_uuid()`, so `pgcrypto` must
be enabled in the database.

## Service API

The generated gRPC service name is `integrationsgrpc.IntegrationService`, with
methods:

- `CreateIntegration(CreateIntegrationRequest) -> CreateIntegrationResponse`
- `GetIntegration(GetIntegrationRequest) -> GetIntegrationResponse`
- `DeleteIntegration(DeleteIntegrationRequest) -> DeleteIntegrationResponse`
- `ProcessWebhookEvent(ProcessWebhookEventRequest) -> ProcessWebhookEventResponse`

The REST gateway exposes the same operations at:

- `POST /v1/public/integrations`
- `GET /v1/public/integrations/{id}`
- `DELETE /v1/public/integrations/{id}`
- `POST /v1/public/webhooks`

The full protobuf schema is
[../protobuf/integrations.proto](../protobuf/integrations.proto).

Example with local reflection enabled:

```bash
grpcurl -plaintext \
  -d '{"provider":"highlevel","locationId":"loc_123","accessToken":"...","refreshToken":"...","tokenExpiresAt":"2026-08-08T00:00:00Z"}' \
  localhost:50051 integrationsgrpc.IntegrationService/CreateIntegration
```

## Database

- `db/migrations` is the database source of truth.
- `db/query/*.sql` contains the SQL queries consumed by sqlc.
- `db/sqlc` is generated code; update migrations/queries and run `make generate`
  rather than editing it manually.
- `db/repo` exposes `Do()` for ordinary queries and `Begin()` for callers that
  need a transaction.

### Migrations

Create a new migration with:

```bash
make create-migration name=descriptive_migration_name
```

### sqlc generation

```bash
make generate
```

This runs the `go:generate` directives in `db/doc.go`, which regenerate the
sqlc data access code and the mocks.

## Protobuf

The protobuf source of truth is `../protobuf/integrations.proto`. Generated
gRPC and gateway code lives in `../grpc/go/integrationsgrpc` and must not be
edited by hand.

Regenerate from the `protobuf/` directory:

```bash
make generate-protos
```

## Docker

Build the container image from the repository root:

```bash
docker build -f integrations/Dockerfile -t rvpay-go-integrations:local .
```

The image uses a multi-stage build: a `golang:1.26.5-alpine` build stage
compiles the service with `CGO_ENABLED=0`, and a
`gcr.io/distroless/static-debian12:nonroot` runtime stage runs it as a
non-root user. Migrations are copied into the image at
`/app/integrations/db/migrations`.

Run the container:

```bash
docker run --rm -p 50051:50051 -p 8080:8080 \
  --env-file .env \
  rvpay-go-integrations:local
```

## Future Integrations

The service is structured so additional providers can be added without
changing the core integration and webhook flows. Reserved sections:

- **HubSpot** — OAuth flow and webhook handling.
- **Salesforce** — OAuth flow and webhook handling.
- **Stripe** — API-key based integration and webhook handling.
- **Microsoft** — OAuth flow and webhook handling.
- **Google** — OAuth flow and webhook handling.

## Current behavior and limitations

- Only the HighLevel provider is implemented for OAuth and webhooks.
- Webhook events are stored and processed asynchronously; internal service
  communication (e.g. notifying the deposits service) is not yet implemented.
- `HIGHLEVEL_SSO_KEY` is defined in configuration but webhook signature
  verification is not yet enforced.
- The `ProcessWebhookEvent` gRPC method is defined in the protobuf contract but
  is not yet implemented in the service.