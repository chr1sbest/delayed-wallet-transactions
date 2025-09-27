package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chris/delayed-wallet-transactions/pkg/models"
	"github.com/chris/delayed-wallet-transactions/pkg/storage"
	"github.com/google/uuid"
)

// SettleTransaction performs the final atomic settlement of a transaction.
// It uses a two-step process to ensure idempotency.
// 1. Attempt to acquire a lock by setting the transaction status to WORKING.
// 2. If the lock is acquired, proceed with the settlement.
// This prevents a transaction from being processed multiple times if the lambda is invoked more than once for the same SQS message.
func (s *Store) SettleTransaction(ctx context.Context, tx *models.Transaction) error {
	// Step 1: Attempt to acquire a lock on the transaction by setting its status to WORKING.
	// This is an atomic operation that will only succeed if the current status is RESERVED.
	if err := s.acquireTransactionLock(ctx, tx.Id); err != nil {
		if errors.Is(err, storage.ErrTransactionAlreadyProcessing) {
			// Another process has already acquired the lock, so we can safely exit.
			return nil
		}
		return fmt.Errorf("failed to acquire transaction lock: %w", err)
	}

	// Step 2: Proceed with the settlement logic.
	return s.executeSettlement(ctx, tx)
}

// acquireTransactionLock atomically updates the transaction status from RESERVED to WORKING.
func (s *Store) acquireTransactionLock(ctx context.Context, txID string) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(s.TransactionsTableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: txID},
		},
		UpdateExpression:    aws.String("SET #status = :working_status"),
		ConditionExpression: aws.String("#status = :reserved_status"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":working_status":  &types.AttributeValueMemberS{Value: string(models.WORKING)},
			":reserved_status": &types.AttributeValueMemberS{Value: string(models.RESERVED)},
		},
	}

	_, err := s.Client.UpdateItem(ctx, input)
	if err != nil {
		var condCheckFailed *types.ConditionalCheckFailedException
		if errors.As(err, &condCheckFailed) {
			return storage.ErrTransactionAlreadyProcessing
		}
		return fmt.Errorf("failed to update transaction status to WORKING: %w", err)
	}

	return nil
}

// executeSettlement performs the actual financial settlement of the transaction.
func (s *Store) executeSettlement(ctx context.Context, tx *models.Transaction) error {
	// 1. Get the current state of both wallets for optimistic locking.
	senderWallet, err := s.GetWallet(ctx, tx.FromUserId)
	if err != nil {
		return fmt.Errorf("failed to get sender's wallet for settlement: %w", err)
	}
	receiverWallet, err := s.GetWallet(ctx, tx.ToUserId)
	if err != nil {
		return fmt.Errorf("failed to get receiver's wallet for settlement: %w", err)
	}

	// 2. Prepare common values.
	now := time.Now()
	amountAV, err := attributevalue.Marshal(tx.Amount)
	if err != nil {
		return fmt.Errorf("failed to marshal amount for settlement: %w", err)
	}

	// 3. Prepare ledger entries.
	debitEntry := models.LedgerEntry{
		TransactionID: tx.Id,
		EntryID:       uuid.New().String(),
		AccountID:     tx.FromUserId,
		Debit:         tx.Amount,
		Description:   fmt.Sprintf("Settlement for transaction %s", tx.Id),
		Timestamp:     now,
		GSI1PK:        "LEDGER_ENTRIES",
	}
	creditEntry := models.LedgerEntry{
		TransactionID: tx.Id,
		EntryID:       uuid.New().String(),
		AccountID:     tx.ToUserId,
		Credit:        tx.Amount,
		Description:   fmt.Sprintf("Settlement for transaction %s", tx.Id),
		Timestamp:     now,
		GSI1PK:        "LEDGER_ENTRIES",
	}
	debitAV, err := attributevalue.MarshalMap(debitEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal debit entry: %w", err)
	}
	creditAV, err := attributevalue.MarshalMap(creditEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal credit entry: %w", err)
	}

	// Prepare attribute values for the transaction status update.
	completedStatusAV, err := attributevalue.Marshal(models.COMPLETED)
	if err != nil {
		return fmt.Errorf("failed to marshal completed status: %w", err)
	}
	workingStatusAV, err := attributevalue.Marshal(models.WORKING)
	if err != nil {
		return fmt.Errorf("failed to marshal working status: %w", err)
	}
	nowAV, err := attributevalue.Marshal(now)
	if err != nil {
		return fmt.Errorf("failed to marshal timestamp for status update: %w", err)
	}

	// 4. Construct the TransactWriteItems input.
	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				// Operation 1: Update sender's wallet.
				Update: &types.Update{
					TableName: aws.String(s.WalletsTableName),
					Key: map[string]types.AttributeValue{"user_id": &types.AttributeValueMemberS{Value: tx.FromUserId}},
					UpdateExpression:    aws.String("SET reserved = reserved - :amount, version = version + :inc, #ttl = :ttl"),
					ConditionExpression: aws.String("reserved >= :amount AND version = :version"),
					ExpressionAttributeNames: map[string]string{
						"#ttl": "ttl",
					},
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":amount":   amountAV,
						":version":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", senderWallet.Version)},
						":inc":      &types.AttributeValueMemberN{Value: "1"},
						":ttl":      &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", time.Now().Add(24*time.Hour).Unix())},
					},
				},
			},
			{
				// Operation 2: Update receiver's wallet.
				Update: &types.Update{
					TableName: aws.String(s.WalletsTableName),
					Key: map[string]types.AttributeValue{"user_id": &types.AttributeValueMemberS{Value: tx.ToUserId}},
					UpdateExpression:    aws.String("SET balance = balance + :amount, version = version + :inc, #ttl = :ttl"),
					ConditionExpression: aws.String("version = :version"),
					ExpressionAttributeNames: map[string]string{
						"#ttl": "ttl",
					},
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":amount":   amountAV,
						":version":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", receiverWallet.Version)},
						":inc":      &types.AttributeValueMemberN{Value: "1"},
						":ttl":      &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", time.Now().Add(24*time.Hour).Unix())},
					},
				},
			},
			{
				// Operation 3: Create debit ledger entry.
				Put: &types.Put{
					TableName:           aws.String(s.LedgerTableName),
					Item:                debitAV,
					ConditionExpression: aws.String("attribute_not_exists(entry_id)"),
				},
			},
			{
				// Operation 4: Create credit ledger entry.
				Put: &types.Put{
					TableName:           aws.String(s.LedgerTableName),
					Item:                creditAV,
					ConditionExpression: aws.String("attribute_not_exists(entry_id)"),
				},
			},
			{
				// Operation 5: Update the transaction status to COMPLETED.
				Update: &types.Update{
					TableName: aws.String(s.TransactionsTableName),
					Key:       map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: tx.Id}},
					UpdateExpression:    aws.String("SET #status = :completed_status, updated_at = :now"),
					ConditionExpression: aws.String("#status = :working_status"),
					ExpressionAttributeNames: map[string]string{
						"#status": "status",
					},
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":completed_status": completedStatusAV,
						":working_status":  workingStatusAV,
						":now":              nowAV,
					},
				},
			},
		},
	}

	// 5. Execute the transaction.
	_, err = s.Client.TransactWriteItems(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to execute settlement transaction: %w", err)
	}

	// After success, the transaction status is now COMPLETED.
	return nil
}
