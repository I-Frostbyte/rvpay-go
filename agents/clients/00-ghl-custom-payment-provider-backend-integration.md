# Agent 00 — GHL Custom Payment Provider Backend Integration

## Role

You are a senior Go backend engineer specializing in payment integrations and HighLevel Marketplace applications.

Your task is to make the existing RVPay backend compatible with **GoHighLevel's Custom Payment Provider infrastructure**.

This is a backend-only task.

The frontend checkout page already exists or will be provided separately by the frontend developer.

DO NOT implement or modify the frontend checkout page.

DO NOT decide the frontend's public URLs.

DO NOT implement a new frontend.

DO NOT replace the existing GHL Marketplace OAuth integration.

The objective is to make the RVPay server capable of participating in HighLevel's payment-provider lifecycle.

---

# 1. CRITICAL SCOPE

This agent concerns:

```text
HighLevel Marketplace
        +
Custom Payment Provider
        +
RVPay Transactions
        +
pawaPay

It does NOT concern:

Frontend checkout UI
Frontend routing
Frontend payment iframe implementation
Frontend custom_provider_ready events
Frontend payment_initiate_props handling

The frontend developer owns those concerns.

The backend must simply provide the server-side contracts that HighLevel requires.
```

---

# 2. READ THESE FILES FIRST

Before modifying anything, read:

```text
README.md
render.yaml
.clinerules
project-context.md
project-checkpoint.md
```

Then inspect the existing Clients and Transactions services, especially:

```text
clients/
clients/cmd/grpc-service/
clients/config/
clients/oauth/
clients/webhooks/
clients/providers/
clients/db/
```

and:

```text
transactions/
transactions/cmd/
transactions/config/
transactions/db/
transactions/...
```

Also inspect the existing pawaPay integration and the current pawaPay SDK/client used by RVPay.

Do not assume the current architecture.

Determine the actual current implementation first.

---

# 3. EXISTING GHL OAUTH MUST BE PRESERVED

The repository already contains GHL Marketplace OAuth functionality.

Preserve:

```text
/oauth/callback
```

and the existing OAuth state protection.

Do NOT replace it.

Do NOT create a second OAuth system.

Do NOT move OAuth into Transactions unless there is a compelling architectural reason.

The existing Clients service remains responsible for GHL integration concerns.

---

# 4. UNDERSTAND HIGHLEVEL'S PAYMENT PROVIDER MODEL

HighLevel's current Custom Payment Provider documentation describes a provider configuration containing values including:

```text
name
description
imageUrl
locationId
queryUrl
paymentsUrl
```

and describes the installation flow in which the app is installed into a location and the payment provider configuration is created for that location.

HighLevel also expects payment-related backend queries to be sent to the configured queryUrl.

The server must therefore support the backend half of this contract.

Reference:

```text
https://marketplace.gohighlevel.com/docs/marketplace-modules/Payments/
```

and:

```text
https://marketplace.gohighlevel.com/docs/ghl/payments/custom-provider/
```

Do not blindly copy documentation examples.

Adapt them to the actual RVPay architecture.

---

# 5. FIRST PERFORM AN ARCHITECTURE AUDIT

Before writing code, determine:

```text
Which service owns GHL integration?
Which service owns transaction state?
Which service owns pawaPay interaction?
Where should HighLevel payment queries enter RVPay?
Where should payment-provider webhook events enter RVPay?
How are transactions currently identified?
How is a pawaPay deposit/payment correlated with a HighLevel transaction?
What database tables already exist?
What existing domain/service methods can be reused?
What existing authentication mechanisms are available?
```

Do not create duplicate functionality.

If an existing component already performs a required operation, reuse it.

---

# 6. TARGET ARCHITECTURE

The intended conceptual architecture is:

```text
                    GoHighLevel
                         |
             +-----------+-----------+
             |                       |
             |                       |
       Payment Query             Payment Webhook
             |                       |
             v                       v
       Clients / GHL            Clients / GHL
       Payment Handler          Payment Webhook
             |                       |
             +-----------+-----------+
                         |
                         v
                    Transactions
                         |
                         v
                      pawaPay
```

The exact package/service placement may differ depending on the existing repository.

Do not force this diagram literally if the existing architecture has a cleaner solution.

The important requirement is separation of responsibilities.

---

# 7. PAYMENT PROVIDER CONFIGURATION LIFECYCLE

Determine whether RVPay currently performs the HighLevel Custom Provider configuration step.

HighLevel provides APIs for:

```text
Create new integration
Create provider config
Fetch provider config
Disconnect provider config
```

The current HighLevel API documentation exposes these under the Custom Provider API.

Review:

```text
POST /payments/custom-provider/provider
POST /payments/custom-provider/connect
GET  /payments/custom-provider/connect
```

where appropriate.

Do NOT expose these HighLevel APIs from RVPay.

Instead, determine whether the existing Clients service needs internal service methods that allow RVPay to:

```text
associate an installed app/location with the payment provider;
create/update the provider configuration;
provide the correct backend query URL;
provide the frontend payment URL as configuration;
associate the configuration with the correct HighLevel location.
```

The frontend URL is configuration supplied to the backend.

Do not invent its value.

---

# 8. LOCATION ASSOCIATION

HighLevel's payment provider configuration is location-specific.

Ensure the existing OAuth flow retains enough information to identify:

```text
locationId
```

for the installed HighLevel account.

Determine whether the existing integration model already stores:

```text
location ID
HighLevel access token
HighLevel refresh token
integration/provider ID
```

If the repository already stores these, reuse them.

Do not introduce duplicate location/integration records.

If a migration is genuinely required, implement the smallest safe migration.

---

# 9. PROVIDER CONFIGURATION

Create or extend the existing GHL integration service so the payment provider configuration can be registered with HighLevel.

The configuration should support:

```text
name
description
imageUrl
locationId
queryUrl
paymentsUrl
supportsSubscriptionSchedule
```

However:

**Important**

The backend must NOT hard-code the frontend's checkout URL.

Use configuration.

For example:

```text
HIGHLEVEL_PAYMENT_URL
```

or the repository's established configuration convention.

The actual production value will be supplied by deployment configuration.

Likewise, the backend query endpoint must be represented by a stable configuration or route appropriate to the deployed service.

Do not hard-code Render hostnames.

---

# 10. PROVIDER CAPABILITIES

RVPay currently supports one-time payments.

Do not enable recurring payment functionality merely because HighLevel supports it.

The provider configuration must represent the actual RVPay capabilities.

If the existing provider configuration contains:

```text
supportsSubscriptionSchedule: false
```

preserve that behavior unless the repository genuinely supports subscription scheduling.

Do not implement subscriptions in this agent.

---

# 11. PAYMENT QUERY ENDPOINT

Implement the backend endpoint required by HighLevel's queryUrl.

The endpoint must accept POST requests.

The endpoint must inspect the HighLevel query operation.

At minimum support:

```text
type = verify
```

HighLevel's documented verification request contains fields such as:

```text
type
transactionId
apiKey
chargeId
subscriptionId
```

Only fields relevant to RVPay should be processed.

Do not blindly trust the incoming apiKey.

---

# 12. VERIFY OPERATION

For:

```text
type = verify
```

the server must determine whether the referenced RVPay transaction/payment has actually succeeded.

The flow should conceptually be:

```text
HighLevel
    |
    | verify transaction
    v
RVPay
    |
    v
Transactions
    |
    v
transaction state
    |
    +---- successful ----> success
    |
    +---- failed --------> failed
    |
    +---- pending -------> pending
```

Use the existing Transactions domain logic where possible.

Do not duplicate transaction-status logic inside the HTTP handler.

---

# 13. VERIFICATION RESPONSE

HighLevel's documented verification response supports:

```text
{
  "success": true
}
```

for success.

For failure:

```text
{
  "failed": true
}
```

For pending:

```text
{
  "success": false
}
```

Implement the response contract according to the current HighLevel documentation.

Do not return internal RVPay transaction objects directly.

Do not leak database models.

---

# 14. HIGHLEVEL API KEY AUTHENTICATION

HighLevel may send an apiKey associated with the configured provider.

This is distinct from:

```text
HIGHLEVEL_CLIENT_SECRET
```

and distinct from:

```text
pawaPay API key
```

Do not confuse these credentials.

Determine how RVPay should securely associate the configured HighLevel provider key with a location/integration.

The expected security model should be:

```text
HighLevel request
       |
       v
provider apiKey
       |
       v
RVPay validates key
       |
       v
location/integration
       |
       v
transaction
```

Do not simply accept any incoming apiKey.

Do not log it.

Do not return it in errors.

Do not store it in plaintext if the existing architecture supports secure credential storage.

If HighLevel requires a publishable key and backend API key configuration, distinguish their purposes.

---

# 15. PAYMENT PROVIDER WEBHOOK

Implement the backend webhook required for HighLevel Custom Payment Provider events.

This is NOT the same as the existing:

```text
/webhooks/highlevel
```

Marketplace application webhook.

Do not replace the existing endpoint.

The two mechanisms serve different purposes.

Existing:

```text
/webhooks/highlevel
```

continues to handle Marketplace/GHL application webhook events.

The new payment-provider webhook handles payment-related events.

---

# 16. PAYMENT PROVIDER WEBHOOK EVENTS

HighLevel documents payment-related events including:

```text
payment.captured
subscription.trialing
subscription.active
subscription.updated
subscription.charged
```

RVPay currently does not support subscriptions.

Therefore:

Implement only the payment event behavior that is relevant to the current RVPay one-time payment flow.

Do not implement subscription processing.

Unknown future event types should be handled safely rather than causing the entire webhook mechanism to become unusable.

---

# 17. PAYMENT CAPTURED

For:

```text
payment.captured
```

the server should be capable of correlating:

```text
chargeId
ghlTransactionId
locationId
```

with the corresponding RVPay transaction.

Do not assume chargeId and ghlTransactionId are interchangeable.

Determine how the existing Transactions model can represent the relationship.

If additional transaction metadata is required, make the smallest appropriate schema/domain change.

---

# 18. WEBHOOK IDEMPOTENCY

Payment webhooks can be delivered more than once.

The webhook handler must therefore be idempotent.

Do not process the same event repeatedly.

Use an existing idempotency mechanism if one already exists.

If one does not exist, introduce the smallest appropriate persistent mechanism.

Do NOT rely solely on:

```text
in-memory maps
```

because the Render service can restart.

Do NOT use:

```text
sleep
```

or timing-based deduplication.

The idempotency mechanism must survive service restarts.

---

# 19. WEBHOOK SECURITY

Do not assume the payment-provider webhook uses the same signature mechanism as the Marketplace app webhook.

The existing GHL Marketplace webhook uses:

```text
X-GHL-Signature
Ed25519
```

and that implementation must remain intact.

For the Custom Payment Provider webhook:

```text
inspect the current HighLevel documentation;
determine the actual authentication mechanism;
implement exactly what the provider contract requires;
do not invent a signature scheme.
```

If the webhook is authenticated using provider credentials rather than a signature, validate those credentials appropriately.

---

# 20. RAW REQUEST HANDLING

If the HighLevel payment webhook includes any signature verification mechanism, verification must operate on the exact raw request body.

Do not:

```text
decode JSON
re-marshal JSON
verify the re-marshaled body
```

Instead:

```text
raw body
   |
   v
verify
   |
   v
decode JSON
```

Preserve the existing Ed25519 implementation for the Marketplace webhook.

---

# 21. TRANSACTION CORRELATION

This is a critical requirement.

Determine how the following identifiers relate:

```text
HighLevel transactionId
HighLevel chargeId
RVPay transaction ID
pawaPay deposit ID
pawaPay provider transaction ID
```

Document the mapping.

Do not invent mappings.

If the existing transaction model lacks the required correlation data, add only the fields necessary to establish reliable correlation.

Avoid storing arbitrary JSON blobs when structured identifiers are sufficient.

---

# 22. PAYMENT INITIATION BOUNDARY

The payment provider flow eventually results in an RVPay payment being initiated.

The actual payment UI is owned by the frontend.

The backend must nevertheless expose the existing payment operation that the frontend will call after receiving HighLevel's payment initiation data.

Do not implement a new frontend-facing API unless the existing architecture genuinely lacks the necessary backend operation.

Inspect the current Transactions and pawaPay integration first.

Reuse the existing payment initiation flow where possible.

---

# 23. PAWAPAY BOUNDARY

The GHL integration must NOT directly contain pawaPay HTTP logic.

Maintain:

```text
GHL
 ↓
RVPay domain
 ↓
Transactions
 ↓
pawaPay client/SDK
 ↓
pawaPay
```

Do not put pawaPay API calls inside:

```text
GHL handlers
OAuth handlers
webhook handlers
```

Handlers should delegate to domain/service logic.

---

# 24. ERROR HANDLING

Do not expose internal errors directly to HighLevel.

Map errors into appropriate provider responses.

Internally preserve:

```text
HTTP status
provider error
transaction ID
underlying error
```

where useful.

Externally return only the contract HighLevel expects.

Never expose:

```text
database credentials
pawaPay API keys
HighLevel client secrets
HighLevel access tokens
```

---

# 25. AUTHENTICATION SEPARATION

Maintain strict separation between:

```text
GHL Marketplace OAuth credentials
GHL provider API key
GHL webhook authentication
pawaPay API credentials
```

Do not reuse one credential for another purpose.

Document each credential's purpose.

---

# 26. CONFIGURATION

Review the existing configuration system.

Add only configuration genuinely required for Custom Payment Provider support.

Potential configuration may include:

```text
HIGHLEVEL_PAYMENT_URL
HIGHLEVEL_QUERY_URL
HIGHLEVEL_PROVIDER_NAME
HIGHLEVEL_PROVIDER_DESCRIPTION
HIGHLEVEL_PROVIDER_IMAGE_URL
```

Use the repository's existing naming/configuration conventions.

Do not require users to configure values that can safely be derived from the deployed environment.

Do not hard-code Render hostnames.

---

# 27. RENDER

Review:

```text
render.yaml
```

Ensure the new backend functionality is deployable using the existing Clients/Transactions architecture.

Do not create another Render service unless the existing architecture makes it unavoidable.

Prefer the existing Clients service for GHL-facing endpoints.

Do not add infrastructure unnecessarily.

Do not modify unrelated Render services.

---

# 28. DATABASE

Before adding migrations, inspect existing schemas.

Do not create duplicate tables for:

```text
integrations
transactions
webhook_events
oauth_states
```

Reuse existing structures.

Only create migrations if required for:

```text
HighLevel provider configuration
transaction correlation
payment-provider webhook idempotency
```

If a migration is necessary:

```text
create up/down migrations;
use safe constraints;
add appropriate indexes;
document shared-database implications.
```

Remember that the current deployment may use a shared PostgreSQL database because of the Render free-tier constraint.

Do not assume each service has its own physical database.

---

# 29. EXISTING GHL WEBHOOK MUST REMAIN FUNCTIONAL

After implementation verify that:

```text
/webhooks/highlevel
```

still works.

Do not accidentally route payment-provider events into the existing Marketplace webhook handler.

The conceptual separation must remain:

```text
Marketplace webhook
    ↓
/webhooks/highlevel
```

and:

```text
Payment-provider webhook
    ↓
new payment-provider endpoint
```

---

# 30. HTTP ROUTING

Register the required backend routes in the appropriate deployed service.

The routes must:

```text
accept only the intended HTTP methods;
return 405 for unsupported methods;
return appropriate 4xx responses for malformed requests;
return appropriate success responses;
support HTTPS deployment;
not interfere with grpc-gateway routes.
```

Do not invent frontend paths.

The frontend developer will provide the payment URL separately.

---

# 31. QUERY ENDPOINT CONTRACT

The query endpoint should be designed so future supported operations can be added cleanly.

Use a structure conceptually similar to:

```text
POST
  |
  +-- verify
  |
  +-- future operation
```

Do not implement:

```text
refund
```

unless it already exists and is required by the current architecture.

Refunds are explicitly deferred.

Unknown operations should return a safe unsupported-operation response.

---

# 32. SUBSCRIPTION SCOPE

RVPay is currently a one-time payment provider.

Do NOT implement:

```text
subscriptions
recurring charges
subscription.trialing
subscription.active
subscription.updated
subscription.charged
```

beyond safely recognizing/ignoring unsupported event types where required.

Do not introduce subscription database models.

Do not introduce subscription workers.

---

# 33. TESTS

Add comprehensive backend tests.

Do NOT call HighLevel production APIs.

Do NOT call pawaPay production APIs.

Use:

```text
httptest.Server
mock services
mock transport
```

where appropriate.

Test at minimum:

Provider configuration

```text
configuration is constructed correctly;
location ID is handled correctly;
payment URL is configurable;
query URL is correct;
unsupported subscription capability is represented correctly.
```

Query endpoint

```text
POST accepted;
GET rejected;
malformed JSON rejected;
missing type rejected;
unsupported type rejected;
valid verify request;
successful transaction;
failed transaction;
pending transaction;
unknown transaction;
invalid provider API key;
credential safety.
```

Webhook

```text
POST accepted;
GET rejected;
malformed JSON;
missing authentication;
invalid authentication;
valid payment event;
duplicate event;
unknown event;
transaction correlation failure;
safe response behavior.
```

Existing GHL OAuth

Verify that existing behavior remains intact.

Existing Marketplace webhook

Verify:

```text
/webhooks/highlevel
```

still behaves correctly.

---

# 34. IDEMPOTENCY TESTS

Explicitly test duplicate payment-provider webhook delivery.

The test should simulate:

```text
event
event
event
```

and confirm that the underlying transaction operation is performed only once.

Test concurrency where practical:

```text
same event
   +---- request A
   |
   +---- request B
```

The final state must remain correct.

---

# 35. SECURITY TESTS

Verify:

```text
provider API keys are not logged;
GHL client secrets are not logged;
pawaPay credentials are not logged;
authentication failures do not reveal secrets;
malformed requests cannot bypass authentication;
webhook authentication cannot be bypassed;
location IDs cannot be arbitrarily used to access another integration.
```

---

# 36. PUBLIC API COMPATIBILITY

Before changing exported types or methods:

Search the repository for all usages.

Do not silently break existing Clients or Transactions APIs.

Prefer adding narrowly scoped methods over replacing working public APIs.

If a breaking change is absolutely necessary:

```text
document it;
explain why;
update tests;
update project-checkpoint.md.
```

---

# 37. DOCUMENTATION

Update:

```text
README.md
clients/README.md
project-context.md
project-checkpoint.md
```

only where necessary.

Document:

```text
GHL Custom Payment Provider architecture
OAuth relationship
provider configuration
query endpoint
payment-provider webhook
authentication
transaction correlation
configuration variables
```

Do not write a giant manual.

Do not document frontend implementation details.

---

# 38. PROJECT CONTEXT

Update:

```text
project-context.md
```

with the minimum information future agents need.

Include:

```text
GHL Marketplace integration
GHL Custom Payment Provider integration
payment query endpoint
payment-provider webhook
provider authentication
transaction correlation
configuration
unsupported capabilities
```

Keep it concise.

---

# 39. PROJECT CHECKPOINT

Update:

```text
project-checkpoint.md
```

to record this agent.

Use the repository's existing agent numbering convention.

Record:

```text
GHL Custom Payment Provider integration: COMPLETED
```

and summarize:

```text
routes;
provider configuration;
authentication;
transaction correlation;
webhook idempotency;
tests;
remaining limitations.
```

---

# 40. VERIFICATION

Run:

```text
gofmt
go test ./...
go vet ./...
go build ./...
```

If practical:

```text
go test -race ./...
```

Do not claim success unless commands actually succeed.

If something fails:

```text
determine whether the failure was caused by this agent;
fix only issues within scope;
document unrelated failures.
```

---

# 41. FINAL REVIEW CHECKLIST

Before finishing, verify:

```text
 Existing GHL OAuth remains functional.
 Existing /oauth/callback remains functional.
 Existing /webhooks/highlevel remains functional.
 Custom Payment Provider configuration is supported.
 Location association is supported.
 Provider configuration can specify the frontend payment URL without hard-coding it.
 Provider configuration exposes the backend query URL.
 Payment query endpoint exists.
 verify operation works.
 Transaction status is sourced from RVPay's transaction domain.
 Payment-provider webhook endpoint exists.
 Payment-provider webhook authentication is implemented according to the current HighLevel contract.
 Webhook processing is idempotent.
 HighLevel transaction identifiers can be correlated with RVPay transactions.
 pawaPay credentials remain isolated.
 GHL credentials remain isolated.
 No frontend code was introduced.
 No frontend URLs were hard-coded.
 No subscription functionality was implemented.
 No refund functionality was implemented.
 No new server/service was unnecessarily introduced.
 Tests do not use live HighLevel or pawaPay services.
 README is accurate.
 project-context.md is updated.
 project-checkpoint.md is updated.
 gofmt passes.
 tests pass.
 vet passes.
 build passes.
```

---

# 42. IMPORTANT NON-GOALS

Do NOT implement:

```text
Frontend checkout page
Frontend payment iframe
custom_provider_ready
payment_initiate_props
custom_element_success_response
custom_element_error_response
custom_element_close_response
Refunds
Subscriptions
Recurring payments
Off-session payments
```

unless a tiny backend change is strictly required to support the server-side contract.

The frontend developer owns frontend payment behavior.

---

# 43. FINAL RESPONSE

Return a concise implementation report using:

```text
Implementation Summary
GHL Payment Provider Architecture

...

Provider Configuration

...

Payment Query

...

Payment Webhook

...

Authentication

...

Transaction Correlation

...

Files Created

...

Files Modified

...

Database Changes

...

Tests

...

Verification

...

Known Issues

...

Deferred Features

...

Agent Status
GHL Custom Payment Provider Backend: COMPLETE
```

Do not provide a long essay.

---

**END OF AGENT**