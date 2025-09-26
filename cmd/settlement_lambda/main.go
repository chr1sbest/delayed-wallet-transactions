package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/chris/delayed-wallet-transactions/pkg/models"
	"github.com/chris/delayed-wallet-transactions/pkg/storage"
	dydbstore "github.com/chris/delayed-wallet-transactions/pkg/storage/dynamodb"
	"github.com/joho/godotenv"
)

var store storage.Storage

func init() {
	// Load environment variables from .env file (useful for local testing).
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize dependencies once.
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	dbClient := dynamodb.NewFromConfig(cfg)
	transactionsTable := os.Getenv("DYNAMODB_TRANSACTIONS_TABLE_NAME")
	walletsTable := os.Getenv("DYNAMODB_WALLETS_TABLE_NAME")
	ledgerTable := os.Getenv("DYNAMODB_LEDGER_TABLE_NAME")

	if transactionsTable == "" || walletsTable == "" || ledgerTable == "" {
		log.Fatal("One or more DynamoDB table name environment variables are not set")
	}

	// The settlement lambda doesn't need a scheduler, so we pass nil.
	store = dydbstore.New(dbClient, nil, transactionsTable, walletsTable, ledgerTable)
}

// HandleRequest processes SQS messages and settles the transactions.
func HandleRequest(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, message := range sqsEvent.Records {
		log.Printf("Processing message %s", message.MessageId)

		var tx models.Transaction
		if err := json.Unmarshal([]byte(message.Body), &tx); err != nil {
			log.Printf("ERROR: failed to unmarshal transaction from SQS message %s: %v", message.MessageId, err)
			// Returning an error will cause SQS to retry the message, which is appropriate here.
			return err
		}

		log.Printf("Attempting to settle transaction %s", tx.Id)

		if err := store.SettleTransaction(ctx, &tx); err != nil {
			log.Printf("ERROR: failed to settle transaction %s: %v", tx.Id, err)
			// In a production system, persistent failures would be sent to a DLQ.
			return err
		}

		log.Printf("Successfully settled transaction %s", tx.Id)
	}

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
