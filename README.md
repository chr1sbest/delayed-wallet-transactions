# Delayed Wallet Transaction Service

This project is a Go-based API service for managing delayed wallet-to-wallet transactions. It is designed to be highly available, scalable, and strongly consistent, making it suitable for financial applications.

The system allows users to schedule a transfer of funds to another user at a future time. The funds are reserved at the time of scheduling to prevent double-spending and are processed asynchronously.

## System Design & Architecture

This service is built using a modern, event-driven architecture on AWS, prioritizing correctness and resilience.

- **Strongly Consistent Data Layer**: The core of the system is a three-table DynamoDB model (`Wallets`, `Transactions`, `LedgerEntries`). Fund reservations are handled atomically using `TransactWriteItems` operations with optimistic locking (via a `version` attribute) to prevent race conditions and ensure data integrity.
- **Asynchronous Processing**: Upon successful creation, transactions are published to an SQS queue. This decouples the API from the processing logic, ensuring the API remains fast and responsive.
- **State Machine**: The `Transaction` object acts as a state machine, transitioning through statuses like `RESERVED`, `APPROVED`, and `COMPLETED`.
- **Idempotency**: All critical operations are designed to be idempotent. For example, creating a transaction uses a conditional write to prevent duplicate records.
- **Double-Entry Ledger**: The design includes an append-only `LedgerEntries` table to provide a complete and immutable audit trail of all financial movements (to be implemented).

## Getting Started

### Prerequisites

- Go (1.18+)
- Docker & Docker Compose (optional, for future containerization)
- An AWS account with credentials configured in your environment.

### Setup

1.  **Clone the repository:**
    ```sh
    git clone <repository-url>
    ```

2.  **Install dependencies:**
    ```sh
    go mod tidy
    ```

3.  **Configure your environment:**
    - Copy the `.env.example` file to `.env`:
      ```sh
      cp .env.example .env
      ```
    - Edit the `.env` file with your specific AWS resource names:
      - `DYNAMODB_TRANSACTIONS_TABLE_NAME`
      - `DYNAMODB_WALLETS_TABLE_NAME`
      - `DYNAMODB_LEDGER_TABLE_NAME`
      - `SQS_QUEUE_URL`

4.  **Set up AWS Resources:**
    - Create the three DynamoDB tables with the specified primary keys:
      - **Transactions Table**: Primary Key `id` (String)
      - **Wallets Table**: Primary Key `user_id` (String)
      - **Ledger Table**: Primary Key `TransactionID` (String), Sort Key `EntryID` (String)
    - Create the SQS queue.

5.  **Run the application:**
    ```sh
    go run ./cmd/app/main.go
    ```
    The server will start on port 8080 by default.

## API Documentation

The API is defined using the OpenAPI 3.0 standard in `api/spec.yaml`.

### Endpoints

- `POST /transactions`
  - **Description**: Schedules a new delayed transaction. Atomically reserves funds from the sender's wallet and creates a transaction record.
  - **Request Body**: `NewTransaction` object (`from_user_id`, `to_user_id`, `amount`, `currency`, `scheduled_at`).
  - **Response**: The created `Transaction` object.

- `GET /transactions/{transactionId}`
  - **Description**: Retrieves the details and current status of a specific transaction.
  - **Response**: The `Transaction` object.

- `GET /wallets/{userId}`
  - **Description**: Retrieves the current state of a user's wallet, including their available balance, reserved funds, and version number.
  - **Response**: The `Wallet` object.
