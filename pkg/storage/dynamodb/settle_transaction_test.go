package dynamodb

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chris/delayed-wallet-transactions/pkg/models"
	"github.com/chris/delayed-wallet-transactions/pkg/storage/dynamodb/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSettleTransaction(t *testing.T) {
	txID := uuid.New().String()
	tx := &models.Transaction{Id: txID, FromUserId: "user1", ToUserId: "user2", Amount: 100, Status: models.RESERVED}
	senderWallet := &models.Wallet{UserId: "user1", Balance: 100, Reserved: 100, Version: 1}
	receiverWallet := &models.Wallet{UserId: "user2", Balance: 50, Version: 1}

	t.Run("Success", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, TransactionsTableName: "transactions", WalletsTableName: "wallets", LedgerTableName: "ledger"}

		// Mock UpdateItem call to acquire lock
		mockClient.On("UpdateItem", mock.Anything, mock.AnythingOfType("*dynamodb.UpdateItemInput")).Return(&dynamodb.UpdateItemOutput{}, nil).Once()

		// Mock GetWallet calls
		senderWalletAV, _ := attributevalue.MarshalMap(senderWallet)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Return(&dynamodb.GetItemOutput{Item: senderWalletAV}, nil).Once()
		receiverWalletAV, _ := attributevalue.MarshalMap(receiverWallet)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Return(&dynamodb.GetItemOutput{Item: receiverWalletAV}, nil).Once()

		// Mock TransactWriteItems call for settlement
		mockClient.On("TransactWriteItems", mock.Anything, mock.AnythingOfType("*dynamodb.TransactWriteItemsInput")).Return(&dynamodb.TransactWriteItemsOutput{}, nil).Once()

		settled, err := store.SettleTransaction(context.Background(), tx)

		assert.NoError(t, err)
		assert.True(t, settled)
		mockClient.AssertExpectations(t)
	})

	t.Run("Get Sender Wallet Fails", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, TransactionsTableName: "transactions"}

		// This test checks what happens if the lock is acquired but the subsequent GetWallet fails.
		mockClient.On("UpdateItem", mock.Anything, mock.AnythingOfType("*dynamodb.UpdateItemInput")).Return(&dynamodb.UpdateItemOutput{}, nil).Once()
		mockClient.On("GetItem", mock.Anything, mock.Anything).Return(nil, errors.New("get wallet failed"))

		settled, err := store.SettleTransaction(context.Background(), tx)

		assert.Error(t, err)
		assert.False(t, settled)
		assert.Contains(t, err.Error(), "failed to get sender's wallet for settlement")
		mockClient.AssertExpectations(t)
	})

	t.Run("Get Receiver Wallet Fails", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, TransactionsTableName: "transactions", WalletsTableName: "wallets"}

		mockClient.On("UpdateItem", mock.Anything, mock.AnythingOfType("*dynamodb.UpdateItemInput")).Return(&dynamodb.UpdateItemOutput{}, nil).Once()
		senderWalletAV, _ := attributevalue.MarshalMap(senderWallet)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Once().Return(&dynamodb.GetItemOutput{Item: senderWalletAV}, nil)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Once().Return(nil, errors.New("get wallet failed"))

		settled, err := store.SettleTransaction(context.Background(), tx)

		assert.Error(t, err)
		assert.False(t, settled)
		assert.Contains(t, err.Error(), "failed to get receiver's wallet for settlement")
		mockClient.AssertExpectations(t)
	})

	t.Run("Transaction Fails", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, TransactionsTableName: "transactions", WalletsTableName: "wallets", LedgerTableName: "ledger"}

		mockClient.On("UpdateItem", mock.Anything, mock.AnythingOfType("*dynamodb.UpdateItemInput")).Return(&dynamodb.UpdateItemOutput{}, nil).Once()
		senderWalletAV, _ := attributevalue.MarshalMap(senderWallet)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Once().Return(&dynamodb.GetItemOutput{Item: senderWalletAV}, nil)
		receiverWalletAV, _ := attributevalue.MarshalMap(receiverWallet)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Once().Return(&dynamodb.GetItemOutput{Item: receiverWalletAV}, nil)

		mockClient.On("TransactWriteItems", mock.Anything, mock.Anything).Return(nil, errors.New("transaction failed"))

		settled, err := store.SettleTransaction(context.Background(), tx)

		assert.Error(t, err)
		assert.False(t, settled)
		assert.Contains(t, err.Error(), "failed to execute settlement transaction")
		mockClient.AssertExpectations(t)
	})

	t.Run("Lock Acquisition Fails", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, TransactionsTableName: "transactions"}

		// Mock UpdateItem call to fail with a conditional check failed exception
		mockClient.On("UpdateItem", mock.Anything, mock.AnythingOfType("*dynamodb.UpdateItemInput")).Return(nil, &types.ConditionalCheckFailedException{}).Once()

		settled, err := store.SettleTransaction(context.Background(), tx)

		assert.NoError(t, err)
		assert.False(t, settled)
		mockClient.AssertExpectations(t)
	})
}
