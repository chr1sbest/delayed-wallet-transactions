package websockets

import (
	"context"
)

// ConnectionManager defines the interface for managing WebSocket connections.
type ConnectionManager interface {
	AddConnection(ctx context.Context, connectionID string) error
	RemoveConnection(ctx context.Context, connectionID string) error
}

// Publisher defines the interface for publishing messages to WebSocket clients.
type Publisher interface {
	Publish(ctx context.Context, message Message) error
}
