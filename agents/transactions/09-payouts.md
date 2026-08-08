# Agent 09 — Transactions Payouts Service

## Objective

Implement the Payouts portion of the Transactions Service.

This agent builds the Payouts capability on top of the Transactions foundation already implemented by Agents 02–08.

The implementation must integrate with:

- the Transactions database
- SQLC-generated persistence
- Transactions repositories
- Merchant domain
- Customer domain where applicable
- Transactions protobuf contracts
- existing Deposits implementation conventions
- the new RVPay domain model
- the documented Transactions architecture

The Payouts implementation is responsible for:

- payout creation
- payout retrieval
- payout listing where defined
- payout lifecycle/status handling
- merchant ownership
- customer association where defined
- payout validation
- repository interaction
- gRPC handlers
- REST/gRPC-gateway exposure where already defined
- domain/API mapping
- payout-specific error handling

This agent must NOT implement:

- deposits
- provider integrations
- webhook processing
- OAuth
- Clients service functionality
- application startup
- runtime wiring
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
- docs/transactions-deposits-review.md

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
- docs/transactions-deposits-review.md

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

The Transactions Merchant, Customer, and Deposit implementations are the primary references for the new Transactions architecture.

Inspect only the relevant files required to understand:

- existing payout behavior
- existing transaction behavior
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

# 1. Review Existing Payout Functionality

Use README.md to locate any existing payout-related implementation.

If payout functionality does not currently exist:

do not invent legacy behavior.

If payout functionality does exist:

inspect only the relevant files needed to understand:

- existing payout domain behavior
- request/response structures
- lifecycle
- repository behavior
- SQL queries
- protobuf contracts
- service implementation
- validation
- tests

Do not copy the old architecture blindly.

The new Transactions architecture is authoritative.

---

# 2. Compare Payout Domain With New Domain

Read:

docs/domain-model.md

Determine exactly how the new system defines a Payout.

Identify:

- payout identity
- merchant relationship
- customer relationship, if any
- amount
- currency
- destination
- status
- reference
- timestamps
- external/provider references
- lifecycle
- immutable fields

Do not invent fields.

Do not silently preserve legacy behavior that conflicts with the new domain.

---

# 3. Deposit vs Payout Boundary

Deposits and Payouts are separate transaction capabilities.

Do not reuse Deposit semantics simply because the structures look similar.

Conceptually:

Deposit:
inbound funds

Payout:
outbound funds

Follow the documented domain model rather than relying on this general distinction if the TDD defines more specific semantics.

Do not create a generic transaction implementation merely to avoid duplication.

---

# 4. Transactions Payout Boundary

The Payout capability belongs inside the Transactions service.

Do not create:

payouts/

as a new top-level microservice.

Follow:

docs/repository-layout.md

for exact package locations.

---

# 5. Inspect Transactions Database

Read:

docs/transactions-database-review.md

Determine the actual persistence model for payouts.

Identify:

- payout table
- primary key
- merchant relationship
- customer relationship where applicable
- amount
- currency
- destination information
- status
- reference
- timestamps
- external/provider references
- constraints
- indexes

Do not modify migrations.

If the database is missing something required by the documented payout domain:

STOP and document it for Agent 02.

---

# 6. Inspect SQLC

Read:

docs/transactions-sqlc-review.md

Determine the generated SQLC operations available for payouts.

Do not modify generated SQLC files.

Do not create raw SQL in the service layer.

If a required query does not exist:

document it for Agent 03.

Do not bypass the repository.

---

# 7. Inspect Transactions Repository

Read:

docs/transactions-repository-review.md

Inspect payout repository interfaces and implementation.

Determine:

- CreatePayout
- GetPayout
- ListPayouts
- UpdatePayout
- status operations
- merchant-scoped operations
- customer-scoped operations where applicable
- not-found behavior
- duplicate behavior
- pagination
- repository error conventions

Use repository abstractions.

Do not call SQLC directly.

---

# 8. Inspect Merchant Implementation

Read:

docs/transactions-merchants-review.md

Follow the same:

- constructor style
- dependency injection
- package organization
- service method conventions
- validation style
- mapping conventions
- error handling
- logging

Do not create a separate architectural pattern.

---

# 9. Inspect Customer Implementation

Read:

docs/transactions-customers-review.md

Determine whether Payouts are associated with Customers.

If they are:

reuse the existing Customer architecture.

Do not duplicate:

- customer lookup logic
- customer validation
- customer repository behavior
- customer ownership rules

If Payouts are not customer-scoped in the documented model:

do not introduce customer requirements.

---

# 10. Inspect Deposit Implementation

Read:

docs/transactions-deposits-review.md

Use the Deposit implementation as the closest Transactions-service reference.

Follow its:

- package organization
- constructor patterns
- mapping patterns
- validation patterns
- gRPC handler style
- error conventions
- documentation approach

Do not blindly duplicate Deposit code.

Payout-specific business rules must remain separate.

---

# 11. Inspect Transactions Protobuf

Read:

docs/transactions-protobuf-review.md

Inspect payout RPC definitions.

Determine:

- create payout RPC
- get payout RPC
- list payout RPC
- update/status RPCs
- request messages
- response messages
- resource representation
- pagination
- field masks
- HTTP annotations
- status representation

Do not modify protobuf contracts.

If a required RPC or field is missing:

document it for Agent 04.

---

# 12. Payout Identity

Determine the authoritative payout identifiers.

Possible identifiers may include:

- internal payout ID
- merchant reference
- external/provider reference

Use the documented model.

Do not invent a second identifier.

Do not use external provider IDs as internal primary keys unless explicitly documented.

---

# 13. Merchant Ownership

Every payout must belong to the correct merchant.

Tenant isolation is mandatory.

A payout belonging to:

Merchant A

must never be retrievable through a request scoped to:

Merchant B.

Prefer repository-level merchant filtering.

Do not retrieve globally and perform tenant checks only after the database query when the repository can enforce ownership.

---

# 14. Customer Association

If the domain associates Payouts with Customers:

validate the relationship.

Do not assume:

customer ID + merchant ID

is automatically valid.

Ensure the customer belongs to the merchant before associating the payout.

Follow the Customer implementation.

---

# 15. Payout Destination

Determine how the domain represents the payout destination.

Possible forms could include:

- bank account
- wallet
- mobile money account
- external beneficiary
- other documented destination

Do not invent destination types.

Use only fields and enums defined by:

docs/domain-model.md

and the protobuf contract.

---

# 16. Sensitive Destination Data

Treat payout destination information as sensitive.

Do not unnecessarily log:

- account numbers
- wallet numbers
- phone numbers
- bank details
- beneficiary credentials
- tokens

Return only the fields explicitly required by the API.

---

# 17. Payout Creation

Implement payout creation according to the protobuf/API contract.

Expected flow:

1. receive request
2. validate request
3. validate merchant context
4. validate customer context if applicable
5. validate amount
6. validate currency
7. validate destination
8. validate reference/idempotency information
9. construct service/domain input
10. call repository
11. map result
12. return response

Do not call an external payout provider.

---

# 18. Amount Validation

Follow the project's monetary representation.

Validate:

- amount is present where required
- amount is positive where required
- amount precision is valid
- amount does not violate documented constraints

Do not use floating-point arithmetic for money if the project uses integer/decimal representations.

Follow existing Transactions conventions.

---

# 19. Currency Validation

Use the existing currency model.

Do not create another currency representation.

Validate supported currencies only where the domain requires it.

Do not hard-code an unrelated provider's currency list.

---

# 20. Destination Validation

Validate payout destination according to the documented domain.

Examples may include:

- required destination identifier
- required destination type
- required beneficiary information

Do not create provider-specific validation here.

Provider-specific requirements belong to provider integration.

---

# 21. Reference Validation

If payout references must be unique:

respect the database uniqueness constraint.

Do not implement application-only check-then-create logic.

If the repository reports a uniqueness violation:

map it to the established conflict error.

---

# 22. Idempotency

If the documented API defines payout idempotency:

preserve it.

A retried request with the same idempotency/reference key must follow the documented behavior.

Do not create duplicate payouts from request retries.

Do not invent a new idempotency mechanism if the architecture does not define one.

---

# 23. Payout Status

Use only statuses defined by:

- docs/domain-model.md
- protobuf definitions
- repository model

Do not create new statuses.

Do not represent status as arbitrary free-form strings if the architecture uses enums.

---

# 24. Payout State Machine

If the domain defines payout transitions:

implement them explicitly.

For example, a payout might move through states such as:

pending → processing → successful

or:

pending → failed

But these are examples only.

Use the actual documented lifecycle.

Do not invent transitions.

---

# 25. Terminal States

Determine whether statuses such as:

- successful
- failed
- cancelled

are terminal.

If the domain says a status is terminal:

do not allow arbitrary transitions away from it.

Do not implement a generic state machine framework.

---

# 26. Invalid State Transitions

Reject transitions that violate the documented lifecycle.

Do not silently accept:

terminal → pending

or any other invalid transition.

Use the project's existing application error convention.

---

# 27. Financial Record Integrity

Payouts are financial records.

Do not expose unrestricted updates.

Protect immutable fields such as:

- payout ID
- merchant ownership
- customer ownership
- creation timestamp
- original amount
- original currency

unless the domain explicitly states otherwise.

---

# 28. Status vs Financial Mutation

Do not treat a status update as permission to modify financial fields.

For example:

changing payout status

must not implicitly modify:

- amount
- currency
- merchant
- customer
- destination

unless explicitly defined.

Keep lifecycle mutation separate from financial record mutation.

---

# 29. Provider Independence

This agent must remain provider-agnostic.

Do not call:

- pawaPay
- banks
- mobile-money APIs
- external payment gateways
- GoHighLevel
- external financial APIs

The Payout service represents internal transaction state.

Provider execution belongs to a separate integration boundary.

---

# 30. Payout Retrieval

Implement GetPayout if defined.

Expected flow:

1. validate payout ID
2. validate merchant context
3. call repository
4. handle not-found
5. map repository result
6. return response

Do not expose SQLC structs.

---

# 31. Merchant-Scoped Retrieval

Where supported:

retrieve payouts using merchant-scoped repository operations.

Do not:

1. fetch a payout globally
2. compare merchant IDs only after retrieval

if the repository can perform the ownership filter.

---

# 32. Customer-Scoped Retrieval

If the API supports customer payout queries:

ensure:

Customer belongs to Merchant

before returning results.

Do not expose cross-merchant payout records.

---

# 33. List Payouts

If defined:

implement using repository pagination.

Do not:

- load the entire payout table
- paginate in memory
- introduce a second pagination system

---

# 34. Payout Filters

Only support documented filters.

Potential filters may include:

- merchant
- customer
- status
- currency
- destination type
- reference
- creation date range

Do not expose arbitrary SQL conditions.

---

# 35. Pagination

Follow:

docs/protobuf-strategy.md

and the Transactions repository pagination behavior.

Preserve:

- page size
- page token
- next page token

semantics.

Do not invent a new token encoding.

---

# 36. Update Payout

Implement updates only where explicitly supported.

Do not create a generic unrestricted update operation.

Determine whether the API allows:

- status updates
- metadata updates
- mutable destination data

based on the actual documented domain.

---

# 37. Status Update Operation

If the protobuf exposes a status update:

validate:

- payout exists
- merchant owns payout
- current state allows transition
- requested state is valid

Then call the repository.

Do not directly modify SQLC state from the handler.

---

# 38. Repository Transactions

If the repository exposes atomic payout operations:

use them.

Do not manually compose multiple database writes if an atomic repository method already exists.

Do not introduce distributed transaction logic.

---

# 39. Protobuf-to-Service Mapping

Do not pass protobuf requests into repositories.

Convert:

protobuf request

→ service/domain input

→ repository input.

Follow the mapping conventions used by Merchants, Customers, and Deposits.

---

# 40. Repository-to-Protobuf Mapping

Do not expose SQLC-generated structs through gRPC.

Map:

repository result

→ service/domain result

→ protobuf response.

---

# 41. Mapping Functions

If mappings become non-trivial:

create focused mapping functions.

Do not introduce a universal mapping framework.

Keep mappings readable and close to the relevant service.

---

# 42. gRPC Handler

Implement payout RPC methods in the Transactions gRPC service.

Handlers must remain thin.

Expected flow:

gRPC request

→ validation/service invocation

→ service logic

→ response/error mapping

→ gRPC response

Do not place database or SQLC logic inside handlers.

---

# 43. REST/gRPC-Gateway

If payout RPCs have HTTP annotations:

use generated grpc-gateway exposure.

Do not create duplicate REST handlers.

Do not manually maintain routes generated from protobuf.

---

# 44. REST Verification

Verify that the payout HTTP endpoints correspond to the protobuf contract.

Check:

- method
- path
- request body
- response
- error behavior

Do not manually modify generated gateway code.

---

# 45. Error Handling

Handle applicable payout errors:

- invalid request
- invalid amount
- invalid currency
- invalid destination
- merchant not found
- customer not found
- payout not found
- duplicate reference
- invalid state transition
- database failure

Use existing project error conventions.

Do not expose:

- PostgreSQL errors
- pgx errors
- SQLC errors
- stack traces

to API clients.

---

# 46. Not Found

Map repository not-found behavior to the established application/API error.

A missing payout should not become an internal server error.

---

# 47. Duplicate Payout

If the database/repository reports a uniqueness violation:

map it to the established conflict/idempotency error.

Do not rely solely on an application-level duplicate check.

---

# 48. Context Propagation

Every request must propagate context.

Do not use:

context.Background()

inside request handling.

Respect cancellation and deadlines.

---

# 49. Logging

Use the established project logger.

Log useful operational information.

Never log:

- API keys
- access tokens
- passwords
- secrets
- authorization headers
- complete bank details
- complete wallet/account identifiers unnecessarily

---

# 50. No Provider Interface

Do not introduce provider abstractions into the Payout service merely because future provider execution will be required.

The payout domain should remain independent of provider implementations.

If the architecture explicitly already defines a provider boundary:

use it.

Otherwise document provider execution as future work.

---

# 51. No Webhooks

Do not implement:

- webhook endpoints
- webhook verification
- callback processing
- provider event consumers

Those belong outside this agent.

---

# 52. No OAuth

Do not implement:

- OAuth
- SSO
- token exchange
- authorization callbacks
- client credentials

Those belong to the Clients service.

---

# 53. No Database Changes

Do not modify:

- migrations
- schema
- SQL query files
- SQLC configuration
- generated SQLC files

If persistence is incomplete:

document the required change for Agent 02 or Agent 03.

---

# 54. No Protobuf Changes

Do not modify:

- .proto files
- generated protobuf files
- generated gRPC files
- generated gateway files

If the payout API contract is incomplete:

document the problem for Agent 04.

---

# 55. No Repository Redesign

Do not redesign repository interfaces.

If a payout repository method is missing:

document it for Agent 05.

Do not bypass the repository.

---

# 56. Existing Code Migration

If existing payout functionality exists elsewhere:

do not delete it automatically.

Use:

docs/migration-plan.md

to determine whether it is:

- retained
- replaced
- migrated
- deprecated

Do not perform broad cleanup in this agent.

---

# 57. Testing Scope

Agent 12 handles comprehensive Transactions testing.

This agent must still make the Payout implementation testable.

Add focused tests where necessary for:

- payout validation
- amount validation
- destination validation
- merchant ownership
- customer association
- state transitions
- mapping
- error handling
- service behavior

Follow existing test conventions.

Do not introduce a new testing framework.

---

# 58. Compile Validation

Run focused validation for affected packages.

At minimum verify:

- payout service compiles
- Transactions service compiles
- gRPC implementation compiles
- repository dependency compiles
- protobuf package compiles

---

# 59. Full Test Suite

If reasonably fast:

go test ./...

If unrelated tests fail:

do not modify unrelated code.

Record the failure.

---

# 60. Git Review

Before completion:

run:

git status --short

Then:

git diff --stat

Then inspect the relevant diff.

Confirm that modifications are limited to Payout implementation and directly required documentation/tests.

---

# 61. Scope Enforcement

Expected changes should generally be limited to:

- Transactions Payout service files
- Transactions gRPC implementation
- directly required Payout tests
- docs/transactions-payouts-review.md

Do not modify:

- Clients service
- Deposits implementation
- Merchant implementation
- Customer implementation
- OAuth
- webhooks
- provider integrations
- deployment
- Docker
- workflows
- third_party/

---

# 62. Generated Files

Never manually modify generated files.

If generated protobuf/SQLC/gateway output is incorrect:

document the issue for the responsible agent.

Do not patch generated output manually.

---

# 63. Deep Folder Protection

Do not modify or regenerate files under:

- third_party/
- third_party/googleapis/
- vendor/
- node_modules/

Do not inspect these directories recursively.

If a tool enters one because of a dependency:

return immediately to project files.

---

# 64. Payout Review Document

Create:

docs/transactions-payouts-review.md

This document is mandatory.

---

# 65. Required Review Document Structure

Use exactly:

# Transactions Payouts Implementation Review

## 1. Source Documents

List every document read.

## 2. Existing Payout Implementation

If legacy payout functionality exists, document:

- files inspected
- behavior retained
- behavior changed
- migration considerations

If no existing payout implementation exists, state that explicitly.

## 3. Domain Model

Document the implemented Payout domain.

## 4. Merchant Relationship

Document merchant ownership and tenant isolation.

## 5. Customer Relationship

Document customer association if applicable.

## 6. Destination

Document the payout destination model.

## 7. Payout Lifecycle

Use:

| Current State | Allowed State | Rule |
|---|---|---|

Document only actual supported transitions.

## 8. Service Structure

Document the Payout service package and responsibilities.

## 9. Repository Integration

Use:

| Service Operation | Repository Operation |
|---|---|

## 10. RPC Implementation

Use:

| RPC | Purpose | Service Operation |
|---|---|---|

## 11. REST Exposure

If applicable:

| Method | Route | RPC |
|---|---|---|

## 12. Validation

Document:

- amount
- currency
- destination
- merchant
- customer
- references
- status

## 13. Error Handling

Use:

| Condition | Application/API Error |
|---|---|

## 14. Mapping

Document:

- protobuf → service
- service → repository
- repository → service
- service → protobuf

## 15. Idempotency

Document how payout idempotency/reference uniqueness is handled if applicable.

## 16. Provider Boundary

Explicitly document what provider functionality is intentionally outside this service.

## 17. Security/Sensitive Data

Document how payout destination information is protected from unnecessary exposure/logging.

## 18. Tests and Validation

Document:

- focused tests
- compilation
- full test results

## 19. Files Changed

List relevant files.

## 20. Risks

Document payout-specific risks.

## 21. Unresolved Questions

Document anything requiring later agents or architectural decisions.

---

# 66. Documentation Check

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
- docs/transactions-deposits-review.md was read

The review document must reflect the actual implementation.

---

# Completion Checklist

Before stopping, verify:

- [ ] Required documents were read.
- [ ] Existing payout functionality was reviewed where present.
- [ ] New Payout domain was reviewed.
- [ ] Transactions database model was reviewed.
- [ ] SQLC payout operations were reviewed.
- [ ] Transactions repository was reviewed.
- [ ] Merchant implementation was reviewed.
- [ ] Customer implementation was reviewed.
- [ ] Deposit implementation was reviewed.
- [ ] Payout protobuf contract was reviewed.
- [ ] Payout service was implemented.
- [ ] Payout creation was implemented where required.
- [ ] Payout retrieval was implemented where required.
- [ ] Payout listing was implemented where required.
- [ ] Payout updates were implemented only where documented.
- [ ] Payout lifecycle rules were implemented.
- [ ] Invalid state transitions are rejected.
- [ ] Merchant ownership is enforced.
- [ ] Customer association is validated where applicable.
- [ ] Amount validation follows the domain.
- [ ] Currency validation follows the domain.
- [ ] Destination validation follows the domain.
- [ ] Reference/idempotency rules are respected.
- [ ] Financial record integrity is protected.
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
- [ ] Sensitive payout data is not unnecessarily logged.
- [ ] Focused validation was performed.
- [ ] Full tests were run if practical.
- [ ] Git diff was reviewed.
- [ ] docs/transactions-payouts-review.md was created.
- [ ] Documentation check was completed.

---

# Final Stop Condition

STOP after completing:

1. Payout domain/service implementation
2. Payout repository integration
3. Payout validation
4. Payout lifecycle handling
5. Merchant/customer ownership enforcement
6. Payout gRPC implementation
7. REST/gRPC-gateway integration where already defined
8. Focused validation
9. docs/transactions-payouts-review.md

Do NOT proceed to:

- runtime wiring
- application startup
- provider execution
- OAuth
- webhooks
- Docker
- deployment

Those responsibilities belong to later agents.

If something required by the Payout implementation is missing:

- missing database capability → document for Agent 02
- missing SQLC operation → document for Agent 03
- missing protobuf contract → document for Agent 04
- missing repository operation → document for Agent 05
- missing merchant capability → document for Agent 06
- missing customer capability → document for Agent 07
- missing deposit behavior that affects shared transaction architecture → document it, but do not modify Agent 08's implementation

Do not silently modify previous agents' work.

STOP.