package dynamodb

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chris/delayed-wallet-transactions/pkg/models"
	"github.com/chris/delayed-wallet-transactions/pkg/storage"
	"github.com/chris/delayed-wallet-transactions/pkg/storage/dynamodb/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateTransaction(t *testing.T) {
	tx := &models.Transaction{FromUserId: "user1", ToUserId: "user2", Amount: 100}
	senderWallet := &models.Wallet{UserId: "user1", Balance: 200, Version: 1}

	t.Run("Success", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, WalletsTableName: "wallets", TransactionsTableName: "transactions"}

		// Mock the initial GetWallet call
		senderWalletAV, _ := attributevalue.MarshalMap(senderWallet)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Once().Return(&dynamodb.GetItemOutput{Item: senderWalletAV}, nil)

		// Mock the TransactWriteItems call
		mockClient.On("TransactWriteItems", mock.Anything, mock.Anything).Once().Return(&dynamodb.TransactWriteItemsOutput{}, nil)

		result, err := store.CreateTransaction(context.Background(), tx)

		assert.NoError(t, err)
		assert.Equal(t, tx, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("GetWallet Fails", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, WalletsTableName: "wallets"}

		mockClient.On("GetItem", mock.Anything, mock.Anything).Return(nil, errors.New("get wallet failed"))

		_, err := store.CreateTransaction(context.Background(), tx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get sender's wallet")
		mockClient.AssertExpectations(t)
	})

	t.Run("Transaction Fails", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, WalletsTableName: "wallets", TransactionsTableName: "transactions"}

		senderWalletAV, _ := attributevalue.MarshalMap(senderWallet)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Return(&dynamodb.GetItemOutput{Item: senderWalletAV}, nil)
		mockClient.On("TransactWriteItems", mock.Anything, mock.Anything).Return(nil, errors.New("transaction failed"))

		_, err := store.CreateTransaction(context.Background(), tx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute transaction")
		mockClient.AssertExpectations(t)
	})

	t.Run("Insufficient Funds", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, WalletsTableName: "wallets", TransactionsTableName: "transactions"}

		senderWalletAV, _ := attributevalue.MarshalMap(senderWallet)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Return(&dynamodb.GetItemOutput{Item: senderWalletAV}, nil)
		cancellationReasons := make([]types.CancellationReason, 1)
		cancellationReasons[0] = types.CancellationReason{Code: aws.String("ConditionalCheckFailed")}
		mockClient.On("TransactWriteItems", mock.Anything, mock.Anything).Return(nil, &types.TransactionCanceledException{CancellationReasons: cancellationReasons})

		_, err := store.CreateTransaction(context.Background(), tx)

		assert.Error(t, err)
		assert.ErrorIs(t, err, storage.ErrInsufficientFunds)
		mockClient.AssertExpectations(t)
	})
}
