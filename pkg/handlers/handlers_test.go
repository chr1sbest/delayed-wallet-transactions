package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chris/delayed-wallet-transactions/pkg/storage"

	"github.com/chris/delayed-wallet-transactions/pkg/api"
	"github.com/chris/delayed-wallet-transactions/pkg/models"
	"github.com/chris/delayed-wallet-transactions/pkg/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestScheduleTransaction(t *testing.T) {
	// Common test data
	newApiTx := api.NewTransaction{
		FromUserId:  "user-a",
		ToUserId:    "user-b",
		Amount:      100,
		ScheduledAt: time.Now().Add(10 * time.Minute),
	}
	// This represents the wallet object that comes back from the database
	expectedWallet := &models.Wallet{
		UserId:   newApiTx.FromUserId,
		Balance:  900,
		Reserved: 100,
		Version:  2,
	}

	t.Run("Success", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.Storage)
		mockStorage.On("CreateTransaction", mock.Anything, mock.Anything).Return(expectedWallet, nil)

		h := NewApiHandler(mockStorage)

		body, _ := json.Marshal(newApiTx)
		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		// Act
		h.ScheduleTransaction(rr, req)

		// Assert
		assert.Equal(t, http.StatusCreated, rr.Code)

		var returnedWallet api.Wallet
		json.Unmarshal(rr.Body.Bytes(), &returnedWallet)
		assert.Equal(t, expectedWallet.UserId, returnedWallet.UserId)
		assert.Equal(t, expectedWallet.Balance, returnedWallet.Balance)
		assert.Equal(t, expectedWallet.Reserved, returnedWallet.Reserved)

		mockStorage.AssertExpectations(t)
	})

	t.Run("Insufficient Funds", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.Storage)
		mockStorage.On("CreateTransaction", mock.Anything, mock.Anything).Return(nil, storage.ErrInsufficientFunds)

		h := NewApiHandler(mockStorage)

		body, _ := json.Marshal(newApiTx)
		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		// Act
		h.ScheduleTransaction(rr, req)

		// Assert
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		assert.Contains(t, rr.Body.String(), "Insufficient funds")
		mockStorage.AssertExpectations(t)
	})

	t.Run("Bad Request - Invalid JSON", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.Storage)
		h := NewApiHandler(mockStorage)

		req := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader("not-json"))
		rr := httptest.NewRecorder()

		// Act
		h.ScheduleTransaction(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		// We don't assert mock expectations because the storage layer should not be called.
	})

	t.Run("Generic Storage Failure", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.Storage)
		mockStorage.On("CreateTransaction", mock.Anything, mock.Anything).Return(nil, errors.New("something went wrong"))

		h := NewApiHandler(mockStorage)

		body, _ := json.Marshal(newApiTx)
		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		// Act
		h.ScheduleTransaction(rr, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		mockStorage.AssertExpectations(t)
	})
}
