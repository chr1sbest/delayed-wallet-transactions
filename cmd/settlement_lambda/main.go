package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/chris/delayed-wallet-transactions/pkg/models"
	dynamo_store "github.com/chris/delayed-wallet-transactions/pkg/storage/dynamodb"
)

var (
	store      *dynamo_store.Store
	apiBaseURL string
)

func init() {
	// Initialize dependencies once.
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	dbClient := dynamodb.NewFromConfig(cfg)
	store = dynamo_store.New(dbClient, os.Getenv("DYNAMODB_TRANSACTIONS_TABLE_NAME"), os.Getenv("DYNAMODB_WALLETS_TABLE_NAME"), os.Getenv("DYNAMODB_LEDGER_TABLE_NAME"), "")
	apiBaseURL = os.Getenv("API_BASE_URL")
}

// HandleRequest processes SQS messages and settles the transactions.
func HandleRequest(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, message := range sqsEvent.Records {
		log.Printf("Processing message %s: %s", message.MessageId, message.Body)

		var tx models.Transaction
		if err := json.Unmarshal([]byte(message.Body), &tx); err != nil {
			log.Printf("ERROR: failed to unmarshal transaction from SQS message %s: %v", message.MessageId, err)
			continue
		}

		if tx.FromUserId == "" || tx.ToUserId == "" {
			log.Printf("ERROR: transaction %s has empty FromUserId or ToUserId", tx.Id)
			continue
		}

		if err := store.SettleTransaction(ctx, &tx); err != nil {
			log.Printf("error settling transaction: %v", err)
			continue
		}

		if err := notifyApi(ctx, &tx); err != nil {
			log.Printf("error notifying API: %v", err)
			// Don't block the main flow if notification fails.
		}
	}
	return nil
}

func notifyApi(ctx context.Context, tx *models.Transaction) error {
	if apiBaseURL == "" {
		log.Println("API_BASE_URL not set, skipping notification.")
		return nil
	}

	url := fmt.Sprintf("%s/transactions/%s/notify-settlement", apiBaseURL, tx.Id)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to send notification to API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("API notification failed with status code: %d", resp.StatusCode)
	}

	log.Printf("Successfully notified API for transaction %s", tx.Id)
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
