package transactions

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/chris/delayed-wallet-transactions/pkg/api"
	"github.com/chris/delayed-wallet-transactions/pkg/mapping"
	"github.com/chris/delayed-wallet-transactions/pkg/storage"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// TransactionsHandler holds the dependencies for transaction-related handlers.
type TransactionsHandler struct {
	Store storage.Storage
}

// NewTransactionsHandler creates a new TransactionsHandler.
func NewTransactionsHandler(store storage.Storage) *TransactionsHandler {
	return &TransactionsHandler{Store: store}
}

// ScheduleTransaction handles the logic for scheduling a new transaction.
func (h *TransactionsHandler) ScheduleTransaction(w http.ResponseWriter, r *http.Request) {
	var newTx api.NewTransaction
	if err := json.NewDecoder(r.Body).Decode(&newTx); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	domainTx := mapping.ToDomainNewTransaction(&newTx)

	updatedWallet, err := h.Store.CreateTransaction(r.Context(), domainTx)
	if err != nil {
		if errors.Is(err, storage.ErrInsufficientFunds) {
			http.Error(w, "Insufficient funds", http.StatusUnprocessableEntity)
		} else {
			http.Error(w, fmt.Sprintf("Failed to schedule transaction: %v", err), http.StatusInternalServerError)
		}
		return
	}

	apiWallet := mapping.ToApiWallet(updatedWallet)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(apiWallet); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", err), http.StatusInternalServerError)
	}
}

// GetTransactionById handles the logic for retrieving a transaction by its ID.
func (h *TransactionsHandler) GetTransactionById(w http.ResponseWriter, r *http.Request, transactionId openapi_types.UUID) {
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
func (h *TransactionsHandler) CancelTransactionById(w http.ResponseWriter, r *http.Request, transactionId openapi_types.UUID) {
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
