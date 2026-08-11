# Platform CI/CD Review

Document Version: 1.0
Status: Complete
System: RVPay
Review: Platform Agent 05 — CI/CD

## 1. Objective

Review, implement, and stabilize the CI/CD pipeline for the RVPay architecture so
the repository can reliably: validate the Go codebase, generate required code,
run tests, verify generated code is current, build the services, build Docker
images, detect integration/build regressions, and provide deterministic feedback
before deployment.

This agent owns CI/CD workflow implementation and validation. No Dockerfile,
Render service configuration, protobuf contract, generated code, migration, or
business logic was modified.

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

All required documents were present and read.

## 3. Existing Workflows

| Workflow | Purpose | Decision |
| --- | --- | --- |
| `.github/workflows/render-deploy.yml` | Active Render delivery pipeline: go generate → protoc → sqlc → drift check → tests → Docker (deposits only) → Render deploy hook (push to main + workflow_dispatch) | **Modified** — extended drift scope, added validate job, Docker matrix, explicit deploy-skip signaling |
| `.github/workflows/deploy.yml` | OCI delivery pipeline: test → build/push ARM64 deposits image to GHCR → SSH restart of OCI Compose stack. Disabled by design (no `on:` trigger beyond workflow_dispatch). | **Preserved unchanged** — intentionally disabled; documented |

`deploy.yml` was preserved per directive #2/#68 (no duplicate deployment
pipelines; the OCI pipeline is a distinct target currently disabled).

## 4. Final Pipeline

```text
generate (go generate → protoc → sqlc → generated-code drift verification)
    ↓
validate (gofmt → go vet → go build)
    ↓
test (go test ./...)
    ↓
docker-build (matrix: deposits, clients, transactions)
    ↓
deploy (Render deploy hook, gated to main; explicit skip when secret absent)
```

- `generate` and `validate`/`test` ordering ensures generated code exists before
  compilation/testing (directives #40, #41).
- `docker-build` requires `validate` (build correctness) and `test` guarantees
  are checked in parallel; `deploy` requires `docker-build` (directive #46).
- Deployment only occurs after generation, validation, tests, and Docker
  validation succeed (directive #47).
- `workflow_dispatch` respects the same dependencies (directive #48).
- No job uses `continue-on-error`, `|| true`, or retries (directives #34, #98,
  #99). Commands fail the job by default.

## 5. Toolchain Versions

| Tool | Version | Source |
| --- | --- | --- |
| Go | 1.26.5 | `go.mod`, existing CI, README |
| protoc | 3.21.12 (Ubuntu `protobuf-compiler`) | `tools/versions.md` |
| protoc-gen-go | v1.36.10 | `tools/versions.md` (pinned `go install`) |
| protoc-gen-go-grpc | v1.5.1 | `tools/versions.md` (pinned `go install`) |
| protoc-gen-grpc-gateway | v2.22.0 | `tools/versions.md` (pinned `go install`) |
| sqlc | v1.29.0 | `deposits/db/doc.go` (pinned `go run`) |

No `@latest` versions are used (directives #7, #37). GitHub Actions use stable
official actions `actions/checkout@v4`, `actions/setup-go@v5` with caching
(directives #31, #32, #36, #85).

## 6. Generation

- **go generate**: `go generate ./...` runs the documented per-service
  `go:generate` directives (sqlc + mockgen, pinned versions in each
  `db/doc.go`).
- **Protobuf generation**: `cd protobuf && make generate-protos` (canonical
  target, produces `grpc/go/` output; directive #6).
- **sqlc generation**: `go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate`
  in `deposits/db` (preserved approach; directive #9).
- **Generated-code verification**: `git diff --exit-code` scoped to every
  generated directory:
  - Protobuf/gRPC/gateway: `grpc/go`
  - sqlc + mocks: `deposits/db/sqlc`, `deposits/db/repo/mocks`,
    `deposits/db/sqlc/mocks`, `clients/db/sqlc`, `clients/db/repo/mocks`,
    `clients/db/sqlc/mocks`, `transactions/db/sqlc`,
    `transactions/db/repo/mocks`, `transactions/db/sqlc/mocks`
  - Failure produces actionable messages (directives #8, #10, #35, #66):
    protobuf → "Run 'cd protobuf && make generate-protos' and commit the
    generated files."; sqlc/mocks → "Run 'go generate ./...' and commit the
    generated files."

This resolves audit finding P-02 (clients/transactions sqlc drift was previously
unverified).

## 7. Testing

- **Unit tests**: `go test ./...` in the `test` job (repository-wide, includes
  `shared/` packages from Agent 04; directives #12, #56, #59).
- **Race detection**: not enabled — the project has no documented race-detection
  requirement and the existing test suite is small; avoiding unbounded CI
  runtime per directive #14.
- **Static analysis**: `go vet ./...` in the `validate` job (directive #16);
  no new lint framework introduced (directive #15).
- **Formatting**: `gofmt -l` over hand-written Go files only (directive #17).
  Generated output (`grpc/go/`, `/db/sqlc/`, `/mocks/`, `*.pb.go`,
  `*.pb.gw.go`) is excluded so CI never rewrites generated files (directive
  #77). Verified the committed hand-written Go is gofmt-clean.
- **Test isolation**: no integration tests require PostgreSQL/Render/HighLevel
  (directives #13, #42, #94). The repository's test suite runs without external
  infrastructure.

## 8. Build

- **Go build validation**: `go build ./...` in the `validate` job compiles all
  service entry points (`clients/cmd/grpc-service`,
  `transactions/cmd/grpc-service`, plus legacy `deposits`/`integrations`) and
  all packages (directives #19, #20).
- **Docker validation**: matrix over `deposits`, `clients`, `transactions`
  Dockerfiles (build context = repository root, matching the Dockerfiles'
  expectations; directive #90). `integrations/` is legacy and not part of the
  documented new-architecture deployment, so its image is not built here
  (directive #23). No Dockerfile content was changed (directive #22, #3).

## 9. Deployment

- **Trigger**: Render deploy hook via `curl -fsS -X POST` on `push` to `main`
  and `workflow_dispatch` (directives #24, #49).
- **Branch restrictions**: deploy job gated with
  `if: github.ref == 'refs/heads/main'` (directive #29).
- **Required secrets**: `RENDER_DEPLOY_HOOK` (from GitHub secrets, never
  hard-coded; directive #25, #26).
- **Failure behavior**: if the hook request fails, the job fails (curl `-f`);
  if the secret is absent, the job emits
  `::notice::Render deploy skipped: RENDER_DEPLOY_HOOK secret is not set.` and
  exits 0 — making the skip explicit in the Actions UI rather than silently
  reporting success (directives #27, #50).

## 10. Security

- **Permissions**: `contents: read` at workflow level; no `write-all`
  (directives #30, #83).
- **Secrets**: only `RENDER_DEPLOY_HOOK` is referenced, via `env:`, never
  echoed or printed (directives #26, #72, #73, #74).
- **Sensitive logs**: no credentials, database URLs, or deploy hooks are
  printed by any step (directives #72, #73).

## 11. Findings

| ID | Severity | File/Area | Finding | Resolution |
| --- | --- | --- | --- | --- |
| CICD-01 | MEDIUM | `render-deploy.yml` | sqlc drift check covered deposits only (audit P-02); clients/transactions sqlc+mocks drift unverified | ✅ Extended `git diff --exit-code` scope to all services' sqlc + mocks with actionable failure messages |
| CICD-02 | LOW | `render-deploy.yml` | No `go vet`, no formatting check, no `go build ./...` in CI (only tests + deposits Docker) | ✅ Added `validate` job: gofmt (hand-written only), `go vet ./...`, `go build ./...` |
| CICD-03 | LOW | `render-deploy.yml` | Docker validation built only deposits; clients/transactions images unverified | ✅ Added matrix over deposits/clients/transactions Dockerfiles |
| CICD-04 | LOW | `render-deploy.yml` | Render deploy silently skipped success when `RENDER_DEPLOY_HOOK` unset (directive #50) | ✅ Emits explicit `::notice::` "deploy skipped" annotation |
| CICD-05 | INFO | `render-deploy.yml` | Generated-code failure messages were generic | ✅ Split into protobuf vs sqlc/mocks with exact local commands (directive #66) |
| CICD-06 | INFO | `deploy.yml` | OCI pipeline is intentionally disabled (no `on:` trigger) | Documented; preserved unchanged |
| CICD-07 | INFO | Local developer machine | `core.autocrlf=true` causes local gofmt to flag CRLF files; CI runners are LF-only | Documented in this review; CI gofmt gate is LF-clean by construction |

## 12. Deferred Work

| Issue | Owner Agent |
| --- | --- |
| Dockerfile content review/validation of matrix images (build context, stages, distroless, EXPOSE) | Agent 06 (Docker) |
| Render Blueprint covering clients/transactions (currently deposits-only) | Agent 07 (Render) |
| Request logging/request-ID/metrics middleware wired through CI tests | Agent 09 (Observability) |
| Auth middleware / CORS policy (no policy defined yet) | Agent 10 (Security) |
| Performance/race-detection pipeline hardening if later required | Agent 11 (Performance) |
| OCI pipeline re-enablement when the OCI stack is evolved to new services | Agent 12 / deployment plan |

## 13. Local Reproduction Commands

| CI Stage | Local Command |
| --- | --- |
| go generate | `go generate ./...` |
| protoc generation | `cd protobuf && make generate-protos` |
| sqlc generation | `cd deposits/db && go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate` |
| Generated-code verification | `git diff --exit-code` over the generated directories listed in §6 |
| Formatting | `gofmt -l $(git ls-files '*.go' \| grep -v -E 'grpc/go/\|/db/sqlc/\|/mocks/\|\.pb\.go$\|\.pb\.gw\.go$')` |
| Static analysis | `go vet ./...` |
| Build | `go build ./...` |
| Tests | `go test ./...` |
| Docker build | `docker build -f <service>/Dockerfile -t rvpay-<service>:ci .` |

## 14. Changes Made

- `.github/workflows/render-deploy.yml` — extended generated-code drift scope
  to all services; added `validate` job (gofmt hand-written-only, `go vet
  ./...`, `go build ./...`); changed `docker-build` to a matrix over
  deposits/clients/transactions; added explicit deploy-skip `::notice::`;
  added `cache: true` to setup-go; split actionable generated-code failure
  messages.
- `shared/logger/logger.go` — added missing trailing newline (real gofmt
  issue found while validating the gate; content unchanged).

## 15. Documentation Check

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

## 16. Final Status

PASS WITH FOLLOW-UP