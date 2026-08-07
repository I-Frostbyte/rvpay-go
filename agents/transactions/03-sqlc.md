# Agent 03 — Transactions SQLC Generation and Verification

## Objective

Implement and verify the SQLC layer for the new Transactions Service.

Agent 02 established the Transactions PostgreSQL schema and migrations.

This agent consumes that database design and creates the SQL query definitions and generated SQLC Go code required by the Transactions repository layer.

This agent must NOT implement repositories, services, protobufs, handlers, runtime wiring, or business logic.

The responsibility of this agent is:

PostgreSQL schema
        ↓
SQL query definitions
        ↓
SQLC configuration
        ↓
Generated Go database models and query methods
        ↓
Validation

The resulting SQLC layer must follow the existing Deposits implementation and the project's documented conventions.

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

These documents are mandatory.

Do not skip any of them.

---

# Architecture Source of Truth

Use the following hierarchy:

1. README.md
2. agents/project-context.md
3. docs/domain-model.md
4. docs/repository-layout.md
5. docs/protobuf-strategy.md
6. docs/migration-plan.md
7. docs/transactions-existing-review.md
8. docs/transactions-database-review.md
9. existing Deposits SQLC implementation

Do not invent database behavior that is not supported by these sources.

If the SQLC implementation appears to require an architectural decision, stop and document the issue rather than silently deciding.

---

# Repository Exploration Rules

Use the root README.md as the repository map.

Perform focused exploration only.

Do not perform an unrestricted repository-wide scan.

Do not recursively inspect:

- third_party/
- googleapis/
- vendor/
- node_modules/
- .git/
- coverage/
- tmp/
- bin/

Do not inspect unrelated services.

Do not recursively inspect generated dependency directories.

Only inspect existing Deposits SQLC/query files when necessary to establish project conventions.

---

# Documentation Check

Before modifying anything, confirm that the following have been read:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md
- docs/transactions-existing-review.md
- docs/transactions-database-review.md

At completion, verify that the implementation still agrees with all of them.

---

# 1. Inspect Existing Deposits SQLC Implementation

Inspect only the relevant existing Deposits files.

Focus on:

deposits/db/query/

deposits/db/sqlc/

deposits/db/doc.go

Any SQLC configuration used by Deposits.

Determine:

- SQLC version
- SQLC configuration format
- schema input
- query directory
- generated output directory
- package name
- generated package name
- SQL driver
- database type mappings
- nullable field handling
- query naming
- parameter naming
- transaction support
- generated interface conventions

Do not modify Deposits SQLC output.

---

# 2. Inspect Project SQLC Version

Determine the SQLC version mandated by the repository.

Use the project documentation and existing configuration.

Do not arbitrarily upgrade SQLC.

Do not introduce a different SQLC version.

If the project pins SQLC through:

go.mod

deposits/db/doc.go

Makefiles

or another documented mechanism,

follow that mechanism.

---

# 3. Review Transactions Database Schema

Read the migration implementation created by Agent 02.

Determine:

- tables
- columns
- types
- primary keys
- foreign keys
- constraints
- indexes
- nullable fields
- status fields
- timestamps
- relationships

The SQLC layer must correspond to the actual Transactions schema.

Do not write queries against tables that do not exist.

Do not modify migrations merely to make SQLC generation easier.

If the schema is incorrect, document the issue instead of silently changing it.

---

# 4. Determine Required Query Surface

Using:

docs/domain-model.md

and:

docs/transactions-database-review.md

determine the database operations required by the Transactions domain.

Group queries logically by entity.

Potential areas include:

- merchants
- customers
- transactions
- deposits
- payouts

Only implement queries required by the documented architecture.

Do not generate every conceivable CRUD operation.

Avoid speculative queries.

---

# 5. Query Design Principles

Every SQL query must have a clear purpose.

Queries should be:

- explicit
- deterministic
- parameterized
- easy to test
- compatible with SQLC
- aligned with repository responsibilities

Do not construct SQL using string concatenation.

Do not embed application logic into SQL unnecessarily.

Do not duplicate queries unnecessarily.

---

# 6. Merchant Queries

If merchants are part of the Transactions database according to the domain model, implement the required SQL queries.

Evaluate the need for:

- create merchant
- get merchant by ID
- get merchant by external identifier
- update merchant
- list merchants where required

Only create operations supported by the documented use cases.

Do not create unnecessary administrative queries.

---

# 7. Customer Queries

If customers are owned by Transactions, implement the required SQL queries.

Evaluate:

- create customer
- get customer by ID
- get customer by merchant
- get customer by external/provider identifier
- update customer
- list/query operations required by the domain

Ensure merchant scoping is respected where applicable.

Do not expose queries that could accidentally cross merchant boundaries if the schema requires tenant isolation.

---

# 8. Transaction Queries

Implement the required SQL operations for the core transaction entity.

Evaluate:

- create transaction
- get transaction by ID
- get transaction by public identifier
- get transaction by provider identifier
- update transaction status
- update transaction lifecycle fields
- list/filter transactions
- retrieve transaction history where required

Only implement state changes supported by the documented lifecycle.

Do not implement business-state transition logic in SQL.

The SQL should persist the requested state; the service layer will validate whether the transition is allowed.

---

# 9. Deposit Queries

Implement SQL operations required by the new Transactions architecture for deposits.

Determine whether Deposits are represented as:

- a transaction subtype
- a related table
- another documented structure

Use the result from:

docs/transactions-database-review.md

Do not duplicate obsolete Deposits queries blindly.

If an existing Deposits query can be adapted conceptually, preserve its useful behavior while matching the new Transactions schema.

---

# 10. Payout Queries

If payouts are part of the Transactions database model, implement the required queries.

Evaluate:

- create payout
- get payout
- update payout status
- provider reference lookup
- transaction relationship
- merchant/customer scoped lookup where required

Do not implement payout business rules.

---

# 11. Query Naming

Follow existing project conventions.

Use descriptive names that make SQLC-generated methods understandable.

Examples of the desired style:

GetTransaction

GetTransactionByProviderReference

CreateTransaction

UpdateTransactionStatus

ListTransactionsByMerchant

Do not blindly copy these names if the existing project uses a different convention.

The actual names must follow the project's established style.

---

# 12. Query Parameters

Use strongly typed SQL parameters where SQLC supports them.

Ensure parameter names are meaningful.

Avoid ambiguous names such as:

id

value

data

unless the context makes them unambiguous.

Prefer names such as:

transaction_id

merchant_id

provider_reference

status

created_after

created_before

where appropriate.

Follow existing SQL formatting conventions.

---

# 13. Nullable Fields

Review nullable database columns carefully.

Ensure SQLC generates the expected Go representation.

Follow existing project conventions for:

- nullable strings
- nullable timestamps
- nullable integers
- nullable UUIDs
- nullable database types

Do not manually modify generated nullable types.

If a generated type is undesirable, determine whether the SQL/schema/query can be adjusted safely and document the reason.

---

# 14. Monetary Values

Ensure queries preserve the database representation of monetary values.

Do not cast monetary values unnecessarily.

Do not introduce floating-point conversion in SQL.

Follow the database design established by Agent 02.

---

# 15. Status Queries

Queries that update status must be explicit.

Avoid broad updates such as:

UPDATE transactions
SET status = $1;

without an appropriate WHERE clause.

Status updates should target a specific record or explicitly scoped set.

Where documented, include appropriate concurrency/state protections.

Do not invent a state machine.

---

# 16. Provider Reference Queries

Where provider references are part of the schema, provide efficient lookup queries where required.

Ensure provider reference queries respect the documented scope.

For example, if provider references are only unique within a merchant/provider combination, do not assume global uniqueness.

Follow the actual database constraints.

---

# 17. Idempotency Queries

If Agent 02 implemented an idempotency key according to the architecture, implement the required lookup/query operation.

The query must align with the actual uniqueness scope.

Do not implement idempotency behavior in this agent.

Only expose the database operation needed by the repository/service layer.

---

# 18. Time-Based Queries

If transaction listing/history requires time filtering, implement appropriately parameterized queries.

Use database-supported timestamp comparisons.

Avoid converting timestamps unnecessarily.

Support documented patterns such as:

- created after
- created before
- completed after
- status + time range

Do not implement pagination unless the architecture or existing conventions require it.

---

# 19. Pagination

Determine whether pagination is required by:

- docs/domain-model.md
- docs/repository-layout.md
- existing project conventions

If pagination is required:

- use deterministic ordering
- use explicit limit/offset or documented cursor strategy
- avoid ambiguous ordering
- ensure stable results

Do not introduce cursor pagination simply because it is fashionable.

---

# 20. Query Security

All queries must use parameters.

Do not concatenate user-controlled values into SQL.

Pay particular attention to:

- search
- filtering
- ordering
- pagination
- provider references

If dynamic ordering is required, use a safe documented approach.

Do not expose arbitrary SQL identifiers from request values.

---

# 21. SQL Transactions

Determine whether any SQLC query groups need to be executed inside database transactions.

Do not create business transactions in this agent.

The generated SQLC code should support the project's existing transaction pattern.

Follow the Deposits implementation where applicable.

---

# 22. SQLC Configuration

Create the Transactions SQLC configuration only if the repository structure requires a dedicated configuration.

Follow:

docs/repository-layout.md

and the existing Deposits SQLC setup.

The configuration must define:

- schema
- queries
- engine
- package
- output directory
- driver
- generation options

Do not invent configuration fields unsupported by the pinned SQLC version.

---

# 23. SQLC Output Location

Generated SQLC output must be placed exactly where:

docs/repository-layout.md

requires it.

If the intended layout is:

transactions/db/sqlc/

use that layout.

Do not place generated code into:

deposits/db/sqlc/

Do not modify existing Deposits generated code.

---

# 24. Generated Code Rule

Generated SQLC code is generated output.

Do not manually edit:

- generated models
- generated query methods
- generated interfaces
- generated type definitions

If generated output is incorrect:

fix the source schema/query/configuration and regenerate.

---

# 25. SQLC Generation

Run SQLC using the project's pinned version and established command.

Do not upgrade SQLC.

Do not install an unrelated version globally.

Use the repository's documented generation mechanism.

After generation, verify:

- generated package compiles
- generated models match schema
- generated methods correspond to queries
- nullable types are correct
- parameter types are correct
- result types are correct

---

# 26. Generated Code Review

Inspect generated output only to verify correctness.

Do not manually rewrite generated files.

Check:

- package name
- imports
- model fields
- query method names
- parameter structs
- result structs
- database interfaces
- transaction support

---

# 27. Query Coverage Review

Compare the SQL query set against the database review.

Every important persistence operation identified in:

docs/transactions-database-review.md

should either:

- have a SQLC query
- explicitly not require one yet

Do not generate speculative queries.

---

# 28. Repository Preparation

The SQLC layer will eventually be consumed by Agent 05.

Make sure the generated API is suitable for a repository layer.

The repository should be able to use:

- generated Queries
- generated models
- generated parameter types
- generated result types
- transaction/database interfaces

Do not create the repository itself.

---

# 29. Existing Deposits Compatibility

If the new Transactions SQLC layer is intentionally derived from Deposits patterns:

preserve the useful conventions.

However:

DO NOT copy obsolete Deposits SQL.

DO NOT copy obsolete tables.

DO NOT preserve architecture that the foundation documents explicitly replaced.

The target architecture wins.

---

# 30. Build Validation

Run the relevant Go validation for the Transactions SQLC package.

At minimum verify:

- generated package compiles
- imports resolve
- types compile
- SQLC output is valid Go

Do not run unrelated long-running repository processes.

Do not run deployment workflows.

---

# 31. Generation Reproducibility

Run the SQLC generation command again.

Confirm that generation is deterministic.

After regeneration:

git diff

should not show unexpected changes.

If repeated generation produces different output:

investigate before completion.

Do not suppress the problem.

---

# 32. Generated Output Stability

Check that generated output does not contain:

- timestamps
- local filesystem paths
- machine-specific values
- random identifiers
- environment-specific data

Generated code must be reproducible in CI.

---

# 33. Documentation

Create:

docs/transactions-sqlc-review.md

This is mandatory.

The document must explain:

- SQLC version used
- SQLC configuration
- query organization
- generated output location
- entities covered
- important query operations
- nullable type decisions
- pagination decisions
- transaction support
- unresolved issues

---

# 34. Required Documentation Structure

Use exactly this structure:

# Transactions SQLC Implementation Review

## 1. Source Documents

List the documents used.

## 2. SQLC Version

Document the exact version used.

## 3. SQLC Configuration

Describe the configuration.

## 4. Query Organization

Describe query files and grouping.

## 5. Generated Output

Document the generated package and location.

## 6. Merchant Queries

Summarize implemented operations.

## 7. Customer Queries

Summarize implemented operations.

## 8. Transaction Queries

Summarize implemented operations.

## 9. Deposit Queries

Summarize implemented operations.

## 10. Payout Queries

Summarize implemented operations.

## 11. Idempotency

Document how it is represented if applicable.

## 12. Pagination

Document the implemented approach if applicable.

## 13. Transaction Support

Document database transaction compatibility.

## 14. Validation

Document commands and results.

## 15. Risks

Document risks.

## 16. Unresolved Questions

Document unresolved issues.

---

# 35. Diff Review

Before completion run:

git status --short

Then:

git diff --stat

Review the complete diff.

The expected changes should be limited to:

- Transactions SQL query files
- Transactions SQLC configuration if required
- generated Transactions SQLC output
- docs/transactions-sqlc-review.md

Do not modify:

- Deposits SQLC output
- Deposits migrations
- Deposits repositories
- protobuf files
- Clients
- unrelated services

---

# 36. Generated Code Check

If the project has a generated-code verification mechanism:

run it.

If the project has a Makefile target for SQLC:

prefer that target.

Do not invent a second generation mechanism if one already exists.

---

# 37. Documentation Check

Before completing, verify that:

- README.md was read
- agents/project-context.md was read
- docs/domain-model.md was read
- docs/repository-layout.md was read
- docs/protobuf-strategy.md was read
- docs/migration-plan.md was read
- docs/transactions-existing-review.md was read
- docs/transactions-database-review.md was read

Verify the generated SQLC implementation follows those documents.

---

# Completion Checklist

Before stopping, confirm:

- [ ] Existing Deposits SQLC conventions were reviewed.
- [ ] Transactions database schema was reviewed.
- [ ] SQLC version was verified.
- [ ] Transactions SQL queries were created.
- [ ] Queries correspond to actual Transactions tables.
- [ ] Queries use parameters safely.
- [ ] Query naming follows project conventions.
- [ ] Nullable fields are handled correctly.
- [ ] Monetary values preserve database representation.
- [ ] Status queries are appropriately scoped.
- [ ] Provider references are queried according to database constraints.
- [ ] Idempotency queries exist if required.
- [ ] Pagination was implemented only if required.
- [ ] SQLC configuration follows project conventions.
- [ ] Generated code was produced using the pinned version.
- [ ] Generated code was not manually modified.
- [ ] Generated code compiles.
- [ ] Generation is reproducible.
- [ ] docs/transactions-sqlc-review.md was created.
- [ ] No unrelated files were modified.
- [ ] Existing Deposits SQLC output was not modified.

---

# Final Stop Condition

STOP after completing:

1. Transactions SQL query definitions
2. SQLC configuration if required
3. SQLC generated output
4. SQLC validation
5. docs/transactions-sqlc-review.md

Do NOT:

- create repositories
- create repository interfaces
- create services
- create protobufs
- create handlers
- create runtime wiring
- create Dockerfiles
- create new tests beyond SQLC/build validation
- modify Deposits
- modify Clients
- begin Agent 04

Agent 04 will handle the Transactions protobuf contracts.