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
  getGetCirclesQueryKey,
  useDeleteCirclesId,
} from '@/lib/api/generated/circles/circles';
import { queryClient } from '@/lib/query-client';
import { useNavigate } from '@tanstack/react-router';
import { useState } from 'react';

interface DissolveCircleDialogProps {
  circleId: string;
  circleName: string;
}

export function DissolveCircleDialog({
  circleId,
  circleName,
}: DissolveCircleDialogProps) {
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const dissolveCircle = useDeleteCirclesId();

  const handleDissolve = async () => {
    setError(null);
    try {
      await dissolveCircle.mutateAsync({ id: circleId });
      await queryClient.invalidateQueries({
        queryKey: getGetCirclesQueryKey(),
      });
      navigate({ to: '/circle' });
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : 'Failed to dissolve this circle. Please try again.'
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
        Dissolve circle
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>
            Dissolve &quot;{circleName}&quot;?
          </AlertDialogTitle>
          <AlertDialogDescription>
            Collaborative Threads return to their creators' personal archives,
            shared-in Moments simply lose their share, and comments/reactions
            made in this circle are removed. This can&apos;t be undone.
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
            onClick={handleDissolve}
            disabled={dissolveCircle.isPending}
          >
            Dissolve
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
