# Protobuf contracts

This directory contains the source protobuf definitions for `rvpay-go` gRPC
APIs. Generated Go code is committed outside this directory under
[`../grpc/go`](../grpc/go).

## Contents

```text
protobuf/
├── deposits.proto  # Deposits API contract
├── Makefile        # protobuf lint and Go code generation targets
└── Dockerfile      # currently empty
```

`deposits.proto` declares package `depositsgrpc` and generates the Go client,
server interface, messages, and enums used by the Deposits service.

## Deposits contract

The service exposes one RPC:

```proto
rpc InitiateDeposit(CreateDepositRequest) returns (CreateDepositResponse)
```

Its gRPC method name is
`/depositsgrpc.DepositsService/InitiateDeposit`. The schema also supplies the
Google API HTTP annotation `POST /v1/public/deposits` with the full request as
the body. This repository currently generates only Go gRPC stubs; it does not
generate or run a gRPC-Gateway HTTP endpoint.

`CreateDepositRequest` contains:

- `amount` as a string (for example, `"1000.00"`)
- `currency` as a string (the database expects an uppercase three-letter code)
- `payer`, containing a deposit type and account details
- `client_id` (defined by the contract but not used by the current service)

`CreateDepositResponse` returns `deposit_id`, a `DepositStatus`, and
`next_step`. The present service returns `DEPOSIT_STATUS_ACCEPTED` and
`FINAL_STATUS` after it successfully initiates the PawaPay request.

The available provider enum values are MTN MoMo Cameroon and Orange MoMo
Cameroon. The available payer type values are MMO and CARD; the current server
only meaningfully handles MMO.

## Generate Go stubs

Prerequisites:

- `protoc`
- `protoc-gen-go`
- `protoc-gen-go-grpc`
- the `third_party/googleapis` submodule, because the schema imports
  `google/api/annotations.proto`

From the repository root, initialise the submodule if necessary:

```bash
git submodule update --init --recursive
```

Then, from this directory:

```bash
make generate-protos
```

For every `*.proto` source, the Makefile creates
`../grpc/go/<proto-name>grpc/` and invokes `protoc` with source-relative Go and
gRPC output. For `deposits.proto`, the expected generated files are:

```text
../grpc/go/depositsgrpc/deposits.pb.go
../grpc/go/depositsgrpc/deposits_grpc.pb.go
```

Generated files should not be edited by hand. Regenerate them after changing a
contract, then review and commit the generated output with the `.proto` change.

## Lint

```bash
make lint
```

This runs `clang-format --dry-run --Werror` over the protobuf sources (and any
files in `types/` if that directory is added). Run `clang-format` to apply
formatting before rerunning the lint target.

## Compatibility notes

- Changes to field numbers or enum numeric values are wire-breaking. Preserve
  existing numbers and reserve removed values instead of reusing them.
- Add fields rather than renaming/removing deployed fields where compatibility
  matters.
- The `go_package` option currently names
  `github.com/rvpay/rvpay-go/grpc/go/depositsgrpc`, while this repository’s Go
  module is `github.com/I-Frostbyte/rvpay-go`. The checked-in generated package
  is used locally by its directory path, but align this option before relying on
  regenerated code as a published Go import path.
