'use client';

import { useEffect, useState, useCallback } from 'react';
import { Wallet, DefaultService, ApiError, OpenAPI, Transaction } from '@/client';
import { LedgerDrawer } from '@/components/wallets/LedgerDrawer';
import { CreateWalletDialog } from '@/components/wallets/CreateWalletDialog';
import { NewTransactionDialog } from '@/components/wallets/NewTransactionDialog';
import { WalletList } from '@/components/wallets/WalletList';
import { WebSocketHandler } from '@/components/wallets/WebSocketHandler';
import { Button } from '@/components/ui/button';
import Link from 'next/link';


export default function HomePage() {
  const [wallets, setWallets] = useState<Wallet[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
    const [selectedWallet, setSelectedWallet] = useState<Wallet | null>(null);
  const [updatedTransaction, setUpdatedTransaction] = useState<Transaction | null>(null);

  const fetchWallets = useCallback(async () => {
    try {
      const walletList = await DefaultService.listWallets();
      setWallets(walletList);
      setError(null);
    } catch (err) {
      setError(err instanceof ApiError ? `API Error: ${err.message}` : 'An unexpected error occurred.');
      console.error(err);
    }
  }, []); // Empty dependency array creates a stable function

  useEffect(() => {
    // Initial fetch with loading indicator
    setIsLoading(true);
    fetchWallets().finally(() => setIsLoading(false));
  }, [fetchWallets]);

  return (
    <main className="container mx-auto p-8">
            <WebSocketHandler
        wallets={wallets}
        onWalletUpdate={fetchWallets}
        onTransactionUpdate={(transaction) => {
          fetchWallets(); // Keep the wallet balance refresh
          setUpdatedTransaction(transaction);
        }}
      />
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Wallets</h1>
        <div className="flex items-center space-x-2">
          <Link href="https://github.com/chr1sbest/delayed-wallet-transactions?tab=readme-ov-file#delayed-wallet-transactions" passHref target="_blank">
             <Button variant="ghost">About</Button>
          </Link>
          <LedgerDrawer />
        </div>
      </div>

      {isLoading && <p>Waiting for Lambda to warm..</p>}
      {error && <p className="text-red-500">{error}</p>}

      {!isLoading && !error && (
        <WalletList wallets={wallets} onWalletClick={setSelectedWallet} />
      )}

      <div className="mt-8 flex justify-center">
        <CreateWalletDialog onWalletCreated={fetchWallets} />
      </div>

      {selectedWallet && (
        <NewTransactionDialog
          sourceWallet={selectedWallet}
          allWallets={wallets}
          isOpen={!!selectedWallet}
          onOpenChange={() => setSelectedWallet(null)}
                    onTransactionScheduled={fetchWallets}
          updatedTransaction={updatedTransaction}
        />
      )}
    </main>
  );
}