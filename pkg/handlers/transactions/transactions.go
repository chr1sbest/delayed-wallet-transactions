package transactions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/chris/delayed-wallet-transactions/pkg/api"
	"github.com/chris/delayed-wallet-transactions/pkg/mapping"
	"github.com/chris/delayed-wallet-transactions/pkg/scheduler"
	"github.com/chris/delayed-wallet-transactions/pkg/storage"
	"github.com/chris/delayed-wallet-transactions/pkg/websockets"
	"github.com/oapi-codegen/runtime/types"
)

// TransactionsHandler holds the dependencies for transaction-related handlers.
type TransactionsHandler struct {
	Store     storage.ApiStore
	Scheduler scheduler.CronScheduler
	Publisher websockets.Publisher
}

// NewTransactionsHandler creates a new TransactionsHandler.
func NewTransactionsHandler(store storage.ApiStore, scheduler scheduler.CronScheduler, publisher websockets.Publisher) *TransactionsHandler {
	return &TransactionsHandler{Store: store, Scheduler: scheduler, Publisher: publisher}
}

// ScheduleTransaction handles the logic for scheduling a new transaction.
func (h *TransactionsHandler) ScheduleTransaction(w http.ResponseWriter, r *http.Request) {
	var newTx api.NewTransaction
	if err := json.NewDecoder(r.Body).Decode(&newTx); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	domainTx := mapping.ToDomainNewTransaction(&newTx)

	createdTx, err := h.Store.CreateTransaction(r.Context(), domainTx)
	if err != nil {
		if errors.Is(err, storage.ErrInsufficientFunds) {
			http.Error(w, "Insufficient funds", http.StatusUnprocessableEntity)
		} else {
			log.Printf("ERROR: Failed to create transaction in store: %v\n", err)
			http.Error(w, fmt.Sprintf("Failed to schedule transaction: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// If the database transaction was successful, enqueue it for processing.
	if h.Scheduler != nil {
		delay := time.Duration(0)
		if newTx.DelaySeconds != nil {
			delay = time.Duration(*newTx.DelaySeconds) * time.Second
		}
		if err := h.Scheduler.ScheduleTransaction(r.Context(), mapping.ToApiTransaction(createdTx), delay); err != nil {
			log.Printf("CRITICAL: transaction %s created but failed to enqueue: %v", createdTx.Id, err)
		}
	}

	// Get the latest wallet balance to update the sender via WebSocket.
	wallet, err := h.Store.GetWallet(r.Context(), createdTx.FromUserId)
	if err != nil {
		log.Printf("ERROR: failed to get wallet for websocket message: %v", err)
		// Do not fail the whole request if the websocket message fails.
	} else {
		msg := websockets.Message{
			Type: websockets.MessageTypeWalletUpdate,
			Payload: websockets.WalletUpdatePayload{
				UserID:        createdTx.FromUserId,
				TransactionID: createdTx.Id,
				Change:        -createdTx.Amount, // Negative because it's a deduction
				NewBalance:    wallet.Balance,
			},
		}
		if err := h.Publisher.Publish(r.Context(), msg); err != nil {
			log.Printf("ERROR: failed to publish websocket message: %v", err)
		}
	}

	// Map the domain model response back to the API model and respond.
	apiTx := mapping.ToApiTransaction(createdTx)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(apiTx); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", err), http.StatusInternalServerError)
	}
}

// GetTransactionById handles the logic for retrieving a transaction by its ID.
func (h *TransactionsHandler) GetTransactionById(w http.ResponseWriter, r *http.Request, transactionId string) {
	domainTx, err := h.Store.GetTransaction(r.Context(), transactionId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve transaction: %v", err), http.StatusNotFound)
		return
	}

	apiTx := mapping.ToApiTransaction(domainTx)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiTx); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", err), http.StatusInternalServerError)
	}
}

// CancelTransactionById handles the logic for cancelling a transaction.
func (h *TransactionsHandler) CancelTransactionById(w http.ResponseWriter, r *http.Request, transactionId string) {
	err := h.Store.CancelTransaction(r.Context(), transactionId)
	if err != nil {
		if errors.Is(err, storage.ErrTransactionNotCancellable) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to cancel transaction: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// NotifySettlement handles the internal callback after a transaction is settled.
func (h *TransactionsHandler) NotifySettlement(w http.ResponseWriter, r *http.Request, transactionId types.UUID) {
	// This handler is called internally, so we use a background context.
	ctx := context.Background()

	// 1. Get the settled transaction details.
	tx, err := h.Store.GetTransaction(ctx, transactionId.String())
	if err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	// 2. Get the recipient's latest wallet state.
	toWallet, err := h.Store.GetWallet(ctx, tx.ToUserId)
	if err != nil {
		log.Printf("ERROR: failed to get recipient's wallet for websocket message: %v", err)
		http.Error(w, "Failed to get recipient wallet", http.StatusInternalServerError)
		return
	}

	// 3. Publish a WebSocket message to the recipient.
	// The sender was already notified when the transaction was created.
	toMsg := websockets.Message{
		Type: websockets.MessageTypeWalletUpdate,
		Payload: websockets.WalletUpdatePayload{
			UserID:        tx.ToUserId,
			TransactionID: tx.Id,
			Change:        tx.Amount,
			NewBalance:    toWallet.Balance,
		},
	}
	if err := h.Publisher.Publish(ctx, toMsg); err != nil {
		log.Printf("ERROR: failed to publish websocket message to recipient: %v", err)
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListTransactionsByUserId handles the logic for retrieving all transactions for a user.
func (h *TransactionsHandler) ListTransactionsByUserId(w http.ResponseWriter, r *http.Request, userId string) {
	domainTxs, err := h.Store.ListTransactionsByUserID(r.Context(), userId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve transactions: %v", err), http.StatusInternalServerError)
		return
	}

	apiTxs := make([]*api.Transaction, len(domainTxs))
	for i, tx := range domainTxs {
		apiTxs[i] = mapping.ToApiTransaction(&tx)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiTxs); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", err), http.StatusInternalServerError)
	}
}
