# Agent 02 — Repository Layout

## Objective

Design the repository structure required to evolve the current repository into the new RVPay architecture.

This task defines structure only.

Do not implement services.

Do not move files.

Do not delete files.

---

## Review Scope

Read only:

- README.md
- docs/00-technical-design-document.md

Review these directories only if necessary to verify naming consistency:

deposits/
integrations/
protobuf/

Ignore:

grpc/go/
third_party/
.github/
deploy/
nginx/
tests/

---

## Existing Layout

Use the repository README as the source of truth.

Preserve existing conventions wherever possible.

The Deposits service represents the preferred implementation style.

---

## Design Goals

Introduce:

Clients Service

Transactions Service

Shared repository assets

without breaking existing repository organization.

---

## Produce

Design the future repository layout.

Identify:

new directories

renamed directories

shared directories

generated directories

documentation locations

---

## Shared Code

Identify code suitable for extraction into shared packages.

Examples:

configuration

logger

database helpers

middleware

Do not move business logic into shared packages.

---

## Deliverable

Create:

docs/repository-layout.md

Include:

- future directory tree
- purpose of every major directory
- migration notes
- directories scheduled for deprecation

Do not modify existing files.