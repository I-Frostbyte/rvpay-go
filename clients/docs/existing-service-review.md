# Existing Integrations Service — Migration Assessment for Clients Service

Document Version: 1.0
Status: Analysis
Source: `integrations/` service review
Target: Clients Service (per `docs/domain-model.md`)

## 1. Purpose

This document reviews the existing Integrations service and produces a
migration assessment for the future Clients Service. It is analysis only; no
source code is modified.

## 2. Current Architecture

### 2.1 Package Layout

```text
integrations/
├── cmd/grpc-service/main.go      # gRPC + HTTP gateway server bootstrap and shutdown
├── config/model.go               # Config and DBConfig environment bindings
├── db/
│   ├── doc.go                    # go:generate directives (sqlc, mockgen)
│   ├── migrations/               # 000001: integrations + webhook_events tables
│   ├── query/                    # integrations.sql, webhook_events.sql
│   ├── repo/                     # IntegrationsRepo interface, Impl, Migrate/MigrateDown
│   └── sqlc/                     # sqlc-generated data access code (not inspected)
├── integrations/
│   ├── service.go                # IntegrationService gRPC implementation
│   ├── converters.go             # sqlc ↔ gRPC converters
│   └── errors.go                 # ErrIntegrationNotFound
├── oauth/
│   ├── service.go                # HighLevel OAuth token exchange + AES-GCM encryption
│   └── handler.go                # GET /oauth/callback
├── webhook/
│   ├── service.go                # HighLevel webhook parse/store/async processing
│   └── handler.go                # POST /webhooks/highlevel
├── README.md
├── Makefile
├── Dockerfile
└── .env.example
```

### 2.2 Runtime Flow

- `main.go` loads config (ardanlabs/conf/v3 + godotenv), parses log level,
  connects and pings PostgreSQL, runs migrations, constructs the repo, the
  IntegrationService, the OAuth service/handler, and the webhook
  service/handler.
- Registers gRPC with reflection, health, and a unary panic-recovery
  interceptor; registers the gRPC-Gateway; serves `/healthz`,
  `/oauth/callback`, and `/webhooks/highlevel`; graceful shutdown on
  SIGINT/SIGTERM.

### 2.3 Database

- `integrations` table: `id`, `provider`, `location_id`, `access_token`,
  `refresh_token`, `token_expires_at`, `created_at`, `updated_at`.
- `webhook_events` table: `id`, `provider`, `event_type`, `payload`,
  `processed`, `created_at`.
- Migrations via golang-migrate; sqlc generates data access code; `db/repo`
  exposes `Do()` and `Begin()`.

### 2.4 API Surface

- gRPC `IntegrationService`: `CreateIntegration`, `GetIntegration`,
  `DeleteIntegration`; `ProcessWebhookEvent` is declared in the protobuf but
  not implemented in the service.
- REST gateway: `POST/GET/DELETE /v1/public/integrations[...]`,
  `POST /v1/public/webhooks`.

## 3. Existing Functionality

### 3.1 What Already Exists

- Integration CRUD (create, get by ID, delete) over gRPC + REST.
- HighLevel OAuth callback flow (code exchange, token storage, AES-GCM
  encryption at rest).
- HighLevel webhook ingestion (payload parsing, event storage, async
  processing stub).
- Token lifecycle (access/refresh tokens with expiry).
- gRPC + HTTP gateway runtime with health checks and graceful shutdown.
- Repository pattern (interface + pgxpool-backed impl) and sqlc workflow.
- Environment-backed configuration.

### 3.2 What Should Remain Unchanged

- The repository pattern (`IntegrationsRepo` interface with `Do()`/`Begin()`).
- The sqlc workflow (migrations → query → generate).
- The configuration loading pattern (`LoadConfig`, ardanlabs/conf/v3 +
  godotenv).
- The runtime bootstrap pattern (config → logger → DB → migrations → services
  → gRPC + gateway → graceful shutdown).
- The gRPC + HTTP gateway wiring, health endpoint, and recovery middleware.
- The token encryption approach (AES-GCM with nonce prefix).
- The converter pattern (`sqlc*ToGrpc`, `grpc*ToSqlc`).
- The gRPC error mapping style (`status.Error` with gRPC codes).
- The zerolog logging style.

### 3.3 What Belongs in the Future Clients Service

Everything in the Integrations service belongs in the Clients Service (the
service evolves into the Clients Service per `docs/domain-model.md`):

- Integration entity (evolves to link `platform_id` and `client_id`).
- Webhook event persistence.
- OAuth token storage.

## 4. Reusable Code

| Component | Verdict | Notes |
|-----------|---------|-------|
| Runtime bootstrap (`main.go`) | Reuse unchanged | Canonical pattern; copy into `clients/cmd/grpc-service/` |
| Config loading (`config/model.go`) | Needs refactor | HighLevel fields become provider-scoped; add `Platform`/`Client` concerns later |
| `db/repo` (IntegrationsRepo, Migrate) | Reuse unchanged | Generalise name to ClientsRepo |
| sqlc workflow (`db/doc.go`, sqlc.yaml) | Reuse unchanged | Add clients/platforms queries |
| Integration repository queries | Reuse unchanged | Evolve the `integrations` table with `platform_id`/`client_id` |
| `integrations/service.go` CRUD | Reuse unchanged | Keep CRUD; implement `ProcessWebhookEvent` |
| `integrations/converters.go` | Reuse unchanged | Extend for new fields |
| `errors.go` | Needs refactor | `ErrIntegrationNotFound` is declared but unused; wire into service or delete |
| OAuth service + handler | Needs refactor | HighLevel-specific; abstract provider boundary |
| Webhook service + handler | Needs refactor | HighLevel-specific; enforce signature verification; implement async processing |
| Token encryption (`encryptToken`) | Reuse unchanged | AES-GCM; may need `decryptToken` for token refresh |
| Dockerfile, Makefile, .env.example | Reuse unchanged | Copy into `clients/` |

## 5. Duplicated Logic

- **Http gateway registration** — duplicated conceptually with deposits
  service; candidate for `shared/` in the future but not urgent.
- **`getPostgresConnectionURL`** — duplicated across services; candidate for
  `shared/` database helpers.
- **Provider string literal** — `"highlevel"` appears in `oauth/service.go`
  and `webhook/service.go`; should be a single constant.
- **Token encryption** — currently only `encryptToken` exists; a matching
  `decryptToken` will be needed for token refresh and is not yet implemented.

## 6. Technical Debt

| Debt | Impact |
|------|--------|
| `ProcessWebhookEvent` gRPC method unimplemented | Public contract promises a method that returns `Unimplemented` |
| `HIGHLEVEL_SSO_KEY` not enforced | Webhook signature verification missing; any caller can POST forged events |
| `processEvent` is a stub with a TODO | Async processing does nothing after storage |
| Hard-coded provider (`"highlevel"`) in multiple places | Adding a provider requires touching service code |
| No `platform_id`/`client_id` on integrations | Does not match the domain model; Integration links Platform and Client |
| No lifecycle state on integrations | Domain model requires Created → OAuth pending → Active → Revoked |
| `ErrIntegrationNotFound` declared but unused | Dead code; service uses `status.Error(codes.NotFound, ...)` directly |
| No token refresh / decrypt | OAuth tokens expire; no re-auth flow implemented |
| No `List`/`Update` integration RPCs | Clients Service CRUD may need them |
| Webhook handler returns 400 for internal storage errors | Should distinguish malformed payload (400) from internal failure (500) |
| HighLevel config fields not validated at load | Misconfiguration surfaces only at OAuth runtime |
| `PORT` read via `os.Getenv` in main | Inconsistent with the config struct pattern |

## 7. Package Boundaries

Current package boundaries are clean and match the repository template:

- `config` — configuration only.
- `db` — database layer (migrations, query, repo, sqlc).
- `integrations` — integration domain logic + gRPC service.
- `oauth` — OAuth flow.
- `webhook` — webhook flow.
- `cmd/grpc-service` — process entry point.

Boundaries requiring change for the Clients Service:

- `integrations/` (domain package) becomes `clients/` with sub-packages for
  Client, Platform, and Integration domains (per `docs/repository-layout.md`).
- `oauth/` becomes provider-agnostic with a HighLevel implementation.
- `webhook/` becomes provider-agnostic with a HighLevel implementation.

## 8. Dependency Boundaries

Current dependencies:

- Service → `repo.IntegrationsRepo` (interface) → sqlc → pgxpool → PostgreSQL.
- OAuth/webhook services → same repo interface.
- gRPC service embeds `UnimplementedIntegrationServiceServer`.
- No cross-service gRPC dependencies exist yet (correct for now).

Desired dependency boundaries for Clients Service:

- Domain packages depend only on the repo interface, never directly on sqlc or
  pgxpool (converter functions currently bridge gRPC ↔ sqlc inside the domain
  package — keep this pattern).
- Provider-specific OAuth/webhook implementations implement provider-agnostic
  interfaces.
- No service package imports another service's package (per
  `docs/protobuf-strategy.md`).

## 9. HighLevel-Specific Code

| Location | HighLevel-specific detail |
|----------|---------------------------|
| `oauth/service.go` | `highLevelTokenURL` constant; token response JSON shape; grant flow; provider string `"highlevel"` |
| `oauth/handler.go` | Endpoint route `/oauth/callback` (generic path, HighLevel-specific handler) |
| `webhook/service.go` | `highLevelProvider` constant; `highLevelWebhookPayload` JSON shape (`type`/`eventType`/`data`) |
| `webhook/handler.go` | Route `/webhooks/highlevel` |
| `config/model.go` | `HighLevelClientID`, `HighLevelClientSecret`, `HighLevelRedirectURL`, `HighLevelSSOKey` |

## 10. Where Provider-Agnostic Abstractions Should Exist

- **OAuth**: an interface (e.g. `OAuthProvider`) with methods to build the
  authorization URL, exchange a code, and refresh tokens. HighLevel implements
  it. The handler stays generic and dispatches by provider.
- **Webhook**: an interface (e.g. `WebhookVerifier`) for signature
  verification, plus a generic event store. HighLevel implements the verifier.
- **Config**: provider config should be grouped (e.g. `HighLevelConfig`
  struct) so a provider registry can be built without changing the top-level
  `Config` shape.
- **Integration routing**: routes like `/oauth/callback` and
  `/webhooks/highlevel` should resolve the provider from the request
  (query/path) rather than hard-coding one handler.

## 11. Migration Strategy — Categorization

### Can Reuse Unchanged

- `db/repo` pattern (`Do()`/`Begin()`, `Migrate`/`MigrateDown`).
- sqlc workflow and config.
- Runtime bootstrap + gRPC/gateway wiring + recovery + health + shutdown.
- Token encryption (AES-GCM).
- Converter pattern.
- CRUD service methods (Create/Get/Delete).
- Dockerfile, Makefile, .env.example.

### Needs Refactor

- `oauth/` — extract provider-agnostic OAuth interface; HighLevel as an
  implementation; add token refresh/decrypt.
- `webhook/` — extract provider-agnostic verifier; enforce SSO signature;
  implement real async processing.
- `config` — group provider-scoped fields; add validation.
- `integrations` domain — evolve into Client/Platform/Integration domains;
  add `platform_id`/`client_id`; add lifecycle state.
- `main.go` wiring — register additional services (Client, Platform) and
  implement `ProcessWebhookEvent`.
- Database migration — add `platform_id`, `client_id`, state to
  `integrations`; add `clients` and `platforms` tables.

### Needs Replacement

- `integrations.proto`/`integrationsgrpc` — replaced by `clients.proto` /
  `clientsgrpc` (per `docs/protobuf-strategy.md`).
- `Integration` message — evolves to link `platform_id` and `client_id`.
- `ProcessWebhookEventRequest` — may evolve into a typed message.

### Needs Deletion

- `ErrIntegrationNotFound` if it remains unused (or wire it in).
- Duplicated `getPostgresConnectionURL` when `shared/` database helpers are
  extracted.
- Hard-coded provider string literals once provider abstractions exist.

## 12. Migration Risks

| Risk | Mitigation |
|------|------------|
| Breaking the existing HighLevel OAuth/webhook flows during refactor | Keep HighLevel behavior identical; add tests before refactor |
| Adding `platform_id`/`client_id` changes the data model | New migration is additive (new nullable/required columns with backfill where needed) |
| Provider-agnostic abstraction over-engineering | Only abstract what today has two or more concrete needs; otherwise keep concrete |
| `ProcessWebhookEvent` implementation changes webhook semantics | Preserve the existing HTTP path; gRPC method delegates to the same service logic |
| Token refresh/decrypt introduces security regressions | Reuse the existing AES-GCM pattern; add unit tests |
| Circular imports between domain packages | Keep domain packages sibling; shared types live in `commongrpc` |

## 13. Suggested Migration Order

1. **Preserve current behavior** — snapshot existing tests; ensure the
   Integrations service still builds and passes tests.
2. **Scaffold `clients/` service** — copy the deposits template structure into
   `clients/` (per `docs/repository-layout.md`).
3. **Port infrastructure unchanged** — config pattern, repo pattern, sqlc
   workflow, runtime bootstrap, Dockerfile, Makefile.
4. **Introduce `common.proto` and `clients.proto`** — add shared enums/messages
   and the Clients Service contract alongside existing protos.
5. **Migrate integration CRUD** — port `CreateIntegration`, `GetIntegration`,
   `DeleteIntegration`; new `clientsgrpc` service.
6. **Migrate OAuth with provider abstraction** — extract OAuth interface;
   HighLevel implementation; handler dispatch by provider.
7. **Migrate webhooks with verification** — extract verifier interface;
   enforce SSO signature; implement async processing.
8. **Implement `ProcessWebhookEvent`** — gRPC method delegates to the webhook
   service.
9. **Evolve the data model** — new migrations for `clients`, `platforms`, and
   `integrations` columns (`platform_id`, `client_id`, state).
10. **Retire the Integrations service in stages** — keep it runnable until the
    Clients Service is verified end-to-end.

## 14. Estimated Implementation Complexity

| Area | Complexity | Rationale |
|------|------------|-----------|
| Scaffold `clients/` service | Low | Copy deposits template |
| Port repo/sqlc/config/runtime | Low | Copy unchanged patterns |
| Port integration CRUD | Low | Mechanical port + new proto |
| Provider abstraction (OAuth) | Medium | Interface extraction; HighLevel impl unchanged |
| Provider abstraction (webhook) | Medium | Interface extraction; signature enforcement |
| Webhook async processing | Medium | New processing logic; store/process semantics |
| `ProcessWebhookEvent` method | Low | Delegates to existing webhook service |
| Data model evolution | Medium | Additive migrations; codegen via sqlc |
| Full replacement of `integrationsgrpc` | Medium | Proto + generated code + callers |

Total: **Low-to-Medium** overall. The service is small, follows the canonical
repository pattern, and most migration work is movement plus additive change
rather than rewrite.

## 15. Package Boundary Recommendations

```text
clients/
├── cmd/grpc-service/     # bootstrap (reuse unchanged pattern)
├── config/               # grouped provider config + validation
├── db/
│   ├── migrations/       # new clients/platforms migrations + integrations evolution
│   ├── query/            # clients.sql, platforms.sql, integrations.sql, webhook_events.sql
│   ├── repo/             # ClientsRepo (renamed from IntegrationsRepo)
│   └── sqlc/             # generated
├── clients/              # ClientService implementation
├── platforms/            # PlatformService implementation
├── integrations/         # IntegrationService implementation (ported CRUD + ProcessWebhookEvent)
├── oauth/                # provider-agnostic OAuth + highlevel/ implementation
├── webhook/              # provider-agnostic webhook + highlevel/ implementation
├── README.md, Makefile, Dockerfile, .env.example
```

## 16. Summary

The Integrations service is well-structured and follows the repository's
canonical patterns. Its core CRUD, repository, sqlc, configuration, and
runtime wiring can be reused almost unchanged. The main refactoring work is

- extracting provider-agnostic OAuth/webhook abstractions,
- evolving the data model to link Clients and Platforms,
- implementing the unimplemented `ProcessWebhookEvent` and real async webhook
  processing,
- enforcing webhook signature verification,
- and migrating the protobuf contract to `clientsgrpc`.

No existing source code was modified as part of this analysis.