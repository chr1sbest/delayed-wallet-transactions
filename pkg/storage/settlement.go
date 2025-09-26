package storage

import (
	"context"

	"github.com/chris/delayed-wallet-transactions/pkg/models"
)

// SettlementStore defines the highly-privileged interface for settling a transaction.
// This operation is complex and involves atomic writes across multiple tables (Transactions, Wallets, Ledger).
// It should only be exposed to the component responsible for final settlement.
type SettlementStore interface {
	// SettleTransaction performs the final atomic settlement of a transaction.
	SettleTransaction(ctx context.Context, tx *models.Transaction) error
}
