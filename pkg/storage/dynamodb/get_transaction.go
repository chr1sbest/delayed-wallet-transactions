package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/chris/delayed-wallet-transactions/pkg/models"
)

// GetTransaction retrieves a transaction from DynamoDB by its ID.
func (s *Store) GetTransaction(ctx context.Context, txID string) (*models.Transaction, error) {
	key, err := attributevalue.MarshalMap(map[string]string{"id": txID})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transaction ID: %w", err)
	}

	input := &dynamodb.GetItemInput{
		TableName: &s.TransactionsTableName,
		Key:       key,
	}

	result, err := s.Client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction from DynamoDB: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("transaction with ID %s not found", txID)
	}

	var tx models.Transaction
	if err := attributevalue.UnmarshalMap(result.Item, &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	return &tx, nil
}
