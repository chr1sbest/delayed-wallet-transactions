package websockets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	apigwtypes "github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi/types"
)

// AllConnectionsGetter defines an interface for getting all connection IDs.
type AllConnectionsGetter interface {
	GetAllConnections(ctx context.Context) ([]string, error)
}

// DefaultPublisher is the default implementation of the Publisher interface.
type DefaultPublisher struct {
	store       AllConnectionsGetter
	connManager ConnectionManager
	apiGwClient *apigatewaymanagementapi.Client
}

// NewPublisher creates a new DefaultPublisher.
func NewPublisher(store AllConnectionsGetter, connManager ConnectionManager, apiEndpoint string) (*DefaultPublisher, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	apiGwClient := apigatewaymanagementapi.NewFromConfig(cfg, func(o *apigatewaymanagementapi.Options) {
		o.BaseEndpoint = aws.String(apiEndpoint)
	})

	return &DefaultPublisher{
		store:       store,
		connManager: connManager,
		apiGwClient: apiGwClient,
	}, nil
}

// Publish sends a message to all connected clients.
func (p *DefaultPublisher) Publish(ctx context.Context, message Message) error {
	connectionIDs, err := p.store.GetAllConnections(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all connections: %w", err)
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	for _, connectionID := range connectionIDs {
		_, err := p.apiGwClient.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: aws.String(connectionID),
			Data:         payload,
		})

		if err != nil {
			var goneErr *apigwtypes.GoneException
			if errors.As(err, &goneErr) {
				slog.Info("stale connection found, deleting", "connectionId", connectionID)
				if err := p.connManager.RemoveConnection(ctx, connectionID); err != nil {
					slog.Error("failed to delete stale connection", "error", err)
				}
			} else {
				slog.Error("failed to post to connection", "connectionId", connectionID, "error", err)
			}
		}
	}

	return nil
}
