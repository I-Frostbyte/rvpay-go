# Transactions Service — Existing Implementation Review

Document Version: 1.0
Status: Foundation Audit
System: RVPay
Audit: Agent 01 — Review Existing Transactions Implementation

## 1. Executive Summary

The RVPay repository currently contains two runnable gRPC microservices:
`deposits/` and `integrations/`. The new architecture in `docs/domain-model.md`
defines a broader Transactions Service that will own Merchants, Customers,
Deposits, and Payouts.

The existing `deposits/` service is the canonical implementation reference.
It is the template that the Transactions Service must follow. The Deposits
service already implements the Deposit flow, client persistence (with a
hard-coded `Socadel` client), PawaPay integration, and the full repository
conventions (runtime, configuration, database, sqlc, repositories, Docker,
Makefile, tests).

This review documents the existing Deposits architecture, establishes the
gap analysis between the current implementation and the target Transactions
architecture, and provides the implementation map for Agents 02–13.

**This agent is a pure audit. No implementation, migration, or modification
was performed.**

## 2. Existing Deposits Architecture

The existing `deposits/` service follows the canonical `cmd/grpc-service/`,
`config/`, `db/`, `<domain>/` layout:

```text
deposits/
├── cmd/grpc-service/main.go      # gRPC server bootstrap, gateway, shutdown
├── config/model.go               # Config and DBConfig environment bindings
├── db/
│   ├── migrations/               # 000001 creates clients and deposits
│   ├── query/                    # Create/Get client and deposit SQL
│   ├── repo/                     # pgx pool adapter + migration helpers
│   ├── sqlc/                     # sqlc-generated data access code
│   └── doc.go                    # go:generate directives
├── deposits/service.go           # InitiateDeposit implementation
├── deposits/service_test.go      # test package
├── .env.example                  # local runtime configuration
├── Dockerfile                    # multi-stage distroless build
├── Makefile                      # local development tasks
└── README.md                     # service documentation
```

### Responsibilities (current)

- Accepts `InitiateDeposit` gRPC + REST requests.
- Creates/stores a client record (hard-coded `Socadel`).
- Creates a deposit record in PostgreSQL.
- Calls PawaPay `InitiateDeposit` to start a mobile-money deposit.
- Exposes gRPC with reflection, unary recovery interceptor, health checks.
- Exposes REST gateway via grpc-gateway v2 with `/healthz`.

## 3. Existing Database

### Migration: `deposits/db/migrations/000001_init_schema.{up,down}.sql`

The up migration creates:

- `clients` — `id UUID PK DEFAULT gen_random_uuid()`, `email UNIQUE`,
  `phone_number UNIQUE`, `client_name`, `created_at`, `updated_at`.
- `deposits` — `id`, `client_id FK`, `amount DECIMAL`, `currency`,
  `status` (enum-like), `payer_*` columns, `provider`,
  `external_reference`, `created_at`, `updated_at`.

### Ownership

The existing `deposits` database mixes two concepts per the target
`docs/domain-model.md`:

- `clients` table — conceptually belongs to the **Clients Service**, but is
  currently owned by the Deposits database.
- `deposits` table — belongs to the **Transactions Service**.

The unique constraints on `email` and `phone_number` currently cause a second
successful request to fail (the hard-coded `Socadel` client is inserted for
every request). This is documented technical debt.

### Determinations

| Concept | Target Ownership | Current Location | Classification |
| --- | --- | --- | --- |
| `clients` | Clients Service | deposits DB | Modify / relocate |
| `deposits` | Transactions Service | deposits DB | Reusable |
| `deposits.status` | Transactions Service | deposits DB | Reusable (extend to 4-state per domain model) |

## 4. Existing SQLC

Config: `deposits/db/sqlc.yaml`. Query inputs: `deposits/db/query/*.sql`.
Generated output: `deposits/db/sqlc/*.sql.go` (`db.go`, `models.go`,
`querier.go`).

### Conventions

- **Query naming** — `CreateClient`, `GetClientByID`, `GetClientByName`,
  `CreateDeposit`, `GetDepositByID`.
- **Models** — generated `Client` and `Deposit` structs.
- **Nullable handling** — uses `pgtype` for nullable fields.
- **Querier interface** — generated `Querier` interface consumed by
  repositories.
- **Generated code** — never hand-edited; regenerated via `make generate`.

### Determinations

The sqlc workflow (migrations → SQL queries → sqlc generate → generated
code → repository) is the canonical pattern and must be reused verbatim by
the Transactions Service. No modification is required to the pattern.

## 5. Existing Repositories

File: `deposits/db/repo/repo.go`.

### Pattern

- Exposes `Do()` for ordinary queries and `Begin()` for transactions.
- `DepositsRepo` wraps the `sqlc.Querier`.
- `Migrate(dbURL, path, logger)` runs golang-migrate migrations at startup.
- Error handling wraps sqlc/pgx errors.

### Determinations

The repository pattern (pool wrapper, `Do()`/`Begin()`, migration runner,
error wrapping) is **Reusable** and must be carried into Transactions. The
existing `deposits` repository concept will be replaced by separate
Merchant, Customer, Deposit, and Payout repositories.

## 6. Existing Service Layer

File: `deposits/deposits/service.go`.

### Pattern

- `DepositsService` struct with `clientsRepo`, `depositsRepo`, `logger`, and a
  `pawapay_client.Client`.
- `NewDepositsService(repo, logger, pawapayClient)` constructor.
- `InitiateDeposit(ctx, req)` — validates request, creates hard-coded client,
  creates deposit, calls PawaPay, returns response.
- Logs via zerolog (value receiver `zerolog.Logger`).
- Uses gRPC `status`/`codes` for error mapping.
- Returns `ACCEPTED`/`FINAL_STATUS` unconditionally.

### Determinations

The service-layer conventions (struct + constructor + transport-independent
business logic + zerolog + gRPC status codes) are **Reusable**. The deposit
flow itself carries into Transactions but:
- the hard-coded `Socadel` client must be replaced with real
  `ClientService.GetClient` resolution,
- the 5-state deposit lifecycle must align with the domain model
  (`Initiated → Processing → Completed / Failed`),
- a Payout flow must be added (merchant, customers, payouts).

## 7. Existing Protobuf

File: `protobuf/deposits.proto` (package `depositsgrpc`).

### Service: `DepositsService`

| RPC | Request | Response |
| --- | --- | --- |
| `InitiateDeposit` | `CreateDepositRequest` | `CreateDepositResponse` |

### Enums

- `DepositStatus` — `UNSPECIFIED`, `PENDING`, `ACCEPTED`, `COMPLETED`,
  `FAILED`, `REJECTED`.
- `DepositType` — `UNSPECIFIED`, `MMO`, `CARD`.
- `DepositProvider` — `UNSPECIFIED`, `MTN_MOMO_CMR`, `ORANGE_MOMO_CMR`.

### Messages

- `AccountDetails` — `phone_number`, `provider`.
- `Participant` — `type`, `account_details`.
- `CreateDepositRequest` — `amount`, `currency`, `payer`, `client_id`.
- `CreateDepositResponse` — `deposit_id`, `status`, `next_step`.

HTTP annotation: `POST /v1/public/deposits`.

Generated code: `grpc/go/depositsgrpc/` (`.pb.go`, `_grpc.pb.go`,
`.pb.gw.go`).

### Determinations

Per `docs/protobuf-strategy.md`, the existing `deposits.proto` is **Replaced**
by `transactions.proto` (package `transactionsgrpc`). The `InitiateDeposit`
RPC and `DepositStatus` enum evolve; `DepositType` and `DepositProvider`
move to shared `commongrpc` enums (`PaymentType`, `Provider`); `Money` becomes
a shared message. Merchant, Customer, Deposit, and Payout RPCs are added.

## 8. Existing Runtime

File: `deposits/cmd/grpc-service/main.go`.

### Startup lifecycle

1. Signal context (SIGINT/SIGTERM).
2. zerolog logger creation with timestamp + caller.
3. `model.Config.LoadConfig()` from environment.
4. zerolog level parse.
5. `pgxpool.New` + `Ping` (eager connection validation).
6. `repo.Migrate` when `RUN_MIGRATIONS` enabled.
7. Repository construction.
8. Service construction (deposits + pawapay client).
9. gRPC server with recovery interceptor, reflection, health server.
10. `depositsgrpc.RegisterDepositsServiceServer`.
11. `net.Listen` + grpc-gateway mux registration.
12. HTTP server with `/healthz`.
13. Concurrent gRPC + HTTP serving with `sync.WaitGroup`.
14. Graceful shutdown on context cancellation (HTTP shutdown, gRPC
    GracefulStop).

### Determinations

The runtime lifecycle is the canonical pattern and is **Reusable** verbatim
for the Transactions Service. The startup sequence
(config → logger → database → migrations → repos → services → handlers →
servers) matches `docs/repository-layout.md` and must be preserved.

## 9. Existing Configuration

File: `deposits/config/model.go`; template: `deposits/.env.example`.

### Variables

| Variable | Required | Purpose |
| --- | --- | --- |
| `PAWAPAY_API_URL` | No | PawaPay API base URL |
| `PAWAPAY_API_KEY` | No | PawaPay credential |
| `LOG_LEVEL` | No; default `debug` | Zerolog level |
| `LISTEN_PORT` | Yes | gRPC TCP port |
| `MIGRATION_PATH` | Yes | Migration directory |
| `DB_USER` / `DB_PASSWORD` / `DB_HOST` / `DB_PORT` / `DB_NAME` | Yes | PostgreSQL |
| `DB_TLS_DISABLED` | No; default `true` | sslmode selection |
| `RUN_MIGRATIONS` | Used | Apply migrations on startup |
| `env` / `BANK_FILE_PATH` | No | Defined but unused |

### Determinations

The configuration pattern (`Config` struct, `LoadConfig()`, `getEnv`,
`getEnvAsInt`, `getEnvAsBool`, `.env.example`) is **Reusable**. Transactions
will reuse the same structure, adding merchant, customer, deposit/payout,
and provider configuration.

## 10. Existing Docker and Build

File: `deposits/Dockerfile`.

### Pattern

- **Build stage** — `golang:1.26.5-alpine`; `go mod download`; copies
  `deposits`, `grpc`; `CGO_ENABLED=0 GOOS=linux go build -trimpath
  -ldflags="-s -w" -o /out/deposits-grpc-service ./deposits/cmd/grpc-service`.
- **Runtime stage** — `gcr.io/distroless/static-debian12:nonroot`;
  copies migrations to `db/migrations`; `USER nonroot:nonroot`;
  `EXPOSE 50051`; `ENTRYPOINT ["./deposits-grpc-service"]`.

Makefile: `deposits/Makefile` provides `install-tools`, `generate`, `lint`,
`test`, `build`, `publish`, `run`, `create-migration`, `rundb`.

### Determinations

The Dockerfile multi-stage/distroless pattern and Makefile targets are
**Reusable** verbatim for the Transactions Service. The binary and
migration paths (`./transactions/cmd/grpc-service`,
`./transactions/db/migrations`) and image name (`rvpay-go-transactions`)
must be adjusted.

## 11. Existing Testing

File: `deposits/deposits/service_test.go`.

### Pattern

- `package deposits` (same-package tests).
- Table-driven tests with `t.Parallel()`.
- `zerolog.Nop()` for loggers.
- `status.Code()` assertions on gRPC errors.
- Tests cover invalid requests, payer/provider enum mapping, and gRPC
  input validation.
- Repository and integration tests not present (require a live PostgreSQL).

### Determinations

The Deposits testing conventions (same-package, table-driven, `t.Parallel`,
`zerolog.Nop`, `status.Code`) are **Reusable**. Transactions will add
repository/service/OAuth/webhook/config test suites following the same style,
plus PostgreSQL integration tests recommended in prior Clients review.

## 12. Deposits → Transactions Gap Analysis

| Area | Existing | Target | Classification |
| --- | --- | --- | --- |
| Service layout | `deposits/cmd/grpc-service/` | `transactions/cmd/grpc-service/` | Reusable |
| Config pattern | `deposits/config/model.go` | `transactions/config/model.go` | Reusable |
| Migration runner | `deposits/db/repo/repo.go` Migrate | `transactions/db/repo/` | Reusable |
| sqlc workflow | `deposits/db/query|sqlc/` | `transactions/db/` | Reusable |
| Deposit repo | `deposits/db/repo` DepositsRepo | `transactions/db/repo` DepositRepo | Modify |
| Client repo | `deposits/db/repo` ClientsRepo | Clients Service (removed) | Deprecated |
| Deposit table | `deposits` deposits table | `transactions` deposits table | Reusable |
| Client table | `deposits` clients table | Clients Service database | Deprecated / move |
| Deposit flow | `deposits/deposits/service.go` | `transactions/deposits/` | Modify |
| Merchant entity | — | `transactions/merchants/` | New |
| Customer entity | — | `transactions/customers/` | New |
| Payout entity | — | `transactions/payouts/` | New |
| PawaPay client | `deposits` service | `transactions/deposits/` + merchants | Reusable |
| Hard-coded client | `Socadel` | `ClientService.GetClient` resolution | Deprecated |
| Deposit lifecycle | 5-state enum | `Initiated → Processing → Completed / Failed` | Modify |
| `deposits.proto` | package `depositsgrpc` | `transactions.proto` package `transactionsgrpc` | Replace |
| `DepositStatus` enum | enum | evolves into Transactions contract | Modify |
| `DepositType` / `DepositProvider` | enums | shared `commongrpc` `PaymentType` / `Provider` | Replace |
| `Money` | duplicate/absent | shared `commongrpc.Money` | New |
| Runtime lifecycle | `deposits/cmd/grpc-service/main.go` | `transactions/cmd/grpc-service/` | Reusable |
| Configuration | `deposits/config` | `transactions/config` | Reusable |
| Docker | `deposits/Dockerfile` | `transactions/Dockerfile` | Reusable |
| Makefile | `deposits/Makefile` | `transactions/Makefile` | Reusable |
| Tests | `deposits/deposits/service_test.go` | `transactions/**` test suites | Reusable |
| Graceful shutdown | `deposits` | `transactions` | Reusable |
| Health checks | `deposits` gRPC + `/healthz` | `transactions` | Reusable |

## 13. Data Migration Considerations

The existing `deposits` database would need data migration if it is reused for
the Transactions Service (Agent 02 will design the schema):

| Table | Column | Consideration |
| --- | --- | --- |
| `clients` | all | Belongs to Clients Service; depends on how Clients DB is provisioned |
| `deposits` | `status` | Must map existing `PENDING/ACCEPTED/COMPLETED/FAILED/REJECTED` to the Transactions 4-state lifecycle |
| `deposits` | `client_id` | Will change from FK to a referenced identifier validated via gRPC |
| `deposits` | `payer_*`, `provider` | Lower camel mapping; provider values may map 1:1 to shared `Provider` or require a compatibility layer |

**No migration scripts are written in this agent.** Agent 02 owns the
Transactions schema; a dedicated data-migration (from legacy deposits to
transactions) is documented as follow-up and depends on whether the legacy
`deposits` DB is preserved, migrated, or rebuilt.

## 14. Protobuf Migration Considerations

Existing `deposits.proto` contracts classification:

| Contract | Classification | Target |
| --- | --- | --- |
| `InitiateDeposit` RPC | Extend / Move | `transactionsgrpc.DepositService.InitiateDeposit` |
| `DepositStatus` enum | Extend | evolves into Transactions lifecycle (add `PROCESSING`, align states) |
| `DepositType` enum | Replace | shared `commongrpc.PaymentType` |
| `DepositProvider` enum | Replace | shared `commongrpc.Provider` |
| `AccountDetails` | Extend / Move | into `transactionsgrpc` |
| `Participant` | Extend / Move | into `transactionsgrpc` |
| `CreateDepositRequest` | Extend / Move | into `transactionsgrpc` (add merchant_id, customer_id) |
| `CreateDepositResponse` | Extend / Move | into `transactionsgrpc` |
| `google.api.http` annotation | Preserve | `POST /v1/public/deposits` |

Agent 04 owns the actual `transactions.proto` + `commongrpc` implementation
per `docs/protobuf-strategy.md`.

## 15. Service Boundary Considerations

Per `docs/domain-model.md`:

| Entity | Service | Notes |
| --- | --- | --- |
| Merchants | Transactions | New |
| Customers | Transactions | New |
| Deposits | Transactions | Evolves from `deposits/` |
| Payouts | Transactions | New |
| Clients | Clients | Moved out of `deposits/` |
| Platforms | Clients | Unrelated to Deposits |
| Integrations | Clients | Unrelated to Deposits |

- The hard-coded client resolution must move to the Clients Service
  (`ClientService.GetClient`).
- Transactions must not own `clients` or `integrations` tables.
- Cross-service access is gRPC-only; no shared database tables.

## 16. Risks

- **Hard-coded client** — replacing `Socadel` with real `ClientService`
  resolution changes deposit behaviour and requires the Clients Service to be
  live.
- **Enum mapping** — mapping legacy `DepositProvider`/`DepositType` to shared
  `Provider`/`PaymentType` may require a compatibility layer.
- **Deposit lifecycle** — aligning legacy `DepositStatus` with the Transactions
  4-state lifecycle may change observable API semantics.
- **Payouts are new** — payout settlement and fee deduction logic has no
  existing reference; fee/wallet entities are unresolved in the domain model.
- **Data migration** — migrating legacy `deposits` client/deposit data depends
  on unverified decisions around the Clients/Transactions database split.
- **Transactionality** — the existing deposit flow has no transaction spanning
  the DB write and the external PawaPay call; moving it must preserve/improve
  consistency.

## 17. Questions Requiring Architectural Decision

- **Fee entity** — the payout flow deducts applicable fees, but no fee entity
  is defined. Must be resolved before Payout implementation (Agent 09).
- **Wallet entity** — the administrator-controlled merchant wallet is
  referenced but not modelled. Balance tracking may be required before
  settlement.
- **Enum mapping** — do `DepositProvider` values map 1:1 to shared `Provider`
  values, or is a compatibility mapping layer required?
- **Client ownership** — is the legacy `clients` table migrated to the Clients
  Service, or is a new Clients database created and the legacy table retired?
- **Deposit lifecycle** — does the Transactions deposit status use
  `Initiated/Processing/Completed/Failed` (domain model) or preserve the legacy
  `PENDING/ACCEPTED/COMPLETED/FAILED/REJECTED` semantics?
- **Pagination** — do `List*` RPCs require pagination in the initial
  `transactions.proto` contract?
- **Cross-service client validation** — is `ClientService.GetClient` callable
  synchronously, or is a cache/local copy needed for deposit processing?
- **Integration ↔ transaction linkage** — must Deposits/Payouts reference an
  Integration ID to attribute transactions to a platform?
- **Customer uniqueness** — can a Customer transact across multiple Clients or
  Merchants over time?

## 18. Recommended Implementation Sequence

The following dependency-aware sequence for Agents 02–13 is proposed:

1. **Agent 02 (Database)** — Design `transactions/db/migrations` + `query` +
   sqlc for merchants, customers, deposits, payouts. Preserve the deposits
   schema conventions. Deprecate the legacy `clients` table.
2. **Agent 03 (SQLC)** — Configure `transactions/db/sqlc.yaml`, generate
   models/queries. Do not hand-edit generated output.
3. **Agent 04 (Protobuf)** — Create `protobuf/common.proto` (shared `Money`,
   `Provider`, `PaymentType`, `UserRole`) then `protobuf/transactions.proto`
   (package `transactionsgrpc`). Regenerate `grpc/go/transactionsgrpc/`.
4. **Agent 05 (Repositories)** — Implement Merchant, Customer, Deposit, and
   Payout repositories on the generated querier using the
   `Do()`/`Begin()`/Migrate pattern.
5. **Agent 06 (Merchants)** — Merchant service implementation.
6. **Agent 07 (Customers)** — Customer service implementation.
7. **Agent 08 (Deposits)** — Migrate the deposit flow from `deposits/`,
   replacing the hard-coded client with `ClientService.GetClient` validation.
8. **Agent 09 (Payouts)** — Implement payout flow (requires fee/wallet
   decision from open questions).
9. **Agent 10 (Runtime)** — `transactions/cmd/grpc-service/main.go` +
   `transactions/config/` wiring all services, gRPC, gateway, health,
   graceful shutdown.
10. **Agent 11 (Scaffolding)** — `transactions/README.md`, `Dockerfile`,
    `Makefile`, `.env.example`, developer-experience review.
11. **Agent 12 (Tests)** — Unit tests for all domains following deposits
    conventions; race detection; full repository suite; test review doc.
12. **Agent 13 (Production Review)** — Final production-readiness audit and
    `docs`-level report.