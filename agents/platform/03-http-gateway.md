# Agent 02 — Protobuf Generation

## Objective

Establish a correct, reproducible, repository-consistent protobuf generation system for RVPay.

This agent is responsible for the protobuf generation TOOLCHAIN and WORKFLOW.

It must ensure that:

- protobuf source files are located correctly
- protoc is invoked correctly
- protoc-gen-go is invoked correctly
- protoc-gen-go-grpc is invoked correctly
- protoc-gen-grpc-gateway is invoked correctly where required
- googleapis dependencies are resolved correctly
- generated Go code is placed in the correct locations
- generation is reproducible
- local generation and CI generation use the same versions
- generated output remains synchronized with protobuf source
- Clients and Transactions protobuf contracts can be generated consistently

This agent is NOT responsible for implementing the HTTP gateway.

That belongs to:

agents/platform/03-http-gateway.md

This agent is NOT responsible for redesigning protobuf contracts.

This agent is NOT responsible for designing business-domain messages.

This agent is responsible for making the existing documented protobuf strategy executable and reproducible.

---

# Required Reading

Read only:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md
- docs/platform-repository-audit.md

The first six documents are mandatory.

The Platform audit is mandatory because Agent 01 establishes the current repository state and identifies existing protobuf tooling.

---

# Documentation Check

Before starting:

confirm that all required documents exist and can be read.

Required:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md
- docs/platform-repository-audit.md

At the end:

perform the check again.

Record the result in the review document:

docs/platform-protobuf-generation-review.md

---

# Repository Exploration Rules

## IMPORTANT

Do NOT perform an unrestricted recursive repository search.

Use:

README.md

as the repository map.

Use:

docs/protobuf-strategy.md

as the primary source for protobuf architecture.

Use:

docs/repository-layout.md

for expected locations.

Use:

agents/project-context.md

for project conventions.

Use:

docs/platform-repository-audit.md

for the known current state.

---

# Do NOT Explore Deep Folders

Do NOT recursively inspect:

- .git/
- vendor/
- node_modules/
- coverage/
- tmp/
- bin/
- third_party/
- third_party/googleapis/

Especially:

DO NOT recursively inspect:

third_party/googleapis/

The googleapis directory is a dependency.

Only verify that the required dependency exists and that the protobuf compiler can resolve imports from it.

If a specific googleapis file is required by an active `.proto` import:

inspect only that specific file if necessary.

---

# Scope Restrictions

This agent may modify only files directly related to protobuf generation.

Potentially relevant files include:

- `.proto` source files, only when generation metadata or options are incorrect
- protobuf Makefiles
- generation scripts
- root Makefile generation targets
- tool version documentation
- generation configuration
- generated protobuf output
- generated gRPC output
- generated gateway output if the existing strategy requires it

Do NOT modify:

- business service implementations
- repositories
- migrations
- SQL
- Dockerfiles
- Render configuration
- CI workflows unless absolutely required to keep generation reproducible

If CI changes are necessary, document them for Agent 05 instead of implementing them unless the change is trivial and directly required by the generation contract.

---

# 1. Read the Existing Protobuf Strategy

Read:

docs/protobuf-strategy.md

Treat this document as authoritative for the intended protobuf architecture.

Extract:

- source directory
- package naming
- Go package naming
- generated output location
- service definitions
- gateway strategy
- googleapis dependency strategy
- generation commands
- committed/generated artifact policy

Do not silently replace the strategy with a different approach.

---

# 2. Read the Repository Layout

Read:

docs/repository-layout.md

Determine where protobuf source and generated code are supposed to live.

Compare this against the actual repository.

Do not reorganize directories simply to match personal preference.

---

# 3. Read the Platform Audit

Read:

docs/platform-repository-audit.md

Pay particular attention to:

- existing `.proto` files
- existing Makefiles
- existing generation commands
- tool versions
- generated output
- CI generation behavior
- known protobuf findings

Treat this as the starting point rather than repeating the entire repository audit.

---

# 4. Establish Git State

Run:

git status --short

Do not modify anything yet.

Record relevant existing changes.

Do not assume existing uncommitted changes belong to this agent.

---

# 5. Locate Protobuf Sources

Locate the active protobuf source files.

Use the documented repository structure.

Identify:

- service `.proto` files
- shared `.proto` files
- imported `.proto` files
- package declarations
- Go package options
- HTTP annotations
- external dependencies

Do not inspect unrelated protobuf files.

---

# 6. Identify Active Services

Verify the protobuf definitions for the current services.

At minimum account for:

- Clients
- Transactions

Also identify any additional active service definitions documented by the repository.

Do not invent new services.

---

# 7. Verify Package Names

For each active `.proto` file verify:

- `package`
- `option go_package`
- service naming
- message naming

They must follow:

docs/protobuf-strategy.md

and:

agents/project-context.md

Do not rename protobuf packages merely for stylistic preference.

---

# 8. Verify Proto Imports

Inspect imports in active `.proto` files.

Determine whether imports resolve from:

- local protobuf source
- googleapis
- protobuf standard definitions
- other documented dependencies

Do not modify third-party protobuf definitions.

---

# 9. Verify Google APIs Dependency

Confirm:

third_party/googleapis/

exists if required by the documented protobuf strategy.

Do not recursively inspect it.

Verify only that the expected import root is available.

For example, if the active proto imports:

google/api/annotations.proto

verify that the expected file exists.

Do not inspect unrelated Google API files.

---

# 10. Inspect Existing Generation Command

Find the existing generation command.

Likely locations include:

- protobuf/Makefile
- root Makefile
- generation scripts
- go generate directives

Determine exactly what command currently generates the code.

Do not replace a working command unnecessarily.

---

# 11. Verify protoc

Determine the required protoc version from the project documentation.

Use:

tools/versions.md

if present.

Verify the locally installed version if practical:

protoc --version

Do not upgrade it simply because a newer version exists.

---

# 12. Verify protoc-gen-go

Determine the project-required version.

If practical:

protoc-gen-go --version

Verify that the binary is available.

Do not blindly install another version.

---

# 13. Verify protoc-gen-go-grpc

Determine the project-required version.

If practical:

protoc-gen-go-grpc --version

Verify that the binary is available.

Do not invent a version.

---

# 14. Verify protoc-gen-grpc-gateway

If gateway generation is part of the documented protobuf strategy:

verify:

protoc-gen-grpc-gateway

and its configured version.

Do not implement gateway runtime behavior here.

---

# 15. Version Consistency

Compare versions across:

- tools/versions.md
- go.mod
- Makefile
- CI workflow
- generated-code headers

Look for mismatches.

Example:

Generated code says:

protoc-gen-go v1.36.11

while tools/versions.md says:

v1.36.10

This is a generation drift issue.

Determine which version the project documentation explicitly requires.

Do not choose arbitrarily.

---

# 16. Generated Code Headers

Inspect generated files only enough to determine:

- generator versions
- source proto
- generated package

Do not manually edit generated files.

Generated headers are evidence, not configuration.

---

# 17. Generation Output Paths

Verify that generation writes files to the documented locations.

For example:

grpc/go/...

or whatever:

docs/repository-layout.md

and:

docs/protobuf-strategy.md

specify.

Do not create a second competing generated-code tree.

---

# 18. source_relative Policy

If the project uses:

paths=source_relative

verify that:

- `--go_opt=paths=source_relative`
- `--go-grpc_opt=paths=source_relative`

are applied consistently.

Do not introduce a different path strategy without explicit documentation.

---

# 19. Go Package Layout

Verify that generated files compile under their intended Go package.

Pay attention to:

- directory location
- `go_package`
- package declarations
- imports

The generated Go package should align with the documented repository structure.

---

# 20. Gateway Generation

If grpc-gateway generation is required:

verify the command includes the appropriate gateway plugin.

Do not implement:

- HTTP handlers
- HTTP server startup
- gateway routing policy

Those belong to Agent 03.

---

# 21. gRPC Generation

Verify that service stubs are generated using:

protoc-gen-go-grpc

and that generated interfaces correspond to the `.proto` service definitions.

Do not modify generated interfaces manually.

---

# 22. Protobuf Runtime Compatibility

Verify that generated code uses dependency versions compatible with:

go.mod

Do not upgrade protobuf libraries.

Do not modify dependency versions unless a concrete compatibility failure requires it.

If a dependency mismatch exists:

document it.

---

# 23. Generation Command Reproducibility

The generation process must be reproducible.

A fresh developer should be able to determine:

1. which tools are required
2. which versions are required
3. where protobuf sources live
4. which command generates them
5. where output is written

Do not rely on undocumented local state.

---

# 24. Local Generation

Run the project's documented protobuf generation command.

Use the existing project command where possible.

Do not invent an alternative command merely because it is shorter.

---

# 25. Inspect Generation Result

After generation inspect:

git status --short

and:

git diff --stat

Determine:

- which generated files changed
- whether changes are expected
- whether unrelated files changed

Do not accept unrelated generated output.

---

# 26. Generated Code Must Be Deterministic

Run the generation command again.

Then inspect:

git status --short

The second generation should not introduce additional differences.

If repeated generation produces differences:

investigate.

This is a generation reproducibility problem.

---

# 27. Do Not Manually Patch Generated Code

If generated output is wrong:

fix the source or generation configuration.

Never manually edit:

- `.pb.go`
- `_grpc.pb.go`
- generated gateway `.go`

unless the repository explicitly treats that file as non-generated.

---

# 28. Verify Generated Code Compilation

Run the narrowest useful Go test/build command.

For example:

go test ./grpc/go/...

if that path is valid.

If the repository structure differs:

use the documented equivalent.

Do not run expensive unrelated tests unless necessary.

---

# 29. Verify Service Compilation

Verify that active services importing generated protobuf code still compile.

At minimum:

- Clients
- Transactions

Use targeted package tests/builds first.

Do not run unrelated integration tests unless required.

---

# 30. Verify Protobuf Contract Compatibility

Do not redesign contracts.

Instead verify that generated code corresponds exactly to source definitions.

Check:

- RPC names
- request messages
- response messages
- package names
- service names

---

# 31. Verify HTTP Annotations

If the protobuf strategy uses grpc-gateway annotations:

verify:

google.api.http

annotations are syntactically correct.

Do not implement HTTP routing outside protobuf annotations.

Do not redesign HTTP endpoints.

---

# 32. Verify Streaming Definitions

If any service uses streaming RPCs:

verify that generated code supports the declared streaming mode.

Do not add streaming where none is documented.

---

# 33. Verify Shared Messages

Identify shared protobuf messages.

Ensure they are not duplicated unnecessarily across active services.

Do not consolidate messages unless the protobuf strategy explicitly requires it.

---

# 34. Verify Naming Consistency

Check:

- PascalCase messages
- RPC naming
- field naming
- package naming
- Go package naming

Follow existing project conventions.

Do not rename working APIs merely for stylistic reasons.

---

# 35. Verify Generated Artifact Policy

Determine whether generated output is expected to be committed.

Use:

README.md

and:

docs/protobuf-strategy.md

Do not assume.

If generated files are expected to be committed:

ensure they are present.

If they are intentionally ignored:

do not force them into git.

---

# 36. Makefile Integration

Inspect the existing protobuf Makefile.

Ensure there is one clear generation entrypoint.

If the project convention is:

make generate-protos

preserve it.

Do not create multiple overlapping targets.

---

# 37. Root Makefile Integration

If the root Makefile invokes protobuf generation:

verify that it calls the canonical protobuf generation target.

Do not duplicate the generation implementation.

---

# 38. go generate Integration

If:

go generate ./...

is part of the repository workflow:

verify that protobuf generation does not conflict with it.

Do not make `go generate` and Makefile generation produce different outputs.

---

# 39. CI Compatibility Preparation

This agent should ensure the generation command can be executed in a clean environment.

Do not redesign CI.

Agent 05 will handle the CI/CD workflow.

If CI currently has a concrete protobuf generation defect:

document it in:

docs/platform-protobuf-generation-review.md

and identify it for Agent 05.

---

# 40. Tool Installation Documentation

Ensure the required toolchain can be understood by a developer.

The project should make clear:

- protoc version
- protoc-gen-go version
- protoc-gen-go-grpc version
- grpc-gateway generator version if used

Prefer existing:

tools/versions.md

rather than introducing another version file.

---

# 41. Avoid Duplicate Version Sources

Do not create:

protobuf-versions.md

if:

tools/versions.md

already serves this purpose.

There should be one authoritative version source.

If multiple sources currently exist:

document the conflict.

---

# 42. Dependency Submodule

If googleapis is a git submodule:

verify:

git submodule status

Do not update the submodule.

Do not change its commit.

Do not modify its contents.

---

# 43. Fresh Checkout Simulation

If practical, verify the generation workflow using the repository's documented setup.

Do not delete local tooling.

Do not modify global developer configuration.

The goal is to determine whether the documented workflow is reproducible.

---

# 44. Clean Generation Check

Where practical:

1. generate protobuf code
2. inspect diff
3. generate again
4. confirm no new diff

Do not blindly delete generated files if they are tracked.

---

# 45. Detect Stale Generated Code

If generated code differs from the committed source:

determine whether:

- source proto changed
- generator version changed
- protoc version changed
- generation flags changed
- output path changed

Do not simply regenerate and commit without understanding why.

---

# 46. Handle Version Drift Carefully

If the generated code indicates a newer generator than the project's documented version:

do NOT immediately update the documented version.

Determine:

1. what version generated the committed file
2. what version local tooling uses
3. what version CI uses
4. what version the project documentation specifies

Then document the mismatch.

---

# 47. Correcting Generation Drift

A correction may be made only when the intended version is unambiguous.

For example:

If:

tools/versions.md

explicitly defines:

protoc-gen-go vX

and generation currently uses vY due to an incorrect local configuration,

correct the generation process to use vX.

Do not update project versions merely to match a local installation.

---

# 48. Do Not Upgrade Protobuf Dependencies

Do not perform broad upgrades such as:

go get -u

or:

go get google.golang.org/protobuf@latest

or:

go get google.golang.org/grpc@latest

unless an explicit documented requirement requires it.

---

# 49. Do Not Modify Business Services

Do not change:

clients/

transactions/

business logic merely because generated code exposes something you would implement differently.

Only make minimal import/package corrections if directly required to compile after correcting protobuf generation.

---

# 50. Do Not Implement HTTP Gateway

Even if the protobuf annotations suggest gateway work:

do not implement:

- HTTP server
- ServeMux
- gateway runtime
- gateway startup
- HTTP middleware

Agent 03 owns that.

---

# 51. Do Not Implement Observability

Do not add:

- metrics
- tracing
- health checks
- OpenTelemetry

Agent 09 owns observability.

---

# 52. Do Not Modify CI

Do not edit:

.github/workflows/

unless a tiny, directly necessary generation correction is unavoidable.

Prefer documenting CI changes for:

Agent 05.

---

# 53. Review Existing Findings

Read:

docs/platform-repository-audit.md

Identify all protobuf-related findings.

For every relevant finding:

- resolve it if this agent owns it
- otherwise document why it is being passed to a later agent

---

# 54. Create Review Document

Create:

docs/platform-protobuf-generation-review.md

Use exactly this structure:

# Platform Protobuf Generation Review

## 1. Objective

Explain the generation work performed.

## 2. Required Documentation

List every required document and whether it was read.

## 3. Existing Protobuf Structure

Describe the source and generated-code structure.

## 4. Toolchain

Document:

| Tool | Required Version | Actual Version | Status |
|---|---|---|---|
| protoc | | | |
| protoc-gen-go | | | |
| protoc-gen-go-grpc | | | |
| protoc-gen-grpc-gateway | | | |

## 5. Generation Commands

Document the canonical generation commands.

## 6. Source Protobuf Verification

Describe:

- package names
- go_package
- imports
- service definitions
- annotations

## 7. Generated Output

Document output locations.

## 8. Googleapis Dependency

Document how googleapis is resolved.

Do not reproduce the contents of the submodule.

## 9. Reproducibility

Document repeated-generation results.

## 10. Compilation Verification

List commands and results.

## 11. CI Compatibility

Document any CI concerns for Agent 05.

## 12. Findings

Use:

| ID | Severity | File/Area | Finding | Resolution |
|---|---|---|---|---|

## 13. Changes Made

List only files actually modified.

## 14. Deferred Work

List issues belonging to other agents.

## 15. Documentation Check

Record the final documentation verification.

## 16. Final Status

Use exactly one:

PASS

PASS WITH FOLLOW-UP

BLOCKED

---

# 55. Final Generation Verification

Run the canonical protobuf generation command one final time.

Then:

git status --short

Then:

git diff --stat

Then inspect relevant generated diffs.

Confirm:

- no unexpected generated files
- no unrelated source changes
- no stale output
- no repeated-generation drift

---

# 56. Final Compilation Verification

Run the narrowest relevant checks.

At minimum verify that generated protobuf packages compile.

Where practical verify:

- Clients
- Transactions

still compile against the generated output.

---

# 57. Final Documentation Check

Confirm the following were read:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md
- docs/platform-repository-audit.md

Record this in:

docs/platform-protobuf-generation-review.md

---

# 58. Final Scope Check

Run:

git status --short

Verify that modifications are limited to protobuf-generation-related files and:

docs/platform-protobuf-generation-review.md

If unrelated changes are present:

DO NOT revert them automatically.

Document them.

---

# 59. Final Completion Checklist

Before stopping:

- [ ] README.md was read.
- [ ] agents/project-context.md was read.
- [ ] docs/domain-model.md was read.
- [ ] docs/repository-layout.md was read.
- [ ] docs/protobuf-strategy.md was read.
- [ ] docs/migration-plan.md was read.
- [ ] docs/platform-repository-audit.md was read.
- [ ] Protobuf source files were identified.
- [ ] Active protobuf services were identified.
- [ ] Package names were verified.
- [ ] go_package values were verified.
- [ ] Imports were verified.
- [ ] googleapis dependency was verified.
- [ ] googleapis was NOT recursively inspected.
- [ ] protoc version was verified.
- [ ] protoc-gen-go version was verified.
- [ ] protoc-gen-go-grpc version was verified.
- [ ] grpc-gateway generator version was verified where applicable.
- [ ] Generation commands were verified.
- [ ] Output paths were verified.
- [ ] Generated code was regenerated.
- [ ] Repeated generation was checked for determinism.
- [ ] Generated code was not manually edited.
- [ ] Generated packages compile.
- [ ] Relevant service packages compile.
- [ ] Version drift was investigated.
- [ ] No unnecessary dependency upgrades were performed.
- [ ] No business-service redesign was performed.
- [ ] No HTTP gateway implementation was performed.
- [ ] No CI redesign was performed.
- [ ] docs/platform-protobuf-generation-review.md was created.
- [ ] Final git state was checked.
- [ ] Final documentation check was recorded.

---

# Final Stop Condition

STOP after:

1. verifying the protobuf architecture,
2. correcting only protobuf-generation issues owned by this agent,
3. successfully generating the protobuf artifacts,
4. verifying repeated generation is deterministic,
5. verifying generated code compiles,
6. documenting any CI follow-up,
7. creating docs/platform-protobuf-generation-review.md,
8. completing the documentation check,
9. checking final git status.

Do NOT proceed to:

- HTTP gateway implementation
- CI/CD redesign
- Docker
- Render
- observability
- security remediation
- performance optimization
- Clients business logic
- Transactions business logic
- database redesign
- provider integrations

Those belong to later agents.

STOP.