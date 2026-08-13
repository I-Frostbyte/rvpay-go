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
| `HIGHLEVEL_REDIRECT_URI` | Yes | Publicly reachable OAuth callback URL, e.g. `https://<render-client-host>/oauth/callback` |
| `HIGHLEVEL_WEBHOOK_PUBLIC_KEY` | Yes | PEM-encoded Ed25519 public key used to verify `X-GHL-Signature` webhook signatures. This is PUBLIC cryptographic material, not a private secret. |
| `HIGHLEVEL_PAYMENT_URL` | No | Frontend checkout URL supplied to HighLevel as the payment provider's `paymentsUrl`. Configuration, never hard-coded. |
| `HIGHLEVEL_QUERY_URL` | No | Backend query URL supplied to HighLevel as the payment provider's `queryUrl`. Configuration, never hard-coded. |
| `HIGHLEVEL_PROVIDER_NAME` | No; defaults to `RVPay` | Display name of the payment provider. |
| `HIGHLEVEL_PROVIDER_DESCRIPTION` | No; defaults to `RVPay payment provider` | Description of the payment provider. |
| `HIGHLEVEL_PROVIDER_IMAGE_URL` | No | Image URL of the payment provider. |
| `TRANSACTIONS_GRPC_ADDR` | No; defaults to `localhost:50052` | gRPC address of the Transactions service, used by the GHL Custom Payment Provider query/webhook endpoints to correlate HighLevel transactions with RVPay deposits. |

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

## GoHighLevel integration

The Clients service is the owner of the GoHighLevel (GHL) Marketplace
integration. It exposes two direct HTTP endpoints (not grpc-gateway RPCs)
because they are external provider/browser-facing:

| Route | Method | Purpose |
| --- | --- | --- |
| `/oauth/callback` | GET | GHL OAuth authorization callback (`code` + `state` query params) |
| `/webhooks/highlevel` | POST | GHL webhook deliveries (`X-GHL-Signature` header) |

### OAuth flow

1. `BeginAuthorization(clientID, platformID)` generates a cryptographically
   random state, persists it (with the client/platform context and a 10-minute
   expiry) in the `oauth_states` table, and returns the GHL authorization URL.
2. GHL redirects the user to `HIGHLEVEL_REDIRECT_URI` (`/oauth/callback`) with
   `code` and `state`.
3. `HandleCallback(code, state)` atomically consumes the state (rejecting
   missing, expired, or already-consumed states to prevent CSRF/replay), recovers
   the client/platform context, exchanges the code for tokens, and creates the
   integration.

### Webhook flow

1. GHL POSTs to `/webhooks/highlevel` with the raw JSON body and an
   `X-GHL-Signature` header (base64-encoded Ed25519 signature over the raw body).
2. The handler reads the raw body bytes and passes them (unmodified) to the
   webhook service.
3. The HighLevel provider verifies the signature against the raw body using the
   configured `HIGHLEVEL_WEBHOOK_PUBLIC_KEY`. Missing, malformed, and invalid
   signatures are rejected with HTTP 400.
4. The event is parsed and recorded in the `webhook_events` table. The unique
   constraint on `(integration_id, provider_event_id)` plus `ON CONFLICT DO
   NOTHING` makes duplicate deliveries race-safe and idempotent; duplicates are
   acknowledged with HTTP 200 so the provider stops retrying.
5. The event is dispatched to the HighLevel dispatcher.

### GHL Marketplace configuration

Configure the GHL Marketplace app with:

- **Client ID** → `HIGHLEVEL_CLIENT_ID`
- **Client Secret** → `HIGHLEVEL_CLIENT_SECRET`
- **Redirect URL** → `https://<render-client-host>/oauth/callback`
- **Webhook URL** → `https://<render-client-host>/webhooks/highlevel`
- **Webhook Verification** → `X-GHL-Signature` / Ed25519
- **Public Key** → `HIGHLEVEL_WEBHOOK_PUBLIC_KEY`

The Render hostname is supplied through deployment configuration
(`HIGHLEVEL_REDIRECT_URI`); it is never hard-coded.

### GHL Custom Payment Provider

The Clients service also implements the backend half of the GHL Custom
Payment Provider contract. It exposes two additional direct HTTP endpoints
(distinct from the Marketplace OAuth callback and webhook):

| Route | Method | Purpose |
| --- | --- | --- |
| `/payments/custom-provider/query` | POST | GHL payment query operations (`type=verify`) |
| `/payments/custom-provider/webhook` | POST | GHL payment-provider webhook events (`payment.captured`) |

#### Payment query flow

1. GHL POSTs to `/payments/custom-provider/query` with a JSON body containing
   `type`, `transactionId`, `apiKey`, `chargeId`, and `subscriptionId`.
2. The handler validates the provider API key against the stored
   `payment_provider_configs` record (constant-time comparison; the key is
   never logged).
3. For `type=verify`, the handler correlates the HighLevel transaction with an
   RVPay deposit by calling the Transactions service
   (`GetDepositByGHLTransactionID` gRPC RPC).
4. The response follows the HighLevel contract: `{"success":true}` for
   completed, `{"failed":true}` for failed, and `{"success":false}` for
   pending. Internal RVPay transaction objects are never returned.

#### Payment webhook flow

1. GHL POSTs to `/payments/custom-provider/webhook` with a payment event
   payload (e.g. `payment.captured`).
2. The handler resolves the integration by `locationId` via the
   `payment_provider_configs` table.
3. The event is recorded in the `webhook_events` table. The unique constraint
   on `(integration_id, provider_event_id)` plus `ON CONFLICT DO NOTHING`
   makes duplicate deliveries race-safe and idempotent; duplicates are
   acknowledged with HTTP 200 so the provider stops retrying.
4. Only `payment.captured` is processed (one-time payment flow). Unknown event
   types (e.g. subscription events) are acknowledged safely without
   processing.

#### Provider configuration

The `payment_provider_configs` table stores per-integration GHL payment
provider configuration: `provider_name`, `provider_description`,
`provider_image_url`, `location_id`, `query_url`, `payments_url`,
`supports_subscription_schedule` (always `false` for one-time payments), and
`provider_api_key`. The provider API key is distinct from the OAuth client
secret and the pawaPay API key.

## Current behavior and limitations

- OAuth flows are implemented for HighLevel provider only.
- Webhook processing is implemented for HighLevel provider only.
- Token refresh is manual; no automatic scheduling is implemented.
- Webhook deduplication is enforced via the `webhook_events` table.
- No authentication or authorization is implemented at the transport layer.
