package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/chris/delayed-wallet-transactions/pkg/api"
	"github.com/chris/delayed-wallet-transactions/pkg/handlers"
	"github.com/chris/delayed-wallet-transactions/pkg/scheduler"
	"github.com/chris/delayed-wallet-transactions/pkg/storage"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// AWS Session
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

	// SQS Client and Scheduler
	sqsClient := sqs.NewFromConfig(cfg)
	sqsQueueURL := os.Getenv("SQS_QUEUE_URL")
	if sqsQueueURL == "" {
		log.Fatal("SQS_QUEUE_URL environment variable not set")
	}
	sqsScheduler := scheduler.NewSQSScheduler(sqsClient, sqsQueueURL)

	// Create our storage implementation
	store := storage.NewDynamoDBStore(dbClient, sqsScheduler, transactionsTable, walletsTable, ledgerTable)

	// Create our handler
	handler := handlers.NewApiHandler(store)

	// Create a new Chi router
	router := chi.NewRouter()

	// Use the generated function to mount our handler on the router
	api.HandlerFromMux(handler, router)

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}

	log.Printf("Starting server on port %s", port)

	// Start the server
	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
