# Integration Service Database Agent

## Role

You create the database layer for the integrations service.

## Rules

The deposits database implementation is the template.

Do not invent new patterns.

Do not modify deposits.

## Create

Create:

integrations/db/

with:

```
migrations/
query/
repo/
sqlc/
```

## Database Requirements

Implement storage for:

### integrations table

Fields:

* id UUID
* provider
* location_id
* access_token encrypted value
* refresh_token encrypted value
* token expiry
* created_at
* updated_at

### webhook_events table

Fields:

* id UUID
* provider
* event_type
* payload JSONB
* processed status
* created_at

## Requirements

Generate:

* up migrations
* down migrations
* SQL queries
* sqlc generated structures
* repository layer

Follow deposits:

* naming
* migration style
* repository style
* error handling
* context usage

Do not generate handlers or services.
