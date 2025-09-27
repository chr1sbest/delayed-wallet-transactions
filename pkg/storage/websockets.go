package storage

import "context"

// WebSocketManager defines the interface for storing and retrieving WebSocket connection IDs.
type WebSocketManager interface {
	AddConnection(ctx context.Context, connectionID string) error
	RemoveConnection(ctx context.Context, connectionID string) error
	GetAllConnections(ctx context.Context) ([]string, error)
}
