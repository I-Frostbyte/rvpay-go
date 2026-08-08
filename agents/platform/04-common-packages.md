# Agent 04 — Common Packages

## Objective

Implement only the genuinely shared platform packages required by the new RVPay architecture.

The purpose of this agent is to establish small, stable, reusable packages that can be safely consumed by multiple services without creating a shared "god package" or moving service-specific business logic into common code.

The implementation must build upon the existing RVPay codebase and the foundation work already completed by previous agents.

The common packages must:

- follow the existing Go conventions
- follow the package structure defined by the documentation
- avoid duplicating service-specific logic
- avoid creating unnecessary abstractions
- avoid introducing premature frameworks
- remain independent of Clients and Transactions business logic
- be usable by future services
- preserve existing working code
- respect the existing service boundaries
- be documented sufficiently for later agents

This agent is responsible for common infrastructure/utilities only.

It is NOT responsible for:

- Clients business logic
- Transactions business logic
- database repositories
- SQL queries
- migrations
- protobuf definitions
- HTTP gateway implementation
- OAuth
- webhooks
- CI/CD
- Docker
- Render
- observability implementation
- security implementation
- performance optimization

---

# Required Reading

Read only:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md
- docs/platform-repository-audit.md
- docs/platform-protobuf-generation-review.md
- docs/platform-http-gateway-review.md

These documents are mandatory.

Do not begin implementation until all required documents have been read.

---

# Documentation Check

Before starting:

verify that every required document exists.

Required:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md
- docs/platform-repository-audit.md
- docs/platform-protobuf-generation-review.md
- docs/platform-http-gateway-review.md

If any required document is missing:

STOP.

Do not recreate missing work from previous agents.

At the end of the task:

perform the documentation check again.

Record the result in:

docs/platform-common-packages-review.md

---

# Repository Exploration Rules

## IMPORTANT

Do NOT perform an unrestricted recursive repository search.

Use:

README.md

as the repository map.

Use:

docs/repository-layout.md

for the target repository structure.

Use:

agents/project-context.md

for coding and package conventions.

Use:

docs/domain-model.md

for domain boundaries.

Use the previous platform review documents to understand completed work.

---

# Do NOT Explore Deep Folders

Do NOT recursively inspect:

- .git/
- vendor/
- node_modules/
- coverage/
- tmp/
- bin/
- third_party/
- third_party/googleapis/

Especially:

DO NOT inspect:

third_party/googleapis/

The existence of protobuf dependencies does not justify exploring the entire submodule.

Do not inspect generated dependency trees unless a concrete compilation issue requires it.

---

# Scope Rules

Only create or modify common packages that are justified by the documented architecture.

Do not create a package simply because the code "might be useful later."

Every new common package must have:

1. a clear consumer,
2. a clear responsibility,
3. a stable API,
4. no service-specific business logic.

If a proposed package has no current consumer:

DO NOT implement it.

Document it under Deferred Work instead.

---

# 1. Determine What "Common" Means

Before writing code, identify actual duplication or shared infrastructure.

Look for concepts such as:

- configuration helpers
- server lifecycle helpers
- error conventions
- context/request metadata
- pagination primitives
- shared validation helpers
- common identifiers
- common time handling
- common database utilities
- shared provider abstractions

However, only implement concepts explicitly justified by the existing codebase and documentation.

Do not assume all of these are required.

---

# 2. Prefer Existing Packages

Before creating anything:

search the locations identified by:

README.md

and:

docs/repository-layout.md

for existing implementations.

If a suitable package already exists:

reuse it.

Do NOT create a second implementation.

---

# 3. Do Not Copy Existing Service Code

If Clients or Transactions contains useful code:

do not blindly copy it into a common package.

First determine whether the logic is genuinely service-independent.

For example:

A provider-specific OAuth helper belongs to Clients.

A deposit-specific status mapper belongs to Transactions.

A generic context helper may belong in a shared package.

---

# 4. Domain Boundary Rule

Common packages must NOT know about service-specific domain concepts unless those concepts are explicitly shared by the domain model.

Do not place:

- Client domain services
- Merchant domain services
- Customer domain services
- Deposit business rules
- Payout business rules
- Provider-specific workflows

inside common packages.

---

# 5. Dependency Direction

Shared packages should sit below service packages in the dependency graph.

The desired direction is:

service
   ↓
common package

NOT:

common package
   ↓
service

A common package must never import:

- clients package
- transactions package
- provider implementations
- service handlers

unless explicitly documented.

---

# 6. Avoid Circular Dependencies

Before finalizing any package:

inspect its imports.

Ensure the dependency graph remains acyclic.

If introducing a package would require awkward dependency inversion solely to make it fit:

STOP.

Document the architectural concern instead of forcing it.

---

# 7. Package Naming

Follow the naming conventions established by:

agents/project-context.md

Use simple Go package names.

Avoid generic names such as:

- utils
- helpers
- misc
- commonstuff

unless that exact convention already exists in the project.

Prefer packages named after a real responsibility.

---

# 8. No God Package

Do NOT create:

common/

containing unrelated functions.

Do not create one enormous package containing:

- logging
- config
- database
- HTTP
- validation
- authentication
- IDs
- pagination
- errors

Keep responsibilities separated.

---

# 9. Configuration

Determine whether configuration helpers are genuinely shared.

The existing services may have configuration structs that are service-specific.

Do not merge all configuration into one global configuration package.

A common configuration package should only exist if it provides genuinely reusable infrastructure.

Examples of potentially shared concerns:

- environment parsing conventions
- common server configuration
- shared runtime defaults

Do not duplicate existing configuration loading mechanisms.

---

# 10. Configuration Ownership

Service-specific configuration remains owned by the service.

For example:

Clients should own configuration related to:

- HighLevel
- OAuth
- webhook settings

Transactions should own configuration related to:

- transaction providers
- transaction processing
- transaction database settings

Common packages must not absorb those settings.

---

# 11. Logging

Inspect the existing logging approach.

If the repository already has a structured logging convention:

do not introduce another logging framework.

Do not create a wrapper around the logger merely for stylistic reasons.

Only create a shared logging package if multiple services already require a common behavior that cannot be cleanly handled through the existing logger.

---

# 12. Error Handling

Inspect existing error conventions.

Determine whether the project needs a shared error package.

Only create one if there is an actual architectural requirement.

If created, it must define generic transport/domain-independent concepts.

It must NOT contain:

- HTTP-specific service logic
- Clients-specific errors
- Transactions-specific errors
- provider-specific error codes

---

# 13. Error Wrapping

Follow idiomatic Go error wrapping.

Preserve underlying errors.

Do not create string-only error handling.

Avoid patterns such as:

fmt.Errorf("something went wrong")

when useful underlying context can be preserved.

---

# 14. Sentinel Errors

Do not create large collections of sentinel errors preemptively.

Only create shared sentinel errors where multiple services genuinely require the same semantic error.

---

# 15. Validation

Determine whether validation is shared.

Do not create a generic validation framework unless the project already uses one or the architecture explicitly requires one.

Prefer small, deterministic validation helpers over large abstractions.

---

# 16. IDs

If the architecture requires a common identifier representation:

implement it only if multiple services actually consume it.

Do not create a custom ID type merely because one service needs an ID.

Follow the domain model.

---

# 17. Time

Do not introduce a custom time abstraction unless explicitly required.

Use:

time.Time

by default.

Avoid creating:

type Timestamp struct ...

without a documented reason.

---

# 18. Pagination

If pagination is explicitly shared across APIs:

implement the documented pagination primitives.

Otherwise:

do not create pagination utilities.

Pagination belongs to the API/domain design, not automatically to common infrastructure.

---

# 19. Context

Shared context helpers must be extremely limited.

Do not use context as a generic storage mechanism.

Do not place arbitrary application state into context.

Only implement context helpers required by the documented architecture.

---

# 20. Request Metadata

If request IDs, correlation IDs, or metadata propagation are already part of the documented platform architecture:

establish only the common primitives required for later observability work.

Do not implement full observability here.

Agent 09 owns observability.

---

# 21. Authentication Metadata

Do not implement authentication in this agent.

If a common package is needed to carry authentication metadata between transports:

implement only the minimal transport-neutral primitive.

Do not validate tokens.

Do not parse JWTs.

Do not perform authorization.

Agent 10 owns security.

---

# 22. Provider Abstractions

Do NOT create a global provider abstraction.

Provider interfaces belong to the service that owns the provider interaction.

For example:

HighLevel integration belongs to Clients.

Transaction providers belong to Transactions.

Do not create:

common.Provider

or equivalent merely to make the architecture appear generic.

---

# 23. Database Helpers

Do not create a generic database abstraction unless existing services demonstrably require one.

Do not hide pgx behind unnecessary interfaces.

Do not create a generic repository framework.

The repository layer remains service-owned.

---

# 24. SQL

Do not add SQL.

Do not add:

- queries
- migrations
- sqlc configuration
- generated sqlc code

This agent does not own database implementation.

---

# 25. Protobuf

Do not modify protobuf contracts.

Do not modify generated protobuf files.

Do not introduce shared protobuf messages unless explicitly required by the protobuf strategy.

Agent 02 owns protobuf generation.

---

# 26. Gateway

Do not modify the HTTP gateway.

Agent 03 owns gateway implementation.

If the common package needs to expose an API for gateway use:

ensure it remains transport-neutral.

---

# 27. Testing

Every common package created by this agent must have focused tests.

Tests must verify behavior rather than implementation details.

Do not create tests for trivial Go behavior.

---

# 28. Test Package Boundaries

Tests should remain close to the package being tested.

Follow the existing project convention.

Do not create a centralized:

common/tests/

directory merely to hold unrelated tests.

---

# 29. Avoid Over-Testing

Do not create massive test matrices for simple helpers.

Focus on:

- valid input
- invalid input
- important boundary conditions
- error behavior

where relevant.

---

# 30. Test Independence

Tests for common packages must not require:

- PostgreSQL
- Render
- HighLevel
- external APIs
- Docker
- production infrastructure

unless the package genuinely requires external infrastructure.

Prefer deterministic unit tests.

---

# 31. Public APIs

Keep common package APIs small.

Every exported:

- type
- function
- method
- constant

must have a clear purpose.

Avoid exporting internal implementation details.

---

# 32. Interfaces

Do not introduce interfaces unnecessarily.

In Go:

prefer concrete types unless multiple implementations or consumer-owned abstractions justify an interface.

Do not create:

type SomethingInterface interface {
    ...
}

simply because interfaces are considered "clean architecture."

---

# 33. Dependency Injection

Do not create a dependency injection framework.

Use normal Go constructors and explicit dependencies.

---

# 34. Generic Programming

Do not introduce generics into common packages unless they materially simplify a real shared problem.

Do not create generic abstractions for hypothetical future services.

---

# 35. Third-Party Dependencies

Before adding a dependency:

verify whether the standard library or an existing project dependency already solves the problem.

Do not add a package for trivial functionality.

---

# 36. Dependency Stability

Shared packages have a larger blast radius than service-local packages.

Therefore:

prefer dependencies that are:

- stable
- already used by the project
- well maintained
- necessary

Avoid experimental dependencies.

---

# 37. Go Version Compatibility

Use the Go version defined by the project.

Do not use language/library features unsupported by the project's configured Go version.

---

# 38. Documentation

Every genuinely reusable common package should have package documentation explaining:

- responsibility
- intended consumers
- non-responsibilities
- important constraints

Do not write large documentation blocks for trivial packages.

---

# 39. README

Only update README.md if the new shared package changes the documented repository architecture or developer workflow.

Do not rewrite unrelated sections.

---

# 40. Repository Layout

The final location of common packages must follow:

docs/repository-layout.md

Do not invent a new top-level directory without architectural justification.

If the documented layout does not provide a suitable location:

STOP and document the discrepancy.

Do not arbitrarily choose a location.

---

# 41. Existing Code Compatibility

Do not rewrite existing service code merely to demonstrate that the new package works.

Adopt shared packages incrementally.

If existing code already works and does not need the common package:

leave it alone.

---

# 42. Migration Strategy

If a common package replaces duplicated existing functionality:

migrate only the clearly identified consumers.

Do not rewrite unrelated code.

Document remaining duplication.

---

# 43. Backward Compatibility

Existing working services must continue to compile.

Do not break:

- deposits
- Clients
- Transactions
- protobuf generation
- gateway

to introduce common packages.

---

# 44. Existing Deposits Service

The legacy deposits service is an important compatibility reference.

Do not rewrite it during this agent.

Use it only to understand:

- existing conventions
- package patterns
- logging conventions
- configuration conventions
- server lifecycle conventions

The new architecture must not be forced into the old structure if the foundation documents define a better structure.

---

# 45. Previous Agent Work

Review:

docs/platform-http-gateway-review.md

Identify any common-package requirements discovered by Agent 03.

Resolve only the items that belong to this agent.

Do not modify gateway behavior unnecessarily.

---

# 46. Architecture Consistency

The common packages must support this architectural relationship:

                ┌───────────────┐
                │   Clients     │
                └───────┬───────┘
                        │
                        ▼
                ┌───────────────┐
                │    Common     │
                │ Infrastructure│
                └───────────────┘
                        ▲
                        │
                ┌───────┴───────┐
                │ Transactions  │
                └───────────────┘

Common packages must remain beneath service/application layers.

---

# 47. No Reverse Dependency

Do not allow:

common → clients

or:

common → transactions

or:

common → provider implementation

or:

common → database schema

unless explicitly justified by the documented architecture.

---

# 48. Package Dependency Audit

After implementation:

inspect imports for every new package.

Verify:

- no circular dependencies
- no service-to-common reverse dependency
- no accidental dependency on generated code
- no dependency on tests
- no dependency on third_party implementation details

---

# 49. Generated Code

Do not modify:

- `.pb.go`
- `_grpc.pb.go`
- grpc-gateway generated files
- sqlc generated files

If a common package requires generated code changes:

document the requirement for the appropriate previous/next agent.

---

# 50. Third-Party Directory

Do not modify:

third_party/

under any circumstances during this agent unless explicitly required by a documented build failure.

If a dependency issue appears:

document it.

Do not start investigating the entire third-party tree.

---

# 51. Security Boundary

Do not implement:

- authentication
- authorization
- JWT validation
- OAuth
- secrets management

Agent 10 owns security.

Only establish transport-neutral primitives if explicitly required.

---

# 52. Observability Boundary

Do not implement:

- metrics
- tracing
- dashboards
- alerting
- OpenTelemetry initialization

Agent 09 owns observability.

Only provide stable extension points if needed.

---

# 53. Performance Boundary

Do not perform broad performance optimization.

Agent 11 owns performance.

However:

avoid obviously inefficient designs in new common packages.

---

# 54. CI/CD Boundary

Do not modify:

.github/workflows/

Agent 05 owns CI/CD.

If CI requires a new common-package test command:

document it.

---

# 55. Docker Boundary

Do not modify Dockerfiles.

Agent 06 owns Docker.

---

# 56. Render Boundary

Do not modify Render configuration.

Agent 07 owns Render.

---

# 57. Final Package Review

For every common package created, answer:

1. Why does this package exist?
2. Who consumes it?
3. What responsibility does it own?
4. What responsibility does it explicitly NOT own?
5. Does it contain business logic?
6. Does it depend on a service package?
7. Does it introduce a new dependency?
8. Could the standard library solve the problem?
9. Could the functionality remain service-local?
10. Is the public API minimal?

If any package fails these questions:

reconsider the package before continuing.

---

# 58. Review Document

Create:

docs/platform-common-packages-review.md

Use exactly this structure:

# Platform Common Packages Review

## 1. Objective

Describe the purpose of the common packages.

## 2. Required Documentation

List every required document and confirm it was read.

## 3. Packages Created

| Package | Purpose | Consumers |
|---|---|---|

## 4. Packages Reused

Document existing shared packages that were reused instead of duplicated.

## 5. Dependency Direction

Describe the dependency relationship between:

- services
- common packages
- generated code
- external dependencies

## 6. Public APIs

Document the exported APIs introduced.

## 7. Testing

Document:

- tests added
- commands executed
- results

## 8. Findings

Use:

| ID | Severity | File/Area | Finding | Resolution |
|---|---|---|---|---|

## 9. Deferred Work

Document common functionality intentionally NOT implemented.

If it belongs to another agent, identify that agent.

## 10. Changes Made

List only files actually modified.

## 11. Documentation Check

Record the final documentation verification.

## 12. Final Status

Use exactly one:

PASS

PASS WITH FOLLOW-UP

BLOCKED

---

# 59. Final Verification

Run:

git status --short

Then:

git diff --stat

Inspect all changed files.

Ensure no unrelated files were modified.

---

# 60. Build Verification

Start with targeted packages.

Run tests for each newly created common package.

Then compile dependent packages.

If practical, run:

go test ./...

Do not spend excessive time investigating unrelated failures.

If an unrelated failure exists:

document it.

---

# 61. Dependency Verification

Run the appropriate Go dependency/build checks.

Do not modify go.mod merely to satisfy an unrelated package.

If a dependency is genuinely required:

add it using the project's normal Go tooling.

---

# 62. Documentation Verification

Before completion verify that:

- package documentation exists where required
- repository documentation is updated where required
- review documentation exists
- no duplicate architectural documentation was created unnecessarily

---

# 63. Final Scope Check

Run:

git status --short

Expected changes should be limited to:

- new common packages
- tests for those packages
- minimal documentation changes
- docs/platform-common-packages-review.md

Do not revert unrelated pre-existing changes automatically.

---

# 64. Final Completion Checklist

Before stopping:

- [ ] README.md was read.
- [ ] agents/project-context.md was read.
- [ ] docs/domain-model.md was read.
- [ ] docs/repository-layout.md was read.
- [ ] docs/protobuf-strategy.md was read.
- [ ] docs/migration-plan.md was read.
- [ ] docs/platform-repository-audit.md was read.
- [ ] docs/platform-protobuf-generation-review.md was read.
- [ ] docs/platform-http-gateway-review.md was read.
- [ ] Repository exploration was limited to relevant areas.
- [ ] Deep third-party folders were not explored.
- [ ] Existing shared packages were inspected before creating new ones.
- [ ] Every new package has a concrete consumer.
- [ ] Every new package has a single clear responsibility.
- [ ] No god package was created.
- [ ] No service-specific business logic was moved into common packages.
- [ ] No provider-specific logic was moved into common packages.
- [ ] No database implementation was added.
- [ ] No protobuf contracts were modified.
- [ ] No generated code was manually edited.
- [ ] No gateway implementation was unnecessarily modified.
- [ ] No CI/CD changes were made.
- [ ] No Docker changes were made.
- [ ] No Render changes were made.
- [ ] No security implementation was added.
- [ ] No observability implementation was added.
- [ ] No premature performance framework was introduced.
- [ ] Public APIs are minimal.
- [ ] Dependencies flow in the correct direction.
- [ ] No circular dependencies exist.
- [ ] Tests were added for new behavior.
- [ ] Relevant tests passed.
- [ ] docs/platform-common-packages-review.md was created.
- [ ] Documentation check was completed.
- [ ] Final git status was inspected.

---

# Final Stop Condition

STOP after:

1. reading all required documentation,
2. auditing existing shared functionality,
3. identifying only genuinely shared concerns,
4. implementing the minimum required common packages,
5. adding focused tests,
6. verifying dependency direction,
7. verifying existing services still compile,
8. creating docs/platform-common-packages-review.md,
9. completing the documentation check,
10. checking final git status.

Do NOT proceed to:

- CI/CD
- Docker
- Render
- observability
- security
- performance optimization
- Clients implementation
- Transactions implementation
- OAuth
- webhooks
- database migrations
- repositories

Those belong to other agents.

STOP.