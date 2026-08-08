# Clients service

The Clients service manages client onboarding, platform integrations, OAuth
flows, and webhook processing for the RVPay platform. It exposes gRPC
operations for managing clients, platforms, integrations, and handles
provider-specific OAuth and webhook communication.

## Runtime flow

```text
gRPC client
    │ CreateClient / ListClients / ActivateClient
    ▼
ClientsService
    ├── validates request
    ├── persists client record
    └── returns client
    ▼
CreateClientResponse (CLIENT)

gRPC client
    │ InstallIntegration / OAuth callback
    ▼
IntegrationsService / OAuthService
    ├── validates client and platform
    ├── initiates OAuth flow with provider
    ├── exchanges authorization code for tokens
    └── creates integration record
    ▼
InstallIntegrationResponse (INTEGRATION)

Webhook endpoint
    │ Provider webhook payload
    ▼
WebhookService
    ├── validates signature
    ├── parses event
    ├── detects duplicates
    └── dispatches to business services
    ▼
200 OK / 4xx / 5xx
```

The process entry point is `cmd/grpc-service/main.go`. It loads `.env` from
its working directory when present, reads environment variables, connects and
pings PostgreSQL, applies migrations, registers providers, initializes business
services, registers gRPC services with reflection and a unary panic-recovery
interceptor, starts a grpc-gateway REST endpoint, then listens on `:$LISTEN_PORT`.
It stops gracefully on `SIGINT` or `SIGTERM`.

## Directory guide

```text
clients/
├── cmd/grpc-service/main.go      # gRPC server bootstrap and shutdown
├── config/model.go               # Config and DBConfig environment bindings
├── db/
│   ├── migrations/               # 000001 creates clients, platforms, integrations, oauth_tokens, webhook_subscriptions
│   ├── query/                    # SQL queries for all entities
│   ├── repo/                     # pgx pool adapter plus migration helpers
│   ├── sqlc/                     # sqlc-generated data access code
│   └── doc.go                    # go:generate directives
├── service/
│   ├── clients_service.go        # Client CRUD operations
│   ├── platforms_service.go      # Platform management
│   └── integrations_service.go   # Integration lifecycle
├── oauth/
│   └── service.go                # OAuth flow orchestration
├── webhooks/
│   └── service.go                # Webhook lifecycle orchestration
├── providers/
│   ├── provider.go               # Provider interfaces and registry
│   ├── highlevel.go              # HighLevel OAuth provider
│   └── highlevel_webhook.go      # HighLevel webhook provider
├── docs/                         # Service documentation
├── .env                          # local runtime configuration; do not commit
└── Makefile                      # local development tasks
```

## Configuration

Copy the repository template and edit it:

```bash
cp .env.example .env
```

| Variable | Required | Purpose |
| --- | --- | --- |
| `LOG_LEVEL` | No; defaults to `info` | Zerolog level |
| `LISTEN_PORT` | Yes | gRPC TCP port |
| `PORT` | No; defaults to `8080` | HTTP gateway port |
| `DB_USER` | Yes | PostgreSQL user |
| `DB_PASSWORD` | Yes | PostgreSQL password |
| `DB_HOST` | Yes | PostgreSQL host |
| `DB_PORT` | Yes | PostgreSQL port |
| `DB_NAME` | Yes | PostgreSQL database |
| `DB_TLS_DISABLED` | No; defaults to `true` | Selects `sslmode=disable`; set `false` for `require` |
| `RUN_MIGRATIONS` | No; defaults to `true` | Apply migrations on startup |
| `MIGRATION_PATH` | No; defaults to `db/migrations` | Migration directory |
| `HIGHLEVEL_CLIENT_ID` | Yes | HighLevel OAuth client ID |
| `HIGHLEVEL_CLIENT_SECRET` | Yes | HighLevel OAuth client secret |
| `HIGHLEVEL_REDIRECT_URI` | No; defaults to `https://api.rvpay.com/v1/public/oauth/callback` | OAuth callback URL |
| `WEBHOOK_SECRET` | Yes | HighLevel webhook verification secret |

## Local startup

Run all commands below from this directory.

```bash
make rundb
docker exec -it clients-postgres psql -U postgres -d clients -c 'CREATE EXTENSION IF NOT EXISTS pgcrypto;'
make run
```

`make rundb` starts a detached PostgreSQL 16 Alpine container named
`clients-postgres`, exposes `DB_PORT`, and uses `DB_USER`, `DB_PASSWORD`, and
`DB_NAME` from `.env`. The service applies migrations automatically at startup.
The initial schema relies on `gen_random_uuid()`, so `pgcrypto` must be enabled
in the database.

## Service API

The generated gRPC service names are:

- `clientsgrpc.ClientsService` — Client CRUD operations
- `clientsgrpc.PlatformsService` — Platform management
- `clientsgrpc.IntegrationsService` — Integration lifecycle

The full protobuf schema is [../protobuf/clients.proto](../protobuf/clients.proto).

Example with local reflection enabled:

```bash
grpcurl -plaintext \
  -d '{"name":"Acme Corp"}' \
  localhost:50051 clientsgrpc.ClientsService/CreateClient
```

## Database and generated code

- `db/migrations` is the database source of truth.
- `db/query/*.sql` contains the SQL queries consumed by sqlc.
- `db/sqlc` is generated code; update migrations/queries and run `make generate`
  rather than editing it manually.
- `db/repo` exposes repository interfaces and implementations.

## Make targets

```bash
make install-tools
make generate
make generate-protos
make generate-sql
make lint
make test
make run
make rundb
make create-migration name=descriptive_migration_name
```

## Generation workflow

```bash
# Install code generation tools
make install-tools

# Generate all code (protos, sqlc, mocks)
make generate

# Generate protobuf code only
make generate-protos

# Generate sqlc code only
make generate-sql
```

## Docker

```bash
# Build Docker image
make docker-build

# Run with Docker
docker run -p 50051:50051 -p 8080:8080 --env-file .env rvpay-go-clients:local
```

## Deployment

The Clients service is compatible with:

- **Render** — See `deploy/render/` for service configuration
- **Docker** — Multi-stage build with distroless runtime
- **Kubernetes** — Standard Go binary deployment

## Current behavior and limitations

- OAuth flows are implemented for HighLevel provider only.
- Webhook processing is implemented for HighLevel provider only.
- Token refresh is manual; no automatic scheduling is implemented.
- Webhook deduplication uses subscription lookup; full event deduplication
  requires a `webhook_events` table.
- No authentication or authorization is implemented at the transport layer.