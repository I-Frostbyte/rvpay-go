# OCI Infrastructure Generator

## Objective

Generate deployment infrastructure for Oracle Cloud Always Free.

Only generate files after inspecting the repository.

If equivalent files already exist:

- review them
- explain improvements
- generate replacements only if requested

---

# Generate

If missing:

- production Dockerfile
- docker-compose.yml
- .dockerignore
- deploy-oci-bootstrap.sh
- systemd service files
- nginx or traefik configuration
- GitHub Actions deployment workflow

---

# Docker Rules

Detect Go version from go.mod.

Use:

golang:<detected-version>-alpine

Build:

GOOS=linux

GOARCH=arm64

Use multi-stage builds.

Prefer:

gcr.io/distroless/static-debian12

Fallback:

alpine

Run containers as a non-root user.

Minimize image size.

---

# Docker Compose

Generate:

- application
- postgres
- migration service

Do not start migrations from every replica.

Prefer:

Dedicated migration service.

---

# Bootstrap Script

Generate:

deploy-oci-bootstrap.sh

It must:

- update packages
- install Docker
- install Docker Compose plugin
- install Git
- install curl
- install jq
- install unzip
- configure firewall
- enable Docker
- configure automatic startup

---

# Reverse Proxy

Generate nginx or Traefik configuration.

Support:

- HTTPS
- gRPC
- HTTP

---

# CI/CD

Generate GitHub Actions workflow.

Pipeline:

Build

↓

Test

↓

Build Docker image

↓

Push

↓

Deploy over SSH

↓

Restart services

---

Every generated artifact must explain:

- why it exists
- why it is configured that way
- how it relates to OCI Always Free