package handlers

import (
	"github.com/chris/delayed-wallet-transactions/pkg/handlers/ledger"
	"github.com/chris/delayed-wallet-transactions/pkg/handlers/transactions"
	"github.com/chris/delayed-wallet-transactions/pkg/handlers/wallets"
	"github.com/chris/delayed-wallet-transactions/pkg/storage"
	"github.com/chris/delayed-wallet-transactions/pkg/scheduler"
)

// ApiHandler implements the generated server interface by embedding entity-specific handlers.
type ApiHandler struct {
	*transactions.TransactionsHandler
	*wallets.WalletsHandler
	*ledger.LedgerHandler
	Scheduler scheduler.CronScheduler
}

// NewApiHandler creates a new ApiHandler with a storage dependency.
func NewApiHandler(store storage.ApiStore, scheduler scheduler.CronScheduler) *ApiHandler {
	return &ApiHandler{
		TransactionsHandler: transactions.NewTransactionsHandler(store, scheduler),
		WalletsHandler:      wallets.NewWalletsHandler(store),
		LedgerHandler:       ledger.NewLedgerHandler(store),
		Scheduler:           scheduler,
	}
}
