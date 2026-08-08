# Agent 07 — Transactions Customers Service

## Objective

Implement the Customers portion of the Transactions Service.

This agent builds the customer-facing application/service layer on top of the Transactions repository layer and the architectural foundations established by Agents 02–06.

The implementation must follow:

- the new RVPay domain model
- the documented Transactions architecture
- the existing Deposits coding conventions
- the Transactions repository implementation
- the Transactions protobuf contract
- the merchant implementation established by Agent 06

The Customers implementation is responsible for:

- customer domain/service logic
- customer validation
- customer lifecycle behavior where documented
- customer repository interaction
- merchant/customer ownership boundaries
- customer service methods
- customer gRPC handlers
- REST/gRPC-gateway behavior where already defined
- protobuf/domain/repository mapping
- customer-specific error handling
- focused customer validation/testing

This agent must NOT implement:

- deposits
- payouts
- transaction processing
- merchant redesign
- OAuth
- webhooks
- external provider integrations
- application startup
- runtime wiring
- Docker
- deployment
- database schema redesign
- SQLC generation
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

The existing Deposits service is the primary coding reference.

The newly implemented Transactions merchant service is the secondary local reference.

Inspect only the relevant files required to understand:

- service structure
- handler structure
- validation
- repository usage
- dependency injection
- error handling
- logging
- protobuf mapping
- package conventions

---

# 1. Inspect Existing Deposits Customer-Related Patterns

Search only the existing Deposits implementation for any customer-related concepts.

Determine whether Deposits already contains:

- customer models
- customer references
- customer validation
- customer identifiers
- payer information
- customer-facing request structures

Use these only as implementation references.

Do NOT copy Deposits domain assumptions if they conflict with:

docs/domain-model.md

The new domain model is authoritative.

---

# 2. Inspect Transactions Merchant Implementation

Review:

docs/transactions-merchants-review.md

Then inspect the actual merchant service implementation.

Use it to maintain consistency for:

- constructor patterns
- package organization
- dependency injection
- validation
- error handling
- mapping
- gRPC implementation
- logging
- context propagation

Customers should feel like part of the same Transactions service.

Do not create a different architectural pattern.

---

# 3. Inspect Customer Domain

Read:

docs/domain-model.md

Determine exactly what a Customer represents.

Identify:

- customer identity
- customer identifiers
- merchant ownership
- customer contact information
- external identifiers
- status
- timestamps
- relationships to transactions
- relationships to merchants
- any other explicitly documented customer attributes

Do not invent additional domain fields.

Do not infer undocumented business rules.

---

# 4. Merchant/Customer Relationship

Determine how Customers belong to Merchants.

The implementation must preserve the documented relationship.

A customer belonging to Merchant A must not accidentally be retrievable through Merchant B.

The merchant/customer relationship is a core tenant boundary.

Follow the actual database and repository design.

Do not create a parallel ownership model.

---

# 5. Inspect Transactions Repository

Read:

docs/transactions-repository-review.md

Inspect the actual customer repository interface.

Determine:

- available customer methods
- parameters
- return types
- ownership filters
- not-found behavior
- duplicate behavior
- pagination behavior
- error conventions

Do not redesign the repository.

If a required repository method is missing:

STOP and document it for Agent 05.

Do not bypass the repository by calling SQLC directly.

---

# 6. Inspect Transactions Protobuf

Read:

docs/transactions-protobuf-review.md

Inspect only the generated protobuf definitions required for customer operations.

Determine:

- customer RPCs
- request messages
- response messages
- customer resource
- identifiers
- pagination
- field masks
- HTTP annotations
- status enums
- error expectations

Do not modify protobuf contracts.

If the contract is insufficient:

document the problem for Agent 04 and STOP at that dependency.

---

# 7. Customer Service Package

Create or complete the Customer service implementation in the location defined by:

docs/repository-layout.md

Follow the exact package naming conventions already established by the Transactions merchant implementation.

Do not create a second service architecture.

---

# 8. Service Constructor

Implement the customer service constructor using the same dependency injection pattern as the merchant service.

The service should receive repository dependencies through interfaces.

Do not:

- create database connections
- instantiate repositories inside methods
- read environment variables
- create global state
- initialize external providers

inside the customer service.

---

# 9. Customer Service Interface

If the project uses service interfaces:

follow the existing convention.

Expose only operations required by the domain and protobuf contract.

Do not expose the entire repository interface.

---

# 10. Create Customer

Implement customer creation where defined by the API contract.

The expected flow is:

1. receive request
2. validate request
3. validate merchant ownership/context
4. normalize permitted fields
5. enforce documented business rules
6. call repository
7. map result
8. return API response

Do not perform database access directly.

---

# 11. Customer Validation

Implement validation based strictly on:

- docs/domain-model.md
- protobuf constraints
- documented business rules
- database-required fields

Validate:

- required customer identifiers
- required merchant association
- required contact data
- valid field formats where documented
- valid status values where applicable

Do not invent validation requirements.

---

# 12. Customer Identifier

Determine which identifier is authoritative.

Potential identifiers may include:

- internal customer ID
- merchant-scoped customer ID
- external customer ID
- reference ID

Follow the documented domain model.

Do not introduce a second identifier simply because it is convenient.

---

# 13. Customer Uniqueness

If customer uniqueness is scoped to a merchant:

respect that scope.

Do not treat:

customer ID X for Merchant A

as automatically equivalent to:

customer ID X for Merchant B.

Database constraints remain authoritative for concurrency-safe uniqueness.

Do not implement in-memory uniqueness checks.

---

# 14. Merchant Validation

If customer creation requires a valid merchant:

use the appropriate repository/service abstraction.

Do not directly query SQLC.

Determine whether merchant existence should be validated:

- through a merchant repository
- through a documented service dependency
- through a database foreign key

Follow the established architecture.

Do not introduce unnecessary duplicate database calls.

---

# 15. Avoid Circular Service Dependencies

Be alert to circular dependencies.

Do not automatically inject the entire MerchantService into CustomerService.

Prefer the smallest abstraction required by the documented architecture.

If customer creation requires merchant validation and the repository/database already guarantees the relationship:

do not create unnecessary service-to-service calls.

---

# 16. Get Customer

Implement customer retrieval according to the protobuf contract.

The service should:

1. validate the identifier
2. validate merchant/tenant context where applicable
3. call repository
4. handle not-found behavior
5. map result
6. return response

Do not expose SQLC types.

---

# 17. Merchant-Scoped Customer Retrieval

If the API requires:

Get customer X belonging to merchant Y

ensure both identifiers are respected.

Do not retrieve globally by customer ID if that would violate tenant isolation.

---

# 18. Customer Not Found

Follow the project's established not-found convention.

Do not expose:

- pgx errors
- SQLC errors
- raw SQL
- database internals

to the API.

Map the repository result into the established service/API error.

---

# 19. Update Customer

Implement update operations only if defined by the protobuf contract.

Follow the documented update semantics.

Determine whether updates use:

- explicit fields
- field masks
- full resource replacement

Do not invent update semantics.

---

# 20. Immutable Customer Fields

Protect fields that the domain defines as immutable.

Potential examples:

- internal customer ID
- merchant ownership
- creation timestamp

Do not allow arbitrary modification of identity/ownership fields.

Only fields explicitly documented as mutable may be changed.

---

# 21. Customer Status

If the domain defines customer statuses:

follow the documented values.

Do not create new status enums.

If status transitions have explicit business rules:

implement those rules in the service layer.

Do not treat the status as an arbitrary string.

---

# 22. Status Transitions

If customer lifecycle transitions are documented:

validate the transition:

current status → requested status

before repository update.

Do not allow invalid transitions.

Do not create a generic state machine framework unless the project already uses one.

---

# 23. List Customers

If customer listing is defined:

implement it using the repository's documented pagination model.

Do not:

- load all customers into memory
- paginate after fetching the complete table
- create a second pagination strategy

---

# 24. Customer Filtering

Support only documented filters.

Potential examples:

- merchant
- status
- external identifier
- creation range

Do not expose arbitrary SQL filters.

Do not dynamically construct SQL in the service.

---

# 25. Pagination Mapping

If the protobuf API has:

- page size
- page token
- next page token

map these consistently with:

docs/protobuf-strategy.md

and the repository's implementation.

Do not invent a new token encoding strategy.

---

# 26. Customer Relationships

Do not automatically load:

- transactions
- deposits
- payouts
- providers

when returning a Customer unless the protobuf response explicitly requires them.

Avoid N+1 queries.

Customer-related transaction operations belong to later agents.

---

# 27. Customer-to-Transaction Boundary

A customer may be referenced by transactions.

However, this agent must NOT implement transaction processing.

Do not:

- create deposits
- create payouts
- modify transaction status
- calculate balances
- call provider APIs

Only establish/maintain the customer resource itself.

---

# 28. Protobuf-to-Service Mapping

Do not pass protobuf messages into repository methods.

Convert request data into the service/repository representation expected by the architecture.

Follow the merchant implementation's mapping conventions.

---

# 29. Repository-to-Protobuf Mapping

Do not return SQLC-generated structs directly from gRPC handlers.

Map:

repository result

→ service/domain representation

→ protobuf response.

Follow the established project pattern.

---

# 30. Mapping Functions

If mappings are sufficiently complex:

create focused mapping functions.

Do not create a large universal mapping framework.

Keep mapping readable and local to the relevant package.

---

# 31. gRPC Service Implementation

Implement customer RPC methods in the generated Transactions gRPC service.

Follow the architecture established by Agent 06.

Handlers should be thin.

Expected flow:

gRPC request

→ validation/service invocation

→ service logic

→ response/error mapping

→ gRPC response

Do not place SQLC or database logic inside handlers.

---

# 32. REST/gRPC-Gateway

If Agent 04 defined HTTP annotations:

allow generated grpc-gateway code to expose the customer operations.

Do not create a second HTTP API.

Do not manually duplicate routes already defined in protobuf.

---

# 33. HTTP Error Behavior

Follow the existing gRPC/gateway error conventions.

Do not create a custom HTTP error framework.

Do not return database errors directly to HTTP clients.

---

# 34. Error Handling

Handle applicable customer errors consistently:

- invalid request
- merchant not found
- customer not found
- duplicate customer
- invalid status transition
- database failure

Use established project error conventions.

Do not introduce a new error hierarchy.

---

# 35. Duplicate Customer

If the repository/database reports a uniqueness violation:

map it using the same application error conventions established by the merchant service.

Do not perform unsafe:

check-then-create

logic as the only uniqueness protection.

---

# 36. Context Propagation

Every customer service method must receive and propagate context.

Do not use:

context.Background()

inside request-handling logic.

Do not discard request cancellation.

---

# 37. Logging

Follow the logging pattern used by the merchant service.

Do not introduce a new logging library.

Never log:

- authorization headers
- passwords
- OAuth tokens
- API keys
- secrets
- sensitive customer data unnecessarily

---

# 38. Provider Independence

Customer service logic must remain provider-agnostic.

Do not call:

- pawaPay
- GoHighLevel
- external payment providers
- external CRM APIs

from this service.

Provider integrations belong to other parts of the architecture.

---

# 39. No OAuth

Do not implement:

- OAuth authorization
- token exchange
- token storage
- SSO
- OAuth callbacks

These belong to the Clients/Integrations architecture.

---

# 40. No Webhooks

Do not implement webhook handlers.

Do not process provider webhook events.

Do not create webhook endpoints in this agent.

---

# 41. No Database Changes

Do not modify:

- migrations
- SQL query files
- SQLC configuration
- generated SQLC output

If a customer persistence problem is discovered:

document it for the relevant previous agent.

---

# 42. No Protobuf Changes

Do not modify:

- `.proto` files
- generated protobuf code
- generated gRPC code

If a customer RPC or field is missing:

document the issue for Agent 04.

---

# 43. No Repository Redesign

Do not modify the repository architecture.

If a customer operation is missing from the repository:

document it for Agent 05.

Do not call SQLC directly as a workaround.

---

# 44. Testing Scope

Agent 12 will handle comprehensive Transactions tests.

This agent must still ensure that the customer implementation is testable and compiles.

Add focused tests only where necessary to verify:

- customer validation
- service behavior
- mapping
- merchant/customer ownership behavior
- error mapping

Follow the project's existing test style.

Do not introduce a new testing framework.

---

# 45. Compile Validation

Run focused validation for the affected packages.

At minimum verify:

- customer service package compiles
- repository package compiles
- generated protobuf package compiles
- gRPC implementation compiles

---

# 46. Full Test Suite

If reasonably fast, run:

go test ./...

If unrelated tests fail:

do not modify unrelated code.

Record the failure.

---

# 47. Git Review

Before completing the agent:

run:

git status --short

Then:

git diff --stat

Then inspect the relevant diff.

Ensure changes are limited to the customer implementation and directly required files.

---

# 48. Scope Enforcement

Expected changes should generally be limited to:

- Transactions customer service files
- Transactions customer gRPC implementation
- directly required customer tests
- docs/transactions-customers-review.md

Do not modify:

- Clients service
- Deposits service
- Payout implementation
- provider integrations
- OAuth
- webhook infrastructure
- deployment
- Docker
- workflows
- third_party/

---

# 49. Generated Files

Do not manually edit generated files.

If generated protobuf/SQLC output appears incorrect:

document the problem for the responsible previous agent.

Do not patch generated output manually.

---

# 50. Deep Folder Protection

Do not modify or regenerate files under:

- third_party/
- third_party/googleapis/
- vendor/
- node_modules/

Do not investigate these directories unless an explicit compiler error requires confirming a dependency path.

If a tool enters one of these directories:

return immediately to the project files.

---

# 51. Customer Review Document

Create:

docs/transactions-customers-review.md

This document is mandatory.

---

# 52. Required Review Document Structure

Use exactly:

# Transactions Customers Implementation Review

## 1. Source Documents

List every document read.

## 2. Existing Implementation References

Document the relevant Deposits and Transactions merchant conventions used.

## 3. Customer Domain

Describe the customer domain implemented.

## 4. Merchant/Customer Relationship

Document ownership and tenant boundaries.

## 5. Service Structure

Document the customer service package and responsibilities.

## 6. Repository Integration

Document the repository methods consumed.

## 7. Validation

Document customer request/domain validation.

## 8. Customer Lifecycle

Document customer status/lifecycle rules if applicable.

## 9. RPC Implementation

Use:

| RPC | Purpose | Repository Operation |
|---|---|---|

## 10. REST Exposure

If applicable, use:

| Method | Route | RPC |
|---|---|---|

## 11. Error Handling

Document customer error mappings.

## 12. Mapping

Document:

- protobuf → service
- service → repository
- repository → service
- service → protobuf

## 13. Tenant Isolation

Document how merchant/customer ownership is protected.

## 14. Tests and Validation

Document tests and compilation performed.

## 15. Files Changed

List relevant files.

## 16. Risks

Document customer-specific risks.

## 17. Unresolved Questions

Document anything requiring later work.

---

# 53. Documentation Check

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
- docs/transactions-merchants-review.md was read

The final review document must describe the actual implementation.

---

# Completion Checklist

Before stopping, verify:

- [ ] Required documents were read.
- [ ] Existing Deposits customer-related conventions were reviewed.
- [ ] Transactions merchant implementation was reviewed.
- [ ] Customer domain was reviewed.
- [ ] Merchant/customer ownership model was understood.
- [ ] Customer repository interface was reviewed.
- [ ] Transactions protobuf contract was reviewed.
- [ ] Customer service package was implemented.
- [ ] Customer service dependencies are injected.
- [ ] Customer creation was implemented where required.
- [ ] Customer retrieval was implemented where required.
- [ ] Customer update was implemented where required.
- [ ] Customer listing was implemented where required.
- [ ] Customer validation was implemented.
- [ ] Merchant/customer ownership boundaries are enforced.
- [ ] Customer uniqueness follows the documented model.
- [ ] Customer lifecycle/status rules are enforced where documented.
- [ ] Repository interfaces are used rather than SQLC directly.
- [ ] SQLC models are not exposed through the API.
- [ ] Protobuf models are not passed into the repository.
- [ ] Error handling follows project conventions.
- [ ] Context is propagated.
- [ ] No provider API calls were added.
- [ ] No OAuth was added.
- [ ] No webhook functionality was added.
- [ ] No database schema was modified.
- [ ] No protobuf contract was modified.
- [ ] No repository redesign was performed.
- [ ] Relevant packages compile.
- [ ] Focused tests were run where practical.
- [ ] Full tests were run if practical.
- [ ] No unrelated files were modified.
- [ ] No third_party/googleapis files were modified.
- [ ] docs/transactions-customers-review.md was created.
- [ ] Git diff was reviewed.

---

# Final Stop Condition

STOP after completing:

1. Customer domain/service implementation
2. Customer gRPC implementation
3. Customer validation
4. Merchant/customer ownership enforcement
5. Repository integration
6. Focused validation
7. docs/transactions-customers-review.md

Do NOT proceed to:

- deposits
- payouts
- transaction processing
- runtime wiring
- application startup
- OAuth
- webhooks
- provider integrations
- Docker
- deployment

Those responsibilities belong to later Transactions agents.

If something required by the customer implementation is missing:

- missing database capability → document for Agent 02
- missing SQLC method → document for Agent 03
- missing protobuf contract → document for Agent 04
- missing repository method → document for Agent 05
- missing merchant capability → document for Agent 06

Do not silently modify previous agents' work.

STOP.