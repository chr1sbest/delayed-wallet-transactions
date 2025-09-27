'use client';

import { useState } from 'react';
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
      alert('Failed to fetch ledger entries.');
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
        <div className="p-4">
          {isLoading ? (
            <p>Loading ledger...</p>
          ) : ledgerEntries.length > 0 ? (
            <ul className="space-y-2">
              {ledgerEntries.map((entry) => (
                <li key={entry.entry_id} className="border-b pb-2 text-sm">
                  <p><strong>Tx ID:</strong> {entry.transaction_id}</p>
                  <p><strong>Account:</strong> {entry.account_id}</p>
                  <p><strong>Description:</strong> {entry.description}</p>
                  <p><strong>Time:</strong> {entry.timestamp ? new Date(entry.timestamp).toLocaleString() : 'N/A'}</p>
                </li>
              ))}
            </ul>
          ) : (
            <p>No ledger entries found.</p>
          )}
        </div>
      </DrawerContent>
    </Drawer>
  );
}
