# Platform Security Review

Document Version: 1.0
Status: Complete
System: RVPay
Review: Platform Agent 10 — Security

## 1. Objective

Audit and strengthen the security foundation of the RVPay platform. The
objective is to ensure the platform:

- handles secrets safely
- protects authentication credentials
- validates external requests
- protects OAuth credentials
- protects webhook endpoints
- uses appropriate transport security
- avoids leaking sensitive information
- validates configuration safely
- applies appropriate authorization boundaries
- maintains secure service-to-service communication
- follows the existing RVPay architecture

Security work answered, for every boundary: what we protect, from whom, at
which boundary, how protection is implemented, how failure is handled, and
how the behavior is tested. Changes were concrete, justified, minimal, and
compatible with the existing repository. No architecture was redesigned and no
performance work was performed.

## 2. Required Documentation

| Document | Read |
| --- | --- |
| README.md | ✅ |
| agents/project-context.md | ✅ |
| docs/domain-model.md | ✅ |
| docs/repository-layout.md | ✅ |
| docs/protobuf-strategy.md | ✅ |
| docs/migration-plan.md | ✅ |
| docs/platform-repository-audit.md | ✅ |
| docs/platform-protobuf-generation-review.md | ✅ |
| docs/platform-http-gateway-review.md | ✅ |
| docs/platform-common-packages-review.md | ✅ |
| docs/platform-ci-cd-review.md | ✅ |
| docs/platform-docker-review.md | ✅ |
| docs/platform-render-review.md | ✅ |
| docs/platform-documentation-review.md | ✅ |
| docs/platform-observability-review.md | ✅ |

All required documents were present and read.

## 3. Existing Security

The following security mechanisms were already present and are preserved:

- **Secrets**: no committed credentials; every `.env.example` uses placeholders;
  secrets are environment-provided only (DB passwords, PawaPay key, HighLevel
  client credentials, webhook secret, SSO key, token encryption key). `.gitignore`
  excludes `**/.env`. `render.yaml` uses manual `sync: false` secrets and
  Render `fromDatabase` wiring. CI references only the `RENDER_DEPLOY_HOOK`
  secret via `env:` and never prints it.
- **OAuth**: HighLevel OAuth flow implemented server-side (authorization URL
  generation, code exchange, token refresh, user info); `GenerateState` uses
  `crypto/rand`; tokens are stored encrypted at rest by the service. OAuth
  errors are mapped to safe gRPC status errors (no credentials leaked).
- **Webhooks**: HighLevel signature verification implemented via HMAC-SHA256
  with `hmac.Equal` (constant-time comparison); malformed payloads and missing
  signatures are rejected with `InvalidArgument`; payload parsing validates
  required fields before processing.
- **Observability**: gRPC interceptor + HTTP middleware log metadata only —
  never Authorization headers, tokens, API keys, SSO/webhook secrets, request
  bodies, or provider credentials (verified in Agent 09 and re-verified here,
  directive #132).
- **Docker**: multi-stage distroless images, non-root execution, no `.env`/keys
  in images, no `ARG`/`ENV` secrets (Agent 06).
- **Render**: TLS from Render, DB `DB_TLS_DISABLED=false`, secrets manual
  `sync: false` (Agent 07).
- **CI/CD**: `contents: read` permissions; no secrets echoed; `go vet ./...`
  and `go test ./...` gating (Agent 05).
- **SQL**: all queries go through sqlc (parameterized); no string-concatenated
  SQL observed; no dynamic SQL construction.
- **Dependencies**: pinned toolchain (`tools/versions.md`); no blind upgrades.

## 4. Security Boundaries

| Boundary | Exposure | Security Controls |
| --- | --- | --- |
| Public HTTP traffic | HTTPS via Render/OCI nginx (`nginx.conf` TLSv1.2/TLSv1.3, HTTP→HTTPS redirect) | TLS termination at edge; `/healthz` only; no auth middleware (documented gap — see Remaining Risks) |
| HTTP gateway | Embedded grpc-gateway on `:PORT` | Request-ID correlation; metadata-only access logs; grpc-gateway error mapping (no internal details leaked) |
| gRPC services | Public gRPC on `:LISTEN_PORT` (reflection enabled) | Recovery + observability interceptors; no auth middleware (documented gap) |
| Service-to-service | Not implemented (no cross-service calls wired yet) | N/A — services are isolated; no shared DB tables (domain-model) |
| Database access | Render-managed PostgreSQL, internal URLs via `fromDatabase` | `DB_TLS_DISABLED=false` (sslmode=require); credentials environment-only; DSN credentials never logged (SEC-01 🔧) |
| External provider APIs | HighLevel OAuth/token/userinfo APIs (HTTPS, 10s timeouts) | Fixed provider endpoints; no user-controlled URLs accepted; provider credentials never logged |
| OAuth callbacks | HighLevel redirect to configured callback | Configured redirect URI (SEC-02 🔧); server-side code exchange; state via `GenerateState` (`crypto/rand`); errors mapped to safe gRPC codes |
| Webhook endpoints | HighLevel POST webhooks | HMAC-SHA256 signature verification with `hmac.Equal`; distinct webhook secret (SEC-03 🔧); malformed payload rejection; no replay protection (no provider nonce — documented) |
| Render infrastructure | Public web services | TLS; `/healthz` health checks; manual secrets; DB not publicly exposed |
| Developer/local environments | `.env` files | `.gitignore` excludes `**/.env`; `.env.example` placeholders only |

## 5. Secrets

- **Sources**: environment variables only (`.env` locally, Render dashboard
  `sync: false` secrets, CI secrets). No secrets exist in Go source, YAML,
  Markdown, Dockerfiles, Makefiles, protobufs, migrations, or tests.
- **Handling**: secrets are loaded into service config structs
  (`clients/config`, `transactions/config`, `deposits/config`,
  `integrations/config`) and passed to the provider/repository layer; never
  logged, never returned through API responses, never committed.
- **Exposure prevention verified**:
  - Targeted secret-pattern scan over `.md`, `.yaml`, `.yml`,
    `.env.example`, `Makefile`, `Dockerfile` files returned no matches.
  - The observability layer (Agent 09) never logs Authorization headers,
    tokens, request bodies, or credentials (re-verified; `TestAccessLog_
    NoPayloadLogging`).
  - **SEC-01 / CRITICAL (fixed)**: `shared/database.Migrate` previously logged
    the entire `*migrate.Migrate` struct (`logger.Info().Msgf("migration
    instance %v", m)`), which contains the full database URL including
    credentials. Removed; only the migration path is logged. Regression test
    `TestMigrate_DoesNotLogDatabaseCredentials` verifies the password and DSN
    never appear in logs.
- **OAuth/webhook secrets**: HighLevel client ID/secret, redirect URI, and
  webhook secret are environment-provided. The webhook secret was previously
  incorrectly reused from the OAuth client secret (SEC-03, fixed below).

## 6. Authentication

- No authentication/authorization middleware exists in any service. This
  matches the legacy Deposits convention and was documented as open finding
  P-07 (repository audit) and F-03 (Transactions production review). Per the
  agent's scope restriction — "Do NOT introduce unrelated security products"
  and "Do NOT replace working authentication mechanisms" — no auth middleware
  was invented; the existing architecture has no authentication mechanism to
  preserve, and adding one requires an architectural decision (token scheme,
  service identity, key distribution) outside this agent's "concrete, minimal"
  mandate. This is recorded in Remaining Risks.
- Webhook signature verification is the one form of request authentication
  actually implemented, and it is preserved and hardened (SEC-03).

## 7. Authorization

- No authorization boundaries are implemented. The domain model defines
  ownership (Clients owns clients/platforms/integrations; Transactions owns
  merchants/customers/deposits/payouts) and each service validates that
  referenced IDs exist and are active before operating (e.g. OAuth callback
  checks platform enabled + client active; webhook processing resolves the
  integration before dispatch). Merchant/customer object-level authorization
  against an authenticated caller is not applicable while no authentication
  exists (see Remaining Risks — F-03 and P-07).
- Repositories expose scoped methods (e.g. `ListByClient`, `GetByIntegrationID`),
  and the service layer controls which queries run. No broad unrestricted
  access to sensitive records was observed.

## 8. OAuth Security

- **State**: `providers.GenerateState()` generates a 32-byte cryptographically
  random state via `crypto/rand` (hex-encoded). The OAuth service accepts a
  caller-supplied `state` and forwards it to the provider. The callback handler
  is not yet wired into a public route (the OAuth service is constructed but
  not registered on the gRPC/gateway), so CSRF state validation at the callback
  is not yet exercised by an HTTP handler — this is recorded in Remaining Risks.
- **Redirect handling (SEC-02 / HIGH, fixed)**: The OAuth service previously
  hard-coded `https://api.rvpay.com/v1/public/oauth/callback` in both
  `AuthorizationURL` and `ProcessCallback`, ignoring the configured
  `HIGHLEVEL_REDIRECT_URI`. The `Service` now carries a `redirectURI` field
  injected via `NewService(..., redirectURI, ...)` and uses it for both
  authorization-URL generation and token exchange. `main.go` passes
  `cfg.HighLevel.RedirectURI`. Regression test `TestAuthorizationURL` asserts
  the generated URL carries the configured `redirect_uri`. No arbitrary
  user-controlled redirect destinations are accepted.
- **Authorization code handling**: codes are exchanged server-side in
  `HighLevelProvider.ExchangeCode`; never logged (error paths in the OAuth
  service log client/platform IDs, not the code); not persisted.
- **Token storage**: OAuth tokens are stored via `oauthRepo.Create` (AES-256
  encrypted at rest per the Clients production review); access/refresh tokens
  are not returned through any ordinary service response (the `CallbackResult`
  is returned to the internal caller only); tokens are never logged.
- **Refresh tokens**: handled server-side (`HighLevelProvider.RefreshToken`);
  never logged, never in HTTP responses, never in error messages, never in
  documentation.
- **SSO key** (legacy Integrations): treated as a secret — loaded from
  `HIGHLEVEL_SSO_KEY` env, never committed/logged/returned.

## 9. Webhook Security

- **Signature verification**: HighLevel webhook requests are HMAC-SHA256 signed
  (`timestamp + body`) using a dedicated webhook secret. Verification runs in
  `HighLevelWebhookProvider.VerifyRequest` and rejects missing signature or
  timestamp headers. Comparison uses `hmac.Equal` (constant-time, directive
  #26/#97).
- **Replay protection**: HighLevel provides a timestamp header but no nonce;
  the timestamp is included in the signature but not currently bound to a
  replay window. Per directive #25 ("Do not implement speculative replay
  systems without provider support") no speculative replay cache was added.
  Documented in Remaining Risks.
- **Payload validation**: `ParseEvent` requires valid JSON; missing
  `integrationId` fails UUID parsing and is rejected as `InvalidPayload`;
  unsupported providers/event types are handled (unknown events log and are
  skipped). Appropriate gRPC status (`InvalidArgument`) maps to HTTP 400 via
  grpc-gateway.
- **Idempotency**: duplicate-event detection is stubbed (webhook_events table
  does not exist); the service logs the provider event ID and continues. No
  duplicate effects are created because the dispatchers are stubs. Documented
  in Remaining Risks (matches existing Clients production-review findings).
- **Webhook secret (SEC-03 / HIGH, fixed)**: `HighLevelProvider.WebhookProvider()`
  previously returned a webhook verifier using the OAuth **client secret**
  (`p.clientSecret`), ignoring the distinct `WEBHOOK_SECRET` configuration
  value. The provider now holds a dedicated `webhookSecret`, set from
  `cfg.Webhook.Secret` in `main.go`. Regression test
  `TestHighLevelWebhookSecretIsUsedForSignatureVerification` verifies that a
  signature computed with the webhook secret verifies and a signature computed
  with the client secret is rejected.

## 10. Input Validation

Security-sensitive validation present at service boundaries:

- **IDs**: UUIDs parsed/validated before repository/provider operations
  (`parseUUID`, `uuid.Parse` in OAuth/webhook/Integrations services); invalid
  IDs are rejected with gRPC `InvalidArgument`.
- **Webhook payloads**: JSON parse, required `integrationId`, supported
  provider, existing integration — validated before processing.
- **Pagination**: `ListIntegrations` clamps page size and derives offset from
  the page token (no negative/unreasonable values accepted as-is).
- **Financial amounts**: Transactions uses `commongrpc.Money` decimal strings
  (NUMERIC(18,2)); no floating point; amount validation is domain-owned.
- **Configuration**: Transactions and Deposits use `ardanlabs/conf` with
  `required` tags for DB secrets/ports — missing values fail startup loudly and
  clearly, and never fall back to embedded credentials.

No oversized-input limits were introduced without justification (directive #46);
gRPC default message limits apply.

## 11. Database Security

- **Credentials**: environment-provided (`DB_*` required tags in
  Transactions/Deposits; env helpers in Clients/Integrations); never committed;
  `render.yaml` uses `fromDatabase` (internal URLs).
- **TLS**: Render sets `DB_TLS_DISABLED=false` → `shared/database.PostgresURL`
  emits `sslmode=require`; local dev uses `sslmode=disable` explicitly
  (`DB_TLS_DISABLED=true`) — production is not made insecure for local
  convenience (directive #75).
- **SQL parameterization**: all queries are sqlc-generated (parameterized); no
  dynamic SQL construction, no user-controlled table/order/filter expressions
  (directives #43–#44). Generated sqlc code was not modified.
- **Access boundaries**: one database per service (no shared tables);
  `ON DELETE RESTRICT` for financial history; repositories expose scoped
  methods. Cross-service access is gRPC-only by architecture.
- **Logging (SEC-01 / CRITICAL, fixed)**: migration instance (which holds the
  full DSN with credentials) is no longer logged.

## 12. HTTP/gRPC Security

- **Transport**: HTTPS/TLS terminated by Render and by OCI `nginx/nginx.conf`
  (TLSv1.2/TLSv1.3, HTTP→HTTPS 301). No custom TLS inside application services
  (directive #35).
- **Authentication**: none implemented (see §6 and Remaining Risks).
- **Authorization**: none implemented; ownership enforced by service logic and
  scoped repositories (see §7).
- **Error handling**: services return gRPC status errors
  (`status.Error(codes.*, "safe message")`); the gateway maps them to HTTP via
  grpc-gateway defaults (verified 404/501 in Agent 03 tests). Internal details,
  SQL, stack traces, credentials, and filesystem paths are never placed in
  public error strings; detailed technical errors are logged server-side only
  (directives #49–#52).
- **Headers**: no CORS policy exists; `Access-Control-Allow-Origin: *` is not
  used (HGW-05). No browser-facing traffic exists today (gRPC/REST API clients
  only), so security headers beyond what the gateway already emits were not
  added (directive #34).
- **gRPC metadata**: request-ID only; no auth metadata is passed; secrets are
  never placed in gRPC metadata.

## 13. Container Security

- **User**: all runtime images run as `nonroot:nonroot` (distroless static)
  — verified in Agent 06 image inspection.
- **Secrets**: no `.env`, keys, or credentials in images; no `ARG`/`ENV`
  secrets; `.dockerignore` excludes `**/.env`; verified by filesystem scan
  (Agent 06).
- **Ports**: `EXPOSE 50051` (documentation only); gateway HTTP on Render-
  injected `PORT`; PostgreSQL never exposed in application images.
- **Debug endpoints**: no pprof/debug/admin HTTP endpoints registered in any
  service; only the documented gRPC services, gateway, and `/healthz`.
- **Dev vs prod**: `DB_TLS_DISABLED` defaults differ per service deliberately;
  production Render sets `false`; no production-insecure defaults.

## 14. CI/CD Security

- **Secrets**: only `RENDER_DEPLOY_HOOK` is referenced, via `env:` from GitHub
  secrets; never echoed or printed; absent secret produces an explicit
  `::notice::` skip (preserved from Agent 05).
- **Permissions**: `contents: read` at the workflow level (no `write-all`).
- **Build logs**: no tokens, passwords, private keys, or database URLs printed
  by any build command; `curl -fsS -X POST` to the deploy hook shows no
  credential in output.
- **Toolchain pinning**: Go 1.26.5 and pinned protoc/plugins/sqlc versions;
  no `@latest` (preserved).
- No CI workflow changes were made by this agent (directives #77–#79; CI/CD is
  Agent 05 scope and was already hardened).

## 15. Render Security

- **Environment variables**: non-sensitive values set in the Blueprint;
  secrets are manual `sync: false` (`HIGHLEVEL_CLIENT_ID`,
  `HIGHLEVEL_CLIENT_SECRET`, `HIGHLEVEL_REDIRECT_URI`, `WEBHOOK_SECRET` for
  clients; `PAWAPAY_API_URL`/`PAWAPAY_API_KEY` for deposits). No values
  committed.
- **Service exposure**: three public web services with `/healthz` health
  checks; no internal services exposed on the public gateway without
  justification.
- **TLS**: provided by Render at the edge; services bind plaintext `:PORT` /
  `:LISTEN_PORT` inside Render's private network (standard Render model).
- **Database exposure**: managed PostgreSQL wired via `fromDatabase` internal
  URLs; not publicly exposed.
- `DB_TLS_DISABLED=false` ensures `sslmode=require` to the managed databases
  (directive #42).

## 16. Tests

Security tests added for the behaviors changed by this agent (directives
#81–#90; no external infrastructure required):

| Test | File | Verifies |
| --- | --- | --- |
| `TestMigrate_DoesNotLogDatabaseCredentials` | `shared/database/database_test.go` | SEC-01: migration logs never contain the DB password or DSN |
| `TestHighLevelWebhookSecretIsUsedForSignatureVerification` | `clients/providers/registry_test.go` | SEC-03: webhook signatures verify with the dedicated `WEBHOOK_SECRET`; signatures made with the OAuth client secret are rejected |
| `TestAuthorizationURL` (extended assertion) | `clients/oauth/service_test.go` | SEC-02: generated OAuth authorization URL carries the **configured** `redirect_uri`, not a hard-coded fallback |

Existing security-relevant tests retained and passing: webhook invalid-signature
rejection (`TestProcessWebhookInvalidSignature`), unknown-provider rejection,
webhook/identity not-found handling, and the Agent 09 `TestAccessLog_NoPayload
Logging` (Authorization header/body never logged).

Commands executed:

```bash
go build ./shared/... ./clients/...          # OK
go test ./shared/database/... ./clients/oauth/... ./clients/webhooks/... ./clients/providers/...  # OK
go test ./...                                # OK (full suite)
go vet ./...                                 # OK
gofmt -l (hand-written Go)                   # flags are the documented pre-existing local
                                             # core.autocrlf=true CRLF artifact (CICD-07/P-GEN-04);
                                             # CI is LF-only and passes. No real drift introduced.
```

## 17. Findings

| ID | Severity | Area | Finding | Resolution |
| --- | --- | --- | --- | --- |
| SEC-01 | CRITICAL | `shared/database/database.go` | `Migrate` logged the full `*migrate.Migrate` struct, which contains the database URL including credentials (`logger.Info().Msgf("migration instance %v", m)`) | ✅ Removed the credential-leaking log; only the migration path is logged; added `TestMigrate_DoesNotLogDatabaseCredentials` |
| SEC-02 | HIGH | `clients/oauth/service.go` | OAuth redirect URI hard-coded to `https://api.rvpay.com/v1/public/oauth/callback` in `AuthorizationURL` and `ProcessCallback`, ignoring configured `HIGHLEVEL_REDIRECT_URI` | ✅ `Service` now receives a configured `redirectURI`; `main.go` passes `cfg.HighLevel.RedirectURI`; regression assertion added to `TestAuthorizationURL` |
| SEC-03 | HIGH | `clients/providers/highlevel.go` | Webhook signature verification reused the OAuth client secret (`p.clientSecret`) instead of the distinct `WEBHOOK_SECRET`; `cfg.Webhook.Secret` was loaded but unused | ✅ `HighLevelProvider` carries a dedicated `webhookSecret`; `main.go` passes `cfg.Webhook.Secret`; regression test added |
| SEC-04 | INFO | `.gitignore`, `.env.example`, `render.yaml`, CI | Secret handling already correct: `**/.env` ignored; placeholders only; `sync:false` secrets; `contents: read`; no secret printing | ✅ Preserved and re-verified this agent |
| SEC-05 | INFO | `shared/observability` | Logging never emits Authorization headers, tokens, API keys, SSO/webhook secrets, payloads, or financial data | ✅ Re-verified (Agent 09 `TestAccessLog_NoPayloadLogging`); no changes needed |
| SEC-06 | INFO | Webhooks | Signature verification already uses `hmac.Equal` (constant-time) | ✅ Preserved (directives #26/#97) |
| SEC-07 | INFO | `clients/oauth` | `GenerateState` uses `crypto/rand` (32-byte hex) — secure randomness | ✅ Preserved (directive #96) |
| SEC-08 | INFO | Local gofmt | `core.autocrlf=true` causes local `gofmt -l` to flag all CRLF working-tree files; CI is LF-only and passes (pre-existing CICD-07/P-GEN-04) | ✅ Documented; not caused by this agent; no `.gitattributes` change made (outside minimal-change scope) |

## 18. Remaining Risks

Documented, consciously NOT solved by this agent (outside "concrete, minimal,
justified" scope or requiring an architectural decision):

- **No authentication/authorization middleware** (audit P-07, Transactions
  F-03): all public HTTP/gRPC endpoints are anonymous. MLS/deposits-style
  authentication does not exist yet. Introducing it requires an architectural
  decision (token scheme, key distribution, service identity) and is tracked
  for a future agent/decision.
- **OAuth callback not wired to a public route**: `clients/oauth.Service` is
  constructed in `main.go` but its result is discarded — the OAuth callback
  handler, webhook processing entry, and OAuth service are not yet registered
  as gRPC/gateway handlers. State validation (CSRF) and callback handling are
  therefore not yet exercised end-to-end over HTTP. The SEC-02/03 fixes make
  the latent code paths correct; wiring them is functional integration work.
- **Webhook replay protection**: HighLevel provides a timestamp but no nonce;
  no replay window is enforced (directive #25 — no speculative replay system).
  Consider a timestamp-window check if HighLevel documentation supports it.
- **Webhook idempotency**: duplicate-event detection is stubbed (no
  `webhook_events` table); the Clients production review already documents this.
- **No rate limiting**: the architecture does not require a rate limiter
  (directive #64); OAuth/webhook endpoints have no quotas. If abused, add at the
  edge (Render/nginx), not in-service.
- **gRPC reflection enabled** in all services: useful in dev, but exposes the
  service API surface publicly. Consider disabling reflection in production
  (Render env flag) as an operator decision.
- **No vulnerability scanning** of dependencies/images (Agent 06 deferred this);
  `govulncheck`/image scanning is not configured.
- **Legacy services** (deposits/integrations) retain pre-existing
  configurations and patterns; this agent did not modify them per
  `.clinerules`.
- **Payment data**: Transactions has no auth and no status-mutation RPCs yet;
  provider execution and reconciliation are future integration work (F-01/F-02).
- **No compliance claims**: this review performs no PCI/SOC 2/GDPR/ISO
  certification or compliance assertion (directive #117).

## 19. Documentation Changes

- `docs/platform-security-review.md` — created (this document).
- `docs/project-checkpoint.md` — updated: Platform 10 Security marked COMPLETE.

No other documentation was modified (directive #112: only security-visible
developer behavior is documented; the README's existing security-relevant
configuration table was already accurate).

## 20. Documentation Check

Final verification — all required documents exist:

| Document | Present |
| --- | --- |
| README.md | ✅ |
| agents/project-context.md | ✅ |
| docs/domain-model.md | ✅ |
| docs/repository-layout.md | ✅ |
| docs/protobuf-strategy.md | ✅ |
| docs/migration-plan.md | ✅ |
| docs/platform-repository-audit.md | ✅ |
| docs/platform-protobuf-generation-review.md | ✅ |
| docs/platform-http-gateway-review.md | ✅ |
| docs/platform-common-packages-review.md | ✅ |
| docs/platform-ci-cd-review.md | ✅ |
| docs/platform-docker-review.md | ✅ |
| docs/platform-render-review.md | ✅ |
| docs/platform-documentation-review.md | ✅ |
| docs/platform-observability-review.md | ✅ |
| docs/platform-security-review.md | ✅ (this document) |

## 21. Final Status

PASS WITH FOLLOW-UP