# Transactions Database Implementation Review

Document Version: 1.0
Status: Complete
System: RVPay
Service: Transactions Service
Review: Agent 02 — Transactions Database Implementation

## 1. Source Documents

The following documents were read and used as the source of truth:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md
- docs/transactions-existing-review.md
- existing `deposits/db/` implementation (migrations, query, repo, sqlc)

## 2. Existing Deposits Database

The existing `deposits/db/migrations/000001_init_schema.{up,down}.sql` defines:

- `clients` — `id UUID PK`, `client_name`, `email UNIQUE`, `phone_number
  UNIQUE`, `created_at`, `updated_at`.
- `deposits` — `id UUID PK`, `amount NUMERIC(18,2)`, `currency VARCHAR(3)`
  with `CHECK (currency ~ '^[A-Z]{3}$')`, `payer_type` enum, `payer_phone_number`,
  `payer_provider` enum, `client_id FK`, `created_at`, `updated_at`.
- Enums: `payment_provider` (`MTN_MOMO_CMR`, `ORANGE_CMR`), `payer_type`
  (`MMO`).

Conventions observed:

- `gen_random_uuid()` for primary keys (requires `pgcrypto` extension).
- `NUMERIC(18,2)` for monetary values (no floating point).
- `VARCHAR(3)` currency with uppercase ISO-4217 regex check.
- `TIMESTAMPTZ NOT NULL DEFAULT NOW()` for `created_at`/`updated_at`.
- `ON DELETE RESTRICT` for transaction-history foreign keys.
- golang-migrate naming: `000001_init_schema.{up,down}.sql`.

## 3. Target Transactions Schema

Per `docs/domain-model.md`, the Transactions Service owns four entities:
**Merchants**, **Customers**, **Deposits**, and **Payouts**. There is no
separate generic "transactions" table — deposits and payouts ARE the
transaction records.

Per the domain model assumption: "Deposits and Payouts reference Clients and
Merchants by identifier resolved through gRPC; foreign keys exist only within
the owning service's database." Therefore `client_id` is stored as a plain
`UUID` (no local FK to a clients table, which belongs to the Clients Service).

The implemented schema creates:

- `merchants` — payment gateways (e.g. PawaPay, Flutterwave).
- `customers` — end users making payments, linked to a client and merchant.
- `deposits` — inbound customer payments.
- `payouts` — outbound settlements.

## 4. Entity Mapping

| Existing Deposits Concept | Transactions Concept | Action |
| --- | --- | --- |
| `clients` table | Clients Service (not Transactions) | Deprecated (moves to Clients Service) |
| `deposits` table | `deposits` table | Reuse / Extend |
| `deposits.amount` | `deposits.amount NUMERIC(18,2)` | Reuse |
| `deposits.currency` | `deposits.currency VARCHAR(3)` + CHECK | Reuse |
| `deposits.payer_type` | `payment_type` enum (`MMO`, `CREDIT_CARD`) | Extend |
| `deposits.payer_provider` | `payment_provider` enum (`MTN_MOMO`, `ORANGE_MOMO`) | Replace |
| `deposits.payer_phone_number` | `deposits.payer_phone_number` | Reuse |
| `deposits.client_id` | `deposits.client_id UUID` (no local FK) | Extend |
| — | `merchants` table | New |
| — | `customers` table | New |
| — | `payouts` table | New |
| — | `deposit_status` enum (`INITIATED/PROCESSING/COMPLETED/FAILED`) | New (4-state per domain model) |
| — | `payout_status` enum (`REQUESTED/PROCESSING/COMPLETED/FAILED`) | New |
| — | `merchant_status` enum (`ONBOARDED/ACTIVE/SUSPENDED/RETIRED`) | New |
| — | `customer_status` enum (`CREATED/ACTIVE`) | New |
| — | `idempotency_key` on deposits/payouts | New |

## 5. Tables Created

| Table | Purpose |
| --- | --- |
| `merchants` | Payment gateway records (name, slug, status, timestamps) |
| `customers` | End-user payment records (client_id, merchant_id, phone_number, status) |
| `deposits` | Inbound payments (client, customer, merchant, amount, currency, payment_type, provider, status, idempotency, lifecycle timestamps) |
| `payouts` | Outbound settlements (client, merchant, amount, currency, provider, destination, status, idempotency, lifecycle timestamps) |

## 6. Relationships

Foreign keys (all `ON DELETE RESTRICT ON UPDATE RESTRICT` to preserve
transaction history):

- `customers.merchant_id` → `merchants.id`
- `deposits.customer_id` → `customers.id`
- `deposits.merchant_id` → `merchants.id`
- `payouts.merchant_id` → `merchants.id`

`client_id` on customers, deposits, and payouts is a plain `UUID` (no local
FK) — it references a Clients Service record resolved via gRPC, per the domain
model. No cascading deletion is used; transaction history is preserved.

## 7. Constraints

- `merchants.slug` — `UNIQUE`.
- `customers` — `UNIQUE (client_id, merchant_id, phone_number)` (a customer is
  unique per client+merchant+phone).
- `deposits.idempotency_key` — `UNIQUE` (idempotency).
- `payouts.idempotency_key` — `UNIQUE` (idempotency).
- `deposits.currency` / `payouts.currency` — `CHECK (currency ~ '^[A-Z]{3}$')`.
- All `id` columns — `UUID PRIMARY KEY DEFAULT gen_random_uuid()`.
- All `created_at`/`updated_at` — `TIMESTAMPTZ NOT NULL DEFAULT NOW()`.
- Status columns — typed enums with documented defaults
  (`ONBOARDED`, `CREATED`, `INITIATED`, `REQUESTED`).

## 8. Indexes

| Index | Table | Query Pattern |
| --- | --- | --- |
| `idx_customers_client_id` | customers | Lookup customers by client |
| `idx_customers_merchant_id` | customers | Lookup customers by merchant |
| `idx_deposits_client_id` | deposits | Lookup deposits by client |
| `idx_deposits_customer_id` | deposits | Lookup deposits by customer |
| `idx_deposits_merchant_id` | deposits | Lookup deposits by merchant |
| `idx_deposits_status` | deposits | Status-based queries |
| `idx_deposits_external_reference` | deposits | Provider reference lookup |
| `idx_deposits_created_at` | deposits | Time-based transaction queries |
| `idx_payouts_client_id` | payouts | Lookup payouts by client |
| `idx_payouts_merchant_id` | payouts | Lookup payouts by merchant |
| `idx_payouts_status` | payouts | Status-based queries |
| `idx_payouts_external_reference` | payouts | Provider reference lookup |
| `idx_payouts_created_at` | payouts | Time-based transaction queries |

Every index supports an identifiable query pattern. No index is created on
every column.

## 9. Migrations

| File | Purpose |
| --- | --- |
| `transactions/db/migrations/000001_init_schema.up.sql` | Creates enums, merchants, customers, deposits, payouts, indexes, constraints |
| `transactions/db/migrations/000001_init_schema.down.sql` | Drops payouts, deposits, customers, merchants, then all enums in reverse dependency order |

Migration ordering follows domain dependencies: foundational enums → merchants
→ customers → deposits → payouts → indexes/constraints.

## 10. Data Migration Considerations

The existing `deposits` database is **not** migrated in this agent. Per
`docs/migration-plan.md`, the legacy `deposits/` service remains runnable until
the Transactions Service is complete. A future data migration (from legacy
`deposits` to `transactions`) is required and depends on:

- Whether the legacy `clients` table is migrated to the Clients Service or
  retired.
- Mapping legacy `deposits.status` (`PENDING/ACCEPTED/COMPLETED/FAILED/REJECTED`)
  to the Transactions 4-state lifecycle (`INITIATED/PROCESSING/COMPLETED/FAILED`).
- Mapping legacy `payment_provider` (`MTN_MOMO_CMR`, `ORANGE_CMR`) to the
  Transactions `payment_provider` (`MTN_MOMO`, `ORANGE_MOMO`).
- Whether legacy `deposits.client_id` (a local FK) maps to the Transactions
  `client_id` (a gRPC-resolved UUID).

This data migration is documented as follow-up work and is outside this
agent's scope.

## 11. Risks

- **Client ownership** — `client_id` is stored without a local FK; orphaned
  client references are possible if the Clients Service deletes a client.
  Mitigated by `ON DELETE RESTRICT` semantics being enforced at the service
  layer via gRPC validation.
- **Enum mapping** — legacy `payment_provider` values differ from the new
  `payment_provider` values; a compatibility mapping is required for any data
  migration.
- **Deposit lifecycle** — the 4-state lifecycle differs from the legacy
  5-state enum; API semantics may change.
- **Payouts are new** — no existing reference; fee/wallet entities are
  unresolved in the domain model and may require schema changes later.
- **Idempotency** — `idempotency_key` is a `UUID`; if callers need string
  idempotency keys, a follow-up migration would be required.

## 12. Unresolved Questions

- **Fee entity** — the payout flow deducts fees, but no fee entity is defined;
  must be resolved before Payout implementation (Agent 09).
- **Wallet entity** — the administrator-controlled merchant wallet is not
  modelled; balance tracking may require a dedicated table.
- **Customer uniqueness** — whether a Customer may transact across multiple
  Clients or Merchants over time is not defined; the current unique constraint
  scopes a customer to one client+merchant+phone.
- **Idempotency key type** — whether `idempotency_key` should be a string
  (client-supplied) or a UUID (server-generated) is not fully specified.
- **Integration ↔ transaction linkage** — whether Deposits/Payouts must
  reference an Integration ID to attribute transactions to a platform is not
  defined.