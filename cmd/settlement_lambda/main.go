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
	dynamo_store "github.com/chris/delayed-wallet-transactions/pkg/storage/dynamodb"
	"github.com/chris/delayed-wallet-transactions/pkg/websockets"
)

var store *dynamo_store.Store
var publisher websockets.Publisher

func init() {
	// Initialize dependencies once.
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	dbClient := dynamodb.NewFromConfig(cfg)
	store = dynamo_store.New(dbClient, os.Getenv("DYNAMODB_TRANSACTIONS_TABLE_NAME"), os.Getenv("DYNAMODB_WALLETS_TABLE_NAME"), os.Getenv("DYNAMODB_LEDGER_TABLE_NAME"), os.Getenv("DYNAMODB_WEBSOCKET_CONNECTIONS_TABLE_NAME"))

	publisher, err = websockets.NewPublisher(store, store, os.Getenv("WEBSOCKET_API_ENDPOINT"))
	if err != nil {
		log.Fatalf("failed to create websocket publisher: %v", err)
	}
}

// HandleRequest processes SQS messages and settles the transactions.
func HandleRequest(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, message := range sqsEvent.Records {
		log.Printf("Processing message %s", message.MessageId)

		var tx models.Transaction
		if err := json.Unmarshal([]byte(message.Body), &tx); err != nil {
			log.Printf("ERROR: failed to unmarshal transaction from SQS message %s: %v", message.MessageId, err)
			continue
		}

		if err := store.SettleTransaction(ctx, &tx); err != nil {
			log.Printf("error settling transaction: %v", err)
			continue
		}

		// Publish wallet update messages
		go func() {
			// Get the latest wallet balances.
			fromWallet, err := store.GetWallet(context.Background(), tx.FromUserId)
			if err != nil {
				log.Printf("ERROR: failed to get wallet for websocket message: %v", err)
				return
			}
			toWallet, err := store.GetWallet(context.Background(), tx.ToUserId)
			if err != nil {
				log.Printf("ERROR: failed to get wallet for websocket message: %v", err)
				return
			}

			// Message for the sender
			fromMsg := websockets.Message{
				Type: websockets.MessageTypeWalletUpdate,
				Payload: websockets.WalletUpdatePayload{
					UserID:        tx.FromUserId,
					TransactionID: tx.Id,
					Change:        -tx.Amount,
					NewBalance:    fromWallet.Balance,
				},
			}
			if err := publisher.Publish(context.Background(), fromMsg); err != nil {
				log.Printf("ERROR: failed to publish websocket message: %v", err)
			}

			// Message for the recipient
			toMsg := websockets.Message{
				Type: websockets.MessageTypeWalletUpdate,
				Payload: websockets.WalletUpdatePayload{
					UserID:        tx.ToUserId,
					TransactionID: tx.Id,
					Change:        tx.Amount,
					NewBalance:    toWallet.Balance,
				},
			}
			if err := publisher.Publish(context.Background(), toMsg); err != nil {
				log.Printf("ERROR: failed to publish websocket message: %v", err)
			}
		}()
	}
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
