package ledger_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chris/delayed-wallet-transactions/pkg/api"
	"github.com/chris/delayed-wallet-transactions/pkg/handlers/ledger"
	"github.com/chris/delayed-wallet-transactions/pkg/models"
	"github.com/chris/delayed-wallet-transactions/pkg/storage/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListLedgerEntries(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.Storage)
		expectedEntries := []models.LedgerEntry{
			{EntryID: uuid.New().String(), Timestamp: time.Now()},
			{EntryID: uuid.New().String(), Timestamp: time.Now().Add(-1 * time.Minute)},
		}
		mockStorage.On("ListLedgerEntries", mock.Anything, int32(20)).Return(expectedEntries, nil)

		h := ledger.NewLedgerHandler(mockStorage)

		req := httptest.NewRequest(http.MethodGet, "/ledger", nil)
		rr := httptest.NewRecorder()

		// Act
		h.ListLedgerEntries(rr, req, api.ListLedgerEntriesParams{})

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		var returnedEntries []api.LedgerEntry
		json.Unmarshal(rr.Body.Bytes(), &returnedEntries)
		assert.Len(t, returnedEntries, 2)
		assert.Equal(t, expectedEntries[0].EntryID, returnedEntries[0].EntryId)

		mockStorage.AssertExpectations(t)
	})
}
