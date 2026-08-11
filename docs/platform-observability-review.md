# Platform Observability Review

Document Version: 1.0
Status: Complete
System: RVPay
Review: Platform Agent 09 — Observability

## 1. Objective

Implement and document the observability foundation for the RVPay platform so
the running system can answer: is a service running; is a request reaching it;
which operation executed; how long it took; did it succeed or fail; what error
occurred; which request does the error belong to; are database/provider
operations failing; is the service restarting.

Observability was implemented to fit the existing RVPay architecture: the
existing zerolog conventions and package structure are preserved, no new
infrastructure was introduced, and no security or performance work was
performed.

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

All required documents were present and read.

## 3. Existing Observability

- **Logger**: zerolog, JSON output with timestamps and caller info; `LOG_LEVEL`
  env-configured (`shared/logger` for Clients/Transactions; service-local for
  the legacy services).
- **Health**: every service exposes `/healthz` (HTTP; 200/405, 503 while
  shutting down) and the gRPC health protocol (SERVING/NOT_SERVING). Render
  uses `/healthz` (see `docs/platform-render-review.md`).
- **Request IDs**: none existed.
- **Tracing**: none existed; no architecture requirement documented.
- **Metrics**: none existed; no architecture requirement documented.
- **Error handling**: gRPC `grpc_recovery` recovery interceptor present;
  services return gRPC status errors; the embedded gateway translates gRPC
  errors to HTTP via grpc-gateway defaults.

## 4. Observability Changes

- **New `shared/observability` package** (place: `shared/observability/` per
  `docs/repository-layout.md` §5.2/§6; directive #67) containing:
  - `requestid.go` — context-based request ID helpers (`Header`,
    `NewRequestID`, `WithRequestID`, `RequestIDFromContext`, `GetOrCreate`,
    `GetOrCreateWithValue`). No global mutable state (directives #53, #54).
  - `grpc.go` — `UnaryServerInterceptor(logger)`: reads/creates the request ID
    from `X-Request-ID` gRPC metadata, attaches it to the context and echoes
    it back in response metadata, and logs `request_id`, `rpc`,
    `grpc_code` (OK/error), `duration_ms`, and the error detail (directives
    #14–#16, #27). Successful RPCs log at INFO; failures at WARN (directive
    #6/#65 style). Payloads are never logged.
  - `http.go` — `AccessLog(logger)` HTTP middleware: reads/creates the request
    ID from the `X-Request-ID` header, sets it on the response, and logs
    `request_id`, `method`, `path`, `status`, `duration_ms` (directives
    #17–#18, #28). `/healthz` probes log at DEBUG to avoid probe noise
    (directive #24). Authorization headers, tokens, and request bodies are
    never logged.
  - `observability_test.go` — focused unit tests (directives #72–#77).
- **Adoption** in the Clients and Transactions service bootstraps
  (`cmd/grpc-service/main.go`): the gRPC unary interceptor is chained after
  the existing `grpc_recovery` interceptor, and the HTTP gateway mux is
  wrapped with `AccessLog`. Legacy deposits/integrations were left untouched
  (no new-service migration required). Existing behavior is unchanged
  (directive #89); no duplicate middleware was created (directive #16).

## 5. Logging

- **Format**: structured JSON to stdout/stderr (zerolog; Timestamp + Caller).
- **Levels**: `debug`, `info`, `warn`, `error` via `LOG_LEVEL` (default info).
- **Fields**: `request_id`, `rpc`, `grpc_code`, `duration_ms`, `method`,
  `path`, `status`, `error` — only where relevant (directive #7). No payloads,
  no secrets (directives #10, #37–#39).
- **Lifecycle**: startup, database connect/ping, migration start/success/
  failure, server listen, shutdown events were already logged and are
  preserved (directives #9–#13, #44).
- **Error behavior**: gRPC failures log the status code and wrapped error;
  HTTP failures log method/path/status/duration; no repository/service/handler
  triple-logging pattern introduced (directives #50–#52) — observability lives
  at the transport interceptor/middleware boundary.

## 6. Request Correlation

- **Mechanism**: `X-Request-ID` (HTTP header and gRPC metadata key).
- **Generation**: created when absent (UUID).
- **Propagation**: HTTP middleware → context → handler; gRPC interceptor →
  context → handler, and echoed back in gRPC response metadata.
- **Returned to clients**: HTTP response header `X-Request-ID`.
- **Usage**: request ID is attached to access and RPC logs so an error can be
  correlated to a request.
- **Trace IDs**: not implemented — no distributed tracing architecture is
  documented (directive #29/#30). If tracing is required later, the request ID
  infrastructure provides the natural correlation hook for Agent 12/work.

## 7. Health

- **Liveness**: `/healthz` returns 200 while the process serves; 503 while
  shutting down. It does not depend on PostgreSQL availability (directives
  #20, #21).
- **Readiness**: Render's health check uses `/healthz` (`docs/platform-render-
  review.md`); gRPC health protocol reports SERVING/NOT_SERVING. No new health
  endpoints were created (directive #19).
- **Dependency checks**: the services already fail startup clearly when the
  database cannot be reached (`Connect` ping); health probes are not spammy
  (`/healthz` access logs at DEBUG).

## 8. Metrics

Not implemented. The repository has no documented metrics requirement
(Prometheus/OpenTelemetry/application counters were assessed in the audit and
are absent). Introducing a metrics stack would add infrastructure without an
architectural requirement (directives #33, #34). High-value metric points
(request count/error count/duration) are already available in structured logs
and can be derived if metrics are later required (Agent 11/12).

## 9. Tracing

Not implemented. OpenTelemetry and distributed tracing are not part of the
repository or its documented architecture, and were not introduced merely
because they are popular (directive #30). Request IDs provide per-request
correlation today; a tracing integration can build on the existing interceptor/
middleware structure without rewiring (directive #64).

## 10. Sensitive Data

- Access and RPC logs record method/path/status/request IDs/durations only —
  never Authorization headers, access/refresh tokens, API keys, SSO/webhook
  secrets, request bodies, or provider credentials (directives #10, #37–#39,
  #46, #47).
- The `TestAccessLog_NoPayloadLogging` test verifies that a request body
  containing a secret-like value and an `Authorization: Bearer` header never
  appear in logs.
- No PII (phone numbers, emails, identity documents) is logged by the new
  observability layer (directive #39).
- Financial data is not logged: no account numbers, payment tokens, or
  provider secrets (directive #38).

## 11. Render

- Logs are written to stdout/stderr, which Render collects as container/service
  logs (directives #60, #61); no log files are created inside containers and no
  log rotation is implemented (directive #62).
- `/healthz` remains the Render health-check path; the observability middleware
  logs health probes at DEBUG so Render's continuous checks do not spam INFO
  logs (directive #24, #23).
- No Render configuration was modified.

## 12. Docker

- The services write operational logs to stdout/stderr (JSON), consistent with
  the distroless non-root images and container logging (directives #61, #62).
- No persistence, log volume, or log-file path is required by the new
  observability layer; no Dockerfile change was needed.

## 13. Tests

`shared/observability/observability_test.go` (10 tests, all passing; no
external infrastructure — directives #72–#77):

- `TestNewRequestID` — unique generated IDs.
- `TestRequestIDContext` — context round-trip, generation, preservation.
- `TestGetOrCreateWithValue` — explicit value wins.
- `TestUnaryServerInterceptor_PropagatesRequestID` — metadata → context →
  handler; log fields `request_id`/`rpc`/`grpc_code`/`duration_ms`.
- `TestUnaryServerInterceptor_GeneratesRequestID` — generated when absent.
- `TestUnaryServerInterceptor_ErrorClassification` — `NotFound` logged at
  WARN with grpc_code and error detail.
- `TestAccessLog_PropagatesAndEchoesRequestID` — header → context → response;
  method/path/status/request_id logged.
- `TestAccessLog_GeneratesRequestIDWhenAbsent`.
- `TestAccessLog_HealthzAtDebug` — `/healthz` probes at DEBUG.
- `TestAccessLog_NoPayloadLogging` — request body and Authorization header
  never logged.

Commands executed: `go build ./shared/... ./clients/... ./transactions/...`
(OK), `go test ./shared/observability/... -v` (10/10 PASS).

## 14. Findings

| ID | Severity | Area | Finding | Resolution |
| --- | --- | --- | --- | --- |
| OBS-01 | MEDIUM | gRPC gateway | No request correlation existed; RPC and HTTP failures could not be tied to a request | ✅ Added `shared/observability` request-ID + gRPC/HTTP logging, adopted in Clients and Transactions |
| OBS-02 | INFO | Health | `/healthz` + gRPC health already present and correct (liveness, 503 on shutdown) | ✅ Preserved; documented |
| OBS-03 | INFO | Metrics/Tracing | No metrics/tracing architecture requirement in the repository | ✅ Assessed and intentionally not implemented; request ID provides correlation; documented for later agents |
| OBS-04 | INFO | Sensitive data | New logging layer must not leak credentials/tokens/PII/financial data | ✅ Verified by design (metadata-only fields) and `TestAccessLog_NoPayloadLogging` |
| OBS-05 | INFO | Legacy services | Deposits/Integrations retain their pre-existing logger/health only | ✅ Documented; not in scope for this agent |

## 15. Documentation Changes

- `README.md` — added a brief "Observability" section (logs, request IDs,
  access logs, health) so developers know how to observe the application
  (directives #78, #79, #83).

## 16. Documentation Check

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
| docs/platform-observability-review.md | ✅ (this document) |

## 17. Final Status

PASS WITH FOLLOW-UP