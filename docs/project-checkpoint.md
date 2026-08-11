# RVPay Project Checkpoint

Document Version: 2.2
Status: Handoff / Navigation Document
System: RVPay
Updated: 2026-08-11 (after Platform Agent 11 — Performance)

## Authoritative Files

A fresh Cline session must read these before making changes:

- README.md — repository map.
- agents/project-context.md — project coding, package, naming, generation,
  testing, and implementation conventions (authority; do not contradict).
- docs/domain-model.md — entity ownership and bounded contexts.
- docs/repository-layout.md — target repository structure.
- docs/protobuf-strategy.md — protobuf ownership, packages, shared types,
  versioning, gateway.
- docs/migration-plan.md — ordered migration roadmap and Phase expectations.
- agents/platform/12-final-review.md — the current (next) agent.
- docs/platform-repository-audit.md — baseline for all Platform work.
- docs/project-checkpoint.md — this document.

## Project Map

- **RVPay** is a Go microservices platform for payment processing
  (deposits/payouts) with marketplace platform integration (GoHighLevel),
  PostgreSQL persistence, gRPC + grpc-gateway (REST), and deployment targets
  of Render and Oracle Cloud (OCI).
- **clients/** — Clients Service (new): client/platform/integration/oauth/
  webhook domain. Owns `clients/db` (migrations, query, sqlc, repo),
  `clients/service`, `clients/oauth`, `clients/webhooks`, `clients/providers`,
  `clients/config`, `clients/cmd/grpc-service`.
- **transactions/** — Transactions Service (new): merchant/customer/deposit/
  payout domain. Owns `transactions/db` (migrations, query, sqlc, repo),
  `transactions/{merchants,customers,deposits,payouts}`, `transactions/config`,
  `transactions/cmd/grpc-service`.
- **deposits/** — legacy Deposits Service (still runnable; evolves into
  Transactions per the migration plan; not deleted).
- **integrations/** — legacy Integrations Service (still runnable; evolves
  into Clients per the migration plan; not deleted).
- **protobuf/** — authoritative protobuf sources: `clients.proto`,
  `transactions.proto`, `common.proto` (shared `commongrpc`), plus legacy
  `deposits.proto`, `integrations.proto`. Includes `Makefile`
  (`make lint`, `make generate-protos`) and `Dockerfile`.
- **grpc/go/** — committed generated Go protobuf/gRPC/gateway code:
  `{clients,transactions,deposits,integrations}grpc/`, `commongrpc/`.
- **shared/** — shared platform infrastructure (non-business):
  `shared/logger`, `shared/database`, `shared/observability`.
- **third_party/googleapis/** — external protobuf dependency (submodule).
  Never inspect recursively.
- **docs/** — architecture documents and the platform review documents
  (`docs/platform-*.md`).
- **clients/docs/** — Clients Service review documents.
- **deploy/** — OCI and Render deployment documentation.
- **nginx/** — TLS termination config for OCI.
- **tools/versions.md** — pinned toolchain versions used by CI.
- **render.yaml** — Render Blueprint (3 web services + 3 managed PostgreSQL).
- **.github/workflows/** — Render pipeline (active) and OCI pipeline (disabled).

## Current Implementation State

The repository is a Go microservices monorepo with four runnable services plus
shared platform infrastructure. The cumulative implementation through Platform
11 is:

- **Foundation** — COMPLETE: domain model, repository layout, protobuf
  strategy, migration plan (`docs/00-foundation`, `docs/`).
- **Clients Service** — COMPLETE and production-reviewed
  (`clients/docs/production-readiness-review.md`, READY WITH WARNINGS).
- **Transactions Service** — COMPLETE and production-reviewed
  (`docs/transactions-production-review.md`, READY WITH CONDITIONS).
- **Platform** — Agents 01–11 COMPLETE; Agent 12 NOT STARTED.
- **PostgreSQL** — IMPLEMENTED as the only datastore (pgxpool per service;
  golang-migrate up/down per service; sqlc v1.29.0).
- **Protobuf/gRPC** — IMPLEMENTED across all services; deterministic
  generation via `protobuf/Makefile`; committed generated output.
- **HTTP gateway** — IMPLEMENTED per service (grpc-gateway v2 + `/healthz`),
  embedded in each service's `main.go`.
- **Shared packages** — IMPLEMENTED: `shared/logger`, `shared/database`,
  `shared/observability` (used by Clients and Transactions).
- **Observability** — IMPLEMENTED: zerolog JSON, `X-Request-ID` correlation,
  gRPC unary logging interceptor, HTTP access-log middleware, `/healthz` +
  gRPC health. Metrics/tracing NOT implemented (no documented requirement).
- **Provider integrations** — PLANNED/IN PROGRESS for Deposits/Payouts:
  Transactions operates on the internal transaction model only; PawaPay client
  exists and is wired only in the legacy Deposits service.
- **OAuth / Webhooks** — IMPLEMENTED in Clients (HighLevel OAuth + webhooks)
  and legacy Integrations.
- **Background processing / queues** — PLANNED/NOT PRESENT. No workers, no
  polling, no reconciliation jobs exist.

## Service Status

### Clients Service

- **Implementation status**: COMPLETE, production-reviewed (READY WITH
  WARNINGS).
- **Major components**: Clients, Platforms, Integrations, OAuth, Webhooks via
  a unified Provider interface (`clients/providers`); gRPC + gateway;
  PostgreSQL; zerolog; config; Docker; tests.
- **Database**: COMPLETE — `clients/db` (migrations, query, sqlc, repo).
- **Protobuf**: COMPLETE — `clientsgrpc` (`ClientsService`,
  `PlatformsService`, `IntegrationsService`).
- **Repository**: COMPLETE — `clients/db/repo` (sqlc-encapsulating repos).
- **Service**: COMPLETE — `clients/service` (business logic, converters).
- **OAuth**: COMPLETE — HighLevel OAuth flow (`clients/oauth`); token
  encryption-at-rest hardening flagged in production review.
- **Webhook**: COMPLETE — HighLevel webhook ingestion/persistence
  (`clients/webhooks`); signature verification not yet enforced.
- **Runtime**: COMPLETE — `clients/cmd/grpc-service/main.go` (gRPC + gateway +
  healthz + graceful shutdown + observability).
- **Tests**: COMPLETE — service tests + gateway wiring tests; race-clean.

### Transactions Service

- **Implementation status**: COMPLETE, production-reviewed (READY WITH
  CONDITIONS).
- **Database**: COMPLETE — `transactions/db` (migrations, query, sqlc, repo).
- **SQLC**: COMPLETE — v1.29.0, committed generated code.
- **Protobuf**: COMPLETE — `transactionsgrpc` (`MerchantService`,
  `CustomerService`, `DepositService`, `PayoutService`).
- **Repositories**: COMPLETE — `transactions/db/repo`.
- **Merchants**: COMPLETE — `transactions/merchants`.
- **Customers**: COMPLETE — `transactions/customers`.
- **Deposits**: COMPLETE — `transactions/deposits` (initiate/get; provider
  execution deferred).
- **Payouts**: COMPLETE — `transactions/payouts` (request/get; not
  customer-scoped).
- **Runtime**: COMPLETE — `transactions/cmd/grpc-service/main.go` (gRPC +
  gateway + healthz + graceful shutdown + observability).
- **Tests**: COMPLETE — service tests + gateway wiring tests; race-clean.

### Platform

- **Platform 01 — Repository Audit**: COMPLETE
  (`docs/platform-repository-audit.md`).
- **Platform 02 — Protobuf Generation**: COMPLETE
  (`docs/platform-protobuf-generation-review.md`).
- **Platform 03 — HTTP Gateway**: COMPLETE
  (`docs/platform-http-gateway-review.md`).
- **Platform 04 — Common Packages**: COMPLETE
  (`docs/platform-common-packages-review.md`).
- **Platform 05 — CI/CD**: COMPLETE (`docs/platform-ci-cd-review.md`).
- **Platform 06 — Docker**: COMPLETE (`docs/platform-docker-review.md`).
- **Platform 07 — Render**: COMPLETE (`docs/platform-render-review.md`).
- **Platform 08 — Documentation**: COMPLETE
  (`docs/platform-documentation-review.md`).
- **Platform 09 — Observability**: COMPLETE
  (`docs/platform-observability-review.md`).
- **Platform 10 — Security**: COMPLETE
  (`docs/platform-security-review.md`).
- **Platform 11 — Performance**: COMPLETE
  (`docs/platform-performance-review.md`).
- **Platform 12 — Final Review**: NOT STARTED.

## Agent Progress

| Area | Agent | Status | Notes |
| ---- | ----- | ------ | ----- |
| Foundation | 00-foundation (01–04) | COMPLETE | domain, layout, protobuf strategy, migration plan |
| Clients | 01–12 | COMPLETE | production-reviewed (READY WITH WARNINGS) |
| Transactions | 01–13 | COMPLETE | production-reviewed (READY WITH CONDITIONS) |
| Platform | 01 Repository Audit | COMPLETE | `docs/platform-repository-audit.md` |
| Platform | 02 Protobuf Generation | COMPLETE | `docs/platform-protobuf-generation-review.md` |
| Platform | 03 HTTP Gateway | COMPLETE | `docs/platform-http-gateway-review.md` |
| Platform | 04 Common Packages | COMPLETE | `docs/platform-common-packages-review.md` |
| Platform | 05 CI/CD | COMPLETE | `docs/platform-ci-cd-review.md` |
| Platform | 06 Docker | COMPLETE | `docs/platform-docker-review.md` |
| Platform | 07 Render | COMPLETE | `docs/platform-render-review.md` |
| Platform | 08 Documentation | COMPLETE | `docs/platform-documentation-review.md` |
| Platform | 09 Observability | COMPLETE | `docs/platform-observability-review.md` |
| Platform | 10 Security | COMPLETE | `docs/platform-security-review.md` |
| Platform | 11 Performance | COMPLETE | `docs/platform-performance-review.md` |
| Platform | 12 Final Review | NOT STARTED | next agent |

## Current Work

The repository confirms Platform 11 is complete
(`docs/platform-performance-review.md` exists; the HighLevel HTTP client is
reused and `ListMerchants` honors pagination with a bounded page size).

**Platform 11 — Performance — COMPLETE**

The next logical unfinished agent is:

**agents/platform/12-final-review.md**

## Next Action

The next implementation task is:

**agents/platform/12-final-review.md**

The next Cline session should read the authoritative project files (README.md,
agents/project-context.md, docs/domain-model.md, docs/repository-layout.md,
docs/protobuf-strategy.md, docs/migration-plan.md, and the relevant platform
reviews) followed by the Platform 12 agent instructions before making changes.

## Rules for Continuing

1. README.md is the repository map.
2. agents/project-context.md contains persistent technical/project conventions.
3. The foundation docs are authoritative for architecture and migration
   decisions.
4. Follow the active agent's directives.
5. Do not overwrite existing working code.
6. Do not perform unrelated refactoring.
7. Do not redesign architecture unless explicitly instructed.
8. Inspect only relevant directories.
9. Do not waste time recursively exploring third_party/googleapis.
10. Do not manually modify generated protobuf code.
11. Do not manually modify generated SQLC code.
12. Follow existing project conventions.
13. Run appropriate tests.
14. Review changes before completing an agent.
15. Update relevant documentation when required.
16. Treat actual repository state as the final source of truth.

## Known Issues

- No compilation failures, failing tests, or migration problems are known at
  this checkpoint; the repository builds and the test suite passes.
- Transactions production review documented HIGH findings (F-01 provider
  execution/reconciliation, F-02 client_id cross-service validation, F-03 no
  auth) and MEDIUM findings (pagination, lifecycle RPCs, fee entity,
  idempotency exposure) — consciously deferred; see
  `docs/transactions-production-review.md`.
- Clients production review documented HIGH/MEDIUM findings (OAuth token
  encryption at rest, redirect URI config wiring, webhook dedup) — see
  `clients/docs/production-readiness-review.md`. The redirect URI config wiring
  finding was addressed by Platform 10 (SEC-02: configured `HIGHLEVEL_REDIRECT_URI`
  is now used by the OAuth service).
- No authentication/authorization middleware in any service (Platform 10).
- No metrics/tracing/request-ID-based distributed tracing (request IDs exist;
  metrics/tracing deferred — Platform 11/12).
- Render free tier includes one managed PostgreSQL; the Blueprint defines three
  databases (one per service), which requires a paid plan (documented fallback
  in `deploy/render/README.md`).
- OCI GitHub Actions pipeline (`deploy.yml`) is intentionally disabled.
- Legacy `deposits/` and `integrations/` services remain runnable and are not
  deleted; they retain their pre-existing (non-shared) infrastructure.

## Important Decisions

- **Deposits → Transactions, Integrations → Clients.** Per
  `docs/migration-plan.md`, the legacy services evolve into the new services.
  The legacy services remain runnable and have NOT been deleted; renames are
  not performed without explicit instruction.
- **Exact service template.** The Deposits service is the canonical template;
  all new services copy its `cmd/grpc-service`, `config`, `db/{migrations,
  query, repo, sqlc}`, `main()→run()`, zerolog, config-loading, Docker,
  Makefile, and error conventions.
- **Protobuf strategy.** Each service owns one `*grpc` package; shared types
  live in `commongrpc`; `go_package` always matches the module; REST is
  generated via `google.api.http`; `/v1/public/...` for public routes.
- **Cross-service communication is gRPC-only.** No shared database tables;
  `client_id` in Transactions is a plain UUID resolved via the Clients Service
  (not yet wired — deferred integration).
- **Generated code is committed and never hand-edited.** Regenerate via the
  protobuf Makefile and per-service `go:generate` (sqlc@v1.29.0; mockgen
  v0.6.0) and verify with `git diff --exit-code`.
- **Monetary values are stored as NUMERIC(18,2)** and passed as decimal
  strings via `commongrpc.Money`; no floating-point arithmetic for money.
- **Financial history is preserved.** Foreign keys use ON DELETE RESTRICT;
  no cascading deletes for transaction history.
- **Transactions deposits/payouts initialize in INITIATED/REQUESTED** and
  have no status-mutation RPC yet; provider execution and status
  reconciliation are future integration work.
- **Transactions payouts are not customer-scoped** (client + merchant only).
- **Shared packages are non-business infrastructure.** `shared/logger`,
  `shared/database`, `shared/observability` never import service packages;
  business logic stays in the owning service.
- **Observability conventions.** zerolog JSON to stdout/stderr; `X-Request-ID`
  correlation (HTTP header + gRPC metadata); gRPC unary logging interceptor +
  HTTP access-log middleware; `/healthz` + gRPC health; metrics/tracing not
  implemented (no documented requirement).
- **Render Blueprint** deploys three web services (deposits, clients,
  transactions), each with its own managed PostgreSQL; manual secrets are
  `sync: false`.
- **Agent files are the working instructions.** Keep all `agents/` files; they
  are the directives the platform agents execute, and they must not be
  deleted or "cleaned up" during implementation.

## Recently Completed

- Clients service implementation completed through its final production
  review (`clients/docs/production-readiness-review.md`).
- Transactions service implementation completed through its final production
  review (`docs/transactions-production-review.md`).
- Platform repository audit completed (`docs/platform-repository-audit.md`).
- Platform protobuf generation completed
  (`docs/platform-protobuf-generation-review.md`).
- Platform HTTP gateway completed (`docs/platform-http-gateway-review.md`).
- Platform common packages completed (`docs/platform-common-packages-review.md`).
- Platform CI/CD completed (`docs/platform-ci-cd-review.md`).
- Platform Docker configuration completed (`docs/platform-docker-review.md`).
- Platform Render configuration completed (`docs/platform-render-review.md`).
- Platform documentation completed (`docs/platform-documentation-review.md`).
- Platform observability completed (`docs/platform-observability-review.md`).
- Platform security completed (`docs/platform-security-review.md`).
- Platform performance completed (`docs/platform-performance-review.md`).

## Current Repository State

- `git status --short`: clean (the separately-updated `agents/project-context.md`
  has been committed by the user).
- `git log --oneline -10` (HEAD on `platform-cleanup`):
  - `7309c13` (HEAD) — project-context update (committed by user)
  - `55b3e4a 09-observability`
  - `0c0b65a 08-documentation`
  - `c2c13b8 07-render`
  - `a2c3d73 06-docker`
  - `f2b341d 05-ci-cd`
  - `397b41f 04-common-packages`
  - `a020431 03-http-gateway`
  - `59f9616 project-checkpoint and chat change`
  - `5a79f21 01-repsitory-audit`
- Generated files (`grpc/go/*`, `*sqlc*`, `*mocks*`) are committed and should
  only change via regeneration; do not edit by hand.

## New Cline Session Handoff

Start by reading README.md, agents/project-context.md, docs/project-checkpoint.md,
and the relevant foundation documents (docs/domain-model.md,
docs/repository-layout.md, docs/protobuf-strategy.md, docs/migration-plan.md).
Then read the active agent file (agents/platform/12-final-review.md). Do not
rely on the previous Cline conversation. Treat the repository and these
documents as the source of truth.

Current completed endpoint:

**Platform 11 — Performance**

Next task:

**Platform 12 — Final Review**
