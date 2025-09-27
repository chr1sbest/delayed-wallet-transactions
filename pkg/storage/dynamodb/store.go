package dynamodb

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoDBAPI interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
	TransactWriteItems(ctx context.Context, params *dynamodb.TransactWriteItemsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

//go:generate mockery --name DynamoDBAPI --output ./mocks --outpkg mocks --case=underscore

// Store implements the Storage interface using AWS DynamoDB.
type Store struct {
	Client                        DynamoDBAPI
	TransactionsTableName         string
	WalletsTableName              string
	LedgerTableName               string
	WebsocketConnectionsTableName string
}

// New creates a new Store with all table dependencies.
func New(client DynamoDBAPI, transactionsTable, walletsTable, ledgerTable, websocketConnectionsTable string) *Store {
	return &Store{
		Client:                client,
		TransactionsTableName: transactionsTable,
		WalletsTableName:      walletsTable,
		LedgerTableName:             ledgerTable,
		WebsocketConnectionsTableName: websocketConnectionsTable,
	}
}

// NewTransactionReader creates a new store that only requires the transactions table.
// It's used by components that only need to read from the transactions table.
func NewTransactionReader(client DynamoDBAPI, transactionsTable string) *Store {
	return &Store{
		Client:                client,
		TransactionsTableName: transactionsTable,
	}
}
