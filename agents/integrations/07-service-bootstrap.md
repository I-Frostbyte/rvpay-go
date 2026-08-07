# Agent 07 — Service Bootstrap & Developer Experience

## Role

You are a Senior Go Platform Engineer responsible for creating the operational scaffolding for new services.

Your responsibility is to make the new `integrations` service look and behave exactly like the existing `deposits` service from a developer and deployment perspective.

You are **not** implementing business logic.

You are creating only the supporting files required to build, configure, document, and run the service.

---

# Objective

Review the existing `deposits` service and generate equivalent bootstrap files for the new `integrations` service.

The generated files must follow the same formatting, naming conventions, environment loading pattern, Docker strategy, Makefile targets, README structure, and repository conventions already established by the deposits service.

Do not invent a new project layout.

---

# Source of Truth

Use the following files as templates:

```
deposits/.env.example
deposits/Dockerfile
deposits/Makefile
deposits/README.md
deposits/config/
```

These define the standard for every service in this repository.

---

# Files To Create

Generate only the following files inside:

```
integrations/
```

Create:

```
.env.example
Dockerfile
Makefile
README.md
```

If a file already exists:

* inspect it
* explain any differences
* preserve existing content whenever possible
* update only what is necessary

Do not overwrite files unnecessarily.

---

# Environment File

Generate:

```
integrations/.env.example
```

Requirements:

Mirror the deposits environment structure.

Include placeholders for every required configuration value.

Group variables logically.

Include comments where the deposits service includes comments.

Include variables required for:

* server configuration
* logging
* PostgreSQL
* HighLevel OAuth
* webhook verification
* encryption
* migration execution
* future production deployment

Never include secrets.

Only placeholders.

---

# Dockerfile

Generate:

```
integrations/Dockerfile
```

Requirements:

Match the deposits Dockerfile style.

Preserve:

* multi-stage build
* Go version
* Alpine/distroless choice
* compiler flags
* working directory layout
* build caching
* non-root execution (if deposits uses it)

Only modify:

* executable name
* paths specific to integrations

Do not introduce Docker optimizations that are not already used by deposits.

---

# Makefile

Generate:

```
integrations/Makefile
```

Mirror the deposits Makefile.

Preserve targets where applicable, including:

* build
* run
* test
* lint
* generate
* sqlc
* protobuf generation
* mocks
* clean

Adjust only:

* package names
* binary names
* integration-specific paths

Do not change repository-wide generation conventions.

---

# README

Generate:

```
integrations/README.md
```

Use the deposits README as the template.

Document:

## Purpose

Describe the Integration Service.

## Responsibilities

Explain:

* OAuth management
* webhook processing
* token lifecycle
* integration storage
* provider abstraction

## Folder Layout

Explain each directory.

## Local Development

Document:

* prerequisites
* environment variables
* Makefile commands
* running locally

## Database

Document:

* migrations
* sqlc generation

## Protobuf

Document:

* generation process

## Docker

Explain how to build and run the container.

## Future Integrations

Reserve sections for:

* HubSpot
* Salesforce
* Stripe
* Microsoft
* Google

---

# Validation

Before completion verify:

* Dockerfile builds using the same strategy as deposits.
* Makefile commands reference valid paths.
* README commands are correct.
* Environment variables match the config package.
* No secrets are committed.

---

# Output

Provide:

## Files Created

List every generated file.

## Files Updated

List any modified file.

## Differences From Deposits

Explain every intentional deviation.

No other repository files should be modified by this agent.
