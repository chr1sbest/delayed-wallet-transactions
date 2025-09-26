package ledger

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/chris/delayed-wallet-transactions/pkg/api"
	"github.com/chris/delayed-wallet-transactions/pkg/mapping"
	"github.com/chris/delayed-wallet-transactions/pkg/storage"
)

// LedgerHandler holds the dependencies for ledger-related handlers.
type LedgerHandler struct {
	Store storage.LedgerReader
}

// NewLedgerHandler creates a new LedgerHandler.
func NewLedgerHandler(store storage.LedgerReader) *LedgerHandler {
	return &LedgerHandler{Store: store}
}

func (h *LedgerHandler) ListLedgerEntries(w http.ResponseWriter, r *http.Request, params api.ListLedgerEntriesParams) {
	limit := int32(20)
	if params.Limit != nil {
		limit = int32(*params.Limit)
	}

	domainEntries, err := h.Store.ListLedgerEntries(r.Context(), limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve ledger entries: %v", err), http.StatusInternalServerError)
		return
	}

	apiEntries := make([]*api.LedgerEntry, len(domainEntries))
	for i, entry := range domainEntries {
		apiEntries[i] = mapping.ToApiLedgerEntry(&entry)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiEntries); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", err), http.StatusInternalServerError)
	}
}
