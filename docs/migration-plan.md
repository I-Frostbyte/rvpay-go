# RVPay Migration Plan

Document Version: 1.0
Status: Foundation
System: RVPay
Architecture: Go Microservices
Database: PostgreSQL
Communication: gRPC + Protocol Buffers + gRPC-Gateway (HTTP)

## 1. Purpose

This document defines the ordered implementation roadmap required to evolve the
current repository into the new RVPay architecture described in
`docs/domain-model.md`, `docs/repository-layout.md`, and
`docs/protobuf-strategy.md`.

This is a migration planning document. No code, database, or protobuf
generation is performed by this document, and no repository files are modified.

## 2. Existing Architecture

### 2.1 Current Services

| Service | Root | Responsibilities |
|---------|------|------------------|
| Deposits | `deposits/` | Accepts `InitiateDeposit`, stores a client and deposit in PostgreSQL, calls PawaPay to initiate external mobile-money deposits |
| Integrations | `integrations/` | Manages third-party provider connections: HighLevel OAuth callback flow, encrypted provider tokens, HighLevel webhook event persistence |

### 2.2 Current Responsibilities

- **Deposits service** — deposit initiation, client persistence (hard-coded
  `Socadel`), PawaPay client integration, gRPC + HTTP gateway with `/healthz`.
- **Integrations service** — HighLevel OAuth callback, encrypted token storage,
  webhook event persistence, gRPC + HTTP gateway with `/healthz`,
  `/oauth/callback`, `/webhooks/highlevel`.

### 2.3 Existing Reusable Work

- **Service template** — the deposits service defines the canonical layout
  (`cmd/grpc-service/`, `config/`, `db/`, `<domain>/`, Dockerfile, Makefile).
- **Configuration pattern** — environment-backed `LoadConfig` used by both
  services.
- **Database layer** — migrations, sqlc query/`sqlc`/`repo` structure.
- **Runtime wiring** — gRPC server, HTTP gateway, health checks, graceful
  shutdown, recovery middleware.
- **OAuth and webhook logic** — HighLevel OAuth handler/service and webhook
  handler/service in the Integrations service.
- **PawaPay client** — external mobile-money deposit initiation in the Deposits
  service.

### 2.4 Technical Debt

- Hard-coded client (`Socadel`) in the Deposits service; `client_id` from the
  request is unused.
- Unique `email`/`phone_number` on clients causes a second successful request
  to fail.
- No transaction spanning the database write and the external PawaPay call.
- No callback handling or status reconciliation for deposits.
- `CreateDepositResponse` always returns `ACCEPTED`/`FINAL_STATUS`.
- Only the HighLevel provider is implemented for OAuth and webhooks.
- Webhook signature verification (`HIGHLEVEL_SSO_KEY`) is not enforced.
- `ProcessWebhookEvent` gRPC method is defined but not implemented.
- Internal service communication (e.g. notifying the deposits service) is not
  implemented.
- `go_package` option in protobufs historically mismatched the Go module
  (now aligned to `github.com/I-Frostbyte/rvpay-go`).

### 2.5 Areas Suitable for Reuse

- The deposits service template for scaffolding the new Clients and
  Transactions services.
- The configuration, database, and runtime wiring patterns.
- The OAuth and webhook logic for the Clients Service.
- The PawaPay client and deposit flow for the Transactions Service.

## 3. Migration Goals

- Produce an ordered, incremental migration plan.
- Prefer incremental evolution over rewrites.
- Reuse existing code whenever practical.
- Avoid duplicate implementations.
- Keep the current services runnable until fully replaced.

## 4. Migration Phases

### Phase 1 — Repository Preparation

**Objective**

Prepare the repository for the new architecture without breaking existing
services.

**Dependencies**

- None (foundation phase).

**Expected Outputs**

- `docs/domain-model.md` (created).
- `docs/repository-layout.md` (created).
- `docs/protobuf-strategy.md` (created).
- `docs/migration-plan.md` (this document).
- `shared/` directory scaffolded with generic infrastructure packages
  (config, logger, database, middleware) extracted from existing services.
- `protobuf/common.proto` created with shared enums (`UserRole`, `Provider`,
  `PaymentType`) and `Money` message.

**Risks**

- Extracting shared code may introduce subtle behaviour changes if not done
  carefully.
- Creating `common.proto` before service contracts may require iteration.

**Rollback Considerations**

- Shared extraction is additive; existing services remain unchanged and
  runnable.
- `common.proto` is additive; existing contracts are untouched.

**Success Criteria**

- Existing services still build and pass tests.
- `shared/` packages are used by at least one service without behaviour change.
- `common.proto` lints and generates cleanly.

### Phase 2 — Clients Service Migration

**Objective**

Evolve the Integrations service into the Clients Service owning Clients,
Platforms, and Integrations.

**Dependencies**

- Phase 1 (shared infrastructure, `common.proto`).

**Expected Outputs**

- `clients/` service scaffolded following the deposits template.
- `protobuf/clients.proto` created (package `clientsgrpc`) with
  `ClientService`, `PlatformService`, and `IntegrationService` RPCs.
- `grpc/go/clientsgrpc/` generated.
- Clients, Platforms, and Integrations database migrations, queries, and sqlc
  code.
- OAuth and webhook logic migrated from `integrations/` into `clients/`.
- `ClientService.GetClient` RPC available for cross-service validation.

**Risks**

- Migrating OAuth/webhook logic may surface provider-specific assumptions.
- Introducing `platform_id`/`client_id` on Integrations changes the data model.

**Rollback Considerations**

- The `integrations/` service remains runnable until `clients/` is complete.
- New contracts are additive; existing `integrationsgrpc` is untouched.

**Success Criteria**

- `clients/` service builds, passes tests, and runs with gRPC + HTTP gateway.
- Clients, Platforms, and Integrations CRUD works end-to-end.
- OAuth and webhook flows work under the Clients Service.
- `ClientService.GetClient` is callable by the Transactions Service.

### Phase 3 — Transactions Service Migration

**Objective**

Evolve the Deposits service into the Transactions Service owning Merchants,
Customers, Deposits, and Payouts.

**Dependencies**

- Phase 1 (shared infrastructure, `common.proto`).
- Phase 2 (`ClientService.GetClient` for client validation).

**Expected Outputs**

- `transactions/` service scaffolded following the deposits template.
- `protobuf/transactions.proto` created (package `transactionsgrpc`) with
  `MerchantService`, `CustomerService`, `DepositService`, and `PayoutService`
  RPCs.
- `grpc/go/transactionsgrpc/` generated.
- Merchants, Customers, Deposits, and Payouts database migrations, queries, and
  sqlc code.
- Deposit flow migrated from `deposits/` into `transactions/`, replacing the
  hard-coded client with `ClientService.GetClient` validation.
- Payout flow implemented.

**Risks**

- Replacing the hard-coded client with real client resolution changes deposit
  behaviour.
- Introducing Payouts adds new settlement logic and fee handling.

**Rollback Considerations**

- The `deposits/` service remains runnable until `transactions/` is complete.
- New contracts are additive; existing `depositsgrpc` is untouched.

**Success Criteria**

- `transactions/` service builds, passes tests, and runs with gRPC + HTTP
  gateway.
- Merchants, Customers, Deposits, and Payouts work end-to-end.
- Deposit flow validates `client_id` via the Clients Service.
- Payout flow works with fee deduction.

### Phase 4 — Shared Infrastructure

**Objective**

Consolidate and harden shared infrastructure across the new services.

**Dependencies**

- Phases 2 and 3 (both services use the shared packages).

**Expected Outputs**

- `shared/` packages fully adopted by both `clients/` and `transactions/`.
- Shared configuration, logger, database helpers, and middleware used
  consistently.
- Duplicate implementations removed from service roots.

**Risks**

- Over-extraction may couple services unnecessarily.
- Behavioural drift if shared code is modified after adoption.

**Rollback Considerations**

- Shared packages are versioned with the repository; services can pin to a
  prior commit if needed.

**Success Criteria**

- Both services use the same shared packages for config, logger, database, and
  middleware.
- No duplicate infrastructure code remains in service roots.

### Phase 5 — Testing

**Objective**

Validate the migrated services and the overall repository.

**Dependencies**

- Phases 2, 3, and 4.

**Expected Outputs**

- Unit tests for Clients, Platforms, Integrations, Merchants, Customers,
  Deposits, and Payouts services.
- Integration tests for cross-service gRPC communication.
- Migration tests for database up/down migrations.
- `go test ./...`, `go generate ./...`, and `go vet ./...` pass.

**Risks**

- Cross-service tests require both services running.
- Database migration tests require a PostgreSQL instance.

**Rollback Considerations**

- Tests are additive; failures do not affect running services.

**Success Criteria**

- All unit, integration, and migration tests pass.
- `go test ./...`, `go generate ./...`, and `go vet ./...` succeed.

### Phase 6 — Deployment

**Objective**

Deploy the new architecture to Oracle Cloud Always Free and Render.

**Dependencies**

- Phases 2, 3, 4, and 5.

**Expected Outputs**

- Dockerfiles for `clients/` and `transactions/`.
- `docker-compose.yml` updated to run both new services.
- CI/CD pipelines (`.github/workflows/`) updated for the new services.
- `render.yaml` updated for the new services.
- Legacy `deposits/` and `integrations/` services retired.

**Risks**

- Deployment configuration drift between OCI and Render.
- Retiring legacy services may break existing consumers if not coordinated.

**Rollback Considerations**

- Keep legacy services deployable until the new services are verified in
  production.
- Deployment artifacts are versioned; rollback to a prior commit restores the
  previous stack.

**Success Criteria**

- Both new services deploy and run on Oracle Cloud Always Free.
- Both new services deploy and run on Render.
- Legacy services are fully retired.
- Health checks and gateway endpoints respond correctly.

## 5. Migration Principles

- **Incremental evolution** — each phase builds on the previous; no big-bang
  rewrite.
- **Reuse over rewrite** — existing deposits template, OAuth/webhook logic, and
  PawaPay client are reused.
- **No duplicate implementations** — shared infrastructure is consolidated in
  `shared/`; business logic stays in the owning service.
- **Additive contracts** — new protobuf packages are added alongside existing
  ones; legacy contracts are retired only after callers migrate.
- **Keep services runnable** — legacy services remain deployable until fully
  replaced.

## 6. Assumptions

- The existing Deposits and Integrations services evolve into the Transactions
  and Clients services respectively.
- The deposits service template is the canonical pattern for new services.
- Cross-service communication occurs only through gRPC contracts.
- No entity is owned by multiple services; no database tables are shared.
- The migration is performed incrementally with legacy services kept runnable
  until replacement.

## 7. Unresolved Questions

- **Fee entity** — the payout flow deducts fees, but no fee entity is defined;
  this must be resolved before Phase 3 payout implementation.
- **Wallet entity** — the administrator-controlled merchant wallet is not
  modelled; balance tracking may be required before settlement logic.
- **Enum value migration** — whether existing `DepositProvider` values map 1:1
  to the shared `Provider` values or require a compatibility layer.
- **Integration message shape** — whether `platform_id`/`client_id` are added
  to the existing `Integration` message or introduced as a new message.
- **Webhook event message** — whether `ProcessWebhookEventRequest` evolves in
  place or is replaced by a typed message.
- **Pagination** — whether `List*` RPCs require pagination in initial contracts.
- **Admin API versioning** — whether administrative RPCs share a package with
  public RPCs or use a separate versioned package.