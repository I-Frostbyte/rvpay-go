# Transactions Payouts Implementation Review

Document Version: 1.0
Status: Complete
System: RVPay
Service: Transactions Service
Review: Agent 09 — Transactions Payouts Service

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
- docs/transactions-repository-review.md
- docs/transactions-merchants-review.md
- docs/transactions-customers-review.md
- docs/transactions-deposits-review.md
- the newly implemented `transactions/merchants/`, `transactions/customers/`,
  and `transactions/deposits/` services
- the generated `transactionsgrpc` and `commongrpc` protobuf packages

## 2. Existing Payout Implementation

**No existing payout implementation exists in the repository.** The legacy
`deposits/` service only handled inbound deposits. Per `docs/migration-plan.md`
and `docs/domain-model.md`, Payouts are a newly introduced Transactions
capability; no legacy behavior was inventoried, retained, or copied. Per
directive #56, no existing payout code was deleted or migrated because none
exists.

## 3. Domain Model

Per `docs/domain-model.md`, a Payout is an outbound settlement: funds transfer
from the administrator-controlled merchant wallet to a Client after fees are
deducted. The implemented `Payout` resource exposes:

- `id` — stable internal identifier (UUID).
- `client_id`, `merchant_id` — ownership references.
- `amount` (`commongrpc.Money`), `currency` — monetary value.
- `provider` — MTN_MOMO or ORANGE_MOMO.
- `destination_reference` — the payout destination account/wallet reference.
- `status` — REQUESTED / PROCESSING / COMPLETED / FAILED.
- `external_reference` — provider reference (nullable).
- `requested_at` / `completed_at` / `failed_at` / `failure_reason` — lifecycle
  fields.
- `created_at` / `updated_at` — audit timestamps.

No invented fields. Per the domain model, Payouts are associated with **Clients
and Merchants only** (not Customers); no customer requirement was introduced.

## 4. Merchant Relationship

Every payout belongs to exactly one merchant. Tenant isolation is enforced:

- **Repository** (`PayoutRepo.Create`) requires `merchant_id`.
- **Service** validates `merchant_id` as a UUID before any repository call;
  `PayoutRepo.GetByID` returns the payout with its `merchant_id`, and
  `ListByMerchant` scopes to a single merchant (available for future
  merchant-scoped listing).
- **Database** has a `FOREIGN KEY (merchant_id) → merchants(id)` with
  `ON DELETE RESTRICT`.

A payout for Merchant A is never retrievable through a request scoped to
Merchant B without an explicit `merchant_id`.

## 5. Customer Relationship

**Payouts are not customer-scoped** per `docs/domain-model.md`. A Payout
references `client_id` and `merchant_id` only. Therefore no customer
validation, customer repository dependency, or customer association logic was
introduced (consistent with directive #9).

## 6. Destination

The payout destination is represented by the `destination_reference` string on
the `Payout` resource (and the `deposits`/`payouts` schema's
`destination_reference` column, nullable). The `CreatePayoutRequest` requires
a non-empty `destination_reference`. No invention of destination types (bank
account/wallet enum) was made — the protobuf contract defines only the
reference string.

## 7. Payout Lifecycle

The domain-model lifecycle `Requested → Processing → Completed / Failed` is
supported:

| Current State | Allowed State | Rule |
| --- | --- | --- |
| (none) | REQUESTED | New payouts are created in REQUESTED |
| REQUESTED | PROCESSING / FAILED | Interim/terminal transitions (to be invoked by future provider/status agents) |
| PROCESSING | COMPLETED / FAILED | Terminal transitions (invoked by future provider agents) |
| COMPLETED | (terminal) | No further transition |
| FAILED | (terminal) | No further transition |

The repository exposes the transition methods (`UpdateStatus`,
`MarkCompleted`, `MarkFailed`). This agent does not expose arbitrary
`UpdatePayoutStatus` RPCs — the protobuf contract defines only `RequestPayout`
and `GetPayout`; lifecycle mutations are reserved for later provider/status
processing agents. The database enforces structural validity; the service layer
will enforce transition rules once a status-mutation operation is introduced.

## 8. Service Structure

Package: `transactions/payouts/`

| File | Purpose |
| --- | --- |
| `service.go` | `Impl` struct, `NewPayoutService`, `RequestPayout`, `GetPayout`, and validation/converter helpers |
| `converters.go` | `payoutToProto`, `sqlcPayoutStatusToGrpc`, `sqlcPaymentProviderToGrpc`, and amount formatting |
| `service_test.go` | Unit tests using generated mocks |

The `Impl` embeds `transactionsgrpc.UnimplementedPayoutServiceServer` (by
value), holds injected `repo.PayoutRepo` and `zerolog.Logger`, and implements
the generated `PayoutServiceServer` interface. It mirrors the
Merchant/Customer/Deposit service conventions (exact same structure and
pattern).

## 9. Repository Integration

| Service Operation | Repository Operation |
| --- | --- |
| `RequestPayout` (persist) | `PayoutRepo.Create` |
| `GetPayout` | `PayoutRepo.GetByID` |

No SQL is written in the service; no SQLC is called directly. All persistence
goes through the injected repository interface.

## 10. RPC Implementation

| RPC | Purpose | Service Operation |
| --- | --- | --- |
| `RequestPayout` | Request an outbound settlement | validate request → `PayoutRepo.Create` → map to protobuf |
| `GetPayout` | Fetch a payout by id | validate id → `PayoutRepo.GetByID` → map to protobuf |

Both RPCs return protobuf `Payout` resources mapped by `payoutToProto`; SQLC
models are never returned directly.

## 11. REST Exposure

REST routes are generated by grpc-gateway from the protobuf annotations
(Agent 04). No manual HTTP server or duplicate route definitions were added.

| Method | Route | RPC |
| --- | --- | --- |
| POST | `/v1/public/payouts` | RequestPayout |
| GET | `/v1/public/payouts/{payout_id}` | GetPayout |

## 12. Validation

- **Required IDs** — `client_id`, `merchant_id` must be valid UUIDs; otherwise
  `InvalidArgument`.
- **Amount** — `Money` must be present, parseable as a decimal, and greater
  than zero (via `pgtype.Numeric.Scan` + `Float64Value`); otherwise
  `InvalidArgument`. No floating-point arithmetic is used for persisted
  values.
- **Currency** — normalized to uppercase and required non-empty.
- **Provider** — must be a supported enum value via `grpcProviderToSqlc`;
  otherwise `InvalidArgument`.
- **Destination** — `destination_reference` required non-empty, trimmed; it is
  treated as sensitive data (not logged in full).
- **Merchant** — existence enforced by DB FK; `repo.ErrConstraint` →
  `NotFound`.

No arbitrary validation rules were added.

## 13. Error Handling

| Condition | Application/API Error |
| --- | --- |
| Nil request | `InvalidArgument` "payout request is required" |
| Invalid `client_id` / `merchant_id` UUID | `InvalidArgument` |
| Invalid / missing / zero / negative amount | `InvalidArgument` |
| Unsupported provider | `InvalidArgument` |
| Empty destination reference | `InvalidArgument` |
| `repo.ErrDuplicate` on create | `AlreadyExists` "payout already exists" |
| `repo.ErrConstraint` on create (FK violation) | `NotFound` "referenced merchant not found" |
| `repo.ErrNotFound` on get | `NotFound` "payout not found" |
| Repository/database failure | `Internal` (logged; no SQL/stack/paths leaked) |

All repository errors are mapped via `errors.Is` against repository sentinels.
No PostgreSQL internals, SQLC errors, pgx errors, or stack traces are exposed
to clients.

## 14. Mapping

- **Protobuf → Service:** `CreatePayoutRequest` fields extracted via
  `GetClientId()`, `GetMerchantId()`, `GetAmount()`, `GetProvider()`,
  `GetDestinationReference()`. No protobuf message is passed into the
  repository.
- **Service → Repository:** typed `uuid.UUID`, `pgtype.Numeric`,
  `sqlc.PaymentProvider`, `sqlc.PayoutStatus` parameters passed to
  `PayoutRepo.Create`. An idempotency key is generated server-side via
  `uuid.New()`.
- **Repository → Service:** `sqlc.Payout` received.
- **Service → Protobuf:** `payoutToProto` converts UUID→string,
  `pgtype.Numeric`→decimal-string (`Float64Value` + `FormatFloat 'f' 2`),
  enums→protobuf enums, and `pgtype.Timestamptz`→`timestamppb.Timestamp`
  (guarded by `.Valid`).

Mapping logic is centralized in `converters.go`.

## 15. Idempotency

An `idempotency_key UUID` is stored on the `payouts` table with a UNIQUE
constraint (from Agent 02). The service generates a fresh server-side
`uuid.New()` per payout creation, and the repository exposes
`GetByIdempotencyKey` for future duplicate detection. Duplicate protection is
authoritative at the database (unique constraint); the service maps
`repo.ErrDuplicate` → `AlreadyExists`. (The protobuf contract does not expose
an idempotency key field.)

## 16. Provider Boundary

**Provider execution is intentionally outside this agent.** The payout service
does not call PawaPay, banks, mobile-money APIs, or any external financial
gateway. It operates purely on the internal transaction model: validate →
persist → return. No provider interface, `ProviderClient`,
`PayoutProvider`, webhook, OAuth, or callback abstraction was introduced.
Provider execution and status reconciliation are documented as future work
(directive #50).

## 17. Security/Sensitive Data

Payout destination information (`destination_reference`) is treated as
sensitive:

- Only the `destination_reference` field is returned in the protobuf response
  as required by the API contract.
- The service logs do NOT include the full destination reference — logs emit
  only the `payout_id` and `merchant_id`. No account numbers, wallet numbers,
  full destination references, provider credentials, or tokens are logged.
- Provider credentials belong to configuration/secret management, not payout
  records or logs.

## 18. Tests and Validation

| Command | Result |
| --- | --- |
| `go build ./transactions/...` | ✅ Exit 0 |
| `go vet ./transactions/payouts/...` | ✅ Exit 0 |
| `go test ./transactions/payouts/... -v` | ✅ All 8 tests pass |
| `go test ./...` | ✅ Full repository — no failures (payouts/customers/deposits/merchants ok) |

Test coverage (using generated `mocks.MockPayoutRepo` + gomock):

- `TestRequestPayoutValidation` — table-driven: nil request, invalid
  client/merchant UUID, missing amount → `InvalidArgument`.
- `TestRequestPayoutZeroAmount` — zero amount → `InvalidArgument`.
- `TestRequestPayoutMissingDestination` — empty destination → `InvalidArgument`.
- `TestRequestPayoutSuccess` — full create path → protobuf payout returned.
- `TestRequestPayoutDuplicate` — `repo.ErrDuplicate` → `AlreadyExists`.
- `TestGetPayout` — get returns protobuf payout.
- `TestGetPayoutNotFound` — `repo.ErrNotFound` → `NotFound`.
- `TestGetPayoutInvalidID` — invalid UUID → `InvalidArgument`.
- `TestGetPayoutRepositoryError` — repository error → `Internal`.

## 19. Files Changed

Created:

- `transactions/payouts/service.go`
- `transactions/payouts/converters.go`
- `transactions/payouts/service_test.go`
- `docs/transactions-payouts-review.md`

No other files were modified. `git status --short` shows only
`transactions/payouts/` (untracked) and the documentation file. No database,
SQLC, protobuf, repository, merchant, customer, deposit, legacy-deposits,
Clients, third_party, or unrelated-service files were touched.

## 20. Risks

- **No legacy reference** — payouts are entirely new; there is no historical
  implementation to validate against, so the lifecycle semantics are
  established entirely from the domain model.
- **Destination free-form** — `destination_reference` is a free-text string;
  there is no destination-type enum or validation. Provider-specific
  destination requirements are deferred to the provider integration boundary.
- **Idempotency key not exposed in the contract** — the server generates a
  UUID per call; retries of the same logical payout currently produce a new
  payout (unique constraint prevents exact key collisions). Full retry
  idempotency with a client-supplied key is deferred to a later contract
  decision.
- **No status-mutation RPC** — lifecycle transitions are not reachable from
  the API in this agent; only `REQUESTED` payouts can be created and fetched.
- **Fee deduction** — the domain model states payouts deduct applicable fees,
  but no fee entity is defined; fee handling is deferred and unresolved.

## 21. Unresolved Questions

- **Fee entity** — the payout flow deducts fees per `docs/domain-model.md`, but
  no fee entity/field/table exists. This must be resolved before any
  fee-deduction logic is built.
- **Wallet/balance** — the administrator-controlled merchant wallet is not
  modelled; balance checks/insufficient-funds handling are unresolved.
- **Idempotency key semantics** — whether clients supply the idempotency key
  (requiring a protobuf field) or the server continues to generate it is
  unresolved.
- **Status-mutation RPCs** — whether the contract needs
  `UpdatePayoutStatus`/`ProcessPayoutStatus` RPCs to drive lifecycle
  transitions is unresolved.
- **Listing** — the contract defines no `ListPayouts` RPC; the repository
  supports `ListByClient/Merchant/Status` for future use.
- **Client validation** — `client_id` is not validated against the Clients
  Service via gRPC in this agent; that cross-service validation is deferred to
  the runtime/application boundary.
- **Destination modelling** — whether a structured destination (bank/wallet/
  mobile-money enum + fields) should replace the free-form
  `destination_reference` string is unresolved.