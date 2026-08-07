# rvpay-go

`rvpay-go` is a Go microservices repository that currently contains two runnable
gRPC services:

- **Deposits** — accepts an `InitiateDeposit` request, stores a client and
  deposit in PostgreSQL, and calls the PawaPay client to initiate the external
  mobile-money deposit.
- **Integrations** — manages third-party provider connections. It handles the
  HighLevel OAuth callback flow, stores encrypted provider tokens, and receives
  and persists HighLevel webhook events.

Both services expose a gRPC API and an HTTP gateway (gRPC-Gateway) with a
`/healthz` endpoint. This README describes the repository **as it is today**.

## Requirements

- Go 1.26.5 (the version declared in `go.mod`)
- Docker, to start the development PostgreSQL container
- PostgreSQL 16-compatible database if you do not use Docker
- PawaPay API URL and API key (Deposits service)
- HighLevel OAuth credentials and a token encryption key (Integrations service)
- `protoc`, `protoc-gen-go`, and `protoc-gen-go-grpc` only when regenerating
  protobuf code

## Repository layout

```text
.
├── deposits/                    # Deposits gRPC service
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
├── integrations/                # Integrations gRPC service
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
├── protobuf/                     # Source protobuf contracts and generation task
├── grpc/go/                      # Generated Go protobuf and gRPC stubs
│   ├── depositsgrpc/             # Generated Deposits service code
│   └── integrationsgrpc/         # Generated Integrations service code
├── third_party/googleapis/       # googleapis Git submodule used by protoc
├── nginx/                        # Nginx TLS termination config for OCI
├── deploy/                       # OCI and Render deployment documentation
├── .github/workflows/            # CI/CD pipelines (OCI and Render)
├── .env.example                  # Environment-variable template
├── docker-compose.yml            # OCI Always Free Compose stack
├── render.yaml                   # Render Blueprint
├── Makefile                      # Repository-wide test tasks
└── layout.md                     # Original layout notes
```

See [deposits/README.md](deposits/README.md) and
[integrations/README.md](integrations/README.md) for each service's runtime and
database details, and [protobuf/README.md](protobuf/README.md) for the API
contract and code-generation workflow.

## Run the Deposits service

1. Initialise the Google API definitions submodule when working from a fresh
   clone (needed only for protobuf imports/regeneration):

   ```bash
   git submodule update --init --recursive
   ```

2. Create the service environment file. `LoadConfig` looks for `.env` in the
   current working directory, so it belongs in `deposits/` when using the
   service Makefile.

   ```bash
   cp .env.example deposits/.env
   ```

3. Edit `deposits/.env` and provide at least the following values:

   ```dotenv
   PAWAPAY_API_URL=<PawaPay API base URL>
   PAWAPAY_API_KEY=<PawaPay API key>
   LOG_LEVEL=debug
   DB_USER=postgres
   DB_PASSWORD=<local password>
   DB_HOST=localhost
   DB_PORT=5432
   DB_NAME=deposits
   DB_TLS_DISABLED=true
   MIGRATION_PATH=db/migrations
   LISTEN_PORT=50051
   ```

   `env` and `BANK_FILE_PATH` are defined by the config but are not currently
   required or used by the server. Keep credentials out of source control.

4. Start PostgreSQL from the `deposits/` directory:

   ```bash
   cd deposits
   make rundb
   ```

   The initial migration uses `gen_random_uuid()`. Ensure the target database
   has the `pgcrypto` extension installed before startup:

   ```bash
   docker exec -it deposits-postgres psql -U postgres -d deposits -c 'CREATE EXTENSION IF NOT EXISTS pgcrypto;'
   ```

5. Start the gRPC service from the same directory:

   ```bash
   make run
   ```

   On startup it loads configuration, connects and pings PostgreSQL, runs all
   up migrations from `MIGRATION_PATH`, registers gRPC reflection and recovery
   middleware, then listens on `:$LISTEN_PORT`. It also starts an HTTP gateway
   on `:$PORT` (default `8080`) serving the REST endpoints and `/healthz`.
   `SIGINT` and `SIGTERM` trigger a graceful shutdown of both servers.

## Run the Integrations service

1. Create the service environment file:

   ```bash
   cp .env.example integrations/.env
   ```

2. Edit `integrations/.env` and provide at least the following values:

   ```dotenv
   HIGHLEVEL_CLIENT_ID=<HighLevel OAuth client ID>
   HIGHLEVEL_CLIENT_SECRET=<HighLevel OAuth client secret>
   HIGHLEVEL_REDIRECT_URL=<OAuth redirect URL>
   HIGHLEVEL_SSO_KEY=<HighLevel webhook verification key>
   TOKEN_ENCRYPTION_KEY=<32-byte AES key>
   LOG_LEVEL=debug
   DB_USER=postgres
   DB_PASSWORD=<local password>
   DB_HOST=localhost
   DB_PORT=5432
   DB_NAME=integrations
   DB_TLS_DISABLED=true
   MIGRATION_PATH=db/migrations
   LISTEN_PORT=50052
   ```

   `TOKEN_ENCRYPTION_KEY` is required and must be a 32-byte AES-256 key.
   `HIGHLEVEL_CLIENT_ID`, `HIGHLEVEL_CLIENT_SECRET`, `HIGHLEVEL_REDIRECT_URL`,
   and `HIGHLEVEL_SSO_KEY` are defined by the config but are not currently
   validated at config load. Keep credentials out of source control.

3. Start PostgreSQL from the `integrations/` directory:

   ```bash
   cd integrations
   make rundb
   ```

   The initial migration uses `gen_random_uuid()`. Ensure the target database
   has the `pgcrypto` extension installed before startup:

   ```bash
   docker exec -it integrations-postgres psql -U postgres -d integrations -c 'CREATE EXTENSION IF NOT EXISTS pgcrypto;'
   ```

4. Start the gRPC service from the same directory:

   ```bash
   make run
   ```

   On startup it loads configuration, connects and pings PostgreSQL, runs all
   up migrations from `MIGRATION_PATH`, registers gRPC reflection, health, and
   recovery middleware, then listens on `:$LISTEN_PORT`. It also starts an HTTP
   gateway on `:$PORT` (default `8080`) serving the REST endpoints, `/healthz`,
   `/oauth/callback`, and `/webhooks/highlevel`. `SIGINT` and `SIGTERM` trigger
   a graceful shutdown of both servers.

## Call the API

With `grpcurl` installed and the service listening, reflection allows a local
plaintext request without supplying proto files.

Deposits service on port `50051`:

```bash
grpcurl -plaintext \
  -d '{
    "amount": "1000.00",
    "currency": "XAF",
    "payer": {
      "type": "DEPOSIT_PORTAL_MMO",
      "accountDetails": {
        "phoneNumber": "+237699541235",
        "provider": "DEPOSIT_PROVIDER_MTN_MOMO_CMR"
      }
    },
    "clientId": "not-currently-used"
  }' \
  localhost:50051 depositsgrpc.DepositsService/InitiateDeposit
```

Integrations service on port `50052`:

```bash
grpcurl -plaintext \
  -d '{
    "provider": "highlevel",
    "locationId": "loc_123",
    "accessToken": "...",
    "refreshToken": "...",
    "tokenExpiresAt": "2026-08-08T00:00:00Z"
  }' \
  localhost:50052 integrationsgrpc.IntegrationService/CreateIntegration
```

Both services also expose their REST endpoints through the HTTP gateway on
`:$PORT` (default `8080`). For example, the Integrations service exposes:

- `POST /v1/public/integrations`
- `GET /v1/public/integrations/{id}`
- `DELETE /v1/public/integrations/{id}`
- `POST /v1/public/webhooks`

## Development tasks

From the repository root:

```bash
make test
make bench-test
go test ./...
```

From `deposits/`:

```bash
make install-tools   # installs sqlc, mockgen, and golangci-lint
make generate        # regenerates sqlc and mocks via go generate
make lint
make test
make create-migration name=add_example
```

From `integrations/`:

```bash
make install-tools   # installs sqlc, mockgen, and golangci-lint
make generate        # regenerates sqlc and mocks via go generate
make lint
make test
make create-migration name=add_example
```

From `protobuf/`:

```bash
make lint
make generate-protos
```

## Deployment

The repository includes deployment artifacts for two targets:

- **Oracle Cloud Always Free** — see [deploy/README.md](deploy/README.md). The
  `docker-compose.yml` stack runs PostgreSQL, a one-shot migration job, the
  Deposits service, and an unprivileged Nginx TLS proxy on a single ARM64
  `VM.Standard.A1.Flex` instance.
- **Render** — see [deploy/render/README.md](deploy/render/README.md). The
  `render.yaml` Blueprint provisions a web service and a managed PostgreSQL
  database.

CI/CD pipelines live in `.github/workflows/`:

- `deploy.yml` — OCI delivery pipeline (currently disabled, no `on:` trigger).
- `render-deploy.yml` — Render delivery pipeline: generate, test, Docker build,
  then trigger a Render deploy hook.

## Current implementation notes

### Deposits

- `InitiateDeposit` persists a hard-coded client (`Socadel`) instead of using
  `client_id` from the request. Because `email` and `phone_number` are unique,
  a second successful request will currently fail when creating that client.
- The service supports the mobile-money payer type and MTN/Orange Cameroon
  providers in the contract. The database stores Orange as `ORANGE_CMR`, while
  the service converts it to `ORANGE_MOMO_CMR` for the PawaPay client.
- Deposits are saved before the PawaPay call. There is no transaction spanning
  the database write and external API request, callback handling, or status
  reconciliation yet.
- `CreateDepositResponse` currently always returns `ACCEPTED` and
  `FINAL_STATUS` after PawaPay accepts the initiation call.

### Integrations

- Only the HighLevel provider is implemented for OAuth and webhooks.
- Webhook events are stored and processed asynchronously; internal service
  communication (e.g. notifying the deposits service) is not yet implemented.
- `HIGHLEVEL_SSO_KEY` is defined in configuration but webhook signature
  verification is not yet enforced.
- The `ProcessWebhookEvent` gRPC method is defined in the protobuf contract but
  is not yet implemented in the service.