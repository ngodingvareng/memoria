import { ConfirmDestructiveDialog } from '@/components/dialogs/confirm-destructive-dialog';
import { Button } from '@/components/ui/button';
import {
  getGetCirclesQueryKey,
  useDeleteCirclesId,
} from '@/lib/api/generated/circles/circles';
import { queryClient } from '@/lib/query-client';
import { useNavigate } from '@tanstack/react-router';

interface DissolveCircleDialogProps {
  circleId: string;
  circleName: string;
}

export function DissolveCircleDialog({
  circleId,
  circleName,
}: DissolveCircleDialogProps) {
  const navigate = useNavigate();
  const dissolveCircle = useDeleteCirclesId();

  return (
    <ConfirmDestructiveDialog
      triggerRender={
        <Button type="button" variant="destructive" className="self-start" />
      }
      triggerLabel="Dissolve circle"
      title={<>Dissolve &quot;{circleName}&quot;?</>}
      description={
        <>
          Collaborative Threads return to their creators' personal archives,
          shared-in Moments simply lose their share, and comments/reactions made
          in this circle are removed. This can&apos;t be undone.
        </>
      }
      confirmLabel="Dissolve"
      errorFallback="Failed to dissolve this circle. Please try again."
      onConfirm={async () => {
        await dissolveCircle.mutateAsync({ id: circleId });
        await queryClient.invalidateQueries({
          queryKey: getGetCirclesQueryKey(),
        });
        navigate({ to: '/circle' });
      }}
    />
  );
}
