# Deposits service

The Deposits service is the currently runnable `rvpay-go` microservice. It
exposes one gRPC operation, writes deposit records to PostgreSQL, and invokes
the PawaPay client to start a mobile-money deposit.

## Runtime flow

```text
gRPC client
    │ InitiateDeposit
    ▼
DepositsService
    ├── creates the current hard-coded client record
    ├── creates a deposit record in PostgreSQL
    └── calls PawaPay InitiateDeposit
    ▼
CreateDepositResponse (ACCEPTED, FINAL_STATUS)
```

The process entry point is `cmd/grpc-service/main.go`. It loads `.env` from
its working directory when present, reads environment variables, connects and
pings PostgreSQL, applies migrations, registers the service with gRPC
reflection and a unary panic-recovery interceptor, then listens on
`:$LISTEN_PORT`. It stops gracefully on `SIGINT` or `SIGTERM`.

## Directory guide

```text
deposits/
├── cmd/grpc-service/main.go      # gRPC server bootstrap and shutdown
├── config/model.go               # Config and DBConfig environment bindings
├── db/
│   ├── migrations/               # 000001 creates clients and deposits
│   ├── query/                    # Create/Get client and deposit SQL
│   ├── repo/                     # pgx pool adapter plus migration helpers
│   ├── sqlc/                     # sqlc-generated data access code
│   └── doc.go                    # go:generate directives
├── deposits/service.go           # InitiateDeposit implementation
├── deposits/service_test.go      # test package (no tests yet)
├── .env                          # local runtime configuration; do not commit
└── Makefile                      # local development tasks
```

## Configuration

Copy the repository template and edit it:

```bash
cp ../.env.example .env
```

| Variable | Required | Purpose |
| --- | --- | --- |
| `PAWAPAY_API_URL` | No config-level validation | PawaPay API base URL |
| `PAWAPAY_API_KEY` | No config-level validation | PawaPay credential |
| `LOG_LEVEL` | No; defaults to `debug` | Zerolog level |
| `LISTEN_PORT` | Yes | gRPC TCP port |
| `MIGRATION_PATH` | Yes | Migration directory, typically `db/migrations` |
| `DB_USER` | Yes | PostgreSQL user |
| `DB_PASSWORD` | Yes | PostgreSQL password |
| `DB_HOST` | Yes | PostgreSQL host |
| `DB_PORT` | Yes | PostgreSQL port |
| `DB_NAME` | Yes | PostgreSQL database |
| `DB_TLS_DISABLED` | No; defaults to `true` | Selects `sslmode=disable`; set `false` for `require` |
| `env` | No | Defined but not currently consumed |
| `BANK_FILE_PATH` | No | Defined but not currently consumed |

## Local startup

Run all commands below from this directory.

```bash
make rundb
docker exec -it deposits-postgres psql -U postgres -d deposits -c 'CREATE EXTENSION IF NOT EXISTS pgcrypto;'
make run
```

`make rundb` starts a detached PostgreSQL 16 Alpine container named
`deposits-postgres`, exposes `DB_PORT`, and uses `DB_USER`, `DB_PASSWORD`, and
`DB_NAME` from `.env`. The service applies migrations automatically at startup.
The initial schema relies on `gen_random_uuid()`, so `pgcrypto` must be enabled
in the database.

## Service API

The generated gRPC service name is
`depositsgrpc.DepositsService`, with method
`InitiateDeposit(CreateDepositRequest) -> CreateDepositResponse`.

The request takes an amount string, three-letter uppercase currency, payer
type, phone number, provider, and a `client_id`. In the current implementation,
`client_id` is ignored; see the current-behavior notes below. The full protobuf
schema is [../protobuf/deposits.proto](../protobuf/deposits.proto).

Example with local reflection enabled:

```bash
grpcurl -plaintext \
  -d '{"amount":"1000.00","currency":"XAF","payer":{"type":"DEPOSIT_PORTAL_MMO","accountDetails":{"phoneNumber":"+237699541235","provider":"DEPOSIT_PROVIDER_MTN_MOMO_CMR"}}}' \
  localhost:50051 depositsgrpc.DepositsService/InitiateDeposit
```

## Database and generated code

- `db/migrations` is the database source of truth.
- `db/query/*.sql` contains the SQL queries consumed by sqlc.
- `db/sqlc` is generated code; update migrations/queries and run `make generate`
  rather than editing it manually.
- `db/repo` exposes `Do()` for ordinary queries and `Begin()` for callers that
  need a transaction. The existing service uses `Do()`.

## Make targets

```bash
make install-tools
make generate
make lint
make test
make run
make rundb
make create-migration name=descriptive_migration_name
```

`build` and `publish` expect a `Dockerfile` and `full_image_tag`; no service
Dockerfile is currently present, so those targets are not ready to use.

## Current behavior and limitations

- The service currently inserts the same hard-coded client for every request.
  Unique email/phone constraints mean requests after the first will fail until
  client lookup/upsert behavior is implemented.
- Only MMO payer data is mapped meaningfully. Unsupported payer types default
  to MMO; unsupported providers default to MTN MoMo Cameroon.
- A database deposit is created before the PawaPay API call. Failures from that
  API leave the locally persisted deposit in place.
- No HTTP server, authentication, webhook/callback processing, deposit-status
  polling, or reconciliation job is implemented in this service.
