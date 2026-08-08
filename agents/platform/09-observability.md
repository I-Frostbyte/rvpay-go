# Agent 09 — Observability

## Objective

Implement and document the observability foundation for the RVPay platform.

The goal is to make the running RVPay system observable enough to answer:

- Is a service running?
- Is a request reaching the service?
- Which service handled the request?
- Which operation was executed?
- How long did it take?
- Did it succeed or fail?
- What error occurred?
- Which request/trace does the error belong to?
- Is the database operation failing?
- Is an external provider call failing?
- Is the service repeatedly restarting?

Observability must be implemented in a way that fits the existing RVPay architecture.

Do NOT redesign the application architecture.

Do NOT introduce unnecessary infrastructure.

Do NOT implement security features.

Do NOT perform performance optimization.

Do NOT rewrite existing services merely to make logging look different.

The implementation must preserve the project's existing logging conventions and package structure.

---

# Required Reading

Read only:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md
- docs/platform-repository-audit.md
- docs/platform-protobuf-generation-review.md
- docs/platform-http-gateway-review.md
- docs/platform-common-packages-review.md
- docs/platform-ci-cd-review.md
- docs/platform-docker-review.md
- docs/platform-render-review.md
- docs/platform-documentation-review.md

Also inspect only the specific application/configuration files referenced by those documents when implementation verification is required.

---

# Documentation Check

Before starting:

verify that all required documents exist.

Required:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md
- docs/platform-repository-audit.md
- docs/platform-protobuf-generation-review.md
- docs/platform-http-gateway-review.md
- docs/platform-common-packages-review.md
- docs/platform-ci-cd-review.md
- docs/platform-docker-review.md
- docs/platform-render-review.md
- docs/platform-documentation-review.md

If any required document is missing:

STOP.

Do not recreate the missing document.

At the end of the task:

perform the documentation check again.

Create:

docs/platform-observability-review.md

and record the result.

---

# Repository Exploration Rules

## IMPORTANT

Do NOT perform an unrestricted repository-wide search.

Use:

README.md

as the primary repository map.

Use:

docs/repository-layout.md

as the authority for the intended structure.

Use:

agents/project-context.md

as the authority for coding and package conventions.

Use the existing platform review documents to determine what has already been inspected.

Only inspect implementation files necessary to understand or integrate observability.

---

# Do NOT Explore Deep Folders

Do NOT recursively inspect:

- .git/
- vendor/
- node_modules/
- coverage/
- tmp/
- bin/
- third_party/
- third_party/googleapis/

Especially:

DO NOT inspect:

third_party/googleapis/

Generated protobuf internals are not the target of this task.

---

# 1. Existing Observability Audit

Before writing code:

determine what observability already exists.

Inspect only the relevant locations identified by the repository documentation.

Look for:

- logging package
- logger initialization
- structured logging
- request logging
- gRPC logging
- HTTP logging
- error logging
- database logging
- startup logging
- shutdown logging
- health endpoints
- metrics
- tracing
- correlation/request IDs

Do not assume that none of these exist.

---

# 2. Preserve Existing Logger

The repository already has established logging conventions.

If the existing system uses a structured logger:

continue using it.

Do NOT replace it with:

- another logging library
- a global logging framework
- an unrelated logging abstraction

unless the existing architecture explicitly requires it.

---

# 3. Logger Configuration

Determine:

- where the logger is initialized
- how log level is configured
- whether timestamps are enabled
- whether caller information is enabled
- whether JSON output is used
- how errors are represented

Preserve these conventions.

---

# 4. Environment Configuration

Determine whether the logger supports configuration through environment variables.

If it already does:

document and preserve it.

If observability configuration is missing:

add only the configuration required by the approved observability design.

Do not add a large number of speculative environment variables.

---

# 5. Log Levels

Use conventional levels where supported by the existing logger:

- debug
- info
- warn
- error

Do not introduce custom levels unnecessarily.

---

# 6. Log Level Rules

Use:

DEBUG

for detailed diagnostic information that should generally not appear in normal production operation.

Use:

INFO

for normal lifecycle and significant operational events.

Use:

WARN

for abnormal but recoverable conditions.

Use:

ERROR

for failed operations requiring investigation.

Do not log every line of application execution.

---

# 7. Structured Logging

Logs should be structured.

Prefer fields such as:

- service
- operation
- request_id
- trace_id
- error
- duration
- status
- method
- path
- rpc
- database
- provider

only where relevant.

Do not create meaningless fields.

---

# 8. No Log Spam

Do not log:

- every variable
- every function call
- every successful database query
- every internal helper call

unless specifically required for debugging.

Production logs must remain useful.

---

# 9. Startup Logging

Every long-running service should provide a useful startup log.

At minimum where appropriate:

- service identity
- environment
- listening address/port
- startup stage

Do NOT log secrets.

---

# 10. Configuration Logging

Never log:

- passwords
- database passwords
- API keys
- OAuth client secrets
- OAuth access tokens
- refresh tokens
- SSO keys
- webhook signing secrets
- Authorization headers

If configuration is logged:

log only safe metadata.

---

# 11. Database Connection Logging

Database initialization should provide useful lifecycle information.

For example:

- beginning connection
- successful connection
- connection failure
- migration start
- migration completion

Do not log:

- database passwords
- full credentials
- connection strings containing secrets

---

# 12. Database Errors

Database errors should preserve the underlying error.

Avoid replacing useful errors with:

"database error"

unless the original error is also retained.

Use structured error logging.

---

# 13. Migration Logging

If migrations run during service startup:

log:

- migration start
- migration success
- migration failure

Do not log migration secrets.

---

# 14. gRPC Observability

Where gRPC is used:

provide visibility into RPC requests.

Relevant fields include:

- service
- RPC method
- status
- duration
- request ID
- trace ID

Do not log complete request payloads by default.

---

# 15. gRPC Errors

Record:

- RPC method
- gRPC status
- duration
- error information

Do not expose secrets or sensitive request payloads.

---

# 16. gRPC Interceptors

If the architecture already uses gRPC interceptors:

integrate observability there.

Do not create duplicate middleware for the same responsibility.

Use the existing interceptor architecture where possible.

---

# 17. HTTP Gateway Observability

If the HTTP gateway exists:

provide request visibility.

Useful fields:

- HTTP method
- path
- status code
- duration
- request ID
- trace ID

Do not log:

- Authorization headers
- access tokens
- API keys
- sensitive request bodies

---

# 18. HTTP Error Logging

Failed HTTP requests should produce useful logs.

At minimum:

- method
- path
- status
- duration
- error where available

Avoid logging full request bodies.

---

# 19. Health Endpoints

Determine whether the platform already provides health endpoints.

If they exist:

document them.

If health endpoints are required by the platform design and do not exist:

implement the minimum required health mechanism.

Do not create a complex health framework.

---

# 20. Liveness

A liveness check answers:

"Is the service process alive?"

It should not necessarily depend on every external dependency being healthy.

Do not make liveness fail merely because PostgreSQL is temporarily unavailable unless the architecture explicitly requires that behavior.

---

# 21. Readiness

A readiness check answers:

"Can this service currently accept traffic?"

Where appropriate, readiness may consider:

- database availability
- required dependencies
- service initialization

Use the project's actual dependency model.

---

# 22. Health Check Separation

Do not confuse:

liveness

with:

readiness.

Document the difference.

---

# 23. Render Health Checks

Use:

docs/platform-render-review.md

to determine how Render health checks are configured.

Ensure observability does not create conflicting health endpoints.

If Render expects a specific endpoint:

use that endpoint.

---

# 24. Health Check Logging

Do not generate excessive logs for health probes.

Health checks may be called frequently.

Avoid producing an INFO log for every probe unless required.

---

# 25. Request IDs

Determine whether the platform already has request/correlation IDs.

If it does:

preserve the implementation.

If it does not:

introduce a minimal request ID mechanism where useful.

---

# 26. Request ID Requirements

A request ID should:

- be generated when absent
- be propagated where appropriate
- be available to logs
- be returned to clients where appropriate

Do not expose internal secrets as request IDs.

---

# 27. gRPC Request ID Propagation

Where practical:

propagate request IDs through gRPC metadata.

Do not invent a proprietary protocol.

Use standard gRPC metadata mechanisms.

---

# 28. HTTP Request ID Propagation

Where appropriate:

use a standard HTTP header.

If the project already establishes a request ID header:

use it.

Do not introduce a second competing header.

---

# 29. Trace IDs

Determine whether distributed tracing already exists.

If tracing already exists:

preserve and document it.

If the approved architecture requires tracing and no implementation exists:

implement the minimum viable tracing integration.

Do not introduce tracing infrastructure without a clear architectural reason.

---

# 30. OpenTelemetry

If OpenTelemetry is already part of the repository:

use the existing implementation.

If OpenTelemetry is not currently present:

do not automatically introduce it merely because it is popular.

Only introduce it if the architecture/repository documentation establishes it as required.

---

# 31. Trace Propagation

If tracing is implemented:

ensure trace context can propagate across:

- HTTP
- gRPC
- service boundaries

where applicable.

Do not modify protobuf contracts solely to carry tracing data.

Use transport metadata/context mechanisms.

---

# 32. Span Naming

If tracing is implemented:

use meaningful names.

Examples:

- grpc/<service>/<method>
- http/<method>/<route>
- db/<operation>

Use the conventions already established by the project if they exist.

---

# 33. Metrics

Determine whether metrics already exist.

Look for:

- Prometheus
- OpenTelemetry metrics
- application counters
- request duration metrics
- database metrics

Do not assume metrics are required simply because observability can include them.

---

# 34. Metrics Scope

If metrics are required:

focus on high-value operational metrics.

Examples:

- request count
- error count
- request duration
- active requests
- database failures
- external provider failures

Do not instrument every function.

---

# 35. Metric Naming

If metrics are introduced:

use consistent names.

Prefer a predictable namespace such as:

rvpay_

only if consistent with the existing project conventions.

Do not create conflicting metric namespaces.

---

# 36. Cardinality

Avoid high-cardinality metric labels.

Never use:

- request ID
- user ID
- transaction ID
- OAuth token
- arbitrary URL
- raw error message

as metric labels.

---

# 37. Sensitive Data

Observability must not become a data-leak mechanism.

Never place sensitive data into:

- logs
- metric labels
- traces
- span attributes

unless explicitly required and approved.

---

# 38. Payment Data

Be especially careful with financial data.

Do not log:

- full account numbers
- authentication credentials
- payment tokens
- provider secrets
- access tokens
- refresh tokens

Use identifiers only when the architecture permits them.

---

# 39. PII

Do not log unnecessary personally identifiable information.

Avoid:

- full phone numbers
- email addresses
- identity documents
- addresses

unless required for operational debugging.

Prefer internal identifiers.

---

# 40. Error Correlation

A production error should be traceable through:

request ID

and/or:

trace ID

when those mechanisms exist.

---

# 41. Error Wrapping

Preserve error context.

Do not discard errors simply to produce a cleaner log message.

Follow the existing Go error-handling conventions.

---

# 42. Panic Recovery

Determine whether the server already uses recovery middleware/interceptors.

If it does:

integrate logging with the existing recovery mechanism.

Do not create a second recovery system.

---

# 43. Panic Logs

Recovered panics should include:

- service
- operation
- error/panic information
- request ID
- trace ID where available

Do not expose sensitive request information.

---

# 44. Graceful Shutdown

Services should log shutdown lifecycle events.

For example:

- shutdown initiated
- server stopped
- database pool closed

Do not log shutdown as an error when it is a normal termination.

---

# 45. Signal Handling

Follow the existing:

context

and:

signal handling

architecture.

Do not replace the server lifecycle implementation merely for logging.

---

# 46. External Provider Calls

For Clients and Transactions provider integrations:

make failures observable.

Useful information:

- provider
- operation
- status
- duration
- error category
- request ID
- trace ID

Never log:

- Authorization headers
- OAuth tokens
- API keys
- full sensitive payloads

---

# 47. Provider Response Logging

Do not automatically log full external provider responses.

Prefer:

- HTTP status
- provider error code
- internal error classification
- duration

where available.

---

# 48. Database Query Logging

Do not automatically log every SQL query.

If query logging is required:

ensure it does not expose sensitive values.

Prefer operation names over raw SQL.

---

# 49. SQLC

Do not modify generated sqlc files.

If observability needs repository-level logging:

integrate it in the repository wrapper or established abstraction.

---

# 50. Repository Logging

Repository methods should not each independently emit duplicate logs for the same failure.

Establish clear ownership:

transport layer:

request lifecycle

service layer:

business operation failures

repository layer:

database-specific failures

---

# 51. Duplicate Logging

Avoid this pattern:

repository logs error

service logs same error

handler logs same error

resulting in three identical production errors.

Choose the appropriate logging boundary.

---

# 52. Log Ownership

Use this general rule:

Transport:

request/response lifecycle.

Service:

business operation context.

Repository:

database/provider technical context when useful.

Application entry point:

startup/shutdown lifecycle.

---

# 53. Context Propagation

Use Go context correctly.

Do not introduce global mutable state to carry request information.

---

# 54. Logger Context

Where the existing logger supports contextual fields:

attach:

- request ID
- trace ID
- service
- operation

to the logger/context.

Do not mutate global logger state per request.

---

# 55. Goroutines

If background goroutines are used:

ensure they have:

- lifecycle ownership
- cancellation
- useful error logging
- graceful shutdown behavior

Do not create background goroutines solely for logging.

---

# 56. Background Jobs

If the repository contains background processing:

instrument:

- start
- completion
- failure
- duration

at a useful level.

Do not log every internal step.

---

# 57. Logging Format

Preserve the existing production logging format.

If the project already produces structured JSON:

continue using it.

Do not switch to human-formatted logs solely for local readability.

---

# 58. Local Development

Local logs should remain readable.

Do not add production-only complexity that makes local development unnecessarily difficult.

---

# 59. Production

Production logs should be:

- structured
- timestamped
- machine-readable
- useful for Render/container logs

where supported by the existing stack.

---

# 60. Render Logs

Use Render's container/service logs as the expected production log destination if that is how the platform is configured.

Do not introduce a log-file architecture inside containers.

---

# 61. Container Logging

Applications running in Docker should generally write operational logs to:

stdout

and:

stderr

according to the project's existing logging implementation.

Do not create persistent application log files inside containers unless explicitly required.

---

# 62. Log Rotation

Do not implement custom log rotation inside application containers.

The deployment platform should own log retention where applicable.

---

# 63. Configuration Errors

Configuration failures at startup must be observable.

Log:

- configuration loading failure
- relevant configuration category

Do not log:

- secret values
- raw secret-bearing configuration

---

# 64. Dependency Failures

If startup fails because a dependency cannot be reached:

the error should clearly identify the dependency.

Example categories:

- PostgreSQL
- provider API
- configuration
- migration

Do not expose credentials.

---

# 65. Retry Logging

If retry logic exists:

avoid logging every retry at ERROR.

Use:

- DEBUG for individual retry attempts where appropriate
- WARN when retries become significant
- ERROR when the operation ultimately fails

Follow existing conventions.

---

# 66. Timeouts

Timeout failures should be distinguishable from generic errors.

Where useful:

include operation and dependency.

Do not expose internal implementation details unnecessarily.

---

# 67. Observability Package

If a shared observability package is appropriate:

place it according to:

docs/repository-layout.md

and:

agents/project-context.md.

Do not create random top-level packages.

---

# 68. Shared Package Rules

Before creating a shared package:

verify that the functionality is genuinely shared.

Do not create:

shared/logger.go

shared/observability.go

shared/utils.go

merely because multiple files could theoretically use them.

---

# 69. Package Ownership

Observability utilities should have one clear owner.

Avoid circular dependencies.

---

# 70. Service Independence

Services must remain independently runnable.

Observability support must not make Clients dependent on Transactions or vice versa unless explicitly required by the architecture.

---

# 71. Configuration Independence

Each service must be able to initialize its own observability configuration.

Do not create hidden cross-service configuration dependencies.

---

# 72. Testing Observability

Add tests only where they provide meaningful value.

Potential tests:

- request ID generation
- request ID propagation
- log level configuration
- health endpoint behavior
- middleware behavior
- error classification

Do not test the logging library itself.

---

# 73. Test Isolation

Observability tests must not require:

- production credentials
- Render
- real external APIs
- production PostgreSQL

unless explicitly defined as integration tests.

---

# 74. Health Tests

If health endpoints are added:

test:

- successful liveness
- readiness behavior
- dependency failure behavior

where applicable.

---

# 75. Metrics Tests

If metrics are added:

test only the project's own metric behavior.

Do not test third-party metrics libraries.

---

# 76. Trace Tests

If tracing is implemented:

test propagation where meaningful.

Do not build an elaborate tracing test framework.

---

# 77. No Test Pollution

Tests must not:

- modify developer environment variables permanently
- write logs to repository directories
- require external services unnecessarily

---

# 78. Documentation

Update documentation to describe the observability behavior actually implemented.

At minimum, update:

README.md

only if a developer needs to know how to use or configure observability.

Create:

docs/platform-observability-review.md

for detailed implementation notes.

---

# 79. Documentation Separation

README.md should explain:

how developers observe the application.

The review document should explain:

what was implemented and why.

Do not dump implementation details into README.md.

---

# 80. Observability Review Document

Create:

docs/platform-observability-review.md

Use exactly this structure:

# Platform Observability Review

## 1. Objective

Describe the observability goals.

## 2. Required Documentation

List every required document read.

## 3. Existing Observability

Describe:

- existing logger
- logging configuration
- health checks
- request IDs
- tracing
- metrics
- error handling

## 4. Observability Changes

Describe every implementation change.

## 5. Logging

Document:

- log format
- levels
- fields
- lifecycle events
- error behavior

## 6. Request Correlation

Document:

- request IDs
- trace IDs
- propagation

## 7. Health

Document:

- liveness
- readiness
- dependency checks

## 8. Metrics

If implemented, document:

- metrics
- labels
- collection mechanism

If not implemented:

state why.

## 9. Tracing

If implemented, document:

- tracing system
- propagation
- span naming

If not implemented:

state why.

## 10. Sensitive Data

Document how observability avoids leaking:

- credentials
- tokens
- secrets
- financial information
- unnecessary PII

## 11. Render

Document how observability works in Render.

## 12. Docker

Document container logging behavior.

## 13. Tests

Document tests added or modified.

## 14. Findings

| ID | Severity | Area | Finding | Resolution |
|---|---|---|---|---|

## 15. Documentation Changes

List modified documentation.

## 16. Documentation Check

Record the final documentation verification.

## 17. Final Status

Use exactly one:

PASS

PASS WITH FOLLOW-UP

BLOCKED

---

# 81. No Security Implementation

Do NOT implement:

- authentication
- authorization
- encryption
- secret rotation
- OAuth security changes
- webhook security changes

Those belong to Agent 10.

You may ensure observability does not leak secrets.

---

# 82. No Performance Implementation

Do NOT:

- optimize database queries
- add caching
- benchmark services
- tune connection pools
- rewrite handlers
- optimize protobuf
- optimize Docker images

Those belong to Agent 11.

---

# 83. No Documentation Overhaul

Agent 08 owns the main documentation work.

Only update README.md when observability introduces developer-visible behavior/configuration.

Do not redo Agent 08's documentation work.

---

# 84. No Deployment Redesign

Do not redesign:

- Render
- Docker
- CI/CD
- deployment topology

Observability must fit the existing platform.

---

# 85. No Database Redesign

Do not add observability tables to PostgreSQL unless explicitly required by the architecture.

Observability should not require application database persistence by default.

---

# 86. No Generated Code Modification

Do not manually modify:

- generated protobuf files
- generated sqlc files

---

# 87. No Protobuf Changes

Do not modify protobuf contracts to support observability.

Use:

context

and:

transport metadata

where applicable.

---

# 88. Error Message Preservation

Do not change public error semantics merely to improve logging.

Observability should add context without breaking service behavior.

---

# 89. Backward Compatibility

Existing service behavior must remain unchanged unless a health/observability endpoint is being added.

---

# 90. Build Verification

After implementation:

run the repository's normal build command.

Use commands documented by:

README.md

and:

agents/project-context.md.

Do not invent build commands.

---

# 91. Test Verification

Run the appropriate test suite.

At minimum:

the tests affected by observability changes.

If the repository's standard test command is documented:

use it.

---

# 92. Formatting

Run the repository's established formatting workflow.

Do not introduce a new formatter.

---

# 93. Static Analysis

Run the project's established static analysis if documented.

Do not install unrelated analysis tools.

---

# 94. Git Diff Review

Before finishing:

run:

git status --short

Then:

git diff --stat

Then:

git diff

Review all changes made by this agent.

---

# 95. Unexpected Changes

If unrelated files were modified:

do not overwrite them.

Determine whether they were pre-existing.

Do not reset the repository.

---

# 96. Generated Files

If commands modify generated files:

determine whether those modifications are expected.

Do not blindly revert generated files.

---

# 97. Final Documentation Check

Verify all required documents still exist:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md
- docs/platform-repository-audit.md
- docs/platform-protobuf-generation-review.md
- docs/platform-http-gateway-review.md
- docs/platform-common-packages-review.md
- docs/platform-ci-cd-review.md
- docs/platform-docker-review.md
- docs/platform-render-review.md
- docs/platform-documentation-review.md
- docs/platform-observability-review.md

Record the result in:

docs/platform-observability-review.md

---

# 98. Completion Checklist

Before stopping:

- [ ] All required documents were read.
- [ ] README.md was read.
- [ ] agents/project-context.md was followed.
- [ ] Repository exploration was restricted.
- [ ] Deep folders were not recursively inspected.
- [ ] third_party/googleapis was not unnecessarily explored.
- [ ] Existing logging was audited.
- [ ] Existing logger conventions were preserved.
- [ ] Startup logging is useful.
- [ ] Shutdown logging is useful.
- [ ] Database failures are observable.
- [ ] gRPC failures are observable.
- [ ] HTTP gateway failures are observable where applicable.
- [ ] Provider failures are observable where applicable.
- [ ] Request correlation is preserved or implemented where required.
- [ ] Health behavior is documented/implemented where required.
- [ ] Health checks do not produce excessive logs.
- [ ] Metrics were assessed.
- [ ] Tracing was assessed.
- [ ] Sensitive information is not logged.
- [ ] No credentials are exposed.
- [ ] No tokens are exposed.
- [ ] No unnecessary PII is logged.
- [ ] Container logs use stdout/stderr where appropriate.
- [ ] Render observability behavior is compatible.
- [ ] Tests were added where useful.
- [ ] Tests pass.
- [ ] Build passes.
- [ ] Formatting passes.
- [ ] No generated code was manually modified.
- [ ] No protobuf contracts were changed.
- [ ] No security architecture was changed.
- [ ] No performance optimization was performed.
- [ ] README was updated only where necessary.
- [ ] docs/platform-observability-review.md was created.
- [ ] Final documentation check was completed.
- [ ] git status was reviewed.
- [ ] git diff was reviewed.

---

# Final Stop Condition

STOP after:

1. reading all required documentation,
2. auditing existing observability,
3. implementing only the required observability functionality,
4. preserving existing logging conventions,
5. ensuring request/error correlation where appropriate,
6. ensuring health behavior is correct where required,
7. ensuring observability does not leak sensitive information,
8. adding meaningful tests where required,
9. updating only necessary documentation,
10. creating docs/platform-observability-review.md,
11. completing the documentation check,
12. running the appropriate verification commands,
13. reviewing git status,
14. reviewing git diff.

Do NOT proceed to:

- security implementation
- performance optimization
- CI/CD redesign
- Render redesign
- Docker redesign
- protobuf redesign
- database redesign
- service redesign
- unrelated refactoring

STOP.