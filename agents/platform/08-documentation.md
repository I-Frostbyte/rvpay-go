# Agent 08 — Platform Documentation

## Objective

Bring the RVPay repository documentation up to date with the platform work completed by Agents 01–07.

The goal is to make the repository understandable to:

- developers
- Cline
- future AI agents
- DevOps engineers
- maintainers
- contributors

The documentation must accurately describe the implementation that actually exists.

Do NOT document planned functionality as if it already exists.

Do NOT redesign the architecture.

Do NOT modify application implementation merely to make documentation easier to write.

Documentation must follow the actual repository state.

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
- docs/platform-docker-review.md
- docs/platform-render-review.md

Also inspect only the specific source/configuration files referenced by those documents when verification is required.

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
- docs/platform-docker-review.md
- docs/platform-render-review.md

If any required document is missing:

STOP.

Do not recreate the missing document.

At the end of the task:

perform the documentation check again.

Create:

docs/platform-documentation-review.md

and record the result.

---

# Repository Exploration Rules

## IMPORTANT

Do NOT perform an unrestricted repository-wide search.

Use:

README.md

as the primary repository map.

Use:

docs/repository-layout.md

as the authority for the intended structure.

Use:

agents/project-context.md

for coding/package conventions.

Use the platform review documents to determine what has already been audited.

Only inspect implementation files when the documentation requires factual verification.

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

The documentation task does not require exploring generated/vendor/submodule contents.

---

# 1. Documentation Philosophy

Documentation must describe:

WHAT exists.

WHERE it exists.

WHY it exists.

HOW the major pieces interact.

Do not document:

imagined functionality.

---

# 2. Source of Truth

When documentation conflicts with implementation:

do not silently choose one.

Determine whether the implementation or documentation represents the current state.

If the implementation is clearly current:

update documentation.

If the architecture documentation is intentionally describing future work:

preserve the distinction.

---

# 3. Existing README

Read:

README.md

before modifying it.

Understand:

- existing project description
- existing setup instructions
- existing architecture explanation
- existing service list
- existing deployment information
- existing development commands

Do not rewrite the README blindly.

---

# 4. README Preservation

Preserve useful existing documentation.

Do not delete historical information unless it is demonstrably obsolete and misleading.

When replacing obsolete information:

prefer updating it rather than removing the entire section.

---

# 5. README Structure

The README should make it easy to understand:

1. What RVPay is.
2. Repository architecture.
3. Services.
4. Shared infrastructure.
5. Local development.
6. Protobuf/gRPC workflow.
7. Database/migration workflow.
8. Testing.
9. Docker.
10. Deployment.
11. Configuration.
12. Important architectural references.

Do not add sections for functionality that does not exist.

---

# 6. Project Overview

Update the README's project description so it reflects the current RVPay architecture.

Use:

docs/domain-model.md

as the architectural authority.

---

# 7. Service Overview

Document each currently implemented service.

At minimum:

- service name
- responsibility
- location
- communication interface
- database ownership where applicable

Do not invent services.

---

# 8. Clients Service

If the Clients service exists:

document:

- its purpose
- repository location
- gRPC/API role
- OAuth responsibilities
- webhook responsibilities where implemented
- database responsibility where implemented

Do not document unfinished features as completed.

---

# 9. Transactions Service

If the Transactions service exists:

document:

- its purpose
- repository location
- transaction responsibilities
- gRPC/API role
- database responsibility

Include only implemented functionality.

---

# 10. Platform Components

Document platform components that actually exist.

Examples:

- protobuf generation
- HTTP gateway
- shared packages
- Docker
- CI/CD
- Render
- PostgreSQL

Do not describe infrastructure that has not been implemented.

---

# 11. Repository Layout

Review:

docs/repository-layout.md

Ensure it accurately represents the current repository.

If the repository layout changed during Agents 01–07:

update the documentation.

---

# 12. Repository Tree

If the README contains a repository tree:

update it.

The tree should focus on meaningful project directories.

Do not produce enormous trees.

Do not include:

- generated/vendor internals
- Google API submodule contents
- .git internals
- node_modules
- temporary files

---

# 13. Service Directory Documentation

Document important service directories.

For example:

service/

├── cmd/

├── config/

├── db/

├── repository/

├── service/

└── ...

Use the actual project layout.

Do not invent directories.

---

# 14. Generated Code

Clearly distinguish:

source code

from:

generated code.

Document where generated protobuf and sqlc files live.

Do not tell developers to manually edit generated files.

---

# 15. Protobuf Documentation

Use:

docs/protobuf-strategy.md

and:

docs/platform-protobuf-generation-review.md

to document:

- protobuf source location
- generated Go location
- generation command
- gRPC stub generation
- expected workflow

Do not modify protobuf contracts.

---

# 16. Protobuf Generation Commands

Document the actual command used by the repository.

Do not invent a new command.

Verify commands against:

Makefile

or the documented generation process.

---

# 17. SQLC Documentation

Document the SQLC workflow if SQLC is part of the implemented repository.

Explain:

- SQL input location
- generated output location
- generation command
- generated-code ownership

Do not modify SQL queries.

---

# 18. Migration Documentation

Use:

docs/migration-plan.md

to document the current migration strategy.

Explain:

- migration location
- migration execution
- development workflow
- production considerations

Do not redesign migration execution.

---

# 19. Database Documentation

Document database ownership accurately.

For each database/service relationship:

state what is actually implemented.

Do not claim that each service has an isolated database if the implementation uses a shared database.

---

# 20. Local PostgreSQL

If local PostgreSQL is required:

document the expected setup.

Include:

- database name
- user
- port
- environment configuration

only where those values are actually established by the repository.

Do not invent credentials.

---

# 21. Environment Variables

Review environment configuration documented by the project.

Document:

- variable name
- purpose
- service
- whether required

Do not include secret values.

---

# 22. .env.example

If:

.env.example

exists:

ensure the README explains how to use it.

If it does not exist:

do not automatically create one unless another agent has established that requirement.

---

# 23. Secrets

Never document actual:

- passwords
- API keys
- OAuth secrets
- database passwords
- webhook signing secrets
- SSO keys

Use placeholders.

---

# 24. OAuth Documentation

If OAuth functionality exists:

document the operational flow at a high level.

Explain:

1. user starts installation/authorization
2. provider redirects to callback
3. authorization code is received
4. backend exchanges the code
5. credentials/tokens are stored
6. provider API calls can be performed

Do not document internal implementation details that are not necessary for developers using the system.

---

# 25. OAuth Configuration

Document the configuration variables required for OAuth.

Use actual variable names.

Do not invent names.

---

# 26. OAuth Redirect URL

Explain that production OAuth redirect URLs must point to the deployed public endpoint.

Do not put a temporary Render hostname into permanent documentation unless it is intentionally the project's configured production URL.

---

# 27. Webhook Documentation

If webhooks are implemented:

document:

- endpoint purpose
- provider
- expected role
- configuration requirements

Do not document undocumented event types.

---

# 28. SSO Documentation

If SSO functionality exists:

document only the operational purpose.

Do not expose the SSO key.

---

# 29. HTTP Gateway

Use:

docs/platform-http-gateway-review.md

to document the gateway.

Include:

- purpose
- location
- public/private role
- routing responsibility
- relationship to gRPC services

Do not document routes that do not exist.

---

# 30. gRPC Architecture

Explain the high-level communication model.

For example:

Client

↓

HTTP Gateway

↓

gRPC Service

↓

Repository

↓

PostgreSQL

Only use this structure where it reflects the actual implementation.

---

# 31. Internal Communication

Document service-to-service communication.

Clearly distinguish:

- external HTTP
- internal HTTP
- gRPC
- database access

Do not claim direct service-to-service communication where none exists.

---

# 32. Docker Documentation

Use:

docs/platform-docker-review.md

to document the Docker workflow.

Include:

- Dockerfiles
- build context
- image creation
- local build
- local execution

Use actual commands.

---

# 33. Docker Build

Document the repository's actual Docker build command.

Do not replace it with a generic command unless it is equivalent to the project's actual workflow.

---

# 34. Docker Run

Document local execution only if it is part of the established development workflow.

Include required environment configuration.

---

# 35. Docker Compose

If Docker Compose exists:

document it.

If it does not:

do not create documentation suggesting that it does.

---

# 36. CI/CD

Use:

docs/platform-ci-cd-review.md

to document CI/CD.

Explain:

- what CI validates
- what triggers it
- how Go is built/tested
- how Docker is involved
- how deployment is connected

Do not redesign CI/CD.

---

# 37. GitHub Actions

If GitHub Actions exists:

document the relevant workflow files.

Do not document unrelated workflows.

---

# 38. Render

Use:

docs/platform-render-review.md

to document Render deployment.

Include:

- Blueprint
- services
- databases
- environment variables
- public/private services
- deployment process

---

# 39. Render Blueprint

Document where the Render Blueprint lives.

Explain that the Blueprint is the version-controlled infrastructure definition.

---

# 40. Render Deployment

Provide the actual deployment procedure.

The procedure should include only steps that are supported by the project.

For example:

1. push changes
2. Render detects the repository change
3. Blueprint updates/deploys services
4. services start
5. health checks run

Do not invent dashboard steps that the repository does not require.

---

# 41. Production Configuration

Document which values must be configured manually in Render.

Examples:

- OAuth secrets
- API credentials
- database secrets
- webhook secrets

Do not expose their values.

---

# 42. Public Endpoints

Document production endpoints only if their structure is stable.

Do not hard-code temporary development URLs.

---

# 43. Health Checks

Document the health checks configured by Agent 07.

Only document actual endpoints.

Do not invent endpoints.

---

# 44. Logging

Document the logging behavior established by the project.

If applications log to stdout/stderr:

state that.

Do not instruct developers to search for application log files if none exist.

---

# 45. Observability

Do not implement observability.

Only document the observability behavior that currently exists.

Agent 09 owns implementation.

---

# 46. Security

Do not redesign security.

Document only current security practices that are actually implemented.

Examples:

- secrets stored in environment variables
- TLS through Render
- authenticated endpoints
- webhook verification

Only document what has been verified.

---

# 47. Performance

Do not redesign performance.

Document only current architecture relevant to performance.

Examples:

- PostgreSQL
- connection pooling
- caching if implemented
- service separation

Do not claim performance characteristics that have not been measured.

---

# 48. Development Workflow

The README should provide a clear development workflow.

At minimum, where applicable:

1. clone repository
2. configure environment
3. start dependencies
4. generate protobuf
5. generate SQLC
6. run migrations
7. run tests
8. run service

Use actual project commands.

---

# 49. Build Workflow

Document:

- Go build
- protobuf generation
- sqlc generation
- Docker build
- tests

Only use commands that actually exist.

---

# 50. Makefile

Inspect the relevant Makefiles.

Document important commands.

Do not create new Makefile targets.

---

# 51. Root Makefile

Document repository-wide commands.

Do not duplicate every service-specific command in the README.

---

# 52. Service Makefiles

Document important service-level commands where developers need them.

---

# 53. Testing

Document the repository's test workflow.

Include:

- unit tests
- service tests
- repository tests
- integration tests

only where they actually exist.

---

# 54. Test Commands

Use the actual commands established by the project.

Do not invent test flags.

---

# 55. Migration Commands

Use actual migration commands from the repository.

Do not provide generic migration tooling instructions if the repository uses a specific implementation.

---

# 56. Troubleshooting

Add concise troubleshooting guidance for real project issues.

Examples:

- database connection failures
- missing environment variables
- protobuf generation failures
- sqlc generation failures
- Docker build failures
- Render configuration failures

Do not create a huge generic troubleshooting guide.

---

# 57. Common Database Error

If the project previously encountered localhost database configuration problems:

document the production distinction:

localhost

is local to the running container/process.

Render services must use the configured Render database/service address.

Do not provide a hard-coded production hostname.

---

# 58. Common Environment Error

If a required environment variable is missing:

document how to identify it.

Do not provide fake values.

---

# 59. Common Docker Error

If the project has known Docker configuration requirements:

document them.

Use:

docs/platform-docker-review.md

as the source.

---

# 60. Common Render Error

If Agent 07 documented a Render-specific issue:

document the resolution.

Do not invent troubleshooting steps.

---

# 61. Architecture References

The README should point developers toward the deeper architecture documents.

At minimum:

- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md

and relevant platform reviews.

---

# 62. Documentation Hierarchy

Clearly distinguish:

README.md

from:

docs/*.md

The README should provide the practical entry point.

The docs should provide deeper architectural detail.

---

# 63. README Length

Do not turn README.md into a massive architecture document.

Keep it navigable.

Use links to deeper documents.

---

# 64. Documentation Links

Verify all internal documentation links.

For example:

docs/domain-model.md

must actually exist.

Do not create links to nonexistent files.

---

# 65. Broken Links

Search only documentation files for broken internal references.

Do not perform a repository-wide content search.

---

# 66. Terminology

Use the terminology established by:

docs/domain-model.md

Do not rename architectural concepts casually.

---

# 67. Service Names

Use exact service names.

Do not alternate between:

clients

client

integrations

integration

unless the documentation explicitly distinguishes them.

---

# 68. Transactions Terminology

Use the terminology established by the domain model.

Do not collapse:

merchant

customer

deposit

payout

transaction

into generic "payment" terminology if the architecture distinguishes them.

---

# 69. Legacy Terminology

If the old system used terminology that has changed:

document the migration relationship clearly.

Example:

Integrations → Clients

Deposits → Transactions

Only use this mapping if established by the architecture documentation.

---

# 70. Architecture History

Do not remove useful historical migration information.

Where appropriate:

explain that the current architecture builds upon the previous RVPay implementation.

---

# 71. Migration Documentation

Ensure the README links to:

docs/migration-plan.md

The migration plan should remain the authoritative source for migration sequencing.

---

# 72. Domain Documentation

Do not duplicate the entire domain model into README.md.

Provide a concise summary and link to:

docs/domain-model.md

---

# 73. Repository Layout Documentation

Do not duplicate the entire repository layout into README.md.

Provide a concise tree or summary and link to:

docs/repository-layout.md

---

# 74. Protobuf Documentation

Do not duplicate protobuf strategy details into README.md.

Summarize the workflow and link to:

docs/protobuf-strategy.md

---

# 75. Platform Reviews

The platform review documents are implementation records.

Do not turn them into user-facing tutorials.

---

# 76. README Audience

The README should answer:

"What is this repository and how do I work with it?"

The architecture documents answer:

"Why is it designed this way?"

The platform reviews answer:

"What platform work was performed and what decisions were made?"

Preserve this distinction.

---

# 77. No Fake Completion

Never write:

"Fully production ready"

"Production tested"

"All services operational"

unless those statements have been verified.

---

# 78. No Unsupported Claims

Do not claim:

- uptime
- scalability
- security guarantees
- throughput
- latency
- zero downtime
- high availability

unless documented evidence supports the claim.

---

# 79. No Secret Values

Before finishing:

search documentation changes for accidental secrets.

Look for obvious patterns such as:

- passwords
- API keys
- bearer tokens
- OAuth secrets
- database URLs containing credentials

Remove any accidental secret.

---

# 80. Documentation Accuracy

For every newly documented command:

verify that the command actually exists.

For every documented path:

verify that the path actually exists.

For every documented service:

verify that the service actually exists.

For every documented environment variable:

verify that the application actually reads it.

---

# 81. README Verification

After editing README.md:

read the modified sections completely.

Check for:

- incorrect commands
- stale paths
- stale service names
- duplicate information
- broken links
- contradictory statements

---

# 82. Markdown Quality

Ensure Markdown is clean.

Check:

- headings
- lists
- code blocks
- tables
- links
- indentation

Avoid unnecessarily complex formatting.

---

# 83. Code Blocks

Use fenced code blocks for commands and configuration.

Use:

```bash

for shell commands.

Use:

```yaml

for YAML.

Use:

```text

for directory trees when appropriate.

---

# 84. Environment Examples

When showing environment configuration:

use placeholders.

Example:

DATABASE_URL=<your-database-url>

Never use real credentials.

---

# 85. Configuration Documentation

Explain what each required variable controls.

Do not merely dump an environment-variable list without context.

---

# 86. Developer Onboarding

Ensure a new developer can understand:

- what to install
- what to configure
- how to generate code
- how to start dependencies
- how to run services
- how to run tests

Do not add unnecessary tooling.

---

# 87. Cline Onboarding

The documentation should also make the repository understandable to future Cline agents.

Clearly reference:

- project-context.md
- domain model
- repository layout
- protobuf strategy
- migration plan

---

# 88. AGENTS.md and .clinerules

If:

AGENTS.md

or:

.clinerules

exist:

do not modify them.

Follow their instructions.

Do not document rules that contradict them.

---

# 89. Existing Agent Instructions

Do not modify:

agents/

during this task.

The agent instruction hierarchy is separate from project documentation.

---

# 90. Generated Documentation

Do not create automatically generated documentation unless the project already uses such a mechanism.

---

# 91. API Documentation

Do not generate a complete API reference unless one already exists or another agent established it as a requirement.

---

# 92. Protobuf API Documentation

If protobuf contracts are already documented:

link to them.

Do not manually duplicate every message and RPC.

---

# 93. Deployment Documentation

The README deployment section should remain concise.

Detailed Render decisions belong in:

docs/platform-render-review.md

---

# 94. Docker Documentation

Detailed Docker decisions belong in:

docs/platform-docker-review.md

---

# 95. CI Documentation

Detailed CI decisions belong in:

docs/platform-ci-cd-review.md

---

# 96. Platform Documentation Review

Create:

docs/platform-documentation-review.md

Use exactly this structure:

# Platform Documentation Review

## 1. Objective

Explain the documentation work performed.

## 2. Required Documentation

List all documents read.

## 3. README Review

Document:

- previous state
- updates made
- sections added/updated
- obsolete content removed or corrected

## 4. Architecture Documentation

Document consistency with:

- domain model
- repository layout
- protobuf strategy
- migration plan

## 5. Platform Documentation

Document consistency with:

- repository audit
- protobuf generation
- HTTP gateway
- common packages
- CI/CD
- Docker
- Render

## 6. Developer Workflow

Document the verified workflow:

- environment setup
- generation
- migrations
- tests
- local execution
- Docker
- deployment

## 7. Links Verified

List important internal documentation links checked.

## 8. Findings

| ID | Severity | Area | Finding | Resolution |
|---|---|---|---|---|

## 9. Documentation Gaps

List anything that could not be documented because implementation information is missing.

## 10. Changes Made

List every modified file.

## 11. Documentation Check

Record the final documentation verification.

## 12. Final Status

Use exactly one:

PASS

PASS WITH FOLLOW-UP

BLOCKED

---

# 97. No Code Refactoring

Do not refactor code to improve documentation.

If documentation exposes a code inconsistency:

document it.

Do not fix it unless it is a trivial documentation/configuration correction.

---

# 98. No Architecture Changes

Do not modify:

- service boundaries
- database ownership
- protobuf design
- communication protocols
- deployment topology

---

# 99. No Platform Changes

Do not modify:

- Render architecture
- Docker architecture
- CI architecture
- gateway architecture

Those were handled by previous platform agents.

---

# 100. Final Repository Review

Run:

git status --short

Then:

git diff --stat

Then inspect every changed documentation file.

Do not inspect unrelated generated/vendor directories.

---

# 101. Changed Files

Expected changes should primarily be:

- README.md
- docs/platform-documentation-review.md

Only modify other documentation files if necessary to correct factual inconsistencies discovered during the task.

Do not modify source code.

---

# 102. Documentation Consistency Review

Verify that:

README.md

does not contradict:

docs/domain-model.md

docs/repository-layout.md

docs/protobuf-strategy.md

docs/migration-plan.md

or the platform review documents.

If a contradiction is discovered:

determine which document is authoritative.

Do not silently guess.

---

# 103. Final Documentation Check

Verify that all required documents still exist:

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
- docs/platform-docker-review.md
- docs/platform-render-review.md
- docs/platform-documentation-review.md

Record the result.

---

# 104. Completion Checklist

Before stopping:

- [ ] All required documentation was read.
- [ ] README.md was read before modification.
- [ ] agents/project-context.md was followed.
- [ ] Repository exploration was restricted.
- [ ] Deep folders were not recursively inspected.
- [ ] third_party/googleapis was not unnecessarily explored.
- [ ] Existing README content was preserved where useful.
- [ ] README reflects the current architecture.
- [ ] Service descriptions match implemented services.
- [ ] Repository layout documentation is accurate.
- [ ] Protobuf workflow is accurately documented.
- [ ] SQLC workflow is accurately documented.
- [ ] Migration workflow is accurately documented.
- [ ] Environment variables are accurately documented.
- [ ] No secrets were exposed.
- [ ] OAuth requirements are accurately documented.
- [ ] Webhook requirements are accurately documented.
- [ ] HTTP gateway is accurately documented.
- [ ] Docker workflow is accurately documented.
- [ ] CI/CD workflow is accurately documented.
- [ ] Render deployment is accurately documented.
- [ ] Developer workflow is documented.
- [ ] Internal documentation links were verified.
- [ ] No unsupported claims were added.
- [ ] No application code was modified unnecessarily.
- [ ] No architecture was redesigned.
- [ ] docs/platform-documentation-review.md was created.
- [ ] Final documentation check was completed.
- [ ] git status was inspected.
- [ ] git diff was inspected.

---

# Final Stop Condition

STOP after:

1. reading all required documentation,
2. reviewing the current README,
3. verifying the current architecture,
4. updating the README where necessary,
5. correcting only documentation that is demonstrably stale,
6. documenting the verified development workflow,
7. documenting the verified deployment workflow,
8. verifying internal documentation links,
9. checking for accidental secrets,
10. creating docs/platform-documentation-review.md,
11. completing the documentation check,
12. reviewing git status,
13. reviewing git diff.

Do NOT proceed to:

- observability implementation
- security implementation
- performance optimization
- application refactoring
- Docker redesign
- Render redesign
- CI/CD redesign
- protobuf changes
- database changes
- service implementation

Those belong to other agents.

STOP.