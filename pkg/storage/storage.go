package storage

import (
	"context"
	"time"

	"github.com/chris/delayed-wallet-transactions/pkg/models"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Storage defines the interface for interacting with the transaction data layer.
type Storage interface {
	// CreateTransaction creates a new transaction and stores it.
	CreateTransaction(ctx context.Context, newTx *models.Transaction) (*models.Transaction, error)

	// GetTransaction retrieves a transaction by its ID.
	GetTransaction(ctx context.Context, txID openapi_types.UUID) (*models.Transaction, error)

	// GetWallet retrieves a user's wallet by their user ID.
	GetWallet(ctx context.Context, userID string) (*models.Wallet, error)

	// CreateWallet creates a new wallet for a user.
	CreateWallet(ctx context.Context, wallet *models.Wallet) (*models.Wallet, error)

	// DeleteWallet deletes a user's wallet.
	DeleteWallet(ctx context.Context, userID string) error

	// SettleTransaction performs the final atomic settlement of a transaction.
	SettleTransaction(ctx context.Context, tx *models.Transaction) error

	// GetStuckTransactions retrieves transactions that are in a 'RESERVED' state for longer than the specified duration.
	GetStuckTransactions(ctx context.Context, maxAge time.Duration) ([]models.Transaction, error)
}
