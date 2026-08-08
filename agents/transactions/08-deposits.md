# Agent 08 — Transactions Deposits Service

## Objective

Implement the Deposits portion of the Transactions Service.

This agent builds the Deposits capability on top of the Transactions foundation already implemented by Agents 02–07.

The implementation must integrate with:

- the Transactions database
- SQLC-generated persistence layer
- Transactions repositories
- Merchant domain
- Customer domain
- Transactions protobuf contracts
- existing Deposits implementation conventions
- the new RVPay domain model
- the documented Transactions architecture

The Deposits implementation is responsible for:

- deposit creation
- deposit retrieval
- deposit listing where defined
- deposit lifecycle/status handling
- merchant ownership
- customer association
- deposit validation
- repository interaction
- gRPC handlers
- REST/gRPC-gateway exposure where already defined
- domain/API mapping
- deposit-specific error handling

This agent must NOT implement:

- payouts
- provider integrations
- webhook processing
- OAuth
- Clients service functionality
- runtime/application startup
- Docker
- deployment
- database migrations
- SQLC generation
- protobuf contract redesign
- repository redesign

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
- docs/transactions-merchants-review.md
- docs/transactions-customers-review.md

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
- docs/transactions-merchants-review.md
- docs/transactions-customers-review.md

At completion, perform the same documentation check again.

---

# Repository Exploration Rules

Use README.md as the repository map.

Perform focused exploration only.

Do NOT perform unrestricted recursive searches.

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

Do not inspect unrelated services.

The existing Deposits service is the primary implementation reference.

The Transactions Merchant and Customer implementations are the primary references for the new Transactions architecture.

Inspect only the relevant files required to understand:

- existing deposit behavior
- service structure
- repository usage
- protobuf mapping
- validation
- error handling
- dependency injection
- lifecycle/state handling
- logging
- context propagation

---

# 1. Review the Existing Deposits Service

The repository already contains a Deposits service.

Use README.md to locate it.

Inspect only the relevant files needed to understand:

- existing deposit domain behavior
- existing deposit request/response structures
- existing deposit lifecycle
- existing repository behavior
- existing SQL queries
- existing protobuf contracts
- existing service implementation
- existing validation
- existing tests

Do not copy the old architecture blindly.

The new Transactions architecture is authoritative.

The old Deposits service is a behavioral and coding-style reference.

---

# 2. Compare Old Deposits With New Domain

Read:

docs/domain-model.md

Compare the existing Deposits implementation against the new domain model.

Identify:

- fields retained
- fields removed
- fields renamed
- fields whose meaning changed
- lifecycle changes
- ownership changes
- customer relationship changes
- merchant relationship changes
- provider-related changes

Do not silently preserve old behavior if the new domain model contradicts it.

---

# 3. Transactions Deposits Boundary

The Deposits capability belongs inside the Transactions service.

Do not create:

deposits/

as a new top-level microservice.

The expected architectural relationship is:

Transactions Service
|
+-- Merchant
+-- Customer
+-- Deposit
+-- Payout
+-- shared transaction concepts

Follow:

docs/repository-layout.md

for the exact package locations.

---

# 4. Inspect Transactions Database

Read:

docs/transactions-database-review.md

Determine the actual persistence model for deposits.

Identify:

- deposit table
- primary key
- merchant relationship
- customer relationship
- amount
- currency
- status
- reference
- timestamps
- external/provider references
- constraints
- indexes

Do not modify migrations.

If the database is missing something required by the documented domain:

STOP and document it for Agent 02.

---

# 5. Inspect SQLC

Read:

docs/transactions-sqlc-review.md

Determine the generated SQLC operations available for deposits.

Do not modify generated SQLC files.

Do not create raw SQL in the service layer.

If a required query does not exist:

document it for Agent 03.

Do not bypass the repository.

---

# 6. Inspect Transactions Repository

Read:

docs/transactions-repository-review.md

Inspect the deposit repository interface and implementation.

Determine:

- CreateDeposit
- GetDeposit
- ListDeposits
- UpdateDeposit
- status operations
- merchant-scoped operations
- customer-scoped operations
- not-found behavior
- duplicate behavior
- pagination
- repository error conventions

Use the repository abstraction.

Do not call SQLC directly.

---

# 7. Inspect Merchant Implementation

Read:

docs/transactions-merchants-review.md

Inspect the actual merchant service.

The Deposit service must follow the same:

- constructor style
- dependency injection
- package organization
- service method conventions
- validation style
- error handling
- logging
- mapping conventions

Do not create a separate architecture for Deposits.

---

# 8. Inspect Customer Implementation

Read:

docs/transactions-customers-review.md

Inspect the actual Customer implementation.

Deposits must reference customers using the existing customer architecture.

Do not duplicate:

- customer validation
- customer lookup logic
- customer repository logic
- customer domain models

---

# 9. Inspect Transactions Protobuf

Read:

docs/transactions-protobuf-review.md

Inspect the deposit RPC definitions.

Determine:

- create deposit RPC
- get deposit RPC
- list deposit RPC
- update/status RPCs
- request messages
- response messages
- resource representation
- pagination
- field masks
- HTTP annotations
- status representations

Do not modify protobuf contracts.

If a required RPC or field is missing:

document it for Agent 04.

---

# 10. Deposit Domain

Implement only fields and semantics supported by:

docs/domain-model.md

A deposit generally represents an inbound movement of funds.

However, do not invent business semantics.

Use the actual documented domain.

---

# 11. Deposit Identity

Determine the authoritative deposit identifiers.

Possible identifiers may include:

- internal deposit ID
- merchant reference
- external/provider reference

Use the documented model.

Do not introduce new identifiers.

Do not treat provider transaction IDs as internal primary keys unless explicitly documented.

---

# 12. Merchant Ownership

Every deposit must belong to the correct merchant according to the domain model.

Tenant isolation is mandatory.

A deposit belonging to:

Merchant A

must never be returned through a request scoped to:

Merchant B.

Use repository-level ownership filters where available.

Do not rely exclusively on application-level filtering after retrieving data.

---

# 13. Customer Association

Where the domain requires a customer:

validate the customer relationship.

Do not assume that a customer ID alone proves ownership.

Ensure:

Customer X belongs to Merchant A

before associating Customer X with a deposit belonging to Merchant A.

Follow the Customer implementation and repository boundaries.

---

# 14. Deposit Creation

Implement deposit creation according to the protobuf/API contract.

Expected flow:

1. receive request
2. validate request
3. validate merchant context
4. validate customer context if applicable
5. validate amount
6. validate currency
7. validate reference/idempotency information
8. construct service/domain input
9. call repository
10. map result
11. return response

Do not perform provider calls.

---

# 15. Amount Validation

Follow the domain model and existing conventions.

Validate:

- amount is present where required
- amount is positive where required
- amount precision is valid
- currency requirements are respected

Do not use floating-point arithmetic for monetary values if the project uses integer/decimal representations.

Follow the existing model.

---

# 16. Currency Validation

Use the currency model defined by the project.

Do not create a second currency representation.

If the protobuf or domain defines supported currency values:

validate against those values.

Do not hard-code unsupported currencies.

---

# 17. Reference Validation

If deposits use merchant references:

validate them according to the documented contract.

Respect uniqueness guarantees.

Do not implement unsafe check-then-insert uniqueness logic.

Database constraints remain authoritative.

---

# 18. Idempotency

If the domain/API defines idempotency:

preserve it.

A repeated request with the same idempotency/reference key must follow the documented behavior.

Do not create a second deposit merely because the same request is retried.

Do not invent an idempotency system if the architecture does not define one.

---

# 19. Deposit Status

Use only statuses defined by the domain/protobuf.

Do not create new statuses.

Do not use arbitrary strings where the project uses enums.

Document the lifecycle implemented.

---

# 20. Deposit State Machine

If docs/domain-model.md defines deposit transitions:

enforce them.

For example:

initial → pending → successful

or whatever transitions the documented model specifies.

Do not assume the example above is the actual project lifecycle.

The documentation is authoritative.

---

# 21. Invalid State Transitions

Reject invalid transitions.

For example, if a terminal state cannot transition elsewhere:

prevent the update.

Do not allow:

successful → pending

unless explicitly documented.

Do not implement a generic state machine framework.

Use straightforward validation consistent with the project.

---

# 22. Provider Independence

This agent must NOT call payment providers.

Do not call:

- pawaPay
- GoHighLevel
- external payment gateways
- external APIs

The Deposit service should operate on the internal transaction model.

Provider orchestration belongs elsewhere.

---

# 23. Deposit Retrieval

Implement GetDeposit if defined.

Expected flow:

1. validate deposit ID
2. validate merchant context
3. call repository
4. handle not-found
5. map repository result
6. return API response

Do not expose SQLC structs.

---

# 24. Merchant-Scoped Retrieval

If the API supports merchant-scoped retrieval:

ensure the repository query includes merchant ownership.

Do not:

1. fetch deposit globally
2. check merchant in application code

if the repository already supports scoped retrieval.

Prefer database-enforced ownership filtering.

---

# 25. Customer-Scoped Retrieval

If listing or retrieving deposits by customer is supported:

ensure the customer belongs to the merchant context.

Do not allow cross-merchant customer access.

---

# 26. List Deposits

If defined by the API:

implement listing using the repository's pagination model.

Do not:

- load all deposits
- paginate in memory
- create a second pagination implementation

---

# 27. Deposit Filters

Only expose documented filters.

Possible filters include:

- merchant
- customer
- status
- currency
- reference
- creation date range

Do not invent arbitrary query parameters.

---

# 28. Pagination

Follow:

docs/protobuf-strategy.md

and the repository's established pagination implementation.

Preserve:

- page size
- page token
- next page token

semantics.

Do not invent a new pagination token format.

---

# 29. Deposit Updates

Implement update operations only if defined by the domain/API.

Do not expose unrestricted deposit updates.

Deposits are financial records.

Mutable fields must be explicitly documented.

---

# 30. Financial Record Integrity

Do not allow arbitrary modification of:

- amount
- merchant ownership
- customer ownership
- creation timestamp
- immutable identifiers

unless the domain explicitly allows it.

Do not expose a generic:

UpdateDeposit(fields map[string]interface{})

style API.

---

# 31. Status Updates

If the protobuf exposes status transitions:

implement them through the documented repository/service boundary.

Validate:

- current state
- requested state
- transition validity
- ownership

Do not directly modify SQLC records from the handler.

---

# 32. Transactional Consistency

Where the repository exposes transactional operations:

use them.

Do not manually compose multiple database writes if the repository already provides an atomic operation.

Do not introduce application-level distributed transaction logic.

---

# 33. Mapping

Do not expose SQLC-generated types through the service API.

Use the established mapping direction:

protobuf
→ service/domain
→ repository

and:

repository
→ service/domain
→ protobuf

Follow the mapping style used by Merchants and Customers.

---

# 34. Deposit Response

The API response should contain only fields defined by the protobuf contract.

Do not expose:

- internal database fields
- raw SQL metadata
- provider credentials
- secrets
- internal connection details

---

# 35. gRPC Handler

Implement the deposit RPC methods in the Transactions gRPC service.

Handlers must remain thin.

Expected flow:

gRPC request

→ validation/service invocation

→ service logic

→ mapping

→ gRPC response

Do not put SQLC calls in the handler.

---

# 36. REST/gRPC-Gateway

If deposit RPCs contain HTTP annotations:

use the generated grpc-gateway exposure.

Do not create a duplicate REST implementation.

Do not manually maintain routes that are already generated from protobuf.

---

# 37. REST Endpoint Verification

Verify the generated REST exposure matches the documented API.

Check:

- HTTP method
- route
- request body
- response
- error behavior

Do not modify generated gateway files manually.

---

# 38. Error Handling

Handle at least the applicable categories:

- invalid request
- invalid amount
- invalid currency
- merchant not found
- customer not found
- deposit not found
- duplicate reference/idempotency key
- invalid state transition
- database failure

Use the project's existing error conventions.

Do not expose:

- PostgreSQL errors
- SQLC errors
- pgx errors
- stack traces

to API clients.

---

# 39. Not Found

Map repository not-found behavior to the project's established application/API error.

Do not return an HTTP 500 for a normal missing deposit.

---

# 40. Duplicate Deposit

If the repository reports a uniqueness violation:

map it using the project's established conflict/error convention.

Do not perform unsafe application-only duplicate detection.

---

# 41. Context

Every request path must propagate context.

Do not use:

context.Background()

inside request processing.

Respect cancellation and deadlines.

---

# 42. Logging

Use the project's established logger.

Log meaningful operational information without leaking sensitive data.

Never log:

- API keys
- access tokens
- authorization headers
- passwords
- secrets
- full payment credentials

---

# 43. No Provider Logic

Do not add provider interfaces to the Deposit service.

Do not add:

ProviderClient

PaymentProvider

PawaPayClient

or similar abstractions here unless they already exist as an established domain boundary explicitly required by the documentation.

Provider integration is outside this agent's scope.

---

# 44. No Webhooks

Do not implement:

- provider webhooks
- callback handlers
- webhook signature validation
- asynchronous provider event processing

Those belong to later/other architecture.

---

# 45. No OAuth

Do not implement:

- OAuth authorization
- OAuth callbacks
- token exchange
- SSO
- client credentials

Those belong to the Clients service.

---

# 46. No Database Changes

Do not modify:

- migrations
- database schema
- SQL query files
- SQLC configuration
- generated SQLC code

If persistence is incomplete:

document the required change for Agent 02 or Agent 03.

---

# 47. No Protobuf Changes

Do not modify:

- .proto files
- generated protobuf files
- generated gRPC files
- generated gateway files

If the contract is insufficient:

document the required change for Agent 04.

---

# 48. No Repository Redesign

Do not redesign repository interfaces.

If a required deposit repository method is missing:

document it for Agent 05.

Do not bypass the repository by calling SQLC directly.

---

# 49. Existing Deposits Migration Strategy

The old Deposits service is being absorbed into Transactions.

Do not leave two competing Deposit implementations active unintentionally.

Determine from the migration documentation:

- which old files are being superseded
- which existing code remains authoritative
- whether old functionality is expected to remain temporarily

Do not delete the existing Deposits service unless explicitly instructed by the migration plan.

Do not perform migration cleanup in this agent.

---

# 50. Testing Scope

Agent 12 will handle comprehensive Transactions testing.

This agent must nevertheless make the Deposit implementation testable.

Add focused tests where necessary for:

- validation
- state transitions
- merchant ownership
- customer association
- mapping
- error handling
- service behavior

Follow existing project test conventions.

Do not introduce a new testing framework.

---

# 51. Compile Validation

Run focused validation for affected packages.

At minimum verify:

- Transactions service compiles
- Deposit package compiles
- gRPC implementation compiles
- repository dependency compiles
- protobuf types compile

---

# 52. Full Tests

If reasonably fast:

go test ./...

If unrelated tests fail:

do not modify unrelated code.

Record the failure in the review document.

---

# 53. Git Review

Before completing the agent:

run:

git status --short

Then:

git diff --stat

Then inspect the relevant diff.

Confirm changes are limited to Deposit implementation and directly required documentation/tests.

---

# 54. Scope Enforcement

Expected changes should generally be limited to:

- Transactions Deposit service files
- Transactions gRPC implementation
- directly required Deposit tests
- docs/transactions-deposits-review.md

Do not modify:

- Clients service
- Payout implementation
- OAuth
- webhooks
- provider integrations
- deployment
- Docker
- workflows
- third_party/

---

# 55. Generated Files

Never manually modify generated files.

If generated protobuf/SQLC/gateway output is incorrect:

document the problem for the responsible agent.

Do not patch generated output manually.

---

# 56. Deep Folder Protection

Do not modify or regenerate files under:

- third_party/
- third_party/googleapis/
- vendor/
- node_modules/

Do not inspect these directories recursively.

If a tool enters one because of a dependency:

return immediately to the project files.

---

# 57. Deposit Review Document

Create:

docs/transactions-deposits-review.md

This document is mandatory.

---

# 58. Required Review Document Structure

Use exactly:

# Transactions Deposits Implementation Review

## 1. Source Documents

List every document read.

## 2. Existing Deposits Implementation

Document:

- relevant old Deposits behavior
- files inspected
- behavior retained
- behavior changed

## 3. Domain Model

Document the implemented Deposit domain.

## 4. Merchant Relationship

Document merchant ownership and tenant isolation.

## 5. Customer Relationship

Document how deposits reference customers.

## 6. Deposit Lifecycle

Use:

| Current State | Allowed State | Rule |
|---|---|---|

Document only transitions actually supported.

## 7. Service Structure

Document the Deposit service package and responsibilities.

## 8. Repository Integration

Use:

| Service Operation | Repository Operation |
|---|---|

## 9. RPC Implementation

Use:

| RPC | Purpose | Service Operation |
|---|---|---|

## 10. REST Exposure

If applicable:

| Method | Route | RPC |
|---|---|---|

## 11. Validation

Document:

- amount
- currency
- references
- merchant
- customer
- status

## 12. Error Handling

Document:

| Condition | Application/API Error |
|---|---|

## 13. Mapping

Document:

- protobuf → service
- service → repository
- repository → service
- service → protobuf

## 14. Idempotency

Document how idempotency/reference uniqueness is handled if applicable.

## 15. Provider Boundary

Document explicitly that provider integration is not performed here.

## 16. Tests and Validation

Document:

- focused tests
- compilation
- full test results

## 17. Files Changed

List relevant files.

## 18. Risks

Document Deposit-specific risks.

## 19. Unresolved Questions

Document anything requiring later agents or architectural decisions.

---

# 59. Documentation Check

Before finishing, verify again that:

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
- docs/transactions-merchants-review.md was read
- docs/transactions-customers-review.md was read

The review document must reflect the actual implementation.

---

# Completion Checklist

Before stopping, verify:

- [ ] Required documents were read.
- [ ] Existing Deposits implementation was reviewed.
- [ ] New Deposit domain was compared against the old implementation.
- [ ] Transactions database model was reviewed.
- [ ] SQLC deposit operations were reviewed.
- [ ] Transactions repository was reviewed.
- [ ] Merchant implementation was reviewed.
- [ ] Customer implementation was reviewed.
- [ ] Deposit protobuf contract was reviewed.
- [ ] Deposit service was implemented.
- [ ] Deposit creation was implemented where required.
- [ ] Deposit retrieval was implemented where required.
- [ ] Deposit listing was implemented where required.
- [ ] Deposit updates were implemented only where documented.
- [ ] Deposit lifecycle rules were implemented.
- [ ] Invalid state transitions are rejected.
- [ ] Merchant ownership is enforced.
- [ ] Customer association is validated.
- [ ] Amount validation follows the domain.
- [ ] Currency validation follows the domain.
- [ ] Reference/idempotency rules are respected.
- [ ] Repository interfaces are used.
- [ ] SQLC types are not exposed through the API.
- [ ] Protobuf messages are not passed into repositories.
- [ ] gRPC handlers remain thin.
- [ ] REST exposure uses generated gateway behavior where applicable.
- [ ] Provider APIs were not called.
- [ ] OAuth was not added.
- [ ] Webhooks were not added.
- [ ] Database migrations were not modified.
- [ ] Protobuf contracts were not modified.
- [ ] Repository architecture was not redesigned.
- [ ] Generated files were not manually modified.
- [ ] No third_party/googleapis files were modified.
- [ ] Focused validation was performed.
- [ ] Full tests were run if practical.
- [ ] Git diff was reviewed.
- [ ] docs/transactions-deposits-review.md was created.
- [ ] Documentation check was completed.

---

# Final Stop Condition

STOP after completing:

1. Deposit domain/service implementation
2. Deposit repository integration
3. Deposit validation
4. Deposit lifecycle handling
5. Merchant/customer ownership enforcement
6. Deposit gRPC implementation
7. REST/gRPC-gateway integration where already defined
8. Focused validation
9. docs/transactions-deposits-review.md

Do NOT proceed to:

- payouts
- runtime wiring
- application startup
- provider integrations
- OAuth
- webhooks
- Docker
- deployment

Those responsibilities belong to later agents.

If something required by the Deposit implementation is missing:

- missing database capability → document for Agent 02
- missing SQLC operation → document for Agent 03
- missing protobuf contract → document for Agent 04
- missing repository operation → document for Agent 05
- missing merchant capability → document for Agent 06
- missing customer capability → document for Agent 07

Do not silently modify previous agents' work.

STOP.