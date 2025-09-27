package websockets

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/chris/delayed-wallet-transactions/pkg/websockets"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Handler handles WebSocket connections.
type Handler struct {
	connManager websockets.ConnectionManager
}

// NewHandler creates a new Handler.
func NewHandler(connManager websockets.ConnectionManager) *Handler {
	return &Handler{
		connManager: connManager,
	}
}

// HandleConnect handles new client connections.
func (h *Handler) HandleConnect(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	slog.Info("Client connected", "connectionId", request.RequestContext.ConnectionID)

	if err := h.connManager.AddConnection(ctx, request.RequestContext.ConnectionID); err != nil {
		slog.Error("failed to save connection ID", "error", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

// HandleDisconnect handles client disconnections.
func (h *Handler) HandleDisconnect(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	slog.Info("Client disconnected", "connectionId", request.RequestContext.ConnectionID)

	if err := h.connManager.RemoveConnection(ctx, request.RequestContext.ConnectionID); err != nil {
		slog.Error("failed to delete connection ID", "error", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

// HandleDefault handles messages sent from a client.
func (h *Handler) HandleDefault(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	slog.Info("Received message", "connectionId", request.RequestContext.ConnectionID, "body", request.Body)
	// We don't expect clients to send messages, but we log them just in case.
	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections by default for local development.
		return true
	},
}

// ServeHTTP handles WebSocket requests for the local development server.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("failed to upgrade connection", "error", err)
		return
	}
	defer conn.Close()

	// Generate a unique connection ID for local connections.
	connectionID := uuid.New().String()
	slog.Info("Client connected locally", "connectionId", connectionID)

	ctx := r.Context()
	if err := h.connManager.AddConnection(ctx, connectionID); err != nil {
		slog.Error("failed to save local connection ID", "error", err)
		return
	}

	// When the function returns (i.e., the client disconnects), remove the connection.
	defer func() {
		slog.Info("Client disconnected locally", "connectionId", connectionID)
		if err := h.connManager.RemoveConnection(ctx, connectionID); err != nil {
			slog.Error("failed to delete local connection ID", "error", err)
		}
	}()

	// Keep the connection alive, waiting for the client to disconnect.
	// The server doesn't process incoming messages in this implementation,
	// but this loop is necessary to detect when the client closes the connection.
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("unexpected close error", "error", err)
			}
			break // Exit the loop on any error, which signifies a disconnection.
		}
	}
}
