package dynamodb

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chris/delayed-wallet-transactions/pkg/mapping"
	"github.com/chris/delayed-wallet-transactions/pkg/models"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const (
	stuckTransactionGSI = "status-created_at-index"
	fromUserIDIndex     = "from_user_id-index"
)

// CreateTransaction atomically reserves funds from the sender's wallet and creates a new transaction record.
func (s *Store) CreateTransaction(ctx context.Context, tx *models.Transaction) (*models.Wallet, error) {
	// 1. Get the current state of the sender's wallet.
	senderWallet, err := s.GetWallet(ctx, tx.FromUserId)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender's wallet: %w", err)
	}

	// 2. Complete the transaction object with server-side details.
	now := time.Now()
	tx.Id = openapi_types.UUID(uuid.New())
	tx.Status = models.RESERVED
	tx.CreatedAt = now
	tx.UpdatedAt = now
	tx.TTL = time.Now().Add(24 * time.Hour).Unix()

	// Marshal the transaction for the Put operation.
	txAV, err := attributevalue.MarshalMap(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transaction: %w", err)
	}

	// Marshal the amount for the wallet update.
	amountAV, err := attributevalue.Marshal(tx.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal amount: %w", err)
	}

	// 4. Construct the TransactWriteItems input.
	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				// Operation 1: Update the sender's wallet.
				Update: &types.Update{
					TableName: aws.String(s.WalletsTableName),
					Key: map[string]types.AttributeValue{
						"user_id": &types.AttributeValueMemberS{Value: tx.FromUserId},
					},
					UpdateExpression:    aws.String("SET balance = balance - :amount, reserved = reserved + :amount, version = version + :inc, ttl = :ttl"),
					ConditionExpression: aws.String("balance >= :amount AND version = :version"),
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":amount":   amountAV,
						":version":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", senderWallet.Version)},
						":inc":      &types.AttributeValueMemberN{Value: "1"},
						":ttl":      &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", time.Now().Add(24*time.Hour).Unix())},
					},
				},
			},
			{
				// Operation 2: Create the new transaction record.
				Put: &types.Put{
					TableName:           aws.String(s.TransactionsTableName),
					Item:                txAV,
					ConditionExpression: aws.String("attribute_not_exists(id)"),
				},
			},
		},
	}

	// 5. Execute the transaction.
	_, err = s.Client.TransactWriteItems(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to execute transaction: %w", err)
	}

	// 6. If the database transaction was successful, enqueue it for processing.
	if s.Scheduler != nil {
		if err := s.Scheduler.ScheduleTransaction(ctx, mapping.ToApiTransaction(tx)); err != nil {
			log.Printf("CRITICAL: transaction %s created but failed to enqueue: %v", tx.Id, err)
		}
	}

	// 7. Get the updated wallet to return to the user.
	updatedWallet, err := s.GetWallet(ctx, tx.FromUserId)
	if err != nil {
		log.Printf("warning: transaction %s created but failed to retrieve updated wallet: %v", tx.Id, err)
		return nil, nil
	}

	return updatedWallet, nil
}

// GetTransaction retrieves a transaction from DynamoDB by its ID.
func (s *Store) GetTransaction(ctx context.Context, txID openapi_types.UUID) (*models.Transaction, error) {
	key, err := attributevalue.MarshalMap(map[string]string{"id": txID.String()})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transaction ID: %w", err)
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(s.TransactionsTableName),
		Key:       key,
	}

	result, err := s.Client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction from DynamoDB: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("transaction with ID %s not found", txID.String())
	}

	var tx models.Transaction
	if err := attributevalue.UnmarshalMap(result.Item, &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	return &tx, nil
}

func (s *Store) CancelTransaction(ctx context.Context, txID openapi_types.UUID) error {
	tx, err := s.GetTransaction(ctx, txID)
	if err != nil {
		return fmt.Errorf("failed to get transaction for cancellation: %w", err)
	}

	if tx.Status != models.RESERVED {
		return ErrTransactionNotCancellable
	}

	senderWallet, err := s.GetWallet(ctx, tx.FromUserId)
	if err != nil {
		return fmt.Errorf("failed to get sender's wallet for cancellation: %w", err)
	}

	now := time.Now()
	amountAV, err := attributevalue.Marshal(tx.Amount)
	if err != nil {
		return fmt.Errorf("failed to marshal amount for cancellation: %w", err)
	}

	cancelledStatusAV, err := attributevalue.Marshal(models.CANCELLED)
	if err != nil {
		return fmt.Errorf("failed to marshal cancelled status: %w", err)
	}
	nowAV, err := attributevalue.Marshal(now)
	if err != nil {
		return fmt.Errorf("failed to marshal timestamp for cancellation: %w", err)
	}

	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Update: &types.Update{
					TableName: aws.String(s.WalletsTableName),
					Key:       map[string]types.AttributeValue{"user_id": &types.AttributeValueMemberS{Value: tx.FromUserId}},
					UpdateExpression:    aws.String("SET balance = balance + :amount, reserved = reserved - :amount, version = version + :inc"),
					ConditionExpression: aws.String("version = :version"),
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":amount":   amountAV,
						":version":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", senderWallet.Version)},
						":inc":      &types.AttributeValueMemberN{Value: "1"},
					},
				},
			},
			{
				Update: &types.Update{
					TableName: aws.String(s.TransactionsTableName),
					Key:       map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: tx.Id.String()}},
					UpdateExpression:    aws.String("SET #status = :cancelled_status, updated_at = :now"),
					ConditionExpression: aws.String("#status = :reserved_status"),
					ExpressionAttributeNames: map[string]string{
						"#status": "status",
					},
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":cancelled_status": cancelledStatusAV,
						":reserved_status":  &types.AttributeValueMemberS{Value: string(models.RESERVED)},
						":now":              nowAV,
					},
				},
			},
		},
	}

	_, err = s.Client.TransactWriteItems(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to execute cancellation transaction: %w", err)
	}

	return nil
}

// SettleTransaction performs the final atomic settlement of a transaction.
func (s *Store) SettleTransaction(ctx context.Context, tx *models.Transaction) error {
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
		TransactionID: tx.Id.String(),
		EntryID:       uuid.New().String(),
		AccountID:     tx.FromUserId,
		Debit:         tx.Amount,
		Description:   fmt.Sprintf("Settlement for transaction %s", tx.Id.String()),
		Timestamp:     now,
		GSI1PK:        "LEDGER_ENTRIES",
	}
	creditEntry := models.LedgerEntry{
		TransactionID: tx.Id.String(),
		EntryID:       uuid.New().String(),
		AccountID:     tx.ToUserId,
		Credit:        tx.Amount,
		Description:   fmt.Sprintf("Settlement for transaction %s", tx.Id.String()),
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
	approvedStatusAV, err := attributevalue.Marshal(models.APPROVED)
	if err != nil {
		return fmt.Errorf("failed to marshal approved status: %w", err)
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
					UpdateExpression:    aws.String("SET reserved = reserved - :amount, version = version + :inc, ttl = :ttl"),
					ConditionExpression: aws.String("reserved >= :amount AND version = :version"),
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
					UpdateExpression:    aws.String("SET balance = balance + :amount, version = version + :inc, ttl = :ttl"),
					ConditionExpression: aws.String("version = :version"),
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
					Key:       map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: tx.Id.String()}},
					UpdateExpression:    aws.String("SET #status = :completed_status, updated_at = :now"),
					ConditionExpression: aws.String("#status = :approved_status"),
					ExpressionAttributeNames: map[string]string{
						"#status": "status",
					},
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":completed_status": completedStatusAV,
						":approved_status":  approvedStatusAV,
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

func (s *Store) GetStuckTransactions(ctx context.Context, maxAge time.Duration) ([]models.Transaction, error) {
	// Calculate the cutoff time.
	cutoffTime := time.Now().Add(-maxAge)
	cutoffTimeStr, err := cutoffTime.MarshalText()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cutoff time: %w", err)
	}

	// Prepare the query input.
	input := &dynamodb.QueryInput{
		TableName:              aws.String(s.TransactionsTableName),
		IndexName:              aws.String(stuckTransactionGSI),
		KeyConditionExpression: aws.String("#status = :status"),
		FilterExpression:       aws.String("created_at < :cutoff"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: string(models.RESERVED)},
			":cutoff": &types.AttributeValueMemberS{Value: string(cutoffTimeStr)},
		},
	}

	// Execute the query.
	result, err := s.Client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query for stuck transactions: %w", err)
	}

	// Unmarshal the results.
	var transactions []models.Transaction
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &transactions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stuck transactions: %w", err)
	}

	return transactions, nil
}

// ListTransactionsByUserID retrieves all transactions for a specific user.
const ledgerGSI = "gsi1pk-timestamp-index"

func (s *Store) ListLedgerEntries(ctx context.Context, limit int32) ([]models.LedgerEntry, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(s.LedgerTableName),
		IndexName:              aws.String(ledgerGSI),
		KeyConditionExpression: aws.String("gsi1pk = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "LEDGER_ENTRIES"},
		},
		ScanIndexForward: aws.Bool(false), // Sort by timestamp in descending order
		Limit:            &limit,
	}

	result, err := s.Client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query for ledger entries: %w", err)
	}

	var entries []models.LedgerEntry
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &entries); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ledger entries: %w", err)
	}

	return entries, nil
}

func (s *Store) ListTransactionsByUserID(ctx context.Context, userID string) ([]models.Transaction, error) {
	// Prepare the query input.
	input := &dynamodb.QueryInput{
		TableName:              aws.String(s.TransactionsTableName),
		IndexName:              aws.String(fromUserIDIndex),
		KeyConditionExpression: aws.String("from_user_id = :userID"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userID": &types.AttributeValueMemberS{Value: userID},
		},
	}

	// Execute the query.
	result, err := s.Client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query for transactions by user ID: %w", err)
	}

	// Unmarshal the results.
	var transactions []models.Transaction
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &transactions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transactions: %w", err)
	}

	return transactions, nil
}
