# HighLevel / Render Connectivity Checklist

## Prerequisites

Before starting, ensure:

- [ ] Oracle Cloud Free Tier or Render account is active
- [ ] RVPay backend is deployed to Render (both Clients and Transactions services)
- [ ] PostgreSQL database is provisioned and migrations have run
- [ ] All required environment variables are configured in Render dashboard

---

## 1. Deploy Clients

```bash
# Render auto-deploys on push to main.
# Verify deployment succeeds in Render dashboard.
# Alternatively, deploy manually:
git push origin main
```

- [ ] Clients service starts without errors
- [ ] Health check responds at `https://<clients-host>/healthz` (HTTP GET → 200)

## 2. Deploy Transactions

- [ ] Transactions service starts without errors
- [ ] Health check responds at `https://<transactions-host>/healthz` (HTTP GET → 200)

## 3. Verify Public URLs

Record the deployed URLs — these are needed for HighLevel configuration:

| Purpose | Example Value |
| --- | --- |
| Clients public URL | `https://rvpay-clients.onrender.com` |
| Transactions public URL | `https://rvpay-transactions.onrender.com` |
| OAuth callback URL | `https://rvpay-clients.onrender.com/oauth/callback` |
| Marketplace webhook URL | `https://rvpay-clients.onrender.com/webhooks/highlevel` |
| Payment query URL | `https://rvpay-clients.onrender.com/payments/custom-provider/query` |
| Payment webhook URL | `https://rvpay-clients.onrender.com/payments/custom-provider/webhook` |
| Frontend payment URL | `https://rvpay-frontend.onrender.com/checkout` (if applicable) |

- [ ] All URLs are publicly reachable (test with `curl -v` or browser)

---

## 4. Configure HighLevel OAuth Redirect URL

In the HighLevel Marketplace app configuration:

| Field | Value |
| --- | --- |
| **Redirect URL** | `https://<clients-host>/oauth/callback` |
| **Webhook URL** | `https://<clients-host>/webhooks/highlevel` |
| **Webhook Verification** | `X-GHL-Signature` / Ed25519 |

- [ ] Redirect URL is set to the deployed Clients OAuth callback
- [ ] Webhook URL is set to the deployed Clients webhook endpoint

## 5. Configure Required Marketplace Scopes

The RVPay Marketplace application requires the following scopes for Custom Payment Provider functionality:

| Scope | Purpose |
| --- | --- |
| `payments/custom-provider.readonly` | Read payment provider configuration |
| `payments/custom-provider.write` | Create/update/delete payment provider configuration |
| `payments/orders.readonly` | Read payment order information |
| `payments/orders.write` | Create payment orders |
| `payments/transactions.readonly` | Read payment transaction history |

Subscription-related scopes (`payments/subscriptions.readonly`) are **not** required — RVPay supports one-time payments only.

- [ ] Required scopes are enabled in the HighLevel Marketplace app
- [ ] Subscription scopes are NOT enabled (prevent unnecessary permission requests)

## 6. Configure RVPay Environment Variables

Set the following environment variables in the Render dashboard for the **Clients** service:

| Variable | Value |
| --- | --- |
| `HIGHLEVEL_CLIENT_ID` | Client ID from HighLevel Marketplace app |
| `HIGHLEVEL_CLIENT_SECRET` | Client Secret from HighLevel Marketplace app |
| `HIGHLEVEL_REDIRECT_URI` | `https://<clients-host>/oauth/callback` |
| `HIGHLEVEL_WEBHOOK_PUBLIC_KEY` | PEM-encoded Ed25519 public key (see `.env.example`) |
| `HIGHLEVEL_API_BASE_URL` | `https://services.leadconnectorhq.com` (default) |
| `HIGHLEVEL_PAYMENT_URL` | Frontend checkout URL (e.g. `https://rvpay-frontend.onrender.com/checkout`) |
| `HIGHLEVEL_QUERY_URL` | `https://<clients-host>/payments/custom-provider/query` |
| `HIGHLEVEL_PROVIDER_NAME` | `RVPay` (default) |
| `HIGHLEVEL_PROVIDER_DESCRIPTION` | `RVPay payment provider` (default) |
| `HIGHLEVEL_PROVIDER_IMAGE_URL` | Publicly accessible image URL (optional) |
| `PUBLIC_BASE_URL` | `https://<clients-host>` |
| `TRANSACTIONS_GRPC_ADDR` | gRPC address of Transactions service (e.g. `rvpay-transactions:50051`) |

> **Security:** Never commit the actual values to source code. Use Render's encrypted environment variables.

- [ ] All variables are set in Render dashboard
- [ ] Variables use the correct deployed hostnames

## 7. Install the Marketplace App into a Test Location

- [ ] Navigate to the HighLevel Marketplace app installation page
- [ ] Install the app into a test location (not production)
- [ ] Complete the OAuth authorization flow

## 8. Complete OAuth

1. HighLevel redirects to `https://<clients-host>/oauth/callback?code=...&state=...`
2. The RVPay backend exchanges the code for tokens
3. The integration record is created in the database
4. The provider association and configuration are automatically registered with HighLevel

- [ ] OAuth callback completes successfully (HTTP 302 redirect to success page)
- [ ] Integration record exists in the `integrations` table
- [ ] OAuth token record exists in the `oauth_tokens` table

## 9. Confirm RVPay Obtains the Location Identity

Check the database:

```sql
SELECT id, location_id, provider FROM integrations WHERE location_id IS NOT NULL;
```

- [ ] `location_id` is populated with the HighLevel location ID
- [ ] `provider` is set to `highlevel`

## 10. Confirm RVPay Calls HighLevel Provider Association API

Check the RVPay logs (Render dashboard or `journalctl`):

```
level=info msg="creating provider association" provider_api_url=https://services.leadconnectorhq.com/payments/custom-provider/provider location_id=...
level=info msg="provider association created" location_id=...
```

- [ ] Logs show provider association API call
- [ ] No errors or 4xx/5xx responses from HighLevel

## 11. Confirm RVPay Calls HighLevel Provider Configuration API

Check the logs:

```
level=info msg="creating provider configuration" provider_api_url=https://services.leadconnectorhq.com/payments/custom-provider/connect location_id=...
level=info msg="provider configuration created" location_id=...
```

- [ ] Logs show provider configuration API call
- [ ] No errors or 4xx/5xx responses from HighLevel

## 12. Confirm HighLevel Recognizes the Provider

In the HighLevel Marketplace dashboard (or via HighLevel API), verify the provider is listed for the test location.

- [ ] Provider appears in HighLevel as an available payment provider
- [ ] Provider name matches `HIGHLEVEL_PROVIDER_NAME`
- [ ] Provider description matches `HIGHLEVEL_PROVIDER_DESCRIPTION`

## 13. Confirm queryUrl is Publicly Reachable

```bash
curl -v -X POST https://<clients-host>/payments/custom-provider/query \
  -H "Content-Type: application/json" \
  -d '{"type":"verify","transactionId":"test","apiKey":"test","chargeId":"test","subscriptionId":""}'
```

Expected response codes:
- **200** — Valid request (may return `{"success":false}` for unknown transaction — that is OK)
- **405** — Wrong HTTP method (GET, PUT, DELETE etc.)
- **400** — Malformed request body

- [ ] POST returns 200 (with valid or invalid transaction)
- [ ] GET returns 405 Method Not Allowed
- [ ] PUT/DELETE returns 405 Method Not Allowed

## 14. Confirm Payment-Provider Webhook URL is Publicly Reachable

```bash
curl -v -X POST https://<clients-host>/payments/custom-provider/webhook \
  -H "Content-Type: application/json" \
  -d '{"type":"payment.captured","chargeId":"test","transactionId":"test","locationId":"test","apiKey":"test"}'
```

Expected:
- **200** — Valid request acknowledged (will not process without valid credentials — that is OK)

- [ ] POST returns 200
- [ ] GET returns 405 Method Not Allowed

## 15. Confirm the Frontend Payment URL is Correctly Configured

Check the `payment_provider_configs` table:

```sql
SELECT integration_id, payments_url, query_url FROM payment_provider_configs;
```

- [ ] `payments_url` matches the configured `HIGHLEVEL_PAYMENT_URL`
- [ ] `query_url` matches the configured `HIGHLEVEL_QUERY_URL`

## 16. Perform a Test Payment

Trigger a payment through the HighLevel checkout flow (requires a frontend or HighLevel test trigger).

- [ ] Payment initiation reaches the Transactions service
- [ ] Transaction record is created in the database
- [ ] pawaPay deposit is created (or mocked in test environment)

## 17. Verify Transaction State

```sql
SELECT id, status, amount, currency, ghl_transaction_id, ghl_charge_id
FROM transactions.deposits
WHERE ghl_transaction_id IS NOT NULL;
```

- [ ] Transaction status matches expected payment outcome
- [ ] `ghl_transaction_id` and `ghl_charge_id` are populated

## 18. Verify HighLevel Verification

Send a verification request to the query endpoint with the actual transaction ID:

```bash
curl -X POST https://<clients-host>/payments/custom-provider/query \
  -H "Content-Type: application/json" \
  -d '{"type":"verify","transactionId":"<ghl-transaction-id>","apiKey":"<provider-api-key>","chargeId":"<ghl-charge-id>","subscriptionId":""}'
```

Expected responses:
- `{"success":true}` — Transaction completed
- `{"failed":true}` — Transaction failed
- `{"success":false}` — Transaction pending

- [ ] Response matches the actual transaction state
- [ ] Unknown transaction IDs return `{"success":false}`

## 19. Verify payment.captured Handling

Send a `payment.captured` event to the payment-provider webhook:

```bash
curl -X POST https://<clients-host>/payments/custom-provider/webhook \
  -H "Content-Type: application/json" \
  -d '{"type":"payment.captured","chargeId":"<charge-id>","transactionId":"<ghl-transaction-id>","locationId":"<location-id>","apiKey":"<provider-api-key>"}'
```

- [ ] Returns HTTP 200
- [ ] Event is recorded in `webhook_events` table
- [ ] Transaction state is updated correctly (if valid)

## 20. Verify Duplicate Webhook Behavior

Send the same `payment.captured` event again:

- [ ] Returns HTTP 200 (duplicate acknowledged, not rejected)
- [ ] Only one transaction state change occurred
- [ ] `webhook_events` table contains exactly one record (or duplicates are idempotently handled)

---

## Failure Diagnosis

### Provider Association Fails

Check:
- `HIGHLEVEL_API_BASE_URL` is correct (default: `https://services.leadconnectorhq.com`)
- Access token is valid (not expired)
- OAuth scopes include `payments/custom-provider.write`
- Location ID is present in the integration record

### Provider Configuration Fails

Check:
- `HIGHLEVEL_QUERY_URL` and `HIGHLEVEL_PAYMENT_URL` are publicly reachable
- `HIGHLEVEL_PROVIDER_NAME` and `HIGHLEVEL_PROVIDER_DESCRIPTION` are set
- Access token has the required scopes

### Query Endpoint Returns Unexpected Response

Check:
- `type` parameter is `verify` (other types are unsupported)
- `transactionId` matches a known HighLevel transaction
- Provider API key is valid (stored in `payment_provider_configs`)
- Transactions service is reachable via gRPC (`TRANSACTIONS_GRPC_ADDR`)

### Webhook Processing Fails

Check:
- Event type is `payment.captured` (subscription events are safely ignored)
- `locationId` matches an installed integration
- Provider API key matches the stored key for that integration

### Log Redaction

When sharing logs for debugging, redact the following values before posting:

| Pattern | Redaction |
| --- | --- |
| `access_token` | `...redacted...` |
| `refresh_token` | `...redacted...` |
| `client_secret` | `...redacted...` |
| `provider_api_key` | `...redacted...` |
| `apiKey` field | `...redacted...` |

---

## Environment Variable Reference

| Variable | Required | Purpose |
| --- | --- | --- |
| `HIGHLEVEL_CLIENT_ID` | Yes | HighLevel OAuth client ID |
| `HIGHLEVEL_CLIENT_SECRET` | Yes | HighLevel OAuth client secret |
| `HIGHLEVEL_REDIRECT_URI` | Yes | OAuth callback URL |
| `HIGHLEVEL_WEBHOOK_PUBLIC_KEY` | Yes | Ed25519 public key for webhook verification |
| `HIGHLEVEL_API_BASE_URL` | No (has default) | HighLevel API base URL |
| `HIGHLEVEL_PAYMENT_URL` | Yes | Frontend checkout URL |
| `HIGHLEVEL_QUERY_URL` | Yes | Payment query endpoint URL |
| `HIGHLEVEL_PROVIDER_NAME` | No (has default) | Payment provider display name |
| `HIGHLEVEL_PROVIDER_DESCRIPTION` | No (has default) | Payment provider description |
| `HIGHLEVEL_PROVIDER_IMAGE_URL` | No | Payment provider image URL |
| `PUBLIC_BASE_URL` | Yes | Public base URL of the backend |
| `TRANSACTIONS_GRPC_ADDR` | Yes | Transactions service gRPC address |
| `DATABASE_URL` | Yes | PostgreSQL connection string (auto-configured by Render) |