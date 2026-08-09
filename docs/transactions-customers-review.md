# Transactions Customers Implementation Review

Document Version: 1.0
Status: Complete
System: RVPay
Service: Transactions Service
Review: Agent 07 — Transactions Customers Service

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
- existing `deposits/deposits/service.go` implementation
- the newly implemented `transactions/merchants/` service

## 2. Existing Implementation References

- The existing Deposits service establishes the canonical service conventions
  (`Impl` struct + embedded `Unimplemented*Server`, constructor injection,
  thin gRPC methods, `status.Error(codes.X, ...)` mapping, converter helpers,
  zerolog).
- The Transactions merchant service (Agent 06) is the local reference for the
  same conventions applied to the Transactions domain. The customer service
  mirrors the merchant package structure exactly:
  `transactions/customers/{service.go, converters.go, service_test.go}`.

## 3. Customer Domain

Per `docs/domain-model.md`, a Customer is an end user making payments. It is
not a registered RVPay user; its interaction is limited to payment initiation.
The customer resource exposes:

- `id` — stable public identifier (UUID, DB-generated).
- `client_id` — the RVPay client the customer belongs to.
- `merchant_id` — the merchant serving the customer (tenant boundary).
- `phone_number` — the customer's payment phone number.
- `status` — customer lifecycle state (`CREATED` → `ACTIVE`).
- `created_at` / `updated_at` — audit timestamps.

The only documented uniqueness constraint is
`UNIQUE (client_id, merchant_id, phone_number)` scoped to a client+merchant.
No additional customer fields were invented.

## 4. Merchant/Customer Relationship

A customer belongs to exactly one client and one merchant (per
`docs/domain-model.md`). The tenant boundary is enforced at multiple layers:

- **Database:** `customers` has `merchant_id` with a `FOREIGN KEY` to
  `merchants(id)` (`ON DELETE RESTRICT`) and the unique constraint
  `(client_id, merchant_id, phone_number)`.
- **Repository:** `CustomerRepo.Create` requires `merchant_id`;
  `GetByClientAndMerchantAndPhone` requires both `client_id` and
  `merchant_id`. No query can cross merchant boundaries.
- **Service:** `CreateCustomer` requires and parses `merchant_id` as a UUID
  before any repository call.

A customer belonging to Merchant A cannot be created for Merchant B via the
customer service without an explicit and validated `merchant_id`.

## 5. Service Structure

Package: `transactions/customers/`

| File | Purpose |
| --- | --- |
| `service.go` | `Impl` struct, `NewCustomerService`, and the CustomerService gRPC methods |
| `converters.go` | `customerToProto` and `sqlcCustomerStatusToGrpc` mapping helpers |
| `service_test.go` | Unit tests using the generated `MockCustomerRepo` |

The `Impl` embeds `transactionsgrpc.UnimplementedCustomerServiceServer` (by
value), holds the injected `repo.CustomerRepo` and `zerolog.Logger`, and
implements the generated `CustomerServiceServer` interface. The gRPC layer is
thin — validate, invoke repository, map errors, return protobuf. No SQL,
SQLC, database, provider, OAuth, or webhook logic exists in the service.

## 6. Repository Integration

The service consumes the `repo.CustomerRepo` interface only:

| Repository Method | Used By |
| --- | --- |
| `Create(clientID, merchantID, phoneNumber, status)` | `CreateCustomer` |
| `GetByID(id)` | `GetCustomer` |

No SQL is written in the service. The repository is injected through the
constructor; no repository is instantiated inside service methods. No
MerchantService dependency is injected (avoiding a circular/service-to-service
call) because merchant existence is enforced by the database foreign key on
`customers.merchant_id`.

## 7. Validation

Implemented client-facing validation before repository calls:

- `CreateCustomer`:
  - nil request → `InvalidArgument`
  - `client_id` not a valid UUID → `InvalidArgument`
  - `merchant_id` not a valid UUID → `InvalidArgument`
  - empty/whitespace `phone_number` → `InvalidArgument`
- `GetCustomer`:
  - nil request → `InvalidArgument`
  - `customer_id` not a valid UUID → `InvalidArgument`

No arbitrary validation rules were introduced. Duplicate customers
(`(client_id, merchant_id, phone_number)`) are enforced authoritatively by
the database unique constraint; the service maps `repo.ErrDuplicate` to
`AlreadyExists` without an unsafe in-memory pre-check.

## 8. Customer Lifecycle

A newly created customer is persisted with `sqlc.CustomerStatusCREATED` (the
initial lifecycle state per the domain model). The `CustomerStatus` enum in
the database (`CREATED`, `ACTIVE`) maps 1:1 to the protobuf
`CustomerStatus_CUSTOMER_STATUS_CREATED/ACTIVE` values via the converter.

No status-change RPC exists in the protobuf contract; the transition
`CREATED → ACTIVE` is therefore not exposed through the API in this agent
and is deferred to later transaction-processing work. The database enforces
structural validity; the service layer will enforce transition rules once a
status-mutation operation is defined.

## 9. RPC Implementation

| RPC | Purpose | Repository Operation |
| --- | --- | --- |
| `CreateCustomer` | Create a customer record in the CREATED state | `CustomerRepo.Create` |
| `GetCustomer` | Fetch a customer by id | `CustomerRepo.GetByID` |

Both RPCs return protobuf `Customer` resources mapped by `customerToProto`;
SQLC models are never returned directly.

## 10. REST Exposure

REST routes are generated by grpc-gateway from the protobuf annotations
(Agent 04). No manual HTTP server or duplicate route definitions were added.

| Method | Route | RPC |
| --- | --- | --- |
| POST | `/v1/public/customers` | CreateCustomer |
| GET | `/v1/public/customers/{customer_id}` | GetCustomer |

## 11. Error Handling

| Condition | gRPC Code |
| --- | --- |
| Nil request | `InvalidArgument` |
| Invalid `client_id` / `merchant_id` / `customer_id` UUID | `InvalidArgument` |
| Empty `phone_number` | `InvalidArgument` |
| `repo.ErrDuplicate` on create | `AlreadyExists` ("customer already exists") |
| `repo.ErrConstraint` on create (FK violation) | `NotFound` ("referenced merchant not found") |
| `repo.ErrNotFound` on get | `NotFound` ("customer not found") |
| Repository/database failure | `Internal` (logged; no SQL/stack/paths leaked) |

All repository errors are mapped via `errors.Is` against the repository
sentinels. No PostgreSQL internals, SQL strings, stack traces, or file paths
are exposed to clients.

## 12. Mapping

- **Protobuf → Service:** request fields are extracted via `req.GetClientId()`,
  `req.GetMerchantId()`, `req.GetPhoneNumber()`, validated as UUIDs, trimmed,
  and passed as `(clientID, merchantID, phoneNumber,
  sqlc.CustomerStatusCREATED)` to `CustomerRepo.Create`. No protobuf message
  is passed into the repository.
- **Service → Repository:** typed `uuid.UUID` and `sqlc.CustomerStatus`
  parameters.
- **Repository → Service:** `sqlc.Customer` is received.
- **Service → Protobuf:** `customerToProto` converts UUID→string,
  `sqlc.CustomerStatus`→`transactionsgrpc.CustomerStatus`, and
  `time.Time`→`timestamppb.Timestamp`.

Mapping logic is centralized in `converters.go`; no large conversion blocks
are embedded in the RPC methods.

## 13. Tenant Isolation

- `CreateCustomer` requires both `client_id` and `merchant_id` as validated
  UUIDs; the caller cannot create a customer without declaring its owning
  client and merchant.
- `GetCustomer` retrieves by customer ID only (the contract exposes no
  merchant-scoped get); the customer record itself carries `merchant_id`, and
  the repository's `Create`/`GetByClientAndMerchantAndPhone` methods preserve
  the tenant boundary. No global customer retrieval by phone number exists,
  so a customer cannot be found across tenants.
- The database FK + unique constraint enforce the boundary authoritatively.

## 14. Tests and Validation

| Command | Result |
| --- | --- |
| `go build ./transactions/...` | ✅ Exit 0 |
| `go vet ./transactions/customers/...` | ✅ Exit 0 |
| `go test ./transactions/customers/... -v` | ✅ All 7 tests pass |
| `go test ./...` | ✅ Full repository — no failures (transactions/customers + transactions/merchants ok) |

Test coverage (using the generated `mocks.MockCustomerRepo` with gomock):

- `TestCreateCustomer` — table-driven: nil request, invalid client id,
  invalid merchant id, empty phone number → `InvalidArgument`.
- `TestCreateCustomerSuccess` — create returns the protobuf customer.
- `TestCreateCustomerDuplicate` — `repo.ErrDuplicate` → `AlreadyExists`.
- `TestCreateCustomerMerchantNotFound` — `repo.ErrConstraint` → `NotFound`.
- `TestGetCustomer` — get returns the protobuf customer.
- `TestGetCustomerNotFound` — `repo.ErrNotFound` → `NotFound`.
- `TestGetCustomerInvalidID` — invalid UUID → `InvalidArgument`.
- `TestGetCustomerRepositoryError` — repository error → `Internal`.

## 15. Files Changed

Created:

- `transactions/customers/service.go`
- `transactions/customers/converters.go`
- `transactions/customers/service_test.go`
- `docs/transactions-customers-review.md`

No other files were modified. `git status --short` shows only
`transactions/customers/` (untracked) and the documentation file. No
database, SQLC, protobuf, repository, merchant, or unrelated-service files
were touched.

## 16. Risks

- **Merchant existence** — validated by the database FK; the service maps
  `repo.ErrConstraint` to `NotFound`. A non-existent `merchant_id` is only
  rejected at persistence time (no pre-query). For high error volumes, a
  defensive merchant existence check could be added, but the DB constraint
  guarantees correctness.
- **Client existence** — `client_id` has no local FK (it references a Clients
  Service record resolved via gRPC). Client existence is not validated in this
  agent; the runtime/service boundary must validate `client_id` against the
  Clients Service for deposits/payouts (per the domain model assumption).
- **Status transitions not exposed** — the contract defines no status-mutation
  RPC; the `CREATED → ACTIVE` transition is deferred.
- **Timestamps** — `customerToProto` assumes non-null `created_at`/`updated_at`
  (both are `NOT NULL DEFAULT NOW()` at the DB).

## 17. Unresolved Questions

- **Customer listing** — the protobuf contract defines no `ListCustomers`
  RPC, so no list/pagination/filtering surface exists yet. Whether listing is
  required (e.g. for client/admin monitoring per the domain model) is
  unresolved; the repository already supports `ListByClient`/`ListByMerchant`.
- **Client validation** — `client_id` is not validated against the Clients
  Service in this agent; the cross-service validation boundary is unresolved
  until the runtime/deposit agents.
- **Status lifecycle RPCs** — whether an `ActivateCustomer`/
  `UpdateCustomerStatus` RPC is needed is unresolved.
- **Immutable fields** — the contract defines no update RPC; whether customer
  phone/status updates are required is unresolved.