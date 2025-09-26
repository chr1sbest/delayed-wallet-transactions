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
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"time"
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
		// This is a bit brittle, but it's the best we can do without a more specific error type.
		assert.Contains(t, err.Error(), "failed to execute transaction")
		mockClient.AssertExpectations(t)
	})
}

func TestGetTransaction(t *testing.T) {
	txID := openapi_types.UUID(uuid.New())
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

func TestCancelTransaction(t *testing.T) {
	txID := openapi_types.UUID(uuid.New())
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

func TestSettleTransaction(t *testing.T) {
	txID := openapi_types.UUID(uuid.New())
	tx := &models.Transaction{Id: txID, FromUserId: "user1", ToUserId: "user2", Amount: 100, Status: models.APPROVED}
	senderWallet := &models.Wallet{UserId: "user1", Balance: 100, Reserved: 100, Version: 1}
	receiverWallet := &models.Wallet{UserId: "user2", Balance: 50, Version: 1}

	t.Run("Success", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, TransactionsTableName: "transactions", WalletsTableName: "wallets", LedgerTableName: "ledger"}

		// Mock GetWallet calls
		senderWalletAV, _ := attributevalue.MarshalMap(senderWallet)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Once().Return(&dynamodb.GetItemOutput{Item: senderWalletAV}, nil)
		receiverWalletAV, _ := attributevalue.MarshalMap(receiverWallet)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Once().Return(&dynamodb.GetItemOutput{Item: receiverWalletAV}, nil)

		// Mock TransactWriteItems call
		mockClient.On("TransactWriteItems", mock.Anything, mock.Anything).Return(&dynamodb.TransactWriteItemsOutput{}, nil)

		err := store.SettleTransaction(context.Background(), tx)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("Get Sender Wallet Fails", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient}

		mockClient.On("GetItem", mock.Anything, mock.Anything).Return(nil, errors.New("get wallet failed"))

		err := store.SettleTransaction(context.Background(), tx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get sender's wallet for settlement")
		mockClient.AssertExpectations(t)
	})

	t.Run("Get Receiver Wallet Fails", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient}

		senderWalletAV, _ := attributevalue.MarshalMap(senderWallet)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Once().Return(&dynamodb.GetItemOutput{Item: senderWalletAV}, nil)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Once().Return(nil, errors.New("get wallet failed"))

		err := store.SettleTransaction(context.Background(), tx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get receiver's wallet for settlement")
		mockClient.AssertExpectations(t)
	})

	t.Run("Transaction Fails", func(t *testing.T) {
		mockClient := new(mocks.DynamoDBAPI)
		store := &Store{Client: mockClient, TransactionsTableName: "transactions", WalletsTableName: "wallets", LedgerTableName: "ledger"}

		senderWalletAV, _ := attributevalue.MarshalMap(senderWallet)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Once().Return(&dynamodb.GetItemOutput{Item: senderWalletAV}, nil)
		receiverWalletAV, _ := attributevalue.MarshalMap(receiverWallet)
		mockClient.On("GetItem", mock.Anything, mock.Anything).Once().Return(&dynamodb.GetItemOutput{Item: receiverWalletAV}, nil)

		mockClient.On("TransactWriteItems", mock.Anything, mock.Anything).Return(nil, errors.New("transaction failed"))

		err := store.SettleTransaction(context.Background(), tx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute settlement transaction")
		mockClient.AssertExpectations(t)
	})
}

func TestGetStuckTransactions(t *testing.T) {
	stuckTxs := []models.Transaction{{Id: openapi_types.UUID(uuid.New())}, {Id: openapi_types.UUID(uuid.New())}}

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
	txs := []models.Transaction{{Id: openapi_types.UUID(uuid.New())}, {Id: openapi_types.UUID(uuid.New())}}

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
