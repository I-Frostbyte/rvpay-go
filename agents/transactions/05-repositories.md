# Agent 05 — Transactions Repository Layer

## Objective

Implement the Transactions Service repository layer.

This agent takes the database schema and generated SQLC code established by Agents 02 and 03 and builds the repository abstraction that the later service layer will use.

The repository layer must follow the existing RVPay Deposits implementation as closely as possible.

The repository is responsible for:

- database connection/pool access
- SQLC query access
- repository interfaces
- repository implementations
- transaction boundaries where required
- database-level persistence operations
- mapping between persistence models and repository-facing models where necessary
- repository error handling
- migration integration where the existing architecture requires it

The repository is NOT responsible for:

- gRPC handlers
- protobuf request/response handling
- HTTP/REST handlers
- OAuth
- webhook processing
- provider integrations
- business workflows
- transaction lifecycle orchestration
- application startup
- Docker
- deployment
- runtime wiring

Do not implement any of those concerns.

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

These documents are mandatory.

Do not skip them.

---

# Repository Exploration Rules

Use README.md as the repository map.

Perform focused exploration.

Do NOT perform unrestricted recursive searches.

Do NOT inspect deeply into:

- third_party/
- third_party/googleapis/
- googleapis/
- vendor/
- node_modules/
- .git/
- coverage/
- tmp/
- bin/

Do not recursively inspect unrelated services.

The existing Deposits service is the primary implementation reference.

Only inspect the relevant Deposits files needed to reproduce its repository conventions.

---

# Documentation Check

Before modifying code, confirm that you have read:

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

At completion, verify the repository implementation against these documents again.

---

# 1. Inspect the Existing Deposits Repository

Use the existing Deposits repository as the primary implementation template.

Inspect only the relevant files under:

deposits/db/

Determine:

- repository package naming
- repository interface conventions
- repository implementation conventions
- database pool ownership
- SQLC query ownership
- constructor patterns
- error handling
- transaction handling
- migration runner integration
- dependency injection conventions
- context usage
- method naming
- return-value conventions

Do not redesign the repository architecture.

Match the established project style.

---

# 2. Inspect the Transactions Database Layer

Review the database artifacts created by Agents 02 and 03.

Inspect:

- migrations
- SQL query files
- generated SQLC models
- generated SQLC queries
- generated interfaces
- database package documentation

Do not modify the schema in this agent.

Do not modify SQL queries unless a repository requirement exposes a clear SQLC-generation issue.

If the schema or queries are fundamentally incorrect:

STOP and document the problem rather than silently redesigning Agent 02/03 work.

---

# 3. Inspect SQLC Output

Determine exactly what SQLC generated.

Identify:

- generated models
- query methods
- query parameters
- result types
- Querier interface
- transaction support
- DBTX abstraction
- generated database interfaces

The repository must consume generated SQLC code.

Do NOT write SQL directly in the repository layer.

---

# 4. Repository Boundary

The repository must provide a stable persistence abstraction above SQLC.

The service layer should not need to know:

- SQL syntax
- SQLC implementation details
- database connection details
- migration mechanics
- PostgreSQL-specific query mechanics

The service layer should interact with repository methods.

Follow the existing Deposits design.

---

# 5. Repository Package Location

Follow:

docs/repository-layout.md

for the exact location.

The expected structure is likely under:

transactions/db/repo/

but do not assume this if the documentation specifies another location.

Follow the documentation.

Do not create alternative repository directories.

---

# 6. Database Pool Ownership

Follow the existing Deposits architecture.

Determine whether the repository owns:

- pgxpool.Pool
- SQLC Queries
- a database wrapper
- another documented abstraction

Do not introduce a new database library.

Do not replace pgx with database/sql.

Do not introduce GORM.

Do not introduce an ORM.

---

# 7. Repository Constructor

Implement the repository constructor using the same pattern as Deposits.

The constructor should receive only dependencies it actually requires.

Avoid global database state.

Avoid package-level mutable variables.

Do not create singleton database connections.

The repository should be dependency-injectable.

---

# 8. Repository Interfaces

Define repository interfaces according to the project's conventions.

The interface should expose domain-relevant persistence operations.

Do not expose SQLC's entire generated interface automatically.

For example, do not simply make:

type Repository interface {
    sqlc.Querier
}

unless that is explicitly how the existing architecture works.

Prefer a repository boundary that expresses what the Transactions service actually needs.

---

# 9. Merchant Repository

Implement persistence operations required by the domain model for merchants.

Use the SQLC-generated methods.

Possible operations may include:

- CreateMerchant
- GetMerchant
- GetMerchantByExternalID
- ListMerchants
- UpdateMerchant

Only implement operations that correspond to:

- the database schema
- generated SQLC methods
- documented domain requirements

Do not automatically implement CRUD.

---

# 10. Customer Repository

Implement the documented customer persistence operations.

Potential operations include:

- CreateCustomer
- GetCustomer
- GetCustomerByExternalID
- ListCustomers
- UpdateCustomer

Follow the actual domain model.

Ensure customer lookup respects merchant ownership/tenant boundaries where required.

Do not allow cross-merchant data access.

---

# 11. Transaction Repository

Implement persistence operations for core transactions.

Potential operations may include:

- CreateTransaction
- GetTransaction
- GetTransactionByReference
- ListTransactions
- UpdateTransaction
- UpdateTransactionStatus

Only implement documented operations.

Do not create generic database mutation methods.

---

# 12. Deposit Repository

Implement the persistence operations required for Deposits within the new Transactions architecture.

Determine from:

docs/domain-model.md

and:

docs/migration-plan.md

whether Deposits are:

- transaction records
- specialized transaction records
- separate persistence resources
- compatibility records

Follow the documented model.

Do not simply copy the old Deposits repository.

The new Transactions repository must represent the new architecture.

---

# 13. Payout Repository

Implement persistence operations for Payouts if defined by the database/domain model.

Potential operations may include:

- CreatePayout
- GetPayout
- GetPayoutByReference
- ListPayouts
- UpdatePayout
- UpdatePayoutStatus

Only implement operations supported by SQLC and the documented domain.

---

# 14. Provider References

If Transactions stores external provider references:

ensure repository methods can retrieve records using the documented provider/reference identifiers.

Examples might include:

- provider transaction ID
- external transaction reference
- provider payout reference

Do not create provider-specific repository abstractions unless the domain explicitly requires them.

The repository should remain provider-agnostic.

---

# 15. Context Handling

Every database operation must use context.

Do not use:

context.Background()

inside repository methods.

The caller must provide the context.

This ensures:

- request cancellation
- timeout propagation
- graceful shutdown behavior
- proper gRPC request lifecycle

---

# 16. Error Handling

Follow the existing Deposits repository error-handling conventions.

Determine how the existing project handles:

- sql.ErrNoRows
- pgx.ErrNoRows
- constraint violations
- connection errors
- transaction failures

Do not convert every database error into a generic string.

Preserve enough information for the service layer to make appropriate decisions.

---

# 17. Not Found Semantics

Determine the project's convention for missing records.

If the existing repository returns a sentinel/domain error:

use that convention.

If it returns a database-specific not-found error:

follow the established pattern.

Do not invent a new error system.

---

# 18. Constraint Errors

Handle PostgreSQL constraint failures consistently with the existing repository.

Potential examples:

- unique constraint violation
- foreign-key violation
- check constraint violation
- not-null violation

Do not silently ignore constraint errors.

Do not convert all constraints into the same application error unless that is the existing project convention.

---

# 19. Transactions

Determine whether repository operations require database transactions.

Use:

docs/domain-model.md

and:

docs/transactions-database-review.md

to identify operations that must be atomic.

For example, if a documented operation creates multiple related records that must succeed together, use a database transaction.

Do not introduce transactions merely because they are technically possible.

---

# 20. SQLC Transaction Support

Use the SQLC-generated transaction support.

Do not manually implement SQL transaction mechanics if SQLC already provides the required abstraction.

Follow the existing Deposits pattern.

---

# 21. Atomic Operations

Where atomicity is required:

1. begin transaction
2. create transaction-bound SQLC queries
3. execute all required operations
4. rollback on failure
5. commit on success

Ensure rollback is handled correctly.

Do not leave transactions open.

---

# 22. Repository Methods Must Be Deterministic

Repository methods should have predictable behavior.

Avoid:

- hidden retries
- arbitrary sleeps
- background goroutines
- network calls
- provider API calls
- business workflow execution

The repository is a persistence boundary.

---

# 23. No Provider API Calls

The repository must NEVER call:

- pawaPay
- GoHighLevel
- external HTTP APIs
- webhook endpoints
- third-party SDKs

External provider interaction belongs to higher layers.

---

# 24. No Business Logic

Do not place business rules inside repository methods.

For example, repository code should not decide:

- whether a payout is allowed
- whether a deposit should be rejected
- whether a merchant has sufficient balance
- whether a transaction can transition state

The repository persists data.

The service layer owns business decisions.

---

# 25. Mapping Between SQLC and Domain Types

Determine whether the project uses direct SQLC models or domain models between repository and service layers.

Follow:

agents/project-context.md

and the existing Deposits repository.

If mapping is required:

implement explicit mapping functions.

Do not scatter mapping logic throughout repository methods.

---

# 26. Nullability

Respect SQLC-generated nullable types.

Do not blindly dereference nullable database fields.

Handle:

- sql.NullString
- sql.NullInt32
- sql.NullInt64
- sql.NullTime
- pgtype values

according to the actual generated types.

Do not alter the SQLC configuration merely to avoid handling nullable fields.

---

# 27. UUIDs and IDs

Follow the project's existing ID strategy.

Do not generate IDs in the repository unless the schema and existing implementation require repository-side generation.

If PostgreSQL generates IDs:

let PostgreSQL/SQLC handle them.

If the application generates UUIDs:

follow the established project pattern.

---

# 28. Timestamps

Follow the database and SQLC timestamp conventions.

Do not manually manipulate timestamps unless required.

Do not introduce application-local timezone conversions into repository methods.

Persist timestamps consistently with the database strategy.

---

# 29. Pagination

If SQLC queries support pagination:

implement repository methods according to the documented pagination strategy.

Do not implement pagination using:

- loading the entire table
- slicing results in Go
- arbitrary offsets

unless the documented architecture explicitly requires it.

---

# 30. Filtering

Repository filtering should use SQLC-generated queries.

Do not construct SQL dynamically in Go.

Do not concatenate user input into SQL.

If a required filter does not exist in the SQLC layer:

STOP and document that Agent 03 must add the required query.

Do not bypass SQLC.

---

# 31. Ordering

Follow the deterministic ordering defined by the SQL queries.

Do not sort large query results in memory unless explicitly required.

Repository-level ordering should generally occur in PostgreSQL.

---

# 32. Repository Method Naming

Follow the naming convention already established by Deposits.

Avoid inconsistent alternatives such as:

- FetchMerchant
- FindMerchant
- RetrieveMerchant

if the project consistently uses:

GetMerchant

Use the existing convention.

---

# 33. No Protobuf Dependencies

The repository layer should not depend on gRPC request/response types.

Do NOT import:

depositsgrpc

or the new Transactions protobuf package into the repository.

The repository must remain transport-independent.

---

# 34. No HTTP Dependencies

Do not import HTTP packages solely for repository behavior.

The repository must remain transport-independent.

---

# 35. No Logging Unless Established

Follow the existing Deposits repository logging conventions.

If Deposits does not log at the repository layer:

do not introduce repository logging.

Avoid noisy SQL-level logging.

---

# 36. Connection Health

If the existing repository exposes a database health/ping method:

follow the same pattern.

Do not create an independent health subsystem.

Runtime/health handling belongs to the later runtime agent.

---

# 37. Migration Runner

If the existing Deposits repository owns or exposes migration execution:

determine whether Transactions should follow the same pattern.

Do not introduce a second migration architecture.

Follow:

docs/migration-plan.md

Do not execute migrations during arbitrary repository method calls.

---

# 38. Testing the Repository

Do not build the complete service test suite in this agent.

However, repository implementation must be compilable and testable.

If the existing project uses repository mocks:

follow that convention.

If mocks are generated:

use the project's existing mock generation strategy.

Do not introduce a new mocking framework.

---

# 39. Generated Mocks

If the repository interface requires generated mocks:

determine how the existing Deposits repository generates them.

Follow that exact mechanism.

Do not manually create hundreds of lines of mock code.

Do not upgrade mockgen.

Do not change generator versions.

---

# 40. Compile Validation

Run focused compilation/tests for the affected packages.

At minimum verify:

- Transactions repository package compiles
- generated SQLC package compiles
- relevant database packages compile

Do not run unrelated deployment workflows.

---

# 41. Full Test Check

If reasonably fast:

run:

go test ./...

If the repository is not yet compatible with the rest of the new architecture and the full test suite cannot pass:

do not modify unrelated services to make the tests pass.

Document the failure.

---

# 42. Review Generated Changes

Before completion:

run:

git status --short

Then:

git diff --stat

Then inspect the actual diff.

Do not assume generated files are correct.

---

# 43. Scope Enforcement

The expected modifications should be limited to the Transactions repository layer and its directly required documentation/mocks.

Potential locations:

- transactions/db/repo/
- transactions/db/repo/mocks/
- directly required Transactions package files
- docs/transactions-repository-review.md

Do NOT modify:

- Clients service
- Deposits service
- protobuf definitions
- unrelated SQLC packages
- third_party/
- deployment files
- Dockerfiles
- workflows

If a protobuf change appears necessary:

STOP and document it for Agent 04 rather than modifying protobuf contracts here.

If a SQL query change appears necessary:

STOP and document it for Agent 03 rather than redesigning SQL here.

---

# 44. Repository Review Document

Create:

docs/transactions-repository-review.md

This document is mandatory.

---

# 45. Required Documentation Structure

Use exactly:

# Transactions Repository Implementation Review

## 1. Source Documents

List every document read.

## 2. Existing Deposits Repository

Describe the relevant Deposits repository conventions.

## 3. Repository Structure

Document the Transactions repository files created.

## 4. Repository Interfaces

List the repository interfaces and their responsibilities.

## 5. Merchant Persistence

Document merchant repository operations.

## 6. Customer Persistence

Document customer repository operations.

## 7. Transaction Persistence

Document transaction repository operations.

## 8. Deposit Persistence

Document deposit persistence operations.

## 9. Payout Persistence

Document payout persistence operations.

## 10. Transactions

Document any multi-record database transactions.

## 11. Error Handling

Document database error behavior.

## 12. SQLC Integration

Explain how the repository uses generated SQLC code.

## 13. Mocks

Document generated repository mocks if applicable.

## 14. Validation

Document compilation/tests performed.

## 15. Files Changed

List relevant files.

## 16. Risks

Document repository-level risks.

## 17. Unresolved Questions

Document anything that requires a later agent or architectural decision.

---

# 46. Documentation Check

Before finishing, verify that:

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

The final repository review document must reflect the actual implementation.

---

# Completion Checklist

Before stopping, verify:

- [ ] Existing Deposits repository implementation was reviewed.
- [ ] Transactions database schema was reviewed.
- [ ] SQLC-generated code was reviewed.
- [ ] Repository package was created in the documented location.
- [ ] Repository interfaces were created.
- [ ] Repository implementations were created.
- [ ] Merchant persistence was implemented where required.
- [ ] Customer persistence was implemented where required.
- [ ] Transaction persistence was implemented where required.
- [ ] Deposit persistence was implemented where required.
- [ ] Payout persistence was implemented where required.
- [ ] SQLC is used for database access.
- [ ] Raw SQL was not introduced into repository Go code.
- [ ] Context is propagated through database operations.
- [ ] Database errors follow project conventions.
- [ ] Not-found behavior follows project conventions.
- [ ] Required database transactions are atomic.
- [ ] No provider API calls were added.
- [ ] No business logic was added.
- [ ] No protobuf dependencies were added to repositories.
- [ ] Repository mocks follow existing project conventions.
- [ ] Relevant packages compile.
- [ ] Tests were run where practical.
- [ ] No unrelated services were modified.
- [ ] docs/transactions-repository-review.md was created.
- [ ] Git diff was reviewed.

---

# Final Stop Condition

STOP after completing:

1. Transactions repository interfaces
2. Transactions repository implementations
3. SQLC integration
4. Repository mocks where required
5. Focused validation
6. docs/transactions-repository-review.md

Do NOT proceed to:

- merchants service logic
- customer service logic
- deposits business logic
- payouts business logic
- gRPC handlers
- REST handlers
- runtime wiring
- provider integrations
- Docker
- deployment

Those belong to later Transactions agents.

If something required by the repository is missing from the SQLC layer, protobuf layer, or database layer:

DO NOT silently modify the previous agent's work.

Document the issue in:

docs/transactions-repository-review.md

and stop.