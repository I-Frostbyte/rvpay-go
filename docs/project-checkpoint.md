# RVPay Project Checkpoint

Document Version: 1.0
Status: Handoff / Navigation Document
System: RVPay
Updated: 2026-08-10 (after Platform Agent 01 — Repository Audit)

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
- agents/platform/02-protobuf-generation.md — the current (next) agent.
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
- **protobuf/** — authoritative protobuf sources: `deposits.proto`,
  `integrations.proto`, `clients.proto`, `transactions.proto`, `common.proto`
  (shared `commongrpc`). Includes `Makefile` (generate-protos) and
  `Dockerfile`.
- **grpc/go/** — committed generated Go protobuf/gRPC/gateway code:
  `{deposits,integrations,clients,transactions}grpc/`, `commongrpc/`.
- **third_party/googleapis/** — external protobuf dependency (submodule).
  Never inspect recursively.
- **docs/** — architecture documents and the transactions/platform review
  documents (`docs/transactions-*.md`, `docs/platform-repository-audit.md`).
- **clients/docs/** — Clients Service review documents
  (`clients/docs/production-readiness-review.md`, etc.).
- **deploy/** — OCI (`.git/` deploy?) and Render deployment documentation:
  `deploy/README.md`, `deploy/render/README.md`.
- **nginx/** — TLS termination config for OCI (`nginx.conf`).
- **tools/versions.md** — pinned toolchain versions used by CI.
- **Shared/platform code:** none yet — no `shared/`, `pkg/`, `internal/`,
  `common/`, `middleware/`, `config/`, `logging/`, `errors/`, or `auth/`
  directory exists at the repository root. This is the primary Platform gap.

## Current Architecture

- **Clients Service** (`clients/`) — IMPLEMENTED and PRODUCTION-REVIEWED
  (`clients/docs/production-readiness-review.md`): Clients, Platforms,
  Integrations, OAuth, Webhooks via unified Provider interface; gRPC +
  gateway; PostgreSQL; zerolog; config; Docker; tests.
- **Transactions Service** (`transactions/`) — IMPLEMENTED and
  PRODUCTION-REVIEWED (`docs/transactions-production-review.md` — decision
  READY WITH CONDITIONS): Merchants, Customers, Deposits, Payouts; gRPC +
  gateway; PostgreSQL; zerolog; config; Docker; tests.
- **Platform infrastructure** — IN PROGRESS. Agent 01 (Repository Audit)
  complete; agents 02–12 not started. No shared/common packages yet.
- **PostgreSQL** — IMPLEMENTED as the only datastore (pgxpool per service;
  goose not used — golang-migrate used via `repo.Migrate`).
- **protobuf/gRPC** — IMPLEMENTED across all services; deterministic
  generation via `protobuf/Makefile` and per-service `go:generate`
  (sqlc@v1.29.0, mockgen v0.6.0).
- **HTTP gateway** — IMPLEMENTED per service (grpc-gateway v2 + `/healthz`),
  duplicated in each service's `main.go` (platform integration pending).
- **Provider integrations** — PLANNED/IN PROGRESS for Deposits/Payouts:
  Transactions currently operates on the internal transaction model only;
  PawaPay client exists and is wired only in the legacy Deposits service.
- **OAuth / Webhooks** — IMPLEMENTED in Clients (GoHighLevel OAuth + webhooks).
- **Background processing / queues** — PLANNED/NOT PRESENT. No workers, no
  polling, no reconciliation jobs exist.

## Implementation Status

| Area | Status | Location | Notes |
| ---- | ------ | -------- | ----- |
| Foundation docs | ✅ Complete | docs/ | domain, layout, protobuf strategy, migration plan, 00-foundation |
| Clients | ✅ Complete | clients/ | Agents 01–12 done; review: `clients/docs/production-readiness-review.md` (READY WITH WARNINGS) |
| Transactions | ✅ Complete | transactions/ | Agents 01–13 done; review: `docs/transactions-production-review.md` (READY WITH CONDITIONS) |
| Platform | 🔄 Agent 01 done | docs/platform-repository-audit.md | Agents 02–12 not started; next is 02-protobuf-generation |
| Database | ✅ Complete | deposits|integrations|clients|transactions/db/migrations | golang-migrate up/down per service |
| SQLC | ✅ Complete | */db/query + sqlc | Deterministic; pinned v1.29.0 |
| Protobuf | ✅ Complete (sources/generation) | protobuf/, grpc/go/ | 5 source files; committed generated output |
| Runtime | ✅ Complete per service | */cmd/grpc-service/main.go | gRPC + gateway + healthz + graceful shutdown |
| OAuth | ✅ Complete | clients/oauth | HighLevel OAuth |
| Webhooks | ✅ Complete | clients/webhooks | HighLevel webhooks |
| Tests | ✅ Service tests | clients/*/service_test.go, transactions/*/*_test.go | Race-clean; full suite passes |
| Docker | ✅ Complete per service | */Dockerfile (multi-stage distroless) | Builds verified |
| CI/CD | ✅ Render pipeline active; OCI disabled | .github/workflows/ | render-deploy.yml active; deploy.yml on workflow_dispatch only |
| Render | ✅ Deposits only | render.yaml | Clients/Transactions not yet in Blueprint |

## Agent Progress

### Clients
01–12 — COMPLETE (`clients/docs/` review documents reference each agent).
No further Clients work is pending.

### Transactions
01–13 — COMPLETE (`docs/transactions-*.md` review documents reference each agent).
Final decision: READY WITH CONDITIONS.

### Platform
- 01 — COMPLETE (this session: `docs/platform-repository-audit.md`).
- 02 — NEXT (protobuf generation).
- 03–12 — PENDING (gateway, common packages, CI/CD, Docker, Render,
  documentation, observability, security, performance, final review).

## Current Work

The repository checkpoint was created immediately after completing **Platform
Agent 01 (Repository Audit)**. The audit report
(`docs/platform-repository-audit.md`) was the last deliverable produced and
the working tree is clean apart from the new audit document (now committed
state as of checkpoint creation). The current agent sequence is **Platform**;
agent 01 is done and agent 02 is the next task.

## Next Action

Read:

agents/platform/02-protobuf-generation.md

Then continue the Platform implementation according to that agent's
directives, using the baseline in `docs/platform-repository-audit.md`.

## Known Issues

- No shared/common package directory exists (config/logger/db/gateway
  bootstrap is duplicated across the four services) — Platform agent 04.
- `make test` is broken in deposits/, integrations/, and clients/ Makefiles
  (broken `find | xargs` variant); transactions/ uses a working
  `go test ./... --cover` — Platform agent 05.
- CI sqlc verification (`render-deploy.yml`) scopes `git diff --exit-code` to
  `deposits/db/sqlc` only; clients/transactions sqlc drift is not explicitly
  verified — Platform agent 05.
- render.yaml deploys only the legacy Deposits service; Clients/Transactions
  not yet deployable via Render — Platform agent 07.
- No authentication middleware in any service; no metrics/tracing/request
  IDs — Platform agents 10, 09.
- Transactions production review documented HIGH findings (F-01 provider
  execution/reconciliation, F-02 client_id cross-service validation,
  F-03 no auth) and MEDIUM findings (pagination, lifecycle RPCs, fee entity,
  idempotency exposure) — consciously deferred; see
  `docs/transactions-production-review.md`.
- Clients production review documented HIGH/MEDIUM findings (OAuth token
  encryption at rest, redirect URI config wiring, webhook dedup) — see
  `clients/docs/production-readiness-review.md`.
- There is no `shared/` platform directory yet; any extraction must not be
  invented — it must follow Platform agent 04 directives.

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
- **Agent files are the working instructions.** Keep all `agents/` files; they
  are the directives the platform agents execute, and they must not be
  deleted or "cleaned up" during implementation.

## Rules for Continuing Work

1. Read README.md first.
2. Read agents/project-context.md before modifying code.
3. Read the relevant foundation documentation.
4. Follow the current agent's directives.
5. Treat the repository as the source of truth.
6. Do not overwrite existing working code.
7. Do not perform unrelated refactoring.
8. Do not redesign architecture unless explicitly instructed.
9. Do not inspect unnecessary deep directories.
10. Do not recursively inspect third_party/googleapis.
11. Do not manually modify generated protobuf code.
12. Do not manually modify generated SQLC code.
13. Regenerate generated code using the project's established commands.
14. Do not rewrite existing migrations unless explicitly instructed.
15. Do not use destructive git commands.
16. Preserve existing project conventions.
17. Run appropriate tests after implementation.
18. Review changes before completing the task.
19. Update relevant documentation when the agent requires it.
20. Do not assume something is implemented merely because documentation exists.
21. Verify implementation against the actual repository.

## New Cline Session Reading Order

Specify this exact order:

### First

README.md

### Second

agents/project-context.md

### Third

docs/domain-model.md

docs/repository-layout.md

docs/protobuf-strategy.md

docs/migration-plan.md

### Fourth

The current agent file (agents/platform/02-protobuf-generation.md).

### Fifth

Only the implementation files relevant to the current agent
(docs/platform-repository-audit.md for the Platform baseline, plus the
protobuf/grpc files the agent will touch).

### Sixth

Relevant review/checkpoint documentation (this file,
docs/platform-repository-audit.md).

Do not recursively explore the repository before determining which files are
relevant.

## Previous Work

- Foundation implemented: domain model, repository layout, protobuf
  strategy, migration plan (docs/00-foundation).
- Clients Service fully implemented and reviewed (Agents 01–12): client/
  platform/integration/oauth/webhook domain, unified Provider interface,
  gRPC + gateway, PostgreSQL, Docker, tests.
- Transactions Service fully implemented and reviewed (Agents 01–13):
  database, SQLC, protobuf, repositories, merchants, customers, deposits,
  payouts, runtime, scaffolding, tests; final decision READY WITH CONDITIONS.
- Generated code is committed and reproducible: protobuf (5 contracts),
  gRPC stubs, gateway stubs, SQLC, mocks.
- Legacy deposits/ and integrations/ services remain runnable.
- Platform audit completed: baseline documented, findings assigned to
  agents 02–12.

## Recently Modified Files

Use `git status` and `git diff` at the current commit. The most recent
documented deliverable before this checkpoint is:
- docs/platform-repository-audit.md (Platform Agent 01).

Prior milestones (committed on the merged Transactions branch, now on `main`/
`platform-cleanup`):
- docs/transactions-production-review.md
- docs/transactions-tests-review.md
- docs/transactions-scaffolding-review.md
- transactions/ scaffolding + runtime + services + db (all committed)
- clients/docs/*.md (Clients reviews, committed)
- clients/ implementation (committed)

Generated files (grpc/go/*, *sqlc*, *mocks*) are committed and should only
change via regeneration; do not edit by hand.

## Do Not Trust This Checkpoint Blindly

"The checkpoint is a navigation and handoff document, not the ultimate source
of truth. When the checkpoint conflicts with the actual repository, README.md,
project documentation, or current source code, verify the repository and
authoritative documentation before proceeding."

## Checkpoint Maintenance

- Update this document when a major agent completes.
- Update the Current Work section when changing tasks.
- Update Known Issues when blockers are resolved.
- Update Next Action when the active agent changes.
- Do not allow this file to become a duplicate README or architecture
  document.