package handlers

import (
	"encoding/json"
	"fmt"
	"errors"
	"net/http"
	"strings"

	"github.com/chris/delayed-wallet-transactions/pkg/api"
	"github.com/chris/delayed-wallet-transactions/pkg/mapping"
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

	// Map the API request to our internal domain model.
	domainTx := mapping.ToDomainNewTransaction(&newTx)

	// Call the storage layer to create the transaction.
	updatedWallet, err := h.Store.CreateTransaction(r.Context(), domainTx)
	if err != nil {
		if errors.Is(err, storage.ErrInsufficientFunds) {
			http.Error(w, "Insufficient funds", http.StatusUnprocessableEntity)
		} else {
			http.Error(w, fmt.Sprintf("Failed to schedule transaction: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Map the domain model response back to the API model and respond.
	apiWallet := mapping.ToApiWallet(updatedWallet)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(apiWallet); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", err), http.StatusInternalServerError)
	}
}

// GetTransactionById handles the logic for retrieving a transaction by its ID.
func (h *ApiHandler) GetTransactionById(w http.ResponseWriter, r *http.Request, transactionId openapi_types.UUID) {
	// Call the storage layer to get the transaction.
	domainTx, err := h.Store.GetTransaction(r.Context(), transactionId)
	if err != nil {
		// A more robust implementation would check for a specific "not found" error.
		http.Error(w, fmt.Sprintf("Failed to retrieve transaction: %v", err), http.StatusNotFound)
		return
	}

	// Map the domain model to the API model and respond.
	apiTx := mapping.ToApiTransaction(domainTx)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiTx); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", err), http.StatusInternalServerError)
	}
}

// GetWalletByUserId handles the logic for retrieving a user's wallet.
// CreateWallet handles the logic for creating a new wallet.
func (h *ApiHandler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	// Decode the request body.
	var newWallet api.NewWallet
	if err := json.NewDecoder(r.Body).Decode(&newWallet); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Map the API request to our internal domain model.
	domainWallet := mapping.ToDomainNewWallet(&newWallet)

	// Call the storage layer to create the wallet.
	createdWallet, err := h.Store.CreateWallet(r.Context(), domainWallet)
	if err != nil {
		if strings.Contains(err.Error(), "wallet already exists") { // This is a simplistic check.
			http.Error(w, "Wallet for this user already exists", http.StatusConflict)
		} else {
			http.Error(w, fmt.Sprintf("Failed to create wallet: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Map the domain model response back to the API model and respond.
	apiWallet := mapping.ToApiWallet(createdWallet)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(apiWallet); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", err), http.StatusInternalServerError)
	}
}

// DeleteWallet handles the logic for deleting a user's wallet.
func (h *ApiHandler) DeleteWallet(w http.ResponseWriter, r *http.Request, userId string) {
	// Call the storage layer to delete the wallet.
	if err := h.Store.DeleteWallet(r.Context(), userId); err != nil {
		// A more robust implementation would check for a specific "not found" error.
		http.Error(w, fmt.Sprintf("Failed to delete wallet: %v", err), http.StatusNotFound)
		return
	}

	// Respond with a success status.
	w.WriteHeader(http.StatusNoContent)
}

// ListWallets handles the logic for retrieving all wallets.
func (h *ApiHandler) ListWallets(w http.ResponseWriter, r *http.Request) {
	// Call the storage layer to get all wallets.
	domainWallets, err := h.Store.ListWallets(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve wallets: %v", err), http.StatusInternalServerError)
		return
	}

	// Map the domain models to the API models.
	apiWallets := make([]*api.Wallet, len(domainWallets))
	for i, wallet := range domainWallets {
		apiWallets[i] = mapping.ToApiWallet(&wallet)
	}

	// Respond with the list of wallets.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiWallets); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", err), http.StatusInternalServerError)
	}
}

func (h *ApiHandler) GetWalletByUserId(w http.ResponseWriter, r *http.Request, userId string) {
	// Call the storage layer to get the wallet.
	domainWallet, err := h.Store.GetWallet(r.Context(), userId)
	if err != nil {
		// A more robust implementation would check for a specific "not found" error.
		http.Error(w, fmt.Sprintf("Failed to retrieve wallet: %v", err), http.StatusNotFound)
		return
	}

	// Map the domain model to the API model and respond.
	apiWallet := mapping.ToApiWallet(domainWallet)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiWallet); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", err), http.StatusInternalServerError)
	}
}
