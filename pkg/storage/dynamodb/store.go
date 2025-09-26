package dynamodb

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/chris/delayed-wallet-transactions/pkg/scheduler"
	"github.com/chris/delayed-wallet-transactions/pkg/storage"
)

// Store implements the Storage interface using AWS DynamoDB.
type Store struct {
	Client                *dynamodb.Client
	Scheduler             scheduler.Scheduler
	TransactionsTableName string
	WalletsTableName      string
	LedgerTableName       string
}

// New creates a new Store.
func New(client *dynamodb.Client, scheduler scheduler.Scheduler, transactionsTable, walletsTable, ledgerTable string) *Store {
	return &Store{
		Client:                client,
		Scheduler:             scheduler,
		TransactionsTableName: transactionsTable,
		WalletsTableName:      walletsTable,
		LedgerTableName:       ledgerTable,
	}
}

// Make sure we conform to the interface
var _ storage.Storage = (*Store)(nil)

// ErrInsufficientFunds is returned when a wallet has an insufficient balance for a transaction.
var ErrInsufficientFunds = errors.New("insufficient funds")

// ErrTransactionNotCancellable is returned when a transaction cannot be cancelled, e.g., because it's already completed or cancelled.
var ErrTransactionNotCancellable = errors.New("transaction not in a cancellable state")
