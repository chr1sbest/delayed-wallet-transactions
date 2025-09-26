package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/chris/delayed-wallet-transactions/pkg/mapping"
	"github.com/chris/delayed-wallet-transactions/pkg/scheduler"
	"github.com/chris/delayed-wallet-transactions/pkg/storage"
	dydbstore "github.com/chris/delayed-wallet-transactions/pkg/storage/dynamodb"
	"github.com/joho/godotenv"
)

var store storage.Storage
var sqsScheduler scheduler.CronScheduler

const stuckTransactionThreshold = 20 * time.Minute

func init() {
	// Load environment variables for local testing.
	godotenv.Load()

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	dbClient := dynamodb.NewFromConfig(cfg)
	sqsClient := sqs.NewFromConfig(cfg)

	// Initialize dependencies.
	sqsQueueURL := os.Getenv("SQS_QUEUE_URL")
	if sqsQueueURL == "" {
		log.Fatal("SQS_QUEUE_URL environment variable not set")
	}
	sqsScheduler = scheduler.NewSQSScheduler(sqsClient, sqsQueueURL)

	transactionsTable := os.Getenv("DYNAMODB_TRANSACTIONS_TABLE_NAME")
	walletsTable := os.Getenv("DYNAMODB_WALLETS_TABLE_NAME")
	ledgerTable := os.Getenv("DYNAMODB_LEDGER_TABLE_NAME")

	store = dydbstore.New(dbClient, transactionsTable, walletsTable, ledgerTable)
}

// HandleRequest is triggered by an EventBridge Schedule.
func HandleRequest(ctx context.Context) error {
	log.Println("Starting reconciliation process for stuck transactions...")

	stuckTxs, err := store.GetStuckTransactions(ctx, stuckTransactionThreshold)
	if err != nil {
		log.Printf("ERROR: failed to get stuck transactions: %v", err)
		return err
	}

	if len(stuckTxs) == 0 {
		log.Println("No stuck transactions found.")
		return nil
	}

	log.Printf("Found %d stuck transactions. Re-enqueuing them...", len(stuckTxs))

	for _, tx := range stuckTxs {
		apiTx := mapping.ToApiTransaction(&tx)
		if err := sqsScheduler.ScheduleTransaction(ctx, apiTx, 0); err != nil {
			log.Printf("ERROR: failed to re-enqueue transaction %s: %v", tx.Id, err)
			// Continue to the next transaction, don't let one failure stop the whole batch.
			continue
		}
		log.Printf("Successfully re-enqueued transaction %s", tx.Id)
	}

	log.Println("Reconciliation process finished.")
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
