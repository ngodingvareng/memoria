import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import Wrapper from '@/components/wrapper';
import { ApiError } from '@/lib/api-client';
import { getGetCirclesQueryKey } from '@/lib/api/generated/circles/circles';
import { usePostCircleJoinRequests } from '@/lib/api/generated/circle-join-requests/circle-join-requests';
import { queryClient } from '@/lib/query-client';
import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { useState } from 'react';

export const Route = createFileRoute('/_app/_circle/c/$id/join/$code')({
  component: RouteComponent,
});

function RouteComponent() {
  const { id, code } = Route.useParams();
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);
  const followLink = usePostCircleJoinRequests();

  const handleJoin = async () => {
    setError(null);
    try {
      const result = await followLink.mutateAsync({ data: { token: code } });
      await queryClient.invalidateQueries({
        queryKey: getGetCirclesQueryKey(),
      });

      if (result.member) {
        navigate({ to: '/c/$id', params: { id } });
      } else {
        setPending(true);
      }
    } catch (err) {
      setError(
        err instanceof ApiError
          ? err.message
          : 'This invite link is no longer valid.'
      );
    }
  };

  return (
    <Wrapper>
      <div className="mx-auto flex max-w-md flex-col items-center gap-4 py-16 text-center">
        <h1 className="text-2xl font-semibold">Join this circle?</h1>

        {pending ? (
          <p className="text-muted-foreground">
            Your request to join has been sent — a member with invite privileges
            needs to approve it before you can see this circle.
          </p>
        ) : (
          <>
            <p className="text-muted-foreground">
              You've been invited to join this circle via link.
            </p>
            {error && (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}
            <Button onClick={handleJoin} disabled={followLink.isPending}>
              Join circle
            </Button>
          </>
        )}
      </div>
    </Wrapper>
  );
}
