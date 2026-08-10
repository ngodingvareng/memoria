import { ConfirmDestructiveDialog } from '@/components/dialogs/confirm-destructive-dialog';
import { Button } from '@/components/ui/button';
import {
  getGetThreadsQueryKey,
  useDeleteThreadsId,
} from '@/lib/api/generated/threads/threads';
import { queryClient } from '@/lib/query-client';
import { useNavigate } from '@tanstack/react-router';

interface DeleteThreadDialogProps {
  threadId: string;
  threadName: string;
}

export function DeleteThreadDialog({
  threadId,
  threadName,
}: DeleteThreadDialogProps) {
  const navigate = useNavigate();
  const deleteThread = useDeleteThreadsId();

  return (
    <ConfirmDestructiveDialog
      triggerRender={
        <Button type="button" variant="destructive" className="self-start" />
      }
      triggerLabel="Delete thread"
      title={<>Delete &quot;{threadName}&quot;?</>}
      description={
        <>
          This removes the thread and its moments from your archive. This
          can&apos;t be undone from here.
        </>
      }
      confirmLabel="Delete"
      errorFallback="Failed to delete this thread. Please try again."
      onConfirm={async () => {
        await deleteThread.mutateAsync({ id: threadId });
        await queryClient.invalidateQueries({
          queryKey: getGetThreadsQueryKey(),
        });
        navigate({ to: '/thread' });
      }}
    />
  );
}
