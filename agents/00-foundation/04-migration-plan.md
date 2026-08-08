# Agent 04 — Migration Strategy

## Objective

Produce the implementation roadmap required to evolve the current repository into the new RVPay architecture.

This is a migration planning task.

No code generation.

No database generation.

No protobuf generation.

---

## Review Scope

Read only:

- README.md
- docs/00-technical-design-document.md

Review only the service roots:

deposits/

integrations/

Do not inspect recursively unless required.

Ignore:

grpc/go/
third_party/
.github/
deploy/
nginx/
tests/

---

## Existing Architecture

Document:

current services

current responsibilities

existing reusable work

technical debt

areas suitable for reuse

---

## Migration Goals

Produce an ordered migration plan.

Prefer incremental evolution over rewrites.

Reuse existing code whenever practical.

Avoid duplicate implementations.

---

## Deliverable

Create:

docs/migration-plan.md

Include:

Phase 1

Repository preparation

Phase 2

Clients Service migration

Phase 3

Transactions Service migration

Phase 4

Shared infrastructure

Phase 5

Testing

Phase 6

Deployment

For every phase include:

objective

dependencies

expected outputs

risks

rollback considerations

success criteria

Do not modify any repository files.