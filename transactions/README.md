# Transactions service

The Transactions service is the `rvpay-go` microservice that owns the payment
transaction domain: merchants, customers, deposits, and payouts. It exposes
gRPC operations plus a REST/gRPC-gateway surface, persists transaction records
to PostgreSQL, and is wired for future provider integration.

## Purpose

The Transactions service implements the documented RVPay transaction
capabilities:

- **Merchants** — payment gateway records (create, get, list).
- **Customers** — end-user payment records tied to a client and merchant
  (create, get).
- **Deposits** — inbound customer payments (initiate, get).
- **Payouts** — outbound settlements (request, get).

It follows the domain model in `docs/domain-model.md` and the layout in
`docs/repository-layout.md`.

## Responsibilities

- Own the Transactions PostgreSQL database (merchants, customers, deposits,
  payouts).
- Expose the Transactions gRPC API (`transactionsgrpc`) and REST gateway.
- Validate merchant/customer ownership and financial record integrity.
- Apply database migrations at startup when enabled.
- Remain provider-agnostic: no external payment provider is called from this
  service yet.

## Service Structure

```text
transactions/
├── cmd/grpc-service/main.go      # gRPC + gateway server bootstrap and shutdown
├── config/model.go               # Config and DBConfig environment bindings
├── db/
│   ├── migrations/               # 000001 creates merchants, customers, deposits, payouts
│   ├── query/                    # SQL inputs consumed by sqlc
│   ├── repo/                     # pgx pool adapter plus domain repositories
│   ├── sqlc/                     # sqlc-generated data access code
│   └── doc.go                    # go:generate directives (sqlc + mocks)
├── merchants/                    # MerchantService implementation
├── customers/                    # CustomerService implementation
├── deposits/                     # DepositService implementation
├── payouts/                      # PayoutService implementation
├── Dockerfile                    # Multi-stage distroless build
├── Makefile                      # Local development tasks
├── README.md                     # This file
└── .env.example                  # Runtime environment template
```

## Running Locally

Run all commands below from the `transactions/` directory.

1. Configure the environment:

```bash
cp .env.example .env
# edit .env with your local PostgreSQL values
```

2. Start PostgreSQL:

```bash
make rundb
docker exec -it transactions-postgres psql -U postgres -d transactions -c 'CREATE EXTENSION IF NOT EXISTS pgcrypto;'
```

`make rundb` starts a detached PostgreSQL 16 Alpine container named
`transactions-postgres`, exposes `DB_PORT`, and uses `DB_USER`, `DB_PASSWORD`,
and `DB_NAME` from `.env`.

3. Run the service:

```bash
make run
```

The service loads `.env`, connects and pings PostgreSQL, applies migrations
(when `RUN_MIGRATIONS=true`), registers the four Transactions services with
gRPC reflection and a unary panic-recovery interceptor, then listens on
`:$LISTEN_PORT`. The REST gateway listens on `:$PORT` (default `8080`). It
stops gracefully on `SIGINT` or `SIGTERM`.

The initial schema relies on `gen_random_uuid()`, so `pgcrypto` must be enabled
in the database.

## Configuration

| Variable | Required | Purpose |
| --- | --- | --- |
| `LOG_LEVEL` | No; defaults to `debug` | Zerolog level |
| `LISTEN_PORT` | Yes | gRPC TCP port |
| `MIGRATION_PATH` | Yes | Migration directory, typically `transactions/db/migrations` |
| `RUN_MIGRATIONS` | No; defaults to `true` | Apply migrations at startup when true |
| `DB_USER` | Yes | PostgreSQL user |
| `DB_PASSWORD` | Yes | PostgreSQL password |
| `DB_HOST` | Yes | PostgreSQL host |
| `DB_PORT` | Yes | PostgreSQL port |
| `DB_NAME` | Yes | PostgreSQL database |
| `DB_TLS_DISABLED` | No; defaults to `false` | Selects `sslmode=disable` when true; set `false` for `require` |

No secrets are committed. All credentials are provided through the environment
(`.env` locally, platform environment variables in production).

## Database

- PostgreSQL is required.
- Migrations are owned by `transactions/db/migrations` and applied at startup
  when `RUN_MIGRATIONS=true`.
- SQLC generates the data-access code from `transactions/db/query/*.sql`.
- `db/repo` exposes `TransactionsRepo.Do()`/`Begin()` plus the domain
  repositories (Merchant, Customer, Deposit, Payout).

## Code Generation

Generation ownership:

| Generated Artifact | Command | Owner |
| --- | --- | --- |
| protobuf + gRPC + gateway | `cd protobuf && make generate-protos` | protobuf Makefile |
| SQLC models/queries | `cd transactions/db && go generate ./...` (or `make generate`) | Transactions DB |
| repository/queries mocks | `cd transactions/db && go generate ./...` (mockgen v0.6.0) | Transactions DB |

Run `make generate` from the `transactions/` directory to regenerate SQLC and
mocks. Do not manually edit generated files.

## Testing

```bash
make test
# or
go test ./...
```

Focused Transactions tests:

```bash
go test ./merchants/...
go test ./customers/...
go test ./deposits/...
go test ./payouts/...
go test ./config/...
```

## Docker

```bash
# from the repository root (the Docker build context)
docker build -f transactions/Dockerfile -t rvpay-go-transactions:local .
```

Run with the required environment variables:

```bash
docker run --rm \
  -e LISTEN_PORT=50051 \
  -e DB_USER=postgres -e DB_PASSWORD=secret -e DB_HOST=host.docker.internal \
  -e DB_PORT=5432 -e DB_NAME=transactions -e DB_TLS_DISABLED=true \
  -e MIGRATION_PATH=/app/transactions/db/migrations \
  -e PORT=8080 \
  -p 50051:50051 -p 8080:8080 \
  rvpay-go-transactions:local
```

The container runs as a non-root distroless user and binds to the configured
ports. No secrets are baked into the image.

## gRPC API

The generated gRPC service package is `transactionsgrpc`, registering four
services:

| Service | RPCs |
| --- | --- |
| `MerchantService` | `CreateMerchant`, `GetMerchant`, `ListMerchants` |
| `CustomerService` | `CreateCustomer`, `GetCustomer` |
| `DepositService` | `InitiateDeposit`, `GetDeposit` |
| `PayoutService` | `RequestPayout`, `GetPayout` |

Example with local reflection enabled:

```bash
grpcurl -plaintext \
  -d '{"name":"PawaPay","slug":"pawapay"}' \
  localhost:50051 transactionsgrpc.MerchantService/CreateMerchant
```

The full protobuf schema is `../protobuf/transactions.proto`.

## REST API

The grpc-gateway exposes REST routes generated from the protobuf annotations:

| Method | Route | RPC |
| --- | --- | --- |
| POST | `/v1/public/merchants` | CreateMerchant |
| GET | `/v1/public/merchants/{merchant_id}` | GetMerchant |
| GET | `/v1/public/merchants` | ListMerchants |
| POST | `/v1/public/customers` | CreateCustomer |
| GET | `/v1/public/customers/{customer_id}` | GetCustomer |
| POST | `/v1/public/deposits` | InitiateDeposit |
| GET | `/v1/public/deposits/{deposit_id}` | GetDeposit |
| POST | `/v1/public/payouts` | RequestPayout |
| GET | `/v1/public/payouts/{payout_id}` | GetPayout |

The gateway listens on `:$PORT` (default `8080`) and serves `/healthz`.

## Migrations

- Up/down migrations live in `transactions/db/migrations` (`000001_init_schema`).
- Migrations run automatically at startup when `RUN_MIGRATIONS=true`.
- Create a new migration with:

```bash
make create-migration name=descriptive_migration_name
```

`make create-migration` requires the `migrate` CLI. Destructive rollback is not
wired into the default Makefile targets.

## Architecture Notes

```text
gRPC API (transactionsgrpc)
    ↓
MerchantService / CustomerService / DepositService / PayoutService
    ↓
MerchantRepo / CustomerRepo / DepositRepo / PayoutRepo
    ↓
SQLC (transactions/db/sqlc)
    ↓
PostgreSQL
```

- Each service embeds the generated `Unimplemented*ServiceServer` for forward
  compatibility.
- The Deposit service validates the customer belongs to the declared
  client+merchant+phone context before persisting.
- Payouts are not customer-scoped; they reference client and merchant only.
- Provider execution (PawaPay etc.) is intentionally not wired in this
  service; the deposit/payout lifecycle is ready for a future integration
  boundary.

## Troubleshooting

- **Database connection failure** — verify `DB_USER`, `DB_PASSWORD`, `DB_HOST`,
  `DB_PORT`, `DB_NAME` in `.env`; ensure PostgreSQL is running and reachable.
- **Missing environment variables** — the service requires `LISTEN_PORT`,
  `MIGRATION_PATH`, and the `DB_*` variables; otherwise startup fails with a
  configuration error.
- **Migrations do not apply** — ensure `MIGRATION_PATH` points to
  `transactions/db/migrations` and that `pgcrypto` is enabled in the database.
- **Port conflicts** — the gRPC port (`LISTEN_PORT`) and gateway port (`PORT`)
  must be free; if a port is already in use, startup fails with
  `net.Listen: bind: address already in use`.
- **Generated code is stale** — after changing migrations or SQL, run
  `make generate`; never hand-edit `db/sqlc` output.
- **Docker build** — the build context is the repository root:
  `docker build -f transactions/Dockerfile ..` from `transactions/`.