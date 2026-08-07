# Agent 02 — Clients Database Design

## Objective

Design and implement the PostgreSQL database schema required for the new Clients Service.

This agent is responsible only for the database layer.

It must create a production-ready relational schema that supports the domain model defined in:

- docs/domain-model.md

while following the repository conventions documented in:

- README.md
- agents/project-context.md

This task ends after database migrations have been created.

Do not implement repositories.

Do not generate sqlc.

Do not generate protobufs.

Do not write Go business logic.

Do not modify existing services.

---

# Required Reading

Read only:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md

Use these documents as the architectural source of truth.

Do not inspect the repository recursively.

Only inspect directories listed below.

---

# Repository Review Scope

Inspect only:

clients/

deposits/db/migrations/

Inspect enough files to understand:

- migration naming convention

- SQL style

- timestamp conventions

- UUID conventions

- foreign key conventions

- indexing conventions

Do not inspect:

grpc/go/

third_party/

.github/

tests/

vendor/

node_modules/

coverage/

bin/

tmp/

---

# Existing Conventions

Mirror the style used by the Deposits service.

Do not invent a new migration style.

Use:

- UUID primary keys

- timestamptz

- snake_case

- explicit foreign keys

- explicit indexes

- NOT NULL where appropriate

- ENUM types where appropriate

Prefer PostgreSQL native types.

Do not use JSON columns unless explicitly justified.

---

# Database Responsibilities

Design the persistence model for:

Clients

Platforms

Integrations

OAuth Tokens

Webhook Subscriptions

Future Provider Metadata

The schema should support adding new providers without future schema redesign.

---

# Entity Requirements

## Clients

Store:

- ID

- Name

- Status

- Created At

- Updated At

One client owns many integrations.

---

## Platforms

Represents supported external platforms.

Examples:

HighLevel

HubSpot

Salesforce

Future providers.

Store:

- ID

- Name

- Display Name

- Slug

- Enabled

- OAuth Capability

- Webhook Capability

- Created At

- Updated At

Do not hardcode providers in Go.

Platforms must be data-driven.

---

## Integrations

Represents a Client connected to a Platform.

Store:

- ID

- Client ID

- Platform ID

- External Account ID

- Status

- Installed At

- Last Sync At

- Created At

- Updated At

Use foreign keys.

Prevent duplicate integrations.

A client may only install a provider once unless explicitly allowed.

---

## OAuth Tokens

Store encrypted OAuth credentials.

Never store plaintext secrets.

Store:

- Integration ID

- Access Token

- Refresh Token

- Expiry

- Scope

- Token Type

- Created At

- Updated At

Leave encryption implementation for later agents.

The schema must support encrypted storage.

---

## Webhook Subscriptions

Store:

- Integration ID

- Endpoint

- Secret

- Status

- Last Delivery

- Created At

- Updated At

---

# Relationships

Enforce:

Client

↓

Integration

↓

OAuth Token

Integration

↓

Webhook Subscription

Platform

↓

Integration

Every relationship must use explicit foreign keys.

---

# Constraints

Use:

UNIQUE

CHECK

FOREIGN KEY

NOT NULL

DEFAULT

where appropriate.

Use ON DELETE behavior intentionally.

Document every cascading delete.

---

# Indexes

Create indexes for:

foreign keys

lookup fields

external IDs

status columns

provider lookups

Do not create speculative indexes.

---

# Migration Strategy

Generate:

up migration

down migration

Use the same naming convention used by Deposits.

Down migrations must:

drop constraints

drop indexes

drop tables

drop custom types

in the correct dependency order.

---

# Deliverables

Create:

clients/db/migrations/

Only create:

SQL migration files.

Do not generate any Go code.

---

# Validation

Before completing verify:

- migrations execute cleanly

- migrations rollback cleanly

- foreign keys validate

- indexes are created

- constraints are valid

- no circular references exist

- schema matches the domain model

---

# Success Criteria

The Clients database should be fully implementable by future agents without further schema redesign.

No application code should exist after this task.

---

# Completion Rules

Before completing verify:

- Existing functionality has not been broken.

- Existing migration conventions have been preserved.

- Existing naming conventions have been preserved.

- Existing directory structure has been preserved.

- No generated code has been modified.

- No unrelated files have been edited.

- Only migration files have been created.

- The project still builds.

If any prerequisite is missing, stop and explain why instead of creating a partial implementation.