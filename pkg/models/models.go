package models

import (
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// TransactionStatus defines the possible statuses of a transaction.
type TransactionStatus string

const (
	RESERVED  TransactionStatus = "RESERVED"
	APPROVED  TransactionStatus = "APPROVED"
	COMPLETED TransactionStatus = "COMPLETED"
	REJECTED  TransactionStatus = "REJECTED"
)

// Transaction represents the internal domain model for a transaction.
// It includes dynamodbav tags for marshalling.
type Transaction struct {
	Id          openapi_types.UUID `dynamodbav:"id"`
	FromUserId  string             `dynamodbav:"from_user_id"`
	ToUserId    string             `dynamodbav:"to_user_id"`
	Amount      float64            `dynamodbav:"amount"`
	Currency    string             `dynamodbav:"currency"`
	Status      TransactionStatus  `dynamodbav:"status"`
	ScheduledAt time.Time          `dynamodbav:"scheduled_at"`
	CreatedAt   time.Time          `dynamodbav:"created_at"`
	UpdatedAt   time.Time          `dynamodbav:"updated_at"`
}

// Wallet represents the internal domain model for a user's wallet.
type Wallet struct {
	UserId   string  `dynamodbav:"user_id"`
	Balance  float64 `dynamodbav:"balance"`
	Reserved float64 `dynamodbav:"reserved"`
	Currency string  `dynamodbav:"currency"`
	Version  int64   `dynamodbav:"version"`
}

// LedgerEntry represents a single entry in the double-entry ledger.
type LedgerEntry struct {
	TransactionID string    `dynamodbav:"transaction_id"`
	EntryID       string    `dynamodbav:"entry_id"`
	AccountID     string    `dynamodbav:"account_id"`
	Debit         float64   `dynamodbav:"debit,omitempty"`
	Credit        float64   `dynamodbav:"credit,omitempty"`
	Description   string    `dynamodbav:"description"`
	Timestamp     time.Time `dynamodbav:"timestamp"`
}
