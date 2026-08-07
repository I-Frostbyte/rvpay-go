# Agent 02 — Transactions Database Implementation

## Objective

Implement the PostgreSQL database foundation for the new Transactions Service.

The existing RVPay repository contains a `deposits/` service with an established PostgreSQL implementation.

The new architecture expands this into a Transactions Service responsible for the transaction domain defined by:

- merchants
- customers
- deposits
- payouts
- transaction lifecycle
- transaction state
- transaction-related persistence

This agent is responsible ONLY for the database layer.

It must implement the database schema and migrations required by the documented Transactions architecture.

It must build upon the existing Deposits implementation where appropriate.

It must NOT implement:

- sqlc generation
- repositories
- protobufs
- gRPC services
- REST handlers
- merchants service logic
- customers service logic
- deposits service logic
- payouts service logic
- runtime wiring
- Docker
- tests beyond database/migration verification

Those responsibilities belong to later agents.

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
8. existing deposits/db implementation

Do not invent domain entities that are not supported by these sources.

Do not silently resolve contradictions.

If the documentation and existing implementation disagree:

- identify the difference
- follow the documented target architecture
- preserve existing data where required by the migration plan
- document any unresolved issue

---

# Repository Exploration Rules

Use the root README.md as the repository map.

Do not perform an unrestricted repository-wide exploration.

Inspect only files required to implement the Transactions database layer.

Do not recursively inspect:

- third_party/
- googleapis/
- vendor/
- node_modules/
- .git/
- coverage/
- tmp/
- bin/

Do not spend time exploring generated code unless required to understand an existing database model.

Do not inspect unrelated services.

---

# Existing Deposits Database Rule

The existing:

deposits/db/

is the primary implementation reference.

Inspect only the relevant database files:

- deposits/db/migrations/
- deposits/db/query/
- deposits/db/repo/
- deposits/db/sqlc/
- deposits/db/doc.go
- relevant migration/configuration files

Do not modify existing Deposits migrations.

Do not rewrite existing Deposits schema history.

Do not rename existing Deposits tables unless explicitly required by:

docs/migration-plan.md

---

# Documentation Check

Before implementing anything, verify that you have read:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md
- docs/transactions-existing-review.md

At the end of the task, verify that the implementation agrees with those documents.

If the implementation requires a deviation, document it.

---

# 1. Determine Transactions Database Ownership

Using:

docs/domain-model.md

determine which entities belong to Transactions.

At minimum, evaluate the documented ownership of:

- merchants
- customers
- transactions
- deposits
- payouts

Do not assume that every concept listed above must become a table.

Only create tables supported by the domain model.

Determine:

- primary ownership
- relationships
- lifecycle
- identifiers
- status/state

---

# 2. Review Existing Deposits Schema

Before creating new migrations, inspect the existing Deposits schema.

Determine:

- existing tables
- existing columns
- primary keys
- foreign keys
- indexes
- constraints
- status fields
- timestamps
- provider identifiers
- external references
- amount/currency representation
- metadata fields

Map each relevant concept to the new Transactions architecture.

Do not modify the existing Deposits migration history.

---

# 3. Build the Migration Map

Create a mental/data-model mapping:

Existing Deposits
        ↓
Transactions Target Schema

Classify existing database concepts as:

- KEEP
- EXTEND
- REPLACE
- MIGRATE
- DEPRECATE
- NEW

Only implement changes supported by:

docs/migration-plan.md

If the migration plan does not specify a required migration, do not invent a destructive migration strategy.

---

# 4. Transactions Schema

Create the Transactions database schema according to:

docs/domain-model.md

and:

docs/migration-plan.md

The schema must represent the documented transaction lifecycle.

Do not introduce additional lifecycle states without documentation support.

---

# 5. Identifiers

Review identifier requirements carefully.

Determine appropriate identifiers for:

- merchants
- customers
- transactions
- deposits
- payouts
- provider references

Use the existing repository conventions unless the target architecture explicitly changes them.

Do not introduce multiple identifiers for the same purpose without a documented reason.

Distinguish clearly between:

- internal primary keys
- public identifiers
- provider identifiers
- external transaction references

---

# 6. Merchant Database Model

If merchants are owned by Transactions according to the domain model, implement the required merchant persistence model.

Verify:

- identity
- unique constraints
- lifecycle state
- timestamps
- external references
- configuration fields

Do not put Clients-service OAuth data into the merchant table unless the architecture explicitly requires it.

The Clients service owns client/integration concerns.

Transactions owns transaction-domain concerns.

Respect this boundary.

---

# 7. Customer Database Model

If customers are owned by Transactions according to the domain model, implement the required customer persistence model.

Verify:

- identity
- uniqueness
- merchant relationship
- timestamps
- required contact/payment identifiers
- provider/external references where documented

Do not store unnecessary sensitive data.

Do not create fields merely because they might be useful in the future.

---

# 8. Transaction Database Model

Implement the core transaction persistence model defined by the architecture.

Verify support for:

- transaction identity
- merchant ownership
- customer relationship where applicable
- transaction type
- transaction direction
- transaction status
- amount
- currency
- provider information
- external/provider reference
- timestamps
- failure information where documented
- metadata where documented

Do not invent additional transaction attributes.

---

# 9. Deposit Database Model

The existing Deposits implementation is being absorbed into the broader Transactions domain.

Determine from:

docs/migration-plan.md

which Deposits data remains necessary.

Do not blindly duplicate the existing Deposits table.

Determine whether Deposits should:

- remain as a transaction subtype
- remain as a related table
- be represented by transaction fields
- be migrated into another structure

Follow the documented architecture.

Do not make this architectural decision independently.

---

# 10. Payout Database Model

If Payouts are defined by the domain model as part of Transactions, implement their persistence requirements.

Verify:

- relationship to transaction
- provider reference
- amount
- currency
- destination/reference information
- status
- timestamps
- failure information

Do not implement payout business logic.

This agent only establishes persistence.

---

# 11. Relationships

Define foreign keys according to documented ownership.

Review relationships between:

merchant

customer

transaction

deposit

payout

Ensure:

- references are valid
- deletion behavior is intentional
- orphan records cannot occur accidentally
- cascading behavior is intentional

Do not use cascading deletion merely for convenience.

Transaction history is generally important data.

Follow the migration/domain documentation.

---

# 12. Status Fields

Review every status/state field.

Verify:

- allowed states are documented
- invalid values are constrained where appropriate
- state transitions can be represented
- terminal states can be represented
- failure states can preserve required information

Do not invent state-machine behavior in the database.

The database should enforce structural validity, while business services will enforce transition rules.

---

# 13. Amount and Currency

Follow the existing project's conventions and the domain documentation for:

- monetary values
- currency
- precision
- storage type

Do not use floating-point types for monetary values unless the existing documented architecture explicitly requires them.

Maintain consistent representation across Transactions tables.

---

# 14. Timestamps

Follow existing repository conventions.

Determine whether timestamps are:

- created_at
- updated_at
- completed_at
- failed_at
- initiated_at

Only create timestamps supported by the domain model.

Ensure database defaults are deterministic.

Prefer database-side timestamp defaults where that matches existing repository conventions.

---

# 15. Provider References

Transactions may interact with external payment providers.

Store provider references only where the Transactions domain requires them.

Clearly distinguish:

internal transaction ID

provider transaction ID

provider name/type

external reference

Do not duplicate provider credentials.

Provider credentials belong in configuration/secret management, not transaction records.

---

# 16. Idempotency

Review the architecture for idempotency requirements.

Determine whether the Transactions database requires an idempotency key or equivalent unique constraint.

If the architecture requires idempotency:

- implement the appropriate field
- add the appropriate uniqueness constraint
- ensure the constraint is scoped correctly

Do not implement application-level idempotency logic in this agent.

---

# 17. Webhook/External Event Data

If the domain or migration documentation requires persistence of external events, implement only the documented persistence model.

Potential concerns include:

- provider event ID
- event type
- received timestamp
- processed timestamp
- processing state
- raw payload reference

Do not store raw sensitive payloads unnecessarily.

Do not create webhook infrastructure here.

---

# 18. Indexing

Create indexes for documented/high-value access patterns.

At minimum evaluate:

- merchant lookups
- customer lookups
- transaction ID lookups
- provider reference lookups
- status queries
- time-based transaction queries
- idempotency lookups

Do not create indexes on every column.

Every index should support an identifiable query pattern.

---

# 19. Constraints

Use database constraints to enforce structural invariants.

Consider:

- NOT NULL
- UNIQUE
- FOREIGN KEY
- CHECK
- appropriate defaults

Do not encode complex business rules as database constraints unless explicitly required.

The service layer will enforce business behavior.

---

# 20. Migration Design

Create migrations that are:

- sequential
- deterministic
- reversible where project conventions require down migrations
- safe
- clearly named
- isolated by logical change

Follow the existing migration naming convention.

Do not rewrite historical migrations.

Do not edit already-existing Deposits migrations.

Create new Transactions migration files for the target schema.

---

# 21. Migration Ordering

Ensure migrations are ordered according to dependencies.

For example:

1. foundational entities
2. merchants/customers
3. transactions
4. deposits/payout relationships
5. indexes/constraints
6. supporting structures

Use the actual domain dependencies from the documentation rather than blindly following this example.

---

# 22. Down Migrations

Where the project uses down migrations:

- provide a corresponding down migration
- remove objects in reverse dependency order
- avoid leaving orphaned database objects
- ensure rollback does not fail because of dependency ordering

Do not use destructive down migrations that violate the documented migration strategy.

---

# 23. Migration Safety

Review for:

- accidental data loss
- unsafe DROP statements
- unsafe ALTER statements
- duplicate constraints
- duplicate indexes
- missing foreign keys
- incorrect dependency ordering

If an existing production table would require destructive modification:

DO NOT perform it automatically.

Document the required migration strategy.

---

# 24. Existing Data Compatibility

If the migration plan requires existing Deposits data to survive:

ensure the new schema can accommodate that data.

Do not write a data migration unless this agent's scope explicitly includes it in:

docs/migration-plan.md

If a data migration is required but outside the scope of this agent:

document it in:

docs/transactions-database-review.md

---

# 25. Database Package Structure

Follow:

docs/repository-layout.md

and the established Deposits layout.

The expected structure should be determined from the documentation.

Do not invent a new database package structure.

If the documentation specifies:

transactions/db/migrations/

transactions/db/query/

transactions/db/repo/

transactions/db/sqlc/

follow that structure exactly.

---

# 26. Existing Coding Conventions

Follow:

agents/project-context.md

for:

- package names
- filenames
- comments
- SQL formatting
- naming
- error handling
- migration style
- Go conventions

Do not introduce a different style.

---

# 27. Implementation Restrictions

Do NOT:

- implement repositories
- generate sqlc
- write service logic
- create protobufs
- create gRPC handlers
- create REST handlers
- implement merchants logic
- implement customers logic
- implement deposits logic
- implement payouts logic
- modify Clients
- modify existing Deposits runtime
- redesign service boundaries

Only database implementation belongs to this agent.

---

# 28. Validation

After creating the database migrations:

Verify migration syntax.

Use the project's existing migration tooling.

If a local PostgreSQL environment is available, perform a clean migration test.

Verify:

1. database starts
2. migrations apply
3. schema is created
4. constraints exist
5. indexes exist
6. foreign keys exist
7. down migrations work if supported
8. migrations can be reapplied cleanly

Do not require a production database.

Do not connect to production infrastructure.

---

# 29. Schema Inspection

After applying migrations, inspect the resulting schema.

Verify:

- expected tables exist
- expected columns exist
- expected types exist
- expected constraints exist
- expected indexes exist
- relationships are correct

Do not rely solely on migration files.

Verify the resulting database state where possible.

---

# 30. Documentation

Create:

docs/transactions-database-review.md

This document must explain:

- what was implemented
- what existing Deposits schema was reused
- what was changed
- what was newly introduced
- migration ordering
- important constraints
- important indexes
- unresolved migration concerns

---

# Required Documentation Structure

The review must contain:

# Transactions Database Implementation Review

## 1. Source Documents

List the required documents used.

## 2. Existing Deposits Database

Summarize the relevant existing schema.

## 3. Target Transactions Schema

Summarize the implemented schema.

## 4. Entity Mapping

Use:

| Existing Deposits Concept | Transactions Concept | Action |
|---|---|---|

Actions:

- Reuse
- Extend
- Migrate
- Replace
- New
- Deprecated

## 5. Tables Created

List every new table.

## 6. Relationships

Describe foreign keys and ownership.

## 7. Constraints

List important constraints.

## 8. Indexes

List indexes and the query patterns they support.

## 9. Migrations

List migration files in execution order.

## 10. Data Migration Considerations

Document any required future migration.

## 11. Risks

Document database risks.

## 12. Unresolved Questions

Document anything that cannot be safely determined.

---

# 31. Final Diff Review

Before completing:

Run:

git status --short

Then:

git diff --stat

Review all changed files.

Ensure changes are limited to:

- Transactions database files
- required documentation

Do not modify unrelated services.

Do not modify existing Deposits migrations.

Do not modify generated code.

---

# 32. Documentation Check

Before completion, verify again:

- README.md was read
- agents/project-context.md was read
- docs/domain-model.md was read
- docs/repository-layout.md was read
- docs/protobuf-strategy.md was read
- docs/migration-plan.md was read
- docs/transactions-existing-review.md was read

Verify the implementation agrees with these documents.

If it does not, document the discrepancy.

---

# Completion Checklist

Before stopping, confirm:

- [ ] Existing Deposits database was inspected.
- [ ] Transactions database ownership was determined from documentation.
- [ ] Existing schema was mapped to target schema.
- [ ] Transactions migrations were created.
- [ ] Historical Deposits migrations were not modified.
- [ ] Foreign keys are correct.
- [ ] Constraints are correct.
- [ ] Indexes support identifiable queries.
- [ ] Monetary fields follow project conventions.
- [ ] Status fields follow the domain model.
- [ ] Idempotency requirements were addressed where documented.
- [ ] Migration ordering is correct.
- [ ] Down migrations exist where required.
- [ ] Migration safety was reviewed.
- [ ] Database validation was performed where possible.
- [ ] docs/transactions-database-review.md was created.
- [ ] No unrelated files were modified.

---

# Final Stop Condition

STOP after completing the database implementation and review.

Do not:

- generate sqlc
- create repositories
- create protobufs
- create services
- create handlers
- implement runtime
- create tests beyond database validation
- modify Deposits
- begin Agent 03

Agent 03 will consume the database schema and generate/verify the SQLC layer.