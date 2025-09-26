package scheduler

import (
	"context"

	"github.com/chris/delayed-wallet-transactions/pkg/api"
)

// Scheduler defines the interface for a component that schedules a transaction for later processing.
type Scheduler interface {
	// ScheduleTransaction enqueues a transaction for asynchronous processing.
	ScheduleTransaction(ctx context.Context, tx *api.Transaction) error
}
