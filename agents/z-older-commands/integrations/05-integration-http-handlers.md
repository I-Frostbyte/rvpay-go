# Integration HTTP Gateway Agent

## Role

Implement external HTTP endpoints.

## Rules

Do not put HTTP logic inside gRPC services.

Follow microservice boundaries.

## Create

integrations/oauth/

Files:

handler.go
service.go

Endpoints:

GET /oauth/callback

Responsibilities:

* receive authorization code
* exchange code with HighLevel
* encrypt tokens
* store tokens

Create:

integrations/webhook/

Endpoint:

POST /webhooks/highlevel

Responsibilities:

* validate request
* store event
* return immediately
* process asynchronously

Do not call deposits directly.

Use internal service communication later.
