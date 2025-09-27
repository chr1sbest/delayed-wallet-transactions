package websockets

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	apigwtypes "github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi/types"
	"github.com/chris/delayed-wallet-transactions/pkg/storage"
)

// WebsocketHandler handles WebSocket connections and messages.
type WebsocketHandler struct {
	store storage.WebSocketManager
}

// NewWebsocketHandler creates a new WebsocketHandler.
func NewWebsocketHandler(store storage.WebSocketManager) *WebsocketHandler {
	return &WebsocketHandler{
		store: store,
	}
}

// HandleConnect handles new client connections.
func (h *WebsocketHandler) HandleConnect(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	slog.Info("Entering HandleConnect", "connectionId", request.RequestContext.ConnectionID)

	slog.Info("Attempting to add connection to store", "connectionId", request.RequestContext.ConnectionID)
	if err := h.store.AddConnection(ctx, request.RequestContext.ConnectionID); err != nil {
		slog.Error("failed to save connection ID", "connectionId", request.RequestContext.ConnectionID, "error", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}
	slog.Info("Successfully added connection to store", "connectionId", request.RequestContext.ConnectionID)

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

// HandleDisconnect handles client disconnections.
func (h *WebsocketHandler) HandleDisconnect(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	slog.Info("Client disconnected", "connectionId", request.RequestContext.ConnectionID)

	if err := h.store.RemoveConnection(ctx, request.RequestContext.ConnectionID); err != nil {
		slog.Error("failed to delete connection ID", "error", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

// HandleDefault handles messages sent from a client and broadcasts them.
func (h *WebsocketHandler) HandleDefault(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	slog.Info("Received message", "connectionId", request.RequestContext.ConnectionID, "body", request.Body)

	if err := h.Broadcast(ctx, request, []byte(request.Body)); err != nil {
		slog.Error("failed to broadcast message", "error", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

// Broadcast sends a message to all connected clients.
func (h *WebsocketHandler) Broadcast(ctx context.Context, request events.APIGatewayWebsocketProxyRequest, message []byte) error {
	connectionIDs, err := h.store.GetAllConnections(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all connections: %w", err)
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load aws config: %w", err)
	}
	endpoint := fmt.Sprintf("https://%s/%s", request.RequestContext.DomainName, request.RequestContext.Stage)
	apiGwClient := apigatewaymanagementapi.NewFromConfig(cfg, func(o *apigatewaymanagementapi.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	for _, connectionID := range connectionIDs {
		_, err := apiGwClient.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: aws.String(connectionID),
			Data:         message,
		})

		if err != nil {
			var goneErr *apigwtypes.GoneException
			if errors.As(err, &goneErr) {
				slog.Info("stale connection found, deleting", "connectionId", connectionID)
				if err := h.store.RemoveConnection(ctx, connectionID); err != nil {
					slog.Error("failed to delete stale connection", "error", err)
				}
			} else {
				slog.Error("failed to post to connection", "connectionId", connectionID, "error", err)
			}
		}
	}

	return nil
}
