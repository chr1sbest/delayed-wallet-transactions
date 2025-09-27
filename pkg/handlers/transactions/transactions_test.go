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
	"github.com/chris/delayed-wallet-transactions/pkg/websockets"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestScheduleTransaction_Success(t *testing.T) {
	t.Run("No Delay", func(t *testing.T) {
		// 1. Setup
		mockStorage := new(storage_mocks.ApiStore)
		mockScheduler := new(scheduler_mocks.CronScheduler)
		mockPublisher := new(websockets.NoOpPublisher)
		handler := NewTransactionsHandler(mockStorage, mockScheduler, mockPublisher)

		newTx := &api.NewTransaction{
			FromUserId: "user1",
			ToUserId:   "user2",
			Amount:     100,
		}

		createdTx := &models.Transaction{
			Id:         uuid.New().String(), // Mock ID
			FromUserId: newTx.FromUserId,
			ToUserId:   newTx.ToUserId,
			Amount:     newTx.Amount,
			Status:     models.RESERVED,
		}

		// 2. Mock expectations
		mockStorage.On("CreateTransaction", mock.Anything, mock.AnythingOfType("*models.Transaction")).Return(createdTx, nil)
		mockStorage.On("GetWallet", mock.Anything, "user1").Return(&models.Wallet{Balance: 1000}, nil).Maybe()
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
		mockStorage := new(storage_mocks.ApiStore)
		mockScheduler := new(scheduler_mocks.CronScheduler)
		mockPublisher := new(websockets.NoOpPublisher)
		handler := NewTransactionsHandler(mockStorage, mockScheduler, mockPublisher)

		delay := int32(60)
		newTx := &api.NewTransaction{
			FromUserId:   "user1",
			ToUserId:     "user2",
			Amount:       100,
			DelaySeconds: &delay,
		}

		createdTx := &models.Transaction{
			Id:         uuid.New().String(), // Mock ID
			FromUserId: newTx.FromUserId,
			ToUserId:   newTx.ToUserId,
			Amount:     newTx.Amount,
			Status:     models.RESERVED,
		}

		// 2. Mock expectations
		mockStorage.On("CreateTransaction", mock.Anything, mock.AnythingOfType("*models.Transaction")).Return(createdTx, nil)
		mockStorage.On("GetWallet", mock.Anything, "user1").Return(&models.Wallet{Balance: 1000}, nil).Maybe()
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
