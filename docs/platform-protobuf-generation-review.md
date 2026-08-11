# Platform Protobuf Generation Review

Document Version: 1.0
Status: Complete
System: RVPay
Review: Platform Agent 02 — Protobuf Generation

## 1. Objective

Establish a correct, reproducible, repository-consistent protobuf generation
system for RVPay. This agent verified and corrected the protobuf generation
TOOLCHAIN and WORKFLOW only:

- protobuf source files located correctly
- protoc invoked correctly
- protoc-gen-go invoked correctly
- protoc-gen-go-grpc invoked correctly
- protoc-gen-grpc-gateway invoked correctly (required — the documented
  strategy uses grpc-gateway annotations)
- googleapis dependencies resolved correctly
- generated Go code placed in the correct locations
- generation is reproducible
- local generation and CI generation use the same versions
- generated output remains synchronized with protobuf source
- Clients and Transactions protobuf contracts generate consistently

Per the agent's scope restrictions, this agent did NOT implement the HTTP
gateway (Agent 03), did NOT redesign protobuf contracts, and did NOT design
business-domain messages.

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

All required documents were present and read.

## 3. Existing Protobuf Structure

### Sources (`protobuf/`)

| File | Package | Go Package Option |
| --- | --- | --- |
| `deposits.proto` | `depositsgrpc` | `github.com/I-Frostbyte/rvpay-go/grpc/go/depositsgrpc` |
| `integrations.proto` | `integrationsgrpc` | `github.com/I-Frostbyte/rvpay-go/grpc/go/integrationsgrpc` |
| `clients.proto` | `clientsgrpc` | `github.com/I-Frostbyte/rvpay-go/grpc/go/clientsgrpc` |
| `transactions.proto` | `transactionsgrpc` | `github.com/I-Frostbyte/rvpay-go/grpc/go/transactionsgrpc` |
| `common.proto` | `commongrpc` | `github.com/I-Frostbyte/rvpay-go/grpc/go/commongrpc` |

All `go_package` options are aligned to the repository Go module
(`github.com/I-Frostbyte/rvpay-go`) as required by `docs/protobuf-strategy.md`.

### Generated output (`grpc/go/`)

Committed generated code per package:

| Package | Files |
| --- | --- |
| `depositsgrpc` | `deposits.pb.go`, `deposits_grpc.pb.go`, `deposits.pb.gw.go` |
| `integrationsgrpc` | `integrations.pb.go`, `integrations_grpc.pb.go`, `integrations.pb.gw.go` |
| `clientsgrpc` | `clients.pb.go`, `clients_grpc.pb.go`, `clients.pb.gw.go` |
| `transactionsgrpc` | `transactions.pb.go`, `transactions_grpc.pb.go`, `transactions.pb.gw.go` |
| `commongrpc` | `common.pb.go` |

`protobuf/Dockerfile` is intentionally empty (matches its documented state).

## 4. Toolchain

| Tool | Required Version | Actual Version | Status |
| --- | --- | --- | --- |
| protoc | v3.21.12 | v3.21.12 (libprotoc 3.21.12) | ✅ |
| protoc-gen-go | v1.36.10 | v1.36.10 | ✅ |
| protoc-gen-go-grpc | v1.5.1 | v1.5.1 | ✅ |
| protoc-gen-grpc-gateway | v2.22.0 | v2.22.0 | ✅ |

Required versions come unambiguously from `tools/versions.md` (project-pinned
versions used by CI). Actual versions were verified via `protoc --version`,
`protoc-gen-go --version`, `protoc-gen-go-grpc --version`, and
`protoc-gen-grpc-gateway --version`. Generated-code headers also confirm the
same versions. No toolchain upgrades were performed.

## 5. Generation Commands

Canonical command (from `protobuf/`):

```bash
make lint
make generate-protos
```

`protobuf/Makefile` `generate-protos` iterates every `*.proto` in `protobuf/`
and invokes protoc with:

- `--proto_path=./` (sources)
- `--proto_path=../third_party/googleapis` (googleapis dependency)
- `--go_out`, `--go-grpc_out`, `--grpc-gateway_out` into `../grpc/go/<pkg>grpc/`
- `paths=source_relative` for all three plugins

`clients/Makefile` provides `make generate-protos` which delegates to the
canonical `protobuf/Makefile` target (corrected this agent).

`go generate ./...` (per-service) is exclusively sqlc + mockgen; protobuf
generation is intentionally NOT wired into `go:generate` (verified — no
`go:generate` directive references protoc/protobuf).

## 6. Source Protobuf Verification

| Check | Result |
| --- | --- |
| Package names (`<service>grpc`, `commongrpc`) | ✅ consistent with strategy |
| `go_package` options | ✅ all match module `github.com/I-Frostbyte/rvpay-go` |
| Imports resolve locally | ✅ `common.proto` cross-package import works |
| Imports resolve from googleapis | ✅ `google/api/annotations.proto` present under `third_party/googleapis/` |
| Standard well-known imports | ✅ `google/protobuf/timestamp.proto` resolves via protoc |
| Service definitions | ✅ Clients (3 services), Transactions (4 services), Deposits (1), Integrations (1) |
| HTTP annotations | ✅ `google.api.http` on every public RPC; paths under `/v1/public/...` |
| Streaming | ❌ none declared (no streaming RPCs) — verified consistent |
| Shared messages | ✅ `commongrpc` types imported, not duplicated |

Active services verified: Clients and Transactions (required by the agent);
legacy Deposits and Integrations also verified and still generate.

## 7. Generated Output

- All output is committed under `grpc/go/<package>/`.
- `paths=source_relative` applied consistently for `--go_opt`,
  `--go-grpc_opt`, and `--grpc-gateway_opt`.
- No second generated-code tree exists.
- Generated packages compile under their intended Go packages
  (`depositsgrpc`, `integrationsgrpc`, `clientsgrpc`, `transactionsgrpc`,
  `commongrpc`).

## 8. Googleapis Dependency

- `third_party/googleapis` is a git submodule (commit
  `f3ff3a1dc91aa7719f98437416fd686fad0296cd`, `common-protos-1_3_1` lineage).
- Initialized at the working-tree commit; `git submodule status` shows it is
  checked out.
- Only the import root was verified (`third_party/googleapis/google/api/
  annotations.proto` exists). The submodule contents were NOT recursively
  inspected, per exploration rules.
- The submodule was not updated and its commit was not changed.

## 9. Reproducibility

- `make generate-protos` was run twice.
- After the first run, `git diff --exit-code -- grpc/go/` returned 0
  (no content differences — committed output is current).
- After the second run (determinism check), `git diff --exit-code -- grpc/go/`
  again returned 0. **Repeated generation is deterministic.**
- Note: because the local machine has `core.autocrlf=true`, regeneration
  rewrites files and Git reports them as stat-only modified (`.M`, porcelain v2)
  without content change. This is a local-node line-ending artifact; the blob
  hashes are identical. See Finding P-GEN-04.

## 10. Compilation Verification

| Command | Result |
| --- | --- |
| `go build ./grpc/go/...` | ✅ |
| `go vet ./grpc/go/...` | ✅ |
| `go build ./clients/...` | ✅ |
| `go build ./transactions/...` | ✅ |
| `go vet ./clients/... ./transactions/...` | ✅ |

## 11. CI Compatibility

- `tools/versions.md` is the single authoritative version source
  (protoc v3.21.12, protoc-gen-go v1.36.10, protoc-gen-go-grpc v1.5.1,
  protoc-gen-grpc-gateway v2.22.0) and matches `render-deploy.yml` installs and
  local tooling.
- The protobuf generation command is executable in a clean environment using
  the documented toolchain + `git submodule update --init --recursive`.
- No CI workflow changes were made. The following are carried forward for
  Agent 05 (CI/CD per P-02 from the Platform audit and Finding P-GEN-05 below):
  - The Render workflow's `git diff --exit-code` verification path
    (`deposits/db/sqlc`) does not explicitly cover `clients/db/sqlc` and
    `transactions/db/sqlc` (existing audit finding P-02).
  - CI does not explicitly verify `make lint` (protobuf clang-format) before
    generation. Now that `make lint` passes, Agent 05 may add it to the
    workflow.

## 12. Findings

| ID | Severity | File/Area | Finding | Resolution |
| --- | --- | --- | --- | --- |
| P-GEN-01 | MEDIUM | `protobuf/Makefile` | `lint` target referenced non-existent `types/*.proto`, so `make lint` always failed ("No such file or directory") and blocked the documented `lint → generate-protos` workflow | Fixed — lint target now lints `*.proto` only; `make lint` passes |
| P-GEN-02 | LOW | `protobuf/clients.proto` | `additional_bindings` block in `ListIntegrations` failed clang-format (`--Werror`) — spacing-only violation | Fixed — formatted per clang-format; zero contract/semantic change (regeneration byte-identical) |
| P-GEN-03 | MEDIUM | `clients/Makefile` | `generate-protos` target invoked `make generate` in `../protobuf` (non-existent target); it errored instead of generating protos | Fixed — now invokes the canonical `make generate-protos`; verified working |
| P-GEN-04 | INFO | Repository/git config | `core.autocrlf=true` (local) causes regenerated committed files to appear stat-modified (`.M`) without content change; CI may show the same on Windows runners | Documented. Optional hardening: add `.gitattributes` with `*.go text eol=lf` (deferred — not required for correctness; committed blobs already LF) |
| P-GEN-05 | INFO | `.github/workflows/render-deploy.yml` | CI runs `make generate-protos` but does not run `make lint`; sqlc diff scope covers `deposits/db/sqlc` only (pre-existing P-02) | Passed to Agent 05 (CI/CD). No CI changes made per scope |
| P-GEN-06 | INFO | `protobuf/deposits.proto` | protoc emits `warning: Import google/protobuf/timestamp.proto is unused` — the import is unused in the legacy contract | Legacy contract deprecated per `docs/migration-plan.md`; retired with the Deposits service. Not modified (contract freeze) |
| P-GEN-07 | INFO | `protobuf/README.md` | Stale: claims `go_package` is `github.com/rvpay/rvpay-go`, lists only `deposits.proto`/`depositsgrpc`, and states gateway stubs are not generated — all contradicted by the actual sources and committed output | Passed to Agent 08 (Documentation) to refresh; no README change made in this agent |

## 13. Changes Made

- `protobuf/Makefile` — removed non-existent `types/*.proto` glob from the
  `lint` target (now `clang-format --dry-run --Werror *.proto`).
- `protobuf/clients.proto` — clang-format formatting of the
  `additional_bindings` block in `ListIntegrations` (spacing only, no
  contract change).
- `clients/Makefile` — corrected `generate-protos` to invoke the canonical
  `../protobuf/Makefile` `generate-protos` target.

## 14. Deferred Work

| Issue | Owner Agent |
| --- | --- |
| HTTP gateway implementation (no changes made here) | Agent 03 |
| Shared/common packages (no `shared/` created here) | Agent 04 |
| CI protobuf lint step + sqlc diff scope coverage | Agent 05 |
| `protobuf/README.md` refresh | Agent 08 |
| Optional `.gitattributes` line-ending hardening (`*.go text eol=lf`) | Agent 05/08 |
| Retirement of legacy contracts (`deposits.proto`, `integrations.proto`) | Migration plan (Phase 6) |

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

## 16. Final Status

PASS WITH FOLLOW-UP