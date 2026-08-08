# Deployment Validation

Remain READ ONLY.

---

Verify

✓ Docker builds

✓ Generated protobufs

✓ sqlc

✓ Tests

    Run tests only after completing the static review.

    Abort testing after 10 minutes if no output is produced.

    Report the reason for the timeout.

✓ gRPC

✓ REST

✓ Health

✓ Logging

✓ Environment

✓ GitHub Actions

✓ Render compatibility

✓ PostgreSQL

✓ Graceful shutdown

✓ Docker image size

✓ Security

---

Produce

Deployment Readiness Score

0–100

List remaining blockers.

Categorize

Critical

High

Medium

Low

Deployment may proceed only if there are no Critical blockers.

Never recursively inspect:

.git
node_modules
vendor
googleapis
grpc/go
generated
dist
coverage
bin

unless explicitly requested.

Never inspect more than 25 files before producing an initial report.

If additional review is required,

ask for permission first.