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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateWallet(t *testing.T) {
	wallet := &models.Wallet{UserId: "test-user"}

	t.Run("Success", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		mockClient.On("PutItem", mock.Anything, mock.Anything).Return(&dynamodb.PutItemOutput{}, nil)

		store := New(mockClient, "transactions", "wallets", "ledger", "")
		createdWallet, err := store.CreateWallet(context.Background(), wallet)

		assert.NoError(t, err)
		assert.Equal(t, wallet, createdWallet)
		mockClient.AssertExpectations(t)
	})

	t.Run("Conflict", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		mockClient.On("PutItem", mock.Anything, mock.Anything).Return(nil, &types.ConditionalCheckFailedException{})

		store := New(mockClient, "transactions", "wallets", "ledger", "")
		_, err := store.CreateWallet(context.Background(), wallet)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "wallet for user ID test-user already exists")
		mockClient.AssertExpectations(t)
	})

	t.Run("Storage Error", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		mockClient.On("PutItem", mock.Anything, mock.Anything).Return(nil, errors.New("some other storage error"))

		store := New(mockClient, "transactions", "wallets", "ledger", "")
		_, err := store.CreateWallet(context.Background(), wallet)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create wallet in DynamoDB")
		mockClient.AssertExpectations(t)
	})
}

func TestDeleteWallet(t *testing.T) {
	userID := "test-user"

	t.Run("Success", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		mockClient.On("DeleteItem", mock.Anything, mock.Anything).Return(&dynamodb.DeleteItemOutput{}, nil)

		store := New(mockClient, "transactions", "wallets", "ledger", "")
		err := store.DeleteWallet(context.Background(), userID)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		mockClient.On("DeleteItem", mock.Anything, mock.Anything).Return(nil, &types.ConditionalCheckFailedException{})

		store := New(mockClient, "transactions", "wallets", "ledger", "")
		err := store.DeleteWallet(context.Background(), userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "wallet for user ID test-user not found")
		mockClient.AssertExpectations(t)
	})

	t.Run("Storage Error", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		mockClient.On("DeleteItem", mock.Anything, mock.Anything).Return(nil, errors.New("some other storage error"))

		store := New(mockClient, "transactions", "wallets", "ledger", "")
		err := store.DeleteWallet(context.Background(), userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete wallet from DynamoDB")
		mockClient.AssertExpectations(t)
	})
}

func TestGetWallet(t *testing.T) {
	userID := "test-user"
	wallet := &models.Wallet{UserId: userID, Balance: 100}

	t.Run("Success", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		walletAV, _ := attributevalue.MarshalMap(wallet)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Return(&dynamodb.GetItemOutput{Item: walletAV}, nil)

		store := New(mockClient, "transactions", "wallets", "ledger", "")
		retrievedWallet, err := store.GetWallet(context.Background(), userID)

		assert.NoError(t, err)
		assert.Equal(t, wallet, retrievedWallet)
		mockClient.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Return(&dynamodb.GetItemOutput{Item: nil}, nil)

		store := New(mockClient, "transactions", "wallets", "ledger", "")
		_, err := store.GetWallet(context.Background(), userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "wallet for user ID test-user not found")
		mockClient.AssertExpectations(t)
	})

	t.Run("Storage Error", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Return(nil, errors.New("some other storage error"))

		store := New(mockClient, "transactions", "wallets", "ledger", "")
		_, err := store.GetWallet(context.Background(), userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get wallet from DynamoDB")
		mockClient.AssertExpectations(t)
	})
}

func TestListWallets(t *testing.T) {
	wallets := []models.Wallet{{UserId: "test-user-1"}, {UserId: "test-user-2"}}

	t.Run("Success", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		var walletsAV []map[string]types.AttributeValue
		for _, w := range wallets {
			av, err := attributevalue.MarshalMap(w)
			assert.NoError(t, err)
			walletsAV = append(walletsAV, av)
		}
		mockClient.On("Scan", mock.Anything, mock.Anything).Return(&dynamodb.ScanOutput{Items: walletsAV}, nil)

		store := New(mockClient, "transactions", "wallets", "ledger", "")
		retrievedWallets, err := store.ListWallets(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, wallets, retrievedWallets)
		mockClient.AssertExpectations(t)
	})

	t.Run("Storage Error", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		mockClient.On("Scan", mock.Anything, mock.Anything).Return(nil, errors.New("some other storage error"))

		store := New(mockClient, "transactions", "wallets", "ledger", "")
		_, err := store.ListWallets(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scan wallets table")
		mockClient.AssertExpectations(t)
	})
}
