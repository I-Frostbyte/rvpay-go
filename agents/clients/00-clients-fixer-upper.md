# Agent 00 — Clients Fixer Upper: Complete GoHighLevel Integration

## Objective

Complete and production-wire the GoHighLevel Marketplace integration in the **Clients service**.

The Clients service currently contains OAuth and webhook domain logic, but the runtime does not expose the required HTTP endpoints. The legacy Integrations service contains older OAuth/webhook handlers, but that service is no longer part of the new Clients deployment.

This task must close that gap.

The final result must allow the deployed `rvpay-clients` service to:

1. Start normally.
2. Expose the Clients gRPC/HTTP gateway.
3. Handle the GoHighLevel OAuth authorization callback.
4. Handle GoHighLevel webhook deliveries.
5. Verify GoHighLevel webhook signatures using the CURRENT GHL Ed25519 `X-GHL-Signature` mechanism.
6. Preserve webhook request-body integrity during signature verification.
7. Return appropriate HTTP responses to GHL.
8. Reuse the existing Clients domain services/provider abstractions wherever appropriate.
9. Be properly wired into `clients/cmd/grpc-service`.
10. Have comprehensive tests.
11. Have the required environment configuration documented.
12. Be deployable through the existing Render Clients service.

This is an implementation task.

You ARE expected to modify the repository.

---

# 1. NON-NEGOTIABLE PROJECT RULES

Before doing anything, read and obey:

* `.clinerules`
* `AGENTS.md`
* `README.md`
* `agents/project-context.md`
* `docs/domain-model.md`
* `docs/repository-layout.md`
* `docs/protobuf-strategy.md`
* `docs/migration-plan.md`
* `docs/project-checkpoint.md`

The above files are authoritative project guidance.

If `.clinerules` or `AGENTS.md` impose additional restrictions, follow them.

Do not override project conventions merely because another implementation would be cleaner.

---

# 2. REPOSITORY EXPLORATION RULE

Use:

* `README.md`
* `agents/project-context.md`
* `docs/repository-layout.md`
* `docs/project-checkpoint.md`

as the repository map.

Do NOT waste time recursively exploring the entire repository.

Do NOT inspect unrelated areas.

Do NOT recursively inspect:

* `.git/`
* `vendor/`
* `third_party/`
* `node_modules/`
* generated dependencies
* unrelated services
* unrelated tests
* unrelated documentation

Only inspect files necessary to complete the Clients GHL integration.

---

# 3. REQUIRED INITIAL REVIEW

Before writing code, inspect the existing Clients implementation.

At minimum inspect:

```text
clients/
├── cmd/
│   └── grpc-service/
├── config/
├── oauth/
├── providers/
├── webhooks/
└── ...
```

Also inspect the relevant:

```text
protobuf/clients.proto
```

and generated protobuf/gateway code if required to understand the current transport.

Inspect the current:

```text
render.yaml
```

only as needed to understand the deployed Clients environment.

---

# 4. LEGACY INTEGRATIONS SERVICE — REFERENCE ONLY

The old `integrations/` service is NOT the target architecture.

You MAY inspect the legacy implementation only to understand:

* how OAuth callback handling previously worked
* how webhook handling previously worked
* how request/response behavior worked
* what GHL data was previously expected
* what migration behavior needs to be preserved

Relevant legacy files include the existing OAuth and webhook handlers.

Do NOT copy the legacy architecture blindly.

Do NOT reintroduce the Integrations service.

Do NOT make the new Clients service depend on the Integrations service.

Do NOT restore legacy package boundaries merely because they already work.

The Clients service is now the owner of the GHL integration.

---

# 5. CURRENT GHL WEBHOOK SIGNATURE STANDARD

The current GoHighLevel webhook mechanism documented for this project uses:

```text
X-GHL-Signature
```

with:

```text
Ed25519
```

verification.

Treat this as authoritative for the new implementation.

Do NOT design the new implementation around the older:

```text
X-HighLevel-Signature
X-HighLevel-Timestamp
HMAC-SHA256
```

mechanism merely because an existing provider implementation currently uses it.

If the existing provider implementation uses HMAC-SHA256, update it to support the current GHL Ed25519 mechanism.

Do not remove useful provider abstraction unnecessarily.

---

# 6. GHL WEBHOOK VERIFICATION

Implement proper verification of:

```text
X-GHL-Signature
```

using the GHL Ed25519 public key.

The implementation MUST:

1. Read the raw HTTP request body.
2. Preserve the exact raw bytes.
3. Read `X-GHL-Signature`.
4. Decode the signature according to the GHL documented encoding.
5. Load the configured GHL Ed25519 public key.
6. Verify the signature against the ORIGINAL request body.
7. Reject invalid signatures.
8. Reject missing signatures.
9. Reject malformed signatures.
10. Never verify a re-marshaled JSON body.
11. Never verify a transformed/normalized JSON representation.

The signature verification must occur BEFORE processing the webhook payload.

Do not parse and re-serialize JSON before verification.

---

# 7. WEBHOOK PUBLIC KEY CONFIGURATION

Inspect the existing configuration model.

Determine whether the current:

```text
WEBHOOK_SECRET
```

variable is appropriate.

Because the GHL Ed25519 value is a PUBLIC verification key and not a shared secret, do not blindly preserve misleading terminology.

Prefer a clearly named configuration variable such as:

```text
HIGHLEVEL_WEBHOOK_PUBLIC_KEY
```

if this is consistent with the project's existing naming conventions.

However:

* do not introduce redundant variables
* do not break existing configuration unnecessarily
* do not silently support multiple conflicting variables

If migration compatibility is required, document it clearly.

The final configuration must make it obvious that this value is a public verification key, not a private credential.

---

# 8. DO NOT STORE SECRETS IN SOURCE

Never hard-code:

* HighLevel Client ID
* HighLevel Client Secret
* webhook credentials
* tokens
* private keys

The GHL Ed25519 public key may be public cryptographic material, but still keep it configurable rather than hard-coded.

Use the existing project configuration system.

---

# 9. OAUTH CALLBACK IMPLEMENTATION

The Clients service already contains OAuth domain logic.

Inspect:

```text
clients/oauth/
```

and the HighLevel provider implementation.

Determine how:

```text
AuthorizationURL
ProcessCallback
RefreshAccessToken
ValidateToken
```

are intended to work.

The missing piece is the HTTP transport.

Implement the OAuth callback so that the Clients service can receive the callback from HighLevel.

The callback must correctly handle:

```text
code
state
```

and any other parameters required by the existing OAuth flow.

Do NOT invent an unrelated OAuth protocol.

Use the existing domain service/provider abstractions.

---

# 10. OAUTH STATE

Inspect how OAuth state is currently generated and validated.

Do not weaken state validation.

The callback MUST NOT simply accept:

```text
code
```

and exchange it without validating state if the architecture expects state protection.

If the current OAuth service does not provide sufficient state validation, implement the missing protection according to project conventions.

Document the behavior.

---

# 11. OAUTH CLIENT IDENTIFICATION

Cline's previous diagnostic found that the existing:

```text
ProcessCallback(ctx, clientID, platformID, code, state)
```

requires identifiers that are not directly supplied by GHL's callback.

Resolve this properly.

Do NOT fabricate UUIDs.

Determine from the existing Clients domain model:

* how the OAuth flow identifies the RVPay client
* how the HighLevel platform is identified
* how the callback can securely recover the required context
* whether the existing `state` value is intended to carry that context
* whether the existing authorization URL generation already has the required information

Preserve the domain model.

If state must encode the necessary callback context, use a secure and project-consistent mechanism.

Do not place sensitive credentials into state.

---

# 12. OAUTH REDIRECT URI

The environment variable:

```text
HIGHLEVEL_REDIRECT_URI
```

must represent the actual publicly reachable callback URL of the deployed Clients service.

Do not leave an obsolete default such as:

```text
https://api.rvpay.com/v1/public/oauth/callback
```

unless that domain and route actually exist in the deployed architecture.

The application must use the configured value consistently when:

1. generating the HighLevel authorization URL
2. exchanging the authorization code

OAuth redirect URI mismatches must be prevented.

---

# 13. WEBHOOK HTTP ENDPOINT

Implement a public HTTP webhook endpoint in the Clients service.

The endpoint must:

1. Accept the HTTP method expected by GHL.
2. Read the raw request body.
3. Verify `X-GHL-Signature`.
4. Reject invalid signatures.
5. Parse the payload only after verification.
6. Determine the appropriate provider/platform.
7. Pass the validated webhook to the existing Clients webhook domain service.
8. Return an appropriate successful response when accepted.
9. Return an appropriate failure response when validation fails.
10. Avoid exposing internal errors or secrets.

Use the existing webhook domain service:

```text
clients/webhooks/
```

where appropriate.

---

# 14. WEBHOOK IDEMPOTENCY

Inspect the existing webhook domain model and database/repository capabilities.

Determine whether webhook events have an event ID or another unique identifier.

If the current design supports idempotency, enforce it.

If it does not yet support idempotency, implement the minimum required mechanism consistent with the project architecture.

The webhook endpoint must be safe against duplicate deliveries.

Remember that external webhook providers can retry requests.

Do not assume a webhook is delivered exactly once.

---

# 15. WEBHOOK RESPONSE BEHAVIOR

The webhook endpoint should acknowledge a valid event quickly.

Do not perform unnecessary long-running processing directly inside the HTTP request.

If the existing architecture already provides an asynchronous mechanism, use it.

If no queue exists yet, keep the implementation compatible with future asynchronous processing.

Do NOT introduce an entirely new queue infrastructure as part of this task unless the existing project architecture explicitly requires it.

This task is about completing GHL integration, not redesigning the platform.

---

# 16. RAW BODY REQUIREMENT

Be extremely careful with Go HTTP request handling.

The verification sequence should conceptually be:

```text
HTTP request
    │
    ├── read raw body bytes
    │
    ├── verify X-GHL-Signature
    │
    ├── parse JSON
    │
    └── process validated event
```

NOT:

```text
HTTP request
    │
    ├── parse JSON
    ├── marshal JSON
    └── verify marshaled JSON
```

The latter can invalidate signatures.

---

# 17. ROUTE REGISTRATION

Wire the new HTTP handlers into:

```text
clients/cmd/grpc-service/main.go
```

or the appropriate runtime/router package according to project conventions.

Do not leave the services instantiated and discarded.

The final runtime must have an actual path:

```text
HTTP
 ↓
OAuth/Webhook handler
 ↓
Clients domain service
 ↓
Provider/repository as appropriate
```

Confirm that the handlers are actually reachable from the deployed HTTP server.

---

# 18. DO NOT BREAK THE EXISTING GRPC GATEWAY

The Clients service already exposes grpc-gateway routes.

Preserve all existing:

```text
/v1/public/...
```

routes.

The new GHL routes must coexist with:

```text
httpMux.Handle("/", gatewayMux)
```

and:

```text
/healthz
```

without accidentally shadowing existing routes.

Pay attention to handler registration order and path matching.

---

# 19. PROTOBUF DECISION

Do not automatically create OAuth/webhook protobuf services.

First determine whether direct HTTP handlers are more appropriate for:

* browser OAuth callbacks
* external provider webhooks

The existing protobuf strategy documentation is authoritative.

Do not create unnecessary RPCs simply to make the routes fit grpc-gateway.

Use direct HTTP handlers if that is consistent with the project's documented transport strategy.

If a protobuf/gateway route is genuinely required by the existing architecture, implement it correctly and regenerate protobuf code using the project's established process.

Do not hand-edit generated protobuf files.

---

# 20. HIGHLEVEL PROVIDER ABSTRACTION

Preserve the provider abstraction.

HighLevel-specific behavior should remain in the HighLevel provider implementation rather than being scattered throughout generic HTTP handlers.

The transport layer should deal with:

* HTTP
* headers
* request body
* response status

The provider layer should deal with:

* HighLevel-specific OAuth behavior
* HighLevel webhook signature verification
* HighLevel payload semantics
* HighLevel API interactions

The domain/service layer should deal with:

* business rules
* client/platform/integration relationships
* persistence orchestration

Keep these responsibilities separate.

---

# 21. ERROR HANDLING

Implement clear errors for:

* missing OAuth code
* invalid OAuth state
* OAuth provider failure
* missing webhook signature
* malformed webhook signature
* invalid webhook signature
* malformed JSON
* unsupported webhook event
* provider processing failure

Do not return stack traces.

Do not expose:

* access tokens
* refresh tokens
* client secrets
* database credentials
* internal connection details

in HTTP responses.

Follow existing project error-handling conventions.

---

# 22. TESTS — OAUTH

Add/update tests covering at minimum:

### Authorization

* authorization URL generation
* correct redirect URI
* correct state behavior

### Callback

* valid callback
* missing code
* missing state
* invalid state
* provider exchange failure
* successful callback

Use mocks/fakes where the existing architecture supports them.

Do not call real HighLevel APIs from unit tests.

---

# 23. TESTS — WEBHOOK SIGNATURE

Add comprehensive tests for Ed25519 verification.

At minimum:

1. valid signature
2. invalid signature
3. modified body
4. missing signature
5. malformed Base64/signature
6. wrong public key
7. empty body
8. valid JSON with formatting preserved
9. body altered after signature generation
10. replay/duplicate event handling if supported by the architecture

The tests must demonstrate that verification operates against the raw body.

---

# 24. TESTS — HTTP HANDLERS

Test the actual HTTP endpoints rather than only testing domain methods.

At minimum:

### OAuth endpoint

* correct method
* valid request
* missing parameters
* invalid state
* provider failure

### Webhook endpoint

* correct method
* valid signed request
* invalid signature
* missing signature
* malformed JSON
* valid JSON
* duplicate event if supported

### Health

Ensure:

```text
GET /healthz
```

continues to return the expected result.

---

# 25. TESTS — RUNTIME WIRING

Add a gateway/runtime test proving that:

```text
Clients runtime
```

actually exposes:

* `/healthz`
* OAuth callback route
* HighLevel webhook route
* existing grpc-gateway routes

Do not merely instantiate the handler in isolation.

The test should prove the route is registered.

---

# 26. CONFIGURATION

Update the Clients configuration model as necessary.

Ensure the required GHL values are clearly represented.

At minimum determine the correct configuration for:

```text
HIGHLEVEL_CLIENT_ID
HIGHLEVEL_CLIENT_SECRET
HIGHLEVEL_REDIRECT_URI
HIGHLEVEL_WEBHOOK_PUBLIC_KEY
```

or the project's chosen equivalent names.

Do not leave obsolete names undocumented.

Do not expose actual credentials.

---

# 27. RENDER CONFIGURATION

Inspect the existing Clients entry in:

```text
render.yaml
```

Ensure the required GHL configuration variables are represented correctly.

Secret/private values must remain:

```yaml
sync: false
```

The public Ed25519 verification key may be configured according to the project's preferred Render configuration strategy.

Do not modify unrelated Render services.

Do not modify Transactions infrastructure.

Do not modify databases unrelated to Clients.

---

# 28. LOCAL ENVIRONMENT

If the repository has a Clients `.env.example` or equivalent configuration documentation, update it.

Do NOT create a real `.env` containing secrets.

Use placeholders such as:

```text
HIGHLEVEL_CLIENT_ID=
HIGHLEVEL_CLIENT_SECRET=
HIGHLEVEL_REDIRECT_URI=
HIGHLEVEL_WEBHOOK_PUBLIC_KEY=
```

if consistent with project conventions.

---

# 29. DOCUMENTATION

Update the appropriate documentation so a future developer knows:

1. Which GHL environment variables are required.
2. Which value is the Client ID.
3. Which value is the Client Secret.
4. Which value is the OAuth Redirect URI.
5. Which value is the GHL Ed25519 public key.
6. Which endpoint receives OAuth callbacks.
7. Which endpoint receives GHL webhooks.
8. How webhook signatures are verified.
9. That the webhook public key is NOT a private secret.
10. How to configure the deployed Render service in the GHL Marketplace.

Do not place real credentials in documentation.

---

# 30. GHL CONFIGURATION COMPATIBILITY

The final implementation must make it possible to configure the GHL Marketplace app with:

```text
Client ID
Client Secret
Redirect URL
Webhook URL
```

and configure the Clients service with the corresponding environment variables.

Do not hard-code the Render hostname.

The hostname must be supplied through deployment configuration.

---

# 31. CURRENT DEPLOYMENT TARGET

The current deployment target is:

```text
Render
    │
    └── rvpay-clients
```

The Clients service must work independently.

Do not require:

```text
rvpay-deposits
```

or:

```text
integrations
```

to be running.

Do not reintroduce the legacy Integrations service as a runtime dependency.

---

# 32. NO LEGACY SERVICE DELETION

Do NOT delete:

```text
integrations/
```

Do NOT delete:

```text
deposits/
```

Do NOT rename them.

Do NOT refactor them.

Do NOT remove their Render configuration unless explicitly required by another task.

This task only makes Clients self-sufficient for GHL integration.

---

# 33. SECURITY REVIEW

Before completing the task, explicitly inspect for:

* OAuth CSRF/state vulnerabilities
* webhook signature bypass
* signature verification against transformed JSON
* missing signature handling
* weak secret handling
* sensitive values in logs
* access token leakage
* refresh token leakage
* insecure error messages
* hard-coded credentials
* accepting unsigned webhook requests

Fix issues found within the scope of the Clients GHL integration.

---

# 34. LOGGING

Ensure webhook and OAuth logs do NOT contain:

* Client Secret
* OAuth access token
* OAuth refresh token
* authorization code
* webhook public key unnecessarily
* raw sensitive webhook payload fields

Log useful operational information such as:

* event type
* success/failure
* validation failure category
* request ID/correlation ID if available

Follow the project's existing logging conventions.

---

# 35. RUN TESTS

Run the narrowest relevant tests first.

For example:

```bash
go test ./clients/...
```

Then run the repository's broader test command according to:

* README.md
* AGENTS.md
* .clinerules
* project-context.md

Do not invent a new test workflow if the repository already has one.

Fix failures caused by your changes.

Do not ignore failing tests.

---

# 36. BUILD VALIDATION

Build the Clients service using the project's normal build process.

Confirm:

* compilation succeeds
* protobuf-generated code remains consistent
* Docker build configuration remains valid if practical
* runtime entrypoint remains valid

Do not make unrelated fixes.

---

# 37. RENDER VALIDATION

If the Render CLI is available, validate the Blueprint after your changes:

```bash
render blueprints validate
```

Do not modify unrelated Blueprint resources.

If validation fails because of unrelated infrastructure already present in the repository, report it separately.

---

# 38. GIT SAFETY

Before finishing:

```bash
git status --short
```

Review every changed file.

Do not revert legitimate user changes.

Do not commit.

Do not push.

---

# 39. REQUIRED FINAL REVIEW

At the end of the implementation, perform a focused review of all changed files.

Check:

### Architecture

* OAuth remains domain/provider driven.
* Webhooks remain domain/provider driven.
* HTTP transport is responsible for HTTP concerns.
* GHL-specific behavior remains in the HighLevel provider.

### OAuth

* callback is reachable
* state is validated
* redirect URI is consistent
* tokens are protected

### Webhooks

* `X-GHL-Signature` is verified
* Ed25519 is used
* raw request body is verified
* invalid signatures are rejected
* duplicate events are handled where supported

### Runtime

* handlers are actually registered
* services are no longer discarded
* gateway routes still work
* `/healthz` still works

### Configuration

* GHL Client ID
* GHL Client Secret
* Redirect URI
* Ed25519 public key

are properly represented.

### Deployment

* Render configuration is consistent
* no credentials are committed
* Clients can operate without legacy Integrations

---

# 40. REQUIRED FINAL REPORT

Your final response MUST contain:

## Implementation Summary

Explain what was implemented.

## OAuth

State:

* exact registered callback route
* HTTP method
* state handling
* redirect URI configuration
* provider integration

## Webhooks

State:

* exact registered webhook route
* HTTP method
* signature header
* cryptographic algorithm
* public key configuration
* raw-body verification
* duplicate/idempotency behavior

## Runtime Wiring

List the files changed to connect OAuth/webhooks to the Clients HTTP server.

## Configuration

List the required environment variable names WITHOUT VALUES.

## Tests

List:

* tests added
* tests modified
* commands executed
* results

## Render

State whether:

```bash
render blueprints validate
```

passed.

## Documentation

List documentation updated.

## Changed Files

Provide the complete list of modified/created files.

## Remaining Issues

List only genuine remaining issues.

## GHL Marketplace Configuration

At the end, provide a concise configuration summary in this form:

```text
GHL Client ID:
    HIGHLEVEL_CLIENT_ID

GHL Client Secret:
    HIGHLEVEL_CLIENT_SECRET

GHL Redirect URL:
    https://<render-client-host>/<exact-oauth-route>

GHL Webhook URL:
    https://<render-client-host>/<exact-webhook-route>

GHL Webhook Verification:
    X-GHL-Signature / Ed25519

GHL Public Key:
    HIGHLEVEL_WEBHOOK_PUBLIC_KEY
```

Use the ACTUAL registered routes discovered/implemented by the code.

Do not invent route names.

Do not provide secret values.

---

# FINAL REQUIREMENT

The implementation is NOT complete merely because the OAuth and webhook domain services compile.

The task is complete only when:

```text
GHL
 │
 ├── OAuth ────────────────┐
 │                         ▼
 │                  Clients HTTP
 │                         │
 │                         ▼
 │                    OAuth service
 │                         │
 │                         ▼
 │                    HighLevel
 │
 └── Webhook ─────────────┐
                          ▼
                   Clients HTTP
                          │
                          ▼
                Ed25519 verification
                          │
                          ▼
                  Webhook service
                          │
                          ▼
                     processing
```

is actually wired and tested.

Do not stop at creating handlers.

Prove that the running Clients server exposes the routes.

After completing the implementation and final review, stop.
