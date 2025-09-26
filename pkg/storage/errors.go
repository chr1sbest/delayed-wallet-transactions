package storage

import "errors"

// ErrInsufficientFunds is returned when a wallet has an insufficient balance for a transaction.
var ErrInsufficientFunds = errors.New("insufficient funds")

// ErrTransactionNotCancellable is returned when a transaction cannot be cancelled, e.g., because it's already completed or cancelled.
var ErrTransactionNotCancellable = errors.New("transaction not in a cancellable state")
