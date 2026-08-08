# Agent 10 — Transactions Service Runtime

## Objective

Implement the runtime and process entry point for the new Transactions microservice.

This agent is responsible for taking the Transactions components already implemented by Agents 02–09 and wiring them together into a runnable service.

The runtime must:

- load Transactions configuration
- initialize logging
- initialize the PostgreSQL connection pool
- initialize the Transactions repository
- initialize Merchant, Customer, Deposit, and Payout services
- construct the Transactions gRPC server
- register the Transactions gRPC service
- register gRPC reflection if that is the established project convention
- configure required interceptors
- expose the configured server port
- handle graceful shutdown
- propagate context correctly
- preserve the project's existing runtime conventions

This agent must NOT redesign any previously implemented domain, repository, protobuf, or service architecture.

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
- docs/transactions-payouts-review.md

Additionally, inspect only the existing Deposits runtime files required to model the Transactions runtime.

The existing Deposits runtime is the primary implementation reference.

---

# Documentation Check

Before modifying code, confirm that all required documents above have been read.

At completion, perform the documentation check again.

The final review document must record that the documentation check was completed.

---

# Repository Exploration Rules

Use README.md as the repository map.

Do not perform unrestricted repository exploration.

Do NOT recursively inspect:

- third_party/
- third_party/googleapis/
- vendor/
- node_modules/
- .git/
- coverage/
- tmp/
- bin/

Do not inspect unrelated services.

Only inspect:

- the existing Deposits runtime
- Transactions runtime-related files
- Transactions configuration
- Transactions service packages
- Transactions repository packages
- relevant protobuf packages
- relevant root configuration files

Do not spend time exploring generated dependency trees.

---

# 1. Establish Existing Runtime Pattern

Locate the existing Deposits runtime.

Likely areas include:

- deposits/cmd/grpc-service/
- deposits/config/
- deposits/db/repo/
- deposits/deposits/

Inspect only the files necessary to understand:

- main()
- run()
- configuration loading
- logger initialization
- database initialization
- repository construction
- service construction
- gRPC server construction
- interceptor configuration
- service registration
- reflection
- shutdown
- error handling

The Transactions runtime must follow the same project conventions.

Do not invent a second runtime architecture.

---

# 2. Compare Existing Deposits Runtime

Document the important runtime conventions found in Deposits.

Pay attention to:

- function naming
- context handling
- logger type
- logger configuration
- error wrapping
- configuration loading
- database connection construction
- connection cleanup
- gRPC server construction
- interceptor ordering
- reflection registration
- signal handling
- graceful shutdown

Use the existing implementation as the pattern.

---

# 3. Transactions Runtime Location

Follow:

docs/repository-layout.md

for the exact location of the Transactions process.

The runtime should be structured consistently with:

deposits/cmd/grpc-service/

If the new repository layout defines:

transactions/cmd/grpc-service/

use that.

Do not create a different executable layout.

---

# 4. Main Entry Point

Create or complete:

transactions/cmd/grpc-service/main.go

The main function should remain small.

Follow the existing Deposits pattern.

The preferred responsibility boundary is:

main()

→ create root context

→ configure signal cancellation

→ configure logger

→ call run()

→ handle returned error

Do not place dependency construction directly into main if Deposits uses run() for that responsibility.

---

# 5. Signal Handling

Use the project's existing signal-handling convention.

The runtime should respond appropriately to:

- SIGINT
- SIGTERM

The exact mechanism should match Deposits.

Do not introduce a custom signal framework.

---

# 6. Context Lifecycle

The runtime must establish a root context.

Use it for:

- configuration-dependent initialization where appropriate
- database lifecycle
- server shutdown

Do not use:

context.Background()

inside service request handling.

Request contexts must come from gRPC.

---

# 7. Logger Initialization

Follow the Deposits logger setup exactly unless the project documentation specifies a deliberate Transactions-specific difference.

Preserve:

- timestamp behavior
- caller information
- log level configuration
- structured logging
- output destination
- error logging conventions

Do not replace the existing logger.

Do not introduce another logging library.

---

# 8. Configuration

Locate the Transactions configuration implementation.

Use the configuration conventions established by the existing project.

The runtime must load configuration before constructing dependent components.

At minimum determine the configuration values required for:

- database connection
- gRPC server address/port
- logging
- migration behavior if the project uses runtime migrations
- environment/runtime mode

Do not invent environment variables.

Use names already defined by the configuration implementation and `.env.example`.

---

# 9. Environment Variables

Do not hard-code:

- database URLs
- usernames
- passwords
- ports
- secrets
- provider credentials

Use environment-backed configuration.

Do not commit secrets.

Do not create a real `.env` file containing credentials.

---

# 10. Configuration Failure

If configuration cannot be loaded:

- log the error
- return the error
- stop startup

Do not continue with partially initialized configuration.

---

# 11. Database Initialization

Use the Transactions database architecture established by Agents 02 and 05.

Initialize the PostgreSQL pool before constructing repositories that depend on it.

Follow the existing Deposits implementation for:

- connection URL construction
- pgxpool creation
- error handling
- connection cleanup

Do not create a second database client.

---

# 12. Database Connection

Use the configured database URL.

Do not assume:

localhost

is correct in production.

The runtime must use the configured PostgreSQL connection information.

Do not hard-code development ports.

---

# 13. Database Cleanup

Ensure the database pool is closed when the process exits.

Follow the Deposits lifecycle pattern.

The pool must not be leaked.

---

# 14. Database Connectivity

Determine whether the existing Deposits runtime explicitly verifies connectivity.

If it does:

follow the same pattern.

If it does not:

do not introduce a new database health-check architecture.

Do not make arbitrary startup behavior changes.

---

# 15. Migrations

Do not redesign migration behavior.

Read:

docs/migration-plan.md

and:

docs/transactions-database-review.md

Determine whether Transactions migrations are:

- run during startup
- run externally
- disabled in production
- controlled by configuration

Follow the documented architecture.

Do not invent migration behavior.

---

# 16. Repository Initialization

Construct the Transactions repository using the repository abstraction already implemented.

Expected dependency direction:

database pool

→ repository

→ service

→ gRPC handler

Do not pass the database pool directly into service methods.

Do not expose SQLC directly to handlers.

---

# 17. Merchant Service Construction

Construct the Merchant service using the implementation created by Agent 06.

Use its existing constructor.

Do not duplicate Merchant initialization.

Do not create a second Merchant repository.

---

# 18. Customer Service Construction

Construct the Customer service using the implementation created by Agent 07.

Use its existing constructor.

Do not duplicate Customer initialization.

---

# 19. Deposit Service Construction

Construct the Deposit service created by Agent 08.

Use the established constructor/dependency pattern.

Do not reimplement Deposit logic.

Do not modify Deposit behavior unless compilation exposes an actual integration defect.

If such a defect exists:

document it rather than expanding scope.

---

# 20. Payout Service Construction

Construct the Payout service created by Agent 09.

Use its existing constructor/dependencies.

Do not duplicate payout repository or business logic.

---

# 21. Service Dependency Graph

The runtime should create dependencies in the correct order.

Conceptually:

Configuration
    ↓
Database Pool
    ↓
Repository
    ↓
Merchant / Customer / Deposit / Payout Services
    ↓
Transactions gRPC Server
    ↓
Service Registration
    ↓
Serve

Follow the actual project package dependencies.

Do not introduce circular dependencies.

---

# 22. gRPC Server

Create the Transactions gRPC server using the same framework and conventions used by Deposits.

Preserve the existing interceptor strategy.

If Deposits uses:

grpc.NewServer(...)

follow the same pattern.

---

# 23. Interceptors

Inspect the existing Deposits runtime for configured interceptors.

Examples may include:

- recovery
- logging
- metrics
- authentication

Only configure interceptors already established by the project.

Do not invent authentication middleware.

Do not introduce unrelated middleware.

Preserve interceptor ordering.

---

# 24. Recovery

If the existing project uses gRPC recovery:

configure it exactly as Deposits does.

The Transactions runtime must not allow a panic in a handler to terminate the process unexpectedly if the established architecture protects against it.

---

# 25. Authentication

Do not invent authentication middleware in this agent.

If authentication is already defined by the architecture:

wire it according to the established implementation.

Otherwise:

leave it unchanged.

Do not create a new authentication system merely because Transactions will eventually need one.

---

# 26. gRPC Service Registration

Register the Transactions gRPC service using the generated protobuf registration function.

Use the generated package from:

grpc/go/

or the location defined by:

docs/repository-layout.md

Do not manually modify generated registration code.

---

# 27. Multiple Transactions Capabilities

The Transactions gRPC server should expose the APIs for:

- Merchants
- Customers
- Deposits
- Payouts

only if the protobuf contract defines them on the same server/service.

Do not register unrelated services.

---

# 28. Reflection

If the Deposits runtime registers gRPC reflection:

register reflection for Transactions as well.

Use the existing convention.

Do not create custom reflection functionality.

---

# 29. Server Address

Use the configured server address/port.

Do not hard-code:

localhost

or:

127.0.0.1

unless that is explicitly the configured default.

For container deployment, ensure the runtime can bind to the configured interface.

Follow the existing service's configuration behavior.

---

# 30. Port Handling

Do not create separate configuration semantics from Deposits.

If the project convention uses:

GRPC_PORT

or an equivalent variable:

follow it.

If the Transactions configuration uses a different documented variable:

follow the documented configuration.

Do not invent multiple aliases.

---

# 31. Server Startup

Log that the Transactions gRPC server is starting.

Use the project's logger.

Do not log secrets or database credentials.

---

# 32. Serve

Start the gRPC server using the established server lifecycle.

The runtime should block while the service is running.

Do not start the server in an unnecessary goroutine if the existing runtime does not do so.

If graceful shutdown requires a serving goroutine:

follow the Deposits pattern.

---

# 33. Graceful Shutdown

Implement graceful shutdown according to the existing runtime pattern.

Expected behavior:

1. receive SIGINT/SIGTERM
2. stop accepting new requests
3. allow active requests to complete within the shutdown behavior supported by the project
4. close the gRPC server
5. close database resources
6. exit

Do not invent arbitrary shutdown timers unless the project already uses them.

---

# 34. Shutdown Ordering

Preserve resource dependency ordering.

The gRPC server should stop accepting requests before the database pool is closed.

Do not close the database pool while requests can still execute.

---

# 35. Shutdown Errors

Follow the existing project's shutdown error handling.

Do not turn a normal shutdown into an error merely because the context was cancelled.

---

# 36. Startup Failure

If any required dependency fails to initialize:

- log the error
- stop startup
- return the error

Do not start a partially configured Transactions service.

---

# 37. Dependency Construction Errors

Wrap initialization errors with useful context.

For example:

- failed to load configuration
- failed to connect database
- failed to initialize repository
- failed to construct service
- failed to start gRPC server

Follow existing error-wrapping conventions.

Do not expose credentials in wrapped errors.

---

# 38. No Business Logic

The runtime must not contain:

- payout validation
- deposit validation
- customer logic
- merchant logic
- SQL queries
- transaction state transitions

Those belong to their respective layers.

---

# 39. No Repository Logic

Do not call SQLC methods directly from:

main.go

or:

run()

Only construct the repository.

---

# 40. No Generated Code Changes

Do not modify:

- `.pb.go`
- `_grpc.pb.go`
- grpc-gateway generated files
- SQLC generated files

If generated code is missing:

document the issue.

---

# 41. Package Boundaries

Follow:

agents/project-context.md

strictly.

Do not introduce imports that violate the established package boundaries.

In particular:

- command packages may construct dependencies
- service packages contain business logic
- repository packages contain persistence behavior
- generated packages remain generated
- configuration packages handle configuration

---

# 42. Dependency Injection

Use explicit constructor injection.

Avoid:

- package-global database pools
- package-global repositories
- package-global services
- hidden initialization
- init() dependency construction

The runtime should make dependencies visible.

---

# 43. No Service Locator

Do not introduce a service locator.

Do not hide dependencies behind global registries.

---

# 44. No Framework Introduction

Do not introduce:

- Wire
- Fx
- Dig
- Uber Fx
- custom dependency injection frameworks

unless the repository already uses one.

Follow the simple construction pattern used by Deposits.

---

# 45. Docker Compatibility

The runtime must be compatible with the existing Transactions Docker strategy.

Do not modify Dockerfiles in this agent unless absolutely required to make the runtime executable and the existing Docker configuration is demonstrably incorrect.

Docker-specific work belongs to Agent 11.

If a Docker issue is discovered:

document it.

---

# 46. Render Compatibility

Do not modify Render workflows or deployment configuration.

The runtime only needs to:

- start successfully
- bind to its configured port
- connect to configured PostgreSQL
- remain alive

Deployment configuration belongs elsewhere.

---

# 47. Health Endpoints

Do not invent HTTP health endpoints in this agent.

If the architecture already defines health checks:

wire them only if required by the documented runtime design.

Otherwise leave health endpoint implementation to the appropriate scaffolding/runtime work.

---

# 48. REST Gateway

If the Transactions service uses gRPC-Gateway:

determine whether the runtime is expected to start:

- gRPC only
- gRPC + gateway
- gateway separately

Use the architecture already documented.

Do not independently create a second HTTP server.

---

# 49. Existing REST Exposure

If REST/gRPC-gateway is already part of the Transactions architecture:

wire it according to:

docs/protobuf-strategy.md

Do not manually implement REST handlers.

Do not manually edit generated gateway files.

---

# 50. Service Separation

Transactions remains its own microservice process.

Do not start:

- Clients
- Deposits as a separate legacy process
- other unrelated services

from Transactions main.go.

The Transactions service owns only its documented capabilities.

---

# 51. Local Development

The runtime should be runnable using the same basic approach as Deposits.

For example:

go run ./transactions/cmd/grpc-service

if that matches the repository structure.

Do not create a custom development launcher.

---

# 52. Compile Validation

Run focused compilation.

At minimum:

go test ./transactions/...

if that package pattern exists.

If the repository layout requires another command:

follow the actual module structure.

---

# 53. Full Test Suite

If practical, run:

go test ./...

Do not modify unrelated failing packages.

If failures are unrelated:

record them.

If failures are caused by Transactions runtime integration:

fix them if they are within this agent's scope.

---

# 54. Runtime Smoke Test

If practical, perform a local startup smoke test.

Verify:

1. configuration loads
2. database connection initializes
3. repository initializes
4. services construct
5. gRPC server starts
6. process remains running
7. shutdown signal is handled

Do not require external provider services.

---

# 55. Port Verification

If a local runtime test is performed:

verify that the configured gRPC port is actually listening.

Do not introduce a new port merely to make the test work.

---

# 56. Configuration Failure Test

If existing runtime tests support it:

verify that invalid configuration causes clean startup failure.

Do not add a large test framework.

---

# 57. Database Failure Test

If practical:

verify that an unavailable database prevents successful startup.

Do not modify production behavior solely for testing.

---

# 58. Logging Review

Review startup and shutdown logs.

Confirm they do not contain:

- passwords
- tokens
- API keys
- database credentials
- authorization headers

---

# 59. Git Status

Before finishing:

run:

git status --short

Then:

git diff --stat

Then inspect the complete relevant diff.

---

# 60. Scope Enforcement

Expected changes should generally be limited to:

- transactions/cmd/grpc-service/
- Transactions runtime configuration if directly required
- directly required runtime tests
- docs/transactions-runtime-review.md

Do not modify:

- Clients service
- Deposits implementation
- Merchant logic
- Customer logic
- Deposit logic
- Payout logic
- database migrations
- SQLC generated files
- protobuf contracts
- generated protobuf files
- provider implementations
- OAuth
- webhook handling
- deployment workflows
- third_party/

---

# 61. Existing Code Protection

Do not overwrite working implementations created by Agents 02–09.

If a component is already implemented:

use it.

Do not replace it simply because you would structure it differently.

---

# 62. Runtime Integration Problems

If a previous service cannot be wired because its constructor, interface, protobuf registration, or repository is missing:

do not rewrite that component.

Document:

- component
- missing dependency
- expected interface
- file/location
- recommended fix

Then continue with the runtime work that can safely be completed.

---

# 63. Runtime Review Document

Create:

docs/transactions-runtime-review.md

This document is mandatory.

---

# 64. Required Review Document Structure

Use exactly:

# Transactions Runtime Implementation Review

## 1. Source Documents

List every document read.

## 2. Existing Runtime Reference

Document the Deposits runtime files inspected.

## 3. Transactions Runtime Location

Document the executable/package location.

## 4. Configuration

Document:

- configuration package
- required runtime variables
- port configuration
- database configuration
- migration configuration where applicable

## 5. Logger

Document how logging is initialized.

## 6. Database Lifecycle

Document:

- pool initialization
- connection handling
- cleanup
- migration behavior if applicable

## 7. Dependency Graph

Document:

Configuration
→ Database
→ Repository
→ Services
→ gRPC Server

Include the actual Transactions dependencies.

## 8. Services Registered

Use:

| Component | Constructor | Registered/Used By |
|---|---|---|

Include:

- Merchant
- Customer
- Deposit
- Payout

where applicable.

## 9. gRPC Configuration

Document:

- server construction
- interceptors
- reflection
- service registration

## 10. REST/Gateway

Document whether the Transactions runtime exposes:

- gRPC only
- gRPC + gateway
- separate gateway process

Use the documented architecture.

## 11. Shutdown

Document:

- signals
- graceful shutdown
- resource cleanup
- shutdown ordering

## 12. Error Handling

Document startup failure behavior.

## 13. Security

Document:

- secret handling
- logging restrictions
- database credential handling
- binding behavior

## 14. Testing

Document:

- compilation
- focused tests
- full test results
- startup smoke test if performed

## 15. Files Changed

List relevant files.

## 16. Risks

Document runtime-specific risks.

## 17. Unresolved Issues

Document missing dependencies or issues belonging to later agents.

---

# 65. Documentation Check

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
- docs/transactions-payouts-review.md was read

The runtime review document must accurately reflect the implementation.

---

# Completion Checklist

Before stopping:

- [ ] Required documents were read.
- [ ] Deposits runtime was inspected as the reference implementation.
- [ ] Transactions runtime location follows repository documentation.
- [ ] Transactions configuration is loaded correctly.
- [ ] Logger is initialized according to project conventions.
- [ ] PostgreSQL pool is initialized.
- [ ] PostgreSQL resources are cleaned up.
- [ ] Migration behavior follows the documented architecture.
- [ ] Transactions repository is constructed.
- [ ] Merchant service is constructed.
- [ ] Customer service is constructed.
- [ ] Deposit service is constructed.
- [ ] Payout service is constructed.
- [ ] Dependencies are injected explicitly.
- [ ] Transactions gRPC server is created.
- [ ] Existing interceptor conventions are preserved.
- [ ] Transactions gRPC service is registered.
- [ ] Reflection is registered if required.
- [ ] Configured port/address is used.
- [ ] Server startup is logged.
- [ ] SIGINT is handled.
- [ ] SIGTERM is handled.
- [ ] Graceful shutdown is implemented.
- [ ] Shutdown ordering protects active requests.
- [ ] Database is not closed before the server stops accepting requests.
- [ ] No business logic was placed in the runtime.
- [ ] No SQLC logic was placed in the runtime.
- [ ] No generated files were manually modified.
- [ ] No protobuf contracts were modified.
- [ ] No database migrations were modified.
- [ ] No provider functionality was introduced.
- [ ] No OAuth functionality was introduced.
- [ ] No webhook functionality was introduced.
- [ ] No unrelated services were started.
- [ ] No third_party/googleapis files were inspected recursively or modified.
- [ ] Focused compilation/tests were performed.
- [ ] Full tests were run if practical.
- [ ] Runtime smoke test was performed if practical.
- [ ] Git status was reviewed.
- [ ] Git diff was reviewed.
- [ ] docs/transactions-runtime-review.md was created.
- [ ] Documentation check was completed.

---

# Final Stop Condition

STOP after completing:

1. Transactions runtime entry point
2. configuration loading
3. logger initialization
4. database lifecycle
5. repository construction
6. Merchant service construction
7. Customer service construction
8. Deposit service construction
9. Payout service construction
10. gRPC server construction
11. service registration
12. reflection/interceptors where required
13. graceful shutdown
14. focused validation
15. runtime review documentation

Do NOT proceed to:

- Docker implementation
- Makefile/scaffolding
- deployment configuration
- provider integrations
- OAuth
- webhooks
- Clients service
- broad repository refactoring

Those responsibilities belong to other agents.

If a previous agent's implementation prevents runtime wiring:

do not rewrite the previous agent's architecture.

Document the exact incompatibility in:

docs/transactions-runtime-review.md

and stop.

STOP.