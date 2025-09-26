package scheduler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/chris/delayed-wallet-transactions/pkg/api"
)

// SQSScheduler implements the Scheduler interface using AWS SQS.
type SQSScheduler struct {
	Client   *sqs.Client
	QueueURL string
}

// NewSQSScheduler creates a new SQSScheduler.
func NewSQSScheduler(client *sqs.Client, queueURL string) *SQSScheduler {
	return &SQSScheduler{
		Client:   client,
		QueueURL: queueURL,
	}
}

// Make sure we conform to the interface
var _ Scheduler = (*SQSScheduler)(nil)

// ScheduleTransaction sends the transaction to an SQS queue for later processing.
func (s *SQSScheduler) ScheduleTransaction(ctx context.Context, tx *api.Transaction) error {
	// Marshal the transaction to JSON.
	body, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction for SQS: %w", err)
	}

	// Send the message to SQS.
	_, err = s.Client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.QueueURL),
		MessageBody: aws.String(string(body)),
	})

	if err != nil {
		return fmt.Errorf("failed to send message to SQS: %w", err)
	}

	return nil
}
