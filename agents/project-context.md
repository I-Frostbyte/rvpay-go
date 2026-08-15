# Project Context

Backend language:
- Go (go 1.26.5 per go.mod)

Architecture:
- gRPC Microservices (monorepo)

Database:
- PostgreSQL

<!-- IGNORE THIS
Migration tool:
- goose
-->

Migration tool (actual):
- golang-migrate (`github.com/golang-migrate/migrate/v4`), NOT goose

Query generation:
- sqlc

Logging:
- zerolog (structured JSON)

Deployment targets:
- Render (primary; active Blueprint)
- Oracle Cloud Always Free / OCI (docker-compose; GitHub Actions pipeline disabled)

Containerization:
- Docker (multi-stage distroless images per service)
- Docker Compose (OCI stack)

CI:
- GitHub Actions (render-deploy.yml active; deploy.yml disabled)

Goals:
- Keep infrastructure simple.
- Prefer ARM64-compatible images.
- Minimize running costs.
- Preserve clean architecture.

Principles:
- Preserve clean architecture.
- Prefer generated code over handwritten boilerplate.
- Keep builds reproducible from a clean clone.
- Avoid duplicating business logic between gRPC and REST.
- Generated code (protobuf, sqlc, mocks, gateway stubs) is committed and NEVER edited by hand; regenerate via documented commands.
- The legacy Deposits service is the architectural template; all new services copy its layout, package naming, constructors, DI, logging, error handling, config style, repository pattern, sqlc workflow, and migration workflow.
- Never modify generated files, existing migrations, or protobuf contracts except through regeneration/approved agent scope.
- Treat the repository and its documentation as source of truth; verify before assuming.

---

## Current Repository Architecture

The repository is a Go microservices monorepo with four runnable services plus
shared platform infrastructure.

Services (each service owns its config, database layer, business logic, gRPC
server, embedded HTTP gateway, Dockerfile, and Makefile):

| Service | Status | Location | Entry Point |
| --- | --- | --- | --- |
| Deposits (legacy) | Runnable | `deposits/` | `deposits/cmd/grpc-service/main.go` |
| Integrations (legacy) | Runnable | `integrations/` | `integrations/cmd/grpc-service/main.go` |
| Clients (new) | Implemented, production-reviewed | `clients/` | `clients/cmd/grpc-service/main.go` |
| Transactions (new) | Implemented, production-reviewed (READY WITH CONDITIONS) | `transactions/` | `transactions/cmd/grpc-service/main.go` |

Per `docs/migration-plan.md`, the legacy services evolve into the new services
(Integrations → Clients, Deposits → Transactions). Legacy services remain
runnable; renames are not performed without explicit instruction.

Platform / shared components (implemented by Platform Agents 02–09):

- `protobuf/` — authoritative protobuf sources (5 files).
- `grpc/go/` — committed generated Go protobuf/gRPC/gateway output.
- `shared/logger` — shared zerolog setup (used by Clients and Transactions).
- `shared/database` — Postgres DSN builder, pgxpool connect + eager ping,
  golang-migrate runner (used by Clients and Transactions).
- `shared/observability` — request-ID correlation, gRPC unary logging
  interceptor, HTTP access-log middleware (used by Clients and Transactions).
- `third_party/googleapis/` — protoc import dependency (git submodule; never
  inspect recursively; do not modify).
- `.github/workflows/` — Render pipeline (active) and OCI pipeline (disabled).
- `render.yaml` — Render Blueprint (3 web services + 3 managed PostgreSQL).
- `docker-compose.yml` — OCI Always Free stack (deposits-only).
- `nginx/` — TLS termination config for OCI.

Documentation locations:

- `README.md` — repository entry point (current architecture, dev workflow).
- `docs/` — foundation architecture docs (domain-model, repository-layout,
  protobuf-strategy, migration-plan) and per-agent platform reviews
  (`docs/platform-*.md`).
- `deploy/` — OCI and Render deployment documentation.
- `agents/` — working agent directives (Cline). Do not delete or "clean up".

---

## Domain Model

Two bounded contexts (see `docs/domain-model.md`):

Clients Context (owned by Clients Service):

- Clients — businesses using RVPay; Admins stored in the same table by role.
- Platforms — supported external marketplaces (e.g. HighLevel), data-driven.
- Integrations — a Client connected to a Platform (OAuth + webhook scope).
- OAuth tokens — encrypted at rest, integration-owned.
- Webhook subscriptions/events — integration-owned.

Transactions Context (owned by Transactions Service):

- Merchants — payment gateways (e.g. PawaPay).
- Customers — end users making payments.
- Deposits — inbound payments (Merchant wallet).
- Payouts — outbound settlements (to Clients).

Entity ownership: each entity lives in exactly one service's database. No
entity is shared across services.

---

## Service Boundaries

- Services never share database tables.
- Cross-service communication is gRPC-only via generated clients (not yet
  wired): Transactions calls `clientsgrpc.ClientService.GetClient` to validate
  `client_id`; Clients may query Transactions for monitoring.
- Shared packages (`shared/logger`, `shared/database`, `shared/observability`)
  are non-business infrastructure; they never import service packages.
- Business services must remain transport-independent and provider-agnostic;
  concrete provider wiring happens only in the composition root
  (`cmd/grpc-service/main.go`).

---

## Database Conventions

- PostgreSQL (pgxpool per service; eager `Ping` at startup).
- One database per service; service owns its migrations, queries, sqlc config
  and generated code, and repository layer.
- Migration files: `up` and `down` SQL migrations in `<service>/db/migrations/`,
  golang-migrate naming style, named `create-...`/additive; down migrations
  drop constraints/indexes/tables/types in dependency order.
- Conventions inherited from Deposits:
  - UUID primary keys (`gen_random_uuid()` requires the `pgcrypto` extension).
  - `timestamptz` columns.
  - snake_case identifiers.
  - explicit foreign keys, indexes, NOT NULL, CHECK, UNIQUE, DEFAULT.
  - PostgreSQL ENUM types where appropriate; no JSON columns unless justified.
  - `ON DELETE` behavior is intentional; financial history uses
    `ON DELETE RESTRICT` (no cascading deletes for transaction history).
- Monetary values are `NUMERIC(18,2)`; passed on the wire as decimal strings
  via `commongrpc.Money`. Never use floating point for money.
- Migrations run at service startup when `RUN_MIGRATIONS=true`, or externally
  (Render/OCI one-shot migration job). Migration failures fail startup.
- Confidential data (OAuth tokens, webhook secrets) is stored encrypted; never
  log secrets.

### SQLC Conventions

- Version pinned: `sqlc v1.29.0`, run via
  `go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate` in each
  service's `db/doc.go` `go:generate` directive.
- Layout per service: `db/sqlc.yaml` (config), `db/query/` (SQL inputs),
  `db/sqlc/` (generated models/methods).
- Repository wrappers in `<service>/db/repo/` encapsulate sqlc; business
  services depend only on repository interfaces (never sqlc/pgx directly).
- Mocks generated by mockgen v0.6.0 into `<service>/db/{sqlc,repo}/mocks/`.
- Never edit sqlc-generated files.

---

## Protobuf / gRPC Conventions

- Sources: `protobuf/*.proto`.
- Generated output: `grpc/go/` — committed; regenerate, never hand-edit.
- Package naming: one package per service — `clientsgrpc`, `transactionsgrpc`,
  `depositsgrpc`, `integrationsgrpc` — plus shared `commongrpc`
  (`protobuf/common.proto`: status enums, `Provider`, `PaymentType`, `Money`,
  pagination, errors, metadata, audit).
- `option go_package` always matches the module:
  `github.com/I-Frostbyte/rvpay-go/grpc/go/<package>`.
- Service packages never import each other; they may import `commongrpc` and
  well-known types.
- Every public RPC carries a `google.api.http` annotation; REST paths are under
  `/v1/public/...` (administrative `/v1/admin/...` reserved for the future).
- Versioning: additive changes in place; breaking changes require a new
  versioned package (`<package>.v2`). Reserved field/enum numbers are never
  reused.
- Generation: `cd protobuf && make lint && make generate-protos`; toolchain
  pinned in `tools/versions.md` (protoc v3.21.12, protoc-gen-go v1.36.10,
  protoc-gen-go-grpc v1.5.1, protoc-gen-grpc-gateway v2.22.0).
- Migration note: legacy `depositsgrpc`/`integrationsgrpc` remain until their
  services are retired; `clientsgrpc`/`transactionsgrpc` replace them.
- HTTP gateway: embedded per service (grpc-gateway v2); generated
  `Register*HandlerServer` wired into each `main.go`; business logic is never
  duplicated between gRPC and REST.

---

## Clients Service

- Purpose: identity, platforms, integrations, OAuth, webhooks.
- Layout: `cmd/grpc-service/`, `config/`, `db/` (migrations, query, repo,
  sqlc), `clients/`, `platforms/`, `integrations/`, `oauth/`, `webhooks/`,
  `providers/`, `service/`, `docs/`, Dockerfile, Makefile, README.
- gRPC services: `clientsgrpc.ClientsService`, `PlatformsService`,
  `IntegrationsService` (`CreateClient`, `GetClient`, `ListClients`,
  `UpdateClient`, `Activate/DeactivateClient`, platform CRUD, integration
  install/uninstall/get/list/reconnect/disconnect/sync).
- OAuth: HighLevel OAuth command flow owned by Clients (token persistence
  encrypted; production review flags encryption-at-rest hardening).
- Webhooks: HighLevel webhook event ingestion/persistence/processing;
  signature verification not yet enforced (production review finding).
- Providers: unified `Provider` interface (`clients/providers`) with a
  concrete `HighLevelProvider`; `ProviderRegistry` registers concrete providers
  in the composition root. The service layer remains provider-agnostic.
- Config: `LOG_LEVEL`, `LISTEN_PORT`, `MIGRATION_PATH`, `RUN_MIGRATIONS`,
  `DB_*`, and HighLevel/webhook secrets (`HIGHLEVEL_CLIENT_ID`,
  `HIGHLEVEL_CLIENT_SECRET`, `HIGHLEVEL_REDIRECT_URI`, `WEBHOOK_SECRET`).
- GHL Custom Payment Provider: Corrected and completed via Agent 01
  (Stages 00–10). Clients owns GHL integration boundary; Transactions owns
  payment domain. Clients exposes `POST /payments/custom-provider/query`
  (verify operation) and `POST /payments/custom-provider/webhook`
  (payment.captured) as thin HTTP adapters that delegate business logic to
  Transactions via gRPC (`GetDepositByGHLTransactionID`). Provider
  configuration (name, description, imageUrl, locationId, queryUrl,
  paymentsUrl, supportsSubscriptionSchedule=false, providerApiKey) is stored
  per-integration in the `payment_provider_configs` table. Webhook idempotency
  reuses the `webhook_events` table unique constraint. Config:
  `HIGHLEVEL_PAYMENT_URL`, `HIGHLEVEL_QUERY_URL`,
  `HIGHLEVEL_PROVIDER_NAME`, `HIGHLEVEL_PROVIDER_DESCRIPTION`,
  `HIGHLEVEL_PROVIDER_IMAGE_URL`, `TRANSACTIONS_GRPC_ADDR`.
- Architecture rules: Clients owns GHL integration and provider registration.
  Transactions owns payment domain, payment state, pawaPay interaction.
  HighLevel registration is outbound from Clients. HighLevel payment
  queries/webhooks delegate to Transactions. pawaPay remains behind
  Transactions provider boundary. Clients never calls pawaPay directly.
  Render is temporary deployment target; AWS is future target. No Render
  hostname is hard-coded; all URLs are configurable via env vars.
- Status: IMPLEMENTED + production-reviewed (READY WITH WARNINGS).

---

## Transactions Service

- Purpose: merchants, customers, deposits, payouts.
- Layout: `cmd/grpc-service/`, `config/`, `db/`, `merchants/`, `customers/`,
  `deposits/`, `payouts/`, `docs/`, Dockerfile, Makefile, README.
- gRPC services: `transactionsgrpc.MerchantService`, `CustomerService`,
  `DepositService`, `PayoutService` (create/get merchants and customers,
  initiate/get deposits, request/get payouts).
- Deposits initialize in `INITIATED`, payouts in `REQUESTED`; no
  status-mutation RPCs yet; provider execution and status reconciliation are
  deferred integration work (documented HIGH findings F-01/F-02).
- Payouts are not customer-scoped (client + merchant only) per domain model.
- Config: `ardanlabs/conf` + `godotenv`; requires `LISTEN_PORT`,
  `MIGRATION_PATH`, and all `DB_*`; `LOG_LEVEL`, `RUN_MIGRATIONS`, `DB_PORT`
  (uint), `DB_TLS_DISABLED` supported.
- Status: IMPLEMENTED + production-reviewed (READY WITH CONDITIONS: client
  validation via Clients service not yet wired, provider execution not
  implemented).

---

## Provider Architecture

- Clients owns the provider abstraction (`clients/providers`): a `Provider`
  interface (capabilities: OAuth, webhooks), a HighLevel implementation, and a
  `ProviderRegistry`. Concrete providers are registered only in the
  composition root; the rest of the service dispatches through interfaces.
- The platform concept is data-driven in Clients (Platforms table), supporting
  future providers without schema redesign.
- Transactions operates on the internal transaction model only; the PawaPay
  client exists and is wired only in the legacy Deposits service.
- There is no global `common.Provider` abstraction; provider boundaries belong
  to the owning service.
- Critical boundary: Clients = GHL integration; Transactions = payment domain.
  This boundary must never be compromised. Payment business logic must never
  reside in Clients; pawaPay calls must never originate from Clients.

---

## OAuth

- HighLevel OAuth is implemented in Clients (`clients/oauth`) and legacy
  Integrations (`integrations/oauth`).
- Flow at runtime: user installs → provider redirects to the callback →
  authorization code is received → backend exchanges the code →
  tokens are stored encrypted → provider API calls can be performed.
- Configuration: `HIGHLEVEL_CLIENT_ID`, `HIGHLEVEL_CLIENT_SECRET`,
  `HIGHLEVEL_REDIRECT_URI` (production redirect must be the deployed public
  URL, never localhost), `TOKEN_ENCRYPTION_KEY` (legacy Integrations,
  32-byte AES-256).
- Never log or commit OAuth client secrets, access tokens, or refresh tokens.

---

## Webhooks

- HighLevel webhook events are ingested, persisted, and processed by Clients
  (`clients/webhooks`) and legacy Integrations (`integrations/webhook`).
- Public endpoints: `POST /v1/public/webhooks` (grpc-gateway →
  `ProcessWebhookEvent`) for Clients; `/webhooks/highlevel` (legacy
  Integrations mux handler).
- Configuration: `WEBHOOK_SECRET` (Clients), `HIGHLEVEL_SSO_KEY` (legacy
  Integrations). Signature verification is not yet enforced.
- No durable queue/worker processing exists yet; events are handled
  synchronously/async in-service.

---

## Runtime

Standard startup convention per service (mirrors Deposits):

1. `main()` creates a signal-aware root context
   (`signal.NotifyContext`, SIGINT/SIGTERM) and initializes the logger once.
2. `run(ctx, logger)`:
   - loads environment-backed config (`LoadConfig`);
   - configures zerolog level;
   - builds the Postgres DSN and connects a pgxpool with eager `Ping`;
   - runs golang-migrate up migrations when `RUN_MIGRATIONS=true`;
   - constructs repositories → provider registry (Clients) → business
     services → gRPC server;
   - creates the gRPC server with reflection, recovery interceptor, and the
     observability unary interceptor (Clients/Transactions); registers the
     gRPC health server;
   - registers generated `Register*Server` on the gRPC server and generated
     `Register*HandlerServer` on a grpc-gateway `runtime.NewServeMux`;
   - mounts the gateway mux (wrapped by the observability access-log
     middleware for Clients/Transactions) plus `/healthz` on
     `http.NewServeMux`;
   - starts HTTP server on `:"$PORT"` (default 8080) and gRPC server on
     `:"$LISTEN_PORT"`; waits on both; aborts startup on either failing;
3. graceful shutdown: context cancel → health NOT_SERVING → HTTP
   `Shutdown` (5s timeout) → `grpcServer.GracefulStop()` → pool close.
4. Startup failures fail fast and exit non-zero; errors are logged via zerolog
   and returned (no panics for ordinary config/connection errors).

Dependency injection: explicit constructor injection; no DI framework, no
global pools/repositories, no service locator.

---

## Platform / Observability

- Structured logging: zerolog JSON to stdout/stderr with timestamp + caller;
  `LOG_LEVEL` (`debug`/`info`/`warn`/`error`) configures verbosity.
- Request correlation: `X-Request-ID` (HTTP header and gRPC metadata).
  Generated when absent; propagated through context; echoed on HTTP response
  and gRPC response metadata; attached to access/RPC logs.
- gRPC instrumentation: `shared/observability.UnaryServerInterceptor` logs
  `request_id`, `rpc`, `grpc_code`, `duration_ms`, and error detail (INFO on
  success, WARN on failure); payloads are never logged.
- HTTP instrumentation: `shared/observability.AccessLog` logs `request_id`,
  `method`, `path`, `status`, `duration_ms`; `/healthz` probes log at DEBUG.
- Health: every service exposes HTTP `/healthz` (200/405, 503 during
  shutdown) and the gRPC health protocol. Render health checks use
  `/healthz`.
- Metrics and tracing are NOT implemented (no documented architecture
  requirement); request IDs provide per-request correlation.
- Observability never logs secrets, tokens, API keys, request bodies,
  authorization headers, financial data, or unnecessary PII.

---

## Configuration

- Environment-driven, per-service config in `<service>/config/`.
- Two established mechanisms (kept separate by design):
  - Clients: bespoke stdlib env helpers (`getEnv`, `getEnvAsInt`,
    `getEnvAsBool`) with defaults.
  - Transactions: `ardanlabs/conf/v3` + `godotenv` with `required` tags.
  - Legacy deposits/integrations follow the deposits config style.
- `.env` files are loaded from the service working directory; never commit
  `.env`; use `.env.example` placeholders.
- Common variables across services: `LOG_LEVEL`, `LISTEN_PORT`, `PORT` (HTTP
  gateway; Render injects), `MIGRATION_PATH`, `RUN_MIGRATIONS`, `DB_HOST`,
  `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_TLS_DISABLED`.
- Secrets are supplied only through environment (`.env`, Render dashboard,
  CI secrets); never hardcode or commit credentials.

---

## Testing

- Repository-wide: `go test ./...`, `go vet ./...`, `gofmt` (hand-written Go).
- Tests require no PostgreSQL, Render, HighLevel, Docker, or external
  infrastructure (unit/service tests run in isolation).
- Mocks: generated by mockgen v0.6.0 (never hand-edited).
- Gateway wiring tests (Agent 03): in-process fakes embedding generated
  `Unimplemented*Server` types; verify route registration, JSON mapping, error
  propagation (gRPC → HTTP status), and `/healthz`.
- Service config tests exist (e.g. `clients/config/model_test.go`,
  `transactions/config/model_test.go`).
- Platform shared packages have focused unit tests (logger, database,
  observability).

---

## Deployment

### Docker

- Multi-stage distroless images per service: `golang:1.26.5-alpine` builder →
  `gcr.io/distroless/static-debian12:nonroot` runtime.
- Build context is the repository root (services depend on root go.mod/go.sum,
  `grpc/go/`, and `shared/`):
  `docker build -f <service>/Dockerfile -t rvpay-<service>:ci .`
- Images contain only the compiled static binary and the service's migration
  files; run as `nonroot`; `EXPOSE 50051` (documentation only); direct
  `ENTRYPOINT` (no shell); CA certs included for HTTPS.
- `CGO_ENABLED=0`; `-trimpath -ldflags="-s -w"`.
- No protoc/sqlc in images; generated code is committed. No `.env`/secrets in
  images (`.dockerignore` excludes them).

### CI/CD

- `.github/workflows/render-deploy.yml` (active on push to `main` +
  `workflow_dispatch`):
  1. `generate` — `go generate ./...` (sqlc + mocks), `cd protobuf && make
     generate-protos`, then `git diff --exit-code` against `grpc/go/` and all
     services' sqlc/mocks dirs (fails on drift with actionable messages);
  2. `validate` — `gofmt -l` (hand-written Go only), `go vet ./...`,
     `go build ./...`;
  3. `test` — `go test ./...`;
  4. `docker-build` — matrix over deposits/clients/transactions Dockerfiles;
  5. `deploy` — POST to `RENDER_DEPLOY_HOOK` secret when set and branch is
     main (explicit skip notice when secret missing).
- Toolchain pinned: Go 1.26.5; protoc/plugins per `tools/versions.md`; sqlc
  v1.29.0 via `go run`.
- `.github/workflows/deploy.yml` (OCI) is intentionally disabled (no `on:`
  trigger).
- CI uses `contents: read` permissions; no secrets printed; no `|| true`,
  `continue-on-error`, or blind retries.

### Render

- `render.yaml` Blueprint (version-controlled infrastructure):
  - `rvpay-deposits` (web) → `rvpay-postgres`;
  - `rvpay-clients` (web) → `rvpay-clients-postgres`, manual secrets
    `HIGHLEVEL_CLIENT_ID`, `HIGHLEVEL_CLIENT_SECRET`,
    `HIGHLEVEL_REDIRECT_URI`, `WEBHOOK_SECRET`;
  - `rvpay-transactions` (web) → `rvpay-transactions-postgres`.
- Each service: Docker runtime, repo-root context, health check `/healthz` on
  injected `PORT`, gRPC on `LISTEN_PORT`, `RUN_MIGRATIONS=true`.
- Manual secrets use `sync: false`; DB credentials wired from Render
  `fromDatabase` (internal URL).
- Free tier includes one managed PostgreSQL; three DBs require a paid plan
  (documented fallback: one DB with distinct `DB_NAME`).

### OCI

- `docker-compose.yml` + `deploy/README.md`: PostgreSQL 16, one-shot
  golang-migrate job, deposits service, non-root Nginx TLS proxy; ARM64,
  Always Free sizing. OCI pipeline disabled.

---

## Coding Conventions

- Go package naming is simple and descriptive; no `utils`/`helpers`/`misc`.
- No god package; responsibilities separated (logger, database, observability).
- No unnecessary interfaces; concrete types unless an interface is justified;
  small, cohesive interfaces for repositories and services; interfaces are
  mockable but not created purely for "clean architecture".
- Explicit constructor injection; no DI framework, no globals, no `init()`
  dependency construction.
- Error handling: idiomatic wrapping with `fmt.Errorf("...: %w", err)`; never
  string-only errors; preserve underlying errors; translate repository errors
  to business errors at the service layer; translate gRPC errors to HTTP via
  grpc-gateway defaults.
- Logging: zerolog structured; never log secrets; log ownership — transport
  logs request lifecycle, service logs business operation context, repository
  logs database-specific failures; no triple-logging of the same error.
- New code must be gofmt-clean, pass `go vet ./...`, and pass the repository
  test suite.
- When uncertain, inspect the existing deposits implementation and follow
  existing patterns; prefer the smallest change; do not invent architecture.