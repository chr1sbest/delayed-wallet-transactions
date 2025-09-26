package transactions

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chris/delayed-wallet-transactions/pkg/api"
	"github.com/chris/delayed-wallet-transactions/pkg/models"
	scheduler_mocks "github.com/chris/delayed-wallet-transactions/pkg/scheduler/mocks"
	storage_mocks "github.com/chris/delayed-wallet-transactions/pkg/storage/mocks"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestScheduleTransaction_Success(t *testing.T) {
	t.Run("No Delay", func(t *testing.T) {
		// 1. Setup
		mockStorage := new(storage_mocks.Storage)
		mockScheduler := new(scheduler_mocks.CronScheduler)
		handler := NewTransactionsHandler(mockStorage, mockScheduler)

		newTx := &api.NewTransaction{
			FromUserId: "user1",
			ToUserId:   "user2",
			Amount:     100,
		}

		createdTx := &models.Transaction{
			Id:         openapi_types.UUID{0x1}, // Mock ID
			FromUserId: newTx.FromUserId,
			ToUserId:   newTx.ToUserId,
			Amount:     newTx.Amount,
			Status:     models.RESERVED,
		}

		// 2. Mock expectations
		mockStorage.On("CreateTransaction", mock.Anything, mock.AnythingOfType("*models.Transaction")).Return(createdTx, nil)
		mockScheduler.On("ScheduleTransaction", mock.Anything, mock.AnythingOfType("*api.Transaction"), time.Duration(0)).Return(nil)

		// 3. Execute
		body, _ := json.Marshal(newTx)
		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		handler.ScheduleTransaction(rr, req)

		// 4. Assert
		assert.Equal(t, http.StatusCreated, rr.Code)
		mockStorage.AssertExpectations(t)
		mockScheduler.AssertExpectations(t)
	})

	t.Run("With Delay", func(t *testing.T) {
		// 1. Setup
		mockStorage := new(storage_mocks.Storage)
		mockScheduler := new(scheduler_mocks.CronScheduler)
		handler := NewTransactionsHandler(mockStorage, mockScheduler)

		delay := int32(60)
		newTx := &api.NewTransaction{
			FromUserId:   "user1",
			ToUserId:     "user2",
			Amount:       100,
			DelaySeconds: &delay,
		}

		createdTx := &models.Transaction{
			Id:         openapi_types.UUID{0x1}, // Mock ID
			FromUserId: newTx.FromUserId,
			ToUserId:   newTx.ToUserId,
			Amount:     newTx.Amount,
			Status:     models.RESERVED,
		}

		// 2. Mock expectations
		mockStorage.On("CreateTransaction", mock.Anything, mock.AnythingOfType("*models.Transaction")).Return(createdTx, nil)
		mockScheduler.On("ScheduleTransaction", mock.Anything, mock.AnythingOfType("*api.Transaction"), time.Duration(delay)*time.Second).Return(nil)

		// 3. Execute
		body, _ := json.Marshal(newTx)
		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		handler.ScheduleTransaction(rr, req)

		// 4. Assert
		assert.Equal(t, http.StatusCreated, rr.Code)
		mockStorage.AssertExpectations(t)
		mockScheduler.AssertExpectations(t)
	})
}
