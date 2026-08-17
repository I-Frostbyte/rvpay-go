# Fix HighLevel INSTALL Webhook Parsing — Minimal POC Fix

Make ONLY the changes required to make the current GHL `INSTALL` webhook work.

The timestamp parsing has ALREADY been fixed locally. Do not modify or undo that fix.

## Actual GHL payload

The webhook payload is:

```json
{
  "type": "INSTALL",
  "appId": "6a5f8aafdb5067f4319b1bb4",
  "versionId": "6a5f8aafdb5067f4319b1bb4",
  "installType": "Location",
  "locationId": "kSRxQkM72aCeYz19uw79",
  "companyId": "f3rBPevH93JANjvqtrK0",
  "userId": "K6MePugfKQPdgEicKzVJ",
  "companyName": "evaristustambua@gmail.com",
  "isWhitelabelCompany": false,
  "whitelabelDetails": {
    "logoUrl": "",
    "domain": ""
  },
  "trial": {},
  "timestamp": "2026-08-17T09:06:59.366Z",
  "webhookId": "f4ef22e3-c4c1-4ce5-996d-297890460e7d"
}
```

## Scope

Read only the relevant production files:

- clients/providers/highlevel_webhook.go
- clients/providers/webhook.go
- clients/webhooks/service.go
- existing payment-provider-config repository/query files
- relevant webhook tests

Do NOT read or modify unrelated services.

## Required changes

### 1. HighLevel payload

Update `HighLevelWebhookProvider.ParseEvent()` to parse the actual GHL schema.

Map:

- `type` → `WebhookEvent.EventType`
- `webhookId` → `WebhookEvent.ProviderEventID`
- `appId` → `WebhookEvent.IntegrationID`
- `companyId` → `WebhookEvent.ClientID`
- `locationId` → `WebhookEvent.LocationID`

Preserve the timestamp fix that is already present locally.

Do not invent fields or change unrelated provider behavior.

### 2. Normalized WebhookEvent

Add:

```go
LocationID string
```

to `providers.WebhookEvent`.

Do not change existing fields unnecessarily.

### 3. HighLevel integration resolution

IMPORTANT:

`appId` is NOT an RVPay UUID.

Do NOT call:

```go
uuid.Parse(event.IntegrationID)
```

for the HighLevel webhook.

The GHL `appId` identifies the Marketplace application and cannot uniquely identify an RVPay integration.

Instead, for HighLevel webhooks, resolve the actual RVPay integration UUID using:

```
event.LocationID
    ↓
existing payment_provider_configs.location_id
    ↓
payment_provider_configs.integration_id
```

The `payment_provider_configs` table already contains `location_id` and `integration_id`.

Reuse the existing payment-provider-config repository/query infrastructure.

DO NOT create a migration.

DO NOT add a new database column.

If an existing lookup by location ID does not exist, add only the smallest repository/query method necessary to retrieve the existing payment-provider configuration by `location_id`.

Then use its existing `integration_id` for the remainder of `ProcessWebhook()`.

### 4. INSTALL event

HighLevel sends:

```
type = INSTALL
```

Add support for:

```go
case "INSTALL":
```

in the HighLevel dispatcher.

Do not redesign the dispatcher.

Use the existing integration-installed handling where appropriate.

### 5. Do NOT change OAuth

Do not modify:

- clients/oauth/service.go
- clients/providers/highlevel.go

The existing OAuth LocationID fix is already correct.

### 6. Tests

Update ONLY the relevant webhook/provider tests.

Add/update a test using the exact GHL INSTALL payload above.

Verify:

- JSON parses successfully.
- `EventType == "INSTALL"`.
- `ProviderEventID == the webhookId`.
- `IntegrationID == the appId`.
- `ClientID == the companyId`.
- `LocationID == the locationId`.
- Timestamp remains correctly parsed using the existing local fix.
- HighLevel processing does NOT attempt to parse `appId` as a UUID.
- The integration is resolved through the existing `payment_provider_configs.location_id` lookup.
- INSTALL is dispatched.

Do not modify unrelated tests.

## Verification

Run only the relevant tests:

```bash
go test ./clients/providers/...
go test ./clients/webhooks/...
```

If one of those paths does not contain tests, run the smallest relevant package test.

Do not run the entire repository test suite unless necessary.

## Final diff audit

Run:

```bash
git diff --stat
git diff
```

The final diff must contain ONLY the minimal changes required for the GHL INSTALL webhook.

- No migrations.
- No architecture changes.
- No unrelated refactoring.
- No changes to payment flows.
- No changes to OAuth.
- No changes to Transactions.
- No commit.
- No push.

## Report

- Files changed.
- How GHL INSTALL is mapped.
- How `locationId` resolves the actual RVPay integration UUID.
- Tests run and results.
- Confirmation that no migration was created.