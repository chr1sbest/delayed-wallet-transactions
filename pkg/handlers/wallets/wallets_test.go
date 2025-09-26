package wallets_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chris/delayed-wallet-transactions/pkg/api"
	"github.com/chris/delayed-wallet-transactions/pkg/handlers/wallets"
	"github.com/chris/delayed-wallet-transactions/pkg/models"
	"github.com/chris/delayed-wallet-transactions/pkg/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateWallet(t *testing.T) {
	newApiWallet := api.NewWallet{UserId: "user-c"}
	expectedWallet := &models.Wallet{UserId: "user-c", Balance: 0, Reserved: 0, Version: 1}

	t.Run("Success", func(t *testing.T) {
		mockStorage := new(mocks.Storage)
		mockStorage.On("CreateWallet", mock.Anything, mock.Anything).Return(expectedWallet, nil)

		h := wallets.NewWalletsHandler(mockStorage)

		body, _ := json.Marshal(newApiWallet)
		req := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		h.CreateWallet(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		mockStorage.AssertExpectations(t)
	})
}

func TestDeleteWallet(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockStorage := new(mocks.Storage)
		mockStorage.On("DeleteWallet", mock.Anything, "user-c").Return(nil)

		h := wallets.NewWalletsHandler(mockStorage)

		req := httptest.NewRequest(http.MethodDelete, "/wallets/user-c", nil)
		rr := httptest.NewRecorder()

		h.DeleteWallet(rr, req, "user-c")

		assert.Equal(t, http.StatusNoContent, rr.Code)
		mockStorage.AssertExpectations(t)
	})
}

func TestListWallets(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.Storage)
		mockStorage.On("ListWallets", mock.Anything).Return([]models.Wallet{}, nil)

		h := wallets.NewWalletsHandler(mockStorage)

		req := httptest.NewRequest(http.MethodGet, "/wallets", nil)
		rr := httptest.NewRecorder()

		// Act
		h.ListWallets(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		mockStorage.AssertExpectations(t)
	})
}

func TestGetWalletByUserId(t *testing.T) {
	expectedWallet := &models.Wallet{UserId: "user-c", Balance: 100, Reserved: 50, Version: 2}

	t.Run("Success", func(t *testing.T) {
		mockStorage := new(mocks.Storage)
		mockStorage.On("GetWallet", mock.Anything, "user-c").Return(expectedWallet, nil)

		h := wallets.NewWalletsHandler(mockStorage)

		req := httptest.NewRequest(http.MethodGet, "/wallets/user-c", nil)
		rr := httptest.NewRecorder()

		h.GetWalletByUserId(rr, req, "user-c")

		assert.Equal(t, http.StatusOK, rr.Code)
		mockStorage.AssertExpectations(t)
	})
}
