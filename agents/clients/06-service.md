# Agent 06 — Business Service Implementation

## Objective

Implement the business service layer for the Clients Service.

This layer contains all application business logic.

It orchestrates repositories, enforces business rules, coordinates transactions, and prepares data for the transport layer.

This agent must not implement runtime wiring.

This agent must not implement OAuth provider communication.

This agent must not implement webhook processing.

This agent must not create the gRPC server.

This agent ends after all business services compile successfully.

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

Do not recursively inspect the repository.

Only inspect the directories listed below.

---

# Repository Review Scope

Inspect only:

clients/

deposits/deposits/

deposits/db/repo/

protobuf/

grpc/go/

Review only enough files to understand:

- service implementation

- constructor patterns

- dependency injection

- repository usage

- logging

- protobuf mapping

- transaction orchestration

- error propagation

Do not inspect:

cmd/

third_party/

vendor/

tests/

.github/

coverage/

bin/

tmp/

---

# Service Responsibilities

Implement the business services responsible for:

Client lifecycle

Platform lifecycle

Integration lifecycle

OAuth orchestration (business decisions only)

Webhook orchestration (business decisions only)

Provider capability management

Service methods should represent business operations.

They should not expose SQL implementation.

They should not expose transport implementation.

---

# Business Rules

The service layer is responsible for enforcing business rules.

Examples include:

A disabled platform cannot receive new integrations.

A client cannot install the same provider twice.

Inactive clients cannot create integrations.

OAuth installation must validate platform capabilities.

Webhook registration requires an active integration.

Provider capability checks must occur before persistence.

Repository methods should assume valid input.

All validation belongs here.

---

# Transaction Coordination

Coordinate repository operations where multiple persistence actions must succeed together.

Examples include:

Create Client

Create Integration

Store OAuth metadata

Register webhook subscription

Rollback the transaction if any step fails.

Do not expose transaction logic outside the service layer.

---

# Repository Usage

Use repository interfaces exclusively.

Never execute SQL directly.

Never depend on sqlc.

Never depend on PostgreSQL packages.

Repositories remain the only persistence abstraction.

---

# Mapping

Map between:

protobuf messages

repository models

domain objects

Avoid leaking repository structures outside the service layer.

Avoid exposing database implementation details.

Centralize mapping logic wherever practical.

---

# Dependency Injection

Construct services using dependency injection.

Dependencies may include:

repositories

logger

configuration interfaces (where required)

clock interfaces (if already used)

Do not construct dependencies internally.

---

# Error Handling

Translate repository errors into business errors.

Do not expose:

SQL errors

constraint names

database implementation

provider SDK errors

Create meaningful application-level errors.

Errors should help API consumers understand what failed.

---

# Logging

Mirror the logging conventions used by the Deposits service.

Log:

business events

state transitions

installation progress

provider selection

Avoid logging:

OAuth tokens

refresh tokens

client secrets

webhook secrets

provider credentials

encrypted values

Personally identifiable information unless already permitted by project conventions.

---

# Idempotency

Ensure service methods are safe to retry where appropriate.

Examples:

Create operations should detect duplicates.

Webhook registration should avoid duplicate subscriptions.

OAuth completion should not produce duplicate integrations.

Business operations should remain deterministic.

---

# Provider Independence

The service layer must remain provider agnostic.

Never implement:

if provider == HighLevel

inside core business services.

Provider-specific implementations belong to later agents.

The service layer should dispatch through interfaces.

---

# Interfaces

Define clean service interfaces.

Avoid large interfaces.

Separate responsibilities by aggregate.

ClientsService

PlatformsService

IntegrationsService

Future provider services

Interfaces should be mockable.

---

# Deliverables

Implement:

clients/service/

(or the equivalent package defined by repository conventions)

Do not modify:

runtime

protobuf definitions

database schema

repositories

---

# Validation

Before completing verify:

- business services compile

- repository interfaces are respected

- dependency injection works

- transaction coordination works

- business validation executes correctly

- mapping functions compile

- services remain provider agnostic

- no SQL exists in the service layer

- no transport logic exists in the service layer

---

# Success Criteria

The Clients Service should now possess a complete business layer capable of coordinating repositories and enforcing application rules.

The service layer should be independent of transport and runtime concerns.

It should be possible to unit test the business layer without running a gRPC server.

---

# Completion Rules

Before completing verify:

- Existing service conventions have been preserved.

- Existing package naming has been preserved.

- Existing dependency injection conventions have been preserved.

- Existing logging conventions have been preserved.

- Existing repository abstractions have been respected.

- Existing protobuf contracts have not been modified unnecessarily.

- No unrelated directories have been modified.

- The project builds successfully.

If a prerequisite from a previous agent is missing, stop and explain why instead of creating a partial implementation.

---

# Service Architecture Review

Before completing, perform a complete review of the service layer.

Confirm:

- business rules exist only in the service layer

- repositories remain persistence-only

- transport concerns remain absent

- services remain provider agnostic

- interfaces remain cohesive

- transaction boundaries are appropriate

- mapping responsibilities are well separated

- dependency injection remains clean

- no circular dependencies exist

- services are easily unit testable

If improvements are discovered that do not affect downstream agents, implement them.

If improvements require redesigning the database, protobuf contracts, or repositories, stop and document the architectural issue instead.

Produce:

clients/docs/service-review.md

The report should summarize:

- implemented business services

- business rule ownership

- transaction strategy

- dependency graph

- service interfaces

- provider abstraction strategy

- remaining work before OAuth implementation

Only after this review is complete should the project proceed to Agent 07.