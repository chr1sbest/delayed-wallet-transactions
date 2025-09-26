# Settlement Lambda

This AWS Lambda function is responsible for the final settlement of delayed wallet-to-wallet transactions. It acts as the consumer in an asynchronous processing flow, ensuring that the main API remains responsive while financial operations are handled reliably in the background.

## Trigger

- **Source**: Amazon SQS Queue
- **Event**: The function is invoked when a new message containing transaction details is available in the queue.

## Core Logic

1.  **Message Consumption**: The lambda receives a batch of SQS messages, with each message body containing a JSON representation of a `Transaction` object.

2.  **Deserialization**: It parses the JSON from the message body into a `Transaction` struct.

3.  **Settlement**: It calls the `SettleTransaction` method from the storage layer. This is a critical, idempotent operation that:
    -   Atomically updates the `balance` and `reserved` funds for both the sender's and receiver's wallets in the `Wallets` DynamoDB table.
    -   Updates the transaction's status from `RESERVED` to `COMPLETED` in the `Transactions` table.
    -   Creates immutable, double-entry records in the `LedgerEntries` table to provide a permanent audit trail.

4.  **Idempotency**: The settlement logic is designed to be idempotent. It includes condition checks to ensure that a transaction can only be settled once, preventing issues like double-payments if the same SQS message is processed multiple times.

## Error Handling

- If the lambda fails to process a message (e.g., due to a transient database error), it returns an error. This causes the message to become visible again in the SQS queue for a retry attempt after its visibility timeout expires.
- In a production environment, a Dead-Letter Queue (DLQ) would be configured to capture messages that fail repeatedly for manual inspection.

## Configuration

The lambda requires the following environment variables to be set:

- `DYNAMODB_TRANSACTIONS_TABLE_NAME`: The name of the DynamoDB table for transactions.
- `DYNAMODB_WALLETS_TABLE_NAME`: The name of the DynamoDB table for wallets.
- `DYNAMODB_LEDGER_TABLE_NAME`: The name of the DynamoDB table for ledger entries.
