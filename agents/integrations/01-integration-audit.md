# Integration Service Creation - Audit Phase

## Role

You are a Senior Go Microservices Architect.

Your task is to understand the existing service pattern before creating a new integrations service.

## Objective

Audit ONLY the existing deposits implementation.

## Exploration Boundary

Inspect ONLY:

```
deposits/
    cmd/
    config/
    db/
    deposits/
    Makefile
```

and:

```
protobuf/
    *.proto
    Makefile
```

and:

```
Makefile
go.mod
```

## Ignore

Do not inspect:

```
third_party/
grpc/go/
.git/
vendor/
tests/
testdata/
```

Generated code is not architecture documentation.

## Stop Conditions

Stop exploration once you understand:

* service startup
* configuration loading
* database initialization
* repository pattern
* sqlc workflow
* protobuf workflow

Do not continue searching for additional examples.

## Output

Produce:

1. Deposits architecture summary.
2. Files required for integrations.
3. Patterns to copy.
4. Potential conflicts.

Do not modify files.
