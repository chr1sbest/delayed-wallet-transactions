package handlers

import (
	"github.com/chris/delayed-wallet-transactions/pkg/api"
	"github.com/chris/delayed-wallet-transactions/pkg/handlers/ledger"
	"github.com/chris/delayed-wallet-transactions/pkg/handlers/transactions"
	"github.com/chris/delayed-wallet-transactions/pkg/handlers/wallets"
	"github.com/chris/delayed-wallet-transactions/pkg/scheduler"
	"github.com/chris/delayed-wallet-transactions/pkg/storage"
	"github.com/chris/delayed-wallet-transactions/pkg/websockets"
)

// ApiHandler implements the generated server interface by embedding entity-specific handlers.
type ApiHandler struct {
	*transactions.TransactionsHandler
	*wallets.WalletsHandler
	*ledger.LedgerHandler
}

// Make sure we conform to the generated server interface.
var _ api.ServerInterface = (*ApiHandler)(nil)

// NewApiHandler creates a new ApiHandler with a storage dependency.
func NewApiHandler(store storage.ApiStore, scheduler scheduler.CronScheduler, publisher websockets.Publisher) *ApiHandler {
	return &ApiHandler{
		TransactionsHandler: transactions.NewTransactionsHandler(store, scheduler, publisher),
		WalletsHandler:      wallets.NewWalletsHandler(store),
		LedgerHandler:       ledger.NewLedgerHandler(store),
	}
}
