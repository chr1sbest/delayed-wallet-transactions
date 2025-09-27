package models

import (
	"time"
)

// TransactionStatus defines the possible states of a transaction.
type TransactionStatus string

const (
	RESERVED  TransactionStatus = "RESERVED"
	APPROVED  TransactionStatus = "APPROVED"
	COMPLETED TransactionStatus = "COMPLETED"
	CANCELLED TransactionStatus = "CANCELLED"
)

// Transaction represents the internal domain model for a transaction.
// It includes dynamodbav tags for marshalling.
type Transaction struct {
	Id          string             `dynamodbav:"id"`
	FromUserId  string             `dynamodbav:"from_user_id"`
	ToUserId    string             `dynamodbav:"to_user_id"`
	Amount      int64              `dynamodbav:"amount"`
	DelaySeconds *int32             `dynamodbav:"delay_seconds,omitempty"`
	Status      TransactionStatus  `dynamodbav:"status"`
	CreatedAt   time.Time          `dynamodbav:"created_at"`
	UpdatedAt   time.Time          `dynamodbav:"updated_at"`
	TTL         int64              `dynamodbav:"ttl,omitempty"`
}

// Wallet represents the internal domain model for a user's wallet.
type Wallet struct {
	UserId    string    `json:"user_id" dynamodbav:"user_id"`
	Name      string    `json:"name" dynamodbav:"name"`
	Balance   int64     `json:"balance" dynamodbav:"balance"`
	Reserved  int64     `json:"reserved" dynamodbav:"reserved"`
	Version   int64     `json:"version" dynamodbav:"version"`
	CreatedAt time.Time `json:"created_at" dynamodbav:"created_at"`
	TTL       int64     `dynamodbav:"ttl,omitempty"`
}

// LedgerEntry represents a single entry in the double-entry ledger.
type LedgerEntry struct {
	EntryID       string    `dynamodbav:"entry_id"`
	TransactionID string    `dynamodbav:"transaction_id"`
	AccountID     string    `dynamodbav:"account_id"`
	Debit         int64     `dynamodbav:"debit,omitempty"`
	Credit        int64     `dynamodbav:"credit,omitempty"`
	Description   string    `dynamodbav:"description"`
	Timestamp     time.Time `dynamodbav:"timestamp"`
	GSI1PK        string    `dynamodbav:"gsi1pk"`
}
