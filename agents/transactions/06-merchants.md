# Agent 06 — Transactions Merchants Service

## Objective

Implement the Merchants portion of the Transactions Service.

This agent builds the merchant-facing application/service layer on top of the repository layer created by Agent 05.

The implementation must follow the existing RVPay Deposits service conventions wherever those conventions remain applicable.

The Merchants implementation is responsible for:

- merchant domain/service logic
- merchant service methods
- validation
- merchant lifecycle rules
- repository interaction
- mapping repository models to service/API models
- gRPC service implementation for merchant operations
- REST/gRPC-gateway behavior where already defined by Agent 04
- merchant-specific error handling
- merchant-specific logging where the existing service convention requires it

This agent must NOT implement:

- customers
- deposits
- payouts
- transaction lifecycle logic unrelated to merchants
- OAuth
- provider integrations
- webhook processing
- application startup
- Docker
- deployment
- database schema changes
- SQLC query generation
- repository redesign
- protobuf redesign

Those responsibilities belong to other agents.

---

# Required Reading

Read only:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md
- docs/transactions-existing-review.md
- docs/transactions-database-review.md
- docs/transactions-sqlc-review.md
- docs/transactions-protobuf-review.md
- docs/transactions-repository-review.md

These documents are mandatory.

Do not skip them.

---

# Documentation Check

Before modifying code, confirm that you have read:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md
- docs/transactions-existing-review.md
- docs/transactions-database-review.md
- docs/transactions-sqlc-review.md
- docs/transactions-protobuf-review.md
- docs/transactions-repository-review.md

At completion, perform the same documentation check again.

---

# Repository Exploration Rules

Use README.md as the repository map.

Perform focused exploration only.

Do NOT perform unrestricted repository-wide exploration.

Do NOT recursively inspect:

- third_party/
- third_party/googleapis/
- googleapis/
- vendor/
- node_modules/
- .git/
- coverage/
- tmp/
- bin/

Do not inspect unrelated services unless a specific implementation convention must be confirmed.

The existing Deposits service is the primary application/service implementation reference.

Inspect only the relevant Deposits files needed to understand:

- service structure
- handler structure
- dependency injection
- validation
- error handling
- logging
- gRPC implementation
- package naming
- constructor conventions

---

# 1. Inspect Existing Deposits Service

Use the existing Deposits implementation as the primary coding reference.

Inspect the relevant files under:

deposits/

Focus on:

- service implementation
- command/runtime wiring only where needed to understand dependency injection
- gRPC implementation
- configuration access
- repository usage
- error handling
- logging
- package structure

Do not copy Deposits business logic.

Copy the established coding patterns.

---

# 2. Inspect Transactions Protobuf Contract

Read:

docs/transactions-protobuf-review.md

Then inspect only the generated Transactions protobuf definitions required to implement merchant operations.

Determine:

- service interface
- merchant RPCs
- request types
- response types
- merchant resource types
- enums
- pagination types
- error expectations
- REST mappings

Do not modify protobuf files in this agent.

If the protobuf contract is insufficient:

STOP and document the problem.

Do not redesign Agent 04's protobuf implementation.

---

# 3. Inspect Transactions Repository

Read:

docs/transactions-repository-review.md

Then inspect the actual repository interface and implementation.

Determine:

- merchant repository methods
- expected parameters
- return types
- not-found behavior
- database error behavior
- context requirements

Do not modify repository code unless there is a trivial compile correction directly caused by this implementation.

If a repository capability is missing:

STOP and document it for Agent 05.

---

# 4. Determine Merchant Domain Boundary

Read:

docs/domain-model.md

Determine exactly what a Merchant represents in RVPay.

Identify:

- merchant identity
- merchant identifiers
- merchant status
- merchant metadata
- external identifiers
- timestamps
- relationships to customers
- relationships to transactions
- provider relationships where documented

Do not invent merchant fields.

Do not add fields simply because they might be useful later.

---

# 5. Merchant Service Package

Create the merchant service implementation in the location defined by:

docs/repository-layout.md

Follow the project's package naming conventions.

Do not create an alternative architecture.

The service must sit above the repository layer and below the transport layer.

---

# 6. Service Constructor

Implement the merchant service constructor using the project's established dependency-injection pattern.

The service should receive the dependencies it actually needs.

At minimum, this will likely include the merchant repository interface.

Do not:

- create database connections inside the service
- instantiate repositories inside service methods
- create global state
- instantiate gRPC clients unnecessarily

Dependencies must be injected.

---

# 7. Merchant Service Interface

If the project convention uses explicit service interfaces:

follow that convention.

Expose only the merchant operations required by the protobuf/domain contract.

Do not automatically expose every repository method.

The service API is an application boundary.

---

# 8. Create Merchant

Implement the merchant creation operation if defined by the protobuf contract.

The flow should be:

1. receive service request
2. validate request
3. normalize values where required
4. enforce documented business rules
5. call repository
6. map repository result
7. return service/API representation

Do not perform database work directly from the service.

---

# 9. Merchant Validation

Implement only validation supported by the domain documentation.

Potential validation areas include:

- required identifiers
- valid merchant references
- required contact information
- valid status
- required metadata

Do not create arbitrary validation rules.

Do not duplicate database constraints unnecessarily.

Application validation should provide useful client-facing errors before database execution.

---

# 10. Merchant Uniqueness

If the domain requires merchant identifiers to be unique:

the service may validate obvious request issues.

However, the database remains the authoritative enforcement mechanism for uniqueness.

Do not implement:

- application-wide mutexes
- in-memory uniqueness maps
- race-prone pre-check-only uniqueness enforcement

Correct behavior must remain safe under concurrent requests.

---

# 11. Get Merchant

Implement the merchant retrieval operation.

The service should:

1. validate the identifier
2. call the repository
3. handle not-found behavior
4. map the result
5. return the API representation

Do not expose SQLC models directly.

---

# 12. Merchant Not Found

Follow the project's established error convention.

Do not invent a new error framework.

A missing merchant should map consistently to the project's expected gRPC/API semantics.

If the repository returns a project-specific not-found error:

use it.

Do not expose raw PostgreSQL or SQLC errors to clients.

---

# 13. Update Merchant

Implement update behavior only if the protobuf contract defines it.

Determine whether the API uses:

- PATCH semantics
- PUT semantics
- explicit update commands

Follow:

docs/protobuf-strategy.md

Do not implement a generic update method merely because the database supports updates.

---

# 14. Partial Updates

If protobuf uses field masks:

follow the existing project convention.

Do not manually interpret arbitrary field names without validation.

Only fields explicitly supported by the domain should be updateable.

Do not allow updates to immutable fields such as:

- primary identifiers
- creation timestamps
- ownership identifiers

unless the domain explicitly permits them.

---

# 15. Merchant Status

If merchants have lifecycle/status states:

follow the domain model exactly.

Do not create new status values.

Do not allow arbitrary status changes if the domain defines a lifecycle.

If status transitions have rules:

implement those rules in the service layer.

The repository should only persist the resulting state.

---

# 16. Status Transitions

If the domain specifies allowed merchant status transitions:

validate:

current status → requested status

before updating.

Do not permit:

arbitrary state jumps

unless the architecture explicitly allows them.

Document the implemented transition rules.

---

# 17. List Merchants

If the protobuf contract exposes merchant listing:

implement it according to the documented pagination strategy.

Use the repository's pagination capabilities.

Do not:

- load all merchants into memory
- paginate in Go unnecessarily
- implement a second pagination model

---

# 18. Merchant Filtering

If merchant filtering is part of the protobuf contract:

support only documented filters.

Examples may include:

- status
- external ID
- creation date

Do not expose arbitrary database filters.

---

# 19. Merchant Ownership / Tenant Isolation

If the domain model defines merchant ownership boundaries:

enforce them.

A request for a merchant must not accidentally return another merchant's data.

Do not rely solely on the caller behaving correctly.

Use the documented identifier/ownership strategy.

---

# 20. Merchant Relationships

Do not eagerly load customers, transactions, deposits, or payouts unless the API contract explicitly requires it.

Avoid creating hidden N+1 queries.

If a merchant response contains only merchant information:

return only merchant information.

Relationships will be handled by the appropriate service operations.

---

# 21. Protobuf-to-Repository Mapping

Do not pass protobuf messages into the repository.

Convert:

protobuf request

→ service/domain representation

→ repository parameters

where required.

Follow the existing Deposits mapping conventions.

---

# 22. Repository-to-Protobuf Mapping

Do not return SQLC models directly from gRPC methods.

Map:

repository result

→ service/domain representation

→ protobuf response

Keep the transport layer independent from persistence implementation details.

---

# 23. Mapping Functions

If mappings are non-trivial:

create focused mapping functions.

Avoid embedding large conversion blocks inside every RPC method.

Follow the existing Deposits implementation style.

Do not create a massive generic mapper framework.

---

# 24. Error Mapping

Implement error mapping consistent with the existing RVPay service.

Handle at minimum where applicable:

- invalid request
- not found
- duplicate merchant
- database failure
- invalid state transition

Do not expose:

- SQL strings
- PostgreSQL internals
- stack traces
- internal file paths

to API clients.

---

# 25. gRPC Service Implementation

Implement the generated Transactions gRPC service interface.

The gRPC implementation should be thin.

It should primarily:

1. receive request
2. validate transport-level requirements
3. invoke merchant service logic
4. convert errors appropriately
5. return protobuf response

Do not put database calls directly inside gRPC handlers.

---

# 26. gRPC Handler Boundary

If the project architecture separates:

- service
- handler

follow that separation.

If the existing Deposits service combines them:

follow the established project convention.

Do not invent a third architectural style.

---

# 27. REST/gRPC-Gateway

If Agent 04 defined REST exposure through grpc-gateway:

do not implement a separate HTTP server.

The generated gateway should route to the Transactions gRPC service according to the protobuf contract.

Only add manual HTTP behavior if the architecture explicitly requires it.

---

# 28. HTTP Semantics

Do not duplicate REST route definitions in Go if they already exist in protobuf annotations.

The protobuf contract is the source of truth for generated REST routing.

---

# 29. Logging

Follow the logging conventions established by Deposits.

Do not introduce a new logging library.

Do not log:

- passwords
- API keys
- tokens
- secrets
- authorization headers
- sensitive customer information

Keep logging useful and appropriately scoped.

---

# 30. Context Propagation

Every service operation must accept and propagate context.

Do not create:

context.Background()

inside request-handling code.

Do not discard cancellation signals.

---

# 31. Timeouts

Do not create arbitrary internal timeouts unless the existing service architecture requires them.

Request-level timeouts should generally originate from the caller/runtime.

Do not hide timeout behavior inside merchant methods.

---

# 32. Concurrency

Do not introduce goroutines for ordinary merchant CRUD operations.

Service methods should execute synchronously unless the architecture explicitly requires asynchronous processing.

Do not create background workers in this agent.

---

# 33. Business Logic Boundary

Business rules belong in the merchant service.

Persistence mechanics belong in the repository.

Transport mechanics belong in the gRPC/REST layer.

Maintain those boundaries.

Do not:

- write SQL here
- call SQLC directly
- call PostgreSQL directly
- call external providers
- implement OAuth
- process webhooks

---

# 34. Provider Independence

Merchant service logic must remain provider-agnostic.

Do not add direct calls to:

- pawaPay
- HighLevel
- Stripe
- other providers

unless the domain documentation explicitly assigns such behavior to merchant creation.

Provider-specific behavior belongs to later integration/provider work.

---

# 35. Transactions

Do not implement transaction/deposit/payout business workflows here.

Merchant creation may create related records only if the documented architecture explicitly requires that atomic operation.

If such a requirement exists but the repository does not support it:

STOP and document the missing repository capability.

Do not bypass the repository.

---

# 36. Configuration

Do not create new configuration fields in this agent unless merchant service behavior explicitly requires them.

Configuration belongs to the runtime/scaffolding work.

Do not read environment variables directly from service methods.

---

# 37. Database Changes

Do not modify:

- migrations
- SQL query files
- SQLC generated code

unless a trivial compile correction is absolutely required.

If a schema/query issue is discovered:

document it and stop at that issue.

The database and SQLC agents own those layers.

---

# 38. Protobuf Changes

Do not modify:

- `.proto` contracts
- generated protobuf code
- generated gRPC code

If the merchant implementation cannot be completed because the protobuf contract is incorrect or incomplete:

document the problem for Agent 04.

Do not silently change the contract.

---

# 39. Repository Changes

Do not redesign the repository.

If a required merchant persistence method is missing:

document the missing method for Agent 05.

Do not move SQL logic into the service to compensate.

---

# 40. Testing Expectations

This agent is not the complete Transactions testing agent.

However, implementation must be testable.

If the existing project has service-level unit-test conventions:

follow them where necessary to validate the implementation.

Do not build the complete repository-wide test suite here.

Agent 12 will perform the comprehensive Transactions testing work.

---

# 41. Compile Validation

After implementation:

run focused compilation/tests for the affected Transactions packages.

At minimum verify:

- merchant service package compiles
- generated Transactions protobuf package compiles
- repository package compiles
- gRPC implementation compiles

---

# 42. Full Test Suite

If reasonably fast, run:

go test ./...

If unrelated packages fail:

do not modify unrelated code.

Record the failure.

---

# 43. Scope Review

Before completion run:

git status --short

Then:

git diff --stat

Then inspect the actual diff.

Expected changes should be limited to:

- Transactions merchant service files
- Transactions merchant handler/service implementation
- directly required tests
- docs/transactions-merchants-review.md

Do not modify unrelated services.

---

# 44. No Deep Repository Exploration

Before completion, confirm that no unnecessary changes were made to:

- third_party/
- googleapis/
- vendor/
- node_modules/
- generated dependency code
- unrelated services

If such changes appear:

investigate and revert unrelated changes.

Do not modify generated dependency/submodule contents.

---

# 45. Merchant Review Document

Create:

docs/transactions-merchants-review.md

This document is mandatory.

---

# 46. Required Documentation Structure

Use exactly:

# Transactions Merchants Implementation Review

## 1. Source Documents

List every required document read.

## 2. Existing Deposits Pattern

Document the Deposits service conventions used as the implementation reference.

## 3. Merchant Domain

Summarize the merchant domain implemented.

## 4. Service Structure

Document the merchant service package and responsibilities.

## 5. Repository Integration

Document the repository methods consumed.

## 6. Validation

Document request and domain validation.

## 7. Merchant Lifecycle

Document merchant status/lifecycle behavior if applicable.

## 8. RPC Implementation

List implemented merchant RPCs.

Use:

| RPC | Purpose | Repository Operation |
|---|---|---|

## 9. REST Exposure

If applicable, document the generated REST routes.

Use:

| Method | Route | RPC |
|---|---|---|

## 10. Error Handling

Document service/API error mappings.

## 11. Mapping

Document protobuf/repository/domain mappings.

## 12. Tests and Validation

Document tests and compilation performed.

## 13. Files Changed

List relevant files.

## 14. Risks

Document merchant-specific risks.

## 15. Unresolved Questions

Document anything requiring later work.

---

# 47. Documentation Check

Before finishing, verify that:

- README.md was read
- agents/project-context.md was read
- docs/domain-model.md was read
- docs/repository-layout.md was read
- docs/protobuf-strategy.md was read
- docs/migration-plan.md was read
- docs/transactions-existing-review.md was read
- docs/transactions-database-review.md was read
- docs/transactions-sqlc-review.md was read
- docs/transactions-protobuf-review.md was read
- docs/transactions-repository-review.md was read

Confirm that the implementation follows the documented architecture.

---

# Completion Checklist

Before stopping, verify:

- [ ] Existing Deposits service conventions were reviewed.
- [ ] Transactions domain documentation was reviewed.
- [ ] Transactions protobuf contract was reviewed.
- [ ] Transactions repository contract was reviewed.
- [ ] Merchant domain boundary was established.
- [ ] Merchant service package was created in the documented location.
- [ ] Merchant service dependencies are injected.
- [ ] Merchant creation was implemented where required.
- [ ] Merchant retrieval was implemented where required.
- [ ] Merchant update was implemented where required.
- [ ] Merchant listing was implemented where required.
- [ ] Merchant validation was implemented.
- [ ] Merchant lifecycle rules were implemented where documented.
- [ ] Merchant ownership/tenant isolation was respected.
- [ ] Repository methods are used rather than direct SQL.
- [ ] SQLC is not called directly from the service layer.
- [ ] Protobuf models are not passed into the repository.
- [ ] SQLC models are not returned directly through gRPC.
- [ ] Error mapping follows project conventions.
- [ ] Context is propagated correctly.
- [ ] No provider APIs were called.
- [ ] No OAuth logic was added.
- [ ] No webhook logic was added.
- [ ] No database schema was modified.
- [ ] No protobuf contract was modified.
- [ ] No repository redesign was performed.
- [ ] Relevant packages compile.
- [ ] Tests were run where practical.
- [ ] No unrelated files were modified.
- [ ] docs/transactions-merchants-review.md was created.
- [ ] Git diff was reviewed.

---

# Final Stop Condition

STOP after completing:

1. Merchant service implementation
2. Merchant gRPC implementation
3. Merchant validation
4. Repository integration
5. Focused validation
6. docs/transactions-merchants-review.md

Do NOT proceed to:

- customers
- deposits
- payouts
- runtime wiring
- application startup
- provider integrations
- OAuth
- webhooks
- Docker
- deployment

Those responsibilities belong to later Transactions agents.

If something required by the merchant implementation is missing:

- missing database capability → document for Agent 02/03
- missing SQLC method → document for Agent 03
- missing protobuf contract → document for Agent 04
- missing repository method → document for Agent 05

Do not silently modify previous agents' work to make the merchant implementation possible.

STOP.