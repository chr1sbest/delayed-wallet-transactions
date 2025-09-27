package websockets

// MessageType defines the type of a WebSocket message.
type MessageType string

const (
	// MessageTypeWalletUpdate is for messages that update wallet balances.
	MessageTypeWalletUpdate MessageType = "walletUpdate"
)

// Message represents a generic WebSocket message.
type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

// WalletUpdatePayload is the payload for a walletUpdate message.
type WalletUpdatePayload struct {
	UserID        string `json:"user_id"`
	TransactionID string `json:"transaction_id"`
	Change        int64  `json:"change"`
	NewBalance    int64  `json:"new_balance"`
}
