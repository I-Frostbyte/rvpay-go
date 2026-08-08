# RVPay Domain Model

Document Version: 1.0
Status: Foundation
System: RVPay
Architecture: Go Microservices
Database: PostgreSQL
Communication: gRPC + Protocol Buffers + gRPC-Gateway (HTTP)

## 1. Overview

This document defines the complete RVPay business domain. It is the authoritative
reference for entity ownership, responsibilities, relationships, and lifecycle
across every service in the repository.

The domain is derived from `docs/00-technical-design-document.md` and the current
repository state described in `README.md`. It reconciles the existing Deposits and
Integrations services with the target Clients and Transactions service architecture.

The domain is organised into two bounded contexts:

- **Clients Context** — identity, marketplace platforms, and integrations.
- **Transactions Context** — payment operations: merchants, customers, deposits,
  and payouts.

## 2. Bounded Contexts

### 2.1 Clients Context

Owns everything related to who uses RVPay and how RVPay is installed.

Responsibilities:

- Client registration and administration.
- Platform (marketplace) management.
- Marketplace integrations, including OAuth token storage and webhook ingestion.

### 2.2 Transactions Context

Owns everything related to moving money.

Responsibilities:

- Merchant (payment gateway) management.
- Customer payment records.
- Deposit processing (inbound payments).
- Payout processing (outbound settlements).

## 3. Entities

### 3.1 Clients

**Description**

Clients represent businesses using RVPay (RVClients) and platform administrators
(Admins). Admins are stored within the Clients table and identified by their
assigned role.

**Responsibilities**

- Integrate RVPay into their CRM or SaaS platform.
- Receive customer payments.
- Request payouts.
- Monitor their own transactions.

**Ownership**

- Service: Clients Service.
- Database: Clients Service database.

**Lifecycle**

1. Registered — a client record is created with role `USER_ROLE_USER` or
   `USER_ROLE_ADMIN`.
2. Active — the client can create integrations and transact.
3. Suspended — the client cannot transact; existing records are preserved.
4. Closed — the client is deactivated; records are retained for audit.

**Relationships**

- A Client has many Integrations.
- A Client has many Customers.
- A Client has many Deposits.
- A Client has many Payouts.

### 3.2 Platforms

**Description**

Platforms represent external marketplaces capable of hosting RVPay. The initial
platform is the GoHighLevel Marketplace; future platforms are supported without
changes to transaction processing.

**Responsibilities**

- Host RVPay installations for clients.

**Ownership**

- Service: Clients Service.
- Database: Clients Service database.

**Lifecycle**

1. Registered — an administrator registers a platform.
2. Active — the platform can host integrations.
3. Retired — the platform no longer accepts new integrations; existing
   integrations remain readable.

**Relationships**

- A Platform has many Integrations.

### 3.3 Integrations

**Description**

Integrations represent installations of RVPay. Each Integration links one
Platform and one Client. Integrations own the OAuth connection and webhook
subscription for a client's platform account.

**Responsibilities**

- Link a Client to a Platform.
- Manage OAuth tokens (encrypted at rest) for the provider connection.
- Receive and persist webhook events from the platform.

**Ownership**

- Service: Clients Service.
- Database: Clients Service database.

**Lifecycle**

1. Created — an integration record links a platform and a client.
2. OAuth pending — the client has not completed the OAuth handshake.
3. Active — OAuth is complete and tokens are stored; webhooks are received.
4. Revoked — the client disconnected the integration; tokens are invalidated.

**Relationships**

- Many Integrations belong to one Platform.
- Many Integrations belong to one Client.

### 3.4 Merchants

**Description**

Merchants represent payment gateways. Examples include PawaPay and Flutterwave.
Each Merchant may support multiple payment providers.

**Responsibilities**

- Process deposits and payouts.
- Maintain provider connectivity (e.g. PawaPay API credentials).

**Ownership**

- Service: Transactions Service.
- Database: Transactions Service database.

**Lifecycle**

1. Onboarded — an administrator registers a merchant and its supported providers.
2. Active — the merchant can process payments.
3. Suspended — the merchant cannot process new payments; in-flight payments
   continue.
4. Retired — the merchant is decommissioned; historical records remain.

**Relationships**

- A Merchant processes many Deposits.
- A Merchant processes many Payouts.
- A Merchant serves many Customers.

### 3.5 Customers

**Description**

Customers are end users making payments. They are not registered users of RVPay;
their interaction is limited to payment initiation.

**Responsibilities**

- Initiate payments (deposits).

**Ownership**

- Service: Transactions Service.
- Database: Transactions Service database.

**Lifecycle**

1. Created — a customer record is created on first payment, associated with a
   Client and a Merchant.
2. Active — the customer can continue initiating payments.

**Relationships**

- Many Customers belong to one Client.
- Many Customers are served by one Merchant.
- A Customer has many Deposits.

### 3.6 Deposits

**Description**

Deposits represent inbound customer payments. Funds move into the
administrator-controlled merchant wallet.

**Responsibilities**

- Record an inbound payment from a Customer through a Merchant for a Client.

**Ownership**

- Service: Transactions Service.
- Database: Transactions Service database.

**Lifecycle**

1. Initiated — a deposit request is accepted.
2. Processing — the merchant is processing the payment.
3. Completed — funds are credited to the merchant wallet.
4. Failed — the payment could not be completed.

**Relationships**

- Many Deposits belong to one Client.
- Many Deposits belong to one Customer.
- Many Deposits are processed by one Merchant.

### 3.7 Payouts

**Description**

Payouts represent outbound payments. Funds transfer from the
administrator-controlled merchant wallet to RVClients after applicable fees are
deducted.

**Responsibilities**

- Record an outbound settlement from the merchant wallet to a Client.

**Ownership**

- Service: Transactions Service.
- Database: Transactions Service database.

**Lifecycle**

1. Requested — a payout is requested by an administrator or a client.
2. Processing — the merchant is processing the payout.
3. Completed — funds are transferred to the client.
4. Failed — the payout could not be completed.

**Relationships**

- Many Payouts belong to one Client.
- Many Payouts are processed by one Merchant.

## 4. Service Ownership

No entity is owned by multiple services. Each entity lives in exactly one
service database. Cross-service communication occurs only through gRPC contracts.

| Entity      | Service              | Database                    |
|-------------|----------------------|-----------------------------|
| Clients     | Clients Service      | Clients Service database    |
| Platforms   | Clients Service      | Clients Service database    |
| Integrations| Clients Service      | Clients Service database    |
| Merchants   | Transactions Service | Transactions Service database |
| Customers   | Transactions Service | Transactions Service database |
| Deposits    | Transactions Service | Transactions Service database |
| Payouts     | Transactions Service | Transactions Service database |

## 5. Cross-Service Communication

Services never share database tables. All cross-service data access uses gRPC
contracts.

### 5.1 Transactions Service → Clients Service

- Validate `client_id` on deposit and payout requests.
- Resolve integration and platform context for a transaction.

### 5.2 Clients Service → Transactions Service

- Query transaction status for client and administrator monitoring.

## 6. Entity Relationships

```text
Platform
    |
    | 1
    |
    | *
Integration
    |
    | *
    |
    | 1
Client
    |
    +--------------------------+
    |                          |
    |                          |
Customers                  Deposits
    |                          |
    |                          |
    +-----------+--------------+
                |
             Merchant
                |
                |
             Payouts
```

## 7. Lifecycle Summary

| Entity       | States                                                              |
|--------------|---------------------------------------------------------------------|
| Client       | Registered → Active → Suspended → Closed                            |
| Platform     | Registered → Active → Retired                                       |
| Integration  | Created → OAuth pending → Active → Revoked                          |
| Merchant     | Onboarded → Active → Suspended → Retired                            |
| Customer     | Created → Active                                                    |
| Deposit      | Initiated → Processing → Completed / Failed                         |
| Payout       | Requested → Processing → Completed / Failed                         |

## 8. Future Extensibility

The domain model intentionally separates concepts to support growth without
architectural change.

- **New marketplace platforms** — add a Platform row; transaction processing is
  unaffected because Integrations decouple Platforms from Clients.
- **New payment providers** — extend the provider enumeration on a Merchant;
  no new service is required.
- **New payment methods** — extend the payment type enumeration
  (`TYPE_MMO`, `TYPE_CREDIT_CARD`).
- **Fee management and settlement** — a future fee/settlement entity can attach
  to Payouts without changing the core model.
- **Wallet and balance tracking** — a future wallet entity can model the
  administrator-controlled merchant wallet explicitly.
- **Reporting and analytics** — read models can be derived from existing
  entities without ownership changes.

## 9. Assumptions

- `docs/00-technical-design-document.md` is the authoritative source for the
  target architecture.
- The existing Deposits and Integrations services evolve into the Transactions
  and Clients services respectively; no new service is introduced.
- The hard-coded client (`Socadel`) in the current Deposits service is a
  temporary implementation detail and is replaced by real client resolution
  through the Clients Service.
- Webhook events belong to the integrations domain and therefore to the Clients
  Service.
- OAuth tokens and webhook event records are owned by the Clients Service as
  part of the Integration entity's lifecycle.
- A Customer is associated with exactly one Client and one Merchant at the time
  of first payment.
- Deposits and Payouts reference Clients and Merchants by identifier resolved
  through gRPC; foreign keys exist only within the owning service's database.

## 10. Unresolved Questions

- **Fee entity** — the payout flow deducts applicable fees, but no fee entity is
  defined. A future fee model must be designed before settlement logic is built.
- **Wallet entity** — the administrator-controlled merchant wallet is referenced
  but not modelled. Balance tracking may require a dedicated entity.
- **Customer uniqueness** — whether a Customer may transact across multiple
  Clients or Merchants over time is not defined.
- **Webhook event ownership** — webhook events are assumed to belong to the
  Clients Service; this must be confirmed when the Clients Service is designed.
- **Provider–Merchant granularity** — the exact relationship between a Merchant
  and its supported providers (enumeration vs. table) is not fully specified.
- **Integration ↔ transaction linkage** — whether Deposits/Payouts must reference
  an Integration (to attribute transactions to a platform) is not defined.