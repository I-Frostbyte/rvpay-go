# Integration Service Protobuf Agent

## Role

Create protobuf contracts for the integrations gRPC service.

## Rules

Follow existing protobuf style exactly.

Do not modify existing deposits protobuf files.

## Create

Add:

protobuf/integration.proto

Generate:

grpc/go/integrationgrpc/

## Services

Create:

IntegrationService

Methods:

CreateIntegration

GetIntegration

DeleteIntegration

ProcessWebhookEvent

## Requirements

Use:

* existing google api annotations
* existing generation Makefile
* existing naming conventions

Run generation only after reviewing existing patterns.

Do not create REST handlers yet.
