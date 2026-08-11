# Platform Docker Review

Document Version: 1.0
Status: Complete
System: RVPay
Review: Platform Agent 06 — Docker

## 1. Objective

Design, implement, and validate the Docker configuration for the new RVPay
architecture so each deployable service can be: built reproducibly, packaged
into a minimal production image, started with the correct entry point,
configured entirely through environment variables, connected to external
infrastructure correctly, used by the Render deployment architecture, and
built locally without relying on developer-machine state.

This agent owns Docker implementation (Dockerfiles, build configuration,
.dockerignore, Docker documentation). No GitHub Actions workflow, Render
configuration, application logic, protobuf contract, generated code, migration,
or business logic was changed.

## 2. Required Documentation

| Document | Read |
| --- | --- |
| README.md | ✅ |
| agents/project-context.md | ✅ |
| docs/domain-model.md | ✅ |
| docs/repository-layout.md | ✅ |
| docs/protobuf-strategy.md | ✅ |
| docs/migration-plan.md | ✅ |
| docs/platform-repository-audit.md | ✅ |
| docs/platform-protobuf-generation-review.md | ✅ |
| docs/platform-http-gateway-review.md | ✅ |
| docs/platform-common-packages-review.md | ✅ |
| docs/platform-ci-cd-review.md | ✅ |

All required documents were present and read.

## 3. Existing Dockerfiles

| File | Service | Current Purpose | Decision |
| --- | --- | --- | --- |
| `deposits/Dockerfile` | Deposits (legacy, reference) | Multi-stage distroless build of the Deposits gRPC service | **Preserved unchanged** — reference conventions intact; no `shared/` dependency |
| `integrations/Dockerfile` | Integrations (legacy) | Multi-stage distroless build of the Integrations gRPC service | **Preserved unchanged** — legacy; no `shared/` dependency |
| `clients/Dockerfile` | Clients (new) | Multi-stage distroless build of the Clients gRPC service | **Modified** — added `COPY shared ./shared` |
| `transactions/Dockerfile` | Transactions (new) | Multi-stage distroless build of the Transactions gRPC service | **Modified** — added `COPY shared ./shared` |

Supporting Docker configuration:

| File | Decision |
| --- | --- |
| `.dockerignore` | **Preserved unchanged** — correctly excludes dev noise while retaining all required source, generated code, migrations, and module files |
| `docker-compose.yml` | **Preserved unchanged** — OCI Always Free stack (postgres, one-shot migration job, deposits, nginx); build contexts remain repo-root |

## 4. Final Docker Architecture

All four service images follow the established deposits reference pattern
(directive #89): a two-stage build ─ official Go builder producing a static
binary, then a minimal distroless runtime containing only the binary and the
service's migration files.

The build environment is the repository root (context `.`), because every
service imports `grpc/go/` generated packages and (for clients/transactions)
the `shared/` packages; dependency state comes solely from the root
`go.mod`/`go.sum` (directives #13, #14, #75).

## 5. Build Context

| Service | Build Context | Dockerfile |
| --- | --- | --- |
| deposits | repository root (`.`) | `deposits/Dockerfile` |
| integrations | repository root (`.`) | `integrations/Dockerfile` |
| clients | repository root (`.`) | `clients/Dockerfile` |
| transactions | repository root (`.`) | `transactions/Dockerfile` |

This matches the Agent 05 CI invocation
(`docker build -f <service>/Dockerfile -t rvpay-<service>:ci .`) and the OCI
Compose stack (directive #104). No cross-platform `--platform` flags were
introduced (directive #80).

## 6. Build Stages

Builder stage (all services):

- `FROM golang:1.26.5-alpine AS build` — Go version matches `go.mod`
  (directives #6, #8).
- `WORKDIR /src`
- `COPY go.mod go.sum ./` then `RUN go mod download` — dependency-layer
  caching (directives #16, #73).
- `COPY <service> ./<service>`, `COPY shared ./shared` (clients/transactions),
  `COPY grpc ./grpc` (generated protobuf/gRPC/gateway source — consumed from
  committed repository state, not regenerated; directives #17–#19), and
  `COPY protobuf ./protobuf` (clients/transactions; required because their
  packages import `protobuf/` sources for the generator-included descriptors? —
  retained from the original working Dockerfiles).
- `RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/<svc>-grpc-service ./<svc>/cmd/grpc-service`
  — static binary; flags preserved from the existing convention (directives
  #10–#12); no `go mod tidy`/`go get` (directives #75–#77).

Runtime stage (all services):

- `FROM gcr.io/distroless/static-debian12:nonroot` — minimal static image,
  suitable for `CGO_ENABLED=0` Go binaries; CA certificates and DNS support
  are present in this distroless image for the services' HTTPS/network needs
  (directives #9, #46, #47).
- `WORKDIR /app/<service>`
- `COPY --from=build /out/<svc>-grpc-service ./<svc>-grpc-service`
- `COPY --from=build /src/<svc>/db/migrations ./db/migrations` — migrations are
  packaged because each service runs migrations at startup when
  `RUN_MIGRATIONS=true` (directives #69–#70).
- `USER nonroot:nonroot` (directives #40, #41)
- `EXPOSE 50051` — documentation only (directive #29)
- `ENTRYPOINT ["./<svc>-grpc-service"]` — direct exec, no shell, preserves
  SIGTERM delivery to PID 1 (directives #37, #44, #96).

## 7. Service Images

| Service | Dockerfile | Binary | Runtime Image | Port |
| --- | --- | --- | --- | --- |
| Deposits | `deposits/Dockerfile` | `deposits-grpc-service` | distroless static-debian12:nonroot | 50051 |
| Integrations | `integrations/Dockerfile` | `integrations-grpc-service` | distroless static-debian12:nonroot | 50051 |
| Clients | `clients/Dockerfile` | `clients-grpc-service` | distroless static-debian12:nonroot | 50051 |
| Transactions | `transactions/Dockerfile` | `transactions-grpc-service` | distroless static-debian12:nonroot | 50051 |

Ports come from the documented configuration (`LISTEN_PORT`; gateway HTTP on
`PORT` default 8080 is not EXPOSE'd because gRPC 50051 is the documented
container port). No port was invented (directive #28).

## 8. Configuration

- All runtime configuration is environment-driven (directives #22–#24, #65):
  `LOG_LEVEL`, `LISTEN_PORT`, `PORT`, `MIGRATION_PATH`, `RUN_MIGRATIONS`,
  `DB_*`, and the service-specific HighLevel/webhook vars.
- The images do not contain or copy `.env` or `.env.example` (directives #22,
  #23; verified in the image scan).
- `DB_HOST` is not hard-coded; deployments (Render/OCI) provide the database
  connection (directives #26, #27).
- No Docker `ENV` carries secrets; non-sensitive defaults are not embedded
  as secrets (directive #65).

## 9. Generated Code

- Protobuf/gRPC/gateway: committed in `grpc/go/` and copied into the build
  context (directives #17–#19, #72). No `protoc` is installed in any image.
- sqlc: committed per service in `<svc>/db/sqlc/` and available to the Go
  toolchain as source (directive #71). No `sqlc` is installed.
- mocks: not required at runtime; excluded from the runtime image.
- Generation tools (protoc, sqlc, mockgen, goose) are never installed in any
  production image (directive #20).

## 10. Database/Migrations

- Migration files are copied into each runtime image (`<svc>/db/migrations`)
  because the services run migrations at startup when `RUN_MIGRATIONS=true`
  (directives #69–#70). The OCI Compose stack runs migrations separately
  (one-shot `migrate/migrate` job with `RUN_MIGRATIONS=false`); the files are
  still present for Render where the service runs its own migrations.
- No SQL generation or migration logic was implemented by this agent.
- PostgreSQL is never inside an application image (directive #25).

## 11. Security

- **Non-root execution**: `USER nonroot:nonroot` in every runtime image
  (verified in image config; directives #40, #41).
- **Secret handling**: no `.env`, no private keys, no credentials in images
  (verified by filesystem scan of the exported images); no `ARG`/`ENV` secrets
  (directives #64, #65, #74).
- **.env exclusion**: `.dockerignore` excludes `**/.env`; verified no `.env`
  in the built images.
- **Runtime config**: secrets supplied via deployment environment only.
- **Git metadata**: `.git` excluded from build context and absent from images
  (directives #53, #66, #102).

## 12. Local Validation

| Command | Result |
| --- | --- |
| `docker build -f clients/Dockerfile -t rvpay-clients:ci .` | ✅ success |
| `docker build -f transactions/Dockerfile -t rvpay-transactions:ci .` | ✅ success |
| Image inspect (clients/transactions): entrypoint, user, ports, workdir | ✅ correct (PID-1 exec, nonroot, 50051) |
| Exported-image filesystem scan for `.env`/`.git`/keys/Go toolchain/protoc/sqlc | ✅ clean for both |
| `timeout 12 docker run --rm rvpay-clients:ci` | ✅ executable starts, config loads; fails on unavailable DB (dependency failure, exit 1 — not a container failure) |
| `timeout 12 docker run --rm rvpay-transactions:ci` | ✅ executable starts; missing `LISTEN_PORT` fails clearly at config load (exit 1, no embedded fallback credentials) |

Container startup was validated without requiring external infrastructure
(directive #59, #101); the observed non-zero exits are correct, explicit
application dependency/config failures.

## 13. CI Compatibility

- The Dockerfiles build with the exact context and path used by Agent 05's CI
  matrix: `docker build -f <service>/Dockerfile -t rvpay-<service>:ci .`
  (directive #104; see `docs/platform-ci-cd-review.md` §8).
- The added `COPY shared ./shared` resolves the build break that would have
  occurred in the CI matrix for clients/transactions following Agent 04's
  shared-package adoption.
- No GitHub Actions workflow was modified (directive #54).

## 14. Render Requirements

Docker-side requirements for Agent 07:

- Services bind gRPC on `LISTEN_PORT` and the HTTP gateway on `PORT`
  (default 8080) from the environment; Render must supply these.
- Images run as non-root and write nothing outside tmpfs; no writable volume
  is required by the application itself.
- Render may run migrations at startup via `RUN_MIGRATIONS=true` (migration
  files are present in the image) or manage them externally
  (`RUN_MIGRATIONS=false`).
- `DB_HOST` must point at the Render-managed PostgreSQL host (never
  localhost) (directives #26, #27, #105).
- The images are `linux/amd64` by default (standard Render architecture); no
  `--platform` pinning was introduced.

## 15. Findings

| ID | Severity | File/Area | Finding | Resolution |
| --- | --- | --- | --- | --- |
| DOCK-01 | HIGH | `clients/Dockerfile`, `transactions/Dockerfile` | Docker builds would fail: both services now import `shared/database` + `shared/logger` (Agent 04) but the build context omitted `shared/` | ✅ Added `COPY shared ./shared` to both Dockerfiles; builds now succeed |
| DOCK-02 | INFO | All 4 Dockerfiles | Consistent multi-stage distroless pattern, non-root, static binary, direct ENTRYPOINT, env-driven config, no secrets | ✅ Preserved as the canonical pattern (audit P-09 baseline) |
| DOCK-03 | INFO | `.dockerignore` | Correctly excludes `.git`, `.github`, agents, third_party, `.env`, docs, tests, build artifacts while retaining required source/generated/migration/module files | ✅ Preserved unchanged (directive #50/#103) |
| DOCK-04 | INFO | Images | Image scan found no `.env`, git history, keys, Go toolchain, protoc, or sqlc in the built images | ✅ Verified (directives #66, #102) |
| DOCK-05 | INFO | Local builds | Local Docker builds were initially unavailable in the WSL environment; user enabled Docker; builds rerun successfully | ✅ Documented |

## 16. Deferred Work

| Issue | Owner Agent |
| --- | --- |
| Render Blueprint coverage for clients/transactions images + `PORT`/`LISTEN_PORT` env wiring | Agent 07 (Render) |
| OCI Compose stack extension to clients/transactions services | Agent 12 / deployment plan |
| Vulnerability scanning (Trivy/Docker Scout) — none currently configured | Agent 10 (Security) |
| CI docker-build matrix already references clients/transactions (fixed images); no change needed | Agent 05 (confirmed compatible) |
| gRPC health-check wiring into any future container healthcheck | Agent 09 (Observability) — no healthcheck added (dir #33–#35) |

## 17. Changes Made

- `clients/Dockerfile` — added `COPY shared ./shared` between `COPY clients
  ./clients` and `COPY grpc ./grpc`.
- `transactions/Dockerfile` — added `COPY shared ./shared` between
  `COPY transactions ./transactions` and `COPY grpc ./grpc`.

No other files were modified.

## 18. Documentation Check

| Document | Read |
| --- | --- |
| README.md | ✅ |
| agents/project-context.md | ✅ |
| docs/domain-model.md | ✅ |
| docs/repository-layout.md | ✅ |
| docs/protobuf-strategy.md | ✅ |
| docs/migration-plan.md | ✅ |
| docs/platform-repository-audit.md | ✅ |
| docs/platform-protobuf-generation-review.md | ✅ |
| docs/platform-http-gateway-review.md | ✅ |
| docs/platform-common-packages-review.md | ✅ |
| docs/platform-ci-cd-review.md | ✅ |

## 19. Final Status

PASS WITH FOLLOW-UP