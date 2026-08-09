# Transactions Repository Implementation Review

Document Version: 1.0
Status: Complete
System: RVPay
Service: Transactions Service
Review: Agent 05 — Transactions Repository Layer

## 1. Source Documents

The following documents were read and used as the source of truth:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md
- docs/transactions-existing-review.md
- docs/transactions-database-review.md
- docs/transactions-sqlc-review.md
- docs/transactions-protobuf-review.md
- existing `deposits/db/` implementation (repo, query, sqlc, doc.go)

## 2. Existing Deposits Repository

The existing `deposits/db/repo/repo.go` defines the canonical repository
conventions:

- `TransactionsRepo`/`DepositsRepo` interface with `Begin(ctx)` (returns
  `sqlc.Querier` + `pgx.Tx`) and `Do()` (returns `sqlc.Querier`).
- `Impl` struct wrapping `*pgxpool.Pool`; `NewTransactionsRepo(db)` /
  `NewDepositsRepo(db)` constructor.
- `Migrate(dbURL, migrationPath, logger)` running golang-migrate at startup,
  and `MigrateDown` for rollback.
- Error conventions: `ErrNotFound`, `ErrDuplicate`, `ErrConstraint` sentinels;
  `wrapError` (pgconn code mapping) and `wrapNotFound` (pgx.ErrNoRows
  mapping) helpers.
- Explicit dependency injection; no global database state.

## 3. Repository Structure

Files created under `transactions/db/repo/`:

| File | Purpose |
| --- | --- |
| `repo.go` | `TransactionsRepo` interface, `Impl`, `NewTransactionsRepo`, `Migrate`, `MigrateDown` |
| `errors.go` | `ErrNotFound`, `ErrDuplicate`, `ErrConstraint` + `wrapError`/`wrapNotFound` |
| `merchant_repo.go` | `MerchantRepo` interface + implementation |
| `customer_repo.go` | `CustomerRepo` interface + implementation |
| `deposit_repo.go` | `DepositRepo` interface + implementation |
| `payout_repo.go` | `PayoutRepo` interface + implementation |

Generated mocks:

- `transactions/db/sqlc/mocks/querier.go` — mock of the generated `sqlc.Querier`.
- `transactions/db/repo/mocks/repo.go` — mocks of `TransactionsRepo`,
  `MerchantRepo`, `CustomerRepo`, `DepositRepo`, `PayoutRepo`.

## 4. Repository Interfaces

| Interface | Responsibility |
| --- | --- |
| `TransactionsRepo` | Pool access: allow the service layer to run queries directly or inside a database transaction |
| `MerchantRepo` | Merchant persistence (create, get by id/slug, list, update status) |
| `CustomerRepo` | Customer persistence (create, get by id, get by client+merchant+phone, list by client/merchant, update status) |
| `DepositRepo` | Deposit persistence (create, get by id/external ref/idempotency key, list by client/customer/merchant/status, status/lifecycle updates) |
| `PayoutRepo` | Payout persistence (create, get by id/external ref/idempotency key, list by client/merchant/status, status/lifecycle updates) |

Each interface expresses only what the Transactions service needs; the full
`sqlc.Querier` is not blindly re-exposed (`TransactionsRepo.Do()`/`Begin()`
provide the pool boundary, while domain repositories provide typed domain
operations).

## 5. Merchant Persistence

Implemented operations (backed by `sqlc.Querier`):

| Repository Method | SQLC Method |
| --- | --- |
| `Create` | `CreateMerchant` |
| `GetByID` | `GetMerchantByID` |
| `GetBySlug` | `GetMerchantBySlug` |
| `List` | `ListMerchants` |
| `UpdateStatus` | `UpdateMerchantStatus` |

## 6. Customer Persistence

Implemented operations:

| Repository Method | SQLC Method |
| --- | --- |
| `Create` | `CreateCustomer` |
| `GetByID` | `GetCustomerByID` |
| `GetByClientAndMerchantAndPhone` | `GetCustomerByClientAndMerchantAndPhone` |
| `ListByClient` | `ListCustomersByClient` |
| `ListByMerchant` | `ListCustomersByMerchant` |
| `UpdateStatus` | `UpdateCustomerStatus` |

Merchant ownership is respected: `GetByClientAndMerchantAndPhone` requires the
merchant ID, and `ListByMerchant` scopes to a single merchant. No query can
cross merchant boundaries.

## 7. Transaction Persistence

As established by the database and protobuf reviews, there is no separate
generic "transactions" table. Deposits and payouts are the transaction
records. Persistence operations for both are exposed via `DepositRepo` and
`PayoutRepo` respectively.

## 8. Deposit Persistence

Implemented operations:

| Repository Method | SQLC Method |
| --- | --- |
| `Create` | `CreateDeposit` |
| `GetByID` | `GetDepositByID` |
| `GetByExternalReference` | `GetDepositByExternalReference` |
| `GetByIdempotencyKey` | `GetDepositByIdempotencyKey` |
| `ListByClient` / `ListByCustomer` / `ListByMerchant` / `ListByStatus` | corresponding `ListDepositsBy*` |
| `UpdateStatus` | `UpdateDepositStatus` |
| `MarkCompleted` | `UpdateDepositStatusAndCompletedAt` |
| `MarkFailed` | `UpdateDepositStatusAndFailedAt` |

The lifecycle update methods map to the domain lifecycle: `UpdateStatus` for
generic state changes (e.g. initiated → processing), `MarkCompleted` sets
`completed_at`, `MarkFailed` sets `failed_at` + `failure_reason`.

## 9. Payout Persistence

Implemented operations:

| Repository Method | SQLC Method |
| --- | --- |
| `Create` | `CreatePayout` |
| `GetByID` | `GetPayoutByID` |
| `GetByExternalReference` | `GetPayoutByExternalReference` |
| `GetByIdempotencyKey` | `GetPayoutByIdempotencyKey` |
| `ListByClient` / `ListByMerchant` / `ListByStatus` | corresponding `ListPayoutsBy*` |
| `UpdateStatus` | `UpdatePayoutStatus` |
| `MarkCompleted` | `UpdatePayoutStatusAndCompletedAt` |
| `MarkFailed` | `UpdatePayoutStatusAndFailedAt` |

No payout business rules (balances, fees, eligibility) are implemented; the
repository only persists the requested state.

## 10. Transactions

`TransactionsRepo` mirrors the Deposits `Begin()`/`Do()` pattern:

- `Begin(ctx)` — starts a `pgx.Tx`, returns a transaction-bound `sqlc.Querier`
  and the raw `pgx.Tx` for commit/rollback by the caller.
- `Do()` — returns the pool-backed `sqlc.Querier`.

The SQLC generated `DBTX` interface accepts both `*pgxpool.Pool` and
`pgx.Tx`, so transaction-bound queries are created via `sqlc.New(tx)`. No
domain repository operation currently requires its own multi-record
transaction per the database review; the service layer can combine domain
operations inside a `Begin()`/`Commit()` block where atomicity is required.
Rollback is the caller's responsibility and no transaction is left open by the
repository.

## 11. Error Handling

- `wrapError` maps PostgreSQL constraint errors:
  - `23505` (unique_violation) → `ErrDuplicate`
  - `23503` (foreign_key_violation) / `23514` (check_violation) →
    `ErrConstraint`
  - otherwise returns the raw error.
- `wrapNotFound` maps `pgx.ErrNoRows` → `ErrNotFound`.
- `ErrNotFound`, `ErrDuplicate`, `ErrConstraint` are package-level sentinel
  errors that the service layer can inspect with `errors.Is`.

## 12. SQLC Integration

The repository consumes the generated `transactions/db/sqlc` package (models,
params structs, `Querier` interface, `DBTX`). No SQL is written in the
repository Go code. Raw SQL is not introduced; all persistence goes through
SQLC-generated query methods.

## 13. Mocks

Mocks are generated with the project's pinned `mockgen v0.6.0` via
`go:generate` directives in `transactions/db/doc.go`:

- `sqlc/mocks/querier.go` — mock of `sqlc.Querier` (for SQLC-level tests).
- `repo/mocks/repo.go` — mocks of `TransactionsRepo`, `MerchantRepo`,
  `CustomerRepo`, `DepositRepo`, `PayoutRepo` (for service-layer tests).

Generation is reproducible and uses the same mechanism as Deposits;
mockgen version was not changed.

## 14. Validation

| Command | Result |
| --- | --- |
| `go build ./transactions/...` | ✅ Exit 0 |
| `go vet ./transactions/...` | ✅ Exit 0 |
| `go test ./transactions/...` | ✅ All packages pass (no test files yet) |
| `go test ./...` | ✅ Full repository — no failures |
| `go generate ./transactions/db` | ✅ Mocks generated (v0.6.0) |

The repository package, generated SQLC package, and generated mocks all
compile. No unrelated services were modified.

## 15. Files Changed

Created:

- `transactions/db/repo/repo.go`
- `transactions/db/repo/errors.go`
- `transactions/db/repo/merchant_repo.go`
- `transactions/db/repo/customer_repo.go`
- `transactions/db/repo/deposit_repo.go`
- `transactions/db/repo/payout_repo.go`
- `transactions/db/repo/mocks/repo.go` (generated)
- `transactions/db/sqlc/mocks/querier.go` (generated)
- `docs/transactions-repository-review.md`

Modified:

- `transactions/db/doc.go` — added sqlc + mockgen `go:generate` directives.

No Deposits, Clients, protobuf, SQL, or SQLC generated files were modified.

## 16. Risks

- **Amount as `pgtype.Numeric`** — `Create`/`CreatePayout` accept
  `pgtype.Numeric` directly; the service layer must construct valid
  `pgtype.Numeric` values from protobuf `Money` (decimal string). Invalid
  numeric input may produce a `CHECK`/parse error surfaced as a raw error.
- **Transactional atomicity** — domain repositories expose single-operation
  methods only; the service layer must call `TransactionsRepo.Begin()` to
  combine operations atomically (e.g. idempotency-check then create). Each
  repository method is itself a single atomic statement.
- **Optional fields** — `external_reference`, `completed_at`, `failed_at`,
  `failure_reason`, `destination_reference` are nullable; the generated SQLC
  types use `pgtype.Timestamptz`/`string`/pgtype types. The service layer must
  not blindly dereference them (the generated `pgtype` types expose
  `.Valid`/`pgtype.Numeric` wrappers).
- **No list pagination** — `List*By*` methods return full result sets with
  deterministic ordering. Unbounded result growth is possible without
  pagination (deferred at the SQL layer; a follow-up is required if production
  volumes demand it).
- **Mock maintenance** — any new interface method requires regenerating mocks
  via `go generate`, otherwise the mocks will not satisfy the interfaces.

## 17. Unresolved Questions

- **Pagination** — list operations have no limit/offset; whether the service
  layer should add pagination parameters is unresolved (SQLC layer has no
  pagination queries).
- **Fee/wallet** — no fee or wallet persistence exists; Payout settlement
  logic (Agent 09) may require an additional fee table and repository.
- **Idempotency key exposure** — `GetByIdempotencyKey` exists, but the
  protobuf `CreateDepositRequest`/`CreatePayoutRequest` do not carry an
  idempotency key field; whether the service layer generates the key
  server-side or accepts it from clients is unresolved.
- **Client validation** — `client_id` is persisted as a plain UUID without a
  local FK; the service layer (runtime agent) must validate `client_id`
  against the Clients Service via gRPC before persisting.