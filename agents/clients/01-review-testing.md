# Agent 01 — Review Existing Integrations Service

## Objective

Review the existing Integrations service and produce a migration assessment for the new Clients Service.

This task is ANALYSIS ONLY.

Do not modify code.

Do not generate files outside the report.

---

## Required Reading

Read only:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md

Use these as the source of truth.

Do not inspect the repository recursively.

---

## Repository Review Scope

Inspect only:

integrations/

Inspect only enough files to understand:

- package layout
- configuration
- repositories
- services
- protobuf usage
- runtime
- OAuth implementation
- webhook implementation

Do not inspect:

tests/

grpc/go/

third_party/

.github/

deploy/

vendor/

node_modules/

---

## Analysis Tasks

Determine:

What functionality already exists.

What functionality should remain unchanged.

What functionality belongs in the future Clients Service.

Identify reusable code.

Identify duplicated logic.

Identify technical debt.

Identify package boundaries.

Identify dependency boundaries.

Determine where HighLevel-specific code exists.

Determine where provider-agnostic abstractions should exist.

---

## Migration Strategy

Produce a migration recommendation.

Categorize:

Can reuse unchanged

Needs refactor

Needs replacement

Needs deletion

---

## Deliverable

Create:

clients/docs/existing-service-review.md

Include:

Current architecture

Strengths

Weaknesses

Migration risks

Suggested migration order

Estimated implementation complexity

Do not modify existing source code.

---

## Completion Rules

(standard completion block)