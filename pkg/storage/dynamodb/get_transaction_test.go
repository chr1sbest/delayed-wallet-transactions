package dynamodb

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/chris/delayed-wallet-transactions/pkg/models"
	"github.com/chris/delayed-wallet-transactions/pkg/storage/dynamodb/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTransaction(t *testing.T) {
	txID := uuid.New().String()
	tx := &models.Transaction{Id: txID, FromUserId: "user1", ToUserId: "user2", Amount: 100}

	t.Run("Success", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, TransactionsTableName: "transactions"}

		txAV, _ := attributevalue.MarshalMap(tx)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Return(&dynamodb.GetItemOutput{Item: txAV}, nil)

		result, err := store.GetTransaction(context.Background(), txID)

		assert.NoError(t, err)
		assert.Equal(t, tx, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, TransactionsTableName: "transactions"}

		mockClient.On("GetItem", mock.Anything, mock.Anything).Return(&dynamodb.GetItemOutput{Item: nil}, nil)

		_, err := store.GetTransaction(context.Background(), txID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		mockClient.AssertExpectations(t)
	})

	t.Run("Storage Error", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, TransactionsTableName: "transactions"}

		mockClient.On("GetItem", mock.Anything, mock.Anything).Return(nil, errors.New("get item failed"))

		_, err := store.GetTransaction(context.Background(), txID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get transaction from DynamoDB")
		mockClient.AssertExpectations(t)
	})
}
