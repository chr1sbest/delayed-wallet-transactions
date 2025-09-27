'use client';

import { useEffect, useState } from 'react';
import { Wallet, DefaultService, ApiError, OpenAPI } from '@/client';
import { LedgerDrawer } from '@/components/wallets/LedgerDrawer';
import { CreateWalletDialog } from '@/components/wallets/CreateWalletDialog';
import { NewTransactionDialog } from '@/components/wallets/NewTransactionDialog';
import { WalletList } from '@/components/wallets/WalletList';

// Configure the API client base URL
OpenAPI.BASE = 'http://localhost:8080';

export default function HomePage() {
  const [wallets, setWallets] = useState<Wallet[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedWallet, setSelectedWallet] = useState<Wallet | null>(null);

  const fetchWallets = async () => {
    try {
      setIsLoading(true);
      const walletList = await DefaultService.listWallets();
      setWallets(walletList);
      setError(null);
    } catch (err) {
      setError(err instanceof ApiError ? `API Error: ${err.message}` : 'An unexpected error occurred.');
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchWallets();
  }, []);

  return (
    <main className="container mx-auto p-8">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Wallets</h1>
        <LedgerDrawer />
      </div>

      {isLoading && <p>Loading wallets...</p>}
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
        />
      )}
    </main>
  );
}


