# Platform Performance Review

Document Version: 1.0
Status: Complete
System: RVPay
Review: Platform Agent 11 — Performance

## 1. Objective

Audit and improve the performance characteristics of the RVPay platform.
The objective is to identify and address concrete performance problems across:

- HTTP gateway
- gRPC services
- database access
- SQL queries
- connection pools
- external provider calls
- serialization
- application startup
- container/runtime configuration
- observability overhead

Performance work followed the philosophy: what is slow, where, why, how was it
determined, what is the smallest appropriate improvement, and how is it
verified. No speculative optimizations, no architectural rewrites, and no
performance of unrelated behavior. Every change is concrete, justified,
minimal, and preserves existing behavior.

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
| docs/platform-security-review.md | ✅ |

All required documents were present and read.

## 3. Performance Baseline

No quantitative baseline measurements existed before this agent (no benchmarks,
no profiling data, no latency/throughput numbers in the repository). This is
stated explicitly: **no benchmark results were invented**.

Known characteristics from the platform reviews:

- **Database**: pgxpool with eager ping per service; sqlc-generated parameterized
  queries; sensible indexes on ownership/status/reference/created_at columns
  (audit P-10). Clients list queries are paginated (`LIMIT/OFFSET`); Transactions
  `ListMerchants` returned the full result set (unbounded).
- **HTTP**: per-service embedded grpc-gateway; `PORT`/`LISTEN_PORT` env-driven;
  `/healthz` is cheap (no DB dependency).
- **gRPC**: per-service server; recovery + observability interceptors log
  metadata only (no payloads); reflection enabled.
- **Providers**: HighLevel HTTP calls had a 10s timeout but a new `http.Client`
  was constructed for every token exchange / refresh / user-info call (no
  connection reuse).
- **Observability**: request-ID correlation; metadata-only access/RPC logs;
  `/healthz` probes logged at DEBUG; no metrics/tracing (no documented
  requirement).

## 4. Architecture Reviewed

- **HTTP gateway**: grpc-gateway per service, embedded in `main.go`; route
  registration static at startup; no per-request router construction; request
  bodies not globally buffered (grpc-gateway streaming default); `/healthz`
  handler is constant-time.
- **gRPC**: per-service server initialized once at startup (recovery +
  observability interceptors chained); no new connections per request (in-process
  gateway); reflection disabled-by-default is not in place (documented in
  Remaining Risks).
- **Services**: Clients and Transactions follow the Deposits template; service
  methods validate input and map repository errors; no triple-logging.
- **PostgreSQL**: one database per service; pgxpool via `shared/database`;
  sqlc-generated queries; migrations owned per service.
- **Providers**: HighLevel OAuth/token/userinfo calls (fixed HTTPS endpoints,
  10s timeouts), webhook signature verification (local HMAC, no network).
- **Render**: three web services (deposits, clients, transactions), one managed
  PostgreSQL each, `/healthz` health checks, Docker runtime.
- **Containers**: multi-stage distroless static images; non-root; direct
  ENTRYPOINT; no shell; migrations packaged in image.

## 5. Database Review

- **Queries**: all client-service list queries are paginated with explicit
  columns and `LIMIT/OFFSET`; no `SELECT *` except `RETURNING *` on inserts
  (required to return the inserted row) and full-row reads in Transactions
  where the full model is consumed. No N+1 queries observed: list operations
  query a single table; existence checks use `EXISTS`; counts use `COUNT(1)`.
- **Round trips**: each service RPC performs one repository call (single
  query); the newly paginated `ListMerchants` performs one page query + one
  count query instead of returning the entire table.
- **Indexes**: existing migrations provide indexes on ownership columns
  (`client_id`, `merchant_id`), `external_reference`, `idempotency_key`,
  `status`, and `created_at` (verified in the repository audit). No new indexes
  were required by the query patterns reviewed.
- **Connection pool**: `shared/database.Connect` uses `pgxpool.New` with pgx
  defaults (max conns based on CPU, no explicit churn reduction). Pool is
  created once per service process (directive #72). No concrete pool defect was
  identified; no values changed blindly (directive #35/#36).
- **Transactions**: no `BEGIN … external HTTP … COMMIT` patterns exist in the
  new services (deposits/payouts initialize status in a single insert); no
  connection held during external calls (directive #38/#39).
- **Row handling**: pgx rows are managed by sqlc-generated methods (closed/
  released by the driver); no leaked rows observed (directive #37/#40).

## 6. HTTP Review

- **Clients**: the only outbound HTTP clients are in `clients/providers`
  (HighLevel). **PERF-01 fixed**: a new `http.Client` was previously created
  inside `ExchangeCode`, `RefreshToken`, and `GetUserInfo` on every call,
  defeating HTTP connection pooling. The provider now holds one shared
  `http.Client{Timeout: 10s}` initialized once in the constructor (directive
  #8, #9).
- **Timeouts**: 10s timeout preserved (not arbitrarily reduced — directive #9).
- **Response bodies**: every provider response body is closed (`defer
  resp.Body.Close()`) so connections are returned to the pool (directive #10).
- **Request allocation**: no request-allocation optimization performed
  (directive #11) — no profiling evidence justifies it.
- **Payload handling**: provider responses are read fully (`io.ReadAll`) for
  JSON parsing; payload sizes are small (token/user-info responses); no
  buffering change warranted (directive #56).

## 7. gRPC Review

- **Connections**: services use in-process gateway registration (no outbound
  gRPC dialing per request); no cross-service gRPC calls are wired yet
  (directive #17).
- **Interceptors**: recovery + observability unary interceptors log metadata
  only (request ID, method, status, duration); no payload serialization, no
  secrets, no expensive per-request work (directive #18). Security/recovery
  interceptors were not removed (directive #122).
- **Serialization**: generated protobuf marshal/unmarshal is used at the
  transport boundary; no unnecessary marshal→unmarshal→marshal cycles exist
  in service code (directive #20).
- **Request size**: protobuf messages are small (IDs, enums, money strings);
  no oversized-message handling concerns.

## 8. Provider Review

- **Unnecessary calls**: OAuth token exchange performs one code exchange + one
  user-info call; no repeated authentication calls per operation; token refresh
  only on demand (directive #12).
- **Retries**: no retry loops exist in the new services (directive #14/#15);
  provider failures return errors without automatic retry, avoiding retry storms.
- **Connection reuse**: **PERF-01** (above) restores HTTP connection pooling for
  repeated HighLevel calls (directive #8, #13).
- **Rate limits**: no provider parallelism was introduced (directive #82/#83);
  HighLevel calls remain sequential and bounded.

## 9. Concurrency Review

- **Goroutines**: the only goroutines are the per-service HTTP/gRPC servers and
  the shutdown watcher (signal-driven); all terminate on shutdown (directive
  #48).
- **Cancellation**: all request-bound work uses `context.Context`; provider
  HTTP calls use `http.NewRequestWithContext`, so cancelled requests abort
  provider work (directive #49/#50).
- **Bounded concurrency**: no unbounded goroutine creation based on requests or
  records (directive #52); no parallel work was introduced — operations are not
  independent in the reviewed paths (directive #46/#47).
- **Race safety**: no concurrency code was changed; no new shared mutable state.

## 10. Memory Review

- No large in-memory datasets are loaded: the previously unbounded
  `ListMerchants` was the only query that could load an entire table; it is now
  bounded by pagination (directive #102).
- Provider responses are small token/user-info payloads; no large byte-slice
  duplication concerns (directive #54).
- No GC/runtime tuning was performed (directive #103); no data structures were
  replaced for style (directive #106).

## 11. Observability Review

- **Logging**: zerolog structured JSON; access/RPC logs emit metadata only
  (request ID, method, path, status, duration); `/healthz` at DEBUG; no request
  bodies or provider responses logged (directive #59). No expensive
  serialization is performed purely for logs.
- **Sensitive logging preserved**: security controls from Agent 10 remain
  intact — no Authorization headers, tokens, or credentials are logged
  (directive #60/#122).
- **Metrics/tracing**: not implemented (no architecture requirement); no
  high-cardinality labels exist (directive #64/#65).
- **Health checks**: `/healthz` is cheap (constant-time, no DB/provider call);
  liveness/readiness requirements satisfied (directive #66–#68).

## 12. Container/Runtime Review

- **Startup**: services initialize config once, connect the pool once, run
  migrations once, and build server/routers once at startup (directive #69–#74);
  no per-request construction. The provider `http.Client` is now also initialized
  once (PERF-01).
- **Docker**: multi-stage distroless images, no shell-script entrypoints, no
  added startup work (directive #75); image contains only the binary and
  migrations.
- **Render**: instances are not resized (directive #77); code is compatible with
  multiple instances (no process-local mutable maps/sessions/queues — directive
  #78/#79); PostgreSQL remains the single source of truth (directive #81).
- **Resources**: no assumption of unlimited CPU/memory/connections (directive
  #76); the pagination cap (100) bounds per-request database result size.

## 13. Changes Implemented

| ID | Change | File(s) |
| --- | --- | --- |
| PERF-01 | Reuse a single shared `http.Client` (10s timeout) across all HighLevel provider calls instead of creating a new client per request (connection reuse) | `clients/providers/highlevel.go` |
| PERF-02 | Honor the existing `ListMerchants` pagination contract: add `LIMIT/OFFSET` and `CountMerchants` to the SQL, regenerate sqlc, extend `MerchantRepo` with `List(ctx, limit, offset)` + `Count(ctx)`, and apply a default page size of 20 with a max cap of 100 in the service | `transactions/db/query/merchants.sql`, `transactions/db/sqlc/*` (generated), `transactions/db/repo/merchant_repo.go`, `transactions/merchants/service.go`, `transactions/db/repo/mocks/repo.go` (generated) |
| PERF-03 | Added regression tests for the pagination behavior (default limit, max cap, next-page token) | `transactions/merchants/service_test.go` |

## 14. Measurements

| Area | Before | After | Change | Measurement Method |
| ---- | -----: | ----: | -----: | ------------------ |
| HighLevel HTTP connection reuse | New `http.Client` per call (no pooling) | Single shared client (pooled connections) | Avoids repeated TCP/TLS handshakes on consecutive provider calls | Code review (directive: no benchmark invented; not measured) |
| ListMerchants result set | Unbounded (full table returned) | Bounded to ≤100 rows per page with default 20 | Bounds memory/DB result size per request regardless of table growth | Code review |
| ListMerchants round trips | 1 (full table) | 2 (page + count) | +1 cheap `COUNT(*)` query in exchange for bounded result sets | Code review |
| Startup | Config/pool/migration/server initialized once | Provider HTTP client also initialized once | Removes per-call client construction | Code review |

No latency, throughput, memory, CPU, or query-duration benchmarks were run;
no numbers are claimed (directives #93, #128).

## 15. Findings

| ID | Severity | Area | Finding | Resolution |
| --- | --- | --- | --- | --- |
| PERF-01 | MEDIUM | `clients/providers/highlevel.go` | A new `http.Client` was created on every `ExchangeCode`, `RefreshToken`, and `GetUserInfo` call, preventing HTTP connection pooling and TLS session reuse | ✅ Provider now holds one shared `http.Client{Timeout: 10s}` initialized once in the constructor |
| PERF-02 | MEDIUM | `transactions/merchants` (SQL/repo/service) | `ListMerchants` accepted a pagination request (`commongrpc.PaginationRequest`) but ignored it and returned the full merchants table with an always-empty next-page token; unbounded over time | ✅ Added `LIMIT/OFFSET` + `CountMerchants`; `MerchantRepo.List(ctx, limit, offset)` + `Count(ctx)`; service honors page size with default 20 / max 100 and returns a real next-page token |
| PERF-03 | INFO | Transactions queries | `ListDepositsBy*`, `ListPayoutsBy*`, `ListCustomersBy*` use `SELECT *` and have no `LIMIT/OFFSET` | Not reachable through any public RPC (only Create/Get RPCs exist for those entities); left unchanged to avoid regenerating/rewriting queries without a concrete consumer. Documented as future work |
| PERF-04 | INFO | HTTP/gRPC/gateway | Per-service embedded architecture reuses in-process handlers; no per-request connection construction; provider responses closed; `defer resp.Body.Close()` present in all provider calls | ✅ Verified and preserved |
| PERF-05 | INFO | Health/observability | `/healthz` constant-time; observability metadata-only with DEBUG probe logging; no high-cardinality metrics (none implemented) | ✅ Verified and preserved |
| PERF-06 | INFO | Database pool | pgxpool uses pgx defaults; pool created once per process; no leak or exhaustion pattern observed | ✅ Preserved; documented (no blind tuning per directive #35) |
| PERF-07 | INFO | Startup/runtime | Config, pool, migrations, HTTP/gRPC servers, provider client all initialized once at startup; no per-request construction | ✅ Verified and preserved |

## 16. Remaining Performance Risks

- **Unpaginated Transactions list queries** (`ListDepositsByClient`, `ListPayoutsByClient`, etc.) remain unpaginated at the SQL layer. They are not reachable through public RPCs today (no list RPCs exist for deposits/payouts/customers in the protobuf contract), so they are latent. When list RPCs are added, pagination must follow the `ListMerchants` pattern.
- **No index on webhook/event dedup**: webhook duplicate detection is stubbed (no `webhook_events` table); when implemented, the provider event ID needs a unique index.
- **Connection pool sizing**: pgx defaults are used; under sustained concurrency on Render, pool sizing may need tuning informed by real load. No evidence of a problem exists now.
- **gRPC reflection** remains enabled (documented in the security review); it has negligible runtime cost but increases API-surface discovery.
- **No load testing or profiling** has been performed; the repository has no documented load test. The changes here are code-level and safe, but end-to-end latency under load is unmeasured.

## 17. Future Recommendations

Out of scope for this agent; listed for future decisions (directive #43/#44/#119/#120):

- **In-memory or distributed caching** (e.g., read-through for platform/config lookups) — only if profiling demonstrates a hot path; requires invalidation and consistency design. Explicitly not added here.
- **Queue-based background processing** for webhook event dispatch and provider status reconciliation — decouples bursty processing from request latency; requires a durable queue (not introduced here per directive #119).
- **Read replicas / reporting queries** if monitoring/analytics workloads grow — Transactional writes must remain on the primary.
- **Horizontal scaling**: services already hold no process-local mutable state, so multiple instances are safe; Render instance sizing should be driven by measured load, not assumed.
- **Load/benchmark harness** for deposits/payouts once provider execution is wired (F-01/F-02 integration work).

## 18. Tests and Verification

Tests added (directive #108/#109/#110):

| Test | File | Verifies |
| --- | --- | --- |
| `TestListMerchants` (updated) | `transactions/merchants/service_test.go` | Default page size (20) + offset 0 used when no page request; count returned; next-page token empty when no more pages |
| `TestListMerchantsHonorsPageSizeAndCap` (new) | `transactions/merchants/service_test.go` | Requested page size capped to max (100); total count returned; next-page token set when more pages exist |
| `TestListMerchantsRepositoryError` (updated) | `transactions/merchants/service_test.go` | List error mapped to `Internal` |

Commands executed:

```bash
cd transactions/db && go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate   # SQLC_OK
cd transactions/db && go run go.uber.org/mock/mockgen@v0.6.0 -destination repo/mocks/repo.go -package mocks ./repo TransactionsRepo,MerchantRepo,CustomerRepo,DepositRepo,PayoutRepo   # MOCKS_OK
go build ./clients/... ./transactions/...                # OK
go test ./clients/providers/... ./transactions/merchants/... ./transactions/db/repo/...   # OK
go test ./...                                           # OK (full suite)
go vet ./...                                            # OK
gofmt -l (hand-written Go)                              # flags are the documented pre-existing local
                                                        # core.autocrlf=true CRLF artifact (CICD-07/P-GEN-04);
                                                        # CI is LF-only and passes. No real drift introduced.
```

Race testing: no concurrency code was changed, so no race-enabled run was
required (directive #108/#132).

## 19. Documentation Changes

- `docs/platform-performance-review.md` — created (this document).
- `docs/project-checkpoint.md` — updated: Platform 11 Performance marked COMPLETE.

No changes to README.md were required (no performance-related configuration
change made existing instructions incorrect; directive #124).

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
| docs/platform-security-review.md | ✅ |
| docs/platform-performance-review.md | ✅ (this document) |

## 21. Final Status

PASS WITH FOLLOW-UP