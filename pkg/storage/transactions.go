package storage

import (
	"context"
	"time"

	"github.com/chris/delayed-wallet-transactions/pkg/models"
)

// TransactionReader defines the interface for reading transaction data.
type TransactionReader interface {
	// GetTransaction retrieves a transaction by its ID.
	GetTransaction(ctx context.Context, txID string) (*models.Transaction, error)

	// GetStuckTransactions retrieves transactions that are in a 'RESERVED' state for longer than the specified duration.
	GetStuckTransactions(ctx context.Context, maxAge time.Duration) ([]models.Transaction, error)

	// ListTransactionsByUserID retrieves all transactions for a specific user.
	ListTransactionsByUserID(ctx context.Context, userID string) ([]models.Transaction, error)
}

// TransactionManager defines the interface for creating and managing transactions before settlement.
// This is suitable for components like the main API service.
type TransactionManager interface {
	// CreateTransaction creates a new transaction and returns the created transaction.
	CreateTransaction(ctx context.Context, newTx *models.Transaction) (*models.Transaction, error)

	// CancelTransaction cancels a transaction if it's in a cancellable state.
	CancelTransaction(ctx context.Context, txID string) error
}

// TransactionStore combines the reader and manager interfaces.
type TransactionStore interface {
	TransactionReader
	TransactionManager
}
