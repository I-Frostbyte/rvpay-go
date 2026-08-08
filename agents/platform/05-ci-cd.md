# Agent 05 — CI/CD

## Objective

Review, implement, and stabilize the CI/CD pipeline for the new RVPay architecture.

The purpose of this agent is to ensure that the repository can reliably:

1. validate the Go codebase,
2. generate required code,
3. run tests,
4. verify generated code is current,
5. build the services,
6. build Docker images where appropriate,
7. detect integration/build regressions,
8. provide deterministic feedback before deployment.

This agent owns CI/CD workflow implementation and validation.

This agent does NOT own:

- Dockerfile implementation
- Render service configuration
- Render deployment configuration
- application runtime behavior
- service business logic
- database schema design
- protobuf contract design
- OAuth
- webhook implementation
- observability implementation
- security implementation
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

Also inspect the existing CI workflow files identified by README.md or
docs/repository-layout.md.

Do not assume a workflow filename.

---

# Documentation Check

Before starting:

verify that every required document exists.

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

If any required document is missing:

STOP.

Do not recreate the missing document.

At the end of the task:

perform the documentation check again.

Record the result in:

docs/platform-ci-cd-review.md

---

# Repository Exploration Rules

## IMPORTANT

Do NOT perform an unrestricted recursive repository search.

Use:

README.md

as the primary repository map.

Use:

docs/repository-layout.md

to understand the intended target layout.

Use:

agents/project-context.md

for coding and repository conventions.

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

The CI pipeline may reference protobuf dependencies without requiring inspection of the entire submodule.

---

# 1. Existing CI Audit

Before changing anything:

identify existing workflow files.

Inspect only the workflow files relevant to:

- Go builds
- tests
- code generation
- Docker builds
- deployment

Do not inspect unrelated workflows unless README.md indicates they are part of the RVPay build/deployment process.

---

# 2. Preserve Existing Workflows

Do not delete or replace existing workflows automatically.

Determine what each workflow currently does.

Document:

- purpose
- trigger
- jobs
- dependencies
- deployment target
- secrets
- generated-code requirements

If a workflow is still required:

preserve it.

If it has become obsolete:

document why before changing it.

---

# 3. Workflow Ownership

This agent owns CI/CD workflows.

It does not own:

Dockerfile content.

If Docker build configuration is incorrect because of a Dockerfile problem:

document the issue for Agent 06.

Do not rewrite the Dockerfile here.

---

# 4. Go Version

Determine the Go version from the authoritative project configuration.

Check:

- go.mod
- README.md
- existing CI configuration
- agents/project-context.md

Do not arbitrarily select a Go version.

The CI Go version must be compatible with the project's declared Go version.

---

# 5. Tooling Versions

Inspect:

- go.mod
- tooling documentation
- existing version documentation
- protobuf generation documentation

Do not silently upgrade tools.

If the project has an established pinned version:

use it.

If a version is inconsistent:

document the inconsistency and resolve it only when the intended version is clearly established by project documentation.

---

# 6. Protobuf Generation

The CI pipeline must support the documented protobuf generation process.

Use:

docs/protobuf-strategy.md

as the authority.

The workflow must install only the required protobuf tooling.

Do not manually edit generated files.

Do not modify protobuf contracts in this agent.

---

# 7. Protobuf Tool Versions

If protobuf generation requires:

- protoc
- protoc-gen-go
- protoc-gen-go-grpc
- protoc-gen-grpc-gateway

ensure their versions are explicitly controlled.

Do not use:

@latest

for reproducible CI.

---

# 8. Protobuf Generation Validation

After generation:

the workflow should detect generated-code drift.

The intended principle is:

generation must be deterministic.

The CI pipeline should fail if generation produces changes that are not committed.

Use an approach consistent with the existing repository.

For example:

generate code

then:

git diff --exit-code

against the generated directories.

Do not hard-code paths that contradict:

docs/repository-layout.md.

---

# 9. SQLC Generation

If sqlc is part of the documented build process:

CI must be able to reproduce sqlc generation.

Use the version documented by the project.

Do not install an unrelated global sqlc version.

If the repository uses:

go run ...

with a pinned version:

preserve that approach unless documentation explicitly requires a different mechanism.

---

# 10. SQLC Validation

Generated sqlc code must be reproducible.

If generation modifies tracked files:

CI must detect the difference.

Do not silently commit or overwrite generated files from CI.

The desired behavior is:

generate

verify

fail if generated output differs.

---

# 11. go generate

Determine whether:

go generate ./...

is part of the repository's documented workflow.

If it is:

include it in CI at the appropriate stage.

Do not invoke arbitrary generators discovered in unrelated directories.

Only use documented repository generation commands.

---

# 12. Test Execution

CI must execute the project's Go tests.

Use the repository's supported command.

Where appropriate:

go test ./...

should be the final repository-wide validation.

If the repository contains services that require external infrastructure:

do not assume CI can run them without configuration.

Follow project documentation.

---

# 13. Test Isolation

Do not introduce external dependencies merely to make CI tests pass.

Avoid requiring:

- local PostgreSQL
- Render
- HighLevel
- external payment APIs
- production credentials

for unit tests.

Tests that genuinely require infrastructure should be clearly separated or documented.

---

# 14. Race Detection

Determine whether race detection is appropriate for the project.

If the repository's runtime and test suite support it:

consider:

go test -race ./...

However:

do not enable race detection blindly if it causes unreasonable CI runtime or incompatibility.

Document the decision.

---

# 15. Static Analysis

Determine whether the project already uses:

- go vet
- golangci-lint
- staticcheck
- another documented analyzer

If an analyzer is already part of the project:

preserve it.

Do not introduce a large linting framework solely because it is popular.

---

# 16. go vet

If appropriate for the project:

run:

go vet ./...

Do not suppress warnings without understanding them.

If existing code produces unrelated warnings:

document them rather than rewriting unrelated code.

---

# 17. Formatting

CI should ensure Go formatting remains consistent.

Use:

gofmt

or the project's established formatting mechanism.

Do not introduce opinionated formatting rules that contradict agents/project-context.md.

---

# 18. Module Validation

CI should detect invalid module state.

Where appropriate:

go mod tidy

should NOT automatically modify the repository during CI.

Instead:

verify the module state without silently changing tracked files.

Do not run commands that unexpectedly rewrite go.mod/go.sum and then ignore the changes.

---

# 19. Build Validation

CI must verify that the services compile.

Determine the service entry points from:

README.md

and:

docs/repository-layout.md

Do not assume there is only one binary.

The new architecture may contain multiple service binaries.

---

# 20. Multi-Service Build

If Clients and Transactions each have their own executable:

CI must validate both.

Use the actual documented command paths.

Do not search the entire repository for binaries.

---

# 21. Build Matrix

Do not introduce a large build matrix unless justified.

Prefer a simple deterministic pipeline.

Avoid testing dozens of unsupported OS/architecture combinations unless the project requires them.

---

# 22. Docker Build

CI may invoke Docker builds to validate Dockerfiles.

However:

Dockerfile implementation belongs to Agent 06.

If the Docker build fails because the Dockerfile is incorrect:

record the failure.

Do not redesign the Dockerfile in this agent.

---

# 23. Docker Build Scope

Only build images that are actually part of the documented deployment architecture.

Do not automatically build:

- every directory containing a Dockerfile
- development-only images
- unrelated historical images

---

# 24. Deployment Boundary

Do NOT implement Render configuration here.

CI may trigger a Render deployment if the existing documented architecture uses a deploy hook.

However:

the actual Render service configuration belongs to Agent 07.

---

# 25. Render Secrets

Never hard-code:

- Render deploy hooks
- API keys
- database passwords
- OAuth credentials
- provider secrets

Use GitHub Actions secrets or the project's documented secret mechanism.

---

# 26. Secrets

Search only the relevant workflow files for credential handling.

Never print secrets.

Never echo environment variables containing secrets.

Do not commit credentials.

---

# 27. Secret Validation

If a required secret is missing:

the workflow should behave intentionally.

For deployment:

a missing deployment secret may either:

- fail deployment explicitly, or
- skip deployment if that is the documented behavior.

Do not accidentally report deployment success when deployment was skipped.

---

# 28. Pull Requests

Determine whether CI runs on:

- push
- pull_request
- workflow_dispatch

The repository should validate changes before they reach main.

If the project currently only validates main:

consider whether PR validation is appropriate.

Do not change branch strategy without documentation.

---

# 29. Main Branch

Do not change branch names.

Use the branch policy documented by the project.

---

# 30. Workflow Permissions

Use the minimum GitHub Actions permissions required.

Prefer:

permissions:
  contents: read

unless a job genuinely requires additional permissions.

Do not grant:

write-all

without justification.

---

# 31. GitHub Actions Versions

Use stable major versions of official actions.

Examples:

actions/checkout

actions/setup-go

Do not use unpinned or experimental action references unnecessarily.

---

# 32. Action Security

Avoid arbitrary third-party GitHub Actions unless the project already uses them or there is a clear documented reason.

Prefer official actions.

---

# 33. Shell Safety

Workflow shell scripts must fail predictably.

Use appropriate shell error handling where necessary.

Do not allow a failed generation command to be hidden by later successful commands.

---

# 34. Command Failures

Every critical command must be capable of failing the job.

Do not append:

|| true

to generation, testing, building, or validation commands merely to keep CI green.

---

# 35. Generated Code Failure

If generated code differs from committed code:

CI must fail.

Provide a clear message explaining:

- which generation step failed
- what developers should run locally

Do not automatically commit generated output.

---

# 36. Dependency Caching

Use caching only where supported by the chosen setup action.

Do not create custom cache logic unless necessary.

Avoid caching generated source or build artifacts that can become stale.

---

# 37. Build Reproducibility

CI should use explicit versions for:

- Go
- protobuf tools
- sqlc
- other required generators

Avoid:

- latest
- floating versions
- unversioned downloads

---

# 38. Timeouts

Do not introduce unnecessarily long CI timeouts.

If a job consistently hangs:

identify the cause.

Do not simply increase timeout values.

---

# 39. Parallelization

Use job dependencies appropriately.

A sensible high-level sequence may be:

generation
    ↓
tests
    ↓
Docker build
    ↓
deployment

Independent validation jobs may run in parallel where safe.

Do not parallelize generation steps that modify the same working tree.

---

# 40. Generated Code Ordering

Do not run:

tests

before required generated code exists.

The pipeline should establish generated code before compilation/testing if the project requires generated artifacts.

---

# 41. Database Generation

If database generation is part of the documented workflow:

ensure it occurs before tests requiring generated database code.

Do not run migrations against production databases from CI.

---

# 42. Database Tests

Do not introduce production database credentials into GitHub Actions.

If integration tests require PostgreSQL:

use the documented project approach.

If no documented approach exists:

do not invent one silently.

Document the requirement.

---

# 43. External APIs

CI tests must not call production external APIs unless explicitly required.

Do not embed:

- HighLevel credentials
- payment provider credentials
- Render credentials

into test workflows.

---

# 44. Workflow Naming

Use descriptive workflow/job names.

A developer should be able to identify:

- generation failure
- test failure
- build failure
- deployment failure

from the GitHub Actions UI.

---

# 45. Job Boundaries

Keep jobs logically separated.

For example:

generate

test

docker-build

deploy

is preferable to one enormous shell script.

However, do not split trivial commands into dozens of jobs.

---

# 46. Job Dependencies

Use:

needs:

where one stage depends on another.

For example:

test:
  needs: generate

Docker build should not occur after failed tests.

Deployment should not occur after failed Docker validation.

---

# 47. Deployment Gate

Deployment must only happen after the required validation stages succeed.

The intended relationship is:

generation
→ tests
→ Docker validation
→ deployment

Do not deploy directly after checkout.

---

# 48. Manual Deployment

If workflow_dispatch is supported:

ensure it still respects validation dependencies.

Do not allow a manually triggered workflow to bypass required tests unless explicitly documented.

---

# 49. Deployment Hook

If Render uses a deploy hook:

treat the hook as a secret.

The workflow should:

- retrieve it from GitHub secrets
- verify whether it exists
- invoke it securely
- fail if an actual deployment request fails

Do not expose the URL in logs.

---

# 50. Deployment Reporting

The CI workflow must make it obvious whether:

- deployment happened
- deployment was skipped
- deployment failed

Do not silently skip deployment and present the workflow as though the service was deployed.

---

# 51. Documentation

If CI commands are changed:

update README.md only where necessary.

Do not rewrite the README.

If developer commands are important but too detailed for README:

document them in the appropriate docs file.

---

# 52. Project Context

All workflow decisions must respect:

agents/project-context.md

especially:

- Go version
- package conventions
- generated code conventions
- service layout
- testing conventions

Do not introduce CI patterns that contradict the repository's documented conventions.

---

# 53. Repository Layout

All paths referenced by workflows must correspond to:

docs/repository-layout.md

Do not hard-code obsolete paths.

If an existing workflow references an old path:

determine whether that path is still valid.

---

# 54. Existing Deposits Service

Deposits is legacy/reference code.

Do not modify deposits merely to make CI easier.

If deposits currently requires special handling:

document it.

Do not break it while implementing CI for the new architecture.

---

# 55. Clients and Transactions

CI must account for the new service layout.

If the repository now contains:

Clients

and:

Transactions

ensure their build/test paths are included where appropriate.

Do not assume they share one executable.

---

# 56. Common Packages

CI must test the common packages introduced by Agent 04.

Do not create separate workflows for every package.

Repository-level testing is preferred where practical.

---

# 57. Protobuf Review Integration

Use:

docs/platform-protobuf-generation-review.md

to understand:

- generated directories
- generator commands
- expected outputs
- tool versions
- unresolved issues

Do not re-design protobuf generation unless the review identifies an unresolved CI problem.

---

# 58. HTTP Gateway Integration

Use:

docs/platform-http-gateway-review.md

to ensure CI validates gateway-generated code if required.

Do not implement gateway behavior here.

---

# 59. Common Package Integration

Use:

docs/platform-common-packages-review.md

to ensure newly created common packages are included in repository tests.

---

# 60. No Unrelated Cleanup

Do not clean up:

- old workflow files
- unrelated YAML formatting
- unrelated documentation
- unrelated shell scripts

unless directly required for the CI architecture.

---

# 61. Workflow Formatting

Keep YAML readable.

Use:

- consistent indentation
- descriptive names
- comments only where useful

Do not add excessive comments.

---

# 62. Environment Variables

Only define environment variables required by the job.

Do not expose the entire repository configuration to every job.

---

# 63. Working Directory

If a command requires a service-specific directory:

use:

working-directory:

where appropriate.

Avoid unnecessary:

cd

chains in long scripts.

Follow existing workflow conventions where they are clear.

---

# 64. Makefile Integration

If the repository has documented Make targets:

prefer those targets where they provide the canonical operation.

Do not duplicate complex Makefile logic in YAML.

However, if CI must explicitly pin tool versions:

do not hide version-critical operations behind opaque commands.

---

# 65. CI vs Local Commands

CI should reproduce the documented local workflow as closely as practical.

Developers should be able to understand:

"CI failed here; this is the equivalent command I can run locally."

Document any important difference.

---

# 66. Failure Messages

Critical validation failures should produce actionable messages.

Examples:

Generated protobuf code is out of date.
Run the documented protobuf generation command and commit the generated files.

Generated sqlc code is out of date.
Run the documented sqlc generation command and commit the generated files.

Do not produce vague:

"Build failed"

messages when a more useful message is available.

---

# 67. Review Existing Render Workflow

If an existing workflow already contains:

- generation
- tests
- Docker build
- Render deployment

review it carefully before replacing it.

Preserve working behavior where possible.

---

# 68. Do Not Duplicate Deployment Pipelines

Do not create a second Render deployment workflow if one already exists.

Consolidate only when the architecture clearly requires it.

If multiple workflows intentionally target different environments:

preserve the distinction.

---

# 69. Branch Protection

Do not modify GitHub branch protection settings.

This agent controls workflow files only.

---

# 70. Release Management

Do not implement:

- semantic-release
- version tagging
- changelog automation
- GitHub releases

unless already required by the documented project architecture.

---

# 71. Artifact Management

Do not introduce artifact uploads unless they provide concrete value.

Do not upload:

- entire repositories
- secrets
- unnecessary build directories

---

# 72. Logs

CI logs should contain enough information to diagnose failures.

Do not print:

- passwords
- API keys
- deploy hooks
- database URLs containing credentials

---

# 73. Sensitive Error Output

Be careful with commands that may print environment variables or connection strings.

If a command can expose secrets:

use a safer invocation.

---

# 74. CI Security

Never use:

curl "$SECRET"

in a way that prints or exposes the secret.

For deployment hooks:

send the request without echoing the value.

---

# 75. Dependency Updates

Do not upgrade unrelated dependencies during this task.

CI implementation is not a dependency modernization exercise.

---

# 76. Go Modules

Do not change:

go.mod

or:

go.sum

unless a CI requirement genuinely requires it.

If changed:

document why.

---

# 77. Generated Files

Do not commit generated files from the CI environment.

CI should validate generated output, not become the source of generated code.

---

# 78. Local Reproduction

For every major CI stage, identify the corresponding local command.

Document these in:

docs/platform-ci-cd-review.md

---

# 79. Workflow Validation

After editing workflows:

inspect the final YAML carefully.

Check:

- indentation
- job names
- dependencies
- expressions
- environment variables
- secret references
- shell syntax

Do not assume a workflow is valid merely because YAML formatting looks correct.

---

# 80. YAML Syntax

Use a YAML parser or an existing project-supported validation method if available.

If no YAML validation tool is available:

perform a careful structural review.

Do not add a dependency solely for YAML validation unless necessary.

---

# 81. GitHub Actions Expressions

Verify expressions such as:

${{ secrets.NAME }}

and:

${{ github.ref }}

are syntactically correct.

Do not quote expressions unnecessarily if doing so changes their behavior.

---

# 82. Matrix Variables

If using a matrix:

verify every matrix value is consumed correctly.

Do not create a matrix unless it is justified.

---

# 83. Permissions

Ensure workflow permissions are no broader than necessary.

Prefer read-only permissions for build/test jobs.

---

# 84. CI Runtime

Avoid:

- repeated dependency downloads
- repeated code generation
- redundant repository checkout
- unnecessary Docker builds

Optimize only where the optimization is obvious and safe.

---

# 85. Cache Correctness

If caching is used:

ensure cache keys account for:

- Go version
- go.sum
- relevant dependency state

Do not cache generated source as a substitute for generation.

---

# 86. Build Reproducibility

A fresh GitHub runner should be able to execute the workflow without relying on:

- developer machine state
- local binaries
- local environment variables
- untracked generated files

---

# 87. Clean Checkout Assumption

Assume the CI runner starts from a clean checkout.

If the pipeline requires generated files:

generate them explicitly.

Do not rely on files existing locally.

---

# 88. Submodules

If the repository uses submodules:

follow the documented checkout configuration.

Do not recursively inspect the submodule contents.

If CI requires a submodule:

checkout it using the appropriate GitHub Actions option.

---

# 89. Third-Party Submodule

The protobuf googleapis submodule is an input dependency.

Do not modify it.

Do not generate files into it.

Do not recursively audit it.

---

# 90. Build Context

If Docker builds depend on files outside the service directory:

ensure the CI Docker build context matches the Dockerfile expectations.

Do not modify the Dockerfile to compensate for an incorrect CI context.

Document mismatches for Agent 06.

---

# 91. Deployment Environment

Do not attempt to replicate Render's production environment entirely inside GitHub Actions.

CI validates the application.

Render runs the deployment environment.

---

# 92. Production Database

Never run destructive migrations against production from CI.

Do not connect CI to the production PostgreSQL database unless explicitly documented and required.

---

# 93. Migration Validation

If migration files need syntax validation:

use a safe mechanism.

Do not apply destructive migrations to production.

If no migration validation mechanism exists:

document it.

---

# 94. Test Database

If integration tests require PostgreSQL:

determine whether the repository already documents a test database strategy.

Do not invent production-like credentials.

---

# 95. Service Ports

Do not hard-code service ports into CI unless required for an integration test.

Do not start production services merely to run unit tests.

---

# 96. Process Lifecycle

If CI starts a background process:

ensure it is cleaned up.

Do not leave orphaned services running.

---

# 97. Integration Tests

If integration tests exist:

identify their dependencies.

Separate them from unit tests where necessary.

Do not make the entire pipeline dependent on external services unless explicitly required.

---

# 98. Flaky Tests

Do not hide flaky tests with:

continue-on-error

or:

|| true

Identify and document flaky behavior.

---

# 99. Retry Logic

Do not add arbitrary retries to mask deterministic failures.

Retries may be appropriate for genuinely transient external infrastructure failures.

If added:

document why.

---

# 100. Final CI Pipeline

The desired high-level pipeline should resemble:

checkout
   ↓
toolchain setup
   ↓
code generation
   ↓
generated-code verification
   ↓
format/static validation
   ↓
tests
   ↓
build
   ↓
Docker validation
   ↓
deployment trigger

Adapt this to the actual project architecture.

Do not blindly reproduce this diagram if the repository documentation specifies another ordering.

---

# 101. Review Document

Create:

docs/platform-ci-cd-review.md

Use exactly this structure:

# Platform CI/CD Review

## 1. Objective

Describe the final CI/CD responsibilities.

## 2. Required Documentation

List every required document and confirm it was read.

## 3. Existing Workflows

| Workflow | Purpose | Decision |
|---|---|---|

## 4. Final Pipeline

Describe the final workflow sequence.

## 5. Toolchain Versions

| Tool | Version | Source |
|---|---|---|

## 6. Generation

Document:

- protobuf generation
- sqlc generation
- go generate
- generated-code verification

## 7. Testing

Document:

- unit tests
- integration tests
- race detection
- static analysis

Only list what is actually implemented.

## 8. Build

Document:

- Go build validation
- service binaries
- Docker validation

## 9. Deployment

Document:

- deployment trigger
- branch restrictions
- required secrets
- failure behavior

## 10. Security

Document:

- GitHub permissions
- secret handling
- sensitive logging precautions

## 11. Findings

Use:

| ID | Severity | File/Area | Finding | Resolution |
|---|---|---|---|---|

## 12. Deferred Work

Document issues intentionally left to:

- Agent 06
- Agent 07
- Agent 09
- Agent 10
- Agent 11

or other relevant agents.

## 13. Local Reproduction Commands

Document the local command corresponding to each major CI stage.

## 14. Changes Made

List only files actually modified.

## 15. Documentation Check

Record the final documentation verification.

## 16. Final Status

Use exactly one:

PASS

PASS WITH FOLLOW-UP

BLOCKED

---

# 102. Final Verification

Run:

git status --short

Then:

git diff --stat

Inspect every changed workflow.

Ensure no unrelated files were modified.

---

# 103. Validation Commands

Run the project's appropriate:

- formatting checks
- generation commands
- tests
- build commands

Do not run destructive production operations.

---

# 104. Final Documentation Check

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

Record this in:

docs/platform-ci-cd-review.md

---

# 105. Final Scope Check

Expected modifications:

- CI workflow files
- possibly minimal CI-related documentation
- docs/platform-ci-cd-review.md

Do NOT modify:

- Dockerfiles
- Render configuration
- protobuf contracts
- generated protobuf files
- database migrations
- sqlc queries
- service business logic
- OAuth
- webhook handlers
- third_party/googleapis

unless a documented CI requirement absolutely requires it.

---

# 106. Completion Checklist

Before stopping:

- [ ] All required documentation was read.
- [ ] README.md was used as the repository map.
- [ ] agents/project-context.md was followed.
- [ ] Repository exploration was restricted.
- [ ] Deep folders were not recursively inspected.
- [ ] third_party/googleapis was not explored unnecessarily.
- [ ] Existing workflows were audited.
- [ ] Existing working workflows were preserved where appropriate.
- [ ] Go version is explicitly controlled.
- [ ] Tool versions are explicitly controlled where required.
- [ ] Protobuf generation is reproducible.
- [ ] sqlc generation is reproducible where required.
- [ ] Generated code drift is detected.
- [ ] Tests run in CI.
- [ ] Builds are validated.
- [ ] Docker builds are validated where required.
- [ ] Deployment is gated behind required validation.
- [ ] Secrets are not hard-coded.
- [ ] Secrets are not printed.
- [ ] GitHub permissions are minimal.
- [ ] CI failures are not hidden.
- [ ] No unnecessary retries were introduced.
- [ ] No unrelated dependencies were upgraded.
- [ ] No Dockerfile implementation was performed.
- [ ] No Render configuration was implemented.
- [ ] No security implementation was performed.
- [ ] No observability implementation was performed.
- [ ] No performance implementation was performed.
- [ ] docs/platform-ci-cd-review.md was created.
- [ ] Final documentation check was completed.
- [ ] git status was inspected.
- [ ] git diff was inspected.

---

# Final Stop Condition

STOP after:

1. reading all required documentation,
2. auditing the existing CI/CD workflows,
3. determining the canonical build and generation commands,
4. implementing the required CI/CD changes,
5. validating generated-code reproducibility,
6. validating tests,
7. validating builds,
8. validating Docker builds where appropriate,
9. ensuring deployment is correctly gated,
10. documenting the final pipeline,
11. completing the documentation check,
12. inspecting the final git diff.

Do NOT proceed to:

- Docker implementation
- Render configuration
- observability
- security
- performance optimization
- Clients implementation
- Transactions implementation
- database implementation
- protobuf contract redesign

Those belong to other agents.

STOP.