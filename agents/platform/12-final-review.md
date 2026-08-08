# Agent 12 — Final Platform Review

## Objective

Perform the final platform-wide review of the RVPay implementation after completion of the Clients, Transactions, and Platform agent sequences.

This is the FINAL platform agent.

The purpose of this agent is to determine whether the implementation produced by the previous agents is:

* structurally consistent
* internally coherent
* compatible with the documented architecture
* buildable
* testable
* deployable
* observable
* secure
* consistent with the existing project conventions
* ready for the next stage of development or deployment

This agent is NOT a new implementation phase.

Do not redesign the system.

Do not introduce new architecture.

Do not add speculative features.

Do not rewrite working code merely because another design might be preferable.

The existing architecture and project documentation are authoritative.

Fix only concrete issues discovered during the review.

---

# Required Reading

Read only:

* README.md
* agents/project-context.md
* docs/domain-model.md
* docs/repository-layout.md
* docs/protobuf-strategy.md
* docs/migration-plan.md
* docs/platform-repository-audit.md
* docs/platform-protobuf-generation-review.md
* docs/platform-http-gateway-review.md
* docs/platform-common-packages-review.md
* docs/platform-ci-cd-review.md
* docs/platform-docker-review.md
* docs/platform-render-review.md
* docs/platform-documentation-review.md
* docs/platform-observability-review.md
* docs/platform-security-review.md
* docs/platform-performance-review.md

Also read the final review documents produced by:

* Clients agents
* Transactions agents

Use the exact filenames documented by those agents.

If a final review document exists for either service:

read it.

If one does not exist:

record the absence.

Do not attempt to reconstruct an entire service history.

---

# Documentation Check

Before beginning:

verify that the required documentation exists.

If a required foundation document is missing:

STOP and report it.

Do not recreate missing foundation documents.

The final review may identify missing documentation, but should not silently invent it.

At the end of the task:

perform the documentation check again.

Create:

docs/platform-final-review.md

---

# Repository Exploration Rules

## IMPORTANT

Do NOT perform an unrestricted repository-wide recursive search.

Use:

README.md

as the primary map.

Use:

docs/repository-layout.md

as the authoritative repository structure.

Use:

agents/project-context.md

as the authority for coding conventions, packages, naming, and implementation style.

Use the platform review documents to understand work already performed.

Use the Clients and Transactions review documents to understand service-level implementation.

Only inspect source files that are relevant to a specific verification question.

---

# Do NOT Explore Deep Folders

Do NOT recursively inspect:

* .git/
* vendor/
* node_modules/
* coverage/
* tmp/
* bin/
* third_party/
* third_party/googleapis/

Especially:

DO NOT recursively inspect:

third_party/googleapis/

Do not spend time reviewing generated protobuf internals.

Do not inspect every generated file simply because it exists.

---

# Existing Code Is Authoritative

The current repository contains an older implementation that has been extended.

Do NOT assume that all code was created by the new agents.

Do not overwrite existing working code.

Do not "clean up" unrelated legacy code.

Do not perform broad refactoring.

---

# 1. Final Review Philosophy

The review must answer:

1. Does the repository match the documented architecture?
2. Do Clients and Transactions services fit together?
3. Does the platform layer support both services?
4. Do protobuf contracts match generated code?
5. Do database boundaries remain correct?
6. Does runtime wiring work?
7. Can the services build?
8. Can the services test?
9. Can the services start?
10. Can the deployment configuration run?
11. Are security controls intact?
12. Is observability present?
13. Are there obvious performance or reliability blockers?
14. Are there unresolved implementation gaps?

---

# 2. Do Not Treat Review as an Invitation to Rewrite

The final review must NOT become:

"Here is how I would redesign RVPay."

Instead:

"Here is whether the implemented RVPay system satisfies its documented design."

---

# 3. Establish the Review Baseline

Before inspecting implementation:

read:

* README.md
* docs/domain-model.md
* docs/repository-layout.md
* docs/protobuf-strategy.md
* docs/migration-plan.md

Understand:

* service boundaries
* domain entities
* database ownership
* protobuf contracts
* migration strategy
* runtime layout

---

# 4. Read Project Context

Read:

agents/project-context.md

Pay particular attention to:

* package naming
* error handling
* logging
* configuration
* database access
* generated code
* testing
* service structure
* naming conventions

Do not introduce conventions that conflict with this document.

---

# 5. Review Foundation Documents

Compare the actual repository against:

docs/domain-model.md

docs/repository-layout.md

docs/protobuf-strategy.md

docs/migration-plan.md

Document deviations.

Do not automatically classify every deviation as a bug.

Some deviations may be justified by the implementation.

---

# 6. Review Agent Outputs

Read the final review documentation generated by:

Clients

Transactions

Platform

agents.

Identify:

* unresolved issues
* follow-ups
* blocked tasks
* warnings
* known compromises

Do not assume previous agents were correct.

Verify important claims against the repository.

---

# 7. Clients Service Review

Inspect the Clients service at a high level.

Verify:

* service exists in expected location
* configuration exists
* database layer exists
* migrations exist
* SQL exists
* SQLC generation exists
* repositories exist
* service implementation exists
* protobuf contracts exist
* runtime wiring exists
* OAuth functionality exists
* webhook functionality exists
* tests exist
* scaffolding exists

Do not recursively inspect every file.

---

# 8. Transactions Service Review

Inspect the Transactions service at a high level.

Verify:

* service exists
* configuration exists
* database layer exists
* migrations exist
* SQL exists
* SQLC generation exists
* repositories exist
* merchant handling exists
* customer handling exists
* deposits exist
* payouts exist
* runtime exists
* tests exist
* scaffolding exists

---

# 9. Platform Review

Verify:

* protobuf generation
* HTTP gateway
* common packages
* CI/CD
* Docker
* Render configuration
* documentation
* observability
* security
* performance

---

# 10. Repository Layout

Compare actual directories against:

docs/repository-layout.md

Check for:

* unexpected service locations
* duplicate packages
* misplaced files
* inconsistent directory names

Do not reorganize directories automatically.

---

# 11. Service Boundaries

Verify that service responsibilities remain separated.

Clients should own client/integration concerns.

Transactions should own transaction concerns.

Platform should provide shared infrastructure.

Do not move code merely to make the boundaries aesthetically cleaner.

---

# 12. Database Ownership

Verify that database access follows the documented ownership model.

Check that one service is not directly reaching into another service's database merely for convenience.

If cross-service communication is required:

verify that the documented mechanism is used.

---

# 13. Cross-Service Dependencies

Look for suspicious imports between services.

For example:

clients importing internal transaction implementation packages.

transactions importing internal clients repository packages.

Such coupling should be treated as a finding unless explicitly documented.

---

# 14. Shared Packages

Verify that shared functionality belongs in the appropriate common package.

Look for duplicated implementations of:

* logging
* configuration
* HTTP handling
* errors
* middleware
* observability

Do not automatically merge packages.

---

# 15. Circular Dependencies

Check for Go package cycles.

Run the appropriate Go build/test command to detect them.

Do not restructure the entire repository unless required to resolve an actual cycle.

---

# 16. Configuration

Review configuration across:

* Clients
* Transactions
* Platform/runtime

Verify:

* environment variables are named consistently
* required variables are validated
* secrets are not committed
* defaults are intentional
* local development remains understandable

---

# 17. Environment Files

Verify:

.env.example

contains required variables without real secrets.

Do not create a real .env file containing credentials.

---

# 18. Secrets

Search only relevant configuration files for:

* API keys
* client secrets
* database passwords
* signing secrets
* OAuth secrets
* tokens

Do not dump secrets into the review document.

If an exposed secret is found:

report its location without reproducing the secret.

---

# 19. Database Configuration

Verify database configuration is compatible with deployment.

Pay particular attention to:

* host
* port
* database name
* username
* password
* SSL settings
* connection URL

Do not hard-code local development addresses into production configuration.

---

# 20. Render Compatibility

Review:

docs/platform-render-review.md

Verify:

* internal database URLs are used where appropriate
* services do not use localhost to reach managed PostgreSQL
* ports are configured correctly
* environment variables are documented
* Docker startup commands are consistent
* health checks are appropriate

---

# 21. Localhost Rule

Remember:

localhost

inside a container refers to that container.

It does NOT refer to:

* another container
* a Render PostgreSQL service
* another Render service

Flag production configurations that incorrectly depend on localhost.

---

# 22. Docker Review

Review:

docs/platform-docker-review.md

Verify:

* Dockerfiles exist where required
* build context is correct
* working directories are correct
* binaries are created in expected locations
* entrypoints are executable
* runtime images are sensible
* environment configuration is not baked with secrets

---

# 23. Docker Build

Run the documented Docker build command where practical.

Do not invent a different container architecture.

---

# 24. Docker Runtime

Where practical:

verify that the built container can start.

Do not require external production credentials merely to perform local structural verification.

---

# 25. Protobuf Source

Verify protobuf source files.

Confirm:

* packages are correct
* service definitions exist
* request/response messages are sensible
* imports are correct

Do not inspect third-party googleapis internals.

---

# 26. Generated Protobuf

Verify generated files exist where expected.

Do not manually edit generated protobuf code.

If regeneration is necessary:

use the documented Makefile workflow.

---

# 27. Protobuf Consistency

Verify:

source protobuf

matches:

generated Go protobuf

and:

generated gRPC interfaces.

If stale generated code is found:

regenerate it using the established workflow.

---

# 28. SQLC

Verify:

SQL source

matches:

generated SQLC code.

If generated code is stale:

regenerate it using the project's established process.

Do not manually edit generated SQLC files.

---

# 29. Migrations

Verify migrations:

* follow naming conventions
* have up/down behavior where required
* do not modify old applied migrations
* match the current domain model
* support the SQL queries

---

# 30. Migration Ordering

Ensure migrations execute in a sensible order.

Look for:

* duplicate numbers
* conflicting names
* missing dependencies
* references to tables not yet created

---

# 31. Database Schema

Verify the implemented schema supports:

* clients/integrations
* transactions
* merchants
* customers
* deposits
* payouts
* webhook state
* OAuth state

Only verify entities actually defined in the TDD/foundation documents.

---

# 32. Repository Layer

Verify repositories:

* use the expected SQLC layer
* respect service ownership
* handle errors consistently
* use context
* do not leak database resources

---

# 33. Service Layer

Verify services:

* contain business logic
* do not directly contain low-level SQL where repositories should be used
* validate requests
* return appropriate errors
* respect context cancellation

---

# 34. Handler Layer

Verify handlers:

* translate transport requests into service operations
* validate transport-level input
* return appropriate gRPC/HTTP errors
* do not contain large amounts of business logic

---

# 35. Runtime Wiring

This is a critical review area.

Verify the central service entry points:

* load configuration
* initialize logging
* initialize database
* initialize repositories
* initialize services
* initialize handlers
* register gRPC services
* configure gateway where required
* configure reflection where appropriate
* start the server
* handle shutdown

---

# 36. Dependency Wiring

Trace the dependency chain:

configuration

→ database

→ repository

→ service

→ handler

→ gRPC server

→ HTTP gateway

where applicable.

Verify there are no missing dependencies.

---

# 37. Startup Failure

Verify startup fails clearly when required dependencies cannot initialize.

Examples:

* invalid configuration
* database unavailable
* invalid migration state
* required provider configuration missing

Do not allow services to appear healthy when critical initialization failed.

---

# 38. Graceful Shutdown

Verify the services respond appropriately to:

* SIGINT
* SIGTERM

Ensure:

* servers stop
* database pools close
* background workers stop
* resources are released

where applicable.

---

# 39. Context

Verify request context is propagated through:

handler

→ service

→ repository

→ database/provider calls.

Do not introduce context.Background() to bypass propagation.

---

# 40. OAuth Review

Verify Clients OAuth implementation at a high level.

Check:

* redirect URI handling
* authorization code handling
* token exchange
* state validation where applicable
* client credentials
* token storage
* callback endpoint
* error handling

Do not expose OAuth secrets in documentation.

---

# 41. Webhook Review

Verify webhook handling:

* endpoint exists
* request validation exists
* provider authenticity/security mechanism is respected
* events are persisted/processed according to the documented architecture
* duplicate events are handled safely
* failures are observable

---

# 42. Idempotency

Verify idempotency for operations where duplicates are possible.

Especially:

* deposits
* payouts
* webhooks
* OAuth installation events

Do not weaken idempotency for performance.

---

# 43. Transaction Safety

Review financial operations.

Verify:

* database transactions are used where required
* state transitions are consistent
* duplicate processing is prevented
* errors do not produce partial invalid state

---

# 44. Provider Integration

Verify provider abstraction and implementation.

Check:

* provider interface exists where required
* concrete provider implementations follow it
* configuration is isolated
* external calls use appropriate clients
* provider errors are handled consistently

Do not introduce new providers during this review.

---

# 45. Provider Credentials

Verify provider credentials come from configuration/secrets.

Do not hard-code credentials.

---

# 46. HTTP Gateway

Review gateway behavior.

Verify:

* routes map to intended gRPC methods
* HTTP status behavior is sensible
* errors are translated correctly
* authentication requirements are preserved
* request/response serialization is consistent

---

# 47. Health Endpoints

Verify appropriate:

* liveness
* readiness

behavior.

Do not make health checks unnecessarily expensive.

---

# 48. Observability

Review:

docs/platform-observability-review.md

Verify:

* structured logging exists
* errors are logged appropriately
* request timing is available where expected
* metrics exist where documented
* tracing exists where documented
* sensitive data is not logged

---

# 49. Security

Review:

docs/platform-security-review.md

Verify:

* authentication
* authorization
* secret handling
* TLS assumptions
* input validation
* webhook verification
* OAuth security
* sensitive logging
* dependency hygiene

Do not weaken security to simplify deployment.

---

# 50. Performance

Review:

docs/platform-performance-review.md

Verify that no obvious performance blockers remain.

Pay attention to:

* N+1 queries
* database connection exhaustion
* repeated client construction
* unnecessary provider requests
* unbounded goroutines
* excessive payload processing

Do not perform a new broad performance optimization project.

---

# 51. CI/CD

Review:

docs/platform-ci-cd-review.md

Verify:

* build works
* tests execute
* generated code requirements are understood
* Docker build works where required
* deployment configuration is coherent

---

# 52. Go Version

Verify the repository's Go version is consistent across:

* go.mod
* Docker
* CI
* documentation

Do not arbitrarily upgrade Go.

If versions differ:

determine whether the difference is intentional.

---

# 53. Dependency Consistency

Check:

go.mod

and:

go.sum

for obvious inconsistencies.

Do not perform broad dependency upgrades.

---

# 54. Build

Run the documented repository build process.

The project must compile.

If it does not:

identify the concrete failure.

Fix only issues necessary to restore the documented build.

---

# 55. Tests

Run the repository's standard test command.

Tests must pass unless:

* an external dependency prevents execution
* the repository contains an explicitly documented unavailable dependency

Do not delete or weaken tests.

---

# 56. Test Failures

For every failure:

determine whether it is:

* implementation failure
* stale generated code
* environment problem
* configuration problem
* pre-existing unrelated issue

Do not blindly change tests.

---

# 57. Race Detection

Where practical:

run race-enabled tests.

Especially if previous agents introduced concurrency.

---

# 58. Static Analysis

Run the project's documented static analysis tools.

Do not introduce unrelated tools.

---

# 59. Formatting

Run:

gofmt

or the project's established formatting workflow.

Do not reformat unrelated files unnecessarily.

---

# 60. Generated Code Verification

If protobuf or SQLC generated code is stale:

regenerate it.

Then inspect:

git diff

carefully.

Generated changes must be explainable.

---

# 61. Git Diff

At this stage:

run:

git status --short

then:

git diff --stat

then:

git diff

Review every changed file relevant to the final implementation.

---

# 62. Unexpected Changes

If unrelated modifications are already present:

DO NOT:

* reset them
* checkout them
* overwrite them
* delete them

Document them if necessary.

---

# 63. No Destructive Git Commands

Do NOT run:

git reset --hard

git clean -fd

git checkout -- .

or equivalent destructive commands.

---

# 64. README Verification

The root:

README.md

must remain an accurate map of the repository.

If implementation changes make the README materially incorrect:

update only the affected sections.

Do not rewrite the entire README.

---

# 65. Documentation Consistency

Verify that documentation does not describe:

* nonexistent services
* old paths
* old commands
* obsolete environment variables
* obsolete deployment architecture

---

# 66. Environment Documentation

Verify .env.example reflects current required configuration.

Do not include secrets.

---

# 67. Makefiles

Verify service/repository Makefiles:

* reference valid directories
* use valid commands
* do not reference removed services
* generate protobuf/SQLC correctly
* run tests correctly

Do not rewrite Makefiles unnecessarily.

---

# 68. Generated Artifacts

Verify generated artifacts are either:

* committed according to project convention

or:

* reproducibly generated by documented commands.

Do not introduce a new generated-artifact policy.

---

# 69. API Compatibility

Verify existing API contracts were not unintentionally broken.

Pay particular attention to the older RVPay functionality described in README.md.

---

# 70. Backward Compatibility

Where the new implementation extends existing behavior:

do not remove working functionality without documented migration requirements.

---

# 71. Migration Safety

Verify the migration path from the existing system to the new implementation.

Use:

docs/migration-plan.md

as the authority.

Do not invent an alternate migration strategy.

---

# 72. Existing Deposits Functionality

The existing deposits implementation must not be accidentally broken by the Transactions service migration.

Verify important existing behavior remains represented.

---

# 73. Existing Repository Compatibility

If old repository code remains:

do not delete it merely because the new architecture could replace it.

Only remove code if the migration plan explicitly requires removal and the new implementation fully replaces its responsibilities.

---

# 74. Clients/Transactions Interaction

Verify the two services communicate through their intended boundaries.

Do not create direct database coupling.

---

# 75. Platform/Service Interaction

Verify platform packages are genuinely shared infrastructure.

Do not allow platform packages to become business-domain packages.

---

# 76. Error Handling

Verify errors are:

* wrapped appropriately
* logged at the correct boundary
* converted correctly for gRPC/HTTP
* not leaking secrets

---

# 77. Logging

Verify logs provide enough information to diagnose:

* startup failure
* database failure
* provider failure
* OAuth failure
* webhook failure
* transaction failure

without logging secrets.

---

# 78. Database Errors

Database errors should not expose raw credentials, connection strings, or other secrets.

---

# 79. Provider Errors

Provider errors should be translated appropriately.

Do not expose sensitive provider configuration.

---

# 80. OAuth Errors

OAuth errors should not leak:

* client secrets
* access tokens
* authorization codes
* refresh tokens

---

# 81. Webhook Security

Webhook payload handling must not trust arbitrary incoming requests without the documented verification mechanism.

---

# 82. Authentication

Verify authentication is applied at the correct boundary.

Do not accidentally make protected transaction endpoints public.

---

# 83. Authorization

Verify authorization decisions are not bypassed by:

* HTTP gateway routes
* internal gRPC methods
* alternate endpoints

---

# 84. CORS

If CORS is configured:

verify it is appropriate for the intended clients.

Do not use wildcard permissive CORS for authenticated sensitive APIs without justification.

---

# 85. Request Validation

Verify malformed input is rejected before dangerous operations.

Especially:

* transaction amounts
* IDs
* provider identifiers
* OAuth parameters
* webhook data

---

# 86. Financial Data

Do not weaken validation around financial values.

Check:

* amount
* currency
* status
* transaction identifiers
* merchant identifiers

---

# 87. Currency

Verify currency handling follows the domain model.

Do not introduce floating-point financial calculations where the architecture expects safe monetary representation.

---

# 88. Status Transitions

Verify transaction status transitions are consistent.

Do not permit arbitrary state transitions merely for convenience.

---

# 89. Database Constraints

Where the domain requires uniqueness or integrity:

verify database constraints exist.

Examples may include:

* external transaction identifiers
* provider references
* OAuth installation identifiers
* webhook event identifiers

Only verify constraints documented by the domain model.

---

# 90. Idempotent Webhooks

Verify duplicate webhook delivery does not create duplicate business records.

---

# 91. Idempotent Transactions

Verify repeated transaction requests do not unintentionally create duplicate financial operations where idempotency is required.

---

# 92. Provider Failure

Verify provider failures produce recoverable states.

Do not leave transactions permanently ambiguous where the documented architecture requires a known failure state.

---

# 93. Deployment Startup

Verify each service has a coherent production startup path.

At minimum:

configuration

→ database

→ repositories

→ services

→ handlers

→ server

---

# 94. Ports

Verify service ports are:

* configured consistently
* documented
* compatible with Docker/Render

Do not hard-code conflicting ports.

---

# 95. Health Checks and Deployment

Verify Render/Docker health checks point to an actual endpoint.

Do not configure a health check for a nonexistent route.

---

# 96. Process Model

Verify each service process runs the intended binary.

Do not run development commands in production containers.

---

# 97. Container Environment

Verify production configuration is injected through environment variables/secrets.

Do not copy .env containing secrets into production images.

---

# 98. Final Build

Run the final build.

Use the project's documented commands.

---

# 99. Final Tests

Run:

* unit tests
* service tests
* repository tests
* relevant integration tests
* race tests where appropriate

Use existing project conventions.

---

# 100. Final Verification

Where practical verify:

* protobuf generation
* SQLC generation
* database migration validation
* Docker build
* application startup

---

# 101. Do Not Require External Credentials

Do not fail the review merely because:

* real OAuth credentials are unavailable
* real provider credentials are unavailable
* production database credentials are unavailable

Instead:

verify configuration structure and local/test behavior.

---

# 102. Do Not Contact External Systems

Do not make destructive requests to:

* payment providers
* production databases
* OAuth production systems
* production webhooks

---

# 103. Production Verification

Production configuration may be reviewed structurally.

Do not perform live financial transactions.

Do not trigger real deposits or payouts.

---

# 104. Fix Policy

Fix issues only when:

1. the issue is concrete,
2. the expected behavior is documented,
3. the fix is localized,
4. the fix does not redesign architecture,
5. the fix can be verified.

---

# 105. When NOT to Fix

Do not fix:

* stylistic preferences
* speculative improvements
* unrelated legacy code
* theoretical performance concerns
* architectural preferences not present in the TDD
* future features

Document them instead.

---

# 106. Priority

Prioritize:

1. build failures
2. test failures
3. broken runtime wiring
4. security defects
5. data integrity problems
6. migration failures
7. deployment blockers
8. broken protobuf/SQLC generation
9. major observability failures
10. significant performance/reliability problems
11. minor consistency issues

---

# 107. Final Review Findings

Use:

| ID | Severity | Area | Finding | Evidence | Action |
| -- | -------- | ---- | ------- | -------- | ------ |

Severity:

* BLOCKER
* HIGH
* MEDIUM
* LOW
* INFO

---

# 108. BLOCKER

Use BLOCKER for issues that prevent:

* compilation
* service startup
* database initialization
* required migrations
* critical service communication
* secure production deployment

---

# 109. HIGH

Use HIGH for serious problems that do not necessarily prevent startup but materially threaten correctness, security, reliability, or deployment.

---

# 110. MEDIUM

Use MEDIUM for meaningful defects that should be addressed but do not block operation.

---

# 111. LOW

Use LOW for minor issues.

---

# 112. INFO

Use INFO for observations or future recommendations.

---

# 113. Final Review Document

Create:

docs/platform-final-review.md

Use exactly this structure:

# RVPay Platform Final Review

## 1. Review Objective

Describe the purpose of the final review.

## 2. Required Documentation

List every document read.

## 3. Repository Scope

Document which areas were inspected.

Explicitly state that deep/generated folders were not recursively explored.

## 4. Architecture Verification

Document whether implementation matches:

* domain model
* repository layout
* protobuf strategy
* migration plan

## 5. Clients Service

Document:

* database
* SQLC
* repositories
* service
* OAuth
* webhooks
* runtime
* tests

## 6. Transactions Service

Document:

* database
* SQLC
* repositories
* merchants
* customers
* deposits
* payouts
* runtime
* tests

## 7. Platform

Document:

* protobuf
* gateway
* common packages
* CI/CD
* Docker
* Render
* documentation
* observability
* security
* performance

## 8. Cross-Service Integration

Document:

* service boundaries
* dependencies
* protobuf contracts
* communication
* database ownership

## 9. Configuration

Document:

* environment variables
* secrets
* database configuration
* provider configuration

## 10. Runtime

Document:

* startup
* dependency wiring
* health
* shutdown
* ports

## 11. Database

Document:

* migrations
* SQL
* SQLC
* repositories
* constraints
* indexes
* transaction handling

## 12. API

Document:

* gRPC
* HTTP gateway
* request validation
* error handling

## 13. Security

Document:

* authentication
* authorization
* OAuth
* webhook security
* secret handling
* sensitive logging

## 14. Observability

Document:

* logs
* metrics
* traces
* health checks

## 15. Performance

Summarize findings from:

docs/platform-performance-review.md

## 16. CI/CD

Document build/test/deployment verification.

## 17. Docker

Document container verification.

## 18. Render

Document deployment configuration verification.

## 19. Tests

Document:

* test commands
* results
* race testing
* integration testing

## 20. Build Verification

Document actual build results.

Do not invent results.

## 21. Generated Code Verification

Document:

* protobuf generation
* SQLC generation

## 22. Findings

Use:

| ID | Severity | Area | Finding | Evidence | Action |
| -- | -------- | ---- | ------- | -------- | ------ |

## 23. Fixes Applied

List only changes made by this agent.

## 24. Remaining Issues

List unresolved issues.

## 25. Future Recommendations

Only include recommendations outside the scope of this implementation.

## 26. Documentation Changes

List documentation changes.

## 27. Final Documentation Check

Record whether all required documents exist.

## 28. Final Status

Use exactly one:

PASS

PASS WITH FOLLOW-UP

BLOCKED

---

# 114. Final Status Rules

Use:

PASS

only when there are no unresolved BLOCKER or HIGH findings and the system is sufficiently verified.

Use:

PASS WITH FOLLOW-UP

when the system is functional but MEDIUM/LOW/INFO items remain.

Use:

BLOCKED

when a BLOCKER remains or required verification cannot be completed because of a fundamental issue.

---

# 115. No False PASS

Do not mark:

PASS

if:

* the repository does not build
* required tests fail due to implementation defects
* runtime wiring is broken
* critical migrations fail
* security controls are broken
* required services cannot start

---

# 116. External Dependency Limitations

If verification is blocked by an external dependency:

document:

* what was attempted
* what could not be tested
* why
* what was verified instead

Do not automatically mark the whole project BLOCKED if the limitation does not prevent meaningful local verification.

---

# 117. Final README Check

Confirm that README.md still accurately describes:

* repository layout
* services
* setup
* environment configuration
* build commands
* test commands
* deployment expectations

Update only incorrect sections.

---

# 118. Final Repository Layout Check

Compare:

actual repository

against:

docs/repository-layout.md

If differences exist:

determine whether they are:

* intentional
* documented
* accidental

Fix accidental differences only when safe.

---

# 119. Final Project Context Check

Before making any code change:

consult:

agents/project-context.md

Ensure the proposed fix follows existing conventions.

---

# 120. Final Generated Code Rule

Never manually edit:

* generated protobuf Go files
* generated gRPC Go files
* generated SQLC files

Regenerate them from source when necessary.

---

# 121. Final Migration Rule

Never rewrite an existing migration merely to make the final review pass.

Create a new migration if a schema correction is required.

---

# 122. Final Git Review

Before completion:

run:

git status --short

then:

git diff --stat

then:

git diff

Review all modifications.

---

# 123. No Destructive Cleanup

Do not remove:

* unrelated untracked files
* existing modified code
* developer work
* legacy files

unless explicitly required by the documented migration plan.

---

# 124. Final Documentation Check

Verify the existence of:

* README.md
* agents/project-context.md
* docs/domain-model.md
* docs/repository-layout.md
* docs/protobuf-strategy.md
* docs/migration-plan.md
* docs/platform-repository-audit.md
* docs/platform-protobuf-generation-review.md
* docs/platform-http-gateway-review.md
* docs/platform-common-packages-review.md
* docs/platform-ci-cd-review.md
* docs/platform-docker-review.md
* docs/platform-render-review.md
* docs/platform-documentation-review.md
* docs/platform-observability-review.md
* docs/platform-security-review.md
* docs/platform-performance-review.md
* docs/platform-final-review.md

Also verify the final review documents produced by:

* Clients
* Transactions

where their filenames are documented.

---

# 125. Completion Checklist

Before stopping:

* [ ] README.md was read.
* [ ] agents/project-context.md was read.
* [ ] All foundation documents were read.
* [ ] All relevant platform review documents were read.
* [ ] Clients final review was read.
* [ ] Transactions final review was read.
* [ ] Repository exploration was restricted.
* [ ] Deep folders were not recursively inspected.
* [ ] third_party/googleapis was not unnecessarily explored.
* [ ] Domain model was verified.
* [ ] Repository layout was verified.
* [ ] Clients service was verified.
* [ ] Transactions service was verified.
* [ ] Platform layer was verified.
* [ ] Service boundaries were verified.
* [ ] Database ownership was verified.
* [ ] Configuration was verified.
* [ ] Secrets were reviewed.
* [ ] Protobuf source was reviewed.
* [ ] Generated protobuf was verified.
* [ ] SQL was reviewed.
* [ ] SQLC generation was verified.
* [ ] Migrations were reviewed.
* [ ] Repository layers were reviewed.
* [ ] Service layers were reviewed.
* [ ] Runtime wiring was verified.
* [ ] HTTP gateway was reviewed.
* [ ] OAuth was reviewed.
* [ ] Webhooks were reviewed.
* [ ] Transaction safety was reviewed.
* [ ] Idempotency was reviewed.
* [ ] Authentication was reviewed.
* [ ] Authorization was reviewed.
* [ ] Observability was reviewed.
* [ ] Performance findings were reviewed.
* [ ] CI/CD was reviewed.
* [ ] Docker was reviewed.
* [ ] Render deployment configuration was reviewed.
* [ ] Build was executed.
* [ ] Tests were executed.
* [ ] Race testing was performed where appropriate.
* [ ] Formatting was verified.
* [ ] Generated code was not manually edited.
* [ ] Existing migrations were not rewritten.
* [ ] README.md was checked.
* [ ] Documentation was checked.
* [ ] git status was reviewed.
* [ ] git diff was reviewed.
* [ ] docs/platform-final-review.md was created.
* [ ] Final documentation check was completed.
* [ ] Final status was assigned honestly.

---

# Final Stop Condition

STOP after:

1. reading all required documentation,
2. reviewing the Clients implementation,
3. reviewing the Transactions implementation,
4. reviewing the Platform implementation,
5. verifying service boundaries,
6. verifying database ownership,
7. verifying protobuf consistency,
8. verifying SQLC consistency,
9. verifying migrations,
10. verifying runtime wiring,
11. verifying configuration,
12. verifying OAuth,
13. verifying webhooks,
14. verifying transaction safety,
15. verifying security,
16. verifying observability,
17. verifying performance findings,
18. verifying Docker,
19. verifying Render configuration,
20. running the appropriate build,
21. running the appropriate tests,
22. reviewing generated code state,
23. reviewing git status,
24. reviewing git diff,
25. fixing only concrete blockers or defects,
26. creating docs/platform-final-review.md,
27. completing the documentation check,
28. assigning the final status.

Do NOT proceed to:

* another architecture redesign
* new services
* new providers
* new infrastructure
* new caching systems
* new queues
* database replacement
* protobuf redesign
* gRPC replacement
* HTTP framework replacement
* broad dependency upgrades
* unrelated refactoring
* speculative optimization
* new business features

This is the FINAL PLATFORM REVIEW.

STOP.
