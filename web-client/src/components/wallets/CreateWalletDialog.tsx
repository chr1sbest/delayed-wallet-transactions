'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { v4 as uuidv4 } from 'uuid';

import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { DefaultService, ApiError } from '@/client';
import { toast } from 'sonner';

const createWalletSchema = z.object({
  user_id: z.string().min(1, 'User ID is required'),
  name: z.string().min(1, 'Wallet name is required'),
});

export function CreateWalletDialog({ onWalletCreated }: { onWalletCreated: () => void }) {
  const [isOpen, setIsOpen] = useState(false);
  const form = useForm<z.infer<typeof createWalletSchema>>({
    resolver: zodResolver(createWalletSchema),
    defaultValues: { user_id: '', name: '' },
  });

  const handleOpenChange = (open: boolean) => {
    if (open) {
      form.reset({ user_id: uuidv4(), name: '' });
    }
    setIsOpen(open);
  };

  const onSubmit = async (values: z.infer<typeof createWalletSchema>) => {
    try {
      await DefaultService.createWallet(values);
      toast.success('Wallet created successfully!');
      form.reset();
      setIsOpen(false);
      onWalletCreated();
    } catch (err) {
      const errorMessage = err instanceof ApiError ? `Failed to create wallet: ${err.body?.message}` : 'An unexpected error occurred.';
      toast.error(errorMessage);
      console.error(err);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button>Create New Wallet</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create a New Wallet</DialogTitle>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField control={form.control} name="name" render={({ field }) => <FormItem><FormLabel>Wallet Name</FormLabel><FormControl><Input placeholder="e.g., My Savings" {...field} /></FormControl><FormMessage /></FormItem>} />
            <Button type="submit" disabled={form.formState.isSubmitting}>{form.formState.isSubmitting ? 'Creating...' : 'Create'}</Button>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
