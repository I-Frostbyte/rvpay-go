# Agent 11 — Transactions Service Scaffolding

## Objective

Complete the operational scaffolding for the Transactions microservice.

Agents 02–10 have established the Transactions:

- database
- SQLC layer
- protobuf contracts
- repositories
- Merchant capability
- Customer capability
- Deposit capability
- Payout capability
- runtime
- gRPC server

This agent makes the Transactions service practical to:

- build
- run
- test
- generate code
- inspect locally
- operate consistently with the existing repository

The implementation must follow the existing Deposits service conventions wherever applicable.

This agent is NOT an architectural redesign agent.

Do not rewrite working Transactions components.

---

# Required Reading

Read only:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md
- docs/transactions-existing-review.md
- docs/transactions-database-review.md
- docs/transactions-sqlc-review.md
- docs/transactions-protobuf-review.md
- docs/transactions-repository-review.md
- docs/transactions-merchants-review.md
- docs/transactions-customers-review.md
- docs/transactions-deposits-review.md
- docs/transactions-payouts-review.md
- docs/transactions-runtime-review.md

Additionally, inspect only the corresponding existing Deposits scaffolding files needed to reproduce project conventions.

---

# Documentation Check

Before modifying anything, confirm that every required document above has been read.

At completion, perform the documentation check again.

The final review document must explicitly record the documentation check.

---

# Repository Exploration Rules

Use README.md as the repository map.

Perform focused exploration only.

Do NOT perform unrestricted recursive searches.

Do NOT recursively inspect:

- third_party/
- third_party/googleapis/
- vendor/
- node_modules/
- .git/
- coverage/
- tmp/
- bin/

Do not inspect unrelated services.

The primary implementation reference is the existing Deposits service.

Only inspect Deposits files necessary to understand:

- Makefile conventions
- Dockerfile conventions
- README conventions
- environment-variable documentation
- generation commands
- local development commands

---

# 1. Establish Existing Deposits Scaffolding

Locate the existing Deposits:

- Dockerfile
- Makefile
- README.md
- `.env.example`
- generation targets
- build targets
- test targets

Inspect only the relevant files.

Do not recursively explore the entire Deposits service.

---

# 2. Compare Transactions Runtime

Read:

docs/transactions-runtime-review.md

Determine the actual runtime entry point.

Use that runtime as the authoritative executable.

Do not assume the runtime location from memory.

---

# 3. Determine Transactions Root

Follow:

docs/repository-layout.md

The Transactions service root should contain the operational files required by the project convention.

Expected examples may include:

transactions/
├── cmd/
├── config/
├── db/
├── merchants/
├── customers/
├── deposits/
├── payouts/
├── Dockerfile
├── Makefile
├── README.md
└── .env.example

Use the actual documented structure.

Do not create directories merely because they appear in this example.

---

# 4. Dockerfile

Create or complete the Transactions Dockerfile.

Model it after the existing Deposits Dockerfile.

Do not blindly copy it.

Verify:

- module path
- build context
- command package
- binary output
- runtime path
- required files
- environment expectations

---

# 5. Docker Build Context

Determine the correct Docker build context from the existing repository.

If the Deposits Dockerfile is built from the repository root:

Transactions must follow the same pattern unless there is a documented reason not to.

Do not assume the Dockerfile's own directory is the build context.

---

# 6. Docker Multi-Stage Build

If Deposits uses a multi-stage build:

follow that convention.

The final image should contain only what the Transactions runtime requires.

Do not add development tools to the final runtime image.

---

# 7. Docker Go Version

Use the Go version established by the repository.

Check:

- go.mod
- existing Dockerfile
- CI configuration
- project documentation

Do not introduce a different Go version.

---

# 8. Docker Generated Code

Determine whether generated protobuf and SQLC output is committed to the repository.

Follow the existing Deposits convention.

Do not create a new generation strategy.

If generation happens before Docker build:

the Dockerfile should consume the generated output.

Do not make Docker silently mutate the Git working tree.

---

# 9. Docker Entrypoint

The container must execute the Transactions runtime.

Use the actual command package from:

docs/transactions-runtime-review.md

Do not run:

- go test
- protoc
- sqlc
- make

as the production container command.

---

# 10. Docker Port

Ensure the Docker configuration corresponds to the Transactions gRPC runtime port.

Do not hard-code a port that conflicts with configuration.

If the repository documents an EXPOSE convention:

follow it.

Remember:

EXPOSE does not publish a port.

Do not confuse Docker metadata with runtime binding.

---

# 11. Docker Environment

Do not place secrets in the Dockerfile.

Do not copy:

- `.env`
- credentials
- API keys
- private certificates

into the image.

Use runtime environment variables.

---

# 12. `.dockerignore`

If the repository uses `.dockerignore`:

follow the existing convention.

If it does not:

do not create a broad unrelated repository-wide cleanup.

If a Transactions-specific `.dockerignore` is required by the existing structure:

create it with focused exclusions.

Do not exclude files required for:

- protobuf generation
- SQLC generation
- Go compilation
- submodules required by the build

unless generation happens before Docker build and the generated output is committed.

---

# 13. Makefile

Create or complete:

transactions/Makefile

Model it after the Deposits Makefile.

Do not invent an unrelated command naming scheme.

---

# 14. Makefile Responsibilities

The Transactions Makefile should provide the documented operational tasks required for the service.

Potential responsibilities include:

- build
- run
- test
- generate protobuf
- generate SQLC
- generate mocks
- migration commands
- Docker build

Only include commands that actually apply to Transactions.

---

# 15. Build Target

Provide a build target consistent with Deposits.

The target should build the Transactions runtime.

Do not compile unrelated services.

Use the actual command package.

---

# 16. Run Target

Provide a local development run target.

It should run the Transactions gRPC service.

Follow Deposits conventions.

Do not create an additional development server.

---

# 17. Test Target

Provide a Transactions test target.

It should run the tests relevant to Transactions.

Do not silently omit packages.

Do not make tests depend on production credentials.

---

# 18. Protobuf Generation

If protobuf generation is owned by the repository-level protobuf Makefile:

do not duplicate the entire protobuf generation system inside Transactions.

Instead, reference the existing generation mechanism appropriately.

Follow:

docs/protobuf-strategy.md

Do not manually maintain generated protobuf commands in multiple unrelated locations.

---

# 19. SQLC Generation

If Transactions SQLC generation is service-local:

provide the appropriate Makefile target.

Use the version established by the project.

Do not install a globally floating SQLC version.

Do not modify generated SQLC files manually.

---

# 20. Mock Generation

If Transactions repositories use generated mocks:

follow the project's existing mock generation strategy.

Do not introduce another mocking library.

Use the pinned version from project documentation.

---

# 21. Migration Commands

Only expose migration commands if the Transactions architecture actually supports developer-controlled migrations.

Follow:

docs/migration-plan.md

Do not create destructive commands by default.

If migration commands exist:

document them clearly.

---

# 22. Migration Safety

Never make the default:

make migration-down

or another destructive operation.

Destructive operations should require an explicit target.

Do not invent a migration system.

---

# 23. `.env.example`

Create or complete:

transactions/.env.example

if the repository structure calls for a service-local environment template.

Follow the actual Transactions configuration implementation.

---

# 24. Environment Variables

Only document environment variables that the Transactions configuration actually consumes.

Do not invent:

- fake variables
- aliases
- unused settings

Each variable must correspond to actual code.

---

# 25. Secret Placeholders

Use safe placeholders.

For example:

DATABASE_URL=

Do not insert real credentials.

Do not copy credentials from any existing `.env`.

---

# 26. Production Environment

Clearly distinguish between:

- local development values
- production-provided values

Do not hard-code Render-specific credentials or deployment secrets.

---

# 27. README

Create or complete:

transactions/README.md

Model it after the Deposits README.

The Transactions README must describe the actual implemented service.

Do not write speculative documentation.

---

# 28. README Structure

Include, where applicable:

# Transactions Service

## Purpose

## Responsibilities

## Service Structure

## Running Locally

## Configuration

## Database

## Code Generation

## Testing

## Docker

## gRPC API

## REST API

## Migrations

## Architecture Notes

## Troubleshooting

Use the existing project terminology.

---

# 29. README Purpose

Clearly explain that Transactions owns the documented transaction capabilities.

At minimum document the capabilities actually implemented:

- Merchants
- Customers
- Deposits
- Payouts

Do not document capabilities that have not been implemented.

---

# 30. README Architecture

Describe the actual dependency flow.

For example:

gRPC API
    ↓
Transactions Services
    ↓
Repositories
    ↓
PostgreSQL

Include the actual structure only.

Do not invent additional infrastructure.

---

# 31. README Configuration

Document every required configuration variable.

For each variable explain:

- purpose
- required/optional status
- local development expectations

Do not document secrets themselves.

---

# 32. README Database

Document:

- PostgreSQL requirement
- database configuration
- migration behavior
- SQLC usage

Do not include real database credentials.

---

# 33. README Code Generation

Document how developers regenerate:

- protobuf output
- SQLC output
- mocks

Use the actual project commands.

Do not invent commands that do not exist.

---

# 34. README Testing

Document the actual test commands.

Include focused Transactions tests.

Include repository-wide tests only if that is the established project convention.

---

# 35. README Docker

Document:

- Docker build command
- image name example
- container run requirements
- required environment variables

Do not include secrets.

---

# 36. README API

Document the available gRPC API at a high level.

If REST/gRPC-gateway is implemented:

document the available HTTP exposure.

Do not manually reproduce generated protobuf definitions.

---

# 37. README Troubleshooting

Include only realistic troubleshooting information.

Potential categories:

- database connection failure
- missing environment variables
- protobuf generation
- SQLC generation
- Docker build
- port conflicts

Do not write speculative troubleshooting.

---

# 38. Makefile Documentation

Every non-obvious Makefile target should be understandable from the README.

Do not create undocumented operational commands.

---

# 39. Root Makefile Integration

Inspect the root Makefile.

Determine whether Transactions should have root-level targets.

If the existing architecture expects services to be invoked from the root Makefile:

add only the required Transactions targets.

Do not redesign the root Makefile.

---

# 40. Root README Integration

Inspect the root README.

Determine whether the new Transactions service needs to be added to the service map.

If the root README already documents service locations:

update it only where necessary.

Do not rewrite unrelated documentation.

---

# 41. Environment Template Integration

Inspect the root `.env.example`.

If the project convention uses one root environment template:

follow it.

Do not create duplicate variables with conflicting names.

If services use service-local environment templates:

follow that convention instead.

---

# 42. Code Generation Ownership

Establish exactly which layer owns each generation task.

For example:

| Generated Artifact | Owner |
|---|---|
| protobuf | protobuf Makefile |
| gRPC gateway | protobuf generation |
| SQLC | Transactions DB |
| mocks | Transactions repository |

Use the actual project architecture.

Do not duplicate ownership.

---

# 43. No Generated File Editing

Never manually edit:

- `.pb.go`
- `_grpc.pb.go`
- gateway generated files
- SQLC generated files
- generated mocks

If generated output is incorrect:

fix the source/configuration/generation command.

---

# 44. No Architecture Changes

Do not:

- merge services
- split services
- rename domains
- redesign repositories
- change protobuf contracts
- redesign database schema

Those decisions have already been made.

---

# 45. No Provider Work

Do not add:

- payment provider clients
- banking integrations
- mobile-money integrations
- webhook processors

This agent is scaffolding only.

---

# 46. No OAuth

Do not implement OAuth.

Do not add OAuth configuration.

OAuth belongs to the Clients service.

---

# 47. No Runtime Redesign

Agent 10 owns the runtime.

If scaffolding reveals a runtime issue:

make the smallest compatibility fix necessary.

Do not redesign `main.go` or `run()`.

If the problem requires architectural changes:

document it for later review.

---

# 48. Local Development Workflow

Ensure the service can reasonably support:

1. configure environment
2. start PostgreSQL
3. run migrations if required
4. generate code
5. run tests
6. start Transactions service

Document this workflow in the README.

---

# 49. Docker Workflow

Ensure the documented Docker workflow is consistent.

The expected workflow should resemble:

docker build
→ configure environment
→ run container
→ connect to PostgreSQL
→ start Transactions

Use the actual Dockerfile and runtime.

---

# 50. Makefile Consistency

Compare Transactions Makefile with Deposits.

Check:

- naming
- indentation
- shell behavior
- variable conventions
- command style
- target naming
- help output if present

Follow existing style.

---

# 51. README Consistency

Compare Transactions README with Deposits README.

Preserve:

- writing style
- command formatting
- architecture explanation style
- section organization

Do not copy Deposits-specific facts.

---

# 52. Dockerfile Consistency

Compare Transactions Dockerfile with Deposits Dockerfile.

Preserve:

- build stages
- Go environment
- binary strategy
- runtime image strategy
- non-root behavior if used
- working directory conventions

Change only what is required for Transactions.

---

# 53. File Ownership

The following files belong to this agent's scope:

- transactions/Dockerfile
- transactions/Makefile
- transactions/README.md
- transactions/.env.example

Potentially:

- root Makefile
- root README.md
- root .env.example

only where directly required by existing repository conventions.

---

# 54. Do Not Create Extra Documentation

Do not create additional architecture documents unless necessary.

The required review document is:

docs/transactions-scaffolding-review.md

---

# 55. Validation — Makefile

Verify the Makefile targets actually work.

At minimum test applicable targets for:

- build
- test
- generation
- run/help where safe

Do not execute destructive migration commands merely to test the Makefile.

---

# 56. Validation — Docker

Build the Transactions Docker image if Docker is available.

Use the actual repository build context.

Do not push the image anywhere.

Do not deploy.

---

# 57. Validation — Docker Runtime

If practical, verify that the container starts with appropriate local environment configuration.

Do not require production credentials.

If PostgreSQL is unavailable:

record that limitation.

---

# 58. Validation — README Commands

Every command documented in the README should correspond to an actual Makefile command or repository command.

Remove commands that do not exist.

---

# 59. Validation — Environment

Verify that every variable documented in `.env.example` is actually consumed by configuration code.

Remove unused variables.

Do not add speculative configuration.

---

# 60. Validation — Generation

Verify the documented generation commands are consistent with:

docs/protobuf-strategy.md

and:

docs/transactions-sqlc-review.md

Do not regenerate unrelated services.

---

# 61. Full Test Suite

If practical:

go test ./...

Do not fix unrelated failures.

If failures are unrelated:

record them.

---

# 62. Git Review

Before completion run:

git status --short

Then:

git diff --stat

Then inspect the relevant diff.

Confirm there are no unexpected generated changes.

---

# 63. Scope Enforcement

Expected changes should generally be limited to:

- transactions/Dockerfile
- transactions/Makefile
- transactions/README.md
- transactions/.env.example
- directly required root documentation/Makefile updates
- docs/transactions-scaffolding-review.md

Do not modify:

- Clients service
- Deposit business logic
- Payout business logic
- Merchant business logic
- Customer business logic
- protobuf definitions
- generated protobuf output
- SQLC generated output
- migrations
- repository interfaces
- OAuth
- webhooks
- provider integrations
- deployment workflows
- third_party/

---

# 64. Existing Code Protection

Do not overwrite working files merely to make them stylistically identical.

Preserve implementation work from Agents 02–10.

Only make changes directly required to expose the service operationally.

---

# 65. Review Document

Create:

docs/transactions-scaffolding-review.md

This document is mandatory.

---

# 66. Required Review Document Structure

Use exactly:

# Transactions Scaffolding Implementation Review

## 1. Source Documents

List every document read.

## 2. Existing Deposits Scaffolding

Document the Deposits files inspected.

## 3. Transactions Structure

Document the actual Transactions directory structure.

## 4. Dockerfile

Document:

- build context
- build stages
- Go version
- binary
- runtime image
- entrypoint

## 5. Makefile

Use:

| Target | Purpose |
|---|---|

## 6. Environment

Document:

| Variable | Required | Purpose |
|---|---|---|

Do not include secret values.

## 7. README

Document the sections created/updated.

## 8. Code Generation

Document:

| Artifact | Command | Owner |
|---|---|---|

## 9. Local Workflow

Document:

configuration
→ generation
→ migration
→ test
→ run

using the actual project workflow.

## 10. Docker Workflow

Document:

build
→ configure
→ run

## 11. Root Integration

Document any changes made to:

- root README
- root Makefile
- root `.env.example`

## 12. Validation

Document:

- Makefile validation
- generation validation
- tests
- Docker build
- Docker smoke test if performed

## 13. Files Changed

List all relevant files.

## 14. Risks

Document operational risks.

## 15. Unresolved Issues

Document anything that should be addressed by Agent 12 or later work.

---

# 67. Documentation Check

Before finishing, verify again that:

- README.md was read
- agents/project-context.md was read
- docs/domain-model.md was read
- docs/repository-layout.md was read
- docs/protobuf-strategy.md was read
- docs/migration-plan.md was read
- docs/transactions-existing-review.md was read
- docs/transactions-database-review.md was read
- docs/transactions-sqlc-review.md was read
- docs/transactions-protobuf-review.md was read
- docs/transactions-repository-review.md was read
- docs/transactions-merchants-review.md was read
- docs/transactions-customers-review.md was read
- docs/transactions-deposits-review.md was read
- docs/transactions-payouts-review.md was read
- docs/transactions-runtime-review.md was read

The review document must accurately describe what was actually implemented.

---

# Completion Checklist

Before stopping:

- [ ] Required documents were read.
- [ ] Existing Deposits scaffolding was inspected.
- [ ] Transactions runtime location was confirmed.
- [ ] Transactions Dockerfile was created or completed.
- [ ] Docker build context is correct.
- [ ] Docker uses the correct Go version.
- [ ] Docker runs the Transactions runtime.
- [ ] Docker does not contain secrets.
- [ ] Docker does not unnecessarily contain development tools.
- [ ] `.dockerignore` behavior was verified where applicable.
- [ ] Transactions Makefile was created or completed.
- [ ] Build target works.
- [ ] Test target works.
- [ ] Generation targets are correct.
- [ ] Migration targets are safe if present.
- [ ] `.env.example` reflects actual configuration.
- [ ] No real secrets were added.
- [ ] Transactions README was created or completed.
- [ ] README commands correspond to real commands.
- [ ] Local development workflow is documented.
- [ ] Docker workflow is documented.
- [ ] Code generation ownership is documented.
- [ ] Root README was updated only if required.
- [ ] Root Makefile was updated only if required.
- [ ] Root `.env.example` was updated only if required.
- [ ] No generated files were manually modified.
- [ ] No protobuf contracts were changed.
- [ ] No database schema was changed.
- [ ] No repository architecture was changed.
- [ ] No business logic was rewritten.
- [ ] No provider integrations were added.
- [ ] No OAuth functionality was added.
- [ ] No webhook functionality was added.
- [ ] No deployment workflow was modified.
- [ ] No third_party/googleapis files were modified.
- [ ] Tests were run where practical.
- [ ] Docker build was run where practical.
- [ ] Git status was reviewed.
- [ ] Git diff was reviewed.
- [ ] docs/transactions-scaffolding-review.md was created.
- [ ] Documentation check was completed.

---

# Final Stop Condition

STOP after completing:

1. Transactions Dockerfile
2. Transactions Makefile
3. Transactions `.env.example`
4. Transactions README
5. necessary root-level integration
6. operational command validation
7. Docker build validation where possible
8. documentation
9. docs/transactions-scaffolding-review.md

Do NOT proceed to:

- comprehensive testing architecture
- broad production review
- provider integration
- OAuth
- webhook processing
- deployment
- unrelated service changes
- architectural redesign

Those responsibilities belong to later agents.

If a required component from Agents 02–10 is missing:

do not recreate it.

Document the missing component in:

docs/transactions-scaffolding-review.md

and identify which previous/later agent owns the correction.

STOP.