package models

import (
	"time"
)

// TransactionStatus defines the possible states of a transaction.
type TransactionStatus string

const (
	RESERVED  TransactionStatus = "RESERVED"
	WORKING   TransactionStatus = "WORKING"
	APPROVED  TransactionStatus = "APPROVED"
	COMPLETED TransactionStatus = "COMPLETED"
	CANCELLED TransactionStatus = "CANCELLED"
)

// Transaction represents the internal domain model for a transaction.
// It includes dynamodbav and json tags for marshalling.
type Transaction struct {
	Id           string            `json:"id" dynamodbav:"id"`
	FromUserId   string            `json:"from_user_id" dynamodbav:"from_user_id"`
	ToUserId     string            `json:"to_user_id" dynamodbav:"to_user_id"`
	Amount       int64             `json:"amount" dynamodbav:"amount"`
	DelaySeconds *int32            `json:"delay_seconds,omitempty" dynamodbav:"delay_seconds,omitempty"`
	Status       TransactionStatus `json:"status" dynamodbav:"status"`
	CreatedAt    time.Time         `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at" dynamodbav:"updated_at"`
	TTL          int64             `json:"ttl,omitempty" dynamodbav:"ttl,omitempty"`
}

// Wallet represents the internal domain model for a user's wallet.
type Wallet struct {
	UserId    string    `json:"user_id" dynamodbav:"user_id"`
	Name      string    `json:"name" dynamodbav:"name"`
	Balance   int64     `json:"balance" dynamodbav:"balance"`
	Reserved  int64     `json:"reserved" dynamodbav:"reserved"`
	Version   int64     `json:"version" dynamodbav:"version"`
	CreatedAt time.Time `json:"created_at" dynamodbav:"created_at"`
	TTL       int64     `json:"ttl,omitempty" dynamodbav:"ttl,omitempty"`
}

// LedgerEntry represents a single entry in the double-entry ledger.
type LedgerEntry struct {
	EntryID       string    `json:"entry_id" dynamodbav:"entry_id"`
	TransactionID string    `json:"transaction_id" dynamodbav:"transaction_id"`
	AccountID     string    `json:"account_id" dynamodbav:"account_id"`
	Debit         int64     `json:"debit,omitempty" dynamodbav:"debit,omitempty"`
	Credit        int64     `json:"credit,omitempty" dynamodbav:"credit,omitempty"`
	Description   string    `json:"description" dynamodbav:"description"`
	Timestamp     time.Time `json:"timestamp" dynamodbav:"timestamp"`
	GSI1PK        string    `json:"gsi1pk" dynamodbav:"gsi1pk"`
}
