# Tooling Versions

This document records the tool versions used for code generation and builds.
The Render CI workflow reads this file to pin the exact versions it installs.

## Locally installed versions

| Tool | Version | Notes |
| --- | --- | --- |
| protoc | 3.21.12 | `libprotoc 3.21.12` |
| protoc-gen-go | v1.36.10 | |
| protoc-gen-go-grpc | 1.5.1 | |
| sqlc | — | Not installed as a standalone binary; run via `go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0` |
| mockgen | v0.5.0 | Project pins v0.6.0 in `go.mod` / `deposits/db/doc.go` |
| goose | — | Not installed; referenced only as a commented-out migration tool in `agents/project-context.md` |

## Project-pinned versions (used by CI)

| Tool | Version | Source |
| --- | --- | --- |
| protoc | 3.21.12 | Local install |
| protoc-gen-go | v1.36.10 | Local install |
| protoc-gen-go-grpc | 1.5.1 | Local install |
| protoc-gen-grpc-gateway | v2.22.0 | `go.mod` (indirect) |
| sqlc | v1.29.0 | `deposits/db/doc.go` |
| mockgen | v0.6.0 | `go.mod` / `deposits/db/doc.go` |

## CI toolchain (`.github/workflows/render-deploy.yml`)

The `Install protoc toolchain` step installs only the tools required for
protobuf generation:

- `protoc` — `3.21.12`
- `protoc-gen-go` — `v1.36.10`
- `protoc-gen-go-grpc` — `1.5.1`
- `protoc-gen-grpc-gateway` — `v2.22.0`

sqlc and mockgen are not installed in the CI toolchain step; they are invoked
via `go run` with the versions pinned in `deposits/db/doc.go`.