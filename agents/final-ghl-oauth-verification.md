# Final GHL OAuth Verification — Do Not Expand Scope

## Objective

Perform one final, narrowly scoped verification/fix of the HighLevel OAuth implementation in the current `fix-oauth-endpoints` branch.

The previous task changed the HighLevel OAuth endpoints in:

`clients/providers/highlevel.go`

The current commit is:

`3df77c2 — fix(clients): conform HighLevel OAuth to current GHL Marketplace endpoints`

The purpose of THIS task is only to ensure that the OAuth implementation exactly matches the current HighLevel Marketplace OAuth flow for our specific use case:

* RVPay is a HighLevel Marketplace app.
* The app targets **Sub-accounts / Locations**.
* A **GHL Sub-account user** installs RVPay.
* A GHL Sub-account corresponds to an RVPay **Client**.
* The GHL `locationId` is therefore the external account identifier used to resolve the RVPay Client.
* The existing stateless callback / first-install mapping implementation is already considered complete.
* Render deployment is already working.
* Do NOT redesign or refactor the application.

## Required Documentation

Before making ANY change, inspect the current official HighLevel documentation for:

1. OAuth authorization for a Sub-account-targeted Marketplace app.
2. Authorization-code token exchange for a Sub-account / Location installation.
3. The current HighLevel OAuth token endpoint.
4. The current HighLevel OAuth authorization/location-selection endpoint.
5. The current HighLevel OAuth user-info endpoint.

Use the official HighLevel Marketplace documentation as the source of truth.

Do NOT rely on old blog posts, Stack Overflow, cached examples, or assumptions.

## Specifically Verify

Inspect the actual implementation in:

* `clients/providers/highlevel.go`
* `clients/oauth/service.go`
* `clients/http/oauth_handler.go`

Also inspect the relevant tests.

Verify the COMPLETE OAuth request/response contract.

### Authorization endpoint

Verify that the authorization URL currently configured is correct for the current HighLevel Marketplace Sub-account installation flow.

Current value:

`https://marketplace.gohighlevel.com/oauth/chooselocation`

Do not change it unless the current official HighLevel documentation demonstrates that it is incorrect for this exact installation flow.

### Token endpoint

Verify that the token endpoint is:

`https://services.leadconnectorhq.com/oauth/token`

Do not revert to `api.highlevel.com`.

### Authorization-code exchange

Verify the exact form body sent by `ExchangeCode`.

For the specific flow we are implementing, verify whether the request requires:

* `grant_type=authorization_code`
* `client_id`
* `client_secret`
* `code`
* `redirect_uri`
* `user_type=Location`

The critical requirement is that the implementation must support the documented Sub-account/Location installation flow.

If the official current HighLevel documentation requires:

`user_type=Location`

then add it.

Do NOT merely claim the existing request is correct.

Inspect the actual generated HTTP request in the code/tests.

### Token response

Verify that the implementation correctly extracts:

* `access_token`
* `refresh_token`
* `expires_in`
* `token_type`
* `scope`
* `locationId`

The existing RVPay architecture depends on `locationId`.

Do not redesign how `locationId` is subsequently resolved.

### User-info endpoint

Verify the current endpoint:

`https://services.leadconnectorhq.com/oauth/userinfo`

Only change it if the official HighLevel documentation says otherwise.

## Testing Requirement

The existing tests use a mock OAuth server.

That is useful, but it is NOT a real end-to-end HighLevel test.

Therefore:

1. Ensure the tests verify the ACTUAL form fields sent by `ExchangeCode`.
2. If `user_type=Location` is required, add/update a focused test assertion proving it is actually sent.
3. Verify that the response containing `locationId` is correctly parsed.
4. Run:

```bash
go build ./clients/...
go test ./clients/...
go vet ./clients/...
gofmt -w clients/providers/highlevel.go
```

All must pass.

## ABSOLUTE SCOPE RESTRICTIONS

Do NOT modify:

* Clients/Customers domain model
* Client repository behavior
* Customer logic
* Transactions service
* Transactions API
* database schema
* Render configuration
* Render architecture
* Docker configuration
* deployment configuration
* GHL payment-provider registration logic
* first-install mapping logic
* stateless callback architecture
* OAuth state architecture
* webhook handling
* unrelated tests
* unrelated files

Do NOT refactor existing code.

Do NOT rename existing functions or types.

Do NOT introduce new abstractions unless absolutely required to implement the documented OAuth request.

Do NOT "improve" anything outside this OAuth verification.

## Important

If the current implementation is already fully compliant with the official HighLevel documentation, DO NOT make unnecessary changes.

Instead, report exactly why it is compliant and stop.

If a discrepancy exists, make ONLY the minimal change required to correct it.

## Git Requirements

After the verification/fix:

1. Show exactly which files changed.
2. Show the exact OAuth request fields that are now being sent.
3. Show the relevant test proving the request is correct.
4. Run all required commands above.
5. Commit the work with:

`fix(clients): finalize GHL marketplace oauth flow`

6. Push the branch:

```bash
git push origin fix-oauth-endpoints
```

Do NOT merge into `main`.

## Final Report

The final report must contain ONLY:

### Changes Made

List the files and exact changes.

### GHL Contract Verified

State the exact current official endpoints and required token-exchange fields.

### Tests

Give the results of:

* `go build ./clients/...`
* `go test ./clients/...`
* `go vet ./clients/...`

### Commit

Give the commit SHA.

### Scope Confirmation

State explicitly:

> No Clients/Customers, Transactions, database, Render, first-install mapping, or payment-provider architecture was changed.

If everything passes, STOP.

Do not propose another task.

Do not make any additional changes.
