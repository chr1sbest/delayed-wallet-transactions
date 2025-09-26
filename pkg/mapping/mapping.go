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
		ScheduledAt: newTx.ScheduledAt,
	}
}

// ToApiWallet converts a domain Wallet model to an API Wallet model.
func ToApiWallet(wallet *models.Wallet) *api.Wallet {
	return &api.Wallet{
		UserId:   wallet.UserId,
		Balance:  wallet.Balance,
		Reserved: wallet.Reserved,
		Version:  wallet.Version,
	}
}

// ToDomainNewWallet converts an API NewWallet model to a domain Wallet model.
func ToDomainNewWallet(newWallet *api.NewWallet) *models.Wallet {
	return &models.Wallet{
		UserId:  newWallet.UserId,
		Balance: 1000, // Seed new wallets with 1000 units.
		Version: 1,
	}
}

// ToDomainTransaction converts an API Transaction model to a domain Transaction model.
func ToApiLedgerEntry(entry *models.LedgerEntry) *api.LedgerEntry {
	return &api.LedgerEntry{
		TransactionId: entry.TransactionID,
		EntryId:       entry.EntryID,
		AccountId:     entry.AccountID,
		Debit:         &entry.Debit,
		Credit:        &entry.Credit,
		Description:   entry.Description,
		Timestamp:     entry.Timestamp,
	}
}

func ToDomainTransaction(tx *api.Transaction) *models.Transaction {
	return &models.Transaction{
		Id:          tx.Id,
		FromUserId:  tx.FromUserId,
		ToUserId:    tx.ToUserId,
		Amount:      tx.Amount,
		Status:      models.TransactionStatus(tx.Status),
		ScheduledAt: tx.ScheduledAt,
		CreatedAt:   tx.CreatedAt,
		UpdatedAt:   tx.UpdatedAt,
	}
}
