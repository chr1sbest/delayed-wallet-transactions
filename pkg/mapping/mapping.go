package mapping

import (
	"github.com/chris/delayed-wallet-transactions/pkg/api"
	"github.com/chris/delayed-wallet-transactions/pkg/models"
)

// ToApiTransaction converts a domain Transaction model to an API Transaction model.
func ToApiTransaction(tx *models.Transaction) *api.Transaction {
	return &api.Transaction{
		Id:          tx.Id,
		FromUserId:  tx.FromUserId,
		ToUserId:    tx.ToUserId,
		Amount:      tx.Amount,
		Currency:    tx.Currency,
		Status:      api.TransactionStatus(tx.Status),
		ScheduledAt: tx.ScheduledAt,
		CreatedAt:   tx.CreatedAt,
		UpdatedAt:   tx.UpdatedAt,
	}
}

// ToDomainNewTransaction converts an API NewTransaction model to a domain Transaction model.
// Note: This is a simplified mapping and does not create the full Transaction object.
func ToDomainNewTransaction(newTx *api.NewTransaction) *models.Transaction {
	return &models.Transaction{
		FromUserId:  newTx.FromUserId,
		ToUserId:    newTx.ToUserId,
		Amount:      newTx.Amount,
		Currency:    newTx.Currency,
		ScheduledAt: newTx.ScheduledAt,
	}
}

// ToApiWallet converts a domain Wallet model to an API Wallet model.
func ToApiWallet(wallet *models.Wallet) *api.Wallet {
	return &api.Wallet{
		UserId:   wallet.UserId,
		Balance:  wallet.Balance,
		Reserved: wallet.Reserved,
		Currency: wallet.Currency,
		Version:  wallet.Version,
	}
}

// ToDomainTransaction converts an API Transaction model to a domain Transaction model.
func ToDomainTransaction(tx *api.Transaction) *models.Transaction {
	return &models.Transaction{
		Id:          tx.Id,
		FromUserId:  tx.FromUserId,
		ToUserId:    tx.ToUserId,
		Amount:      tx.Amount,
		Currency:    tx.Currency,
		Status:      models.TransactionStatus(tx.Status),
		ScheduledAt: tx.ScheduledAt,
		CreatedAt:   tx.CreatedAt,
		UpdatedAt:   tx.UpdatedAt,
	}
}
