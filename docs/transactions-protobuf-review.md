# Transactions Protobuf Implementation Review

Document Version: 1.0
Status: Complete
System: RVPay
Service: Transactions Service
Review: Agent 04 — Transactions Protobuf Contracts and Generation

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
- existing protobuf implementation (`protobuf/*.proto`, `protobuf/Makefile`,
  `grpc/go/`)

## 2. Existing Deposits API

The existing `protobuf/deposits.proto` (package `depositsgrpc`) defines:

- Service `DepositsService` with a single RPC `InitiateDeposit`.
- Enums: `DepositStatus` (6 states), `DepositType` (MMO/CARD),
  `DepositProvider` (`MTN_MOMO_CMR`, `ORANGE_MOMO_CMR`).
- Messages: `AccountDetails`, `Participant`, `CreateDepositRequest`,
  `CreateDepositResponse`.
- HTTP annotation: `POST /v1/public/deposits`.

The existing `protobuf/common.proto` (package `commongrpc`) contained shared
types for the Clients Service: `ClientStatus`, `PlatformStatus`,
`IntegrationStatus`, `OAuthStatus`, `WebhookStatus`, `ProviderCapability`,
`PaginationRequest/Response`, `Error`, `Metadata`, `AuditInformation`,
`HealthStatus`.

Per `docs/protobuf-strategy.md`, the shared `Provider`, `PaymentType`, and
`Money` types were missing from `common.proto` and have been added in this
agent.

## 3. Transactions Service

The Transactions Service is defined in `protobuf/transactions.proto`
(package `transactionsgrpc`). It exposes four services following the RPC
ownership table in `docs/protobuf-strategy.md`:

- `MerchantService` — CreateMerchant, GetMerchant, ListMerchants.
- `CustomerService` — CreateCustomer, GetCustomer.
- `DepositService` — InitiateDeposit, GetDeposit.
- `PayoutService` — RequestPayout, GetPayout.

## 4. RPCs

| RPC | Purpose | HTTP Exposure |
| --- | --- | --- |
| `CreateMerchant` | Register a new merchant | `POST /v1/public/merchants` |
| `GetMerchant` | Fetch a merchant by id | `GET /v1/public/merchants/{merchant_id}` |
| `ListMerchants` | List merchants | `GET /v1/public/merchants` |
| `CreateCustomer` | Create a customer record | `POST /v1/public/customers` |
| `GetCustomer` | Fetch a customer by id | `GET /v1/public/customers/{customer_id}` |
| `InitiateDeposit` | Initiate a customer deposit | `POST /v1/public/deposits` |
| `GetDeposit` | Fetch a deposit by id | `GET /v1/public/deposits/{deposit_id}` |
| `RequestPayout` | Request an outbound settlement | `POST /v1/public/payouts` |
| `GetPayout` | Fetch a payout by id | `GET /v1/public/payouts/{payout_id}` |

All RPCs are public (`/v1/public/...`), consistent with the existing gateway
strategy. No speculative administrative, debug, or raw database RPCs were
added. Lifecycle state mutation is not exposed as generic
`UpdateTransactionStatus`; persistence is handled by the repository/service
layers per the documented lifecycle.

## 5. Messages

Resource messages:

| Message | Purpose |
| --- | --- |
| `Merchant` | Payment gateway resource (id, name, slug, status, timestamps) |
| `Customer` | End-user payment record (id, client_id, merchant_id, phone_number, status, timestamps) |
| `Deposit` | Inbound payment (id, client_id, customer_id, merchant_id, amount, payment_type, payer_phone_number, provider, status, external_reference, lifecycle timestamps) |
| `Payout` | Outbound settlement (id, client_id, merchant_id, amount, provider, destination_reference, status, external_reference, lifecycle timestamps) |

Request/response messages:

- `CreateMerchantRequest/Response`, `GetMerchantRequest/Response`,
  `ListMerchantsRequest/Response`.
- `CreateCustomerRequest/Response`, `GetCustomerRequest/Response`.
- `CreateDepositRequest/Response`, `GetDepositRequest/Response`.
- `CreatePayoutRequest/Response`, `GetPayoutRequest/Response`.

Each RPC uses explicit request/response messages with distinct semantics; no
messages are reused across operations with different meanings.

## 6. Enums

| Enum | Values | Source |
| --- | --- | --- |
| `MerchantStatus` | UNSPECIFIED, ONBOARDED, ACTIVE, SUSPENDED, RETIRED | transactions.proto |
| `CustomerStatus` | UNSPECIFIED, CREATED, ACTIVE | transactions.proto |
| `DepositStatus` | UNSPECIFIED, INITIATED, PROCESSING, COMPLETED, FAILED | transactions.proto |
| `PayoutStatus` | UNSPECIFIED, REQUESTED, PROCESSING, COMPLETED, FAILED | transactions.proto |
| `Provider` | UNSPECIFIED, MTN_MOMO, ORANGE_MOMO | common.proto (shared) |
| `PaymentType` | UNSPECIFIED, MMO, CREDIT_CARD | common.proto (shared) |

Enum values match the database enums defined by Agent 02
(`docs/transactions-database-review.md`) and the domain-model lifecycle
states. Every enum has an explicit `UNSPECIFIED` zero value.

**Status mapping to database:** the protobuf `DepositStatus`
(`INITIATED/PROCESSING/COMPLETED/FAILED`) and `PayoutStatus`
(`REQUESTED/PROCESSING/COMPLETED/FAILED`) map 1:1 to the
`deposit_status`/`payout_status` PostgreSQL enums. No database status exists
that cannot be represented, and no protobuf status exists that is not
representable in the database.

## 7. Shared Types

The Transactions contract reuses the following shared types from
`commongrpc`:

- `commongrpc.Money` — monetary value (`amount` decimal string, `currency`
  ISO 4217 code), added to `common.proto` in this agent per
  `docs/protobuf-strategy.md`.
- `commongrpc.Provider` — payment provider enum (`MTN_MOMO`,
  `ORANGE_MOMO`), added to `common.proto`.
- `commongrpc.PaymentType` — payment method enum (`MMO`, `CREDIT_CARD`),
  added to `common.proto`.
- `commongrpc.PaginationRequest/Response` — reused by `ListMerchants`.

No Transactions-specific duplicate of `Money`, `Provider`, or `PaymentType`
was created. The `DepositType`/`DepositProvider` enums from the legacy
`deposits.proto` are not duplicated.

## 8. REST Exposure

| Method | Route | RPC |
| --- | --- | --- |
| POST | `/v1/public/merchants` | CreateMerchant |
| GET | `/v1/public/merchants/{merchant_id}` | GetMerchant |
| GET | `/v1/public/merchants` | ListMerchants |
| POST | `/v1/public/customers` | CreateCustomer |
| GET | `/v1/public/customers/{customer_id}` | GetCustomer |
| POST | `/v1/public/deposits` | InitiateDeposit |
| GET | `/v1/public/deposits/{deposit_id}` | GetDeposit |
| POST | `/v1/public/payouts` | RequestPayout |
| GET | `/v1/public/payouts/{payout_id}` | GetPayout |

Routes are resource-oriented, unambiguous, and follow the existing
`/v1/public/...` convention. HTTP methods are used correctly (GET for
retrieval, POST for creation/command). No DELETE RPCs were added since
deletion is not part of the documented Transactions domain.

Note: `POST /v1/public/deposits` is the same route as the legacy Deposits
service — this is intentional per the migration strategy: the Transactions
`DepositService.InitiateDeposit` supersedes the legacy `DepositsService`
contract, and the legacy service remains runnable until migration completes.

## 9. Compatibility

Per `docs/protobuf-strategy.md` and `docs/migration-plan.md`, the legacy
`deposits.proto` (package `depositsgrpc`) is **replaced** by
`transactions.proto` (package `transactionsgrpc`). Compatibility analysis:

| Existing Deposits Contract | Classification | Transaction Target |
| --- | --- | --- |
| `DepositsService.InitiateDeposit` | Replace / Extend | `DepositService.InitiateDeposit` (same HTTP route, richer request/response) |
| `DepositStatus` enum | Replace | `transactionsgrpc.DepositStatus` (4-state lifecycle per domain model) |
| `DepositType` enum | Replace | `commongrpc.PaymentType` |
| `DepositProvider` enum | Replace | `commongrpc.Provider` |
| `AccountDetails` / `Participant` | Replace | `commongrpc.Money` + inline fields in `CreateDepositRequest` |

The legacy `deposits.proto` and `grpc/go/depositsgrpc/` were **not modified**
or deleted. They remain for the Deposits service until it is fully replaced.
The additive change to `common.proto` (adding `Provider`, `PaymentType`,
`Money`) does not break existing `commongrpc` consumers (Clients Service).

## 10. Generated Output

Generated output is placed in the documented locations:

- `grpc/go/commongrpc/common.pb.go` — modified; regenerated with the new
  shared `Provider`, `PaymentType`, `Money` definitions.
- `grpc/go/transactionsgrpc/transactions.pb.go` — generated messages.
- `grpc/go/transactionsgrpc/transactions_grpc.pb.go` — generated gRPC stubs.
- `grpc/go/transactionsgrpc/transactions.pb.gw.go` — generated grpc-gateway
  handlers.

No other generated packages were regenerated. The existing
`depositsgrpc`, `integrationsgrpc`, and `clientsgrpc` generated code is
untouched (verified via `git status`). Generated files were not manually
edited.

## 11. Generator Versions

Generation used the repository's existing `protobuf/Makefile` workflow
(identical `protoc` flags), executed per-file to avoid regenerating unrelated
packages:

- `protoc` — as pinned by the repository environment.
- `protoc-gen-go` — as pinned by the repository environment.
- `protoc-gen-go-grpc` — as pinned by the repository environment.
- `protoc-gen-grpc-gateway` — as pinned by the repository environment.

No generator versions were changed or upgraded.

## 12. Validation

| Command | Result |
| --- | --- |
| `protoc` generation (common.proto + transactions.proto) | ✅ Exit 0 |
| `go build ./grpc/go/transactionsgrpc/...` | ✅ Exit 0 |
| `go build ./grpc/go/commongrpc/...` | ✅ Exit 0 |
| `go build ./...` | ✅ Exit 0 (full repository, including Clients which imports commongrpc) |
| `go vet ./grpc/go/transactionsgrpc/... ./grpc/go/commongrpc/...` | ✅ Exit 0 |
| `go test ./...` | ✅ All packages pass |
| Regeneration (2nd run) | ✅ Deterministic — no unexpected diff |
| `git status --short` | ✅ Only `protobuf/common.proto`, `protobuf/transactions.proto`, `grpc/go/commongrpc/common.pb.go`, `grpc/go/transactionsgrpc/` changed |

## 13. Risks

- **Route collision** — `POST /v1/public/deposits` exists on both the legacy
  Deposits service and the Transactions `DepositService`. During the migration
  window (both services deployed), the gateway must route to the intended
  service; the migration plan requires the legacy service to be retired after
  the Transactions service is verified.
- **Backward compatibility** — the Transactions `CreateDepositRequest` differs
  from the legacy `CreateDepositRequest` (Money message vs. amount string,
  merchant/customer IDs added, shared Provider/PaymentType enums). Legacy
  clients must migrate to the new contract.
- **Enum value changes** — the shared `Provider` values (`MTN_MOMO`,
  `ORANGE_MOMO`) differ textually from the legacy `DepositProvider` values
  (`MTN_MOMO_CMR`, `ORANGE_MOMO_CMR`); a mapping is required for migrated data.
- **Pagination** — `ListMerchants` uses the shared `PaginationRequest/Response`
  types, but no pagination was implemented at the SQL layer in Agent 03. The
  ListMerchants response currently returns all merchants; pagination support
  must be aligned between the protobuf contract and the repository/service
  implementation.

## 14. Unresolved Questions

- **Fee entity** — the payout flow deducts fees, but no fee entity/field is
  defined in the domain; the `Payout` message has no fee representation. Must
  be resolved before Payout implementation (Agent 09).
- **Wallet entity** — the merchant wallet is not modelled; no wallet-balance
  field exists on `Merchant`. Balance tracking may require a protobuf addition.
- **Idempotency key exposure** — `deposits`/`payouts` have a database
  `idempotency_key`, but no idempotency field is exposed in the
  `CreateDepositRequest`/`CreatePayoutRequest` messages. Whether clients
  supply an idempotency key is unresolved.
- **Transaction list/status RPCs** — the Clients Service requires transaction
  status queries (per `docs/domain-model.md` §5.2), but no
  `ListDeposits`/`ListPayouts`/`GetTransactionStatus` RPC was defined. Whether
  these are needed in the initial contract is unresolved.
- **Provider scoping** — the `merchants` table supports multiple providers,
  but no `providers` field/relationship exists on the `Merchant` message;
  `CreateMerchantRequest` takes only name/slug. How merchant-supported
  providers are represented is unresolved.