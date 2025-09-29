'use client';

import { useState } from 'react';
import { toast } from 'sonner';
import { Button } from '@/components/ui/button';
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from '@/components/ui/drawer';
import { DefaultService, LedgerEntry } from '@/client';

export function LedgerDrawer() {
  const [ledgerEntries, setLedgerEntries] = useState<LedgerEntry[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const fetchLedger = async () => {
    setIsLoading(true);
    try {
      const entries = await DefaultService.listLedgerEntries();
      setLedgerEntries(entries);
    } catch (err) {
      console.error('Failed to fetch ledger entries:', err);
      toast.error('Failed to fetch ledger entries.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Drawer>
      <DrawerTrigger asChild>
        <Button variant="outline" onClick={fetchLedger}>Ledger</Button>
      </DrawerTrigger>
      <DrawerContent>
        <DrawerHeader>
          <DrawerTitle>Recent Ledger Entries</DrawerTitle>
          <DrawerDescription>Showing the most recent ledger activities.</DrawerDescription>
        </DrawerHeader>
        <div className="px-4 max-h-[60vh] overflow-y-auto">
          {isLoading ? (
            <p className="text-center">Loading ledger...</p>
          ) : ledgerEntries.length > 0 ? (
            <ul className="space-y-4">
              {ledgerEntries.map((entry) => {
                const isCredit = entry.credit && entry.credit > 0;
                const amount = isCredit ? entry.credit : entry.debit;
                const amountColor = isCredit ? 'text-green-500' : 'text-red-500';
                const amountPrefix = isCredit ? '+' : '-';

                return (
                  <li key={entry.entry_id} className="border-b pb-3 text-sm flex items-center gap-4">
                    <p className={`font-bold text-lg w-20 sm:w-24 text-right ${amountColor}`}>
                      {amountPrefix}{amount}
                    </p>
                    <div className="truncate">
                      <p className="font-semibold truncate">{entry.description}</p>
                      <p className="text-xs text-gray-500">Tx: {entry.transaction_id?.substring(0, 8)}...</p>
                      <p className="text-xs text-gray-500">{entry.timestamp ? new Date(entry.timestamp).toLocaleString() : 'N/A'}</p>
                    </div>
                  </li>
                );
              })}
            </ul>
          ) : (
            <p className="text-center">No ledger entries found.</p>
          )}
        </div>
      </DrawerContent>
    </Drawer>
  );
}
