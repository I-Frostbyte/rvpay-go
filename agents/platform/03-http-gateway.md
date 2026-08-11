# Agent 03 — HTTP Gateway

## Objective

Implement and validate the HTTP gateway layer for the new RVPay platform architecture.

The gateway must expose the documented HTTP API for the protobuf-defined services while preserving the existing gRPC service architecture.

The gateway must:

- use the protobuf contracts as the source of truth
- use grpc-gateway where specified by the protobuf strategy
- route HTTP requests to the appropriate gRPC services
- preserve the service boundaries defined by the architecture
- use the generated gateway code from the protobuf generation pipeline
- avoid duplicating business logic
- avoid creating HTTP-specific business implementations
- fit the existing repository structure
- be compatible with the Clients and Transactions services
- remain suitable for deployment behind the platform's eventual infrastructure

This agent is responsible for the HTTP gateway implementation and wiring.

It is NOT responsible for:

- business logic
- database logic
- repository implementation
- OAuth implementation
- webhook business processing
- CI/CD
- Docker
- Render configuration
- observability implementation
- security hardening beyond gateway-specific correctness

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

The first seven documents are mandatory.

The protobuf generation review is mandatory because Agent 02 establishes the generated protobuf and gateway artifacts that this agent must consume.

---

# Documentation Check

Before starting:

verify that every required document exists.

Required:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md
- docs/platform-repository-audit.md
- docs/platform-protobuf-generation-review.md

If the protobuf generation review does not exist:

STOP.

Do not recreate Agent 02's work.

At the end of the task:

perform the documentation check again.

Record the result in:

docs/platform-http-gateway-review.md

---

# Repository Exploration Rules

## IMPORTANT

Do NOT perform an unrestricted recursive repository search.

Use:

README.md

as the repository map.

Use:

docs/repository-layout.md

for the intended structure.

Use:

docs/protobuf-strategy.md

for API and gateway architecture.

Use:

docs/platform-protobuf-generation-review.md

for the actual generated protobuf/gateway state.

Use:

agents/project-context.md

for coding and package conventions.

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

DO NOT recursively inspect:

third_party/googleapis/

The gateway agent only needs to consume generated gateway artifacts and protobuf definitions.

If an import requires a specific Google API definition, rely on Agent 02's verification rather than exploring the dependency tree.

---

# Scope Restrictions

This agent may modify only files directly required for HTTP gateway implementation and its documentation.

Potentially relevant areas include:

- gateway command/package
- gateway runtime wiring
- gateway registration
- gateway configuration
- protobuf HTTP annotations if a documented correction is necessary
- Makefile targets directly required to run the gateway
- platform documentation

Do NOT modify:

- Clients business logic
- Transactions business logic
- repositories
- migrations
- SQL
- generated `.pb.go` files manually
- generated gateway files manually
- CI workflows
- Dockerfiles
- Render configuration
- observability packages
- security architecture

If a required change belongs to another agent:

document it.

Do not take over that agent's work.

---

# 1. Read the Protobuf Strategy

Read:

docs/protobuf-strategy.md

Determine:

- whether grpc-gateway is required
- which services are exposed over HTTP
- HTTP endpoint conventions
- HTTP method conventions
- path conventions
- request/response mapping
- error handling expectations
- gateway registration strategy
- gateway package location
- whether the gateway is a separate process or part of an existing process

Treat this document as authoritative.

---

# 2. Read the Protobuf Generation Review

Read:

docs/platform-protobuf-generation-review.md

Identify:

- generated gateway files
- generated gRPC clients
- generated package locations
- gateway generation command
- generator version
- known generation limitations
- deferred protobuf work

Do not regenerate protobuf code unless a concrete gateway issue requires it.

---

# 3. Understand the Service Architecture

Use:

docs/domain-model.md

and:

docs/repository-layout.md

to determine which services exist.

At minimum account for:

- Clients
- Transactions

Do not assume that every internal gRPC method should automatically become a public HTTP endpoint.

Use the documented API exposure strategy.

---

# 4. Understand Existing Runtime Structure

Read:

README.md

and:

agents/project-context.md

to determine:

- how services currently start
- how configuration is loaded
- how logging is initialized
- how errors are handled
- how servers are wired
- package naming conventions
- command naming conventions

Do not introduce a completely different application lifecycle.

---

# 5. Inspect Existing Gateway State

Locate only the relevant gateway files.

Look for:

- grpc-gateway packages
- gateway commands
- generated gateway files
- `Register...Handler`
- `Register...HandlerFromEndpoint`
- `runtime.ServeMux`
- `grpc.Dial`
- HTTP server setup

Do not search the entire repository recursively.

---

# 6. Determine Gateway Deployment Model

Determine from the documentation whether the gateway is intended to be:

1. a separate HTTP process,
2. embedded alongside the gRPC server,
3. or another explicitly documented arrangement.

Do NOT choose a new architecture based on personal preference.

If documentation conflicts:

document the conflict.

Do not silently redesign the platform.

---

# 7. Gateway Must Not Contain Business Logic

The gateway must perform:

HTTP
→ generated gateway handler
→ gRPC client
→ gRPC service
→ repository/domain logic

It must NOT perform:

HTTP
→ database

or:

HTTP
→ business logic

or:

HTTP
→ direct provider API

unless explicitly required by the documented architecture.

---

# 8. Use Generated Gateway Code

Use the generated grpc-gateway code produced by Agent 02.

Do not manually recreate generated handlers.

For example, if generated code provides:

RegisterClientsHandler

or:

RegisterClientsHandlerFromEndpoint

use those functions.

Do not duplicate their functionality.

---

# 9. Verify Gateway Registration

For each HTTP-exposed service:

verify the generated registration function exists.

Examples may resemble:

RegisterClientsHandlerFromEndpoint

RegisterTransactionsHandlerFromEndpoint

Use the actual generated names.

Do not invent names.

---

# 10. Gateway Multiplexing

If multiple services share the gateway:

use a single gateway multiplexer where the documented architecture calls for one.

The expected flow should resemble:

HTTP request
    ↓
runtime.ServeMux
    ↓
generated handler
    ↓
gRPC service

Do not create a separate mux for every RPC unless explicitly required.

---

# 11. Gateway-to-gRPC Connection

Configure the gateway to communicate with the appropriate gRPC service.

Determine:

- host
- port
- protocol
- connection lifecycle
- local development defaults

Do not hard-code production infrastructure addresses.

Use configuration.

---

# 12. Configuration

Gateway configuration must follow the existing project configuration conventions.

Read:

agents/project-context.md

before adding configuration.

Do not create a second configuration framework.

Do not introduce unrelated environment variables.

Document every gateway-specific variable.

---

# 13. Local Development Defaults

If the existing architecture supports local development:

provide sensible local defaults.

For example:

localhost:<grpc-port>

Only use actual ports defined by the project.

Do not invent production ports.

---

# 14. Production Configuration

The gateway must not assume:

localhost

for production service communication.

Use environment-backed configuration where required.

Do not hard-code Render hostnames.

---

# 15. HTTP Listen Address

Configure the HTTP gateway's listen address according to the deployment architecture.

For containerized deployment:

the gateway generally needs to bind to an externally reachable interface rather than only loopback.

Use the project's documented conventions.

Do not change unrelated service binding behavior.

---

# 16. HTTP Port

The gateway's HTTP port must come from configuration where the architecture requires it.

Do not hard-code a platform-specific port.

If the deployment platform expects a `PORT` environment variable:

document that dependency.

Do not modify Render configuration in this agent.

---

# 17. gRPC Port

The gateway's gRPC upstream address must be configurable.

Do not assume:

localhost:50051

unless that is actually the project's documented value.

---

# 18. Context and Shutdown

The gateway server must support graceful shutdown.

Use the project's existing context/signal conventions where available.

The gateway should respond to:

- SIGINT
- SIGTERM

without abruptly terminating active requests when graceful shutdown is possible.

Do not redesign the entire application lifecycle.

---

# 19. HTTP Server

Use Go's standard HTTP server infrastructure unless the project explicitly specifies another framework.

Avoid unnecessary dependencies.

The gateway should expose:

- the grpc-gateway mux
- documented middleware
- the configured HTTP listener

Do not add business handlers manually.

---

# 20. Gateway Middleware

Only add middleware explicitly required by the documented architecture.

Potential concerns include:

- request logging
- CORS
- request IDs
- authentication

However:

DO NOT implement the complete observability or security architecture here.

Agent 09 handles observability.

Agent 10 handles security.

Only establish the minimum gateway structure necessary for those later components.

---

# 21. CORS

Determine from the documentation whether CORS is required.

If it is required:

implement it according to the documented API requirements.

Do not use:

Allow-Origin: *

for production merely because it is convenient.

If the allowed-origin policy is not yet defined:

document it for Agent 10 rather than inventing a security policy.

---

# 22. Authentication

Do not implement authentication middleware unless explicitly defined by the architecture.

The gateway may need to preserve authorization metadata for downstream gRPC services.

If this requirement exists:

implement only the documented transport mechanism.

Do not invent JWT/session/OAuth validation here.

---

# 23. OAuth Routes

Do NOT implement HighLevel OAuth flows in this agent.

OAuth belongs to the Clients service.

This agent may expose the Clients OAuth endpoints only if:

- the endpoints are represented by the documented HTTP API,
- the Clients service owns the actual implementation,
- and the gateway only forwards the request.

The gateway must not:

- exchange OAuth codes
- decrypt OAuth tokens
- store OAuth credentials
- call HighLevel directly

---

# 24. Webhook Routes

Do NOT implement webhook business processing here.

If webhook endpoints are exposed through grpc-gateway:

the gateway should simply route them to the appropriate gRPC service.

Do not:

- validate provider signatures here unless explicitly defined as gateway responsibility
- update database records
- process events
- call external providers

Those belong to the service layer.

---

# 25. HTTP Error Translation

Use grpc-gateway's standard error translation unless the architecture specifies custom behavior.

Do not invent a second error format.

If custom error mapping is required:

implement it centrally.

Do not duplicate error conversion in every service.

---

# 26. HTTP Status Codes

Verify that HTTP status codes originate from the documented protobuf/API contract.

Do not manually force every RPC to return:

200 OK

when the HTTP semantics require another status.

---

# 27. JSON Serialization

Use the generated gateway behavior and protobuf JSON conventions.

Do not create separate JSON DTOs unless explicitly required.

Avoid duplicate representations of protobuf messages.

---

# 28. Content Types

Ensure the gateway correctly handles:

- application/json
- protobuf-generated request/response serialization

according to grpc-gateway defaults and documented requirements.

Do not introduce custom serialization without a concrete requirement.

---

# 29. HTTP Path Parameters

Verify that path parameters map correctly to protobuf request fields.

For example:

/clients/{client_id}

must map to the documented protobuf field.

Do not add routes not defined by the protobuf contract.

---

# 30. Query Parameters

Verify that query parameters are mapped through the generated gateway layer.

Do not manually parse query parameters in business handlers.

---

# 31. Request Bodies

Verify that HTTP request bodies map to the correct protobuf request messages.

Do not create parallel request structs solely for HTTP.

---

# 32. Service Registration

Register all HTTP-exposed services required by the architecture.

At minimum inspect:

- Clients
- Transactions

Do not register internal-only services unless documented.

---

# 33. gRPC Endpoint Discovery

Use the documented configuration mechanism for locating gRPC services.

Do not introduce service discovery infrastructure in this agent.

If service discovery is planned but not yet implemented:

document the required integration.

---

# 34. Connection Lifecycle

Ensure gateway gRPC connections remain available for the lifetime of the gateway.

Do not create a new gRPC connection per HTTP request.

---

# 35. Connection Shutdown

Ensure gateway connections are closed during shutdown where the connection API requires explicit cleanup.

Avoid resource leaks.

---

# 36. TLS

Do not invent TLS architecture.

If TLS termination is expected to happen at:

- reverse proxy
- Render
- load balancer
- ingress

follow the documented architecture.

If the gateway itself must use TLS:

implement only if explicitly specified.

---

# 37. Health Endpoint

Do not invent health endpoints here unless the platform documentation specifies them.

If a health endpoint is required:

document the requirement for the appropriate platform/observability agent if it falls outside the gateway's scope.

---

# 38. Readiness

Do not add a custom readiness system.

The gateway should at minimum fail appropriately if its upstream gRPC service cannot be established when the architecture requires startup-time connection validation.

Do not introduce service discovery or health orchestration.

---

# 39. Gateway Logging

Follow the existing logging convention.

Read:

agents/project-context.md

Do not introduce:

- fmt.Println
- log.Printf

if the project uses structured logging.

Use the project's established logger.

---

# 40. Gateway Errors

Errors during startup should be:

- logged
- returned
- associated with the correct failure context

Do not swallow startup failures.

Do not use panic for ordinary configuration or connection errors.

---

# 41. Graceful Shutdown

The gateway should:

1. receive shutdown signal
2. stop accepting new requests
3. allow active requests to complete where possible
4. close upstream connections
5. exit cleanly

Follow existing project lifecycle patterns.

---

# 42. Testing

Write targeted gateway tests where appropriate.

Test at minimum:

- gateway initialization
- route registration
- configuration handling
- shutdown behavior if practical
- HTTP-to-gRPC forwarding where practical

Do not create a huge integration-test framework.

---

# 43. Do Not Test Generated Code Internals

Do not write tests for grpc-gateway generated implementation details.

Test the application's gateway wiring.

---

# 44. Gateway Integration Test

If practical:

create a small in-process test setup where:

HTTP request
→ gateway
→ test gRPC server

can be verified.

Do not require external Render infrastructure.

---

# 45. Test HTTP Routes

Verify representative routes for the active services.

At minimum consider:

- one Clients route
- one Transactions route

Use actual routes defined by the protobuf contracts.

Do not invent fake endpoints.

---

# 46. Test Error Propagation

Verify that a representative gRPC error becomes the expected HTTP response through grpc-gateway.

Do not duplicate grpc-gateway's implementation.

---

# 47. Test JSON Mapping

Verify one representative request/response mapping.

Ensure protobuf fields serialize as expected.

---

# 48. Build Verification

Run the narrowest relevant Go build/test commands first.

Do not immediately run the entire repository test suite.

Start with the gateway package.

Then test dependent service packages where appropriate.

---

# 49. Generated Code Verification

Do not modify generated files manually.

If generated code is missing:

document the issue for Agent 02.

If the gateway cannot compile because generation is stale:

STOP and report the dependency rather than manually patching generated code.

---

# 50. Makefile

If the repository already has gateway targets:

use them.

If a gateway target is missing and the architecture requires one:

add a minimal target following existing Makefile conventions.

Do not redesign the root Makefile.

---

# 51. Documentation

Document:

- gateway command
- required environment variables
- HTTP port
- gRPC upstream configuration
- local startup
- exposed service registration
- shutdown behavior

Do not create a second API specification.

The protobuf definitions remain the API source of truth.

---

# 52. README Changes

Only modify README.md if the gateway implementation changes the documented developer workflow.

Do not rewrite the entire README.

Preserve existing structure and terminology.

---

# 53. Avoid Architecture Drift

Do not introduce:

- REST controllers
- Gin
- Echo
- Fiber
- custom HTTP routers
- GraphQL
- direct database HTTP handlers

unless explicitly required by the existing architecture.

The documented grpc-gateway approach must remain the default.

---

# 54. Avoid Duplicate API Layers

There must not be both:

HTTP handler → service

and:

grpc-gateway → gRPC service

implementing the same API unless explicitly documented.

The gateway should remain a transport adapter.

---

# 55. Service Boundary

The architecture should remain:

                    ┌───────────────┐
HTTP ──────────────>│ HTTP Gateway  │
                    └───────┬───────┘
                            │
                            │ gRPC
                            ▼
                 ┌─────────────────────┐
                 │ gRPC Service        │
                 │                     │
                 │ Clients             │
                 │ Transactions        │
                 └──────────┬──────────┘
                            │
                            ▼
                       Repositories

The gateway must not bypass the service layer.

---

# 56. Multiple Service Registration

If the gateway connects to multiple independently running services:

follow the documented service topology.

Do not merge service implementations into the gateway.

The gateway is a transport layer only.

---

# 57. Separate Service Processes

If Clients and Transactions are separate processes:

do not assume they share the same gRPC server.

Configure their upstream endpoints independently where required.

Use configuration.

---

# 58. Embedded Gateway Case

If the documented architecture explicitly embeds the gateway alongside a gRPC server:

follow that architecture.

Do not split it into separate processes merely for preference.

---

# 59. Configuration Validation

Gateway startup should fail clearly when required configuration is missing.

Do not allow empty upstream addresses to reach a connection attempt that produces an unclear error.

Use the existing configuration validation conventions.

---

# 60. No Secrets in Source

Do not commit:

- OAuth client secrets
- API keys
- database passwords
- tokens
- Render deploy hooks

The gateway should not contain provider credentials.

---

# 61. No Production URLs in Source

Do not hard-code:

- Render URLs
- production hostnames
- database hostnames
- provider URLs

Use configuration.

---

# 62. Security Boundary

The gateway is an entry point.

Do not weaken security for convenience.

Do not add permissive CORS or unauthenticated production endpoints without documented justification.

If the required security policy is not yet defined:

document it for Agent 10.

---

# 63. Performance

Do not prematurely optimize the gateway.

At minimum:

- reuse gRPC connections
- avoid per-request connection creation
- avoid unnecessary JSON transformations
- avoid unnecessary allocations where obvious

Do not implement caching here.

Agent 11 handles performance.

---

# 64. Observability Boundary

Do not implement full metrics/tracing.

Ensure the gateway has enough structure that Agent 09 can later add:

- request metrics
- latency metrics
- tracing
- request IDs

without rewriting the gateway.

---

# 65. Docker Compatibility

Do not modify Dockerfiles.

However, ensure the gateway command can run as a normal process in a container.

If the current Docker architecture cannot support it:

document the issue for Agent 06.

---

# 66. Render Compatibility

Do not modify Render configuration.

Ensure the gateway:

- reads configuration from environment
- binds to the configured interface
- binds to the configured port
- exits with non-zero status on startup failure

Document any Render-specific requirement for Agent 07.

---

# 67. Do Not Touch Third-Party Dependencies

Do not modify:

third_party/googleapis/

Do not update submodules.

Do not vendor additional dependencies.

---

# 68. Review Existing Findings

Review:

docs/platform-repository-audit.md

and:

docs/platform-protobuf-generation-review.md

For each gateway-related finding:

- resolve it if this agent owns it
- document it if it belongs to another agent
- do not silently ignore it

---

# 69. Create Review Document

Create:

docs/platform-http-gateway-review.md

Use exactly this structure:

# Platform HTTP Gateway Review

## 1. Objective

Describe the gateway implementation.

## 2. Required Documentation

List each required document and whether it was read.

## 3. Gateway Architecture

Describe:

- HTTP entry point
- gateway mux
- gRPC upstream
- service registration

## 4. Gateway Services

| Service | HTTP Exposure | Registration | Status |
|---|---|---|---|
| Clients | | | |
| Transactions | | | |

## 5. Configuration

Document gateway configuration variables.

## 6. HTTP Routes

Document the source of truth for routes.

Do not duplicate the entire protobuf specification.

## 7. Error Handling

Describe grpc-gateway error handling.

## 8. Shutdown

Describe graceful shutdown behavior.

## 9. Testing

Document tests performed and results.

## 10. Findings

Use:

| ID | Severity | File/Area | Finding | Resolution |
|---|---|---|---|---|

## 11. Deferred Work

List issues belonging to:

- Agent 05
- Agent 06
- Agent 07
- Agent 09
- Agent 10
- Agent 11

where applicable.

## 12. Changes Made

List only files actually modified.

## 13. Documentation Check

Record the final documentation verification.

## 14. Final Status

Use exactly one:

PASS

PASS WITH FOLLOW-UP

BLOCKED

---

# 70. Final Verification

Run:

git status --short

Then:

git diff --stat

Inspect all changed files.

Ensure no unrelated changes were introduced.

---

# 71. Final Gateway Generation Check

Do not regenerate protobuf artifacts unless required.

If generation is required:

use the canonical command documented by Agent 02.

After generation:

verify that generated output is clean.

Do not manually modify generated files.

---

# 72. Final Build/Test

Run the narrowest relevant checks first.

Then, if practical:

go test ./...

Do not spend excessive time running unrelated tests if targeted tests already establish correctness.

If a repository-wide failure is unrelated:

document it.

---

# 73. Final Documentation Check

Confirm all required documents were read:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md
- docs/platform-repository-audit.md
- docs/platform-protobuf-generation-review.md

Record the result in:

docs/platform-http-gateway-review.md

---

# 74. Final Scope Check

Run:

git status --short

Verify that changes are limited to:

- gateway implementation
- gateway configuration where required
- gateway tests
- minimal Makefile changes if required
- minimal README documentation if required
- docs/platform-http-gateway-review.md

Do not revert unrelated pre-existing changes automatically.

Document them if present.

---

# 75. Final Completion Checklist

Before stopping:

- [ ] README.md was read.
- [ ] agents/project-context.md was read.
- [ ] docs/domain-model.md was read.
- [ ] docs/repository-layout.md was read.
- [ ] docs/protobuf-strategy.md was read.
- [ ] docs/migration-plan.md was read.
- [ ] docs/platform-repository-audit.md was read.
- [ ] docs/platform-protobuf-generation-review.md was read.
- [ ] Existing gateway state was inspected.
- [ ] Gateway architecture was confirmed from documentation.
- [ ] Generated gateway code was identified.
- [ ] HTTP-exposed services were identified.
- [ ] Gateway registration was verified.
- [ ] Gateway configuration was implemented consistently.
- [ ] gRPC upstream connections are reusable.
- [ ] HTTP server lifecycle is correct.
- [ ] Graceful shutdown is implemented.
- [ ] No business logic was added to the gateway.
- [ ] No direct database access was added.
- [ ] No direct provider API calls were added.
- [ ] OAuth business logic was not added.
- [ ] Webhook business logic was not added.
- [ ] Generated files were not manually edited.
- [ ] Third-party googleapis files were not modified.
- [ ] Gateway tests were added where appropriate.
- [ ] Gateway compilation succeeded.
- [ ] Relevant service compilation succeeded.
- [ ] HTTP-to-gRPC behavior was verified where practical.
- [ ] No CI redesign was performed.
- [ ] No Docker changes were performed.
- [ ] No Render changes were performed.
- [ ] docs/platform-http-gateway-review.md was created.
- [ ] Final documentation check was recorded.
- [ ] Final git state was inspected.

---

# Final Stop Condition

STOP after:

1. reading all required documentation,
2. confirming the gateway architecture,
3. implementing only the documented HTTP gateway layer,
4. registering the required generated gateway handlers,
5. verifying HTTP-to-gRPC routing,
6. verifying configuration,
7. verifying graceful shutdown,
8. adding focused gateway tests,
9. creating docs/platform-http-gateway-review.md,
10. completing the documentation check,
11. checking final git status.

Do NOT proceed to:

- common packages
- CI/CD
- Docker
- Render
- observability implementation
- security implementation
- performance optimization
- database changes
- repository changes
- Clients business logic
- Transactions business logic
- OAuth implementation
- webhook processing

Those belong to later agents.

STOP.