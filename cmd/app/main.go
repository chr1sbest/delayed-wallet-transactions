package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	chiadapter "github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/chris/delayed-wallet-transactions/pkg/api"
	"github.com/chris/delayed-wallet-transactions/pkg/handlers"
	"github.com/chris/delayed-wallet-transactions/pkg/scheduler"
	dydbstore "github.com/chris/delayed-wallet-transactions/pkg/storage/dynamodb"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/swaggest/swgui/v5emb"

	"github.com/aws/aws-lambda-go/lambda"
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

	// --- Add Swagger UI endpoint for local development --- //
	// This will only work locally and will not be available in the deployed Lambda
	// because the spec file is not included in the build artifact.
	chiRouter.Get("/docs/*", func(w http.ResponseWriter, r *http.Request) {
		// Read the spec file on every request to ensure it's up to date.
		spec, err := os.ReadFile("api/spec.yaml")
		if err != nil {
			http.Error(w, "Failed to read OpenAPI spec", http.StatusInternalServerError)
			return
		}

		// Create a new Swagger UI handler on each request.
		swguiHandler := v5emb.New("Delayed Wallet API", "/docs/openapi.yaml", "/docs")

		// Serve the raw spec file.
		chiRouter.Get("/docs/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/x-yaml")
			_, _ = w.Write(spec)
		})

		swguiHandler.ServeHTTP(w, r)
	})

	// If we're running in a Lambda environment, use the chiadapter.
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		chiLambda := chiadapter.New(chiRouter)
		lambda.Start(chiLambda.ProxyWithContext)
	} else {
		// Otherwise, start a local HTTP server.
		log.Println("Server starting on port 8080...")
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
