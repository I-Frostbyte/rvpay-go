# Agent 01 — Platform Repository Audit

## Objective

Perform a focused audit of the existing RVPay repository before any Platform implementation work begins.

This agent does NOT implement the Platform layer.

Its purpose is to establish an accurate baseline of:

- the current repository structure
- existing shared infrastructure
- protobuf generation
- HTTP/gRPC gateway infrastructure
- common packages
- CI/CD
- Docker
- Render deployment
- documentation
- observability
- security
- performance-related infrastructure
- existing conventions
- existing implementation decisions

The output of this agent will be used by Platform Agents 02–12.

Create the final audit document at:

docs/platform-repository-audit.md

Do not create new application architecture during this task.

Do not redesign the repository.

Do not modify Clients or Transactions.

---

# Required Reading

Read only:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md

Also inspect the existing Platform-related documentation if it exists.

If a referenced document does not exist:

- do not create it
- do not invent its contents
- record that it is missing

---

# Documentation Check

Before beginning the audit, confirm that the required documents are present and readable.

Required:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md

At the end of the task, perform the documentation check again.

Record the result in:

docs/platform-repository-audit.md

---

# Repository Exploration Rules

## IMPORTANT

Do NOT perform an unrestricted recursive repository scan.

The repository is large and contains generated code, submodules, dependencies, and other directories that are irrelevant to this audit.

Use:

README.md

as the primary map.

Use:

agents/project-context.md

for coding and package conventions.

Use:

docs/repository-layout.md

for expected repository structure.

Use:

docs/protobuf-strategy.md

for protobuf and generated-code expectations.

Use:

docs/migration-plan.md

for migration and compatibility expectations.

---

# Directories You MUST NOT Explore Recursively

Do not recursively inspect:

- .git/
- third_party/
- third_party/googleapis/
- vendor/
- node_modules/
- coverage/
- tmp/
- bin/
- generated dependency directories

Do not descend deeply into generated protobuf output.

Do not inspect every generated SQLC file.

Do not inspect every mock.

Do not inspect unrelated historical or archived material.

If a specific file in one of these areas is directly referenced by an active configuration or build process, inspect only that specific file when necessary.

---

# Repository Exploration Strategy

Start from:

README.md

Determine:

1. repository root structure
2. service locations
3. protobuf locations
4. generated-code locations
5. database locations
6. configuration locations
7. build files
8. CI files
9. Docker files
10. deployment files
11. documentation locations

Then inspect only the files relevant to those areas.

---

# 1. Establish Git State

Run:

git status --short

Then:

git branch --show-current

Then:

git log -5 --oneline

Do not modify anything.

Record only information useful to understanding the current implementation state.

---

# 2. Establish Repository Baseline

Read:

README.md

carefully.

Determine:

- what RVPay currently is
- what services exist
- which services are active
- which services are being introduced
- which services were migrated from the older implementation
- how the repository expects developers to work with the project

Do not reinterpret the architecture.

The README is the repository map.

---

# 3. Compare README With Actual Layout

Compare the documented structure against the actual root directory.

Identify:

- documented directories that exist
- documented directories that no longer exist
- directories that exist but are not documented
- files that have moved
- infrastructure that is undocumented

Do not fix the README.

Record discrepancies in the audit.

---

# 4. Platform Scope

Establish which repository components belong to the Platform layer.

At minimum investigate whether the repository contains:

- protobuf source definitions
- generated protobuf Go code
- gRPC gateway code
- common packages
- shared configuration
- logging
- middleware
- error handling
- CI workflows
- Docker infrastructure
- Render configuration
- observability
- security configuration
- repository-wide Makefiles

Do not assume every shared package belongs in Platform.

Use the documentation and existing implementation to determine ownership.

---

# 5. Service Boundary Audit

Inspect the existing service directories only at a reasonable depth.

The known services include:

- clients
- transactions

and potentially older services or supporting components.

Determine:

- what each service owns
- what each service shares
- whether services directly import each other's implementation packages
- whether services rely on shared packages
- whether service boundaries are respected

Do not refactor service boundaries.

---

# 6. Clients and Transactions Relationship

Because Clients and Transactions are the primary application services being introduced:

inspect only enough of each service to determine how they depend on Platform functionality.

Look for:

- protobuf imports
- shared packages
- logging
- configuration
- middleware
- database helpers
- HTTP gateway
- common errors
- authentication infrastructure

Do not audit the business logic of Clients or Transactions.

That belongs to their dedicated agents.

---

# 7. Existing Shared Packages

Locate shared packages.

Examples may include:

- internal/
- pkg/
- common/
- middleware/
- config/
- logging/
- errors/
- auth/

Do not assume these packages should be retained.

For each relevant package determine:

- purpose
- importers
- ownership
- whether it is generic
- whether it contains business logic
- whether it is service-specific code incorrectly placed in a shared package

Do not move anything.

---

# 8. Shared Package Dependency Direction

Check dependency direction.

The desired principle is:

business services
→ shared platform infrastructure

rather than:

shared platform infrastructure
→ business service implementation

Look for obvious violations.

Do not redesign the dependency graph.

Document findings.

---

# 9. Go Module Review

Inspect:

go.mod

and, where relevant:

go.sum

Determine:

- Go version
- important shared dependencies
- protobuf dependencies
- gRPC dependencies
- gateway dependencies
- logging dependencies
- database-related shared dependencies
- testing dependencies

Do not upgrade dependencies.

Do not run a broad dependency upgrade.

---

# 10. Existing Makefiles

Locate:

- root Makefile
- service Makefiles
- protobuf Makefile
- other active build files

Determine:

- generation commands
- test commands
- build commands
- Docker commands
- migration commands
- lint commands
- formatting commands

Do not modify Makefiles.

Record existing conventions.

---

# 11. Protobuf Source Inventory

Locate the source protobuf definitions.

Use:

docs/protobuf-strategy.md

to understand their intended ownership.

Identify:

- source `.proto` files
- service definitions
- shared messages
- imports
- package declarations
- HTTP annotations
- googleapis dependencies

Do NOT recursively inspect:

third_party/googleapis/

Only verify that the required dependency exists and is referenced appropriately.

---

# 12. Protobuf Generated-Code Inventory

Locate generated Go protobuf files.

Identify:

- generated protobuf output
- generated gRPC output
- generated gateway output if present

Determine:

- output locations
- generation commands
- whether generated files are committed
- whether generated code is referenced by services

Do not modify generated files.

---

# 13. Protobuf Toolchain Inventory

Determine the currently configured versions of:

- protoc
- protoc-gen-go
- protoc-gen-go-grpc
- protoc-gen-grpc-gateway

Use existing project documentation such as:

tools/versions.md

if present.

Do not assume versions from general knowledge.

Do not upgrade the toolchain.

---

# 14. Protobuf Generation Workflow

Identify exactly how protobuf generation currently happens.

Look for:

- Makefiles
- go generate directives
- shell scripts
- CI commands
- documentation

Determine whether there are multiple competing generation mechanisms.

Do not consolidate them.

Document the current situation.

---

# 15. HTTP Gateway Inventory

Determine whether the repository currently uses:

- grpc-gateway
- handwritten HTTP handlers
- reverse proxying
- another HTTP layer

Identify:

- gateway source
- generated gateway code
- registration
- routing
- server startup

Do not implement the gateway.

That belongs to:

agents/platform/03-http-gateway.md

---

# 16. CI/CD Inventory

Inspect:

.github/workflows/

BUT:

do not recursively inspect unrelated historical workflows.

Identify active workflows based on:

- README.md
- filenames
- workflow triggers

Document:

- Go version
- generation steps
- tests
- Docker build
- deployment
- secrets
- Render integration

Do not modify CI.

---

# 17. CI Dependency on Generated Code

Determine whether CI verifies generated code.

Look for patterns such as:

git diff --exit-code

after generation.

Record:

- what gets generated
- what gets verified
- which versions are pinned

Do not change the workflow.

---

# 18. Docker Inventory

Locate active Dockerfiles.

Determine:

- service image structure
- build context
- build stage
- runtime stage
- binary location
- entrypoint
- environment assumptions
- port assumptions

Do not optimize Dockerfiles yet.

That belongs to:

agents/platform/06-docker.md

---

# 19. Render Inventory

Locate existing Render-related configuration.

Examples:

- render.yaml
- deployment documentation
- deploy hooks
- environment references

Determine:

- whether Render is currently used
- which services are deployed
- whether deployment is per-service
- whether Blueprint is used
- whether Docker is used

Do not modify Render configuration.

That belongs to:

agents/platform/07-render.md

---

# 20. Environment Configuration Inventory

Locate:

- `.env.example`
- service-specific environment templates
- configuration packages
- deployment environment documentation

Determine:

- naming conventions
- configuration ownership
- required variables
- service-specific variables
- shared variables

Do not add variables.

Do not create secrets.

---

# 21. Logging Inventory

Determine the existing logging approach.

Identify:

- logger package
- logger initialization
- log format
- structured logging
- log levels
- request logging
- error logging

Determine whether services use a consistent logger.

Do not replace the logger.

---

# 22. Error Handling Inventory

Determine how the repository handles:

- domain errors
- repository errors
- gRPC errors
- HTTP errors
- validation errors
- provider errors

Look for shared error packages.

Determine whether error mapping is centralized or duplicated.

Do not redesign error handling.

---

# 23. Middleware Inventory

Identify existing:

- gRPC interceptors
- HTTP middleware
- recovery handlers
- logging middleware
- authentication middleware
- request ID handling

Determine where middleware is registered.

Do not implement new middleware.

---

# 24. Authentication Inventory

Determine whether the repository already contains authentication infrastructure.

Look for:

- authentication middleware
- token validation
- API keys
- OAuth-related shared components
- authorization checks

Do not implement authentication.

Do not inspect production credentials.

---

# 25. Observability Inventory

Determine whether the repository already contains:

- metrics
- tracing
- health checks
- readiness checks
- structured logs
- request IDs
- error reporting

Document what exists.

Do not install an observability platform.

---

# 26. Security Configuration Inventory

Inspect only active relevant configuration for:

- secrets handling
- environment variables
- Docker secrets assumptions
- CI secrets
- Render secrets
- TLS configuration
- authentication configuration

Do not perform a full penetration test.

Do not expose any secret values.

---

# 27. Performance Infrastructure Inventory

Determine whether the repository already uses:

- connection pooling
- caching
- HTTP client pooling
- gRPC connection reuse
- pagination
- indexes
- background workers

Do not benchmark anything in this agent.

Do not optimize anything.

Record existing mechanisms.

---

# 28. Database Infrastructure Boundary

Transactions and Clients own their application databases.

Determine whether there are shared database utilities.

Examples:

- connection helpers
- migration runners
- pool configuration
- common database interfaces

Do not move database ownership.

Do not create a shared database abstraction merely for convenience.

---

# 29. Generated Files Policy

Determine which generated artifacts are committed.

Examples:

- protobuf
- gRPC
- gateway
- SQLC
- mocks

Document the project's current policy.

Do not change it.

---

# 30. Documentation Inventory

Locate relevant:

- README files
- docs/
- architecture documents
- agent documents
- tool version documentation

Do not read every markdown file in the repository.

Only inspect documents relevant to Platform.

---

# 31. Existing Platform Problems

Identify concrete problems that later Platform agents must address.

Examples:

- broken protobuf generation
- duplicated gateway configuration
- inconsistent shared packages
- CI generation mismatch
- inconsistent Docker setup
- missing deployment configuration
- logging inconsistencies
- missing observability
- security configuration gaps
- performance bottlenecks

Do not fix them in this agent.

---

# 32. Do Not Treat Every Difference as a Problem

A difference from the desired architecture is only a finding if it is supported by:

- README.md
- project-context.md
- domain documentation
- repository-layout.md
- protobuf-strategy.md
- migration-plan.md
- existing explicit project requirements

Do not invent requirements.

---

# 33. Do Not Review Deep Folders

If you encounter a directory that appears unrelated:

STOP.

Use the README and documentation to determine whether it matters.

Do not explore it merely because it exists.

This is especially important for:

- googleapis
- generated code
- dependency trees
- historical files
- test fixtures
- unrelated services

---

# 34. No Code Changes

This agent is an audit.

Do not:

- create source code
- modify Go files
- modify protobuf files
- modify SQL
- modify migrations
- modify Dockerfiles
- modify Makefiles
- modify CI
- modify Render configuration
- modify service implementation

The only file this agent should create is:

docs/platform-repository-audit.md

If another change appears necessary:

document it.

---

# 35. Audit Findings Must Be Evidence-Based

Every finding should contain:

- location
- evidence
- impact
- recommendation for the appropriate later agent

Do not write:

"CI could be better."

Instead write:

"`.github/workflows/render-deploy.yml` installs protoc-gen-go-grpc version X while generated output indicates version Y. This may cause generated-code drift. Agent 02/05 should verify the toolchain."

Use exact file paths where possible.

---

# 36. Classify Findings

Use:

### BLOCKER

Prevents the Platform layer from functioning or causes serious correctness/security issues.

### HIGH

Likely to cause production failure or major integration problems.

### MEDIUM

Important but not immediately blocking.

### LOW

Minor issue or maintainability concern.

### INFORMATIONAL

Observation that later agents should be aware of.

---

# 37. Assign Findings to Later Agents

Where possible, associate findings with the appropriate Platform agent.

Examples:

- protobuf issue → Agent 02
- gateway issue → Agent 03
- common package issue → Agent 04
- CI issue → Agent 05
- Docker issue → Agent 06
- Render issue → Agent 07
- documentation issue → Agent 08
- observability issue → Agent 09
- security issue → Agent 10
- performance issue → Agent 11

Do not assign work to an agent merely to create a task.

Only assign concrete findings.

---

# 38. Platform Dependency Map

Create a dependency map showing:

Clients
↓
Platform components

Transactions
↓
Platform components

For each shared dependency identify:

- package/path
- purpose
- current implementation
- whether it is stable
- whether a later Platform agent needs to modify it

---

# 39. Platform Risk Register

Create:

| ID | Area | Severity | Evidence | Affected Component | Responsible Agent |
|---|---|---|---|---|---|

Every significant Platform finding should appear here.

---

# 40. Platform Implementation Inventory

Create a table:

| Component | Exists? | Location | Current State | Later Agent |
|---|---|---|---|---|
| Protobuf generation | | | | 02 |
| HTTP gateway | | | | 03 |
| Common packages | | | | 04 |
| CI/CD | | | | 05 |
| Docker | | | | 06 |
| Render | | | | 07 |
| Documentation | | | | 08 |
| Observability | | | | 09 |
| Security | | | | 10 |
| Performance | | | | 11 |

Use actual repository evidence.

---

# 41. Expected Platform Structure

Do NOT invent the final Platform structure.

Instead:

compare the existing structure against:

docs/repository-layout.md

and identify where Platform responsibilities currently live.

The later Platform agents will implement according to the documented architecture.

---

# 42. Cross-Service Consistency

Compare Clients and Transactions only where necessary to identify Platform-level inconsistencies.

Examples:

- different logging setup
- different protobuf package conventions
- different server startup patterns
- different environment naming
- different Docker conventions

Do not fix those differences here.

---

# 43. Existing Conventions Take Priority

If the repository already has a working convention:

do not replace it merely because another convention is theoretically better.

Follow:

agents/project-context.md

and the existing implementation.

---

# 44. Do Not Introduce New Dependencies

Do not run:

go get

for Platform work.

Do not modify:

go.mod

or:

go.sum

unless required merely to inspect or verify something.

This agent is read/audit only.

---

# 45. Do Not Run Deployment

Do not:

- deploy to Render
- trigger deployment hooks
- modify cloud infrastructure
- modify production databases
- modify production secrets

---

# 46. Do Not Use Production Credentials

Never request or use:

- production database credentials
- Render credentials
- API keys
- OAuth secrets
- provider credentials

The audit must be performed using repository contents and local/test configuration only.

---

# 47. Audit Report

Create:

docs/platform-repository-audit.md

Use exactly this structure:

# Platform Repository Audit

## 1. Executive Summary

Summarize the current state of Platform infrastructure.

## 2. Scope

Explain what was inspected and what was deliberately excluded.

## 3. Required Documentation

List:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md

Record whether each was successfully read.

## 4. Repository Baseline

Describe the current repository layout.

## 5. Platform Component Inventory

Use the Platform Implementation Inventory table.

## 6. Service Boundary Findings

Describe relevant Clients/Transactions → Platform relationships.

## 7. Protobuf Infrastructure

Describe current source, generation, and generated-code structure.

## 8. HTTP Gateway

Describe the current gateway infrastructure.

## 9. Common Packages

Describe shared package structure and dependency direction.

## 10. CI/CD

Describe current workflows and generation/build/test behavior.

## 11. Docker

Describe current Docker conventions.

## 12. Render

Describe current Render deployment structure.

## 13. Documentation

Describe current Platform-related documentation.

## 14. Observability

Describe current logging, metrics, tracing, and health infrastructure.

## 15. Security

Describe current security-related infrastructure.

Do not include secret values.

## 16. Performance

Describe existing performance-related infrastructure.

Do not perform benchmarking.

## 17. Cross-Service Consistency

Document relevant inconsistencies between Clients and Transactions.

## 18. Findings

Use:

| ID | Severity | Area | Evidence | Impact | Responsible Agent |
|---|---|---|---|---|---|

## 19. Platform Risk Register

Include the risk register.

## 20. Recommended Agent Order

Provide the dependency-aware order for Agents 02–12.

Do not redesign the sequence unless a concrete dependency requires it.

## 21. Out-of-Scope Areas

Explicitly list directories and areas deliberately not inspected.

## 22. Documentation Check

Record the final documentation verification.

## 23. Final Repository State

Record:

git status --short

and relevant observations.

---

# 48. Recommended Agent Order

Unless the repository reveals a concrete dependency requiring otherwise, use:

1. Repository Audit
2. Protobuf Generation
3. HTTP Gateway
4. Common Packages
5. CI/CD
6. Docker
7. Render
8. Documentation
9. Observability
10. Security
11. Performance
12. Final Review

Do not execute any of those agents.

This agent only prepares the information they need.

---

# 49. Final Validation

Before completing:

run:

git status --short

Confirm that the only intended new file is:

docs/platform-repository-audit.md

If other files were modified accidentally:

DO NOT automatically revert them.

Report them.

---

# 50. Final Completion Checklist

Before stopping:

- [ ] README.md was read.
- [ ] agents/project-context.md was read.
- [ ] docs/domain-model.md was read.
- [ ] docs/repository-layout.md was read.
- [ ] docs/protobuf-strategy.md was read.
- [ ] docs/migration-plan.md was read.
- [ ] Repository structure was audited.
- [ ] Platform responsibilities were identified.
- [ ] Clients/Transactions dependencies were inspected only where relevant.
- [ ] Protobuf infrastructure was inventoried.
- [ ] HTTP gateway was inventoried.
- [ ] Common packages were inventoried.
- [ ] CI/CD was inventoried.
- [ ] Docker was inventoried.
- [ ] Render was inventoried.
- [ ] Documentation was inventoried.
- [ ] Observability was inventoried.
- [ ] Security infrastructure was inventoried.
- [ ] Performance infrastructure was inventoried.
- [ ] No deep irrelevant directories were explored.
- [ ] third_party/googleapis was not recursively explored.
- [ ] Generated code was not unnecessarily inspected line-by-line.
- [ ] No source code was modified.
- [ ] No dependencies were upgraded.
- [ ] No deployment was performed.
- [ ] No production credentials were used.
- [ ] Findings were evidence-based.
- [ ] Findings were assigned to appropriate later agents.
- [ ] docs/platform-repository-audit.md was created.
- [ ] Final git state was checked.
- [ ] Final documentation check was recorded.

---

# Final Stop Condition

STOP after:

1. completing the repository audit,
2. creating docs/platform-repository-audit.md,
3. documenting all relevant findings,
4. assigning findings to the appropriate Platform agents,
5. completing the final documentation check,
6. checking git status.

Do NOT:

- implement Platform components
- modify existing services
- modify protobufs
- modify SQL
- modify CI
- modify Docker
- modify Render
- add dependencies
- deploy anything
- perform performance optimization
- perform security remediation
- redesign the architecture

This agent establishes the baseline.

The subsequent Platform agents will perform the implementation work.