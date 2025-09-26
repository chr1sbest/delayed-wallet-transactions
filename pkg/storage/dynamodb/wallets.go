package dynamodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chris/delayed-wallet-transactions/pkg/models"
)

// CreateWallet creates a new wallet record in DynamoDB.
func (s *Store) CreateWallet(ctx context.Context, wallet *models.Wallet) (*models.Wallet, error) {
	wallet.TTL = time.Now().Add(24 * time.Hour).Unix()
	// Marshal the wallet object for the Put operation.
	walletAV, err := attributevalue.MarshalMap(wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal wallet: %w", err)
	}

	// Construct the PutItem input.
	input := &dynamodb.PutItemInput{
		TableName:           aws.String(s.WalletsTableName),
		Item:                walletAV,
		ConditionExpression: aws.String("attribute_not_exists(user_id)"), // Prevent overwriting existing wallets.
	}

	// Execute the PutItem operation.
	_, err = s.Client.PutItem(ctx, input)
	if err != nil {
		var condCheckFailed *types.ConditionalCheckFailedException
		if errors.As(err, &condCheckFailed) {
			return nil, fmt.Errorf("wallet for user ID %s already exists", wallet.UserId)
		}
		return nil, fmt.Errorf("failed to create wallet in DynamoDB: %w", err)
	}

	// Return the wallet object as it was successfully created.
	return wallet, nil
}

// DeleteWallet deletes a wallet record from DynamoDB.
func (s *Store) DeleteWallet(ctx context.Context, userID string) error {
	// Marshal the key for the DeleteItem operation.
	key, err := attributevalue.MarshalMap(map[string]string{"user_id": userID})
	if err != nil {
		return fmt.Errorf("failed to marshal wallet user ID for deletion: %w", err)
	}

	// Construct the DeleteItem input.
	input := &dynamodb.DeleteItemInput{
		TableName:           aws.String(s.WalletsTableName),
		Key:                 key,
		ConditionExpression: aws.String("attribute_exists(user_id)"), // Ensure the wallet exists before deleting.
	}

	// Execute the DeleteItem operation.
	_, err = s.Client.DeleteItem(ctx, input)
	if err != nil {
		var condCheckFailed *types.ConditionalCheckFailedException
		if errors.As(err, &condCheckFailed) {
			return fmt.Errorf("wallet for user ID %s not found", userID)
		}
		return fmt.Errorf("failed to delete wallet from DynamoDB: %w", err)
	}

	return nil
}

// GetWallet retrieves a user's wallet from DynamoDB by their user ID.
func (s *Store) GetWallet(ctx context.Context, userID string) (*models.Wallet, error) {
	key, err := attributevalue.MarshalMap(map[string]string{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal wallet user ID: %w", err)
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(s.WalletsTableName),
		Key:       key,
	}

	result, err := s.Client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet from DynamoDB: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("wallet for user ID %s not found", userID)
	}

	var wallet models.Wallet
	if err := attributevalue.UnmarshalMap(result.Item, &wallet); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet: %w", err)
	}

	return &wallet, nil
}

// ListWallets retrieves all wallets from DynamoDB.
func (s *Store) ListWallets(ctx context.Context) ([]models.Wallet, error) {
	// Prepare the Scan input.
	input := &dynamodb.ScanInput{
		TableName: aws.String(s.WalletsTableName),
	}

	// Execute the Scan operation.
	result, err := s.Client.Scan(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan wallets table: %w", err)
	}

	// Unmarshal the results.
	var wallets []models.Wallet
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &wallets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallets: %w", err)
	}

	return wallets, nil
}
