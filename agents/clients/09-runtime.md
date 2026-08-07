# Agent 09 — Runtime & Service Bootstrap

## Objective

Implement the runtime for the Clients Service.

This agent is responsible for wiring together every component created by previous agents into a runnable microservice.

The runtime should closely mirror the architecture and conventions used by the Deposits service.

This agent must not introduce new business logic.

This agent must not redesign repositories.

This agent must not redesign protobuf contracts.

This agent must only compose existing components.

The Clients service should be fully executable after this task.

---

# Required Reading

Read only:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md

Then inspect only:

deposits/

clients/

protobuf/

grpc/go/

Review only enough files to understand:

- runtime layout

- configuration

- dependency injection

- grpc startup

- grpc registration

- gateway startup

- logger creation

- migration execution

- graceful shutdown

Do not recursively inspect the repository.

---

# Repository Review Scope

Inspect only:

clients/cmd/

clients/config/

deposits/cmd/

deposits/config/

protobuf/

grpc/go/

Do not inspect:

tests/

third_party/

vendor/

.github/

coverage/

tmp/

bin/

---

# Responsibilities

Wire together:

configuration

logger

database

repositories

business services

provider registry

OAuth providers

Webhook providers

gRPC services

REST gateway

health endpoints

migration runner

signal handling

Nothing else.

No business rules belong here.

---

# Runtime Layout

Create a runtime matching the Deposits service.

Implement:

clients/

├── cmd/
│   └── grpc-service/
│       └── main.go
│
├── config/
│
└── Makefile

Mirror the existing project layout wherever possible.

Do not invent new runtime patterns.

---

# Configuration

Implement configuration loading.

Support:

environment variables

.env

Docker runtime

Render runtime

future Kubernetes runtime

Configuration should remain strongly typed.

Provide safe defaults where appropriate.

Never hardcode secrets.

---

# Logger

Configure logging using the same conventions as Deposits.

Logger should be created once.

Pass through dependency injection.

Do not create multiple logger instances.

---

# Database

Create database connection pool.

Support:

connection retries

connection validation

graceful shutdown

migration execution

Repositories should receive the shared pool.

---

# Migration Runner

Integrate database migrations into startup.

Run migrations only when configuration enables them.

Startup should fail if required migrations cannot be applied.

Never silently ignore migration failures.

---

# Dependency Injection

Construct dependencies in the following order.

Configuration

↓

Logger

↓

Database

↓

Repositories

↓

Provider Registry

↓

Business Services

↓

gRPC Handlers

↓

gRPC Server

↓

REST Gateway

Every dependency should be constructed exactly once.

---

# Provider Registration

Register every provider with the Provider Registry.

Only concrete providers should be referenced here.

The remainder of the service should communicate only through interfaces.

The runtime becomes the composition root for provider implementations.

---

# gRPC Server

Create the Clients gRPC server.

Register:

ClientsService

PlatformService

IntegrationsService

Health Service (if already used elsewhere)

Mirror Deposits startup conventions.

---

# REST Gateway

Configure grpc-gateway.

Expose REST endpoints defined by protobuf annotations.

Reuse existing gateway conventions where possible.

Do not manually duplicate endpoint definitions.

---

# Health Endpoints

Implement:

startup readiness

liveness

dependency health

database connectivity

provider registry initialization

Health endpoints should expose operational state only.

Never expose secrets.

---

# Graceful Shutdown

Support graceful shutdown.

Handle:

SIGINT

SIGTERM

Stop accepting new requests.

Complete in-flight requests.

Shutdown:

gateway

gRPC server

database pool

provider resources

Exit cleanly.

---

# Error Handling

Startup failures should fail fast.

Examples:

configuration errors

database unavailable

migration failures

provider registration failures

port already in use

Never continue with a partially initialized runtime.

---

# Logging

Log:

startup

configuration loaded

database connected

migrations complete

providers registered

gRPC started

REST started

shutdown initiated

shutdown complete

Never log:

OAuth secrets

provider credentials

database passwords

access tokens

refresh tokens

---

# Deliverables

Implement:

clients/cmd/

clients/config/

clients/Makefile

Update runtime files only.

Do not modify repositories.

Do not modify protobufs.

Do not modify OAuth implementation.

Do not modify webhook implementation.

---

# Validation

Before completing verify:

- configuration loads correctly

- logger initializes once

- database connects successfully

- migrations execute

- repositories initialize

- provider registry initializes

- providers register successfully

- services initialize

- gRPC starts

- REST gateway starts

- graceful shutdown works

- project builds successfully

---

# Success Criteria

The Clients Service should now be executable as an independent microservice.

The service should mirror the runtime architecture already established by the Deposits service.

No additional implementation should be required before testing.

---

# Completion Rules

Before completing verify:

- Existing runtime conventions have been preserved.

- Existing package naming has been preserved.

- Existing dependency injection conventions have been preserved.

- Existing logging conventions have been preserved.

- Existing Makefile conventions have been preserved.

- Existing Docker conventions have been preserved.

- Existing Render compatibility has been preserved.

- No unrelated directories have been modified.

- Project builds successfully.

If a prerequisite from a previous agent is missing, stop and explain why instead of creating a partial runtime.

---

# Runtime Architecture Review

Before completing, perform a comprehensive runtime review.

Confirm:

- startup sequence is deterministic

- dependency initialization order is correct

- every dependency is initialized exactly once

- graceful shutdown cleans up every resource

- configuration remains environment-driven

- providers are registered only in the composition root

- repositories remain independent

- business services remain transport-independent

- REST and gRPC share the same business layer

- migrations execute safely

- runtime contains no business logic

- no circular dependencies exist

If improvements are discovered that do not affect previous agents, implement them.

If improvements require redesigning repositories, protobufs, OAuth, or webhook architecture, stop and document the architectural issue instead.

Produce:

clients/docs/runtime-review.md

The report should summarize:

- startup lifecycle

- dependency graph

- provider registration

- runtime architecture

- health checks

- graceful shutdown strategy

- deployment readiness

- remaining work before scaffolding and testing

Only after this review is complete should the project proceed to Agent 10.