# Render deployment

This directory documents deploying the Deposits service to
[Render](https://render.com). Render builds the Docker image from
`deposits/Dockerfile`, runs the service, and provisions a managed PostgreSQL
database.

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
```

- Render injects `PORT` automatically; the service binds the HTTP gateway to it.
- The gRPC server binds to `LISTEN_PORT` (default `50051`).
- Render's health check hits `/healthz` on `PORT`.
- The service applies migrations at startup because `RUN_MIGRATIONS=true`.

## Blueprint

`render.yaml` at the repository root is the Render Blueprint. It declares:

- a **web service** `rvpay-deposits` using the Docker runtime
- a **managed PostgreSQL** database `rvpay-postgres`
- environment wiring from the database to the service (`DB_HOST`, `DB_PORT`,
  `DB_USER`, `DB_PASSWORD`, `DB_NAME`)
- `PAWAPAY_API_URL` and `PAWAPAY_API_KEY` as manual secrets (`sync: false`)

## One-time setup

1. Create a Render account and connect the `I-Frostbyte/rvpay-go` GitHub
   repository.
2. From the Render dashboard, choose **New → Blueprint** and select the
   repository. Render reads `render.yaml` and provisions the service and
   database.
3. Set the two manual secrets on the `rvpay-deposits` service:
   - `PAWAPAY_API_URL` — PawaPay API base URL
   - `PAWAPAY_API_KEY` — PawaPay credential
4. (Optional) Create a **Deploy Hook** for the service and store its URL in the
   GitHub secret `RENDER_DEPLOY_HOOK` to enable the CI deploy step.

## Environment variables

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

## CI/CD

`.github/workflows/render-deploy.yml` runs on pushes to `main`:

1. `generate` — runs `go generate ./...`, then `protoc` generation via
   `protobuf/Makefile`, then sqlc generation, and fails if the working tree
   differs (ensures generated code is committed and current)
2. `test` — runs `go test ./...`
3. `docker-build` — builds the image with `deposits/Dockerfile`
4. `deploy` — POSTs to the `RENDER_DEPLOY_HOOK` secret if set

The existing OCI workflow (`.github/workflows/deploy.yml`) is left untouched.

## Migrations

`RUN_MIGRATIONS=true` makes the service run `golang-migrate` at startup against
the managed database. This is appropriate for a single-instance Render service.
If you later scale to multiple instances, move migrations to a separate job and
set `RUN_MIGRATIONS=false` to avoid migration races.

## Notes

- Render currently runs x86_64; the Dockerfile builds for `linux/amd64` and does
  not force ARM.
- The service runs as the non-root `nonroot` user in a distroless image.
- Generated protobuf code is committed under `grpc/go/`, so the Docker build
  works from a clean clone without running `protoc`.