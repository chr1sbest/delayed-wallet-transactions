package storage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chris/delayed-wallet-transactions/pkg/api"
	"github.com/chris/delayed-wallet-transactions/pkg/scheduler"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// DynamoDBStore implements the Storage interface using AWS DynamoDB.
type DynamoDBStore struct {
	Client                *dynamodb.Client
	Scheduler             scheduler.Scheduler
	TransactionsTableName string
	WalletsTableName      string
	LedgerTableName       string
}

// NewDynamoDBStore creates a new DynamoDBStore.
func NewDynamoDBStore(client *dynamodb.Client, scheduler scheduler.Scheduler, transactionsTable, walletsTable, ledgerTable string) *DynamoDBStore {
	return &DynamoDBStore{
		Client:                client,
		Scheduler:             scheduler,
		TransactionsTableName: transactionsTable,
		WalletsTableName:      walletsTable,
		LedgerTableName:       ledgerTable,
	}
}

// Make sure we conform to the interface
var _ Storage = (*DynamoDBStore)(nil)

// CreateTransaction atomically reserves funds from the sender's wallet and creates a new transaction record.
func (s *DynamoDBStore) CreateTransaction(ctx context.Context, newTx api.NewTransaction) (*api.Transaction, error) {
	// 1. Get the current state of the sender's wallet.
	senderWallet, err := s.GetWallet(ctx, newTx.FromUserId)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender's wallet: %w", err)
	}

	// 2. Generate a new UUID for the transaction.
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate transaction ID: %w", err)
	}

	// 3. Create the full transaction object.
	now := time.Now()
	tx := &api.Transaction{
		Id:          openapi_types.UUID(id),
		FromUserId:  newTx.FromUserId,
		ToUserId:    newTx.ToUserId,
		Amount:      newTx.Amount,
		Currency:    newTx.Currency,
		Status:      api.RESERVED,
		ScheduledAt: newTx.ScheduledAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Marshal the transaction for the Put operation.
	txAV, err := attributevalue.MarshalMap(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transaction: %w", err)
	}

	// Marshal the amount for the wallet update.
	amountAV, err := attributevalue.Marshal(newTx.Amount)
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
						"user_id": &types.AttributeValueMemberS{Value: newTx.FromUserId},
					},
					UpdateExpression:    aws.String("SET balance = balance - :amount, reserved = reserved + :amount, version = version + :inc"),
					ConditionExpression: aws.String("balance >= :amount AND version = :version"),
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":amount":   amountAV,
						":version":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", senderWallet.Version)},
						":inc":      &types.AttributeValueMemberN{Value: "1"},
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
		// Check for specific transaction cancellation reasons.
		var txc *types.TransactionCanceledException
		if errors.As(err, &txc) {
			for _, reason := range txc.CancellationReasons {
				if *reason.Code == "ConditionalCheckFailed" {
					// This could be due to insufficient funds, a version mismatch (race condition), or the transaction already existing.
					// A more robust implementation would inspect the reason message or the item index to provide a more specific error.
					return nil, fmt.Errorf("transaction failed: conditional check failed (insufficient funds, race condition, or duplicate transaction)")
				}
			}
		}
		return nil, fmt.Errorf("failed to execute transaction: %w", err)
	}

	// 6. If the database transaction was successful, enqueue it for processing.
	if err := s.Scheduler.ScheduleTransaction(ctx, tx); err != nil {
		// If this fails, the transaction is in the DB but not scheduled.
		// A scavenger service would be needed to find and re-enqueue these.
		// For now, we will log this critical error.
		log.Printf("CRITICAL: transaction %s created but failed to enqueue: %v", tx.Id, err)
	}

	return tx, nil
}

// GetTransaction retrieves a transaction from DynamoDB by its ID.
func (s *DynamoDBStore) GetTransaction(ctx context.Context, txID openapi_types.UUID) (*api.Transaction, error) {
	// Create the key for the GetItem request.
	key, err := attributevalue.MarshalMap(map[string]string{"id": txID.String()})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transaction ID: %w", err)
	}

	// Create the GetItem input.
	input := &dynamodb.GetItemInput{
		TableName: aws.String(s.TransactionsTableName),
		Key:       key,
	}

	// Execute the GetItem request.
	result, err := s.Client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction from DynamoDB: %w", err)
	}

	// Check if the item was found.
	if result.Item == nil {
		return nil, fmt.Errorf("transaction with ID %s not found", txID.String())
	}

	// Unmarshal the result into a Transaction struct.
	var tx api.Transaction
	if err := attributevalue.UnmarshalMap(result.Item, &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	return &tx, nil
}

// GetWallet retrieves a user's wallet from DynamoDB by their user ID.
func (s *DynamoDBStore) GetWallet(ctx context.Context, userID string) (*api.Wallet, error) {
	// Create the key for the GetItem request.
	key, err := attributevalue.MarshalMap(map[string]string{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal wallet user ID: %w", err)
	}

	// Create the GetItem input.
	input := &dynamodb.GetItemInput{
		TableName: aws.String(s.WalletsTableName),
		Key:       key,
	}

	// Execute the GetItem request.
	result, err := s.Client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet from DynamoDB: %w", err)
	}

	// Check if the item was found.
	if result.Item == nil {
		return nil, fmt.Errorf("wallet for user ID %s not found", userID)
	}

	// Unmarshal the result into a Wallet struct.
	var wallet api.Wallet
	if err := attributevalue.UnmarshalMap(result.Item, &wallet); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet: %w", err)
	}

	return &wallet, nil
}

// LedgerEntry represents a single entry in the double-entry ledger.
// TODO: This should be moved to the api package and defined in the spec.
type LedgerEntry struct {
	TransactionID string    `json:"transaction_id"`
	EntryID       string    `json:"entry_id"`
	AccountID     string    `json:"account_id"`
	Debit         float64   `json:"debit,omitempty"`
	Credit        float64   `json:"credit,omitempty"`
	Description   string    `json:"description"`
	Timestamp     time.Time `json:"timestamp"`
}

// SettleTransaction performs the final atomic settlement of a transaction.
func (s *DynamoDBStore) SettleTransaction(ctx context.Context, tx *api.Transaction) error {
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
	debitEntry := LedgerEntry{
		TransactionID: tx.Id.String(),
		EntryID:       uuid.New().String(),
		AccountID:     tx.FromUserId,
		Debit:         tx.Amount,
		Description:   fmt.Sprintf("Settlement for transaction %s", tx.Id.String()),
		Timestamp:     now,
	}
	creditEntry := LedgerEntry{
		TransactionID: tx.Id.String(),
		EntryID:       uuid.New().String(),
		AccountID:     tx.ToUserId,
		Credit:        tx.Amount,
		Description:   fmt.Sprintf("Settlement for transaction %s", tx.Id.String()),
		Timestamp:     now,
	}
	debitAV, err := attributevalue.MarshalMap(debitEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal debit entry: %w", err)
	}
	creditAV, err := attributevalue.MarshalMap(creditEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal credit entry: %w", err)
	}

	// 4. Construct the TransactWriteItems input.
	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				// Operation 1: Update sender's wallet.
				Update: &types.Update{
					TableName: aws.String(s.WalletsTableName),
					Key: map[string]types.AttributeValue{"user_id": &types.AttributeValueMemberS{Value: tx.FromUserId}},
					UpdateExpression:    aws.String("SET reserved = reserved - :amount, version = version + :inc"),
					ConditionExpression: aws.String("reserved >= :amount AND version = :version"),
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":amount":   amountAV,
						":version":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", senderWallet.Version)},
						":inc":      &types.AttributeValueMemberN{Value: "1"},
					},
				},
			},
			{
				// Operation 2: Update receiver's wallet.
				Update: &types.Update{
					TableName: aws.String(s.WalletsTableName),
					Key: map[string]types.AttributeValue{"user_id": &types.AttributeValueMemberS{Value: tx.ToUserId}},
					UpdateExpression:    aws.String("SET balance = balance + :amount, version = version + :inc"),
					ConditionExpression: aws.String("version = :version"),
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":amount":   amountAV,
						":version":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", receiverWallet.Version)},
						":inc":      &types.AttributeValueMemberN{Value: "1"},
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
		},
	}

	// 5. Execute the transaction.
	_, err = s.Client.TransactWriteItems(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to execute settlement transaction: %w", err)
	}

	// After success, we would also update the transaction status to COMPLETED.
	// This is omitted for brevity but would be an UpdateItem call on the Transactions table.

	return nil
}
