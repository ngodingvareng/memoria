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
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { ApiError } from '@/lib/api-client';
import {
  getGetThreadsQueryKey,
  useDeleteThreadsId,
} from '@/lib/api/generated/threads/threads';
import { queryClient } from '@/lib/query-client';
import { useNavigate } from '@tanstack/react-router';
import { useState } from 'react';

interface DeleteThreadDialogProps {
  threadId: string;
  threadName: string;
}

export function DeleteThreadDialog({
  threadId,
  threadName,
}: DeleteThreadDialogProps) {
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const deleteThread = useDeleteThreadsId();

  const handleDelete = async () => {
    setError(null);
    try {
      await deleteThread.mutateAsync({ id: threadId });
      await queryClient.invalidateQueries({
        queryKey: getGetThreadsQueryKey(),
      });
      navigate({ to: '/thread' });
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : 'Failed to delete this thread. Please try again.'
      );
    }
  };

  return (
    <AlertDialog>
      <AlertDialogTrigger
        render={
          <Button type="button" variant="destructive" className="self-start" />
        }
      >
        Delete thread
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete &quot;{threadName}&quot;?</AlertDialogTitle>
          <AlertDialogDescription>
            This removes the thread and its moments from your archive. This
            can&apos;t be undone from here.
          </AlertDialogDescription>
        </AlertDialogHeader>
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            variant="destructive"
            onClick={handleDelete}
            disabled={deleteThread.isPending}
          >
            Delete
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
