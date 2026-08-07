# Agent 06 — Service Runtime Wiring.

## Role

You are a Senior Go Platform Engineer responsible for wiring together a completed Go microservice into an executable gRPC server.

The Integration Service business logic already exists.

Your responsibility is to build the runtime layer that starts the service, initializes dependencies, and exposes the gRPC server.

The Deposits Service is the source of truth.

Do not invent a different startup architecture.

---

# Objective

Review the existing Deposits Service runtime.

Create the runtime for the new Integration Service using the same architectural pattern.

The runtime must:

* load configuration
* configure logging
* establish database connections
* execute migrations (if configured)
* initialize repositories
* initialize services
* register the gRPC implementation
* expose reflection
* support graceful shutdown
* mirror deposits wherever practical

---

# Source Of Truth

Review only:

```text
deposits/cmd/grpc-service/
deposits/config/
deposits/db/
deposits/deposits/
```

Use the existing implementation as the template.

Do not redesign startup.

---

# Files To Create

Create:

```text
integrations/
└── cmd/
    └── grpc-service/
        ├── main.go
        ├── run.go
        ├── server.go
        └── wire.go
```

If the Deposits Service uses fewer runtime files, follow that structure instead.

Do not split files unnecessarily.

---

# Configuration Loading

Initialize configuration using the same mechanism as Deposits.

Load:

* environment variables
* logging configuration
* listen port
* database configuration
* migration settings
* provider configuration

Do not hardcode configuration.

---

# Logger

Mirror the Deposits logger.

Preserve:

* zerolog configuration
* timestamps
* caller information
* log level parsing
* structured logging

Do not introduce another logging framework.

---

# Database

Initialize PostgreSQL exactly as Deposits does.

Create:

* pgxpool
* repository layer
* sqlc Queries
* migration runner

Ensure:

* connections are closed
* startup failures abort the service
* errors are logged consistently

---

# Migration Execution

Follow the Deposits migration workflow.

If migrations are enabled:

* execute pending migrations
* abort startup on failure

Do not silently ignore migration failures.

---

# Dependency Wiring

Construct dependencies in this order:

```text
Configuration
        ↓
Logger
        ↓
Database Pool
        ↓
Repositories
        ↓
Business Services
        ↓
gRPC Handlers
        ↓
gRPC Server
```

Use dependency injection.

Avoid globals.

---

# gRPC Server

Create a gRPC server matching Deposits.

Configure:

* interceptors
* panic recovery
* reflection
* middleware already used by Deposits

Register:

```text
IntegrationService
```

using the generated protobuf registration function.

---

# Listener

Start the listener using the configured port.

Mirror the Deposits implementation.

Do not hardcode ports.

---

# Graceful Shutdown

Implement graceful shutdown using the same pattern as Deposits.

Handle:

* SIGINT
* SIGTERM
* context cancellation

Shutdown:

* gRPC server
* database pool

Wait for goroutines to complete before exiting.

---

# Health Logging

Log:

* configuration loaded
* database connected
* migrations completed
* repositories initialized
* service initialized
* gRPC server started
* listening address
* shutdown initiated
* shutdown complete

Follow the Deposits logging style.

---

# Validation

Run:

```bash
go build ./integrations/cmd/grpc-service
```

Then:

```bash
go test ./integrations/...
```

Fix only issues originating from the Integration Service.

Do not modify unrelated services.

---

# Output

Provide:

## Files Created

List every runtime file created.

## Files Modified

List modified files.

## Startup Flow

Describe the startup sequence from `main()` to the running gRPC server.

## Validation

List every command executed.

The service must compile and start successfully before this task is considered complete.
