package storage

import "errors"

// ErrInsufficientFunds is returned when a wallet has an insufficient balance for a transaction.
var ErrInsufficientFunds = errors.New("insufficient funds")
var ErrTransactionAlreadyProcessing = errors.New("transaction is already being processed")

// ErrTransactionNotCancellable is returned when a transaction cannot be cancelled, e.g., because it's already completed or cancelled.
var ErrTransactionNotCancellable = errors.New("transaction not in a cancellable state")

// ErrTransactionNotProcessable is returned when a transaction is not in a state that allows processing (e.g., it's already cancelled).
var ErrTransactionNotProcessable = errors.New("transaction not in a processable state")
