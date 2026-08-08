# Agent 04 — Transactions Protobuf Contracts and Generation

## Objective

Implement and verify the protobuf and gRPC API contracts for the new Transactions Service.

The existing RVPay repository contains protobuf contracts for the existing Deposits service.

The new Transactions Service expands the transaction domain to include the entities and operations defined by the foundation documentation.

This agent is responsible for:

- defining Transactions protobuf contracts
- defining Transactions gRPC services
- defining request/response messages
- defining enums where required
- defining HTTP/REST mappings where documented
- configuring protobuf generation
- generating Go protobuf code
- generating Go gRPC code
- generating grpc-gateway code where required
- validating generated output
- documenting the resulting API contract

This agent must NOT implement:

- database schema
- SQL queries
- repositories
- business logic
- service implementations
- authentication
- OAuth
- webhook processing
- runtime wiring
- Docker
- deployment
- application tests

Those responsibilities belong to later agents.

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

These documents are mandatory.

Do not skip any of them.

---

# Architecture Source of Truth

Use the following hierarchy:

1. README.md
2. agents/project-context.md
3. docs/domain-model.md
4. docs/repository-layout.md
5. docs/protobuf-strategy.md
6. docs/migration-plan.md
7. docs/transactions-existing-review.md
8. docs/transactions-database-review.md
9. docs/transactions-sqlc-review.md
10. existing protobuf implementation

Do not invent API behavior that is not supported by the documentation.

If an API decision is unclear:

- identify the ambiguity
- document it
- do not silently invent a contract

---

# Repository Exploration Rules

Use the root README.md as the repository map.

Perform focused exploration only.

Do not perform an unrestricted repository-wide scan.

Do not recursively inspect:

- third_party/
- googleapis/
- vendor/
- node_modules/
- .git/
- coverage/
- tmp/
- bin/

In particular:

DO NOT recursively inspect:

third_party/googleapis/

The googleapis directory is a dependency/submodule.

Only inspect specific imported Google API definitions if required to understand an existing protobuf annotation.

Do not spend time reviewing unrelated services.

---

# Documentation Check

Before modifying anything, confirm that you have read:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md
- docs/transactions-existing-review.md
- docs/transactions-database-review.md
- docs/transactions-sqlc-review.md

At completion, verify that the protobuf implementation agrees with those documents.

---

# 1. Inspect Existing Protobuf Structure

Inspect the existing:

protobuf/

directory.

Determine:

- `.proto` file naming
- package naming
- directory structure
- Go package options
- service definitions
- message naming
- enum conventions
- field numbering conventions
- HTTP annotations
- imports
- generation commands
- Makefile integration
- generated output location

Use the existing Deposits protobuf as the primary implementation reference.

Do not modify existing Deposits protobuf files unless explicitly required by the migration plan.

---

# 2. Inspect Existing Deposits API

Inspect only the relevant Deposits `.proto` files.

Determine:

- existing RPC style
- request naming
- response naming
- service naming
- field naming
- status representation
- error representation
- pagination conventions
- timestamps
- resource identifiers
- HTTP annotations
- REST exposure

Do not blindly copy the Deposits API.

Preserve established conventions while adapting them to the new Transactions architecture.

---

# 3. Review Protobuf Strategy

Read:

docs/protobuf-strategy.md

This document is authoritative for:

- service boundaries
- protobuf package structure
- shared messages
- generated code
- REST/gRPC strategy
- API versioning
- resource naming
- compatibility

Follow it exactly.

Do not introduce an alternative protobuf architecture.

---

# 4. Determine Transactions API Boundary

Use:

docs/domain-model.md

to determine what Transactions exposes.

Evaluate the documented API surface for:

- merchants
- customers
- transactions
- deposits
- payouts

Do not expose internal database concepts simply because tables exist.

The protobuf API represents the service contract, not the database schema.

---

# 5. Service Definition

Create the Transactions gRPC service according to:

docs/protobuf-strategy.md

Use the project's established service naming convention.

The service should expose only documented operations.

Do not add speculative administrative RPCs.

Do not add debugging RPCs.

Do not expose raw database operations.

---

# 6. RPC Design

For every RPC:

- use explicit request messages
- use explicit response messages
- use appropriate resource identifiers
- use google.protobuf.Timestamp where the project convention requires it
- use documented enums
- avoid leaking database implementation details

Do not reuse request messages merely to reduce the number of protobuf definitions if the semantics differ.

---

# 7. Merchant API

If merchants are owned by Transactions according to the domain model, define the required RPCs.

Possible operations may include:

- CreateMerchant
- GetMerchant
- UpdateMerchant
- ListMerchants

Only implement operations supported by the architecture.

Do not automatically create full CRUD.

Do not expose database-specific fields.

---

# 8. Customer API

If customers are owned by Transactions, define the required API.

Possible operations may include:

- CreateCustomer
- GetCustomer
- UpdateCustomer
- ListCustomers

Follow the actual documented use cases.

Ensure merchant/customer relationships are represented correctly.

Do not expose internal database foreign keys unless they are part of the public API contract.

---

# 9. Transaction API

Define the core Transactions API.

Potential operations include:

- CreateTransaction
- GetTransaction
- ListTransactions
- UpdateTransaction
- GetTransactionStatus

Only expose operations supported by the domain model.

Do not expose unrestricted state mutation.

For example, avoid generic:

UpdateTransactionStatus

if the architecture requires specific lifecycle operations instead.

Follow the documented transaction lifecycle.

---

# 10. Deposit API

Determine how Deposits are represented in the new Transactions API.

Use:

docs/domain-model.md

and:

docs/migration-plan.md

to determine whether Deposits are:

- a transaction type
- a transaction-specific resource
- a sub-resource
- a compatibility API
- another documented structure

Do not preserve the old Deposits API automatically.

If backward compatibility is required, document how it will be handled.

---

# 11. Payout API

If Payouts are part of Transactions, define their protobuf API according to the architecture.

Potential operations include:

- CreatePayout
- GetPayout
- ListPayouts
- GetPayoutStatus

Only implement documented operations.

Do not create payout operations that belong to another service.

---

# 12. Resource Naming

Follow:

docs/protobuf-strategy.md

for resource naming.

Use stable public identifiers.

Do not expose:

- SQL row IDs
- internal database implementation details
- database-specific naming

unless explicitly required by the API strategy.

---

# 13. Field Naming

Follow standard protobuf naming conventions and the existing project style.

Use:

snake_case

inside `.proto` definitions.

Ensure generated Go fields follow the expected conventions.

Do not introduce inconsistent abbreviations.

For example, use:

provider_reference

rather than an arbitrary variation such as:

providerRef

inside the protobuf schema.

---

# 14. Field Numbering

Treat protobuf field numbers as permanent API identifiers.

Rules:

- never reuse a previously assigned field number
- never change the meaning of an existing field number
- reserve removed fields where appropriate
- do not reorder fields merely for aesthetics

If modifying an existing message:

inspect its current field numbers first.

Do not alter existing Deposits messages casually.

---

# 15. Enum Design

Use protobuf enums where the domain requires a finite set of states.

Potential examples:

- transaction type
- transaction status
- payout status
- deposit status
- provider type

Only define enums supported by the architecture.

Use an explicit unspecified/unknown zero value when consistent with project conventions.

Do not invent status values.

The database and domain documentation must agree with the protobuf enum values.

---

# 16. Status Representation

Compare:

- docs/domain-model.md
- docs/transactions-database-review.md
- existing Deposits protobuf

Ensure status values are consistent.

Do not create protobuf statuses that cannot be represented by the database.

Do not create database statuses solely because they are convenient for protobuf.

If the representations differ intentionally, document the mapping.

---

# 17. Monetary Representation

Follow the existing project and protobuf strategy for representing monetary values.

Do not use floating-point protobuf fields for money unless explicitly required.

If the project defines a shared money representation:

reuse it.

Do not create a second money representation for Transactions.

---

# 18. Currency Representation

Follow the documented currency strategy.

Use the project's existing convention for:

- currency codes
- currency strings
- enums
- validation

Do not introduce a custom currency enum unless explicitly required.

---

# 19. Timestamp Representation

Use the project's established timestamp convention.

Prefer:

google.protobuf.Timestamp

where the project strategy requires it.

Do not represent timestamps as arbitrary strings.

Do not introduce Unix integer timestamps if the project uses protobuf timestamps.

---

# 20. Pagination

Determine whether Transactions list operations require pagination.

Use:

docs/protobuf-strategy.md

as the authority.

If pagination is required:

follow the existing project's pagination request/response convention.

Do not invent a new pagination format.

Ensure pagination is deterministic.

---

# 21. Filtering

If list operations support filtering:

use explicit request fields.

Do not accept arbitrary SQL-like filter expressions.

Potential filters may include documented:

- merchant
- customer
- status
- type
- provider
- date range

Only include filters supported by the architecture.

---

# 22. Error Contract

Review the existing protobuf/error strategy.

Do not invent a custom error envelope unless the project strategy requires it.

Use the project's established gRPC error approach.

Do not encode errors as arbitrary strings inside every response.

---

# 23. HTTP/REST Exposure

If:

docs/protobuf-strategy.md

requires grpc-gateway REST exposure, add the appropriate HTTP annotations.

Ensure HTTP routes are:

- stable
- resource-oriented
- unambiguous
- consistent with existing API conventions

Do not expose internal gRPC methods over HTTP automatically.

Only annotate public operations.

---

# 24. REST Route Design

Follow the project's existing REST convention.

For example, use resource-oriented routes such as:

GET /v1/transactions/{transaction_id}

rather than database-style routes.

Do not choose routes based solely on convenience.

Use the documented API strategy.

---

# 25. HTTP Method Semantics

Use HTTP methods correctly.

Generally:

GET

for retrieval.

POST

for creation or command-like operations.

PATCH

for partial updates where supported.

PUT

only where the API strategy requires replacement semantics.

DELETE

only where deletion is part of the documented domain.

Do not expose DELETE merely because CRUD exists in the database.

---

# 26. Query Parameters

Use query parameters for documented filtering/pagination.

Do not encode arbitrary application commands into query strings.

Keep route variables for resource identifiers.

---

# 27. API Versioning

Follow:

docs/protobuf-strategy.md

for package and API versioning.

Do not invent a versioning system.

If the existing project uses:

v1

continue the convention.

If the strategy requires another structure, follow the strategy.

---

# 28. Shared Protobuf Types

Before creating a new message:

check whether the protobuf strategy defines an existing shared message.

Avoid duplicating:

- money
- pagination
- timestamps
- common identifiers
- common metadata

If a shared type belongs in the shared protobuf layer:

use it rather than creating a Transactions-specific duplicate.

---

# 29. Protobuf Package

Place Transactions protobuf definitions exactly where:

docs/repository-layout.md

requires.

Do not place Transactions definitions inside:

deposits/

Do not modify the Clients protobuf package.

Use the documented package structure.

---

# 30. Go Package Options

Follow the existing protobuf conventions for:

option go_package

Ensure generated Go code lands in the expected:

grpc/go/

location.

Do not invent a new generated-code directory.

---

# 31. Google API Imports

Only import Google API protobuf definitions actually required by the contracts.

Examples may include:

google/api/annotations.proto

google/protobuf/timestamp.proto

google/protobuf/empty.proto

Do not add imports speculatively.

Do not recursively inspect:

third_party/googleapis/

Only use the required definitions.

---

# 32. Protobuf Generation

Use the existing repository generation mechanism.

If the project uses:

protobuf/Makefile

then use it.

If generation is defined by the root Makefile, follow that mechanism.

Do not invent a separate generation workflow.

---

# 33. Generator Versions

Use the exact generator versions defined by the project.

Do not upgrade:

- protoc
- protoc-gen-go
- protoc-gen-go-grpc
- protoc-gen-grpc-gateway

unless the project documentation explicitly requires it.

Generated code must remain reproducible with CI.

---

# 34. Generated Output

Generate the appropriate Go output for:

- protobuf messages
- gRPC service stubs
- grpc-gateway handlers where required

Generated output must go into the documented location.

Do not manually edit generated files.

If generated output is wrong:

fix the `.proto` files or generation configuration and regenerate.

---

# 35. Existing Generated Code

Do not modify unrelated generated code.

Especially do not regenerate the entire repository if the generation command would produce unrelated changes.

If the repository generation mechanism necessarily regenerates multiple services:

inspect the diff carefully.

Only commit changes that are expected and relevant.

---

# 36. Backward Compatibility

Review:

docs/migration-plan.md

for compatibility requirements.

If existing Deposits clients depend on the current protobuf API:

do not break them accidentally.

Determine whether the migration strategy requires:

- preserving existing messages
- preserving existing RPCs
- introducing new RPCs
- deprecating old RPCs
- maintaining old field numbers

Do not delete an existing API merely because the new architecture looks cleaner.

---

# 37. API Compatibility Review

For every existing Deposits protobuf definition that is affected:

record:

- preserve
- extend
- replace
- deprecate
- remove

Only perform the documented changes.

If the migration plan is unclear:

do not guess.

Document the issue.

---

# 38. Compile Validation

After generation:

run the relevant Go compilation checks.

Verify:

- generated protobuf package compiles
- generated gRPC package compiles
- generated gateway package compiles where applicable
- imports resolve
- package names are correct

Do not proceed to service implementation.

---

# 39. Protobuf Validation

If the repository has protobuf linting or validation:

run it.

Check for:

- duplicate field numbers
- invalid imports
- invalid package names
- missing go_package
- incompatible types
- invalid HTTP annotations

Do not suppress errors simply to make generation pass.

---

# 40. Generated Code Reproducibility

Run protobuf generation a second time.

Confirm that the generated output is stable.

A second generation should not introduce unexpected changes.

If generated output changes between runs:

investigate.

Do not ignore nondeterministic generation.

---

# 41. API Documentation

Create:

docs/transactions-protobuf-review.md

This is mandatory.

The document must describe the resulting Transactions API.

---

# 42. Required Documentation Structure

Use exactly:

# Transactions Protobuf Implementation Review

## 1. Source Documents

List the documents used.

## 2. Existing Deposits API

Summarize the relevant existing protobuf contracts.

## 3. Transactions Service

Describe the Transactions gRPC service.

## 4. RPCs

List every RPC.

Use:

| RPC | Purpose | HTTP Exposure |
|---|---|---|

## 5. Messages

List the major request/response/resource messages.

## 6. Enums

List the enums and their meanings.

## 7. Shared Types

Document reused shared protobuf types.

## 8. REST Exposure

Document HTTP routes.

Use:

| Method | Route | RPC |
|---|---|---|

## 9. Compatibility

Explain how the new API relates to the existing Deposits API.

## 10. Generated Output

Document generated files/directories.

## 11. Generator Versions

Document the exact versions used.

## 12. Validation

Document generation and compile checks.

## 13. Risks

Document compatibility/API risks.

## 14. Unresolved Questions

Document anything that remains unclear.

---

# 43. Diff Review

Before completing:

run:

git status --short

Then:

git diff --stat

Then inspect the relevant protobuf/generated-code diff.

Expected changes should be limited to:

- Transactions `.proto` files
- protobuf generation configuration if required
- generated Transactions Go code
- generated gateway code where required
- docs/transactions-protobuf-review.md

Do not modify unrelated services.

Do not modify Deposits generated code unless explicitly required by the migration plan.

---

# 44. Documentation Check

Before completion, verify again that:

- README.md was read
- agents/project-context.md was read
- docs/domain-model.md was read
- docs/repository-layout.md was read
- docs/protobuf-strategy.md was read
- docs/migration-plan.md was read
- docs/transactions-existing-review.md was read
- docs/transactions-database-review.md was read
- docs/transactions-sqlc-review.md was read

Confirm that the protobuf implementation follows these documents.

---

# Completion Checklist

Before stopping, confirm:

- [ ] Existing Deposits protobuf conventions were reviewed.
- [ ] Transactions API ownership was established.
- [ ] Transactions protobuf package was created in the correct location.
- [ ] Transactions gRPC service was defined.
- [ ] Required request/response messages were defined.
- [ ] Required enums were defined.
- [ ] Field numbering was reviewed.
- [ ] Shared protobuf types were reused where appropriate.
- [ ] Monetary representation follows project conventions.
- [ ] Timestamp representation follows project conventions.
- [ ] Pagination follows project conventions where required.
- [ ] REST annotations were added where required.
- [ ] REST routes follow the documented strategy.
- [ ] API versioning follows project conventions.
- [ ] Existing Deposits compatibility was reviewed.
- [ ] Generator versions were not changed.
- [ ] Generated Go protobuf code was created.
- [ ] Generated gRPC code was created.
- [ ] Generated gateway code was created where required.
- [ ] Generated code was not manually edited.
- [ ] Protobuf generation is reproducible.
- [ ] Generated packages compile.
- [ ] docs/transactions-protobuf-review.md was created.
- [ ] No unrelated files were modified.

---

# Final Stop Condition

STOP after completing:

1. Transactions `.proto` definitions
2. protobuf/gRPC/gateway generation
3. generated Go output
4. protobuf validation
5. docs/transactions-protobuf-review.md

Do NOT:

- create repositories
- create service implementations
- create handlers
- create runtime wiring
- implement business logic
- implement merchants
- implement customers
- implement deposits
- implement payouts
- create Dockerfiles
- modify Clients
- modify unrelated services
- begin Agent 05

Agent 05 will consume the Transactions SQLC layer and protobuf contracts to implement the repository layer.