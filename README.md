# Delayed Wallet Transactions

## Goal
The system allows users to schedule a transfer of funds to another user at a future time. The funds are reserved at the time of scheduling to prevent double-spending and are processed asynchronously.

### Cruxes
1. **Consistency** - Preventing Double Spend
2. **Scale** - 1M Scheduled Payments at Once
3. **Delay** - Asynchronous Processing of Payments
4. **Delivery** - Asynchronous Delivery of Events to the Client

![System Design](DelayedWallet.png)

## Architecture
This service is built using a modern, event-driven architecture on AWS. The main components are:

- **Scheduler Service (`cmd/app`):** A Go-based HTTP server that exposes the primary API for creating and viewing transactions and wallets. It is responsible for initial request validation and authentication.

- **DynamoDB Tables:** A set of three purpose-built tables form the core of our data layer:
  - **`Wallets`**: Stores the current state of each user's wallet, including their available `balance`, `reserved` funds, and a `version` number for optimistic locking.
  - **`Transactions`**: Acts as a state machine for each financial movement, tracking its status from `RESERVED` to `COMPLETED`.
  - **`LedgerEntries`**: An append-only, immutable ledger that provides a permanent, double-entry audit trail of all fund movements.

- **Asynchronous Processing Flow:** To ensure the API is responsive and resilient, transaction processing is handled asynchronously:
  1. The API service reserves funds and publishes a transaction message to an **SQS Queue**.
  2. A **Settlement Lambda (`cmd/settlement_lambda`)** consumes this message, performs the final settlement, and creates the ledger entries. This flow uses SQS's `DelaySeconds` feature for transactions scheduled in the future (up to 15 minutes).

- **Reconciliation Lambda (`cmd/reconciliation_lambda`):** A scheduled Lambda that runs periodically (e.g., every 20 minutes) to find and re-enqueue transactions that may have become "stuck" in a `RESERVED` state due to transient failures. This makes the system self-healing.

## Consistency & Idempotency

Ensuring financial correctness in a distributed system is the primary challenge. We address this with the following strategies:

- **Atomic Operations with `TransactWriteItems`:** All critical state changes are performed inside a single, atomic `TransactWriteItems` call. This guarantees that an operation (like reserving funds or settling a transaction) either completely succeeds or completely fails, leaving the system in a consistent state. There are no partial updates.

- **Race Condition Prevention via Optimistic Locking:** To prevent double-spends, all wallet updates are protected by a `version` number. A transaction will only succeed if the wallet's `version` has not changed since it was read, preventing two concurrent operations from corrupting the balance.

- **Idempotent Settlement:** The final settlement operation is designed to be idempotent by including a condition check that the transaction's status must be `APPROVED`. This means that even if the same settlement message is processed multiple times (a guarantee in distributed systems), the funds will only be moved once. Subsequent attempts will fail safely, preventing double-payments.

## Getting Started

### Prerequisites

- Go (1.24+)
- Docker & Docker Compose (optional, for future containerization)
- An AWS account with credentials configured in your environment.

### Setup

1.  **Install dependencies:**
    ```sh
    go mod tidy
    ```

2.  **Configure your environment:**
    - Copy the `.env.example` file to `.env`:
      ```sh
      cp .env.example .env
      ```
    - Edit the `.env` file with your specific AWS resource names:
      - `DYNAMODB_TRANSACTIONS_TABLE_NAME`
      - `DYNAMODB_WALLETS_TABLE_NAME`
      - `DYNAMODB_LEDGER_TABLE_NAME`
      - `SQS_QUEUE_URL`

3.  **Deploy AWS Resources with SAM:**

    This project uses the AWS Serverless Application Model (SAM) to define and deploy the required AWS resources. The `template.yaml` file in the root of the project contains the definitions for the DynamoDB tables and the SQS queue.

    To deploy the resources, run the following commands from the project root:

    First, build the SAM application:
    ```sh
    sam build
    ```

    Then, deploy the resources using the guided deployment process. This will prompt you for configuration details, such as the AWS region and a stack name.
    ```sh
    sam deploy --guided
    ```

    After deployment, the SAM CLI will output the names and ARNs of the created resources. You will need to use these outputs to update your `.env` file.

4.  **Run the application:**
    ```sh
    go run ./cmd/app/main.go
    ```
    The server will start on port 8080 by default.

## Scheduler API Documentation

The Scheduler API is defined using the OpenAPI 3.0 standard in `api/spec.yaml`.

### Endpoints

- `POST /transactions`
  - **Description**: Schedules a new delayed transaction. Atomically reserves funds from the sender's wallet and creates a transaction record.
  - **Request Body**: `NewTransaction` object (`from_user_id`, `to_user_id`, `amount`, `scheduled_at`).
  - **Response**: The created `Transaction` object.

- `GET /transactions/{transactionId}`
  - **Description**: Retrieves the details and current status of a specific transaction.
  - **Response**: The `Transaction` object.

- `GET /wallets`
  - **Description**: Retrieves a list of all wallets.
  - **Response**: An array of `Wallet` objects.

- `POST /wallets`
  - **Description**: Creates a new wallet for a user with a default starting balance.
  - **Request Body**: `NewWallet` object (`user_id`).
  - **Response**: The created `Wallet` object.

- `GET /wallets/{userId}`
  - **Description**: Retrieves the current state of a user's wallet, including their available balance, reserved funds, and version number.
  - **Response**: The `Wallet` object.

- `DELETE /wallets/{userId}`
  - **Description**: Deletes a user's wallet. This is a destructive operation.
  - **Response**: `204 No Content` on success.
