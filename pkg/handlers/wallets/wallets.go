package wallets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/chris/delayed-wallet-transactions/pkg/api"
	"github.com/chris/delayed-wallet-transactions/pkg/mapping"
	"github.com/chris/delayed-wallet-transactions/pkg/storage"
)

// WalletsHandler holds the dependencies for wallet-related handlers.
type WalletsHandler struct {
	Store storage.WalletStore
}

// NewWalletsHandler creates a new WalletsHandler.
func NewWalletsHandler(store storage.WalletStore) *WalletsHandler {
	return &WalletsHandler{Store: store}
}

// CreateWallet handles the logic for creating a new wallet.
func (h *WalletsHandler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	var newWallet api.NewWallet
	if err := json.NewDecoder(r.Body).Decode(&newWallet); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	domainWallet := mapping.ToDomainNewWallet(&newWallet)
	domainWallet.CreatedAt = time.Now()

	createdWallet, err := h.Store.CreateWallet(r.Context(), domainWallet)
	if err != nil {
		if strings.Contains(err.Error(), "wallet already exists") { // This is a simplistic check.
			http.Error(w, "Wallet for this user already exists", http.StatusConflict)
		} else {
			http.Error(w, fmt.Sprintf("Failed to create wallet: %v", err), http.StatusInternalServerError)
		}
		return
	}

	apiWallet := mapping.ToApiWallet(createdWallet)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(apiWallet); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", err), http.StatusInternalServerError)
	}
}

// DeleteWallet handles the logic for deleting a user's wallet.
func (h *WalletsHandler) DeleteWallet(w http.ResponseWriter, r *http.Request, userId string) {
	if err := h.Store.DeleteWallet(r.Context(), userId); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete wallet: %v", err), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListWallets handles the logic for retrieving all wallets.

func (h *WalletsHandler) ListWallets(w http.ResponseWriter, r *http.Request) {
	domainWallets, err := h.Store.ListWallets(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve wallets: %v", err), http.StatusInternalServerError)
		return
	}

	// Sort wallets by CreatedAt in descending order.
	sort.Slice(domainWallets, func(i, j int) bool {
		return domainWallets[i].CreatedAt.After(domainWallets[j].CreatedAt)
	})

	apiWallets := make([]*api.Wallet, len(domainWallets))
	for i, wallet := range domainWallets {
		apiWallets[i] = mapping.ToApiWallet(&wallet)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiWallets); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", err), http.StatusInternalServerError)
	}
}

// GetWalletByUserId handles the logic for retrieving a user's wallet.
func (h *WalletsHandler) GetWalletByUserId(w http.ResponseWriter, r *http.Request, userId string) {
	domainWallet, err := h.Store.GetWallet(r.Context(), userId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve wallet: %v", err), http.StatusNotFound)
		return
	}

	apiWallet := mapping.ToApiWallet(domainWallet)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiWallet); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", err), http.StatusInternalServerError)
	}
}
