# Agent 08 — Webhook Implementation

## Objective

Implement the webhook lifecycle for the Clients Service.

This agent owns every aspect of webhook communication between external providers and RVPay.

The webhook layer is responsible for:

- webhook registration
- webhook verification
- signature validation
- payload parsing
- event routing
- idempotency
- persistence of webhook metadata
- dispatching validated events into the application layer

This task ends after webhook handling is fully implemented and integrated into the Clients service.

Do not implement runtime wiring.

Do not implement gRPC server startup.

Do not implement payment or transaction business logic.

Do not modify protobuf contracts unless absolutely required.

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

Do not recursively inspect the repository.

Only inspect the directories listed below.

---

# Repository Review Scope

Inspect only:

clients/

protobuf/

grpc/go/

Review only enough files to understand:

- OAuth implementation

- repository usage

- service implementation

- protobuf mappings

- logging conventions

- configuration loading

- dependency injection

Do not inspect:

cmd/

third_party/

vendor/

tests/

.github/

coverage/

tmp/

bin/

---

# Responsibilities

Implement the webhook lifecycle.

Support:

Webhook registration

Webhook updates

Webhook deletion

Webhook verification

Signature validation

Payload parsing

Payload validation

Duplicate detection

Event persistence

Retry handling

Dispatch to business services

Health monitoring

Future providers should require implementing only provider-specific adapters.

---

# Provider Architecture

Implement provider webhooks through interfaces.

Create abstractions similar to:

WebhookProvider

WebhookValidator

WebhookDispatcher

WebhookRegistry

The Clients service must never contain provider-specific branching.

Avoid:

if provider == HighLevel

Provider implementations belong behind interfaces.

---

# HighLevel Webhook Provider

Implement webhook support for HighLevel.

Support:

Webhook registration

Webhook verification

Webhook deletion

Signature validation

Payload parsing

Provider event normalization

Map provider events into internal domain events.

No provider-specific payloads should escape the provider package.

---

# Incoming Webhook Flow

Implement the following flow.

Receive request

↓

Validate HTTP method

↓

Validate headers

↓

Identify provider

↓

Verify webhook signature

↓

Validate payload

↓

Detect duplicate event

↓

Persist webhook metadata

↓

Convert provider payload into domain event

↓

Dispatch to appropriate business service

↓

Return provider response

Abort immediately if validation fails.

---

# Signature Verification

Implement provider signature verification.

Never trust incoming payloads.

Validate:

signature

timestamp

provider identity

request integrity

Reject invalid requests immediately.

---

# Payload Parsing

Convert provider payloads into internal domain models.

Avoid leaking provider payload structures.

Normalize all incoming webhook events.

The remainder of the system should consume provider-independent events.

---

# Event Dispatch

Dispatch validated events into application services.

The webhook layer should never execute business logic directly.

Examples:

Client installed

Integration removed

OAuth revoked

Token expired

Provider disconnected

Dispatch only.

Business services remain responsible for business rules.

---

# Idempotency

Implement duplicate detection.

Receiving the same webhook multiple times must not create duplicate business operations.

Persist provider event identifiers where appropriate.

Handle retries safely.

---

# Retry Strategy

Implement retry-safe processing.

The webhook layer should tolerate:

duplicate deliveries

provider retries

network interruptions

temporary persistence failures

Do not assume providers deliver events exactly once.

---

# Security

Validate:

HTTP method

provider signature

content type

timestamp

payload integrity

Reject malformed requests immediately.

Never expose provider secrets.

Never trust provider payloads before validation.

---

# Logging

Mirror Deposits logging conventions.

Log:

webhook received

provider identified

validation passed

validation failed

duplicate ignored

event dispatched

Never log:

provider secrets

OAuth tokens

refresh tokens

client secrets

raw webhook payloads containing sensitive information

signature secrets

---

# Error Handling

Translate provider errors into business errors.

Return provider-appropriate HTTP responses.

Do not expose:

stack traces

database errors

provider SDK errors

internal implementation

---

# Configuration

Load webhook configuration from environment variables.

Never hardcode:

webhook secret

callback URL

provider endpoints

signature keys

Future providers should require configuration only.

---

# Deliverables

Implement:

clients/webhooks/

clients/providers/

(or equivalent directories matching repository conventions)

Only create files required for webhook functionality.

Do not modify unrelated services.

---

# Validation

Before completing verify:

- webhook registration works

- webhook verification succeeds

- signature validation succeeds

- payload parsing succeeds

- duplicate detection works

- retries are safe

- provider abstraction remains intact

- repositories remain abstracted

- business logic remains outside webhook handlers

- project compiles

---

# Success Criteria

The Clients Service should now support secure provider webhooks through a provider-independent architecture.

Future providers should require only new provider implementations.

Business services should receive normalized events regardless of provider.

---

# Completion Rules

Before completing verify:

- Existing coding conventions have been preserved.

- Existing logging conventions have been preserved.

- Existing dependency injection conventions have been preserved.

- Existing repository abstractions have been respected.

- Existing protobuf contracts have not been modified unnecessarily.

- No unrelated directories have been modified.

- The project builds successfully.

If a prerequisite from a previous agent is missing, stop and explain why instead of creating a partial implementation.

---

# Webhook Architecture Review

Before completing, perform a complete review of the webhook implementation.

Confirm:

- providers remain interface-driven

- webhook validation is secure

- signature verification is isolated

- payload normalization is provider-independent

- business logic remains outside webhook handlers

- duplicate detection is implemented

- retries are safe

- dispatch responsibilities are clearly separated

- repositories remain persistence-only

- no circular dependencies exist

If improvements are discovered that do not require redesigning previous agents, implement them.

If improvements require changes to protobuf contracts, repositories, OAuth abstractions, or database schema, stop and document the architectural issue instead.

Produce:

clients/docs/webhook-review.md

The report should summarize:

- implemented providers

- webhook lifecycle

- validation strategy

- signature verification strategy

- event normalization

- dispatch architecture

- retry strategy

- future provider extensibility

- remaining work before runtime implementation

Only after this review is complete should the project proceed to Agent 09.