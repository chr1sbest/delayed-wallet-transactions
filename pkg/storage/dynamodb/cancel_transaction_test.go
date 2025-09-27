package dynamodb

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/chris/delayed-wallet-transactions/pkg/models"
	"github.com/chris/delayed-wallet-transactions/pkg/storage"
	"github.com/chris/delayed-wallet-transactions/pkg/storage/dynamodb/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCancelTransaction(t *testing.T) {
	txID := uuid.New().String()
	tx := &models.Transaction{Id: txID, FromUserId: "user1", Status: models.RESERVED, Amount: 100}
	senderWallet := &models.Wallet{UserId: "user1", Balance: 100, Reserved: 100, Version: 1}

	t.Run("Success", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, TransactionsTableName: "transactions", WalletsTableName: "wallets"}

		// Mock GetTransaction's GetItem call
		txAV, _ := attributevalue.MarshalMap(tx)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Once().Return(&dynamodb.GetItemOutput{Item: txAV}, nil)

		// Mock GetWallet's GetItem call
		walletAV, _ := attributevalue.MarshalMap(senderWallet)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Once().Return(&dynamodb.GetItemOutput{Item: walletAV}, nil)

		// Mock TransactWriteItems call
		mockClient.On("TransactWriteItems", mock.Anything, mock.Anything).Return(&dynamodb.TransactWriteItemsOutput{}, nil)

		err := store.CancelTransaction(context.Background(), txID)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("GetTransaction Fails", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, TransactionsTableName: "transactions"}

		mockClient.On("GetItem", mock.Anything, mock.Anything).Return(nil, errors.New("get item failed"))

		err := store.CancelTransaction(context.Background(), txID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get transaction for cancellation")
		mockClient.AssertExpectations(t)
	})

	t.Run("Transaction Not Cancellable", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, TransactionsTableName: "transactions"}

		completedTx := &models.Transaction{Id: txID, Status: models.COMPLETED}
		txAV, _ := attributevalue.MarshalMap(completedTx)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Return(&dynamodb.GetItemOutput{Item: txAV}, nil)

		err := store.CancelTransaction(context.Background(), txID)

		assert.Error(t, err)
		assert.Equal(t, storage.ErrTransactionNotCancellable, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("GetWallet Fails", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, TransactionsTableName: "transactions", WalletsTableName: "wallets"}

		txAV, _ := attributevalue.MarshalMap(tx)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Once().Return(&dynamodb.GetItemOutput{Item: txAV}, nil)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Once().Return(nil, errors.New("get wallet failed"))

		err := store.CancelTransaction(context.Background(), txID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get sender's wallet for cancellation")
		mockClient.AssertExpectations(t)
	})

	t.Run("Transaction Fails", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, TransactionsTableName: "transactions", WalletsTableName: "wallets"}

		txAV, _ := attributevalue.MarshalMap(tx)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Once().Return(&dynamodb.GetItemOutput{Item: txAV}, nil)

		walletAV, _ := attributevalue.MarshalMap(senderWallet)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Once().Return(&dynamodb.GetItemOutput{Item: walletAV}, nil)

		mockClient.On("TransactWriteItems", mock.Anything, mock.Anything).Return(nil, errors.New("transaction failed"))

		err := store.CancelTransaction(context.Background(), txID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute cancellation transaction")
		mockClient.AssertExpectations(t)
	})
}
