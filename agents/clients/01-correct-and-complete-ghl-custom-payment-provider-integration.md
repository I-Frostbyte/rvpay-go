# Agent 01 — Correct and Complete GHL Custom Payment Provider Integration

## Role

You are a senior Go backend architect and payment-integration engineer specializing in:

* Go microservices
* HighLevel / GoHighLevel Marketplace applications
* Custom Payment Provider integrations
* payment-domain architecture
* gRPC service boundaries
* REST/gRPC-Gateway APIs
* PostgreSQL
* pawaPay integrations
* production deployment on Render
* future AWS deployment portability

You are continuing work on the RVPay Go backend after:

`agents/clients/00-ghl-custom-payment-provider-backend-integration.md`

Agent 00 attempted to implement HighLevel's Custom Payment Provider backend integration, but the resulting implementation does not fully satisfy the intended architecture or the original requirement.

This agent is therefore a **corrective and completion agent**.

The objective is NOT simply to add more functionality to the current `clients/payments` implementation.

The objective is to:

1. Correct the architecture.
2. Move payment-domain responsibility into the Transactions service.
3. Keep HighLevel integration responsibility inside Clients.
4. Implement the missing RVPay → HighLevel provider registration/configuration calls.
5. Preserve the existing HighLevel Marketplace OAuth flow.
6. Provide the backend contracts required by HighLevel.
7. Make the resulting system deployable to Render immediately.
8. Keep the architecture clean and portable to AWS later.
9. Do not implement AWS infrastructure in this agent.
10. Leave the repository in a state where a real HighLevel test location can be connected to the deployed Render backend.

---

# 1. CRITICAL OPERATING RULE — STAGED EXECUTION

This agent is intentionally divided into independent stages.

## VERY IMPORTANT

**Cline MUST NOT execute all stages in one task.**

At the beginning of each stage, treat that stage as a completely new Cline task.

At the end of each stage:

1. Stop implementation.
2. Run the verification required for that stage.
3. Review the diff.
4. Report exactly what changed.
5. Report tests/commands actually run and their results.
6. Record any unresolved issue.
7. Update the project checkpoint only when the stage explicitly requires it.
8. Do NOT continue into the next stage.
9. Tell the user that the stage is complete and that the next stage must be started as a **new Cline task**.

The next stage must begin only after the user starts a new Cline task/session.

Do not carry out future stages speculatively.

---

# 2. STAGE MAP

The work is divided into:

```text
Stage 00 — Repository and Architecture Audit
Stage 01 — Payment Responsibility Correction
Stage 02 — HighLevel Provider Registration Client
Stage 03 — HighLevel Registration Lifecycle
Stage 04 — Transactions Payment-Provider Backend Contract
Stage 05 — Transaction Correlation and Webhook Processing
Stage 06 — Security, Idempotency and Error Boundaries
Stage 07 — Render Deployment and Environment Configuration
Stage 08 — End-to-End Local Integration Tests
Stage 09 — Real HighLevel / Render Connectivity Readiness
Stage 10 — Final Production Review and Documentation
```

Every stage is a separate Cline task.

---

# 3. SOURCE-OF-TRUTH FILES

Before Stage 00 begins, read:

```text
README.md
agents/project-context.md
docs/project-checkpoint.md
render.yaml
.clinerules
```

Then read:

```text
agents/clients/00-ghl-custom-payment-provider-backend-integration.md
```

Then inspect:

```text
clients/
transactions/
```

especially:

```text
clients/cmd/
clients/config/
clients/oauth/
clients/providers/
clients/webhooks/
clients/db/

transactions/cmd/
transactions/config/
transactions/db/
transactions/
```

Also inspect:

```text
integrations/
deposits/
```

because they may contain legacy behavior that must not be accidentally duplicated or deleted.

Inspect the current pawaPay integration and the PawaPay Go SDK/client used by RVPay.

Do not assume that documentation accurately describes the current code.

The repository is the implementation source of truth.

---

# 4. CURRENT ARCHITECTURAL INTENT

The intended ownership is:

## Clients service

Clients owns:

```text
HighLevel Marketplace integration
HighLevel OAuth
OAuth callback
OAuth state protection
HighLevel access/refresh credentials
HighLevel location association
HighLevel Marketplace webhook
HighLevel API communication
HighLevel Custom Payment Provider registration
HighLevel Custom Payment Provider configuration
HighLevel provider installation lifecycle
```

Clients does NOT own:

```text
Deposits
Payouts
Payment state
Payment initiation
Payment status
pawaPay execution
Payment reconciliation
Payment-domain business rules
```

---

## Transactions service

Transactions owns:

```text
Merchants
Customers
Deposits
Payouts
Payment state
Payment lifecycle
Payment initiation
Provider execution
Provider reconciliation
HighLevel transaction correlation
Payment-provider webhook business processing
Payment verification business logic
```

Transactions is the payment domain.

---

## pawaPay

pawaPay remains behind the provider boundary:

```text
Transactions
     |
     v
pawaPay client / SDK
     |
     v
pawaPay
```

GHL code must never directly call pawaPay.

---

# 5. IMPORTANT ARCHITECTURAL CORRECTION

The current repository contains a Clients payment implementation created by Agent 00.

Do NOT assume that because it exists it should remain.

Inspect it carefully.

If the current implementation contains:

```text
clients/payments/
```

or equivalent payment-domain logic, determine which responsibilities belong in Transactions.

The desired result is:

```text
Clients
    |
    +---- GHL integration
    +---- OAuth
    +---- GHL provider registration
    +---- GHL transport adapter
    |
    +---- no payment-domain ownership


Transactions
    |
    +---- payment query
    +---- payment verification
    +---- payment webhook processing
    +---- transaction correlation
    +---- payment state
    +---- pawaPay
```

Do not preserve an incorrect architecture merely to avoid moving code.

---

# 6. HIGHLEVEL DOCUMENTATION

Use the current official HighLevel documentation as the external API source of truth.

Relevant documentation includes:

```text
https://marketplace.gohighlevel.com/docs/marketplace-modules/Payments/

https://marketplace.gohighlevel.com/docs/ghl/payments/create-integration/

https://marketplace.gohighlevel.com/docs/ghl/payments/fetch-config/

https://marketplace.gohighlevel.com/docs/2021-04-15/ghl/payments/create-config/

https://marketplace.gohighlevel.com/docs/Authorization/Scopes/
```

HighLevel currently documents:

```text
POST /payments/custom-provider/provider
```

for creating the app/location provider association.

It also documents:

```text
POST /payments/custom-provider/connect
```

for creating the provider configuration.

It documents fetching the configuration using:

```text
GET /payments/custom-provider/connect
```

and provider disconnection using the corresponding Custom Provider API.

The agent must verify the current request/response schemas from the official documentation before implementation.

Do not invent request fields.

Do not assume old examples are still authoritative.

---

# 7. HIGHLEVEL REQUIRED SCOPES

The implementation must account for the Custom Payment Provider scopes:

```text
payments/custom-provider.readonly
payments/custom-provider.write
```

The existing Marketplace application may also require payment-related scopes such as:

```text
payments/orders.readonly
payments/orders.write
payments/transactions.readonly
payments/subscriptions.readonly
```

Do not automatically add unnecessary scopes.

Determine which scopes the current RVPay Marketplace application actually requires.

If the repository cannot modify Marketplace configuration automatically, document the required scope changes clearly.

Do not pretend that a backend code change modifies the Marketplace application's scopes.

---

# 8. PROVIDER CONFIGURATION CONCEPT

The provider configuration is conceptually:

```text
name
description
imageUrl
locationId
queryUrl
paymentsUrl
supportsSubscriptionSchedule
```

The actual HighLevel API request schema must be verified against the current official documentation.

The backend must never hard-code:

```text
Render hostname
AWS hostname
frontend hostname
frontend route
```

The frontend payment URL must come from configuration.

For example:

```text
HIGHLEVEL_PAYMENT_URL
```

or the repository's established equivalent.

The backend query URL may be derived from the deployed public base URL if the repository has a safe mechanism for doing so.

If deriving URLs from a base URL, use a configuration variable such as:

```text
PUBLIC_BASE_URL
```

rather than detecting the Render hostname from the runtime environment.

The resulting architecture must remain portable to AWS.

---

# 9. ONE-TIME PAYMENT SCOPE

RVPay currently supports one-time payment behavior.

Therefore:

```text
supportsSubscriptionSchedule = false
```

unless the repository genuinely supports subscriptions.

Do NOT implement:

```text
recurring payments
subscription lifecycle
off-session payments
subscription workers
subscription database models
```

Do not claim support for functionality that does not exist.

---

# 10. STAGE 00 — REPOSITORY AND ARCHITECTURE AUDIT

## Objective

Understand exactly what Agent 00 changed and establish the correction plan.

## Tasks

Inspect:

```text
clients/payments/
clients/providers/
clients/oauth/
clients/webhooks/
clients/db/
transactions/
transactions/db/
transactions/cmd/
```

Determine:

1. Where GHL OAuth currently lives.
2. Where OAuth state is stored.
3. How location IDs are stored.
4. How HighLevel access tokens are stored.
5. How refresh tokens are stored.
6. Where HighLevel provider configuration is currently stored.
7. Where payment query handling currently lives.
8. Where payment webhook handling currently lives.
9. Where transaction state currently lives.
10. How deposits are represented.
11. How pawaPay IDs are represented.
12. Whether GHL transaction IDs already exist.
13. Whether GHL charge IDs already exist.
14. Whether persistent webhook idempotency already exists.
15. How Clients communicates with Transactions.
16. How Transactions communicates with pawaPay.
17. How REST endpoints are exposed.
18. How gRPC endpoints are exposed.
19. How Render currently exposes each service.
20. Whether Clients and Transactions share a physical database on Render.
21. Whether generated code is involved.
22. Which existing tests cover Agent 00 functionality.

Produce an explicit architecture map.

Do not modify code during the audit unless required to make a test or inspection possible.

## Stage 00 completion requirements

The stage is complete only when:

```text
current ownership is documented;
incorrect Clients payment responsibilities are identified;
missing GHL outbound registration is identified;
existing OAuth lifecycle is understood;
existing Transactions payment lifecycle is understood;
Render deployment topology is understood;
required migrations are identified;
no code changes were made unnecessarily.
```

## STOP

At the end of Stage 00:

* run relevant inspection/tests;
* provide an audit report;
* do not begin Stage 01.

The user must start a new Cline task for Stage 01.

---

# 11. STAGE 01 — PAYMENT RESPONSIBILITY CORRECTION

## Objective

Remove payment-domain ownership from Clients without breaking the GHL integration.

The result must be:

```text
Clients = GHL integration boundary

Transactions = payment boundary
```

## Tasks

Inspect the current:

```text
clients/payments/
```

implementation.

Separate responsibilities into:

### Clients responsibilities

Keep only:

```text
GHL request parsing where required
GHL authentication/credential validation where appropriate
GHL provider registration
GHL provider configuration
GHL API communication
GHL transport adapters
```

### Transactions responsibilities

Move:

```text
payment verification
transaction lookup
payment state interpretation
payment webhook business processing
transaction correlation
payment-domain decisions
```

into Transactions.

Do not blindly move HTTP handlers if doing so would violate the service's public architecture.

The key requirement is ownership of business logic.

---

# 12. PAYMENT QUERY LOCATION

The preferred architecture is:

```text
GoHighLevel
     |
     | POST queryUrl
     v
Transactions public HTTP endpoint
     |
     v
Transactions payment domain
     |
     v
transaction state
```

If the current Render/gRPC-Gateway architecture makes this impossible without unnecessary infrastructure, determine the smallest clean alternative.

A Clients HTTP adapter may exist only if necessary.

If Clients receives the request, it must behave as a thin adapter:

```text
GHL
 |
 v
Clients adapter
 |
 | gRPC
 v
Transactions
 |
 v
payment domain
```

Do NOT create payment business logic inside Clients merely because Clients receives HTTP traffic.

---

# 13. PAYMENT WEBHOOK LOCATION

Prefer:

```text
GHL
 |
 v
Transactions payment-provider webhook
 |
 v
Transactions payment domain
```

If the existing deployment topology requires Clients to be the public entry point, use:

```text
GHL
 |
 v
Clients transport adapter
 |
 | gRPC
 v
Transactions
```

The adapter must not contain payment-domain logic.

The existing Marketplace webhook:

```text
/webhooks/highlevel
```

must remain separate and functional.

---

# 14. REMOVE DUPLICATION

Do not maintain two independent implementations of:

```text
payment verification
payment correlation
payment status
payment webhook processing
```

If the existing Clients implementation and Transactions implementation overlap, establish one authoritative implementation in Transactions.

If an existing Clients package becomes unused after the migration, remove it rather than leaving dead architecture behind.

Do not remove code until repository-wide usage has been checked.

---

# 15. STAGE 01 TESTS

Run:

```text
gofmt
go test ./...
```

Also run focused tests for:

```text
Clients
Transactions
OAuth
webhooks
```

Verify:

```text
existing OAuth still works;
existing Marketplace webhook still works;
Transactions owns payment logic;
Clients no longer owns payment state/business logic.
```

## STOP

Do not begin Stage 02.

Start a new Cline task for Stage 02.

---

# 16. STAGE 02 — HIGHLEVEL PROVIDER REGISTRATION CLIENT

## Objective

Implement the missing outbound RVPay → HighLevel Custom Payment Provider API client.

This is the most important missing functionality from Agent 00.

The Clients service must be capable of making authenticated HighLevel API calls for an installed location.

---

# 17. HIGHLEVEL API CLIENT RESPONSIBILITY

Create or extend the existing HighLevel provider/client implementation.

It should support the necessary operations conceptually:

```text
CreateProviderAssociation
CreateProviderConfig
FetchProviderConfig
DisconnectProvider
```

Do not expose these as RVPay public endpoints.

They are outbound integrations.

Conceptually:

```text
Clients
   |
   v
HighLevel API client
   |
   | HTTPS
   v
HighLevel API
```

---

# 18. CREATE PROVIDER ASSOCIATION

Implement the outbound call to:

```text
POST /payments/custom-provider/provider
```

This creates the association between the Marketplace app and the HighLevel location.

The implementation must:

* use the installed location's OAuth credentials;
* use the correct HighLevel API base URL;
* send the correct authorization header;
* use the current documented request body;
* handle 2xx responses;
* handle 400 responses;
* handle 401 responses;
* handle 422 responses;
* avoid logging access tokens;
* return typed/domain errors.

Do not hard-code a location ID.

---

# 19. CREATE PROVIDER CONFIGURATION

Implement the outbound call to:

```text
POST /payments/custom-provider/connect
```

The request must be built from RVPay configuration and the correct location.

Conceptually:

```text
buildProviderConfig()
    |
    +-- name
    +-- description
    +-- imageUrl
    +-- locationId
    +-- paymentsUrl
    +-- queryUrl
    +-- capability flags
```

Do not blindly reproduce the old Next.js object.

Adapt it to the actual current HighLevel API request schema.

---

# 20. OLD FRONTEND POC

The legacy frontend used a configuration conceptually equivalent to:

```text
{
    name: PROVIDER_NAME,
    description: PROVIDER_DESCRIPTION,
    imageUrl: `${baseUrl}/logo.jpg`,
    paymentsUrl: `${baseUrl}/payment/checkout`,
    queryUrl: `${baseUrl}/api/pawa/payments/query`,
    webhookUrl: `${baseUrl}/api/pawa/webhook`,
    supportsSubscriptionSchedule: false,
}
```

This is the historical POC behavior.

Use it as a functional reference, NOT as an authoritative API schema.

The new backend must reproduce the required behavior while respecting the current RVPay architecture.

Do not reintroduce the Next.js frontend.

---

# 21. PROVIDER URL CONFIGURATION

The backend must support:

```text
HIGHLEVEL_PAYMENT_URL
```

or equivalent.

For the backend query URL, prefer a configuration such as:

```text
PUBLIC_BASE_URL
```

plus a stable backend route.

For example:

```text
PUBLIC_BASE_URL=https://rvpay-transactions.onrender.com
```

could produce:

```text
https://rvpay-transactions.onrender.com/...
```

Do not hard-code:

```text
.onrender.com
```

anywhere in Go code.

The same configuration should work later with:

```text
https://api.example.com
```

on AWS.

---

# 22. IMAGE URL

The provider configuration requires a publicly accessible image URL if HighLevel requires it for the current provider configuration.

Do not invent a frontend implementation.

Use configuration:

```text
HIGHLEVEL_PROVIDER_IMAGE_URL
```

unless an existing repository convention provides a better mechanism.

Do not create a new image server.

Do not embed a binary asset into the Go application unless required.

---

# 23. PROVIDER CONFIGURATION STORAGE

Determine whether RVPay actually needs to persist a provider configuration record.

If it does, the authoritative payment-provider configuration should belong to the payment/integration model that logically owns it.

Do not create:

```text
clients/payment_provider_configs
```

merely because the configuration contains payment terminology.

Determine whether the configuration belongs to:

```text
Clients integration
```

or:

```text
Transactions payment provider
```

using the ownership rule:

```text
GHL installation identity
    -> Clients

payment-provider configuration/state
    -> Transactions
```

Avoid duplicate storage.

If a migration is required, create both:

```text
up migration
down migration
```

with:

```text
foreign keys
unique constraints
indexes
```

as appropriate.

---

# 24. STAGE 02 TESTS

Use:

```text
httptest.Server
```

or an injected HTTP transport.

Do NOT call HighLevel production.

Test:

```text
CreateProviderAssociation
CreateProviderConfig
FetchProviderConfig
DisconnectProvider
```

including:

```text
success
400
401
422
network error
malformed response
missing access token
context cancellation
timeout
```

Verify tokens are not present in logs or error strings.

## STOP

Do not continue to Stage 03.

Start a new Cline task.

---

# 25. STAGE 03 — HIGHLEVEL REGISTRATION LIFECYCLE

## Objective

Connect the HighLevel API client to the existing OAuth/install lifecycle.

The critical flow must become:

```text
HighLevel installation
        |
        v
/oauth/callback
        |
        v
exchange OAuth code
        |
        v
persist integration
        |
        v
obtain locationId
        |
        v
CreateProviderAssociation
        |
        v
CreateProviderConfig
        |
        v
provider available to location
```

---

# 26. DO NOT REPLACE OAUTH

The existing:

```text
/oauth/callback
```

must remain.

Do not create:

```text
/oauth/highlevel/callback
```

or another parallel OAuth system.

Preserve:

```text
state generation
state validation
code exchange
token persistence
refresh token behavior
```

unless a specific defect is discovered.

---

# 27. REGISTRATION FAILURE HANDLING

Determine how provider registration failures should affect OAuth installation.

Do not automatically roll back a successful OAuth installation unless the existing domain model explicitly supports transactional rollback.

A sensible model may be:

```text
OAuth installation
        |
        +---- success
        |
        v
provider registration
        |
        +---- success -> configured
        |
        +---- failure -> installed but provider configuration pending/failed
```

Do not leave the database in an ambiguous state.

The final model must make it possible to retry provider registration safely.

---

# 28. IDEMPOTENT REGISTRATION

The registration process may run more than once.

Therefore:

```text
CreateProviderAssociation
CreateProviderConfig
```

must be handled safely when HighLevel reports that the provider already exists.

Determine the correct current HighLevel response behavior.

Do not blindly treat every 400/422 as fatal.

If the provider is already associated/configured:

```text
fetch existing configuration
```

or otherwise perform the appropriate idempotent operation.

Do not create duplicate provider records.

---

# 29. RETRY BEHAVIOR

Do not implement aggressive retry loops.

For transient failures:

```text
network timeout
5xx
temporary connection failure
```

use the repository's existing retry conventions if present.

If none exist, keep retries minimal and bounded.

Do not block OAuth indefinitely.

---

# 30. STAGE 03 TESTS

Test:

```text
OAuth success
OAuth state failure
token exchange failure
provider association success
provider association failure
provider config success
provider config failure
already-configured location
retry-safe behavior
```

Use mocked HighLevel HTTP endpoints.

## STOP

Do not continue.

Start a new Cline task for Stage 04.

---

# 31. STAGE 04 — TRANSACTIONS PAYMENT-PROVIDER BACKEND CONTRACT

## Objective

Implement the actual HighLevel payment-provider backend contract inside Transactions.

HighLevel documents that its payment flow uses a configured:

```text
paymentsUrl
queryUrl
```

and sends verification requests to `queryUrl`.

The current documentation describes a verification payload containing fields such as:

```text
type
transactionId
apiKey
chargeId
subscriptionId
```

and expects responses such as:

```json
{"success":true}
```

or:

```json
{"failed":true}
```

or a pending representation.

Verify the exact current contract before implementation.

---

# 32. QUERY ENDPOINT

Implement:

```text
POST <configured-query-route>
```

for:

```text
type = verify
```

The HTTP layer must be thin.

Conceptually:

```text
HTTP
 |
 v
decode request
 |
 v
authenticate provider request
 |
 v
Transactions service
 |
 v
payment verification
 |
 v
response
```

Do not place transaction status logic in the HTTP handler.

---

# 33. VERIFICATION

The Transactions service must determine:

```text
successful
failed
pending
unknown
```

from authoritative transaction state.

Do not infer success merely from:

```text
chargeId exists
```

Do not infer success merely from:

```text
HighLevel webhook received
```

Use the Transactions domain.

---

# 34. HIGHLEVEL TRANSACTION IDENTIFIERS

Explicitly model the relationship between:

```text
HighLevel transactionId
HighLevel chargeId
RVPay transaction ID
pawaPay deposit ID
pawaPay provider transaction ID
```

Do not assume they are interchangeable.

Document the mapping.

If the current database lacks the required fields, add only the fields needed.

Prefer structured columns over arbitrary JSON.

---

# 35. PAYMENT INITIATION

Determine the actual current Transactions payment initiation API.

The frontend may eventually send HighLevel's:

```text
payment_initiate_props
```

information to the frontend/backend flow.

Do not implement the frontend.

Do ensure the backend exposes whatever payment operation the existing architecture requires.

If the Transactions service already has the correct payment initiation operation:

```text
reuse it
```

Do not create a duplicate payment operation.

---

# 36. PAWAPAY

The payment flow must remain:

```text
HighLevel
    |
    v
RVPay Transactions
    |
    v
pawaPay client/SDK
    |
    v
pawaPay
```

Never:

```text
HighLevel
    |
    v
Clients
    |
    v
pawaPay
```

Never put pawaPay HTTP logic into:

```text
GHL handlers
OAuth handlers
HighLevel registration client
```

---

# 37. STAGE 04 TESTS

Test:

```text
verify success
verify failure
verify pending
verify unknown transaction
invalid transaction
malformed request
unsupported query type
unsupported HTTP method
provider authentication failure
```

## STOP

Start a new Cline task for Stage 05.

---

# 38. STAGE 05 — TRANSACTION CORRELATION AND PAYMENT WEBHOOK

## Objective

Implement the HighLevel payment-provider webhook as a Transactions payment-domain operation.

HighLevel documents payment-related webhook events including:

```text
payment.captured
```

and subscription events.

RVPay only supports one-time payments.

Therefore implement:

```text
payment.captured
```

and safely ignore/reject unsupported subscription events.

---

# 39. WEBHOOK ENDPOINT

The endpoint must be separate from:

```text
/webhooks/highlevel
```

The existing Marketplace webhook must remain untouched.

The payment-provider webhook must have its own route.

The exact route may follow the existing repository naming convention, but it must be stable and configurable.

---

# 40. WEBHOOK AUTHENTICATION

Do NOT assume the existing Marketplace webhook signature applies.

The existing Marketplace webhook uses its own authentication mechanism.

For Custom Payment Provider requests, inspect the current HighLevel documentation and implement the authentication mechanism actually required.

If HighLevel provides:

```text
apiKey
```

as the provider credential:

```text
validate it
```

Do not accept arbitrary keys.

Do not log it.

Do not return it in errors.

Do not confuse it with:

```text
HIGHLEVEL_CLIENT_SECRET
```

or:

```text
PAWAPAY_API_KEY
```

---

# 41. PAYMENT CAPTURED

For:

```text
payment.captured
```

correlate:

```text
chargeId
ghlTransactionId
locationId
```

with the RVPay transaction.

The resulting transaction operation must be performed by Transactions.

Clients must not independently update payment state.

---

# 42. IDEMPOTENCY

Payment webhooks can be delivered multiple times.

The implementation must use persistent idempotency.

Do NOT use:

```text
map[string]bool
```

Do NOT use:

```text
sync.Map
```

as the authoritative mechanism.

Do NOT use:

```text
sleep
```

or timing tricks.

Use PostgreSQL or the existing persistent event/idempotency mechanism.

The database operation must safely handle concurrent duplicate deliveries.

Prefer a unique constraint on an appropriate event identifier.

If HighLevel does not provide a globally unique event ID, derive a deterministic event identity from the documented event fields only if this is safe and collision-resistant for the actual contract.

Do not invent a weak key.

---

# 43. CONCURRENT WEBHOOKS

The following must be safe:

```text
request A -> payment.captured
request B -> same payment.captured
```

Both arriving simultaneously.

Only one should perform the underlying state-changing operation.

The final transaction state must remain correct.

---

# 44. STAGE 05 TESTS

Test:

```text
valid payment.captured
duplicate payment.captured
three identical events
concurrent duplicate events
invalid credentials
unknown event
subscription event
missing transaction
wrong location
wrong charge ID
correlation failure
database conflict
```

## STOP

Start a new Cline task for Stage 06.

---

# 45. STAGE 06 — SECURITY, IDEMPOTENCY AND ERROR BOUNDARIES

## Objective

Perform a dedicated security pass.

Verify separation of:

```text
GHL Marketplace client ID
GHL Marketplace client secret
GHL access token
GHL refresh token
GHL provider API key
GHL webhook authentication
pawaPay API key
```

These credentials must never be interchangeable.

---

# 46. SECRET LOGGING

Search the entire repository for logging of:

```text
accessToken
refreshToken
clientSecret
apiKey
PAWAPAY_API_KEY
```

Ensure secrets cannot appear in:

```text
logs
errors
HTTP responses
test failure messages
```

---

# 47. LOCATION ISOLATION

A request for:

```text
location A
```

must not be able to retrieve or manipulate:

```text
location B
```

Provider credentials must be associated with the correct location/integration.

Test cross-location access explicitly.

---

# 48. ERROR MAPPING

External GHL responses must never expose internal details.

Never expose:

```text
SQL errors
database credentials
pawaPay errors containing secrets
HighLevel tokens
stack traces
internal filesystem paths
```

Return only the documented provider contract.

Internally retain useful diagnostics.

---

# 49. HTTP METHOD SAFETY

Verify:

```text
GET query endpoint -> 405
GET webhook -> 405
POST accepted
malformed POST -> 400
unauthenticated POST -> appropriate auth failure
```

Do not let unsupported methods enter domain logic.

---

# 50. STAGE 06 TESTS

Run security-focused tests.

Also run:

```text
go test ./...
go test -race ./...
```

if practical.

## STOP

Start a new Cline task for Stage 07.

---

# 51. STAGE 07 — RENDER DEPLOYMENT AND CONFIGURATION

## Objective

Make the corrected system deployable to Render.

Do NOT implement AWS infrastructure.

However, the design must remain AWS-portable.

---

# 52. RENDER RULES

Inspect:

```text
render.yaml
```

Do not create a new service unless absolutely necessary.

Prefer the existing service topology.

Do not introduce:

```text
new Redis
new queue
new worker
new load balancer
new database
```

unless the existing architecture requires one.

---

# 53. PUBLIC URL REQUIREMENT

HighLevel requires publicly reachable URLs.

The deployed configuration must make it possible to configure:

```text
PUBLIC_BASE_URL
HIGHLEVEL_PAYMENT_URL
HIGHLEVEL_PROVIDER_IMAGE_URL
```

and any other required values.

Do not hard-code Render URLs.

---

# 54. TEMPORARY RENDER DEPLOYMENT

The deployment must support:

```text
Render
    |
    +-- Clients
    |
    +-- Transactions
    |
    +-- PostgreSQL
```

according to the repository's existing deployment model.

If the current free-tier architecture uses a shared PostgreSQL database:

```text
preserve it
```

and document the temporary nature.

Do not redesign the database topology.

---

# 55. AWS PORTABILITY

Do NOT create:

```text
AWS CDK
Terraform
CloudFormation
ECS
EKS
Lambda
API Gateway
RDS
```

in this agent.

However:

* do not hard-code Render hostnames;
* do not depend on Render-specific APIs;
* do not store URLs in Go constants;
* use environment configuration;
* keep HTTP/gRPC boundaries portable;
* keep PostgreSQL usage standard;
* keep external provider integrations behind interfaces.

Later AWS deployment should primarily require infrastructure/configuration changes, not payment-domain rewrites.

---

# 56. RENDER ENVIRONMENT VARIABLES

Determine the actual required environment variables.

At minimum consider:

```text
HIGHLEVEL_CLIENT_ID
HIGHLEVEL_CLIENT_SECRET
HIGHLEVEL_REDIRECT_URI

HIGHLEVEL_API_BASE_URL

HIGHLEVEL_PROVIDER_NAME
HIGHLEVEL_PROVIDER_DESCRIPTION
HIGHLEVEL_PROVIDER_IMAGE_URL

HIGHLEVEL_PAYMENT_URL
PUBLIC_BASE_URL
```

and the existing:

```text
DATABASE_URL
PAWAPAY_API_URL
PAWAPAY_API_KEY
```

Use the repository's actual configuration naming conventions.

Do not duplicate existing variables.

---

# 57. RENDER HEALTH CHECKS

Verify the deployed services expose appropriate health endpoints.

Do not make HighLevel's provider endpoint the health check.

Health checks must not require:

```text
HighLevel authentication
pawaPay
database writes
```

unless the existing architecture intentionally requires it.

---

# 58. STAGE 07 VERIFICATION

Verify:

```text
render.yaml
environment configuration
service startup
database migration
Clients startup
Transactions startup
gRPC connectivity
REST/gRPC-Gateway routing
public URL construction
```

Do not claim Render readiness unless the configuration has actually been validated.

## STOP

Start a new Cline task for Stage 08.

---

# 59. STAGE 08 — END-TO-END LOCAL INTEGRATION TESTS

## Objective

Prove the entire server-side flow without contacting production HighLevel or production pawaPay.

Use:

```text
httptest.Server
mock HighLevel API
mock pawaPay transport/client
test PostgreSQL where appropriate
```

---

# 60. TEST FLOW A — OAUTH

Simulate:

```text
HighLevel
    |
    v
/oauth/callback
    |
    v
OAuth token exchange
    |
    v
integration persistence
    |
    v
provider association
    |
    v
provider config
```

Verify all stages.

---

# 61. TEST FLOW B — PROVIDER CONFIGURATION

Verify the generated provider configuration contains the expected values:

```text
name
description
imageUrl
locationId
queryUrl
paymentsUrl
capabilities
```

Verify:

```text
no Render hostname is hard-coded;
frontend URL comes from configuration;
locationId comes from the installed integration;
subscription support is disabled.
```

---

# 62. TEST FLOW C — PAYMENT VERIFICATION

Simulate:

```text
HighLevel
    |
    v
queryUrl
    |
    v
Transactions
    |
    v
transaction state
```

Test:

```text
success
failure
pending
unknown
```

---

# 63. TEST FLOW D — PAYMENT WEBHOOK

Simulate:

```text
HighLevel
    |
    v
payment.captured
    |
    v
Transactions
    |
    v
correlation
    |
    v
transaction state
```

Then send the same event repeatedly.

Verify the operation is idempotent.

---

# 64. TEST FLOW E — PAWAPAY

Mock the pawaPay boundary.

Verify:

```text
Transactions
    |
    v
pawaPay client
```

and ensure Clients never calls pawaPay directly.

---

# 65. TEST FLOW F — EXISTING FEATURES

Regression-test:

```text
HighLevel OAuth
/oauth/callback
/webhooks/highlevel
legacy services
Transactions
pawaPay integration
```

Do not break unrelated functionality.

---

# 66. STAGE 08 VERIFICATION

Run:

```text
gofmt
go test ./...
go test -race ./...
go vet ./...
go build ./...
```

Do not claim success unless each actually succeeds.

## STOP

Start a new Cline task for Stage 09.

---

# 67. STAGE 09 — REAL HIGHLEVEL / RENDER CONNECTIVITY READINESS

## Objective

Prepare the application to be connected to a real HighLevel test location.

This stage must not call HighLevel production automatically during tests.

It prepares the operator/developer workflow.

---

# 68. REQUIRED HIGHLEVEL MARKETPLACE CONFIGURATION

Document the required Marketplace application configuration:

```text
Client ID
Client Secret
Redirect URL
required scopes
Marketplace webhook URL
```

The Redirect URL must point to the deployed Clients service:

```text
https://<deployed-client-host>/oauth/callback
```

Do not hard-code the actual hostname into source code.

---

# 69. HIGHLEVEL PAYMENT PROVIDER CONFIGURATION

After deployment, the system must be able to register:

```text
provider association
provider configuration
```

through HighLevel's API.

The user must not need to manually reproduce the old Next.js POC logic.

The backend must perform the registration as part of the installation/configuration lifecycle.

---

# 70. REAL CONNECTIVITY CHECKLIST

Prepare an operator checklist containing:

```text
1. Deploy Clients.
2. Deploy Transactions.
3. Verify public URLs.
4. Configure HighLevel OAuth redirect URL.
5. Configure required Marketplace scopes.
6. Configure RVPay environment variables.
7. Install the Marketplace app into a test location.
8. Complete OAuth.
9. Confirm RVPay obtains the location identity.
10. Confirm RVPay calls HighLevel provider association API.
11. Confirm RVPay calls HighLevel provider configuration API.
12. Confirm HighLevel recognizes the provider.
13. Confirm queryUrl is publicly reachable.
14. Confirm payment-provider webhook URL is publicly reachable.
15. Confirm the frontend payment URL is correctly configured.
16. Perform a test payment.
17. Verify transaction state.
18. Verify HighLevel verification.
19. Verify payment.captured handling.
20. Verify duplicate webhook behavior.
```

Do not claim that an item passed unless it was actually tested.

---

# 71. HIGHLEVEL MANUAL VERIFICATION

The agent should provide the exact HTTP/API evidence needed to diagnose failures.

For example:

```text
provider association request
provider configuration request
response status
response body
locationId
```

Never print:

```text
access token
refresh token
client secret
provider API key
pawaPay API key
```

Use redacted values in logs and reports.

---

# 72. STAGE 09 COMPLETION

The stage is complete when:

```text
Render deployment configuration is ready;
HighLevel OAuth redirect is documented;
required scopes are documented;
provider registration lifecycle is implemented;
query URL is known;
webhook URL is known;
payment URL is configurable;
operator verification procedure exists.
```

## STOP

Start a new Cline task for Stage 10.

---

# 73. STAGE 10 — FINAL PRODUCTION REVIEW

## Objective

Perform a complete review of the implementation.

This is a review stage, not a feature-development stage.

---

# 74. ARCHITECTURE REVIEW

Confirm:

```text
Clients owns GHL integration.
Transactions owns payments.
Transactions owns payment state.
Transactions owns pawaPay interaction.
GHL registration is performed by Clients.
GHL payment queries reach Transactions.
GHL payment webhook business logic reaches Transactions.
No duplicate payment-domain implementation exists.
```

---

# 75. SECURITY REVIEW

Confirm:

```text
OAuth credentials are isolated.
Provider API keys are isolated.
Webhook authentication is isolated.
pawaPay credentials are isolated.
Secrets are not logged.
Location boundaries are enforced.
Tokens are not exposed through API responses.
```

---

# 76. DEPLOYMENT REVIEW

Confirm:

```text
Render deployment is supported.
Environment variables are documented.
No Render hostname is hard-coded.
No AWS-specific code exists.
AWS portability is preserved.
Database migrations are included.
Health checks work.
Service startup works.
```

---

# 77. PAYMENT REVIEW

Confirm:

```text
payment initiation belongs to Transactions;
verification belongs to Transactions;
payment state belongs to Transactions;
pawaPay interaction belongs to Transactions;
GHL registration belongs to Clients;
GHL OAuth belongs to Clients;
payment webhook business logic belongs to Transactions;
```

---

# 78. HIGHLEVEL REVIEW

Confirm against current HighLevel documentation:

```text
provider association endpoint
provider configuration endpoint
query contract
webhook contract
authentication mechanism
required scopes
one-time capability configuration
```

If HighLevel documentation changed during implementation, use the current official documentation and update the implementation accordingly.

Do not rely solely on the original Agent 00 document.

---

# 79. FINAL COMMANDS

Run:

```text
gofmt
go test ./...
go test -race ./...
go vet ./...
go build ./...
```

Also run:

```text
git diff --check
```

and inspect:

```text
git status
```

Do not claim success unless the commands actually succeed.

---

# 80. DATABASE REVIEW

Review all migrations.

Verify:

```text
up migrations
down migrations
indexes
unique constraints
foreign keys
idempotency constraints
transaction correlation fields
```

If Render currently uses a shared PostgreSQL database:

```text
document the shared-database implications
```

Do not create duplicate domain tables.

---

# 81. DOCUMENTATION UPDATES

Update only where necessary:

```text
README.md
clients/README.md
agents/project-context.md
docs/project-checkpoint.md
```

The documentation must clearly state:

```text
Clients:
    GHL integration and provider registration.

Transactions:
    payment domain and payment-provider operations.

HighLevel:
    provider association/configuration is performed by RVPay backend.

Render:
    temporary deployment target.

AWS:
    future deployment target; no AWS infrastructure implemented here.
```

Remove any documentation that incorrectly says Clients owns payment business logic.

---

# 82. PROJECT CHECKPOINT UPDATE

Update:

```text
docs/project-checkpoint.md
```

to record this corrective agent.

Do not mark the work complete merely because code compiles.

The checkpoint should state what was actually completed.

It must explicitly distinguish:

```text
GHL registration implemented
GHL provider configuration implemented
payment logic owned by Transactions
query endpoint implemented
payment webhook implemented
transaction correlation implemented
persistent webhook idempotency implemented
Render deployment ready
real HighLevel verification status
```

If real HighLevel connectivity was not tested, say so.

Do not claim production verification.

---

# 83. PROJECT CONTEXT UPDATE

Update:

```text
agents/project-context.md
```

with concise architectural rules:

```text
Clients owns GHL integration.
Transactions owns payments.
HighLevel registration is outbound from Clients.
HighLevel payment queries/webhooks delegate to Transactions.
pawaPay remains behind Transactions provider boundary.
Render is temporary deployment target.
AWS is future target.
```

Keep it concise.

---

# 84. FINAL ACCEPTANCE CRITERIA

The agent is COMPLETE only if all of the following are true.

## Architecture

```text
[ ] Clients owns GHL integration.
[ ] Transactions owns payment business logic.
[ ] No duplicate payment-domain logic remains in Clients.
[ ] pawaPay calls remain inside Transactions/provider boundary.
```

## HighLevel

```text
[ ] Existing OAuth remains functional.
[ ] /oauth/callback remains functional.
[ ] Provider association API is implemented.
[ ] Provider configuration API is implemented.
[ ] Provider configuration uses the correct locationId.
[ ] Provider configuration uses configurable URLs.
[ ] No Render hostname is hard-coded.
[ ] Required HighLevel scopes are documented.
[ ] One-time payment capability is represented correctly.
```

## Query

```text
[ ] Query endpoint exists.
[ ] POST is accepted.
[ ] Unsupported methods are rejected.
[ ] verify is implemented.
[ ] Success is returned correctly.
[ ] Failure is returned correctly.
[ ] Pending is returned correctly.
[ ] Transaction state comes from Transactions.
[ ] Provider API key is authenticated.
```

## Webhook

```text
[ ] Payment-provider webhook exists.
[ ] Marketplace webhook remains separate.
[ ] payment.captured is supported.
[ ] Unsupported subscription events are safely handled.
[ ] Authentication is implemented according to HighLevel's contract.
[ ] Webhook processing is persistent/idempotent.
[ ] Concurrent duplicate events are safe.
```

## Correlation

```text
[ ] GHL transactionId mapping is documented.
[ ] GHL chargeId mapping is documented.
[ ] RVPay transaction ID mapping is documented.
[ ] pawaPay deposit ID mapping is documented.
[ ] pawaPay provider transaction ID mapping is documented.
```

## Deployment

```text
[ ] Render deployment is supported.
[ ] Required environment variables are documented.
[ ] Public URLs are configurable.
[ ] Existing Render architecture is preserved where practical.
[ ] No unnecessary service was introduced.
[ ] No AWS infrastructure was added.
[ ] Architecture remains AWS-portable.
```

## Testing

```text
[ ] Unit tests pass.
[ ] Integration tests pass.
[ ] Race tests pass where practical.
[ ] go vet passes.
[ ] go build passes.
[ ] git diff --check passes.
[ ] No production HighLevel calls are made by automated tests.
[ ] No production pawaPay calls are made by automated tests.
```

## Documentation

```text
[ ] README is accurate.
[ ] project-context.md is accurate.
[ ] project-checkpoint.md is accurate.
[ ] Clients README is accurate.
[ ] Incorrect Agent 00 architectural claims are corrected.
```

---

# 85. IMPORTANT NON-GOALS

Do NOT implement:

```text
AWS infrastructure
Terraform
AWS CDK
CloudFormation
ECS
EKS
Lambda
API Gateway
RDS
```

Do NOT implement:

```text
subscription payments
recurring payments
off-session payments
refunds
chargebacks
```

unless a tiny compatibility change is absolutely required.

Do NOT implement:

```text
frontend checkout UI
frontend iframe
custom_provider_ready
payment_initiate_props frontend handling
custom_element_success_response frontend handling
custom_element_error_response frontend handling
custom_element_close_response frontend handling
```

Do NOT replace the existing HighLevel OAuth implementation.

Do NOT create a second OAuth system.

Do NOT introduce a new Render service merely to avoid making a clean architectural change.

Do NOT hard-code Render URLs.

Do NOT hard-code frontend URLs.

Do NOT directly call pawaPay from Clients.

---

# 86. REQUIRED STAGE REPORT FORMAT

At the end of EVERY stage, Cline must return:

```text
Stage:
<stage number and name>

Status:
COMPLETE / BLOCKED

Objective:
<one paragraph>

Changes:
- ...
- ...
- ...

Files Created:
- ...

Files Modified:
- ...

Files Removed:
- ...

Database Changes:
- ...

Tests Run:
- command
- command

Verification:
- PASS / FAIL

Unresolved Issues:
- ...

Architecture Impact:
- ...

Next Stage:
<next stage number>

IMPORTANT:
The next stage must be started as a NEW Cline task.
Do not continue automatically.
```

---

# 87. REQUIRED FINAL RESPONSE FORMAT

After Stage 10 only, provide:

```text
Implementation Summary

Architecture Correction
...

HighLevel Registration
...

Provider Configuration
...

Transactions Payment Domain
...

Payment Query
...

Payment Webhook
...

Authentication
...

Transaction Correlation
...

Idempotency
...

Render Deployment
...

Configuration
...

Files Created
...

Files Modified
...

Files Removed
...

Database Changes
...

Tests
...

Verification
...

HighLevel Connectivity Readiness
...

Known Issues
...

Deferred Features
...

AWS Portability
...

Agent Status
GHL Custom Payment Provider Integration: COMPLETE
```

Do not provide a long essay.

Do not claim real HighLevel connectivity unless it was actually tested.

Do not claim pawaPay production connectivity unless it was actually tested.

---

# 88. FINAL PRINCIPLE

The final architecture must make this statement true:

```text
HighLevel knows RVPay as a Custom Payment Provider.

Clients knows how to talk to HighLevel.

Transactions knows how to process payments.

pawaPay knows nothing about HighLevel.

HighLevel knows nothing about RVPay's internal database.

The frontend knows nothing about RVPay's internal payment implementation.

Render is only the temporary deployment environment.

AWS can replace Render later without redesigning the payment domain.
```

The most important architectural boundary is:

```text
                    GHL
                     |
                     v
                  Clients
                     |
          GHL integration only
                     |
                     v
               Transactions
                     |
              Payment domain
                     |
                     v
                  pawaPay
```

Do not compromise this boundary merely to make the implementation easier.

END OF AGENT
