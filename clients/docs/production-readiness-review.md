# Clients Service — Production Readiness Review

## Executive Summary

The Clients Service implementation is architecturally consistent with the
RVPay repository, functionally complete for its defined scope, and passes all
automated validation. The service builds, tests pass, race detection is clean,
vet passes, and the Docker image builds successfully. The service follows the
Deposits conventions for runtime, configuration, logging, and error handling.

No production blockers were identified. The service is **READY WITH WARNINGS**
— several medium/high-risk items should be addressed before broader production
deployment, primarily around OAuth token encryption at rest, webhook
deduplication, and configuration-driven redirect URIs.

## Architecture

The implementation matches the documented domain model, repository layout, and
protobuf strategy:

- **Domain model** — Implements clients, platforms, integrations, OAuth tokens,
  and webhook subscriptions. No accidental duplication with Deposits.
- **Repository layout** — Mirrors `deposits/` exactly: `cmd/grpc-service/`,
  `config/`, `db/`, `service/`, `docs/`, `Makefile`, `Dockerfile`.
- **Service boundaries** — Clients owns only client/platform/integration/OAuth/
  webhook concepts. No transaction, payment, or deposit logic.
- **Provider architecture** — Unified `Provider` interface with `OAuthProvider`
  and `WebhookProvider` accessors. Registry is thread-safe via `sync.RWMutex`.
- **Protobuf strategy** — Uses `protobuf/clients.proto` as source, generated
  code in `grpc/go/clientsgrpc/`. Contracts unchanged.

## Database

The schema is sound:

- **Schema** — 6 tables (clients, platforms, integrations, oauth_tokens,
  webhook_subscriptions, platform_metadata) with proper enums.
- **Migrations** — Sequential `000001_init_schema.up.sql` and `.down.sql`;
  up/down paths are present and symmetric.
- **Indexes** — Indexes exist for platform enabled, integration platform/external
  account/status, and webhook subscription status.
- **Constraints** — Foreign keys with RESTRICT for integrations, CASCADE for
  oauth_tokens and webhook_subscriptions. Unique constraints enforce:
  - client_name is NOT unique (potential for duplicate client names)
  - platform slug is unique
  - client + platform is unique (prevents duplicate integrations)
  - oauth token per integration is unique
  - webhook per integration+endpoint is unique
- **sqlc** — Generated code corresponds to current SQL; no manual edits.

## Runtime

The runtime mirrors Deposits:

- **Startup** — Deterministic 16-step sequence: signal context → logger → config
  → database → migrations → repositories → providers → services → gRPC →
  gateway → HTTP → servers.
- **Dependency injection** — Explicit constructor injection; every dependency
  constructed exactly once.
- **gRPC** — Standard gRPC with recovery interceptor, health checks, reflection.
- **REST** — grpc-gateway v2.
- **Health** — gRPC health protocol + HTTP `/healthz`.
- **Shutdown** — SIGINT/SIGTERM with 5-second timeout; graceful gRPC and HTTP
  shutdown; db.Close deferred.

## OAuth

- **Flow** — Authorization URL generation, code exchange, token persistence,
  refresh, validation.
- **State handling** — `GenerateState()` creates cryptographically random state.
  Transport-layer state validation is the verified responsibility of the
  runtime (not yet implemented).
- **Token handling** — Persisted via `OAuthTokenRepo`; not returned through
  protobuf responses; never logged.
- **Security** — Client secrets loaded from configuration, never hardcoded.
  Token exchange and refresh errors are handled.

## Webhooks

- **Validation** — Signature validation (HMAC-SHA256) occurs before processing.
  Malformed payloads are rejected. Unknown providers fail safely.
- **Processing** — Parse → normalize → lookup subscription → update delivery →
  dispatch.
- **Idempotency** — Subscription lookup provides basic duplicate detection.
  No `webhook_events` table exists yet for provider event ID deduplication.
- **Security** — Provider secrets never logged; payloads don't leak credentials.

## Testing

- **go test ./clients/...** — PASS (all packages)
- **go test ./...** — PASS (full repository)
- **go vet ./clients/...** — PASS
- **go test -race ./clients/...** — PASS (no race conditions)
- **Integration tests** — No PostgreSQL-backed repository integration tests
  implemented yet.

## Deployment

- **Docker** — Multi-stage build from `clients/Dockerfile` builds successfully
  (verified: `docker build -f clients/Dockerfile -t rvpay-go-clients:review .`).
  Minimal distroless runtime, non-root user, correct entrypoint.
- **Render** — Compatible with environment-variable-driven configuration;
  Render env vars can supply DB, ports, OAuth, webhook secrets.
- **Configuration** — All variables in `.env.example` documented; defaults safe;
  no secrets committed.
- **Migrations** — Applied automatically when `RUN_MIGRATIONS=true`.

## Security

- **Secrets** — No real credentials committed. `.env.example` uses placeholders.
- **OAuth security** — Client secrets never logged; tokens never logged.
- **Webhook security** — HMAC validation before processing.
- **Logging** — Logs contain no OAuth secrets, tokens, passwords, or provider
  credentials.

## Findings

| Severity | Area | Finding | Required Action |
| --- | --- | --- | --- |
| HIGH | OAuth tokens | Tokens stored in plaintext in `oauth_tokens` table | Encrypt access/refresh tokens at rest (repository/database layer) |
| HIGH | OAuth redirect | `redirectURI` hardcoded to `https://api.rvpay.com/v1/public/oauth/callback` in `clients/oauth/service.go`; config `HIGHLEVEL_REDIRECT_URI` not consumed by service | Wire config into OAuth service |
| MEDIUM | Webhooks | No `webhook_events` table for provider event ID deduplication; replay protection incomplete | Add `webhook_events` migration + dedup logic |
| MEDIUM | OAuth refresh | No automatic token refresh scheduling; tokens expire without proactive refresh | Add background refresh job |
| MEDIUM | Database | `updated_at` columns NOT NULL DEFAULT NOW() but no trigger to auto-update on UPDATE | Add trigger (follow-up migration) |
| LOW | Clients | `client_name` is not unique; duplicate client names possible | Consider unique constraint (requires product decision) |
| INFORMATIONAL | Tests | No repository integration tests (require PostgreSQL test database) | Add integration test suite |
| INFORMATIONAL | Tests | No gRPC/API translation tests or REST/gateway tests | Add transport-layer tests |

## Blockers

No production blockers identified.

## Remaining Risks

1. OAuth tokens unencrypted at rest — a data breach would expose valid
   credentials.
2. Webhook deduplication relies solely on subscription lookup; provider retries
   could trigger duplicate business operations.
3. No automatic token refresh — integrations will break when tokens expire
   until manual refresh is invoked.
4. Redirect URI hardcoded — deployment-specific URLs cannot be configured
   without code changes.

## Recommended Follow-ups

1. Encrypt OAuth tokens at rest (AES-GCM or similar) in the repository layer.
2. Add a `webhook_events` table to persist provider event IDs for idempotent
   processing.
3. Add a background refresh scheduler for expiring OAuth tokens.
4. Wire `HIGHLEVEL_REDIRECT_URI` from configuration into the OAuth service.
5. Add PostgreSQL-backed integration tests for the repository layer.
6. Add gRPC/API translation tests.
7. Add an `updated_at` trigger for consistency.

## Final Verdict

**READY WITH WARNINGS**

The Clients Service is deployable and architecturally sound. It passes all
automated validation, builds successfully, and follows the established RVPay
conventions. However, the OAuth token encryption, webhook deduplication, and
config-driven redirect URI issues should be addressed before broader production
deployment to ensure security, idempotency, and operational reliability.