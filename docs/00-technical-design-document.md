RVPay
Technical Design Document (TDD)
Document Version: 0.1 (Draft)
 Status: Initial Design
 System: RVPay
 Architecture: Golang Microservices Backend
 Database: PostgreSQL
 Communication: gRPC + Protocol Buffers + gRPC-Gateway (HTTP)

1. Overview
1.1 Purpose
RVPay is a backend payment platform designed to facilitate digital payment collection and settlement between businesses and their customers.
The platform is intended to be published on third-party marketplaces such as the GoHighLevel Marketplace while remaining flexible enough to support additional platforms in the future.
RVPay serves as an intermediary payment layer between:
Platform Administrators
Registered Business Clients (RVClients)
End Customers
Customer payments are collected into an administrator-controlled merchant wallet before being settled to the appropriate RVClient after deducting applicable fees.

2. Goals
The primary goals of RVPay are:
Provide a centralized payment platform.
Support multiple marketplace platforms.
Support multiple payment providers.
Expose both gRPC and HTTP APIs.
Support future payment providers without architectural changes.
Separate business capabilities into independent microservices.
Maintain strongly typed database access using sqlc.
Maintain strongly typed RPC contracts using Protocol Buffers.

3. High-Level Architecture
                +----------------------+
                |      GoHighLevel     |
                |     Marketplace      |
                +----------+-----------+
                           |
                           |
                    Install RVPay
                           |
                           v
                  +------------------+
                  |      RVPay       |
                  | Microservices    |
                  +------------------+
                           |
          +----------------+----------------+
          |                                 |
          |                                 |
   Clients Service                 Transactions Service
          |                                 |
          |                                 |
          +---------------+-----------------+
                          |
                     PostgreSQL
Future marketplace platforms can integrate with RVPay without changes to the transaction system.

4. Technology Stack
Component
Technology
Language
Go
Database
PostgreSQL
RPC
gRPC
HTTP Gateway
gRPC-Gateway
Serialization
Protocol Buffers
SQL Generation
sqlc
Protobuf Generation
protoc


5. Core Concepts
RVPay distinguishes between three principal actors.
5.1 Admin
Represents the platform owner.
Responsibilities include:
managing platforms
managing RVClients
monitoring transactions
controlling merchant wallets
Admins are stored within the Clients table and identified by their assigned role.

5.2 RVClient
Represents businesses using RVPay.
Examples include businesses integrating RVPay into their CRM or SaaS platform.
Responsibilities include:
integrating RVPay
receiving customer payments
requesting payouts
monitoring transactions
RVClients are registered users.

5.3 Customer
Customers are end users making payments.
Customers are not registered users of RVPay.
Their interaction is limited to payment initiation.

6. Service Architecture
RVPay consists of two primary microservices.
6.1 Clients Service
Responsible for identity and marketplace integrations.
Responsibilities include:
Client registration
Platform management
Marketplace integrations
Each service owns its own:
database layer
repository layer
business service
gRPC interface

6.2 Transactions Service
Responsible for payment operations.
Responsibilities include:
Deposits
Payouts
Customer payment records
Merchant management
Each service owns its own:
database layer
repository layer
business service
gRPC interface

7. Database Design

7.1 Clients Service Database
Clients
Represents both administrators and RVClients.
Primary fields include:
id
name
user_role
additional client metadata

Platforms
Represents external platforms capable of hosting RVPay.
Examples include:
GoHighLevel Marketplace
Future platforms may also be supported.
Primary fields:
id
name

Integrations
Represents installations of RVPay.
Each Integration links:
one Platform
one Client
Primary fields:
id
platform_id
client_id

User Role Enumeration
USER_ROLE_USER
USER_ROLE_ADMIN
The role determines authorization within RVPay.

8. Transactions Service Database

Merchants
Represents payment gateways.
Examples include:
PawaPay
Flutterwave
Fields include:
id
name
supported providers
Each Merchant may support multiple payment providers.

Deposits
Represents inbound customer payments.
Fields include:
id
client_id
customer_id
merchant_id
currency
amount
payment type
provider
Deposits move funds into the administrator-controlled merchant wallet.

Payouts
Represents outbound payments.
Fields include:
id
client_id
merchant_id
currency
amount
payment type
provider
Payouts transfer funds from the administrator-controlled wallet to RVClients.

Customers
Represents end users making payments.
Fields include:
id
client_id
merchant_id
name
phone number
Customers are associated with a Client and Merchant but are not authenticated users of RVPay.

Provider Enumeration
PROVIDER_MTN_MOMO
PROVIDER_ORANGE_MOMO

Payment Type Enumeration
TYPE_MMO
TYPE_CREDIT_CARD

9. Entity Relationships
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

10. Payment Flow
Deposit Flow
Customer initiates payment.
Deposit request includes Client ID.
Deposit is processed through a Merchant.
Funds are deposited into the administrator-controlled wallet.
Deposit is recorded.

Payout Flow
Administrator or Client requests payout.
Funds are withdrawn from the merchant wallet.
Applicable platform fee is deducted.
Remaining funds are transferred to the Client.
Payout is recorded.

11. Authorization Model
Administrator
Stored as:
USER_ROLE_ADMIN
Permissions include:
Platform management
Client management
Transaction monitoring
Administrative APIs

Client
Stored as:
USER_ROLE_CLIENT
Permissions include:
View own transactions
Manage integrations
Request payouts

Customer
Customers are unauthenticated users.
Their access is limited to deposit endpoints.

12. API Structure
Authentication
Responsibilities include:
Client registration
Platform integration

Administration
Capabilities include:
Platform management
Client management
Transaction monitoring

Client
Capabilities include:
Registration
Integration management
Transaction monitoring

Deposits
Accessible by customers.
Deposit requests require a Client ID.

Payouts
Accessible by:
Administrators
Clients

13. Extensibility
The design intentionally separates Platforms from Integrations.
This allows RVPay to support additional marketplaces without modifying transaction processing.
Similarly, Merchants are independent of payment providers.
A Merchant may support multiple payment providers through the Provider enumeration.
This separation enables future support for additional payment ecosystems while preserving the existing data model.

14. Future Considerations
The current design establishes the foundational architecture for RVPay. Areas identified for future expansion include:
Additional marketplace platform support.
Additional payment providers.
Expanded payment methods.
Enhanced authorization and authentication.
Fee management and settlement logic.
Reporting and analytics.
Audit logging.
Operational monitoring and observability.
Administrative dashboards.

Appendix A – Defined Enumerations
User Roles
Name
USER_ROLE_USER
USER_ROLE_ADMIN

Providers
Name
PROVIDER_MTN_MOMO
PROVIDER_ORANGE_MOMO

Payment Types
Name
TYPE_MMO
TYPE_CREDIT_CARD
