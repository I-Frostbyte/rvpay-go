# Transactions Deposits Implementation Review

Document Version: 1.0
Status: Complete
System: RVPay
Service: Transactions Service
Review: Agent 08 — Transactions Deposits Service

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
- existing `deposits/` service implementation
- the newly implemented `transactions/merchants/` and `transactions/customers/` services

## 2. Existing Deposits Implementation

The legacy `deposits/` service was inspected as a behavioral and coding-style
reference (its `InitiateDeposit` flow, `serdepositAmount` parsing, and
protobuf-converter pattern). Per the migration strategy, the legacy service is
**not deleted** and remains runnable until the Transactions service is
verified; this agent does not perform migration cleanup.

Behavior retained conceptually in the new Transactions deposit flow:

- Inbound "move money in" semantics (Initiated → Processing → Completed / Failed).
- Amount validation (positive, decimal) and currency representation.
- Customer/payer association via phone number.

Behavior changed for the new domain model:

- The deposit is now owned by **Merchant + Customer + Client** (no hard-coded
  `Socadel` client).
- The 4-state domain lifecycle replaces the legacy 5-state enum.
- Provider references use the shared `Provider`/`PaymentType` enums from
  `commongrpc` instead of legacy `DepositProvider`/`DepositType`.
- `Money` is a protobuf message (decimal string + currency) rather than a bare
  string amount.
- An idempotency key is generated server-side for duplicate detection.

## 3. Domain Model

Per `docs/domain-model.md`, a Deposit is an inbound customer payment moving
into the merchant wallet. The implemented `Deposit` resource exposes:

- `id` — stable internal identifier (UUID).
- `client_id`, `customer_id`, `merchant_id` — ownership references.
- `amount` (`commongrpc.Money`), `currency` — monetary value.
- `payment_type` — MMO or CREDIT_CARD.
- `payer_phone_number` — the payer's phone.
- `provider` — MTN_MOMO or ORANGE_MOMO.
- `status` — INITIATED / PROCESSING / COMPLETED / FAILED.
- `external_reference` — provider reference (nullable).
- `initiated_at` / `completed_at` / `failed_at` / `failure_reason` — lifecycle
  fields.
- `created_at` / `updated_at` — audit timestamps.

No invented fields.

## 4. Merchant Relationship

Every deposit belongs to exactly one merchant. Tenant isolation is enforced:

- **Repository** (`DepositRepo.Create`) requires `merchant_id`.
- **Service** validates `merchant_id` as a UUID before any repository call, and
  `DepositRepo.GetByID` returns the deposit with its `merchant_id`; the
  `ListByMerchant` repository method scopes to a single merchant (available for
  future merchant-scoped listing).
- **Database** has a `FOREIGN KEY (merchant_id) → merchants(id)` with
  `ON DELETE RESTRICT`.

A deposit for Merchant A is never retrievable through a request scoped to
Merchant B without an explicit `merchant_id`.

## 5. Customer Relationship

Each deposit references a customer. The service **validates the customer
belongs to the client+merchant context** before associating the deposit:

- `CreateDepositRequest` carries `client_id`, `customer_id`, `merchant_id`,
  and `payer_phone_number`.
- Before persisting, `CustomerRepo.GetByClientAndMerchantAndPhone(
  clientID, merchantID, phoneNumber)` is called. If the customer is not found,
  the deposit is rejected with `NotFound`. This prevents associating a
  customer that does not belong to the declared merchant/client/phone context.
- The persisted deposit uses the **validated customer's `ID`**, not the raw
  `customer_id` from the request.

This closes the tenant-isolation gap (directive #13) by proving customer
ownership before deposit creation.

## 6. Deposit Lifecycle

The domain-model lifecycle `Initiated → Processing → Completed / Failed` is
supported:

| Current State | Allowed State | Rule |
| --- | --- | --- |
| (none) | INITIATED | New deposits are created in INITIATED |
| INITIATED | PROCESSING / FAILED | Interim/terminal transitions (to be invoked by future provider/status agents) |
| PROCESSING | COMPLETED / FAILED | Terminal transitions (invoked by future provider agents) |
| COMPLETED | (terminal) | No further transition |
| FAILED | (terminal) | No further transition |

The repository exposes the transition methods (`UpdateStatus`,
`MarkCompleted`, `MarkFailed`). This agent does not expose arbitrary
`UpdateDepositStatus` RPCs — the protobuf contract defines only
`InitiateDeposit` and `GetDeposit`; lifecycle mutations are reserved for later
provider/status processing agents. The database enforces structural validity;
the service layer will enforce transition rules once a status-mutation
operation is introduced.

## 7. Service Structure

Package: `transactions/deposits/`

| File | Purpose |
| --- | --- |
| `service.go` | `Impl` struct, `NewDepositService`, `InitiateDeposit`, `GetDeposit`, and validation/converter helpers |
| `converters.go` | `depositToProto`, `sqlcDepositStatusToGrpc`, `sqlcPaymentTypeToGrpc`, `sqlcPaymentProviderToGrpc`, and amount formatting |
| `service_test.go` | Unit tests using generated mocks |

The `Impl` embeds `transactionsgrpc.UnimplementedDepositServiceServer` (by
value), holds injected `repo.DepositRepo`, `repo.CustomerRepo`, and
`zerolog.Logger`, and implements the generated `DepositServiceServer`
interface. It mirrors the Deposits/Merchant/Customer service conventions
(exact same structure and pattern).

## 8. Repository Integration

| Service Operation | Repository Operation |
| --- | --- |
| `InitiateDeposit` (customer validation) | `CustomerRepo.GetByClientAndMerchantAndPhone` |
| `InitiateDeposit` (persist) | `DepositRepo.Create` |
| `GetDeposit` | `DepositRepo.GetByID` |

No SQL is written in the service; no SQLC is called directly. All persistence
goes through injected repository interfaces.

## 9. RPC Implementation

| RPC | Purpose | Service Operation |
| --- | --- | --- |
| `InitiateDeposit` | Initiate a customer deposit | validate request → validate customer ownership → `DepositRepo.Create` → map to protobuf |
| `GetDeposit` | Fetch a deposit by id | validate id → `DepositRepo.GetByID` → map to protobuf |

Both RPCs return protobuf `Deposit` resources mapped by `depositToProto`;
SQLC models are never returned directly.

## 10. REST Exposure

REST routes are generated by grpc-gateway from the protobuf annotations
(Agent 04). No manual HTTP server or duplicate route definitions were added.

| Method | Route | RPC |
| --- | --- | --- |
| POST | `/v1/public/deposits` | InitiateDeposit |
| GET | `/v1/public/deposits/{deposit_id}` | GetDeposit |

## 11. Validation

- **Required IDs** — `client_id`, `customer_id`, `merchant_id` must be valid
  UUIDs; otherwise `InvalidArgument`.
- **Amount** — `Money` must be present, parseable as a decimal, and greater
  than zero (via `pgtype.Numeric.Scan` + `Float64Value`); otherwise
  `InvalidArgument`. No floating-point arithmetic is used for persisted
  values — the input decimal string is scanned directly into `pgtype.Numeric`.
- **Currency** — normalized to uppercase and required non-empty (`Money.GetCurrency`).
- **Payment type / provider** — must be a supported enum value via
  `grpcPaymentTypeToSqlc` / `grpcProviderToSqlc`; otherwise `InvalidArgument`.
- **Payer phone** — required non-empty, trimmed.
- **Customer ownership** — the customer must belong to the declared
  client+merchant+phone; otherwise `NotFound`.

No arbitrary validation rules were added beyond the schema/domain requirements.

## 12. Error Handling

| Condition | Application/API Error |
| --- | --- |
| Nil request | `InvalidArgument` "deposit request is required" |
| Invalid `client_id` / `customer_id` / `merchant_id` UUID | `InvalidArgument` |
| Invalid / missing / zero / negative amount | `InvalidArgument` |
| Unsupported payment type / provider | `InvalidArgument` |
| Empty payer phone | `InvalidArgument` |
| Customer not found in client+merchant+phone context | `NotFound` "customer not found for the given client, merchant, and phone number" |
| `repo.ErrDuplicate` on create | `AlreadyExists` "deposit already exists" |
| `repo.ErrConstraint` on create (FK violation) | `NotFound` "referenced merchant or customer not found" |
| `repo.ErrNotFound` on get | `NotFound` "deposit not found" |
| Repository/database failure | `Internal` (logged; no SQL/stack/paths leaked) |

All repository errors are mapped via `errors.Is` against repository sentinels.
No PostgreSQL internals, SQLC errors, pgx errors, or stack traces are exposed
to clients.

## 13. Mapping

- **Protobuf → Service:** `CreateDepositRequest` fields extracted via
  `GetClientId()`, `GetCustomerId()`, `GetMerchantId()`, `GetAmount()`,
  `GetPaymentType()`, `GetPayerPhoneNumber()`, `GetProvider()`. No protobuf
  message is passed into the repository.
- **Service → Repository:** typed `uuid.UUID`, `pgtype.Numeric`,
  `sqlc.PaymentType`, `sqlc.PaymentProvider`, `sqlc.DepositStatus`
  parameters passed to `DepositRepo.Create`. An idempotency key is generated
  server-side via `uuid.New()`.
- **Repository → Service:** `sqlc.Deposit` received.
- **Service → Protobuf:** `depositToProto` converts UUID→string,
  `pgtype.Numeric`→decimal-string (`Float64Value` + `FormatFloat 'f' 2`),
  enums→protobuf enums, and `pgtype.Timestamptz`→`timestamppb.Timestamp`
  (guarded by `.Valid`).

Mapping logic is centralized in `converters.go`; no large conversion blocks
are embedded in the RPC methods.

## 14. Idempotency

An `idempotency_key UUID` is stored on the `deposits` table with a UNIQUE
constraint (from Agent 02). The service generates a fresh server-side
`uuid.New()` per deposit creation, and the repository exposes
`GetByIdempotencyKey` for future duplicate detection. Duplicate protection is
authoritative at the database (unique constraint); the service maps
`repo.ErrDuplicate` → `AlreadyExists` without unsafe in-memory pre-checks.
(The protobuf contract does not currently expose an idempotency key field,
matching the documented `deposits-protobuf-review` open question.)

## 15. Provider Boundary

**Provider integration is not performed in this agent.** The deposit service
does not call PawaPay, GenerateHighLevel, Stripe, or any external provider.
It operates purely on the internal transaction model: validate → persist →
return. No `ProviderClient`, `PawaPayClient`, payment-provider, webhook, OAuth,
or callback abstraction was introduced. Provider orchestration and status
reconciliation belong to later/other architecture.

## 16. Tests and Validation

| Command | Result |
| --- | --- |
| `go build ./transactions/...` | ✅ Exit 0 |
| `go vet ./transactions/deposits/...` | ✅ Exit 0 |
| `go test ./transactions/deposits/... -v` | ✅ All 7 tests pass |
| `go test ./...` | ✅ Full repository — no failures (deposits/customers/merchants ok) |

Test coverage (using generated `mocks.MockDepositRepo`, `mocks.MockCustomerRepo` + gomock):

- `TestInitiateDepositValidation` — table-driven: nil request, invalid
  client/customer/merchant UUID, missing amount → `InvalidArgument`.
- `TestInitiateDepositZeroAmount` — zero amount → `InvalidArgument`.
- `TestInitiateDepositCustomerNotFound` — customer not in context →
  `NotFound`.
- `TestInitiateDepositSuccess` — full create path with validated customer →
  protobuf deposit returned.
- `TestGetDeposit` — get returns protobuf deposit.
- `TestGetDepositNotFound` — `repo.ErrNotFound` → `NotFound`.
- `TestGetDepositRepositoryError` — repository error → `Internal`.

## 17. Files Changed

Created:

- `transactions/deposits/service.go`
- `transactions/deposits/converters.go`
- `transactions/deposits/service_test.go`
- `docs/transactions-deposits-review.md`

No other files were modified. `git status --short` shows only
`transactions/deposits/` (untracked) and the documentation file. No database,
SQLC, protobuf, repository, merchant, customer, legacy-deposits,
third_party, or unrelated-service files were touched.

## 18. Risks

- **Merchant/customer existence relies on DB FK** — a bad `merchant_id`/
  `customer_id` is rejected only at persistence time (`repo.ErrConstraint` →
  `NotFound`). The customer is pre-validated in context by
  `GetByClientAndMerchantAndPhone`, but the merchant row itself is confirmed
  authoritatively by the FK.
- **Idempotency key not exposed in the contract** — the server generates a
  UUID per call; retries of the same logical deposit currently produce a new
  deposit (the unique constraint prevents exact key collisions). An atomic
  check-then-create with a client-supplied key would be needed for full
  retry idempotency (deferred to a later agent/contract decision).
- **No status-mutation RPC** — lifecycle transitions are not reachable from
  the API in this agent; only `INITIATED` deposits can be created and fetched.
- **`external_reference`** is nullable; `depositToProto` handles it as a plain
  string (empty when null) — acceptable for this scope.

## 19. Unresolved Questions

- **Provider orchestration** — the deposit flow currently does not initiate an
  external payment; how/where PawaPay integration and status reconciliation
  mount is deferred (later provider/status agent or runtime).
- **Idempotency key semantics** — whether clients supply the idempotency key
  (requiring a protobuf field) or the server continues to generate it is
  unresolved per the `deposits-protobuf-review` open question.
- **Status-mutation RPCs** — whether the contract needs
  `UpdateDepositStatus`/`ProcessDepositStatus` RPCs to drive lifecycle
  transitions is unresolved.
- **Listing** — the contract defines no `ListDeposits` RPC; the repository
  supports `ListByClient/Customer/Merchant/Status` for future use.
- **Client validation** — `client_id` is not validated against the Clients
  Service via gRPC in this agent; that cross-service validation is deferred to
  the runtime/application boundary.