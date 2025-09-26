package storage

import (
	"context"

	"github.com/chris/delayed-wallet-transactions/pkg/models"
)

// WalletStore defines the interface for managing wallets.
type WalletStore interface {
	// GetWallet retrieves a user's wallet by their user ID.
	GetWallet(ctx context.Context, userID string) (*models.Wallet, error)

	// CreateWallet creates a new wallet for a user.
	CreateWallet(ctx context.Context, wallet *models.Wallet) (*models.Wallet, error)

	// DeleteWallet deletes a user's wallet.
	DeleteWallet(ctx context.Context, userID string) error

	// ListWallets retrieves all wallets from the storage.
	ListWallets(ctx context.Context) ([]models.Wallet, error)
}
