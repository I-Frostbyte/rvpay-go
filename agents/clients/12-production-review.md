# Agent 12 — Production Review & Deployment Readiness

## Objective

Perform the final production-readiness review of the Clients Service.

This is the final review agent for the Clients Service implementation.

The purpose of this agent is to determine whether the Clients Service is:

- architecturally consistent
- functionally complete
- secure
- operationally sound
- deployable
- compatible with the existing RVPay repository
- ready for integration with the rest of the system

This agent is primarily an AUDIT agent.

It must not redesign the Clients Service.

It must not introduce new architectural patterns.

It must not expand the scope of the Clients Service.

It must not refactor working code merely for stylistic preference.

It must not modify unrelated services.

Small, obvious defects may be fixed when the correct fix is unambiguous and does not alter the architecture.

Anything requiring an architectural decision must be reported rather than implemented.

---

# Required Reading

Read only:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md

Then read the Clients implementation and the review documents produced by previous agents.

Read:

- clients/docs/runtime-review.md
- clients/docs/developer-experience-review.md
- clients/docs/test-review.md

Then inspect:

- clients/
- deposits/

Use Deposits only as a reference for established repository conventions.

Do not recursively inspect the entire repository.

---

# Repository Exploration Rules

Use the root README.md as the repository map.

Use agents/project-context.md as the coding and package conventions.

Use the foundation documents as the architectural source of truth.

Do not explore unrelated directories.

Do not inspect large dependency or generated-code trees unless a specific failure requires it.

Do not inspect:

- third_party/
- googleapis/
- vendor/
- node_modules/
- .git/
- coverage/
- tmp/
- bin/

Do not recursively inspect the entire repository looking for potential problems.

Only inspect files directly relevant to the Clients Service or to a concrete dependency discovered during review.

---

# Source of Truth Hierarchy

When evaluating the implementation, use this order of authority:

1. Existing repository architecture and conventions
2. README.md
3. agents/project-context.md
4. docs/domain-model.md
5. docs/repository-layout.md
6. docs/protobuf-strategy.md
7. docs/migration-plan.md
8. Clients implementation
9. Previous agent review documents

Do not silently replace documented project decisions with personal architectural preferences.

If two sources appear contradictory:

1. identify the contradiction
2. determine whether it is intentional
3. do not silently choose a new architecture
4. document the issue

---

# Review Scope

Review the following areas:

- domain model
- package structure
- database schema
- migrations
- sqlc
- repositories
- provider architecture
- business services
- OAuth
- webhooks
- protobuf
- gRPC
- REST/gateway
- runtime
- configuration
- logging
- error handling
- graceful shutdown
- Docker
- Makefile
- environment configuration
- tests
- CI compatibility
- Render compatibility
- security
- secrets management
- observability
- operational readiness

---

# 1. Domain Model Review

Verify that the implementation matches:

docs/domain-model.md

Check:

- entities
- relationships
- identifiers
- lifecycle states
- ownership
- provider concepts
- OAuth concepts
- webhook concepts
- integration boundaries

Ensure no accidental duplication exists between Clients and Deposits.

The Clients Service should own only the concepts assigned to it.

Do not move business ownership between services during this review.

---

# 2. Repository Layout Review

Compare:

clients/

against:

docs/repository-layout.md

Verify:

- package locations
- command location
- config location
- database layout
- migration location
- query location
- repository location
- generated-code location
- protobuf usage
- service implementation location

Ensure the Clients service follows the repository's established structure.

Do not reorganize directories unless required to correct a clear violation of the documented architecture.

---

# 3. Database Review

Review:

clients/db/migrations/

Verify:

- migrations are sequential
- migrations have up/down paths where required
- migrations are deterministic
- foreign keys are appropriate
- indexes exist for important lookup paths
- unique constraints enforce required invariants
- nullable fields are intentional
- timestamps are handled consistently
- provider identifiers are constrained appropriately
- OAuth data is persisted safely
- webhook events support idempotency

Check for:

- destructive migrations
- accidental data loss
- missing constraints
- missing indexes
- duplicate schema concepts
- unsafe defaults

Do not rewrite migrations that have already been applied in a real environment.

If a migration problem is discovered, document the required follow-up migration rather than modifying historical migrations.

---

# 4. SQLC Review

Verify:

- generated code corresponds to current SQL
- queries follow project conventions
- query parameters are appropriate
- transaction handling is correct
- generated models are not manually modified
- nullable fields are handled correctly
- database errors are propagated correctly

Regenerate generated code only if required to verify consistency.

Do not manually edit generated files.

---

# 5. Repository Review

Review repository implementations for:

- correct database usage
- context propagation
- error handling
- transaction boundaries
- resource cleanup
- idempotency
- concurrency safety

Verify repositories do not contain business logic that belongs in services.

Verify repositories do not depend on transport-layer types.

---

# 6. Provider Architecture Review

Review the Provider interface and provider registry.

Verify:

- interfaces are appropriately small
- provider implementations are isolated
- provider registration is deterministic
- unknown providers fail safely
- duplicate registration is handled
- provider-specific logic remains inside provider implementations
- business services do not depend directly on HighLevel-specific implementation details

Do not introduce additional provider abstractions.

Do not add future providers.

The purpose of this review is to verify the existing architecture, not expand it.

---

# 7. OAuth Security Review

Perform a security-focused review of the OAuth implementation.

Verify:

- state validation exists
- authorization codes are handled safely
- client secrets are never exposed
- access tokens are not logged
- refresh tokens are not logged
- tokens are not returned unnecessarily
- callback parameters are validated
- token exchange errors are handled
- expired tokens are handled
- refresh failures are handled
- provider credentials come from configuration
- secrets are not committed

Search only relevant Clients files.

Do not perform an unrestricted repository-wide search.

---

# 8. Webhook Security Review

Verify:

- webhook signatures are validated
- validation occurs before processing
- malformed payloads are rejected
- unsupported events are handled safely
- duplicate events are handled
- replay protection exists where required by the design
- provider secrets are not logged
- webhook payloads do not leak credentials
- processing failures are observable

Verify webhook processing cannot accidentally trigger duplicate business operations.

---

# 9. API Review

Review the protobuf contracts and generated transport code used by Clients.

Verify:

- request validation exists
- response types are appropriate
- error mapping is correct
- HTTP annotations match intended REST endpoints
- gRPC methods map to the correct business operations
- no business logic exists inside handlers
- REST and gRPC use the same service layer

Do not redesign protobuf contracts.

If a protobuf change is necessary, document it as a follow-up architectural change.

---

# 10. Runtime Review

Verify:

- configuration loads correctly
- logger initializes correctly
- database initializes correctly
- migrations execute correctly
- repositories initialize correctly
- providers initialize correctly
- services initialize correctly
- gRPC server starts
- REST gateway starts
- health checks work
- graceful shutdown works

Verify dependency initialization follows:

Configuration
↓
Logger
↓
Database
↓
Repositories
↓
Providers
↓
Services
↓
Handlers
↓
Servers

Startup must fail fast when critical dependencies cannot initialize.

---

# 11. Configuration Review

Review:

clients/config/

and:

clients/.env.example

Verify:

- every required environment variable is documented
- variable names are consistent
- types are correct
- defaults are safe
- production secrets are not hardcoded
- local development remains possible
- Render environment variables can supply configuration
- `.env` is not committed

Pay particular attention to:

database configuration

OAuth configuration

provider configuration

server ports

migration flags

logging

environment name

---

# 12. Secret Review

Search only the relevant Clients implementation and configuration files for:

- API keys
- client secrets
- access tokens
- refresh tokens
- database passwords
- signing secrets
- private keys
- bearer tokens

Verify no real credentials are committed.

Do not print discovered secrets into the final report.

If a secret is discovered:

1. identify the file
2. redact the value in the report
3. remove it from source if appropriate
4. recommend rotation if it may have been exposed

Do not expose the secret value.

---

# 13. Logging Review

Verify logs do not contain:

- OAuth client secrets
- authorization codes
- access tokens
- refresh tokens
- webhook signatures
- database passwords
- provider credentials

Logs should provide enough information to diagnose:

startup

database failures

OAuth failures

webhook failures

provider failures

request failures

shutdown

without exposing sensitive information.

---

# 14. Error Handling Review

Verify errors:

- preserve useful context
- use wrapping appropriately
- map to appropriate gRPC statuses
- do not expose internal database details unnecessarily
- do not expose secrets
- do not silently disappear
- distinguish expected business errors from infrastructure errors

Look for:

ignored errors

empty error handling

panic-based control flow

incorrect status codes

misleading error messages

---

# 15. Concurrency Review

Review concurrency-sensitive components.

Pay particular attention to:

- provider registry
- webhook processing
- OAuth state
- runtime shutdown
- shared caches/state
- goroutines
- channels
- background workers

Verify:

- goroutines have termination paths
- contexts are propagated
- shutdown does not leak goroutines
- shared mutable state is protected
- race conditions are not obvious

Run race detection if practical.

Do not redesign concurrency architecture during this review.

---

# 16. Docker Review

Review the Clients Dockerfile.

Verify:

- multi-stage build
- correct Go version
- correct target architecture
- minimal runtime image
- no development credentials
- no unnecessary build artifacts
- correct binary path
- correct entrypoint
- correct port behavior
- environment variables are runtime-provided

Build the image.

Do not introduce a new container strategy.

---

# 17. Makefile Review

Verify Makefile commands for:

- build
- run
- generate
- protobuf generation
- sqlc generation
- tests
- Docker build
- cleanup

Ensure commands work from the expected working directory.

Do not introduce unnecessary Makefile targets.

---

# 18. CI Review

Review only the CI workflow(s) directly responsible for the Clients service or repository build.

Verify:

- Go version compatibility
- generated code verification
- protobuf generation
- sqlc generation
- tests
- Docker build
- deployment trigger

Ensure generated files are reproducible.

Do not redesign CI unless there is a concrete failure.

---

# 19. Render Review

Verify the Clients service is compatible with the existing Render deployment approach.

Check:

- Docker build context
- Dockerfile path
- exposed port
- environment variables
- database connection
- migration behavior
- health check behavior
- startup command
- shutdown behavior

Do not create a new Render architecture during this agent.

---

# 20. Test Review

Read:

clients/docs/test-review.md

Verify that the reported tests actually correspond to the implementation.

Run the critical test suite again.

At minimum verify:

go test ./clients/...

Then, if practical:

go test ./...

Also run:

go vet ./clients/...

where supported.

Run:

go test -race ./clients/...

where practical.

Do not modify tests simply to make the review pass.

---

# 21. Build Verification

Verify:

go build ./...

If the repository has an established build command, use it as well.

Verify the Clients service executable can be built successfully.

---

# 22. Generated Code Verification

Verify generated code is current.

Use the project's established generation commands.

Check:

protobuf

grpc

grpc-gateway

sqlc

mock generation

Do not manually modify generated output.

If generation produces unexpected changes:

1. inspect the diff
2. determine why it changed
3. verify tooling versions
4. do not blindly commit generated changes

---

# 23. Migration Verification

Verify migrations can be applied in a clean database.

Where practical:

1. create clean test database
2. apply migrations
3. verify schema
4. execute relevant repository operations
5. rollback where the project's migration tooling supports it
6. verify rollback behavior

Do not modify historical migrations simply because rollback exposes a design problem.

Document required follow-up migrations.

---

# 24. API Smoke Test

Where practical, perform a basic local smoke test.

Verify:

service starts

gRPC endpoint becomes available

REST gateway becomes available

health endpoint responds

configuration is accepted

database connection succeeds

shutdown completes

Do not require real OAuth credentials.

Do not call production HighLevel endpoints.

---

# 25. Scope Review

Confirm that the Clients implementation has not accidentally introduced:

- transaction business logic
- payment logic
- deposit logic
- unrelated infrastructure
- unrelated database tables
- unrelated protobuf APIs
- unrelated configuration
- duplicate service implementations

The Clients Service should remain within its documented responsibility.

---

# 26. Changes Allowed During This Agent

You may fix only:

- obvious compilation errors
- broken imports
- incorrect configuration defaults
- incorrect documentation
- incorrect Makefile commands
- obvious security leaks
- incorrect Docker paths
- obvious test defects
- small runtime defects
- missing error handling

Only make a change if its intended behavior is unambiguous.

---

# 27. Changes Forbidden During This Agent

Do not:

- redesign the service
- redesign the database
- redesign protobuf
- redesign OAuth
- redesign providers
- redesign repositories
- change service boundaries
- merge Clients with Deposits
- introduce unrelated services
- introduce new frameworks
- replace working libraries
- rewrite working code for style
- modify unrelated services
- rewrite historical migrations

If one of these appears necessary, document it as a blocker.

---

# 28. Production Readiness Classification

Every discovered issue must be classified as:

## BLOCKER

The service cannot safely be deployed or operated.

Examples:

- compilation failure
- broken startup
- exposed secret
- broken database migration
- broken OAuth security
- broken webhook signature validation
- unrecoverable runtime failure

## HIGH RISK

The service may run but has a significant production reliability or security concern.

Examples:

- missing idempotency
- unsafe concurrency
- missing error handling
- unreliable shutdown
- serious observability gap

## MEDIUM RISK

The service can operate but should be improved before broader production use.

## LOW RISK

Non-critical improvements.

## INFORMATIONAL

Observations that do not currently require action.

---

# 29. Final Review Report

Create:

clients/docs/production-readiness-review.md

The report must contain:

# Clients Service — Production Readiness Review

## Executive Summary

Provide a concise assessment.

## Architecture

Describe whether the implementation matches:

- domain model
- repository layout
- service boundaries
- provider architecture
- protobuf strategy

## Database

Summarize:

- schema
- migrations
- indexes
- constraints
- sqlc

## Runtime

Summarize:

- startup
- dependency injection
- gRPC
- REST
- health
- shutdown

## OAuth

Summarize:

- flow
- state handling
- token handling
- security

## Webhooks

Summarize:

- validation
- processing
- idempotency
- security

## Testing

Summarize:

- tests executed
- test results
- race testing
- static analysis
- integration testing

## Deployment

Summarize:

- Docker
- Render compatibility
- configuration
- database
- migrations

## Security

Summarize:

- secret handling
- OAuth security
- webhook security
- logging

## Findings

Use a table:

| Severity | Area | Finding | Required Action |
|---|---|---|---|

## Blockers

List all deployment blockers.

If none exist, explicitly state:

"No production blockers identified."

## Remaining Risks

List risks that remain after implementation.

## Recommended Follow-ups

List non-blocking improvements.

## Final Verdict

Choose exactly one:

READY

READY WITH WARNINGS

NOT READY

Explain the decision.

---

# 30. Final Repository Check

Before completing:

Run:

git status --short

Review every changed file.

Ensure changes belong to the Clients implementation.

Do not modify unrelated files.

If unexpected files changed:

- identify why
- revert unrelated changes if safe
- do not delete user work

Review:

git diff --stat

and relevant diffs.

---

# Completion Rules

The agent is complete only when:

- Clients builds successfully.

- Clients tests pass.

- Full repository tests have been executed where practical.

- Configuration has been reviewed.

- Database migrations have been reviewed.

- SQLC has been verified.

- Protobuf generation has been verified.

- Runtime has been reviewed.

- OAuth has been security-reviewed.

- Webhooks have been security-reviewed.

- Provider architecture has been reviewed.

- Docker has been reviewed and built.

- Render compatibility has been reviewed.

- Secrets have been checked.

- Logging has been checked.

- graceful shutdown has been reviewed.

- generated code is reproducible.

- no unrelated files were modified.

- production-readiness-review.md has been created.

Do not declare the service READY merely because tests pass.

---

# Final Stop Condition

After producing the production-readiness report:

STOP.

Do not begin implementing the next RVPay service.

Do not expand the Clients Service.

Do not create future providers.

Do not redesign the Transactions Service.

Do not modify unrelated services.

The final report is the handoff to the developer for human review and approval.