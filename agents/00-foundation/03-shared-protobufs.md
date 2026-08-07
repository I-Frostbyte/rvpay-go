# Agent 03 — Shared Protobuf Strategy

## Objective

Design the protobuf contract strategy for the new RVPay architecture.

Do not generate protobuf files.

Do not run protoc.

Do not modify generated code.

---

## Review Scope

Read only:

- README.md
- docs/00-technical-design-document.md

protobuf/

Do not inspect:

grpc/go/

Generated code is intentionally excluded.

---

## Existing Contracts

Identify:

current services

current messages

current RPCs

Determine which contracts remain valid.

Determine which contracts require replacement.

---

## Produce

Define protobuf ownership.

Determine:

service boundaries

shared messages

common enums

message reuse

package naming

versioning strategy

HTTP gateway strategy

---

## Naming Rules

Every service owns its own protobuf package.

Avoid circular imports.

Avoid duplicated messages.

Shared types must be isolated into common protobuf definitions.

---

## Deliverable

Create:

docs/protobuf-strategy.md

Include:

service packages

RPC ownership

shared enums

shared messages

generation workflow

versioning guidelines

Do not generate code.