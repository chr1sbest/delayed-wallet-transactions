package dynamodb

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chris/delayed-wallet-transactions/pkg/models"
	"github.com/chris/delayed-wallet-transactions/pkg/storage/dynamodb/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetStuckTransactions(t *testing.T) {
	stuckTxs := []models.Transaction{{Id: uuid.New().String()}, {Id: uuid.New().String()}}

	t.Run("Success", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, TransactionsTableName: "transactions"}

		var stuckTxsAV []map[string]types.AttributeValue
		for _, tx := range stuckTxs {
			av, err := attributevalue.MarshalMap(tx)
			assert.NoError(t, err)
			stuckTxsAV = append(stuckTxsAV, av)
		}
		mockClient.On("Query", mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{Items: stuckTxsAV}, nil)

		result, err := store.GetStuckTransactions(context.Background(), time.Minute)

		assert.NoError(t, err)
		assert.Equal(t, stuckTxs, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("Storage Error", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, TransactionsTableName: "transactions"}

		mockClient.On("Query", mock.Anything, mock.Anything).Return(nil, errors.New("query failed"))

		_, err := store.GetStuckTransactions(context.Background(), time.Minute)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query for stuck transactions")
		mockClient.AssertExpectations(t)
	})
}

func TestListTransactionsByUserID(t *testing.T) {
	userID := "test-user"
	txs := []models.Transaction{{Id: uuid.New().String()}, {Id: uuid.New().String()}}

	t.Run("Success", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, TransactionsTableName: "transactions"}

		var txsAV []map[string]types.AttributeValue
		for _, tx := range txs {
			av, err := attributevalue.MarshalMap(tx)
			assert.NoError(t, err)
			txsAV = append(txsAV, av)
		}
		mockClient.On("Query", mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{Items: txsAV}, nil)

		result, err := store.ListTransactionsByUserID(context.Background(), userID)

		assert.NoError(t, err)
		assert.Equal(t, txs, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("Storage Error", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, TransactionsTableName: "transactions"}

		mockClient.On("Query", mock.Anything, mock.Anything).Return(nil, errors.New("query failed"))

		_, err := store.ListTransactionsByUserID(context.Background(), userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query for transactions by user ID")
		mockClient.AssertExpectations(t)
	})
}

func TestListLedgerEntries(t *testing.T) {
	entries := []models.LedgerEntry{{EntryID: uuid.New().String()}, {EntryID: uuid.New().String()}}

	t.Run("Success", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, LedgerTableName: "ledger"}

		var entriesAV []map[string]types.AttributeValue
		for _, entry := range entries {
			av, err := attributevalue.MarshalMap(entry)
			assert.NoError(t, err)
			entriesAV = append(entriesAV, av)
		}
		mockClient.On("Query", mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{Items: entriesAV}, nil)

		result, err := store.ListLedgerEntries(context.Background(), 2)

		assert.NoError(t, err)
		assert.Equal(t, entries, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("Storage Error", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, LedgerTableName: "ledger"}

		mockClient.On("Query", mock.Anything, mock.Anything).Return(nil, errors.New("query failed"))

		_, err := store.ListLedgerEntries(context.Background(), 2)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query for ledger entries")
		mockClient.AssertExpectations(t)
	})
}
