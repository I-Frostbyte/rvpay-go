# Agent 07 — Render Deployment

## Objective

Implement and validate the Render deployment configuration for the new RVPay architecture.

The objective is to ensure that the RVPay services and supporting infrastructure can be deployed to Render as independently managed production components while preserving the service boundaries defined by the project documentation.

This agent owns:

- Render Blueprint configuration
- Render service definitions
- Render deployment configuration
- Render environment-variable wiring
- Render PostgreSQL configuration
- Render service-to-service networking configuration
- Render health-check configuration
- Render deployment dependencies
- Render-specific operational configuration
- Render deployment documentation

This agent does NOT own:

- application business logic
- database schema design
- SQL queries
- sqlc
- protobuf contracts
- gRPC service implementation
- OAuth implementation
- webhook implementation
- Dockerfile implementation
- GitHub Actions workflow design
- observability architecture
- security architecture
- performance optimization

If another agent owns the underlying functionality:

do not implement it here.

Document the dependency instead.

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

Also inspect only the Render-related configuration and deployment files identified by the documentation above.

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

If any required document is missing:

STOP.

Do not recreate the missing document.

At the end of the task:

perform the documentation check again.

Create:

docs/platform-render-review.md

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

for coding and package conventions.

Use:

docs/platform-docker-review.md

to understand the Docker images and build contexts that this agent must deploy.

Use:

docs/platform-ci-cd-review.md

to understand the expected deployment/build workflow.

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

Render configuration does not require exploring the contents of the Google APIs submodule.

---

# 1. Existing Render Configuration Audit

Identify existing Render configuration from:

- README.md
- repository layout documentation
- Docker documentation
- CI/CD documentation

Look for:

- render.yaml
- render.yml
- Render Blueprint configuration
- deployment scripts
- Render-specific documentation
- environment-variable documentation

Do not assume a Render Blueprint already exists.

---

# 2. Determine Current Deployment Model

Determine whether RVPay currently uses:

- individual Render services
- Render Blueprint
- manually created Render services
- a mixture
- no Render configuration yet

Document the result.

---

# 3. Target Deployment Architecture

Use:

docs/repository-layout.md

and:

docs/domain-model.md

to determine the services that should be independently deployable.

Do not collapse independently deployable services into one Render service merely for convenience.

---

# 4. Blueprint Preference

The target RVPay deployment should use a Render Blueprint when the architecture contains multiple related services and infrastructure resources.

The Blueprint should describe the infrastructure as one coherent deployment configuration while preserving service independence.

The Blueprint should allow:

- services to be created consistently
- environment variables to be wired consistently
- databases to be referenced consistently
- deployment configuration to be version controlled
- future services to be added predictably

---

# 5. Blueprint File

If the project does not already contain a Render Blueprint:

create the appropriate:

render.yaml

or project-standard Render Blueprint file.

Do not create multiple competing Blueprint files.

---

# 6. Existing Blueprint

If a Blueprint already exists:

do not replace it blindly.

Audit it first.

Determine:

- which resources it defines
- which services it deploys
- which environment variables it defines
- which databases it references
- which services are legacy
- which services belong to the new architecture

Modify only what is required.

---

# 7. Service Inventory

Build a service inventory from the documentation.

For each deployable service determine:

- service name
- repository path
- Dockerfile path
- Docker build context
- binary
- exposed/listening port
- public/private requirement
- dependencies
- environment variables
- health check
- database dependency

Document this inventory in:

docs/platform-render-review.md

---

# 8. Clients Service

If the Clients service is part of the target architecture:

define a Render service for it.

Use the Dockerfile identified by:

docs/platform-docker-review.md

Do not create another Dockerfile.

---

# 9. Transactions Service

If the Transactions service is part of the target architecture:

define a Render service for it.

Use the Dockerfile identified by:

docs/platform-docker-review.md

Do not modify transaction application code.

---

# 10. Future Services

The Blueprint should be structured so additional services can be added without redesigning the deployment architecture.

Do not create placeholder services for services that do not yet exist.

Do not invent future services.

---

# 11. Public vs Private Services

Determine whether each service must be publicly reachable.

Use the architecture documentation.

A service should not be publicly exposed merely because it uses HTTP or gRPC.

Prefer private/internal networking for internal services where Render supports the required configuration.

---

# 12. HTTP Gateway

If the project contains an HTTP gateway:

determine whether it is:

- public
- internal
- optional

Use:

docs/platform-http-gateway-review.md

as the authority.

If the gateway is public:

configure the Render service accordingly.

Do not implement the gateway itself.

---

# 13. gRPC Services

Determine whether Render needs to expose gRPC directly.

Do not assume every gRPC service requires a public Render URL.

For internal communication:

prefer Render's internal networking where supported.

Document any platform limitation.

---

# 14. Public URLs

Do not hard-code Render-generated public URLs into application source code.

If application configuration requires a URL:

provide it through environment configuration.

---

# 15. Internal Service URLs

Where services communicate internally:

use Render's internal networking mechanism where supported.

Do not use:

localhost

for inter-service communication in production.

---

# 16. Localhost Rule

The following must NOT be used for production inter-service communication:

localhost

127.0.0.1

::1

unless the dependency genuinely runs inside the same container.

---

# 17. PostgreSQL

Determine the PostgreSQL resources required by the architecture.

Use:

docs/domain-model.md

and:

docs/migration-plan.md

to determine database boundaries.

Do not redesign database schemas.

---

# 18. Database Ownership

Determine whether the architecture expects:

- one shared PostgreSQL database
- one database per service
- multiple logical databases
- another documented arrangement

Follow the architecture documentation.

Do not create databases simply because multiple services exist.

---

# 19. Existing Render PostgreSQL

If an existing Render PostgreSQL database exists:

audit its documented purpose.

Do not delete it.

Do not migrate data.

Do not modify production database settings unless explicitly required.

---

# 20. Database Environment Variables

Where Render provides database connection information:

wire the application configuration through environment variables.

Do not hard-code:

- hostname
- username
- password
- port
- database name

into source code or Dockerfiles.

---

# 21. Internal Database URL

If the Render PostgreSQL resource exposes an internal database connection URL:

prefer the internal URL for services deployed within Render.

Do not use an externally routed database URL for internal Render-to-Render traffic unless required.

---

# 22. External Database URL

Do not expose the external database URL to services unless there is a concrete requirement.

Internal connectivity is preferred for Render-hosted services.

---

# 23. Database Credentials

Never place database credentials directly into:

- render.yaml
- Dockerfiles
- source code
- README.md
- committed configuration files

Use Render's secret/environment mechanisms.

---

# 24. Secret Environment Variables

Secrets should be marked and managed appropriately through Render.

Do not commit:

- passwords
- API keys
- OAuth client secrets
- webhook secrets
- signing keys
- SSO keys
- private keys

---

# 25. Environment Variable Inventory

For every Render service:

identify all required runtime environment variables.

Use:

- service configuration
- README.md
- project-context.md
- service documentation
- architecture documentation

Do not invent variable names.

---

# 26. Environment Variable Categories

Separate variables conceptually into:

1. service configuration
2. database configuration
3. internal service URLs
4. external API configuration
5. authentication credentials
6. OAuth credentials
7. webhook secrets
8. observability configuration

Only implement variables supported by the actual application.

---

# 27. Sensitive Variables

Ensure secrets are represented using Render's secret/environment configuration.

Do not place actual secret values into source control.

---

# 28. Non-Secret Variables

Non-sensitive configuration may be represented directly in the Blueprint where appropriate.

Examples may include:

- service mode
- environment name
- non-sensitive ports

Only use variables that actually exist.

---

# 29. Environment Separation

Determine whether the project expects:

- development
- staging
- production

Do not create an elaborate multi-environment Blueprint unless the project documentation requires it.

---

# 30. Production

The Blueprint should be suitable for production.

Production services must not depend on:

- local .env files
- localhost databases
- developer machines
- local Docker networks

---

# 31. Render Service Type

Use the appropriate Render resource type for each component.

Examples:

- web service
- private service
- PostgreSQL
- worker

Do not use a worker for a continuously running network server.

---

# 32. Long-Running Services

gRPC servers and HTTP servers are long-running processes.

Configure them as persistent services appropriate to the Render architecture.

Do not configure them as one-shot jobs.

---

# 33. Workers

Only use a Render worker where the application genuinely performs background processing without serving network traffic.

Do not create workers merely because a service performs asynchronous work internally.

---

# 34. Jobs

Do not introduce Render jobs for:

- migrations
- database maintenance
- tests

unless the architecture explicitly requires a separate deployment job.

---

# 35. Docker Deployment

Use the Docker deployment mechanism established by:

docs/platform-docker-review.md

Do not switch back to native Go builds.

---

# 36. Dockerfile Path

Ensure every Render Docker service references the correct Dockerfile.

The path must correspond to the repository layout.

Do not assume:

./Dockerfile

if the Dockerfile lives under a service directory.

---

# 37. Docker Build Context

Ensure Render uses the build context required by the Dockerfile.

This is especially important if the Dockerfile depends on:

- root go.mod
- root go.sum
- generated protobuf code
- common packages
- repository-wide files

Use:

docs/platform-docker-review.md

as the authority.

---

# 38. Root Context

If the Docker review requires repository-root context:

preserve it in Render.

Do not set the service directory as the build context simply because the Dockerfile is located there.

---

# 39. Service Name

Use stable, descriptive Render service names.

Do not use names tied to temporary development branches.

Follow existing project naming conventions.

---

# 40. Service Naming

Service names should make their role obvious.

Avoid names such as:

service1

backend

test-server

new-api

unless those names are already project conventions.

---

# 41. Region

Determine the intended Render region from project requirements.

If no region has been specified:

do not invent a business-critical regional strategy.

Use the documented/default project expectation and record the decision.

---

# 42. Instance Type

Do not optimize instance size in this agent.

Use the project's documented deployment size if one exists.

If no size is specified:

use a conservative documented default where appropriate.

Record the choice.

---

# 43. Scaling

Do not design horizontal autoscaling here.

Agent 11 owns performance considerations.

Only configure the minimum deployment behavior necessary to run the services.

---

# 44. Health Checks

Configure Render health checks only against endpoints that actually exist.

Do not invent:

/health

/status

/ready

unless the application actually implements them or another agent explicitly established them.

---

# 45. gRPC Health Checks

If the service exposes gRPC health checking:

determine whether Render can use it directly.

Do not create an HTTP health endpoint solely for Render without documenting the application change.

---

# 46. Health Check Port

Ensure the Render health check targets the correct service port.

Do not use an arbitrary port.

---

# 47. Health Check Path

If using HTTP:

use the documented health path.

Do not assume:

/

is a valid health endpoint.

---

# 48. Startup Time

Do not compensate for slow application startup by blindly setting very large health-check delays.

If startup is unexpectedly slow:

document the cause.

---

# 49. Database Startup

Do not assume PostgreSQL is immediately available during service startup.

The application should already handle database connection behavior according to its runtime design.

Do not implement retry logic here.

---

# 50. Database Migrations

Determine how migrations are intended to run.

Use:

docs/migration-plan.md

as the authority.

Do not invent a new migration strategy.

---

# 51. Migration Execution

If migrations are run by the application at startup:

ensure Render supplies:

RUN_MIGRATIONS

or the project's actual configuration variable.

Do not invent a variable.

---

# 52. Migration Safety

Do not enable automatic production migrations unless the architecture explicitly requires them.

If the documented architecture requires migration execution:

follow it exactly.

---

# 53. Migration Concurrency

If multiple service instances could attempt migrations simultaneously:

document the risk.

Do not redesign the migration system here.

---

# 54. Service Dependencies

Render service definitions should express deployment dependencies where supported.

Determine which services require:

- database
- internal gateway
- other services

Do not create unnecessary dependency chains.

---

# 55. Avoid Dependency Loops

Do not create circular Render dependencies.

For example:

Clients → Transactions → Clients

should not be introduced merely to express logical relationships.

---

# 56. Runtime Environment

Set the correct runtime environment variables.

Do not use build-time secrets.

---

# 57. Build Environment vs Runtime Environment

Understand the distinction:

Build environment:

used to create the Docker image.

Runtime environment:

used when the service executes.

Secrets required only at runtime must not be injected into the Docker build.

---

# 58. OAuth

For the Clients service:

ensure Render can provide the required OAuth configuration.

Examples may include:

- client ID
- client secret
- redirect URL
- scopes

Use the actual application variable names.

Do not implement OAuth code.

---

# 59. OAuth Redirect URL

The production redirect URL must not be:

localhost

or a developer-machine URL.

Use the public URL provided by the deployment architecture.

If the final URL depends on Render's generated hostname:

document the required post-deployment configuration.

---

# 60. Webhooks

For webhook-enabled services:

ensure the production webhook endpoint is externally reachable if required.

Do not make an internal-only service the public webhook endpoint.

---

# 61. Webhook URL

Do not hard-code a temporary Render URL into application source.

Use deployment configuration.

---

# 62. External Providers

If the application connects to:

- HighLevel
- payment providers
- external APIs

ensure the required credentials can be supplied through Render.

Do not place credentials into the Blueprint.

---

# 63. CORS

Do not implement CORS in the Render Blueprint.

If CORS configuration is required:

document it for the HTTP gateway/application layer.

---

# 64. TLS

Do not manually implement TLS termination in the application unless the architecture requires it.

Render normally handles public HTTPS at the platform edge.

Do not add certificates to the repository for Render unless explicitly required.

---

# 65. Custom Domains

Do not configure a production custom domain unless one is already documented.

If the application requires a stable domain for:

- OAuth
- webhooks
- SSO

document the requirement.

---

# 66. Render URL Dependencies

Do not make the architecture depend permanently on automatically generated hostnames if a stable custom domain is required by external integrations.

Document the distinction between:

development Render URL

and:

production public domain.

---

# 67. SSO

If Clients requires an SSO key:

ensure Render can supply it as a secret.

Do not place it into the Blueprint as plaintext.

---

# 68. Webhook Signing Secrets

Treat webhook secrets as runtime secrets.

Never commit them.

---

# 69. OAuth Client Secrets

Treat OAuth client secrets as runtime secrets.

Never commit them.

---

# 70. Database Passwords

Treat database passwords as runtime secrets.

Never commit them.

---

# 71. Render Blueprint Secrets

Do not place literal production credentials inside:

render.yaml

Even if the file is private.

---

# 72. Secret References

Use Render's supported secret/environment reference mechanism where appropriate.

Follow the actual Render Blueprint syntax supported by the project's deployment model.

Do not invent unsupported YAML keys.

---

# 73. Render YAML Validation

After modifying the Blueprint:

validate the YAML syntax.

Ensure:

- indentation is correct
- resource definitions are valid
- environment variable definitions are valid
- Docker configuration is valid

---

# 74. Unsupported Configuration

Do not add Render properties merely because they appear in an online example.

Only use configuration known to be supported by the Render deployment model being used.

---

# 75. Render API

Do not write scripts against the Render API unless the project explicitly requires them.

The Blueprint should be the primary infrastructure definition.

---

# 76. Render CLI

Do not require the Render CLI unless already used by the project.

---

# 77. Manual Dashboard Configuration

Prefer version-controlled Blueprint configuration for infrastructure that can be represented declaratively.

Document anything that must still be configured manually.

---

# 78. PostgreSQL Dashboard Configuration

Do not attempt to reproduce every Render PostgreSQL dashboard setting inside the Blueprint.

Only configure settings supported and required by the architecture.

---

# 79. Database Region

Keep application services and PostgreSQL in a compatible region where possible.

If the project has a documented region:

follow it.

---

# 80. Internal Networking

Use internal networking for service-to-service communication when supported.

Do not route internal traffic unnecessarily through public URLs.

---

# 81. Public Exposure

Minimize public exposure.

Only externally reachable components should be publicly exposed.

Examples:

- public HTTP gateway
- OAuth callback
- webhook endpoint

Internal:

- transaction service
- internal clients service
- internal gRPC services

where architecture permits.

---

# 82. Service-to-Service URLs

Use environment variables for internal service addresses.

Do not hard-code Render hostnames into Go source code.

---

# 83. Service Port Variables

Use the application's actual port configuration.

If Render supplies PORT:

ensure application configuration is compatible.

Do not modify application source unless required.

---

# 84. Logs

Render should receive application logs through stdout/stderr.

Do not configure Render to collect application log files.

---

# 85. Observability

Do not implement observability in this agent.

Agent 09 owns observability.

Only ensure the Render configuration does not interfere with application logging.

---

# 86. Metrics

Do not configure a metrics platform here unless already part of the documented Render deployment.

---

# 87. Security

Do not redesign application security here.

Agent 10 owns security.

Only ensure Render configuration does not expose secrets or unnecessarily expose private services.

---

# 88. Performance

Do not redesign scaling or caching here.

Agent 11 owns performance.

---

# 89. CI/CD

Review:

docs/platform-ci-cd-review.md

but do not redesign CI.

Render configuration should be compatible with the existing CI/deployment approach.

---

# 90. Deploy Hooks

If the project currently uses a Render deploy hook:

document it.

Do not create a new deploy hook merely because one is possible.

---

# 91. Render Deploy Hook Secrets

If a deploy hook is required:

it must be stored securely.

Do not commit the URL into the repository.

---

# 92. Auto Deploy

Determine whether the architecture expects Render to auto-deploy from the repository.

Use the documented CI/CD strategy.

Do not enable/disable automatic deployments arbitrarily.

---

# 93. Branch

Use the documented deployment branch.

Do not automatically assume:

main

unless the repository documentation establishes it.

---

# 94. Pull Requests

Do not create preview environments unless explicitly required.

---

# 95. Production Deployments

Do not trigger a production deployment merely to test YAML syntax.

Validate locally first.

If deployment testing is possible and safe:

perform it only according to project procedures.

---

# 96. Existing Production Resources

Do not delete or recreate existing Render resources blindly.

If the new Blueprint would conflict with existing resources:

document the migration path.

---

# 97. Migration from Existing Render Services

The new Blueprint should account for existing RVPay deployment resources.

Determine:

- what can be preserved
- what must be renamed
- what must be replaced
- what requires manual migration

Do not delete production resources.

---

# 98. Database Migration

Do not migrate PostgreSQL data in this agent.

Only document required infrastructure changes.

---

# 99. Zero-Downtime Assumptions

Do not claim zero-downtime deployment unless the architecture actually provides it.

---

# 100. Rollbacks

Document the expected Render rollback mechanism if relevant.

Do not implement custom rollback tooling.

---

# 101. Resource Naming

Use stable names across the Blueprint.

Avoid embedding:

- timestamps
- developer usernames
- temporary branch names

---

# 102. Environment Naming

Do not use developer-specific environment names in production.

---

# 103. Blueprint Comments

Keep YAML comments concise.

Explain only non-obvious deployment decisions.

---

# 104. Render Documentation

Create:

docs/platform-render-review.md

Use exactly this structure:

# Platform Render Review

## 1. Objective

Describe the target Render deployment architecture.

## 2. Required Documentation

List all documents read.

## 3. Existing Render Configuration

Document existing Render resources/configuration.

## 4. Target Service Inventory

| Service | Render Type | Public/Private | Dockerfile | Port | Database |
|---|---|---|---|---|---|

## 5. PostgreSQL Resources

| Resource | Purpose | Consumers | Internal Connection |
|---|---|---|---|

## 6. Environment Variables

| Service | Variable | Secret? | Source | Purpose |
|---|---|---|---|---|

Do not include secret values.

## 7. Service Dependencies

Describe service-to-service and database dependencies.

## 8. Networking

Describe:

- public services
- private services
- internal URLs
- external URLs

## 9. Health Checks

Document configured health checks.

## 10. Docker Integration

Document:

- Dockerfile paths
- build contexts
- image expectations

## 11. CI/CD Integration

Describe how Render interacts with the CI/CD architecture.

## 12. OAuth/Webhooks

Document production URL requirements for externally integrated endpoints.

## 13. Security Considerations

Document only Render-specific security considerations.

## 14. Existing Resource Migration

Document any existing Render resources that must be preserved or migrated.

## 15. Manual Configuration

Document anything that cannot be represented safely in the Blueprint.

## 16. Findings

| ID | Severity | Area | Finding | Resolution |
|---|---|---|---|---|

## 17. Deferred Work

Document work belonging to other agents.

## 18. Changes Made

List only files actually modified.

## 19. Documentation Check

Record the final documentation verification.

## 20. Final Status

Use exactly one:

PASS

PASS WITH FOLLOW-UP

BLOCKED

---

# 105. No Application-Code Changes

Do not modify Go source code unless a Render deployment requirement exposes an unavoidable configuration defect.

If application code appears incompatible with Render:

STOP.

Document:

- file
- problem
- expected change
- owning agent

Do not silently modify application logic.

---

# 106. No Dockerfile Changes

Agent 06 owns Docker.

If a Render deployment problem is caused by a Dockerfile:

document the required Docker change.

Only modify the Dockerfile if the change is trivial, clearly Docker-owned, and necessary to complete the Render configuration.

Prefer handing it back to Agent 06.

---

# 107. No CI Changes

Do not modify:

.github/workflows/

unless a Render-specific integration cannot function without a trivial configuration correction.

Agent 05 owns CI/CD.

---

# 108. No Protobuf Changes

Do not modify:

protobuf/

or generated gRPC packages.

---

# 109. No Database Schema Changes

Do not modify:

- migrations
- SQL queries
- sqlc configuration
- database schema

---

# 110. No OAuth Implementation

Do not implement:

- OAuth handlers
- token exchange
- state validation
- provider interfaces

Only configure the deployment environment required by the already implemented OAuth system.

---

# 111. No Webhook Implementation

Do not implement webhook handlers.

Only ensure the deployment exposes the required endpoint.

---

# 112. Render YAML Validation

Before finishing:

validate the Blueprint syntax.

Then inspect the final file manually.

Check:

- resources
- names
- Docker paths
- contexts
- ports
- environment variables
- database references
- health checks
- dependencies

---

# 113. Render Blueprint Consistency

Ensure the Blueprint matches:

docs/repository-layout.md

and:

docs/platform-docker-review.md

Any mismatch must be resolved or documented.

---

# 114. Service Count

Verify that every intended deployable service is represented.

Do not add services that are not part of the current architecture.

---

# 115. Database Count

Verify that PostgreSQL resources match the documented database architecture.

Do not create duplicate databases without architectural justification.

---

# 116. Environment Variable Consistency

For every service:

compare Render variables against the service configuration.

Look for:

- missing variables
- incorrect names
- obsolete variables
- secrets accidentally exposed
- incorrect service URLs

---

# 117. URL Consistency

Verify:

- public URLs are used only where required
- internal URLs are used for internal communication
- localhost is not used for production dependencies
- OAuth callback URLs are externally reachable
- webhook URLs are externally reachable

---

# 118. Dependency Consistency

Verify service dependencies against:

docs/domain-model.md

and:

docs/repository-layout.md

Do not introduce dependencies that do not exist in the architecture.

---

# 119. Existing Deposits Deployment

If the old Deposits service has an existing Render deployment:

use it as a migration reference.

Determine:

- existing service type
- existing environment variables
- existing Docker deployment
- existing database connection

Do not blindly copy obsolete configuration.

---

# 120. Legacy Compatibility

The new Render deployment must support the new architecture without unnecessarily breaking the currently functioning system.

Where migration requires replacing a legacy resource:

document the transition.

---

# 121. Final Deployment Readiness

Before marking PASS, verify:

- Dockerfiles exist
- Docker build contexts are correct
- services have valid Render definitions
- databases are defined/referenced correctly
- required environment variables are represented
- secrets are not committed
- internal networking is correctly represented
- public endpoints are correctly represented
- health checks target real endpoints
- service dependencies are valid
- Blueprint syntax is valid
- CI/CD expectations are documented
- no unrelated application code was changed

---

# 122. Final Git Review

Run:

git status --short

Then:

git diff --stat

Then inspect every changed file.

Pay particular attention to:

- render.yaml
- Render-related configuration
- docs/platform-render-review.md

Ensure no unrelated files were modified.

---

# 123. Documentation Check

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
- docs/platform-docker-review.md exists

Record this in:

docs/platform-render-review.md

---

# 124. Completion Checklist

Before stopping:

- [ ] All required documentation was read.
- [ ] README.md was used as the repository map.
- [ ] agents/project-context.md was followed.
- [ ] Repository exploration was restricted.
- [ ] Deep folders were not recursively inspected.
- [ ] third_party/googleapis was not unnecessarily explored.
- [ ] Existing Render configuration was audited.
- [ ] Existing production resources were preserved.
- [ ] Target service inventory was established.
- [ ] Each deployable service has the correct Render resource type.
- [ ] Dockerfile paths match Agent 06.
- [ ] Docker build contexts match Agent 06.
- [ ] PostgreSQL resources match the documented architecture.
- [ ] Internal database connectivity is configured correctly.
- [ ] Internal service networking is configured correctly.
- [ ] localhost is not used for production service dependencies.
- [ ] Public services are intentionally public.
- [ ] Private services remain private where appropriate.
- [ ] Required environment variables are represented.
- [ ] Secrets are not committed.
- [ ] OAuth deployment requirements are documented.
- [ ] Webhook deployment requirements are documented.
- [ ] Health checks reference real endpoints.
- [ ] Render Blueprint syntax was validated.
- [ ] CI/CD compatibility was checked.
- [ ] No unrelated application code was modified.
- [ ] docs/platform-render-review.md was created.
- [ ] Final documentation check was completed.
- [ ] git status was inspected.
- [ ] git diff was inspected.

---

# Final Stop Condition

STOP after:

1. reading all required documentation,
2. auditing existing Render configuration,
3. determining the target Render resources,
4. implementing the Render Blueprint/configuration,
5. wiring services to the correct Dockerfiles,
6. wiring databases correctly,
7. configuring runtime environment variables,
8. documenting secret requirements without exposing values,
9. configuring internal/public networking,
10. configuring valid health checks,
11. documenting OAuth and webhook deployment requirements,
12. validating the Blueprint,
13. completing the documentation check,
14. inspecting the final git diff.

Do NOT proceed to:

- Docker redesign
- CI/CD redesign
- application implementation
- protobuf implementation
- database redesign
- OAuth implementation
- webhook implementation
- observability implementation
- security architecture
- performance optimization

Those belong to other agents.

STOP.