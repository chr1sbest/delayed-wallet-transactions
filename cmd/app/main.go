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
	dydbstore "github.com/chris/delayed-wallet-transactions/pkg/storage/dynamodb"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/swaggest/swgui/v5emb"
	"github.com/aws/aws-lambda-go/lambda"
	chiadapter "github.com/awslabs/aws-lambda-go-api-proxy/chi"
)

func main() {
	// Load environment variables from .env file (useful for local testing).
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables.")
	}

	// Get environment variables.
	transactionsTable := getEnv("DYNAMODB_TRANSACTIONS_TABLE_NAME", "Transactions")
	walletsTable := getEnv("DYNAMODB_WALLETS_TABLE_NAME", "Wallets")
	ledgerTable := getEnv("DYNAMODB_LEDGER_TABLE_NAME", "LedgerEntries")
	sqsQueueURL := getEnv("SQS_QUEUE_URL", "")

	// Load the AWS SDK configuration.
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("backendbest"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create clients.
	dbClient := dynamodb.NewFromConfig(cfg)
	sqsClient := sqs.NewFromConfig(cfg)

	// Initialize components.
	store := dydbstore.New(dbClient, transactionsTable, walletsTable, ledgerTable)
	sqsScheduler := scheduler.NewSQSScheduler(sqsClient, sqsQueueURL)
	apiHandler := handlers.NewApiHandler(store, sqsScheduler)

	// Use oapi-codegen's generated handler to mount the API routes.
	apiRouter := api.Handler(apiHandler)

	// Create a new Chi router and add middleware.
	chiRouter := chi.NewRouter()
	chiRouter.Use(middleware.Logger)
	chiRouter.Use(middleware.Recoverer)
	chiRouter.Mount("/", apiRouter)

	// --- Add Swagger UI endpoint --- //
	spec, err := os.ReadFile("api/spec.yaml")
	if err != nil {
		log.Fatalf("Failed to read OpenAPI spec: %v", err)
	}

	// Create the Swagger UI handler.
	swguiHandler := v5emb.New("Delayed Wallet API", "/docs/openapi.yaml", "/docs")
	chiRouter.Mount("/docs", swguiHandler)

	// Add an endpoint to serve the raw spec file.
	chiRouter.Get("/docs/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-yaml")
		_, _ = w.Write(spec)
	})

	// If we're running in a Lambda environment, use the chiadapter.
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		chiLambda := chiadapter.New(chiRouter)
		lambda.Start(chiLambda.ProxyWithContext)
	} else {
		// Otherwise, start a local HTTP server.
		log.Println("Server starting on port 8080...")
		log.Println("API documentation available at http://localhost:8080/docs")
		if err := http.ListenAndServe(":8080", chiRouter); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}
}

// getEnv reads an environment variable or returns a default value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
