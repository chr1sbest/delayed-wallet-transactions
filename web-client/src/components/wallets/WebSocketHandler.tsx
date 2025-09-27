'use client';

import { useEffect, useRef } from 'react';
import { toast } from 'sonner';
import { webSocketClient, WebSocketMessage } from '@/client/websocket';
import { Wallet } from '@/client';

// Type guard to check if the payload is a valid WalletUpdatePayload
function isWalletUpdatePayload(payload: unknown): payload is { user_id: string; transaction_id?: string; change: number; new_balance: number; } {
  return (
    typeof payload === 'object' &&
    payload !== null &&
    'user_id' in payload &&
    'change' in payload &&
    'new_balance' in payload
  );
}

interface WebSocketHandlerProps {
  wallets: Wallet[];
  onWalletUpdate: () => void;
  onTransactionUpdate: () => void;
}

export function WebSocketHandler({ wallets, onWalletUpdate, onTransactionUpdate }: WebSocketHandlerProps) {
  const processedIds = useRef(new Set<string>());

  useEffect(() => {
    webSocketClient.connect();

    const handleWalletUpdate = (message: WebSocketMessage) => {
      if (message.type === 'walletUpdate' && isWalletUpdatePayload(message.payload)) {
        onWalletUpdate(); // Keep this to refresh wallet balances

        const { user_id, transaction_id, change, new_balance } = message.payload;

        if (transaction_id) {
          onTransactionUpdate(); // Refresh transaction list if a tx is involved
        }

        // Deduplication check
        if (transaction_id && processedIds.current.has(transaction_id)) {
          console.log(`Duplicate message received for transaction ${transaction_id}. Ignoring.`);
          return;
        }

        if (transaction_id) {
          processedIds.current.add(transaction_id);
          // Remove the ID after a short period to allow for legitimate future updates
          setTimeout(() => {
            processedIds.current.delete(transaction_id);
          }, 2000); // 2-second deduplication window
        }

        const wallet = wallets.find((w) => w.user_id === user_id);
        const ownerName = wallet ? wallet.name : 'Unknown';

        if (change > 0) {
          toast.success(`${ownerName}'s wallet was credited!`, {
            description: `+${change} units. New balance: ${new_balance} units.`,
            icon: <div style={{ color: 'oklch(var(--chart-4))' }}>✓</div>,
          });
        }
      }
    };

    const unsubscribe = webSocketClient.subscribe(handleWalletUpdate);

    return () => {
      unsubscribe();
    };
  }, [onWalletUpdate, onTransactionUpdate, wallets]);

  return null;
}
