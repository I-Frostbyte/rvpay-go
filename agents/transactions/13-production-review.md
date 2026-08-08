# Agent 13 — Transactions Production Review

## Objective

Perform the final production-readiness review of the Transactions microservice.

This is the FINAL Transactions agent.

The purpose of this agent is to determine whether the Transactions service, as implemented by Agents 01–12, is ready to be considered production-ready within the RVPay architecture.

This is a REVIEW AND VALIDATION task.

Do not use this task as an opportunity to redesign the Transactions service.

Do not introduce new architecture.

Do not add unrelated features.

Do not expand the scope of the Transactions service.

The final outcome must clearly classify discovered issues as:

- BLOCKER
- HIGH
- MEDIUM
- LOW
- INFORMATIONAL

A production-readiness report must be created at:

docs/transactions-production-review.md

---

# Required Reading

Read only:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md

Then read the review documents produced by the previous Transactions agents:

- docs/transactions-existing-review.md
- docs/transactions-database-review.md
- docs/transactions-sqlc-review.md
- docs/transactions-protobuf-review.md
- docs/transactions-repository-review.md
- docs/transactions-merchants-review.md
- docs/transactions-customers-review.md
- docs/transactions-deposits-review.md
- docs/transactions-payouts-review.md
- docs/transactions-runtime-review.md
- docs/transactions-scaffolding-review.md
- docs/transactions-tests-review.md

Also inspect the root README.md carefully.

The README is the repository map.

Use it to understand:

- what existed before this implementation
- what has been replaced
- what remains
- how the repository is currently structured
- how services are expected to run
- what commands are expected to work

---

# Documentation Check

Before beginning the review:

confirm that all required documents are available and readable.

At completion:

perform the documentation check again.

Record the result in:

docs/transactions-production-review.md

---

# Repository Exploration Rules

Use:

README.md

as the primary repository map.

Use:

agents/project-context.md

for coding, package, naming, and implementation conventions.

Use:

docs/repository-layout.md

for repository structure.

Use:

docs/domain-model.md

for domain expectations.

Use:

docs/protobuf-strategy.md

for protobuf expectations.

Use:

docs/migration-plan.md

for compatibility and migration expectations.

Do NOT perform unrestricted recursive searches.

Do NOT inspect the entire repository simply because this is a production review.

Do NOT descend into irrelevant directories.

Do NOT inspect:

- third_party/
- third_party/googleapis/
- vendor/
- node_modules/
- .git/
- coverage/
- tmp/
- bin/

unless a concrete finding requires inspecting a specific file there.

Do not recursively inspect unrelated services.

Do not inspect generated protobuf code line-by-line.

Do not inspect every SQLC generated file individually.

---

# Review Philosophy

The objective is to answer:

> "Can the Transactions service safely be considered production-ready within the current RVPay architecture?"

Do not ask:

> "Can this code be made more sophisticated?"

Do not introduce improvements merely because another architecture could be better.

Evaluate the implementation that exists.

---

# 1. Establish Current Repository State

Run:

git status --short

Then inspect:

git diff --stat

Do not modify anything yet.

Determine whether there are:

- uncommitted Transactions changes
- generated files
- unexpected files
- unrelated modifications

Record relevant findings.

---

# 2. Verify Repository Scope

Confirm that the Transactions implementation is contained within the intended repository structure.

Expected areas may include:

transactions/

protobuf/

grpc/go/

database-related directories

documentation

configuration

root build files

Only treat additional modifications as acceptable if the previous agents explicitly required them.

Flag unrelated changes.

---

# 3. Verify Legacy Compatibility

Use:

README.md

to determine what the old RVPay system contained.

Determine:

- what the Transactions service replaced
- what was preserved
- what was migrated
- what was intentionally changed

Do not assume every old component must remain.

The migration plan is authoritative for intentional changes.

---

# 4. Verify Domain Model

Compare the implemented Transactions service against:

docs/domain-model.md

Check:

- Merchant
- Customer
- Transaction
- Deposit
- Payout
- transaction relationships
- identifiers
- statuses
- ownership boundaries

Do not redesign the domain model.

Flag implementation deviations.

---

# 5. Verify Repository Layout

Compare the actual repository against:

docs/repository-layout.md

Check:

- service boundaries
- package placement
- database ownership
- protobuf ownership
- generated-code placement
- runtime placement
- configuration placement

Flag unexpected structural deviations.

Do not move directories during this review unless explicitly required by the repository documentation.

---

# 6. Verify Protobuf Strategy

Read:

docs/protobuf-strategy.md

Verify:

- source protobuf contracts are in the expected location
- generated Go code is generated rather than manually written
- package names are consistent
- service definitions match implementation
- HTTP/gRPC gateway behavior matches the documented strategy where applicable

Do not edit generated protobuf files.

Do not change protobuf contracts during this review.

If a protobuf contract is wrong:

document it as a finding.

---

# 7. Verify Database Ownership

Inspect the Transactions database implementation.

Confirm:

- Transactions owns its intended database objects
- migrations exist
- migrations are ordered
- migrations are coherent
- schema matches the domain
- foreign keys are appropriate
- indexes exist where clearly required
- timestamps are handled consistently
- status fields are represented consistently

Do not perform a complete database optimization exercise.

Only flag issues that materially affect correctness or production readiness.

---

# 8. Migration Safety

Inspect all Transactions migrations.

Verify:

- up migrations exist
- down migrations exist where project conventions require them
- migration ordering is correct
- migrations are deterministic
- migrations do not silently destroy unrelated data
- migration names are consistent
- schema changes correspond to the intended domain

Pay particular attention to destructive operations.

A destructive migration without a clear reason is a potential BLOCKER or HIGH finding.

---

# 9. Migration Compatibility

Compare migrations against:

docs/migration-plan.md

Determine whether the implementation preserves the migration strategy.

Check:

- old data compatibility
- table ownership
- renamed entities
- identifier compatibility
- status compatibility

Do not invent migration requirements.

---

# 10. SQLC Verification

Verify that SQLC generation is reproducible.

Inspect:

- SQL inputs
- SQLC configuration
- generated output
- generation commands

Do not manually modify generated SQLC files.

Where the project already has a generation command:

run it.

If generation produces unexpected differences:

document them.

---

# 11. Generated-Code Verification

Verify that generated files are current.

Use the repository's established generation commands.

Where practical run the same generation process used by CI.

Check:

git diff

after generation.

Unexpected generated changes should be investigated.

---

# 12. Repository Layer Review

Review:

- merchant repositories
- customer repositories
- transaction repositories
- deposit repositories
- payout repositories

Check:

- context propagation
- error handling
- resource cleanup
- query usage
- transaction handling
- mapping
- consistency
- repository interfaces

Do not redesign repository abstractions.

---

# 13. Database Connection Handling

Verify:

- connection configuration
- pool creation
- pool lifecycle
- context handling
- connection cleanup
- startup failure behavior
- shutdown behavior

Check that the service does not silently continue when its required database cannot be reached.

---

# 14. Transaction Boundaries

Where database transactions are required:

verify that they are actually used.

Pay particular attention to operations that modify multiple related records.

Check for:

- partial writes
- missing rollback
- commits occurring too early
- ignored rollback errors
- inconsistent state

Do not introduce transactions where they are not required.

---

# 15. Merchant Service Review

Review the Merchant service.

Check:

- request validation
- repository calls
- error handling
- response construction
- gRPC status mapping
- context propagation
- database interaction

Verify behavior against the domain model.

---

# 16. Customer Service Review

Review the Customer service.

Check:

- merchant relationship
- request validation
- repository behavior
- error handling
- response construction
- gRPC status mapping
- context propagation

Flag ownership or authorization ambiguities if they materially affect correctness.

---

# 17. Deposit Service Review

Review Deposits.

Check:

- lifecycle
- validation
- merchant relationship
- customer relationship
- repository usage
- status handling
- provider abstraction
- error handling
- idempotency where explicitly required
- gRPC response behavior

Do not invent additional deposit states.

---

# 18. Payout Service Review

Review Payouts.

Check:

- lifecycle
- validation
- merchant relationship
- customer relationship where applicable
- repository usage
- status handling
- provider abstraction
- error handling
- idempotency where explicitly required

Do not invent additional payout behavior.

---

# 19. Provider Abstractions

Where provider interfaces exist:

verify that the Transactions service depends on abstractions rather than hard-coded external implementations.

Check:

- interface boundaries
- dependency injection
- error handling
- timeout handling
- context propagation

Do not add new providers.

Do not add SDKs.

Do not perform real external API calls.

---

# 20. External API Safety

If Transactions communicates with external providers:

verify:

- request timeouts
- context propagation
- error handling
- response validation
- no secrets in source code
- no credentials in logs

Do not perform live calls.

---

# 21. Secrets Review

Search only relevant Transactions configuration and source files for:

- API keys
- tokens
- passwords
- database credentials
- private keys
- secrets

Do NOT perform an unrestricted repository-wide secret scan if the repository already has a dedicated security process.

Inspect:

- configuration
- Dockerfile
- deployment configuration relevant to Transactions
- `.env.example`
- service source

Never expose discovered secrets in the review document.

If a real secret is found:

classify it as a BLOCKER.

Do not copy the secret into the report.

---

# 22. Environment Configuration

Verify:

- required environment variables are defined
- names match the implementation
- types are parsed correctly
- booleans are handled correctly
- URLs are validated where necessary
- defaults are intentional
- production configuration does not rely on `.env`

The service must not require a local `.env` file in production.

---

# 23. Configuration Failure Behavior

Verify that missing critical configuration causes clear startup failure.

Avoid silent defaults for critical production settings.

Examples:

- database URL
- service port
- provider credentials
- required API endpoints

Do not classify optional configuration as a blocker.

---

# 24. Dockerfile Review

Inspect the Transactions Dockerfile.

Verify:

- correct build context
- correct service entrypoint
- correct binary path
- correct working directory
- required files are copied
- multi-stage build where appropriate
- runtime image is not unnecessarily large
- no development credentials
- no unnecessary tools

Do not optimize Docker images beyond obvious correctness and security issues.

---

# 25. Container Startup

Verify the container startup command.

Confirm:

- executable exists
- executable path is correct
- permissions are correct
- required environment variables are read
- service starts in the expected container environment

If practical:

build the Docker image.

Do not deploy it.

---

# 26. Runtime Lifecycle

Review the Transactions runtime.

Check:

- startup logging
- configuration loading
- database initialization
- service construction
- gRPC server construction
- service registration
- graceful shutdown
- signal handling
- resource cleanup

The service must not leave database connections or other resources open unnecessarily.

---

# 27. Logging

Review logs for:

- useful startup information
- useful failure information
- structured logging consistency
- absence of secrets
- absence of sensitive credentials
- absence of excessive noisy logging

Do not require a new logging framework.

---

# 28. Error Handling

Look for:

- ignored errors
- swallowed errors
- panic-based control flow
- generic errors hiding root causes
- incorrect gRPC status codes
- database errors exposed directly to clients
- provider errors exposed unsafely

Classify according to actual impact.

---

# 29. Context Handling

Verify that request contexts are propagated through:

gRPC handler
→ service
→ repository
→ database

and external provider calls where applicable.

Check for:

- `context.Background()` replacing request context unnecessarily
- missing context parameters
- operations that cannot be cancelled

---

# 30. Timeout Handling

Verify that operations involving:

- external APIs
- database operations
- server requests

have appropriate timeout behavior according to project conventions.

Do not introduce arbitrary timeout values.

---

# 31. Idempotency

Check the TDD, domain documentation, and existing implementation for explicitly required idempotency.

If idempotency is required:

verify:

- idempotency key handling
- database uniqueness
- duplicate request behavior
- retry safety

If idempotency is not currently specified:

do not invent a new mechanism.

Document the absence only if it creates a concrete production risk.

---

# 32. Status Transitions

Verify transaction status transitions.

Check:

- valid transitions
- invalid transitions
- persistence
- returned API status

Do not invent new statuses.

---

# 33. Concurrency

Inspect for obvious concurrency hazards.

Examples:

- shared mutable state
- unsafe maps
- race-prone globals
- unsynchronized state

Do not perform a full performance benchmark.

Do not redesign concurrency unless a concrete correctness issue exists.

---

# 34. Goroutine Lifecycle

Check for obvious goroutine leaks.

Pay particular attention to:

- background workers
- provider polling
- listeners
- shutdown routines

Do not create workers merely because they might be useful.

---

# 35. gRPC Server Review

Verify:

- server creation
- service registration
- interceptors
- recovery
- logging
- reflection where intentionally enabled
- graceful shutdown

Do not expose additional endpoints.

---

# 36. gRPC Error Semantics

Verify that API errors use appropriate gRPC status codes.

Check examples such as:

- InvalidArgument
- NotFound
- AlreadyExists
- FailedPrecondition
- Unauthenticated
- PermissionDenied
- Internal
- Unavailable

Only use statuses supported by the actual application semantics.

---

# 37. API Contract Consistency

Compare protobuf definitions against service implementations.

Verify:

- method names
- request types
- response types
- field semantics
- optionality
- status representation

Do not modify protobuf contracts during this review.

---

# 38. Tests

Review the work produced by Agent 12.

Run:

go test ./...

where practical.

Run:

go test -race ./...

where practical.

Do not weaken tests.

Do not skip failing tests to obtain a green build.

---

# 39. Test Results

Classify failures:

1. Transactions defect
2. unrelated repository defect
3. environment issue
4. flaky test
5. dependency/toolchain issue

Do not automatically classify every failure as a Transactions problem.

---

# 40. Static Analysis

If existing project tooling supports:

- go vet
- staticcheck
- golangci-lint

run the established checks.

Do not introduce new tooling.

Record actual commands and results.

---

# 41. Formatting

Verify that modified Go code is formatted.

Use:

gofmt

only where necessary.

Do not reformat unrelated files.

---

# 42. Build Verification

Run the appropriate build command for the Transactions service.

Use the project's existing Makefile or documented build command.

Verify:

- compilation
- package dependencies
- generated code
- service entrypoint

Do not introduce new build systems.

---

# 43. Docker Build Verification

If the repository's workflow uses Docker:

build the Transactions image.

Verify:

- Docker build succeeds
- expected binary exists
- image starts
- required configuration is recognized

Do not deploy.

---

# 44. Dependency Review

Review relevant Transactions dependencies.

Look for:

- unnecessary dependencies
- suspicious replacements
- local paths
- accidental test dependencies in production
- incompatible versions

Do not upgrade dependencies merely because newer versions exist.

---

# 45. Go Module Integrity

Verify:

go.mod

and:

go.sum

are consistent.

Run:

go mod tidy

ONLY if the repository's established workflow permits it.

Do not blindly rewrite dependency versions.

If tidy would produce unrelated changes:

do not commit them.

Document the issue.

---

# 46. Generated Artifacts

Verify generated output is reproducible.

Check:

- protobuf
- gRPC
- SQLC
- mocks

Do not manually edit generated files.

If generated files differ:

determine whether the source or toolchain is responsible.

---

# 47. CI Compatibility

Inspect only the CI workflows directly relevant to Transactions.

Verify that:

- required tools are available
- generation commands are valid
- tests run
- Docker builds
- generated code remains synchronized

Do not redesign CI.

---

# 48. Deployment Compatibility

Use the repository's existing deployment documentation and README.

Verify:

- required environment variables
- container startup
- port configuration
- database configuration
- service command
- health behavior where implemented

Do not create new deployment infrastructure.

---

# 49. Render Compatibility

If the Transactions service is intended for Render:

verify the container and configuration are compatible with the existing Render deployment strategy.

Check:

- Dockerfile
- exposed/listening port
- environment variables
- startup command
- database URL usage

Do not create Render resources.

Do not deploy.

---

# 50. Health Checks

If the project already defines health/readiness behavior:

verify it.

If no health endpoint exists:

do not automatically create one.

Determine whether its absence is a production concern according to the TDD and deployment architecture.

Document rather than redesign.

---

# 51. Database Startup Behavior

Verify what happens when the database is unavailable.

Expected behavior should be intentional.

The service should not:

- falsely report readiness
- silently continue without required persistence
- loop indefinitely without clear behavior

---

# 52. Migration Startup Behavior

If the service runs migrations automatically:

verify:

- correct ordering
- failure handling
- database connection handling
- logging
- production safety

If migrations are externally managed:

verify that the service does not unexpectedly attempt to run them.

Follow the established architecture.

---

# 53. Security Boundaries

Review obvious application security issues.

Check:

- authentication expectations
- authorization boundaries
- secret handling
- sensitive data logging
- input validation
- unsafe SQL construction
- unsafe shell execution
- unsafe file access

Do not turn this into a full penetration test.

---

# 54. SQL Safety

Verify that application SQL is parameterized through SQLC or the established repository mechanism.

Flag:

- string-concatenated SQL
- user-controlled SQL fragments
- unsafe dynamic queries

Do not rewrite generated SQL manually.

---

# 55. Sensitive Data

Check whether logs or API responses expose:

- passwords
- API keys
- authentication tokens
- payment credentials
- sensitive customer information

Do not reproduce sensitive values in the review.

---

# 56. Input Limits

Look for obvious absence of limits around:

- request sizes
- strings
- numeric values
- pagination

Only flag issues where the absence creates a realistic production risk.

Do not introduce arbitrary limits.

---

# 57. Pagination

Where list endpoints exist:

verify pagination behavior.

Check:

- page size
- limits
- offsets/cursors
- deterministic ordering

Do not redesign pagination.

---

# 58. Database Indexes

Review obvious indexes needed for production query paths.

Examples:

- merchant IDs
- customer IDs
- transaction IDs
- status
- external/provider references
- timestamps

Only flag missing indexes where query patterns make the production impact plausible.

Do not perform premature database optimization.

---

# 59. Foreign Keys

Verify appropriate foreign-key relationships.

Check:

- merchant ownership
- customer ownership
- transaction ownership
- deposit ownership
- payout ownership

Do not add relationships not defined by the domain model.

---

# 60. Uniqueness

Verify uniqueness constraints where the domain requires them.

Examples may include:

- public identifiers
- provider references
- idempotency keys
- merchant identifiers

Only enforce uniqueness where the domain or implementation requires it.

---

# 61. Nullability

Review database nullability.

Check that required domain values cannot unexpectedly become NULL.

Do not change schema solely for stylistic reasons.

---

# 62. Timestamp Handling

Verify consistency of:

- created_at
- updated_at
- completed_at
- failed_at

where applicable.

Check timezone assumptions.

Do not redesign timestamp handling.

---

# 63. Monetary Values

Review handling of financial amounts.

Verify:

- appropriate database type
- appropriate Go representation
- no accidental floating-point use for money
- precision consistency

This is a high-priority area because Transactions handles financial operations.

---

# 64. Currency Handling

Verify currency representation.

Check:

- database representation
- protobuf representation
- validation
- consistency between deposit and payout operations

Do not introduce multi-currency behavior unless already defined.

---

# 65. Financial State Integrity

Verify that the system does not easily create impossible financial states.

Look for:

- negative amounts where prohibited
- duplicate transactions
- impossible status transitions
- orphaned transactions
- missing merchant/customer relationships

Flag concrete integrity risks.

---

# 66. Provider Reference Integrity

Where external provider references exist:

verify they are stored consistently.

Check:

- uniqueness where required
- nullable behavior
- mapping
- status synchronization

Do not introduce provider polling or reconciliation systems.

---

# 67. Transaction Atomicity

Where a single business operation modifies multiple records:

verify atomicity.

Examples:

- creating transaction + related record
- updating transaction + status
- provider response + transaction state

Only flag actual partial-state risks.

---

# 68. Retry Safety

Consider whether operations can safely be retried.

Pay particular attention to:

- deposits
- payouts
- external provider calls
- database writes

Do not implement a new retry framework.

Flag concrete retry hazards.

---

# 69. External Provider Failure

Verify behavior when an external provider:

- times out
- returns an error
- returns malformed data
- becomes unavailable

The application should fail predictably.

Do not add resilience infrastructure during this agent.

---

# 70. Error Recovery

Review whether recoverable failures are handled appropriately.

Examples:

- database temporary failure
- provider unavailable
- invalid provider response

Do not hide failures.

---

# 71. Resource Cleanup

Verify cleanup of:

- database pools
- HTTP clients where lifecycle requires it
- gRPC servers
- background workers

Check shutdown paths.

---

# 72. Graceful Shutdown

Verify that SIGINT/SIGTERM handling:

- stops accepting new requests
- allows active requests to finish where appropriate
- closes resources
- exits cleanly

Follow existing runtime conventions.

---

# 73. Startup Ordering

Verify:

configuration
→ database
→ dependencies
→ services
→ server

or the documented equivalent.

The service should not expose itself as ready before required dependencies are initialized.

---

# 74. Runtime Logging

Verify startup logs clearly indicate:

- service startup
- configuration success/failure
- database initialization
- server startup
- shutdown

Do not log secrets.

---

# 75. Configuration Naming

Verify consistency among:

- `.env.example`
- config structs
- deployment configuration
- Docker usage
- documentation

Flag mismatches.

---

# 76. Documentation

Check whether the Transactions service documentation accurately describes:

- setup
- configuration
- running locally
- migrations
- tests
- Docker
- runtime

Do not create a separate documentation system.

Update documentation only if it is demonstrably incorrect and the correction belongs within the review scope.

---

# 77. README Consistency

The root README is authoritative as the repository map.

If the Transactions implementation changed the repository structure:

verify that README.md still accurately describes the relevant structure.

Do not rewrite unrelated README sections.

---

# 78. Makefile Verification

If Transactions has a Makefile:

verify commands for:

- generation
- testing
- migration
- build
- run

Ensure commands correspond to actual files.

Do not add large numbers of convenience commands.

---

# 79. Scaffolding Verification

Review the work from Agent 11.

Check:

- Dockerfile
- Makefile
- README
- `.env.example`
- directory structure

Ensure scaffolding reflects the actual service.

---

# 80. Test Review

Review:

docs/transactions-tests-review.md

Check whether:

- tests actually cover critical behavior
- failures were resolved
- known limitations remain

Do not duplicate Agent 12's test implementation work.

---

# 81. Previous-Agent Review Consistency

Read all previous review documents.

Look for unresolved findings.

Create a tracking table:

| Finding | Source Agent | Severity | Resolved? | Evidence |
|---|---|---|---|---|

Every significant previous finding must be accounted for.

---

# 82. Blocker Identification

A BLOCKER is something that prevents safe production use.

Examples:

- service does not build
- service cannot connect to its required database
- secrets committed
- destructive migration
- corrupted financial state possible
- critical transaction operation broken
- generated code cannot be reproduced
- runtime cannot start
- required configuration is missing or broken

Do not classify minor issues as blockers.

---

# 83. High-Severity Findings

HIGH findings are serious production risks that should be resolved before normal production use.

Examples:

- unsafe retry behavior
- incorrect transaction state transitions
- missing critical database constraint
- serious error handling problem
- incorrect provider failure behavior
- significant data integrity risk

---

# 84. Medium-Severity Findings

MEDIUM findings are important but do not necessarily prevent initial deployment.

Examples:

- incomplete validation
- operational visibility gaps
- non-critical missing index
- incomplete documentation
- limited test coverage

---

# 85. Low-Severity Findings

LOW findings include:

- minor inconsistencies
- small maintainability concerns
- documentation improvements
- minor code quality issues

Do not inflate their importance.

---

# 86. Informational Findings

Use INFORMATIONAL for:

- intentional architectural choices
- known limitations
- future considerations
- areas that may deserve monitoring

These are not defects.

---

# 87. No Speculative Findings

Do not report hypothetical issues without evidence.

Every finding must reference:

- a specific file
- relevant code
- a documented requirement
- a reproducible command/result
- or an explicit architectural rule

---

# 88. No Architectural Redesign

Do NOT propose:

- event-driven architecture
- Kafka
- RabbitMQ
- Redis
- CQRS
- event sourcing
- service mesh
- Kubernetes
- new API gateway
- new authentication system
- new provider abstraction

unless such technology is already part of the documented architecture and the issue directly concerns its implementation.

---

# 89. No Performance Project

Do not perform:

- load testing
- stress testing
- benchmarking
- profiling
- database query optimization campaigns

unless specifically required by existing project documentation.

Document obvious performance blockers only.

---

# 90. No Deployment

Do not:

- deploy to Render
- create cloud services
- modify production databases
- trigger deployment hooks
- modify production secrets

This is a readiness review.

---

# 91. No Production Credentials

Never request or use:

- production database credentials
- provider API keys
- Render secrets
- customer credentials

The review must operate using local/test configuration.

---

# 92. Fix Policy

This agent may make SMALL corrective changes only when:

1. the issue is clearly demonstrated,
2. the fix is unambiguous,
3. the fix is within Transactions scope,
4. the fix does not redesign architecture.

Examples:

- obvious incorrect configuration variable
- missing error return
- incorrect Docker entrypoint
- broken test command
- incorrect generated-code command

---

# 93. Do Not Fix Architectural Findings

If fixing a finding would require:

- schema redesign
- protobuf redesign
- service boundary changes
- new infrastructure
- new external services
- broad refactoring

do NOT implement the fix.

Document it.

---

# 94. Re-run Validation After Fixes

If any corrective changes are made:

run the relevant:

- tests
- generation checks
- build
- Docker build

again.

Do not report a fix without verification.

---

# 95. Final Generation Check

If generated code is part of the service:

run the documented generation process.

Then:

git diff

Confirm generated code is synchronized.

---

# 96. Final Test Check

Run:

go test ./...

where practical.

Run:

go test -race ./...

where practical.

Record actual results.

---

# 97. Final Build Check

Run the documented Transactions build command.

Record:

- success/failure
- command
- relevant output

---

# 98. Final Docker Check

If Docker is part of the service:

build the image.

If practical, run it locally using test configuration.

Do not connect it to production infrastructure.

---

# 99. Final Git Review

Run:

git status --short

Then:

git diff --stat

Then inspect all relevant changes.

Confirm:

- no unrelated modifications
- no secrets
- no generated-code surprises
- no accidental deletions
- no third-party changes
- no unrelated service changes

---

# 100. Production Readiness Matrix

Create a final matrix:

| Area | Status | Severity | Notes |
|---|---|---|---|
| Domain model | PASS/FAIL | | |
| Repository layout | PASS/FAIL | | |
| Database | PASS/FAIL | | |
| Migrations | PASS/FAIL | | |
| SQLC | PASS/FAIL | | |
| Protobuf | PASS/FAIL | | |
| Repositories | PASS/FAIL | | |
| Merchants | PASS/FAIL | | |
| Customers | PASS/FAIL | | |
| Deposits | PASS/FAIL | | |
| Payouts | PASS/FAIL | | |
| Runtime | PASS/FAIL | | |
| Configuration | PASS/FAIL | | |
| Docker | PASS/FAIL | | |
| Tests | PASS/FAIL | | |
| Security | PASS/FAIL | | |
| Financial integrity | PASS/FAIL | | |
| Documentation | PASS/FAIL | | |
| CI compatibility | PASS/FAIL | | |

Use:

PASS

only when evidence supports it.

---

# 101. Production Readiness Decision

At the end of the review provide exactly one overall status:

## READY

No BLOCKER or HIGH findings remain.

OR:

## READY WITH CONDITIONS

No BLOCKER remains, but documented MEDIUM/LOW issues should be addressed or consciously accepted.

OR:

## NOT READY

One or more BLOCKER or HIGH findings remain.

Do not label the service READY simply because tests pass.

---

# 102. Required Production Review Document

Create:

docs/transactions-production-review.md

Use this structure:

# Transactions Production Readiness Review

## 1. Executive Summary

Briefly explain the final state.

## 2. Scope

Explain what was reviewed.

## 3. Required Documents

List every required document.

## 4. Repository State

Record relevant git status/diff findings.

## 5. Architecture Verification

Document alignment with:

- domain model
- repository layout
- protobuf strategy
- migration plan

## 6. Database Review

Document:

- schema
- migrations
- constraints
- indexes
- ownership
- transaction safety

## 7. Repository Review

Summarize repository findings.

## 8. Service Review

Summarize:

- Merchants
- Customers
- Deposits
- Payouts

## 9. Runtime Review

Document:

- configuration
- startup
- gRPC
- shutdown
- dependencies
- logging

## 10. Security Review

Document:

- secrets
- credentials
- sensitive logging
- SQL safety
- input validation

Do not include actual secret values.

## 11. Testing

Document commands and results.

## 12. Build

Document build results.

## 13. Docker

Document Docker build/runtime results.

## 14. CI Compatibility

Document relevant CI validation.

## 15. Previous Agent Findings

Use a table:

| Finding | Agent | Severity | Status |
|---|---|---|---|

## 16. Current Findings

Use:

| ID | Severity | Area | Finding | Evidence | Recommendation |
|---|---|---|---|---|---|

## 17. Blockers

List all blockers.

If none:

None identified.

## 18. High-Severity Issues

List all HIGH issues.

If none:

None identified.

## 19. Medium/Low Issues

List remaining issues.

## 20. Corrective Changes

List any changes made by this agent.

## 21. Validation Commands

List commands actually executed.

## 22. Production Readiness Matrix

Include the matrix defined above.

## 23. Final Decision

Use exactly one:

READY

READY WITH CONDITIONS

NOT READY

## 24. Documentation Check

Confirm all required documents were read.

## 25. Final Scope Check

Confirm that unrelated services and directories were not modified.

---

# 103. Final Documentation Check

Before finishing, confirm that the following were read:

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
- docs/transactions-runtime-review.md
- docs/transactions-scaffolding-review.md
- docs/transactions-tests-review.md

Record this explicitly in:

docs/transactions-production-review.md

---

# 104. Final Scope Check

Before stopping:

run:

git status --short

Then:

git diff --stat

Verify that changes are limited to:

- Transactions implementation fixes explicitly required by this review
- generated output that is legitimately required
- docs/transactions-production-review.md

If unrelated changes are present:

DO NOT clean them up automatically.

Document them.

---

# 105. Final Completion Checklist

Before stopping:

- [ ] README.md was read.
- [ ] agents/project-context.md was read.
- [ ] docs/domain-model.md was read.
- [ ] docs/repository-layout.md was read.
- [ ] docs/protobuf-strategy.md was read.
- [ ] docs/migration-plan.md was read.
- [ ] All Transactions agent review documents were read.
- [ ] Agent 12 test review was read.
- [ ] Repository state was inspected.
- [ ] Domain alignment was verified.
- [ ] Repository layout was verified.
- [ ] Database schema was reviewed.
- [ ] Migrations were reviewed.
- [ ] SQLC generation was reviewed.
- [ ] Protobuf generation was reviewed.
- [ ] Repository layer was reviewed.
- [ ] Merchant service was reviewed.
- [ ] Customer service was reviewed.
- [ ] Deposit service was reviewed.
- [ ] Payout service was reviewed.
- [ ] Runtime was reviewed.
- [ ] Configuration was reviewed.
- [ ] Dockerfile was reviewed.
- [ ] Secrets were checked in relevant scope.
- [ ] Financial data handling was reviewed.
- [ ] Error handling was reviewed.
- [ ] Context handling was reviewed.
- [ ] Transaction boundaries were reviewed.
- [ ] Provider boundaries were reviewed.
- [ ] Tests were executed.
- [ ] Race detection was executed where practical.
- [ ] Static validation was executed where applicable.
- [ ] Build was verified.
- [ ] Docker build was verified where applicable.
- [ ] Previous agent findings were accounted for.
- [ ] Current findings were categorized.
- [ ] No speculative findings were introduced.
- [ ] No architectural redesign was performed.
- [ ] No deployment was performed.
- [ ] No production credentials were used.
- [ ] Final git status was inspected.
- [ ] Final diff was inspected.
- [ ] docs/transactions-production-review.md was created.
- [ ] Final production-readiness decision was recorded.
- [ ] Documentation check was recorded.

---

# Final Stop Condition

This is the final Transactions agent.

STOP after:

1. completing the production-readiness review,
2. resolving only small, unambiguous defects that are safe to correct,
3. running the required validation,
4. reviewing the final diff,
5. documenting all findings,
6. creating docs/transactions-production-review.md,
7. recording the final production-readiness decision.

Do NOT proceed to:

- new service creation
- architecture redesign
- deployment
- cloud infrastructure
- Render configuration changes
- production database changes
- provider integration
- OAuth
- webhooks
- Redis
- Kafka
- queues
- performance engineering
- unrelated services

If significant work remains:

DO NOT implement it.

Document it clearly in:

docs/transactions-production-review.md

The final report is the handoff point to the developer.

STOP.