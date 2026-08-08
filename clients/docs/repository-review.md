# Clients Service Repository Review

Document Version: 1.0
Status: Complete
System: RVPay
Service: Clients Service
Review: Agent 05 — Repository Implementation

## 1. Purpose

This document records the repository layer implementation for the Clients
Service. It summarizes the implemented repositories, their dependency graph,
transaction strategy, interface design, future extensibility, and remaining
risks before service implementation begins.

## 2. Implemented Repositories

Five per-aggregate repositories were implemented, each backed by the sqlc
`Querier` interface:

| Repository | File | Aggregate |
|---|---|---|
| `ClientRepo` | `clients/db/repo/client_repo.go` | Clients |
| `PlatformRepo` | `clients/db/repo/platform_repo.go` | Platforms |
| `IntegrationRepo` | `clients/db/repo/integration_repo.go` | Integrations |
| `OAuthTokenRepo` | `clients/db/repo/oauth_token_repo.go` | OAuth Tokens |
| `WebhookSubscriptionRepo` | `clients/db/repo/webhook_subscription_repo.go` | Webhook Subscriptions |

Each repository exposes a small, cohesive interface with CRUD operations:

- **Create** — insert a new record.
- **GetByID / FindByID** — fetch a single record by primary key.
- **GetByName / GetBySlug / GetByExternalAccountID / GetByIntegrationIDAndEndpoint** — lookup by natural keys.
- **List / ListActive / ListByClient / ListByPlatform / ListByIntegrationID** — paginated list operations.
- **Count** — aggregate counts.
- **Exists / ExistsByID / ExistsBySlug / ExistsByClientAndPlatform / ExistsByIntegrationID / Exists** — existence checks.
- **Update / UpdateStatus / UpdateLastSyncAt / UpdateLastDelivery** — partial updates.
- **Delete / DeleteByIntegrationID** — deletion.

### Error Handling

A shared `errors.go` provides sentinel errors and a `wrapError` helper that
translates raw PostgreSQL errors into repository-level errors without exposing
SQL implementation details:

- `ErrNotFound` — `pgx.ErrNoRows` (no rows returned).
- `ErrDuplicate` — PostgreSQL `23505` (unique violation).
- `ErrConstraint` — PostgreSQL `23503` (foreign key violation) and `23514`
  (check violation).
- Connection failures are returned as-is (wrapped by the caller).

No raw PostgreSQL errors or SQL strings are returned to callers.

### Logging

No logging is performed in the repository layer. The Deposits repository does
not log database operations, so this convention is preserved. Sensitive data
(OAuth tokens, refresh tokens, provider secrets, encrypted payloads, client
secrets) is never logged.

## 3. Dependency Graph

```
clients/db/repo/
├── errors.go              (sentinel errors, error translation)
├── client_repo.go         (ClientRepo interface + impl)
├── platform_repo.go       (PlatformRepo interface + impl)
├── integration_repo.go    (IntegrationRepo interface + impl)
├── oauth_token_repo.go    (OAuthTokenRepo interface + impl)
├── webhook_subscription_repo.go (WebhookSubscriptionRepo interface + impl)
└── repo.go                (ClientsRepo pool wrapper, Begin/Do, Migrate)

clients/db/
├── doc.go                 (go:generate directives)
├── sqlc/                  (generated sqlc code — not modified)
├── sqlc/mocks/            (generated sqlc.Querier mock — not modified)
└── repo/mocks/            (generated repository mocks)
```

Dependencies:

- Repositories depend on `clients/db/sqlc` (generated queries) and
  `github.com/jackc/pgx/v5` (for `pgx.ErrNoRows` and `pgconn.PgError`).
- Repositories do **not** depend on gRPC, protobuf, OAuth providers, HTTP
  clients, service implementations, or runtime packages.
- The existing `ClientsRepo` interface (pool wrapper with `Begin`/`Do`) is
  preserved unchanged. Its generated mock remains compatible.

## 4. Transaction Strategy

The transaction strategy mirrors the Deposits service:

- `ClientsRepo.Begin(ctx)` returns a `sqlc.Querier` bound to a `pgx.Tx` and
  the `pgx.Tx` handle.
- `ClientsRepo.Do()` returns the pool-bound `sqlc.Querier`.
- Per-aggregate repository constructors accept a `sqlc.Querier`, so they work
  transparently with either the pool or a transaction:

```go
// Pool mode:
clientRepo := repo.NewClientRepo(clientsRepo.Do())

// Transaction mode:
queries, tx, err := clientsRepo.Begin(ctx)
clientRepo := repo.NewClientRepo(queries)
// ... multiple repository operations ...
tx.Commit(ctx)
```

This allows the service layer (Agent 06) to compose multi-aggregate
transactions without the repositories knowing whether they are operating in a
transaction or not.

## 5. Interface Design

- Each aggregate has its own small, cohesive interface (no "god interface").
- Interfaces are split by aggregate: Client, Platform, Integration, OAuth
  Token, Webhook Subscription.
- Each interface is independently mockable via `mockgen`.
- Constructors return the interface type, enabling dependency injection and
  test substitution.
- No business rules exist in the repository layer — only persistence and
  retrieval.

## 6. Future Extensibility

- New queries can be added to the `.sql` files and regenerated with sqlc
  without changing existing repository interfaces.
- New repository methods can be added to interfaces without breaking existing
  consumers (Go interfaces are satisfied structurally).
- The `wrapError` helper can be extended to handle additional PostgreSQL error
  codes as needed.
- The `ClientsRepo` pool wrapper is unchanged, so existing transaction and
  migration workflows remain stable.

## 7. Remaining Risks Before Service Implementation

1. **OAuth and webhook registration RPCs** — The API contract (reviewed in
   Agent 04.5) does not yet expose gRPC RPCs for OAuth token management or
   webhook subscription registration. The repository layer supports these
   entities, but the service layer (Agent 06) will need to decide whether to
   expose them as gRPC RPCs or handle them via HTTP handlers (as the existing
   integrations service does). This is an API-level decision, not a
   repository-level one.

2. **Domain model conversion** — The repositories return sqlc model types
   directly. The service layer (Agent 06) will need converters to map these
   to gRPC response types, following the pattern established by the
   integrations service (`integrations/integrations/converters.go`).

3. **Mock generation for aggregate repos** — Mocks for the five new
   repository interfaces have been generated via `go generate`. The tests
   agent (Agent 11) should verify mock coverage and add test doubles as
   needed.

## 8. Validation Results

- ✅ Repositories compile (`go build ./clients/...`, exit 0)
- ✅ Constructors compile
- ✅ Interfaces compile
- ✅ sqlc integration compiles
- ✅ `go vet ./clients/...` passes (exit 0)
- ✅ Mock generation succeeds (`go generate ./clients/db/...`, exit 0)
- ✅ Generated mocks remain compatible (existing `ClientsRepo` mock unchanged)
- ✅ No SQL is duplicated outside sqlc
- ✅ No business rules in repository layer
- ✅ sqlc remains encapsulated
- ✅ Repositories remain provider-agnostic
- ✅ Repositories remain unit testable

## 9. Files Created

- `clients/db/repo/errors.go`
- `clients/db/repo/client_repo.go`
- `clients/db/repo/platform_repo.go`
- `clients/db/repo/integration_repo.go`
- `clients/db/repo/oauth_token_repo.go`
- `clients/db/repo/webhook_subscription_repo.go`
- `clients/db/repo/mocks/client_repo.go` (generated)
- `clients/db/repo/mocks/platform_repo.go` (generated)
- `clients/db/repo/mocks/integration_repo.go` (generated)
- `clients/db/repo/mocks/oauth_token_repo.go` (generated)
- `clients/db/repo/mocks/webhook_subscription_repo.go` (generated)
- `clients/docs/repository-review.md`

## 10. Files Modified

- `clients/db/doc.go` — added `go:generate` directives for the five new
  repository interface mocks.

## 11. Commands Executed

- `go generate ./clients/db/...` (exit 0)
- `go build ./clients/...` (exit 0)
- `go vet ./clients/...` (exit 0)

## 12. Issues Found

- None blocking. The repository layer is complete and ready for service
  implementation.