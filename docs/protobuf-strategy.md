# RVPay Protobuf Strategy

Document Version: 1.0
Status: Foundation
System: RVPay
Architecture: Go Microservices
Communication: gRPC + Protocol Buffers + gRPC-Gateway (HTTP)

## 1. Purpose

This document defines the protobuf contract strategy for the new RVPay
architecture described in `docs/domain-model.md` and
`docs/repository-layout.md`.

This is a design document only. No protobuf files are generated, no `protoc`
invocations are run, and no generated code is modified.

## 2. Current Contracts

### 2.1 `protobuf/deposits.proto` — package `depositsgrpc`

Service: `DepositsService`

| RPC | Request | Response |
|-----|---------|----------|
| `InitiateDeposit` | `CreateDepositRequest` | `CreateDepositResponse` |

Enums:

- `DepositStatus` — `UNSPECIFIED`, `PENDING`, `ACCEPTED`, `COMPLETED`,
  `FAILED`, `REJECTED`
- `DepositType` — `UNSPECIFIED`, `MMO`, `CARD`
- `DepositProvider` — `UNSPECIFIED`, `MTN_MOMO_CMR`, `ORANGE_MOMO_CMR`

Messages:

- `AccountDetails` — `phone_number`, `provider`
- `Participant` — `type`, `account_details`
- `CreateDepositRequest` — `amount`, `currency`, `payer`, `client_id`
- `CreateDepositResponse` — `deposit_id`, `status`, `next_step`

HTTP annotation: `POST /v1/public/deposits`

### 2.2 `protobuf/integrations.proto` — package `integrationsgrpc`

Service: `IntegrationService`

| RPC | Request | Response |
|-----|---------|----------|
| `CreateIntegration` | `CreateIntegrationRequest` | `CreateIntegrationResponse` |
| `GetIntegration` | `GetIntegrationRequest` | `GetIntegrationResponse` |
| `DeleteIntegration` | `DeleteIntegrationRequest` | `DeleteIntegrationResponse` |
| `ProcessWebhookEvent` | `ProcessWebhookEventRequest` | `ProcessWebhookEventResponse` |

Messages:

- `Integration` — `id`, `provider`, `location_id`, `access_token`,
  `refresh_token`, `token_expires_at`, `created_at`, `updated_at`
- `CreateIntegrationRequest` — `provider`, `location_id`, `access_token`,
  `refresh_token`, `token_expires_at`
- `CreateIntegrationResponse` — `integration`
- `GetIntegrationRequest` — `id`
- `GetIntegrationResponse` — `integration`
- `DeleteIntegrationRequest` — `id`
- `DeleteIntegrationResponse` — `id`
- `ProcessWebhookEventRequest` — `provider`, `event_type`, `payload`
- `ProcessWebhookEventResponse` — `id`, `processed`

HTTP annotations:

- `POST /v1/public/integrations`
- `GET /v1/public/integrations/{id}`
- `DELETE /v1/public/integrations/{id}`
- `POST /v1/public/webhooks`

## 3. Contract Assessment

### 3.1 Contracts That Remain Valid

| Contract | Verdict | Rationale |
|----------|---------|-----------|
| `InitiateDeposit` RPC | Valid, evolves | Core deposit flow remains; moves to the Transactions Service contract |
| `DepositStatus` enum | Valid | Status lifecycle (pending/accepted/completed/failed/rejected) is unchanged |
| `Integration` CRUD RPCs | Valid, evolves | Integration management remains; moves to the Clients Service contract |
| `ProcessWebhookEvent` RPC | Valid, evolves | Webhook ingestion remains; moves to the Clients Service contract |
| `google.api.http` annotations | Valid | Gateway strategy is preserved |

### 3.2 Contracts That Require Replacement

| Contract | Reason |
|----------|--------|
| `deposits.proto` | Replaced by `transactions.proto`; the Deposits service evolves into the Transactions Service owning Merchants, Customers, Deposits, and Payouts |
| `integrations.proto` | Replaced by `clients.proto`; the Integrations service evolves into the Clients Service owning Clients, Platforms, and Integrations |
| `DepositType` enum | Replaced by a shared `PaymentType` enum (`TYPE_MMO`, `TYPE_CREDIT_CARD`) per the TDD |
| `DepositProvider` enum | Replaced by a shared `Provider` enum (`PROVIDER_MTN_MOMO`, `PROVIDER_ORANGE_MOMO`) per the TDD |
| `Integration` message | Evolves to link `platform_id` and `client_id` per the domain model; tokens remain integration-owned |
| `AccountDetails` / `Participant` | Evolve into the Transactions contract; payer details remain deposit-specific |

## 4. Protobuf Ownership

Every service owns its own protobuf package. No service imports another
service's package. Shared types are isolated into a common package.

| Package | Owner | Source File |
|---------|-------|-------------|
| `clientsgrpc` | Clients Service | `protobuf/clients.proto` |
| `transactionsgrpc` | Transactions Service | `protobuf/transactions.proto` |
| `commongrpc` | Shared (no service) | `protobuf/common.proto` |

## 5. Package Naming

Package naming follows the existing repository convention: `<service>grpc`.

| Package | Go Package Option |
|---------|-------------------|
| `clientsgrpc` | `github.com/I-Frostbyte/rvpay-go/grpc/go/clientsgrpc` |
| `transactionsgrpc` | `github.com/I-Frostbyte/rvpay-go/grpc/go/transactionsgrpc` |
| `commongrpc` | `github.com/I-Frostbyte/rvpay-go/grpc/go/commongrpc` |

Rules:

- Every service owns exactly one protobuf package.
- Service packages never import each other.
- Service packages may import only `commongrpc` and well-known types
  (`google/protobuf/*.proto`, `google/api/annotations.proto`).
- `go_package` must always match the repository Go module
  (`github.com/I-Frostbyte/rvpay-go`) to keep generated code importable.

## 6. Shared Messages and Enums

Shared types live in `protobuf/common.proto` (package `commongrpc`). They are
imported by service packages and never duplicated.

### 6.1 Shared Enums

```proto
enum UserRole {
  USER_ROLE_UNSPECIFIED = 0;
  USER_ROLE_USER = 1;
  USER_ROLE_ADMIN = 2;
}

enum Provider {
  PROVIDER_UNSPECIFIED = 0;
  PROVIDER_MTN_MOMO = 1;
  PROVIDER_ORANGE_MOMO = 2;
}

enum PaymentType {
  PAYMENT_TYPE_UNSPECIFIED = 0;
  PAYMENT_TYPE_MMO = 1;
  PAYMENT_TYPE_CREDIT_CARD = 2;
}
```

- `UserRole` — used by the Clients Service for client records.
- `Provider` — used by the Transactions Service for merchants, deposits, and
  payouts.
- `PaymentType` — used by the Transactions Service for deposits and payouts.

### 6.2 Shared Messages

```proto
message Money {
  string amount = 1;    // decimal string, e.g. "1000.00"
  string currency = 2;  // uppercase ISO 4217 code, e.g. "XAF"
}
```

- `Money` — used by the Transactions Service for deposit and payout amounts.

### 6.3 Message Reuse Rules

- A message used by exactly one service stays in that service's package.
- A message used by two or more services moves to `commongrpc`.
- Service packages must not redefine types that exist in `commongrpc`.
- Well-known types (`google.protobuf.Timestamp`, etc.) are used directly and
  are not wrapped.

## 7. RPC Ownership

### 7.1 Clients Service — `clientsgrpc`

| Service | RPC | Request | Response |
|---------|-----|---------|----------|
| `ClientService` | `CreateClient` | `CreateClientRequest` | `CreateClientResponse` |
| `ClientService` | `GetClient` | `GetClientRequest` | `GetClientResponse` |
| `ClientService` | `ListClients` | `ListClientsRequest` | `ListClientsResponse` |
| `ClientService` | `UpdateClient` | `UpdateClientRequest` | `UpdateClientResponse` |
| `ClientService` | `DeactivateClient` | `DeactivateClientRequest` | `DeactivateClientResponse` |
| `PlatformService` | `CreatePlatform` | `CreatePlatformRequest` | `CreatePlatformResponse` |
| `PlatformService` | `GetPlatform` | `GetPlatformRequest` | `GetPlatformResponse` |
| `PlatformService` | `ListPlatforms` | `ListPlatformsRequest` | `ListPlatformsResponse` |
| `IntegrationService` | `CreateIntegration` | `CreateIntegrationRequest` | `CreateIntegrationResponse` |
| `IntegrationService` | `GetIntegration` | `GetIntegrationRequest` | `GetIntegrationResponse` |
| `IntegrationService` | `DeleteIntegration` | `DeleteIntegrationRequest` | `DeleteIntegrationResponse` |
| `IntegrationService` | `ProcessWebhookEvent` | `ProcessWebhookEventRequest` | `ProcessWebhookEventResponse` |

### 7.2 Transactions Service — `transactionsgrpc`

| Service | RPC | Request | Response |
|---------|-----|---------|----------|
| `MerchantService` | `CreateMerchant` | `CreateMerchantRequest` | `CreateMerchantResponse` |
| `MerchantService` | `GetMerchant` | `GetMerchantRequest` | `GetMerchantResponse` |
| `MerchantService` | `ListMerchants` | `ListMerchantsRequest` | `ListMerchantsResponse` |
| `CustomerService` | `CreateCustomer` | `CreateCustomerRequest` | `CreateCustomerResponse` |
| `CustomerService` | `GetCustomer` | `GetCustomerRequest` | `GetCustomerResponse` |
| `DepositService` | `InitiateDeposit` | `CreateDepositRequest` | `CreateDepositResponse` |
| `DepositService` | `GetDeposit` | `GetDepositRequest` | `GetDepositResponse` |
| `PayoutService` | `RequestPayout` | `CreatePayoutRequest` | `CreatePayoutResponse` |
| `PayoutService` | `GetPayout` | `GetPayoutRequest` | `GetPayoutResponse` |

### 7.3 Cross-Service RPC Usage

Cross-service communication occurs only through gRPC contracts:

- Transactions Service calls `clientsgrpc.ClientService.GetClient` to validate
  `client_id` on deposit and payout requests.
- Clients Service calls `transactionsgrpc` query RPCs for client and
  administrator transaction monitoring.

## 8. Versioning Strategy

The existing contracts are unversioned (`depositsgrpc`, `integrationsgrpc`).
The new architecture preserves this convention for the initial contracts:

- `clientsgrpc`
- `transactionsgrpc`
- `commongrpc`

Guidelines:

- **Additive changes** (new fields, new RPCs, new enum values) are made in
  place. Field numbers and enum numeric values are never reused.
- **Breaking changes** (removing/renaming fields, changing types, changing
  semantics) require a new versioned package, e.g. `clientsgrpc.v2`,
  `transactionsgrpc.v2`, `commongrpc.v2`.
- Removed enum values are reserved, never deleted.
- Removed field numbers are reserved, never reused.
- A contract is frozen once deployed; evolution happens by addition or by a new
  versioned package.

## 9. HTTP Gateway Strategy

The existing `google.api.http` annotations are preserved as the gateway
strategy.

- Every public RPC carries a `google.api.http` annotation.
- REST paths remain under `/v1/public/...` for public endpoints.
- Administrative endpoints use `/v1/admin/...` when introduced.
- The gateway generates `.pb.gw.go` stubs into `grpc/go/<package>/` alongside
  the gRPC stubs.
- Business logic is never duplicated between gRPC and REST; the gateway
  forwards to the same service implementation.

## 10. Generation Workflow

The existing `protobuf/Makefile` workflow is preserved and extended.

```text
protobuf/
├── common.proto          # Shared enums and messages (commongrpc)
├── clients.proto         # Clients Service contract (clientsgrpc)
├── transactions.proto    # Transactions Service contract (transactionsgrpc)
├── Makefile              # protobuf lint and Go code generation targets
└── README.md
```

Steps:

1. Edit the `.proto` source in `protobuf/`.
2. Run `make lint` (clang-format dry-run over sources).
3. Run `make generate-protos` from `protobuf/`.
4. Generated Go code is written to `grpc/go/<package>/`:
   - `grpc/go/commongrpc/common.pb.go`
   - `grpc/go/clientsgrpc/clients.pb.go`, `clients_grpc.pb.go`,
     `clients.pb.gw.go`
   - `grpc/go/transactionsgrpc/transactions.pb.go`,
     `transactions_grpc.pb.go`, `transactions.pb.gw.go`
5. Review and commit the generated output with the `.proto` change.

Prerequisites (unchanged):

- `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`
- `third_party/googleapis` submodule for `google/api/annotations.proto`

Generated files are never edited by hand.

## 11. Migration Notes

1. **Create `common.proto` first.** Shared enums and `Money` are the foundation
   for both service contracts.
2. **Create `clients.proto` and `transactions.proto`** alongside the existing
   contracts. Do not delete `deposits.proto` or `integrations.proto` until
   callers migrate.
3. **Regenerate `grpc/go/`** for the new packages. Existing generated packages
   (`depositsgrpc`, `integrationsgrpc`) remain until their services are
   replaced.
4. **Migrate callers in stages.** Keep the current services runnable while the
   new contracts are adopted.
5. **Retire legacy contracts** (`deposits.proto`, `integrations.proto`) only
   after the Deposits and Integrations services are fully replaced by the
   Transactions and Clients services.

## 12. Assumptions

- The existing `depositsgrpc` and `integrationsgrpc` packages are superseded by
  `transactionsgrpc` and `clientsgrpc` respectively.
- `commongrpc` is a new shared package; no existing contract is modified to
  extract shared types until the new contracts are introduced.
- The `go_package` option always matches the repository Go module
  (`github.com/I-Frostbyte/rvpay-go`).
- Public REST paths remain under `/v1/public/...`; administrative paths are
  introduced under `/v1/admin/...` when needed.

## 13. Unresolved Questions

- **Enum value migration** — whether existing `DepositProvider` values
  (`MTN_MOMO_CMR`, `ORANGE_MOMO_CMR`) map 1:1 to the shared `Provider` values
  (`MTN_MOMO`, `ORANGE_MOMO`) or require a compatibility mapping layer.
- **Integration message shape** — whether `platform_id` and `client_id` are
  added to the existing `Integration` message or introduced as a new message in
  `clients.proto`.
- **Webhook event message** — whether `ProcessWebhookEventRequest` evolves in
  place or is replaced by a typed webhook event message in `clients.proto`.
- **Pagination** — whether `List*` RPCs require pagination fields in the
  initial contracts or defer them to a later additive change.
- **Admin API versioning** — whether administrative RPCs share the same
  package as public RPCs or use a separate versioned package.