package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	chiadapter "github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/chris/delayed-wallet-transactions/pkg/api"
	"github.com/chris/delayed-wallet-transactions/pkg/handlers"
	ws "github.com/chris/delayed-wallet-transactions/pkg/handlers/websockets"
	customMiddleware "github.com/chris/delayed-wallet-transactions/pkg/middleware"
	"github.com/chris/delayed-wallet-transactions/pkg/scheduler"
	dydbstore "github.com/chris/delayed-wallet-transactions/pkg/storage/dynamodb"
	"github.com/chris/delayed-wallet-transactions/pkg/websockets"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github.com/swaggest/swgui/v5emb"

	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
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
	websocketConnectionsTable := getEnv("DYNAMODB_WEBSOCKET_CONNECTIONS_TABLE_NAME", "WebsocketConnections")
	sqsQueueURL := getEnv("SQS_QUEUE_URL", "")
	websocketAPIEndpoint := getEnv("WEBSOCKET_API_ENDPOINT", "")

	// Load the AWS SDK configuration.
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create clients.
	dbClient := dynamodb.NewFromConfig(cfg)
	sqsClient := sqs.NewFromConfig(cfg)

	// Initialize components.
	store := dydbstore.New(dbClient, transactionsTable, walletsTable, ledgerTable, websocketConnectionsTable)
	sqsScheduler := scheduler.NewSQSScheduler(sqsClient, sqsQueueURL)
	publisher, err := websockets.NewPublisher(store, store, websocketAPIEndpoint)
	if err != nil {
		log.Fatalf("failed to create websocket publisher: %v", err)
	}
	apiHandler := handlers.NewApiHandler(store, sqsScheduler, publisher)
	websocketHandler := ws.NewHandler(store)

	// Use oapi-codegen's generated handler to mount the API routes.
	apiRouter := api.Handler(apiHandler)

	// Create a new Chi router and add middleware.
	chiRouter := chi.NewRouter()
	// Set up CORS middleware.
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://chr1sbest.github.io"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any major browsers
	})
	chiRouter.Use(c.Handler)

	// Set up structured logging.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	chiRouter.Use(customMiddleware.NewStructuredLogger(logger))
	chiRouter.Use(middleware.Recoverer)

	// Health check endpoint
	chiRouter.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	chiRouter.Mount("/", apiRouter)

	// --- Add WebSocket endpoint for local development ---
	chiRouter.Get("/ws", websocketHandler.ServeHTTP)

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

	// Start the lambda handler. This will be used for both AWS Lambda and local development.
	lambda.Start(NewCombinedHandler(chiRouter, websocketHandler))
}

// getEnv reads an environment variable or returns a default value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// CombinedHandler can handle both API Gateway (HTTP) and API Gateway v2 (WebSocket) events.
func NewCombinedHandler(httpHandler chi.Router, wsHandler *ws.Handler) func(ctx context.Context, request json.RawMessage) (interface{}, error) {
	chiLambda := chiadapter.New(httpHandler.(*chi.Mux))

	return func(ctx context.Context, request json.RawMessage) (interface{}, error) {
		// Try to unmarshal as a WebSocket request first.
		var wsRequest events.APIGatewayWebsocketProxyRequest
		if err := json.Unmarshal(request, &wsRequest); err == nil && wsRequest.RequestContext.RouteKey != "" {
			switch wsRequest.RequestContext.RouteKey {
			case "$connect":
				return wsHandler.HandleConnect(ctx, wsRequest)
			case "$disconnect":
				return wsHandler.HandleDisconnect(ctx, wsRequest)
			default:
				return wsHandler.HandleDefault(ctx, wsRequest)
			}
		}

		// If it's not a WebSocket request, assume it's an HTTP request.
		var httpRequest events.APIGatewayProxyRequest
		if err := json.Unmarshal(request, &httpRequest); err != nil {
			return nil, err
		}

		return chiLambda.ProxyWithContext(ctx, httpRequest)
	}
}
