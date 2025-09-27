'use client';

import { useState, useEffect, useCallback } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';

import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { DefaultService, Wallet, ApiError, NewTransaction, Transaction } from '@/client';
import { StatusBadge } from './StatusBadge';
import { CountdownTimer } from './CountdownTimer';
import { toast } from 'sonner';

const newTransactionSchema = z.object({
    to_user_id: z.string().min(1, 'Destination wallet is required.'),
    amount: z.string().min(1, 'Amount is required.'),
    delay_seconds: z.string().optional(),
});

type TransactionFormValues = z.infer<typeof newTransactionSchema>;

const CANCELLABLE_STATUSES: Array<Transaction['status']> = [
  Transaction.status.RESERVED,
  Transaction.status.PENDING_APPROVAL,
  Transaction.status.APPROVED,
];

const NON_CANCELLABLE_STATUSES: Array<Transaction['status']> = [
  Transaction.status.COMPLETED,
  Transaction.status.REJECTED,
  Transaction.status.FAILED,
];

export function NewTransactionDialog({ sourceWallet, allWallets, isOpen, onOpenChange, onTransactionScheduled, updatedTransaction }: {
  sourceWallet: Wallet;
  allWallets: Wallet[];
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
    onTransactionScheduled: () => void;
  updatedTransaction: Partial<Transaction> | null;
}) {
  const [view, setView] = useState<'form' | 'transactions'>('form');
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [transactionToCancel, setTransactionToCancel] = useState<Transaction | null>(null);

  const form = useForm<TransactionFormValues>({
    resolver: zodResolver(newTransactionSchema),
    defaultValues: { to_user_id: '', amount: '', delay_seconds: '' },
  });

  const destinationWallets = allWallets.filter(w => w.user_id && w.user_id !== sourceWallet.user_id);

  const showTransactions = useCallback(async () => {
    if (!sourceWallet.user_id) return;
    setIsLoading(true);
    try {
      const userTransactions = await DefaultService.listTransactionsByUserId(sourceWallet.user_id);
      setTransactions(userTransactions);
      setView('transactions');
    } catch (err) {
      toast.error('Failed to fetch transactions.');
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  }, [sourceWallet.user_id]);

  useEffect(() => {
    if (view === 'transactions') {
      showTransactions();
    }
    }, [view, showTransactions]);

  useEffect(() => {
    if (updatedTransaction && updatedTransaction.id) {
      setTransactions(prevTransactions =>
        prevTransactions.map(tx =>
          tx.id === updatedTransaction.id
            ? { ...tx, status: updatedTransaction.status || tx.status }
            : tx
        )
      );
    }
  }, [updatedTransaction]);

  const onSubmit = async (values: TransactionFormValues) => {
    const amount = parseInt(values.amount, 10);
    const delay_seconds = values.delay_seconds ? parseInt(values.delay_seconds, 10) : 0;

    if (isNaN(amount) || amount <= 0) {
      toast.error('Amount must be a positive number.');
      return;
    }
    if (isNaN(delay_seconds) || delay_seconds < 0 || delay_seconds > 900) {
      toast.error('Delay must be between 0 and 900 seconds.');
      return;
    }
    if (!sourceWallet.user_id) {
      toast.error('Source wallet has no ID!');
      return;
    }

    const transactionData: NewTransaction = {
      from_user_id: sourceWallet.user_id,
      to_user_id: values.to_user_id,
      amount: amount,
      delay_seconds: delay_seconds,
    };

    try {
      await DefaultService.scheduleTransaction(transactionData);
      toast.success('Transaction scheduled successfully!');
      setView('transactions');
      onTransactionScheduled();
    } catch (err) {
      let errorMessage = 'An unexpected error occurred.';
      if (err instanceof ApiError) {
        const body = err.body as { message?: string };
        errorMessage = `Failed to schedule transaction: ${body?.message || err.statusText}`;
      } else if (err instanceof Error) {
        errorMessage = err.message;
      }
      toast.error(errorMessage);
      console.error(err);
    }
  };

  const handleCancelTransaction = async () => {
    if (!transactionToCancel || !transactionToCancel.id) return;
    const txIdToCancel = transactionToCancel.id;

    try {
      await DefaultService.cancelTransactionById(txIdToCancel);

      setTransactions(prevTransactions =>
        prevTransactions.map(tx =>
          tx.id === txIdToCancel ? { ...tx, status: Transaction.status.REJECTED } : tx
        )
      );

      onTransactionScheduled();
      
      toast.success('Transaction canceled successfully!');
      setTransactionToCancel(null);
    } catch (err) {
      let errorMessage = 'An unexpected error occurred.';
      if (err instanceof ApiError) {
        const body = err.body as { message?: string };
        errorMessage = `Failed to cancel transaction: ${body?.message}`;
      } else if (err instanceof Error) {
        errorMessage = err.message;
      }
      toast.error(errorMessage);
      console.error(err);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="w-full max-w-lg">
        <DialogHeader>
          <DialogTitle>New Transaction from {sourceWallet.name}</DialogTitle>
          <DialogDescription>Balance: {sourceWallet.balance}</DialogDescription>
        </DialogHeader>
        
        {view === 'form' ? (
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="to_user_id"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>To</FormLabel>
                    <Select onValueChange={field.onChange} defaultValue={field.value}>
                      <FormControl><SelectTrigger><SelectValue placeholder="Select a destination wallet" /></SelectTrigger></FormControl>
                      <SelectContent>
                        {destinationWallets.map(w => (
                          <SelectItem key={w.user_id} value={w.user_id!}>{w.name} ({w.user_id})</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField control={form.control} name="amount" render={({ field }) => <FormItem><FormLabel>Amount</FormLabel><FormControl><Input type="number" placeholder="0" {...field} /></FormControl><FormMessage /></FormItem>} />
              <FormField control={form.control} name="delay_seconds" render={({ field }) => <FormItem><FormLabel>Delay (seconds)</FormLabel><FormControl><Input type="number" placeholder="0" {...field} /></FormControl><FormMessage /></FormItem>} />
              <div className="flex flex-col sm:flex-row sm:justify-between gap-2">
                <Button type="submit" disabled={form.formState.isSubmitting}>{form.formState.isSubmitting ? 'Scheduling...' : 'Schedule Transaction'}</Button>
                <Button type="button" variant="outline" onClick={showTransactions}>Outgoing Transactions</Button>
              </div>
            </form>
          </Form>
        ) : (
          <div>
            <Button variant="outline" onClick={() => setView('form')} className="mb-4">Back to Form</Button>
            <h3 className="text-lg font-semibold mb-2">Outgoing Transactions</h3>
            {isLoading ? (
              <p>Loading transactions...</p>
            ) : transactions.length > 0 ? (
              <ul className="space-y-2 max-h-[40vh] sm:max-h-60 overflow-y-auto">
                {transactions.map(tx => (
                  <li key={tx.id} className="flex justify-between items-center border-b pb-2 text-sm">
                    <div>
                      <p><strong>To:</strong> {tx.to_user_id}</p>
                      <p><strong>Amount:</strong> {tx.amount}</p>
                      <div className="flex items-center gap-2">
                        <strong>Status:</strong>
                        <StatusBadge status={tx.status} />
                        {tx.created_at && tx.delay_seconds && tx.delay_seconds > 0 && CANCELLABLE_STATUSES.includes(tx.status) && (
                          <CountdownTimer executionTime={new Date(new Date(tx.created_at).getTime() + tx.delay_seconds * 1000)} />
                        )}
                      </div>
                    </div>
                    {tx.status && CANCELLABLE_STATUSES.includes(tx.status) && (
                      <Button variant="destructive" size="sm" onClick={() => setTransactionToCancel(tx)}>Cancel</Button>
                    )}
                  </li>
                ))}
              </ul>
            ) : (
              <p>No outgoing transactions found.</p>
            )}
          </div>
        )}

        <AlertDialog open={!!transactionToCancel} onOpenChange={() => setTransactionToCancel(null)}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Are you sure?</AlertDialogTitle>
              <AlertDialogDescription>
                This action will cancel the transaction. This cannot be undone.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Back</AlertDialogCancel>
              <AlertDialogAction onClick={handleCancelTransaction}>Yes, cancel</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </DialogContent>
    </Dialog>
  );
}