'use client';

import { Wallet } from '@/client';

interface WalletListProps {
  wallets: Wallet[];
  onWalletClick: (wallet: Wallet) => void;
}

export function WalletList({ wallets, onWalletClick }: WalletListProps) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {wallets.length > 0 ? (
        wallets.map((wallet) => (
          <div
            key={wallet.user_id}
            className="border rounded-lg p-4 shadow-sm cursor-pointer hover:shadow-md transition-shadow"
            onClick={() => onWalletClick(wallet)}
          >
            <h2 className="text-xl font-semibold mb-2">{wallet.name}</h2>
            <p className="text-sm text-gray-500">User ID: {wallet.user_id}</p>
            <p className="text-lg font-mono mt-2">Balance: {wallet.balance}</p>
          </div>
        ))
      ) : (
        <p>No wallets found. Create one to get started!</p>
      )}
    </div>
  );
}
