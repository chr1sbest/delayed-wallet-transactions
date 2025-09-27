'use client';

import { useEffect } from 'react';
import { toast } from 'sonner';
import { webSocketClient } from '@/client/websocket';

interface WebSocketHandlerProps {
  onWalletUpdate: () => void;
}

export function WebSocketHandler({ onWalletUpdate }: WebSocketHandlerProps) {
  useEffect(() => {
    // Connect to WebSocket and subscribe to messages
    webSocketClient.connect();

    const unsubscribe = webSocketClient.subscribe((message) => {
      console.log('WebSocket message received:', message);

      if (message.type === 'walletUpdate') {
        // Always refetch wallet data to update the UI
        onWalletUpdate();

        const { change, new_balance } = message.payload;

        // Only show a notification if it's not a deduction (i.e., change is positive or zero)
        if (change >= 0) {
          const amount = Math.abs(change);

          toast.info(`Wallet Updated: +${amount} units`, {
            description: `New balance is ${new_balance} units.`,
            style: {
              backgroundColor: 'hsl(142.1 76.2% 36.3%)', // A green color for additions
              color: 'white',
              border: '1px solid hsl(142.1 76.2% 36.3%)',
            },
          });
        }
      }
    });

    // Cleanup on component unmount
    return () => {
      unsubscribe();
      // We can leave the connection open or disconnect if preferred
      // webSocketClient.disconnect();
    };
  }, [onWalletUpdate]);

  // This component does not render anything to the DOM
  return null;
}
