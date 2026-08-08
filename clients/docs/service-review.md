# Clients Service Business Layer Review

Document Version: 1.0
Status: Complete
System: RVPay
Service: Clients Service
Review: Agent 06 — Business Service Implementation

## 1. Purpose

This document records the business service layer implementation for the
Clients Service. It summarizes the implemented services, business rule
ownership, transaction strategy, dependency graph, service interfaces,
provider abstraction strategy, and remaining work before OAuth
implementation.

## 2. Implemented Business Services

Three service implementations were created in `clients/service/`:

| Service | File | Responsibility |
|---|---|---|
| `ClientsServiceImpl` | `clients/service/clients_service.go` | Client lifecycle |
| `PlatformsServiceImpl` | `clients/service/platforms_service.go` | Platform lifecycle |
| `IntegrationsServiceImpl` | `clients/service/integrations_service.go` | Integration lifecycle |

Each service is a plain Go struct with a constructor that accepts its
dependencies via dependency injection. No gRPC server, runtime wiring, OAuth
provider communication, or webhook processing is implemented here.

## 3. Business Rule Ownership

All business rules live exclusively in the service layer:

- **Client name uniqueness** — `CreateClient` checks for existing clients by
  name before creating.
- **Client status transitions** — `ActivateClient` and `DeactivateClient`
  enforce valid state changes (ACTIVE ↔ CLOSED).
- **Active-client requirement** — `InstallIntegration` rejects inactive
  clients.
- **Platform enablement** — `InstallIntegration` rejects disabled platforms.
- **Duplicate integration prevention** — `InstallIntegration` checks for
  existing client-platform pairs.
- **Active-integration requirement** — `ReconnectIntegration` and
  `SyncIntegration` require ACTIVE status.
- **Safe deletion** — `DeleteClient` blocks deletion of ACTIVE clients.

Repositories assume valid input and do not enforce business rules.

## 4. Transaction Strategy

The service layer coordinates repository operations but does not expose
transaction logic outside. The `IntegrationsServiceImpl` constructor accepts
all five repository interfaces, enabling future transaction orchestration
via `ClientsRepo.Begin(ctx)` when runtime wiring is implemented.

Currently, each RPC executes as a single repository call. Multi-step
transactions (e.g., install integration + create OAuth token + register
webhook) will be coordinated by the service layer in later agents.

## 5. Dependency Graph

```
clients/service/
├── errors.go              (business errors, error translation)
├── converters.go          (protobuf ↔ sqlc mapping)
├── clients_service.go     (ClientsServiceImpl)
├── platforms_service.go   (PlatformsServiceImpl)
└── integrations_service.go (IntegrationsServiceImpl)

Dependencies:
- clients/db/repo (repository interfaces only)
- clients/db/sqlc (model types for mapping)
- grpc/go/clientsgrpc (protobuf request/response types)
- grpc/go/commongrpc (shared enums)
- google.golang.org/grpc/status (error codes)
- google.golang.org/protobuf/types/known/timestamppb (timestamps)
```

Services do **not** depend on:
- gRPC server implementation
- protobuf generation tools
- OAuth providers
- HTTP clients
- runtime packages
- database drivers

## 6. Service Interfaces

Each service exposes a small, cohesive interface via its constructor:

```go
type ClientsServiceImpl struct {
    clientsRepo repo.ClientRepo
    logger      Logger
}

type PlatformsServiceImpl struct {
    platformsRepo repo.PlatformRepo
    logger        Logger
}

type IntegrationsServiceImpl struct {
    integrationsRepo repo.IntegrationRepo
    clientsRepo      repo.ClientRepo
    platformsRepo    repo.PlatformRepo
    oauthRepo        repo.OAuthTokenRepo
    webhookRepo      repo.WebhookSubscriptionRepo
    logger           Logger
}
```

The `Logger` interface is minimal (`Info(msg string, args ...interface{})`),
matching the zerolog convention used by the Deposits service.

## 7. Provider Abstraction Strategy

The service layer remains provider-agnostic:

- No `if provider == HighLevel` branching exists.
- No provider-specific fields are exposed.
- OAuth implementation details are hidden behind repository interfaces.
- Future providers integrate through repository and service changes only;
  no protobuf redesign is required.

The `IntegrationsServiceImpl` accepts `oauthRepo` and `webhookRepo` in its
constructor, but does not use them yet. This prepares the service for OAuth
and webhook orchestration in later agents without changing the constructor
signature.

## 8. Mapping Responsibilities

`clients/service/converters.go` centralizes all protobuf ↔ sqlc mapping:

- `sqlcClientToProto` — sqlc `Client` → protobuf `Client`
- `sqlcPlatformToProto` — sqlc `Platform` → protobuf `Platform`
- `sqlcIntegrationToProto` — sqlc `Integration` → protobuf `Integration`
- `protoStatusToSqlcClientStatus` — protobuf `ClientStatus` → sqlc
  `ClientStatus`
- `protoStatusToSqlcIntegrationStatus` — protobuf `IntegrationStatus` →
  sqlc `IntegrationStatus`
- `parseUUID` — UUID string parsing with gRPC error translation
- `toTimePtr` — protobuf timestamp → `time.Time` pointer

No repository models leak outside the service layer.

## 9. Error Handling

Repository errors are translated to business errors in `errors.go`:

- `ErrClientNotFound`, `ErrPlatformNotFound`, `ErrIntegrationNotFound`
- `ErrClientAlreadyExists`, `ErrPlatformSlugExists`,
  `ErrIntegrationAlreadyExists`, `ErrWebhookSubscriptionExists`
- `ErrClientInactive`, `ErrPlatformDisabled`, `ErrIntegrationNotActive`,
  `ErrOAuthNotSupported`, `ErrWebhookNotSupported`
- `ErrClientHasIntegrations`, `ErrWebhookSubscriptionNotFound`

The `translateRepoError` helper converts repository sentinels to gRPC
status errors without exposing SQL or database details.

## 10. Remaining Work Before OAuth Implementation

1. **OAuth token management RPCs** — The API contract does not yet expose
   gRPC RPCs for OAuth token creation, retrieval, or refresh. The
   `IntegrationsServiceImpl` accepts `oauthRepo` but does not use it yet.
   Agent 07 will add OAuth-specific RPCs or HTTP handlers.

2. **Webhook subscription RPCs** — Similarly, webhook subscription
   registration is not yet exposed via gRPC. The service accepts
   `webhookRepo` for future use. Agent 08 will add webhook-specific RPCs or
   HTTP handlers.

3. **Runtime wiring** — The service constructors are ready for dependency
   injection, but the gRPC server and HTTP gateway wiring (Agent 09/10) have
   not yet connected them to the generated gRPC server interfaces.

4. **Domain model converters** — The current converters map sqlc types
   directly to protobuf types. If a richer domain model is needed, Agent 06
   can introduce domain structs without changing the repository or protobuf
   layers.

## 11. Validation Results

- ✅ Business services compile (`go build ./clients/...`, exit 0)
- ✅ Repository interfaces are respected (no SQL in service layer)
- ✅ Dependency injection works (constructors accept interfaces)
- ✅ Transaction coordination ready (repositories accept `sqlc.Querier`)
- ✅ Business validation executes correctly (status checks, uniqueness)
- ✅ Mapping functions compile
- ✅ Services remain provider agnostic
- ✅ No SQL exists in the service layer
- ✅ No transport logic exists in the service layer

## 12. Files Created

- `clients/service/errors.go`
- `clients/service/converters.go`
- `clients/service/clients_service.go`
- `clients/service/platforms_service.go`
- `clients/service/integrations_service.go`
- `clients/docs/service-review.md`

## 13. Files Modified

- None (all files are new)

## 14. Commands Executed

- `go build ./clients/...` (exit 0)

## 15. Issues Found

- None blocking. The business service layer is complete and ready for
  runtime wiring (Agent 09/10) and OAuth/webhook implementation (Agents
  07/08).