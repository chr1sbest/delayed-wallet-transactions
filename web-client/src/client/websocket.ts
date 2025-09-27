const WEBSOCKET_URL = process.env.NEXT_PUBLIC_WEBSOCKET_URL || 'wss://1ex23kq855.execute-api.us-west-2.amazonaws.com/ws';

type MessageCallback = (message: any) => void;

class WebSocketClient {
    private socket: WebSocket | null = null;
    private subscribers: Set<MessageCallback> = new Set();
    private reconnectAttempts = 0;
    private maxReconnectAttempts = 5;
    private reconnectInterval = 3000; // 3 seconds

    public connect(): void {
        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
            console.log("WebSocket is already connected.");
            return;
        }

        this.socket = new WebSocket(WEBSOCKET_URL);

        this.socket.onopen = () => {
            console.log("WebSocket connection established.");
            this.reconnectAttempts = 0; // Reset on successful connection
        };

        this.socket.onmessage = (event) => {
            console.log("Message from server: ", event.data);
            this.handleMessage(event.data);
        };

        this.socket.onclose = (event) => {
            console.log("WebSocket connection closed.", event);
            if (this.reconnectAttempts < this.maxReconnectAttempts) {
                setTimeout(() => {
                    this.reconnectAttempts++;
                    console.log(`Reconnecting... Attempt ${this.reconnectAttempts}`);
                    this.connect();
                }, this.reconnectInterval);
            } else {
                console.error("Max reconnect attempts reached.");
            }
        };

        this.socket.onerror = (error) => {
            console.error("WebSocket error: ", error);
        };
    }

    public disconnect(): void {
        if (this.socket) {
            this.socket.close();
            this.socket = null;
            console.log("WebSocket connection closed by client.");
        }
    }

    private handleMessage(data: any): void {
        // Parse the message if it's a JSON string
        let message;
        try {
            message = JSON.parse(data);
        } catch (error) {
            message = data;
        }

        // Notify all subscribers
        this.subscribers.forEach(callback => callback(message));
    }

    public subscribe(callback: (message: any) => void): () => void {
        this.subscribers.add(callback);
        console.log("New subscriber added.");

        // Return an unsubscribe function
        return () => {
            this.subscribers.delete(callback);
            console.log("Subscriber removed.");
        };
    }

}

export const webSocketClient = new WebSocketClient();
