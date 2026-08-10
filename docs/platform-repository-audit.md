# Platform Repository Audit

Document Version: 1.0
Status: Complete
System: RVPay
Review: Platform Agent 01 — Repository Audit

## 1. Executive Summary

The RVPay repository contains the Deposits service (legacy, still runnable),
the Integrations service (legacy, still runnable), the Clients service (new,
fully implemented across the Clients agents), the Transactions service (new,
fully implemented across the Transactions agents), shared protobuf sources,
generated Go protobuf/gRPC/gateway code, deployment documentation, and CI/CD.

The Platform layer is **largely absent as an explicit shared layer**:
- There is **no** `shared/`, `pkg/`, `internal/`, `common/`, `middleware/`,
  `config/`, `logging/`, `errors/`, or `auth/` directory.
- Shared infrastructure (config loading, logger setup, database helpers,
  gRPC recovery, gateway registration, health checks) is **duplicated across
  the four services** rather than extracted.
- Protobuf generation, gateway registration, Docker, and Makefile conventions
  exist and are consistent, but are **not centralized** into a Platform layer.

This is the baseline the Platform agents (02–12) will build upon.

## 2. Scope

### Inspected
- Git state, repository root, README structure.
- `go.mod` and shared dependencies.
- Root Makefile (`test`, `bench-test`) and service Makefiles.
- CI workflows (`.github/workflows/deploy.yml`,
  `.github/workflows/render-deploy.yml`), `tools/versions.md`.
- Protobuf sources (`protobuf/*.proto`), `protobuf/Makefile`, generated code
  (`grpc/go/`).
- Dockerfiles (clients, transactions, deposits, integrations, protobuf),
  `docker-compose.yml`, `nginx/nginx.conf`.
- Render configuration (`render.yaml`), deployment documentation
  (`deploy/README.md`, `deploy/render/README.md`).
- Shared-package presence and gateway usage across the four services.

### Deliberately excluded
- `third_party/`, `third_party/googleapis/` (recursive contents).
- `vendor/`, `node_modules/`, `.git/`, `coverage/`, `tmp/`, `bin/`.
- Generated dependency trees and historical archives.
- Deep business-logic review of Clients/Transactions (owned by their agents).
- Any code, protobuf, SQL, migration, Dockerfile, Makefile, CI, or Render
  modification. **This agent is read/audit only.**

## 3. Required Documentation

| Document | Present | Read |
| --- | --- | --- |
| README.md | ✅ | ✅ |
| agents/project-context.md | ✅ | ✅ |
| docs/domain-model.md | ✅ | ✅ |
| docs/repository-layout.md | ✅ | ✅ |
| docs/protobuf-strategy.md | ✅ | ✅ |
| docs/migration-plan.md | ✅ | ✅ |

All required documents were read successfully.

## 4. Repository Baseline

```text
rvpay-go/
├── .clinerules, AGENTS.md, Makefile, README.md
├── .env.example, .dockerignore, .gitignore, .gitmodules
├── agents/                  # agent directives (clients, transactions, platform)
├── clients/                 # Clients service (new)
├── transactions/            # Transactions service (new)
├── deposits/                # Deposits service (legacy, runnable)
├── integrations/            # Integrations service (legacy, runnable)
├── protobuf/                # shared protobuf sources + Makefile + Dockerfile
├── grpc/go/                 # generated Go protobuf/gRPC/gateway output
├── third_party/googleapis/  # protoc dependency (submodule)
├── nginx/                   # nginx.conf (OCI TLS termination)
├── deploy/                  # deployment documentation
├── docker-compose.yml       # OCI Compose stack
├── render.yaml              # Render Blueprint
├── docs/                    # architecture + agent review documents
├── tools/versions.md        # toolchain version pinning
├── layout.md                # historical layout notes
```

### Git state
- Branch: `platform-cleanup` (HEAD aligned with `main`).
- `git status --short`: clean (no uncommitted files at audit start).
- Latest commits: `13-production-review`, `12-tests`, `11-scaffolding`,
  `10-runtime` (Transactions implementation) on the merged branch.

### Go module
- `go 1.26.5`.
- Direct shared dependencies: `github.com/rs/zerolog v1.35.1`,
  `github.com/ardanlabs/conf/v3 v3.13.0`, `github.com/joho/godotenv v1.5.1`,
  `github.com/jackc/pgx/v5 v5.10.0`, `github.com/golang-migrate/migrate/v4
  v4.19.1`, `github.com/google/uuid v1.6.0`, `go.uber.org/mock v0.6.0`,
  `google.golang.org/grpc v1.83.0`, `google.golang.org/protobuf v1.36.11`,
  `google.golang.org/genproto/googleapis/api` (latest),
  `github.com/grpc-ecosystem/grpc-gateway/v2 v2.22.0` (indirect),
  `github.com/grpc-ecosystem/go-grpc-middleware v1.4.0`,
  `github.com/I-Frostbyte/pawapay_client` (PawaPay client).

## 5. Platform Component Inventory

| Component | Exists? | Location | Current State | Later Agent |
| --- | --- | --- | --- | --- |
| Protobuf generation | ✅ | `protobuf/` + `protobuf/Makefile` | Works; `make generate-protos` | 02 |
| HTTP gateway | ✅ (per-service) | `clients|transactions|deposits|integrations/cmd/grpc-service/main.go` | Duplicated registration per service | 03 |
| Common packages | ❌ | No `shared/`, `pkg/`, `internal/` | Not extracted; duplicated per service | 04 |
| CI/CD | ✅ (partial) | `.github/workflows/deploy.yml`, `render-deploy.yml` | Render pipeline active; OCI disabled | 05 |
| Docker | ✅ | per-service `Dockerfile`, `docker-compose.yml` | Multi-stage distroless per service; Compose OCI | 06 |
| Render | ✅ (deposits only) | `render.yaml` | Blueprint deploys only `rvpay-deposits` | 07 |
| Documentation | ✅ | `deploy/`, `docs/`, `README.md`, `tools/versions.md` | Deployment + architecture + version docs exist | 08 |
| Observability | ✅ (minimal) | per-service `/healthz`, gRPC health, zerolog | No metrics/tracing/request IDs | 09 |
| Security | ❌ | — | No auth middleware; secrets via env only | 10 |
| Performance | ✅ (basic) | pgxpool, sqlc, indexes | No caching/benchmarks | 11 |

## 6. Service Boundary Findings

- **Deposits** (legacy) — owns `deposits/`, `depositsgrpc`, its own
  config/db/repo, PawaPay client, Dockerfile, Makefile, README,
  `.env.example`.
- **Integrations** (legacy) — owns `integrations/`, `integrationsgrpc`, OAuth
  + webhook logic, Dockerfile, Makefile, README, `.env.example`.
- **Clients** (new) — owns `clients/`, `clientsgrpc`, platforms/integrations/
  oauth/webhooks, Dockerfile, Makefile, README, `.env.example`.
- **Transactions** (new) — owns `transactions/`, `transactionsgrpc`,
  merchants/customers/deposits/payouts, Dockerfile, Makefile, README,
  `.env.example`.

**Shared dependencies:** all services import the shared generated
`grpc/go/` packages (`depositsgrpc`, `integrationsgrpc`, `clientsgrpc`,
`transactionsgrpc`, `commongrpc`) and all depend on `zerolog`, `conf`,
`godotenv`, `pgx`, `golang-migrate`. No service imports another service's
business-implementation package (e.g. `clients` does not import `transactions`
business packages). **Service boundaries are respected.**

## 7. Protobuf Infrastructure

- **Sources:** `protobuf/deposits.proto` (`depositsgrpc`),
  `protobuf/integrations.proto` (`integrationsgrpc`),
  `protobuf/clients.proto` (`clientsgrpc`), `protobuf/transactions.proto`
  (`transactionsgrpc`), `protobuf/common.proto` (`commongrpc` shared types:
  `Provider`, `PaymentType`, `Money`, `Pagination`, status enums).
- **Imports:** `google/api/annotations.proto` (from `third_party/googleapis`),
  `google/protobuf/timestamp.proto`, and `common.proto` cross-package import.
- **Generation:** `protobuf/Makefile` (`make generate-protos`) iterates all
  `.proto` files, writing to `grpc/go/<package>/` with
  `paths=source_relative`. `protobuf/Dockerfile` exists for generated-code
  image builds.
- **Generated output:** committed. `grpc/go/{deposits,integrations,clients,
  transactions}grpc/*.pb.go`, `*_grpc.pb.go`, `*.pb.gw.go`; `grpc/go/commongrpc/`.
- **Toolchain:** `tools/versions.md` pins protoc v3.21.12, protoc-gen-go
  v1.36.10, protoc-gen-go-grpc v1.5.1, protoc-gen-grpc-gateway v2.22.0.
- **Compatibility note:** the `clients_proto`/`transactions_proto` package
  names and `go_package` options match the module; `deposits.proto`/
  `integrations.proto` remain for the legacy services.

## 8. HTTP Gateway

- grpc-gateway v2 is used by **all four services**, each wiring
  `runtime.NewServeMux` + `Register*HandlerServer` inside
  `cmd/grpc-service/main.go` (evidence: `grep -rl "runtime.NewServeMux"`).
- The HTTP server pattern (`PORT` env, default `8080`, `/healthz`) is
  duplicated across the four `main.go` files.
- No centralized gateway helper or shared middleware exists.

## 9. Common Packages

- **No shared code directory exists.** `internal/`, `pkg/`, `common/`,
  `shared/`, `middleware/`, `config/`, `logging/`, `errors/`, `auth/` are all
  absent at the repository root.
- Config loading (`ardanlabs/conf` + `godotenv`), logger setup (zerolog),
  gRPC recovery interceptor, health server, gateway+healthz bootstrap, and
  pgxpool+migration bootstrap are **duplicated** across `clients`,
  `transactions`, `deposits`, and `integrations`.
- Dependency direction is currently: services use only the generated
  protobuf packages and imports; the desired
  `business services → shared platform infrastructure` direction has no
  shared target because no shared package exists yet.
- This is the primary gap addressed by Platform Agent 04 (Common packages).

## 10. CI/CD

- **`.github/workflows/render-deploy.yml`** (active, `push` on `main` +
  `workflow_dispatch`): recursive submodule checkout → Go 1.26.5 → install
  protoc/plugins (pinned) → `go generate ./...` → `make generate-protos` →
  sqlc (v1.29.0 via `go run` for `deposits/db` only) →
  `git diff --exit-code` on generated protobuf/sqlc/mocks → tests → Docker →
  Render deploy hook.
- **`.github/workflows/deploy.yml`** (OCI; `on: workflow_dispatch` only, so
  effectively **disabled**): Go 1.26.5 → `go test ./...` → build ARM64
  `deposits` image → push GHCR → SSH restart of the OCI Compose stack.
- **Generated-code verification:** render-deploy verifies protobuf + sqlc +
  mocks via `git diff --exit-code`. Note: the sqlc step runs only
  `deposits/db`; `clients/db` and `transactions/db` sqlc regeneration is
  covered by `go generate ./...` earlier in the job (their doc.go pins
  sqlc@v1.29.0), **but the diff check path is `deposits/db/sqlc` only** —
  see Findings.
- `tools/versions.md` is read by the Render workflow to pin toolchain
  versions.

## 11. Docker

- **Dockerfiles:** `deposits/Dockerfile`, `integrations/Dockerfile`,
  `clients/Dockerfile`, `transactions/Dockerfile`, `protobuf/Dockerfile`.
- All service Dockerfiles follow the **multi-stage distroless** pattern:
  `golang:1.26.5-alpine` build stage → `gcr.io/distroless/static-debian12:nonroot`
  runtime, `CGO_ENABLED=0`, `-trimpath -ldflags="-s -w"`, non-root user,
  `EXPOSE 50051`, migrations copied.
- **`docker-compose.yml`** (OCI): services include PostgreSQL 16,
  golang-migrate (`migrate/migrate:v4.19.1`), the Deposits service
  (`${DEPOSITS_IMAGE:-rvpay/deposits:local}`), and nginx unprivileged.
- **`nginx/nginx.conf`** exists for OCI TLS termination.
- Build context is the repository root for all service Dockerfiles.

## 12. Render

- **`render.yaml`** is a Render Blueprint (`type: web`, `runtime: docker`,
  `dockerfilePath: ./deposits/Dockerfile`, `dockerContext: .`).
- It deploys **only the legacy Deposits service** (`rvpay-deposits`), wired to
  a managed PostgreSQL database, with env vars for `LISTEN_PORT`,
  `MIGRATION_PATH`, `RUN_MIGRATIONS`, DB_*, `PAWAPAY_API_URL` (sync:false),
  and a `/healthz` health check.
- **Clients and Transactions are NOT yet wired into Render.**
- `deploy/render/README.md` documents the Render deployment strategy.

## 13. Documentation

- `docs/` contains architecture documents (`domain-model.md`,
  `repository-layout.md`, `protobuf-strategy.md`, `migration-plan.md`) and the
  per-agent review documents (stable baseline).
- `deploy/README.md` and `deploy/render/README.md` document deployment.
- `tools/versions.md` documents the toolchain (protoc/plugins/sqlc/mockgen).
- Each service has a `README.md` (deposits, integrations, clients,
  transactions have service-level setup/config/run documentation).

## 14. Observability

- **Logging:** `zerolog` everywhere (structured, timestamp+caller).
- **Health checks:** gRPC health server + HTTP `/healthz` per service.
- **Metrics:** none observed.
- **Tracing:** none observed.
- **Request IDs:** none observed.
- **Error reporting:** none observed.

## 15. Security

- **Secrets:** no committed credentials; `.env.example` files use
  placeholders; secrets are environment-provided (DB passwords, PawaPay key,
  Render/OCI secrets referenced by name only).
- **Authentication:** none observed in any service (no auth middleware;
  matches the Deposits legacy convention). Confirmed as an open HIGH item in
  the Transactions production review (F-03) and relevant to Clients too.
- **Authorization:** none observed.
- **SQL safety:** all SQL goes through sqlc (parameterized); no string-
  concatenated SQL observed.
- **TLS:** nginx.conf exists (OCI); DB `sslmode` selectable per service.

## 16. Performance

- **Database:** pgxpool connection pooling + eager Ping; sqlc-generated
  queries; sensible indexes on ownership/status/reference/created_at columns.
- **HTTP/gRPC:** gRPC connection reuse via `grpc.Dial`; gateway reuses the same
  service instance.
- **Caching:** none observed.
- **Background workers:** none observed (no polling/reconciliation jobs).
- **Pagination:** present in protobuf for merchants only; deposits/payouts/
  customers/customers lists are not paginated (documented MEDIUM findings).
- No benchmarking or optimization performed (out of scope).

## 17. Cross-Service Consistency

| Area | Observations |
| --- | --- |
| Logging | Consistent — zerolog with timestamp+caller across all services |
| Config | Consistent — `ardanlabs/conf` + `godotenv`, same env naming (`LOG_LEVEL`, `LISTEN_PORT`, `MIGRATION_PATH`, `RUN_MIGRATIONS`, `DB_*`) |
| Server startup | Consistent — `main() → run()`, signal context, pgxpool+Ping, migrate, gRPC+gateway+healthz, graceful shutdown |
| Protobuf | Consistent — per-service package + shared `commongrpc` |
| Docker | Consistent — multi-stage distroless, same Go version, same EXPOSE |
| Gateway | Consistent but duplicated — each `main.go` wires its own mux |
| Makefiles | Mostly consistent — Deposits `test` target is broken (`find | xargs`); Transactions fixed it (`go test ./... --cover`); Clients/Integrations/Deposits still use the old broken variant |

## 18. Findings

| ID | Severity | Area | Evidence | Impact | Responsible Agent |
| --- | --- | --- | --- | --- | --- |
| P-01 | HIGH | Common packages | No `shared/`, `pkg/`, `internal/`, `middleware/`, `config/`, `logging/`, `errors/` directory exists; config/logger/gateway/database bootstrap is copied across clients, transactions, deposits, integrations `main.go` | Duplication and drift risk across four services; Platform layer absent | 04 (Common packages) |
| P-02 | MEDIUM | CI/CD | `render-deploy.yml` sqlc step runs only `deposits/db` and the `git diff --exit-code` path covers only `deposits/db/sqlc`; `clients/db` and `transactions/db` sqlc output is regenerated via `go generate` but **not** verified by the explicit diff scope | Possible unverified sqlc drift for clients/transactions | 05 (CI/CD) |
| P-03 | MEDIUM | Render | `render.yaml` deploys only the legacy Deposits service; Clients and Transactions have no Render service entries | New services are not deployable via Render yet | 07 (Render) |
| P-04 | MEDIUM | Gateway | grpc-gateway + `/healthz` + HTTP bootstrap is duplicated in all four `cmd/grpc-service/main.go` files | Operational drift; no shared gateway/middleware | 03 (HTTP gateway), 04 |
| P-05 | LOW | Makefiles | `deposits/Makefile` (and copied variants in Clients/Integrations) `test` target uses a broken `find | xargs` command; Transactions fixed it to `go test ./... --cover` | `make test` is broken in three services | 05 (CI/CD) |
| P-06 | LOW | Observability | No metrics, tracing, request IDs, or error reporting infrastructure anywhere | Limited production observability | 09 (Observability) |
| P-07 | LOW | Security | No authentication/authorization middleware in any service (matches legacy convention and prior reviews F-03) | Public endpoints unauthenticated | 10 (Security) |
| P-08 | INFO | Toolchain | `tools/versions.md` exists and pins protoc/plugins/sqlc/mockgen; render-deploy workflow installs the pinned versions | Good reproducibility baseline | 02 (should preserve) |
| P-09 | INFO | Docker | All four service Dockerfiles follow a consistent multi-stage distroless pattern; docker-compose OCI stack is deposits-only | Consistent baseline; Compose needs new services when deployed on OCI | 06 (Docker) |
| P-10 | INFO | Performance | pgxpool, sqlc, indexes present; no caching or background workers | Fine for current scope | 11 (Performance) |

## 19. Platform Risk Register

| ID | Area | Severity | Evidence | Affected Component | Responsible Agent |
| --- | --- | --- | --- | --- | --- |
| P-01 | Common packages | HIGH | No shared infrastructure directory; duplicates across 4 services | Config/logger/db/gateway bootstrap | 04 |
| P-02 | CI/CD | MEDIUM | sqlc diff scope covers deposits only | Generated-code sync for clients/transactions | 05 |
| P-03 | Render | MEDIUM | render.yaml lists only deposits | Deployability of new services | 07 |
| P-04 | Gateway | MEDIUM | Duplicated mux/healthz setup | All services | 03 |
| P-05 | Makefiles | LOW | Broken `test` target in 3 services | Developer experience | 05 |
| P-06 | Observability | LOW | No metrics/tracing/request IDs | Operations | 09 |
| P-07 | Security | LOW | No auth middleware | Public endpoints | 10 |
| P-08 | Toolchain | INFO | Versions pinned; CI installs them | Protobuf generation | 02 |
| P-09 | Docker | INFO | Consistent distroless pattern; Compose deposits-only | Deployment | 06 |
| P-10 | Performance | INFO | Pooling/sqlc/indexes; no caching/workers | Performance | 11 |

## 20. Recommended Agent Order

The repository does not reveal a concrete dependency requiring deviation from
the prescribed order:

1. Repository Audit (this agent — complete)
2. Protobuf Generation (02) — preserves/verifies toolchain and generated code
3. HTTP Gateway (03) — centralizes the duplicated gateway bootstrap
4. Common Packages (04) — extracts config/logger/db/middleware shared code
5. CI/CD (05) — fixes sqlc diff scope and broken Makefile test targets
6. Docker (06) — verifies/extends the consistent Docker conventions
7. Render (07) — adds Clients/Transactions to the Render Blueprint
8. Documentation (08) — updates README/deploy docs for the new services
9. Observability (09) — adds metrics/tracing/request ID infrastructure
10. Security (10) — adds authentication/authorization middleware
11. Performance (11) — reviews pooling/caching/pagination
12. Final Review (12)

## 21. Out-of-Scope Areas

Deliberately not inspected recursively:

- `third_party/`, `third_party/googleapis/`
- `vendor/`, `node_modules/`, `.git/`, `coverage/`, `tmp/`, `bin/`
- Generated dependency trees (deep `grpc/go/*.pb.go` internals, sqlc internals,
  mocks)
- Historical archives (`layout.md` contents, `agents/z-older-commands/`)
- Business-logic internals of Clients/Transactions/Deposits/Integrations

## 22. Documentation Check

All six required documents (README.md, agents/project-context.md,
docs/domain-model.md, docs/repository-layout.md, docs/protobuf-strategy.md,
docs/migration-plan.md) were present and read successfully. In addition,
`tools/versions.md`, `deploy/README.md`, `deploy/render/README.md`, the
individual service READMEs, and the CI workflow files were inspected as
relevant infrastructure.

## 23. Final Repository State

`git status --short` at the end of this audit (before creating this report)
was clean. The only file created by this agent is
`docs/platform-repository-audit.md`. No source code, protobuf, SQL,
migration, Dockerfile, Makefile, CI, Render, or service-modification files
were changed.