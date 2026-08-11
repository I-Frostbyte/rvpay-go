# Protobuf contracts

This directory contains the source protobuf definitions for `rvpay-go` gRPC
APIs. Generated Go code is committed outside this directory under
[`../grpc/go`](../grpc/go).

## Contents

```text
protobuf/
├── clients.proto        # Clients Service contract (clientsgrpc)
├── transactions.proto   # Transactions Service contract (transactionsgrpc)
├── common.proto         # Shared enums and messages (commongrpc)
├── deposits.proto       # Legacy Deposits Service contract (depositsgrpc)
├── integrations.proto   # Legacy Integrations Service contract (integrationsgrpc)
├── Makefile             # protobuf lint and Go code generation targets
└── README.md            # This file
```

Each service owns one protobuf package (`<service>grpc`); shared types live in
`commongrpc` (`common.proto`). Every `go_package` option matches the repository
Go module: `github.com/I-Frostbyte/rvpay-go/grpc/go/<package>`.

## Contracts

| File | Package | Service(s) | HTTP annotation examples |
| --- | --- | --- | --- |
| `clients.proto` | `clientsgrpc` | `ClientsService`, `PlatformsService`, `IntegrationsService` | `POST /v1/public/clients`, `GET /v1/public/clients/{id}`, `GET /v1/public/platforms`, `POST /v1/public/integrations`, `GET /v1/public/clients/{client_id}/integrations` |
| `transactions.proto` | `transactionsgrpc` | `MerchantService`, `CustomerService`, `DepositService`, `PayoutService` | `POST/GET /v1/public/merchants[/{merchant_id}]`, `POST/GET /v1/public/customers[/{customer_id}]`, `POST/GET /v1/public/deposits[/{deposit_id}]`, `POST/GET /v1/public/payouts[/{payout_id}]` |
| `common.proto` | `commongrpc` | Shared types (status enums, `Provider`, `PaymentType`, `Money`, pagination, errors) | none |
| `deposits.proto` | `depositsgrpc` | `DepositsService` (legacy) | `POST /v1/public/deposits` |
| `integrations.proto` | `integrationsgrpc` | `IntegrationService` (legacy) | `POST/GET/DELETE /v1/public/integrations`, `POST /v1/public/webhooks` |

All public RPCs carry a `google.api.http` annotation; REST paths are under
`/v1/public/...`. See [`../docs/protobuf-strategy.md`](../docs/protobuf-strategy.md)
for the authoritative protobuf architecture.

## Generate Go stubs

Prerequisites:

- `protoc`
- `protoc-gen-go`
- `protoc-gen-go-grpc`
- `protoc-gen-grpc-gateway`
- the `third_party/googleapis` submodule, because the schemas import
  `google/api/annotations.proto`

From the repository root, initialise the submodule if necessary:

```bash
git submodule update --init --recursive
```

Then, from this directory:

```bash
make lint              # clang-format dry-run over *.proto
make generate-protos   # generate Go, gRPC, and gateway stubs
```

For every `*.proto` source, the Makefile creates `../grpc/go/<proto-name>grpc/`
and invokes `protoc` with source-relative Go, gRPC, and gateway output. For
`clients.proto`, the expected generated files are:

```text
../grpc/go/clientsgrpc/clients.pb.go
../grpc/go/clientsgrpc/clients_grpc.pb.go
../grpc/go/clientsgrpc/clients.pb.gw.go
```

Every HTTP-exposed service also gets a generated gateway file (`*.pb.gw.go`) in
its package; the runtime gateway is wired in each service's
`cmd/grpc-service/main.go`.

Generated files should not be edited by hand. Regenerate them after changing a
contract, then review and commit the generated output with the `.proto` change.

## Lint

```bash
make lint
```

This runs `clang-format --dry-run --Werror` over the protobuf sources.

## Compatibility notes

- Changes to field numbers or enum numeric values are wire-breaking. Preserve
  existing numbers and reserve removed values instead of reusing them.
- Add fields rather than renaming/removing deployed fields where compatibility
  matters.
- The `go_package` option for every contract matches the repository Go module
  (`github.com/I-Frostbyte/rvpay-go`), so generated code is importable as a
  standard Go package.
- Legacy `deposits.proto` and `integrations.proto` remain committed while the
  Deposits and Integrations services are runnable; they are retired with those
  services per `docs/migration-plan.md`.