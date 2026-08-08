# RVPay Repository Layout

Document Version: 1.0
Status: Foundation
System: RVPay
Architecture: Go Microservices

## 1. Purpose

This document defines the target repository structure required to evolve the
current repository into the new RVPay architecture described in
`docs/00-technical-design-document.md` and `docs/domain-model.md`.

This is a structure-only design. No files are moved or deleted by this document.

## 2. Current Layout Summary

The repository currently contains two gRPC services plus shared infrastructure:

| Directory     | Purpose                                                          |
|---------------|------------------------------------------------------------------|
| `deposits/`   | Deposits service: clients + deposits, PawaPay integration        |
| `integrations/`| Integrations service: HighLevel OAuth + webhooks                |
| `protobuf/`   | Source protobuf contracts and generation                         |
| `grpc/go/`    | Generated Go protobuf and gRPC/gateway stubs                     |
| `third_party/`| googleapis Git submodule for protoc                              |
| `nginx/`      | TLS termination config for OCI                                   |
| `deploy/`     | OCI and Render deployment documentation                          |
| `docs/`       | Technical design and domain documentation                        |
| `.github/`    | CI/CD pipelines                                                  |

The deposits template is the canonical service layout, applied consistently by
both services:

```text
<service>/
├── cmd/grpc-service/   # Process entry point and server lifecycle
├── config/             # Environment-backed service configuration
├── db/
│   ├── migrations/     # PostgreSQL up/down migrations
│   ├── query/          # SQL inputs for sqlc
│   ├── repo/           # Pool/query wrapper and migration runner
│   └── sqlc/           # Generated query models and methods
├── <domain>/           # Service implementation (business logic)
├── Dockerfile
├── Makefile
└── README.md
```

## 3. Service Evolution

Per `docs/domain-model.md`, the existing services evolve into the target
bounded contexts.

| Current Service | Evolves Into          | Bounded Context |
|-----------------|------------------------|-----------------|
| `integrations/` | Clients Service        | Clients Context |
| `deposits/`     | Transactions Service   | Transactions Context |

## 4. Future Directory Tree

```text
rvpay-go/
├── clients/                     # Clients Service (evolves from integrations/)
│   ├── cmd/grpc-service/         # Process entry point and server lifecycle
│   ├── config/                   # Environment-backed service configuration
│   ├── db/
│   │   ├── migrations/           # PostgreSQL up/down migrations
│   │   ├── query/                # SQL inputs for sqlc
│   │   ├── repo/                 # Pool/query wrapper and migration runner
│   │   └── sqlc/                 # Generated query models and methods
│   ├── clients/                  # ClientService implementation
│   ├── platforms/                # Platform entity logic
│   ├── integrations/             # Integration entity logic
│   ├── oauth/                    # OAuth callback handler and service
│   ├── webhook/                  # Webhook handler and service
│   ├── Dockerfile                # Multi-stage distroless container build
│   ├── Makefile                  # Service tasks
│   └── README.md                 # Service documentation
├── transactions/                # Transactions Service (evolves from deposits/)
│   ├── cmd/grpc-service/         # Process entry point and server lifecycle
│   ├── config/                   # Environment-backed service configuration
│   ├── db/
│   │   ├── migrations/           # PostgreSQL up/down migrations
│   │   ├── query/                # SQL inputs for sqlc
│   │   ├── repo/                 # Pool/query wrapper and migration runner
│   │   └── sqlc/                 # Generated query models and methods
│   ├── merchants/                # Merchant entity logic
│   ├── customers/                # Customer entity logic
│   ├── deposits/                 # Deposit processing
│   ├── payouts/                  # Payout processing
│   ├── Dockerfile                # Multi-stage distroless container build
│   ├── Makefile                  # Service tasks
│   └── README.md                 # Service documentation
├── shared/                      # Shared infrastructure packages (non-business)
│   ├── config/                   # Shared configuration helpers
│   ├── logger/                   # Shared zerolog setup
│   ├── database/                 # Shared PostgreSQL pool/helpers
│   └── middleware/               # Shared gRPC/gateway middleware
├── protobuf/                    # Source protobuf contracts and generation
│   ├── clients.proto             # Clients Service contract
│   ├── transactions.proto        # Transactions Service contract
│   ├── Makefile                  # protoc generation tasks
│   └── README.md
├── grpc/go/                     # Generated Go protobuf and gRPC stubs
│   ├── clientsgrpc/              # Generated Clients Service code
│   └── transactionsgrpc/         # Generated Transactions Service code
├── docs/                        # Repository documentation
│   ├── 00-technical-design-document.md
│   ├── domain-model.md
│   └── repository-layout.md
├── third_party/                 # External protobuf dependencies (submodule)
├── nginx/                       # TLS termination config for OCI
├── deploy/                      # OCI and Render deployment documentation
├── .github/workflows/           # CI/CD pipelines
├── docker-compose.yml           # OCI Always Free Compose stack
├── render.yaml                  # Render Blueprint
├── Makefile                     # Repository-wide test tasks
└── layout.md                    # Original layout notes (historical)
```

## 5. Purpose of Every Major Directory

### 5.1 Service Directories (`clients/`, `transactions/`)

Each service is fully self-contained and owns its own configuration, database
layer, repositories, business logic, and handlers. Services never share
database tables; cross-service access is gRPC-only.

- `cmd/grpc-service/` — process entry point; loads config, connects to
  PostgreSQL, runs migrations, registers gRPC and HTTP gateway, graceful
  shutdown.
- `config/` — environment-backed service configuration using the existing
  `LoadConfig` pattern.
- `db/` — database layer:
  - `migrations/` — up/down goose-style migrations.
  - `query/` — SQL inputs consumed by sqlc.
  - `repo/` — pool/query wrapper and migration runner.
  - `sqlc/` — generated query models and methods (never hand-edited).
- Domain folders — business logic per entity, following the deposits
  implementation style (service.go, errors.go, converters.go where relevant).

### 5.2 `shared/`

Shared infrastructure candidates extracted from across services. Only
non-business, generic infrastructure belongs here. Business logic stays in the
owning service.

- `config/` — shared configuration loading helpers.
- `logger/` — shared zerolog logger setup.
- `database/` — shared PostgreSQL pool construction and migration runner
  helpers (where identical across services).
- `middleware/` — shared gRPC recovery and gateway middleware.

### 5.3 `protobuf/` and `grpc/go/`

- `protobuf/` — the authoritative API contracts. `clients.proto` and
  `transactions.proto` replace/extend the current `deposits.proto` and
  `integrations.proto`.
- `grpc/go/` — generated Go stubs. Generated output is never hand-edited;
  regenerate via the protobuf Makefile.

### 5.4 `docs/`

Single documentation location for architecture-level documents
(`00-technical-design-document.md`, `domain-model.md`, `repository-layout.md`).

### 5.5 Infrastructure

- `third_party/` — external protobuf dependencies (googleapis submodule).
- `nginx/`, `deploy/`, `.github/`, `docker-compose.yml`, `render.yaml` —
  deployment and CI/CD as they exist today.

## 6. New Directories

| Directory            | Purpose                                             |
|----------------------|-----------------------------------------------------|
| `clients/`           | New Clients Service root (evolves from `integrations/`) |
| `transactions/`      | New Transactions Service root (evolves from `deposits/`) |
| `shared/`            | Shared infrastructure packages (config, logger, database, middleware) |
| `docs/`              | Already exists; becomes the canonical docs location  |

## 7. Renamed Directories

The service roots evolve by renaming existing service folders to their target
canonical names under the new two-service architecture:

| Current    | Target          | Note                                             |
|------------|-----------------|--------------------------------------------------|
| `deposits/`| `transactions/` | Grows merchants/, customers/, payouts/ contexts  |
| `integrations/`| `clients/`  | Grows clients/, platforms/ contexts              |

Per repository rules, folder renames must not be performed without explicit
instruction, and the existing deposits template must remain the pattern source
during any such migration. New services are created first; renames are applied
only when explicitly requested and are coordinated with protobuf, grpc/go, and
CI/CD references.

## 8. Shared Directories

- `shared/` — new shared infrastructure package location.
- `protobuf/` and `grpc/go/` — already shared across services today.
- `docs/`, `deploy/`, `.github/` — already shared today.

## 9. Generated Directories

- `grpc/go/` — protobuf/gRPC/gateway generated Go code.
- `clients/db/sqlc/` — sqlc generated models and methods.
- `transactions/db/sqlc/` — sqlc generated models and methods.

Generated code is output, never modified by hand.

## 10. Shared Code Candidates

Candidate code for extraction into `shared/` (generic infrastructure only):

- Service configuration loading (`LoadConfig`, env parsing) — currently
  duplicated per service.
- zerolog logger setup and configuration.
- PostgreSQL pool construction and migration runner helpers.
- gRPC recovery middleware and HTTP gateway middleware.

Business logic (deposit flow, OAuth, webhook handling, etc.) is explicitly
excluded from extraction and remains in the owning service.

## 11. Migration Notes

1. **Create new services first.** Scaffold `clients/` and `transactions/`
   following the deposits template. Do not delete or move existing code.
2. **Extract shared code.** Move generic infrastructure (config, logger,
   database helpers, middleware) into `shared/` incrementally; keep identical
   behaviour.
3. **Introduce protobuf contracts.** Add `clients.proto` and
   `transactions.proto` alongside existing contracts; regenerate `grpc/go/`.
   Do not rename existing contracts until callers are updated.
4. **Coordinate renames.** Applying the `integrations/ → clients/` and
   `deposits/ → transactions/` renames requires explicit approval and must be
   coordinated with protobuf, grpc/go, `.github/workflows/`, docker-compose,
   and render blueprint references.
5. **Retire legacy layout in stages.** Keep the current services runnable until
   their functionality is fully replaced by the new services.

## 12. Directories Scheduled for Deprecation

| Directory      | Status      | Plan                                             |
|----------------|-------------|--------------------------------------------------|
| `deposits/`    | Replace     | Evolves into `transactions/`; retired in stages  |
| `integrations/`| Replace     | Evolves into `clients/`; retired in stages       |
| `layout.md`    | Historical  | Retained as original layout notes                |
| `protobuf/deposits.proto`, `protobuf/integrations.proto` | Replace | Superseded by `clients.proto`, `transactions.proto`; retired after callers migrate |
| `grpc/go/depositsgrpc/`, `grpc/go/integrationsgrpc/` | Replace | Regenerated from new contracts; retired with the services |

Deprecation plans are descriptive. No files are removed by this document.