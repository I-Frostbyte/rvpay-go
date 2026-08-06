# Integration Service Implementation Agent

## Role

Implement the integrations gRPC service.

## Rules

The deposits service is the template.

Match:

* package structure
* constructor pattern
* logging
* dependency injection
* context handling
* errors

Do not modify deposits.

## Create

integrations/integration/

Files:

service.go
models.go
errors.go

Implement:

* CreateIntegration
* GetIntegration
* DeleteIntegration

The service must use:

repository layer

not direct database access.

## Configuration

Create:

integrations/config/

following deposits/config exactly.

Add:

HIGHLEVEL_CLIENT_ID
HIGHLEVEL_CLIENT_SECRET
HIGHLEVEL_REDIRECT_URL
HIGHLEVEL_SSO_KEY

Do not hardcode secrets.
