package dynamodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/chris/delayed-wallet-transactions/pkg/models"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	stuckTransactionGSI = "status-created_at-index"
	fromUserIDIndex     = "from_user_id-index"
)

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
