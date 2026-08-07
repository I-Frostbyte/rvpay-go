# Agent 01 — Review Existing Transactions Implementation

## Objective

Perform the initial audit of the existing RVPay repository before implementing the new Transactions Service.

The existing RVPay system contains a `deposits/` service.

The new architecture defines a broader Transactions Service that will eventually contain:

- merchants
- customers
- deposits
- payouts
- transaction lifecycle and state
- transaction-related persistence
- transaction-related APIs

This agent must determine how the existing `deposits/` implementation can be carried forward into the new Transactions Service.

This is an AUDIT AND DISCOVERY agent.

Do not implement the new Transactions Service.

Do not create new database tables.

Do not create new protobufs.

Do not modify the existing Deposits service.

Do not move files.

Do not rename packages.

Do not refactor existing code.

Do not create compatibility code.

The purpose of this agent is to understand the existing implementation and produce a precise implementation map for Agents 02–13.

---

# Required Reading

Read only:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md

These documents are mandatory.

Do not skip any of them.

Treat them as the primary architectural source of truth.

---

# Repository Map Rule

The root `README.md` is the repository map.

Use it to understand:

- what currently exists
- where existing services live
- what has already been implemented
- previous architectural decisions
- naming conventions
- service boundaries
- current deployment structure

Do not assume the repository has the structure described by the new architecture unless the documentation confirms it.

---

# Project Context Rule

`agents/project-context.md` contains the project's coding, package, dependency, generation, testing, and implementation conventions.

Follow it strictly.

Do not introduce conventions that contradict it.

Do not substitute personal Go style for the project's existing style.

---

# Foundation Documentation Rule

The following documents define the intended new architecture:

- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md

Use them to determine:

- what Transactions should own
- what Deposits currently owns
- what should move
- what should remain
- what should be shared
- what should eventually be replaced
- what should be preserved

Do not silently resolve contradictions.

If the existing implementation conflicts with the foundation documents, record the conflict.

---

# Exploration Rules

Use focused exploration.

Do not perform an unrestricted repository-wide review.

Do not recursively inspect directories that are unrelated to Transactions.

Do not inspect:

- third_party/
- googleapis/
- vendor/
- node_modules/
- .git/
- coverage/
- tmp/
- bin/

Do not recursively inspect generated dependency trees.

Do not spend time reviewing unrelated services.

---

# Allowed Existing Implementation Scope

Inspect only the following areas:

deposits/

protobuf/

grpc/go/

Makefile

go.mod

go.sum

README.md

.github/workflows/

Only inspect `.github/workflows/` files that are relevant to:

- Go generation
- testing
- Docker
- deployment
- Transactions/Deposits

Do not audit unrelated workflows.

---

# Deposits Audit

The existing `deposits/` service is the primary implementation reference.

Review its:

- directory structure
- command entrypoint
- configuration
- database structure
- migrations
- SQL queries
- sqlc output
- repositories
- service implementation
- protobuf contracts
- generated gRPC code
- Makefile
- Dockerfile
- tests
- runtime lifecycle

Do not modify any of these files.

---

# Existing Database Audit

Inspect:

deposits/db/

Determine:

- existing tables
- relationships
- primary keys
- foreign keys
- indexes
- unique constraints
- nullable fields
- timestamps
- status fields
- migration ordering
- migration naming
- up/down migration conventions
- transaction boundaries
- existing database ownership

Determine which database concepts are clearly Transactions-related.

Do not create or modify migrations.

---

# Existing SQLC Audit

Inspect:

deposits/db/query/

and:

deposits/db/sqlc/

Determine:

- SQL query conventions
- sqlc configuration
- generated model conventions
- generated query conventions
- nullable type handling
- transaction usage
- query naming
- repository interaction

Do not modify generated code.

Do not regenerate code.

---

# Existing Repository Audit

Inspect:

deposits/db/repo/

Determine:

- repository interfaces
- concrete repository implementations
- pool handling
- transaction handling
- error handling
- context propagation
- query usage
- mocking strategy

Determine which repository patterns should be reused by Transactions.

Do not refactor repositories.

---

# Existing Service Audit

Inspect the Deposits service implementation.

Determine:

- service interfaces
- business logic
- validation
- error handling
- dependency injection
- repository usage
- provider interaction
- transaction lifecycle
- logging
- context handling

Identify which responsibilities belong specifically to Deposits and which appear to be broader transaction responsibilities.

Do not change the implementation.

---

# Existing Protobuf Audit

Inspect:

protobuf/

and the generated code relevant to Deposits.

Determine:

- current package naming
- service definitions
- RPC naming
- request messages
- response messages
- HTTP annotations
- REST exposure
- field naming
- enum conventions
- error handling
- generation workflow

Do not modify protobuf files.

Do not regenerate protobuf output.

---

# Existing Runtime Audit

Inspect:

deposits/cmd/grpc-service/

Determine:

- startup lifecycle
- configuration loading
- logger setup
- database initialization
- migration execution
- repository construction
- service construction
- gRPC registration
- REST/gateway registration if present
- health checks
- signal handling
- graceful shutdown
- goroutine usage

Record which patterns should be preserved for Transactions.

Do not modify runtime code.

---

# Existing Configuration Audit

Inspect the existing Deposits configuration.

Determine:

- environment variables
- types
- defaults
- required values
- database configuration
- server configuration
- migration flags
- logging configuration
- provider configuration

Determine which configuration concepts should migrate to Transactions.

Do not change configuration.

---

# Existing Makefile Audit

Inspect:

deposits/Makefile

and the repository-level:

Makefile

Determine:

- generation commands
- test commands
- build commands
- run commands
- Docker commands
- protobuf commands
- sqlc commands
- migration commands

Record conventions that Transactions should follow.

Do not modify Makefiles.

---

# Existing Docker Audit

Inspect the Deposits Dockerfile.

Determine:

- build image
- runtime image
- build context
- binary path
- architecture
- environment handling
- port exposure
- entrypoint
- health check
- generation requirements

Determine what the Transactions Dockerfile should eventually follow.

Do not create a Dockerfile in this agent.

---

# Existing Test Audit

Inspect only tests directly associated with:

deposits/

Determine:

- test naming
- test package conventions
- mocks
- repository testing
- service testing
- integration testing
- fixtures
- database testing
- test helpers

Do not modify tests.

Do not run the complete repository test suite during this agent unless required to establish the existing baseline.

---

# Transactions Gap Analysis

Compare the existing Deposits implementation against the intended Transactions architecture.

Identify:

## Existing and Reusable

Components that can be carried forward with minimal change.

## Existing but Requires Modification

Components that exist but need to evolve for Transactions.

## New

Components that do not currently exist and must be implemented.

## Deprecated

Existing concepts that should not be carried into the new design.

## Unclear

Areas where the documentation does not provide enough information to determine the correct implementation.

Do not resolve the "Unclear" category by guessing.

---

# Service Boundary Analysis

Determine the intended boundaries between:

Transactions

Clients

other future RVPay services

Pay particular attention to:

- merchants
- customers
- deposits
- payouts
- transaction records
- payment provider integrations
- client/platform integrations

Do not move ownership between services based on personal preference.

Use:

docs/domain-model.md

and:

docs/repository-layout.md

as the authority.

---

# Deposits-to-Transactions Migration Analysis

Determine how the existing Deposits service relates to the new Transactions service.

Document:

- what is preserved
- what is renamed
- what is relocated
- what is expanded
- what is replaced
- what becomes shared
- what must remain backward-compatible

Do not perform the migration.

Do not move files.

Do not rename packages.

This agent only documents the required migration.

---

# Data Migration Analysis

Determine whether existing Deposits data would need to migrate into the new Transactions schema.

Identify:

- tables
- columns
- identifiers
- relationships
- status mappings
- provider identifiers
- timestamps
- historical records

Do not write migration scripts.

Do not modify existing migrations.

Document any required migration work for Agent 02.

---

# Protobuf Migration Analysis

Determine whether existing Deposits protobuf contracts can be reused.

Classify each existing RPC/message as:

- preserve
- extend
- replace
- deprecate
- unclear

Do not modify protobuf definitions.

Agent 04 will handle the actual protobuf implementation.

---

# Dependency Analysis

Identify dependencies that Transactions will inherit from Deposits.

Examples include:

- PostgreSQL
- pgx
- sqlc
- gRPC
- grpc-gateway
- zerolog
- migration tooling
- provider clients
- generated mocks

Do not add dependencies.

Do not upgrade dependencies.

Record the current versions where directly relevant.

---

# Security Audit

Review the existing Deposits implementation only for security characteristics relevant to Transactions.

Look for:

- hardcoded credentials
- secrets in source
- unsafe logging
- database credentials
- authentication assumptions
- authorization assumptions
- insecure configuration defaults
- sensitive data exposure

Do not perform a repository-wide secret scan.

Do not expose discovered secret values in the report.

---

# Operational Audit

Determine whether the existing Deposits patterns provide:

- health checks
- graceful shutdown
- structured logging
- startup validation
- database connectivity checks
- migration handling
- container readiness
- deployment compatibility

Identify which patterns Transactions should inherit.

---

# Do Not Implement

This agent must NOT:

- create `transactions/`
- create migrations
- modify Deposits
- modify protobufs
- generate sqlc
- generate mocks
- create repositories
- create services
- create handlers
- create configuration
- create Dockerfiles
- create Makefiles
- create tests
- change dependencies

This is strictly an audit.

---

# Required Output

Create:

docs/transactions-existing-review.md

Do not create any other implementation files.

---

# Review Document Structure

The document must contain:

# Transactions Service — Existing Implementation Review

## 1. Executive Summary

Summarize the current state of the repository and the existing Deposits implementation.

## 2. Existing Deposits Architecture

Document the current structure.

## 3. Existing Database

Document relevant tables, relationships, migrations, and ownership.

## 4. Existing SQLC

Document generation and query conventions.

## 5. Existing Repositories

Document repository patterns.

## 6. Existing Service Layer

Document business-service patterns.

## 7. Existing Protobuf

Document current RPC and message structure.

## 8. Existing Runtime

Document startup, dependency injection, server lifecycle, and shutdown.

## 9. Existing Configuration

Document relevant environment configuration.

## 10. Existing Docker and Build

Document container and build conventions.

## 11. Existing Testing

Document testing conventions.

## 12. Deposits → Transactions Gap Analysis

Use:

| Area | Existing | Target | Classification |
|---|---|---|---|

Classifications:

- Reusable
- Modify
- New
- Deprecated
- Unclear

## 13. Data Migration Considerations

Document possible migration requirements.

## 14. Protobuf Migration Considerations

Document existing contracts that need to evolve.

## 15. Service Boundary Considerations

Document ownership boundaries.

## 16. Risks

List architectural, technical, security, or migration risks.

## 17. Questions Requiring Architectural Decision

List unresolved questions.

Do not answer them by guessing.

## 18. Recommended Implementation Sequence

Provide a dependency-aware sequence for Agents 02–13.

Do not implement anything.

---

# Documentation Check

Before completing, verify that:

docs/transactions-existing-review.md

exists and accurately reflects the findings.

Do not create implementation files.

Do not modify:

README.md

agents/project-context.md

docs/domain-model.md

docs/repository-layout.md

docs/protobuf-strategy.md

docs/migration-plan.md

---

# Final Review

Before finishing:

1. Confirm all required documents were read.

2. Confirm the root README.md was used as the repository map.

3. Confirm agents/project-context.md was followed.

4. Confirm the foundation documents were used as the architecture source of truth.

5. Confirm only the permitted repository areas were inspected.

6. Confirm unrelated directories were not recursively explored.

7. Confirm no implementation changes were made.

8. Confirm no existing Deposits files were modified.

9. Confirm no generated code was modified.

10. Confirm no migrations were created or changed.

11. Confirm docs/transactions-existing-review.md exists.

12. Confirm the report distinguishes existing implementation from target architecture.

---

# Stop Condition

STOP after completing:

docs/transactions-existing-review.md

Do not proceed to database implementation.

Do not create Transactions code.

Do not modify Deposits.

Do not begin Agent 02 work.

Agent 02 will use this review to implement the Transactions database layer.