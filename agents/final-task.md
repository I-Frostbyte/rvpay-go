# FINAL TASK — Migrate RVPay HighLevel OAuth to Current v3 Contract

## READ THIS FIRST

This is the FINAL coding task for the current GHL OAuth installation problem.

Do not expand the scope.

Do not redesign RVPay.

Do not modify unrelated architecture.

Do not create another OAuth abstraction.

Do not touch Render, PostgreSQL, Transactions, Clients/Customers domain behavior, first-install mapping, webhook architecture, or payment-provider registration.

The only objective is to make RVPay's HighLevel OAuth implementation conform to the CURRENT HighLevel OAuth v3 `/oauth/token` contract and ensure every affected caller, test, mock, request/response structure, and OAuth path is consistent with that contract.

After this task, STOP.

---

# Repository / Current State

Repository:

`github.com/I-Frostbyte/rvpay-go`

Current deployed commit:

`32adb76fb4b971a45aaba54f37a98b5224e247b7`

The application is already successfully receiving the HighLevel Marketplace OAuth callback.

The current live error is:

```text
token exchange failed:
{"error":"invalid_grant","error_description":"Invalid grant: authorization code is invalid"}
```

The current GHL installation is:

* HighLevel Marketplace app
* Target user: Sub-account / Location
* Agency and Sub-account users can install
* Installation is initiated by clicking Install App from the GHL Sub-account Marketplace
* Redirect URLs have already been verified and match
* Render is running the recent `main` deployment

The previous task correctly changed the OAuth hosts and added `user_type=Location`, but the current HighLevel v3 API has a newer token contract.

---

# REQUIRED OFFICIAL DOCUMENTATION

Before changing code, read the CURRENT official HighLevel documentation.

Use ONLY official HighLevel documentation as the authority.

At minimum inspect:

* Current Get Access Token documentation
* Current HighLevel OAuth changelog
* Current Marketplace/Sub-account OAuth documentation

The relevant current contract is:

```text
POST https://services.leadconnectorhq.com/oauth/token
```

Required header:

```text
Version: v3
```

Current request property names include:

```text
clientId
clientSecret
grantType
code
redirectUri
userType
```

The old snake_case properties are no longer the current v3 contract:

```text
client_id
client_secret
grant_type
redirect_uri
refresh_token
user_type
```

The current response uses camelCase fields including:

```text
accessToken
refreshToken
expiresIn
tokenType
```

Do not rely on the previous Cline report saying the old request body was already correct. Verify the actual current specification yourself.

---

# IMPORTANT: TRACE THE WHOLE IMPACT

Do NOT only edit:

`clients/providers/highlevel.go`

First search the entire repository for every HighLevel OAuth-related field and operation.

Search for at least:

```text
client_id
client_secret
grant_type
redirect_uri
refresh_token
user_type
access_token
expires_in
token_type
clientId
clientSecret
grantType
redirectUri
refreshToken
userType
accessToken
expiresIn
tokenType
ExchangeCode
RefreshToken
oauth/token
HighLevel
highlevel
```

Determine which occurrences are:

1. HighLevel OAuth implementation
2. HighLevel OAuth tests/mocks
3. Shared/internal OAuth abstractions
4. Configuration
5. Documentation/examples
6. Unrelated APIs that must NOT be changed

Only change occurrences that are actually part of RVPay's HighLevel OAuth contract.

Do NOT globally replace strings.

---

# 1. Authorization Flow

Inspect the existing HighLevel authorization URL generation.

Verify that the current Marketplace installation flow remains correct.

Do NOT change the authorization endpoint unless the current official HighLevel documentation requires it.

Current expected endpoint:

```text
https://marketplace.gohighlevel.com/oauth/chooselocation
```

Do not modify the existing first-install/state architecture.

---

# 2. Authorization-Code Token Exchange

Inspect `ExchangeCode` and every function it calls.

The actual HTTP request sent to:

```text
https://services.leadconnectorhq.com/oauth/token
```

must conform to the current v3 contract.

Verify the actual generated request, not just the source code.

Required:

### HTTP method

```text
POST
```

### Headers

At minimum:

```text
Content-Type: application/x-www-form-urlencoded
Accept: application/json
Version: v3
```

Use JSON instead only if the existing implementation requires it; do not introduce unnecessary transport changes.

### Authorization-code request fields

Use the current v3 property names:

```text
clientId
clientSecret
grantType=authorization_code
code
redirectUri
userType=Location
```

Do NOT send the obsolete snake_case equivalents.

Specifically:

```text
client_id       ❌
client_secret   ❌
grant_type      ❌
redirect_uri    ❌
user_type       ❌
```

must become:

```text
clientId        ✅
clientSecret    ✅
grantType       ✅
redirectUri     ✅
userType        ✅
```

The authorization code itself remains:

```text
code
```

---

# 3. Token Response

Inspect the response structure used by RVPay.

The current v3 response uses camelCase fields.

Verify that RVPay correctly parses:

```text
accessToken
refreshToken
expiresIn
tokenType
scope
userType
locationId
companyId
userId
```

where applicable.

Most importantly:

```text
locationId
```

must continue to be available to the existing RVPay stateless first-install flow.

DO NOT change how `locationId` is subsequently used.

DO NOT change Client resolution.

DO NOT change the integration model.

Only correct the OAuth response parsing if required by the current HighLevel response schema.

---

# 4. Refresh Token Flow

This is REQUIRED.

Do not fix only `ExchangeCode`.

Inspect `RefreshToken` and every caller.

The current HighLevel v3 token endpoint also uses the current request property names for refresh-token grants.

Verify the actual request against the current official documentation.

The request must use the v3 contract, including the appropriate current fields such as:

```text
clientId
clientSecret
grantType=refresh_token
refreshToken
```

and:

```text
Version: v3
```

Do not leave the refresh flow using the obsolete:

```text
client_id
client_secret
grant_type
refresh_token
```

If the current HighLevel documentation specifies additional fields for the refresh flow, follow the documentation.

---

# 5. Tests

This is extremely important.

Do not simply update tests until they pass.

Update the tests so they verify the REAL v3 contract.

Inspect all relevant tests, including but not limited to:

```text
clients/providers/*
clients/oauth/*
```

and any integration/mock OAuth server used by those tests.

The mock token endpoint must inspect the actual incoming request.

For the authorization-code exchange, tests must verify:

```text
Version == v3

clientId
clientSecret
grantType == authorization_code
code
redirectUri
userType == Location
```

and verify that obsolete fields are NOT being sent:

```text
client_id
client_secret
grant_type
redirect_uri
user_type
```

For refresh-token tests, verify the current v3 request fields as well.

---

# 6. Response Tests

Update mock token responses to use the current camelCase response schema where appropriate:

```text
accessToken
refreshToken
expiresIn
tokenType
```

Verify that RVPay still produces the same internal token representation expected by the rest of the application.

Do not change the internal application contract merely because HighLevel changed its external JSON schema.

The adapter/provider should translate:

```text
HighLevel v3 API
        ↓
RVPay internal OAuth representation
```

---

# 7. Verify Every Caller

Search for every call to:

```text
ExchangeCode(...)
RefreshToken(...)
```

and every use of the returned token structure.

Ensure that the v3 migration does not break:

* stateless callback
* stateful callback
* first-install resolution
* existing integration reuse
* token persistence
* token refresh
* user-info handling
* any other existing HighLevel OAuth path

Do NOT redesign these flows.

Only make compatibility changes required because of the HighLevel v3 contract.

---

# 8. User Info

Inspect the existing HighLevel user-info implementation.

Current endpoint:

```text
https://services.leadconnectorhq.com/oauth/userinfo
```

Verify it against the current official HighLevel documentation.

Do not change it unless documentation requires a change.

Also inspect its response structure and make sure the current implementation still parses the actual current response.

Do not modify unrelated user-info behavior.

---

# 9. Do NOT Chase the Current `invalid_grant`

The current live error is:

```text
invalid_grant
Invalid grant: authorization code is invalid
```

Do not add retry logic.

Do not cache authorization codes.

Do not store authorization codes in the database.

Do not reuse authorization codes.

Do not weaken validation.

Do not modify the callback architecture to work around the error.

The purpose of this task is to ensure RVPay sends the authorization code to HighLevel using the exact current v3 OAuth contract.

After that, a fresh real Marketplace installation will be used to determine whether the live `invalid_grant` disappears.

---

# 10. No Changes Outside OAuth Compatibility

DO NOT modify:

* Clients domain model
* Customers domain model
* Transactions service
* transaction APIs
* PostgreSQL schema
* database migrations
* Render configuration
* Docker configuration
* deployment configuration
* webhook architecture
* GHL webhook handling
* first-install mapping logic
* location resolution logic
* payment-provider registration
* gRPC APIs
* service boundaries
* authentication architecture
* unrelated providers
* unrelated APIs

No refactoring.

No cleanup.

No "while we're here" changes.

No dependency upgrades.

No new packages unless absolutely unavoidable for the documented OAuth contract.

---

# 11. Required Verification Commands

After making the minimal changes, run:

```bash
gofmt -w <all modified Go files>
go build ./clients/...
go test ./clients/...
go vet ./clients/...
```

Then, if practical, run the broader repository test suite:

```bash
go test ./...
```

If the full repository test suite has unrelated existing failures, clearly distinguish those from failures caused by this task.

---

# 12. Inspect the Git Diff Carefully

Before committing:

```bash
git diff --stat
git diff
```

Confirm that every changed line is directly related to the HighLevel OAuth v3 migration or its tests.

If you find unrelated changes, revert them.

---

# 13. Required Final Report

The final report MUST contain:

## Files Changed

List every changed file and why it changed.

## Current GHL v3 Contract

Show:

* token endpoint
* required headers
* authorization-code request fields
* refresh-token request fields
* response fields

## Repository Impact

List every caller/test/mock/structure that had to change.

Explicitly state that you searched the repository for both old and new OAuth field names and did not blindly replace unrelated occurrences.

## Tests

Report:

```text
gofmt
go build ./clients/...
go test ./clients/...
go vet ./clients/...
go test ./...
```

with exact results.

## Diff Scope

Confirm that no changes were made to:

* Clients/Customers domain
* Transactions
* database
* Render
* first-install mapping
* webhook architecture
* payment-provider architecture

## Commit

Create exactly ONE commit with:

```text
fix(clients): migrate HighLevel OAuth to v3
```

Push the branch to GitHub.

Do NOT merge into `main`.

## STOP

Once the above is complete:

STOP.

Do not propose another coding task.

Do not perform unrelated improvements.

Do not modify anything else.

The next step after this task is a fresh real-world GHL Marketplace installation test against the deployed service.
