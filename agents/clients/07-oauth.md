# Agent 07 — OAuth Implementation

## Objective

Implement the OAuth functionality for the Clients Service.

This agent is responsible for implementing the complete OAuth lifecycle for provider integrations.

The implementation must follow the provider abstraction defined in:

- docs/domain-model.md
- docs/protobuf-strategy.md

while adhering to the coding standards documented in:

- README.md
- agents/project-context.md

This task includes:

- OAuth provider interfaces
- OAuth service implementation
- authorization URL generation
- callback handling
- access token exchange
- refresh token handling
- token persistence
- provider registration

This task ends after OAuth flows compile successfully.

Do not implement webhook processing.

Do not implement runtime wiring.

Do not modify protobuf contracts unless absolutely necessary.

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

clients/

deposits/

protobuf/

grpc/go/

Review only enough files to understand:

- service layout

- repository usage

- configuration conventions

- dependency injection

- protobuf mappings

- error handling

- logging conventions

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

Implement the OAuth lifecycle.

Support:

Authorization URL generation

Authorization callback processing

Authorization code exchange

Refresh token exchange

Token persistence

Token validation

Token expiration detection

Token revocation

Provider registration

The implementation should support adding future providers without architectural redesign.

---

# Provider Architecture

OAuth providers must be implemented behind interfaces.

Never hardcode provider logic into the Clients service.

Create provider abstractions similar to:

Provider

OAuthProvider

ProviderRegistry

The service layer should communicate only with interfaces.

---

# HighLevel Provider

Implement the first provider:

HighLevel

Support:

Generate authorization URL

Exchange authorization code

Retrieve access token

Retrieve refresh token

Retrieve expiry

Retrieve provider account identifier

Persist OAuth credentials using repositories.

No provider-specific logic should leak outside the provider implementation.

---

# OAuth Flow

Implement the following sequence:

Generate authorization URL

↓

Redirect user

↓

Receive callback

↓

Validate callback

↓

Exchange authorization code

↓

Persist OAuth credentials

↓

Create Integration

↓

Return successful installation

Rollback persistence if any step fails.

---

# Token Storage

Persist:

Access Token

Refresh Token

Expiry

Scope

Provider Account ID

Never store plaintext secrets in logs.

Never expose tokens through protobuf responses.

Repositories remain responsible for persistence.

---

# Token Refresh

Implement refresh support.

Automatically determine when refresh is required.

Refresh only when necessary.

Persist refreshed credentials.

Handle refresh failures gracefully.

---

# Security

Validate:

state

provider identity

redirect URI

authorization code

Never trust callback parameters.

Prevent replay attacks where possible.

Do not expose implementation secrets.

---

# Configuration

Load provider configuration from environment variables.

Never hardcode:

Client ID

Client Secret

Redirect URL

OAuth endpoints

Scopes

Future providers should require configuration only.

---

# Error Handling

Translate provider errors into business errors.

Do not expose:

HTTP client errors

provider JSON

provider SDK errors

authentication secrets

Errors should remain meaningful to API consumers.

---

# Logging

Mirror Deposits logging conventions.

Log:

OAuth start

OAuth completion

provider selected

token refresh

integration installation

Never log:

access tokens

refresh tokens

client secrets

authorization codes

provider secrets

webhook secrets

---

# Dependency Rules

OAuth implementation may depend on:

repositories

configuration

logger

HTTP client

provider interfaces

standard library

Do not depend directly on:

gRPC runtime

database packages

sqlc

transport implementation

---

# Deliverables

Implement:

clients/providers/

clients/oauth/

(or equivalent directories matching repository conventions)

Only create files required for OAuth implementation.

Do not modify unrelated services.

---

# Validation

Before completing verify:

- authorization URL generation works

- callback processing works

- state validation succeeds

- token exchange succeeds

- refresh flow succeeds

- token persistence succeeds

- repositories remain abstracted

- providers remain interchangeable

- configuration loads correctly

- project compiles

---

# Success Criteria

The Clients service should support installing a provider through OAuth without exposing provider implementation details.

Adding a second provider should require implementing only another provider adapter.

No runtime wiring should exist after this task.

No webhook processing should exist after this task.

---

# Completion Rules

Before completing verify:

- Existing coding conventions have been preserved.

- Existing dependency injection conventions have been preserved.

- Existing repository abstractions have been respected.

- Existing logging conventions have been preserved.

- Existing protobuf contracts have not been modified unnecessarily.

- No unrelated directories have been modified.

- The project builds successfully.

If a prerequisite from a previous agent is missing, stop and explain why instead of creating a partial implementation.

---

# OAuth Architecture Review

Before completing, perform a comprehensive review of the OAuth implementation.

Confirm:

- providers are interface-driven

- HighLevel implementation remains isolated

- service layer remains provider agnostic

- repositories own persistence

- OAuth secrets are never logged

- configuration is environment-driven

- token refresh is encapsulated

- callback validation is secure

- provider registry supports future providers

- no circular dependencies exist

If improvements are discovered that do not require redesigning previous agents, implement them.

If improvements require changes to protobuf contracts, repositories, or database schema, stop and document the architectural issue instead.

Produce:

clients/docs/oauth-review.md

The report should summarize:

- implemented OAuth providers

- provider abstraction strategy

- callback lifecycle

- token lifecycle

- refresh strategy

- security considerations

- extensibility for future providers

- remaining work before webhook implementation

Only after this review is complete should the project proceed to Agent 08.