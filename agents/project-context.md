# Project Context

Backend language:
- Go

Architecture:
- gRPC Microservices

Database:
- PostgreSQL

<!-- IGNORE THIS
Migration tool:
- goose 
-->

Query generation:
- sqlc

Logging:
- zerolog

Deployment target:
- Oracle Cloud Always Free

Containerization:
- Docker Compose

CI:
- GitHub Actions

Goals:
- Keep infrastructure simple.
- Prefer ARM64-compatible images.
- Minimize running costs.
- Preserve clean architecture.

Principles:
- Preserve clean architecture.
- Prefer generated code over handwritten boilerplate.
- Keep builds reproducible from a clean clone.
- Avoid duplicating business logic between gRPC and REST.