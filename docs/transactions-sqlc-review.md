# Transactions SQLC Implementation Review

Document Version: 1.0
Status: Complete
System: RVPay
Service: Transactions Service
Review: Agent 03 — Transactions SQLC Generation and Verification

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
- existing `deposits/db/` SQLC implementation (sqlc.yaml, query/, sqlc/, doc.go)

## 2. SQLC Version

SQLC **v1.29.0** is used, pinned by the repository convention in
`deposits/db/doc.go`:

```go
//go:generate go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate
```

The Transactions `transactions/db/doc.go` uses the identical pinned version.
No SQLC upgrade was performed. Generation was executed with
`go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate` from
`transactions/db/`.

## 3. SQLC Configuration

File: `transactions/db/sqlc.yaml`

The configuration is an exact copy of the Deposits convention:

```yaml
version: "2"
sql:
- schema: "./migrations"
  queries: "./query"
  engine: "postgresql"
  gen:
    go:
      package: "sqlc"
      out: "./sqlc"
      sql_package: "pgx/v5"
      emit_json_tags: true
      emit_interface: true
      emit_empty_slices: true
      emit_exact_table_names: false
      emit_pointers_for_null_types: true
      overrides:
        - db_type: "date"
          nullable: true
          go_type: "time.Time"
        - db_type: "timestamptz"
          nullable: true
          go_type: "github.com/jackc/pgx/v5/pgtype.Timestamptz"
        - db_type: "timestamptz"
          nullable: false
          go_type: "time.Time"
        - db_type: "uuid"
          nullable: true
          go_type: "github.com/google/uuid.UUID"
        - db_type: "uuid"
          nullable: false
          go_type: "github.com/google/uuid.UUID"
        - db_type: "text"
          nullable: true
          go_type: "string"
        - db_type: "text"
          go_type: "string"
```

Key settings:

- `engine: postgresql`, `sql_package: pgx/v5`.
- Emits JSON tags, an interface (`Querier`), empty slices, and pointers for
  null types.
- Nullable `timestamptz` maps to `pgtype.Timestamptz`; non-nullable maps to
  `time.Time`.
- UUID maps to `github.com/google/uuid.UUID`.
- Text maps to `string` for both nullable and non-nullable.

## 4. Query Organization

Query files live in `transactions/db/query/`, grouped by entity:

| File | Entity | Operations |
| --- | --- | --- |
| `merchants.sql` | Merchants | Create, GetByID, GetBySlug, List, UpdateStatus |
| `customers.sql` | Customers | Create, GetByID, GetByClientAndMerchantAndPhone, ListByClient, ListByMerchant, UpdateStatus |
| `deposits.sql` | Deposits | Create, GetByID, GetByExternalReference, GetByIdempotencyKey, ListByClient, ListByCustomer, ListByMerchant, ListByStatus, UpdateStatus, UpdateStatusAndCompletedAt, UpdateStatusAndFailedAt |
| `payouts.sql` | Payouts | Create, GetByID, GetByExternalReference, GetByIdempotencyKey, ListByClient, ListByMerchant, ListByStatus, UpdateStatus, UpdateStatusAndCompletedAt, UpdateStatusAndFailedAt |

Query naming follows the project convention: `Create<Entity>`,
`Get<Entity>By<Field>`, `List<Entity>By<Scope>`, `Update<Entity>Status`.

## 5. Generated Output

Generated output is written to `transactions/db/sqlc/`:

- `db.go` — `DBTX` interface and `New()`.
- `models.go` — generated model structs and enum types.
- `querier.go` — generated `Querier` interface (32 methods).
- `merchants.sql.go`, `customers.sql.go`, `deposits.sql.go`,
  `payouts.sql.go` — generated query implementations.

Generated package name: `sqlc`. Generated code is never hand-edited.

## 6. Merchant Queries

| Method | Purpose |
| --- | --- |
| `CreateMerchant` | Insert a merchant with name, slug, status |
| `GetMerchantByID` | Fetch by primary key |
| `GetMerchantBySlug` | Fetch by unique slug |
| `ListMerchants` | List all merchants ordered by creation |
| `UpdateMerchantStatus` | Update status scoped to a specific merchant ID |

## 7. Customer Queries

| Method | Purpose |
| --- | --- |
| `CreateCustomer` | Insert a customer with client, merchant, phone, status |
| `GetCustomerByID` | Fetch by primary key |
| `GetCustomerByClientAndMerchantAndPhone` | Fetch by the unique (client, merchant, phone) constraint |
| `ListCustomersByClient` | List customers scoped to a client |
| `ListCustomersByMerchant` | List customers scoped to a merchant |
| `UpdateCustomerStatus` | Update status scoped to a specific customer ID |

Merchant scoping is respected; no query crosses merchant boundaries.

## 8. Transaction Queries

The Transactions architecture does not define a separate generic
"transactions" table. Deposits and payouts are the transaction records
(per `docs/transactions-database-review.md`). No speculative generic
"transaction" queries were created.

## 9. Deposit Queries

| Method | Purpose |
| --- | --- |
| `CreateDeposit` | Insert a deposit with client, customer, merchant, amount, currency, payment type, provider, status, idempotency key |
| `GetDepositByID` | Fetch by primary key |
| `GetDepositByExternalReference` | Fetch by provider external reference |
| `GetDepositByIdempotencyKey` | Fetch by unique idempotency key |
| `ListDepositsByClient` | List deposits scoped to a client |
| `ListDepositsByCustomer` | List deposits scoped to a customer |
| `ListDepositsByMerchant` | List deposits scoped to a merchant |
| `ListDepositsByStatus` | List deposits by status |
| `UpdateDepositStatus` | Update status scoped to a specific deposit ID |
| `UpdateDepositStatusAndCompletedAt` | Update status + completed_at |
| `UpdateDepositStatusAndFailedAt` | Update status + failed_at + failure_reason |

Status updates target a specific record via `WHERE id = $1`. Lifecycle
fields (`completed_at`, `failed_at`) are set by the specific transition
queries, not by generic broad updates.

The legacy `deposits/` SQL was not copied blindly; the new schema fields
(`client_id`, `customer_id`, `merchant_id`, `idempotency_key`,
`payment_type`) replace the obsolete legacy fields (`payer_type`,
`payer_provider`).

## 10. Payout Queries

| Method | Purpose |
| --- | --- |
| `CreatePayout` | Insert a payout with client, merchant, amount, currency, provider, destination, status, idempotency key |
| `GetPayoutByID` | Fetch by primary key |
| `GetPayoutByExternalReference` | Fetch by provider external reference |
| `GetPayoutByIdempotencyKey` | Fetch by unique idempotency key |
| `ListPayoutsByClient` | List payouts scoped to a client |
| `ListPayoutsByMerchant` | List payouts scoped to a merchant |
| `ListPayoutsByStatus` | List payouts by status |
| `UpdatePayoutStatus` | Update status scoped to a specific payout ID |
| `UpdatePayoutStatusAndCompletedAt` | Update status + completed_at |
| `UpdatePayoutStatusAndFailedAt` | Update status + failed_at + failure_reason |

No payout business rules are implemented in SQL.

## 11. Idempotency

Idempotency is represented by the `idempotency_key UUID NOT NULL UNIQUE`
column on `deposits` and `payouts` (from Agent 02).

Queries:

- `GetDepositByIdempotencyKey`
- `GetPayoutByIdempotencyKey`

These expose the database lookup operation needed by the repository/service
layer. No application-level idempotency logic is implemented in this agent.

## 12. Pagination

Pagination was **not implemented**. Per `docs/domain-model.md`,
`docs/repository-layout.md`, and the existing Deposits conventions, no
pagination requirement is documented for the initial Transactions queries.
List queries use deterministic `ORDER BY created_at [DESC]` ordering. If
pagination is required later, it will be added as an explicit limit/offset
pattern.

## 13. Transaction Support

The generated SQLC code supports the project's existing transaction pattern
(the same as Deposits):

- `db.go` generates `DBTX` (interface compatible with `pgxpool.Pool` and
  `pgx.Tx`).
- `New(db DBTX) *Queries` accepts a transaction handle.
- The repository layer (Agent 05) can use `Do()`/`Begin()` exactly as
  `deposits/db/repo/repo.go` does.

No business transactions are created in this agent.

## 14. Validation

| Command | Result |
| --- | --- |
| `go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate` | ✅ Exit 0 |
| `go build ./transactions/db/...` | ✅ Exit 0 |
| `go build ./transactions/...` | ✅ Exit 0 |
| `go vet ./transactions/db/...` | ✅ Exit 0 |
| Regeneration (2nd run) | ✅ Deterministic — no diff |
| Machine-specific output scan | ✅ No `/home/`, `workdir`, or `timestamp` in generated files |

The generated package compiles, imports resolve, types compile, and SQLC
output is valid Go.

## 15. Risks

- **Pagination deferred** — List queries have no limit/offset; large result
  sets are unbounded. Must be addressed if production volumes require it.
- **Null external references** — `external_reference`, `completed_at`,
  `failed_at`, `failure_reason` are nullable; generated types use
  `pgtype.Timestamptz`/`string` appropriately per the overrides. Repository
  code must handle nulls.
- **Enum alignment** — the new `payment_provider`/`payment_type` enums
  replace the legacy `payer_provider`/`payer_type`; any data migration must
  map values.
- **Idempotency key type** — `idempotency_key` is a UUID; if a string key is
  required by callers, a schema + query change would be needed.

## 16. Unresolved Questions

- **Fee entity** — no payout fee entity exists; payout queries may need a
  companion `fees` table and query file when settlement logic is designed
  (Agent 09).
- **Pagination** — whether `List*` queries require explicit limit/offset in
  the initial contract is unresolved; currently deferred.
- **Idempotency key semantics** — whether `idempotency_key` is
  client-supplied (string) or server-generated (UUID) is not fully specified;
  the UUID column supports server-generated keys.
- **Transaction reconciliation** — no query exists to fetch deposits/payouts
  by time range (e.g. `created_after`/`created_before`); required if a
  reconciliation job is introduced.