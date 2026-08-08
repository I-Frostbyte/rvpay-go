# Agent 06 — Docker

## Objective

Design, implement, and validate the Docker configuration for the new RVPay architecture.

The objective is to ensure that each deployable RVPay service can be:

1. built reproducibly,
2. packaged into a minimal production image,
3. started using the correct service entry point,
4. configured entirely through environment variables,
5. connected to external infrastructure correctly,
6. used by the Render deployment architecture,
7. built locally without relying on developer-machine state.

This agent owns Docker implementation.

This agent does NOT own:

- GitHub Actions workflow design
- Render service configuration
- Render deployment hooks
- PostgreSQL infrastructure
- application business logic
- protobuf contract design
- HTTP gateway design
- OAuth
- webhook behavior
- observability implementation
- security architecture
- performance optimization

Those concerns belong to other agents.

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
- docs/platform-protobuf-generation-review.md
- docs/platform-http-gateway-review.md
- docs/platform-common-packages-review.md
- docs/platform-ci-cd-review.md

Also inspect:

- existing Dockerfiles identified by README.md
- existing Docker-related configuration identified by docs/repository-layout.md
- existing .dockerignore files, if present

Do not inspect unrelated repository directories.

---

# Documentation Check

Before starting:

verify that all required documents exist.

Required:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md
- docs/platform-repository-audit.md
- docs/platform-protobuf-generation-review.md
- docs/platform-http-gateway-review.md
- docs/platform-common-packages-review.md
- docs/platform-ci-cd-review.md

If any required document is missing:

STOP.

Do not recreate the missing document.

At the end of the task:

perform the documentation check again.

Create:

docs/platform-docker-review.md

and record the result.

---

# Repository Exploration Rules

## IMPORTANT

Do NOT perform an unrestricted recursive repository search.

Use:

README.md

as the primary repository map.

Use:

docs/repository-layout.md

as the authority for the intended repository structure.

Use:

agents/project-context.md

for implementation conventions.

Use the previous platform review documents to understand work already performed.

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

DO NOT inspect:

third_party/googleapis/

Docker may require files from the protobuf dependency tree, but that does not justify recursively reviewing the submodule.

---

# 1. Existing Docker Audit

Identify existing Dockerfiles from:

- README.md
- docs/repository-layout.md
- platform repository review documentation

Inspect only Dockerfiles relevant to deployable RVPay services.

Determine:

- base image
- build context
- working directory
- build command
- generated-code requirements
- copied files
- runtime command
- exposed ports
- environment assumptions
- entrypoint behavior

---

# 2. Preserve Working Dockerfiles

Do not automatically replace existing Dockerfiles.

Determine whether each Dockerfile:

- is still used,
- belongs to a legacy service,
- belongs to the new architecture,
- is referenced by CI,
- is referenced by deployment documentation.

Preserve working Dockerfiles unless the new architecture explicitly requires modification.

---

# 3. Dockerfile Ownership

This agent owns:

- Dockerfiles
- Docker build configuration
- .dockerignore
- Docker-specific documentation

This agent does NOT own:

- GitHub Actions workflow implementation
- Render service creation
- Render blueprint configuration

If CI references an incorrect Dockerfile path:

document it for Agent 05 unless fixing the path is clearly required by the Docker implementation.

---

# 4. Service Architecture

Use:

docs/repository-layout.md

to identify deployable services.

Do not assume:

one repository = one Docker image.

The new RVPay architecture may contain multiple independently deployable services.

At minimum, account for the services documented by the project.

---

# 5. Service Entry Points

Determine each service's actual executable entry point.

Use the repository documentation and source structure.

Do not invent commands such as:

go run main.go

unless the repository actually uses that entry point.

For each service determine:

- package containing main()
- expected binary name
- runtime command

---

# 6. Go Version

Determine the authoritative Go version from:

- go.mod
- agents/project-context.md
- README.md
- existing Dockerfiles

Do not arbitrarily choose a version.

The Docker build environment must be compatible with the project's Go version.

---

# 7. Multi-Stage Builds

Prefer a multi-stage Docker build for production Go services.

The general architecture should be:

builder stage
    ↓
compiled binary
    ↓
minimal runtime image

Do not copy the Go toolchain into the final runtime image unless explicitly required.

---

# 8. Builder Image

Use an appropriate official Go image.

The version must match the project's documented Go toolchain.

Do not use:

golang:latest

Do not silently upgrade Go.

---

# 9. Runtime Image

Select the smallest practical runtime environment compatible with the service.

Consider:

- distroless
- minimal Debian
- Alpine

based on the application's actual requirements.

Do not choose an image solely because it is smaller.

Verify compatibility with:

- TLS
- DNS
- CA certificates
- timezone behavior if required
- Go networking
- database connectivity

---

# 10. CGO

Determine whether the application requires CGO.

Do not blindly set:

CGO_ENABLED=0

if the application or dependencies require CGO.

If CGO is disabled:

ensure the resulting binary works in the selected runtime image.

---

# 11. Static Binary

If the application supports a static binary:

prefer a statically linked production binary.

However:

do not introduce linker flags solely for size without verifying runtime compatibility.

---

# 12. Build Flags

Inspect existing project conventions.

If the existing Dockerfile uses:

-ldflags '-s -w'

determine whether it remains appropriate.

Do not change build flags merely for cosmetic reasons.

---

# 13. Build Context

Determine the correct Docker build context.

This is particularly important because service directories may depend on:

- root go.mod
- root go.sum
- protobuf-generated code
- common packages
- shared configuration
- repository-wide files

Do not assume the Docker build context should be the service directory.

---

# 14. Root Build Context

If the service depends on repository-level Go modules:

the Docker build should use the repository root as the build context where necessary.

Example:

docker build -f clients/Dockerfile .

Do not change the Docker build context merely because the Dockerfile lives inside a service directory.

---

# 15. Copy Strategy

Copy only the files required to build the service.

Avoid:

COPY . .

as the first choice when the repository contains large unrelated directories.

However:

do not create an excessively complicated COPY sequence that becomes fragile.

Balance:

- build correctness
- caching
- maintainability
- security

---

# 16. Dependency Caching

Structure Dockerfiles so dependency downloads can be cached.

For example:

copy go.mod/go.sum

download dependencies

then copy source

then build

Use the actual repository module structure.

Do not assume a service has its own go.mod.

---

# 17. Generated Code

Determine whether the service requires generated:

- protobuf
- gRPC
- grpc-gateway
- sqlc

code.

Docker builds must use the correct generated source.

Do not generate code by modifying the developer repository during a runtime container build.

---

# 18. Generation Strategy

Determine whether generated code is:

1. committed to the repository, or
2. generated during Docker build.

Use:

docs/protobuf-strategy.md

and:

docs/platform-protobuf-generation-review.md

as the authority.

Do not invent a new generation strategy.

---

# 19. Prefer Repository State

If the project's documented strategy is to commit generated source:

the Docker build should consume the committed generated source.

Do not add protoc/sqlc installations to Docker simply because those tools exist in the repository.

---

# 20. Docker Toolchain

Do not install:

- protoc
- sqlc
- mockgen
- goose

into the production runtime image.

If generation is intentionally performed in the builder stage:

use only the tools explicitly required.

---

# 21. Build Reproducibility

A Docker build from a clean checkout must succeed.

It must not depend on:

- developer-installed protoc
- developer-installed sqlc
- developer-installed Go tools
- local .env
- local database
- local filesystem state

---

# 22. Environment Files

Do NOT copy:

.env

into production images.

Do not use:

COPY .env ...

Do not hard-code production secrets into Dockerfiles.

---

# 23. .env.example

If the repository uses:

.env.example

it may remain part of the source repository.

Do not copy it into the production runtime image unless there is a concrete documented reason.

---

# 24. Runtime Configuration

Production configuration must come from environment variables or the project's documented configuration mechanism.

The Docker image should not contain environment-specific secrets.

---

# 25. PostgreSQL

The Docker image should not contain PostgreSQL.

The application connects to PostgreSQL externally.

Do not add:

postgres

services to the application Dockerfile.

---

# 26. Database Host

Do not hard-code:

localhost

for production database connections.

Render or another deployment environment will provide the correct database connection configuration.

---

# 27. Local Development

Do not confuse local development networking with production networking.

For example:

localhost

inside a container refers to that container.

Do not assume it refers to the developer's host machine.

---

# 28. Ports

Determine the service port from:

- configuration
- README.md
- existing service implementation
- repository documentation

Do not invent a port.

---

# 29. EXPOSE

Use:

EXPOSE

only as documentation of the container port.

Do not assume EXPOSE automatically publishes the port.

---

# 30. Render Compatibility

The Docker image must be compatible with Render deployment.

Follow the project's documented Render architecture.

Do not implement Render configuration here.

Agent 07 owns Render configuration.

---

# 31. PORT Environment Variable

Determine whether Render requires the service to bind to a platform-provided port.

If the application currently uses a fixed port:

determine whether the new deployment architecture requires configuration through:

PORT

or another documented variable.

Do not blindly introduce PORT handling into the application here.

If application code must change:

document it for the appropriate service agent.

---

# 32. Bind Address

Production network services should generally bind to an address reachable from the container network.

If the application currently binds only to:

127.0.0.1

determine whether that prevents Render/container networking.

Do not silently rewrite service runtime code.

Document the required application change if necessary.

---

# 33. Health Checks

Determine whether the project exposes a health endpoint.

Do not invent an HTTP health endpoint if one does not exist.

If Docker health checks are appropriate:

implement only if compatible with the service.

Do not add curl/wget merely to satisfy a health check.

---

# 34. Docker HEALTHCHECK

Only add:

HEALTHCHECK

when the runtime image contains a reliable mechanism for performing it.

Do not add a health check that always succeeds.

Do not add a health check that depends on unavailable shell utilities.

---

# 35. gRPC Services

For gRPC-only services:

do not assume an HTTP health endpoint exists.

Determine whether the service implements the gRPC health checking protocol.

If not:

do not invent one inside the Dockerfile.

Document the requirement.

---

# 36. Process Model

The container should run one primary service process.

Do not add:

- supervisors
- init systems
- shell process managers

unless explicitly required.

---

# 37. ENTRYPOINT vs CMD

Use the mechanism consistent with the repository's Docker conventions.

For a fixed service executable:

ENTRYPOINT may be appropriate.

For a configurable executable:

CMD may be appropriate.

Do not introduce both unnecessarily.

---

# 38. Binary Location

Use a predictable binary path.

For example:

/app/server

or another project-standard location.

Do not create a directory with the same name as the executable.

Verify the final Dockerfile avoids the common error:

exec: "/app/server": is a directory

---

# 39. Working Directory

Set:

WORKDIR

only where useful.

Do not depend on implicit working directories.

---

# 40. File Ownership

Consider running the application as a non-root user.

If the selected runtime image supports a non-root user:

prefer it where practical.

However:

do not introduce complicated permission handling without need.

---

# 41. Root User

Avoid running the production service as root unless there is a documented reason.

If the application requires root:

document why.

---

# 42. Filesystem

The application should not require write access to the entire container filesystem.

If it writes temporary files:

identify the required directory.

Do not grant broad write permissions unnecessarily.

---

# 43. Logs

Applications should normally write logs to stdout/stderr.

Do not configure Docker to manage application log files inside the container unless explicitly required.

This allows Render to collect logs.

---

# 44. Signals

Ensure the application receives termination signals correctly.

The container's primary process should be the application.

Avoid wrapping the application in shell commands that interfere with signal delivery.

---

# 45. Graceful Shutdown

Do not implement graceful shutdown in the Dockerfile.

If the application fails to handle SIGTERM correctly:

document it for the appropriate service/runtime agent.

---

# 46. CA Certificates

If the service makes HTTPS requests:

ensure the runtime image contains CA certificates.

This is especially important for:

- HighLevel API
- payment providers
- Render services
- external HTTP APIs

Do not assume a minimal image contains certificates.

---

# 47. DNS

Ensure the runtime image supports DNS resolution.

Do not remove required networking components merely to minimize image size.

---

# 48. Timezone

Do not install timezone databases unless the application actually requires local timezone information.

Go applications should generally operate using UTC where appropriate.

Follow existing application conventions.

---

# 49. Image Size

Reduce image size where practical.

Prioritize:

1. correctness
2. reproducibility
3. security
4. maintainability
5. size

Do not sacrifice reliability for a few megabytes.

---

# 50. Docker Ignore

Inspect whether:

.dockerignore

exists.

If not:

create one when useful.

It should exclude unnecessary content such as:

- .git
- .github
- local environment files
- coverage output
- temporary files
- IDE metadata
- local build artifacts

Do not exclude source files required for the build.

---

# 51. Third-Party Dependencies

Do not exclude required module files or source directories from Docker build context.

In particular:

do not add broad patterns that accidentally remove required Go source.

---

# 52. Googleapis

Do not modify:

third_party/googleapis/

If Docker requires protobuf source from this submodule:

ensure the checkout/build context contains it.

Do not copy the entire submodule into the final runtime image unless required.

---

# 53. Git Metadata

Do not include:

.git/

in the Docker image.

---

# 54. GitHub Actions

Do not modify GitHub Actions workflows in this agent unless a Docker-specific path is clearly broken.

Agent 05 owns CI/CD.

If a workflow references the wrong Dockerfile:

document the required adjustment.

---

# 55. Existing Docker Build Commands

Inspect:

docs/platform-ci-cd-review.md

to determine the expected Docker build invocation.

The Dockerfile must work with that build context and path.

---

# 56. Local Build

Perform a local Docker build using the project's expected command.

For example:

docker build -f <service>/Dockerfile .

Use the actual repository paths.

Do not invent paths.

---

# 57. Image Naming

Use a descriptive local image name for testing.

Do not hard-code a production registry.

Do not push images to Docker Hub or another registry from this agent.

---

# 58. Container Startup

After building:

run the image locally where practical.

Verify:

- binary starts
- configuration loading occurs
- expected startup logs appear
- no immediate executable/path error occurs

---

# 59. External Dependencies

If the service requires PostgreSQL or external APIs:

do not expect the local container to connect unless those dependencies are intentionally available.

A failed connection to an unavailable local database is not automatically a Docker failure.

Distinguish:

container startup failure

from:

application dependency failure.

---

# 60. Configuration Testing

Test that missing configuration fails clearly.

Do not embed fallback production credentials.

---

# 61. Database URL

If configuration requires a PostgreSQL URL:

ensure Docker passes it through environment variables.

Do not hard-code:

postgres://postgres:...

into the Dockerfile.

---

# 62. OAuth Secrets

If Clients requires OAuth secrets:

do not put them in:

- Dockerfile
- image
- build arguments
- source code

They must be supplied at runtime.

---

# 63. Webhook Secrets

Same rule applies to webhook-related secrets.

Never bake them into the image.

---

# 64. Build Arguments

Do not use Docker ARG for secrets.

Docker build arguments can become visible through image history or build metadata.

---

# 65. Environment Variables

Runtime secrets belong in:

ENVIRONMENT

or the deployment platform's secret configuration.

Do not use Dockerfile ENV for secret values.

Non-sensitive defaults may be used where appropriate.

---

# 66. Production Image Inspection

After building:

inspect the resulting image where practical.

Verify that it does not contain:

- .env
- private keys
- credentials
- unnecessary source files
- Git history
- local artifacts

---

# 67. Source Code in Runtime Image

Prefer copying only the compiled binary and required runtime assets into the final image.

Do not copy the complete source repository into the production image unless required.

---

# 68. Configuration Files

Only copy runtime configuration files that are explicitly required.

Do not copy:

- development configs
- test fixtures
- migration source unless runtime migration execution requires it
- local scripts

without justification.

---

# 69. Database Migrations

Determine how production migrations are handled.

If migrations execute as part of application startup:

the image may need migration files.

If migrations are handled separately:

do not unnecessarily package them into the runtime image.

Use the documented architecture.

---

# 70. Migration Runner

Do not implement migration logic in this agent.

If the service requires migration files at runtime:

ensure the Docker image contains the required files.

---

# 71. SQLC Generated Code

Generated sqlc code is source code from Docker's perspective.

It must be available during compilation.

Do not install sqlc in the runtime image.

---

# 72. Protobuf Generated Code

Generated protobuf code is source code from Docker's perspective.

Ensure the build context contains the generated packages required by the service.

Do not install protoc in the runtime image.

---

# 73. Build Cache

Use Docker layer ordering to improve rebuild speed.

A reasonable structure is:

COPY go.mod/go.sum
download dependencies
COPY required source
build

Adapt to actual repository layout.

---

# 74. Build Secrets

Do not pass production secrets into Docker build.

Build-time configuration should not require production credentials.

---

# 75. Reproducible Dependencies

Use the repository's:

go.mod

and:

go.sum

as authoritative dependency state.

Do not run:

go get

during Docker builds.

---

# 76. Go Module Downloads

Prefer:

go mod download

or the project's documented dependency mechanism.

Do not mutate module files during the Docker build.

---

# 77. No go mod tidy

Do not run:

go mod tidy

inside the Docker build.

It may alter module state and make builds non-reproducible.

---

# 78. Compiler Errors

If Docker compilation exposes existing Go errors:

do not modify unrelated application code.

Determine whether the error is:

- Docker build context problem
- missing generated code
- missing dependency
- existing source-code problem

Fix only Docker-owned issues.

---

# 79. Architecture Compatibility

Determine the deployment architecture expected by Render.

Do not hard-code an architecture-specific binary unless required.

Use the standard architecture expected by the deployment platform.

---

# 80. Build Platform

Do not introduce:

--platform

unless there is a documented cross-platform requirement.

If cross-compilation is necessary:

document it.

---

# 81. Development Docker Compose

Do not create docker-compose files unless explicitly required by the project.

This agent is for deployable Docker images.

---

# 82. Local PostgreSQL Containers

Do not create a PostgreSQL container as part of this agent unless existing project documentation explicitly requires it.

---

# 83. Docker Networking

Do not redesign application networking.

Only ensure the container behaves correctly within the documented deployment architecture.

---

# 84. Image Tags

Do not implement production image tagging or registry publishing here.

CI/CD owns that process.

---

# 85. Registry Authentication

Do not add registry credentials to Dockerfiles or scripts.

---

# 86. Vulnerability Scanning

Determine whether the project already uses:

- Trivy
- Docker Scout
- another scanner

If already documented:

preserve it.

Do not introduce a new security platform in this agent.

Agent 10 owns security architecture.

---

# 87. Dockerfile Comments

Keep comments concise.

Explain only non-obvious decisions.

Do not turn the Dockerfile into a tutorial.

---

# 88. Dockerfile Maintainability

Avoid unnecessarily clever Dockerfile techniques.

Prefer:

- explicit stages
- predictable paths
- straightforward commands
- documented assumptions

---

# 89. Legacy Deposits Dockerfile

Deposits is the reference implementation.

Inspect its Dockerfile.

Use its conventions where they remain compatible with the new architecture.

Do not blindly copy it.

The new service may require different:

- build paths
- binary names
- migrations
- runtime configuration
- generated code

---

# 90. Clients Service

If the Clients service exists:

verify its Dockerfile builds the correct executable.

Verify:

- Clients source is included
- common packages are available
- generated protobuf code is available
- configuration is runtime-driven

Do not implement Clients application logic here.

---

# 91. Transactions Service

If the Transactions service exists:

verify its Dockerfile builds the correct executable.

Verify:

- Transactions source is included
- common packages are available
- generated protobuf code is available
- sqlc output is available
- migrations are handled according to project architecture

Do not implement Transactions application logic here.

---

# 92. Multiple Dockerfiles

If multiple services require separate images:

each deployable service should have its own Dockerfile unless the documented architecture specifies another approach.

Do not create separate Dockerfiles for internal packages.

---

# 93. Shared Dockerfile

Do not create an abstract shared Dockerfile merely to eliminate duplication.

Prefer service-specific Dockerfiles when services have meaningful differences.

---

# 94. Runtime Environment

Production images should contain only what is required to execute the service.

Do not include:

- Go compiler
- Git
- protoc
- sqlc
- mockgen
- Make
- editor tools

unless actually required at runtime.

---

# 95. Make

Do not install Make in the production image.

Make belongs to development/build orchestration.

---

# 96. Shell

If using a distroless image:

remember that shell-based ENTRYPOINT commands may not work.

Prefer direct executable invocation.

---

# 97. Debugging

Do not add permanent debugging tools to production images.

If debugging is required:

use temporary local tooling or document the limitation.

---

# 98. Error Handling

If Docker build fails:

identify the exact failing layer.

Do not blindly modify multiple layers at once.

---

# 99. Incremental Validation

After changing a Dockerfile:

perform:

1. Dockerfile inspection
2. local build
3. image inspection
4. container startup where practical

Do not make many unrelated changes before testing.

---

# 100. Final Docker Build

Perform a clean build where practical.

Avoid relying only on Docker's existing local cache.

Use a clean/no-cache build if necessary to verify reproducibility.

Do not repeatedly perform expensive clean builds when one validation is sufficient.

---

# 101. Runtime Validation

Run the resulting container where practical.

Verify:

- executable starts
- expected process is PID 1
- configuration is read
- no missing-library error occurs
- no executable/path error occurs
- expected port is used

Do not require unavailable external infrastructure merely to prove the binary starts.

---

# 102. Image Contents

Inspect the final image sufficiently to verify that sensitive or development-only files are not included.

Do not recursively dump the entire filesystem into logs.

---

# 103. Docker Ignore Validation

Ensure .dockerignore does not exclude:

- go.mod
- go.sum
- required Go source
- required generated source
- required migrations
- required protobuf-generated packages

---

# 104. CI Compatibility

Review:

docs/platform-ci-cd-review.md

and verify the Docker implementation matches the CI build command.

If a mismatch exists:

fix the Docker-side issue if appropriate.

Otherwise document it for Agent 05.

---

# 105. Render Compatibility

Review the Render-related architectural expectations from the documentation.

Do not implement Render settings.

Document any Docker requirements Agent 07 must satisfy.

---

# 106. Documentation

Create:

docs/platform-docker-review.md

Use exactly this structure:

# Platform Docker Review

## 1. Objective

Describe the Docker architecture.

## 2. Required Documentation

List all required documents and confirm they were read.

## 3. Existing Dockerfiles

| File | Service | Current Purpose | Decision |
|---|---|---|---|

## 4. Final Docker Architecture

Describe the final image/build strategy.

## 5. Build Context

Document the build context for each service.

## 6. Build Stages

Document the builder and runtime stages.

## 7. Service Images

| Service | Dockerfile | Binary | Runtime Image | Port |
|---|---|---|---|---|

## 8. Configuration

Describe how runtime configuration is supplied.

## 9. Generated Code

Document protobuf/sqlc/generated-code requirements.

## 10. Database/Migrations

Document whether migration files are included and why.

## 11. Security

Document:

- non-root execution
- secret handling
- .env exclusion
- build secret handling

Only record what is actually implemented.

## 12. Local Validation

Document the Docker build and startup commands used.

## 13. CI Compatibility

Explain how Docker integrates with Agent 05's CI workflow.

## 14. Render Requirements

Document Docker-side requirements that Agent 07 must account for.

## 15. Findings

| ID | Severity | File/Area | Finding | Resolution |
|---|---|---|---|---|

## 16. Deferred Work

Document issues belonging to other agents.

## 17. Changes Made

List only files actually modified.

## 18. Documentation Check

Record the final documentation verification.

## 19. Final Status

Use exactly one:

PASS

PASS WITH FOLLOW-UP

BLOCKED

---

# 107. No Unrelated Changes

Do not modify:

- Go business logic
- protobuf contracts
- generated protobuf code
- SQL queries
- migrations
- repository implementations
- OAuth code
- webhook code
- Render configuration
- GitHub Actions workflows
- observability implementation
- security implementation
- performance implementation

unless a Docker-specific issue absolutely requires a coordinated change.

If such a change is required:

document it before making it.

---

# 108. Final Verification

Run:

git status --short

Then:

git diff --stat

Then inspect every changed Docker-related file.

Ensure there are no unexpected changes.

---

# 109. Documentation Check

Verify again:

- README.md exists
- agents/project-context.md exists
- docs/domain-model.md exists
- docs/repository-layout.md exists
- docs/protobuf-strategy.md exists
- docs/migration-plan.md exists
- docs/platform-repository-audit.md exists
- docs/platform-protobuf-generation-review.md exists
- docs/platform-http-gateway-review.md exists
- docs/platform-common-packages-review.md exists
- docs/platform-ci-cd-review.md exists

Record this in:

docs/platform-docker-review.md

---

# 110. Completion Checklist

Before stopping:

- [ ] All required documentation was read.
- [ ] README.md was used as the repository map.
- [ ] agents/project-context.md was followed.
- [ ] Repository exploration was restricted.
- [ ] Deep folders were not recursively inspected.
- [ ] third_party/googleapis was not unnecessarily explored.
- [ ] Existing Dockerfiles were audited.
- [ ] Existing working Dockerfiles were preserved where appropriate.
- [ ] Correct service entry points were identified.
- [ ] Correct Go version was used.
- [ ] Multi-stage builds were considered.
- [ ] Production images do not contain the Go toolchain unnecessarily.
- [ ] Production images do not contain .env files.
- [ ] Production secrets are not embedded.
- [ ] Docker build context is correct.
- [ ] Required generated code is available.
- [ ] sqlc is not unnecessarily installed in the runtime image.
- [ ] protoc is not unnecessarily installed in the runtime image.
- [ ] Required migrations are handled correctly.
- [ ] Runtime configuration comes from the environment.
- [ ] Service ports are correct.
- [ ] Container process starts correctly.
- [ ] Containers do not accidentally execute a directory as a binary.
- [ ] Non-root execution was considered.
- [ ] Logs are compatible with container logging.
- [ ] CI Docker build expectations were reviewed.
- [ ] Render Docker requirements were documented.
- [ ] No unrelated code was modified.
- [ ] docs/platform-docker-review.md was created.
- [ ] Final documentation check was completed.
- [ ] git status was inspected.
- [ ] git diff was inspected.

---

# Final Stop Condition

STOP after:

1. reading all required documentation,
2. auditing existing Dockerfiles,
3. determining the deployable service images,
4. implementing the required Docker changes,
5. validating Docker build contexts,
6. validating production image construction,
7. validating container startup where practical,
8. verifying runtime configuration behavior,
9. verifying secrets are not embedded,
10. checking CI compatibility,
11. documenting Render-side Docker requirements,
12. completing the documentation check,
13. inspecting the final git diff.

Do NOT proceed to:

- CI/CD redesign
- Render configuration
- observability implementation
- security architecture
- performance optimization
- service business logic
- database redesign
- protobuf redesign

Those belong to other agents.

STOP.