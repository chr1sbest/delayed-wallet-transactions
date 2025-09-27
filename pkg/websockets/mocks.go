package websockets

import "context"

// NoOpPublisher is a mock publisher that does nothing.
type NoOpPublisher struct{}

// Publish does nothing.
func (p *NoOpPublisher) Publish(ctx context.Context, message Message) error {
	return nil
}
