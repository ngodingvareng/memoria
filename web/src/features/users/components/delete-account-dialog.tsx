import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { getApiErrorMessage } from '@/lib/api-client';
import { useDeleteUsersMe } from '@/lib/api/generated/users/users';
import { setSession } from '@/lib/session';
import { useNavigate } from '@tanstack/react-router';
import * as React from 'react';

const CONFIRMATION_WORD = 'DELETE';

// Not built on ConfirmDestructiveDialog — that component has no slot
// for the extra confirmation input this needs between the description
// and the footer, so the AlertDialog primitives are composed directly
// here instead (same primitives DissolveCircleDialog's simpler case
// gets away with reusing).
export function DeleteAccountDialog() {
  const navigate = useNavigate();
  const [open, setOpen] = React.useState(false);
  const [confirmation, setConfirmation] = React.useState('');
  const [error, setError] = React.useState<string | null>(null);
  const deleteAccount = useDeleteUsersMe();

  const handleOpenChange = (next: boolean) => {
    if (deleteAccount.isPending) return;
    if (!next) {
      setConfirmation('');
      setError(null);
    }
    setOpen(next);
  };

  const handleConfirm = async () => {
    setError(null);
    try {
      await deleteAccount.mutateAsync({
        data: { confirmation: CONFIRMATION_WORD },
      });
      // Only on success — a failed delete must not sign the user out of
      // their still-existing account, unlike logout's own unconditional
      // clear.
      setSession(null);
      navigate({ to: '/signin' });
    } catch (err) {
      setError(
        getApiErrorMessage(err, 'Failed to delete account. Please try again.')
      );
    }
  };

  return (
    <AlertDialog open={open} onOpenChange={handleOpenChange}>
      <AlertDialogTrigger
        render={<Button type="button" variant="destructive" />}
      >
        Delete account
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete your account?</AlertDialogTitle>
          <AlertDialogDescription>
            Your Moments, Threads, and photos are deleted. Your comments and
            reactions on other people's Moments stay, attributed to a former
            member. Circles you're the sole admin of are handed off to another
            member, or dissolved if you're the only one left. This can't be
            undone.
          </AlertDialogDescription>
        </AlertDialogHeader>

        <Input
          placeholder={`Type "${CONFIRMATION_WORD}" to confirm`}
          value={confirmation}
          onChange={(e) => setConfirmation(e.target.value)}
          autoComplete="off"
        />

        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        <AlertDialogFooter>
          <AlertDialogCancel disabled={deleteAccount.isPending}>
            Cancel
          </AlertDialogCancel>
          <AlertDialogAction
            variant="destructive"
            onClick={handleConfirm}
            disabled={
              confirmation !== CONFIRMATION_WORD || deleteAccount.isPending
            }
          >
            Delete account
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
