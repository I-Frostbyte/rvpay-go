# Agent 01 — Domain Model Foundation

## Objective

Design the new RVPay domain model that will be used by every service in the repository.

This is a planning and modelling task.

Do not begin implementing services.

Do not generate SQL.

Do not generate protobufs.

Do not generate Go code.

The output of this task is a documented domain model that future agents will implement.

---

## Review Scope

Read only:

- README.md
- docs/00-technical-design-document.md

Do not inspect:

.github/
deploy/
nginx/
third_party/
grpc/go/
protobuf/
tests/
vendor/
node_modules/

Use README.md as the authoritative map of the repository.

---

## Existing System

Identify the current bounded contexts.

Identify which concepts already exist.

Identify which concepts must evolve.

Do not redesign existing functionality that already aligns with the new architecture.

---

## Produce

Define the complete business domain including:

• Clients

• Platforms

• Integrations

• Merchants

• Customers

• Deposits

• Payouts

Define:

• responsibilities

• ownership

• relationships

• lifecycle

• service ownership

---

## Service Ownership

Determine which service owns each entity.

No entity may be owned by multiple services.

Cross-service communication must occur through gRPC contracts.

Never share database tables across services.

---

## Deliverable

Create:

docs/domain-model.md

Include:

- entity descriptions
- ownership
- relationships
- future extensibility
- assumptions
- unresolved questions

Do not modify source code.