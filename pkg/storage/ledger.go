package storage

import (
	"context"

	"github.com/chris/delayed-wallet-transactions/pkg/models"
)

// LedgerReader defines the interface for reading ledger data.
type LedgerReader interface {
	// ListLedgerEntries retrieves the most recent ledger entries.
	ListLedgerEntries(ctx context.Context, limit int32) ([]models.LedgerEntry, error)
}
