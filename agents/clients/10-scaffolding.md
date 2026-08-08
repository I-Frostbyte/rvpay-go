# Agent 10 — Service Scaffolding & Developer Experience

## Objective

Complete the Clients Service by creating and validating every supporting artifact required for development, deployment, and long-term maintenance.

The Clients service should mirror the structure and developer experience of the Deposits service.

This agent focuses on project scaffolding only.

It must not introduce new business logic.

It must not redesign repositories.

It must not redesign protobuf contracts.

It must not redesign runtime wiring.

The outcome should be that a developer unfamiliar with RVPay can clone the repository and immediately understand how to build, run, generate, test, and deploy the Clients service.

---

# Required Reading

Read only:

- README.md
- agents/project-context.md
- docs/domain-model.md
- docs/repository-layout.md
- docs/protobuf-strategy.md
- docs/migration-plan.md

Then inspect only:

clients/

deposits/

Review only enough files to understand:

- repository layout

- Makefiles

- Dockerfiles

- README structure

- configuration files

- environment variables

- generation workflow

- deployment workflow

Do not recursively inspect the repository.

---

# Repository Review Scope

Inspect only:

clients/

deposits/

.github/workflows/

protobuf/

Do not inspect:

tests/

third_party/

vendor/

coverage/

tmp/

bin/

---

# Responsibilities

Complete the developer-facing scaffolding for the Clients Service.

Ensure consistency with Deposits wherever possible.

This includes:

README

Dockerfile

Makefile

.env.example

development scripts

generation commands

deployment compatibility

documentation

No business functionality should be introduced.

---

# Service README

Create:

clients/README.md

Document:

purpose

architecture

directory structure

configuration

build instructions

generation workflow

database migrations

protobuf generation

sqlc generation

running locally

running with Docker

deployment notes

Render compatibility

The style should closely mirror the Deposits documentation.

---

# Dockerfile

Create a Dockerfile matching the conventions used by Deposits.

Verify:

multi-stage build

minimal runtime image

Go version consistency

protobuf generation compatibility

health check compatibility

environment variable support

Do not introduce a different container strategy.

---

# Makefile

Create a Makefile mirroring Deposits.

Support commands such as:

build

run

generate

generate-protos

generate-sql

test

lint (if already used)

docker-build

clean

The Makefile should integrate naturally with repository-wide tooling.

---

# Environment Template

Create:

clients/.env.example

Include every required configuration variable.

Provide safe placeholder values.

Do not include:

real secrets

OAuth credentials

API tokens

database passwords

The template should document the purpose of every variable.

---

# Documentation

Verify documentation exists for:

service overview

configuration

repository layout

generation workflow

protobuf generation

sqlc generation

Docker usage

Render deployment

common troubleshooting

Keep documentation concise.

Avoid duplicating repository-level documentation.

---

# Generation Workflow

Verify that:

protobuf generation

sqlc generation

mock generation

go generate

continue working without modification.

Do not duplicate generation logic.

Reuse repository tooling.

---

# Deployment Compatibility

Verify compatibility with:

Render

Docker

future container orchestration

Repository CI

Do not implement deployment infrastructure.

Only ensure the Clients service integrates cleanly with existing deployment tooling.

---

# Repository Consistency

Ensure the Clients service matches repository conventions.

Examples include:

directory names

package names

logging

Makefile style

Docker conventions

README formatting

environment naming

Avoid introducing new conventions.

---

# Developer Experience

Review the Clients service as though a new engineer has just joined the project.

Confirm they can determine:

how to build

how to run

how to generate code

how to configure environment variables

how to execute migrations

how to deploy

without reading unrelated services.

---

# Deliverables

Create or update only:

clients/

README.md

Dockerfile

Makefile

.env.example

supporting documentation

Do not modify:

repositories

business services

protobufs

runtime

OAuth

webhooks

---

# Validation

Before completing verify:

- README is complete

- Dockerfile builds successfully

- Makefile commands execute successfully

- environment template is complete

- generation commands work

- service builds

- Docker image builds

- documentation matches implementation

- project compiles successfully

---

# Success Criteria

The Clients service should now appear as a complete first-class microservice within the RVPay repository.

A developer should be able to build, run, configure, generate, and deploy the service using only the service documentation.

---

# Completion Rules

Before completing verify:

- Existing repository conventions have been preserved.

- Existing Makefile conventions have been preserved.

- Existing Docker conventions have been preserved.

- Existing README conventions have been preserved.

- Existing CI compatibility has been preserved.

- Existing generation workflow has been preserved.

- No unrelated directories have been modified.

- Project builds successfully.

If a prerequisite from a previous agent is missing, stop and explain why instead of creating incomplete scaffolding.

---

# Developer Experience Review

Before completing, perform a complete developer experience review.

Evaluate the Clients service from the perspective of a new contributor.

Confirm:

- the service is easy to discover

- documentation is sufficient

- setup requires minimal effort

- configuration is obvious

- build commands are intuitive

- generation commands are documented

- deployment steps are discoverable

- Docker workflow matches the rest of the repository

- Makefile targets remain consistent

- repository conventions are preserved

- no documentation contradicts implementation

If improvements are discovered that do not affect previous implementation agents, implement them.

If improvements require redesigning runtime, repositories, protobufs, or business services, stop and document the issue instead.

Produce:

clients/docs/developer-experience-review.md

The report should summarize:

- onboarding experience

- documentation quality

- build workflow

- generation workflow

- deployment readiness

- consistency with Deposits

- remaining work before testing

Only after this review is complete should the project proceed to Agent 11.