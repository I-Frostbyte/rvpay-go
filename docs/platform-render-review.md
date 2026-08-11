# Platform Render Review

Document Version: 1.0
Status: Complete
System: RVPay
Review: Platform Agent 07 — Render Deployment

## 1. Objective

Implement and validate the Render deployment configuration for the new RVPay
architecture so the services and supporting infrastructure can be deployed to
Render as independently managed production components while preserving the
service boundaries defined by the project documentation.

This agent owns Render Blueprint configuration, service definitions, env
wiring, PostgreSQL configuration, health-check configuration, deployment
dependencies, and Render deployment documentation. No application logic,
Dockerfile, CI workflow, protobuf, migration, or business logic was changed.

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

All required documents were present and read.

## 3. Existing Render Configuration

| File | Purpose | Decision |
| --- | --- | --- |
| `render.yaml` | Render Blueprint: `rvpay-deposits` web service + `rvpay-postgres` managed DB | **Modified** — added `rvpay-clients` and `rvpay-transactions` services + two databases (additive; deposits + existing DB preserved) |
| `deploy/render/README.md` | Render deployment documentation (deposits-only) | **Modified** — updated for the three-service architecture |

## 4. Deployment Model

- **Before:** a single Render Blueprint deploying only the legacy Deposits
  service and one managed PostgreSQL database (audit finding P-03).
- **After:** a Render Blueprint deploying three independent web services
  (deposits, clients, transactions), each with its own managed PostgreSQL
  database, wired via Render's `fromDatabase` env mechanism.

## 5. Service Inventory

| Service | Dockerfile | Build Context | Binary | Ports | Public | Health Check | DB |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `rvpay-deposits` | `deposits/Dockerfile` | repo root (`.`) | `deposits-grpc-service` | gRPC 50051, HTTP `PORT` | public (web) | `/healthz` on `PORT` | `rvpay-postgres` |
| `rvpay-clients` | `clients/Dockerfile` | repo root (`.`) | `clients-grpc-service` | gRPC 50051, HTTP `PORT` | public (web) | `/healthz` on `PORT` | `rvpay-clients-postgres` |
| `rvpay-transactions` | `transactions/Dockerfile` | repo root (`.`) | `transactions-grpc-service` | gRPC 50051, HTTP `PORT` | public (web) | `/healthz` on `PORT` | `rvpay-transactions-postgres` |

All three are long-running network servers → `type: web` (directives #31, #32).
No workers or one-shot jobs were introduced (directives #33, #34). No
placeholder/future services were created (directive #10).

## 6. Blueprint

`render.yaml` now declares three `type: web` services and three databases:

- `rvpay-deposits` → `rvpay-postgres` (unchanged)
- `rvpay-clients` → `rvpay-clients-postgres` (new)
- `rvpay-transactions` → `rvpay-transactions-postgres` (new)

Each service uses `runtime: docker`, `dockerfilePath: ./<svc>/Dockerfile`,
`dockerContext: .` (repository root — required by the Dockerfiles for root
`go.mod`/`go.sum`, `grpc/go/`, and `shared/`; directives #36–#38), and
`healthCheckPath: /healthz` (an endpoint that actually exists; directives
#44–#47). `autoDeploy: true` and `branch: main` follow the existing deposits
convention.

## 7. Service Definitions

### rvpay-clients

- `dockerfilePath: ./clients/Dockerfile`, `dockerContext: .`
- Env: `LISTEN_PORT=50051`, `MIGRATION_PATH=/app/clients/db/migrations`,
  `RUN_MIGRATIONS=true`, `LOG_LEVEL=info`, `DB_TLS_DISABLED=false`, and
  `DB_*` from `rvpay-clients-postgres`.
- Manual secrets (`sync: false`): `HIGHLEVEL_CLIENT_ID`,
  `HIGHLEVEL_CLIENT_SECRET`, `HIGHLEVEL_REDIRECT_URI`, `WEBHOOK_SECRET`
  (directives #58–#61; no values committed).

### rvpay-transactions

- `dockerfilePath: ./transactions/Dockerfile`, `dockerContext: .`
- Env: `LISTEN_PORT=50051`, `MIGRATION_PATH=/app/transactions/db/migrations`,
  `RUN_MIGRATIONS=true`, `LOG_LEVEL=info`, `DB_TLS_DISABLED=false`, and
  `DB_*` from `rvpay-transactions-postgres`.
- No external API secrets are required by the current Transactions config.

## 8. Configuration

- All runtime configuration is environment-driven via the Blueprint
  (directives #20, #25–#28). No hostname, username, password, port, or
  database name is hard-coded in source or Dockerfiles.
- Non-sensitive values (`LISTEN_PORT`, `MIGRATION_PATH`, `RUN_MIGRATIONS`,
  `LOG_LEVEL`, `DB_TLS_DISABLED`) are set directly in the Blueprint
  (directive #28).
- Secrets are `sync: false` manual secrets — never committed (directives
  #23, #24, #27).
- No `localhost`/`127.0.0.1`/`::1` is used for production inter-service or
  database communication (directives #15, #16); Render's `fromDatabase`
  provides the internal DB host.

## 9. PostgreSQL

- One managed database per service, matching the domain model's
  service-owned-database boundary (directives #17, #18; docs/domain-model.md
  §4). No shared tables.
- Existing `rvpay-postgres` (deposits) preserved; no data migration, no
  production settings changed (directive #19).
- Database credentials are wired via Render's `fromDatabase` mechanism
  (directives #20–#22); the internal connection is used for Render-hosted
  services.

## 10. Health Checks

- Each service's health check targets `/healthz` on Render-injected `PORT`
  (directives #44–#47). This endpoint exists in every service (verified in
  Agent 03). No invented `/health`, `/status`, or `/ready` endpoints.
- No gRPC health-check-only configuration was added (directive #45); the
  HTTP `/healthz` is the documented health path.
- No large health-check delays were introduced (directive #48).

## 11. Migrations

- `RUN_MIGRATIONS=true` per service: each service runs `golang-migrate` at
  startup against its own database (directives #50–#52; docs/migration-plan.md
  and the Docker review §10).
- Migration files are present in each image at the configured
  `MIGRATION_PATH`.
- Migration concurrency risk for multi-instance scaling is documented in
  `deploy/render/README.md` (directive #53).

## 12. Security

- No secrets in `render.yaml` or committed files (directives #23, #24, #27).
- Secrets are manual (`sync: false`) and supplied at runtime (directives
  #56, #57).
- No build-time secrets; runtime-only secrets are not injected into Docker
  builds (directive #57).
- OAuth redirect URL is documented as requiring the public Render URL, not
  localhost (directive #59).

## 13. Local Validation

| Command | Result |
| --- | --- |
| `python3 -c "import yaml; yaml.safe_load(open('render.yaml'))"` | ✅ YAML valid |
| Blueprint parse: services | ✅ `['rvpay-deposits', 'rvpay-clients', 'rvpay-transactions']` |
| Blueprint parse: databases | ✅ `['rvpay-postgres', 'rvpay-clients-postgres', 'rvpay-transactions-postgres']` |

## 14. CI Compatibility

- The Render Blueprint is consumed by Render's own Blueprint provisioning
  (not GitHub Actions). The CI `render-deploy.yml` deploy step POSTs to
  `RENDER_DEPLOY_HOOK`; the Blueprint's `autoDeploy: true` also triggers
  deploys on `main` pushes. No CI workflow was modified.
- The Docker build contexts in the Blueprint match the Agent 05 CI matrix and
  the Agent 06 Docker review (directives #35–#38).

## 15. Findings

| ID | Severity | File/Area | Finding | Resolution |
| --- | --- | --- | --- | --- |
| RND-01 | MEDIUM | `render.yaml` | Blueprint deployed only deposits (audit P-03); clients/transactions not deployable via Render | ✅ Added `rvpay-clients` + `rvpay-transactions` services and their databases |
| RND-02 | INFO | `render.yaml` | Three managed PostgreSQL databases required (one per service) | ✅ Implemented; free-tier single-DB limit documented in `deploy/render/README.md` with a shared-DB fallback |
| RND-03 | INFO | `deploy/render/README.md` | Docs described deposits-only architecture | ✅ Updated for three services, env tables, setup, and free-tier note |
| RND-04 | INFO | `render.yaml` | Clients secrets (HighLevel/Webhook) not previously wired | ✅ Added as `sync: false` manual secrets |

## 16. Deferred Work

| Issue | Owner Agent |
| --- | --- |
| OAuth redirect URL final value depends on Render's generated hostname — must be set post-provisioning | Agent 07 follow-up / operator |
| Webhook endpoint public reachability for HighLevel delivery | Agent 07 follow-up / operator |
| Free-tier single-PostgreSQL constraint (if staying free, consolidate to one DB with distinct `DB_NAME`) | Operator decision |
| Vulnerability scanning of deployed images | Agent 10 (Security) |
| Scaling/autoscaling of Render services | Agent 11 (Performance) |

## 17. Changes Made

- `render.yaml` — added `rvpay-clients` and `rvpay-transactions` web services
  and `rvpay-clients-postgres` + `rvpay-transactions-postgres` databases;
  deposits service and `rvpay-postgres` preserved unchanged.
- `deploy/render/README.md` — updated architecture diagram, Blueprint
  description, env tables, one-time setup, CI/CD section, and added a
  free-tier note.

No other files were modified.

## 18. Documentation Check

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

## 19. Final Status

PASS WITH FOLLOW-UP