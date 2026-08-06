# rvpay-go

`rvpay-go` is a Go microservices repository that currently contains one runnable
gRPC service: **Deposits**. The service accepts an `InitiateDeposit` request,
stores a client and deposit in PostgreSQL, and calls the PawaPay client to
initiate the external mobile-money deposit.

This README describes the repository **as it is today**. It is not an HTTP
server: the protobuf contains an HTTP annotation, but no gRPC-Gateway or HTTP
handler is started.

## Requirements

- Go 1.26.5 (the version declared in `go.mod`)
- Docker, to start the development PostgreSQL container
- PostgreSQL 16-compatible database if you do not use Docker
- PawaPay API URL and API key
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
│   └── Makefile                  # Service tasks
├── protobuf/                     # Source protobuf contracts and generation task
├── grpc/go/depositsgrpc/         # Generated Go protobuf and gRPC stubs
├── third_party/googleapis/       # googleapis Git submodule used by protoc
├── .env.example                  # Environment-variable template
├── Makefile                      # Repository-wide test tasks
└── layout.md                     # Original layout notes
```

See [deposits/README.mds](deposits/README.mds) for the service’s runtime and
database details, and [protobuf/README.mds](protobuf/README.mds) for the API
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
   middleware, then listens on `:$LISTEN_PORT`. `SIGINT` and `SIGTERM` trigger
   a graceful gRPC stop.

## Call the API

With `grpcurl` installed and the service listening on port `50051`, reflection
allows a local plaintext request without supplying proto files:

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

The contract declares `POST /v1/public/deposits`, but this repository does not
run an HTTP gateway, so send the request over gRPC.

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

From `protobuf/`:

```bash
make lint
make generate-protos
```

## Current implementation notes

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
