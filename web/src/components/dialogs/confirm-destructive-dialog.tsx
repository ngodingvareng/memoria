import * as React from 'react';
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
import { getApiErrorMessage } from '@/lib/api-client';

interface ConfirmDestructiveDialogProps {
  // Renders an AlertDialogTrigger wrapping this element; omit to drive the
  // dialog fully via `open`/`onOpenChange` from an external trigger (e.g. a
  // dropdown menu item).
  triggerRender?: React.ReactElement;
  triggerLabel?: React.ReactNode;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  title: React.ReactNode;
  description: React.ReactNode;
  confirmLabel: string;
  cancelLabel?: string;
  errorFallback: string;
  onConfirm: () => Promise<void>;
}

export function ConfirmDestructiveDialog({
  triggerRender,
  triggerLabel,
  open,
  onOpenChange,
  title,
  description,
  confirmLabel,
  cancelLabel = 'Cancel',
  errorFallback,
  onConfirm,
}: ConfirmDestructiveDialogProps) {
  const [uncontrolledOpen, setUncontrolledOpen] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);
  const [isPending, setIsPending] = React.useState(false);

  const isOpen = open ?? uncontrolledOpen;
  const setOpen = onOpenChange ?? setUncontrolledOpen;

  const handleOpenChange = (next: boolean) => {
    if (isPending) return;
    if (!next) setError(null);
    setOpen(next);
  };

  const handleConfirm = async () => {
    setError(null);
    setIsPending(true);
    try {
      await onConfirm();
      setOpen(false);
    } catch (err) {
      setError(getApiErrorMessage(err, errorFallback));
    } finally {
      setIsPending(false);
    }
  };

  return (
    <AlertDialog open={isOpen} onOpenChange={handleOpenChange}>
      {triggerRender && (
        <AlertDialogTrigger render={triggerRender}>
          {triggerLabel}
        </AlertDialogTrigger>
      )}
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{title}</AlertDialogTitle>
          <AlertDialogDescription>{description}</AlertDialogDescription>
        </AlertDialogHeader>
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isPending}>
            {cancelLabel}
          </AlertDialogCancel>
          <AlertDialogAction
            variant="destructive"
            onClick={handleConfirm}
            disabled={isPending}
          >
            {confirmLabel}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
