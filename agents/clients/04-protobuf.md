# Agent 04 — Protobuf Contracts & API Design

## Objective

Design and implement the protobuf contracts for the Clients Service.

This agent owns the external API contract.

The protobuf definitions created by this agent become the public interface of the Clients Service and must remain stable.

Every future service must communicate with the Clients Service exclusively through these contracts.

The protobufs generated here should follow the architecture defined by:

- docs/domain-model.md
- docs/protobuf-strategy.md

while matching the repository conventions documented in:

- README.md
- agents/project-context.md

This task includes:

- protobuf definitions
- gRPC services
- request messages
- response messages
- shared enums
- grpc-gateway HTTP annotations
- protobuf generation
- generated Go code

This task ends after protobuf generation succeeds.

Do not implement repositories.

Do not implement business logic.

Do not implement runtime wiring.

Do not implement OAuth.

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

Only inspect the directories listed below.

---

# Repository Review Scope

Inspect only:

protobuf/

grpc/go/

deposits/

Inspect enough files to understand:

- protobuf naming

- package naming

- grpc-gateway annotations

- file organization

- Makefile conventions

- protoc generation

- generated output layout

- HTTP endpoint conventions

Do not inspect:

clients/db/

third_party/

tests/

.github/

vendor/

node_modules/

coverage/

tmp/

bin/

---

# Existing Conventions

Mirror the Deposits service.

Do not invent a new protobuf style.

Reuse:

package names

option go_package

HTTP annotation style

comment style

directory layout

generation workflow

Makefile conventions

grpc-gateway conventions

---

# Service Responsibilities

The Clients Service owns:

Clients

Platforms

Integrations

OAuth lifecycle

Webhook registration

No transaction-related RPCs belong here.

---

# API Design Goals

Design a clean, provider-agnostic API.

The service must support future providers without changing RPC contracts.

Avoid provider-specific messages.

Provider-specific logic belongs in implementation.

The protobuf API must remain generic.

---

# Services

Generate the following gRPC services.

## ClientsService

Responsible for:

CreateClient

UpdateClient

DeleteClient

GetClient

ListClients

ActivateClient

DeactivateClient

---

## PlatformsService

Responsible for:

ListPlatforms

GetPlatform

EnablePlatform

DisablePlatform

Future providers should require database changes only.

Avoid hardcoded provider RPCs.

---

## IntegrationsService

Responsible for:

InstallIntegration

UninstallIntegration

GetIntegration

ListIntegrations

ReconnectIntegration

DisconnectIntegration

SyncIntegration

---

# Messages

Design request and response messages for every RPC.

Every message should:

be explicit

avoid nested complexity

avoid nullable ambiguity

prefer repeated fields over maps

use UUID strings consistently

use Timestamp for temporal fields

use enums for finite state

avoid provider-specific fields

---

# Shared Enums

Create reusable enums for:

ClientStatus

PlatformStatus

IntegrationStatus

OAuthStatus

WebhookStatus

ProviderCapability

Future services should reuse these enums whenever appropriate.

Avoid duplicate enum definitions.

---

# Shared Messages

Create reusable messages for:

PaginationRequest

PaginationResponse

Error

Metadata

AuditInformation

HealthStatus

Only introduce shared messages that are genuinely reusable.

Avoid premature abstraction.

---

# HTTP Gateway

Annotate every public RPC using grpc-gateway.

Follow the same conventions used by Deposits.

Design REST endpoints that are predictable.

Example patterns:

GET

POST

PATCH

DELETE

Avoid RPC-style URLs.

Prefer resource-oriented endpoints.

---

# Validation

Design protobuf fields suitable for future validation.

Do not implement validation.

Reserve field numbering carefully.

Never reuse field numbers.

Leave room for future expansion.

---

# Versioning

Use stable package names.

Avoid experimental suffixes.

Do not introduce v2 unless required.

Document reserved field numbers where appropriate.

Reserve deleted fields.

Never renumber protobuf fields.

---

# Package Layout

Generate protobufs under:

protobuf/

Generated Go code should be written into:

grpc/go/

using the existing repository conventions.

Do not manually edit generated files.

---

# Generation

Update generation tooling only if required.

Reuse existing Makefile targets whenever possible.

Do not create duplicate generation scripts.

Generation must produce:

protobuf definitions

Go message structs

gRPC server interfaces

grpc-gateway bindings

OpenAPI output (if already supported by repository)

---

# Documentation

Every service.

Every RPC.

Every message.

Every enum.

Every field.

must include meaningful protobuf comments.

Comments should explain business purpose.

Do not generate placeholder comments.

---

# API Principles

Every RPC should be:

idempotent where appropriate

predictable

strongly typed

provider agnostic

future extensible

Every response should expose only business information.

Never leak database implementation.

Never expose encrypted values.

Never expose internal identifiers unless required.

---

# Deliverables

Create or update:

protobuf/

grpc/go/

Update protobuf generation only if necessary.

Do not modify unrelated services.

---

# Validation

Before completing verify:

- protoc generation succeeds

- grpc generation succeeds

- grpc-gateway generation succeeds

- generated Go code compiles

- no generated files were manually edited

- every RPC has documentation

- every message has documentation

- every enum has documentation

- every REST annotation is valid

- protobuf packages compile

- imports are correct

- no circular protobuf imports exist

---

# Success Criteria

The Clients Service should expose a complete, stable, provider-agnostic public API suitable for production use.

Future providers should integrate without requiring protobuf redesign.

No business logic should exist after this task.

No repositories should be modified.

No runtime code should exist.

---

# Completion Rules

Before completing verify:

- Existing protobuf conventions have been preserved.

- Existing grpc-gateway conventions have been preserved.

- Existing package naming has been preserved.

- Existing generation workflow has been preserved.

- Existing Makefile conventions have been preserved.

- Generated files have not been manually edited.

- Project builds successfully.

- No unrelated directories were modified.

If any prerequisite is missing, stop and explain why instead of producing a partial implementation.