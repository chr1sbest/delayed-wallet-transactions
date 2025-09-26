package scheduler

import (
	"context"

	"github.com/chris/delayed-wallet-transactions/pkg/api"
)

import "time"

// CronScheduler defines the interface for a component that schedules a transaction for later processing.
type CronScheduler interface {
	// ScheduleTransaction enqueues a transaction for asynchronous processing with an optional delay.
	ScheduleTransaction(ctx context.Context, tx *api.Transaction, delay time.Duration) error
}
