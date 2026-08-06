# Render Infrastructure Generator

## Objective

Generate or improve deployment artifacts for Render.

Review existing artifacts first.

Do not overwrite working implementations.

---

## Generate

If missing:

- Render-compatible Dockerfile
- .dockerignore
- render.yaml (optional)
- GitHub Actions workflow
- deployment documentation

---

## Docker Rules

Detect Go version from go.mod.

Generate multi-stage Docker builds.

Build inside Docker.

Support:

linux/amd64

Render currently runs x86_64.

Do not force ARM.

---

## Protobuf

Inspect protobuf generation.

Determine whether protoc generation happens:

- before Docker build
- during Docker build

Recommend the most reproducible approach.

If generated code is absent during container builds,

move protobuf generation into Docker.

Explain why.

---

## GitHub Actions

Review workflow.

Ensure:

tests

↓

protobuf generation

↓

sqlc generation

↓

build

↓

Docker build

↓

deploy

---

## Environment

Support:

PORT

from Render.

Do not hardcode ports.

---

## Explain every generated artifact.