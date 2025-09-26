package storage

import (
	"context"

	"github.com/chris/delayed-wallet-transactions/pkg/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Storage defines the interface for interacting with the transaction data layer.
type Storage interface {
	// CreateTransaction creates a new transaction and stores it.
	// This is where the "reservation" of funds will happen.
	CreateTransaction(ctx context.Context, newTx api.NewTransaction) (*api.Transaction, error)

	// GetTransaction retrieves a transaction by its ID.
	GetTransaction(ctx context.Context, txID openapi_types.UUID) (*api.Transaction, error)

	// GetWallet retrieves a user's wallet by their user ID.
	GetWallet(ctx context.Context, userID string) (*api.Wallet, error)

	// SettleTransaction performs the final atomic settlement of a transaction.
	SettleTransaction(ctx context.Context, tx *api.Transaction) error
}
