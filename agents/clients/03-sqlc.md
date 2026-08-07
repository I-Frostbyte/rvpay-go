# Agent 03 — SQLC Generation & Persistence Layer

## Objective

Generate the persistence layer for the Clients service from the database schema created in Agent 02.

This agent owns the SQL abstraction layer only.

It is responsible for:

- SQL query definitions
- sqlc configuration
- generated models
- generated query methods
- repository interfaces
- repository mocks

This task ends once the persistence layer has been generated successfully.

Do not implement business logic.

Do not implement gRPC services.

Do not implement OAuth.

Do not generate protobufs.

Do not implement runtime wiring.

---

# Required Reading

Read only:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md

Use these documents as the architectural source of truth.

Only inspect additional directories listed below.

Do not recursively inspect the repository.

---

# Repository Review Scope

Inspect only:

clients/db/

deposits/db/query/

deposits/db/sqlc/

deposits/db/repo/

deposits/db/doc.go

Review only enough files to understand:

- sqlc.yaml conventions
- query layout
- generated package structure
- repository interfaces
- repository implementations
- mock generation
- go:generate usage
- naming conventions

Do not inspect:

grpc/go/

protobuf/

third_party/

.github/

tests/

vendor/

coverage/

bin/

tmp/

---

# Existing Conventions

Mirror the Deposits service exactly.

Do not invent a new sqlc layout.

Use identical:

directory names

file naming

repository naming

interface naming

mock generation

package naming

code generation strategy

---

# SQL Query Responsibilities

Create query definitions for:

Clients

Platforms

Integrations

OAuth Tokens

Webhook Subscriptions

Support:

Create

Update

Delete

Soft Delete (if supported)

Find By ID

Find By Name

Find By Slug

Find By Client

Find By Platform

Find Active

List

Pagination

Existence checks

Count queries where appropriate.

---

# Query Style

Use sqlc named queries.

Prefer explicit SQL.

Avoid SELECT *.

Return only required columns.

Keep queries readable.

Follow formatting used by Deposits.

---

# SQLC Configuration

Generate sqlc configuration following the same conventions as Deposits.

Reuse:

package names

output directories

emit settings

type mappings

UUID mappings

timestamp mappings

Do not introduce custom plugins.

---

# Repository Layer

Generate repository interfaces matching the generated sqlc package.

Repository responsibilities should include:

CRUD operations

lookup operations

transaction support where required

context-aware methods

error propagation

Do not add business validation.

Repositories perform persistence only.

---

# Mock Generation

Generate mocks following the Deposits pattern.

Reuse:

go:generate

mockgen

package naming

directory layout

Do not manually edit generated mocks.

---

# Transactions

Where multiple SQL operations must be atomic:

Design repository methods that support database transactions.

Do not implement service-level transaction orchestration.

Leave orchestration to later agents.

---

# Error Handling

Follow Deposits conventions.

Wrap database errors consistently.

Avoid leaking PostgreSQL-specific details outside repositories.

---

# Deliverables

Generate:

clients/db/query/

clients/db/sqlc/

clients/db/repo/

clients/db/sqlc/mocks/

clients/db/repo/mocks/

Update:

clients/db/doc.go

Only if required to support generation.

---

# Validation

Before completing verify:

- sqlc generation succeeds

- mock generation succeeds

- go generate succeeds

- generated code compiles

- interfaces compile

- repository package builds

- queries reference existing tables

- generated models match schema

- no manual edits exist in generated files

---

# Success Criteria

The Clients service should now possess a fully generated persistence layer equivalent in quality and structure to the Deposits service.

No business logic should exist.

No protobufs should exist.

No runtime wiring should exist.

---

# Completion Rules

Before completing verify:

- Existing repository conventions have been preserved.

- Existing package naming has been preserved.

- Existing sqlc conventions have been preserved.

- Existing mock generation conventions have been preserved.

- Existing code has not been modified unnecessarily.

- Generated code has not been manually edited.

- Project builds successfully.

- No unrelated directories were modified.

If any prerequisite is missing, stop and explain why instead of producing a partial implementation.