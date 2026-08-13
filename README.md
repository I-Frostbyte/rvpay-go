# rvpay-go

`rvpay-go` is a Go microservices platform for payment processing (deposits and
payouts) with marketplace platform integration (GoHighLevel), PostgreSQL
persistence, gRPC + gRPC-Gateway (REST), and deployment targets of Render and
Oracle Cloud Infrastructure (OCI).

The repository currently contains four runnable gRPC services:

- **Deposits** (legacy) — accepts an `InitiateDeposit` request, stores a client
  and deposit in PostgreSQL, and calls the PawaPay client to initiate the
  external mobile-money deposit.
- **Integrations** (legacy) — manages third-party provider connections. It
  handles the HighLevel OAuth callback flow, stores encrypted provider tokens,
  and receives and persists HighLevel webhook events.
- **Clients** — client/platform/integration/oauth/webhook domain (Clients,
  Platforms, Integrations, OAuth, Webhooks via a unified Provider interface).
- **Transactions** — merchant/customer/deposit/payout domain (Merchants,
  Customers, Deposits, Payouts).

Every service exposes a gRPC API and an embedded HTTP gateway (gRPC-Gateway)
with a `/healthz` endpoint. Per `docs/migration-plan.md`, the legacy services
evolve into the new services (Integrations → Clients, Deposits → Transactions);
the legacy services remain runnable and are not deleted. This README describes
the repository **as it is today**.

## Requirements

- Go 1.26.5 (the version declared in `go.mod`)
- Docker, to start the development PostgreSQL container or build images
- PostgreSQL 16-compatible database if you do not use Docker
- PawaPay API URL and API key (Deposits service)
- HighLevel OAuth credentials and a webhook verification key (Clients and
  Integrations services)
- `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`, and
  `protoc-gen-grpc-gateway` only when regenerating protobuf code (versions
  pinned in `tools/versions.md`)

## Repository layout

```text
.
├── deposits/                    # Deposits gRPC service (legacy, runnable)
│   ├── cmd/grpc-service/         # Process entry point and server lifecycle
│   ├── config/                   # Environment-backed service configuration
│   ├── db/
│   │   ├── migrations/           # PostgreSQL up/down migrations
│   │   ├── query/                # SQL inputs for sqlc
│   │   ├── repo/                 # Pool/query wrapper and migration runner
│   │   └── sqlc/                 # Generated query models and methods
│   ├── deposits/                 # DepositsService implementation
│   ├── Dockerfile                # Multi-stage distroless container build
│   ├── Makefile                  # Service tasks
│   └── README.md                 # Service documentation
├── integrations/                # Integrations gRPC service (legacy, runnable)
│   ├── cmd/grpc-service/         # Process entry point and server lifecycle
│   ├── config/                   # Environment-backed service configuration
│   ├── db/
│   │   ├── migrations/           # PostgreSQL up/down migrations
│   │   ├── query/                # SQL inputs for sqlc
│   │   ├── repo/                 # Pool/query wrapper and migration runner
│   │   └── sqlc/                 # Generated query models and methods
│   ├── integrations/             # IntegrationService implementation
│   ├── oauth/                    # HighLevel OAuth callback handler and service
│   ├── webhook/                  # HighLevel webhook handler and service
│   ├── Dockerfile                # Multi-stage distroless container build
│   ├── Makefile                  # Service tasks
│   └── README.md                 # Service documentation
├── clients/                     # Clients gRPC service
│   ├── cmd/grpc-service/         # Process entry point and server lifecycle
│   ├── config/                   # Environment-backed service configuration
│   ├── db/
│   │   ├── migrations/           # PostgreSQL up/down migrations
│   │   ├── query/                # SQL inputs for sqlc
│   │   ├── repo/                 # Pool/query wrapper and migration runner
│   │   └── sqlc/                 # Generated query models and methods
│   ├── clients/                  # Client domain logic
│   ├── platforms/                # Platform domain logic
│   ├── integrations/             # Integration domain logic
│   ├── oauth/                    # HighLevel OAuth callback handler and service
│   ├── webhooks/                 # HighLevel webhook handler and service
│   ├── providers/                # Unified Provider interface + HighLevel implementation
│   ├── service/                  # gRPC service implementations + converters
│   ├── Dockerfile                # Multi-stage distroless container build
│   ├── Makefile                  # Service tasks
│   └── README.md                 # Service documentation
├── transactions/                # Transactions gRPC service
│   ├── cmd/grpc-service/         # Process entry point and server lifecycle
│   ├── config/                   # Environment-backed service configuration
│   ├── db/
│   │   ├── migrations/           # PostgreSQL up/down migrations
│   │   ├── query/                # SQL inputs for sqlc
│   │   ├── repo/                 # Pool/query wrapper and migration runner
│   │   └── sqlc/                 # Generated query models and methods
│   ├── merchants/                # Merchant domain logic
│   ├── customers/                # Customer domain logic
│   ├── deposits/                 # Deposit domain logic
│   ├── payouts/                  # Payout domain logic
│   ├── Dockerfile                # Multi-stage distroless container build
│   ├── Makefile                  # Service tasks
│   └── README.md                 # Service documentation
├── shared/                      # Shared infrastructure packages (non-business)
│   ├── logger/                   # Shared zerolog setup
│   └── database/                 # Shared PostgreSQL DSN/pool/migration helpers
├── protobuf/                     # Source protobuf contracts and generation task
│   ├── clients.proto             # Clients Service contract (clientsgrpc)
│   ├── transactions.proto        # Transactions Service contract (transactionsgrpc)
│   ├── common.proto              # Shared types (commongrpc)
│   ├── deposits.proto            # Legacy Deposits contract (depositsgrpc)
│   ├── integrations.proto        # Legacy Integrations contract (integrationsgrpc)
│   ├── Makefile                  # protobuf lint and Go code generation targets
│   └── README.md                 # API contract and generation workflow
├── grpc/go/                      # Generated Go protobuf/gRPC/gateway stubs
│   ├── clientsgrpc/             # Generated Clients service code
│   ├── transactionsgrpc/        # Generated Transactions service code
│   ├── commongrpc/              # Generated shared types
│   ├── depositsgrpc/            # Generated legacy Deposits service code
│   └── integrationsgrpc/        # Generated legacy Integrations service code
├── third_party/googleapis/       # googleapis Git submodule used by protoc
├── nginx/                        # Nginx TLS termination config for OCI
├── deploy/                       # OCI and Render deployment documentation
├── .github/workflows/            # CI/CD pipelines (OCI disabled; Render active)
├── .env.example                  # Environment-variable template
├── docker-compose.yml            # OCI Always Free Compose stack
├── render.yaml                   # Render Blueprint (3 services + 3 databases)
├── Makefile                      # Repository-wide test tasks
└── layout.md                     # Original layout notes
```

Generated code (`grpc/go/*`, `<service>/db/sqlc/*`, mocks) is committed and
must be regenerated, never edited by hand.

## Services

The service boundaries follow `docs/domain-model.md`. Each service owns its own
configuration, database layer (migrations, sqlc, repo), business logic, gRPC
handlers, Dockerfile, and Makefile. Services never share database tables;
cross-service communication is gRPC-only.

| Service | Responsibility | Location | Interfaces | Database |
| --- | --- | --- | --- | --- |
| Deposits (legacy) | Initiate external mobile-money deposits via PawaPay | `deposits/` | gRPC `depositsgrpc.DepositsService` + HTTP `/v1/public/deposits` | `deposits` |
| Integrations (legacy) | Manage third-party provider connections, OAuth, webhooks | `integrations/` | gRPC `integrationsgrpc.IntegrationService` + HTTP `/v1/public/integrations`, `/v1/public/webhooks`, `/oauth/callback`, `/webhooks/highlevel` | `integrations` |
| Clients | Clients, Platforms, and Integrations; HighLevel OAuth and webhooks | `clients/` | gRPC `clientsgrpc.{ClientsService, PlatformsService, IntegrationsService}` + HTTP `/v1/public/clients`, `/v1/public/platforms`, `/v1/public/integrations` | Owns its own PostgreSQL database |
| Transactions | Merchants, Customers, Deposits, and Payouts | `transactions/` | gRPC `transactionsgrpc.{MerchantService, CustomerService, DepositService, PayoutService}` + HTTP `/v1/public/merchants`, `/v1/public/customers`, `/v1/public/deposits`, `/v1/public/payouts` | Owns its own PostgreSQL database |

### Clients

- Purpose: identity, marketplace platforms, and integrations (Clients,
  Platforms, Integrations, OAuth tokens, webhook subscriptions).
- Repository location: `clients/`.
- gRPC/API role: `clientsgrpc.ClientsService`, `clientsgrpc.PlatformsService`,
  `clientsgrpc.IntegrationsService`, exposed over HTTP through
  gRPC-Gateway at `/v1/public/...`.
- OAuth: HighLevel OAuth callback flow (tokens stored encrypted by the service).
- Webhooks: HighLevel webhook event ingestion and persistence.
- Database: PostgreSQL, owned by the service (`clients/db`).
- The `ClientService.GetClient` RPC is the documented cross-service validation
  path for Transactions (not yet wired; see the migration plan).

### Transactions

- Purpose: payment operations — merchants, customers, deposits, and payouts.
- Repository location: `transactions/`.
- gRPC/API role: `transactionsgrpc.MerchantService`,
  `transactionsgrpc.CustomerService`, `transactionsgrpc.DepositService`,
  `transactionsgrpc.PayoutService`, exposed over HTTP through gRPC-Gateway at
  `/v1/public/...`.
- Database: PostgreSQL, owned by the service (`transactions/db`).
- Deposits/payouts initialize in INITIATED/REQUESTED; provider execution and
  status reconciliation are future integration work (see the production
  review).

## Shared infrastructure

- `shared/logger` — zerolog logger setup (timestamps, caller, validated level).
- `shared/database` — Postgres DSN builder, pgxpool connect with eager ping,
  golang-migrate runner.

These are consumed by the Clients and Transactions service bootstraps.
Business logic stays in the owning service; see
`docs/platform-common-packages-review.md`.

## Local development

### 1. Clone and initialise submodules

```bash
git clone git@github.com:I-Frostbyte/rvpay-go.git
cd rvpay-go
git submodule update --init --recursive   # googleapis, needed only for protobuf regeneration
```

### 2. Create service environment files

Each service's `LoadConfig` looks for `.env` in its own directory. Start from
the templates:

```bash
cp .env.example deposits/.env
cp .env.example integrations/.env
cp clients/.env.example clients/.env
cp transactions/.env.example transactions/.env
```

Edit each `.env` with the values documented in the service README and the
configuration section below. Keep credentials out of source control.

### 3. Start PostgreSQL

From the service directory (each service uses its own database and port):

```bash
cd deposits && make rundb
cd integrations && make rundb
cd clients && make rundb
cd transactions && make rundb
```

The initial migrations use `gen_random_uuid()`. Ensure the target database has
the `pgcrypto` extension:

```bash
docker exec -it <service>-postgres psql -U postgres -d <db> -c 'CREATE EXTENSION IF NOT EXISTS pgcrypto;'
```

### 4. Run a service

From the service directory:

```bash
make run
```

On startup the service loads configuration, connects and pings PostgreSQL,
runs up migrations (`RUN_MIGRATIONS`, default true), registers gRPC reflection
and recovery middleware, then listens on `:$LISTEN_PORT` with an HTTP gateway
on `:$PORT` (default `8080`) serving REST endpoints and `/healthz`. `SIGINT`
and `SIGTERM` trigger graceful shutdown.

### 5. Call the API

gRPC reflection allows `grpcurl` plaintext requests without proto files. For
example:

```bash
grpcurl -plaintext -d '{"amount":"1000.00","currency":"XAF","payer":{"type":"DEPOSIT_PORTAL_MMO","accountDetails":{"phoneNumber":"+237699541235","provider":"DEPOSIT_PROVIDER_MTN_MOMO_CMR"}},"clientId":"not-currently-used"}' \
  localhost:50051 depositsgrpc.DepositsService/InitiateDeposit
```

REST endpoints mirror the gRPC API under `/v1/public/...` (see the protobuf
contracts for the exact routes).

## Protobuf / gRPC workflow

The protobuf contracts are the source of truth (`protobuf/*.proto`); generated
Go code is committed under `grpc/go/` and is output only.

From `protobuf/`:

```bash
make lint              # clang-format dry-run over sources
make generate-protos   # generate Go, gRPC, and gateway stubs into ../grpc/go
```

Tool versions are pinned in `tools/versions.md` (protoc v3.21.12,
protoc-gen-go v1.36.10, protoc-gen-go-grpc v1.5.1,
protoc-gen-grpc-gateway v2.22.0). See `docs/protobuf-strategy.md` and
`docs/platform-protobuf-generation-review.md`.

Generated files are never edited by hand. After changing a `.proto`, regenerate
and commit the output with the contract change.

## SQLC workflow

Each service defines its sqlc configuration and SQL inputs:

- `clients/db/sqlc.yaml` + `clients/db/query/` → `clients/db/sqlc/`
- `transactions/db/sqlc.yaml` + `transactions/db/query/` → `transactions/db/sqlc/`
- (legacy deposits/integrations follow the same layout)

Generation uses sqlc v1.29.0 pinned via `go:generate` in each service's
`db/doc.go`. Run from a service directory:

```bash
make generate   # go generate ./... (sqlc + mocks)
```

Generated sqlc code is committed and never edited by hand. See
`docs/platform-protobuf-generation-review.md` (CI drift verification uses
`go generate ./...`).

## Database and migrations

- Every service uses golang-migrate with up/down SQL migrations in
  `<service>/db/migrations/`.
- Migrations run at service startup when `RUN_MIGRATIONS=true`, or externally
  (Render/OCI one-shot migration job) when false.
- The migration roadmap is `docs/migration-plan.md`; the domain boundaries are
  `docs/domain-model.md`.
- Monetary values are `NUMERIC(18,2)` passed as decimal strings via
  `commongrpc.Money`; no floating point for money.
- Financial history is preserved: foreign keys use `ON DELETE RESTRICT`.

## Testing

```bash
go test ./...                  # repository-wide
go vet ./...                   # static analysis
gofmt -l .                     # formatting check (hand-written Go)
```

Service-level suites (e.g. clients, transactions) pass without external
infrastructure; the gateway wiring tests use in-process fakes.

## Observability

- **Logs**: structured JSON to stdout/stderr (zerolog). `LOG_LEVEL` controls
  verbosity (`debug`, `info`, `warn`, `error`).
- **Request IDs**: every HTTP/gRPC request receives or propagates an
  `X-Request-ID`; it is echoed in the response and attached to access/RPC
  logs for error correlation.
- **Access logs**: HTTP gateway requests are logged with method, path, status,
  duration, and request ID. `/healthz` probes are logged at debug to avoid
  noise.
- **Health**: each service exposes `/healthz` (HTTP) and the gRPC health
  protocol. Render uses `/healthz`.

## Docker

Every service has a multi-stage distroless Dockerfile. The build context is the
repository root (the services import root `go.mod`, generated `grpc/go/`, and
`shared/`):

```bash
docker build -f clients/Dockerfile -t rvpay-clients:ci .
docker build -f transactions/Dockerfile -t rvpay-transactions:ci .
docker build -f deposits/Dockerfile -t rvpay-deposits:ci .
```

Images run as the non-root `nonroot` user, contain only the compiled binary and
migrations, and take all configuration from the environment. The OCI Compose
stack (`docker-compose.yml`) is deposits-only; the Render Blueprint
(`render.yaml`) covers all three web services. See
`docs/platform-docker-review.md`.

## CI/CD

`.github/workflows/render-deploy.yml` (active on pushes to `main` and
`workflow_dispatch`):

1. `generate` — `go generate ./...`, protoc generation, sqlc generation, then
   fails if generated code differs from the committed state.
2. `validate` — `gofmt` (hand-written Go), `go vet ./...`, `go build ./...`.
3. `test` — `go test ./...`.
4. `docker-build` — matrix over `deposits`, `clients`, `transactions`
   Dockerfiles.
5. `deploy` — POSTs to the `RENDER_DEPLOY_HOOK` GitHub secret when set and the
   branch is `main`.

`.github/workflows/deploy.yml` (OCI) is intentionally disabled (no `on:`
trigger). See `docs/platform-ci-cd-review.md`.

## Deployment

### Render (primary)

`render.yaml` is the Render Blueprint (version-controlled infrastructure). It
deploys:

- `rvpay-deposits` → database `rvpay-postgres`
- `rvpay-clients` → database `rvpay-clients-postgres`
- `rvpay-transactions` → database `rvpay-transactions-postgres`

Each service is a Docker web service using the repository root as the build
context, listens on the Render-injected `PORT` for the HTTP gateway (health
check `/healthz`) and `LISTEN_PORT` for gRPC, and runs migrations at startup.
Manual `sync: false` secrets must be set in the Render dashboard:
`PAWAPAY_API_URL`/`PAWAPAY_API_KEY` (deposits) and
`HIGHLEVEL_CLIENT_ID`/`HIGHLEVEL_CLIENT_SECRET`/`HIGHLEVEL_REDIRECT_URI`/
`WEBHOOK_SECRET` (clients). See `deploy/render/README.md` and
`docs/platform-render-review.md`. Render's free tier includes one managed
PostgreSQL; three databases require a paid plan (see the deploy docs for the
single-database fallback).

### OCI

`docker-compose.yml` + `deploy/README.md` describe the Oracle Cloud Always Free
stack (PostgreSQL, one-shot migration job, deposits service, non-root Nginx TLS
proxy). The OCI GitHub Actions pipeline is currently disabled.

## Configuration

All configuration is environment-driven. Each service documents its own
variables in its README and `.env.example`. The common set:

| Variable | Services | Purpose | Required |
| --- | --- | --- | --- |
| `LOG_LEVEL` | all | zerolog level (`debug`, `info`, ...) | no (default `info`) |
| `LISTEN_PORT` | all | gRPC server port | yes (transactions requires it) |
| `PORT` | all | HTTP gateway port (Render injects it) | no (default `8080`) |
| `MIGRATION_PATH` | all | migration directory (e.g. `db/migrations`) | transactions requires it |
| `RUN_MIGRATIONS` | all | run migrations at startup | no (default `true`) |
| `DB_HOST` | all | PostgreSQL host | yes |
| `DB_PORT` | all | PostgreSQL port | yes |
| `DB_USER` | all | PostgreSQL user | yes |
| `DB_PASSWORD` | all | PostgreSQL password (secret) | yes |
| `DB_NAME` | all | PostgreSQL database | yes |
| `DB_TLS_DISABLED` | all | `true` → `sslmode=disable` | no |
| `PAWAPAY_API_URL` | deposits | PawaPay API base URL (secret) | deposits |
| `PAWAPAY_API_KEY` | deposits | PawaPay credential (secret) | deposits |
| `HIGHLEVEL_CLIENT_ID` | clients, integrations | HighLevel OAuth client ID (secret) | clients |
| `HIGHLEVEL_CLIENT_SECRET` | clients, integrations | HighLevel OAuth client secret (secret) | clients |
| `HIGHLEVEL_REDIRECT_URI` | clients, integrations | OAuth redirect URL (production must be the deployed public URL) | clients |
| `HIGHLEVEL_SSO_KEY` | integrations | HighLevel webhook verification key (secret) | integrations |
| `WEBHOOK_SECRET` | clients | Client webhook verification secret (secret) | clients |
| `TOKEN_ENCRYPTION_KEY` | integrations | 32-byte AES-256 key for token encryption (secret) | integrations |
| `HIGHLEVEL_PAYMENT_URL` | clients | GHL Custom Payment Provider frontend checkout URL (`paymentsUrl`) | no |
| `HIGHLEVEL_QUERY_URL` | clients | GHL Custom Payment Provider backend query URL (`queryUrl`) | no |
| `HIGHLEVEL_PROVIDER_NAME` | clients | GHL Custom Payment Provider display name | no |
| `HIGHLEVEL_PROVIDER_DESCRIPTION` | clients | GHL Custom Payment Provider description | no |
| `HIGHLEVEL_PROVIDER_IMAGE_URL` | clients | GHL Custom Payment Provider image URL | no |
| `TRANSACTIONS_GRPC_ADDR` | clients | Transactions service gRPC address used by the GHL Custom Payment Provider query/webhook endpoints | no |

Secrets are placeholders only in this repository; real values are supplied via
environment (`.env`, Render dashboard, or CI secrets). See each service README
for the full variable list.

### OAuth and webhooks

- The Clients service owns the HighLevel OAuth flow: user installs →
  provider redirects to the callback → authorization code is exchanged →
  tokens are encrypted and stored by the service.
- Production OAuth redirect URLs must point at the deployed public endpoint,
  never `localhost`.
- HighLevel webhook events are ingested, persisted, and processed by the
  Clients (and legacy Integrations) service.

## Troubleshooting

- **Database connection failures:** `localhost` inside a container refers to
  the container itself. In Docker/Render, point `DB_HOST` at the actual
  PostgreSQL host (e.g. Render's `fromDatabase` value) and verify `DB_PORT`/
  `DB_NAME`/credentials. If the pool can't be created or pinging fails, the
  service exits with a clear error.
- **Missing environment variables:** transactions fails config load if
  `LISTEN_PORT`, `MIGRATION_PATH`, or any `DB_*` is missing; clients defaults
  several values. Check the service startup log for the exact variable.
- **Generated-code drift (CI):** run `go generate ./...` and
  `cd protobuf && make generate-protos`, then commit the generated output.
- **SQLC drift (CI):** run `go generate ./...` from a service directory and
  commit the generated output.
- **Docker build failure:** the build context must be the repository root
  (`docker build -f <service>/Dockerfile .`) so root `go.mod`/`go.sum`,
  `grpc/go/`, and `shared/` are available; generated code is committed, so no
  protoc/sqlc install is needed to build an image.
- **Render health check failing:** confirm `PORT`/`LISTEN_PORT` are set and
  `/healthz` returns 200 on `PORT`; the health check path is `/healthz`, not
  `/` (see `docs/platform-render-review.md`).

## Architecture references

- `docs/domain-model.md` — entities, bounded contexts, service ownership.
- `docs/repository-layout.md` — target repository structure.
- `docs/protobuf-strategy.md` — protobuf ownership, packages, shared types,
  versioning, gateway.
- `docs/migration-plan.md` — ordered migration roadmap and phases.
- `docs/platform-repository-audit.md` — platform baseline.
- `docs/platform-*.md` — platform reviews (generation, gateway, common
  packages, CI/CD, Docker, Render).
- `agents/project-context.md` — coding, package, naming, generation, testing
  conventions.
- `agents/platform/*.md` — working agent directives (Cline).

See each service's `README.md` for its runtime and database details, and
`protobuf/README.md` for the API contract and code-generation workflow.