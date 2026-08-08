# Agent 10 — Security

## Objective

Audit and strengthen the security foundation of the RVPay platform.

The objective is to ensure that the platform:

- handles secrets safely
- protects authentication credentials
- validates external requests
- protects OAuth credentials
- protects webhook endpoints
- uses appropriate transport security
- avoids leaking sensitive information
- validates configuration safely
- applies appropriate authorization boundaries
- maintains secure service-to-service communication
- follows the existing RVPay architecture

Security improvements must be implemented within the architecture already established by:

- the foundation documents
- Clients service
- Transactions service
- platform agents 01–09

Do NOT redesign the architecture.

Do NOT introduce unrelated security products.

Do NOT replace working authentication mechanisms merely because another mechanism is more fashionable.

Do NOT perform performance optimization.

Do NOT perform general refactoring.

Do NOT modify unrelated application behavior.

Security changes must be concrete, justified, minimal, and compatible with the existing repository.

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
- docs/platform-documentation-review.md
- docs/platform-observability-review.md

Also inspect only the specific source/configuration files referenced by those documents when security verification requires them.

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
- docs/platform-documentation-review.md
- docs/platform-observability-review.md

If any required document is missing:

STOP.

Do not recreate the missing document.

At the end of the task:

perform the documentation check again.

Create:

docs/platform-security-review.md

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

as the authority for repository structure.

Use:

agents/project-context.md

as the authority for coding and package conventions.

Use:

docs/platform-repository-audit.md

to understand what has already been inspected.

Use:

docs/platform-observability-review.md

to understand the existing observability implementation.

Only inspect security-relevant implementation files.

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

Do not spend time inspecting generated protobuf internals.

---

# 1. Security Philosophy

Security work must answer:

What are we protecting?

From whom?

At which boundary?

How is the protection implemented?

How is failure handled?

How is the behavior tested?

Do not add security mechanisms without identifying the boundary they protect.

---

# 2. Security Boundaries

Identify the major boundaries:

- public HTTP traffic
- HTTP gateway
- gRPC services
- service-to-service communication
- database access
- external provider APIs
- OAuth callbacks
- webhook endpoints
- Render infrastructure
- developer/local environments

Document which security controls apply to each.

---

# 3. Existing Security Audit

Before making changes:

inspect existing security mechanisms.

Look for:

- authentication
- authorization
- OAuth
- webhook verification
- API keys
- bearer tokens
- TLS assumptions
- secret handling
- environment variables
- database credentials
- CORS
- request validation
- input validation
- rate limiting
- security headers
- gRPC metadata
- service authentication

Do not assume these mechanisms are absent.

---

# 4. Preserve Existing Security

If an existing security mechanism is correct:

preserve it.

Do not replace working code without a concrete reason.

---

# 5. Secrets

Secrets must never be committed to the repository.

Never store production secrets in:

- Go source
- YAML
- JSON
- Markdown
- Dockerfiles
- Makefiles
- protobuf files
- migration files
- tests

---

# 6. Environment Variables

Secrets should be supplied through environment/configuration mechanisms appropriate to the deployment environment.

Examples include:

- database credentials
- OAuth client secrets
- API keys
- webhook secrets
- SSO keys
- JWT signing secrets

Use the actual project configuration mechanism.

Do not invent a second configuration system.

---

# 7. .env Files

Treat:

.env

as sensitive.

Do not commit it.

Do not add real credentials to:

.env.example.

Use placeholders.

---

# 8. Gitignore

Verify that sensitive local configuration is excluded from Git where appropriate.

Check:

.gitignore

Do not modify unrelated ignore rules.

---

# 9. Secret Detection

Search only relevant configuration/documentation files for obvious secret leakage.

Look for patterns involving:

- passwords
- API keys
- bearer tokens
- OAuth secrets
- database URLs
- private keys
- signing secrets

Do not perform a huge repository-wide scan.

---

# 10. Documentation Secrets

Never place real secrets in:

- README.md
- docs/
- agent review documents
- code comments

Use:

<REDACTED>

or:

<your-secret>

style placeholders.

---

# 11. Logging Security

Use:

docs/platform-observability-review.md

to understand logging behavior.

Ensure secrets are not written to:

- application logs
- traces
- metrics
- error messages

---

# 12. Authorization Headers

Never log:

Authorization

headers.

This includes:

- Bearer tokens
- API credentials
- provider credentials

---

# 13. OAuth Tokens

Never log:

- authorization codes
- access tokens
- refresh tokens
- client secrets

---

# 14. OAuth Client Credentials

Verify that OAuth client credentials are loaded securely.

The application must not hard-code:

- client ID where it should be configurable
- client secret
- signing secret

Client IDs may be non-secret in some systems, but follow the actual provider requirements.

---

# 15. OAuth Callback

Inspect the OAuth callback implementation.

Verify:

- authorization code is handled securely
- callback state is validated where applicable
- tokens are not exposed in URLs beyond the provider's callback mechanism
- token exchange occurs server-side
- errors do not expose credentials

Do not redesign the OAuth flow unless a concrete vulnerability exists.

---

# 16. OAuth State

If OAuth uses a state parameter:

verify that it protects against CSRF.

If it does not:

determine whether the current architecture requires adding it.

Do not introduce an incompatible OAuth flow.

---

# 17. OAuth Redirect URLs

Validate OAuth redirect URL configuration.

Do not allow arbitrary user-controlled redirect destinations.

Use configured callback URLs.

---

# 18. Authorization Code Handling

Authorization codes must:

- be exchanged server-side
- not be logged
- not be persisted unnecessarily
- not be returned to unrelated clients

---

# 19. Token Storage

Inspect how provider tokens are stored.

Determine:

- where tokens are stored
- whether they are encrypted/protected
- who can access them
- whether they are unnecessarily returned through APIs

Do not expose provider access tokens through ordinary service responses.

---

# 20. Refresh Tokens

Refresh tokens are sensitive credentials.

They must not appear in:

- logs
- protobuf responses unless explicitly required
- HTTP responses
- error messages
- documentation

---

# 21. SSO Key

Treat the SSO key as a secret.

Never log it.

Never commit it.

Never return it through an API.

---

# 22. SSO Token Handling

If the application receives encrypted authentication material:

verify that:

- decryption happens server-side
- the secret key is securely loaded
- decrypted credentials are not logged
- invalid data is rejected safely

Do not expose decrypted authentication material.

---

# 23. Webhook Security

Inspect all webhook endpoints.

Determine whether the provider supplies:

- signature
- secret
- timestamp
- token
- authentication header

Use the actual provider mechanism implemented by the project.

---

# 24. Webhook Verification

Webhook requests must not be trusted merely because they reach a known URL.

Where the provider supports signing:

verify the signature before processing.

---

# 25. Webhook Replay Protection

If the provider provides timestamps/nonces:

consider replay protection.

Do not implement speculative replay systems without provider support.

---

# 26. Webhook Timing Attacks

If signature comparison is implemented:

use constant-time comparison where appropriate.

Do not use ordinary string comparison for cryptographic signatures if timing attacks are relevant.

---

# 27. Webhook Payload Validation

Validate webhook input before processing.

Reject:

- malformed payloads
- invalid signatures
- unsupported events
- missing required fields

Use appropriate HTTP status responses.

---

# 28. Webhook Idempotency

Security and reliability overlap here.

If webhook events may be retried:

ensure the implementation does not create unintended duplicate effects.

Do not redesign transaction processing unless required.

---

# 29. Public Endpoints

Identify every intentionally public endpoint.

For each:

determine:

- why it is public
- what authentication it requires
- what input validation it performs
- what data it exposes

---

# 30. Internal Endpoints

Internal endpoints should not accidentally become public.

Verify Render/gateway configuration where applicable.

Do not expose internal services through the public gateway without justification.

---

# 31. HTTP Gateway

Use:

docs/platform-http-gateway-review.md

to understand the gateway.

Verify:

- public routes
- internal routes
- authentication boundaries
- CORS
- headers
- request validation

Do not redesign gateway routing.

---

# 32. CORS

If browser clients access the API:

verify CORS configuration.

Do not use:

Access-Control-Allow-Origin: *

for credentialed/authenticated applications unless explicitly justified.

Prefer configured origins.

---

# 33. CORS Configuration

CORS origins must be configurable where required.

Do not hard-code temporary development domains into production behavior.

---

# 34. Security Headers

If the gateway serves browser-facing traffic:

consider appropriate headers.

Examples:

- Content-Security-Policy
- X-Content-Type-Options
- Referrer-Policy
- Strict-Transport-Security

Only implement headers appropriate to the actual gateway behavior.

Do not add headers that break legitimate API behavior.

---

# 35. TLS

Production traffic must use HTTPS/TLS where supported.

Do not implement custom TLS termination inside application services if Render/gateway infrastructure already provides TLS.

---

# 36. Internal gRPC

Determine whether internal gRPC traffic is:

- private network traffic
- public traffic
- authenticated traffic

Use the actual Render/deployment topology.

Do not assume public exposure is required.

---

# 37. gRPC Authentication

If internal gRPC endpoints are externally reachable:

ensure they cannot be called anonymously unless explicitly intended.

If authentication is required:

use the architecture's existing authentication mechanism.

Do not invent a proprietary authentication protocol.

---

# 38. gRPC Metadata

If authentication information is passed through gRPC metadata:

ensure:

- secrets are not logged
- metadata is validated
- invalid credentials are rejected

---

# 39. Service-to-Service Authentication

If services communicate directly:

determine how they authenticate.

If the current architecture does not require service authentication because services are isolated behind private networking:

document that boundary.

Do not add unnecessary complexity.

---

# 40. Database Security

Verify:

- credentials are environment-provided
- production credentials are not committed
- connections use the correct deployment configuration
- database endpoints are not publicly exposed unnecessarily

---

# 41. Database Credentials

Never log full database connection URLs if they contain credentials.

If connection information must be logged:

log safe metadata only.

---

# 42. Database SSL

Determine whether production PostgreSQL requires SSL.

Use the actual Render PostgreSQL configuration.

Do not force an incompatible local development configuration.

---

# 43. SQL Injection

Review SQL inputs where security-relevant.

Because RVPay uses SQLC:

ensure user input is passed as parameters rather than concatenated into SQL.

Do not rewrite generated SQLC code.

---

# 44. Dynamic SQL

Pay special attention to:

- dynamically constructed queries
- ORDER BY clauses
- table names
- filtering
- search expressions

If dynamic SQL exists:

verify that untrusted input cannot directly alter SQL syntax.

---

# 45. Input Validation

Validate external input at the service boundary.

Examples:

- IDs
- phone numbers
- email addresses
- amounts
- currency codes
- provider identifiers
- pagination values
- webhook payloads

Use domain rules established by the foundation documents.

---

# 46. Input Size

Prevent unreasonably large inputs where appropriate.

Do not introduce arbitrary restrictive limits without justification.

---

# 47. Integer/Amount Validation

Financial values require special care.

Verify:

- negative values are rejected where invalid
- zero values are rejected where invalid
- currency precision is respected
- integer overflow cannot create invalid amounts

Do not redesign the financial model.

---

# 48. Identifier Validation

Validate identifiers before database/provider operations.

Do not trust IDs supplied by external clients.

---

# 49. Error Responses

Public errors must not reveal:

- stack traces
- database credentials
- SQL statements containing secrets
- internal filesystem paths
- environment variables
- OAuth tokens
- provider credentials

---

# 50. Internal Error Logging

Detailed technical errors may be logged internally.

Public responses should expose only appropriate information.

---

# 51. gRPC Error Mapping

Ensure internal errors are mapped to appropriate gRPC status codes.

Do not leak internal errors directly when they contain sensitive information.

---

# 52. HTTP Error Mapping

Ensure HTTP errors do not expose internal implementation details.

Use appropriate status codes.

---

# 53. Authentication Failures

Authentication failures should not reveal whether sensitive credentials partially matched.

Avoid unnecessarily detailed responses such as:

"user exists but password is wrong."

Follow the application's actual authentication model.

---

# 54. Authorization

Identify authorization boundaries.

Determine:

- who can access merchant data
- who can access customer data
- who can initiate transactions
- who can access provider credentials
- who can administer services

Use the domain model and existing architecture.

---

# 55. Object-Level Authorization

Where APIs accept resource IDs:

verify that the caller is authorized to access that resource.

Do not assume that possession of an ID is sufficient authorization.

---

# 56. Merchant Isolation

Transactions and client/integration data must not accidentally cross merchant boundaries.

Where merchant ownership exists:

ensure queries and service methods enforce the appropriate scope.

Do not redesign database ownership.

---

# 57. Customer Data

Customer records must only be accessible within the appropriate merchant/application scope.

Do not expose unrelated customers through broad queries.

---

# 58. Transaction Data

Transaction access must respect ownership boundaries.

Do not return transaction information solely because a valid transaction ID was supplied.

---

# 59. Provider Credentials

Provider credentials must only be accessible to the code paths that require them.

Do not expose them through generic repository methods or APIs.

---

# 60. Mass Assignment

Do not allow clients to arbitrarily populate sensitive fields.

Examples:

- merchant IDs
- ownership fields
- internal status
- authorization state
- provider credentials
- timestamps

Explicitly control fields accepted from clients.

---

# 61. Status Fields

Clients should not be allowed to arbitrarily change system-controlled statuses.

Examples:

- transaction status
- OAuth installation status
- webhook state
- payout state

---

# 62. Database Ownership

Do not allow clients to choose arbitrary database identifiers where the server should determine ownership.

---

# 63. Pagination

Validate pagination inputs.

Prevent:

- negative offsets
- unreasonable limits
- malformed cursors

Do not optimize pagination here.

---

# 64. Rate Limiting

Determine whether rate limiting already exists.

If it is explicitly required by the platform architecture:

implement the minimum appropriate mechanism.

Do not introduce an elaborate distributed rate limiter unless required.

---

# 65. Authentication Endpoints

If authentication endpoints exist:

protect them against:

- brute force
- credential enumeration
- token leakage

Only implement controls relevant to the actual authentication architecture.

---

# 66. OAuth Rate Limiting

Do not allow uncontrolled OAuth authorization/token exchange requests.

Use provider-supported behavior where appropriate.

Do not invent arbitrary quotas.

---

# 67. Webhook Rate Limiting

Do not implement rate limiting that could cause legitimate provider retries to be rejected unless carefully justified.

---

# 68. Dependency Security

Do not blindly upgrade every dependency.

Only address security-relevant dependency issues that are verified.

---

# 69. Go Dependencies

If a dependency has a known security issue:

document it and update it only if the update is compatible with the project.

Do not perform broad dependency upgrades.

---

# 70. Docker Security

Review:

docs/platform-docker-review.md

for the current container setup.

Check:

- non-root execution where appropriate
- unnecessary packages
- secret handling
- exposed ports
- filesystem permissions

Do not redesign Docker.

---

# 71. Docker Secrets

Never bake secrets into Docker images.

Do not put secrets in:

- Dockerfile ENV
- Docker build ARG
- image layers
- source code

---

# 72. Container User

If the current Docker setup runs as root:

determine whether changing to a non-root user is safe.

If implemented:

ensure the application still has access to required files/directories.

Do not make the container change solely for stylistic reasons.

---

# 73. Exposed Ports

Expose only ports required by the service.

Do not expose:

- PostgreSQL
- internal admin ports
- debug ports

unless explicitly required.

---

# 74. Debug Endpoints

Ensure production does not expose:

- pprof
- debug endpoints
- development admin endpoints

unless explicitly protected.

---

# 75. Development vs Production

Security settings may differ between:

development

and:

production

but the distinction must be explicit.

Do not make production insecure merely to simplify local development.

---

# 76. Render Security

Use:

docs/platform-render-review.md

to verify:

- environment variable handling
- service exposure
- private/public services
- database exposure
- TLS
- health checks

Do not change Render architecture.

---

# 77. CI Security

Use:

docs/platform-ci-cd-review.md

to inspect:

- secret handling
- GitHub Actions permissions
- dependency installation
- build steps
- secret exposure

Do not redesign CI/CD.

---

# 78. CI Secrets

Secrets must come from:

GitHub Actions secrets

or:

approved environment mechanisms.

Never commit CI credentials.

---

# 79. CI Logs

Ensure secrets are not echoed by build commands.

Do not print:

- tokens
- passwords
- private keys
- database URLs

---

# 80. Git History

Do not rewrite Git history.

If a secret is found in current files:

remove it from the working tree.

If a historical secret is suspected:

document the issue.

Do not run destructive history rewriting automatically.

---

# 81. Security Tests

Add tests for security behavior that the implementation changes.

Examples:

- unauthorized request rejected
- invalid signature rejected
- invalid OAuth state rejected
- invalid input rejected
- unauthorized resource access rejected
- secret values absent from logs where testable

---

# 82. Security Test Isolation

Tests must not require:

- production credentials
- real OAuth providers
- real webhook providers
- Render
- production databases

unless explicitly defined as integration tests.

---

# 83. Webhook Tests

Where applicable, test:

- valid signature accepted
- invalid signature rejected
- malformed payload rejected
- replay protection if implemented

---

# 84. OAuth Tests

Where applicable, test:

- invalid state rejected
- callback errors handled
- token exchange failures handled
- credentials not exposed

Do not test against real provider credentials.

---

# 85. Authorization Tests

Where authorization exists:

test:

- authorized access succeeds
- unauthorized access fails
- cross-merchant access fails

Use test fixtures.

---

# 86. Input Validation Tests

Test security-sensitive boundary validation.

Include relevant:

- invalid IDs
- invalid amounts
- invalid statuses
- malformed payloads
- oversized inputs

Do not create exhaustive tests for every field.

---

# 87. Error Leakage Tests

Where practical:

verify that public errors do not expose:

- SQL
- credentials
- filesystem paths
- stack traces
- provider tokens

---

# 88. Logging Security Tests

Where practical:

verify that sensitive values do not appear in security-sensitive logs.

Do not build a complex logging test framework.

---

# 89. No Sensitive Test Fixtures

Do not use real:

- API keys
- OAuth tokens
- passwords
- production database URLs
- customer credentials

in tests.

---

# 90. Security Headers Tests

If headers are added:

test that required headers are present.

Do not test the behavior of the HTTP library itself.

---

# 91. TLS Tests

Do not create elaborate TLS integration tests if TLS is terminated by Render.

Document the deployment boundary instead.

---

# 92. No Custom Cryptography

Do not implement custom cryptographic algorithms.

Use established Go/provider libraries.

If cryptography is already implemented:

verify it uses established primitives.

---

# 93. Password Handling

If passwords exist anywhere:

do not store plaintext passwords.

Use established password hashing mechanisms.

If RVPay does not handle passwords:

do not introduce password infrastructure.

---

# 94. Token Encryption

If provider tokens require encryption at rest:

use established cryptographic libraries and key management.

Do not invent an encryption scheme.

If the architecture does not require application-level encryption:

document the existing storage/security boundary instead of adding unnecessary encryption.

---

# 95. Key Management

Never hard-code encryption/signing keys.

Keys must come from secure configuration.

Do not commit development keys that could be mistaken for production keys.

---

# 96. Cryptographic Randomness

Security-sensitive random values must use:

crypto/rand

or an established secure mechanism.

Do not use:

math/rand

for:

- tokens
- secrets
- nonces
- authorization state
- password reset tokens

---

# 97. Constant-Time Comparisons

Use constant-time comparisons for cryptographic secrets/signatures where appropriate.

Do not use ordinary string comparison for sensitive signature verification.

---

# 98. URL Security

Validate URLs where user-controlled URLs are accepted.

Prevent:

- arbitrary redirects
- unintended internal access
- malformed callback targets

Do not implement a generic SSRF framework unless required.

---

# 99. SSRF

Identify any server-side HTTP request that accepts a user-controlled URL.

If such functionality exists:

ensure the allowed destinations are constrained appropriately.

Do not introduce arbitrary outbound URL support.

---

# 100. Provider URLs

External provider API URLs should come from controlled configuration.

Do not allow ordinary users to select arbitrary provider API endpoints.

---

# 101. HTTP Client Security

External HTTP clients should:

- use HTTPS where required
- have reasonable timeouts
- not follow unsafe redirects blindly where credentials could leak
- avoid logging sensitive headers

Do not perform performance tuning.

---

# 102. Authentication Token Placement

Prefer secure transport mechanisms appropriate to the API.

Do not put bearer tokens into:

- URL query parameters
- logs
- error messages

unless the external provider explicitly requires it.

---

# 103. Query Parameters

Do not place sensitive information into URLs.

URLs commonly appear in:

- logs
- tracing
- browser history
- monitoring systems

---

# 104. Database Queries

Verify that security-sensitive filtering happens server-side.

Do not rely exclusively on clients to supply ownership filters.

---

# 105. Repository Boundaries

Repositories should not expose broad unrestricted access to sensitive records unless the service layer explicitly controls access.

---

# 106. Service Boundaries

Do not allow one service to bypass another service's security boundary through direct database access unless that architecture explicitly permits it.

---

# 107. Cross-Service Database Access

If services have separate database ownership:

respect those boundaries.

If the current architecture uses a shared database:

do not redesign it.

Document the security implications.

---

# 108. Audit Logging

Determine whether security-sensitive actions require audit logging.

Examples:

- OAuth installation
- OAuth removal
- credential changes
- transaction initiation
- payout initiation
- administrative actions

If audit logging is required by the architecture:

implement only the necessary events.

Do not build a full SIEM.

---

# 109. Audit Data

Audit logs must not contain:

- access tokens
- passwords
- API keys
- full payment credentials

Store only the information needed to establish:

who

did:

what

and:

when.

---

# 110. Time Synchronization

If security mechanisms rely on timestamps:

use server-side time.

Do not trust client-provided timestamps without validation.

---

# 111. Replay Protection

For signed/authenticated requests:

consider whether replay protection is required.

Use provider-supported timestamps/nonces where available.

---

# 112. Security Documentation

Update documentation only where security behavior is developer-visible.

Do not rewrite the README.

Create:

docs/platform-security-review.md

as the detailed security implementation record.

---

# 113. Security Review Document

Create:

docs/platform-security-review.md

Use exactly this structure:

# Platform Security Review

## 1. Objective

Describe the security objectives.

## 2. Required Documentation

List every required document read.

## 3. Existing Security

Describe the security mechanisms already present.

## 4. Security Boundaries

Document:

- public HTTP
- gateway
- gRPC
- services
- database
- provider APIs
- OAuth
- webhooks
- Render

## 5. Secrets

Document:

- secret sources
- secret handling
- secret exposure prevention

## 6. Authentication

Document authentication mechanisms actually implemented.

## 7. Authorization

Document authorization boundaries actually implemented.

## 8. OAuth Security

Document:

- state
- redirect handling
- authorization code handling
- token storage
- token protection

## 9. Webhook Security

Document:

- signature verification
- replay protection
- payload validation
- idempotency considerations

## 10. Input Validation

Document security-sensitive validation.

## 11. Database Security

Document:

- credentials
- TLS
- SQL parameterization
- access boundaries

## 12. HTTP/gRPC Security

Document:

- transport
- authentication
- authorization
- error handling
- headers

## 13. Container Security

Document:

- user
- secrets
- ports
- debug endpoints

## 14. CI/CD Security

Document:

- secrets
- GitHub Actions permissions
- build logs

## 15. Render Security

Document:

- environment variables
- service exposure
- TLS
- database exposure

## 16. Tests

Document security tests added or modified.

## 17. Findings

| ID | Severity | Area | Finding | Resolution |
|---|---|---|---|---|

## 18. Remaining Risks

List security risks that remain outside this agent's scope.

Do not pretend they were solved.

## 19. Documentation Changes

List modified documentation.

## 20. Documentation Check

Record the final documentation verification.

## 21. Final Status

Use exactly one:

PASS

PASS WITH FOLLOW-UP

BLOCKED

---

# 114. Security Severity

Use:

CRITICAL

for vulnerabilities that could directly compromise credentials, funds, or broad system access.

HIGH

for serious exploitable vulnerabilities.

MEDIUM

for meaningful weaknesses with limited impact.

LOW

for minor weaknesses.

INFO

for observations.

Do not inflate severity.

---

# 115. Security Findings

Every finding should contain:

- what is wrong
- why it matters
- where it occurs
- what was changed
- how it was tested

---

# 116. No False Security Claims

Do not write:

"secure"

"fully secure"

"production secure"

"zero vulnerabilities"

unless the claim is justified by the actual verification performed.

---

# 117. No Compliance Claims

Do not claim:

- PCI compliance
- SOC 2 compliance
- GDPR compliance
- ISO certification
- regulatory compliance

unless the repository contains authoritative evidence.

---

# 118. No Security Product Shopping

Do not introduce:

- WAF
- SIEM
- secrets manager
- API gateway
- IAM platform
- vulnerability scanner

unless already established by the architecture or explicitly required.

---

# 119. Dependency Changes

If dependency changes are necessary:

make the smallest compatible change.

Document:

- dependency
- reason
- version
- security issue addressed

Do not upgrade unrelated dependencies.

---

# 120. Generated Code

Do not manually modify:

- generated protobuf files
- generated sqlc files

---

# 121. Protobuf

Do not change protobuf contracts solely for security.

Use:

- context
- metadata
- transport-level authentication

where appropriate.

---

# 122. Database

Do not modify schema unless a concrete security requirement requires it.

If schema changes are necessary:

follow the project's established migration process.

Do not manually modify generated sqlc code.

---

# 123. Migration Security

Database migrations must not contain:

- production credentials
- secret values
- environment-specific passwords

---

# 124. Build Verification

After implementation:

run the repository's normal build command.

Use commands documented by:

README.md

and:

agents/project-context.md.

Do not invent commands.

---

# 125. Test Verification

Run the appropriate test suite.

At minimum:

run tests affected by security changes.

If the repository has a documented standard test command:

use it.

---

# 126. Formatting

Run the repository's established formatting workflow.

Do not introduce a new formatter.

---

# 127. Static Analysis

Run the project's established static analysis if documented.

Do not install unrelated analysis tools.

---

# 128. Git Status

Before finishing:

run:

git status --short

---

# 129. Git Diff

Then run:

git diff --stat

Then:

git diff

Review every change made by this agent.

---

# 130. Unexpected Changes

If unrelated files are already modified:

do not overwrite them.

Do not reset the repository.

---

# 131. Security Review Scope

The final review must specifically verify:

- authentication
- authorization
- OAuth
- webhooks
- secrets
- input validation
- database access
- HTTP
- gRPC
- Docker
- CI/CD
- Render
- logging

---

# 132. Observability Integration

Verify that security-sensitive information is not exposed by the observability mechanisms implemented by Agent 09.

Check:

- logs
- traces
- metrics
- errors
- request metadata

---

# 133. Documentation Check

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
- docs/platform-observability-review.md
- docs/platform-security-review.md

Record the result in:

docs/platform-security-review.md

---

# 134. Completion Checklist

Before stopping:

- [ ] All required documents were read.
- [ ] README.md was read.
- [ ] agents/project-context.md was followed.
- [ ] Repository exploration was restricted.
- [ ] Deep folders were not recursively inspected.
- [ ] third_party/googleapis was not unnecessarily explored.
- [ ] Existing authentication was audited.
- [ ] Existing authorization was audited.
- [ ] OAuth security was audited.
- [ ] Webhook security was audited.
- [ ] Secrets were audited.
- [ ] .env handling was audited.
- [ ] .gitignore was checked.
- [ ] Logging was checked for secret leakage.
- [ ] Trace/metric exposure was considered.
- [ ] HTTP security was audited.
- [ ] gRPC security was audited.
- [ ] Database security was audited.
- [ ] SQL injection risks were checked.
- [ ] Input validation was audited.
- [ ] Error leakage was checked.
- [ ] Provider credentials were protected.
- [ ] SSO credentials were protected.
- [ ] Docker security was audited.
- [ ] CI/CD security was audited.
- [ ] Render security was audited.
- [ ] Security tests were added where necessary.
- [ ] No real secrets were added.
- [ ] No custom cryptography was introduced.
- [ ] No protobuf contracts were unnecessarily changed.
- [ ] No generated code was manually modified.
- [ ] No database redesign was performed.
- [ ] No performance optimization was performed.
- [ ] docs/platform-security-review.md was created.
- [ ] Documentation check was completed.
- [ ] Build verification was completed.
- [ ] Tests were completed.
- [ ] Formatting was completed.
- [ ] git status was reviewed.
- [ ] git diff was reviewed.

---

# Final Stop Condition

STOP after:

1. reading all required documentation,
2. auditing the current security posture,
3. identifying concrete security weaknesses,
4. implementing only justified security improvements,
5. protecting secrets,
6. protecting OAuth credentials,
7. protecting webhook endpoints,
8. validating relevant external input,
9. verifying authentication and authorization boundaries,
10. verifying database security,
11. verifying HTTP/gRPC security,
12. verifying Docker/CI/Render security,
13. ensuring observability does not leak sensitive information,
14. adding meaningful security tests,
15. creating docs/platform-security-review.md,
16. completing the documentation check,
17. running the appropriate verification commands,
18. reviewing git status,
19. reviewing git diff.

Do NOT proceed to:

- performance optimization
- caching
- load testing
- database performance tuning
- query optimization
- Docker redesign
- Render redesign
- CI/CD redesign
- protobuf redesign
- unrelated refactoring

STOP.