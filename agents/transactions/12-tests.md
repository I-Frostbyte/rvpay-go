# Agent 12 — Transactions Tests

## Objective

Create, complete, and execute the test suite for the Transactions microservice.

The Transactions service has now been implemented across the previous agents:

- database
- SQLC
- protobuf
- repositories
- merchants
- customers
- deposits
- payouts
- runtime
- scaffolding

This agent is responsible for establishing confidence that those components work together correctly.

The objective is NOT to redesign the Transactions service.

The objective is to:

1. inspect the existing implementation,
2. identify the appropriate test boundaries,
3. create missing tests,
4. execute the tests,
5. diagnose failures,
6. fix only test-related or clearly demonstrated implementation defects,
7. document the results.

Do not turn this task into a general code review.

Do not redesign working architecture.

---

# Required Reading

Read only:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md

Then read the Transactions implementation review documents:

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

Also inspect the existing Deposits tests that are directly relevant to understanding project testing conventions.

Do NOT recursively inspect the entire Deposits service.

---

# Documentation Check

Before doing implementation work:

confirm that all required documents have been read.

At completion:

perform the documentation check again.

The final review document must explicitly record this check.

---

# Repository Exploration Rules

Use README.md as the repository map.

Perform focused exploration only.

Do NOT perform unrestricted recursive searches.

Do NOT inspect:

- third_party/
- third_party/googleapis/
- vendor/
- node_modules/
- .git/
- coverage/
- tmp/
- bin/

unless a specific test failure directly requires examining a file there.

Do not inspect unrelated services.

Do not scan generated protobuf directories recursively.

Do not inspect the entire repository merely to discover tests.

---

# 1. Establish Testing Conventions

Inspect the existing Deposits tests.

Determine:

- test naming conventions
- package conventions
- table-driven test conventions
- assertion library
- mock usage
- database test strategy
- service test strategy
- test helper conventions
- fixture conventions

Transactions tests must follow the existing project style.

Do not introduce a new testing framework unless the project explicitly requires it.

---

# 2. Inspect Transactions Test Coverage

Determine which Transactions components already have tests.

Focus only on:

- transactions repositories
- transactions services
- transaction handlers
- runtime wiring where practical
- database-related behavior
- generated-code boundaries

Do not treat generated code itself as a test target.

---

# 3. Establish Test Boundaries

Tests must follow the actual architecture.

Use the existing separation between:

service
→ repository
→ database

where applicable.

Do not test implementation details unnecessarily.

Prefer testing observable behavior.

---

# 4. Repository Tests

Inspect:

- merchant repository
- customer repository
- deposit repository
- payout repository

and any shared repository implementation.

Create missing repository tests where practical.

Tests should verify:

- successful operations
- expected not-found behavior
- invalid input behavior where repository owns validation
- database errors
- correct mapping between database models and domain models

Do not duplicate SQLC-generated tests.

---

# 5. Repository Mock Tests

If repository mocks are used:

verify that service tests can substitute repositories cleanly.

Do not test the generated mock implementation itself.

Use mocks to test service behavior.

---

# 6. Merchant Tests

Test the Merchant service behavior.

At minimum cover applicable operations such as:

- create merchant
- retrieve merchant
- update merchant
- list merchant
- validation failures
- repository failures

Use the actual API and service interfaces implemented by the project.

Do not invent operations that do not exist.

---

# 7. Customer Tests

Test the Customer service behavior.

Cover applicable operations such as:

- create customer
- retrieve customer
- update customer
- list customer
- merchant/customer relationships
- validation failures
- repository failures

Use the actual implementation.

---

# 8. Deposit Tests

Test the Deposit service behavior.

At minimum cover the implemented lifecycle and business operations.

Depending on the actual implementation, test:

- creation
- retrieval
- status handling
- validation
- repository failures
- provider-related behavior where the provider is already abstracted

Do not introduce provider integration logic into this agent.

---

# 9. Payout Tests

Test the Payout service behavior.

Cover the actual implemented lifecycle.

Depending on the implementation, test:

- creation
- retrieval
- status handling
- validation
- repository failures
- provider abstraction behavior where applicable

Do not add new provider functionality.

---

# 10. Service Error Handling

Tests should verify that repository failures are correctly translated by service layers.

Where the project defines specific gRPC status codes:

verify them.

For example:

- invalid argument
- not found
- already exists
- internal
- unavailable

Do not invent error mappings.

Use the project's actual conventions.

---

# 11. Input Validation

Test validation at the correct layer.

Examples may include:

- missing required identifiers
- invalid IDs
- invalid amounts
- invalid currency
- missing merchant
- missing customer
- invalid transaction state

Only test rules actually defined by the implementation or documentation.

Do not invent business rules.

---

# 12. Domain Relationships

Where the domain model defines relationships such as:

Merchant
→ Customers

Merchant
→ Transactions

Customer
→ Transactions

tests should verify the implemented relationship behavior.

Do not create tests for relationships that the current implementation does not expose.

---

# 13. Database Tests

Determine whether the project has an established integration-test strategy.

If the existing project uses a real PostgreSQL test database:

follow that strategy.

If the project uses repository mocks for service tests:

do not introduce PostgreSQL solely for service unit tests.

Do not introduce Docker-based test infrastructure unless the repository already uses it or it is clearly required.

---

# 14. Test Database Safety

If database integration tests are used:

they must never target production.

Use:

- dedicated test database
- test environment
- explicit test connection configuration

Never use production credentials.

Never run destructive migrations against an unknown database.

---

# 15. Migration Tests

If migrations are part of the established test strategy:

verify that migrations can be applied successfully.

If down migrations are part of the project workflow:

test them only against a dedicated test database.

Do not run down migrations against the developer's normal database unless explicitly instructed.

---

# 16. SQLC Boundary

Do not write tests for generated SQLC implementation internals.

Instead test:

- repository behavior
- expected query results
- mapping behavior

The purpose is to verify the application's use of SQLC.

---

# 17. Protobuf Boundary

Do not test generated protobuf implementation internals.

Test:

- service methods
- request handling
- response behavior
- gRPC error codes
- serialization boundaries where practical

Do not modify generated protobuf files.

---

# 18. gRPC Tests

Where practical, create service-level gRPC tests.

Tests should verify the externally observable API behavior.

Use the project's established gRPC testing conventions.

Do not require a deployed environment.

---

# 19. Runtime Tests

Do not attempt to comprehensively unit-test `main()`.

If the runtime exposes testable wiring:

test only what is practical.

Focus on:

- configuration
- dependency wiring
- service registration
- server construction

Do not redesign the runtime to make it easier to test.

---

# 20. Configuration Tests

If the Transactions configuration package contains meaningful parsing or validation logic:

test it.

Cover:

- valid configuration
- missing required configuration
- invalid values
- boolean parsing
- numeric parsing
- environment handling

Use the actual configuration structure.

---

# 21. Environment Isolation

Tests must not depend on the developer's personal `.env`.

Do not assume:

- local credentials
- Render credentials
- production environment
- developer-specific ports

Use explicit test configuration.

---

# 22. Mocking Rules

Use mocks only where appropriate.

Good candidates:

- repositories
- external provider interfaces
- infrastructure abstractions

Do not mock the component under test.

Do not create elaborate mock hierarchies unnecessarily.

Follow the project's existing mock conventions.

---

# 23. Provider Interfaces

If deposit or payout logic already depends on a provider interface:

test the service against that interface.

Do not implement a real provider.

Do not add provider SDKs.

Do not make network requests during normal unit tests.

---

# 24. Network Isolation

Normal unit tests must not make external network calls.

Do not call:

- payment providers
- banking APIs
- HighLevel
- Render
- external HTTP APIs

Tests must remain deterministic.

---

# 25. Test Determinism

Avoid tests dependent on:

- current time
- random values
- network availability
- machine-specific paths
- environment-specific ports

Where the existing implementation requires time:

use the project's established strategy.

Do not redesign production code solely to introduce a testing abstraction.

---

# 26. Test Data

Use clear, deterministic test data.

Avoid:

- real personal information
- real credentials
- real API keys
- production transaction IDs
- production database records

Use synthetic values.

---

# 27. Table-Driven Tests

Where the existing project uses table-driven tests:

follow that convention.

Use table-driven tests especially for:

- validation
- error mapping
- status transitions
- configuration parsing

Do not force every test into table-driven form.

---

# 28. Test Names

Follow existing naming conventions.

Names should describe behavior.

Prefer:

TestCreateDeposit_InvalidMerchant

over:

TestDeposit1

Use the actual project naming style.

---

# 29. Test Helpers

Before creating new helpers:

search only the relevant Transactions package and existing Deposits test helpers.

Reuse existing helpers where appropriate.

Do not create duplicate helper packages.

---

# 30. Test Fixtures

If the repository already has fixture conventions:

follow them.

Do not create a large fixture framework for a small number of tests.

Keep fixtures close to the tests when practical.

---

# 31. Test Organization

Place tests beside the implementation they exercise unless the existing project uses another convention.

Do not create a global:

tests/

directory unless the repository already uses one.

---

# 32. Generated Code

Never modify generated files to make tests pass.

Do not add tests inside generated files.

If generated output is incorrect:

document the issue.

---

# 33. First Test Run

Before modifying tests:

run the existing Transactions tests.

Use:

go test ./...

or the narrower Transactions command established by the repository.

Record the baseline.

Do not immediately start changing implementation code.

---

# 34. Diagnose Existing Failures

Classify failures into:

1. missing test coverage
2. test defect
3. implementation defect
4. environment/setup problem
5. unrelated repository failure

Do not assume every failure is caused by Transactions.

---

# 35. Implementation Fixes

If a test exposes a genuine Transactions implementation defect:

make the smallest correct fix.

Only modify implementation code when the failure demonstrates a real defect.

Do not perform unrelated refactoring.

Document implementation fixes in:

docs/transactions-tests-review.md

---

# 36. No Architectural Changes

Do not:

- split services
- merge services
- redesign repositories
- change database schema
- change protobuf contracts
- redesign runtime
- introduce event buses
- introduce caching
- introduce queues

Testing is not an excuse for architectural changes.

---

# 37. No New Infrastructure

Do not add:

- Redis
- Kafka
- RabbitMQ
- external testing platforms
- CI services
- cloud databases

unless already required by the existing test architecture.

---

# 38. Test Coverage

Aim for meaningful behavioral coverage.

Do not chase an arbitrary percentage.

Prioritize:

- business rules
- validation
- repository behavior
- service behavior
- error handling
- transaction lifecycle behavior

---

# 39. Edge Cases

Include meaningful edge cases supported by the implementation.

Examples:

- missing IDs
- nonexistent merchant
- nonexistent customer
- duplicate records
- invalid status transitions
- repository errors
- empty result sets
- invalid configuration

Do not invent unsupported behavior.

---

# 40. Concurrency

If Transactions contains explicitly concurrent behavior:

test it where appropriate.

Do not introduce concurrency tests merely because the service is a microservice.

Do not create flaky timing-dependent tests.

---

# 41. Race Detection

If practical, run:

go test -race ./...

If the repository already uses race detection:

follow that convention.

If race detection reveals genuine issues:

document them.

Do not rewrite unrelated concurrency code.

---

# 42. Formatting

Run:

gofmt

only on files changed by this agent.

Do not reformat the entire repository.

Do not produce unrelated formatting changes.

---

# 43. Static Validation

If the repository already uses:

- go vet
- staticcheck
- golangci-lint

follow the existing project configuration.

Do not introduce a new linter.

---

# 44. Full Test Run

After creating tests:

run the focused Transactions tests.

Then, if practical:

go test ./...

Do not hide unrelated failures.

---

# 45. Test Failure Handling

If tests fail:

investigate the actual failure.

Do not:

- delete the failing test
- weaken assertions
- skip tests
- add broad retries
- add arbitrary sleeps

unless there is a documented reason consistent with project conventions.

---

# 46. No `t.Skip` Workarounds

Do not use:

t.Skip()

to conceal an implementation failure.

Only use skips where the existing project explicitly requires environment-dependent tests to be skipped.

Document the reason.

---

# 47. No Weak Assertions

Do not replace meaningful assertions with:

- `require.NoError` alone
- `assert.NotNil` alone
- generic success checks

when the actual response should be verified.

Verify:

- returned values
- status
- identifiers
- errors
- state changes

where appropriate.

---

# 48. Repository Error Tests

Repository tests should verify meaningful database/repository failures.

Examples:

- not found
- constraint violation
- connection failure where practical

Do not require unrealistic database failure simulation.

---

# 49. Service Error Tests

Service tests should verify that repository errors are translated correctly.

For example:

repository not found
→ expected service/gRPC not-found behavior

Use the actual error conventions.

---

# 50. Transaction Lifecycle Tests

For Deposits and Payouts:

test the lifecycle states actually supported by the implementation.

Do not invent additional states.

Do not alter lifecycle rules just to simplify testing.

---

# 51. Merchant/Customer Dependencies

Where a transaction requires a Merchant or Customer:

test the dependency behavior.

Examples:

- valid merchant
- missing merchant
- valid customer
- missing customer

Only where the service implementation actually performs those checks.

---

# 52. Test Naming Consistency

Review all new test names.

They must:

- describe behavior
- use project conventions
- remain searchable
- avoid generic numbering

---

# 53. Test Documentation

Tests should be understandable from the code.

Do not add verbose comments explaining obvious Go behavior.

Use comments only where the business reason is not obvious.

---

# 54. Avoid Over-Testing Implementation Details

Do not assert:

- private helper call counts
- exact internal function ordering
- irrelevant SQL implementation details
- generated-code internals

unless the project's existing testing strategy explicitly requires it.

Test behavior.

---

# 55. Review Existing Tests

Before adding duplicate tests:

inspect existing Transactions test files.

Do not create duplicate coverage merely to increase file count.

Improve existing tests where appropriate.

---

# 56. Test File Scope

Expected test locations should remain inside the Transactions service.

Do not modify tests for unrelated services.

---

# 57. Git Scope

At this stage inspect:

git status --short

before modifications.

Then inspect again after modifications.

The expected changes should primarily be:

- Transactions test files
- small implementation fixes proven necessary by tests
- docs/transactions-tests-review.md

Do not modify unrelated files.

---

# 58. No Third-Party Changes

Do not modify:

third_party/

or:

third_party/googleapis/

unless a specific test failure proves that a dependency file there is directly relevant.

If that happens:

STOP and document the issue rather than modifying it automatically.

---

# 59. Review Document

Create:

docs/transactions-tests-review.md

This document is mandatory.

---

# 60. Required Review Document Structure

Use:

# Transactions Tests Review

## 1. Source Documents

List every required document read.

## 2. Existing Testing Conventions

Document the conventions discovered from Deposits.

## 3. Baseline

Document the test state before changes.

## 4. Test Coverage Added

Use:

| Component | Tests Added | Coverage Focus |
|---|---:|---|

## 5. Repository Tests

Document repository test coverage.

## 6. Service Tests

Document:

- merchants
- customers
- deposits
- payouts

## 7. gRPC Tests

Document API-level tests added.

## 8. Configuration Tests

Document configuration coverage.

## 9. Integration Tests

Document database/integration testing if applicable.

## 10. Mocking

Document repository/provider mocking strategy.

## 11. Validation

Document commands executed.

Example:

go test ./...

go test -race ./...

go vet ./...

Only list commands actually executed.

## 12. Failures

For every failure document:

- test
- cause
- whether it was fixed
- remaining impact

## 13. Implementation Fixes

List any production-code changes made because of tests.

## 14. Unresolved Issues

List remaining issues.

## 15. Scope Verification

Confirm that unrelated services and directories were not modified.

## 16. Documentation Check

Confirm every required document was read.

---

# 61. Documentation Check

Before completion verify again that the following were read:

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

Record this in:

docs/transactions-tests-review.md

---

# 62. Final Git Review

Run:

git status --short

Then:

git diff --stat

Then inspect the relevant diff.

Check specifically for:

- accidental generated changes
- unrelated formatting
- unrelated service modifications
- deleted files
- modified protobuf files
- modified migrations
- third-party changes

---

# 63. Generated Output Check

If tests or generation commands changed generated files:

determine why.

Do not automatically commit generated changes.

If generated output changed unexpectedly:

investigate and document it.

---

# 64. Scope Verification

The final implementation should not contain unrelated changes.

Do not modify:

- Clients
- Integrations
- Deposits legacy implementation
- unrelated services
- third_party
- deployment configuration
- CI workflows
- protobuf source contracts
- database migrations

unless a specific Transactions test proves a necessary correction.

---

# 65. Completion Checklist

Before stopping:

- [ ] Required documents were read.
- [ ] Existing Deposits testing conventions were inspected.
- [ ] Existing Transactions tests were inspected.
- [ ] Baseline test results were recorded.
- [ ] Repository tests were created or improved where necessary.
- [ ] Merchant tests were created or improved.
- [ ] Customer tests were created or improved.
- [ ] Deposit tests were created or improved.
- [ ] Payout tests were created or improved.
- [ ] Relevant gRPC tests were created.
- [ ] Configuration tests were created where applicable.
- [ ] Database integration tests were added only if consistent with project conventions.
- [ ] Provider interfaces were mocked where applicable.
- [ ] No external network calls are made by normal unit tests.
- [ ] No production credentials are used.
- [ ] No real customer data is used.
- [ ] No generated files were manually modified.
- [ ] No protobuf contracts were changed.
- [ ] No database schema was changed.
- [ ] No unrelated service was modified.
- [ ] No third_party files were modified.
- [ ] Tests were executed.
- [ ] Race detection was executed where practical.
- [ ] Static validation was executed where already supported.
- [ ] Failures were investigated.
- [ ] No tests were weakened or skipped to hide failures.
- [ ] Necessary implementation defects were fixed minimally.
- [ ] `git status` was reviewed.
- [ ] `git diff` was reviewed.
- [ ] `docs/transactions-tests-review.md` was created.
- [ ] Documentation check was completed.

---

# Final Stop Condition

STOP after:

1. inspecting existing tests,
2. establishing baseline results,
3. adding meaningful Transactions tests,
4. executing the tests,
5. fixing only demonstrated defects,
6. reviewing the resulting changes,
7. creating `docs/transactions-tests-review.md`.

Do NOT proceed to:

- production readiness review
- deployment
- infrastructure changes
- performance optimization
- load testing
- security audit
- architecture redesign
- new service creation
- provider integration
- OAuth
- webhook implementation

Those responsibilities belong elsewhere.

If a significant architectural or production issue is discovered:

document it in:

docs/transactions-tests-review.md

Do not attempt to solve it outside the scope of this agent.

STOP.