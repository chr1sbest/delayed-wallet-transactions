# Reconciliation Lambda

This AWS Lambda function serves as a self-healing mechanism for the transaction processing system. Its primary purpose is to find and re-process transactions that may have become "stuck" in a `RESERVED` state due to transient failures or other unexpected issues in the settlement flow.

## Trigger

- **Source**: Amazon EventBridge (or CloudWatch Events) Schedule
- **Event**: The function is invoked on a fixed schedule (every 30 minutes).

## Core Logic

1.  **Scan for Stuck Transactions**: The lambda queries the `Transactions` DynamoDB table to find all transactions that have been in the `RESERVED` state for longer than a predefined `stuckTransactionThreshold` (30 minutes).

2.  **Re-enqueue**: For each stuck transaction found, the lambda re-enqueues it into the main settlement SQS queue.

3.  **Graceful Continuation**: The process is designed to be robust. If re-enqueuing a specific transaction fails, the error is logged, and the function continues to the next stuck transaction without halting the entire batch.

## Goal

The goal of this lambda is to ensure the durability and eventual consistency of the financial system. By periodically reconciling the state of transactions, it guarantees that no transaction is permanently lost or left in an intermediate state, even if the settlement lambda experiences temporary downtime or errors.

## Configuration

The lambda requires the following environment variables to be set:

- `SQS_QUEUE_URL`: The URL of the settlement SQS queue where stuck transactions will be re-enqueued.
- `DYNAMODB_TRANSACTIONS_TABLE_NAME`: The name of the DynamoDB table for transactions.
