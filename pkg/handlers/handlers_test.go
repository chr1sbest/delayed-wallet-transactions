package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chris/delayed-wallet-transactions/pkg/api"
	"github.com/chris/delayed-wallet-transactions/pkg/models"
	"github.com/chris/delayed-wallet-transactions/pkg/storage/mocks"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestScheduleTransaction(t *testing.T) {
	// Common test data
	newTx := api.NewTransaction{
		FromUserId:  "user-a",
		ToUserId:    "user-b",
		Amount:      100.0,
		Currency:    "USD",
		ScheduledAt: time.Now().Add(10 * time.Minute),
	}

	txID := openapi_types.UUID(uuid.New())
	expectedTx := &api.Transaction{
		Id:          txID,
		FromUserId:  newTx.FromUserId,
		ToUserId:    newTx.ToUserId,
		Amount:      newTx.Amount,
		Currency:    newTx.Currency,
		Status:      api.TransactionStatus(models.RESERVED),
		ScheduledAt: newTx.ScheduledAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	t.Run("Success", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.Storage)
		// The mock now expects a models.Transaction object.
		mockStorage.On("CreateTransaction",
			mock.AnythingOfType("context.Context"),
			mock.AnythingOfType("*models.Transaction"),
		).Return(mapping.ToDomainTransaction(expectedTx), nil)

		h := NewApiHandler(mockStorage)

		body, _ := json.Marshal(newTx)
		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		// Act
		h.ScheduleTransaction(rr, req)

		// Assert
		assert.Equal(t, http.StatusCreated, rr.Code)

		var returnedTx api.Transaction
		json.Unmarshal(rr.Body.Bytes(), &returnedTx)
		assert.Equal(t, expectedTx.Id, returnedTx.Id)
		assert.Equal(t, expectedTx.Amount, returnedTx.Amount)

		mockStorage.AssertExpectations(t)
	})

	t.Run("Storage Failure - Conditional Check Failed", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.Storage)
		mockStorage.On("CreateTransaction",
			mock.AnythingOfType("context.Context"),
			mock.AnythingOfType("*models.Transaction"),
		).Return(nil, errors.New("conditional check failed"))

		h := NewApiHandler(mockStorage)

		body, _ := json.Marshal(newTx)
		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		// Act
		h.ScheduleTransaction(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "conditional check failed")
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
}
