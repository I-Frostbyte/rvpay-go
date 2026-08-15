# Agent 07 — Fix GHL Location ID + Transactions gRPC Address


## Purpose


Make ONLY these two targeted fixes in `rvpay-go`:


1. Fix the HighLevel OAuth flow so the actual HighLevel `locationId` is used when registering the Custom Payment Provider.
2. Fix the Clients → Transactions gRPC address configuration so the Clients service does not fall back to an incorrect localhost address in deployed environments.


Do NOT make any other code, architectural, refactoring, cleanup, naming, documentation, dependency, database, API, or GHL integration changes.


This is a surgical patch intended to be completed quickly and cheaply.


---


# Rules


- Work from the current repository state.
- Read only the minimum files necessary.
- Do NOT read the entire repository.
- Do NOT run broad agents or exploratory analysis.
- Do NOT redesign the OAuth flow.
- Do NOT redesign the Clients service.
- Do NOT modify the Transactions service.
- Do NOT modify payment flows.
- Do NOT modify the GHL provider registration payload except where required to supply the correct `locationId`.
- Do NOT modify frontend code.
- Do NOT modify Render infrastructure beyond the minimum configuration needed for the Transactions gRPC address.
- Do NOT add dependencies unless absolutely required. Prefer zero new dependencies.
- Do NOT change existing behavior unrelated to these two fixes.
- Do NOT perform unrelated cleanup.
- Do NOT update documentation unless required to accurately record the exact configuration change.
- Keep the diff as small as possible.


---


# Stage 1 — Inspect Only


Read ONLY:


1. `agents/project-context.md`
2. `README.md`
3. `docs/project-checkpoint.md`
4. `clients/providers/highlevel.go`
5. `clients/oauth/service.go`
6. `clients/cmd/grpc-service/main.go`
7. `render.yaml`


Determine:


### A. Location ID


Find:


- HighLevel OAuth token response struct.
- How the OAuth callback currently obtains the HighLevel user ID.
- How `RegisterProvider()` receives its `locationID`.
- Where the incorrect user ID is currently being passed as the location ID.


### B. gRPC


Find:


- `TRANSACTIONS_GRPC_ADDR`.
- Its current default value.
- How `main.go` creates the Transactions gRPC client.
- How `render.yaml` currently configures the Clients service.


Do NOT modify anything during this stage.


---


# Stage 2 — Fix ONLY the HighLevel Location ID


Modify the existing HighLevel OAuth token response handling so that the actual HighLevel OAuth response `locationId` is captured.


The token response should expose the real HighLevel location ID, for example:


```go
LocationID string `json:"locationId"`
```


Preserve all existing token fields and behavior.


Then modify the OAuth callback flow so that the provider registration receives:


```go
tokenResp.LocationID
```


as the locationID.


Do NOT use:


```go
userInfo.ID
providerUserID
user ID
```


as the provider locationID.


The existing user information retrieval may remain if it is still required elsewhere. Do not remove it unless it becomes demonstrably unnecessary for this exact fix.


The resulting flow must be:


```
HighLevel OAuth
      ↓
token response
      ↓
locationId
      ↓
RegisterProvider(... locationId ...)
```


Do not change the provider registration API, payload structure, provider name, URLs, or payment behavior.


---


# Stage 3 — Fix ONLY the Transactions gRPC Address


Inspect the existing:


```
TRANSACTIONS_GRPC_ADDR
```


handling.


The Clients service must NOT use an inappropriate hard-coded localhost address when deployed to Render.


The existing configuration should remain environment-variable driven.


Use the Render service-to-service/internal address mechanism already appropriate for this project.


Do NOT invent a public URL.


Do NOT change the Transactions service.


Do NOT change the gRPC port unless the existing project configuration proves that the port is wrong.


The objective is simply:


```
rvpay-clients
      ↓
TRANSACTIONS_GRPC_ADDR
      ↓
rvpay-transactions internal Render address
```


If render.yaml can provide the correct service-to-service value through the existing Render configuration, make only that minimal change.


If the correct internal hostname cannot safely be hard-coded because Render supplies it dynamically, preserve the environment-variable approach and make the smallest necessary change to prevent the deployed Clients service from falling back to localhost.


Do not redesign service discovery.


---


# Stage 4 — Minimal Verification


Run only targeted verification.


At minimum:


```sh
go test ./clients/...
```


If that is unnecessarily broad or expensive, run the smallest relevant package tests first.


Also verify:


```sh
go test ./clients/providers/...
go test ./clients/oauth/...
```


if those packages exist independently.


Check that:


- The project compiles.
- The HighLevel token response contains locationId.
- RegisterProvider() receives the actual OAuth locationId.
- No code path passes the HighLevel user ID as the provider location ID.
- TRANSACTIONS_GRPC_ADDR remains environment-driven.
- The Clients service no longer incorrectly relies on localhost:50052 in the deployed configuration.


Do NOT run the entire repository test suite unless necessary.


---


# Stage 5 — Diff Audit


Before finishing, inspect:


```sh
git diff --stat
git diff
```


The final diff must contain ONLY changes related to:


- HighLevel OAuth locationId.
- Clients → Transactions gRPC address configuration.


If unrelated changes are present, revert them.


Do NOT commit.


Do NOT push.


Do NOT modify unrelated files.


---


## Completion Criteria


The task is complete only when all of the following are true:


- HighLevel OAuth locationId is captured.
- RegisterProvider() receives the actual HighLevel locationId.
- HighLevel user ID is no longer incorrectly used as locationId.
- Transactions gRPC address is correctly environment-driven for Render.
- Clients does not incorrectly fall back to localhost for the deployed Transactions service.
- Relevant tests/build pass.
- Final git diff contains only these two fixes.
- No commit or push was performed.


## Final Response


Report ONLY:


- What changed for locationId.
- What changed for TRANSACTIONS_GRPC_ADDR.
- Tests run and their result.
- Files changed.
- Confirmation that no unrelated changes were made.


Do not suggest additional improvements.


Do not continue into GHL configuration.


Do not modify anything else.


Stop immediately after reporting completion.