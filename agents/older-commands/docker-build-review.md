# Docker Build Review

## Objective

Determine whether Docker builds are fully reproducible.

Remain READ ONLY.

---

## Review

Inspect:

Dockerfile

Makefile

sqlc

protobuf

go:generate

generated code

---

## Determine

Can Docker build successfully from a clean clone?

Assume:

No generated files exist.

If not,

recommend changes.

---

## Verify

Should Docker perform:

protoc

↓

sqlc

↓

go generate

↓

go build

or should CI generate these beforehand?

Recommend the most reproducible solution.

Explain tradeoffs.

---

## Output

Current build pipeline

↓

Recommended build pipeline

↓

Required changes