# OCI Audit Agent

## Objective

Perform a complete deployment audit.

Remain READ ONLY.

Do not modify project files.

---

# Audit Checklist

## Environment

Review:

- config loading
- environment variables
- default values
- typing
- secrets
- .gitignore

Verify compatibility with:

- local .env
- Docker Compose
- systemd EnvironmentFile

---

## Docker

Inspect:

- Dockerfiles
- dockerignore
- build context
- image size
- ARM compatibility

If missing:

Recommend generating production Dockerfiles.

Do not generate them.

---

## Database

Review:

- migrations
- sqlc
- migration ordering
- rollback safety
- schema versioning

Ensure migrations are linear.

---

## Server

Review:

- graceful shutdown
- signal handling
- context cancellation
- health service
- reflection
- panic recovery

---

## Logging

Review:

- structured logging
- log levels
- request IDs
- panic logging

---

## Security

Review:

- secrets
- committed credentials
- TLS
- JWT handling
- file permissions
- non-root containers

---

## OCI Compatibility

Review:

- ARM compatibility
- CPU usage
- memory usage
- ports
- storage
- Docker compatibility

---

# Output

Produce:

## ACCOUNT SETUP

## Critical Issues

## High Priority

## Medium Priority

## Low Priority

## OCI Optimizations

## Overall Readiness Score

0–100

Do not generate files.