# Render deployment

This directory documents deploying the RVPay services to
[Render](https://render.com). Render builds each Docker image from the
service's Dockerfile, runs the service, and provisions a managed PostgreSQL
database per service.

## Architecture on Render

```text
Internet
    │  HTTPS (Render-managed TLS)
    ▼
Render Web Service (rvpay-deposits)
    ├── HTTP gateway on :PORT  →  /healthz, /v1/public/deposits
    └── gRPC server on :LISTEN_PORT
    ▼
Render Managed PostgreSQL (rvpay-postgres)

Internet
    │  HTTPS (Render-managed TLS)
    ▼
Render Web Service (rvpay-clients)
    ├── HTTP gateway on :PORT  →  /healthz, /v1/public/clients, /v1/public/platforms, /v1/public/integrations
    └── gRPC server on :LISTEN_PORT
    ▼
Render Managed PostgreSQL (rvpay-clients-postgres)

Internet
    │  HTTPS (Render-managed TLS)
    ▼
Render Web Service (rvpay-transactions)
    ├── HTTP gateway on :PORT  →  /healthz, /v1/public/merchants, /v1/public/customers, /v1/public/deposits, /v1/public/payouts
    └── gRPC server on :LISTEN_PORT
    ▼
Render Managed PostgreSQL (rvpay-transactions-postgres)
```

- Render injects `PORT` automatically; each service binds its HTTP gateway to it.
- Each gRPC server binds to `LISTEN_PORT` (default `50051`).
- Render's health check hits `/healthz` on `PORT` for each service.
- Each service applies its own migrations at startup because
  `RUN_MIGRATIONS=true`.
- Each service owns its own database (per the domain model); services never
  share database tables.

## Blueprint

`render.yaml` at the repository root is the Render Blueprint. It declares:

- three **web services** using the Docker runtime:
  - `rvpay-deposits` (`deposits/Dockerfile`)
  - `rvpay-clients` (`clients/Dockerfile`)
  - `rvpay-transactions` (`transactions/Dockerfile`)
- three **managed PostgreSQL** databases:
  - `rvpay-postgres` (deposits)
  - `rvpay-clients-postgres` (clients)
  - `rvpay-transactions-postgres` (transactions)
- environment wiring from each database to its service (`DB_HOST`, `DB_PORT`,
  `DB_USER`, `DB_PASSWORD`, `DB_NAME`)
- manual secrets (`sync: false`):
  - `PAWAPAY_API_URL`, `PAWAPAY_API_KEY` (deposits)
  - `HIGHLEVEL_CLIENT_ID`, `HIGHLEVEL_CLIENT_SECRET`,
    `HIGHLEVEL_REDIRECT_URI`, `WEBHOOK_SECRET` (clients)

All services use the repository root as the Docker build context
(`dockerContext: .`), matching the Dockerfiles' requirement for the root
`go.mod`/`go.sum`, the generated `grpc/go/` packages, and the `shared/`
packages.

## One-time setup

1. Create a Render account and connect the `I-Frostbyte/rvpay-go` GitHub
   repository.
2. From the Render dashboard, choose **New → Blueprint** and select the
   repository. Render reads `render.yaml` and provisions the services and
   databases.
3. Set the manual secrets on each service:
   - `rvpay-deposits`: `PAWAPAY_API_URL`, `PAWAPAY_API_KEY`
   - `rvpay-clients`: `HIGHLEVEL_CLIENT_ID`, `HIGHLEVEL_CLIENT_SECRET`,
     `HIGHLEVEL_REDIRECT_URI`, `WEBHOOK_SECRET`
4. (Optional) Create a **Deploy Hook** for each service and store its URL in
   the GitHub secret `RENDER_DEPLOY_HOOK` to enable the CI deploy step.

> **Free-tier note:** Render's free tier includes one free managed PostgreSQL
> database. Deploying three managed databases (one per service) requires a
> paid plan. If you must stay on the free tier, point all three services at a
> single database and set distinct `DB_NAME` values (logical databases) — the
> services remain isolated by schema/table ownership.

## Environment variables

### rvpay-deposits

| Variable | Source | Purpose |
| --- | --- | --- |
| `PORT` | Render (injected) | HTTP gateway port |
| `LISTEN_PORT` | Blueprint (`50051`) | gRPC server port |
| `MIGRATION_PATH` | Blueprint (`/app/deposits/db/migrations`) | Migration directory |
| `RUN_MIGRATIONS` | Blueprint (`true`) | Apply migrations at startup |
| `LOG_LEVEL` | Blueprint (`info`) | Zerolog level |
| `DB_TLS_DISABLED` | Blueprint (`false`) | Use `sslmode=require` |
| `DB_HOST` | Blueprint (from database) | PostgreSQL host |
| `DB_PORT` | Blueprint (from database) | PostgreSQL port |
| `DB_USER` | Blueprint (from database) | PostgreSQL user |
| `DB_PASSWORD` | Blueprint (from database) | PostgreSQL password |
| `DB_NAME` | Blueprint (from database) | PostgreSQL database |
| `PAWAPAY_API_URL` | Manual secret | PawaPay API base URL |
| `PAWAPAY_API_KEY` | Manual secret | PawaPay credential |

### rvpay-clients

| Variable | Source | Purpose |
| --- | --- | --- |
| `PORT` | Render (injected) | HTTP gateway port |
| `LISTEN_PORT` | Blueprint (`50051`) | gRPC server port |
| `MIGRATION_PATH` | Blueprint (`/app/clients/db/migrations`) | Migration directory |
| `RUN_MIGRATIONS` | Blueprint (`true`) | Apply migrations at startup |
| `LOG_LEVEL` | Blueprint (`info`) | Zerolog level |
| `DB_TLS_DISABLED` | Blueprint (`false`) | Use `sslmode=require` |
| `DB_HOST` | Blueprint (from database) | PostgreSQL host |
| `DB_PORT` | Blueprint (from database) | PostgreSQL port |
| `DB_USER` | Blueprint (from database) | PostgreSQL user |
| `DB_PASSWORD` | Blueprint (from database) | PostgreSQL password |
| `DB_NAME` | Blueprint (from database) | PostgreSQL database |
| `HIGHLEVEL_CLIENT_ID` | Manual secret | HighLevel OAuth client ID |
| `HIGHLEVEL_CLIENT_SECRET` | Manual secret | HighLevel OAuth client secret |
| `HIGHLEVEL_REDIRECT_URI` | Manual secret | HighLevel OAuth redirect URL (must be the public Render URL) |
| `WEBHOOK_SECRET` | Manual secret | HighLevel webhook verification secret |

### rvpay-transactions

| Variable | Source | Purpose |
| --- | --- | --- |
| `PORT` | Render (injected) | HTTP gateway port |
| `LISTEN_PORT` | Blueprint (`50051`) | gRPC server port |
| `MIGRATION_PATH` | Blueprint (`/app/transactions/db/migrations`) | Migration directory |
| `RUN_MIGRATIONS` | Blueprint (`true`) | Apply migrations at startup |
| `LOG_LEVEL` | Blueprint (`info`) | Zerolog level |
| `DB_TLS_DISABLED` | Blueprint (`false`) | Use `sslmode=require` |
| `DB_HOST` | Blueprint (from database) | PostgreSQL host |
| `DB_PORT` | Blueprint (from database) | PostgreSQL port |
| `DB_USER` | Blueprint (from database) | PostgreSQL user |
| `DB_PASSWORD` | Blueprint (from database) | PostgreSQL password |
| `DB_NAME` | Blueprint (from database) | PostgreSQL database |

## CI/CD

`.github/workflows/render-deploy.yml` runs on pushes to `main`:

1. `generate` — runs `go generate ./...`, then `protoc` generation via
   `protobuf/Makefile`, then sqlc generation, and fails if the working tree
   differs (ensures generated code is committed and current)
2. `validate` — `gofmt` (hand-written Go), `go vet ./...`, `go build ./...`
3. `test` — runs `go test ./...`
4. `docker-build` — builds the images with `deposits/Dockerfile`,
   `clients/Dockerfile`, and `transactions/Dockerfile`
5. `deploy` — POSTs to the `RENDER_DEPLOY_HOOK` secret if set

The existing OCI workflow (`.github/workflows/deploy.yml`) is left untouched.

## Migrations

`RUN_MIGRATIONS=true` makes each service run `golang-migrate` at startup
against its own managed database. This is appropriate for single-instance
Render services. If you later scale a service to multiple instances, move
migrations to a separate job and set `RUN_MIGRATIONS=false` to avoid migration
races.

## Notes

- Render currently runs x86_64; the Dockerfiles build for `linux/amd64` and do
  not force ARM.
- The services run as the non-root `nonroot` user in distroless images.
- Generated protobuf code is committed under `grpc/go/`, so the Docker builds
  work from a clean clone without running `protoc`.
- The `shared/` packages (Agent 04) are included in the clients and
  transactions Docker build contexts.