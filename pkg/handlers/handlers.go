package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/chris/delayed-wallet-transactions/pkg/api"
	"github.com/chris/delayed-wallet-transactions/pkg/storage"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// ApiHandler implements the generated server interface.
// It holds our application's dependencies, including the storage layer.
type ApiHandler struct {
	Store storage.Storage
}

// NewApiHandler creates a new ApiHandler with a storage dependency.
func NewApiHandler(store storage.Storage) *ApiHandler {
	return &ApiHandler{Store: store}
}

// Make sure we conform to the interface
var _ api.ServerInterface = (*ApiHandler)(nil)

// ScheduleTransaction handles the logic for scheduling a new transaction.
func (h *ApiHandler) ScheduleTransaction(w http.ResponseWriter, r *http.Request) {
	// Decode the request body.
	var newTx api.NewTransaction
	if err := json.NewDecoder(r.Body).Decode(&newTx); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Call the storage layer to create the transaction.
	tx, err := h.Store.CreateTransaction(r.Context(), newTx)
	if err != nil {
		// Check for specific, user-facing errors like conditional check failures.
		if strings.Contains(err.Error(), "conditional check failed") {
			http.Error(w, fmt.Sprintf("Failed to schedule transaction: %v", err), http.StatusBadRequest)
		} else {
			http.Error(w, fmt.Sprintf("Failed to schedule transaction: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Respond with the created transaction.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(tx); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", err), http.StatusInternalServerError)
	}
}

// GetTransactionById handles the logic for retrieving a transaction by its ID.
func (h *ApiHandler) GetTransactionById(w http.ResponseWriter, r *http.Request, transactionId openapi_types.UUID) {
	// Call the storage layer to get the transaction.
	tx, err := h.Store.GetTransaction(r.Context(), transactionId)
	if err != nil {
		// A more robust implementation would check for a specific "not found" error.
		http.Error(w, fmt.Sprintf("Failed to retrieve transaction: %v", err), http.StatusNotFound)
		return
	}

	// Respond with the transaction details.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tx); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", err), http.StatusInternalServerError)
	}
}

// GetWalletByUserId handles the logic for retrieving a user's wallet.
func (h *ApiHandler) GetWalletByUserId(w http.ResponseWriter, r *http.Request, userId string) {
	// Call the storage layer to get the wallet.
	wallet, err := h.Store.GetWallet(r.Context(), userId)
	if err != nil {
		// A more robust implementation would check for a specific "not found" error.
		http.Error(w, fmt.Sprintf("Failed to retrieve wallet: %v", err), http.StatusNotFound)
		return
	}

	// Respond with the wallet details.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(wallet); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", err), http.StatusInternalServerError)
	}
}
