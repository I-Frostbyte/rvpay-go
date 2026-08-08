# Agent 11 — Performance

## Objective

Audit and improve the performance characteristics of the RVPay platform.

The objective is to identify and address concrete performance problems across:

* HTTP gateway
* gRPC services
* database access
* SQL queries
* connection pools
* external provider calls
* serialization
* application startup
* container/runtime configuration
* observability overhead

Performance improvements must remain consistent with the architecture established by:

* the foundation documents
* the Clients service
* the Transactions service
* Platform Agents 01–10

Do NOT redesign the architecture.

Do NOT introduce unnecessary infrastructure.

Do NOT replace working implementations simply because another implementation may be theoretically faster.

Do NOT perform speculative optimization.

Do NOT rewrite entire services.

Every performance change must have a concrete reason and must preserve existing behavior.

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

Also inspect only the specific source files referenced by these documents when performance verification requires them.

---

# Documentation Check

Before starting:

verify that all required documents exist.

If any required document is missing:

STOP.

Do not recreate missing documentation.

At the end of the task:

perform the documentation check again.

Create:

docs/platform-performance-review.md

and record the final result.

---

# Repository Exploration Rules

## IMPORTANT

Do NOT perform an unrestricted repository-wide search.

Use:

README.md

as the primary repository map.

Use:

docs/repository-layout.md

as the authority for repository structure.

Use:

agents/project-context.md

as the authority for coding and package conventions.

Use:

docs/platform-repository-audit.md

to understand what has already been inspected.

Use:

docs/platform-observability-review.md

to understand the existing performance/observability instrumentation.

Use:

docs/platform-security-review.md

to ensure performance changes do not weaken security.

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

DO NOT inspect:

third_party/googleapis/

Do not spend time reading generated protobuf internals.

Do not inspect generated code unless a specific performance problem requires verifying generated behavior.

---

# 1. Performance Philosophy

Performance work must answer:

1. What is slow?
2. Where is it slow?
3. Why is it slow?
4. How was that determined?
5. What is the smallest appropriate improvement?
6. How will the improvement be verified?

Do not optimize based solely on intuition.

---

# 2. Establish the Performance Baseline

Before changing code:

identify the major performance-sensitive paths.

At minimum inspect:

* HTTP requests
* gRPC requests
* database operations
* external API requests
* startup
* health checks
* webhook processing
* OAuth operations

Document what is known and what is unknown.

Do not invent benchmark numbers.

---

# 3. Existing Performance Characteristics

Review:

docs/platform-repository-audit.md

and identify:

* service boundaries
* database boundaries
* connection handling
* HTTP clients
* gRPC configuration
* repository patterns
* generated code
* deployment topology

Do not re-audit unrelated architecture.

---

# 4. Existing Observability

Review:

docs/platform-observability-review.md

Determine whether the application already provides:

* request duration
* database timing
* provider call timing
* error counts
* request counts

Use existing instrumentation wherever possible.

Do not introduce duplicate metrics.

---

# 5. Performance Measurements

Where practical:

measure before changing code.

Useful measurements may include:

* benchmark duration
* query execution time
* request latency
* allocation count
* memory usage
* startup duration

If no baseline measurement is available:

document that fact.

Do not manufacture benchmark results.

---

# 6. Avoid Premature Optimization

Do not optimize code merely because it:

* looks verbose
* allocates occasionally
* contains a loop
* uses an abstraction
* uses a struct
* calls a function
* performs a database query

There must be a meaningful performance reason.

---

# 7. HTTP Gateway Performance

Review:

docs/platform-http-gateway-review.md

Inspect:

* middleware
* request parsing
* serialization
* connection handling
* timeouts
* routing
* request body handling

---

# 8. HTTP Connection Reuse

Verify that outbound HTTP clients are reused rather than recreated for every request.

Avoid patterns such as:

creating a new:

http.Client

for every request.

Use the existing application/client structure.

---

# 9. HTTP Timeouts

Verify external HTTP clients have appropriate timeouts.

Do not use:

http.Client

with no timeout for external provider calls.

Do not choose arbitrary extremely short timeouts that break legitimate requests.

---

# 10. HTTP Response Bodies

Ensure response bodies are closed correctly.

Where appropriate:

ensure connections can be reused by properly consuming/closing response bodies.

---

# 11. HTTP Request Allocation

Do not optimize request allocation unless profiling demonstrates it matters.

Favor correctness and readability.

---

# 12. Provider API Calls

Inspect external provider calls.

Look for:

* unnecessary duplicate requests
* sequential requests that could safely be combined
* repeated authentication calls
* unnecessary token exchanges
* repeated configuration lookups

Do not introduce concurrency unless the operations are independent and safe.

---

# 13. Provider Call Latency

External provider calls are usually network-bound.

Do not attempt to optimize them through micro-level Go changes.

Focus on:

* reducing unnecessary calls
* connection reuse
* sensible timeouts
* avoiding duplicate operations

---

# 14. Provider Retries

Review retry behavior.

Retries must not accidentally multiply provider traffic.

Do not introduce aggressive retries.

Where retries exist:

ensure they have appropriate limits.

---

# 15. Retry Storms

Avoid retry behavior that can produce:

request

→ retry

→ retry

→ retry

across multiple services.

Do not create cascading retry amplification.

---

# 16. gRPC Performance

Review the existing gRPC setup.

Focus on:

* connection reuse
* interceptor overhead
* serialization
* request size
* unnecessary conversions

Do not modify generated protobuf code.

---

# 17. gRPC Connections

If services communicate through gRPC:

reuse client connections/channels where appropriate.

Do not create a new gRPC connection for every request.

---

# 18. gRPC Interceptors

Inspect interceptors for:

* expensive logging
* repeated serialization
* unnecessary allocations
* excessive stack/error processing

Do not remove security or recovery interceptors simply for performance.

---

# 19. Protobuf

Use:

docs/protobuf-strategy.md

as the authority for protobuf design.

Do not redesign protobuf contracts for marginal performance gains.

---

# 20. Serialization

Look for unnecessary:

marshal

→ unmarshal

→ marshal

cycles.

Where possible:

avoid converting the same payload repeatedly between representations.

---

# 21. JSON

Do not replace standard encoding/json merely because another library claims higher benchmarks.

Only change JSON implementation if:

* profiling demonstrates a meaningful bottleneck
* compatibility is preserved
* the dependency is justified

---

# 22. Database Performance

Database performance is one of the highest-priority areas.

Inspect:

* connection pools
* query patterns
* indexes
* transactions
* query frequency
* result sizes

---

# 23. SQLC

The project uses SQLC.

Do not manually modify generated SQLC code.

Performance improvements belong in:

* SQL query files
* migrations
* repository usage
* service behavior

---

# 24. Query Performance

Look for:

* repeated queries
* unnecessary queries
* SELECT *
* missing filters
* unnecessarily large result sets
* queries executed inside loops

---

# 25. SELECT *

Avoid:

SELECT *

when only a subset of columns is required.

Only change existing queries where the returned data can safely be reduced.

---

# 26. N+1 Queries

Look specifically for:

for each item:

query database

patterns.

If an N+1 query exists:

determine whether it can safely be replaced by:

* a joined query
* a batched query
* an IN query
* a repository-level batch operation

Do not rewrite unrelated queries.

---

# 27. Query Batching

Where several independent database operations can safely be combined:

consider batching.

Do not batch operations that require different transactional semantics.

---

# 28. Database Indexes

Inspect existing migrations for indexes relevant to frequently queried fields.

Do not create indexes blindly.

Every index has:

* storage cost
* write cost
* maintenance cost

---

# 29. Index Candidates

Potential index candidates may include:

* merchant IDs
* provider IDs
* transaction IDs
* external IDs
* status fields
* timestamps
* webhook identifiers

Only add indexes when the query pattern justifies them.

---

# 30. Composite Indexes

Consider composite indexes only when actual query predicates commonly use the same column combination.

Do not create every possible combination.

---

# 31. Unique Constraints

Remember that unique constraints may already provide indexes.

Do not create duplicate indexes unnecessarily.

---

# 32. Query Result Size

Avoid loading unnecessarily large result sets.

Use:

* explicit columns
* appropriate filters
* pagination

where the existing API semantics support it.

---

# 33. Pagination

Review endpoints returning potentially large datasets.

Where pagination already exists:

verify reasonable limits.

Where pagination is required by the architecture:

implement it consistently with existing conventions.

Do not redesign APIs unnecessarily.

---

# 34. Pagination Limits

Do not allow clients to request unlimited result sets.

Use a reasonable maximum configured consistently with project conventions.

Do not invent arbitrary limits without documenting the reasoning.

---

# 35. Database Connection Pool

Inspect the pgxpool configuration.

Review:

* maximum connections
* minimum connections
* idle connections
* connection lifetime
* acquisition behavior

Do not change values blindly.

---

# 36. Connection Pool Sizing

Pool size depends on:

* Render instance resources
* database limits
* service count
* request concurrency
* query duration

Do not assume:

more connections = more performance.

---

# 37. Database Connection Exhaustion

Look for code that:

* fails to release rows
* holds transactions too long
* performs network calls inside database transactions
* leaks connections

Fix concrete leaks.

---

# 38. Transactions

Transactions should be kept as short as correctness permits.

Avoid:

BEGIN

→ database work

→ external HTTP request

→ more database work

→ COMMIT

unless the architecture explicitly requires it.

---

# 39. External Calls Inside Transactions

Avoid holding database connections while waiting for external providers.

If existing code does this:

determine whether the workflow can be safely restructured.

Do not break transaction consistency.

---

# 40. Database Rows

Ensure database rows are closed where required.

Do not retain large result sets unnecessarily.

---

# 41. Prepared Statements

Do not introduce manual prepared statement infrastructure unless there is a demonstrated need.

Use SQLC/pgx behavior already established by the project.

---

# 42. Database Round Trips

Reducing round trips is generally preferable to micro-optimizing Go code.

Look for:

query

→ query

→ query

patterns that can safely become:

one query.

---

# 43. Caching

Do NOT introduce a distributed cache in this agent.

Do not add:

* Redis
* Memcached
* external caching systems

unless already established by the architecture.

If caching would be beneficial:

document it as a future recommendation.

---

# 44. In-Memory Caching

Do not add in-memory caches merely to improve benchmark results.

Caching introduces:

* invalidation
* memory consumption
* consistency concerns

Only add it if the existing design explicitly requires it.

---

# 45. Database Caching

Do not rely on undocumented PostgreSQL caching behavior.

Optimize queries and indexes instead.

---

# 46. Concurrency

Go provides concurrency through:

goroutines

and:

channels

but concurrency is not automatically a performance improvement.

Only introduce concurrency when operations are genuinely independent.

---

# 47. Safe Concurrency

Before parallelizing operations:

verify:

* no shared mutable state
* no ordering dependency
* no transaction dependency
* no provider rate-limit violation
* no database overload

---

# 48. Goroutine Leaks

Look for goroutines that:

* never terminate
* block forever
* wait on channels that are never closed
* outlive their request unnecessarily

Fix concrete leaks.

---

# 49. Context Propagation

All request-bound work should use:

context.Context

where appropriate.

Do not use:

context.Background()

inside request processing merely to avoid passing context.

---

# 50. Context Cancellation

Ensure external/database operations can stop when the request is cancelled.

This prevents unnecessary work after the client disconnects.

---

# 51. Background Work

Do not introduce background workers merely for performance.

If background processing already exists:

verify that work has:

* lifecycle management
* cancellation
* bounded concurrency

---

# 52. Unbounded Concurrency

Avoid patterns that create unlimited goroutines based on incoming requests or records.

Use bounded concurrency where parallel work is genuinely necessary.

---

# 53. Memory Allocation

Do not optimize allocations without measurement.

Use benchmarks/profiling where practical.

---

# 54. Large Payloads

Inspect endpoints that may process large:

* webhook payloads
* transaction responses
* provider responses
* database results

Avoid unnecessary duplication of large byte slices or strings.

---

# 55. Request Bodies

Where appropriate:

limit request body sizes.

Do not impose limits that reject legitimate provider payloads.

---

# 56. Response Bodies

Avoid buffering entire external responses when streaming or bounded processing is sufficient.

Only change this where actual payload sizes justify it.

---

# 57. String Building

Do not perform broad micro-optimization such as replacing every string concatenation.

Only optimize allocation-heavy paths identified through profiling.

---

# 58. Reflection

Do not remove reflection merely because it exists.

Protobuf/gRPC frameworks may legitimately use reflection internally.

Do not modify generated code.

---

# 59. Logging Performance

Use:

docs/platform-observability-review.md

to inspect logging.

Look for:

* logging entire request bodies
* logging huge provider responses
* excessive debug logs in production
* expensive serialization performed solely for logs

---

# 60. Sensitive Logging

Security takes precedence over performance.

Do not reduce security logging merely because logging costs CPU.

Use:

docs/platform-security-review.md

to preserve security requirements.

---

# 61. Structured Logging

Use the existing structured logging system.

Do not introduce another logging library.

---

# 62. Log Levels

Production should not perform expensive debug logging unless enabled.

If the existing logger supports levels:

use them appropriately.

---

# 63. Tracing Overhead

Review tracing implementation.

Avoid:

* tracing every internal micro-operation
* unnecessarily huge attributes
* full request/response payload capture

Do not remove useful distributed tracing.

---

# 64. Metrics Overhead

Metrics should be cheap.

Avoid generating extremely high-cardinality labels such as:

* transaction IDs
* customer IDs
* authorization tokens
* arbitrary URLs

unless explicitly required.

---

# 65. Metric Cardinality

Use stable dimensions such as:

* service
* method
* status
* provider
* operation type

where appropriate.

Do not introduce unbounded label values.

---

# 66. Health Checks

Health checks should be cheap.

Do not perform expensive external provider calls on every health check.

Do not perform full database diagnostics unless required by the deployment architecture.

---

# 67. Readiness

If readiness requires database connectivity:

perform the minimum appropriate check.

Do not execute expensive queries.

---

# 68. Liveness

Liveness should generally indicate whether the process is alive.

Do not make liveness depend on external services unless explicitly required.

---

# 69. Startup Performance

Inspect service startup.

Look for:

* unnecessary database queries
* repeated configuration parsing
* expensive initialization
* unnecessary external API calls

---

# 70. Configuration

Configuration should be loaded once during startup where practical.

Do not repeatedly parse environment variables during every request.

---

# 71. Provider Configuration

Provider configuration should be initialized once when possible.

Do not recreate clients unnecessarily.

---

# 72. Database Initialization

Database pools should be created once per service process.

Do not create a pool for each request.

---

# 73. gRPC Server Initialization

The gRPC server should be initialized once.

Do not repeatedly construct server infrastructure during requests.

---

# 74. HTTP Server Initialization

HTTP server/router initialization belongs to startup.

Do not dynamically construct routing structures per request.

---

# 75. Docker Startup

Review:

docs/platform-docker-review.md

Avoid adding unnecessary startup work.

Do not add arbitrary shell scripts.

---

# 76. Container Resources

Use:

docs/platform-render-review.md

to understand deployment resources.

Do not assume the container has unlimited:

* CPU
* memory
* connections

---

# 77. Render Scaling

Do not automatically increase Render instance size.

If the current resource allocation is clearly insufficient:

document the evidence.

Do not make infrastructure scaling changes unless explicitly required.

---

# 78. Horizontal Scaling

Ensure application code is compatible with multiple instances where the architecture expects horizontal scaling.

Avoid process-local state that must be shared across instances.

---

# 79. Process-Local State

Inspect for:

* global mutable maps
* in-memory sessions
* process-local queues
* singleton state

Do not automatically remove all global state.

Only identify state that breaks the intended deployment model.

---

# 80. Distributed State

Do not introduce Redis or another distributed state system in this agent.

Document future requirements if necessary.

---

# 81. Database as Source of Truth

Where the architecture uses PostgreSQL as the source of truth:

do not create competing state stores.

---

# 82. External Provider Rate Limits

Performance improvements must respect provider rate limits.

Faster internal processing is not useful if it causes provider throttling.

---

# 83. Provider Concurrency

Do not parallelize provider requests without confirming:

* provider supports it
* requests are independent
* rate limits are respected

---

# 84. Backpressure

If the application can receive work faster than it can process it:

avoid creating unbounded queues or goroutines.

Document the existing backpressure behavior.

---

# 85. Timeouts

Use appropriate timeouts for:

* database operations
* provider HTTP requests
* gRPC calls

Do not use excessively large timeout values that hold resources indefinitely.

---

# 86. Retries and Timeouts

Retries multiply work.

For every retry:

consider:

timeout

×

attempts

and:

concurrent requests

Do not create retry storms.

---

# 87. Error Handling

Do not repeatedly retry permanent failures.

Examples:

* invalid credentials
* invalid request
* invalid payer information
* invalid provider configuration

---

# 88. Idempotency

Performance changes must preserve idempotency.

This is especially important for:

* deposits
* payouts
* webhooks
* OAuth installation

Do not optimize away safeguards against duplicate operations.

---

# 89. Financial Operations

Never sacrifice correctness for speed.

For:

* deposits
* payouts
* balances
* transaction state

correctness takes precedence over latency.

---

# 90. Database Consistency

Do not remove transactions merely because they reduce latency.

Only modify transaction boundaries when correctness is preserved.

---

# 91. Benchmarking

Where code-level optimization is proposed:

create a benchmark when practical.

Use standard Go benchmarks:

BenchmarkXxx

Do not create benchmarks for trivial code with no meaningful performance impact.

---

# 92. Benchmark Scope

Benchmarks should focus on:

* database-independent logic
* serialization
* transformations
* provider payload processing
* computationally expensive code

Do not benchmark external network calls as ordinary unit benchmarks.

---

# 93. Benchmark Results

Record meaningful before/after results where available.

Include:

* operation
* baseline
* optimized
* change
* environment

Do not claim results that were not measured.

---

# 94. Profiling

If performance problems are significant:

use appropriate Go profiling tools.

Possible tools include:

* pprof
* go test -bench
* go test -benchmem

Only use them when justified.

---

# 95. Profiling Restrictions

Do not leave profiling/debug endpoints enabled in production unless explicitly secured.

Security requirements from Agent 10 remain authoritative.

---

# 96. Query Analysis

If query performance is a concern:

use PostgreSQL query analysis where available.

Do not run destructive SQL.

Do not modify production data.

---

# 97. EXPLAIN

Where appropriate:

use:

EXPLAIN

or:

EXPLAIN ANALYZE

against controlled/test data.

Do not run expensive EXPLAIN ANALYZE operations against production merely for this task.

---

# 98. Migration Performance

If adding an index:

consider the cost of creating it.

Do not perform dangerous table-wide migrations without appropriate safeguards.

---

# 99. Large Tables

Do not assume tables are small.

Consider the impact of:

* indexes
* migrations
* backfills
* full-table scans

---

# 100. Query Timeouts

Do not allow pathological queries to run indefinitely.

Use the project's established database/context mechanisms.

---

# 101. Connection Lifetime

Do not arbitrarily reduce connection lifetime.

Connection churn can reduce performance.

---

# 102. Memory Limits

Avoid loading entire large datasets into memory.

Prefer:

* pagination
* bounded processing
* streaming

where appropriate.

---

# 103. Garbage Collection

Do not modify Go GC settings without evidence.

Do not set:

GOGC

or runtime tuning parameters arbitrarily.

---

# 104. CPU Optimization

Do not perform low-level CPU optimization unless profiling identifies a real bottleneck.

---

# 105. Algorithmic Complexity

If profiling identifies an algorithmic bottleneck:

prefer a simpler complexity improvement over micro-optimizations.

For example:

O(n²)

to:

O(n)

may justify a change.

But verify correctness.

---

# 106. Data Structures

Do not replace data structures purely for style.

Only change them when:

* performance impact is demonstrated
* semantics remain unchanged
* code remains maintainable

---

# 107. Concurrency Safety

Any concurrency optimization must be checked for:

* race conditions
* deadlocks
* goroutine leaks

---

# 108. Race Detection

If concurrency code is changed:

run the appropriate race-enabled tests where practical.

Use the project's supported Go testing workflow.

---

# 109. Tests

All performance changes must preserve existing tests.

Do not remove tests to make performance tests pass.

---

# 110. Performance Tests

Add tests/benchmarks only where they provide meaningful regression protection.

Do not create dozens of synthetic benchmarks.

---

# 111. Regression Protection

For every significant optimization:

ensure there is a way to detect regression.

This may be:

* unit test
* benchmark
* query review
* documented measurement

---

# 112. Generated Code

Do not manually modify:

* generated protobuf files
* generated SQLC files

---

# 113. Protobuf Generation

If protobuf changes are absolutely required:

modify source protobuf definitions.

Use the established generation workflow.

Do not edit generated output directly.

---

# 114. SQLC Generation

If SQL changes are required:

modify the source SQL.

Regenerate SQLC using the existing Makefile/workflow.

Do not manually modify generated SQLC output.

---

# 115. Database Migrations

If indexes or query-supporting schema changes are required:

create proper migrations.

Do not modify existing applied migrations.

---

# 116. Migration Naming

Follow the project's existing migration naming convention.

Do not invent a second naming style.

---

# 117. Common Packages

Use existing common packages where applicable.

Do not create duplicate utility packages for:

* HTTP
* logging
* database
* configuration
* metrics

---

# 118. Dependency Discipline

Do not add dependencies unless:

* the functionality is genuinely required
* the standard library is insufficient
* the dependency is compatible with the project

---

# 119. No Premature Infrastructure

Do not introduce:

* Redis
* Kafka
* RabbitMQ
* NATS
* Elasticsearch
* distributed cache
* service mesh

for performance purposes.

If future scale may require one:

document it as a future recommendation.

---

# 120. No Architectural Rewrite

Do not:

* split services
* merge services
* redesign protobufs
* redesign database ownership
* replace gRPC
* replace PostgreSQL
* replace Render

---

# 121. Compatibility

Performance improvements must preserve:

* API behavior
* protobuf behavior
* database semantics
* transaction semantics
* OAuth behavior
* webhook behavior
* provider behavior

---

# 122. Security Compatibility

Review:

docs/platform-security-review.md

after performance changes.

Ensure no optimization:

* removes authentication
* bypasses authorization
* exposes secrets
* weakens TLS
* disables validation
* exposes debug endpoints

---

# 123. Observability Compatibility

Review:

docs/platform-observability-review.md

after changes.

Ensure useful:

* metrics
* logs
* traces
* request timings

remain available.

---

# 124. Documentation

Create:

docs/platform-performance-review.md

Document:

* baseline
* findings
* changes
* measurements
* remaining concerns

Do not rewrite:

README.md

unless a performance-related configuration change makes existing instructions incorrect.

---

# 125. Performance Review Document

Use exactly this structure:

# Platform Performance Review

## 1. Objective

Describe the performance objectives.

## 2. Required Documentation

List every document read.

## 3. Performance Baseline

Describe known measurements.

If none exist:

state that explicitly.

## 4. Architecture Reviewed

Document:

* HTTP gateway
* gRPC
* services
* PostgreSQL
* providers
* Render
* containers

## 5. Database Review

Document:

* queries
* indexes
* connection pool
* transactions
* round trips

## 6. HTTP Review

Document:

* clients
* connection reuse
* timeouts
* payload handling

## 7. gRPC Review

Document:

* connections
* interceptors
* serialization
* request handling

## 8. Provider Review

Document:

* unnecessary calls
* retries
* connection reuse
* rate limits

## 9. Concurrency Review

Document:

* goroutines
* cancellation
* bounded concurrency
* leaks

## 10. Memory Review

Document meaningful allocation/memory findings.

## 11. Observability Review

Document performance overhead from:

* logging
* metrics
* tracing

## 12. Container/Runtime Review

Document:

* startup
* resources
* process behavior

## 13. Changes Implemented

List every change.

## 14. Measurements

| Area | Before | After | Change | Measurement Method |
| ---- | -----: | ----: | -----: | ------------------ |

If no reliable measurement exists:

state:

"Not measured."

Do not invent numbers.

## 15. Findings

| ID | Severity | Area | Finding | Resolution |
| -- | -------- | ---- | ------- | ---------- |

Severity:

* HIGH
* MEDIUM
* LOW
* INFO

## 16. Remaining Performance Risks

Document issues that remain.

## 17. Future Recommendations

Only list recommendations that are outside this agent's scope.

Examples:

* caching
* horizontal scaling
* queue-based processing
* read replicas

Do not implement them here unless explicitly required.

## 18. Tests and Verification

Document:

* tests
* benchmarks
* profiling
* query analysis
* build verification

## 19. Documentation Changes

List documentation changes.

## 20. Documentation Check

Record the final documentation verification.

## 21. Final Status

Use exactly one:

PASS

PASS WITH FOLLOW-UP

BLOCKED

---

# 126. Performance Severity

Use:

HIGH

for a performance problem that materially threatens production reliability or causes severe resource exhaustion.

MEDIUM

for a meaningful performance bottleneck.

LOW

for a minor optimization opportunity.

INFO

for observations.

Do not inflate severity.

---

# 127. Performance Claims

Do not claim:

"high performance"

"production ready"

"optimized"

"zero bottlenecks"

without supporting evidence.

---

# 128. No Fake Benchmarks

Never invent:

* latency
* throughput
* memory
* CPU
* query duration
* benchmark results

If something was not measured:

say:

"Not measured."

---

# 129. Build Verification

After implementation:

run the repository's documented build process.

Use commands from:

README.md

and:

agents/project-context.md.

Do not invent build commands.

---

# 130. Test Verification

Run the relevant tests.

If the repository has a standard test command:

use it.

---

# 131. Benchmark Verification

If benchmarks were created:

run them.

Record the actual results.

---

# 132. Race Verification

If concurrency was changed:

run race-enabled testing where practical.

---

# 133. Formatting

Run the established formatting workflow.

Do not introduce a new formatting tool.

---

# 134. Static Analysis

Run existing project static analysis if documented.

Do not install unrelated tools.

---

# 135. Git Status

Before finishing:

run:

git status --short

---

# 136. Git Diff

Then run:

git diff --stat

and:

git diff

Review every change made by this agent.

---

# 137. Unexpected Changes

If unrelated files are already modified:

do not overwrite them.

Do not reset the repository.

---

# 138. Final Performance Review

Before stopping, verify:

* database queries
* database pool
* indexes
* transactions
* HTTP clients
* gRPC clients
* provider calls
* retries
* concurrency
* goroutines
* context cancellation
* memory
* serialization
* logging
* metrics
* tracing
* health checks
* startup
* Docker
* Render

---

# 139. Documentation Check

Verify that all required documents still exist:

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

Record the result in:

docs/platform-performance-review.md

---

# 140. Completion Checklist

Before stopping:

* [ ] All required documents were read.
* [ ] README.md was read.
* [ ] agents/project-context.md was followed.
* [ ] Repository exploration was restricted.
* [ ] Deep folders were not recursively inspected.
* [ ] third_party/googleapis was not unnecessarily explored.
* [ ] Existing performance characteristics were audited.
* [ ] A baseline was established where possible.
* [ ] No benchmark results were invented.
* [ ] HTTP performance was reviewed.
* [ ] gRPC performance was reviewed.
* [ ] Database performance was reviewed.
* [ ] SQLC queries were reviewed.
* [ ] Indexes were reviewed.
* [ ] Database connection pooling was reviewed.
* [ ] Transaction duration was reviewed.
* [ ] External provider calls were reviewed.
* [ ] HTTP connection reuse was reviewed.
* [ ] Retry behavior was reviewed.
* [ ] Timeout behavior was reviewed.
* [ ] Concurrency was reviewed.
* [ ] Goroutine leaks were considered.
* [ ] Context cancellation was reviewed.
* [ ] Memory behavior was reviewed where justified.
* [ ] Logging overhead was reviewed.
* [ ] Metrics overhead was reviewed.
* [ ] Tracing overhead was reviewed.
* [ ] Health checks were reviewed.
* [ ] Startup behavior was reviewed.
* [ ] Docker/runtime behavior was reviewed.
* [ ] Render resource assumptions were reviewed.
* [ ] No unnecessary infrastructure was introduced.
* [ ] No cache was introduced unnecessarily.
* [ ] No architectural rewrite was performed.
* [ ] Security controls remain intact.
* [ ] Observability remains intact.
* [ ] Generated code was not manually modified.
* [ ] SQLC generated code was not manually modified.
* [ ] Existing migrations were not modified.
* [ ] Appropriate migrations were created where required.
* [ ] Tests were run.
* [ ] Benchmarks were run where created.
* [ ] Race testing was performed where appropriate.
* [ ] Formatting was completed.
* [ ] Static analysis was completed where applicable.
* [ ] docs/platform-performance-review.md was created.
* [ ] Documentation check was completed.
* [ ] git status was reviewed.
* [ ] git diff was reviewed.

---

# Final Stop Condition

STOP after:

1. reading all required documentation,
2. establishing the current performance baseline,
3. identifying concrete performance issues,
4. implementing only justified performance improvements,
5. verifying database performance,
6. verifying HTTP/gRPC performance,
7. verifying provider-call behavior,
8. verifying concurrency,
9. verifying memory behavior where relevant,
10. verifying observability overhead,
11. verifying Docker/runtime behavior,
12. preserving security controls,
13. preserving existing architecture,
14. adding meaningful benchmarks/tests where appropriate,
15. creating docs/platform-performance-review.md,
16. completing the documentation check,
17. running the appropriate verification commands,
18. reviewing git status,
19. reviewing git diff.

Do NOT proceed to:

* platform redesign
* new caching infrastructure
* Redis
* Kafka
* RabbitMQ
* service decomposition
* database replacement
* protobuf redesign
* gRPC replacement
* Render architecture redesign
* broad dependency upgrades
* unrelated refactoring

STOP.