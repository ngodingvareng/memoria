import { ConfirmDestructiveDialog } from '@/components/dialogs/confirm-destructive-dialog';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import Wrapper from '@/components/wrapper';
import { useUserRelationship } from '@/features/users';
import { getApiErrorMessage } from '@/lib/api-client';
import { useGetUserByUsername } from '@/lib/api/generated/users/users';
import { useSession } from '@/lib/session';
import { createFileRoute } from '@tanstack/react-router';
import { useState } from 'react';

export const Route = createFileRoute('/_app/@{$username}')({
  component: RouteComponent,
});

function RouteComponent() {
  const { username } = Route.useParams();
  const session = useSession();
  const [error, setError] = useState<string | null>(null);

  const profileQuery = useGetUserByUsername(username);
  const relationship = useUserRelationship(username);

  const isOwnProfile = session?.user.username === username;

  const handleToggleKnown = async () => {
    setError(null);
    try {
      await relationship.toggleKnown();
    } catch (err) {
      setError(getApiErrorMessage(err, 'Something went wrong. Please try again.'));
    }
  };

  const handleToggleMute = async () => {
    setError(null);
    try {
      await relationship.toggleMute();
    } catch (err) {
      setError(getApiErrorMessage(err, 'Something went wrong. Please try again.'));
    }
  };

  const handleUnblock = async () => {
    setError(null);
    try {
      await relationship.toggleBlock();
    } catch (err) {
      setError(getApiErrorMessage(err, 'Something went wrong. Please try again.'));
    }
  };

  if (profileQuery.isPending) {
    return (
      <Wrapper>
        <div className="flex gap-4">
          <Skeleton className="size-20 rounded-full" />
          <div className="flex flex-col gap-2">
            <Skeleton className="h-6 w-40" />
            <Skeleton className="h-5 w-28" />
          </div>
        </div>
      </Wrapper>
    );
  }

  if (profileQuery.isError || !profileQuery.data) {
    return (
      <Wrapper>
        <Alert variant="destructive">
          <AlertDescription>Couldn't find this user.</AlertDescription>
        </Alert>
      </Wrapper>
    );
  }

  const user = profileQuery.data;

  return (
    <Wrapper>
      <div className="flex gap-4">
        <Avatar className="size-20">
          <AvatarImage src={user.image_path ?? undefined} alt={user.name} />
          <AvatarFallback>{user.name?.charAt(0).toUpperCase()}</AvatarFallback>
        </Avatar>
        <div className="flex flex-col gap-1">
          <div>
            <p className="text-2xl/5 font-medium">{user.name}</p>
            {user.username && (
              <p className="text-muted-foreground text-lg/5">
                @{user.username}
              </p>
            )}
          </div>

          {user.bio && <p className="max-w-prose">{user.bio}</p>}

          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {!isOwnProfile && (
            <div className="flex flex-wrap gap-2">
              <Button
                onClick={handleToggleKnown}
                disabled={relationship.isTogglingKnown}
              >
                {relationship.isKnown
                  ? 'You know this person'
                  : 'I know this person'}
              </Button>

              {relationship.isBlocked ? (
                <Button
                  variant="outline"
                  onClick={handleUnblock}
                  disabled={relationship.isTogglingBlock}
                >
                  Unblock
                </Button>
              ) : (
                <ConfirmDestructiveDialog
                  triggerRender={<Button variant="outline" />}
                  triggerLabel="Block"
                  title={`Block ${user.name}?`}
                  description="Neither of you will be able to see, mention, comment on, or react to the other's Moments, in any context."
                  confirmLabel="Block"
                  errorFallback="Failed to block this user. Please try again."
                  onConfirm={relationship.toggleBlock}
                />
              )}

              <Button
                variant="outline"
                onClick={handleToggleMute}
                disabled={relationship.isTogglingMute}
              >
                {relationship.isMuted ? 'Unmute' : 'Mute'}
              </Button>
            </div>
          )}
        </div>
      </div>
    </Wrapper>
  );
}
