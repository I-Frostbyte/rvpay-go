# Transactions Merchants Implementation Review

Document Version: 1.0
Status: Complete
System: RVPay
Service: Transactions Service
Review: Agent 06 — Transactions Merchants Service

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
- existing `deposits/deposits/service.go` implementation

## 2. Existing Deposits Pattern

The existing Deposits service defines the implementation conventions copied
for the merchant service:

- `Impl` struct embedding `UnimplementedDepositsServiceServer` for forward
  compatibility.
- `NewDepositsService(repo, logger, ...)` constructor with explicit
  dependency injection.
- Thin gRPC methods that validate, normalize, call the repository through the
  `repo` interface, and map results to protobuf responses.
- `status.Error(codes.X, "message")` for all error mapping; never exposing
  raw PostgreSQL/SQLC errors.
- Converter helper functions for enum/type mapping between protobuf and sqlc.
- zerolog via a value-receiver `logger zerolog.Logger` field.

## 3. Merchant Domain

Per `docs/domain-model.md`, a Merchant is a payment gateway (e.g. PawaPay,
Flutterwave) with the lifecycle `Onboarded → Active → Suspended → Retired`.
Merchants process deposits and payouts and serve customers.

The merchant resource implemented in this agent exposes:

- `id` — stable public identifier (UUID, DB-generated).
- `name` — merchant display name.
- `slug` — stable URL-friendly unique identifier.
- `status` — merchant lifecycle state.
- `created_at` / `updated_at` — audit timestamps.

No additional merchant fields were invented. Customers, deposits, and payouts
are NOT eagerly loaded (the merchant response contains only merchant
information per the API contract).

## 4. Service Structure

Package: `transactions/merchants/`

| File | Purpose |
| --- | --- |
| `service.go` | `Impl` struct, `NewMerchantService`, and the three MerchantService gRPC methods |
| `converters.go` | `merchantToProto` and `sqlcMerchantStatusToGrpc` mapping helpers |
| `service_test.go` | Unit tests using the generated `MockMerchantRepo` |

The `Impl` struct embeds `transactionsgrpc.UnimplementedMerchantServiceServer`
(by value) for forward compatibility, holds the injected `repo.MerchantRepo`
and `zerolog.Logger`, and implements the generated
`MerchantServiceServer` interface. The gRPC layer is thin: it validates,
invokes the repository, maps errors, and returns protobuf responses. No SQL,
SQLC calls, or database access exist in the service.

## 5. Repository Integration

The service consumes only the `repo.MerchantRepo` interface:

| Repository Method | Used By |
| --- | --- |
| `Create` | `CreateMerchant` |
| `GetByID` | `GetMerchant` |
| `List` | `ListMerchants` |

No SQL is written in the service. The repository interface is injected through
the constructor; no repository is instantiated inside service methods.

## 6. Validation

Implemented validation (client-facing, before repository calls):

- `CreateMerchant`:
  - nil request → `InvalidArgument`
  - empty/whitespace `name` → `InvalidArgument`
  - empty/whitespace `slug` → `InvalidArgument`
- `GetMerchant`:
  - nil request → `InvalidArgument`
  - `merchant_id` not a valid UUID → `InvalidArgument`
- `ListMerchants`:
  - nil request → `InvalidArgument`

No arbitrary validation rules were introduced. Uniqueness of `slug` is
enforced authoritatively by the database unique constraint; the service maps
`repo.ErrDuplicate` to `AlreadyExists` without pre-checking via in-memory
maps or mutexes (safe under concurrent requests).

## 7. Merchant Lifecycle

A newly created merchant is persisted with `sqlc.MerchantStatusONBOARDED`
(the initial lifecycle state per the domain model). No arbitrary status
transition RPCs were exposed — the protobuf contract (Agent 04) defines only
`CreateMerchant`, `GetMerchant`, and `ListMerchants`. Status transition rules
are not implemented because the API contract does not expose a status-change
operation; this is deferred to later agents if the architecture requires it.
The `sqlcMerchantStatusToGrpc` converter maps all four domain states
(ONBOARDED/ACTIVE/SUSPENDED/RETIRED) plus the UNSPECIFIED fallback.

## 8. RPC Implementation

| RPC | Purpose | Repository Operation |
| --- | --- | --- |
| `CreateMerchant` | Register a new merchant in the ONBOARDED state | `MerchantRepo.Create` |
| `GetMerchant` | Fetch a merchant by id | `MerchantRepo.GetByID` |
| `ListMerchants` | List all merchants | `MerchantRepo.List` |

All three RPCs return protobuf `Merchant` resources mapped by
`merchantToProto`; SQLC models are never returned directly.

## 9. REST Exposure

REST routes are generated by grpc-gateway from the protobuf annotations
(Agent 04). No manual HTTP server or duplicate route definitions were added.

| Method | Route | RPC |
| --- | --- | --- |
| POST | `/v1/public/merchants` | CreateMerchant |
| GET | `/v1/public/merchants/{merchant_id}` | GetMerchant |
| GET | `/v1/public/merchants` | ListMerchants |

## 10. Error Handling

| Condition | gRPC Code |
| --- | --- |
| Nil request | `InvalidArgument` |
| Missing/empty name or slug | `InvalidArgument` |
| Invalid UUID | `InvalidArgument` |
| `repo.ErrDuplicate` on create | `AlreadyExists` ("merchant already exists") |
| `repo.ErrNotFound` on get | `NotFound` ("merchant not found") |
| Repository/database failure | `Internal` (logged; no SQL/stack/paths leaked) |

All repository errors are mapped via `errors.Is` against the repository
sentinels (`repo.ErrDuplicate`, `repo.ErrNotFound`); any other error is logged
with zerolog and returned as a generic `Internal` error. No PostgreSQL
internals, SQL strings, stack traces, or file paths are exposed to clients.

## 11. Mapping

- **Protobuf → Repository:** `CreateMerchantRequest` fields are extracted via
  `req.GetName()`/`req.GetSlug()`, trimmed, and passed as
  `(name, slug, sqlc.MerchantStatusONBOARDED)` to `MerchantRepo.Create`. No
  protobuf message is passed into the repository.
- **Repository → Protobuf:** `sqlc.Merchant` is mapped by `merchantToProto`
  (converting UUID→string, `sqlc.MerchantStatus`→`transactionsgrpc.MerchantStatus`,
  and `time.Time`→`timestamppb.Timestamp`). SQLC models are never returned
  directly through gRPC.

Mapping logic is centralized in `converters.go`; no large conversion blocks
are embedded in the RPC methods.

## 12. Tests and Validation

| Command | Result |
| --- | --- |
| `go build ./transactions/... ./grpc/go/transactionsgrpc/...` | ✅ Exit 0 |
| `go vet ./transactions/...` | ✅ Exit 0 |
| `go test ./transactions/merchants/... -v` | ✅ All 8 tests pass |
| `go test ./...` | ✅ Full repository — no failures |

Test coverage (using the generated `mocks.MockMerchantRepo` with gomock):

- `TestCreateMerchant` — table-driven: nil request, empty name, empty slug
  → `InvalidArgument`.
- `TestCreateMerchantSuccess` — create returns the protobuf merchant.
- `TestCreateMerchantDuplicate` — `repo.ErrDuplicate` → `AlreadyExists`.
- `TestGetMerchant` — get returns the protobuf merchant.
- `TestGetMerchantNotFound` — `repo.ErrNotFound` → `NotFound`.
- `TestGetMerchantInvalidID` — invalid UUID → `InvalidArgument`.
- `TestListMerchants` — two merchants returned with `TotalCount`.
- `TestListMerchantsRepositoryError` — repository error → `Internal`.

## 13. Files Changed

Created:

- `transactions/merchants/service.go`
- `transactions/merchants/converters.go`
- `transactions/merchants/service_test.go`
- `docs/transactions-merchants-review.md`

No other files were modified. `git status --short` shows only
`transactions/merchants/` (untracked) and the documentation file. No
database, SQLC, protobuf, repository, or unrelated-service files were touched.

## 14. Risks

- **Pagination deferred** — `ListMerchants` returns the full result set with
  `TotalCount` and an empty `next_page_token`. The SQLC/repository layer has no
  limit/offset support; unbounded result growth is possible for large merchant
  tables.
- **Status transitions not exposed** — the API contract has no status-change
  RPC, so lifecycle transitions (Active/Suspended/Retired) cannot currently be
  invoked by clients. This is a contract limitation, not a service defect.
- **Slug uniqueness** — enforced by the database; the service maps
  `ErrDuplicate` to `AlreadyExists`. Duplicate detection is authoritative
  under concurrency.
- **Timestamps** — `merchantToProto` assumes non-null `created_at`/`updated_at`
  (both are `NOT NULL DEFAULT NOW()` at the DB); a nil timestamp would produce
  an invalid `timestamppb` and is not expected from the schema.

## 15. Unresolved Questions

- **Pagination** — whether `ListMerchants` should support
  limit/offset/cursor pagination in the initial contract is unresolved (the
  SQLC/repository layer has no pagination queries).
- **Merchant update** — the protobuf contract defines no `UpdateMerchant`
  RPC; whether merchant updates (name/slug) are required is unresolved.
- **Status transitions** — the domain defines a lifecycle but no API operation
  to transition status; whether an `ActivateMerchant`/`SuspendMerchant`/
  `RetireMerchant` RPC set is required is unresolved.
- **Provider relationship** — the domain says a merchant may support multiple
  payment providers, but the merchant resource has no provider field or
  relationship; how merchant-provider association is represented is
  unresolved.