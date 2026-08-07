# Agent 11 — Tests & Verification

## Objective

Create, implement, execute, and validate the automated test suite for the Clients Service.

This agent is responsible for establishing confidence that the Clients Service implemented by Agents 01–10 behaves correctly and integrates correctly with the existing RVPay architecture.

The testing strategy must follow the existing testing conventions used by the repository and the Deposits service.

This agent must test the implementation rather than redesign it.

This agent may fix implementation defects discovered during testing when the fix is:

- local
- clearly understood
- consistent with the existing architecture
- required for the failing test to pass

Do not redesign the Clients Service during this agent.

Do not introduce new architectural patterns.

Do not restructure packages.

Do not redesign repositories.

Do not redesign protobuf contracts.

Do not redesign OAuth architecture.

Do not redesign webhook architecture.

If a failure indicates an architectural problem rather than a localized defect, stop and document it instead of redesigning the system.

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

- clients/
- deposits/

Use the Deposits service only as a reference for existing testing conventions.

Do not recursively inspect the repository.

---

# Repository Exploration Rules

Use the root README.md as the repository map.

Use agents/project-context.md as the coding and testing conventions source.

Use the foundation documents as the architectural source of truth.

Do not explore unrelated directories simply because they exist.

Do not recursively inspect:

- third_party/
- googleapis/
- vendor/
- .git/
- node_modules/
- coverage/
- tmp/
- bin/
- generated dependency trees

Do not inspect the entire repository looking for tests.

Only locate tests that are directly relevant to:

- deposits/
- clients/
- shared protobuf contracts used by clients
- repository-level Go test configuration

---

# Existing Code Protection

Do not overwrite existing tests merely to make them pass.

Do not delete existing tests.

Do not weaken assertions.

Do not remove test cases because implementation currently fails.

Do not modify unrelated tests.

If an existing test exposes a real implementation defect, fix the implementation rather than weakening the test.

If an existing test reflects an obsolete contract that was intentionally changed by the foundation documentation, document the discrepancy before modifying it.

---

# Testing Strategy

Test the Clients Service in layers.

The testing sequence is:

1. Repository tests
2. Business service tests
3. OAuth tests
4. Webhook tests
5. Provider registry tests
6. Runtime/configuration tests
7. gRPC/API tests where practical
8. Full repository test suite

Do not skip lower-level tests merely because an integration test passes.

---

# Step 1 — Test Discovery

First inspect the Clients Service and identify:

- existing test files
- untested public methods
- repository interfaces
- service interfaces
- provider interfaces
- OAuth interfaces
- webhook interfaces
- runtime components
- generated mocks
- existing test helpers

Compare the testing style with:

deposits/

Do not blindly copy Deposits tests.

Use Deposits only to understand conventions.

Before writing tests, produce a short internal testing plan.

Do not create a separate planning document unless necessary.

---

# Step 2 — Repository Tests

Create tests for the Clients repository layer.

Test:

Create

FindByID

Update

Delete

Exists

List

Count

provider lookups

integration lookups

OAuth token persistence

webhook persistence

event lookup

duplicate detection

transaction behavior where applicable

Use the repository abstractions already created by Agent 05.

Do not bypass the repository layer in repository tests.

---

# Database Testing

Use the existing repository/database testing strategy where one exists.

If the project already uses a PostgreSQL test database or integration-test pattern, follow it.

Do not introduce a new database testing framework unless required.

Do not require a developer's production database.

Do not modify development database credentials simply to make tests pass.

Tests must be deterministic.

---

# Step 3 — Business Service Tests

Test business logic independently from the database wherever possible.

Use mocks for repositories.

Test:

client creation

client retrieval

client updates

client deletion

platform handling

integration creation

integration lookup

duplicate integration prevention

client state validation

provider capability validation

transaction coordination

error propagation

idempotency

business validation

Do not use real external providers.

Do not make network requests.

---

# Step 4 — Provider Registry Tests

Test:

provider registration

provider lookup

unknown provider handling

duplicate registration

provider capabilities

provider availability

provider interface compliance

Ensure the registry behaves deterministically.

Test HighLevel as the currently implemented provider.

Do not create fake future providers unless a test double is required.

---

# Step 5 — OAuth Tests

Test OAuth behavior without making real HighLevel API calls.

Test:

authorization URL generation

state generation

state validation

callback validation

authorization code exchange

token parsing

token persistence

refresh token handling

token expiry

refresh failure

provider error handling

invalid callback

missing parameters

invalid state

duplicate installation

OAuth failure rollback

Use mocked HTTP clients or test HTTP servers where appropriate.

Never use real:

client IDs

client secrets

authorization codes

access tokens

refresh tokens

provider credentials

---

# OAuth Security Tests

Explicitly test that sensitive values are not accidentally exposed.

Verify that:

access tokens are not returned where they should not be

refresh tokens are not logged

client secrets are not logged

authorization codes are not logged

invalid state values are rejected

invalid callbacks are rejected

---

# Step 6 — Webhook Tests

Test webhook processing without depending on the real HighLevel service.

Test:

valid webhook

invalid webhook

invalid signature

missing signature

invalid timestamp

malformed payload

unsupported event

unknown provider

duplicate event

retry behavior

event persistence

event dispatch

provider normalization

registration

unregistration

verification

provider errors

database failures

---

# Webhook Security Tests

Explicitly verify:

signature validation occurs before processing

invalid requests are rejected

malformed payloads are rejected

duplicate events are handled safely

sensitive webhook data is not logged

provider secrets are never exposed

---

# Step 7 — Service/Transport Tests

Where practical, test the gRPC service handlers.

Verify:

valid request

invalid request

missing required fields

not-found behavior

duplicate behavior

business validation errors

repository errors

provider errors

correct protobuf response

correct gRPC status code

Do not duplicate the entire business-service test suite here.

Transport tests should verify translation between:

protobuf request

business service

protobuf response

---

# REST / Gateway Testing

If the Clients Service exposes REST endpoints through grpc-gateway, test the gateway surface where practical.

Verify:

HTTP method

path

request translation

response translation

HTTP status

gRPC error translation

JSON serialization

Do not implement a second business layer for REST.

REST and gRPC must ultimately invoke the same business services.

---

# Step 8 — Configuration Tests

Test configuration loading.

Verify:

required variables

optional variables

defaults

boolean parsing

integer parsing

URL parsing

invalid values

missing required values

environment overrides

Do not place real credentials in tests.

Use test-specific environment values.

Ensure tests do not depend on the developer's `.env`.

---

# Step 9 — Runtime Tests

Test runtime initialization where practical.

Verify:

configuration loads

dependencies are constructed

provider registry initializes

providers register

repositories initialize

services initialize

server initialization succeeds

shutdown does not panic

resources are released

Do not require an externally running Render deployment for unit tests.

---

# Graceful Shutdown Testing

Test:

SIGINT handling

SIGTERM handling

gRPC shutdown

HTTP/gateway shutdown

database pool closure

in-flight request handling where practical

shutdown should be deterministic.

Do not introduce arbitrary sleeps simply to make shutdown tests pass.

Prefer channels, contexts, synchronization primitives, and explicit readiness signals.

---

# Step 10 — Integration Tests

Where practical, create integration tests covering the complete Clients flow.

At minimum, validate the major lifecycle:

Client creation

↓

Integration creation

↓

OAuth installation

↓

Token persistence

↓

Webhook registration

↓

Webhook reception

↓

Event dispatch

Use mocks or local test infrastructure for external providers.

Never call the real HighLevel service from CI.

---

# Test Isolation

Tests must not depend on:

developer `.env`

production credentials

Render environment variables

external internet connectivity

a developer's local PostgreSQL instance unless explicitly running an integration-test environment

unrelated services

test execution order

local filesystem state

---

# Test Data

Create deterministic test fixtures.

Do not reuse production-like credentials.

Do not hardcode secrets.

Use clearly identifiable test values.

Clean up database state after integration tests.

---

# Generated Code

Do not manually modify generated:

protobuf

grpc

sqlc

mock

files.

If generated code is incorrect:

1. identify the generator input
2. identify the generator version
3. regenerate using the project-approved tooling
4. verify the generated diff
5. commit only the legitimate generated changes

---

# Test Commands

Run the narrowest relevant test first.

For example:

go test ./clients/...

Then run broader tests.

Finally run:

go test ./...

Do not immediately start with a full repository test if a narrow test can identify the problem faster.

---

# Race Detection

Where practical, run:

go test -race ./clients/...

Use race detection particularly for:

webhooks

provider registry

runtime

concurrent requests

shutdown

shared state

If race detection exposes a genuine concurrency issue, fix it only if the fix is local and consistent with the existing architecture.

---

# Static Validation

Run the repository's existing validation commands.

Where already supported, verify:

go vet

gofmt

existing lint commands

Do not introduce a new linter solely for this agent.

---

# Test Failures

When a test fails:

1. identify the exact failing test
2. determine whether the failure is caused by:
   - test defect
   - implementation defect
   - environment problem
   - generated-code mismatch
   - architectural issue
3. fix only local implementation defects
4. rerun the narrow test
5. rerun the relevant package tests
6. continue toward the full test suite

Do not repeatedly rerun the entire repository without understanding the failure.

---

# Architectural Failures

Stop and document the issue if a failure requires:

- redesigning protobuf contracts

- changing database architecture

- redesigning repository interfaces

- changing provider architecture

- changing service boundaries

- changing the Clients/Transactions boundary

- modifying unrelated microservices

Do not make architectural changes during this agent.

---

# Test Coverage

Measure coverage for the Clients Service where tooling is already available.

Focus on meaningful coverage rather than a percentage target.

Prioritize:

business rules

error paths

OAuth security

webhook security

idempotency

repository behavior

provider registry

runtime lifecycle

Do not write meaningless tests solely to increase coverage.

---

# Deliverables

Create or update tests under:

clients/

using the repository's established test naming and package conventions.

Tests should be located beside the code they test unless the existing project convention requires otherwise.

Do not create a new top-level testing framework.

---

# Validation Sequence

Complete validation in this order:

1. gofmt

2. targeted repository tests

3. targeted service tests

4. OAuth tests

5. webhook tests

6. provider tests

7. runtime tests

8. Clients package tests

9. race tests where practical

10. go vet where supported

11. full repository tests

12. Docker build if practical

Do not declare success until the final repository test suite has been executed.

---

# Full Repository Verification

Run:

go test ./...

Record:

- passed packages
- failed packages
- skipped tests
- environment-dependent tests
- known unrelated failures

If unrelated existing tests fail, do not modify them.

Determine whether the Clients implementation caused the failure.

---

# Completion Rules

Before completing verify:

- Existing tests were preserved.

- Existing assertions were not weakened.

- New tests follow repository conventions.

- No real credentials exist in test files.

- No external provider calls occur during CI tests.

- No generated files were manually modified.

- Tests are deterministic.

- Tests do not depend on developer environment state.

- Tests do not require unrelated services.

- Project formatting is clean.

- Project builds successfully.

- Clients tests pass.

- Full repository tests have been executed.

If a prerequisite from a previous agent is missing, stop and explain why instead of creating a parallel implementation.

---

# Testing Review

Before completing, perform a comprehensive testing review.

Confirm:

- repository behavior is covered

- business rules are covered

- error paths are covered

- provider registry is covered

- OAuth lifecycle is covered

- OAuth security is covered

- webhook lifecycle is covered

- webhook security is covered

- idempotency is covered

- runtime initialization is covered where practical

- graceful shutdown is covered where practical

- gRPC translation is covered where practical

- REST/gateway translation is covered where applicable

- tests remain isolated

- tests remain deterministic

- no production credentials are present

- no real external provider calls occur

- race conditions were considered

- full repository tests were executed

Produce:

clients/docs/test-review.md

The report must contain:

## Test Coverage Summary

Summarize which components were tested.

## Tests Added

List the major test suites and their purpose.

## Commands Executed

List the actual validation commands executed.

## Results

Record whether each command passed or failed.

## Known Failures

Document any failures that are unrelated to the Clients implementation.

## Defects Fixed

List implementation defects discovered and fixed during testing.

## Remaining Risks

Document areas that could not be fully tested.

## Production Confidence

Give a concise assessment of whether the Clients Service is ready for the final production review.

Do not claim full production readiness solely because tests pass.

Only after this review is complete should the project proceed to Agent 12.