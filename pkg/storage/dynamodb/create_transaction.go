package dynamodb

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"errors"

	"github.com/chris/delayed-wallet-transactions/pkg/models"
	"github.com/chris/delayed-wallet-transactions/pkg/storage"
	"github.com/google/uuid"
)

// CreateTransaction atomically reserves funds from the sender's wallet and creates a new transaction record.
func (s *Store) CreateTransaction(ctx context.Context, tx *models.Transaction) (*models.Transaction, error) {
	// 1. Get the current state of the sender's wallet.
	senderWallet, err := s.GetWallet(ctx, tx.FromUserId)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender's wallet: %w", err)
	}

	// 2. Complete the transaction object with server-side details.
	now := time.Now()
	tx.Id = uuid.New().String()
	tx.Status = models.RESERVED
	tx.CreatedAt = now
	tx.UpdatedAt = now
	tx.TTL = time.Now().Add(24 * time.Hour).Unix()

	slog.Log(ctx, slog.LevelDebug, "creating transaction", "transaction", tx)

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
					UpdateExpression:    aws.String("SET balance = balance - :amount, reserved = reserved + :amount, version = version + :inc, #ttl = :ttl"),
					ConditionExpression: aws.String("balance >= :amount AND version = :version"),
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
		var tce *types.TransactionCanceledException
		if errors.As(err, &tce) {
			// Check if the first operation (updating the sender's wallet) failed due to a conditional check.
			if len(tce.CancellationReasons) > 0 && *tce.CancellationReasons[0].Code == "ConditionalCheckFailed" {
				return nil, storage.ErrInsufficientFunds
			}
		}
		return nil, fmt.Errorf("failed to execute transaction: %w", err)
	}

	return tx, nil
}
